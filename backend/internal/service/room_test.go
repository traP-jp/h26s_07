package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
	"github.com/traP-jp/h26_07/backend/internal/testutil"
)

type testEvents struct {
	events []any
	fail   bool
	check  func()
}

func (e *testEvents) SendRoom(_ context.Context, _ model.RoomID, event any) error {
	if e.check != nil {
		e.check()
	}
	e.events = append(e.events, event)
	if e.fail {
		return errors.New("broadcast failed")
	}
	return nil
}
func (e *testEvents) SendParticipant(ctx context.Context, id model.RoomID, _ model.UserID, event any) error {
	return e.SendRoom(ctx, id, event)
}

func testService(t *testing.T) (*RoomService, *repository.GormRoomRepository, *testEvents, model.RoomID) {
	t.Helper()
	db := testutil.NewRoomDB(t)
	repo := repository.NewGormRoomRepository(db)
	events := &testEvents{}
	service := NewRoomService(repository.NewGormTransactionRunner(db), repo, events)
	room, err := service.CreateRoom(context.Background(), model.RoomSettings{Name: "test"}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.PostParticipants(context.Background(), room.RoomID, "alice"); err != nil {
		t.Fatal(err)
	}
	events.events = nil
	return service, repo, events, room.RoomID
}

func TestJoinBroadcastsParticipantCountOnce(t *testing.T) {
	s, _, events, id := testService(t)
	ctx := context.Background()
	mustSucceed(t, s.PostParticipants(ctx, id, "bob"))
	if len(events.events) != 1 {
		t.Fatalf("got %d events, want 1", len(events.events))
	}
	display, ok := events.events[0].(openapi.DisplayParticipantJoinedEvent)
	if !ok || display.Body.Participant.UserID != "bob" || display.Body.ParticipantCount != 2 {
		t.Fatalf("unexpected display event: %#v", events.events[0])
	}
	events.events = nil
	mustSucceed(t, s.PostParticipants(ctx, id, "bob"))
	if len(events.events) != 0 {
		t.Fatal("repeat join broadcast an event")
	}
}
func mustSucceed(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentMessagesAndPicks(t *testing.T) {
	s, repo, _, id := testService(t)
	ctx := context.Background()
	mustSucceed(t, s.StartGame(ctx, id, "admin"))
	before, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	var wg sync.WaitGroup
	errs := make(chan error, 301)
	for user := 0; user < 10; user++ {
		wg.Add(1)
		go func(user int) {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				_, err := s.PostMessage(ctx, id, "alice", fmt.Sprintf("%d:%d", user, i))
				if err != nil {
					errs <- err
				}
			}
		}(user)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			if err := s.StartPick(ctx, id, "admin"); err != nil {
				errs <- err
				return
			}
			if err := s.FinishPick(ctx, id, "admin"); err != nil {
				errs <- err
				return
			}
		}
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	room, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if len(room.Messages) != 300 || len(room.PickedBalls) != 30 || room.PickState != model.RoomPickStateIdle || len(room.Cards) != 1 || len(room.Participants) != 1 {
		t.Fatalf("lost data: messages=%d picks=%d cards=%d participants=%d state=%s", len(room.Messages), len(room.PickedBalls), len(room.Cards), len(room.Participants), room.PickState)
	}
	seen := map[string]bool{}
	for _, message := range room.Messages {
		if seen[message.Content] {
			t.Fatal("duplicate message")
		}
		seen[message.Content] = true
	}
	if room.Cards[0].CardID != before.Cards[0].CardID {
		t.Fatal("card replaced")
	}
	for i, cell := range room.Cards[0].Cells {
		if !reflect.DeepEqual(cell.Number, before.Cards[0].Cells[i].Number) {
			t.Fatal("cell number changed")
		}
		wantOpen := cell.Number == nil || slices.Contains(room.PickedBalls, *cell.Number)
		if (cell.CellState == model.CardCellStateOpen || cell.CellState == model.CardCellStateBingo) != wantOpen {
			t.Fatalf("cell %d inconsistent with picks", i)
		}
	}
}

func TestRetrySafeControlsAndEventOrder(t *testing.T) {
	s, repo, events, id := testService(t)
	ctx := context.Background()
	mustSucceed(t, s.StartGame(ctx, id, "admin"))
	before, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	events.events = nil
	mustSucceed(t, s.StartGame(ctx, id, "admin"))
	after, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if !reflect.DeepEqual(before, after) || len(events.events) != 0 {
		t.Fatal("start retry changed room")
	}
	mustSucceed(t, s.PostParticipants(ctx, id, "alice"))
	mustSucceed(t, s.CancelPick(ctx, id, "admin"))
	if err := s.FinishPick(ctx, id, "admin"); !errors.Is(err, model.ErrRoomPickNotFinishable) {
		t.Fatalf("finish idle: %v", err)
	}
	mustSucceed(t, s.StartPick(ctx, id, "admin"))
	events.events = nil
	mustSucceed(t, s.StartPick(ctx, id, "admin"))
	if len(events.events) != 2 {
		t.Fatal("start retry did not rebroadcast")
	}
	if _, ok := events.events[0].(openapi.DisplayPickStartedEvent); !ok {
		t.Fatal("missing display start")
	}
	if _, ok := events.events[1].(openapi.ParticipantPickStartedEvent); !ok {
		t.Fatal("missing participant start")
	}
	for _, control := range []func(context.Context, model.RoomID, model.UserID) error{s.StartGame, s.FinishGame, s.StartPick, s.CancelPick} {
		if err := control(ctx, id, "alice"); !errors.Is(err, model.ErrRoomForbidden) {
			t.Fatalf("permission error: %v", err)
		}
	}
	events.events = nil
	mustSucceed(t, s.FinishGame(ctx, id, "admin"))
	if len(events.events) != 4 {
		t.Fatalf("events: %d", len(events.events))
	}
	if _, ok := events.events[0].(openapi.DisplayPickCanceledEvent); !ok {
		t.Fatal("missing display cancel")
	}
	if _, ok := events.events[1].(openapi.ParticipantPickCanceledEvent); !ok {
		t.Fatal("missing participant cancel")
	}
	if _, ok := events.events[2].(openapi.ParticipantGameFinishedEvent); !ok {
		t.Fatal("missing participant finish")
	}
	if _, ok := events.events[3].(openapi.DisplayGameFinishedEvent); !ok {
		t.Fatal("missing display finish")
	}
	events.events = nil
	mustSucceed(t, s.FinishGame(ctx, id, "admin"))
	mustSucceed(t, s.PostParticipants(ctx, id, "alice"))
	if len(events.events) != 0 {
		t.Fatal("finish retry broadcast")
	}
}

func TestBroadcastFailureAfterCommit(t *testing.T) {
	s, repo, events, id := testService(t)
	ctx := context.Background()
	events.fail = true
	events.check = func() {
		room, err := repo.FindByID(ctx, id)
		mustSucceed(t, err)
		if len(room.Messages) != 1 {
			t.Fatal("broadcast happened before commit")
		}
	}
	message, err := s.PostMessage(ctx, id, "alice", "hello")
	mustSucceed(t, err)
	if message == nil || len(events.events) != 2 {
		t.Fatal("broadcast failure interrupted remaining events")
	}
}

type collisionRepository struct {
	repository.RoomRepository
	checks int
}

func (r *collisionRepository) ActiveCodeExists(ctx context.Context, code model.RoomCode) (bool, error) {
	r.checks++
	if r.checks == 1 {
		return true, nil
	}
	return r.RoomRepository.ActiveCodeExists(ctx, code)
}
func TestCreateRetriesActiveCodeCollision(t *testing.T) {
	s, repo, _, _ := testService(t)
	collision := &collisionRepository{RoomRepository: repo}
	s.roomRepository = collision
	_, err := s.CreateRoom(context.Background(), model.RoomSettings{Name: "second"}, "admin")
	mustSucceed(t, err)
	if collision.checks != 2 {
		t.Fatalf("code checks: %d", collision.checks)
	}
}

func TestRoomLockCoversBroadcastAndSnapshotRegistration(t *testing.T) {
	s, repo, events, id := testService(t)
	ctx := context.Background()
	// Simulate snapshot loading and registration while a mutation waits.
	unlock := s.LockRoom(id)
	snapshot, err := s.GetRoom(ctx, id)
	mustSucceed(t, err)
	done := make(chan error, 1)
	go func() { _, err := s.PostMessage(ctx, id, "alice", "first"); done <- err }()
	registered := false
	events.check = func() {
		if !registered {
			t.Error("broadcast before registration")
		}
	}
	if len(snapshot.Messages) != 0 {
		t.Fatal("unexpected snapshot")
	}
	registered = true
	unlock()
	mustSucceed(t, <-done)
	// A blocked broadcaster must prevent the next mutation from committing.
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	events.check = func() { once.Do(func() { close(entered); <-release }) }
	go func() { _, err := s.PostMessage(ctx, id, "alice", "second"); done <- err }()
	<-entered
	room, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if len(room.Messages) != 2 {
		t.Fatal("broadcast before commit")
	}
	next := make(chan error, 1)
	go func() { _, err := s.PostMessage(ctx, id, "alice", "third"); next <- err }()
	// The room mutex remains held until both events have been broadcast.
	s.locks.mu.Lock()
	entry := s.locks.entries[id]
	if entry == nil {
		t.Error("room lock released during broadcast")
	} else if entry.mu.TryLock() {
		entry.mu.Unlock()
		t.Error("room lock not held during broadcast")
	}
	s.locks.mu.Unlock()
	close(release)
	mustSucceed(t, <-done)
	mustSucceed(t, <-next)
	room, err = repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if len(room.Messages) != 3 {
		t.Fatal("mutation lost after broadcast")
	}
}

type failingTransaction struct{ repository.TransactionRunner }

func (r failingTransaction) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return r.TransactionRunner.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := fn(ctx); err != nil {
			return err
		}
		return errors.New("transaction failed")
	})
}
func TestTransactionFailureDoesNotBroadcast(t *testing.T) {
	s, repo, events, id := testService(t)
	s.transactionRunner = failingTransaction{s.transactionRunner}
	_, err := s.PostMessage(context.Background(), id, "alice", "rolled back")
	if err == nil {
		t.Fatal("expected transaction failure")
	}
	if len(events.events) != 0 {
		t.Fatal("broadcast on rollback")
	}
	room, err := repo.FindByID(context.Background(), id)
	mustSucceed(t, err)
	if len(room.Messages) != 0 {
		t.Fatal("failed mutation persisted")
	}
}
func TestActiveCodeExcludesFinishedRooms(t *testing.T) {
	s, repo, _, id := testService(t)
	ctx := context.Background()
	room, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	exists, err := repo.ActiveCodeExists(ctx, room.RoomCode)
	mustSucceed(t, err)
	if !exists {
		t.Fatal("waiting code not reserved")
	}
	mustSucceed(t, s.StartGame(ctx, id, "admin"))
	exists, err = repo.ActiveCodeExists(ctx, room.RoomCode)
	mustSucceed(t, err)
	if !exists {
		t.Fatal("playing code not reserved")
	}
	mustSucceed(t, s.FinishGame(ctx, id, "admin"))
	exists, err = repo.ActiveCodeExists(ctx, room.RoomCode)
	mustSucceed(t, err)
	if exists {
		t.Fatal("finished code remains reserved")
	}
}

func TestStartPickExhaustionPersists(t *testing.T) {
	s, repo, events, id := testService(t)
	ctx := context.Background()
	mustSucceed(t, s.StartGame(ctx, id, "admin"))
	room, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	for number := 1; number <= 75; number++ {
		room.PickedBalls = append(room.PickedBalls, model.BallNumber(number))
	}
	mustSucceed(t, repo.Save(ctx, room))
	events.events = nil
	if err := s.StartPick(ctx, id, "admin"); !errors.Is(err, model.ErrNoDrawableBalls) {
		t.Fatalf("exhausted start: %v", err)
	}
	room, err = repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if room.PickState != model.RoomPickStateExhausted || len(events.events) != 0 {
		t.Fatal("exhaustion was rolled back or broadcast started")
	}
	if err := s.FinishPick(ctx, id, "alice"); !errors.Is(err, model.ErrRoomForbidden) {
		t.Fatalf("permission: %v", err)
	}
}
