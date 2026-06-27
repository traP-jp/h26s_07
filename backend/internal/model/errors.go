package model

import (
	"errors"
)

var (
	ErrRoomNotJoinable       = errors.New("room is not joinable")
	ErrRoomMassageNotAllowed = errors.New("room massage is not allowed")
	ErrMassageInvalid        = errors.New("massage is invalid")
)
