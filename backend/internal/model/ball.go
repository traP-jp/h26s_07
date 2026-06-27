package model

const (
	MinBallNumber BallNumber = 1
	MaxBallNumber BallNumber = 75
)

func (number BallNumber) Valid() bool {
	return number >= MinBallNumber && number <= MaxBallNumber
}

func AllBallNumbers() []BallNumber {
	numbers := make([]BallNumber, 0, MaxBallNumber)
	for number := MinBallNumber; number <= MaxBallNumber; number++ {
		numbers = append(numbers, number)
	}
	return numbers
}

func DrawableNumbers(pickedBalls []BallNumber) []BallNumber {
	picked := make(map[BallNumber]struct{}, len(pickedBalls))
	for _, number := range pickedBalls {
		if number.Valid() {
			picked[number] = struct{}{}
		}
	}

	result := make([]BallNumber, 0, int(MaxBallNumber)-len(picked))
	for number := MinBallNumber; number <= MaxBallNumber; number++ {
		if _, ok := picked[number]; !ok {
			result = append(result, number)
		}
	}
	return result
}
