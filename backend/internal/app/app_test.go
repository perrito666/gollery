package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/perrito666/gollery/backend/internal/api"
	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/fswalk"
	"github.com/perrito666/gollery/backend/internal/index"
)

func TestExtractConfigs(t *testing.T) {
	cfg := &config.AlbumConfig{Title: "Test"}
	scan := &fswalk.ScanResult{
		Albums: map[string]*fswalk.ScannedAlbum{
			"":       {Path: "", Config: nil},
			"photos": {Path: "photos", Config: cfg},
		},
	}

	configs := extractConfigs(scan)
	if len(configs) != 1 {
		t.Fatalf("got %d configs, want 1", len(configs))
	}
	if configs["photos"].Title != "Test" {
		t.Errorf("title = %q, want Test", configs["photos"].Title)
	}
}

func TestRun_InvalidConfig(t *testing.T) {
	// Non-existent config path.
	err := Run(context.Background(), "/nonexistent/config.json")
	if err == nil {
		t.Fatal("expected error for non-existent config")
	}
}

func TestRun_InvalidConfigContent(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	err := Run(context.Background(), cfgPath)
	if err == nil {
		t.Fatal("expected validation error for empty config")
	}
}

func TestRun_ValidConfig_ScanFails(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	cfg := config.ServerConfig{
		ContentRoot: "/nonexistent/content/root",
		CacheDir:    filepath.Join(dir, "cache"),
		ListenAddr:  ":0",
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	err := Run(context.Background(), cfgPath)
	if err == nil {
		t.Fatal("expected error for non-existent content root")
	}
}

func TestRun_GracefulShutdown(t *testing.T) {
	dir := t.TempDir()
	contentRoot := filepath.Join(dir, "content")
	cacheDir := filepath.Join(dir, "cache")
	if err := os.MkdirAll(contentRoot, 0755); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(dir, "config.json")
	cfg := config.ServerConfig{
		ContentRoot: contentRoot,
		CacheDir:    cacheDir,
		ListenAddr:  "127.0.0.1:0",
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, cfgPath)
	}()

	// Cancel immediately to trigger shutdown.
	cancel()

	err := <-errCh
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupAuth_UnsupportedProvider(t *testing.T) {
	dir := t.TempDir()
	srv := createTestServer(t, dir)
	cfg := &config.ServerConfig{
		Auth: &config.AuthConfig{
			Provider:      "oidc",
			SessionSecret: "test-secret",
		},
	}

	err := setupAuth(srv, cfg, "gollery.json")
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func createTestServer(t *testing.T, contentRoot string) *api.Server {
	t.Helper()
	scan, err := fswalk.Scan(contentRoot)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := index.BuildSnapshot(contentRoot, scan)
	if err != nil {
		t.Fatal(err)
	}
	return api.NewServer(snap, extractConfigs(scan))
}

func TestAnalyticsIPSalt_ConfiguredSaltWins(t *testing.T) {
	cfg := &config.ServerConfig{
		Auth:      &config.AuthConfig{SessionSecret: "session-secret"},
		Analytics: &config.GlobalAnalyticsConfig{IPSalt: "explicit-salt"},
	}

	if got := analyticsIPSalt(cfg); got != "explicit-salt" {
		t.Errorf("salt = %q, want the configured ip_salt", got)
	}
}

func TestAnalyticsIPSalt_DerivedFromSessionSecret(t *testing.T) {
	cfg := &config.ServerConfig{
		Auth: &config.AuthConfig{SessionSecret: "session-secret"},
	}

	got := analyticsIPSalt(cfg)
	if got == "" {
		t.Fatal("salt is empty")
	}
	// Stable across calls, so visitor hashes survive a restart.
	if second := analyticsIPSalt(cfg); second != got {
		t.Errorf("salt changed between calls: %q then %q", got, second)
	}
	// Derived, not the secret itself.
	if got == "session-secret" {
		t.Error("salt is the session secret verbatim")
	}

	other := analyticsIPSalt(&config.ServerConfig{
		Auth: &config.AuthConfig{SessionSecret: "another-secret"},
	})
	if other == got {
		t.Error("different session secrets produced the same salt")
	}
}

func TestAnalyticsIPSalt_RandomWithoutSecret(t *testing.T) {
	cfg := &config.ServerConfig{}

	first := analyticsIPSalt(cfg)
	if first == "" {
		t.Fatal("salt is empty")
	}
	if second := analyticsIPSalt(cfg); second == first {
		t.Error("fallback salt is not random")
	}
}
