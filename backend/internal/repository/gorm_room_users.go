package repository

import (
	"fmt"
	"slices"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

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

func loadRoomAdmins(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	rows, err := roomAdminRows(db, roomIDs)
	if err != nil {
		return err
	}

	for _, row := range rows {
		room := roomByID[row.RoomID]
		room.Settings.Admins = append(room.Settings.Admins, model.UserID(row.UserID))
	}
	return nil
}

func loadRoomSummaryAdmins(db *gorm.DB, roomIDs []string, summaryByID map[string]*model.RoomSummary) error {
	rows, err := roomAdminRows(db, roomIDs)
	if err != nil {
		return err
	}

	for _, row := range rows {
		summary := summaryByID[row.RoomID]
		summary.Settings.Admins = append(summary.Settings.Admins, model.UserID(row.UserID))
	}
	return nil
}

func roomAdminRows(db *gorm.DB, roomIDs []string) ([]dbmodel.RoomAdmin, error) {
	var rows []dbmodel.RoomAdmin
	// User ID breaks equal-time ties so admin order remains deterministic.
	err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("added_at ASC").
		Order("user_id ASC").
		Find(&rows).
		Error
	return rows, err
}

func loadRoomParticipants(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	rows, err := roomParticipantRows(db, roomIDs)
	if err != nil {
		return err
	}

	for _, row := range rows {
		room := roomByID[row.RoomID]
		room.Participants = append(room.Participants, participantFromRow(row))
	}
	return nil
}

func loadRoomSummaryParticipants(db *gorm.DB, roomIDs []string, summaryByID map[string]*model.RoomSummary) error {
	rows, err := roomParticipantRows(db, roomIDs)
	if err != nil {
		return err
	}

	for _, row := range rows {
		summary := summaryByID[row.RoomID]
		summary.Participants = append(summary.Participants, participantFromRow(row))
	}
	return nil
}

func roomParticipantRows(db *gorm.DB, roomIDs []string) ([]dbmodel.RoomParticipant, error) {
	var rows []dbmodel.RoomParticipant
	// User ID breaks equal-time ties so participant order remains deterministic.
	err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("joined_at ASC").
		Order("user_id ASC").
		Find(&rows).
		Error
	return rows, err
}

func participantFromRow(row dbmodel.RoomParticipant) model.Participant {
	return model.Participant{
		UserID:   model.UserID(row.UserID),
		JoinedAt: row.JoinedAt,
	}
}

func roomUserRows(room *model.Room) ([]dbmodel.RoomUser, error) {
	sortedUserIDs, err := roomUserIDs(room)
	if err != nil {
		return nil, err
	}

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

func roomUserIDs(room *model.Room) ([]model.UserID, error) {
	userIDs := make([]model.UserID, 0)
	userIDs = append(userIDs, room.Settings.Admins...)
	for _, participant := range room.Participants {
		userIDs = append(userIDs, participant.UserID)
	}
	for _, card := range room.Cards {
		userIDs = append(userIDs, card.OwnerUserID)
	}
	for _, record := range room.BingoRecords {
		userIDs = append(userIDs, record.UserID)
	}
	for _, record := range room.ReachRecords {
		userIDs = append(userIDs, record.UserID)
	}
	for _, message := range room.Messages {
		userIDs = append(userIDs, message.Author)
	}
	return uniqueValidatedRoomUserIDs(userIDs)
}

func uniqueValidatedRoomUserIDs(userIDs []model.UserID) ([]model.UserID, error) {
	seen := make(map[model.UserID]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if err := validateUserID(userID); err != nil {
			return nil, err
		}
		seen[userID] = struct{}{}
	}

	// Sorting makes batch insertion deterministic despite map iteration order.
	sortedUserIDs := make([]model.UserID, 0, len(seen))
	for userID := range seen {
		sortedUserIDs = append(sortedUserIDs, userID)
	}
	slices.SortFunc(sortedUserIDs, func(a, b model.UserID) int {
		return cmpString(string(a), string(b))
	})
	return sortedUserIDs, nil
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
