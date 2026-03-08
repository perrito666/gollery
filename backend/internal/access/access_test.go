package access

import (
	"testing"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/domain"
)

func TestCheckView_NilACL(t *testing.T) {
	if d := CheckView(nil, nil); d != Allow {
		t.Error("nil ACL should allow")
	}
}

func TestCheckView_EmptyView(t *testing.T) {
	acl := &config.AccessConfig{}
	if d := CheckView(acl, nil); d != Allow {
		t.Error("empty view should default to allow")
	}
}

func TestCheckView_Public(t *testing.T) {
	acl := &config.AccessConfig{View: "public"}
	if d := CheckView(acl, nil); d != Allow {
		t.Error("public should allow anonymous")
	}
	p := &domain.Principal{Username: "alice"}
	if d := CheckView(acl, p); d != Allow {
		t.Error("public should allow authenticated")
	}
}

func TestCheckView_Authenticated_Anonymous(t *testing.T) {
	acl := &config.AccessConfig{View: "authenticated"}
	if d := CheckView(acl, nil); d != Deny {
		t.Error("authenticated should deny anonymous")
	}
}

func TestCheckView_Authenticated_LoggedIn(t *testing.T) {
	acl := &config.AccessConfig{View: "authenticated"}
	p := &domain.Principal{Username: "alice"}
	if d := CheckView(acl, p); d != Allow {
		t.Error("authenticated should allow logged-in user")
	}
}

func TestCheckView_Restricted_Anonymous(t *testing.T) {
	acl := &config.AccessConfig{
		View:         "restricted",
		AllowedUsers: []string{"alice"},
	}
	if d := CheckView(acl, nil); d != Deny {
		t.Error("restricted should deny anonymous")
	}
}

func TestCheckView_Restricted_AllowedUser(t *testing.T) {
	acl := &config.AccessConfig{
		View:         "restricted",
		AllowedUsers: []string{"alice", "bob"},
	}
	p := &domain.Principal{Username: "bob"}
	if d := CheckView(acl, p); d != Allow {
		t.Error("restricted should allow listed user")
	}
}

func TestCheckView_Restricted_DeniedUser(t *testing.T) {
	acl := &config.AccessConfig{
		View:         "restricted",
		AllowedUsers: []string{"alice"},
	}
	p := &domain.Principal{Username: "eve"}
	if d := CheckView(acl, p); d != Deny {
		t.Error("restricted should deny unlisted user")
	}
}

func TestCheckView_Restricted_AllowedGroup(t *testing.T) {
	acl := &config.AccessConfig{
		View:          "restricted",
		AllowedGroups: []string{"family"},
	}
	p := &domain.Principal{Username: "bob", Groups: []string{"family"}}
	if d := CheckView(acl, p); d != Allow {
		t.Error("restricted should allow user in listed group")
	}
}

func TestCheckView_Restricted_ObjectAdmin(t *testing.T) {
	acl := &config.AccessConfig{
		View:   "restricted",
		Admins: []string{"horacio"},
	}
	p := &domain.Principal{Username: "horacio"}
	if d := CheckView(acl, p); d != Allow {
		t.Error("object admin should be allowed in restricted")
	}
}

func TestCheckView_Restricted_GlobalAdmin(t *testing.T) {
	acl := &config.AccessConfig{
		View:         "restricted",
		AllowedUsers: []string{"alice"},
	}
	p := &domain.Principal{Username: "superadmin", IsAdmin: true}
	if d := CheckView(acl, p); d != Allow {
		t.Error("global admin should always be allowed")
	}
}

func TestCheckView_UnknownMode(t *testing.T) {
	acl := &config.AccessConfig{View: "secret"}
	p := &domain.Principal{Username: "alice"}
	if d := CheckView(acl, p); d != Deny {
		t.Error("unknown mode should deny")
	}
}

func TestIsObjectAdmin_GlobalAdmin(t *testing.T) {
	p := &domain.Principal{Username: "root", IsAdmin: true}
	if !IsObjectAdmin(nil, p) {
		t.Error("global admin should be object admin")
	}
}

func TestIsObjectAdmin_Listed(t *testing.T) {
	acl := &config.AccessConfig{Admins: []string{"horacio"}}
	p := &domain.Principal{Username: "horacio"}
	if !IsObjectAdmin(acl, p) {
		t.Error("listed admin should be object admin")
	}
}

func TestIsObjectAdmin_NotListed(t *testing.T) {
	acl := &config.AccessConfig{Admins: []string{"horacio"}}
	p := &domain.Principal{Username: "eve"}
	if IsObjectAdmin(acl, p) {
		t.Error("unlisted user should not be object admin")
	}
}

func TestIsObjectAdmin_NilPrincipal(t *testing.T) {
	acl := &config.AccessConfig{Admins: []string{"horacio"}}
	if IsObjectAdmin(acl, nil) {
		t.Error("nil principal should not be object admin")
	}
}

func TestEffectiveAssetACL_NilOverride(t *testing.T) {
	albumACL := &config.AccessConfig{View: "restricted", AllowedUsers: []string{"alice"}}
	effective := EffectiveAssetACL(albumACL, nil)
	if effective != albumACL {
		t.Error("nil override should return album ACL")
	}
}

func TestEffectiveAssetACL_EmptyOverride(t *testing.T) {
	albumACL := &config.AccessConfig{View: "restricted"}
	override := &domain.AccessOverride{}
	effective := EffectiveAssetACL(albumACL, override)
	if effective != albumACL {
		t.Error("empty override should return album ACL")
	}
}

func TestEffectiveAssetACL_ViewOverride(t *testing.T) {
	albumACL := &config.AccessConfig{View: "public", AllowedUsers: []string{"alice"}}
	override := &domain.AccessOverride{View: "restricted"}
	effective := EffectiveAssetACL(albumACL, override)

	if effective.View != "restricted" {
		t.Errorf("view = %q, want restricted", effective.View)
	}
	// AllowedUsers inherited from album.
	if len(effective.AllowedUsers) != 1 || effective.AllowedUsers[0] != "alice" {
		t.Errorf("AllowedUsers = %v, want [alice]", effective.AllowedUsers)
	}
}

func TestEffectiveAssetACL_UsersOverride(t *testing.T) {
	albumACL := &config.AccessConfig{View: "restricted", AllowedUsers: []string{"alice"}}
	override := &domain.AccessOverride{AllowedUsers: []string{"bob", "carol"}}
	effective := EffectiveAssetACL(albumACL, override)

	if effective.View != "restricted" {
		t.Errorf("view = %q, want restricted", effective.View)
	}
	if len(effective.AllowedUsers) != 2 || effective.AllowedUsers[0] != "bob" {
		t.Errorf("AllowedUsers = %v, want [bob carol]", effective.AllowedUsers)
	}
}

func TestEffectiveAssetACL_FullOverride(t *testing.T) {
	albumACL := &config.AccessConfig{View: "public"}
	override := &domain.AccessOverride{
		View:          "restricted",
		AllowedUsers:  []string{"bob"},
		AllowedGroups: []string{"editors"},
	}
	effective := EffectiveAssetACL(albumACL, override)

	// Asset-level override restricts access even though album is public.
	p := &domain.Principal{Username: "eve"}
	if d := CheckView(effective, p); d != Deny {
		t.Error("non-allowed user should be denied on restricted asset")
	}

	pBob := &domain.Principal{Username: "bob"}
	if d := CheckView(effective, pBob); d != Allow {
		t.Error("allowed user should be permitted")
	}
}

func TestEffectiveAssetACL_AdminBypass(t *testing.T) {
	albumACL := &config.AccessConfig{View: "public"}
	override := &domain.AccessOverride{View: "restricted", AllowedUsers: []string{"alice"}}
	effective := EffectiveAssetACL(albumACL, override)

	admin := &domain.Principal{Username: "superadmin", IsAdmin: true}
	if d := CheckView(effective, admin); d != Allow {
		t.Error("global admin should bypass asset-level restriction")
	}
}
