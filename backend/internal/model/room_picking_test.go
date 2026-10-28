package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFinishPickEdgeCases(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name             string
		picked           BallNumber
		prepare          func(*testing.T, *Room)
		finishBeforeTest bool
		wantErr          error
		wantPickState    RoomPickState
		wantBallCount    int
		wantAllPicked    bool
	}{
		{
			name:   "last drawable ball exhausts the pool",
			picked: MaxBallNumber,
			prepare: func(_ *testing.T, room *Room) {
				room.PickedBalls = AllBallNumbers()[:int(MaxBallNumber)-1]
			},
			wantPickState: RoomPickStateExhausted,
			wantBallCount: int(MaxBallNumber),
			wantAllPicked: true,
		},
		{
			name:   "canceled pick cannot finish",
			picked: 1,
			prepare: func(t *testing.T, room *Room) {
				if err := room.CancelPick("admin", now); err != nil {
					t.Fatal(err)
				}
			},
			wantErr:       ErrRoomPickNotFinishable,
			wantPickState: RoomPickStateIdle,
		},
		{
			name:             "repeated finish is rejected",
			picked:           1,
			finishBeforeTest: true,
			wantErr:          ErrRoomPickNotFinishable,
			wantPickState:    RoomPickStateIdle,
			wantBallCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := NewRoom(RoomID(uuid.New()), "123456", RoomSettings{Admins: []UserID{"admin"}}, now)
			room.State = RoomStatePlaying
			room.PickState = RoomPickStatePicking
			if tt.prepare != nil {
				tt.prepare(t, room)
			}

			nextRecordID := func() (RecordID, error) { return RecordID(uuid.New()), nil }
			if tt.finishBeforeTest {
				if _, err := room.FinishPick("admin", tt.picked, nextRecordID, now); err != nil {
					t.Fatalf("first finish: %v", err)
				}
			}

			result, err := room.FinishPick("admin", tt.picked, nextRecordID, now)
			if err != tt.wantErr {
				t.Fatalf("FinishPick() error = %v, want %v", err, tt.wantErr)
			}
			if room.PickState != tt.wantPickState || len(room.PickedBalls) != tt.wantBallCount {
				t.Fatalf("room after finish = state %q, %d picked balls; want state %q, %d", room.PickState, len(room.PickedBalls), tt.wantPickState, tt.wantBallCount)
			}
			if result.AllPicked != tt.wantAllPicked {
				t.Fatalf("AllPicked = %t, want %t", result.AllPicked, tt.wantAllPicked)
			}
			if tt.wantErr == nil && (result.Room != room || result.PickedBall != tt.picked) {
				t.Fatalf("unexpected successful result: %+v", result)
			}
		})
	}
}
