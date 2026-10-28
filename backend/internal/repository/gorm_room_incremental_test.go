package repository

import (
	"context"
	"errors"
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
