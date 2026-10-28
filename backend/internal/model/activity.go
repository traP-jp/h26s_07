package model

import (
	"github.com/google/uuid"
	"slices"
	"strings"
	"time"
)

// ActivityLogEntry retains every first-reach record and every bingo-line record.
// Triggering balls cannot be recovered: the aggregate stores no pick timestamps.
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
	for _, r := range room.BingoRecords {
		entries = append(entries, ActivityLogEntry{ID: r.RecordID, Kind: "bingo", UserID: r.UserID, CreatedAt: r.CreatedAt})
	}
	slices.SortFunc(entries, func(a, b ActivityLogEntry) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return strings.Compare(uuid.UUID(a.ID).String(), uuid.UUID(b.ID).String())
	})
	ranks := map[UserID]int{}
	counts := map[UserID]int{}
	for i := range entries {
		e := &entries[i]
		if e.Kind != "bingo" {
			continue
		}
		if ranks[e.UserID] == 0 {
			ranks[e.UserID] = len(ranks) + 1
		}
		counts[e.UserID]++
		rank, count := ranks[e.UserID], counts[e.UserID]
		e.BingoRank = &rank
		e.LineCount = &count
	}
	return entries
}
