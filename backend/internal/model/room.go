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
	Admins      []User
}

type Participant struct {
	User     User
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
type Card struct {
	CardID      CardID
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

type MassageID uuid.UUID
type Massage struct {
	MassageID MassageID
	Content   string
	Author    User
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
	Massages      []Massage
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
		Massages:      []Massage{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (settings *RoomSettings) HasAdmin(userID UserID) bool {
	for _, admin := range settings.Admins {
		if admin.UserID == userID {
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
		if participant.User.UserID == userID {
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

func (room *Room) CanStartPick(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID) && room.PickState == RoomPickStateIdle
}

func (room *Room) CanCancelPick(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID) && room.PickState == RoomPickStatePicking
}

func (room *Room) CanFinishPick(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID) && room.PickState == RoomPickStatePicking
}

func (room *Room) ParticipantCount() int {
	return len(room.Participants)
}

//Todo
//func (room *Room) DrawableNumbers() []BallNumber {
//}

//Todo
//func (room *Room) AllPicked() bool {
//	return len(room.PickedBalls) == room.Settings.MaxBalls

func (room *Room) Join(user User, now time.Time) error {
	if !room.CanJoin() {
		return ErrRoomNotJoinable
	}
	if room.IsParticipant(user.UserID) {
		return nil
	}

	room.Participants = append(room.Participants, Participant{
		User:     user,
		JoinedAt: now,
	})
	room.UpdatedAt = now
	return nil
}

func (room *Room) CanPostMassage(user User) bool {
	return (room.IsParticipant(user.UserID) || room.IsAdmin(user.UserID)) && room.State != RoomStateFinished
}

func (room *Room) PostMassage(user User, content string, now time.Time, massageID MassageID) (Massage, error) {
	if !room.CanPostMassage(user) {
		return Massage{}, ErrRoomMassageNotAllowed
	}
	if content == "" || len(content) > 500 {
		return Massage{}, ErrMassageInvalid
	}
	massage := Massage{
		MassageID: massageID,
		Content:   content,
		Author:    user,
		CreatedAt: now,
	}
	room.Massages = append(room.Massages, massage)
	room.UpdatedAt = now
	return massage, nil
}
