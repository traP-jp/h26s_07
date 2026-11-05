package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/traP-jp/h26_07/backend/internal/model"
)

// Entries are removed when the last holder or waiter releases the lock.
type roomLocks struct {
	mu      sync.Mutex
	entries map[model.RoomID]*roomLock
}
type roomLock struct {
	mu   sync.Mutex
	refs int
}

func (l *roomLocks) lock(id model.RoomID) func() {
	l.mu.Lock()
	if l.entries == nil {
		l.entries = make(map[model.RoomID]*roomLock)
	}
	entry := l.entries[id]
	if entry == nil {
		entry = &roomLock{}
		l.entries[id] = entry
	}
	entry.refs++
	l.mu.Unlock()
	entry.mu.Lock()
	return func() {
		entry.mu.Unlock()
		l.mu.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(l.entries, id)
		}
		l.mu.Unlock()
	}
}

// LockRoom coordinates snapshot loading and subscription with mutations and broadcasts.
// The caller must release the lock immediately after registering the subscription.
func (s *RoomService) LockRoom(id model.RoomID) func() { return s.locks.lock(id) }

type pendingEvent func(context.Context) error
type eventQueueKey struct{}

// mutate applies a room write in lock, transaction, row-lock, mutation, commit order.
// The process lock prevents concurrent requests here from losing updates, while the
// repository's SELECT FOR UPDATE prevents lost updates across service instances.
// Events stay queued until commit so a rollback cannot be observed by clients.
func (s *RoomService) mutate(ctx context.Context, id model.RoomID, fn func(context.Context) error) error {
	unlock := s.LockRoom(id)
	defer unlock()
	var events []pendingEvent
	ctx = context.WithValue(ctx, eventQueueKey{}, &events)
	// StartPick records exhaustion while still reporting the domain conflict.
	var domainErr error
	if err := s.transactionRunner.WithinTransaction(ctx, func(ctx context.Context) error {
		err := fn(ctx)
		if errors.Is(err, model.ErrNoDrawableBalls) {
			domainErr = err
			return nil
		}
		return err
	}); err != nil {
		return err
	}
	for _, send := range events {
		if err := send(ctx); err != nil {
			slog.ErrorContext(ctx, "room event broadcast failed", "room_id", id, "error", err)
		}
	}
	return domainErr
}
func (s *RoomService) queueRoomEvent(ctx context.Context, id model.RoomID, event any) error {
	queue := ctx.Value(eventQueueKey{}).(*[]pendingEvent)
	*queue = append(*queue, func(ctx context.Context) error { return s.events.SendRoom(ctx, id, event) })
	return nil
}
func (s *RoomService) queueParticipantEvent(ctx context.Context, id model.RoomID, user model.UserID, event any) error {
	queue := ctx.Value(eventQueueKey{}).(*[]pendingEvent)
	*queue = append(*queue, func(ctx context.Context) error { return s.events.SendParticipant(ctx, id, user, event) })
	return nil
}
