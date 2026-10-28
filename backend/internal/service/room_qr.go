package service

import (
	"context"
	"time"

	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
)

func (s *RoomService) showQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.LockRoomForUser(ctx, roomID, user)
	if err != nil {
		if err == repository.ErrRoomNotFound {
			return model.ErrRoomNotFound
		}
		return err
	}
	if err := room.ShowQRCode(user, time.Now()); err != nil {
		return err
	}
	err = s.roomRepository.UpdateQRCode(ctx, room)
	if err != nil {
		return err
	}
	err = s.queueRoomEvent(ctx, roomID, openapi.DisplayShowQRCodeEvent{
		Type: openapi.ShowQRCode,
		Body: openapi.ShowQRCodeBody{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *RoomService) hideQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.LockRoomForUser(ctx, roomID, user)
	if err != nil {
		if err == repository.ErrRoomNotFound {
			return model.ErrRoomNotFound
		}
		return err
	}
	if err := room.HideQRCode(user, time.Now()); err != nil {
		return err
	}
	err = s.roomRepository.UpdateQRCode(ctx, room)
	if err != nil {
		return err
	}
	err = s.queueRoomEvent(ctx, roomID, openapi.DisplayHideQRCodeEvent{
		Type: openapi.DisplayHideQRCodeEventTypeHideQRCode,
		Body: openapi.HideQRCodeBody{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *RoomService) ShowQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.showQRCode(ctx, roomID, user) })
}

func (s *RoomService) HideQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.hideQRCode(ctx, roomID, user) })
}
