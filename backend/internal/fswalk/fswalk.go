// Package fswalk implements filesystem scanning and album discovery.
package fswalk

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perrito666/gollery/backend/internal/config"
)

// ImageExtensions lists file extensions recognized as image assets.
var ImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".tiff": true,
	".bmp":  true,
}

// ScannedAsset holds basic info about a discovered image file.
type ScannedAsset struct {
	Filename  string
	ModTime   time.Time
	SizeBytes int64
}

// ScannedAlbum is a folder discovered during scanning that belongs
// to a published subtree.
type ScannedAlbum struct {
	// Path is relative to the content root.
	Path string

	// Config is the resolved (merged) album configuration.
	Config *config.AlbumConfig

	// Assets lists the image files found in this folder.
	Assets []ScannedAsset

	// ChildPaths lists relative paths of direct child albums.
	ChildPaths []string
}

// ScanResult holds the output of a content tree scan.
type ScanResult struct {
	// Albums keyed by relative path.
	Albums map[string]*ScannedAlbum

	// Errors collects non-fatal issues encountered during scanning.
	Errors []ScanError
}

// ScanError records a non-fatal issue found during scanning.
type ScanError struct {
	Path string
	Err  error
}

// Scan walks the content root and discovers published albums and their assets.
// It applies config inheritance following the merge rules from the design doc.
// Folders outside published subtrees are silently ignored.
func Scan(contentRoot string) (*ScanResult, error) {
	result := &ScanResult{
		Albums: make(map[string]*ScannedAlbum),
	}

	// resolved tracks the merged config for each published path.
	resolved := make(map[string]*config.AlbumConfig)

	err := filepath.WalkDir(contentRoot, func(absPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		// Skip hidden directories (like .gallery, .git).
		if d.Name() != "." && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(contentRoot, absPath)
		if err != nil {
			return err
		}
		if relPath == "." {
			relPath = ""
		}

		// Try loading album.json in this directory.
		albumJSONPath := filepath.Join(absPath, "album.json")
		localCfg, loadErr := config.LoadAlbumConfig(albumJSONPath)
		if loadErr != nil && !os.IsNotExist(loadErr) {
			// Config exists but is invalid — record error, treat as absent.
			result.Errors = append(result.Errors, ScanError{
				Path: relPath,
				Err:  loadErr,
			})
			localCfg = nil
		}

		// Find the parent's resolved config.
		var parentCfg *config.AlbumConfig
		if relPath != "" {
			parentPath := filepath.Dir(relPath)
			if parentPath == "." {
				parentPath = ""
			}
			parentCfg = resolved[parentPath]
		}

		// Determine the resolved config for this folder.
		var cfg *config.AlbumConfig
		switch {
		case localCfg != nil && parentCfg != nil:
			cfg = config.MergeAlbumConfigs(parentCfg, localCfg)
		case localCfg != nil:
			cfg = localCfg
		case parentCfg != nil:
			cfg = parentCfg
		default:
			// No album.json here or above — not in a published subtree.
			return nil
		}

		if err := cfg.Validate(); err != nil {
			result.Errors = append(result.Errors, ScanError{Path: relPath, Err: err})
			// Use parent config if available, otherwise skip.
			if parentCfg != nil {
				cfg = parentCfg
			} else {
				return nil
			}
		}

		resolved[relPath] = cfg

		// Scan for image assets in this directory.
		assets, assetErr := scanAssets(absPath)
		if assetErr != nil {
			result.Errors = append(result.Errors, ScanError{Path: relPath, Err: assetErr})
		}

		album := &ScannedAlbum{
			Path:   relPath,
			Config: cfg,
			Assets: assets,
		}
		result.Albums[relPath] = album

		// Register as child of parent.
		if relPath != "" {
			parentPath := filepath.Dir(relPath)
			if parentPath == "." {
				parentPath = ""
			}
			if parent, ok := result.Albums[parentPath]; ok {
				parent.ChildPaths = append(parent.ChildPaths, relPath)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// scanAssets reads a directory and returns recognized image files.
func scanAssets(dirPath string) ([]ScannedAsset, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var assets []ScannedAsset
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !ImageExtensions[ext] {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		assets = append(assets, ScannedAsset{
			Filename:  e.Name(),
			ModTime:   info.ModTime(),
			SizeBytes: info.Size(),
		})
	}
	return assets, nil
}
