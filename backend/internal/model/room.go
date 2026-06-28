package model

import (
	"cmp"
	"slices"
	"time"

	"github.com/google/uuid"
)

type RoomID uuid.UUID
type RoomCode string
type RoomState string

const (
	RoomStateWaiting  RoomState = "waiting"
	RoomStatePlaying  RoomState = "playing"
	RoomStateFinished RoomState = "finished"
)

type RoomPickState string

const (
	RoomPickStateIdle      RoomPickState = "idle"
	RoomPickStatePicking   RoomPickState = "picking"
	RoomPickStateExhausted RoomPickState = "exhausted"
)

type UserID string

type User struct {
	UserID UserID
}

type RoomSettings struct {
	Name        string
	Description string
	Admins      []UserID
}

type Participant struct {
	UserID   UserID
	JoinedAt time.Time
}

type CardCellState string

const (
	CardCellStateBingo  CardCellState = "bingo"
	CardCellStateReach  CardCellState = "reach"
	CardCellStateOpen   CardCellState = "open"
	CardCellStateClosed CardCellState = "closed"
)

type CellIndex uint8
type BallNumber uint8

type CardCell struct {
	Index     CellIndex
	Number    *BallNumber
	CellState CardCellState
}

type CardID uuid.UUID
type CardNumber string

func (number CardNumber) Valid() bool {
	if len(number) != 36 {
		return false
	}
	for _, r := range number {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

type Card struct {
	CardID      CardID
	CardNumber  CardNumber
	OwnerUserID UserID
	Cells       [25]CardCell
}

type RecordID uuid.UUID
type LineIndex uint8
type BingoOrder uint

type BingoRecord struct {
	RecordID  RecordID
	UserID    UserID
	Line      LineIndex
	Order     BingoOrder
	CreatedAt time.Time
}

type BingoSummary struct {
	UserID      UserID
	BingoOrders []BingoOrder
}

type ReachSummary struct {
	UserID UserID
}

type ReachRecord struct {
	RecordID      RecordID
	UserID        UserID
	Line          LineIndex
	LastCellIndex CellIndex
	CreatedAt     time.Time
}

type MessageID uuid.UUID
type Message struct {
	MessageID MessageID
	Content   string
	Author    UserID
	CreatedAt time.Time
}

type Room struct {
	RoomID        RoomID
	RoomCode      RoomCode
	State         RoomState
	PickState     RoomPickState
	QrCodeVisible bool
	Settings      RoomSettings
	Participants  []Participant
	Cards         []Card
	PickedBalls   []BallNumber
	BingoRecords  []BingoRecord
	ReachRecords  []ReachRecord
	Messages      []Message
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (id RoomID) String() string {
	return uuid.UUID(id).String()
}

type RoomSummary struct {
	RoomID         RoomID
	RoomCode       RoomCode
	State          RoomState
	PickState      RoomPickState
	QrCodeVisible  bool
	Settings       RoomSettings
	Participants   []Participant
	BingoSummaries []BingoSummary
	ReachSummaries []ReachSummary
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewRoom(
	roomID RoomID,
	roomCode RoomCode,
	roomsettings RoomSettings,
	now time.Time,
) *Room {
	return &Room{
		RoomID:        roomID,
		RoomCode:      roomCode,
		State:         RoomStateWaiting,
		PickState:     RoomPickStateIdle,
		QrCodeVisible: false,
		Settings:      roomsettings,
		Participants:  []Participant{},
		Cards:         []Card{},
		PickedBalls:   []BallNumber{},
		BingoRecords:  []BingoRecord{},
		ReachRecords:  []ReachRecord{},
		Messages:      []Message{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (settings RoomSettings) HasAdmin(userID UserID) bool {
	for _, admin := range settings.Admins {
		if admin == userID {
			return true
		}
	}
	return false
}

func (room *Room) IsAdmin(userID UserID) bool {
	return room.Settings.HasAdmin(userID)
}

func (room *Room) IsParticipant(userID UserID) bool {
	for _, participant := range room.Participants {
		if participant.UserID == userID {
			return true
		}
	}
	return false
}

func (room *Room) CanJoin() bool {
	return room.State == RoomStateWaiting
}

func (room *Room) CanStart(userID UserID) bool {
	return room.State == RoomStateWaiting && len(room.Participants) > 0 && room.IsAdmin(userID)
}

func (room *Room) CanFinish(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID)
}

// TODO あとで消して別々にチェックして別のエラーを返すようにする
func (room *Room) CanFinishPick(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID) && room.PickState == RoomPickStatePicking
}

func (room *Room) ParticipantCount() int {
	return len(room.Participants)
}

func (room *Room) Join(userID UserID, now time.Time) error {
	if !room.CanJoin() {
		return ErrRoomNotJoinable
	}
	if room.IsParticipant(userID) {
		return nil
	}

	room.Participants = append(room.Participants, Participant{
		UserID:   userID,
		JoinedAt: now,
	})
	room.UpdatedAt = now
	return nil
}

func (room *Room) CanPostMessage(user UserID) bool {
	return (room.IsParticipant(user) || room.IsAdmin(user)) && room.State != RoomStateFinished
}

func (room *Room) PostMessage(user UserID, content string, now time.Time, messageID MessageID) (Message, error) {
	if !room.IsParticipant(user) && !room.IsAdmin(user) {
		return Message{}, ErrRoomMessageNotAllowed
	}
	if room.State == RoomStateFinished {
		return Message{}, ErrRoomMessageNotPostable
	}
	if content == "" || len(content) > 500 {
		return Message{}, ErrMessageInvalid
	}
	message := Message{
		MessageID: messageID,
		Content:   content,
		Author:    user,
		CreatedAt: now,
	}
	room.Messages = append(room.Messages, message)
	room.UpdatedAt = now
	return message, nil
}

type CardChanges struct {
	OpenedCellIndices []CellIndex
	NewReachLines     []LineIndex
	NewBingoLines     []LineIndex
}

type ParticipantCardUpdate struct {
	UserID      UserID
	Card        Card
	CardChanges CardChanges
}

type BingoUpdate struct {
	UserID         UserID
	NewBingoOrders []BingoOrder
	BingoOrders    []BingoOrder
}

type ReachUpdate struct {
	UserID UserID
}

type GameStartedResult struct {
	ParticipantCards []ParticipantCardUpdate
}

type GameFinishedResult struct {
	PickCanceled       bool
	ParticipantUpdates []ParticipantCardUpdate
}

type PickFinishedResult struct {
	Room               *Room
	PickedBall         BallNumber
	ParticipantUpdates []ParticipantCardUpdate
	NewBingos          []BingoUpdate
	NewReaches         []ReachUpdate
	AllPicked          bool
}

func (room *Room) CanView(userID UserID) bool {
	return room.IsParticipant(userID) || room.IsAdmin(userID)
}

func (room *Room) CanViewChat(userID UserID) bool {
	return room.CanView(userID)
}

func (room *Room) CanUpdateSettings(userID UserID) bool {
	return room.State != RoomStateFinished && room.IsAdmin(userID)
}

func (room *Room) CanFinishGame(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID)
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
		summaries = append(summaries, ReachSummary{UserID: record.UserID})
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
	for _, record := range room.ReachRecords {
		if record.UserID == userID {
			return true
		}
	}
	return false
}

type RecordIDGenerator func() (RecordID, error)

func (settings RoomSettings) Valid() bool {
	return len(settings.Admins) > 0
}

func (room *Room) UpdateSettings(actor UserID, settings RoomSettings, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	if room.State == RoomStateFinished {
		return ErrRoomNotConfigurable
	}
	if !settings.Valid() {
		return ErrRoomSettingsInvalid
	}
	room.Settings = settings
	room.UpdatedAt = now
	return nil
}

func (room *Room) StartGame(actor UserID, cards []Card, now time.Time) (GameStartedResult, error) {
	if room.State != RoomStateWaiting || len(room.Participants) == 0 {
		return GameStartedResult{}, ErrRoomNotStartable
	} else if !room.IsAdmin(actor) {
		return GameStartedResult{}, ErrRoomForbidden
	}
	if !cardsMatchParticipants(cards, room.Participants) {
		return GameStartedResult{}, ErrInvalidCard
	}

	room.State = RoomStatePlaying
	room.PickState = RoomPickStateIdle
	room.Cards = append([]Card(nil), cards...)
	room.UpdatedAt = now

	updates := make([]ParticipantCardUpdate, 0, len(cards))
	for _, card := range cards {
		updates = append(updates, ParticipantCardUpdate{
			UserID: card.OwnerUserID,
			Card:   card,
		})
	}
	return GameStartedResult{ParticipantCards: updates}, nil
}

func (room *Room) FinishGame(actor UserID, now time.Time) (GameFinishedResult, error) {
	if !room.CanFinishGame(actor) {
		return GameFinishedResult{}, ErrRoomNotFinishable
	}

	pickCanceled := room.PickState == RoomPickStatePicking
	room.State = RoomStateFinished
	room.PickState = RoomPickStateIdle
	room.UpdatedAt = now

	updates := make([]ParticipantCardUpdate, 0, len(room.Cards))
	for _, card := range room.Cards {
		updates = append(updates, ParticipantCardUpdate{
			UserID: card.OwnerUserID,
			Card:   card,
		})
	}
	return GameFinishedResult{
		PickCanceled:       pickCanceled,
		ParticipantUpdates: updates,
	}, nil
}

func (room *Room) StartPick(actor UserID, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	if room.State != RoomStatePlaying || room.PickState != RoomPickStateIdle {
		return ErrRoomPickNotStartable
	}
	if !room.HasDrawableBalls() {
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
	if !room.CanFinishPick(actor) {
		return PickFinishedResult{}, ErrRoomPickNotFinishable
	}
	if !picked.Valid() {
		return PickFinishedResult{}, ErrInvalidBallNumber
	}
	if ballPicked(room.PickedBalls, picked) {
		return PickFinishedResult{}, ErrBallAlreadyPicked
	}
	if nextRecordID == nil {
		return PickFinishedResult{}, ErrRecordIDRequired
	}

	cards := append([]Card(nil), room.Cards...)
	pickedBalls := append(append([]BallNumber(nil), room.PickedBalls...), picked)
	bingoRecords := append([]BingoRecord(nil), room.BingoRecords...)
	reachRecords := append([]ReachRecord(nil), room.ReachRecords...)
	newBingoRecords := make([]BingoRecord, 0)
	newReachRecords := make([]ReachRecord, 0)
	participantUpdates := make([]ParticipantCardUpdate, 0, len(cards))

	for i := range cards {
		before := cards[i]
		beforeReachLines := before.ReachLines(bingoRecords)

		cards[i].OpenNumber(picked)
		newBingoLines := cards[i].NewBingoLines(bingoRecords)
		for _, line := range newBingoLines {
			recordID, err := nextRecordID()
			if err != nil {
				return PickFinishedResult{}, err
			}
			record := BingoRecord{
				RecordID:  recordID,
				UserID:    cards[i].OwnerUserID,
				Line:      line,
				Order:     BingoOrder(len(bingoRecords) + 1),
				CreatedAt: now,
			}
			bingoRecords = append(bingoRecords, record)
			newBingoRecords = append(newBingoRecords, record)
		}
		cards[i].MarkBingoLines(newBingoLines)

		reachLines := cards[i].ReachLines(bingoRecords)
		newReachLines := lineDifference(reachLines, beforeReachLines)
		if len(newReachLines) > 0 && !hasReachRecord(reachRecords, cards[i].OwnerUserID) {
			recordID, err := nextRecordID()
			if err != nil {
				return PickFinishedResult{}, err
			}
			lastCellIndex, _ := cards[i].LastMissingCellIndex(newReachLines[0])
			record := ReachRecord{
				RecordID:      recordID,
				UserID:        cards[i].OwnerUserID,
				Line:          newReachLines[0],
				LastCellIndex: lastCellIndex,
				CreatedAt:     now,
			}
			reachRecords = append(reachRecords, record)
			newReachRecords = append(newReachRecords, record)
		}
		cards[i].MarkReachLines(reachLines)

		participantUpdates = append(participantUpdates, ParticipantCardUpdate{
			UserID: cards[i].OwnerUserID,
			Card:   cards[i],
			CardChanges: CardChanges{
				OpenedCellIndices: cards[i].NewlyOpenedCells(before),
				NewReachLines:     newReachLines,
				NewBingoLines:     newBingoLines,
			},
		})
	}

	room.Cards = cards
	room.PickedBalls = pickedBalls
	room.BingoRecords = bingoRecords
	room.ReachRecords = reachRecords
	if len(DrawableNumbers(pickedBalls)) == 0 {
		room.PickState = RoomPickStateExhausted
	} else {
		room.PickState = RoomPickStateIdle
	}
	room.UpdatedAt = now

	return PickFinishedResult{
		Room:               room,
		PickedBall:         picked,
		ParticipantUpdates: participantUpdates,
		NewBingos:          bingoUpdates(newBingoRecords, bingoRecords),
		NewReaches:         reachUpdates(newReachRecords),
		AllPicked:          room.AllPicked(),
	}, nil
}

func (room *Room) ShowQRCode(actor UserID, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	room.QrCodeVisible = true
	room.UpdatedAt = now
	return nil
}

func (room *Room) HideQRCode(actor UserID, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	room.QrCodeVisible = false
	room.UpdatedAt = now
	return nil
}

func cardsMatchParticipants(cards []Card, participants []Participant) bool {
	if len(cards) != len(participants) {
		return false
	}
	participantSet := make(map[UserID]struct{}, len(participants))
	for _, participant := range participants {
		participantSet[participant.UserID] = struct{}{}
	}
	for _, card := range cards {
		if _, ok := participantSet[card.OwnerUserID]; !ok {
			return false
		}
		delete(participantSet, card.OwnerUserID)
	}
	return len(participantSet) == 0
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
