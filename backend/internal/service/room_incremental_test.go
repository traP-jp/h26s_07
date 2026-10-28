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
