// Package rbac provides Casbin authorization model loading and enforcer
// construction. P2 pins the model shape (sub, obj, act) matching the
// permission set in architecture §12.2 (e.g. post:create, user:ban,
// ai_task:retry). Policy storage and enforcement middleware are added in P4.
package rbac

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"

	"ai-forum/backend/internal/auth"

	"github.com/casbin/casbin/v2"
)

// NewEnforcer constructs a Casbin enforcer from the model file at modelPath
// with no policy adapter attached. Callers add policies in-memory (P2) or
// supply a real adapter (P4). The returned enforcer is ready to evaluate
// r = sub, obj, act requests against the loaded model.
func NewEnforcer(modelPath string) (*casbin.Enforcer, error) {
	e, err := casbin.NewEnforcer(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer from %q: %w", modelPath, err)
	}
	return e, nil
}

// DefaultModelPath returns the absolute path to the bundled model.conf that
// ships with this package. It resolves the path relative to the source file
// via runtime.Caller so callers (P4) can do rbac.NewEnforcer(rbac.DefaultModelPath())
// without hardcoding filesystem layout.
func DefaultModelPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		// runtime.Caller failing is extraordinarily unlikely; fall back to a
		// best-effort relative path so the error surfaces at the caller.
		return "internal/rbac/model.conf"
	}
	return filepath.Join(filepath.Dir(file), "model.conf")
}

type Authorizer struct {
	enforcer *casbin.Enforcer
}

func NewAuthorizer(modelPath string) (*Authorizer, error) {
	e, err := NewEnforcer(modelPath)
	if err != nil {
		return nil, err
	}
	return &Authorizer{enforcer: e}, nil
}

func (a *Authorizer) AddPolicy(sub, obj, act string) error {
	_, err := a.enforcer.AddPolicy(sub, obj, act)
	return err
}

func (a *Authorizer) SeedAdminPolicies() error {
	for _, policy := range [][3]string{
		{"ADMIN", "post", "create"},
		{"ADMIN", "post", "delete-any"},
		{"ADMIN", "user", "ban"},
		{"ADMIN", "ai_task", "retry"},
	} {
		if err := a.AddPolicy(policy[0], policy[1], policy[2]); err != nil {
			return err
		}
	}
	return nil
}

func (a *Authorizer) Enforce(sub, obj, act string) (bool, error) {
	return a.enforcer.Enforce(sub, obj, act)
}

func (a *Authorizer) Require(sub, obj, act string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed, err := a.Enforce(sub, obj, act)
		if err != nil {
			http.Error(w, "authorize", http.StatusInternalServerError)
			return
		}
		if !allowed {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *Authorizer) RequireSubject(obj, act string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub, ok := auth.SubjectFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		a.Require(sub.Role, obj, act, next).ServeHTTP(w, r)
	})
}
