package model

import "time"

func (room *Room) ShowQRCode(actor UserID, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	room.QrCodeVisible = true
	room.UpdatedAt = now
	return nil
}

func (room *Room) HideQRCode(actor UserID, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	room.QrCodeVisible = false
	room.UpdatedAt = now
	return nil
}
