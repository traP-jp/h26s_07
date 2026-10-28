package model

import (
	"cmp"
	"slices"
	"time"
)

type BingoSummary struct {
	UserID      UserID
	BingoOrders []BingoOrder
	CreatedAt   time.Time
}

type ReachSummary struct {
	UserID    UserID
	CreatedAt time.Time
}

type RoomSummary struct {
	ActivityLogEntries []ActivityLogEntry
	RoomID             RoomID
	RoomCode           RoomCode
	State              RoomState
	PickState          RoomPickState
	QrCodeVisible      bool
	Settings           RoomSettings
	Participants       []Participant
	BingoSummaries     []BingoSummary
	ReachSummaries     []ReachSummary
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (room *Room) CardByUserID(userID UserID) (Card, bool) {
	for _, card := range room.Cards {
		if card.OwnerUserID == userID {
			return card, true
		}
	}
	return Card{}, false
}

func (room *Room) BingoSummaries() []BingoSummary {
	records := slices.Clone(room.BingoRecords)
	slices.SortFunc(records, func(a, b BingoRecord) int {
		if n := cmp.Compare(a.Order, b.Order); n != 0 {
			return n
		}
		return cmp.Compare(a.UserID, b.UserID)
	})

	summaries := make([]BingoSummary, 0)
	indexByUserID := make(map[UserID]int)
	for _, record := range records {
		index, ok := indexByUserID[record.UserID]
		if !ok {
			index = len(summaries)
			indexByUserID[record.UserID] = index
			summaries = append(summaries, BingoSummary{
				UserID:      record.UserID,
				BingoOrders: []BingoOrder{},
				CreatedAt:   record.CreatedAt,
			})
		}
		summaries[index].BingoOrders = append(summaries[index].BingoOrders, record.Order)
	}
	return summaries
}

func (room *Room) ReachSummaries() []ReachSummary {
	return ReachSummariesFromRecords(room.ReachRecords, room.BingoRecords)
}

func ReachSummariesFromRecords(reachRecords []ReachRecord, bingoRecords []BingoRecord) []ReachSummary {
	records := slices.Clone(reachRecords)
	slices.SortFunc(records, func(a, b ReachRecord) int {
		if n := a.CreatedAt.Compare(b.CreatedAt); n != 0 {
			return n
		}
		return cmp.Compare(a.UserID, b.UserID)
	})

	bingoUserIDs := make(map[UserID]struct{})
	for _, record := range bingoRecords {
		bingoUserIDs[record.UserID] = struct{}{}
	}

	summaries := make([]ReachSummary, 0)
	addedUserIDs := make(map[UserID]struct{})
	for _, record := range records {
		if _, ok := bingoUserIDs[record.UserID]; ok {
			continue
		}
		if _, ok := addedUserIDs[record.UserID]; ok {
			continue
		}
		addedUserIDs[record.UserID] = struct{}{}
		summaries = append(summaries, ReachSummary{UserID: record.UserID, CreatedAt: record.CreatedAt})
	}
	return summaries
}

func (room *Room) DrawableNumbers() []BallNumber {
	return DrawableNumbers(room.PickedBalls)
}

func (room *Room) HasDrawableBalls() bool {
	return len(room.DrawableNumbers()) > 0
}

func (room *Room) AllPicked() bool {
	return len(room.DrawableNumbers()) == 0
}

func (room *Room) HasBingoRecord(userID UserID, line LineIndex) bool {
	return hasBingoRecord(room.BingoRecords, userID, line)
}

func (room *Room) HasReachRecord(userID UserID) bool {
	return hasReachRecord(room.ReachRecords, userID)
}
