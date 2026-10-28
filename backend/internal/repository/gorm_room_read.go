package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

func (r *GormRoomRepository) FindByID(ctx context.Context, roomID model.RoomID) (*model.Room, error) {
	var roomRow dbmodel.Room
	err := gormDB(ctx, r.db).
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

func (r *GormRoomRepository) ActiveCodeExists(ctx context.Context, code model.RoomCode) (bool, error) {
	var count int64
	err := gormDB(ctx, r.db).
		Model(&dbmodel.Room{}).
		Where("room_code = ? AND state <> ?", string(code), string(model.RoomStateFinished)).
		Count(&count).
		Error
	return count > 0, err
}

func (r *GormRoomRepository) buildRooms(ctx context.Context, roomRows []dbmodel.Room) ([]*model.Room, error) {
	if len(roomRows) == 0 {
		return []*model.Room{}, nil
	}

	rooms := make([]*model.Room, 0, len(roomRows))
	roomByID := make(map[string]*model.Room, len(roomRows))
	roomIDs := make([]string, 0, len(roomRows))
	for _, row := range roomRows {
		room, err := roomFromRow(row)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
		roomID := roomIDString(room.RoomID)
		roomIDs = append(roomIDs, roomID)
		roomByID[roomID] = room
	}

	db := gormDB(ctx, r.db)
	if err := loadRoomAdmins(db, roomIDs, roomByID); err != nil {
		return nil, err
	}
	if err := loadRoomParticipants(db, roomIDs, roomByID); err != nil {
		return nil, err
	}
	if err := loadRoomCards(db, roomIDs, roomByID); err != nil {
		return nil, err
	}
	if err := loadPickedBalls(db, roomIDs, roomByID); err != nil {
		return nil, err
	}
	if err := loadBingoRecords(db, roomIDs, roomByID); err != nil {
		return nil, err
	}
	if err := loadReachRecords(db, roomIDs, roomByID); err != nil {
		return nil, err
	}
	if err := loadMessages(db, roomIDs, roomByID); err != nil {
		return nil, err
	}

	return rooms, nil
}
