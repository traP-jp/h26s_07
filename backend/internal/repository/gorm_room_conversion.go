package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

type roomRowFields struct {
	roomID        model.RoomID
	roomCode      model.RoomCode
	state         model.RoomState
	pickState     model.RoomPickState
	qrCodeVisible bool
	settings      model.RoomSettings
	createdAt     time.Time
	updatedAt     time.Time
}

func roomFieldsFromRow(row dbmodel.Room) (roomRowFields, error) {
	roomID, err := parseRoomID(row.RoomID)
	if err != nil {
		return roomRowFields{}, err
	}

	return roomRowFields{
		roomID:        roomID,
		roomCode:      model.RoomCode(row.RoomCode),
		state:         model.RoomState(row.State),
		pickState:     model.RoomPickState(row.PickState),
		qrCodeVisible: row.QRCodeVisible,
		settings: model.RoomSettings{
			Name:        row.SettingsName,
			Description: row.SettingsDescription,
			Admins:      []model.UserID{},
		},
		createdAt: row.CreatedAt,
		updatedAt: row.UpdatedAt,
	}, nil
}

func roomFromRow(row dbmodel.Room) (*model.Room, error) {
	fields, err := roomFieldsFromRow(row)
	if err != nil {
		return nil, err
	}

	return &model.Room{
		RoomID:        fields.roomID,
		RoomCode:      fields.roomCode,
		State:         fields.state,
		PickState:     fields.pickState,
		QrCodeVisible: fields.qrCodeVisible,
		Settings:      fields.settings,
		Participants:  []model.Participant{},
		Cards:         []model.Card{},
		PickedBalls:   []model.BallNumber{},
		BingoRecords:  []model.BingoRecord{},
		ReachRecords:  []model.ReachRecord{},
		Messages:      []model.Message{},
		CreatedAt:     fields.createdAt,
		UpdatedAt:     fields.updatedAt,
	}, nil
}

func aggregateTimestamp(room *model.Room) time.Time {
	if !room.UpdatedAt.IsZero() {
		return room.UpdatedAt
	}
	if !room.CreatedAt.IsZero() {
		return room.CreatedAt
	}
	return time.Now().UTC()
}
func parseRoomID(value string) (model.RoomID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return model.RoomID{}, fmt.Errorf("%w: invalid room id %q: %v", ErrInvalidRoomAggregate, value, err)
	}
	return model.RoomID(id), nil
}
func parseCardID(value string) (model.CardID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return model.CardID{}, fmt.Errorf("%w: invalid card id %q: %v", ErrInvalidRoomAggregate, value, err)
	}
	return model.CardID(id), nil
}
func parseRecordID(value string) (model.RecordID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return model.RecordID{}, fmt.Errorf("%w: invalid record id %q: %v", ErrInvalidRoomAggregate, value, err)
	}
	return model.RecordID(id), nil
}
func parseMessageID(value string) (model.MessageID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return model.MessageID{}, fmt.Errorf("%w: invalid message id %q: %v", ErrInvalidRoomAggregate, value, err)
	}
	return model.MessageID(id), nil
}
func roomIDString(roomID model.RoomID) string {
	return uuid.UUID(roomID).String()
}
func cardIDString(cardID model.CardID) string {
	return uuid.UUID(cardID).String()
}
func recordIDString(recordID model.RecordID) string {
	return uuid.UUID(recordID).String()
}
func messageIDString(messageID model.MessageID) string {
	return uuid.UUID(messageID).String()
}
func ballNumberPtrToUint8Ptr(number *model.BallNumber) *uint8 {
	if number == nil {
		return nil
	}
	value := uint8(*number)
	return &value
}
func uint8PtrToBallNumberPtr(number *uint8) *model.BallNumber {
	if number == nil {
		return nil
	}
	value := model.BallNumber(*number)
	return &value
}
