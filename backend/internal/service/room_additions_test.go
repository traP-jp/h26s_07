package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
	"github.com/traP-jp/h26_07/backend/internal/testutil"
	"reflect"
	"testing"
	"time"
)

func TestParticipantJoinedOnlyForNewParticipant(t *testing.T) {
	s, _, events, id := testService(t)
	ctx := context.Background()
	mustSucceed(t, s.PostParticipants(ctx, id, "alice"))
	if len(events.events) != 0 {
		t.Fatal("idempotent join broadcast")
	}
	mustSucceed(t, s.PostParticipants(ctx, id, "bob"))
	if len(events.events) != 2 {
		t.Fatal("expected both modes")
	}
	display, ok := events.events[0].(openapi.DisplayParticipantJoinedEvent)
	if !ok || display.Body.Participant.UserID != "bob" || display.Body.ParticipantCount != 2 || display.Body.JoinedAt.IsZero() {
		t.Fatalf("bad join event: %+v", events.events)
	}
	if _, ok := events.events[1].(openapi.ParticipantParticipantJoinedEvent); !ok {
		t.Fatal("missing participant event")
	}
}

func TestDeleteRoomPermissionsCascadeAndEvents(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewRoomDB(t)
	repo := repository.NewGormRoomRepository(db)
	events := &testEvents{}
	s := NewRoomService(repository.NewGormTransactionRunner(db), repo, events)
	room, err := s.CreateRoom(ctx, model.RoomSettings{Name: "test"}, "admin")
	mustSucceed(t, err)
	id := room.RoomID
	mustSucceed(t, s.PostParticipants(ctx, id, "alice"))
	mustSucceed(t, s.StartGame(ctx, id, "admin"))
	_, err = s.PostMessage(ctx, id, "alice", "hello")
	mustSucceed(t, err)
	mustSucceed(t, s.StartPick(ctx, id, "admin"))
	mustSucceed(t, s.FinishPick(ctx, id, "admin"))
	stored, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	stored.ReachRecords = []model.ReachRecord{{RecordID: model.RecordID(uuid.New()), UserID: "alice", CreatedAt: time.Now()}}
	stored.BingoRecords = []model.BingoRecord{{RecordID: model.RecordID(uuid.New()), UserID: "alice", Order: 1, CreatedAt: time.Now()}}
	mustSucceed(t, repo.Save(ctx, stored))
	events.events = nil
	if err := s.DeleteRoom(ctx, id, "alice"); !errors.Is(err, model.ErrRoomForbidden) {
		t.Fatal(err)
	}
	if len(events.events) != 0 {
		t.Fatal("forbidden delete broadcast")
	}
	events.check = func() {
		if _, err := repo.FindByID(ctx, id); !errors.Is(err, repository.ErrRoomNotFound) {
			t.Error("broadcast before deletion committed")
		}
	}
	mustSucceed(t, s.DeleteRoom(ctx, id, "admin"))
	if len(events.events) != 2 {
		t.Fatal("missing delete events")
	}
	if _, ok := events.events[0].(openapi.DisplayRoomDeletedEvent); !ok {
		t.Fatal("display delete event")
	}
	if _, ok := events.events[1].(openapi.ParticipantRoomDeletedEvent); !ok {
		t.Fatal("participant delete event")
	}
	for _, table := range []string{"rooms", "room_users", "room_admins", "room_participants", "room_cards", "room_card_cells", "room_picked_balls", "room_bingo_records", "room_reach_records", "room_messages"} {
		var count int64
		mustSucceed(t, db.Table(table).Count(&count).Error)
		if count != 0 {
			t.Errorf("%s retains %d rows", table, count)
		}
	}
	if err := s.DeleteRoom(ctx, id, "admin"); !errors.Is(err, repository.ErrRoomNotFound) {
		t.Fatal(err)
	}
	if len(events.events) != 2 {
		t.Fatal("not found broadcast")
	}
}

func TestActivityLogReloadAndListFiltering(t *testing.T) {
	s, repo, _, id := testService(t)
	ctx := context.Background()
	room, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	now := time.Now().UTC().Truncate(time.Microsecond)
	room.ReachRecords = []model.ReachRecord{{RecordID: model.RecordID(uuid.New()), UserID: "alice", CreatedAt: now}}
	// StartGame supplies a card required by persisted bingo records.
	mustSucceed(t, s.StartGame(ctx, id, "admin"))
	playing, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	playing.ReachRecords = room.ReachRecords
	playing.BingoRecords = []model.BingoRecord{{RecordID: model.RecordID(uuid.New()), UserID: "alice", Order: 1, CreatedAt: now.Add(time.Second)}}
	mustSucceed(t, repo.Save(ctx, playing))
	want := playing.ActivityLog()
	for range 2 {
		reloaded, err := repo.FindByID(ctx, id)
		mustSucceed(t, err)
		if !reflect.DeepEqual(want, reloaded.ActivityLog()) {
			t.Fatal("history changed after reload")
		}
	}
	mustSucceed(t, s.FinishGame(ctx, id, "admin"))
	rooms, err := s.ListRooms(ctx)
	mustSucceed(t, err)
	if len(rooms) != 0 {
		t.Fatal("finished room included by default")
	}
	rooms, err = s.ListRooms(ctx, true)
	mustSucceed(t, err)
	if len(rooms) != 1 || len(rooms[0].ActivityLogEntries) != 2 {
		t.Fatal("includeFinished omitted room or history")
	}
}

func TestPickFinishedNewActivityLogMatchesReload(t *testing.T) {
	s, repo, events, id := testService(t)
	ctx := context.Background()
	mustSucceed(t, s.StartGame(ctx, id, "admin"))
	seen := map[string]bool{}
	// Drawing all balls guarantees reach and bingo records regardless of card layout.
	for range 75 {
		events.events = nil
		mustSucceed(t, s.StartPick(ctx, id, "admin"))
		mustSucceed(t, s.FinishPick(ctx, id, "admin"))
		var display []openapi.ActivityLogEntry
		var participant []openapi.ActivityLogEntry
		for _, event := range events.events {
			switch e := event.(type) {
			case openapi.DisplayPickFinishedEvent:
				display = e.Body.NewActivityLog
			case openapi.ParticipantPickFinishedEvent:
				participant = e.Body.NewActivityLog
			}
		}
		if !reflect.DeepEqual(display, participant) {
			t.Fatal("modes disagree on activity delta")
		}
		for _, entry := range display {
			if seen[entry.ID] {
				t.Fatal("repeated activity")
			}
			seen[entry.ID] = true
		}
		room, err := repo.FindByID(ctx, id)
		mustSucceed(t, err)
		if len(seen) != len(room.ActivityLog()) {
			t.Fatal("activity delta missed records")
		}
	}
	room, err := repo.FindByID(ctx, id)
	mustSucceed(t, err)
	if len(room.ActivityLog()) < 2 {
		t.Fatal("expected reach and bingo history")
	}
}
