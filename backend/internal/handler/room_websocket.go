package handler

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/traP-jp/h26_07/backend/internal/utils"

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

func NewRoomWebSocketHandler(roomService *service.RoomService, hub *WebSocketHub, originPatterns []string) *RoomWebSocketHandler {
	return &RoomWebSocketHandler{
		roomService:    roomService,
		hub:            hub,
		originPatterns: append([]string(nil), originPatterns...),
	}
}

func (h *RoomWebSocketHandler) Connect(c *echo.Context) error {
	user, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}
	roomID, ok := parseRoomIDParam(c)
	if !ok {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "Invalid roomId"})
	}
	mode := openapi.WebSocketMode(c.QueryParam("mode"))
	if !mode.Valid() {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "invalid websocket mode"})
	}

	room, err := h.roomService.GetRoom(c.Request().Context(), roomID)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return c.JSON(http.StatusNotFound, openapi.Error{Message: "room not found"})
		}
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "Internal Server Error"})
	}

	userID := model.UserID(user.Name)
	if mode == openapi.Participant && !room.IsParticipant(userID) {
		return c.JSON(http.StatusForbidden, openapi.Error{Message: "participant websocket requires room participant"})
	}

	initialEvent, ok, err := buildInitializedEvent(room, mode, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "Internal Server Error"})
	}
	if !ok {
		return c.JSON(http.StatusBadRequest, openapi.Error{Message: "invalid websocket mode"})
	}
	initialPayload, err := marshalWebSocketEvent(mode, initialEvent)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, openapi.Error{Message: "Internal Server Error"})
	}

	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
		OriginPatterns: h.originPatterns,
	})
	if err != nil {
		return nil
	}
	ctx := conn.CloseRead(context.Background())

	client := &webSocketClient{
		hub:    h.hub,
		roomID: roomID,
		userID: userID,
		mode:   mode,
		conn:   conn,
		ctx:    ctx,
		send:   make(chan []byte, 16),
	}
	if !client.enqueue(initialPayload) {
		client.close()
		return nil
	}
	h.hub.register(client)
	go client.writeLoop()

	<-ctx.Done()
	client.close()
	return nil
}

func parseRoomIDParam(c *echo.Context) (model.RoomID, bool) {
	roomID, err := uuid.Parse(c.Param("roomId"))
	if err != nil {
		return model.RoomID{}, false
	}
	return model.RoomID(roomID), true
}

type WebSocketEventSender struct {
	hub *WebSocketHub
}

func NewWebSocketEventSender(hub *WebSocketHub) *WebSocketEventSender {
	return &WebSocketEventSender{hub: hub}
}

func (s *WebSocketEventSender) SendRoom(_ context.Context, roomID model.RoomID, event any) error {
	return s.hub.sendRoom(roomID, event)
}

func (s *WebSocketEventSender) SendParticipant(_ context.Context, roomID model.RoomID, userID model.UserID, event any) error {
	return s.hub.sendParticipant(roomID, userID, event)
}

type WebSocketHub struct {
	mu      sync.RWMutex
	clients map[model.RoomID]map[*webSocketClient]struct{}
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients: make(map[model.RoomID]map[*webSocketClient]struct{}),
	}
}

func (h *WebSocketHub) register(client *webSocketClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.roomID]; !ok {
		h.clients[client.roomID] = make(map[*webSocketClient]struct{})
	}
	h.clients[client.roomID][client] = struct{}{}
}

func (h *WebSocketHub) unregister(client *webSocketClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.clients[client.roomID]
	if clients == nil {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.clients, client.roomID)
	}
}

func (h *WebSocketHub) sendRoom(roomID model.RoomID, event any) error {
	h.mu.RLock()
	clientsByRoom := h.clients[roomID]
	clients := make([]*webSocketClient, 0, len(clientsByRoom))
	for client := range clientsByRoom {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		payload, ok, err := marshalWebSocketEventForMode(client.mode, event)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if !client.enqueue(payload) {
			client.close()
		}
	}
	return nil
}

func (h *WebSocketHub) sendParticipant(roomID model.RoomID, userID model.UserID, event any) error {
	payload, err := marshalWebSocketEvent(openapi.Participant, event)
	if err != nil {
		return err
	}

	h.mu.RLock()
	clientsByRoom := h.clients[roomID]
	clients := make([]*webSocketClient, 0, len(clientsByRoom))
	for client := range clientsByRoom {
		if client.mode == openapi.Participant && client.userID == userID {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if !client.enqueue(payload) {
			client.close()
		}
	}
	return nil
}

type webSocketClient struct {
	hub    *WebSocketHub
	roomID model.RoomID
	userID model.UserID
	mode   openapi.WebSocketMode
	conn   *websocket.Conn
	ctx    context.Context
	send   chan []byte

	mu        sync.Mutex
	closed    bool
	closeOnce sync.Once
}

func (c *webSocketClient) enqueue(payload []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return false
	}
	select {
	case c.send <- payload:
		return true
	default:
		return false
	}
}

func (c *webSocketClient) writeLoop() {
	for payload := range c.send {
		ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
		err := c.conn.Write(ctx, websocket.MessageText, payload)
		cancel()
		if err != nil {
			break
		}
	}
	c.close()
}

func (c *webSocketClient) close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		close(c.send)
		c.mu.Unlock()

		c.hub.unregister(c)
		_ = c.conn.CloseNow()
	})
}

func buildInitializedEvent(room *model.Room, mode openapi.WebSocketMode, userID model.UserID) (any, bool, error) {
	switch mode {
	case openapi.Participant:
		body := openapi.ParticipantInitializedBody{
			State:          openapi.RoomState(room.State),
			Settings:       convertRoomSettingsToOpenAPI(room.Settings),
			PickState:      openapi.PickState(room.PickState),
			PickedBalls:    utils.ConvertPickedBallsToOpenAPI(room.PickedBalls),
			BingoSummaries: convertBingoSummariesToOpenAPI(room.BingoSummaries()),
			ReachSummaries: convertReachSummariesToOpenAPI(room.ReachSummaries()),
		}
		if room.State != model.RoomStateWaiting {
			if card, ok := utils.FindCard(room, userID); ok {
				body.Card = new(utils.ConvertCardToOpenAPI(room, card))
			}
		}
		return openapi.ParticipantInitializedEvent{
			Type: openapi.ParticipantInitializedEventTypeInitialized,
			Body: body,
		}, true, nil
	case openapi.Display:
		return openapi.DisplayInitializedEvent{
			Type: openapi.DisplayInitializedEventTypeInitialized,
			Body: openapi.DisplayInitializedBody{
				State:            openapi.RoomState(room.State),
				Settings:         convertRoomSettingsToOpenAPI(room.Settings),
				PickState:        openapi.PickState(room.PickState),
				ParticipantCount: room.ParticipantCount(),
				PickedBalls:      utils.ConvertPickedBallsToOpenAPI(room.PickedBalls),
				QrCodeVisible:    room.QrCodeVisible,
				BingoSummaries:   convertBingoSummariesToOpenAPI(room.BingoSummaries()),
				ReachSummaries:   convertReachSummariesToOpenAPI(room.ReachSummaries()),
			},
		}, true, nil
	default:
		return nil, false, nil
	}
}
