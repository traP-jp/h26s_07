package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/utils"
)

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

func (s *RoomService) startPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	room, err := s.roomRepository.LockPickRoom(ctx, roomID, actor)
	if err != nil {
		return err
	}

	if err := room.StartPick(actor, time.Now()); err != nil {
		if errors.Is(err, model.ErrNoDrawableBalls) {
			if saveErr := s.roomRepository.UpdatePickState(ctx, room); saveErr != nil {
				return saveErr
			}
		}
		return err
	}

	if err := s.roomRepository.UpdatePickState(ctx, room); err != nil {
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
	room, err := s.roomRepository.LockRoomForUser(ctx, roomID, actor)
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

	if err := s.roomRepository.UpdatePickState(ctx, room); err != nil {
		return err
	}

	return s.queuePickCanceledEvents(ctx, roomID)
}

func (s *RoomService) finishPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	room, err := s.roomRepository.FindByIDForUpdate(ctx, roomID)
	if err != nil {
		return err
	}

	if !room.IsAdmin(actor) {
		return model.ErrRoomForbidden
	}
	before, result, oldIDs, err := finishPickState(room, actor)
	if err != nil {
		return err
	}
	if err := s.roomRepository.SaveFinishedPick(ctx, before, result.Room); err != nil {
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

func finishPickState(room *model.Room, actor model.UserID) (model.Room, model.PickFinishedResult, map[model.RecordID]bool, error) {
	if !room.CanFinishPick(actor) {
		return model.Room{}, model.PickFinishedResult{}, nil, model.ErrRoomPickNotFinishable
	}
	picked, err := randomDrawableNumber(room.DrawableNumbers())
	if err != nil {
		return model.Room{}, model.PickFinishedResult{}, nil, err
	}
	oldIDs := make(map[model.RecordID]bool)
	for _, entry := range room.ActivityLog() {
		oldIDs[entry.ID] = true
	}
	before := *room
	now := time.Now()
	result, err := room.FinishPick(actor, picked, func() (model.RecordID, error) {
		recordID, err := uuid.NewV7()
		if err != nil {
			return model.RecordID{}, err
		}
		return model.RecordID(recordID), nil
	}, now)
	if err != nil {
		return model.Room{}, model.PickFinishedResult{}, nil, err
	}
	return before, result, oldIDs, nil
}

type pickFinishedEventData struct {
	newActivityLog []openapi.ActivityLogEntry
	bingoSummaries []openapi.BingoSummary
	reachSummaries []openapi.ReachSummary
	newBingos      []openapi.BingoUpdate
	newReaches     []openapi.ReachUpdate
	pickedBalls    []openapi.PickedBall
}

func newPickFinishedEventData(result model.PickFinishedResult, oldIDs map[model.RecordID]bool) pickFinishedEventData {
	room := result.Room
	entries := slices.DeleteFunc(room.ActivityLog(), func(e model.ActivityLogEntry) bool { return oldIDs[e.ID] })
	return pickFinishedEventData{
		newActivityLog: utils.ConvertActivityLogToOpenAPI(entries),
		bingoSummaries: convertBingoSummariesToOpenAPI(room.BingoSummaries()),
		reachSummaries: convertReachSummariesToOpenAPI(room.ReachSummaries()),
		newBingos:      convertBingoUpdatesToOpenAPI(result.NewBingos),
		newReaches:     convertReachUpdatesToOpenAPI(result.NewReaches),
		pickedBalls:    convertPickedBallsToOpenAPI(room.PickedBalls),
	}
}

func (s *RoomService) sendPickFinishedEvents(ctx context.Context, roomID model.RoomID, result model.PickFinishedResult, oldIDs map[model.RecordID]bool) error {
	data := newPickFinishedEventData(result, oldIDs)
	if err := s.queueDisplayPickFinishedEvent(ctx, roomID, result, data); err != nil {
		return err
	}
	return s.queueParticipantPickFinishedEvents(ctx, roomID, result, data)
}

func (s *RoomService) queueDisplayPickFinishedEvent(ctx context.Context, roomID model.RoomID, result model.PickFinishedResult, data pickFinishedEventData) error {
	room := result.Room
	return s.queueRoomEvent(ctx, roomID, openapi.DisplayPickFinishedEvent{
		Type: openapi.DisplayPickFinishedEventTypePickFinished,
		Body: openapi.DisplayPickFinishedBody{
			NewActivityLog:   data.newActivityLog,
			PickedBall:       openapi.PickedBall(result.PickedBall),
			PickState:        openapi.PickState(room.PickState),
			ParticipantCount: room.ParticipantCount(),
			BingoSummaries:   data.bingoSummaries,
			ReachSummaries:   data.reachSummaries,
			NewBingos:        data.newBingos,
			NewReaches:       data.newReaches,
			PickedBalls:      data.pickedBalls,
		},
	})
}

func (s *RoomService) queueParticipantPickFinishedEvents(ctx context.Context, roomID model.RoomID, result model.PickFinishedResult, data pickFinishedEventData) error {
	room := result.Room
	for _, update := range result.ParticipantUpdates {
		if err := s.queueParticipantEvent(ctx, roomID, update.UserID, openapi.ParticipantPickFinishedEvent{
			Type: openapi.ParticipantPickFinishedEventTypePickFinished,
			Body: openapi.ParticipantPickFinishedBody{
				NewActivityLog: data.newActivityLog,
				PickedBall:     openapi.PickedBall(result.PickedBall),
				PickState:      openapi.PickState(room.PickState),
				Card:           convertCardToOpenAPI(room, update.Card),
				CardChanges:    convertCardChangesToOpenAPI(update.CardChanges),
				PickedBalls:    data.pickedBalls,
				BingoSummaries: data.bingoSummaries,
				ReachSummaries: data.reachSummaries,
				NewBingos:      data.newBingos,
				NewReaches:     data.newReaches,
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

func (s *RoomService) StartPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.startPick(ctx, roomID, actor) })
}

func (s *RoomService) CancelPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.cancelPick(ctx, roomID, actor) })
}

func (s *RoomService) FinishPick(ctx context.Context, roomID model.RoomID, actor model.UserID) error {
	return s.mutate(ctx, roomID, func(ctx context.Context) error { return s.finishPick(ctx, roomID, actor) })
}
