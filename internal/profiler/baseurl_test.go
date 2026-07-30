package profiler

import (
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
)

// TestSyncableBaseURLPrecedence locks the host a syncable resource's sync
// requests target after the multi-host sync fix: a per-operation (endpoint)
// server URL wins over the resource-level base, and a trailing slash is
// trimmed. Mirrors the generator's effectiveEndpointBaseURL so sync and the
// endpoint-mirror commands agree on the host for multi-host APIs. Empty
// (single-host) stays empty so the emitted path is relative and byte-identical
// to pre-fix output.
func TestSyncableBaseURLPrecedence(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		resource spec.Resource
		endpoint spec.Endpoint
		want     string
	}{
		{
			name:     "endpoint base wins over resource base",
			resource: spec.Resource{BaseURL: "https://res.example.com/v1"},
			endpoint: spec.Endpoint{BaseURL: "https://op.example.com/v1"},
			want:     "https://op.example.com/v1",
		},
		{
			name:     "falls back to resource base with trailing slash trimmed",
			resource: spec.Resource{BaseURL: "https://res.example.com/v1/"},
			endpoint: spec.Endpoint{},
			want:     "https://res.example.com/v1",
		},
		{
			name:     "single-host empty when neither declares a base",
			resource: spec.Resource{},
			endpoint: spec.Endpoint{},
			want:     "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, syncableBaseURL(tc.resource, tc.endpoint))
		})
	}
}

// TestDependentBaseURLInheritance locks the dependent host resolution: the
// child's own base wins; when the child declares none it inherits the parent
// syncable resource's resolved base; when neither has one (or the parent is
// not syncable) it is empty, leaving the emitted path relative so the client
// falls back to its compiled default host.
func TestDependentBaseURLInheritance(t *testing.T) {
	t.Parallel()
	syncable := map[string]syncableMeta{
		"accounts": {BaseURL: "https://admin.example.com/v1"},
		"events":   {},
	}
	assert.Equal(t, "https://child.example.com/v1",
		dependentBaseURL("https://child.example.com/v1", "accounts", syncable),
		"child base wins over parent")
	assert.Equal(t, "https://admin.example.com/v1",
		dependentBaseURL("", "accounts", syncable),
		"child inherits parent base when it declares none")
	assert.Equal(t, "",
		dependentBaseURL("", "events", syncable),
		"empty when neither child nor parent declares a base")
	assert.Equal(t, "",
		dependentBaseURL("", "unknown", syncable),
		"empty when the parent is not a syncable resource")
}
