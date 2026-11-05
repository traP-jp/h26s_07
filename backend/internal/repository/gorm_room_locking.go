package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

// FindByIDForUpdate locks the room row before loading it. Call it in the same transaction as the mutation and save so concurrent read-modify-write operations serialize.
func (r *GormRoomRepository) FindByIDForUpdate(ctx context.Context, roomID model.RoomID) (*model.Room, error) {
	var roomRow dbmodel.Room
	err := gormDB(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("room_id = ?", roomIDString(roomID)).
		First(&roomRow).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}

	rooms, err := r.buildRooms(ctx, []dbmodel.Room{roomRow})
	if err != nil {
		return nil, err
	}
	return rooms[0], nil
}
