package model

import "time"

type BingoUpdate struct {
	UserID         UserID
	NewBingoOrders []BingoOrder
	BingoOrders    []BingoOrder
}

type ReachUpdate struct {
	UserID UserID
}

type PickFinishedResult struct {
	Room               *Room
	PickedBall         BallNumber
	ParticipantUpdates []ParticipantCardUpdate
	NewBingos          []BingoUpdate
	NewReaches         []ReachUpdate
	AllPicked          bool
}

type RecordIDGenerator func() (RecordID, error)

func (room *Room) StartPick(actor UserID, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	// A retry during an active draw leaves that draw and its animation in place.
	if room.State == RoomStatePlaying && room.PickState == RoomPickStatePicking {
		return nil
	}
	if room.State != RoomStatePlaying || room.PickState != RoomPickStateIdle {
		return ErrRoomPickNotStartable
	}
	if !room.HasDrawableBalls() {
		// Persist exhaustion even though the request reports the domain conflict.
		room.PickState = RoomPickStateExhausted
		room.UpdatedAt = now
		return ErrNoDrawableBalls
	}
	room.PickState = RoomPickStatePicking
	room.UpdatedAt = now
	return nil
}

func (room *Room) CancelPick(actor UserID, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	if room.State != RoomStatePlaying || room.PickState != RoomPickStatePicking {
		return ErrRoomPickNotCancelable
	}
	room.PickState = RoomPickStateIdle
	room.UpdatedAt = now
	return nil
}

func (room *Room) FinishPick(actor UserID, picked BallNumber, nextRecordID RecordIDGenerator, now time.Time) (PickFinishedResult, error) {
	if err := room.validateFinishPick(actor, picked, nextRecordID); err != nil {
		return PickFinishedResult{}, err
	}

	changes := newPickChanges(room)
	changes.recordPickedBall(picked)
	if err := changes.markCards(picked, nextRecordID, now); err != nil {
		return PickFinishedResult{}, err
	}
	room.commitPick(changes, now)

	return buildPickFinishedResult(room, picked, changes), nil
}

func (room *Room) validateFinishPick(actor UserID, picked BallNumber, nextRecordID RecordIDGenerator) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	if room.State != RoomStatePlaying || room.PickState != RoomPickStatePicking {
		return ErrRoomPickNotFinishable
	}
	if !picked.Valid() {
		return ErrInvalidBallNumber
	}
	if ballPicked(room.PickedBalls, picked) {
		return ErrBallAlreadyPicked
	}
	if nextRecordID == nil {
		return ErrRecordIDRequired
	}
	return nil
}

type pickChanges struct {
	cards              []Card
	pickedBalls        []BallNumber
	bingoRecords       []BingoRecord
	reachRecords       []ReachRecord
	newBingoRecords    []BingoRecord
	newReachRecords    []ReachRecord
	participantUpdates []ParticipantCardUpdate
}

func newPickChanges(room *Room) pickChanges {
	return pickChanges{
		cards:              append([]Card(nil), room.Cards...),
		pickedBalls:        append([]BallNumber(nil), room.PickedBalls...),
		bingoRecords:       append([]BingoRecord(nil), room.BingoRecords...),
		reachRecords:       append([]ReachRecord(nil), room.ReachRecords...),
		newBingoRecords:    make([]BingoRecord, 0),
		newReachRecords:    make([]ReachRecord, 0),
		participantUpdates: make([]ParticipantCardUpdate, 0, len(room.Cards)),
	}
}

func (changes *pickChanges) recordPickedBall(picked BallNumber) {
	changes.pickedBalls = append(changes.pickedBalls, picked)
}

func (changes *pickChanges) markCards(picked BallNumber, nextRecordID RecordIDGenerator, now time.Time) error {
	for i := range changes.cards {
		update, err := updateCardForPick(changes.cards[i], picked, changes, nextRecordID, now)
		if err != nil {
			return err
		}
		changes.cards[i] = update.Card
		changes.participantUpdates = append(changes.participantUpdates, update)
	}
	return nil
}

func updateCardForPick(card Card, picked BallNumber, changes *pickChanges, nextRecordID RecordIDGenerator, now time.Time) (ParticipantCardUpdate, error) {
	before := card
	beforeReachLines := card.ReachLines(changes.bingoRecords)
	card.OpenNumber(picked)

	newBingoLines := card.NewBingoLines(changes.bingoRecords)
	if err := recordBingoLines(card, newBingoLines, changes, nextRecordID, now); err != nil {
		return ParticipantCardUpdate{}, err
	}
	card.MarkBingoLines(newBingoLines)

	reachLines := card.ReachLines(changes.bingoRecords)
	newReachLines := lineDifference(reachLines, beforeReachLines)
	if err := recordFirstReach(card, newReachLines, changes, nextRecordID, now); err != nil {
		return ParticipantCardUpdate{}, err
	}
	card.MarkReachLines(reachLines)

	return ParticipantCardUpdate{
		UserID: card.OwnerUserID,
		Card:   card,
		CardChanges: CardChanges{
			OpenedCellIndices: card.NewlyOpenedCells(before),
			NewReachLines:     newReachLines,
			NewBingoLines:     newBingoLines,
		},
	}, nil
}

func recordBingoLines(card Card, lines []LineIndex, changes *pickChanges, nextRecordID RecordIDGenerator, now time.Time) error {
	for _, line := range lines {
		recordID, err := nextRecordID()
		if err != nil {
			return err
		}
		record := BingoRecord{
			RecordID:  recordID,
			UserID:    card.OwnerUserID,
			Line:      line,
			Order:     BingoOrder(len(changes.bingoRecords) + 1),
			CreatedAt: now,
		}
		changes.bingoRecords = append(changes.bingoRecords, record)
		changes.newBingoRecords = append(changes.newBingoRecords, record)
	}
	return nil
}

func recordFirstReach(card Card, newLines []LineIndex, changes *pickChanges, nextRecordID RecordIDGenerator, now time.Time) error {
	if len(newLines) == 0 || hasReachRecord(changes.reachRecords, card.OwnerUserID) {
		return nil
	}
	recordID, err := nextRecordID()
	if err != nil {
		return err
	}
	lastCellIndex, _ := card.LastMissingCellIndex(newLines[0])
	record := ReachRecord{
		RecordID:      recordID,
		UserID:        card.OwnerUserID,
		Line:          newLines[0],
		LastCellIndex: lastCellIndex,
		CreatedAt:     now,
	}
	changes.reachRecords = append(changes.reachRecords, record)
	changes.newReachRecords = append(changes.newReachRecords, record)
	return nil
}

func (room *Room) commitPick(changes pickChanges, now time.Time) {
	// Apply changes only after all IDs exist so a generator error cannot leave a partial pick.
	room.Cards = changes.cards
	room.PickedBalls = changes.pickedBalls
	room.BingoRecords = changes.bingoRecords
	room.ReachRecords = changes.reachRecords
	if len(DrawableNumbers(changes.pickedBalls)) == 0 {
		room.PickState = RoomPickStateExhausted
	} else {
		room.PickState = RoomPickStateIdle
	}
	room.UpdatedAt = now
}

func buildPickFinishedResult(room *Room, picked BallNumber, changes pickChanges) PickFinishedResult {
	return PickFinishedResult{
		Room:               room,
		PickedBall:         picked,
		ParticipantUpdates: changes.participantUpdates,
		NewBingos:          bingoUpdates(changes.newBingoRecords, changes.bingoRecords),
		NewReaches:         reachUpdates(changes.newReachRecords),
		AllPicked:          room.AllPicked(),
	}
}

func ballPicked(pickedBalls []BallNumber, number BallNumber) bool {
	for _, picked := range pickedBalls {
		if picked == number {
			return true
		}
	}
	return false
}

func hasReachRecord(records []ReachRecord, userID UserID) bool {
	for _, record := range records {
		if record.UserID == userID {
			return true
		}
	}
	return false
}

func lineDifference(lines []LineIndex, existing []LineIndex) []LineIndex {
	existingSet := make(map[LineIndex]struct{}, len(existing))
	for _, line := range existing {
		existingSet[line] = struct{}{}
	}
	result := make([]LineIndex, 0, len(lines))
	for _, line := range lines {
		if _, ok := existingSet[line]; !ok {
			result = append(result, line)
		}
	}
	return result
}

func bingoUpdates(newRecords []BingoRecord, allRecords []BingoRecord) []BingoUpdate {
	updates := make([]BingoUpdate, 0)
	indexByUserID := make(map[UserID]int)
	for _, record := range newRecords {
		index, ok := indexByUserID[record.UserID]
		if !ok {
			index = len(updates)
			indexByUserID[record.UserID] = index
			updates = append(updates, BingoUpdate{
				UserID:         record.UserID,
				NewBingoOrders: []BingoOrder{},
				BingoOrders:    bingoOrders(allRecords, record.UserID),
			})
		}
		updates[index].NewBingoOrders = append(updates[index].NewBingoOrders, record.Order)
	}
	return updates
}

func bingoOrders(records []BingoRecord, userID UserID) []BingoOrder {
	orders := make([]BingoOrder, 0)
	for _, record := range records {
		if record.UserID == userID {
			orders = append(orders, record.Order)
		}
	}
	return orders
}

func reachUpdates(records []ReachRecord) []ReachUpdate {
	updates := make([]ReachUpdate, 0, len(records))
	for _, record := range records {
		updates = append(updates, ReachUpdate{UserID: record.UserID})
	}
	return updates
}
