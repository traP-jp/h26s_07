package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func (h *RoomHandler) PostPickStart(c *echo.Context) error {
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
	err = h.roomService.StartPick(c.Request().Context(), model.RoomID(roomID), user)

	if err != nil {
		return writeRoomServiceError(
			c, err, "internal server error", roomNotFoundMapping, adminRequiredMapping, pickNotStartableMapping,
		)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *RoomHandler) PostPickCancel(c *echo.Context) error {
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
	err = h.roomService.CancelPick(c.Request().Context(), model.RoomID(roomID), user)

	if err != nil {
		return writeRoomServiceError(
			c, err, "internal server error", roomNotFoundMapping, adminRequiredMapping, pickNotCancelableMapping,
		)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *RoomHandler) StartGame(c *echo.Context) error {
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
	err = h.roomService.StartGame(c.Request().Context(), model.RoomID(roomID), user)
	if err != nil {
		return writeRoomServiceError(
			c, err, "internal server error", roomNotFoundMapping, adminRequiredMapping, roomNotStartableMapping,
		)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *RoomHandler) FinishGame(c *echo.Context) error {
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
	err = h.roomService.FinishGame(c.Request().Context(), model.RoomID(roomID), user)
	if err != nil {
		return writeRoomServiceError(
			c, err, "internal server error", roomNotFoundMapping, adminRequiredMapping, roomNotFinishableMapping,
		)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *RoomHandler) PostPickFinish(c *echo.Context) error {
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
	err = h.roomService.FinishPick(c.Request().Context(), model.RoomID(roomID), user)

	if err != nil {
		return writeRoomServiceError(
			c, err, "internal server error", roomNotFoundMapping, adminRequiredMapping, pickNotFinishableMapping,
		)
	}
	return c.NoContent(http.StatusNoContent)
}
