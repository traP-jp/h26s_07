package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/traP-jp/h26_07/backend/internal/model"
)

func parseRoomIDParam(c *echo.Context) (model.RoomID, bool) {
	roomID, err := uuid.Parse(c.Param("roomId"))
	if err != nil {
		return model.RoomID{}, false
	}
	return model.RoomID(roomID), true
}
