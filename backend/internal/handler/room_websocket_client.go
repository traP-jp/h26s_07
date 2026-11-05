package handler

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

type webSocketClient struct {
	hub    *WebSocketHub
	roomID model.RoomID
	userID model.UserID
	mode   openapi.WebSocketMode
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
	send   chan []byte

	mu        sync.Mutex
	closed    bool
	closeOnce sync.Once
}

func newWebSocketClient(hub *WebSocketHub, request *webSocketConnectRequest, conn *websocket.Conn) *webSocketClient {
	// CloseRead reads in the background so pongs and the peer's close frame are processed,
	// and it closes the connection with StatusPolicyViolation if a data message arrives.
	// Its context is not canceled by our own close, so the close handshake can finish;
	// the child context is canceled when writeLoop exits, which also stops pingLoop.
	ctx, cancel := context.WithCancel(conn.CloseRead(context.Background()))
	return &webSocketClient{
		hub:    hub,
		roomID: request.roomID,
		userID: request.userID,
		mode:   request.mode,
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
		send:   make(chan []byte, 64),
	}
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
	// Stop the ping loop and release Connect only after the writer has drained the queue,
	// so events queued before close are still attempted, as before the ping loop was split out.
	defer c.cancel()
	defer c.close()
	for {
		select {
		case <-c.ctx.Done():
			return
		case payload, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.write(payload); err != nil {
				return
			}
			if isRoomDeletedPayload(payload) {
				return
			}
		}
	}
}

func (c *webSocketClient) write(payload []byte) error {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()
	return c.conn.Write(ctx, websocket.MessageText, payload)
}

func (c *webSocketClient) pingLoop() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	defer c.close()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
			err := c.conn.Ping(ctx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func isRoomDeletedPayload(payload []byte) bool {
	var envelope struct {
		Type string `json:"type"`
	}
	return json.Unmarshal(payload, &envelope) == nil && envelope.Type == "RoomDeleted"
}

func (c *webSocketClient) close() {
	c.closeWithStatus(websocket.StatusNormalClosure, "connection closed")
}

func (c *webSocketClient) closeWithStatus(code websocket.StatusCode, reason string) {
	c.closeOnce.Do(func() {
		// Serialize queue closure with enqueue so broadcasts cannot send to a closed channel.
		c.mu.Lock()
		c.closed = true
		close(c.send)
		c.mu.Unlock()

		c.hub.unregister(c)
		// Keep close handshakes out of the broadcast path, where they could delay other clients.
		go func() { _ = c.conn.Close(code, reason) }()
	})
}
