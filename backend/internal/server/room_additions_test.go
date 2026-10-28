package server

import (
	"encoding/json"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoomDeleteAndListHTTP(t *testing.T) {
	e := New(testConfig(), testutil.NewRoomDB(t))
	request := func(method, path, user, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("X-Forwarded-User", user)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != status {
			t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
		}
		return rec
	}
	rec := request(http.MethodPost, "/api/rooms", "admin", `{"settings":{"name":"test","description":""}}`, 200)
	var room openapi.Room
	if err := json.Unmarshal(rec.Body.Bytes(), &room); err != nil {
		t.Fatal(err)
	}
	path := "/api/rooms/" + room.RoomID
	request(http.MethodDelete, path, "other", "", 403)
	request(http.MethodPost, path+"/participants", "alice", "", 204)
	request(http.MethodPost, path+"/control/start", "admin", "", 204)
	request(http.MethodPost, path+"/control/finish", "admin", "", 204)
	rec = request(http.MethodGet, "/api/rooms", "admin", "", 200)
	var rooms []openapi.Room
	if err := json.Unmarshal(rec.Body.Bytes(), &rooms); err != nil || len(rooms) != 0 {
		t.Fatalf("default list: %s", rec.Body.String())
	}
	rec = request(http.MethodGet, "/api/rooms?includeFinished=true", "admin", "", 200)
	if err := json.Unmarshal(rec.Body.Bytes(), &rooms); err != nil || len(rooms) != 1 {
		t.Fatalf("full list: %s", rec.Body.String())
	}
	request(http.MethodGet, "/api/rooms?includeFinished=invalid", "admin", "", 400)
	request(http.MethodDelete, path, "admin", "", 204)
	request(http.MethodDelete, path, "admin", "", 404)
}
