package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/testutil"
)

func TestAppendMessageUsesTransactionAndReadOrdering(t *testing.T) {
	db := testutil.NewRoomDB(t)
	repo := NewGormRoomRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	room := model.NewRoom(model.RoomID(uuid.New()), "123456", model.RoomSettings{Name: "append", Admins: []model.UserID{"admin"}}, now)
	if err := repo.Save(ctx, room); err != nil {
		t.Fatal(err)
	}
	runner := NewGormTransactionRunner(db)
	for _, offset := range []time.Duration{2 * time.Second, time.Second} {
		message := model.Message{MessageID: model.MessageID(uuid.New()), Author: "admin", Content: "hello", CreatedAt: now.Add(offset)}
		if err := runner.WithinTransaction(ctx, func(ctx context.Context) error { return repo.AppendMessage(ctx, room.RoomID, message) }); err != nil {
			t.Fatal(err)
		}
	}
	rollback := errors.New("rollback")
	err := runner.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := repo.AppendMessage(ctx, room.RoomID, model.Message{MessageID: model.MessageID(uuid.New()), Author: "admin", Content: "discard", CreatedAt: now.Add(3 * time.Second)}); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	stored, err := repo.FindByID(ctx, room.RoomID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Messages) != 2 || !stored.Messages[0].CreatedAt.Equal(now.Add(time.Second)) || !stored.Messages[1].CreatedAt.Equal(now.Add(2*time.Second)) || !stored.UpdatedAt.Equal(now.Add(time.Second)) {
		t.Fatalf("unexpected persistence: %+v", stored)
	}
}

func TestIncrementalPicksMatchFullSave(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewRoomDB(t)
	repo := NewGormRoomRepository(db)
	reference := NewGormRoomRepository(testutil.NewRoomDB(t))
	runner := NewGormTransactionRunner(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	room := model.NewRoom(model.RoomID(uuid.New()), "123456", model.RoomSettings{Name: "picks", Admins: []model.UserID{"admin"}}, now)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(room.Join("alice", now))
	_, err := room.StartGame("admin", []model.Card{testCard(uuid.NewString(), "alice")}, now)
	must(err)
	_, err = room.PostMessage("admin", "preserve", now, model.MessageID(uuid.New()))
	must(err)
	must(repo.Save(ctx, room))
	room, err = repo.FindByID(ctx, room.RoomID)
	must(err)
	must(reference.Save(ctx, room))
	for _, table := range []string{"room_users", "room_admins", "room_participants", "room_cards", "room_card_cells", "room_picked_balls", "room_bingo_records", "room_reach_records", "room_messages"} {
		must(db.Exec("CREATE TRIGGER guard_" + table + " BEFORE DELETE ON " + table + " BEGIN SELECT RAISE(ABORT, 'aggregate rewrite'); END").Error)
	}
	compare := func() {
		t.Helper()
		got, err := repo.FindByID(ctx, room.RoomID)
		must(err)
		want, err := reference.FindByID(ctx, room.RoomID)
		must(err)
		// Both paths refresh updated_at at persistence time; the writes occur separately.
		want.UpdatedAt = got.UpdatedAt
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(got.ActivityLog(), want.ActivityLog()) {
			t.Fatalf("incremental aggregate differs from full Save\ngot: %+v\nwant: %+v", got, want)
		}
	}
	for pick := 1; pick <= 75; pick++ {
		now = now.Add(time.Second)
		must(runner.WithinTransaction(ctx, func(ctx context.Context) error {
			projection, err := repo.LockPickRoom(ctx, room.RoomID, "admin")
			if err != nil {
				return err
			}
			if err := projection.StartPick("admin", now); err != nil {
				return err
			}
			if err := room.StartPick("admin", now); err != nil {
				return err
			}
			return repo.UpdatePickState(ctx, projection)
		}))
		must(reference.Save(ctx, room))
		compare()
		if pick == 1 {
			must(runner.WithinTransaction(ctx, func(ctx context.Context) error {
				projection, err := repo.LockRoomForUser(ctx, room.RoomID, "admin")
				if err != nil {
					return err
				}
				if err := projection.CancelPick("admin", now); err != nil {
					return err
				}
				return repo.UpdatePickState(ctx, projection)
			}))
			must(room.CancelPick("admin", now))
			must(reference.Save(ctx, room))
			compare()
			must(room.StartPick("admin", now))
			must(repo.UpdatePickState(ctx, room))
		}
		must(runner.WithinTransaction(ctx, func(ctx context.Context) error {
			loaded, err := repo.FindByIDForUpdate(ctx, room.RoomID)
			if err != nil {
				return err
			}
			before := *loaded
			_, err = loaded.FinishPick("admin", model.BallNumber(pick), func() (model.RecordID, error) { return model.RecordID(uuid.New()), nil }, now)
			if err != nil {
				return err
			}
			room = loaded
			return repo.SaveFinishedPick(ctx, before, loaded)
		}))
		must(reference.Save(ctx, room))
		compare()
	}
	if len(room.ReachRecords) != 1 || len(room.BingoRecords) != 12 || room.PickState != model.RoomPickStateExhausted {
		t.Fatalf("unexpected final records/state: %+v", room)
	}
	for _, visible := range []bool{false, true} {
		must(runner.WithinTransaction(ctx, func(ctx context.Context) error {
			projection, err := repo.LockRoomForUser(ctx, room.RoomID, "admin")
			if err != nil {
				return err
			}
			if visible {
				err = projection.ShowQRCode("admin", now)
			} else {
				err = projection.HideQRCode("admin", now)
			}
			if err != nil {
				return err
			}
			return repo.UpdateQRCode(ctx, projection)
		}))
		if visible {
			must(room.ShowQRCode("admin", now))
		} else {
			must(room.HideQRCode("admin", now))
		}
		must(reference.Save(ctx, room))
		compare()
	}
}
