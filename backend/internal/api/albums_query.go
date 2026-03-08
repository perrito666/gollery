package api

import (
	"net/http"
	"strings"
)

func (s *Server) handleAlbumsByPath(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	// Normalize: strip leading/trailing slashes.
	path = strings.Trim(path, "/")

	s.mu.RLock()
	defer s.mu.RUnlock()

	album, ok := s.albumsByPath[path]
	if !ok {
		writeError(w, http.StatusNotFound, "album not found")
		return
	}

	if !s.checkAlbumAccess(w, r, album) {
		return
	}

	offset, limit := parsePagination(r)
	writeJSON(w, http.StatusOK, albumToResponse(album, offset, limit))
}
