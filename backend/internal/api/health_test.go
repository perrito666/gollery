package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/domain"
)

func TestHealthz(t *testing.T) {
	snap := &domain.Snapshot{Albums: map[string]*domain.Album{}}
	srv := NewServer(snap, nil)
	handler := srv.Handler()

	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("body = %v, want status=ok", body)
	}
}

func TestHTTPServer_Defaults(t *testing.T) {
	snap := &domain.Snapshot{Albums: map[string]*domain.Album{}}
	srv := NewServer(snap, nil)
	httpSrv := srv.HTTPServer(":8080", nil)

	if httpSrv.Addr != ":8080" {
		t.Errorf("addr = %q, want :8080", httpSrv.Addr)
	}
	if httpSrv.ReadTimeout != DefaultReadTimeout {
		t.Errorf("ReadTimeout = %v, want %v", httpSrv.ReadTimeout, DefaultReadTimeout)
	}
	if httpSrv.WriteTimeout != DefaultWriteTimeout {
		t.Errorf("WriteTimeout = %v, want %v", httpSrv.WriteTimeout, DefaultWriteTimeout)
	}
	if httpSrv.ReadHeaderTimeout != DefaultReadHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want %v", httpSrv.ReadHeaderTimeout, DefaultReadHeaderTimeout)
	}
	if httpSrv.IdleTimeout != DefaultIdleTimeout {
		t.Errorf("IdleTimeout = %v, want %v", httpSrv.IdleTimeout, DefaultIdleTimeout)
	}
	if httpSrv.MaxHeaderBytes != DefaultMaxHeaderBytes {
		t.Errorf("MaxHeaderBytes = %d, want %d", httpSrv.MaxHeaderBytes, DefaultMaxHeaderBytes)
	}
}

func TestHTTPServer_CustomTimeouts(t *testing.T) {
	snap := &domain.Snapshot{Albums: map[string]*domain.Album{}}
	srv := NewServer(snap, nil)
	timeouts := &config.TimeoutConfig{
		ReadTimeoutSecs:       20,
		WriteTimeoutSecs:      60,
		ReadHeaderTimeoutSecs: 10,
		IdleTimeoutSecs:       300,
	}
	httpSrv := srv.HTTPServer(":9090", timeouts)

	if httpSrv.ReadTimeout != 20*time.Second {
		t.Errorf("ReadTimeout = %v, want 20s", httpSrv.ReadTimeout)
	}
	if httpSrv.WriteTimeout != 60*time.Second {
		t.Errorf("WriteTimeout = %v, want 60s", httpSrv.WriteTimeout)
	}
	if httpSrv.ReadHeaderTimeout != 10*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 10s", httpSrv.ReadHeaderTimeout)
	}
	if httpSrv.IdleTimeout != 300*time.Second {
		t.Errorf("IdleTimeout = %v, want 300s", httpSrv.IdleTimeout)
	}
}
