package handler

import (
	"context"
	"errors"
	"sync"

	"github.com/coder/websocket"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

type WebSocketEventSender struct {
	hub *WebSocketHub
}

type WebSocketHub struct {
	mu      sync.RWMutex
	clients map[model.RoomID]map[*webSocketClient]struct{}
}

func NewWebSocketEventSender(hub *WebSocketHub) *WebSocketEventSender {
	return &WebSocketEventSender{hub: hub}
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients: make(map[model.RoomID]map[*webSocketClient]struct{}),
	}
}

func (s *WebSocketEventSender) SendRoom(_ context.Context, roomID model.RoomID, event any) error {
	return s.hub.sendRoom(roomID, event)
}

func (s *WebSocketEventSender) SendParticipant(
	_ context.Context,
	roomID model.RoomID,
	userID model.UserID,
	event any,
) error {
	return s.hub.sendParticipant(roomID, userID, event)
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

	var sendErr error
	for _, client := range clients {
		payload, ok, err := marshalWebSocketEventForMode(client.mode, event)
		if err != nil {
			sendErr = errors.Join(sendErr, err)
			continue
		}
		if !ok {
			continue
		}
		if !client.enqueue(payload) {
			client.closeWithStatus(websocket.StatusTryAgainLater, "client is too slow")
			sendErr = errors.Join(sendErr, errors.New("websocket client send buffer is full"))
		}
	}
	return sendErr
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

	var sendErr error
	for _, client := range clients {
		if !client.enqueue(payload) {
			client.closeWithStatus(websocket.StatusTryAgainLater, "client is too slow")
			sendErr = errors.Join(sendErr, errors.New("websocket client send buffer is full"))
		}
	}
	return sendErr
}
