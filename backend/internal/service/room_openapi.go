package service

import (
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func convertUserIDsToOpenAPI(userIDs []model.UserID) []openapi.User {
	var result []openapi.User
	for _, userID := range userIDs {
		result = append(result, openapi.User{UserID: openapi.UserID(userID)})
	}
	return result
}

func convertRoomSettingsToOpenAPI(settings model.RoomSettings) openapi.GameSettings {
	return openapi.GameSettings{
		Name:        settings.Name,
		Description: settings.Description,
		Admins:      convertUserIDsToOpenAPI(settings.Admins),
	}
}

func convertPickedBallsToOpenAPI(pickedBalls []model.BallNumber) []openapi.PickedBall {
	result := make([]openapi.PickedBall, 0, len(pickedBalls))
	for _, ball := range pickedBalls {
		result = append(result, openapi.PickedBall(ball))
	}
	return result
}

func convertBingoSummariesToOpenAPI(bingoSummaries []model.BingoSummary) []openapi.BingoSummary {
	result := make([]openapi.BingoSummary, 0, len(bingoSummaries))
	for _, bingoSummary := range bingoSummaries {
		bingoOrders := make([]int, 0, len(bingoSummary.BingoOrders))
		for _, bingoOrder := range bingoSummary.BingoOrders {
			bingoOrders = append(bingoOrders, int(bingoOrder))
		}
		result = append(result, openapi.BingoSummary{
			BingoOrders: bingoOrders,
			CreatedAt:   bingoSummary.CreatedAt,
			User:        openapi.User{UserID: openapi.UserID(bingoSummary.UserID)},
		})
	}
	return result
}

func convertReachSummariesToOpenAPI(reachSummaries []model.ReachSummary) []openapi.ReachSummary {
	result := make([]openapi.ReachSummary, 0, len(reachSummaries))
	for _, reachSummary := range reachSummaries {
		result = append(result, openapi.ReachSummary{
			CreatedAt: reachSummary.CreatedAt,
			User:      openapi.User{UserID: openapi.UserID(reachSummary.UserID)},
		})
	}
	return result
}

func convertBingoUpdatesToOpenAPI(updates []model.BingoUpdate) []openapi.BingoUpdate {
	result := make([]openapi.BingoUpdate, 0, len(updates))
	for _, update := range updates {
		result = append(result, openapi.BingoUpdate{
			User:           openapi.User{UserID: openapi.UserID(update.UserID)},
			NewBingoOrders: convertBingoOrdersToOpenAPI(update.NewBingoOrders),
			BingoOrders:    convertBingoOrdersToOpenAPI(update.BingoOrders),
		})
	}
	return result
}

func convertBingoOrdersToOpenAPI(orders []model.BingoOrder) []int {
	result := make([]int, 0, len(orders))
	for _, order := range orders {
		result = append(result, int(order))
	}
	return result
}

func convertReachUpdatesToOpenAPI(updates []model.ReachUpdate) []openapi.ReachUpdate {
	result := make([]openapi.ReachUpdate, 0, len(updates))
	for _, update := range updates {
		result = append(result, openapi.ReachUpdate{
			User: openapi.User{UserID: openapi.UserID(update.UserID)},
		})
	}
	return result
}

func convertCardToOpenAPI(room *model.Room, card model.Card) openapi.Card {
	cells := make([]openapi.CardCell, 0, len(card.Cells))
	for _, cell := range card.Cells {
		cells = append(cells, convertCardCellToOpenAPI(cell))
	}
	return openapi.Card{
		CardID:      uuid.UUID(card.CardID).String(),
		CardNumber:  string(card.CardNumber),
		OwnerUserID: openapi.UserID(card.OwnerUserID),
		Cells:       cells,
		BingoLines:  convertLineIndexesToOpenAPI(bingoLineIndexes(room, card.OwnerUserID)),
		ReachLines:  convertLineIndexesToOpenAPI(card.ReachLines(room.BingoRecords)),
	}
}

func convertCardCellToOpenAPI(cell model.CardCell) openapi.CardCell {
	var number *int
	if cell.Number != nil {
		value := int(*cell.Number)
		number = &value
	}

	displayText := "FREE"
	if number != nil {
		displayText = strconv.Itoa(*number)
	}

	return openapi.CardCell{
		Index:       int(cell.Index),
		Number:      number,
		DisplayText: displayText,
		CellState:   openapi.CardCellState(cell.CellState),
	}
}

func convertCardChangesToOpenAPI(changes model.CardChanges) openapi.CardChanges {
	openedCellIndices := make([]int, 0, len(changes.OpenedCellIndices))
	for _, index := range changes.OpenedCellIndices {
		openedCellIndices = append(openedCellIndices, int(index))
	}
	return openapi.CardChanges{
		OpenedCellIndices: openedCellIndices,
		NewReachLines:     convertLineIndexesToOpenAPI(changes.NewReachLines),
		NewBingoLines:     convertLineIndexesToOpenAPI(changes.NewBingoLines),
	}
}

func convertLineIndexesToOpenAPI(lines []model.LineIndex) []openapi.Line {
	result := make([]openapi.Line, 0, len(lines))
	for _, line := range lines {
		indices, ok := model.LineCells(line)
		if !ok {
			continue
		}
		openapiLine := make(openapi.Line, 0, len(indices))
		for _, index := range indices {
			openapiLine = append(openapiLine, int(index))
		}
		result = append(result, openapiLine)
	}
	return result
}

func bingoLineIndexes(room *model.Room, userID model.UserID) []model.LineIndex {
	lines := make([]model.LineIndex, 0)
	for _, record := range room.BingoRecords {
		if record.UserID == userID {
			lines = append(lines, record.Line)
		}
	}
	slices.Sort(lines)
	return lines
}
