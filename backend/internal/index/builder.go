// Package index builds the in-memory [domain.Snapshot] from scanner output
// and sidecar state.
//
// # How it works
//
// [BuildSnapshot] takes the content root path and a [fswalk.ScanResult]
// (which contains discovered albums and their assets) and produces a
// [domain.Snapshot]:
//
//  1. For each scanned album, it calls [state.EnsureAlbumID] to load or
//     create the album's stable ID from the sidecar file.
//  2. For each asset, it calls [state.EnsureAssetID] similarly, and loads
//     any per-asset ACL overrides from the sidecar.
//  3. It assembles [domain.Album] and [domain.Asset] objects and stores
//     them in the snapshot's Albums map (keyed by relative path).
//
// # Memory model
//
// The resulting Snapshot is a fully self-contained, read-only data structure.
// It holds the complete album tree with all assets and their metadata.
// The API server stores one Snapshot at a time behind a [sync.RWMutex]:
// reads (all API requests) take the read lock, while re-indexing takes the
// write lock to swap in a new Snapshot.
//
// There is no incremental update mechanism. Every re-index rebuilds the
// full Snapshot from a fresh filesystem scan. This is simple and avoids
// consistency issues but means rebuild time is proportional to the total
// number of albums and assets.
//
// # Sidecar side effects
//
// [BuildSnapshot] writes sidecar state files for albums and assets that
// don't have stable IDs yet. This is the only place the server writes to
// the content tree (aside from admin-triggered state updates). The writes
// use temp-file + rename for atomicity.
package index

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/perrito666/gollery/backend/internal/domain"
	"github.com/perrito666/gollery/backend/internal/fswalk"
	"github.com/perrito666/gollery/backend/internal/state"
)

// BuildSnapshot combines scanner output with sidecar state to produce
// a point-in-time Snapshot with stable IDs and album hierarchy.
func BuildSnapshot(contentRoot string, scan *fswalk.ScanResult) (*domain.Snapshot, error) {
	snap := &domain.Snapshot{
		GeneratedAt: time.Now(),
		Albums:      make(map[string]*domain.Album, len(scan.Albums)),
	}

	for relPath, scanned := range scan.Albums {
		absPath := filepath.Join(contentRoot, relPath)

		// Ensure album has a stable ID.
		albumState, _, err := state.EnsureAlbumID(absPath)
		if err != nil {
			return nil, fmt.Errorf("ensuring album ID for %q: %w", relPath, err)
		}

		// Resolve title and description from config.
		var title, description string
		if scanned.Config != nil {
			title = scanned.Config.Title
			description = scanned.Config.Description
		}

		// Determine parent path.
		parentPath := ""
		if relPath != "" {
			parentPath = filepath.Dir(relPath)
			if parentPath == "." {
				parentPath = ""
			}
		}

		// Build assets with stable IDs.
		assets := make([]domain.Asset, 0, len(scanned.Assets))
		for _, sa := range scanned.Assets {
			assetState, _, err := state.EnsureAssetID(absPath, sa.Filename)
			if err != nil {
				return nil, fmt.Errorf("ensuring asset ID for %q in %q: %w", sa.Filename, relPath, err)
			}
			asset := domain.Asset{
				ID:        assetState.ObjectID,
				Filename:  sa.Filename,
				AlbumPath: relPath,
				ModTime:   sa.ModTime,
				SizeBytes: sa.SizeBytes,
			}
			if assetState.AccessOverride != nil {
				asset.Access = &domain.AccessOverride{
					View:          assetState.AccessOverride.View,
					AllowedUsers:  assetState.AccessOverride.AllowedUsers,
					AllowedGroups: assetState.AccessOverride.AllowedGroups,
				}
			}
			assets = append(assets, asset)
		}

		album := &domain.Album{
			ID:          albumState.ObjectID,
			Path:        relPath,
			Title:       title,
			Description: description,
			ParentPath:  parentPath,
			Children:    scanned.ChildPaths,
			Assets:      assets,
		}
		snap.Albums[relPath] = album
	}

	return snap, nil
}
