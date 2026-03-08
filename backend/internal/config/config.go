// Package config handles server and album configuration loading and inheritance.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ServerConfig holds the global server configuration.
type ServerConfig struct {
	// ContentRoot is the filesystem path to the root content directory.
	ContentRoot string `json:"content_root"`

	// CacheDir is the path to the gallery-cache directory for derivatives.
	CacheDir string `json:"cache_dir"`

	// ListenAddr is the address the server listens on (e.g. ":8080").
	ListenAddr string `json:"listen_addr"`

	// Auth holds authentication configuration.
	Auth *AuthConfig `json:"auth,omitempty"`

	// AllowedOrigins configures CORS allowed origins.
	AllowedOrigins []string `json:"allowed_origins,omitempty"`

	// Analytics holds global analytics configuration.
	Analytics *GlobalAnalyticsConfig `json:"analytics,omitempty"`

	// Timeouts configures HTTP server timeouts.
	Timeouts *TimeoutConfig `json:"timeouts,omitempty"`
}

// TimeoutConfig holds HTTP server timeout settings.
type TimeoutConfig struct {
	ReadTimeoutSecs       int `json:"read_timeout_seconds,omitempty"`
	WriteTimeoutSecs      int `json:"write_timeout_seconds,omitempty"`
	ReadHeaderTimeoutSecs int `json:"read_header_timeout_seconds,omitempty"`
	IdleTimeoutSecs       int `json:"idle_timeout_seconds,omitempty"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	// Provider is the auth provider type (e.g. "static", "oidc").
	Provider string `json:"provider"`

	// SessionSecret is used to sign session tokens.
	SessionSecret string `json:"session_secret"`

	// RateLimit configures rate limiting on auth endpoints.
	RateLimit *RateLimitConfig `json:"rate_limit,omitempty"`
}

// RateLimitConfig holds rate limiter settings.
type RateLimitConfig struct {
	// Rate is the number of requests per second.
	Rate float64 `json:"rate"`
	// Burst is the maximum burst size.
	Burst int `json:"burst"`
}

// GlobalAnalyticsConfig holds server-level analytics settings.
type GlobalAnalyticsConfig struct {
	Enabled           bool   `json:"enabled"`
	Backend           string `json:"backend"`
	PostgresDSNEnv    string `json:"postgres_dsn_env"`
	HashIP            bool   `json:"hash_ip"`
	DedupWindowSecs   int    `json:"dedup_window_seconds"`
	RetainEventsDays  int    `json:"retain_events_days"`
}

// AlbumConfig represents the declarative configuration from album.json.
// It defines publication rules, visibility, discussion policy, analytics
// policy, and derivative defaults. It must not contain mutable state.
type AlbumConfig struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// Inherit controls whether parent config is inherited.
	// If set to false, inheritance stops but server defaults still apply.
	Inherit *bool `json:"inherit,omitempty"`

	// Access defines visibility and ACL defaults for this album.
	Access *AccessConfig `json:"access,omitempty"`

	// Discussion defines the discussion policy for this album.
	Discussion *DiscussionPolicy `json:"discussion,omitempty"`

	// Analytics defines per-album analytics policy.
	Analytics *AlbumAnalyticsConfig `json:"analytics,omitempty"`

	// Derivatives defines default derivative generation settings.
	Derivatives *DerivativesConfig `json:"derivatives,omitempty"`
}

// AccessConfig defines visibility and ACL rules.
type AccessConfig struct {
	View          string   `json:"view,omitempty"`
	AllowedUsers  []string `json:"allowed_users,omitempty"`
	AllowedGroups []string `json:"allowed_groups,omitempty"`
	Admins        []string `json:"admins,omitempty"`
}

// DiscussionPolicy defines the discussion rules for an album.
type DiscussionPolicy struct {
	Enabled   *bool    `json:"enabled,omitempty"`
	Providers []string `json:"providers,omitempty"`
}

// AlbumAnalyticsConfig defines per-album analytics policy (declarative).
type AlbumAnalyticsConfig struct {
	Enabled           *bool `json:"enabled,omitempty"`
	TrackAlbumViews   *bool `json:"track_album_views,omitempty"`
	TrackAssetViews   *bool `json:"track_asset_views,omitempty"`
	TrackOriginalHits *bool `json:"track_original_hits,omitempty"`
	ExposePopularity  *bool `json:"expose_popularity,omitempty"`
}

// DerivativesConfig defines defaults for derivative generation.
type DerivativesConfig struct {
	ThumbnailSizes []int `json:"thumbnail_sizes,omitempty"`
	PreviewSizes   []int `json:"preview_sizes,omitempty"`
}

// ValidAccessModes lists the allowed values for AccessConfig.View.
var ValidAccessModes = map[string]bool{
	"public":        true,
	"authenticated": true,
	"restricted":    true,
}

// LoadAlbumConfig reads and parses an album.json file.
func LoadAlbumConfig(path string) (*AlbumConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading album config: %w", err)
	}
	var cfg AlbumConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing album config: %w", err)
	}
	return &cfg, nil
}

// Validate checks an AlbumConfig for structural correctness.
func (c *AlbumConfig) Validate() error {
	if c.Access != nil && c.Access.View != "" {
		if !ValidAccessModes[c.Access.View] {
			return fmt.Errorf("invalid access mode: %q", c.Access.View)
		}
	}
	return nil
}

// MergeAlbumConfigs merges a parent and child AlbumConfig following the
// inheritance rules from the technical design:
//   - scalar values: child overrides parent
//   - objects/maps: merge by key
//   - arrays/lists: child replaces parent
//
// If child.Inherit is false, the parent is ignored and only the child
// is returned (server defaults are applied elsewhere).
func MergeAlbumConfigs(parent, child *AlbumConfig) *AlbumConfig {
	if parent == nil {
		return child
	}
	if child == nil {
		return parent
	}

	// If child explicitly opts out of inheritance, return child as-is.
	if child.Inherit != nil && !*child.Inherit {
		return child
	}

	merged := *parent

	// Scalars: child overrides parent.
	if child.Title != "" {
		merged.Title = child.Title
	}
	if child.Description != "" {
		merged.Description = child.Description
	}
	merged.Inherit = child.Inherit

	// Objects: merge by key.
	merged.Access = mergeAccess(parent.Access, child.Access)
	merged.Discussion = mergeDiscussion(parent.Discussion, child.Discussion)
	merged.Analytics = mergeAnalytics(parent.Analytics, child.Analytics)
	merged.Derivatives = mergeDerivatives(parent.Derivatives, child.Derivatives)

	return &merged
}

func mergeAccess(parent, child *AccessConfig) *AccessConfig {
	if parent == nil {
		return child
	}
	if child == nil {
		return parent
	}
	merged := *parent
	if child.View != "" {
		merged.View = child.View
	}
	// Lists: child replaces parent.
	if child.AllowedUsers != nil {
		merged.AllowedUsers = child.AllowedUsers
	}
	if child.AllowedGroups != nil {
		merged.AllowedGroups = child.AllowedGroups
	}
	if child.Admins != nil {
		merged.Admins = child.Admins
	}
	return &merged
}

func mergeDiscussion(parent, child *DiscussionPolicy) *DiscussionPolicy {
	if parent == nil {
		return child
	}
	if child == nil {
		return parent
	}
	merged := *parent
	if child.Enabled != nil {
		merged.Enabled = child.Enabled
	}
	// List: child replaces parent.
	if child.Providers != nil {
		merged.Providers = child.Providers
	}
	return &merged
}

func mergeAnalytics(parent, child *AlbumAnalyticsConfig) *AlbumAnalyticsConfig {
	if parent == nil {
		return child
	}
	if child == nil {
		return parent
	}
	merged := *parent
	if child.Enabled != nil {
		merged.Enabled = child.Enabled
	}
	if child.TrackAlbumViews != nil {
		merged.TrackAlbumViews = child.TrackAlbumViews
	}
	if child.TrackAssetViews != nil {
		merged.TrackAssetViews = child.TrackAssetViews
	}
	if child.TrackOriginalHits != nil {
		merged.TrackOriginalHits = child.TrackOriginalHits
	}
	if child.ExposePopularity != nil {
		merged.ExposePopularity = child.ExposePopularity
	}
	return &merged
}

func mergeDerivatives(parent, child *DerivativesConfig) *DerivativesConfig {
	if parent == nil {
		return child
	}
	if child == nil {
		return parent
	}
	merged := *parent
	// Lists: child replaces parent.
	if child.ThumbnailSizes != nil {
		merged.ThumbnailSizes = child.ThumbnailSizes
	}
	if child.PreviewSizes != nil {
		merged.PreviewSizes = child.PreviewSizes
	}
	return &merged
}

// LoadServerConfig reads a JSON config file and applies environment variable
// overrides for sensitive fields:
//   - GOLLERY_LISTEN_ADDR overrides listen_addr
//   - GOLLERY_POSTGRES_DSN overrides analytics.postgres_dsn_env
//   - GOLLERY_SESSION_SECRET overrides auth.session_secret
func LoadServerConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading server config: %w", err)
	}
	var cfg ServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing server config: %w", err)
	}

	// Environment variable overrides.
	if v := os.Getenv("GOLLERY_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("GOLLERY_POSTGRES_DSN"); v != "" {
		if cfg.Analytics == nil {
			cfg.Analytics = &GlobalAnalyticsConfig{}
		}
		cfg.Analytics.PostgresDSNEnv = v
	}
	if v := os.Getenv("GOLLERY_SESSION_SECRET"); v != "" {
		if cfg.Auth == nil {
			cfg.Auth = &AuthConfig{}
		}
		cfg.Auth.SessionSecret = v
	}

	return &cfg, nil
}

// Validate checks server config for required fields.
func (c *ServerConfig) Validate() error {
	var errs []error
	if c.ContentRoot == "" {
		errs = append(errs, fmt.Errorf("content_root is required"))
	}
	if c.CacheDir == "" {
		errs = append(errs, fmt.Errorf("cache_dir is required"))
	}
	if c.ListenAddr == "" {
		errs = append(errs, fmt.Errorf("listen_addr is required"))
	}
	if c.Auth != nil {
		if c.Auth.Provider == "" {
			errs = append(errs, fmt.Errorf("auth.provider is required when auth is configured"))
		}
		if c.Auth.SessionSecret == "" {
			errs = append(errs, fmt.Errorf("auth.session_secret is required when auth is configured"))
		}
	}
	if c.Analytics != nil && c.Analytics.Enabled {
		if c.Analytics.Backend == "" {
			errs = append(errs, fmt.Errorf("analytics.backend is required when analytics is enabled"))
		}
		if c.Analytics.Backend == "postgres" && c.Analytics.PostgresDSNEnv == "" {
			errs = append(errs, fmt.Errorf("analytics.postgres_dsn_env is required for postgres backend"))
		}
	}
	return errors.Join(errs...)
}
