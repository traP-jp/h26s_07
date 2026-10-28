package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/utils"
)

func random6Digits() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *RoomService) CreateRoom(ctx context.Context, settings model.RoomSettings, creator model.UserID) (*model.Room, error) {
	s.createMu.Lock()
	defer s.createMu.Unlock()
	if !settings.HasAdmin(creator) {
		settings.Admins = append(settings.Admins, creator)
	}
	if !settings.Valid() {
		return nil, model.ErrRoomSettingsInvalid
	}
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	generatedCode, err := random6Digits()
	if err != nil {
		return nil, err
	}
	for {
		exists, err := s.roomRepository.ActiveCodeExists(ctx, model.RoomCode(generatedCode))
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		generatedCode, err = random6Digits()
		if err != nil {
			return nil, err
		}
	}
	room := model.NewRoom(
		model.RoomID(uuid),
		model.RoomCode(generatedCode),
		settings,
		time.Now(),
	)
	if err := s.transactionRunner.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.roomRepository.Save(ctx, room)
	}); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) putSettings(ctx context.Context, roomID model.RoomID, user model.UserID, input RoomSettingsUpdate) (model.RoomSettings, error) {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
	if err != nil {
		return model.RoomSettings{}, err
	}
	admins := append([]model.UserID(nil), room.Settings.Admins...)
	if input.Admins != nil {
		admins = append([]model.UserID(nil), *input.Admins...)
	}
	settings := model.RoomSettings{
		Name:        input.Name,
		Description: input.Description,
		Admins:      admins,
	}
	if err := room.UpdateSettings(user, settings, time.Now()); err != nil {
		return model.RoomSettings{}, err
	}
	if err := s.roomRepository.Save(ctx, room); err != nil {
		return model.RoomSettings{}, err
	}
	err = s.queueRoomEvent(ctx, roomID, openapi.DisplayGameSettingsUpdatedEvent{
		Type: openapi.DisplayGameSettingsUpdatedEventTypeGameSettingsUpdated,
		Body: openapi.DisplayGameSettingsUpdatedBody{
			Settings: convertRoomSettingsToOpenAPI(room.Settings),
		},
	})
	if err != nil {
		return model.RoomSettings{}, err
	}
	err = s.queueRoomEvent(ctx, roomID, openapi.ParticipantGameSettingsUpdatedEvent{
		Type: openapi.ParticipantGameSettingsUpdatedEventTypeGameSettingsUpdated,
		Body: openapi.ParticipantGameSettingsUpdatedBody{
			Settings: convertRoomSettingsToOpenAPI(room.Settings),
		},
	})
	if err != nil {
		return model.RoomSettings{}, err
	}
	return room.Settings, nil
}

func (s *RoomService) startGame(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
	if err != nil {
		return err
	}
	if !room.IsAdmin(user) {
		return model.ErrRoomForbidden
	}
	if room.State == model.RoomStatePlaying {
		return nil
	}

	cards := make([]model.Card, 0, len(room.Participants))
	for _, participant := range room.Participants {
		card, err := model.MakeRandomCard(participant.UserID)
		if err != nil {
			return err
		}
		cards = append(cards, card)
	}
	result, err := room.StartGame(user, cards, time.Now())
	if err != nil {
		return err
	}
	if err := s.roomRepository.Save(ctx, room); err != nil {
		return err
	}
	if err := s.queueRoomEvent(ctx, roomID, openapi.DisplayGameStartedEvent{
		Type: openapi.DisplayGameStartedEventTypeGameStarted,
		Body: openapi.DisplayGameStartedBody{ParticipantCount: len(room.Participants)},
	}); err != nil {
		return err
	}
	for _, update := range result.ParticipantCards {
		if err := s.queueParticipantEvent(ctx, roomID, update.UserID, openapi.ParticipantGameStartedEvent{
			Type: openapi.ParticipantGameStartedEventTypeGameStarted,
			Body: openapi.ParticipantGameStartedBody{
				Card: utils.ConvertCardToOpenAPI(room, update.Card),
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *RoomService) finishGame(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
	if err != nil {
		return err
	}
	if !room.IsAdmin(user) {
		return model.ErrRoomForbidden
	}
	if room.State == model.RoomStateFinished {
		return nil
	}

	result, err := room.FinishGame(user, time.Now())
	if err != nil {
		return err
	}
	if err := s.roomRepository.Save(ctx, room); err != nil {
		return err
	}
	if result.PickCanceled {
		if err := s.queuePickCanceledEvents(ctx, roomID); err != nil {
			return err
		}
	}
	return s.queueGameFinishedEvents(ctx, roomID, room, result)
}

// queuePickCanceledEvents tells display and participants to stop the drumroll.
// CancelPick and FinishGame (when it interrupts a draw) both use it so the two paths send the same events in the same order.
func (s *RoomService) queuePickCanceledEvents(ctx context.Context, roomID model.RoomID) error {
	if err := s.queueRoomEvent(ctx, roomID, openapi.DisplayPickCanceledEvent{
		Type: openapi.DisplayPickCanceledEventTypePickCanceled,
		Body: openapi.PickCanceledBody{},
	}); err != nil {
		return err
	}
	return s.queueRoomEvent(ctx, roomID, openapi.ParticipantPickCanceledEvent{
		Type: openapi.ParticipantPickCanceledEventTypePickCanceled,
		Body: openapi.PickCanceledBody{},
	})
}

func (s *RoomService) queueGameFinishedEvents(ctx context.Context, roomID model.RoomID, room *model.Room, result model.GameFinishedResult) error {
	reachSummaries := convertReachSummariesToOpenAPI(room.ReachSummaries())
	bingoSummaries := convertBingoSummariesToOpenAPI(room.BingoSummaries())
	for _, update := range result.ParticipantUpdates {
		if err := s.queueParticipantEvent(ctx, roomID, update.UserID, openapi.ParticipantGameFinishedEvent{
			Type: openapi.ParticipantGameFinishedEventTypeGameFinished,
			Body: openapi.ParticipantGameFinishedBody{
				BingoSummaries: bingoSummaries,
				Card:           utils.ConvertCardToOpenAPI(room, update.Card),
				PickState:      openapi.ParticipantGameFinishedBodyPickState(room.PickState),
				ReachSummaries: reachSummaries,
				State:          openapi.ParticipantGameFinishedBodyState(room.State),
			},
		}); err != nil {
			return err
		}
	}
	return s.queueRoomEvent(ctx, roomID, openapi.DisplayGameFinishedEvent{
		Type: openapi.DisplayGameFinishedEventTypeGameFinished,
		Body: openapi.DisplayGameFinishedBody{
			ParticipantCount: len(room.Participants),
			BingoSummaries:   bingoSummaries,
			PickState:        openapi.DisplayGameFinishedBodyPickState(room.PickState),
			ReachSummaries:   reachSummaries,
			State:            openapi.DisplayGameFinishedBodyState(room.State),
		},
	})
}

func (s *RoomService) PutSettings(ctx context.Context, roomID model.RoomID, user model.UserID, input RoomSettingsUpdate) (model.RoomSettings, error) {
	var result model.RoomSettings
	err := s.mutate(ctx, roomID, func(ctx context.Context) error {
		var err error
		result, err = s.putSettings(ctx, roomID, user, input)
		return err
	})
	return result, err
}

func (s *RoomService) StartGame(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.startGame(ctx, roomID, user) })
}

func (s *RoomService) FinishGame(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.finishGame(ctx, roomID, user) })
}

func (s *RoomService) DeleteRoom(ctx context.Context, id model.RoomID, user model.UserID) error {
	return s.mutate(ctx, id, func(ctx context.Context) error {
		room, err := s.roomRepository.FindByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if !room.IsAdmin(user) {
			return model.ErrRoomForbidden
		}
		if err := s.roomRepository.Delete(ctx, id); err != nil {
			return err
		}
		if err := s.queueRoomEvent(ctx, id, openapi.DisplayRoomDeletedEvent{Type: openapi.DisplayRoomDeletedEventTypeRoomDeleted, Body: openapi.RoomDeletedBody{}}); err != nil {
			return err
		}
		return s.queueRoomEvent(ctx, id, openapi.ParticipantRoomDeletedEvent{Type: openapi.ParticipantRoomDeletedEventTypeRoomDeleted, Body: openapi.RoomDeletedBody{}})
	})
}
