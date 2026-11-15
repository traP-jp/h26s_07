package model

import (
	"cmp"
	"github.com/google/uuid"
	"slices"
	"strings"
	"time"
)

// ActivityLogEntry は初回リーチと各ビンゴ成立を履歴として表す。
// 判定に使った球の時刻は保存していないため、履歴には含めない。
type ActivityLogEntry struct {
	ID        RecordID
	Kind      string
	UserID    UserID
	CreatedAt time.Time
	BingoRank *int
	LineCount *int
}

func (room *Room) ActivityLog() []ActivityLogEntry {
	entries := make([]ActivityLogEntry, 0, len(room.ReachRecords)+len(room.BingoRecords))
	for _, r := range room.ReachRecords {
		entries = append(entries, ActivityLogEntry{ID: r.RecordID, Kind: "reach", UserID: r.UserID, CreatedAt: r.CreatedAt})
	}
	bingoRecords := slices.Clone(room.BingoRecords)
	// 同じ抽選で時刻が同じ記録の順位は、保存済みの順序で決める。
	slices.SortFunc(bingoRecords, func(a, b BingoRecord) int { return cmp.Compare(a.Order, b.Order) })
	ranks := map[UserID]int{}
	counts := map[UserID]int{}
	for _, r := range bingoRecords {
		if _, exists := ranks[r.UserID]; !exists {
			ranks[r.UserID] = len(ranks) + 1
		}
		counts[r.UserID]++
		rank, count := ranks[r.UserID], counts[r.UserID]
		entries = append(entries, ActivityLogEntry{ID: r.RecordID, Kind: "bingo", UserID: r.UserID, CreatedAt: r.CreatedAt, BingoRank: &rank, LineCount: &count})
	}
	slices.SortFunc(entries, func(a, b ActivityLogEntry) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(uuid.UUID(a.ID).String(), uuid.UUID(b.ID).String())
	})
	return entries
}
