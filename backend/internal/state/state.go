// Package state manages mutable sidecar state in .gallery/ directories.
//
// # What sidecar state is
//
// Sidecar state is data the server persists alongside the content tree
// to track things that aren't part of the source images or album.json:
//
//   - Stable object IDs (alb_<hex>, ast_<hex>) — needed because filesystem
//     paths can change, but discussions and analytics need permanent refs.
//   - Discussion bindings — links to external threads (Mastodon, Bluesky).
//   - Per-asset ACL overrides — allow individual images to have different
//     visibility than their containing album.
//
// # File layout
//
// Each album directory may have a .gallery/ subdirectory:
//
//	<album>/
//	├── album.json              (declarative config, never written by server)
//	├── image1.jpg
//	├── image2.jpg
//	└── .gallery/
//	    ├── album.state.json    (album ID, discussion bindings)
//	    └── assets/
//	        ├── image1.jpg.json (asset ID, ACL override, discussion bindings)
//	        └── image2.jpg.json
//
// # Atomicity
//
// All writes use [atomicWriteJSON]: marshal to temp file in the same
// directory, then [os.Rename]. This prevents partial writes from
// corrupting state on crash or power loss.
//
// # ID generation
//
// IDs are 16 random bytes from crypto/rand, hex-encoded, with a prefix
// (alb_ or ast_). [EnsureAlbumID] and [EnsureAssetID] are idempotent:
// they load existing state if present, or generate and persist a new ID
// if absent. This means the first scan of new content creates sidecar
// files, and subsequent scans reuse them.
//
// # Important: the server writes here
//
// The .gallery/ directories are the only place the server writes into the
// content tree. album.json is strictly read-only (declarative config owned
// by the user). This separation is a core invariant of the architecture.
package state

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	galleryDir     = ".gallery"
	albumStateFile = "album.state.json"
	assetsDir      = "assets"
	albumIDPrefix  = "alb_"
	assetIDPrefix  = "ast_"
	idRandomBytes  = 16
)

// AlbumState holds the mutable editorial state for an album.
type AlbumState struct {
	ObjectID    string              `json:"object_id"`
	Discussions []DiscussionBinding `json:"discussions,omitempty"`
}

// AssetState holds the mutable editorial state for an asset.
type AssetState struct {
	ObjectID       string              `json:"object_id"`
	Title          string              `json:"title,omitempty"`
	Description    string              `json:"description,omitempty"`
	Discussions    []DiscussionBinding `json:"discussions,omitempty"`
	AccessOverride *AccessOverride     `json:"access_override,omitempty"`
}

// AccessOverride stores per-asset ACL overrides in sidecar state.
type AccessOverride struct {
	View          string   `json:"view,omitempty"`
	AllowedUsers  []string `json:"allowed_users,omitempty"`
	AllowedGroups []string `json:"allowed_groups,omitempty"`
}

// DiscussionBinding represents a link between a gallery object and an
// external discussion thread. Stored in sidecar state.
type DiscussionBinding struct {
	Provider     string            `json:"provider"`
	RemoteID     string            `json:"remote_id"`
	URL          string            `json:"url"`
	CreatedAt    string            `json:"created_at"`
	CreatedBy    string            `json:"created_by"`
	ProviderMeta map[string]string `json:"provider_meta,omitempty"`
}

// GenerateAlbumID creates a new stable album identifier.
func GenerateAlbumID() (string, error) {
	return generateID(albumIDPrefix)
}

// GenerateAssetID creates a new stable asset identifier.
func GenerateAssetID() (string, error) {
	return generateID(assetIDPrefix)
}

func generateID(prefix string) (string, error) {
	b := make([]byte, idRandomBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating ID: %w", err)
	}
	return prefix + hex.EncodeToString(b), nil
}

// GalleryDir returns the path to the .gallery directory for a given album path.
func GalleryDir(albumAbsPath string) string {
	return filepath.Join(albumAbsPath, galleryDir)
}

// LoadAlbumState reads the album state from .gallery/album.state.json.
// Returns nil without error if the file does not exist.
func LoadAlbumState(albumAbsPath string) (*AlbumState, error) {
	path := filepath.Join(albumAbsPath, galleryDir, albumStateFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading album state: %w", err)
	}
	var s AlbumState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing album state: %w", err)
	}
	return &s, nil
}

// SaveAlbumState writes the album state atomically to .gallery/album.state.json.
func SaveAlbumState(albumAbsPath string, s *AlbumState) error {
	dir := filepath.Join(albumAbsPath, galleryDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating .gallery dir: %w", err)
	}
	path := filepath.Join(dir, albumStateFile)
	return atomicWriteJSON(path, s)
}

// LoadAssetState reads the asset state from .gallery/assets/<filename>.json.
// Returns nil without error if the file does not exist.
func LoadAssetState(albumAbsPath, filename string) (*AssetState, error) {
	path := filepath.Join(albumAbsPath, galleryDir, assetsDir, filename+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading asset state: %w", err)
	}
	var s AssetState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing asset state: %w", err)
	}
	return &s, nil
}

// SaveAssetState writes the asset state atomically to .gallery/assets/<filename>.json.
func SaveAssetState(albumAbsPath, filename string, s *AssetState) error {
	dir := filepath.Join(albumAbsPath, galleryDir, assetsDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating .gallery/assets dir: %w", err)
	}
	path := filepath.Join(dir, filename+".json")
	return atomicWriteJSON(path, s)
}

// EnsureAlbumID loads existing album state or creates a new one with a fresh ID.
// Returns the state (possibly newly created) and whether it was newly created.
func EnsureAlbumID(albumAbsPath string) (*AlbumState, bool, error) {
	s, err := LoadAlbumState(albumAbsPath)
	if err != nil {
		return nil, false, err
	}
	if s != nil && s.ObjectID != "" {
		return s, false, nil
	}

	id, err := GenerateAlbumID()
	if err != nil {
		return nil, false, err
	}
	s = &AlbumState{ObjectID: id}
	if err := SaveAlbumState(albumAbsPath, s); err != nil {
		return nil, false, err
	}
	return s, true, nil
}

// EnsureAssetID loads existing asset state or creates a new one with a fresh ID.
// Returns the state (possibly newly created) and whether it was newly created.
func EnsureAssetID(albumAbsPath, filename string) (*AssetState, bool, error) {
	s, err := LoadAssetState(albumAbsPath, filename)
	if err != nil {
		return nil, false, err
	}
	if s != nil && s.ObjectID != "" {
		return s, false, nil
	}

	id, err := GenerateAssetID()
	if err != nil {
		return nil, false, err
	}
	s = &AssetState{ObjectID: id}
	if err := SaveAssetState(albumAbsPath, filename, s); err != nil {
		return nil, false, err
	}
	return s, true, nil
}

// atomicWriteJSON marshals v as indented JSON and writes it atomically
// by writing to a temp file then renaming.
func atomicWriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}
