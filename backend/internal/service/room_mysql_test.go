package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"os"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/repository"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysqlTestEvents struct{}

func (mysqlTestEvents) SendRoom(context.Context, model.RoomID, any) error { return nil }
func (mysqlTestEvents) SendParticipant(context.Context, model.RoomID, model.UserID, any) error {
	return nil
}

// Run against the migrated dev database (schema: cmd/migrate):
// TEST_MYSQL_DSN='<user>:<password>@tcp(127.0.0.1:3306)/h26_07?parseTime=true' go test ./internal/service -run TestMySQL -count=1
// Only the uniquely generated room and its children are removed; other data is untouched.
func TestMySQLRoomMutationsAcrossInstances(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set; run cmd/migrate first")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	mustSucceed(t, err)
	sqlDB, err := db.DB()
	mustSucceed(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	// Use the existing migrated schema, never replace it with the SQLite test schema.
	for _, row := range []any{&dbmodel.Room{}, &dbmodel.RoomCard{}, &dbmodel.RoomMessage{}} {
		if !db.Migrator().HasTable(row) {
			t.Fatal("missing schema: run go run ./cmd/migrate")
		}
	}
	repo := repository.NewGormRoomRepository(db)
	services := []*RoomService{
		NewRoomService(repository.NewGormTransactionRunner(db), repo, mysqlTestEvents{}),
		NewRoomService(repository.NewGormTransactionRunner(db), repository.NewGormRoomRepository(db), mysqlTestEvents{}),
	}
	ctx := context.Background()
	room, err := services[0].CreateRoom(ctx, model.RoomSettings{Name: "mysql integration"}, "admin")
	mustSucceed(t, err)
	id := room.RoomID
	t.Cleanup(func() {
		if err := repo.Delete(ctx, id); err != nil && !errors.Is(err, repository.ErrRoomNotFound) {
			t.Error(err)
		}
	})
	mustSucceed(t, services[0].PostParticipants(ctx, id, "alice"))
	mustSucceed(t, services[0].StartGame(ctx, id, "admin"))
	before, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	start := make(chan struct{})
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	// Four writers share the DB but only their own service's in-process lock.
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			<-start
			for i := 0; i < 20; i++ {
				if _, err := services[worker%2].PostMessage(ctx, id, "alice", fmt.Sprintf("%d:%d", worker, i)); err != nil {
					errs <- err
					return
				}
			}
		}(worker)
	}
	// Alternate which instance performs picks while both receive concurrent chats.
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 75; i++ {
			if err := services[i%2].StartPick(ctx, id, "admin"); err != nil {
				errs <- err
				return
			}
			if err := services[(i+1)%2].FinishPick(ctx, id, "admin"); err != nil {
				errs <- err
				return
			}
		}
	}()
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	stored, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if len(stored.Messages) != 80 || len(stored.PickedBalls) != 75 || len(stored.Cards) != 1 || len(stored.Participants) != 1 {
		t.Fatalf("lost data: messages=%d picks=%d cards=%d participants=%d", len(stored.Messages), len(stored.PickedBalls), len(stored.Cards), len(stored.Participants))
	}
	messages := map[string]bool{}
	for _, m := range stored.Messages {
		if messages[m.Content] {
			t.Fatal("duplicate message")
		}
		messages[m.Content] = true
	}
	balls := map[model.BallNumber]bool{}
	for _, b := range stored.PickedBalls {
		if balls[b] {
			t.Fatal("duplicate pick")
		}
		balls[b] = true
	}
	if stored.Cards[0].CardID != before.Cards[0].CardID {
		t.Fatal("card identity changed")
	}
	for i, cell := range stored.Cards[0].Cells {
		if !reflect.DeepEqual(cell.Number, before.Cards[0].Cells[i].Number) {
			t.Fatal("card number changed")
		}
		wantOpen := cell.Number == nil || slices.Contains(stored.PickedBalls, *cell.Number)
		if (cell.CellState == model.CardCellStateOpen || cell.CellState == model.CardCellStateBingo) != wantOpen {
			t.Fatalf("cell %d inconsistent", i)
		}
	}
	if len(stored.BingoRecords) == 0 || len(stored.ReachRecords) == 0 {
		t.Fatal("expected populated activity records")
	}
	cardID := stored.Cards[0].CardID
	mustSucceed(t, services[1].DeleteRoom(ctx, id, "admin"))
	if _, err := repo.FindByID(ctx, id); !errors.Is(err, repository.ErrRoomNotFound) {
		t.Fatalf("deleted room: %v", err)
	}
	for _, table := range []string{"rooms", "room_users", "room_admins", "room_participants", "room_cards", "room_picked_balls", "room_bingo_records", "room_reach_records", "room_messages"} {
		var count int64
		mustSucceed(t, db.Table(table).Where("room_id = ?", id.String()).Count(&count).Error)
		if count != 0 {
			t.Errorf("%s retains %d rows", table, count)
		}
	}
	var cells int64
	mustSucceed(t, db.Model(&dbmodel.RoomCardCell{}).Where("card_id = ?", uuid.UUID(cardID).String()).Count(&cells).Error)
	if cells != 0 {
		t.Fatalf("deleted card retains %d cells", cells)
	}
}

// Each writer has its own service lock, exercising database serialization.
func TestMySQLConcurrentIncrementalChatsAndJoins(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set; run cmd/migrate first")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	mustSucceed(t, err)
	sqlDB, err := db.DB()
	mustSucceed(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	repo := repository.NewGormRoomRepository(db)
	events := &concurrentMessageEvents{ids: map[string]int{}, counts: map[int]int{}}
	newService := func() *RoomService {
		return NewRoomService(repository.NewGormTransactionRunner(db), repository.NewGormRoomRepository(db), events)
	}
	s := newService()
	room, err := s.CreateRoom(ctx, model.RoomSettings{Name: "concurrent incremental"}, "admin")
	mustSucceed(t, err)
	t.Cleanup(func() {
		if err := repo.Delete(context.Background(), room.RoomID); err != nil {
			t.Error(err)
		}
	})
	run := func(n int, fn func(int) error) {
		t.Helper()
		var wg sync.WaitGroup
		start := make(chan struct{})
		errs := make(chan error, n)
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int) { defer wg.Done(); <-start; errs <- fn(i) }(i)
		}
		close(start)
		wg.Wait()
		close(errs)
		for err := range errs {
			mustSucceed(t, err)
		}
	}
	run(20, func(i int) error {
		return newService().PostParticipants(ctx, room.RoomID, model.UserID(fmt.Sprintf("member-%d", i%10)))
	})
	run(50, func(i int) error {
		_, err := newService().PostMessage(ctx, room.RoomID, model.UserID(fmt.Sprintf("member-%d", i%10)), fmt.Sprintf("chat-%d", i))
		return err
	})
	stored, err := repo.FindByID(ctx, room.RoomID)
	mustSucceed(t, err)
	if len(stored.Participants) != 10 || len(stored.Messages) != 50 {
		t.Fatalf("participants=%d messages=%d", len(stored.Participants), len(stored.Messages))
	}
	events.mu.Lock()
	defer events.mu.Unlock()
	for _, message := range stored.Messages {
		if events.ids[uuid.UUID(message.MessageID).String()] != 2 {
			t.Fatal("missing MessageCreated events")
		}
	}
	if len(events.ids) != 50 {
		t.Fatalf("message events: %d", len(events.ids))
	}
	for count := 1; count <= 10; count++ {
		if events.counts[count] != 2 {
			t.Fatalf("join count %d emitted %d times", count, events.counts[count])
		}
	}
}

type concurrentMessageEvents struct {
	mu     sync.Mutex
	ids    map[string]int
	counts map[int]int
}

func (e *concurrentMessageEvents) SendRoom(_ context.Context, _ model.RoomID, event any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	switch event := event.(type) {
	case openapi.ParticipantMessageCreatedEvent:
		e.ids[event.Body.Message.MessageID]++
	case openapi.DisplayMessageCreatedEvent:
		e.ids[event.Body.Message.MessageID]++
	case openapi.ParticipantParticipantJoinedEvent:
		e.counts[event.Body.ParticipantCount]++
	case openapi.DisplayParticipantJoinedEvent:
		e.counts[event.Body.ParticipantCount]++
	}
	return nil
}
func (e *concurrentMessageEvents) SendParticipant(context.Context, model.RoomID, model.UserID, any) error {
	return nil
}

func TestMySQLConcurrentPickTransitionsAcrossInstances(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set; run cmd/migrate first")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	mustSucceed(t, err)
	sqlDB, err := db.DB()
	mustSucceed(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	repo := repository.NewGormRoomRepository(db)
	services := []*RoomService{
		NewRoomService(repository.NewGormTransactionRunner(db), repo, mysqlTestEvents{}),
		NewRoomService(repository.NewGormTransactionRunner(db), repository.NewGormRoomRepository(db), mysqlTestEvents{}),
	}
	room, err := services[0].CreateRoom(ctx, model.RoomSettings{Name: "racing picks"}, "admin")
	mustSucceed(t, err)
	t.Cleanup(func() { mustSucceed(t, repo.Delete(context.Background(), room.RoomID)) })
	mustSucceed(t, services[0].PostParticipants(ctx, room.RoomID, "alice"))
	mustSucceed(t, services[0].StartGame(ctx, room.RoomID, "admin"))
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for worker, svc := range services {
		wg.Add(1)
		go func(worker int, svc *RoomService) {
			defer wg.Done()
			<-start
			for i := 0; i < 150; i++ {
				var err error
				switch (i + worker) % 3 {
				case 0:
					err = svc.StartPick(ctx, room.RoomID, "admin")
				case 1:
					err = svc.FinishPick(ctx, room.RoomID, "admin")
				case 2:
					err = svc.CancelPick(ctx, room.RoomID, "admin")
				}
				if err != nil && !errors.Is(err, model.ErrRoomPickNotStartable) && !errors.Is(err, model.ErrRoomPickNotFinishable) && !errors.Is(err, model.ErrRoomPickNotCancelable) && !errors.Is(err, model.ErrNoDrawableBalls) {
					errs <- err
					return
				}
			}
		}(worker, svc)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		mustSucceed(t, err)
	}
	stored, err := repo.FindByID(ctx, room.RoomID)
	mustSucceed(t, err)
	// Drain the remaining balls to prove that racing cancellations leave a usable state.
	for len(stored.PickedBalls) < 75 {
		mustSucceed(t, services[0].StartPick(ctx, room.RoomID, "admin"))
		mustSucceed(t, services[1].FinishPick(ctx, room.RoomID, "admin"))
		stored, err = repo.FindByID(ctx, room.RoomID)
		mustSucceed(t, err)
	}
	if stored.PickState != model.RoomPickStateExhausted {
		t.Fatalf("final pick state: %s", stored.PickState)
	}
	seen := map[model.BallNumber]bool{}
	for _, ball := range stored.PickedBalls {
		if seen[ball] {
			t.Fatalf("duplicate ball: %d", ball)
		}
		seen[ball] = true
	}
	if len(seen) != 75 || len(stored.BingoRecords) != 12 || len(stored.ReachRecords) != 1 {
		t.Fatalf("inconsistent final aggregate: %+v", stored)
	}
	for _, cell := range stored.Cards[0].Cells {
		if cell.CellState != model.CardCellStateBingo {
			t.Fatalf("unopened final cell: %+v", cell)
		}
	}
}
