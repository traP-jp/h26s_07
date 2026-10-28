package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

func (r *GormRoomRepository) replaceRoomCards(tx *gorm.DB, room *model.Room) error {
	roomID := roomIDString(room.RoomID)
	if err := deleteRoomCards(tx, roomID); err != nil {
		return err
	}
	if len(room.Cards) == 0 {
		return nil
	}

	cardRows, cellRows, err := roomCardRows(roomID, room)
	if err != nil {
		return err
	}
	if err := createInBatches(tx, cardRows); err != nil {
		return err
	}
	return createInBatches(tx, cellRows)
}

// deleteRoomCards removes child cells first to satisfy their foreign key.
func deleteRoomCards(tx *gorm.DB, roomID string) error {
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
	return tx.Where("room_id = ?", roomID).Delete(&dbmodel.RoomCard{}).Error
}

func roomCardRows(roomID string, room *model.Room) ([]dbmodel.RoomCard, []dbmodel.RoomCardCell, error) {
	cardRows := make([]dbmodel.RoomCard, 0, len(room.Cards))
	cellRows := make([]dbmodel.RoomCardCell, 0, len(room.Cards)*25)
	createdAt := aggregateTimestamp(room)
	seenCardIDs := make(map[string]struct{}, len(room.Cards))
	seenCardNumbers := make(map[model.CardNumber]struct{}, len(room.Cards))
	for _, card := range room.Cards {
		cardID := cardIDString(card.CardID)
		if _, ok := seenCardIDs[cardID]; ok {
			return nil, nil, fmt.Errorf("%w: duplicate card id %s", ErrInvalidRoomAggregate, cardID)
		}
		seenCardIDs[cardID] = struct{}{}
		if err := validateUserID(card.OwnerUserID); err != nil {
			return nil, nil, err
		}
		if err := validateCardNumber(card.CardNumber); err != nil {
			return nil, nil, err
		}
		if _, ok := seenCardNumbers[card.CardNumber]; ok {
			return nil, nil, fmt.Errorf("%w: duplicate card number %s", ErrInvalidRoomAggregate, card.CardNumber)
		}
		seenCardNumbers[card.CardNumber] = struct{}{}

		cardRows = append(cardRows, dbmodel.RoomCard{
			CardID:      cardID,
			RoomID:      roomID,
			CardNumber:  string(card.CardNumber),
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
	return cardRows, cellRows, nil
}

func loadRoomCards(db *gorm.DB, roomIDs []string, roomByID map[string]*model.Room) error {
	cardRows, err := roomCardRowsForRooms(db, roomIDs)
	if err != nil {
		return err
	}
	if len(cardRows) == 0 {
		return nil
	}

	cardIDs, cardsByID, err := loadedCardsFromRows(cardRows)
	if err != nil {
		return err
	}
	cellRows, err := roomCardCellRows(db, cardIDs)
	if err != nil {
		return err
	}
	if err := loadRoomCardCells(cellRows, cardsByID); err != nil {
		return err
	}
	return appendLoadedCards(cardRows, cardsByID, roomByID)
}

func roomCardRowsForRooms(db *gorm.DB, roomIDs []string) ([]dbmodel.RoomCard, error) {
	var rows []dbmodel.RoomCard
	// Owner ID breaks creation-time ties to keep the exposed card order stable.
	if err := db.Where("room_id IN ?", roomIDs).
		Order("room_id ASC").
		Order("created_at ASC").
		Order("owner_user_id ASC").
		Find(&rows).
		Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func loadedCardsFromRows(cardRows []dbmodel.RoomCard) ([]string, map[string]*loadedCard, error) {
	cardIDs := make([]string, 0, len(cardRows))
	cardsByID := make(map[string]*loadedCard, len(cardRows))
	for _, row := range cardRows {
		cardID, err := parseCardID(row.CardID)
		if err != nil {
			return nil, nil, err
		}
		cardIDs = append(cardIDs, row.CardID)
		cardsByID[row.CardID] = &loadedCard{
			roomID: row.RoomID,
			card: model.Card{
				CardID:      cardID,
				CardNumber:  model.CardNumber(row.CardNumber),
				OwnerUserID: model.UserID(row.OwnerUserID),
				Cells:       [25]model.CardCell{},
			},
			cellIndices: map[uint8]struct{}{},
		}
	}
	return cardIDs, cardsByID, nil
}

func roomCardCellRows(db *gorm.DB, cardIDs []string) ([]dbmodel.RoomCardCell, error) {
	var rows []dbmodel.RoomCardCell
	if err := db.Where("card_id IN ?", cardIDs).
		Order("card_id ASC").
		Order("cell_index ASC").
		Find(&rows).
		Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func loadRoomCardCells(cellRows []dbmodel.RoomCardCell, cardsByID map[string]*loadedCard) error {
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
	return nil
}

func appendLoadedCards(cardRows []dbmodel.RoomCard, cardsByID map[string]*loadedCard, roomByID map[string]*model.Room) error {
	for _, row := range cardRows {
		loaded := cardsByID[row.CardID]
		// Reject incomplete cards instead of silently filling missing cells with zero values.
		if len(loaded.cellIndices) != 25 {
			return fmt.Errorf("%w: card %s has %d cells", ErrInvalidRoomAggregate, row.CardID, len(loaded.cellIndices))
		}
		roomByID[loaded.roomID].Cards = append(roomByID[loaded.roomID].Cards, loaded.card)
	}
	return nil
}

type loadedCard struct {
	roomID      string
	card        model.Card
	cellIndices map[uint8]struct{}
}

func validateCardNumber(cardNumber model.CardNumber) error {
	if !cardNumber.Valid() {
		return fmt.Errorf("%w: invalid card number %q", ErrInvalidRoomAggregate, cardNumber)
	}
	return nil
}
