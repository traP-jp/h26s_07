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

func TestAddedEventsAndInitializedModes(t *testing.T) {
	room := model.NewRoom(model.RoomID(uuid.New()), "123456", model.RoomSettings{Name: "test", Admins: []model.UserID{"admin"}}, time.Now())
	if err := room.Join("alice", time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []openapi.WebSocketMode{openapi.Display, openapi.Participant} {
		event, ok, err := buildInitializedEvent(room, mode, "alice")
		if err != nil || !ok {
			t.Fatal(err)
		}
		payload, err := marshalWebSocketEvent(mode, event)
		if err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{`"participantCount":1`, `"participants":[{"userId":"alice"}]`, `"activityLog":[]`} {
			if !strings.Contains(string(payload), field) {
				t.Fatalf("missing %s in %s", field, payload)
			}
		}
	}
	for _, event := range []any{openapi.DisplayRoomDeletedEvent{Type: openapi.DisplayRoomDeletedEventTypeRoomDeleted}, openapi.DisplayParticipantJoinedEvent{Type: openapi.DisplayParticipantJoinedEventTypeParticipantJoined}} {
		if _, err := marshalWebSocketEvent(openapi.Display, event); err != nil {
			t.Fatal(err)
		}
		if _, ok, err := marshalWebSocketEventForMode(openapi.Participant, event); err != nil || ok {
			t.Fatal("display event leaked to participant", err)
		}
	}
	for _, event := range []any{openapi.ParticipantRoomDeletedEvent{Type: openapi.ParticipantRoomDeletedEventTypeRoomDeleted}, openapi.ParticipantParticipantJoinedEvent{Type: openapi.ParticipantParticipantJoinedEventTypeParticipantJoined}} {
		if _, err := marshalWebSocketEvent(openapi.Participant, event); err != nil {
			t.Fatal(err)
		}
		if _, ok, err := marshalWebSocketEventForMode(openapi.Display, event); err != nil || ok {
			t.Fatal("participant event leaked to display", err)
		}
	}
}

func TestRoomDeletedDeliveredBeforeNormalClosure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if errors.Is(err, syscall.EPERM) {
		t.Skip("sandbox does not allow listening sockets")
	}
	if err != nil {
		t.Fatal(err)
	}
	hub := NewWebSocketHub()
	id := model.RoomID(uuid.New())
	ready := make(chan struct{})
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		client := &webSocketClient{hub: hub, roomID: id, mode: openapi.Display, conn: conn, ctx: conn.CloseRead(ctx), send: make(chan []byte, 64)}
		hub.register(client)
		close(ready)
		client.writeLoop()
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
	<-ready
	if err := hub.sendRoom(id, openapi.DisplayRoomDeletedEvent{Type: openapi.DisplayRoomDeletedEventTypeRoomDeleted, Body: openapi.RoomDeletedBody{}}); err != nil {
		t.Fatal(err)
	}
	_, payload, err := conn.Read(ctx)
	if err != nil || !strings.Contains(string(payload), `"type":"RoomDeleted"`) {
		t.Fatalf("delete event: %s %v", payload, err)
	}
	_, _, err = conn.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		t.Fatalf("close: %v", err)
	}
}
