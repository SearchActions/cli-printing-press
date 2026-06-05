package generator

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// replacePathParamForTest mirrors the function the helpers.go.tmpl template
// emits into every generated CLI. Keep this byte-identical with the template
// body — TestReplacePathParamTemplateMatchesLocalCopy is the guardrail that
// blocks template drift from silently invalidating these table-driven cases.
// See INC-2026-147.
func replacePathParamForTest(path, name, value string) string {
	segments := strings.Split(value, "/")
	for _, s := range segments {
		if s == "" || s == "." || s == ".." {
			return strings.ReplaceAll(path, "{"+name+"}", url.PathEscape(value))
		}
	}
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.ReplaceAll(path, "{"+name+"}", strings.Join(segments, "/"))
}

// TestReplacePathParamBehavior locks in the per-segment escape + traversal
// rejection contract that every generated CLI inherits from the template. The
// negative cases (empty, ".", ".." segments and ?/# injection in any segment)
// are the path-traversal defenses introduced by INC-2026-147.
func TestReplacePathParamBehavior(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value string
		want  string
	}{
		// Legitimate composite resource names — structural "/" must survive.
		{"google-ga4-property", "properties/456314183", "properties/456314183"},
		{"two-level-resource", "accounts/123/users/456", "accounts/123/users/456"},

		// Non-traversal chars in a segment must be escaped, "/" must not be.
		{"space-in-segment", "properties/foo bar", "properties/foo%20bar"},

		// Traversal-allowing patterns fall back to whole-value escape, which
		// encodes every "/" too — the API then 404s instead of letting the URL
		// router resolve the dots to a different endpoint.
		{"parent-dir-traversal", "properties/../accounts/123", "properties%2F..%2Faccounts%2F123"},
		{"current-dir-traversal", "properties/./foo", "properties%2F.%2Ffoo"},
		{"leading-empty-segment", "/properties/123", "%2Fproperties%2F123"},
		{"trailing-empty-segment", "properties/123/", "properties%2F123%2F"},
		{"empty-mid-segment", "properties//456314183", "properties%2F%2F456314183"},

		// Per-segment PathEscape neutralizes attempts to graft a query string
		// or fragment onto the URL via an injected ? or #.
		{"query-char-in-segment", "foo?bar=1", "foo%3Fbar=1"},
		{"fragment-char-in-segment", "foo#bar", "foo%23bar"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := replacePathParamForTest("{p}", "p", tc.value)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestReplacePathParamTemplateMatchesLocalCopy is the guardrail bridging the
// template emission and the table-driven cases above. If someone edits
// helpers.go.tmpl and breaks the traversal-rejection check, the local copy in
// replacePathParamForTest would still pass — this test reads the generated
// helpers.go from a fresh generation and asserts the safety check survives.
func TestReplacePathParamTemplateMatchesLocalCopy(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("path-param-guardrail")
	apiSpec.Resources = map[string]spec.Resource{
		"items": {
			Description: "Items",
			Endpoints: map[string]spec.Endpoint{
				"get": {
					Method:      "GET",
					Path:        "/items/{itemId}",
					Description: "Get item",
					Params: []spec.Param{
						{Name: "itemId", Type: "string", PathParam: true, Description: "Item ID"},
					},
				},
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), "path-param-guardrail-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	helpersGo, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "helpers.go"))
	require.NoError(t, err)
	src := string(helpersGo)

	// The template must emit the traversal-rejection check in the body of
	// replacePathParam. Without this, the per-segment escape path would let
	// "..", ".", and empty segments survive into the templated path.
	assert.Contains(t, src, `s == "" || s == "." || s == ".."`,
		"helpers.go.tmpl must emit the traversal-segment rejection check; see INC-2026-147")

	// The whole-value escape fallback is what makes the API 404 instead of
	// silently routing an authenticated request to an attacker-chosen URL.
	assert.Contains(t, src, `url.PathEscape(value)`,
		"helpers.go.tmpl must fall back to whole-value url.PathEscape when a traversal segment is detected")

	// And per-segment PathEscape is what preserves structural "/" while
	// neutralizing per-segment injection (?, #, spaces).
	assert.Contains(t, src, `segments[i] = url.PathEscape(s)`,
		"helpers.go.tmpl must percent-encode each segment individually for the happy path")

	// net/url must be imported when HasPathParams is true. Without this the
	// generated CLI fails to compile.
	assert.Contains(t, src, `"net/url"`,
		"helpers.go.tmpl must import net/url when HasPathParams is true")
}
