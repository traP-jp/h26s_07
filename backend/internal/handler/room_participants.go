package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func (h *RoomHandler) PostParticipant(c *echo.Context) error {
	userRaw, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}
	roomIDString := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDString)
	if err != nil {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "Invalid roomId"})
	}
	user := model.UserID(userRaw.Name)
	err = h.roomService.PostParticipants(c.Request().Context(), model.RoomID(roomID), user)
	if err != nil {
		return writeRoomServiceError(c, err, "internal server error", roomNotFoundMapping, roomNotJoinableMapping)
	}

	return c.NoContent(http.StatusNoContent)
}
