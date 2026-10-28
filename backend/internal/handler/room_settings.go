package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func (h *RoomHandler) PutSettings(c *echo.Context) error {
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
	var settingsRaw openapi.UpdateGameSettingsRequest
	if err := c.Bind(&settingsRaw); err != nil {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "Invalid Settings"})
	}
	if adminUserIDsSpecifiedEmpty(settingsRaw.Settings) {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "adminUserIds must not be empty"})
	}
	settings, err := h.roomService.PutSettings(
		c.Request().Context(),
		model.RoomID(roomID),
		user,
		convertRoomSettingsUpdateToService(settingsRaw.Settings),
	)
	if err != nil {
		return writeRoomServiceError(
			c,
			err,
			"internal server error",
			roomNotFoundMapping,
			adminRequiredMapping,
			roomSettingsInvalidMapping,
			roomNotConfigurableMapping,
		)
	}
	return c.JSON(http.StatusOK, convertRoomSettingsToOpenAPI(settings))
}
