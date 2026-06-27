package handler

import (
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func convertPickedBallsToOpenAPI(pickedBalls []model.BallNumber) []openapi.PickedBall {
	result := make([]openapi.PickedBall, 0, len(pickedBalls))
	for _, ball := range pickedBalls {
		result = append(result, openapi.PickedBall(ball))
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
		OwnerUserID: openapi.UserID(card.OwnerUserID),
		Cells:       cells,
		BingoLines:  convertLineIndexesToOpenAPI(bingoLineIndexes(room, card.OwnerUserID)),
		ReachLines:  convertLineIndexesToOpenAPI(card.ReachLines(room.BingoRecords)),
	}
}

func convertCardCellToOpenAPI(cell model.CardCell) openapi.CardCell {
	var number *int
	if cell.Number != nil {
		number = new(int(*cell.Number))
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

func convertLineIndexesToOpenAPI(lines []model.LineIndex) []openapi.Line {
	result := make([]openapi.Line, 0, len(lines))
	for _, line := range lines {
		if indices, ok := lineCells(line); ok {
			openapiLine := make(openapi.Line, 0, len(indices))
			for _, index := range indices {
				openapiLine = append(openapiLine, int(index))
			}
			result = append(result, openapiLine)
		}
	}
	return result
}

func findCard(room *model.Room, userID model.UserID) (model.Card, bool) {
	for _, card := range room.Cards {
		if card.OwnerUserID == userID {
			return card, true
		}
	}
	return model.Card{}, false
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

func lineCells(line model.LineIndex) ([5]model.CellIndex, bool) {
	return model.LineCells(line)
}
