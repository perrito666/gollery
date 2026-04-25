// Package fswalk implements filesystem scanning and album discovery.
//
// # Scanning algorithm
//
// [Scan] walks the content root directory tree using [filepath.WalkDir]
// and discovers published albums:
//
//  1. For each directory, it tries to load album.json. If found, the
//     directory (and all descendants) are part of a published subtree.
//  2. Config inheritance is applied: child directories without album.json
//     inherit their parent's resolved config. Children with album.json
//     get a merged config (child overrides parent).
//  3. Directories outside any published subtree are silently ignored.
//  4. Hidden directories (starting with ".") are always skipped.
//  5. Image files are recognized by extension (.jpg, .jpeg, .png, .gif,
//     .webp, .tiff, .bmp).
//
// # Output
//
// The result is a [ScanResult] containing a map of [ScannedAlbum] keyed
// by relative path, plus a list of non-fatal errors (invalid configs,
// unreadable directories, etc.). Non-fatal errors are recorded but do not
// stop the scan.
//
// # Performance
//
// Scanning is synchronous and single-threaded. It reads directory listings
// and album.json files but does not open image files. For a content tree
// with 10,000 images across 500 albums, scanning typically completes in
// under a second on local disk. Network filesystems may be slower due to
// metadata latency.
//
// # Relationship to index
//
// The scan result feeds into [index.BuildSnapshot], which loads sidecar
// state (stable IDs, ACL overrides) and produces the final in-memory
// [domain.Snapshot].
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

	// GPXFiles lists absolute paths to .gpx files found in this directory.
	GPXFiles []string
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

		// Scan for image assets and GPX files in this directory.
		assets, gpxFiles, scanErr := scanDir(absPath)
		if scanErr != nil {
			result.Errors = append(result.Errors, ScanError{Path: relPath, Err: scanErr})
		}

		album := &ScannedAlbum{
			Path:     relPath,
			Config:   cfg,
			Assets:   assets,
			GPXFiles: gpxFiles,
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

// scanDir reads a directory and returns recognized image files and GPX file paths.
func scanDir(dirPath string) ([]ScannedAsset, []string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, nil, err
	}

	var assets []ScannedAsset
	var gpxFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".gpx" {
			gpxFiles = append(gpxFiles, filepath.Join(dirPath, e.Name()))
			continue
		}
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
	return assets, gpxFiles, nil
}
