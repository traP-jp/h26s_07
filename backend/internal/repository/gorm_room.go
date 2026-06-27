package repository

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *GormRoomRepository) List(ctx context.Context) ([]model.RoomSummary, error) {
	var roomRows []dbmodel.Room
	if err := gormDB(ctx, r.db).
		Order("created_at DESC").
		Order("room_id ASC").
		Find(&roomRows).
		Error; err != nil {
		return nil, err
	}
	return r.buildRoomSummaries(ctx, roomRows)
}

func (r *GormRoomRepository) saveRoomUsers(tx *gorm.DB, room *model.Room) error {
	rows, err := roomUserRows(room)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "room_id"},
			{Name: "user_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"user_name", "updated_at"}),
	}).CreateInBatches(&rows, createBatchSize).Error
}

func (r *GormRoomRepository) replaceRoomAdmins(tx *gorm.DB, room *model.Room) error {
	roomID := roomIDString(room.RoomID)
	if err := tx.Where("room_id = ?", roomID).Delete(&dbmodel.RoomAdmin{}).Error; err != nil {
		return err
	}

	adminIDs := uniqueUserIDs(room.Settings.Admins)
	if len(adminIDs) == 0 {
		return nil
	}

	rows := make([]dbmodel.RoomAdmin, 0, len(adminIDs))
	addedAt := aggregateTimestamp(room)
	for _, userID := range adminIDs {
		rows = append(rows, dbmodel.RoomAdmin{
			RoomID:  roomID,
			UserID:  string(userID),
			AddedAt: addedAt,
		})
	}
	return createInBatches(tx, rows)
}

func (r *GormRoomRepository) replaceRoomParticipants(tx *gorm.DB, room *model.Room) error {
	roomID := roomIDString(room.RoomID)
	if err := tx.Where("room_id = ?", roomID).Delete(&dbmodel.RoomParticipant{}).Error; err != nil {
		return err
	}
	if len(room.Participants) == 0 {
		return nil
	}

	rows := make([]dbmodel.RoomParticipant, 0, len(room.Participants))
	seen := make(map[model.UserID]struct{}, len(room.Participants))
	for _, participant := range room.Participants {
		userID := participant.UserID
		if err := validateUserID(userID); err != nil {
			return err
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		rows = append(rows, dbmodel.RoomParticipant{
			RoomID:   roomID,
			UserID:   string(userID),
			JoinedAt: participant.JoinedAt,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	return createInBatches(tx, rows)
}

func (r *GormRoomRepository) replaceRoomCards(tx *gorm.DB, room *model.Room) error {
	roomID := roomIDString(room.RoomID)

	var oldCardIDs []string
	if err := tx.Model(&dbmodel.RoomCard{}).
		Where("room_id = ?", roomID).
		Pluck("card_id", &oldCardIDs).
		Error; err != nil {
		return err
	}
	if len(oldCardIDs) > 0 {
		if err := tx.Where("card_id IN ?", oldCardIDs).Delete(&dbmodel.RoomCardCell{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("room_id = ?", roomID).Delete(&dbmodel.RoomCard{}).Error; err != nil {
		return err
	}
	if len(room.Cards) == 0 {
		return nil
	}

	cardRows := make([]dbmodel.RoomCard, 0, len(room.Cards))
	cellRows := make([]dbmodel.RoomCardCell, 0, len(room.Cards)*25)
	createdAt := aggregateTimestamp(room)
	seenCardIDs := make(map[string]struct{}, len(room.Cards))
	for _, card := range room.Cards {
		cardID := cardIDString(card.CardID)
		if _, ok := seenCardIDs[cardID]; ok {
			return fmt.Errorf("%w: duplicate card id %s", ErrInvalidRoomAggregate, cardID)
		}
		seenCardIDs[cardID] = struct{}{}
		if err := validateUserID(card.OwnerUserID); err != nil {
			return err
		}

		cardRows = append(cardRows, dbmodel.RoomCard{
			CardID:      cardID,
			RoomID:      roomID,
			OwnerUserID: string(card.OwnerUserID),
			CreatedAt:   createdAt,
		})
		for _, cell := range card.Cells {
			cellRows = append(cellRows, dbmodel.RoomCardCell{
				CardID:    cardID,
				CellIndex: uint8(cell.Index),
				Number:    ballNumberPtrToUint8Ptr(cell.Number),
				CellState: string(cell.CellState),
			})
		}
	}
	if err := createInBatches(tx, cardRows); err != nil {
		return err
	}
	return createInBatches(tx, cellRows)
}

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
	if len(room.Massages) == 0 {
		return nil
	}

	rows := make([]dbmodel.RoomMessage, 0, len(room.Massages))
	for _, message := range room.Massages {
		if err := validateUserID(message.Author); err != nil {
			return err
		}
		rows = append(rows, dbmodel.RoomMessage{
			MessageID:    messageIDString(message.MassageID),
			RoomID:       roomID,
			AuthorUserID: string(message.Author),
			Content:      message.Content,
			CreatedAt:    message.CreatedAt,
		})
	}
	return createInBatches(tx, rows)
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

func (r *GormRoomRepository) buildRoomSummaries(ctx context.Context, roomRows []dbmodel.Room) ([]model.RoomSummary, error) {
	if len(roomRows) == 0 {
		return []model.RoomSummary{}, nil
	}

	summaries := make([]model.RoomSummary, len(roomRows))
	summaryByID := make(map[string]*model.RoomSummary, len(roomRows))
	roomIDs := make([]string, 0, len(roomRows))
	for i, row := range roomRows {
		summary, err := roomSummaryFromRow(row)
		if err != nil {
			return nil, err
		}
		summaries[i] = summary
		roomID := roomIDString(summary.RoomID)
		roomIDs = append(roomIDs, roomID)
		summaryByID[roomID] = &summaries[i]
	}

	db := gormDB(ctx, r.db)
	if err := loadRoomSummaryAdmins(db, roomIDs, summaryByID); err != nil {
		return nil, err
	}
	if err := loadRoomSummaryParticipants(db, roomIDs, summaryByID); err != nil {
		return nil, err
	}
	if err := loadRoomSummaryBingoSummaries(db, roomIDs, summaryByID); err != nil {
		return nil, err
	}

	return summaries, nil
}

func roomFromRow(row dbmodel.Room) (*model.Room, error) {
	roomID, err := parseRoomID(row.RoomID)
	if err != nil {
		return nil, err
	}

	return &model.Room{
		RoomID:        roomID,
		RoomCode:      model.RoomCode(row.RoomCode),
		State:         model.RoomState(row.State),
		PickState:     model.RoomPickState(row.PickState),
		QrCodeVisible: row.QRCodeVisible,
		Settings: model.RoomSettings{
			Name:        row.SettingsName,
			Description: row.SettingsDescription,
			Admins:      []model.UserID{},
		},
		Participants: []model.Participant{},
		Cards:        []model.Card{},
		PickedBalls:  []model.BallNumber{},
		BingoRecords: []model.BingoRecord{},
		ReachRecords: []model.ReachRecord{},
		Massages:     []model.Massage{},
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func roomSummaryFromRow(row dbmodel.Room) (model.RoomSummary, error) {
	roomID, err := parseRoomID(row.RoomID)
	if err != nil {
		return model.RoomSummary{}, err
	}

	return model.RoomSummary{
		RoomID:        roomID,
		RoomCode:      model.RoomCode(row.RoomCode),
		State:         model.RoomState(row.State),
		PickState:     model.RoomPickState(row.PickState),
		QrCodeVisible: row.QRCodeVisible,
		Settings: model.RoomSettings{
			Name:        row.SettingsName,
			Description: row.SettingsDescription,
			Admins:      []model.UserID{},
		},
		Participants:   []model.Participant{},
		BingoSummaries: []model.BingoSummary{},
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}

func loadRoomAdmins(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	var rows []dbmodel.RoomAdmin
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("added_at ASC").
		Order("user_id ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	for _, row := range rows {
		room := roomByID[row.RoomID]
		room.Settings.Admins = append(room.Settings.Admins, model.UserID(row.UserID))
	}
	return nil
}

func loadRoomSummaryAdmins(db *gorm.DB, roomIDs []string, summaryByID map[string]*model.RoomSummary) error {
	var rows []dbmodel.RoomAdmin
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("added_at ASC").
		Order("user_id ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	for _, row := range rows {
		summary := summaryByID[row.RoomID]
		summary.Settings.Admins = append(summary.Settings.Admins, model.UserID(row.UserID))
	}
	return nil
}

func loadRoomParticipants(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	var rows []dbmodel.RoomParticipant
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("joined_at ASC").
		Order("user_id ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	for _, row := range rows {
		room := roomByID[row.RoomID]
		room.Participants = append(room.Participants, model.Participant{
			UserID:   model.UserID(row.UserID),
			JoinedAt: row.JoinedAt,
		})
	}
	return nil
}

func loadRoomSummaryParticipants(db *gorm.DB, roomIDs []string, summaryByID map[string]*model.RoomSummary) error {
	var rows []dbmodel.RoomParticipant
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("joined_at ASC").
		Order("user_id ASC").
		Find(&rows).
		Error; err != nil {
		return err
	}

	for _, row := range rows {
		summary := summaryByID[row.RoomID]
		summary.Participants = append(summary.Participants, model.Participant{
			UserID:   model.UserID(row.UserID),
			JoinedAt: row.JoinedAt,
		})
	}
	return nil
}

func loadRoomCards(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	var cardRows []dbmodel.RoomCard
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("created_at ASC").
		Order("owner_user_id ASC").
		Find(&cardRows).
		Error; err != nil {
		return err
	}
	if len(cardRows) == 0 {
		return nil
	}

	cardIDs := make([]string, 0, len(cardRows))
	cardsByID := make(map[string]*loadedCard, len(cardRows))
	for _, row := range cardRows {
		cardID, err := parseCardID(row.CardID)
		if err != nil {
			return err
		}
		cardIDs = append(cardIDs, row.CardID)
		cardsByID[row.CardID] = &loadedCard{
			roomID: row.RoomID,
			card: model.Card{
				CardID:      cardID,
				OwnerUserID: model.UserID(row.OwnerUserID),
				Cells:       [25]model.CardCell{},
			},
			cellIndices: map[uint8]struct{}{},
		}
	}

	var cellRows []dbmodel.RoomCardCell
	if err := db.Where("card_id IN ?", cardIDs).
		Order("card_id ASC").
		Order("cell_index ASC").
		Find(&cellRows).
		Error; err != nil {
		return err
	}
	for _, row := range cellRows {
		loaded, ok := cardsByID[row.CardID]
		if !ok {
			continue
		}
		if row.CellIndex >= 25 {
			return fmt.Errorf("%w: card %s has out-of-range cell index %d", ErrInvalidRoomAggregate, row.CardID, row.CellIndex)
		}
		loaded.card.Cells[row.CellIndex] = model.CardCell{
			Index:     model.CellIndex(row.CellIndex),
			Number:    uint8PtrToBallNumberPtr(row.Number),
			CellState: model.CardCellState(row.CellState),
		}
		loaded.cellIndices[row.CellIndex] = struct{}{}
	}

	for _, row := range cardRows {
		loaded := cardsByID[row.CardID]
		if len(loaded.cellIndices) != 25 {
			return fmt.Errorf("%w: card %s has %d cells", ErrInvalidRoomAggregate, row.CardID, len(loaded.cellIndices))
		}
		roomByID[loaded.roomID].Cards = append(roomByID[loaded.roomID].Cards, loaded.card)
	}
	return nil
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
			})
		}

		summary.BingoSummaries[index].BingoOrders = append(
			summary.BingoSummaries[index].BingoOrders,
			model.BingoOrder(row.BingoOrder),
		)
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
		roomByID[row.RoomID].Massages = append(roomByID[row.RoomID].Massages, model.Massage{
			MassageID: messageID,
			Content:   row.Content,
			Author:    model.UserID(row.AuthorUserID),
			CreatedAt: row.CreatedAt,
		})
	}
	return nil
}

type loadedCard struct {
	roomID      string
	card        model.Card
	cellIndices map[uint8]struct{}
}

func roomUserRows(room *model.Room) ([]dbmodel.RoomUser, error) {
	userIDs := make(map[model.UserID]struct{})
	addUserID := func(userID model.UserID) error {
		if err := validateUserID(userID); err != nil {
			return err
		}
		userIDs[userID] = struct{}{}
		return nil
	}
	for _, admin := range room.Settings.Admins {
		if err := addUserID(admin); err != nil {
			return nil, err
		}
	}
	for _, participant := range room.Participants {
		if err := addUserID(participant.UserID); err != nil {
			return nil, err
		}
	}
	for _, card := range room.Cards {
		if err := addUserID(card.OwnerUserID); err != nil {
			return nil, err
		}
	}
	for _, record := range room.BingoRecords {
		if err := addUserID(record.UserID); err != nil {
			return nil, err
		}
	}
	for _, record := range room.ReachRecords {
		if err := addUserID(record.UserID); err != nil {
			return nil, err
		}
	}
	for _, message := range room.Massages {
		if err := addUserID(message.Author); err != nil {
			return nil, err
		}
	}

	sortedUserIDs := make([]model.UserID, 0, len(userIDs))
	for userID := range userIDs {
		sortedUserIDs = append(sortedUserIDs, userID)
	}
	slices.SortFunc(sortedUserIDs, func(a, b model.UserID) int {
		return cmpString(string(a), string(b))
	})

	roomID := roomIDString(room.RoomID)
	timestamp := aggregateTimestamp(room)
	rows := make([]dbmodel.RoomUser, 0, len(sortedUserIDs))
	for _, userID := range sortedUserIDs {
		rows = append(rows, dbmodel.RoomUser{
			RoomID:    roomID,
			UserID:    string(userID),
			UserName:  string(userID),
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		})
	}
	return rows, nil
}

func createInBatches[T any](tx *gorm.DB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}
	return tx.CreateInBatches(&rows, createBatchSize).Error
}

func uniqueUserIDs(userIDs []model.UserID) []model.UserID {
	seen := make(map[model.UserID]struct{}, len(userIDs))
	uniqueUserIDs := make([]model.UserID, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		uniqueUserIDs = append(uniqueUserIDs, userID)
	}
	sort.SliceStable(uniqueUserIDs, func(i, j int) bool {
		return uniqueUserIDs[i] < uniqueUserIDs[j]
	})
	return uniqueUserIDs
}

func validateUserID(userID model.UserID) error {
	if userID == "" {
		return fmt.Errorf("%w: empty user id", ErrInvalidRoomAggregate)
	}
	return nil
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

func parseMessageID(value string) (model.MassageID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return model.MassageID{}, fmt.Errorf("%w: invalid message id %q: %v", ErrInvalidRoomAggregate, value, err)
	}
	return model.MassageID(id), nil
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

func messageIDString(messageID model.MassageID) string {
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

func cmpString(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
