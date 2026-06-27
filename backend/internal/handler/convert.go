package handler

import (
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

var bingoLineCellIndices = [12][5]model.CellIndex{
	{0, 1, 2, 3, 4},
	{5, 6, 7, 8, 9},
	{10, 11, 12, 13, 14},
	{15, 16, 17, 18, 19},
	{20, 21, 22, 23, 24},
	{0, 5, 10, 15, 20},
	{1, 6, 11, 16, 21},
	{2, 7, 12, 17, 22},
	{3, 8, 13, 18, 23},
	{4, 9, 14, 19, 24},
	{0, 6, 12, 18, 24},
	{4, 8, 12, 16, 20},
}

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
		ReachLines:  convertLineIndexesToOpenAPI(reachLineIndexes(card)),
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

func reachLineIndexes(card model.Card) []model.LineIndex {
	lines := make([]model.LineIndex, 0)
	for line := model.LineIndex(0); int(line) < len(bingoLineCellIndices); line++ {
		indices := bingoLineCellIndices[line]
		for _, index := range indices {
			if card.Cells[index].CellState == model.CardCellStateReach {
				lines = append(lines, line)
				break
			}
		}
	}
	return lines
}

func lineCells(line model.LineIndex) ([5]model.CellIndex, bool) {
	if int(line) < 0 || int(line) >= len(bingoLineCellIndices) {
		return [5]model.CellIndex{}, false
	}
	return bingoLineCellIndices[line], true
}
