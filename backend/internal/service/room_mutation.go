package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/traP-jp/h26_07/backend/internal/model"
)

// 最後の保持者または待機者がロックを解放すると、項目を削除する。
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

// LockRoom はスナップショット読み込み、購読登録、状態変更、イベント配信の順序を調整する。
// 呼び出し側は購読登録の直後にロックを解放すること。
func (s *RoomService) LockRoom(id model.RoomID) func() { return s.locks.lock(id) }

type pendingEvent func(context.Context) error
type eventQueueKey struct{}

func (s *RoomService) mutate(ctx context.Context, id model.RoomID, fn func(context.Context) error) error {
	unlock := s.LockRoom(id)
	defer unlock()
	var events []pendingEvent
	ctx = context.WithValue(ctx, eventQueueKey{}, &events)
	// 抽選開始不可の競合を返しつつ、球切れ状態を記録する。
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
