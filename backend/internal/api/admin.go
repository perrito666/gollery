package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/perrito666/gollery/backend/internal/auth"
)

// StatusResponse is the JSON body for GET /api/v1/admin/status.
type StatusResponse struct {
	Uptime      string `json:"uptime"`
	SnapshotAge string `json:"snapshot_age"`
	ContentRoot string `json:"content_root"`
	AlbumCount  int    `json:"album_count"`
	AssetCount  int    `json:"asset_count"`
}

// DiagnosticsResponse is the JSON body for GET /api/v1/admin/diagnostics.
type DiagnosticsResponse struct {
	ScanErrors []string `json:"scan_errors"`
}

// requireGlobalAdmin checks that the principal is a global admin.
func (s *Server) requireGlobalAdmin(w http.ResponseWriter, r *http.Request) bool {
	p := auth.PrincipalFromContext(r.Context())
	if p == nil || !p.IsAdmin {
		writeError(w, http.StatusForbidden, "admin access required")
		return false
	}
	return true
}

func (s *Server) handleAdminReindex(w http.ResponseWriter, r *http.Request) {
	if !s.requireGlobalAdmin(w, r) {
		return
	}

	if s.reindexFunc == nil {
		writeError(w, http.StatusServiceUnavailable, "reindex not configured")
		return
	}

	if err := s.reindexFunc(); err != nil {
		slog.Error("reindex failed", "error", err)
		writeError(w, http.StatusInternalServerError, "reindex failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "reindex complete"})
}

func (s *Server) handleAdminStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireGlobalAdmin(w, r) {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := StatusResponse{
		Uptime:      time.Since(s.startTime).Truncate(time.Second).String(),
		SnapshotAge: time.Since(s.snapshot.GeneratedAt).Truncate(time.Second).String(),
		ContentRoot: s.contentRoot,
		AlbumCount:  len(s.albumsByID),
		AssetCount:  len(s.assetsByID),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAdminDiagnostics(w http.ResponseWriter, r *http.Request) {
	if !s.requireGlobalAdmin(w, r) {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	errs := s.lastScanErrors
	if errs == nil {
		errs = []string{}
	}
	writeJSON(w, http.StatusOK, DiagnosticsResponse{ScanErrors: errs})
}
