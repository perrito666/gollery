package api

import (
	"net/http"
	"strconv"
)

func (s *Server) handleAlbumsRoot(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	root, ok := s.snapshot.Albums[""]
	if !ok {
		writeError(w, http.StatusNotFound, "no root album")
		return
	}

	if !s.checkAlbumAccess(w, r, root) {
		return
	}

	offset, limit := parsePagination(r)
	opts := s.responseOpts(r)
	writeJSON(w, http.StatusOK, albumToResponse(root, opts, offset, limit))
}

func (s *Server) handleAlbumByID(w http.ResponseWriter, r *http.Request) {
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

	offset, limit := parsePagination(r)
	opts := s.responseOpts(r)
	writeJSON(w, http.StatusOK, albumToResponse(album, opts, offset, limit))
}

// parsePagination extracts offset and limit from query parameters.
func parsePagination(r *http.Request) (offset, limit int) {
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	return offset, limit
}
