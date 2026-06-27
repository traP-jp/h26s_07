package service

import (
	"context"

	"github.com/traP-jp/h26_07/backend/internal/model"
)

type RoomEventSender interface {
	SendRoom(ctx context.Context, roomID model.RoomID, event any) error
	SendParticipant(ctx context.Context, roomID model.RoomID, userID model.UserID, event any) error
}
