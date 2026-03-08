// Package access implements ACL evaluation for albums and assets.
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
