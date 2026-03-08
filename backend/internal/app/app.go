// Package app wires together the application components.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/perrito666/gollery/backend/internal/analytics"
	pganalytics "github.com/perrito666/gollery/backend/internal/analytics/postgres"
	"github.com/perrito666/gollery/backend/internal/api"
	"github.com/perrito666/gollery/backend/internal/auth"
	"github.com/perrito666/gollery/backend/internal/cache"
	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/fswalk"
	"github.com/perrito666/gollery/backend/internal/index"
	"github.com/perrito666/gollery/backend/internal/logging"
	"github.com/perrito666/gollery/backend/internal/watch"
)

// Run starts the gallery application. It loads configuration, initializes
// all subsystems, starts the HTTP server, and blocks until ctx is cancelled.
func Run(ctx context.Context, configPath string) error {
	// 1. Load and validate configuration.
	cfg, err := config.LoadServerConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 2. Set up structured logging.
	logging.Setup()
	slog.Info("starting gollery", "listen_addr", cfg.ListenAddr, "content_root", cfg.ContentRoot)

	// 3. Initial filesystem scan and snapshot.
	scan, err := fswalk.Scan(cfg.ContentRoot)
	if err != nil {
		return fmt.Errorf("initial scan: %w", err)
	}

	snap, err := index.BuildSnapshot(cfg.ContentRoot, scan)
	if err != nil {
		return fmt.Errorf("building snapshot: %w", err)
	}

	configs := extractConfigs(scan)
	slog.Info("initial scan complete", "albums", len(snap.Albums))

	// 4. Initialize API server.
	srv := api.NewServer(snap, configs)
	cacheLayout := cache.NewLayout(cfg.CacheDir)
	srv.SetContentRoot(cfg.ContentRoot, cacheLayout)

	// Collect scan errors for diagnostics.
	scanErrors := make([]string, len(scan.Errors))
	for i, e := range scan.Errors {
		scanErrors[i] = fmt.Sprintf("%s: %v", e.Path, e.Err)
	}
	srv.SetScanErrors(scanErrors)

	// 5. Initialize auth if configured.
	if cfg.Auth != nil {
		if err := setupAuth(srv, cfg, configPath); err != nil {
			return fmt.Errorf("setting up auth: %w", err)
		}
	}

	// 6. Optionally connect to PostgreSQL for analytics.
	var analyticsStore analytics.Store
	if cfg.Analytics != nil && cfg.Analytics.Enabled {
		store, err := setupAnalytics(ctx, cfg.Analytics)
		if err != nil {
			return fmt.Errorf("setting up analytics: %w", err)
		}
		analyticsStore = store
		defer store.Close()

		// 8. Start retention jobs.
		retainDays := cfg.Analytics.RetainEventsDays
		if retainDays <= 0 {
			retainDays = 90
		}
		analytics.StartRetentionJobs(ctx, store, analytics.RetentionConfig{
			RetainEventsDays: retainDays,
		})
	}

	// 7. Start filesystem watcher.
	reindex := func() error {
		return doReindex(srv, cfg.ContentRoot, cacheLayout)
	}
	srv.SetAdmin(reindex)

	w := watch.New(watch.Config{
		ContentRoot: cfg.ContentRoot,
		Reconcile: func(ctx context.Context, dirtyPaths []string) error {
			slog.Info("reconciling changes", "dirty_paths", len(dirtyPaths))
			return doReindex(srv, cfg.ContentRoot, cacheLayout)
		},
	})
	go func() {
		if err := w.Run(ctx); err != nil && ctx.Err() == nil {
			slog.Error("watcher stopped", "error", err)
		}
	}()

	// Set up analytics recorder if analytics are available.
	if analyticsStore != nil {
		// The analytics recorder uses the analytics.Store directly.
		// The api.AnalyticsStore interface is for query routes and
		// will need an adapter when wired. For now we skip the query
		// adapter as it requires additional implementation.
		_ = analyticsStore
	}

	// 9. Start HTTP server.
	httpSrv := srv.HTTPServer(cfg.ListenAddr, cfg.Timeouts)
	slog.Info("HTTP server starting", "addr", cfg.ListenAddr)

	errCh := make(chan error, 1)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// 10. Wait for shutdown signal.
	select {
	case <-ctx.Done():
		slog.Info("shutting down")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	// Graceful shutdown with 10-second deadline.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	slog.Info("server stopped")
	return nil
}

// doReindex performs a full rescan and updates the API server's snapshot.
// If cacheLayout is non-nil, it purges orphaned derivative cache files.
func doReindex(srv *api.Server, contentRoot string, cacheLayout *cache.Layout) error {
	scan, err := fswalk.Scan(contentRoot)
	if err != nil {
		return fmt.Errorf("rescan: %w", err)
	}

	snap, err := index.BuildSnapshot(contentRoot, scan)
	if err != nil {
		return fmt.Errorf("rebuild snapshot: %w", err)
	}

	configs := extractConfigs(scan)

	// Purge orphaned cache files before updating snapshot.
	if cacheLayout != nil {
		knownIDs := make(map[string]bool)
		for _, album := range snap.Albums {
			for _, asset := range album.Assets {
				knownIDs[asset.ID] = true
			}
		}
		removed, err := cache.PurgeOrphans(cacheLayout, knownIDs)
		if err != nil {
			slog.Error("cache purge failed", "error", err)
		} else if removed > 0 {
			slog.Info("purged orphaned cache files", "count", removed)
		}
	}

	srv.SetSnapshot(snap, configs)

	scanErrors := make([]string, len(scan.Errors))
	for i, e := range scan.Errors {
		scanErrors[i] = fmt.Sprintf("%s: %v", e.Path, e.Err)
	}
	srv.SetScanErrors(scanErrors)

	slog.Info("reindex complete", "albums", len(snap.Albums))
	return nil
}

// extractConfigs pulls album configs from the scan result.
func extractConfigs(scan *fswalk.ScanResult) map[string]*config.AlbumConfig {
	configs := make(map[string]*config.AlbumConfig, len(scan.Albums))
	for path, album := range scan.Albums {
		if album.Config != nil {
			configs[path] = album.Config
		}
	}
	return configs
}

// setupAuth initializes authentication from config.
func setupAuth(srv *api.Server, cfg *config.ServerConfig, configPath string) error {
	// For now, only "static" provider is supported (file-based user store).
	if cfg.Auth.Provider != "static" {
		return fmt.Errorf("unsupported auth provider: %q", cfg.Auth.Provider)
	}

	// Resolve users.json path: explicit config > next to config file > cwd > content root.
	usersPath := cfg.Auth.UsersFile
	if usersPath == "" {
		candidates := []string{
			filepath.Join(filepath.Dir(configPath), "users.json"),
			"users.json",
			filepath.Join(cfg.ContentRoot, "users.json"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				usersPath = c
				break
			}
		}
		if usersPath == "" {
			return fmt.Errorf("users.json not found (searched next to config, cwd, and content root)")
		}
	}

	userStore, err := auth.LoadFileUserStore(usersPath)
	if err != nil {
		return fmt.Errorf("loading user store: %w", err)
	}

	sessions := auth.NewCookieSessionStore(cfg.Auth.SessionSecret)

	var rateLimitCfg *api.RateLimitConfig
	if cfg.Auth.RateLimit != nil {
		rateLimitCfg = &api.RateLimitConfig{
			Rate:  cfg.Auth.RateLimit.Rate,
			Burst: cfg.Auth.RateLimit.Burst,
		}
	}

	srv.SetAuth(userStore, sessions, cfg.Auth.SessionSecret, rateLimitCfg)
	return nil
}

// setupAnalytics connects to PostgreSQL and runs migrations.
func setupAnalytics(ctx context.Context, cfg *config.GlobalAnalyticsConfig) (*pganalytics.Store, error) {
	dsn := cfg.PostgresDSNEnv
	if dsn == "" {
		return nil, fmt.Errorf("postgres DSN is required for analytics")
	}

	pool, err := pganalytics.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	store := pganalytics.New(pool)
	if err := store.Migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	slog.Info("analytics store connected")
	return store, nil
}
