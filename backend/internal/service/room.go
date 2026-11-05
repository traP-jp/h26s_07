package service

import (
	"context"
	"slices"
	"sync"

	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/repository"
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

func (s *RoomService) GetRoom(ctx context.Context, roomID model.RoomID) (*model.Room, error) {
	return s.roomRepository.FindByID(ctx, roomID)
}

func (s *RoomService) ListRooms(ctx context.Context, includeFinished ...bool) ([]model.RoomSummary, error) {
	rooms, err := s.roomRepository.List(ctx)
	if err != nil {
		return nil, err
	}
	if len(includeFinished) > 0 && includeFinished[0] {
		return rooms, nil
	}
	return slices.DeleteFunc(rooms, func(r model.RoomSummary) bool { return r.State == model.RoomStateFinished }), nil
}
