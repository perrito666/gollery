// Package access implements ACL evaluation for albums and assets.
//
// # ACL model
//
// Every album can have an access configuration (set in album.json) with a
// "view" mode that determines who can see the album and its contents:
//
//   - "public" — anyone, including anonymous visitors.
//   - "authenticated" — any logged-in user.
//   - "restricted" — only users explicitly listed in allowed_users,
//     members of allowed_groups, object-level admins, or global admins.
//
// If no access configuration exists (nil), the resource defaults to public.
// Unknown view modes are denied by default for safety.
//
// # Asset-level overrides
//
// Individual assets can override their album's ACL through sidecar state
// files (.gallery/assets/<filename>.json). [EffectiveAssetACL] merges the
// album-level ACL with any per-asset override. Override fields that are
// set replace the album-level values; unset fields inherit from the album.
//
// # In-memory evaluation
//
// All ACL data lives in memory as part of the server's snapshot. The
// evaluation path is:
//
//  1. [index.BuildSnapshot] loads album configs and asset sidecar state
//     from the filesystem into [domain.Snapshot].
//  2. [api.Server.SetSnapshot] rebuilds in-memory indexes (albumsByID,
//     albumsByPath, assetsByID) and stores the config map.
//  3. On each request, the API handler looks up the album/asset, retrieves
//     the config, and calls [CheckView] or [EffectiveAssetACL] + [CheckView].
//
// The entire ACL evaluation is a pure function over in-memory data — there
// are no database queries, no file reads, and no network calls at request
// time. This makes it fast but means the server must be restarted (or
// re-indexed via the admin endpoint) to pick up ACL changes.
//
// # Scaling considerations
//
// Because the full album tree, asset list, and ACL configs are held in
// memory, the server's RAM usage scales linearly with the number of albums
// and assets. Approximate per-object overhead:
//
//   - Each album: ~500 bytes (ID, path, title, description, parent, children slice)
//   - Each asset: ~200 bytes (ID, filename, album path, metadata pointers)
//   - Each album config: ~300 bytes (merged config with access, analytics, etc.)
//
// For a gallery with 1,000 albums and 100,000 assets, this is roughly
// 20–25 MB of heap — well within reason for a single-instance server.
// At 1 million assets the index would use ~200 MB, which is still
// practical on modern hardware but worth monitoring.
//
// The snapshot is rebuilt atomically on re-index: a new Snapshot is
// constructed, then swapped in under a write lock. During the swap,
// in-flight requests continue using the old snapshot (read lock).
// There is no incremental update — every re-index rebuilds everything.
// This is simple and correct but means re-index time scales with the
// total content size (typically seconds for tens of thousands of files).
//
// # Admin bypass
//
// Global admins (Principal.IsAdmin == true) bypass all access restrictions.
// Object-level admins (listed in the ACL's "admins" field) also have full
// access to that specific resource.
package access

import (
	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/domain"
)

// Decision represents the result of an access check.
type Decision int

const (
	// Deny means the principal cannot access the resource.
	Deny Decision = iota
	// Allow means the principal can access the resource.
	Allow
)

// CheckView evaluates whether a principal may view a resource
// protected by the given AccessConfig.
//
// Rules from the technical design (§6):
//   - public: anyone can view
//   - authenticated: any non-nil principal can view
//   - restricted: only allowed users/groups, object admins, or global admins
//   - global admin always has access
//
// A nil principal represents an anonymous (unauthenticated) visitor.
// A nil AccessConfig is treated as public.
func CheckView(acl *config.AccessConfig, p *domain.Principal) Decision {
	// No ACL configured — default to public.
	if acl == nil || acl.View == "" {
		return Allow
	}

	switch acl.View {
	case "public":
		return Allow

	case "authenticated":
		if p != nil {
			return Allow
		}
		return Deny

	case "restricted":
		if p == nil {
			return Deny
		}
		// Global admin always overrides.
		if p.IsAdmin {
			return Allow
		}
		// Check object-level admins.
		if containsStr(acl.Admins, p.Username) {
			return Allow
		}
		// Check allowed users.
		if containsStr(acl.AllowedUsers, p.Username) {
			return Allow
		}
		// Check allowed groups.
		for _, g := range p.Groups {
			if containsStr(acl.AllowedGroups, g) {
				return Allow
			}
		}
		return Deny

	default:
		// Unknown mode — deny by default.
		return Deny
	}
}

// IsObjectAdmin checks whether a principal is an admin for a specific
// resource (listed in the ACL's admins field or a global admin).
func IsObjectAdmin(acl *config.AccessConfig, p *domain.Principal) bool {
	if p == nil {
		return false
	}
	if p.IsAdmin {
		return true
	}
	if acl == nil {
		return false
	}
	return containsStr(acl.Admins, p.Username)
}

// EffectiveAssetACL returns the effective AccessConfig for an asset.
// If the asset has an access override, it takes priority over the album ACL.
func EffectiveAssetACL(albumACL *config.AccessConfig, assetOverride *domain.AccessOverride) *config.AccessConfig {
	if assetOverride == nil || (assetOverride.View == "" && assetOverride.AllowedUsers == nil && assetOverride.AllowedGroups == nil) {
		return albumACL
	}

	// Start from album ACL as base.
	effective := &config.AccessConfig{}
	if albumACL != nil {
		*effective = *albumACL
	}

	// Apply overrides.
	if assetOverride.View != "" {
		effective.View = assetOverride.View
	}
	if assetOverride.AllowedUsers != nil {
		effective.AllowedUsers = assetOverride.AllowedUsers
	}
	if assetOverride.AllowedGroups != nil {
		effective.AllowedGroups = assetOverride.AllowedGroups
	}
	return effective
}

func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
