package model

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJoinIdempotentAfterStart(t *testing.T) {
	now := time.Now()
	room := NewRoom(RoomID(uuid.New()), "123456", RoomSettings{Name: "test", Admins: []UserID{"admin"}}, now)
	if err := room.Join("alice", now); err != nil {
		t.Fatal(err)
	}
	card, err := MakeRandomCard("alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := room.StartGame("admin", []Card{card}, now); err != nil {
		t.Fatal(err)
	}
	for _, state := range []RoomState{RoomStatePlaying, RoomStateFinished} {
		room.State = state
		if err := room.Join("alice", now.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		if len(room.Participants) != 1 || !room.Participants[0].JoinedAt.Equal(now) {
			t.Fatal("join changed existing participant")
		}
		if err := room.Join("bob", now); err != ErrRoomNotJoinable {
			t.Fatalf("new join: %v", err)
		}
	}
}

func TestMessageRuneLimit(t *testing.T) {
	room := NewRoom(RoomID(uuid.New()), "123456", RoomSettings{Name: "test", Admins: []UserID{"admin"}}, time.Now())
	for _, content := range []string{strings.Repeat("あ", 500), strings.Repeat("🎉", 500), strings.Repeat("a", 500)} {
		if _, err := room.PostMessage("admin", content, time.Now(), MessageID(uuid.New())); err != nil {
			t.Fatal(err)
		}
	}
	for _, content := range []string{"", strings.Repeat("あ", 501), strings.Repeat("🎉", 501)} {
		if _, err := room.PostMessage("admin", content, time.Now(), MessageID(uuid.New())); err == nil {
			t.Fatal("invalid content accepted")
		}
	}
}

func TestFinishGamePickCanceled(t *testing.T) {
	for _, picking := range []bool{false, true} {
		room := NewRoom(RoomID(uuid.New()), "123456", RoomSettings{Name: "test", Admins: []UserID{"admin"}}, time.Now())
		room.State = RoomStatePlaying
		if picking {
			room.PickState = RoomPickStatePicking
		}
		result, err := room.FinishGame("admin", time.Now())
		if err != nil {
			t.Fatal(err)
		}
		if result.PickCanceled != picking || room.State != RoomStateFinished || room.PickState != RoomPickStateIdle {
			t.Fatalf("unexpected finish: %+v", result)
		}
	}
}
