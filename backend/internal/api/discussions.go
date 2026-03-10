package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/perrito666/gollery/backend/internal/auth"
	"github.com/perrito666/gollery/backend/internal/state"
)

const maxDiscussionURLLen = 2048

// isValidDiscussionURL returns true if rawURL is a valid HTTPS URL
// with a non-empty host and length within bounds.
func isValidDiscussionURL(rawURL string) bool {
	if !strings.HasPrefix(rawURL, "https://") {
		return false
	}
	if len(rawURL) > maxDiscussionURLLen {
		return false
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Host != ""
}

func (s *Server) handleAlbumDiscussionsList(w http.ResponseWriter, r *http.Request) {
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

	albumAbsPath := filepath.Join(s.contentRoot, album.Path)
	bindings, err := s.discussions.ListBindings(albumAbsPath, "album", "")
	if err != nil {
		slog.Error("listing album discussions", "album_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list discussions")
		return
	}

	writeJSON(w, http.StatusOK, bindingsToResponse(bindings))
}

func (s *Server) handleAlbumDiscussionsCreate(w http.ResponseWriter, r *http.Request) {
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

	if !s.requireAdmin(w, r, album) {
		return
	}

	var req CreateDiscussionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	principal := auth.PrincipalFromContext(r.Context())
	createdBy := ""
	if principal != nil {
		createdBy = principal.Username
	}

	albumAbsPath := filepath.Join(s.contentRoot, album.Path)

	var binding *state.DiscussionBinding
	var err error

	if req.URL != "" {
		if !isValidDiscussionURL(req.URL) {
			writeError(w, http.StatusBadRequest, "invalid URL: must be a valid https:// URL")
			return
		}
		provider := req.Provider
		if provider == "" {
			provider = "mastodon"
		}
		binding, err = s.discussions.LinkBinding(provider, req.URL, albumAbsPath, "album", "", createdBy)
	} else {
		binding, err = s.discussions.CreateBinding(r.Context(), req.Provider, albumAbsPath, "album", "", req.Title, req.Body, createdBy)
	}
	if err != nil {
		slog.Error("creating album discussion", "album_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create discussion")
		return
	}

	writeJSON(w, http.StatusCreated, DiscussionBindingResponse{
		Provider:  binding.Provider,
		URL:       binding.URL,
		CreatedAt: binding.CreatedAt,
		CreatedBy: binding.CreatedBy,
	})
}

func (s *Server) handleAssetDiscussionsList(w http.ResponseWriter, r *http.Request) {
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

	albumAbsPath := filepath.Join(s.contentRoot, asset.AlbumPath)
	bindings, err := s.discussions.ListBindings(albumAbsPath, "asset", asset.Filename)
	if err != nil {
		slog.Error("listing asset discussions", "asset_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list discussions")
		return
	}

	writeJSON(w, http.StatusOK, bindingsToResponse(bindings))
}

func (s *Server) handleAssetDiscussionsCreate(w http.ResponseWriter, r *http.Request) {
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

	if !s.requireAdmin(w, r, album) {
		return
	}

	var req CreateDiscussionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	principal := auth.PrincipalFromContext(r.Context())
	createdBy := ""
	if principal != nil {
		createdBy = principal.Username
	}

	albumAbsPath := filepath.Join(s.contentRoot, asset.AlbumPath)

	var binding *state.DiscussionBinding
	var err error

	if req.URL != "" {
		if !isValidDiscussionURL(req.URL) {
			writeError(w, http.StatusBadRequest, "invalid URL: must be a valid https:// URL")
			return
		}
		provider := req.Provider
		if provider == "" {
			provider = "mastodon"
		}
		binding, err = s.discussions.LinkBinding(provider, req.URL, albumAbsPath, "asset", asset.Filename, createdBy)
	} else {
		binding, err = s.discussions.CreateBinding(r.Context(), req.Provider, albumAbsPath, "asset", asset.Filename, req.Title, req.Body, createdBy)
	}
	if err != nil {
		slog.Error("creating asset discussion", "asset_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create discussion")
		return
	}

	writeJSON(w, http.StatusCreated, DiscussionBindingResponse{
		Provider:  binding.Provider,
		URL:       binding.URL,
		CreatedAt: binding.CreatedAt,
		CreatedBy: binding.CreatedBy,
	})
}

func bindingsToResponse(bindings []state.DiscussionBinding) []DiscussionBindingResponse {
	resp := make([]DiscussionBindingResponse, len(bindings))
	for i, b := range bindings {
		resp[i] = DiscussionBindingResponse{
			Provider:  b.Provider,
			URL:       b.URL,
			CreatedAt: b.CreatedAt,
			CreatedBy: b.CreatedBy,
		}
	}
	return resp
}
