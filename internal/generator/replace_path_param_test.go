package generator

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// escapePathParamForTest mirrors cliutil.EscapePathParam, the shared escaper
// that both the emitted CLI (replacePathParam) and MCP (mcpPathValue) surfaces
// route path-param values through. Keep this byte-identical with the template
// body in cliutil_text.go.tmpl. The template-emission side is pinned separately
// by TestReplacePathParamPercentEncodesValue / TestEscapePathParamPreserves-
// HierarchicalIdentifiers in path_param_encoding_test.go, which assert the
// generated cliutil/text.go and the CLI+MCP call sites. This table locks the
// adversarial behavior contract.
func escapePathParamForTest(value string) string {
	if value == "" || strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") || strings.Contains(value, "//") {
		return url.PathEscape(value)
	}
	segments := strings.Split(value, "/")
	for i, segment := range segments {
		if segment == "." || segment == ".." {
			segments[i] = strings.Repeat("%2E", len(segment))
			continue
		}
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

// TestEscapePathParamBehavior locks the per-segment escape + traversal
// neutralization contract every generated CLI inherits from cliutil. The dot
// and empty-segment cases are the path-traversal / route-selection defenses:
// "." and ".." are encoded to %2E per segment (preserving legitimate composite
// IDs), and any empty segment falls back to a whole-value escape.
func TestEscapePathParamBehavior(t *testing.T) {
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

		// Dot segments are encoded per-segment to %2E, so the URL router cannot
		// resolve traversal, while the structural "/" is preserved.
		{"parent-dir-embedded", "properties/../accounts/123", "properties/%2E%2E/accounts/123"},
		{"current-dir-embedded", "properties/./foo", "properties/%2E/foo"},
		{"bare-parent-dir", "..", "%2E%2E"},
		{"bare-current-dir", ".", "%2E"},

		// Empty segments (leading/trailing/doubled "/") fall back to whole-value
		// escape, which encodes every "/" too — the API 404s instead of letting a
		// slash-collapsing proxy select a different route.
		{"leading-empty-segment", "/properties/123", "%2Fproperties%2F123"},
		{"trailing-empty-segment", "properties/123/", "properties%2F123%2F"},
		{"empty-mid-segment", "properties//456314183", "properties%2F%2F456314183"},

		// Per-segment PathEscape neutralizes attempts to graft a query string
		// or fragment onto the URL via an injected ? or #.
		{"query-char-in-segment", "foo?bar=1", "foo%3Fbar=1"},
		{"fragment-char-in-segment", "foo#bar", "foo%23bar"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := escapePathParamForTest(tc.value)
			assert.Equal(t, tc.want, got)
		})
	}
}
