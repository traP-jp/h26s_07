package repository

import (
	"context"
	"errors"
	"time"

	"github.com/traP-jp/h26_07/backend/internal/model"
)

var (
	ErrRoomNotFound         = errors.New("room not found")
	ErrInvalidRoomAggregate = errors.New("invalid room aggregate")
)

type RoomRepository interface {
	LockRoomForUser(ctx context.Context, roomID model.RoomID, actor model.UserID) (*model.Room, error)
	LockPickRoom(ctx context.Context, roomID model.RoomID, actor model.UserID) (*model.Room, error)
	UpdatePickState(ctx context.Context, room *model.Room) error
	UpdateQRCode(ctx context.Context, room *model.Room) error
	SaveFinishedPick(ctx context.Context, before model.Room, room *model.Room) error
	AppendMessage(ctx context.Context, roomID model.RoomID, message model.Message) error
	JoinParticipant(ctx context.Context, roomID model.RoomID, user model.UserID, now time.Time) (joined bool, count int, err error)
	Delete(ctx context.Context, roomID model.RoomID) error
	Save(ctx context.Context, room *model.Room) error
	FindByID(ctx context.Context, roomID model.RoomID) (*model.Room, error)
	FindByIDForUpdate(ctx context.Context, roomID model.RoomID) (*model.Room, error)
	ActiveCodeExists(ctx context.Context, code model.RoomCode) (bool, error)
	List(ctx context.Context) ([]model.RoomSummary, error)
}
