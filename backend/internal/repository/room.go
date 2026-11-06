package repository

import (
	"context"
	"errors"

	"github.com/traP-jp/h26_07/backend/internal/model"
)

var (
	ErrRoomNotFound         = errors.New("room not found")
	ErrInvalidRoomAggregate = errors.New("invalid room aggregate")
)

type RoomRepository interface {
	Save(ctx context.Context, room *model.Room) error
	FindByID(ctx context.Context, roomID model.RoomID) (*model.Room, error)
	FindByIDForUpdate(ctx context.Context, roomID model.RoomID) (*model.Room, error)
	ActiveCodeExists(ctx context.Context, code model.RoomCode) (bool, error)
	List(ctx context.Context) ([]model.RoomSummary, error)
}
