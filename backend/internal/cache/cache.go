// Package cache manages the gallery-cache directory for generated artifacts.
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
