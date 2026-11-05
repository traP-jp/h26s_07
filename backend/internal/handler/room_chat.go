package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func (h *RoomHandler) PostMessage(c *echo.Context) error {
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
	var createMessageRequest openapi.CreateMessageRequest
	if err := c.Bind(&createMessageRequest); err != nil {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "message invalid"})
	}
	content := createMessageRequest.Content
	message, err := h.roomService.PostMessage(c.Request().Context(), model.RoomID(roomID), user, content)
	if err != nil {
		return writeRoomServiceError(
			c,
			err,
			"internal server error",
			messageInvalidMapping,
			roomNotFoundMapping,
			messageNotAllowedMapping,
			messageNotPostableMapping,
		)
	}
	return c.JSON(http.StatusOK, convertMessageToOpenAPI(*message))
}

func (h *RoomHandler) GetMessages(c *echo.Context) error {
	userRaw, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}
	roomIDString := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDString)
	if err != nil {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "Invalid roomId"})
	}
	messages, err := h.roomService.GetMessages(c.Request().Context(), model.RoomID(roomID), model.UserID(userRaw.Name))
	if err != nil {
		return writeRoomServiceError(c, err, "internal server error", roomNotFoundMapping, roomMessagesForbiddenMapping)
	}
	result := make([]openapi.Message, 0, len(messages))
	for _, message := range messages {
		result = append(result, convertMessageToOpenAPI(message))
	}
	return c.JSON(http.StatusOK, result)
}
