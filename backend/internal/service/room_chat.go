package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func (s *RoomService) postMessage(ctx context.Context, roomID model.RoomID, user model.UserID, content string) (*model.Message, error) {
	messageID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	createdMessage := model.Message{MessageID: model.MessageID(messageID), Author: user, Content: content, CreatedAt: time.Now()}
	if err := s.roomRepository.AppendMessage(ctx, roomID, createdMessage); err != nil {
		return nil, err
	}
	err = s.queueRoomEvent(ctx, roomID, openapi.ParticipantMessageCreatedEvent{
		Type: openapi.ParticipantMessageCreatedEventTypeMessageCreated,
		Body: openapi.MessageCreatedBody{Message: openapi.Message{
			Author:    openapi.User{UserID: openapi.UserID(createdMessage.Author)},
			Content:   createdMessage.Content,
			CreatedAt: createdMessage.CreatedAt,
			MessageID: openapi.UUID(uuid.UUID(createdMessage.MessageID).String()),
		}},
	})
	if err != nil {
		return nil, err
	}
	err = s.queueRoomEvent(ctx, roomID, openapi.DisplayMessageCreatedEvent{
		Type: openapi.DisplayMessageCreatedEventTypeMessageCreated,
		Body: openapi.MessageCreatedBody{Message: openapi.Message{
			Author:    openapi.User{UserID: openapi.UserID(createdMessage.Author)},
			Content:   createdMessage.Content,
			CreatedAt: createdMessage.CreatedAt,
			MessageID: openapi.UUID(uuid.UUID(createdMessage.MessageID).String()),
		}},
	})
	if err != nil {
		return nil, err
	}
	return &createdMessage, nil
}

func (s *RoomService) GetMessages(ctx context.Context, roomID model.RoomID, user model.UserID) ([]model.Message, error) {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !room.CanViewChat(user) {
		return nil, model.ErrRoomForbidden
	}
	return append([]model.Message(nil), room.Messages...), nil
}

func (s *RoomService) PostMessage(ctx context.Context, roomID model.RoomID, user model.UserID, content string) (*model.Message, error) {
	var result *model.Message
	err := s.mutate(ctx, roomID, func(ctx context.Context) error {
		var err error
		result, err = s.postMessage(ctx, roomID, user, content)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, err
}
