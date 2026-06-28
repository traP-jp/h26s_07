package utils

import (
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func ConvertPickedBallsToOpenAPI(pickedBalls []model.BallNumber) []openapi.PickedBall {
	result := make([]openapi.PickedBall, 0, len(pickedBalls))
	for _, ball := range pickedBalls {
		result = append(result, openapi.PickedBall(ball))
	}
	return result
}

func ConvertCardToOpenAPI(room *model.Room, card model.Card) openapi.Card {
	cells := make([]openapi.CardCell, 0, len(card.Cells))
	for _, cell := range card.Cells {
		cells = append(cells, ConvertCardCellToOpenAPI(cell))
	}
	return openapi.Card{
		CardID:      uuid.UUID(card.CardID).String(),
		CardNumber:  string(card.CardNumber),
		OwnerUserID: openapi.UserID(card.OwnerUserID),
		Cells:       cells,
		BingoLines:  ConvertLineIndexesToOpenAPI(BingoLineIndexes(room, card.OwnerUserID)),
		ReachLines:  ConvertLineIndexesToOpenAPI(card.ReachLines(room.BingoRecords)),
	}
}

func ConvertCardCellToOpenAPI(cell model.CardCell) openapi.CardCell {
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

func ConvertLineIndexesToOpenAPI(lines []model.LineIndex) []openapi.Line {
	result := make([]openapi.Line, 0, len(lines))
	for _, line := range lines {
		if indices, ok := LineCells(line); ok {
			openapiLine := make(openapi.Line, 0, len(indices))
			for _, index := range indices {
				openapiLine = append(openapiLine, int(index))
			}
			result = append(result, openapiLine)
		}
	}
	return result
}

func FindCard(room *model.Room, userID model.UserID) (model.Card, bool) {
	for _, card := range room.Cards {
		if card.OwnerUserID == userID {
			return card, true
		}
	}
	return model.Card{}, false
}

func BingoLineIndexes(room *model.Room, userID model.UserID) []model.LineIndex {
	lines := make([]model.LineIndex, 0)
	for _, record := range room.BingoRecords {
		if record.UserID == userID {
			lines = append(lines, record.Line)
		}
	}
	slices.Sort(lines)
	return lines
}

func LineCells(line model.LineIndex) ([5]model.CellIndex, bool) {
	return model.LineCells(line)
}
