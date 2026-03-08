// Package watch monitors the filesystem for content changes using polling.
//
// # How it works
//
// [Watcher] periodically walks the content tree and compares file
// modification times and sizes against its last scan. When differences
// are detected, the affected directory paths are marked dirty. After a
// configurable debounce delay (to batch rapid changes), the watcher
// calls the [ReconcileFunc] with the list of dirty paths.
//
// The reconciliation function typically triggers a full re-scan and
// snapshot rebuild (via [fswalk.Scan] + [index.BuildSnapshot]), then
// swaps the new snapshot into the API server.
//
// # Why polling
//
// Polling was chosen over inotify/fsnotify because:
//   - It works on all platforms and filesystems (including NFS/CIFS).
//   - No limit on the number of watched directories.
//   - Simpler to reason about correctness.
//
// The tradeoff is latency: changes are detected at the poll interval
// (default 5 seconds) plus the debounce delay (default 2 seconds),
// so worst-case detection is ~7 seconds.
//
// # Memory usage
//
// The watcher maintains a map of every file/directory path to its last
// known modtime and size. This is separate from the API server's snapshot.
// For a tree with 100,000 entries, this map uses roughly 10-15 MB.
//
// # Baseline scan
//
// The first scan establishes a baseline without marking anything dirty.
// This prevents a full reconciliation on server startup (the initial
// snapshot is built directly by the app startup code).
package watch

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ReconcileFunc is called when dirty paths need to be reconciled.
// It receives the set of dirty relative paths.
type ReconcileFunc func(ctx context.Context, dirtyPaths []string) error

// fileState tracks the last known state of a file/directory.
type fileState struct {
	modTime time.Time
	size    int64
	isDir   bool
}

// Watcher polls the content directory for changes and triggers
// reconciliation when modifications are detected.
type Watcher struct {
	contentRoot string
	interval    time.Duration
	debounce    time.Duration
	reconcile   ReconcileFunc

	mu         sync.Mutex
	dirtyPaths map[string]bool
	lastScan   map[string]fileState
	lastDirty  time.Time
	baselined  bool
}

// Config holds watcher configuration.
type Config struct {
	// ContentRoot is the filesystem path to watch.
	ContentRoot string

	// PollInterval is how often to scan for changes.
	PollInterval time.Duration

	// DebounceDelay is how long to wait after the last change
	// before triggering reconciliation.
	DebounceDelay time.Duration

	// Reconcile is called when changes are detected and debounced.
	Reconcile ReconcileFunc
}

// New creates a new filesystem watcher.
func New(cfg Config) *Watcher {
	interval := cfg.PollInterval
	if interval == 0 {
		interval = 5 * time.Second
	}
	debounce := cfg.DebounceDelay
	if debounce == 0 {
		debounce = 2 * time.Second
	}
	return &Watcher{
		contentRoot: cfg.ContentRoot,
		interval:    interval,
		debounce:    debounce,
		reconcile:   cfg.Reconcile,
		dirtyPaths:  make(map[string]bool),
		lastScan:    make(map[string]fileState),
	}
}

// Run starts the watcher loop. It blocks until the context is cancelled.
func (w *Watcher) Run(ctx context.Context) error {
	// Initial scan to establish baseline.
	w.scan()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.scan()
			w.maybeReconcile(ctx)
		}
	}
}

// DirtyPaths returns the current set of dirty paths (for testing).
func (w *Watcher) DirtyPaths() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	paths := make([]string, 0, len(w.dirtyPaths))
	for p := range w.dirtyPaths {
		paths = append(paths, p)
	}
	return paths
}

// scan walks the content directory and detects changes.
func (w *Watcher) scan() {
	current := make(map[string]fileState)

	filepath.WalkDir(w.contentRoot, func(absPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Skip hidden directories.
		if d.IsDir() && d.Name() != "." && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(w.contentRoot, absPath)
		if err != nil {
			return nil
		}
		if relPath == "." {
			relPath = ""
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		current[relPath] = fileState{
			modTime: info.ModTime(),
			size:    info.Size(),
			isDir:   d.IsDir(),
		}
		return nil
	})

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.baselined {
		// First scan establishes the baseline without marking dirty.
		w.lastScan = current
		w.baselined = true
		return
	}

	// Detect new or modified entries.
	for path, state := range current {
		old, exists := w.lastScan[path]
		if !exists || old.modTime != state.modTime || old.size != state.size {
			dir := path
			if !state.isDir {
				dir = filepath.Dir(path)
				if dir == "." {
					dir = ""
				}
			}
			w.dirtyPaths[dir] = true
			w.lastDirty = time.Now()
		}
	}

	// Detect deleted entries.
	for path, state := range w.lastScan {
		if _, exists := current[path]; !exists {
			dir := path
			if !state.isDir {
				dir = filepath.Dir(path)
				if dir == "." {
					dir = ""
				}
			}
			w.dirtyPaths[dir] = true
			w.lastDirty = time.Now()
		}
	}

	w.lastScan = current
}

// maybeReconcile triggers reconciliation if the debounce period has elapsed.
func (w *Watcher) maybeReconcile(ctx context.Context) {
	w.mu.Lock()
	if len(w.dirtyPaths) == 0 {
		w.mu.Unlock()
		return
	}
	if time.Since(w.lastDirty) < w.debounce {
		w.mu.Unlock()
		return
	}

	paths := make([]string, 0, len(w.dirtyPaths))
	for p := range w.dirtyPaths {
		paths = append(paths, p)
	}
	w.dirtyPaths = make(map[string]bool)
	w.mu.Unlock()

	if w.reconcile != nil {
		w.reconcile(ctx, paths)
	}
}

// ScanOnce performs a single scan cycle (for testing without Run loop).
func (w *Watcher) ScanOnce() {
	w.scan()
}

// ForceReconcile triggers reconciliation regardless of debounce (for testing).
func (w *Watcher) ForceReconcile(ctx context.Context) {
	w.mu.Lock()
	w.lastDirty = time.Time{} // Force debounce to pass.
	w.mu.Unlock()
	w.maybeReconcile(ctx)
}

// FullReconcile creates a standard reconciliation function that
// re-scans the content root and calls the given callback with dirty paths.
func FullReconcile(fn func(ctx context.Context, dirtyPaths []string) error) ReconcileFunc {
	return fn
}

// DetectChanges performs a scan and returns the current dirty paths
// without triggering reconciliation. Useful for testing.
func (w *Watcher) DetectChanges() []string {
	w.scan()
	return w.DirtyPaths()
}

// MarkClean is used by tests to simulate a completed reconciliation,
// explicitly removing a set of paths from the dirty set. In production,
// the reconcile callback handles this implicitly via ForceReconcile.
func (w *Watcher) MarkClean(paths ...string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, p := range paths {
		delete(w.dirtyPaths, p)
	}
}

// PendingDir checks if the given directory (relative path) is dirty.
// Exported for external reconcilers or tests that need to know what changed.
func (w *Watcher) PendingDir(dir string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.dirtyPaths[dir]
}

// WatchFile checks the current state of a specific file path
// relative to the content root. Returns true if the file exists.
func WatchFile(contentRoot, relPath string) bool {
	_, err := os.Stat(filepath.Join(contentRoot, relPath))
	return err == nil
}
