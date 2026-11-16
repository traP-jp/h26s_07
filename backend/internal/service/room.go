package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"sync"
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
	locks             roomLocks
	createMu          sync.Mutex
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

func (s *RoomService) GetRoom(ctx context.Context, roomID model.RoomID) (*model.Room, error) {
	return s.roomRepository.FindByID(ctx, roomID)
}

func (s *RoomService) ListRooms(ctx context.Context) ([]model.RoomSummary, error) {
	return s.roomRepository.List(ctx)
}

func (s *RoomService) postParticipants(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
	if err != nil {
		return err
	}
	alreadyJoined := room.IsParticipant(user)
	now := time.Now()
	if err := room.Join(user, now); err != nil {
		return err
	}
	if alreadyJoined {
		return nil
	}
	if err := s.roomRepository.Save(ctx, room); err != nil {
		return err
	}
	// 参加一覧を使う表示画面にだけ参加者追加を通知する。
	body := openapi.ParticipantJoinedBody{Participant: openapi.User{UserID: string(user)}}
	return s.queueRoomEvent(ctx, roomID, openapi.DisplayParticipantJoinedEvent{Type: openapi.ParticipantJoined, Body: body})
}

func (s *RoomService) postMessage(ctx context.Context, roomID model.RoomID, user model.UserID, content string) (*model.Message, error) {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
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
	err = s.queueRoomEvent(ctx, roomID, openapi.ParticipantMessageCreatedEvent{
		Type: openapi.ParticipantMessageCreatedEventTypeMessageCreated,
		Body: openapi.MessageCreatedBody{Message: openapi.Message{
			Author:    openapi.User{UserID: openapi.UserID(createdMessage.Author)},
			Content:   createdMessage.Content,
			CreatedAt: createdMessage.CreatedAt,
			MessageID: uuid.UUID(createdMessage.MessageID),
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
			MessageID: uuid.UUID(createdMessage.MessageID),
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

func convertCardToOpenAPI(room *model.Room, card model.Card) openapi.Card {
	cells := make([]openapi.CardCell, 0, len(card.Cells))
	for _, cell := range card.Cells {
		cells = append(cells, convertCardCellToOpenAPI(cell))
	}
	return openapi.Card{
		CardID:      uuid.UUID(card.CardID),
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
		Body: openapi.EmptyObject{},
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
		s.queueRoomEvent(ctx, roomID, openapi.DisplayPickCanceledEvent{Type: openapi.DisplayPickCanceledEventTypePickCanceled, Body: openapi.PickCanceledBody{}})
		s.queueRoomEvent(ctx, roomID, openapi.ParticipantPickCanceledEvent{Type: openapi.ParticipantPickCanceledEventTypePickCanceled, Body: openapi.PickCanceledBody{}})
	}
	for _, update := range result.ParticipantUpdates {
		if err := s.queueParticipantEvent(ctx, roomID, update.UserID, openapi.ParticipantGameFinishedEvent{
			Type: openapi.ParticipantGameFinishedEventTypeGameFinished,
			Body: openapi.ParticipantGameFinishedBody{
				Card:      utils.ConvertCardToOpenAPI(room, update.Card),
				PickState: openapi.ParticipantGameFinishedBodyPickState(room.PickState),
				State:     openapi.ParticipantGameFinishedBodyState(room.State),
			},
		}); err != nil {
			return err
		}
	}
	err = s.queueRoomEvent(ctx, roomID, openapi.DisplayGameFinishedEvent{
		Type: openapi.DisplayGameFinishedEventTypeGameFinished,
		Body: openapi.DisplayGameFinishedBody{
			PickState: openapi.DisplayGameFinishedBodyPickState(room.PickState),
			State:     openapi.DisplayGameFinishedBodyState(room.State),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *RoomService) showQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
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
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
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
	err = s.queueRoomEvent(ctx, roomID, openapi.DisplayHideQRCodeEvent{
		Type: openapi.DisplayHideQRCodeEventTypeHideQRCode,
		Body: openapi.HideQRCodeBody{},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *RoomService) startPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
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

	if err := s.queueRoomEvent(ctx, roomID, openapi.DisplayPickStartedEvent{
		Type: openapi.DisplayPickStartedEventTypePickStarted,
		Body: openapi.PickStartedBody{},
	}); err != nil {
		return err
	}

	if err := s.queueRoomEvent(ctx, roomID, openapi.ParticipantPickStartedEvent{
		Type: openapi.ParticipantPickStartedEventTypePickStarted,
		Body: openapi.PickStartedBody{},
	}); err != nil {
		return err
	}

	return nil
}

func (s *RoomService) cancelPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
	if err != nil {
		return err
	}
	if !room.IsAdmin(actor) {
		return model.ErrRoomForbidden
	}
	if room.PickState == model.RoomPickStateIdle {
		return nil
	}

	if err := room.CancelPick(actor, time.Now()); err != nil {
		return err
	}

	if err := s.roomRepository.Save(ctx, room); err != nil {
		return err
	}

	if err := s.queueRoomEvent(ctx, roomID, openapi.DisplayPickCanceledEvent{
		Type: openapi.DisplayPickCanceledEventTypePickCanceled,
		Body: openapi.PickCanceledBody{},
	}); err != nil {
		return err
	}

	if err := s.queueRoomEvent(ctx, roomID, openapi.ParticipantPickCanceledEvent{
		Type: openapi.ParticipantPickCanceledEventTypePickCanceled,
		Body: openapi.PickCanceledBody{},
	}); err != nil {
		return err
	}

	return nil
}

func (s *RoomService) finishPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
	if err != nil {
		return err
	}

	if !room.IsAdmin(actor) {
		return model.ErrRoomForbidden
	}
	if !room.CanFinishPick(actor) {
		return model.ErrRoomPickNotFinishable
	}

	picked, err := randomDrawableNumber(room.DrawableNumbers())
	if err != nil {
		return err
	}
	oldIDs := map[model.RecordID]bool{}
	for _, record := range room.BingoRecords {
		oldIDs[record.RecordID] = true
	}
	for _, record := range room.ReachRecords {
		oldIDs[record.RecordID] = true
	}
	now := time.Now()
	result, err := room.FinishPick(actor, picked, func() (model.RecordID, error) {
		recordID, err := uuid.NewV7()
		if err != nil {
			return model.RecordID{}, err
		}
		return model.RecordID(recordID), nil
	}, now)
	if err != nil {
		return err
	}

	if err := s.roomRepository.Save(ctx, result.Room); err != nil {
		return err
	}

	if err := s.sendPickFinishedEvents(ctx, roomID, result, oldIDs); err != nil {
		return err
	}
	if result.AllPicked {
		if err := s.sendAllPickedEvents(ctx, roomID, result.Room.PickedBalls); err != nil {
			return err
		}
	}
	return nil
}

func (s *RoomService) sendPickFinishedEvents(ctx context.Context, roomID model.RoomID, result model.PickFinishedResult, oldIDs map[model.RecordID]bool) error {
	room := result.Room
	entries := slices.DeleteFunc(room.ActivityLog(), func(entry model.ActivityLogEntry) bool { return oldIDs[entry.ID] })
	newActivityLog := utils.ConvertActivityLogToOpenAPI(entries)
	pickedBalls := convertPickedBallsToOpenAPI(room.PickedBalls)

	if err := s.queueRoomEvent(ctx, roomID, openapi.DisplayPickFinishedEvent{
		Type: openapi.DisplayPickFinishedEventTypePickFinished,
		Body: openapi.DisplayPickFinishedBody{
			NewActivityLog: newActivityLog,
			PickedBall:     openapi.PickedBall(result.PickedBall),
			PickState:      openapi.PickState(room.PickState),
			PickedBalls:    pickedBalls,
		},
	}); err != nil {
		return err
	}

	for _, update := range result.ParticipantUpdates {
		if err := s.queueParticipantEvent(ctx, roomID, update.UserID, openapi.ParticipantPickFinishedEvent{
			Type: openapi.ParticipantPickFinishedEventTypePickFinished,
			Body: openapi.ParticipantPickFinishedBody{
				NewActivityLog: newActivityLog,
				PickedBall:     openapi.PickedBall(result.PickedBall),
				PickState:      openapi.PickState(room.PickState),
				Card:           convertCardToOpenAPI(room, update.Card),
				CardChanges:    convertCardChangesToOpenAPI(update.CardChanges),
				PickedBalls:    pickedBalls,
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
	if err := s.queueRoomEvent(ctx, roomID, openapi.DisplayAllPickedEvent{
		Type: openapi.DisplayAllPickedEventTypeAllPicked,
		Body: body,
	}); err != nil {
		return err
	}
	return s.queueRoomEvent(ctx, roomID, openapi.ParticipantAllPickedEvent{
		Type: openapi.ParticipantAllPickedEventTypeAllPicked,
		Body: body,
	})
}

func (s *RoomService) PostParticipants(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.postParticipants(ctx, roomID, user) })
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

func (s *RoomService) StartPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.startPick(ctx, roomID, actor) })
}

func (s *RoomService) CancelPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.cancelPick(ctx, roomID, actor) })
}

func (s *RoomService) FinishPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.finishPick(ctx, roomID, actor) })
}

func (s *RoomService) ShowQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.showQRCode(ctx, roomID, user) })
}

func (s *RoomService) HideQRCode(ctx context.Context, roomID model.RoomID, user model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.hideQRCode(ctx, roomID, user) })
}
