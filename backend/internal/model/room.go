package model

import (
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

func (id RoomID) String() string {
	return uuid.UUID(id).String()
}
