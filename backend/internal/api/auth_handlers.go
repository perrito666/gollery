package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/perrito666/gollery/backend/internal/auth"
)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	p, err := s.authenticator.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := s.sessions.Create(r.Context(), p)
	if err != nil {
		slog.Error("session creation failed", "error", err)
		writeError(w, http.StatusInternalServerError, "session creation failed")
		return
	}

	s.sessions.SetCookie(w, token)
	writeJSON(w, http.StatusOK, MeResponse{
		Username: p.Username,
		Groups:   p.Groups,
		IsAdmin:  p.IsAdmin,
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	p := auth.PrincipalFromContext(r.Context())
	if p == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	writeJSON(w, http.StatusOK, MeResponse{
		Username: p.Username,
		Groups:   p.Groups,
		IsAdmin:  p.IsAdmin,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := auth.TokenFromRequest(r)
	if token != "" {
		_ = s.sessions.Delete(r.Context(), token)
	}
	s.sessions.ClearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCSRFToken(w http.ResponseWriter, r *http.Request) {
	sessionToken := auth.TokenFromRequest(r)
	if sessionToken == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	token := auth.GenerateCSRFToken(sessionToken, s.csrfSecret)
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
