// Package cache manages the gallery-cache directory layout for generated
// image derivatives (thumbnails and previews).
//
// # Directory structure
//
// The cache lives outside the content tree to keep source directories clean.
// A typical layout:
//
//	<cache-root>/
//	├── thumbs/          # thumbnails (small, square-ish images for grids)
//	│   ├── ast_a1b2c3_400.jpg
//	│   └── ast_d4e5f6_400.jpg
//	└── previews/        # previews (larger images for detail views)
//	    ├── ast_a1b2c3_1600.jpg
//	    └── ast_d4e5f6_1600.jpg
//
// The cache root is configured via [config.ServerConfig].DerivativeCacheDir
// and defaults to ".gallery-cache" relative to the content root.
//
// # Filename convention
//
// Every cached file is named <assetID>_<size>.jpg, where assetID is the
// stable sidecar ID (e.g. "ast_a1b2c3") and size is the longest-edge pixel
// count requested by the client. This naming scheme means a single asset can
// have multiple cached sizes (e.g. ast_x_200.jpg and ast_x_400.jpg).
//
// # Cache lifecycle
//
//   - Generation: the [derive] package calls [Layout.ThumbPath] or
//     [Layout.PreviewPath] to obtain the expected output path, checks
//     [Exists], and writes the file only on a miss.
//   - Eviction: [PurgeOrphans] scans both subdirectories and removes any
//     file whose asset ID prefix is not in the supplied known-ID set.
//     This is called after a re-index to clean up derivatives for deleted
//     or renamed assets.
//   - No TTL: cached files are valid indefinitely because source images
//     are immutable. A changed source image gets a new asset ID (new
//     sidecar entry), so old cache entries become orphans and are purged.
//
// # Path safety
//
// All paths are constructed by [Layout] methods using [filepath.Join] on
// the configured root plus a filename built from the asset ID and size.
// Asset IDs originate from the sidecar state layer (format "ast_<hex>")
// and are never derived from user input. The API layer looks up asset IDs
// from an in-memory index keyed by the URL path parameter; it never passes
// raw user strings into cache path construction.
//
// # Concurrency
//
// The cache directory is written to by HTTP handlers under the server's
// read-lock (multiple concurrent requests may generate derivatives for
// different assets simultaneously). File creation uses [os.Create] which
// is atomic at the filesystem level for distinct paths. Two concurrent
// requests for the same derivative may race, but both produce identical
// output so the last writer wins harmlessly.
package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

// Layout defines the cache directory structure.
type Layout struct {
	Root string
}

// NewLayout creates a cache layout rooted at the given directory.
func NewLayout(root string) *Layout {
	return &Layout{Root: root}
}

// ThumbDir returns the path to the thumbnails directory.
func (l *Layout) ThumbDir() string {
	return filepath.Join(l.Root, "thumbs")
}

// PreviewDir returns the path to the previews directory.
func (l *Layout) PreviewDir() string {
	return filepath.Join(l.Root, "previews")
}

// ThumbPath returns the cache path for a thumbnail.
func (l *Layout) ThumbPath(assetID string, size int) string {
	return filepath.Join(l.ThumbDir(), fmt.Sprintf("%s_%d.jpg", assetID, size))
}

// PreviewPath returns the cache path for a preview.
func (l *Layout) PreviewPath(assetID string, size int) string {
	return filepath.Join(l.PreviewDir(), fmt.Sprintf("%s_%d.jpg", assetID, size))
}

// EnsureDirs creates all cache subdirectories if they don't exist.
func (l *Layout) EnsureDirs() error {
	for _, dir := range []string{l.ThumbDir(), l.PreviewDir()} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating cache dir %s: %w", dir, err)
		}
	}
	return nil
}

// Exists checks if a cached file exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// PurgeOrphans removes cached derivative files whose asset IDs are not
// in knownAssetIDs. Returns the number of files removed.
func PurgeOrphans(layout *Layout, knownAssetIDs map[string]bool) (int, error) {
	removed := 0
	for _, dir := range []string{layout.ThumbDir(), layout.PreviewDir()} {
		n, err := purgeDir(dir, knownAssetIDs)
		if err != nil {
			return removed, err
		}
		removed += n
	}
	return removed, nil
}

// purgeDir removes files from dir whose asset ID prefix is not in known.
func purgeDir(dir string, known map[string]bool) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("reading cache dir %s: %w", dir, err)
	}

	removed := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		assetID := extractAssetID(name)
		if assetID == "" || known[assetID] {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
			return removed, fmt.Errorf("removing %s: %w", name, err)
		}
		removed++
	}
	return removed, nil
}

// extractAssetID extracts the asset ID from a cache filename like "ast_abc123_400.jpg".
// It finds the last underscore before the extension and returns everything before it.
func extractAssetID(name string) string {
	// Strip extension.
	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]
	// Find last underscore (separating asset ID from size).
	idx := lastIndex(base, '_')
	if idx <= 0 {
		return ""
	}
	return base[:idx]
}

func lastIndex(s string, b byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == b {
			return i
		}
	}
	return -1
}
