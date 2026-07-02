package rbac

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelLoadsAndEvaluates verifies the bundled model.conf loads into a
// Casbin enforcer and evaluates a trivial in-memory policy without error
// (spec: casbin-model, "Model loads").
func TestModelLoadsAndEvaluates(t *testing.T) {
	// Arrange — load the model that ships with this package.
	e, err := NewEnforcer(DefaultModelPath())
	require.NoError(t, err, "model.conf must load into a casbin enforcer")

	// Add a trivial in-memory policy (no adapter yet — P4 supplies storage).
	// Casbin v2 AddPolicy returns (bool, error): ok=false means the policy
	// already existed, not a failure.
	ok, err := e.AddPolicy("admin", "post", "create")
	require.NoError(t, err, "AddPolicy must not error on the in-memory enforcer")
	require.True(t, ok, "AddPolicy must report the policy was added")

	// Act + Assert — allowed pair evaluates true.
	ok, err = e.Enforce("admin", "post", "create")
	require.NoError(t, err, "Enforce must not error on a valid request")
	assert.True(t, ok, "admin/post/create must be allowed by the policy")

	// Act + Assert — ungranted pair evaluates false (not an error).
	ok, err = e.Enforce("admin", "post", "delete")
	require.NoError(t, err, "Enforce must not error on an unmatched request")
	assert.False(t, ok, "admin/post/delete must be denied (no policy added)")
}

// TestDefaultModelPathIsAbsolute guards against runtime.Caller regressions
// silently returning a relative path, which would break callers that resolve
// the model from a different working directory.
func TestDefaultModelPathIsAbsolute(t *testing.T) {
	p := DefaultModelPath()
	assert.True(t, filepath.IsAbs(p),
		"DefaultModelPath must return an absolute path, got %q", p)
}
