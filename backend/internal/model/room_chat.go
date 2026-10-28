package model

import (
	"time"
	"unicode/utf8"
)

func (room *Room) CanPostMessage(user UserID) bool {
	return (room.IsParticipant(user) || room.IsAdmin(user)) && room.State != RoomStateFinished
}

func (room *Room) PostMessage(user UserID, content string, now time.Time, messageID MessageID) (Message, error) {
	if !room.IsParticipant(user) && !room.IsAdmin(user) {
		return Message{}, ErrRoomMessageNotAllowed
	}
	if room.State == RoomStateFinished {
		return Message{}, ErrRoomMessageNotPostable
	}
	if content == "" || utf8.RuneCountInString(content) > 500 {
		return Message{}, ErrMessageInvalid
	}
	message := Message{
		MessageID: messageID,
		Content:   content,
		Author:    user,
		CreatedAt: now,
	}
	room.Messages = append(room.Messages, message)
	room.UpdatedAt = now
	return message, nil
}
