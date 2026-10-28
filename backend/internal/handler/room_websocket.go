package handler

import (
	"errors"
	"net/http"

	"github.com/coder/websocket"
	"github.com/labstack/echo/v5"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
	"github.com/traP-jp/h26_07/backend/internal/service"
)

type RoomWebSocketHandler struct {
	roomService    *service.RoomService
	hub            *WebSocketHub
	originPatterns []string
}

type webSocketConnectRequest struct {
	userID model.UserID
	roomID model.RoomID
	mode   openapi.WebSocketMode
}

func NewRoomWebSocketHandler(
	roomService *service.RoomService,
	hub *WebSocketHub,
	originPatterns []string,
) *RoomWebSocketHandler {
	return &RoomWebSocketHandler{
		roomService:    roomService,
		hub:            hub,
		originPatterns: append([]string(nil), originPatterns...),
	}
}

func (h *RoomWebSocketHandler) Connect(c *echo.Context) error {
	request, err := h.validateConnect(c)
	if request == nil {
		return err
	}

	unlock := h.roomService.LockRoom(request.roomID)
	defer func() {
		if unlock != nil {
			unlock()
		}
	}()

	initialPayload, err := h.initialPayload(c, request)
	if initialPayload == nil {
		return err
	}
	conn, err := h.upgrade(c)
	if err != nil {
		return nil
	}

	client := newWebSocketClient(h.hub, request, conn)
	if !h.subscribeWithSnapshot(client, initialPayload) {
		return nil
	}

	// Keep the room lock through registration so no update committed by this instance can land between
	// the snapshot and subscription. The lock is in-memory, so it does not cover other server instances.
	unlock()
	unlock = nil

	go client.writeLoop()
	go client.pingLoop()
	<-client.ctx.Done()
	client.close()
	return nil
}

func (h *RoomWebSocketHandler) validateConnect(c *echo.Context) (*webSocketConnectRequest, error) {
	user, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return nil, c.NoContent(http.StatusUnauthorized)
	}
	roomID, ok := parseRoomIDParam(c)
	if !ok {
		return nil, c.JSON(http.StatusBadRequest, openapi.Error{Message: "Invalid roomId"})
	}
	mode := openapi.WebSocketMode(c.QueryParam("mode"))
	if !mode.Valid() {
		return nil, c.JSON(http.StatusBadRequest, openapi.Error{Message: "invalid websocket mode"})
	}
	return &webSocketConnectRequest{
		userID: model.UserID(user.Name),
		roomID: roomID,
		mode:   mode,
	}, nil
}

func (h *RoomWebSocketHandler) initialPayload(c *echo.Context, request *webSocketConnectRequest) ([]byte, error) {
	room, err := h.roomService.GetRoom(c.Request().Context(), request.roomID)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return nil, c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		}
		return nil, c.JSON(http.StatusInternalServerError, openapi.Error{Message: "Internal Server Error"})
	}
	if request.mode == openapi.Participant && !room.IsParticipant(request.userID) {
		return nil, c.JSON(http.StatusForbidden, openapi.Error{Message: "participant websocket requires room participant"})
	}

	initialEvent, ok, err := buildInitializedEvent(room, request.mode, request.userID)
	if err != nil {
		return nil, c.JSON(http.StatusInternalServerError, openapi.Error{Message: "Internal Server Error"})
	}
	if !ok {
		return nil, c.JSON(http.StatusBadRequest, openapi.Error{Message: "invalid websocket mode"})
	}
	initialPayload, err := marshalWebSocketEvent(request.mode, initialEvent)
	if err != nil {
		return nil, c.JSON(http.StatusInternalServerError, openapi.Error{Message: "Internal Server Error"})
	}
	return initialPayload, nil
}

func (h *RoomWebSocketHandler) upgrade(c *echo.Context) (*websocket.Conn, error) {
	return websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
		OriginPatterns: h.originPatterns,
	})
}

func (h *RoomWebSocketHandler) subscribeWithSnapshot(client *webSocketClient, initialPayload []byte) bool {
	if !client.enqueue(initialPayload) {
		client.close()
		return false
	}
	h.hub.register(client)
	return true
}
