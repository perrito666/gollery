package config

import (
	"os"
	"path/filepath"
	"testing"
)

func boolPtr(v bool) *bool { return &v }

func TestLoadAlbumConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "album.json")

	data := `{
		"title": "Vacation",
		"description": "Summer 2024",
		"access": {"view": "public"},
		"analytics": {"enabled": true}
	}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadAlbumConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Title != "Vacation" {
		t.Errorf("title = %q, want %q", cfg.Title, "Vacation")
	}
	if cfg.Access == nil || cfg.Access.View != "public" {
		t.Error("access.view should be public")
	}
	if cfg.Analytics == nil || cfg.Analytics.Enabled == nil || !*cfg.Analytics.Enabled {
		t.Error("analytics.enabled should be true")
	}
}

func TestLoadAlbumConfig_NotFound(t *testing.T) {
	_, err := LoadAlbumConfig("/nonexistent/album.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadAlbumConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "album.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAlbumConfig(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAlbumConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AlbumConfig
		wantErr bool
	}{
		{
			name: "valid public",
			cfg:  AlbumConfig{Access: &AccessConfig{View: "public"}},
		},
		{
			name: "valid authenticated",
			cfg:  AlbumConfig{Access: &AccessConfig{View: "authenticated"}},
		},
		{
			name: "valid restricted",
			cfg:  AlbumConfig{Access: &AccessConfig{View: "restricted"}},
		},
		{
			name: "no access config is valid",
			cfg:  AlbumConfig{Title: "Test"},
		},
		{
			name:    "invalid access mode",
			cfg:     AlbumConfig{Access: &AccessConfig{View: "secret"}},
			wantErr: true,
		},
		{
			name: "valid sort_order filename",
			cfg:  AlbumConfig{SortOrder: "filename"},
		},
		{
			name: "valid sort_order date",
			cfg:  AlbumConfig{SortOrder: "date"},
		},
		{
			name: "valid sort_order empty",
			cfg:  AlbumConfig{SortOrder: ""},
		},
		{
			name:    "invalid sort_order",
			cfg:     AlbumConfig{SortOrder: "random"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestMergeAlbumConfigs_NilCases(t *testing.T) {
	cfg := &AlbumConfig{Title: "A"}

	if got := MergeAlbumConfigs(nil, cfg); got != cfg {
		t.Error("merge(nil, child) should return child")
	}
	if got := MergeAlbumConfigs(cfg, nil); got != cfg {
		t.Error("merge(parent, nil) should return parent")
	}
}

func TestMergeAlbumConfigs_ScalarOverride(t *testing.T) {
	parent := &AlbumConfig{Title: "Parent", Description: "Parent desc"}
	child := &AlbumConfig{Title: "Child"}

	merged := MergeAlbumConfigs(parent, child)

	if merged.Title != "Child" {
		t.Errorf("title = %q, want %q", merged.Title, "Child")
	}
	if merged.Description != "Parent desc" {
		t.Errorf("description = %q, want %q", merged.Description, "Parent desc")
	}
}

func TestMergeAlbumConfigs_AccessMergeByKey(t *testing.T) {
	parent := &AlbumConfig{
		Access: &AccessConfig{
			View:         "public",
			AllowedUsers: []string{"alice"},
		},
	}
	child := &AlbumConfig{
		Access: &AccessConfig{
			View: "restricted",
		},
	}

	merged := MergeAlbumConfigs(parent, child)

	if merged.Access.View != "restricted" {
		t.Errorf("view = %q, want %q", merged.Access.View, "restricted")
	}
	// AllowedUsers not overridden by child, so parent's value persists.
	if len(merged.Access.AllowedUsers) != 1 || merged.Access.AllowedUsers[0] != "alice" {
		t.Errorf("allowed_users = %v, want [alice]", merged.Access.AllowedUsers)
	}
}

func TestMergeAlbumConfigs_ListReplacesParent(t *testing.T) {
	parent := &AlbumConfig{
		Access: &AccessConfig{
			AllowedUsers: []string{"alice", "bob"},
		},
	}
	child := &AlbumConfig{
		Access: &AccessConfig{
			AllowedUsers: []string{"charlie"},
		},
	}

	merged := MergeAlbumConfigs(parent, child)

	if len(merged.Access.AllowedUsers) != 1 || merged.Access.AllowedUsers[0] != "charlie" {
		t.Errorf("allowed_users = %v, want [charlie]", merged.Access.AllowedUsers)
	}
}

func TestMergeAlbumConfigs_InheritFalse(t *testing.T) {
	parent := &AlbumConfig{Title: "Parent", Description: "Parent desc"}
	child := &AlbumConfig{Title: "Child", Inherit: boolPtr(false)}

	merged := MergeAlbumConfigs(parent, child)

	if merged.Title != "Child" {
		t.Errorf("title = %q, want %q", merged.Title, "Child")
	}
	if merged.Description != "" {
		t.Errorf("description = %q, want empty (inherit=false)", merged.Description)
	}
}

func TestMergeAlbumConfigs_SortOrderInheritance(t *testing.T) {
	parent := &AlbumConfig{SortOrder: "date"}
	child := &AlbumConfig{Title: "Child"}

	merged := MergeAlbumConfigs(parent, child)
	if merged.SortOrder != "date" {
		t.Errorf("sort_order = %q, want %q (inherited from parent)", merged.SortOrder, "date")
	}

	// Child overrides parent.
	child2 := &AlbumConfig{SortOrder: "filename"}
	merged2 := MergeAlbumConfigs(parent, child2)
	if merged2.SortOrder != "filename" {
		t.Errorf("sort_order = %q, want %q (overridden by child)", merged2.SortOrder, "filename")
	}
}

func TestMergeAlbumConfigs_AnalyticsMerge(t *testing.T) {
	parent := &AlbumConfig{
		Analytics: &AlbumAnalyticsConfig{
			Enabled:         boolPtr(true),
			TrackAlbumViews: boolPtr(true),
		},
	}
	child := &AlbumConfig{
		Analytics: &AlbumAnalyticsConfig{
			ExposePopularity: boolPtr(false),
		},
	}

	merged := MergeAlbumConfigs(parent, child)

	if merged.Analytics.Enabled == nil || !*merged.Analytics.Enabled {
		t.Error("analytics.enabled should be inherited as true")
	}
	if merged.Analytics.ExposePopularity == nil || *merged.Analytics.ExposePopularity {
		t.Error("analytics.expose_popularity should be false from child")
	}
}

func TestMergeAlbumConfigs_DerivativesListReplace(t *testing.T) {
	parent := &AlbumConfig{
		Derivatives: &DerivativesConfig{
			ThumbnailSizes: []int{200, 400},
			PreviewSizes:   []int{1600},
		},
	}
	child := &AlbumConfig{
		Derivatives: &DerivativesConfig{
			ThumbnailSizes: []int{300},
		},
	}

	merged := MergeAlbumConfigs(parent, child)

	if len(merged.Derivatives.ThumbnailSizes) != 1 || merged.Derivatives.ThumbnailSizes[0] != 300 {
		t.Errorf("thumbnail_sizes = %v, want [300]", merged.Derivatives.ThumbnailSizes)
	}
	// PreviewSizes not overridden, inherits.
	if len(merged.Derivatives.PreviewSizes) != 1 || merged.Derivatives.PreviewSizes[0] != 1600 {
		t.Errorf("preview_sizes = %v, want [1600]", merged.Derivatives.PreviewSizes)
	}
}

func TestServerConfigValidate(t *testing.T) {
	valid := ServerConfig{
		ContentRoot: "/data/photos",
		CacheDir:    "/data/cache",
		ListenAddr:  ":8080",
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid config should pass: %v", err)
	}

	empty := ServerConfig{}
	if err := empty.Validate(); err == nil {
		t.Error("empty config should fail validation")
	}
}

func TestServerConfigValidate_Auth(t *testing.T) {
	cfg := ServerConfig{
		ContentRoot: "/data",
		CacheDir:    "/cache",
		ListenAddr:  ":8080",
		Auth:        &AuthConfig{Provider: "static", SessionSecret: "secret123"},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid auth config should pass: %v", err)
	}

	cfg.Auth.Provider = ""
	if err := cfg.Validate(); err == nil {
		t.Error("missing auth.provider should fail")
	}

	cfg.Auth.Provider = "static"
	cfg.Auth.SessionSecret = ""
	if err := cfg.Validate(); err == nil {
		t.Error("missing auth.session_secret should fail")
	}
}

func TestServerConfigValidate_Analytics(t *testing.T) {
	cfg := ServerConfig{
		ContentRoot: "/data",
		CacheDir:    "/cache",
		ListenAddr:  ":8080",
		Analytics: &GlobalAnalyticsConfig{
			Enabled:        true,
			Backend:        "postgres",
			PostgresDSNEnv: "postgres://localhost/test",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid analytics config should pass: %v", err)
	}

	cfg.Analytics.Backend = ""
	if err := cfg.Validate(); err == nil {
		t.Error("missing analytics.backend should fail when enabled")
	}

	cfg.Analytics.Backend = "postgres"
	cfg.Analytics.PostgresDSNEnv = ""
	if err := cfg.Validate(); err == nil {
		t.Error("missing postgres_dsn_env should fail for postgres backend")
	}
}

func TestLoadServerConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := `{
		"content_root": "/photos",
		"cache_dir": "/cache",
		"listen_addr": ":8080",
		"auth": {"provider": "static", "session_secret": "s3cret"},
		"analytics": {
			"enabled": true,
			"backend": "postgres",
			"postgres_dsn_env": "postgres://localhost/gallery",
			"dedup_window_seconds": 300,
			"retain_events_days": 90
		}
	}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadServerConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ContentRoot != "/photos" {
		t.Errorf("content_root = %q", cfg.ContentRoot)
	}
	if cfg.Auth == nil || cfg.Auth.Provider != "static" {
		t.Error("auth.provider should be static")
	}
	if cfg.Auth.SessionSecret != "s3cret" {
		t.Errorf("session_secret = %q", cfg.Auth.SessionSecret)
	}
	if cfg.Analytics.DedupWindowSecs != 300 {
		t.Errorf("dedup_window_seconds = %d", cfg.Analytics.DedupWindowSecs)
	}
	if cfg.Analytics.RetainEventsDays != 90 {
		t.Errorf("retain_events_days = %d", cfg.Analytics.RetainEventsDays)
	}
}

func TestLoadServerConfig_NotFound(t *testing.T) {
	_, err := LoadServerConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadServerConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{bad"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadServerConfig(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadServerConfig_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := `{
		"content_root": "/photos",
		"cache_dir": "/cache",
		"listen_addr": ":8080"
	}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GOLLERY_LISTEN_ADDR", ":9090")
	t.Setenv("GOLLERY_POSTGRES_DSN", "postgres://override/db")
	t.Setenv("GOLLERY_SESSION_SECRET", "env-secret")

	cfg, err := LoadServerConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("listen_addr = %q, want :9090", cfg.ListenAddr)
	}
	if cfg.Analytics == nil || cfg.Analytics.PostgresDSNEnv != "postgres://override/db" {
		t.Error("GOLLERY_POSTGRES_DSN should override analytics.postgres_dsn_env")
	}
	if cfg.Auth == nil || cfg.Auth.SessionSecret != "env-secret" {
		t.Error("GOLLERY_SESSION_SECRET should override auth.session_secret")
	}
}
