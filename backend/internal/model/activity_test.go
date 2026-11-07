package model

import (
	"github.com/google/uuid"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestActivityLogRetainsReachAndOrdersRecords(t *testing.T) {
	now := time.Now()
	id := func(n byte) RecordID { var u uuid.UUID; u[15] = n; return RecordID(u) }
	room := &Room{ReachRecords: []ReachRecord{{RecordID: id(1), UserID: "alice", CreatedAt: now}}, BingoRecords: []BingoRecord{{RecordID: id(4), UserID: "alice", Order: 3, CreatedAt: now.Add(time.Second)}, {RecordID: id(3), UserID: "bob", Order: 1, CreatedAt: now.Add(time.Second)}, {RecordID: id(2), UserID: "alice", Order: 2, CreatedAt: now.Add(time.Second)}}}
	log := room.ActivityLog()
	if len(log) != 4 || log[0].Kind != "reach" || log[1].ID != id(2) || *log[1].BingoRank != 2 || *log[2].BingoRank != 1 || *log[3].BingoRank != 2 || *log[3].LineCount != 2 {
		t.Fatalf("unexpected log: %+v", log)
	}
	slices.Reverse(room.BingoRecords)
	if !reflect.DeepEqual(log, room.ActivityLog()) {
		t.Fatal("record loading order changed log")
	}
}
