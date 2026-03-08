package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/state"
)

func (s *Server) handleAlbumAccess(w http.ResponseWriter, r *http.Request) {
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

	var acl *config.AccessConfig
	if cfg, ok := s.configs[album.Path]; ok {
		acl = cfg.Access
	}

	writeJSON(w, http.StatusOK, aclToResponse(acl))
}

func (s *Server) handleAssetAccess(w http.ResponseWriter, r *http.Request) {
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

	var acl *config.AccessConfig
	if cfg, ok := s.configs[album.Path]; ok {
		acl = cfg.Access
	}

	writeJSON(w, http.StatusOK, aclToResponse(acl))
}

func (s *Server) handleAssetAccessPatch(w http.ResponseWriter, r *http.Request) {
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

	var req AccessPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate view mode if provided.
	if req.View != nil && *req.View != "" {
		if !config.ValidAccessModes[*req.View] {
			writeError(w, http.StatusBadRequest, "invalid access mode: "+*req.View)
			return
		}
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

	if st.AccessOverride == nil {
		st.AccessOverride = &state.AccessOverride{}
	}
	if req.View != nil {
		st.AccessOverride.View = *req.View
	}
	if req.AllowedUsers != nil {
		st.AccessOverride.AllowedUsers = req.AllowedUsers
	}
	if req.AllowedGroups != nil {
		st.AccessOverride.AllowedGroups = req.AllowedGroups
	}

	if err := state.SaveAssetState(albumAbsPath, asset.Filename, st); err != nil {
		slog.Error("saving asset state", "asset_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to save asset state")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func aclToResponse(acl *config.AccessConfig) AccessResponse {
	resp := AccessResponse{
		View:          "public",
		AllowedUsers:  []string{},
		AllowedGroups: []string{},
		Admins:        []string{},
	}
	if acl != nil {
		if acl.View != "" {
			resp.View = acl.View
		}
		if acl.AllowedUsers != nil {
			resp.AllowedUsers = acl.AllowedUsers
		}
		if acl.AllowedGroups != nil {
			resp.AllowedGroups = acl.AllowedGroups
		}
		if acl.Admins != nil {
			resp.Admins = acl.Admins
		}
	}
	return resp
}
