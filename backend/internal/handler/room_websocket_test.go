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

func TestParticipantJoinedEventIsDisplayOnly(t *testing.T) {
	body := openapi.ParticipantJoinedBody{Participant: openapi.User{UserID: "alice"}}
	event := openapi.DisplayParticipantJoinedEvent{Type: openapi.ParticipantJoined, Body: body}
	payload, err := marshalWebSocketEvent(openapi.Display, event)
	if err != nil || strings.Contains(string(payload), `"participantCount"`) || !strings.Contains(string(payload), `"userId":"alice"`) {
		t.Fatalf("payload %s, error %v", payload, err)
	}
	if _, ok, err := marshalWebSocketEventForMode(openapi.Participant, event); err != nil || ok {
		t.Fatalf("participant mode accepted display event: ok=%v, error=%v", ok, err)
	}
}

func TestInitializedEventIncludesFullActivityLogAndDisplayCounts(t *testing.T) {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	room := model.NewRoom(
		model.RoomID(uuid.New()),
		"123456",
		model.RoomSettings{Name: "test", Admins: []model.UserID{"admin"}},
		createdAt,
	)
	room.State = model.RoomStatePlaying
	room.ReachRecords = []model.ReachRecord{{
		RecordID: model.RecordID(uuid.New()), UserID: "alice", Line: 1, CreatedAt: createdAt,
	}}
	room.BingoRecords = []model.BingoRecord{{
		RecordID: model.RecordID(uuid.New()), UserID: "bob", Line: 2, Order: 1, CreatedAt: createdAt.Add(time.Minute),
	}}

	for _, mode := range []openapi.WebSocketMode{openapi.Participant, openapi.Display} {
		event, ok, err := buildInitializedEvent(room, mode, "alice")
		if err != nil || !ok {
			t.Fatalf("buildInitializedEvent(%s): ok=%v, error=%v", mode, ok, err)
		}
		payload, err := marshalWebSocketEvent(mode, event)
		if err != nil {
			t.Fatalf("marshalWebSocketEvent(%s): %v", mode, err)
		}
		text := string(payload)
		if !strings.Contains(text, `"activityLog":[`) ||
			!strings.Contains(text, createdAt.Format(time.RFC3339)) ||
			!strings.Contains(text, createdAt.Add(time.Minute).Format(time.RFC3339)) {
			t.Fatalf("initialized payload is missing full activity history: %s", text)
		}
		for _, obsolete := range []string{`"bingoSummaries"`, `"reachSummaries"`, `"participantCount"`, `"bingoParticipantCount"`, `"reachParticipantCount"`} {
			if strings.Contains(text, obsolete) {
				t.Fatalf("initialized payload contains obsolete field %s: %s", obsolete, text)
			}
		}
	}
}
