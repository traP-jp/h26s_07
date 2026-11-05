package service

import (
	"context"
	"time"

	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func (s *RoomService) postParticipants(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	now := time.Now()
	joined, count, err := s.roomRepository.JoinParticipant(ctx, roomID, user, now)
	if err != nil {
		return err
	}
	if !joined {
		return nil
	}
	body := openapi.ParticipantJoinedBody{Participant: openapi.User{UserID: string(user)}, ParticipantCount: count, JoinedAt: now}
	if err := s.queueRoomEvent(ctx, roomID, openapi.DisplayParticipantJoinedEvent{Type: openapi.DisplayParticipantJoinedEventTypeParticipantJoined, Body: body}); err != nil {
		return err
	}
	return s.queueRoomEvent(ctx, roomID, openapi.ParticipantParticipantJoinedEvent{Type: openapi.ParticipantParticipantJoinedEventTypeParticipantJoined, Body: body})
}

func (s *RoomService) PostParticipants(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.postParticipants(ctx, roomID, user) })
}
