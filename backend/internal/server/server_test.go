package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/traP-jp/h26_07/backend/internal/config"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func TestHealthz(t *testing.T) {
	e := New(testConfig())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetMeUsesForwardedUser(t *testing.T) {
	e := New(testConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("X-Forwarded-User", "mumumu")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body openapi.User
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.ID != "mumumu" || body.Name != "mumumu" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestGetMeFallsBackToDeveloper(t *testing.T) {
	e := New(testConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body openapi.User
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.ID != "traP" || body.Name != "traP" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func testConfig() config.Config {
	return config.Config{
		Port:             "8080",
		CORSAllowOrigins: []string{"http://localhost:5173"},
	}
}
