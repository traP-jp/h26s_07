package model

import "time"

func (settings RoomSettings) HasAdmin(userID UserID) bool {
	for _, admin := range settings.Admins {
		if admin == userID {
			return true
		}
	}
	return false
}

func (room *Room) IsAdmin(userID UserID) bool {
	return room.Settings.HasAdmin(userID)
}

func (room *Room) IsParticipant(userID UserID) bool {
	for _, participant := range room.Participants {
		if participant.UserID == userID {
			return true
		}
	}
	return false
}

func (room *Room) CanJoin() bool {
	return room.State == RoomStateWaiting
}

func (room *Room) CanStart(userID UserID) bool {
	return room.State == RoomStateWaiting && len(room.Participants) > 0 && room.IsAdmin(userID)
}

func (room *Room) CanFinish(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID)
}

// TODO あとで消して別々にチェックして別のエラーを返すようにする
func (room *Room) CanFinishPick(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID) && room.PickState == RoomPickStatePicking
}

func (room *Room) ParticipantCount() int {
	return len(room.Participants)
}

func (room *Room) Join(userID UserID, now time.Time) error {
	// A retried join stays successful after the room starts if the user already joined.
	if room.IsParticipant(userID) {
		return nil
	}
	if !room.CanJoin() {
		return ErrRoomNotJoinable
	}

	room.Participants = append(room.Participants, Participant{
		UserID:   userID,
		JoinedAt: now,
	})
	room.UpdatedAt = now
	return nil
}

func (room *Room) CanView(userID UserID) bool {
	return room.IsParticipant(userID) || room.IsAdmin(userID)
}

func (room *Room) CanViewChat(userID UserID) bool {
	return room.CanView(userID)
}

func (room *Room) CanUpdateSettings(userID UserID) bool {
	return room.State != RoomStateFinished && room.IsAdmin(userID)
}

func (room *Room) CanFinishGame(userID UserID) bool {
	return room.State == RoomStatePlaying && room.IsAdmin(userID)
}
