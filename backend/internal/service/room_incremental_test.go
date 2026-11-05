package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
	"github.com/traP-jp/h26_07/backend/internal/testutil"
	"github.com/traP-jp/h26_07/backend/internal/utils"
)

func TestIncrementalMessagesAndJoins(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewRoomDB(t)
	repo := repository.NewGormRoomRepository(db)
	events := &testEvents{}
	s := NewRoomService(repository.NewGormTransactionRunner(db), repo, events)
	room, err := s.CreateRoom(ctx, model.RoomSettings{Name: "incremental"}, "admin")
	mustSucceed(t, err)
	id := room.RoomID
	// Any aggregate rewrite would trip these guards, even with empty children.
	for _, table := range []string{"room_admins", "room_participants", "room_cards", "room_messages"} {
		mustSucceed(t, db.Exec("CREATE TRIGGER guard_"+table+" BEFORE DELETE ON "+table+" BEGIN SELECT RAISE(ABORT, 'aggregate rewrite'); END").Error)
	}
	for i, user := range []model.UserID{"alice", "bob", "admin"} {
		mustSucceed(t, s.PostParticipants(ctx, id, user))
		event := events.events[len(events.events)-1].(openapi.ParticipantParticipantJoinedEvent)
		if event.Body.ParticipantCount != i+1 {
			t.Fatalf("count: %+v", event.Body)
		}
	}
	before, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	n := len(events.events)
	mustSucceed(t, s.PostParticipants(ctx, id, "alice"))
	after, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if len(events.events) != n || !reflect.DeepEqual(before, after) {
		t.Fatal("rejoin changed room or emitted events")
	}
	events.events = nil
	var posted []model.Message
	events.check = func() {
		// A separate read during delivery observes the committed insert.
		messages, err := s.GetMessages(ctx, id, "alice")
		mustSucceed(t, err)
		if len(messages) == 0 {
			t.Fatal("event delivered before persistence")
		}
	}
	for _, user := range []model.UserID{"alice", "admin", "bob"} {
		message, err := s.PostMessage(ctx, id, user, string(user))
		mustSucceed(t, err)
		posted = append(posted, *message)
	}
	events.check = nil
	messages, err := s.GetMessages(ctx, id, "alice")
	mustSucceed(t, err)
	stored, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if len(messages) != len(posted) || len(stored.Messages) != len(posted) {
		t.Fatal("message count")
	}
	for i, want := range posted {
		for _, got := range []model.Message{messages[i], stored.Messages[i]} {
			if got.MessageID != want.MessageID || got.Author != want.Author || got.Content != want.Content || !got.CreatedAt.Equal(want.CreatedAt) {
				t.Fatalf("message/order: got %+v want %+v", got, want)
			}
		}
	}
	if len(events.events) != 6 {
		t.Fatalf("events: %d", len(events.events))
	}
	for i, message := range posted {
		p := events.events[2*i].(openapi.ParticipantMessageCreatedEvent)
		d := events.events[2*i+1].(openapi.DisplayMessageCreatedEvent)
		if p.Body.Message.Content != message.Content || p.Body.Message.MessageID != uuid.UUID(message.MessageID).String() || !reflect.DeepEqual(p.Body, d.Body) {
			t.Fatal("message event payload")
		}
	}
	for _, tc := range []struct {
		user    model.UserID
		content string
		want    error
	}{
		{"outsider", "hello", model.ErrRoomMessageNotAllowed},
		{"alice", "", model.ErrMessageInvalid},
		{"alice", strings.Repeat("あ", 501), model.ErrMessageInvalid},
	} {
		_, err := s.PostMessage(ctx, id, tc.user, tc.content)
		if !errors.Is(err, tc.want) {
			t.Fatalf("got %v want %v", err, tc.want)
		}
	}
	mustSucceed(t, db.Table("rooms").Where("room_id = ?", id.String()).Update("state", "finished").Error)
	_, err = s.PostMessage(ctx, id, "alice", "hello")
	if !errors.Is(err, model.ErrRoomMessageNotPostable) {
		t.Fatal(err)
	}
	mustSucceed(t, s.PostParticipants(ctx, id, "alice"))
	if err := s.PostParticipants(ctx, id, "new"); !errors.Is(err, model.ErrRoomNotJoinable) {
		t.Fatal(err)
	}
	_, err = s.PostMessage(ctx, model.RoomID(uuid.New()), "alice", "hello")
	if !errors.Is(err, repository.ErrRoomNotFound) {
		t.Fatal(err)
	}
	stored, err = repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if len(stored.Messages) != 3 || len(events.events) != 6 {
		t.Fatal("rejected mutation persisted or emitted event")
	}
}

func TestIncrementalMessageRollbackDoesNotBroadcast(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewRoomDB(t)
	repo := repository.NewGormRoomRepository(db)
	events := &testEvents{}
	s := NewRoomService(repository.NewGormTransactionRunner(db), repo, events)
	room, err := s.CreateRoom(ctx, model.RoomSettings{Name: "rollback"}, "admin")
	mustSucceed(t, err)
	mustSucceed(t, db.Exec("CREATE TRIGGER fail_update BEFORE UPDATE ON rooms BEGIN SELECT RAISE(ABORT, 'fail'); END").Error)
	if _, err := s.PostMessage(ctx, room.RoomID, "admin", "hello"); err == nil {
		t.Fatal("expected update failure")
	}
	stored, err := repo.FindByID(ctx, room.RoomID)
	mustSucceed(t, err)
	if len(stored.Messages) != 0 || len(events.events) != 0 {
		t.Fatal("failed transaction persisted or broadcast")
	}
}

func TestIncrementalPickEventsMatchPersistedActivity(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewRoomDB(t)
	repo := repository.NewGormRoomRepository(db)
	events := &testEvents{}
	svc := NewRoomService(repository.NewGormTransactionRunner(db), repo, events)
	room, err := svc.CreateRoom(ctx, model.RoomSettings{Name: "pick events"}, "admin")
	mustSucceed(t, err)
	id := room.RoomID
	mustSucceed(t, svc.PostParticipants(ctx, id, "alice"))
	mustSucceed(t, svc.StartGame(ctx, id, "admin"))
	for _, table := range []string{"room_users", "room_admins", "room_participants", "room_cards", "room_card_cells", "room_picked_balls", "room_bingo_records", "room_reach_records", "room_messages"} {
		mustSucceed(t, db.Exec("CREATE TRIGGER guard_"+table+" BEFORE DELETE ON "+table+" BEGIN SELECT RAISE(ABORT, 'aggregate rewrite'); END").Error)
	}
	mustSucceed(t, svc.HideQRCode(ctx, id, "admin"))
	stored, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if stored.QrCodeVisible {
		t.Fatal("QR code still visible")
	}
	mustSucceed(t, svc.ShowQRCode(ctx, id, "admin"))
	mustSucceed(t, svc.StartPick(ctx, id, "admin"))
	before, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	mustSucceed(t, svc.StartPick(ctx, id, "admin"))
	stored, err = repo.FindByID(ctx, id)
	mustSucceed(t, err)
	before.UpdatedAt = stored.UpdatedAt // Save also refreshed this on repeated starts.
	if !reflect.DeepEqual(before, stored) {
		t.Fatal("idempotent start changed aggregate")
	}
	mustSucceed(t, svc.CancelPick(ctx, id, "admin"))
	events.events = nil
	mustSucceed(t, svc.CancelPick(ctx, id, "admin"))
	if len(events.events) != 0 {
		t.Fatal("idle cancellation emitted events")
	}
	equalEntry := func(a, b openapi.ActivityLogEntry) bool {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			return false
		}
		a.CreatedAt = b.CreatedAt
		return reflect.DeepEqual(a, b)
	}
	seen := map[string]bool{}
	for i := 0; i < 75; i++ {
		mustSucceed(t, svc.StartPick(ctx, id, "admin"))
		events.events = nil
		mustSucceed(t, svc.FinishPick(ctx, id, "admin"))
		stored, err = repo.FindByID(ctx, id)
		mustSucceed(t, err)
		if len(stored.PickedBalls) != i+1 {
			t.Fatal("pick not persisted")
		}
		display := events.events[0].(openapi.DisplayPickFinishedEvent)
		participant := events.events[1].(openapi.ParticipantPickFinishedEvent)
		if !reflect.DeepEqual(display.Body.NewActivityLog, participant.Body.NewActivityLog) {
			t.Fatal("activity payload mismatch")
		}
		for _, entry := range stored.ActivityLog() {
			key := uuid.UUID(entry.ID).String()
			if !seen[key] {
				expected := utils.ConvertActivityLogToOpenAPI([]model.ActivityLogEntry{entry})
				found := false
				for _, emitted := range display.Body.NewActivityLog {
					if equalEntry(emitted, expected[0]) {
						found = true
						break
					}
				}
				if !found {
					t.Fatal("new persisted activity missing from event")
				}
				seen[key] = true
			}
		}
		want := utils.ConvertActivityLogToOpenAPI(stored.ActivityLog())
		for _, entry := range display.Body.NewActivityLog {
			found := false
			for _, persisted := range want {
				if equalEntry(entry, persisted) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("event activity not persisted: %+v", entry)
			}
		}
		if display.Body.ParticipantCount != 1 || !reflect.DeepEqual(display.Body.PickedBalls, convertPickedBallsToOpenAPI(stored.PickedBalls)) || !reflect.DeepEqual(participant.Body.Card, convertCardToOpenAPI(stored, stored.Cards[0])) {
			t.Fatal("finish payload differs from aggregate")
		}
	}
	if len(seen) != 13 || len(stored.ActivityLog()) != 13 || stored.PickState != model.RoomPickStateExhausted {
		t.Fatal("missing reach/bingo activity or exhaustion")
	}
	if err := svc.StartPick(ctx, id, "admin"); !errors.Is(err, model.ErrRoomPickNotStartable) {
		t.Fatalf("exhausted start: %v", err)
	}
}
