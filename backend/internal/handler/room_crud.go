package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func (h *RoomHandler) PostRoom(c *echo.Context) error {
	user, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}
	var req openapi.CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, openapi.Error{
			Message: "invalid request body",
		})
	}
	if adminUserIDsSpecifiedEmpty(req.Settings) {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "adminUserIds must not be empty"})
	}

	room, err := h.roomService.CreateRoom(
		c.Request().Context(),
		convertRoomSettingsToModel(req.Settings),
		model.UserID(user.Name),
	)
	if err != nil {
		return writeRoomServiceError(c, err, "failed to create room", roomSettingsInvalidMapping)
	}
	return c.JSON(http.StatusOK, convertRoom(room))
}

func (h *RoomHandler) GetRoom(c *echo.Context) error {
	userRaw, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}
	roomIDString := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDString)
	if err != nil {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "Invalid roomId"})
	}
	room, err := h.roomService.GetRoom(c.Request().Context(), model.RoomID(roomID))
	if err != nil {
		return writeRoomServiceError(c, err, "internal server error", repositoryRoomNotFoundMapping)
	}
	if !room.CanView(model.UserID(userRaw.Name)) {
		return c.JSON(http.StatusForbidden, openapi.Error{Message: "room view forbidden"})
	}
	return c.JSON(http.StatusOK, convertRoom(room))
}

func (h *RoomHandler) ListRooms(c *echo.Context) error {
	_, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}
	includeFinished := false
	if raw := c.QueryParam("includeFinished"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return c.JSON(http.StatusBadRequest, openapi.Error{Message: "invalid includeFinished"})
		}
		includeFinished = value
	}
	roomSummary, err := h.roomService.ListRooms(c.Request().Context(), includeFinished)
	if err != nil {
		return writeRoomServiceError(c, err, "Internal Server Error")
	}
	return c.JSON(http.StatusOK, convertRoomSummary(roomSummary))

}

func (h *RoomHandler) DeleteRoom(c *echo.Context) error {
	user, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}
	id, ok := parseRoomIDParam(c)
	if !ok {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "Invalid roomId"})
	}
	err := h.roomService.DeleteRoom(c.Request().Context(), id, model.UserID(user.Name))
	if err != nil {
		return writeRoomServiceError(c, err, "internal server error", roomNotFoundMapping, adminRequiredMapping)
	}
	return c.NoContent(http.StatusNoContent)
}
