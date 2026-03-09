package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/state"
)

func (s *Server) handleAssetMetadataPatch(w http.ResponseWriter, r *http.Request) {
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

	if !s.requireAdmin(w, r, album) {
		return
	}

	var req MetadataPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	albumAbsPath := filepath.Join(s.contentRoot, asset.AlbumPath)
	st, err := state.LoadAssetState(albumAbsPath, asset.Filename)
	if err != nil {
		slog.Error("loading asset state", "asset_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to load asset state")
		return
	}
	if st == nil {
		st = &state.AssetState{}
	}

	if req.Title != nil {
		st.Title = *req.Title
	}
	if req.Description != nil {
		st.Description = *req.Description
	}

	if err := state.SaveAssetState(albumAbsPath, asset.Filename, st); err != nil {
		slog.Error("saving asset state", "asset_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to save asset state")
		return
	}

	// Update in-memory snapshot.
	asset.Title = st.Title
	asset.Description = st.Description

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleAlbumMetadataPatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	album, ok := s.albumsByID[id]
	if !ok {
		writeError(w, http.StatusNotFound, "album not found")
		return
	}

	if !s.requireAdmin(w, r, album) {
		return
	}

	var req MetadataPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Load album.json, patch, and save.
	albumAbsPath := filepath.Join(s.contentRoot, album.Path)
	albumJSONPath := filepath.Join(albumAbsPath, "album.json")

	cfg, err := config.LoadAlbumConfig(albumJSONPath)
	if err != nil {
		slog.Error("loading album config", "album_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to load album config")
		return
	}

	if req.Title != nil {
		cfg.Title = *req.Title
	}
	if req.Description != nil {
		cfg.Description = *req.Description
	}

	if err := config.SaveAlbumConfig(albumJSONPath, cfg); err != nil {
		slog.Error("saving album config", "album_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to save album config")
		return
	}

	// Update in-memory snapshot.
	album.Title = cfg.Title
	album.Description = cfg.Description

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
