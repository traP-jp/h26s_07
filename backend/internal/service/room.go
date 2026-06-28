package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
	"github.com/traP-jp/h26_07/backend/internal/utils"
)

type RoomService struct {
	transactionRunner repository.TransactionRunner
	roomRepository    repository.RoomRepository
	events            RoomEventSender
}

type RoomSettingsUpdate struct {
	Name        string
	Description string
	Admins      *[]model.UserID
}

func NewRoomService(transactionRunner repository.TransactionRunner, roomRepository repository.RoomRepository, events RoomEventSender) *RoomService {
	return &RoomService{
		transactionRunner: transactionRunner,
		roomRepository:    roomRepository,
		events:            events,
	}
}
func random6Digits() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *RoomService) CreateRoom(ctx context.Context, settings model.RoomSettings, creator model.UserID) (*model.Room, error) {
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

func (s *RoomService) GetRoom(ctx context.Context, roomID model.RoomID) (*model.Room, error) {
	return s.roomRepository.FindByID(ctx, roomID)
}

func (s *RoomService) ListRooms(ctx context.Context) ([]model.RoomSummary, error) {
	return s.roomRepository.List(ctx)
}

func (s *RoomService) PostParticipants(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if err := room.Join(user, time.Now()); err != nil {
		return err
	}
	return s.roomRepository.Save(ctx, room)
}

func (s *RoomService) PostMessage(ctx context.Context, roomID model.RoomID, user model.UserID, content string) (*model.Message, error) {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	messageID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	createdMessage, err := room.PostMessage(user, content, time.Now(), model.MessageID(messageID))
	if err != nil {
		return nil, err
	}
	err = s.roomRepository.Save(ctx, room)
	if err != nil {
		return nil, err
	}
	err = s.events.SendRoom(ctx, roomID, openapi.ParticipantMessageCreatedEvent{
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
	err = s.events.SendRoom(ctx, roomID, openapi.DisplayMessageCreatedEvent{
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

func convertUserIDsToOpenAPI(userIDs []model.UserID) []openapi.User {
	var result []openapi.User
	for _, userID := range userIDs {
		result = append(result, openapi.User{UserID: openapi.UserID(userID)})
	}
	return result
}

func convertRoomSettingsToOpenAPI(settings model.RoomSettings) openapi.GameSettings {
	return openapi.GameSettings{
		Name:        settings.Name,
		Description: settings.Description,
		Admins:      convertUserIDsToOpenAPI(settings.Admins),
	}
}

func (s *RoomService) PutSettings(ctx context.Context, roomID model.RoomID, user model.UserID, input RoomSettingsUpdate) (model.RoomSettings, error) {
	room, err := s.roomRepository.FindByID(ctx, roomID)
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
	err = s.events.SendRoom(ctx, roomID, openapi.DisplayGameSettingsUpdatedEvent{
		Type: openapi.DisplayGameSettingsUpdatedEventTypeGameSettingsUpdated,
		Body: openapi.DisplayGameSettingsUpdatedBody{
			Settings: convertRoomSettingsToOpenAPI(room.Settings),
		},
	})
	if err != nil {
		return model.RoomSettings{}, err
	}
	err = s.events.SendRoom(ctx, roomID, openapi.ParticipantGameSettingsUpdatedEvent{
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

func (s *RoomService) StartGame(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		return err
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
	if err := s.events.SendRoom(ctx, roomID, openapi.DisplayGameStartedEvent{
		Type: openapi.DisplayGameStartedEventTypeGameStarted,
		Body: openapi.DisplayGameStartedBody{ParticipantCount: len(room.Participants)},
	}); err != nil {
		return err
	}
	for _, update := range result.ParticipantCards {
		if err := s.events.SendParticipant(ctx, roomID, update.UserID, openapi.ParticipantGameStartedEvent{
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

func (s *RoomService) FinishGame(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	result, err := room.FinishGame(user, time.Now())
	if err != nil {
		return err
	}
	room.State = model.RoomStateFinished
	s.roomRepository.Save(ctx, room)
	reachSummaries := make([]openapi.ReachSummary, 0, len(room.ReachSummaries()))
	for _, reachSumamry := range room.ReachSummaries() {
		reachSummaries = append(reachSummaries, openapi.ReachSummary{User: openapi.User{UserID: openapi.UserID(reachSumamry.UserID)}})
	}
	bingoSummaries := make([]openapi.BingoSummary, 0, len(room.BingoSummaries()))
	for _, bingoSummary := range room.BingoSummaries() {
		bingoOrders := make([]int, 0, len(bingoSummary.BingoOrders))
		for _, bingoOrder := range bingoSummary.BingoOrders {
			bingoOrders = append(bingoOrders, int(bingoOrder))
		}
		bingoSummaries = append(bingoSummaries, openapi.BingoSummary{BingoOrders: bingoOrders, User: openapi.User{UserID: openapi.UserID(bingoSummary.UserID)}})
	}
	for _, update := range result.ParticipantUpdates {
		err = s.events.SendRoom(ctx, roomID, openapi.ParticipantGameFinishedEvent{
			Type: openapi.ParticipantGameFinishedEventTypeGameFinished,
			Body: openapi.ParticipantGameFinishedBody{
				BingoSummaries: bingoSummaries,
				Card:           utils.ConvertCardToOpenAPI(room, update.Card),
				PickState:      openapi.ParticipantGameFinishedBodyPickState(room.PickState),
				ReachSummaries: reachSummaries,
				State:          openapi.ParticipantGameFinishedBodyState(room.State),
			},
		})
		if err != nil {
			return err
		}
	}
	err = s.events.SendRoom(ctx, roomID, openapi.DisplayGameFinishedEvent{
		Type: openapi.DisplayGameFinishedEventTypeGameFinished,
		Body: openapi.DisplayGameFinishedBody{
			ParticipantCount: len(room.Participants),
			BingoSummaries:   bingoSummaries,
			PickState:        openapi.DisplayGameFinishedBodyPickState(room.PickState),
			ReachSummaries:   reachSummaries,
			State:            openapi.DisplayGameFinishedBodyState(room.State),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *RoomService) ShowQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		if err == repository.ErrRoomNotFound {
			return model.ErrRoomNotFound
		}
		return err
	}
	if err := room.ShowQRCode(user, time.Now()); err != nil {
		return err
	}
	err = s.roomRepository.Save(ctx, room)
	if err != nil {
		return err
	}
	err = s.events.SendRoom(ctx, roomID, openapi.DisplayShowQRCodeEvent{
		Type: openapi.ShowQRCode,
		Body: openapi.ShowQRCodeBody{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *RoomService) HideQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		if err == repository.ErrRoomNotFound {
			return model.ErrRoomNotFound
		}
		return err
	}
	if err := room.HideQRCode(user, time.Now()); err != nil {
		return err
	}
	err = s.roomRepository.Save(ctx, room)
	if err != nil {
		return err
	}
	err = s.events.SendRoom(ctx, roomID, openapi.DisplayHideQRCodeEvent{
		Type: openapi.DisplayHideQRCodeEventTypeHideQRCode,
		Body: openapi.HideQRCodeBody{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *RoomService) StartPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		return err
	}

	if err := room.StartPick(actor, time.Now()); err != nil {
		if errors.Is(err, model.ErrNoDrawableBalls) {
			if saveErr := s.roomRepository.Save(ctx, room); saveErr != nil {
				return saveErr
			}
		}
		return err
	}

	if err := s.roomRepository.Save(ctx, room); err != nil {
		return err
	}

	if err := s.events.SendRoom(ctx, roomID, openapi.DisplayPickStartedEvent{
		Type: openapi.DisplayPickStartedEventTypePickStarted,
		Body: openapi.PickStartedBody{},
	}); err != nil {
		return err
	}

	if err := s.events.SendRoom(ctx, roomID, openapi.ParticipantPickStartedEvent{
		Type: openapi.ParticipantPickStartedEventTypePickStarted,
		Body: openapi.PickStartedBody{},
	}); err != nil {
		return err
	}

	return nil
}

func (s *RoomService) CancelPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	room, err := s.roomRepository.FindByID(ctx, roomID)
	if err != nil {
		return err
	}

	if err := room.CancelPick(actor, time.Now()); err != nil {
		return err
	}

	if err := s.roomRepository.Save(ctx, room); err != nil {
		return err
	}

	if err := s.events.SendRoom(ctx, roomID, openapi.DisplayPickCanceledEvent{
		Type: openapi.DisplayPickCanceledEventTypePickCanceled,
		Body: openapi.PickCanceledBody{},
	}); err != nil {
		return err
	}

	if err := s.events.SendRoom(ctx, roomID, openapi.ParticipantPickCanceledEvent{
		Type: openapi.ParticipantPickCanceledEventTypePickCanceled,
		Body: openapi.PickCanceledBody{},
	}); err != nil {
		return err
	}

	return nil
}
