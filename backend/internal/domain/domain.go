// Package domain defines core types shared across the backend.
//
// These types are the lingua franca between packages. They carry no behavior
// beyond being data containers. Packages like [fswalk], [index], [access],
// and [api] all operate on domain types.
//
// # Key types
//
//   - [Snapshot] — a point-in-time view of the entire gallery, built by
//     [index.BuildSnapshot]. Held in memory by the API server and swapped
//     atomically on re-index.
//   - [Album] — a discovered album with its stable ID, path, children,
//     and assets.
//   - [Asset] — an image file with its stable ID, filename, and optional
//     per-asset ACL override.
//   - [Principal] — an authenticated user, carried in request context.
//     A nil Principal means anonymous access.
//   - [AccessOverride] — per-asset ACL fields that override the album-level
//     access config. Loaded from sidecar state files.
//
// # Identity model
//
// Albums and assets have stable IDs (alb_<hex>, ast_<hex>) that are
// persisted in sidecar state files. IDs are not derived from paths because
// paths can change (rename/move), but discussions and analytics need stable
// references. IDs are generated once by the [state] package using
// crypto/rand and reused on subsequent scans.
package domain

import "time"

// Album represents a discovered album in the content tree.
type Album struct {
	// ID is the stable object identifier (e.g. "alb_01J...").
	// Assigned by the state layer, not derived from paths.
	ID string

	// Path is the filesystem path relative to the content root.
	Path string

	// Title from the resolved (merged) album config.
	Title string

	// Description from the resolved album config.
	Description string

	// ParentPath is the relative path of the parent album, empty for root.
	ParentPath string

	// Children contains relative paths of direct child albums.
	Children []string

	// Assets contains the assets discovered in this album.
	Assets []Asset
}

// Asset represents an image file discovered in an album.
type Asset struct {
	// ID is the stable object identifier (e.g. "ast_01J...").
	ID string

	// Filename is the base name of the file.
	Filename string

	// AlbumPath is the relative path of the containing album.
	AlbumPath string

	// ModTime is the last modification time from the filesystem.
	ModTime time.Time

	// SizeBytes is the file size in bytes.
	SizeBytes int64

	// Access holds per-asset ACL overrides from sidecar state.
	// If nil, the containing album's ACL applies.
	Access *AccessOverride

	// Metadata holds extracted EXIF/image metadata.
	// May be nil if metadata extraction was not performed.
	Metadata *ImageMetadata
}

// ImageMetadata holds extracted image metadata (EXIF, dimensions, etc.).
type ImageMetadata struct {
	CameraMake  string     `json:"camera_make,omitempty"`
	CameraModel string     `json:"camera_model,omitempty"`
	DateTaken   *time.Time `json:"date_taken,omitempty"`
	Width       int        `json:"width,omitempty"`
	Height      int        `json:"height,omitempty"`
	Orientation int        `json:"orientation,omitempty"`
	Latitude    *float64   `json:"latitude,omitempty"`
	Longitude   *float64   `json:"longitude,omitempty"`
}

// AccessOverride holds per-asset ACL overrides.
type AccessOverride struct {
	View          string
	AllowedUsers  []string
	AllowedGroups []string
}

// Snapshot represents a point-in-time view of the entire gallery.
type Snapshot struct {
	// GeneratedAt is when this snapshot was built.
	GeneratedAt time.Time

	// Albums is the full set of discovered albums keyed by relative path.
	Albums map[string]*Album
}

// Principal represents an authenticated user for ACL evaluation.
type Principal struct {
	// Username is the unique identifier of the user.
	Username string

	// Groups lists the groups this user belongs to.
	Groups []string

	// IsAdmin indicates whether this is a global admin.
	IsAdmin bool
}

// DiscussionBinding represents a link between a gallery object and an
// external discussion thread. This is editorial state stored in sidecars.
type DiscussionBinding struct {
	// Provider identifies the discussion service (e.g. "mastodon", "bluesky").
	Provider string `json:"provider"`

	// RemoteID is the provider-specific identifier for the thread.
	RemoteID string `json:"remote_id"`

	// URL is the public URL of the discussion thread.
	URL string `json:"url"`

	// CreatedAt is when the binding was created.
	CreatedAt time.Time `json:"created_at"`

	// CreatedBy is the username that created the binding.
	CreatedBy string `json:"created_by"`

	// ProviderMeta holds provider-specific metadata.
	ProviderMeta map[string]string `json:"provider_meta,omitempty"`
}
