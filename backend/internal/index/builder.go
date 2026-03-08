// Package index builds and maintains the in-memory album/asset index.
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
