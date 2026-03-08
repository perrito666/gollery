// Package api implements the REST API handlers and routing.
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/perrito666/gollery/backend/internal/access"
	"github.com/perrito666/gollery/backend/internal/auth"
	"github.com/perrito666/gollery/backend/internal/cache"
	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/discussion"
	"github.com/perrito666/gollery/backend/internal/domain"
)

// APIError is the standard error response body.
type APIError struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
}

// AlbumResponse is the JSON representation of an album.
type AlbumResponse struct {
	ID          string              `json:"id"`
	Path        string              `json:"path"`
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	ParentPath  string              `json:"parent_path,omitempty"`
	Children    []ChildAlbumSummary `json:"children"`
	Assets      []AssetSummary      `json:"assets"`
	TotalAssets int                 `json:"total_assets"`
}

// ChildAlbumSummary is a brief representation of a child album.
type ChildAlbumSummary struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Title string `json:"title,omitempty"`
}

// AssetSummary is a brief representation of an asset within an album listing.
type AssetSummary struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

// AssetResponse is the JSON representation of an asset.
type AssetResponse struct {
	ID           string  `json:"id"`
	Filename     string  `json:"filename"`
	AlbumPath    string  `json:"album_path"`
	AlbumID      string  `json:"album_id"`
	SizeBytes    int64   `json:"size_bytes"`
	PrevAssetID  *string `json:"prev_asset_id"`
	NextAssetID  *string `json:"next_asset_id"`
}

// LoginRequest is the JSON body for POST /api/v1/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// MeResponse is the JSON body for GET /api/v1/auth/me.
type MeResponse struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
	IsAdmin  bool     `json:"is_admin"`
}

// DiscussionBindingResponse is the JSON representation of a discussion binding.
type DiscussionBindingResponse struct {
	Provider  string `json:"provider"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

// CreateDiscussionRequest is the JSON body for creating discussion threads.
type CreateDiscussionRequest struct {
	Provider string `json:"provider"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

// AccessResponse is the JSON representation of effective access configuration.
type AccessResponse struct {
	View          string   `json:"view"`
	AllowedUsers  []string `json:"allowed_users"`
	AllowedGroups []string `json:"allowed_groups"`
	Admins        []string `json:"admins"`
}

// AccessPatchRequest is the JSON body for PATCH asset access.
type AccessPatchRequest struct {
	View          *string  `json:"view,omitempty"`
	AllowedUsers  []string `json:"allowed_users,omitempty"`
	AllowedGroups []string `json:"allowed_groups,omitempty"`
}

// AnalyticsStore is the interface the API needs from the analytics backend.
type AnalyticsStore interface {
	QueryPopularity(ctx context.Context, objectID string) (totalViews, views7d, views30d int64, err error)
	QueryPopularAssets(ctx context.Context, albumID string, limit int) ([]PopularAsset, error)
	QueryOverview(ctx context.Context) (*AnalyticsOverview, error)
}

// PopularAsset represents a popular asset from the analytics store.
type PopularAsset struct {
	AssetID    string
	TotalViews int64
}

// AnalyticsOverview holds aggregate analytics data.
type AnalyticsOverview struct {
	TotalEvents      int64 `json:"total_events"`
	UniqueVisitors7d int64 `json:"unique_visitors_7d"`
	TotalViews7d     int64 `json:"total_views_7d"`
	TotalViews30d    int64 `json:"total_views_30d"`
}

// Server holds the API state and handlers.
type Server struct {
	mu          sync.RWMutex
	snapshot    *domain.Snapshot
	configs     map[string]*config.AlbumConfig // keyed by album path
	contentRoot string
	cacheLayout *cache.Layout

	// indexes built from snapshot
	albumsByID   map[string]*domain.Album
	albumsByPath map[string]*domain.Album
	assetsByID   map[string]*domain.Asset

	// auth dependencies (optional, nil means no auth endpoints)
	authenticator auth.Authenticator
	sessions      *auth.CookieSessionStore
	csrfSecret    string
	rateLimitCfg  *RateLimitConfig

	// discussion service (optional)
	discussions *discussion.Service

	// analytics store (optional)
	analyticsStore AnalyticsStore

	// admin support
	reindexFunc    func() error
	startTime      time.Time
	lastScanErrors []string
}

// NewServer creates a new API server with the given snapshot and configs.
// contentRoot is the filesystem content root. cacheLayout may be nil to
// disable derivative serving.
func NewServer(snap *domain.Snapshot, configs map[string]*config.AlbumConfig) *Server {
	s := &Server{startTime: time.Now()}
	s.SetSnapshot(snap, configs)
	return s
}

// SetAuth configures the authentication backend and session store.
// csrfSecret is used for CSRF token generation/validation.
// rateLimitCfg is optional; if nil, no rate limiting is applied.
func (s *Server) SetAuth(authenticator auth.Authenticator, sessions *auth.CookieSessionStore, csrfSecret string, rateLimitCfg *RateLimitConfig) {
	s.authenticator = authenticator
	s.sessions = sessions
	s.csrfSecret = csrfSecret
	s.rateLimitCfg = rateLimitCfg
}

// SetContentRoot configures the filesystem paths for derivative generation.
func (s *Server) SetContentRoot(contentRoot string, cacheLayout *cache.Layout) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contentRoot = contentRoot
	s.cacheLayout = cacheLayout
}

// SetDiscussions configures the discussion service.
func (s *Server) SetDiscussions(svc *discussion.Service) {
	s.discussions = svc
}

// SetAnalytics configures the analytics store.
func (s *Server) SetAnalytics(store AnalyticsStore) {
	s.analyticsStore = store
}

// SetAdmin configures admin-only capabilities.
func (s *Server) SetAdmin(reindexFunc func() error) {
	s.reindexFunc = reindexFunc
}

// SetScanErrors stores the last scan errors for diagnostics.
func (s *Server) SetScanErrors(errs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastScanErrors = errs
}

// SetSnapshot replaces the current snapshot and rebuilds indexes.
func (s *Server) SetSnapshot(snap *domain.Snapshot, configs map[string]*config.AlbumConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshot = snap
	s.configs = configs
	s.albumsByID = make(map[string]*domain.Album, len(snap.Albums))
	s.albumsByPath = make(map[string]*domain.Album, len(snap.Albums))
	s.assetsByID = make(map[string]*domain.Asset)

	for _, album := range snap.Albums {
		s.albumsByID[album.ID] = album
		s.albumsByPath[album.Path] = album
		for i := range album.Assets {
			asset := &album.Assets[i]
			s.assetsByID[asset.ID] = asset
		}
	}
}

// Handler returns an http.Handler with all API routes registered.
// The auth session middleware is applied to all routes when auth is configured.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health check — no auth required.
	mux.HandleFunc("GET /healthz", handleHealthz)

	mux.HandleFunc("GET /api/v1/albums/root", s.handleAlbumsRoot)
	mux.HandleFunc("GET /api/v1/albums/{id}", s.handleAlbumByID)
	mux.HandleFunc("GET /api/v1/assets/{id}", s.handleAssetByID)
	mux.HandleFunc("GET /api/v1/assets/{id}/thumbnail", s.handleAssetThumbnail)
	mux.HandleFunc("GET /api/v1/assets/{id}/preview", s.handleAssetPreview)
	mux.HandleFunc("GET /api/v1/assets/{id}/original", s.handleAssetOriginal)

	if s.sessions != nil {
		mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
		mux.HandleFunc("GET /api/v1/auth/me", s.handleMe)
		mux.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)
		mux.HandleFunc("GET /api/v1/auth/csrf-token", s.handleCSRFToken)
	}

	// Admin routes
	mux.HandleFunc("POST /api/v1/admin/reindex", s.handleAdminReindex)
	mux.HandleFunc("GET /api/v1/admin/status", s.handleAdminStatus)
	mux.HandleFunc("GET /api/v1/admin/diagnostics", s.handleAdminDiagnostics)

	// Analytics routes
	if s.analyticsStore != nil {
		mux.HandleFunc("GET /api/v1/albums/{id}/stats", s.handleAlbumStats)
		mux.HandleFunc("GET /api/v1/assets/{id}/stats", s.handleAssetStats)
		mux.HandleFunc("GET /api/v1/albums/{id}/popular-assets", s.handlePopularAssets)
		mux.HandleFunc("GET /api/v1/admin/analytics/overview", s.handleAnalyticsOverview)
	}
	mux.HandleFunc("GET /api/v1/popularity/status", s.handlePopularityStatus)

	// Album path query
	mux.HandleFunc("GET /api/v1/albums", s.handleAlbumsByPath)

	mux.HandleFunc("GET /api/v1/albums/{id}/access", s.handleAlbumAccess)
	mux.HandleFunc("GET /api/v1/assets/{id}/access", s.handleAssetAccess)
	mux.HandleFunc("PATCH /api/v1/assets/{id}/access", s.handleAssetAccessPatch)

	if s.discussions != nil {
		mux.HandleFunc("GET /api/v1/albums/{id}/discussion-threads", s.handleAlbumDiscussionsList)
		mux.HandleFunc("POST /api/v1/albums/{id}/discussion-threads", s.handleAlbumDiscussionsCreate)
		mux.HandleFunc("GET /api/v1/assets/{id}/discussion-threads", s.handleAssetDiscussionsList)
		mux.HandleFunc("POST /api/v1/assets/{id}/discussion-threads", s.handleAssetDiscussionsCreate)
	}

	var handler http.Handler = mux
	if s.sessions != nil {
		handler = s.csrfMiddleware(s.authMiddleware(mux))
		if s.rateLimitCfg != nil {
			ratePaths := map[string]bool{
				"/api/v1/auth/login":  true,
				"/api/v1/auth/logout": true,
			}
			handler = RateLimitMiddleware(*s.rateLimitCfg, ratePaths, handler)
		}
	}
	return handler
}

// authMiddleware extracts the session cookie, looks up the session,
// and sets the principal in the request context.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := auth.TokenFromRequest(r)
		if token != "" {
			if p, err := s.sessions.Lookup(r.Context(), token); err == nil {
				r = r.WithContext(auth.WithPrincipal(r.Context(), p))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// csrfMiddleware validates CSRF tokens on state-changing methods.
// Skips GET/HEAD/OPTIONS and the login endpoint (no session yet).
func (s *Server) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safe methods don't need CSRF.
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Login is exempt — there's no session token yet.
		if r.URL.Path == "/api/v1/auth/login" {
			next.ServeHTTP(w, r)
			return
		}

		sessionToken := auth.TokenFromRequest(r)
		if sessionToken == "" {
			writeError(w, http.StatusForbidden, "CSRF validation failed")
			return
		}

		csrfToken := r.Header.Get("X-CSRF-Token")
		if !auth.ValidateCSRFToken(csrfToken, sessionToken, s.csrfSecret) {
			writeError(w, http.StatusForbidden, "CSRF validation failed")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkAlbumAccess checks ACL for an album and writes the error response
// if denied. Returns true if access is allowed.
func (s *Server) checkAlbumAccess(w http.ResponseWriter, r *http.Request, album *domain.Album) bool {
	var acl *config.AccessConfig
	if cfg, ok := s.configs[album.Path]; ok {
		acl = cfg.Access
	}

	principal := auth.PrincipalFromContext(r.Context())
	if access.CheckView(acl, principal) == access.Deny {
		if principal == nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
		} else {
			writeError(w, http.StatusForbidden, "access denied")
		}
		return false
	}
	return true
}

// checkAssetAccess checks ACL for an asset, using asset-level overrides
// if set, falling back to the album ACL. Returns true if access is allowed.
func (s *Server) checkAssetAccess(w http.ResponseWriter, r *http.Request, asset *domain.Asset, album *domain.Album) bool {
	var albumACL *config.AccessConfig
	if cfg, ok := s.configs[album.Path]; ok {
		albumACL = cfg.Access
	}

	effectiveACL := access.EffectiveAssetACL(albumACL, asset.Access)

	principal := auth.PrincipalFromContext(r.Context())
	if access.CheckView(effectiveACL, principal) == access.Deny {
		if principal == nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
		} else {
			writeError(w, http.StatusForbidden, "access denied")
		}
		return false
	}
	return true
}

// requireAdmin checks if the principal is an admin for the album.
// Returns true if the request should proceed.
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request, album *domain.Album) bool {
	var acl *config.AccessConfig
	if cfg, ok := s.configs[album.Path]; ok {
		acl = cfg.Access
	}

	principal := auth.PrincipalFromContext(r.Context())
	if !access.IsObjectAdmin(acl, principal) {
		writeError(w, http.StatusForbidden, "admin access required")
		return false
	}
	return true
}

// responseOpts builds albumResponseOpts from the current request and server state.
// Must be called while s.mu is held.
func (s *Server) responseOpts(r *http.Request) albumResponseOpts {
	return albumResponseOpts{
		albumsByPath: s.albumsByPath,
		configs:      s.configs,
		principal:    auth.PrincipalFromContext(r.Context()),
	}
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DefaultReadTimeout is the default read timeout for the HTTP server.
const DefaultReadTimeout = 10 * time.Second

// DefaultWriteTimeout is the default write timeout for the HTTP server.
const DefaultWriteTimeout = 30 * time.Second

// DefaultReadHeaderTimeout is the default read header timeout.
const DefaultReadHeaderTimeout = 5 * time.Second

// DefaultIdleTimeout is the default idle timeout.
const DefaultIdleTimeout = 120 * time.Second

// DefaultMaxHeaderBytes is the max header size (1 MB).
const DefaultMaxHeaderBytes = 1 << 20

// HTTPServer returns a configured *http.Server using the given address
// and optional timeout config. If timeouts is nil, sensible defaults apply.
func (s *Server) HTTPServer(addr string, timeouts *config.TimeoutConfig) *http.Server {
	readTimeout := DefaultReadTimeout
	writeTimeout := DefaultWriteTimeout
	readHeaderTimeout := DefaultReadHeaderTimeout
	idleTimeout := DefaultIdleTimeout

	if timeouts != nil {
		if timeouts.ReadTimeoutSecs > 0 {
			readTimeout = time.Duration(timeouts.ReadTimeoutSecs) * time.Second
		}
		if timeouts.WriteTimeoutSecs > 0 {
			writeTimeout = time.Duration(timeouts.WriteTimeoutSecs) * time.Second
		}
		if timeouts.ReadHeaderTimeoutSecs > 0 {
			readHeaderTimeout = time.Duration(timeouts.ReadHeaderTimeoutSecs) * time.Second
		}
		if timeouts.IdleTimeoutSecs > 0 {
			idleTimeout = time.Duration(timeouts.IdleTimeoutSecs) * time.Second
		}
	}

	return &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
	}
}

const (
	defaultPageLimit = 100
	maxPageLimit     = 500
)

// albumResponseOpts holds contextual data needed to filter album responses
// by the current user's access level.
type albumResponseOpts struct {
	albumsByPath map[string]*domain.Album
	configs      map[string]*config.AlbumConfig
	principal    *domain.Principal
}

func albumToResponse(a *domain.Album, opts albumResponseOpts, offset, limit int) AlbumResponse {
	albumACL := effectiveAlbumACL(opts.configs, a.Path)

	// Filter assets the principal can view.
	var visible []domain.Asset
	for _, ast := range a.Assets {
		effectiveACL := access.EffectiveAssetACL(albumACL, ast.Access)
		if access.CheckView(effectiveACL, opts.principal) == access.Allow {
			visible = append(visible, ast)
		}
	}

	total := len(visible)

	// Clamp offset.
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}

	// Clamp limit.
	if limit <= 0 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}

	end := offset + limit
	if end > total {
		end = total
	}

	page := visible[offset:end]
	assets := make([]AssetSummary, len(page))
	for i, ast := range page {
		assets[i] = AssetSummary{ID: ast.ID, Filename: ast.Filename}
	}

	// Filter children the principal can view.
	children := make([]ChildAlbumSummary, 0, len(a.Children))
	for _, childPath := range a.Children {
		childACL := effectiveAlbumACL(opts.configs, childPath)
		if access.CheckView(childACL, opts.principal) == access.Deny {
			continue
		}
		cs := ChildAlbumSummary{Path: childPath}
		if child, ok := opts.albumsByPath[childPath]; ok {
			cs.ID = child.ID
			cs.Title = child.Title
		}
		children = append(children, cs)
	}

	return AlbumResponse{
		ID:          a.ID,
		Path:        a.Path,
		Title:       a.Title,
		Description: a.Description,
		ParentPath:  a.ParentPath,
		Children:    children,
		Assets:      assets,
		TotalAssets: total,
	}
}

// effectiveAlbumACL returns the AccessConfig for an album path, or nil.
func effectiveAlbumACL(configs map[string]*config.AlbumConfig, path string) *config.AccessConfig {
	if cfg, ok := configs[path]; ok {
		return cfg.Access
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, APIError{
		Error:   http.StatusText(status),
		Status:  status,
		Message: msg,
	})
}
