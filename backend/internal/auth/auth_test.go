package auth

import (
	"context"
	"testing"

	"github.com/perrito666/gollery/backend/internal/domain"
)

func TestWithPrincipal_RoundTrip(t *testing.T) {
	p := &domain.Principal{Username: "alice", Groups: []string{"editors"}}
	ctx := WithPrincipal(context.Background(), p)

	got := PrincipalFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil principal")
	}
	if got.Username != "alice" {
		t.Errorf("username = %q, want %q", got.Username, "alice")
	}
	if len(got.Groups) != 1 || got.Groups[0] != "editors" {
		t.Errorf("groups = %v", got.Groups)
	}
}

func TestPrincipalFromContext_Empty(t *testing.T) {
	got := PrincipalFromContext(context.Background())
	if got != nil {
		t.Error("expected nil for empty context")
	}
}

func TestPrincipalFromContext_Admin(t *testing.T) {
	p := &domain.Principal{Username: "root", IsAdmin: true}
	ctx := WithPrincipal(context.Background(), p)

	got := PrincipalFromContext(ctx)
	if !got.IsAdmin {
		t.Error("expected admin flag preserved")
	}
}
