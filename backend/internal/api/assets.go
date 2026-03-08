package api

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/perrito666/gollery/backend/internal/derive"
	"github.com/perrito666/gollery/backend/internal/domain"
)

// findAdjacentAssets returns the previous and next asset IDs relative to
// assetID within the album's asset list, sorted by filename.
func findAdjacentAssets(album *domain.Album, assetID string) (prev, next *string) {
	if len(album.Assets) <= 1 {
		return nil, nil
	}

	// Build sorted list of assets by filename.
	sorted := make([]domain.Asset, len(album.Assets))
	copy(sorted, album.Assets)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Filename < sorted[j].Filename
	})

	for i, a := range sorted {
		if a.ID == assetID {
			if i > 0 {
				prev = &sorted[i-1].ID
			}
			if i < len(sorted)-1 {
				next = &sorted[i+1].ID
			}
			return prev, next
		}
	}
	return nil, nil
}

func (s *Server) handleAssetByID(w http.ResponseWriter, r *http.Request) {
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

	if !s.checkAssetAccess(w, r, asset, album) {
		return
	}

	prev, next := findAdjacentAssets(album, asset.ID)
	resp := AssetResponse{
		ID:          asset.ID,
		Filename:    asset.Filename,
		AlbumPath:   asset.AlbumPath,
		AlbumID:     album.ID,
		SizeBytes:   asset.SizeBytes,
		PrevAssetID: prev,
		NextAssetID: next,
	}
	writeJSON(w, http.StatusOK, resp)
}

// resolveAssetForDerivative is a shared helper for derivative endpoints.
func (s *Server) resolveAssetForDerivative(w http.ResponseWriter, r *http.Request) (*domain.Asset, string, bool) {
	id := r.PathValue("id")

	asset, ok := s.assetsByID[id]
	if !ok {
		writeError(w, http.StatusNotFound, "asset not found")
		return nil, "", false
	}

	album, ok := s.snapshot.Albums[asset.AlbumPath]
	if !ok {
		writeError(w, http.StatusNotFound, "containing album not found")
		return nil, "", false
	}

	if !s.checkAssetAccess(w, r, asset, album) {
		return nil, "", false
	}

	srcPath := filepath.Join(s.contentRoot, asset.AlbumPath, asset.Filename)
	return asset, srcPath, true
}

func (s *Server) handleAssetThumbnail(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cacheLayout == nil {
		writeError(w, http.StatusServiceUnavailable, "derivatives not configured")
		return
	}

	asset, srcPath, ok := s.resolveAssetForDerivative(w, r)
	if !ok {
		return
	}

	size := 400
	if qs := r.URL.Query().Get("size"); qs != "" {
		if parsed, err := strconv.Atoi(qs); err == nil && parsed > 0 && parsed <= 2000 {
			size = parsed
		}
	}

	outPath, err := derive.GenerateThumbnail(s.cacheLayout, asset.ID, srcPath, size)
	if err != nil {
		slog.Error("thumbnail generation failed", "asset_id", asset.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "thumbnail generation failed")
		return
	}

	http.ServeFile(w, r, outPath)
}

func (s *Server) handleAssetPreview(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cacheLayout == nil {
		writeError(w, http.StatusServiceUnavailable, "derivatives not configured")
		return
	}

	asset, srcPath, ok := s.resolveAssetForDerivative(w, r)
	if !ok {
		return
	}

	size := 1600
	if qs := r.URL.Query().Get("size"); qs != "" {
		if parsed, err := strconv.Atoi(qs); err == nil && parsed > 0 && parsed <= 4000 {
			size = parsed
		}
	}

	outPath, err := derive.GeneratePreview(s.cacheLayout, asset.ID, srcPath, size)
	if err != nil {
		slog.Error("preview generation failed", "asset_id", asset.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "preview generation failed")
		return
	}

	http.ServeFile(w, r, outPath)
}

func (s *Server) handleAssetOriginal(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, srcPath, ok := s.resolveAssetForDerivative(w, r)
	if !ok {
		return
	}

	http.ServeFile(w, r, srcPath)
}
