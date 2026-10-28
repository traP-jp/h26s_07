package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

func (r *GormRoomRepository) List(ctx context.Context) ([]model.RoomSummary, error) {
	var roomRows []dbmodel.Room
	// Room ID breaks creation-time ties so list order is stable.
	if err := gormDB(ctx, r.db).
		Order("created_at DESC").
		Order("room_id ASC").
		Find(&roomRows).
		Error; err != nil {
		return nil, err
	}
	return r.buildRoomSummaries(ctx, roomRows)
}

func (r *GormRoomRepository) buildRoomSummaries(ctx context.Context, roomRows []dbmodel.Room) ([]model.RoomSummary, error) {
	if len(roomRows) == 0 {
		return []model.RoomSummary{}, nil
	}

	summaries, roomIDs, summaryByID, err := roomSummariesFromRows(roomRows)
	if err != nil {
		return nil, err
	}
	db := gormDB(ctx, r.db)
	if err := loadRoomSummaries(db, roomIDs, summaryByID); err != nil {
		return nil, err
	}
	if err := loadRoomSummaryActivityLogs(db, roomIDs, summaryByID); err != nil {
		return nil, err
	}
	return summaries, nil
}

func roomSummariesFromRows(roomRows []dbmodel.Room) ([]model.RoomSummary, []string, map[string]*model.RoomSummary, error) {
	summaries := make([]model.RoomSummary, len(roomRows))
	summaryByID := make(map[string]*model.RoomSummary, len(roomRows))
	roomIDs := make([]string, 0, len(roomRows))
	for i, row := range roomRows {
		summary, err := roomSummaryFromRow(row)
		if err != nil {
			return nil, nil, nil, err
		}
		summaries[i] = summary
		roomID := roomIDString(summary.RoomID)
		roomIDs = append(roomIDs, roomID)
		summaryByID[roomID] = &summaries[i]
	}
	return summaries, roomIDs, summaryByID, nil
}

func loadRoomSummaries(db *gorm.DB, roomIDs []string, summaryByID map[string]*model.RoomSummary) error {
	if err := loadRoomSummaryAdmins(db, roomIDs, summaryByID); err != nil {
		return err
	}
	if err := loadRoomSummaryParticipants(db, roomIDs, summaryByID); err != nil {
		return err
	}
	if err := loadRoomSummaryBingoSummaries(db, roomIDs, summaryByID); err != nil {
		return err
	}
	if err := loadRoomSummaryReachSummaries(db, roomIDs, summaryByID); err != nil {
		return err
	}
	return nil
}

func loadRoomSummaryActivityLogs(db *gorm.DB, roomIDs []string, summaryByID map[string]*model.RoomSummary) error {
	recordRooms := make(map[string]*model.Room, len(roomIDs))
	for _, id := range roomIDs {
		recordRooms[id] = &model.Room{}
	}
	// ActivityLog combines reach and bingo records, so summaries need both ordered record streams.
	if err := loadBingoRecords(db, roomIDs, recordRooms); err != nil {
		return err
	}
	if err := loadReachRecords(db, roomIDs, recordRooms); err != nil {
		return err
	}
	for _, id := range roomIDs {
		summaryByID[id].ActivityLogEntries = recordRooms[id].ActivityLog()
	}
	return nil
}

func roomSummaryFromRow(row dbmodel.Room) (model.RoomSummary, error) {
	fields, err := roomFieldsFromRow(row)
	if err != nil {
		return model.RoomSummary{}, err
	}

	return model.RoomSummary{
		RoomID:         fields.roomID,
		RoomCode:       fields.roomCode,
		State:          fields.state,
		PickState:      fields.pickState,
		QrCodeVisible:  fields.qrCodeVisible,
		Settings:       fields.settings,
		Participants:   []model.Participant{},
		BingoSummaries: []model.BingoSummary{},
		ReachSummaries: []model.ReachSummary{},
		CreatedAt:      fields.createdAt,
		UpdatedAt:      fields.updatedAt,
	}, nil
}
func loadRoomSummaryBingoSummaries(db *gorm.DB, roomIDs []string, summaryByID map[string]*model.RoomSummary) error {
	var rows []dbmodel.RoomBingoRecord
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("bingo_order ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	summaryIndexByRoomUser := make(map[string]map[string]int, len(roomIDs))
	for _, row := range rows {
		indexByUser, ok := summaryIndexByRoomUser[row.RoomID]
		if !ok {
			indexByUser = map[string]int{}
			summaryIndexByRoomUser[row.RoomID] = indexByUser
		}

		summary := summaryByID[row.RoomID]
		index, ok := indexByUser[row.UserID]
		if !ok {
			index = len(summary.BingoSummaries)
			indexByUser[row.UserID] = index
			summary.BingoSummaries = append(summary.BingoSummaries, model.BingoSummary{
				UserID:      model.UserID(row.UserID),
				BingoOrders: []model.BingoOrder{},
				CreatedAt:   row.CreatedAt,
			})
		}

		summary.BingoSummaries[index].BingoOrders = append(
			summary.BingoSummaries[index].BingoOrders,
			model.BingoOrder(row.BingoOrder),
		)
	}
	return nil
}
func loadRoomSummaryReachSummaries(db *gorm.DB, roomIDs []string, summaryByID map[string]*model.RoomSummary) error {
	var reachRows []dbmodel.RoomReachRecord
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("created_at ASC").
		Order("user_id ASC").
		Find(&reachRows).
		Error; err != nil {
		return err
	}

	reachRecordsByRoomID := make(map[string][]model.ReachRecord, len(roomIDs))
	for _, row := range reachRows {
		reachRecordsByRoomID[row.RoomID] = append(reachRecordsByRoomID[row.RoomID], model.ReachRecord{
			UserID:    model.UserID(row.UserID),
			CreatedAt: row.CreatedAt,
		})
	}

	bingoRecordsByRoomID := make(map[string][]model.BingoRecord, len(roomIDs))
	for _, roomID := range roomIDs {
		for _, bingoSummary := range summaryByID[roomID].BingoSummaries {
			bingoRecordsByRoomID[roomID] = append(bingoRecordsByRoomID[roomID], model.BingoRecord{
				UserID: bingoSummary.UserID,
			})
		}
	}

	for _, roomID := range roomIDs {
		summaryByID[roomID].ReachSummaries = model.ReachSummariesFromRecords(
			reachRecordsByRoomID[roomID],
			bingoRecordsByRoomID[roomID],
		)
	}
	return nil
}
