package model

import "time"

func (settings RoomSettings) Valid() bool {
	return len(settings.Admins) > 0
}

func (room *Room) UpdateSettings(actor UserID, settings RoomSettings, now time.Time) error {
	if !room.IsAdmin(actor) {
		return ErrRoomForbidden
	}
	if room.State == RoomStateFinished {
		return ErrRoomNotConfigurable
	}
	if !settings.Valid() {
		return ErrRoomSettingsInvalid
	}
	room.Settings = settings
	room.UpdatedAt = now
	return nil
}
