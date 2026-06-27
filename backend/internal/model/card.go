package model

const (
	CardCellCount = 25
	FreeCellIndex = 12
)

func NewCard(cardID CardID, owner UserID, numbers [24]BallNumber) (Card, error) {
	var cells [CardCellCount]CardCell
	numberIndex := 0
	seenByColumn := [5]map[BallNumber]struct{}{}
	for i := range seenByColumn {
		seenByColumn[i] = map[BallNumber]struct{}{}
	}

	for i := range cells {
		index := CellIndex(i)
		if index == FreeCellIndex {
			cells[i] = CardCell{
				Index:     index,
				Number:    nil,
				CellState: CardCellStateOpen,
			}
			continue
		}

		number := numbers[numberIndex]
		numberIndex++
		if !number.Valid() || !numberInColumn(index, number) {
			return Card{}, ErrInvalidCard
		}
		column := int(index) % 5
		if _, ok := seenByColumn[column][number]; ok {
			return Card{}, ErrInvalidCard
		}
		seenByColumn[column][number] = struct{}{}
		cells[i] = CardCell{
			Index:     index,
			Number:    &number,
			CellState: CardCellStateClosed,
		}
	}

	return Card{
		CardID:      cardID,
		OwnerUserID: owner,
		Cells:       cells,
	}, nil
}

func (card *Card) Cell(index CellIndex) (CardCell, bool) {
	if int(index) < 0 || int(index) >= len(card.Cells) {
		return CardCell{}, false
	}
	return card.Cells[index], true
}

func (card *Card) HasNumber(number BallNumber) bool {
	for _, cell := range card.Cells {
		if cell.Number != nil && *cell.Number == number {
			return true
		}
	}
	return false
}

func (card *Card) IsOpenLike(index CellIndex) bool {
	cell, ok := card.Cell(index)
	if !ok {
		return false
	}
	return cell.CellState == CardCellStateOpen || cell.CellState == CardCellStateBingo
}

func (card *Card) OpenNumber(number BallNumber) []CellIndex {
	opened := make([]CellIndex, 0)
	for i := range card.Cells {
		cell := &card.Cells[i]
		if cell.Number == nil || *cell.Number != number {
			continue
		}
		if cell.CellState == CardCellStateOpen || cell.CellState == CardCellStateBingo {
			continue
		}
		cell.CellState = CardCellStateOpen
		opened = append(opened, cell.Index)
	}
	return opened
}

func (card *Card) MarkBingoLines(lines []LineIndex) {
	for _, line := range lines {
		indices, ok := LineCells(line)
		if !ok {
			continue
		}
		for _, index := range indices {
			card.Cells[index].CellState = CardCellStateBingo
		}
	}
}

func (card *Card) MarkReachLines(lines []LineIndex) {
	for _, line := range lines {
		index, ok := card.LastMissingCellIndex(line)
		if !ok {
			continue
		}
		if card.Cells[index].CellState == CardCellStateClosed {
			card.Cells[index].CellState = CardCellStateReach
		}
	}
}

func (card *Card) NewBingoLines(records []BingoRecord) []LineIndex {
	lines := make([]LineIndex, 0)
	for _, line := range AllBingoLines() {
		if hasBingoRecord(records, card.OwnerUserID, line) {
			continue
		}
		if card.lineIsBingo(line) {
			lines = append(lines, line)
		}
	}
	return lines
}

func (card *Card) ReachLines(records []BingoRecord) []LineIndex {
	lines := make([]LineIndex, 0)
	for _, line := range AllBingoLines() {
		if hasBingoRecord(records, card.OwnerUserID, line) {
			continue
		}
		if card.lineIsReach(line) {
			lines = append(lines, line)
		}
	}
	return lines
}

func (card *Card) LastMissingCellIndex(line LineIndex) (CellIndex, bool) {
	indices, ok := LineCells(line)
	if !ok {
		return 0, false
	}

	var missing CellIndex
	missingCount := 0
	for _, index := range indices {
		if card.IsOpenLike(index) {
			continue
		}
		missing = index
		missingCount++
	}
	return missing, missingCount == 1
}

func (card *Card) NewlyOpenedCells(before Card) []CellIndex {
	opened := make([]CellIndex, 0)
	for i, cell := range card.Cells {
		beforeCell := before.Cells[i]
		if isOpenLikeState(beforeCell.CellState) {
			continue
		}
		if isOpenLikeState(cell.CellState) {
			opened = append(opened, cell.Index)
		}
	}
	return opened
}

func (card *Card) ChangesFrom(before Card, newReachLines []LineIndex, newBingoLines []LineIndex) CardChanges {
	return CardChanges{
		OpenedCellIndices: card.NewlyOpenedCells(before),
		NewReachLines:     append([]LineIndex(nil), newReachLines...),
		NewBingoLines:     append([]LineIndex(nil), newBingoLines...),
	}
}

func (card *Card) lineIsBingo(line LineIndex) bool {
	indices, ok := LineCells(line)
	if !ok {
		return false
	}
	for _, index := range indices {
		if !card.IsOpenLike(index) {
			return false
		}
	}
	return true
}

func (card *Card) lineIsReach(line LineIndex) bool {
	_, ok := card.LastMissingCellIndex(line)
	return ok
}

func isOpenLikeState(state CardCellState) bool {
	return state == CardCellStateOpen || state == CardCellStateBingo
}

func hasBingoRecord(records []BingoRecord, userID UserID, line LineIndex) bool {
	for _, record := range records {
		if record.UserID == userID && record.Line == line {
			return true
		}
	}
	return false
}

func numberInColumn(index CellIndex, number BallNumber) bool {
	column := int(index) % 5
	min := BallNumber(column*15 + 1)
	max := BallNumber((column + 1) * 15)
	return number >= min && number <= max
}
