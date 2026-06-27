package model

import (
	"errors"
)

var (
	ErrRoomNotJoinable       = errors.New("room is not joinable")
	ErrRoomMessageNotAllowed = errors.New("room message is not allowed")
	ErrMessageInvalid        = errors.New("message is invalid")
	ErrRoomSettingsInvalid   = errors.New("room settings are invalid")
	ErrRoomNotConfigurable   = errors.New("room is not configurable")
	ErrRoomNotStartable      = errors.New("room is not startable")
	ErrRoomNotFinishable     = errors.New("room is not finishable")
	ErrRoomPickNotStartable  = errors.New("room pick is not startable")
	ErrRoomPickNotCancelable = errors.New("room pick is not cancelable")
	ErrRoomPickNotFinishable = errors.New("room pick is not finishable")
	ErrNoDrawableBalls       = errors.New("no drawable balls")
	ErrBallAlreadyPicked     = errors.New("ball already picked")
	ErrInvalidBallNumber     = errors.New("invalid ball number")
	ErrInvalidCard           = errors.New("invalid card")
	ErrInvalidLine           = errors.New("invalid line")
	ErrRecordIDRequired      = errors.New("record id is required")
	ErrRoomNotFound          = errors.New("room is not found")
	ErrNotForbidden          = errors.New("room settings change forbidden")
)
