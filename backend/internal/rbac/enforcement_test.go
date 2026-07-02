package rbac

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-forum/backend/internal/auth"
)

func TestAuthorizerAllowsAndDeniesActions(t *testing.T) {
	authz, err := NewAuthorizer(DefaultModelPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := authz.AddPolicy("ADMIN", "post", "delete-any"); err != nil {
		t.Fatal(err)
	}

	allowed, err := authz.Enforce("ADMIN", "post", "delete-any")
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected admin policy to allow delete-any")
	}

	allowed, err = authz.Enforce("USER", "post", "delete-any")
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("expected user without policy to be denied")
	}
}

func TestRequireDeniesUnauthorizedAction(t *testing.T) {
	authz, err := NewAuthorizer(DefaultModelPath())
	if err != nil {
		t.Fatal(err)
	}
	h := authz.Require("USER", "post", "delete-any", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/posts/1", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestRequireSubjectUsesAuthenticatedRole(t *testing.T) {
	authz, err := NewAuthorizer(DefaultModelPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := authz.AddPolicy("ADMIN", "post", "delete-any"); err != nil {
		t.Fatal(err)
	}
	h := authz.RequireSubject("post", "delete-any", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/posts/1/status", nil)
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 1, Role: "USER"}))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("USER status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/admin/posts/1/status", nil)
	req = req.WithContext(auth.ContextWithSubject(req.Context(), auth.Subject{UserID: 2, Role: "ADMIN"}))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("ADMIN status = %d, want 204", rec.Code)
	}
}

func TestSeedAdminPoliciesAllowsConfiguredPermissions(t *testing.T) {
	authz, err := NewAuthorizer(DefaultModelPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := authz.SeedAdminPolicies(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		obj string
		act string
	}{
		{"post", "create"},
		{"post", "delete-any"},
		{"user", "ban"},
		{"ai_task", "retry"},
	} {
		allowed, err := authz.Enforce("ADMIN", tc.obj, tc.act)
		if err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Fatalf("ADMIN must allow %s:%s", tc.obj, tc.act)
		}
	}
}
