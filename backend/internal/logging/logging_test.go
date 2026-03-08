package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/albums/root", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v\nraw: %s", err, buf.String())
	}

	if entry["method"] != "GET" {
		t.Errorf("method = %v", entry["method"])
	}
	if entry["path"] != "/api/v1/albums/root" {
		t.Errorf("path = %v", entry["path"])
	}
	if entry["status"] != float64(200) {
		t.Errorf("status = %v", entry["status"])
	}
	if _, ok := entry["request_id"]; !ok {
		t.Error("missing request_id")
	}
	if _, ok := entry["duration_ms"]; !ok {
		t.Error("missing duration_ms")
	}
}

func TestMiddleware_CapturesNonOKStatus(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest("GET", "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}
	if entry["status"] != float64(404) {
		t.Errorf("status = %v, want 404", entry["status"])
	}
}

func TestMiddleware_SetsRequestID(t *testing.T) {
	var gotID string
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// Suppress log output for this test.
	slog.SetDefault(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if gotID == "" {
		t.Error("expected non-empty request ID in context")
	}
	if len(gotID) != 16 { // 8 bytes = 16 hex chars
		t.Errorf("request ID length = %d, want 16", len(gotID))
	}
}

func TestMiddleware_DefaultStatusOnWrite(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}
	if entry["status"] != float64(200) {
		t.Errorf("status = %v, want 200", entry["status"])
	}
}

func TestRequestID_EmptyWithoutMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if id := RequestID(req.Context()); id != "" {
		t.Errorf("expected empty request ID, got %q", id)
	}
}
