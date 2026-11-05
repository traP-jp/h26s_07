package model

import "time"

type CardChanges struct {
	OpenedCellIndices []CellIndex
	NewReachLines     []LineIndex
	NewBingoLines     []LineIndex
}

type ParticipantCardUpdate struct {
	UserID      UserID
	Card        Card
	CardChanges CardChanges
}

type GameStartedResult struct {
	ParticipantCards []ParticipantCardUpdate
}

type GameFinishedResult struct {
	PickCanceled       bool
	ParticipantUpdates []ParticipantCardUpdate
}

func (room *Room) StartGame(actor UserID, cards []Card, now time.Time) (GameStartedResult, error) {
	if room.State != RoomStateWaiting || len(room.Participants) == 0 {
		return GameStartedResult{}, ErrRoomNotStartable
	}
	if !room.IsAdmin(actor) {
		return GameStartedResult{}, ErrRoomForbidden
	}
	if !cardsMatchParticipants(cards, room.Participants) {
		return GameStartedResult{}, ErrInvalidCard
	}

	room.State = RoomStatePlaying
	room.PickState = RoomPickStateIdle
	room.Cards = append([]Card(nil), cards...)
	room.UpdatedAt = now

	updates := make([]ParticipantCardUpdate, 0, len(cards))
	for _, card := range cards {
		updates = append(updates, ParticipantCardUpdate{
			UserID: card.OwnerUserID,
			Card:   card,
		})
	}
	return GameStartedResult{ParticipantCards: updates}, nil
}

func (room *Room) FinishGame(actor UserID, now time.Time) (GameFinishedResult, error) {
	if !room.IsAdmin(actor) {
		return GameFinishedResult{}, ErrRoomForbidden
	}
	if room.State != RoomStatePlaying {
		return GameFinishedResult{}, ErrRoomNotFinishable
	}

	// A game can end during a draw; callers need to notify clients to stop that pick.
	pickCanceled := room.PickState == RoomPickStatePicking
	room.State = RoomStateFinished
	room.PickState = RoomPickStateIdle
	room.UpdatedAt = now

	updates := make([]ParticipantCardUpdate, 0, len(room.Cards))
	for _, card := range room.Cards {
		updates = append(updates, ParticipantCardUpdate{
			UserID: card.OwnerUserID,
			Card:   card,
		})
	}
	return GameFinishedResult{
		PickCanceled:       pickCanceled,
		ParticipantUpdates: updates,
	}, nil
}

func cardsMatchParticipants(cards []Card, participants []Participant) bool {
	if len(cards) != len(participants) {
		return false
	}
	participantSet := make(map[UserID]struct{}, len(participants))
	for _, participant := range participants {
		participantSet[participant.UserID] = struct{}{}
	}
	for _, card := range cards {
		if _, ok := participantSet[card.OwnerUserID]; !ok {
			return false
		}
		delete(participantSet, card.OwnerUserID)
	}
	return len(participantSet) == 0
}
