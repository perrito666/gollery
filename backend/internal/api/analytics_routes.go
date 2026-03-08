package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/perrito666/gollery/backend/internal/auth"
)

// StatsResponse is the JSON body for stats endpoints.
type StatsResponse struct {
	TotalViews int64 `json:"total_views"`
	Views7d    int64 `json:"views_7d"`
	Views30d   int64 `json:"views_30d"`
}

// PopularAssetResponse is one entry in the popular-assets list.
type PopularAssetResponse struct {
	ID         string `json:"id"`
	Filename   string `json:"filename"`
	TotalViews int64  `json:"total_views"`
}

func (s *Server) handleAlbumStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	album, ok := s.albumsByID[id]
	if !ok {
		writeError(w, http.StatusNotFound, "album not found")
		return
	}

	if !s.checkAlbumAccess(w, r, album) {
		return
	}

	total, v7d, v30d, err := s.analyticsStore.QueryPopularity(r.Context(), album.ID)
	if err != nil {
		slog.Error("querying album stats", "album_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to query stats")
		return
	}

	writeJSON(w, http.StatusOK, StatsResponse{TotalViews: total, Views7d: v7d, Views30d: v30d})
}

func (s *Server) handleAssetStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	asset, ok := s.assetsByID[id]
	if !ok {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}

	album, ok := s.snapshot.Albums[asset.AlbumPath]
	if !ok {
		writeError(w, http.StatusNotFound, "containing album not found")
		return
	}

	if !s.checkAlbumAccess(w, r, album) {
		return
	}

	total, v7d, v30d, err := s.analyticsStore.QueryPopularity(r.Context(), asset.ID)
	if err != nil {
		slog.Error("querying asset stats", "asset_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to query stats")
		return
	}

	writeJSON(w, http.StatusOK, StatsResponse{TotalViews: total, Views7d: v7d, Views30d: v30d})
}

func (s *Server) handlePopularAssets(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	album, ok := s.albumsByID[id]
	if !ok {
		writeError(w, http.StatusNotFound, "album not found")
		return
	}

	if !s.checkAlbumAccess(w, r, album) {
		return
	}

	limit := 10
	if qs := r.URL.Query().Get("limit"); qs != "" {
		if parsed, err := strconv.Atoi(qs); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	popular, err := s.analyticsStore.QueryPopularAssets(r.Context(), album.ID, limit)
	if err != nil {
		slog.Error("querying popular assets", "album_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to query popular assets")
		return
	}

	resp := make([]PopularAssetResponse, 0, len(popular))
	for _, p := range popular {
		if asset, ok := s.assetsByID[p.AssetID]; ok {
			resp = append(resp, PopularAssetResponse{
				ID:         p.AssetID,
				Filename:   asset.Filename,
				TotalViews: p.TotalViews,
			})
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFromContext(r.Context())
	if p == nil || !p.IsAdmin {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	overview, err := s.analyticsStore.QueryOverview(r.Context())
	if err != nil {
		slog.Error("querying analytics overview", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to query overview")
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handlePopularityStatus(w http.ResponseWriter, r *http.Request) {
	enabled := s.analyticsStore != nil
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}
