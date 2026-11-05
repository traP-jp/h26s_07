package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

type GormRoomRepository struct {
	db *gorm.DB
}

const createBatchSize = 500

func NewGormRoomRepository(db *gorm.DB) *GormRoomRepository {
	return &GormRoomRepository{db: db}
}

// Save persists a complete room snapshot atomically. Chat, join, pick, and QR
// mutations use incremental methods to avoid rewriting historical rows.
func (r *GormRoomRepository) Save(ctx context.Context, room *model.Room) error {
	if room == nil {
		return fmt.Errorf("%w: room is nil", ErrInvalidRoomAggregate)
	}

	roomRow := dbmodel.Room{
		RoomID:              roomIDString(room.RoomID),
		RoomCode:            string(room.RoomCode),
		State:               string(room.State),
		PickState:           string(room.PickState),
		QRCodeVisible:       room.QrCodeVisible,
		SettingsName:        room.Settings.Name,
		SettingsDescription: room.Settings.Description,
		CreatedAt:           room.CreatedAt,
		UpdatedAt:           room.UpdatedAt,
	}

	if hasGormTransaction(ctx) {
		return r.save(ctx, room, roomRow)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, gormTransactionContextKey{}, tx)
		return r.save(txCtx, room, roomRow)
	})
}

// Delete removes children before the room row so foreign keys remain valid throughout the transaction.
func (r *GormRoomRepository) Delete(ctx context.Context, id model.RoomID) error {
	if !hasGormTransaction(ctx) {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := context.WithValue(ctx, gormTransactionContextKey{}, tx)
			return r.Delete(txCtx, id)
		})
	}
	tx := gormDB(ctx, r.db)
	roomID := roomIDString(id)
	for _, row := range []any{&dbmodel.RoomMessage{}, &dbmodel.RoomBingoRecord{}, &dbmodel.RoomReachRecord{}} {
		if err := tx.Where("room_id = ?", roomID).Delete(row).Error; err != nil {
			return err
		}
	}
	cards := tx.Model(&dbmodel.RoomCard{}).Select("card_id").Where("room_id = ?", roomID)
	if err := tx.Where("card_id IN (?)", cards).Delete(&dbmodel.RoomCardCell{}).Error; err != nil {
		return err
	}
	for _, row := range []any{&dbmodel.RoomCard{}, &dbmodel.RoomPickedBall{}, &dbmodel.RoomParticipant{}, &dbmodel.RoomAdmin{}, &dbmodel.RoomUser{}} {
		if err := tx.Where("room_id = ?", roomID).Delete(row).Error; err != nil {
			return err
		}
	}
	result := tx.Where("room_id = ?", roomID).Delete(&dbmodel.Room{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRoomNotFound
	}
	return nil
}

// save keeps the room row and all replacement child writes in one transaction.
func (r *GormRoomRepository) save(ctx context.Context, room *model.Room, roomRow dbmodel.Room) error {
	tx := gormDB(ctx, r.db)
	if err := tx.Save(&roomRow).Error; err != nil {
		return err
	}

	if err := r.saveRoomUsers(tx, room); err != nil {
		return err
	}
	if err := r.replaceRoomAdmins(tx, room); err != nil {
		return err
	}
	if err := r.replaceRoomParticipants(tx, room); err != nil {
		return err
	}
	if err := r.replaceRoomCards(tx, room); err != nil {
		return err
	}
	if err := r.replacePickedBalls(tx, room); err != nil {
		return err
	}
	if err := r.replaceBingoRecords(tx, room); err != nil {
		return err
	}
	if err := r.replaceReachRecords(tx, room); err != nil {
		return err
	}
	if err := r.replaceMessages(tx, room); err != nil {
		return err
	}

	return nil
}

func createInBatches[T any](tx *gorm.DB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}
	return tx.CreateInBatches(&rows, createBatchSize).Error
}
