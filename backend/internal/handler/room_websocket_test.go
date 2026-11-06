package handler

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func TestSendRoomContinuesAfterClientError(t *testing.T) {
	hub := NewWebSocketHub()
	id := model.RoomID(uuid.New())
	bad := &webSocketClient{hub: hub, roomID: id, mode: openapi.WebSocketMode("invalid"), ctx: context.Background(), send: make(chan []byte, 64)}
	good := &webSocketClient{hub: hub, roomID: id, mode: openapi.Display, ctx: context.Background(), send: make(chan []byte, 64)}
	hub.register(bad)
	hub.register(good)
	event := openapi.DisplayPickStartedEvent{Type: openapi.DisplayPickStartedEventTypePickStarted, Body: openapi.PickStartedBody{}}
	if err := hub.sendRoom(id, event); err == nil {
		t.Fatal("expected invalid client mode error")
	}
	if len(good.send) != 1 {
		t.Fatal("healthy client missed event")
	}
}

func TestSlowClientReceivesCloseStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	accepted := make(chan *websocket.Conn, 1)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if errors.Is(err, syscall.EPERM) {
		t.Skip("sandbox does not allow listening sockets")
	}
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		accepted <- conn
		<-ctx.Done()
	}))
	_ = server.Listener.Close()
	server.Listener = listener
	server.Start()
	defer server.Close()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	serverConn := <-accepted
	defer serverConn.CloseNow()
	hub := NewWebSocketHub()
	id := model.RoomID(uuid.New())
	slow := &webSocketClient{hub: hub, roomID: id, mode: openapi.Display, ctx: serverConn.CloseRead(ctx), conn: serverConn, send: make(chan []byte, 1)}
	slow.send <- []byte("full")
	good := &webSocketClient{hub: hub, roomID: id, mode: openapi.Display, ctx: ctx, send: make(chan []byte, 64)}
	hub.register(slow)
	hub.register(good)
	event := openapi.DisplayPickStartedEvent{Type: openapi.DisplayPickStartedEventTypePickStarted, Body: openapi.PickStartedBody{}}
	if err := hub.sendRoom(id, event); err == nil {
		t.Fatal("expected slow client error")
	}
	if len(good.send) != 1 {
		t.Fatal("healthy client missed event")
	}
	_, _, err = conn.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusTryAgainLater {
		t.Fatalf("close status: %v", err)
	}
	if !strings.Contains(err.Error(), "client is too slow") {
		t.Fatalf("close reason: %v", err)
	}
	cancel()
}

func TestParticipantJoinedEventUsesMatchingMode(t *testing.T) {
	body := openapi.ParticipantJoinedBody{Participant: openapi.User{UserID: "alice"}, ParticipantCount: 2, JoinedAt: time.Now()}
	for _, tc := range []struct {
		mode  openapi.WebSocketMode
		event any
	}{
		{openapi.Display, openapi.DisplayParticipantJoinedEvent{Type: openapi.DisplayParticipantJoinedEventTypeParticipantJoined, Body: body}},
		{openapi.Participant, openapi.ParticipantParticipantJoinedEvent{Type: openapi.ParticipantParticipantJoinedEventTypeParticipantJoined, Body: body}},
	} {
		payload, err := marshalWebSocketEvent(tc.mode, tc.event)
		if err != nil || !strings.Contains(string(payload), `"participantCount":2`) || !strings.Contains(string(payload), `"userId":"alice"`) {
			t.Fatalf("mode %s: payload %s, error %v", tc.mode, payload, err)
		}
	}
}
