package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
	"github.com/traP-jp/h26_07/backend/internal/service"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}
func convertUserIDsToModel(userIDS []openapi.UserID) []model.UserID {
	var result []model.UserID
	for _, userID := range userIDS {
		result = append(result, model.UserID(userID))
	}
	return result
}

func convertOptionalUserIDsToModel(userIDs *[]openapi.UserID) []model.UserID {
	if userIDs == nil {
		return []model.UserID{}
	}
	return convertUserIDsToModel(*userIDs)
}

func convertOptionalUserIDsToModelPointer(userIDs *[]openapi.UserID) *[]model.UserID {
	if userIDs == nil {
		return nil
	}
	converted := convertUserIDsToModel(*userIDs)
	return &converted
}

func convertUserIDsToOpenAPI(userIDs []model.UserID) []openapi.User {
	var result []openapi.User
	for _, userID := range userIDs {
		result = append(result, openapi.User{UserID: openapi.UserID(userID)})
	}
	return result
}

func convertRoomSettingsToModel(settings openapi.GameSettingsInput) model.RoomSettings {
	return model.RoomSettings{
		Name:        settings.Name,
		Description: settings.Description,
		Admins:      convertOptionalUserIDsToModel(settings.AdminUserIds),
	}
}

func convertRoomSettingsUpdateToService(settings openapi.GameSettingsInput) service.RoomSettingsUpdate {
	return service.RoomSettingsUpdate{
		Name:        settings.Name,
		Description: settings.Description,
		Admins:      convertOptionalUserIDsToModelPointer(settings.AdminUserIds),
	}
}

func adminUserIDsSpecifiedEmpty(settings openapi.GameSettingsInput) bool {
	return settings.AdminUserIds != nil && len(*settings.AdminUserIds) == 0
}

func isRoomNotFoundError(err error) bool {
	return errors.Is(err, repository.ErrRoomNotFound) || errors.Is(err, model.ErrRoomNotFound)
}

func convertRoomSettingsToOpenAPI(settings model.RoomSettings) openapi.GameSettings {
	return openapi.GameSettings{
		Name:        settings.Name,
		Description: settings.Description,
		Admins:      convertUserIDsToOpenAPI(settings.Admins),
	}
}

func convertParticipantsToOpenAPI(participants []model.Participant) []openapi.ParticipantSummary {
	result := make([]openapi.ParticipantSummary, 0, len(participants))
	for _, participant := range participants {
		result = append(result, openapi.ParticipantSummary{
			User:     openapi.User{UserID: openapi.UserID(participant.UserID)},
			JoinedAt: participant.JoinedAt,
		})
	}
	return result
}

func convertBingoSummariesToOpenAPI(bingoSummaries []model.BingoSummary) []openapi.BingoSummary {
	result := make([]openapi.BingoSummary, 0, len(bingoSummaries))
	for _, bingoSummary := range bingoSummaries {
		bingoOrders := make([]int, 0, len(bingoSummary.BingoOrders))
		for _, bingoOrder := range bingoSummary.BingoOrders {
			bingoOrders = append(bingoOrders, int(bingoOrder))
		}
		result = append(result, openapi.BingoSummary{
			BingoOrders: bingoOrders,
			User:        openapi.User{UserID: openapi.UserID(bingoSummary.UserID)},
		})
	}
	return result
}

func convertReachSummariesToOpenAPI(reachSummaries []model.ReachSummary) []openapi.ReachSummary {
	result := make([]openapi.ReachSummary, 0, len(reachSummaries))
	for _, reachSummary := range reachSummaries {
		result = append(result, openapi.ReachSummary{
			User: openapi.User{UserID: openapi.UserID(reachSummary.UserID)},
		})
	}
	return result
}

func convertRoom(room *model.Room) openapi.Room {
	return openapi.Room{
		RoomID:         openapi.RoomID(room.RoomID.String()),
		Settings:       convertRoomSettingsToOpenAPI(room.Settings),
		State:          openapi.RoomState(room.State),
		PickState:      openapi.PickState(room.PickState),
		QrCodeVisible:  room.QrCodeVisible,
		Participants:   convertParticipantsToOpenAPI(room.Participants),
		BingoSummaries: convertBingoSummariesToOpenAPI(room.BingoSummaries()),
		ReachSummaries: convertReachSummariesToOpenAPI(room.ReachSummaries()),
		CreatedAt:      room.CreatedAt,
		UpdatedAt:      room.UpdatedAt,
		RoomCode:       openapi.RoomCode(room.RoomCode),
	}
}

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

	room, err := h.roomService.CreateRoom(c.Request().Context(), convertRoomSettingsToModel(req.Settings), model.UserID(user.Name))
	if err != nil {
		if errors.Is(err, model.ErrRoomSettingsInvalid) {
			return c.JSON(http.StatusBadRequest, openapi.Error{Message: "invalid room settings"})
		}
		return c.JSON(
			http.StatusInternalServerError,
			openapi.Error{
				Message: "failed to create room",
			},
		)
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
		switch {
		case errors.Is(err, repository.ErrRoomNotFound):
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		default:
			return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
		}
	}
	if !room.CanView(model.UserID(userRaw.Name)) {
		return c.JSON(http.StatusForbidden, openapi.Error{Message: "room view forbidden"})
	}
	return c.JSON(http.StatusOK, convertRoom(room))
}

func convertRoomSummary(rooms []model.RoomSummary) []openapi.Room {
	result := make([]openapi.Room, 0, len(rooms))
	for _, room := range rooms {
		result = append(result,
			openapi.Room{
				RoomID:         openapi.RoomID(room.RoomID.String()),
				Settings:       convertRoomSettingsToOpenAPI(room.Settings),
				State:          openapi.RoomState(room.State),
				PickState:      openapi.PickState(room.PickState),
				QrCodeVisible:  room.QrCodeVisible,
				Participants:   convertParticipantsToOpenAPI(room.Participants),
				BingoSummaries: convertBingoSummariesToOpenAPI(room.BingoSummaries),
				ReachSummaries: convertReachSummariesToOpenAPI(room.ReachSummaries),
				CreatedAt:      room.CreatedAt,
				UpdatedAt:      room.UpdatedAt,
				RoomCode:       openapi.RoomCode(room.RoomCode),
			},
		)
	}
	return result
}

func (h *RoomHandler) ListRooms(c *echo.Context) error {
	_, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}
	roomSummary, err := h.roomService.ListRooms(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "Internal Server Error"})
	}
	return c.JSON(http.StatusOK, convertRoomSummary(roomSummary))

}

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
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		}
		if errors.Is(err, model.ErrRoomNotJoinable) {
			return c.JSON(http.StatusConflict, openapi.Error{Message: "room is not joinable"})
		}
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
	}

	return c.NoContent(http.StatusNoContent)
}

func convertMessageToOpenAPI(message model.Message) openapi.Message {
	return openapi.Message{
		MessageID: uuid.UUID(message.MessageID).String(),
		Content:   message.Content,
		Author: openapi.User{
			UserID: openapi.UserID(message.Author),
		},
		CreatedAt: message.CreatedAt,
	}
}

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
		if errors.Is(err, model.ErrMessageInvalid) {
			return c.JSON(http.StatusBadRequest, openapi.Error{Message: "message invalid"})
		} else if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		} else if errors.Is(err, model.ErrRoomMessageNotAllowed) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "message not allowed"})
		} else if errors.Is(err, model.ErrRoomMessageNotPostable) {
			return c.JSON(http.StatusConflict, openapi.Error{Message: "message not postable"})
		} else {
			return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
		}
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
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		} else if errors.Is(err, model.ErrRoomForbidden) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "room message not allowed"})
		}
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
	}
	result := make([]openapi.Message, 0, len(messages))
	for _, message := range messages {
		result = append(result, convertMessageToOpenAPI(message))
	}
	return c.JSON(http.StatusOK, result)
}

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
	settings, err := h.roomService.PutSettings(c.Request().Context(), model.RoomID(roomID), user, convertRoomSettingsUpdateToService(settingsRaw.Settings))
	if err != nil {
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		} else if errors.Is(err, model.ErrRoomForbidden) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "admin required"})
		} else if errors.Is(err, model.ErrRoomSettingsInvalid) {
			return c.JSON(http.StatusBadRequest, openapi.Error{Message: "invalid room settings"})
		} else if errors.Is(err, model.ErrRoomNotConfigurable) {
			return c.JSON(http.StatusConflict, openapi.Error{Message: "room is not configurable"})
		}
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
	}
	return c.JSON(http.StatusOK, convertRoomSettingsToOpenAPI(settings))
}

func (h *RoomHandler) ShowQRCode(c *echo.Context) error {
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
	err = h.roomService.ShowQRCode(c.Request().Context(), model.RoomID(roomID), user)
	if err != nil {
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		} else if errors.Is(err, model.ErrRoomForbidden) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "admin required"})
		}
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *RoomHandler) HideQRCode(c *echo.Context) error {
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
	err = h.roomService.HideQRCode(c.Request().Context(), model.RoomID(roomID), user)
	if err != nil {
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		} else if errors.Is(err, model.ErrRoomForbidden) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "admin required"})
		}
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
	}
	return c.NoContent(http.StatusNoContent)
}

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
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		}
		if errors.Is(err, model.ErrRoomForbidden) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "admin required"})
		}
		if errors.Is(err, model.ErrRoomPickNotStartable) || errors.Is(err, model.ErrNoDrawableBalls) {
			return c.JSON(http.StatusConflict, openapi.Error{Message: "pick is not startable"})
		}

		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
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
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		}
		if errors.Is(err, model.ErrRoomForbidden) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "admin required"})
		}
		if errors.Is(err, model.ErrRoomPickNotCancelable) {
			return c.JSON(http.StatusConflict, openapi.Error{Message: "pick is not cancelable"})
		}

		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
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
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		} else if errors.Is(err, model.ErrRoomForbidden) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "admin required"})
		} else if errors.Is(err, model.ErrRoomNotStartable) {
			return c.JSON(http.StatusConflict, openapi.Error{Message: "room not startable"})
		}
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
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
		if isRoomNotFoundError(err) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		} else if errors.Is(err, model.ErrRoomForbidden) {
			return c.JSON(http.StatusForbidden, openapi.Error{Message: "admin required"})
		} else if errors.Is(err, model.ErrRoomNotFinishable) {
			return c.JSON(http.StatusConflict, openapi.Error{Message: "room not finishable"})
		}
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "internal server error"})
	}
	return c.NoContent(http.StatusNoContent)
}
