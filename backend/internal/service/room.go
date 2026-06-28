package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
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

func randomDrawableNumber(numbers []model.BallNumber) (model.BallNumber, error) {
	if len(numbers) == 0 {
		return 0, model.ErrNoDrawableBalls
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(numbers))))
	if err != nil {
		return 0, err
	}
	return numbers[n.Int64()], nil
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

func convertPickedBallsToOpenAPI(pickedBalls []model.BallNumber) []openapi.PickedBall {
	result := make([]openapi.PickedBall, 0, len(pickedBalls))
	for _, ball := range pickedBalls {
		result = append(result, openapi.PickedBall(ball))
	}
	return result
}

func convertBingoSummariesToOpenAPI(bingoSummaries []model.BingoSummary) []openapi.BingoSummary {
	result := make([]openapi.BingoSummary, 0, len(bingoSummaries))
	for _, bingoSummary := range bingoSummaries {
		bingoOrders := make([]int, 0, len(bingoSummary.BingoOrders))
		for _, bingoOrder := range bingoSummary.BingoOrders {
			bingoOrders = append(bingoOrders, int(bingoOrder))
		}
		result = append(result, openapi.BingoSummary{
			BingoOrders: bingoOrders,
			User:        openapi.User{UserID: openapi.UserID(bingoSummary.UserID)},
		})
	}
	return result
}

func convertReachSummariesToOpenAPI(reachSummaries []model.ReachSummary) []openapi.ReachSummary {
	result := make([]openapi.ReachSummary, 0, len(reachSummaries))
	for _, reachSummary := range reachSummaries {
		result = append(result, openapi.ReachSummary{
			User: openapi.User{UserID: openapi.UserID(reachSummary.UserID)},
		})
	}
	return result
}

func convertBingoUpdatesToOpenAPI(updates []model.BingoUpdate) []openapi.BingoUpdate {
	result := make([]openapi.BingoUpdate, 0, len(updates))
	for _, update := range updates {
		result = append(result, openapi.BingoUpdate{
			User:           openapi.User{UserID: openapi.UserID(update.UserID)},
			NewBingoOrders: convertBingoOrdersToOpenAPI(update.NewBingoOrders),
			BingoOrders:    convertBingoOrdersToOpenAPI(update.BingoOrders),
		})
	}
	return result
}

func convertBingoOrdersToOpenAPI(orders []model.BingoOrder) []int {
	result := make([]int, 0, len(orders))
	for _, order := range orders {
		result = append(result, int(order))
	}
	return result
}

func convertReachUpdatesToOpenAPI(updates []model.ReachUpdate) []openapi.ReachUpdate {
	result := make([]openapi.ReachUpdate, 0, len(updates))
	for _, update := range updates {
		result = append(result, openapi.ReachUpdate{
			User: openapi.User{UserID: openapi.UserID(update.UserID)},
		})
	}
	return result
}

func convertCardToOpenAPI(room *model.Room, card model.Card) openapi.Card {
	cells := make([]openapi.CardCell, 0, len(card.Cells))
	for _, cell := range card.Cells {
		cells = append(cells, convertCardCellToOpenAPI(cell))
	}
	return openapi.Card{
		CardID:      uuid.UUID(card.CardID).String(),
		CardNumber:  string(card.CardNumber),
		OwnerUserID: openapi.UserID(card.OwnerUserID),
		Cells:       cells,
		BingoLines:  convertLineIndexesToOpenAPI(bingoLineIndexes(room, card.OwnerUserID)),
		ReachLines:  convertLineIndexesToOpenAPI(card.ReachLines(room.BingoRecords)),
	}
}

func convertCardCellToOpenAPI(cell model.CardCell) openapi.CardCell {
	var number *int
	if cell.Number != nil {
		value := int(*cell.Number)
		number = &value
	}

	displayText := "FREE"
	if number != nil {
		displayText = strconv.Itoa(*number)
	}

	return openapi.CardCell{
		Index:       int(cell.Index),
		Number:      number,
		DisplayText: displayText,
		CellState:   openapi.CardCellState(cell.CellState),
	}
}

func convertCardChangesToOpenAPI(changes model.CardChanges) openapi.CardChanges {
	openedCellIndices := make([]int, 0, len(changes.OpenedCellIndices))
	for _, index := range changes.OpenedCellIndices {
		openedCellIndices = append(openedCellIndices, int(index))
	}
	return openapi.CardChanges{
		OpenedCellIndices: openedCellIndices,
		NewReachLines:     convertLineIndexesToOpenAPI(changes.NewReachLines),
		NewBingoLines:     convertLineIndexesToOpenAPI(changes.NewBingoLines),
	}
}

func convertLineIndexesToOpenAPI(lines []model.LineIndex) []openapi.Line {
	result := make([]openapi.Line, 0, len(lines))
	for _, line := range lines {
		indices, ok := model.LineCells(line)
		if !ok {
			continue
		}
		openapiLine := make(openapi.Line, 0, len(indices))
		for _, index := range indices {
			openapiLine = append(openapiLine, int(index))
		}
		result = append(result, openapiLine)
	}
	return result
}

func bingoLineIndexes(room *model.Room, userID model.UserID) []model.LineIndex {
	lines := make([]model.LineIndex, 0)
	for _, record := range room.BingoRecords {
		if record.UserID == userID {
			lines = append(lines, record.Line)
		}
	}
	slices.Sort(lines)
	return lines
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

func (s *RoomService) FinishPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	var result model.PickFinishedResult
	if err := s.transactionRunner.WithinTransaction(ctx, func(ctx context.Context) error {
		room, err := s.roomRepository.FindByID(ctx, roomID)
		if err != nil {
			return err
		}

		picked, err := randomDrawableNumber(room.DrawableNumbers())
		if err != nil {
			return err
		}
		now := time.Now()
		result, err = room.FinishPick(actor, picked, func() (model.RecordID, error) {
			recordID, err := uuid.NewV7()
			if err != nil {
				return model.RecordID{}, err
			}
			return model.RecordID(recordID), nil
		}, now)
		if err != nil {
			return err
		}

		return s.roomRepository.Save(ctx, result.Room)
	}); err != nil {
		return err
	}

	if err := s.sendPickFinishedEvents(ctx, roomID, result); err != nil {
		return err
	}
	if result.AllPicked {
		if err := s.sendAllPickedEvents(ctx, roomID, result.Room.PickedBalls); err != nil {
			return err
		}
	}
	return nil
}

func (s *RoomService) sendPickFinishedEvents(ctx context.Context, roomID model.RoomID, result model.PickFinishedResult) error {
	room := result.Room
	bingoSummaries := convertBingoSummariesToOpenAPI(room.BingoSummaries())
	reachSummaries := convertReachSummariesToOpenAPI(room.ReachSummaries())
	newBingos := convertBingoUpdatesToOpenAPI(result.NewBingos)
	newReaches := convertReachUpdatesToOpenAPI(result.NewReaches)
	pickedBalls := convertPickedBallsToOpenAPI(room.PickedBalls)

	if err := s.events.SendRoom(ctx, roomID, openapi.DisplayPickFinishedEvent{
		Type: openapi.DisplayPickFinishedEventTypePickFinished,
		Body: openapi.DisplayPickFinishedBody{
			PickedBall:       openapi.PickedBall(result.PickedBall),
			PickState:        openapi.PickState(room.PickState),
			ParticipantCount: room.ParticipantCount(),
			BingoSummaries:   bingoSummaries,
			ReachSummaries:   reachSummaries,
			NewBingos:        newBingos,
			NewReaches:       newReaches,
			PickedBalls:      pickedBalls,
		},
	}); err != nil {
		return err
	}

	for _, update := range result.ParticipantUpdates {
		if err := s.events.SendParticipant(ctx, roomID, update.UserID, openapi.ParticipantPickFinishedEvent{
			Type: openapi.ParticipantPickFinishedEventTypePickFinished,
			Body: openapi.ParticipantPickFinishedBody{
				PickedBall:     openapi.PickedBall(result.PickedBall),
				PickState:      openapi.PickState(room.PickState),
				Card:           convertCardToOpenAPI(room, update.Card),
				CardChanges:    convertCardChangesToOpenAPI(update.CardChanges),
				PickedBalls:    pickedBalls,
				BingoSummaries: bingoSummaries,
				ReachSummaries: reachSummaries,
				NewBingos:      newBingos,
				NewReaches:     newReaches,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *RoomService) sendAllPickedEvents(ctx context.Context, roomID model.RoomID, pickedBalls []model.BallNumber) error {
	body := openapi.AllPickedBody{
		PickedBalls: convertPickedBallsToOpenAPI(pickedBalls),
	}
	if err := s.events.SendRoom(ctx, roomID, openapi.DisplayAllPickedEvent{
		Type: openapi.DisplayAllPickedEventTypeAllPicked,
		Body: body,
	}); err != nil {
		return err
	}
	return s.events.SendRoom(ctx, roomID, openapi.ParticipantAllPickedEvent{
		Type: openapi.ParticipantAllPickedEventTypeAllPicked,
		Body: body,
	})
}
