package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

func (r *GormRoomRepository) replacePickedBalls(tx *gorm.DB, room *model.Room) error {
	roomID := roomIDString(room.RoomID)
	if err := tx.Where("room_id = ?", roomID).Delete(&dbmodel.RoomPickedBall{}).Error; err != nil {
		return err
	}
	if len(room.PickedBalls) == 0 {
		return nil
	}

	rows := make([]dbmodel.RoomPickedBall, 0, len(room.PickedBalls))
	pickedAt := aggregateTimestamp(room)
	for i, number := range room.PickedBalls {
		rows = append(rows, dbmodel.RoomPickedBall{
			PickedBallID: uuid.NewString(),
			RoomID:       roomID,
			PickOrder:    uint16(i + 1),
			Number:       uint8(number),
			PickedAt:     pickedAt,
		})
	}
	return createInBatches(tx, rows)
}

func (r *GormRoomRepository) replaceBingoRecords(tx *gorm.DB, room *model.Room) error {
	roomID := roomIDString(room.RoomID)
	if err := tx.Where("room_id = ?", roomID).Delete(&dbmodel.RoomBingoRecord{}).Error; err != nil {
		return err
	}
	if len(room.BingoRecords) == 0 {
		return nil
	}

	cardIDByOwner := make(map[model.UserID]string, len(room.Cards))
	for _, card := range room.Cards {
		cardIDByOwner[card.OwnerUserID] = cardIDString(card.CardID)
	}

	rows := make([]dbmodel.RoomBingoRecord, 0, len(room.BingoRecords))
	for _, record := range room.BingoRecords {
		cardID, ok := cardIDByOwner[record.UserID]
		if !ok {
			return fmt.Errorf("%w: bingo record user %s has no card", ErrInvalidRoomAggregate, record.UserID)
		}
		createdAt := record.CreatedAt
		if createdAt.IsZero() {
			createdAt = aggregateTimestamp(room)
		}
		rows = append(rows, dbmodel.RoomBingoRecord{
			RecordID:   recordIDString(record.RecordID),
			RoomID:     roomID,
			UserID:     string(record.UserID),
			CardID:     cardID,
			LineIndex:  uint8(record.Line),
			BingoOrder: uint(record.Order),
			CreatedAt:  createdAt,
		})
	}
	return createInBatches(tx, rows)
}

func (r *GormRoomRepository) replaceReachRecords(tx *gorm.DB, room *model.Room) error {
	roomID := roomIDString(room.RoomID)
	if err := tx.Where("room_id = ?", roomID).Delete(&dbmodel.RoomReachRecord{}).Error; err != nil {
		return err
	}
	if len(room.ReachRecords) == 0 {
		return nil
	}

	rows := make([]dbmodel.RoomReachRecord, 0, len(room.ReachRecords))
	for _, record := range room.ReachRecords {
		createdAt := record.CreatedAt
		if createdAt.IsZero() {
			createdAt = aggregateTimestamp(room)
		}
		rows = append(rows, dbmodel.RoomReachRecord{
			RecordID:      recordIDString(record.RecordID),
			RoomID:        roomID,
			UserID:        string(record.UserID),
			LineIndex:     uint8(record.Line),
			LastCellIndex: uint8(record.LastCellIndex),
			CreatedAt:     createdAt,
		})
	}
	return createInBatches(tx, rows)
}

func (r *GormRoomRepository) replaceMessages(tx *gorm.DB, room *model.Room) error {
	roomID := roomIDString(room.RoomID)
	if err := tx.Where("room_id = ?", roomID).Delete(&dbmodel.RoomMessage{}).Error; err != nil {
		return err
	}
	if len(room.Messages) == 0 {
		return nil
	}

	rows := make([]dbmodel.RoomMessage, 0, len(room.Messages))
	for _, message := range room.Messages {
		if err := validateUserID(message.Author); err != nil {
			return err
		}
		rows = append(rows, dbmodel.RoomMessage{
			MessageID:    messageIDString(message.MessageID),
			RoomID:       roomID,
			AuthorUserID: string(message.Author),
			Content:      message.Content,
			CreatedAt:    message.CreatedAt,
		})
	}
	return createInBatches(tx, rows)
}

func loadPickedBalls(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	var rows []dbmodel.RoomPickedBall
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("pick_order ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	for _, row := range rows {
		roomByID[row.RoomID].PickedBalls = append(roomByID[row.RoomID].PickedBalls, model.BallNumber(row.Number))
	}
	return nil
}

func loadBingoRecords(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	var rows []dbmodel.RoomBingoRecord
	// Bingo order is the persisted event sequence, so hydration must preserve it.
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("bingo_order ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	for _, row := range rows {
		recordID, err := parseRecordID(row.RecordID)
		if err != nil {
			return err
		}
		roomByID[row.RoomID].BingoRecords = append(roomByID[row.RoomID].BingoRecords, model.BingoRecord{
			RecordID:  recordID,
			UserID:    model.UserID(row.UserID),
			Line:      model.LineIndex(row.LineIndex),
			Order:     model.BingoOrder(row.BingoOrder),
			CreatedAt: row.CreatedAt,
		})
	}
	return nil
}

func loadReachRecords(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	var rows []dbmodel.RoomReachRecord
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("created_at ASC").
		Order("user_id ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	for _, row := range rows {
		recordID, err := parseRecordID(row.RecordID)
		if err != nil {
			return err
		}
		roomByID[row.RoomID].ReachRecords = append(roomByID[row.RoomID].ReachRecords, model.ReachRecord{
			RecordID:      recordID,
			UserID:        model.UserID(row.UserID),
			Line:          model.LineIndex(row.LineIndex),
			LastCellIndex: model.CellIndex(row.LastCellIndex),
			CreatedAt:     row.CreatedAt,
		})
	}
	return nil
}

func loadMessages(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	var rows []dbmodel.RoomMessage
	// Message IDs break timestamp ties so loaded chat history is stable.
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("created_at ASC").
		Order("message_id ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	for _, row := range rows {
		messageID, err := parseMessageID(row.MessageID)
		if err != nil {
			return err
		}
		roomByID[row.RoomID].Messages = append(roomByID[row.RoomID].Messages, model.Message{
			MessageID: messageID,
			Content:   row.Content,
			Author:    model.UserID(row.AuthorUserID),
			CreatedAt: row.CreatedAt,
		})
	}
	return nil
}
