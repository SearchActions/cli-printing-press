package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/require"
)

// walkerQuerySpec builds a spec whose child endpoint is reached by a walker but
// whose path carries no {placeholder} — the parent key rides in the query
// string instead (the Slack Web API shape: /conversations.history?channel=<id>).
// No endpoint anywhere in this spec has a path param, which is the condition
// that used to omit replacePathParam's definition while still emitting a call.
func walkerQuerySpec(name string) *spec.APISpec {
	apiSpec := minimalSpec(name)
	apiSpec.Resources = map[string]spec.Resource{
		"channels": {
			Description: "Channels",
			Endpoints: map[string]spec.Endpoint{
				"list": {
					Method:      "GET",
					Path:        "/conversations.list",
					Description: "List channels",
				},
			},
		},
		"messages": {
			Description: "Messages",
			Endpoints: map[string]spec.Endpoint{
				"list": {
					Method:      "GET",
					Path:        "/conversations.history",
					Description: "List messages for a channel",
					Syncable:    true,
					Walker: &spec.WalkerConfig{
						Parent:   "channels",
						KeyField: "id",
						KeyParam: "channel",
					},
					Params: []spec.Param{
						{Name: "channel", Type: "string", Required: true, Description: "Channel ID"},
					},
				},
			},
		},
	}
	return apiSpec
}

// A walker on a placeholder-free path emits a replacePathParam call site in
// sync.go. Before this was fixed the definition stayed gated on the spec
// carrying a path param, so the generated module did not compile at all.
// requireGeneratedCompiles is the assertion that matters here — a
// strings.Contains check on the call site would have passed against the broken
// generator.
func TestWalkerWithQueryKeyParamEmitsReplacePathParamDefinition(t *testing.T) {
	t.Parallel()

	outputDir := filepath.Join(t.TempDir(), "walker-query-pp-cli")
	require.NoError(t, New(walkerQuerySpec("walker-query"), outputDir).Generate())

	syncSrc := readGeneratedFile(t, outputDir, "internal", "cli", "sync.go")
	require.Contains(t, syncSrc, "replacePathParam(path, pathParam.Param",
		"sync.go should call replacePathParam on the dependent-resource path")

	helpersSrc := readGeneratedFile(t, outputDir, "internal", "cli", "helpers.go")
	require.Contains(t, helpersSrc, "func replacePathParam(path, name, value string) string",
		"helpers.go must define replacePathParam whenever sync.go calls it, even though no spec path carries a {placeholder}")

	requireGeneratedCompiles(t, outputDir)
}

// The parent key must reach the request's query params. Path substitution is a
// no-op when there is no placeholder, so without this the child request goes out
// unscoped — fetching the whole collection instead of one parent's slice, with
// no error and a completely plausible response shape.
func TestWalkerWithQueryKeyParamScopesTheChildRequest(t *testing.T) {
	t.Parallel()

	outputDir := filepath.Join(t.TempDir(), "walker-query-scope-pp-cli")
	require.NoError(t, New(walkerQuerySpec("walker-query-scope"), outputDir).Generate())

	syncSrc := readGeneratedFile(t, outputDir, "internal", "cli", "sync.go")

	require.Contains(t, syncSrc, "scopeParams := map[string]string{}",
		"sync.go should collect parent keys that have no path placeholder to substitute into")
	require.Contains(t, syncSrc, `if strings.Contains(path, "{"+pathParam.Param+"}")`,
		"the path-vs-query decision must be made per param, so path-templated walkers keep substituting")
	require.Contains(t, syncSrc, "scopeParams[pathParam.Param] = value",
		"a parent key with no path placeholder must be routed to the query params")
	require.Contains(t, syncSrc, "for scopeParam, scopeValue := range scopeParams {",
		"the collected scope must be applied to the outgoing request params")

	// Ordering is load-bearing: parent scope is structural, not a preference.
	// If user flags were applied last they could retarget the parent key, making
	// every iteration of the fan-out request the same wrong parent.
	userFlagIdx := indexOfGenerated(t, syncSrc, "userParams.applyTo(dep.Name, params, true)")
	scopeIdx := indexOfGenerated(t, syncSrc, "for scopeParam, scopeValue := range scopeParams {")
	require.Less(t, userFlagIdx, scopeIdx,
		"parent scope must be applied AFTER user flags so --global-param cannot retarget the parent key")

	// The dependent must actually carry the query-param name from the walker.
	require.Contains(t, syncSrc, `ParentIDParam: "channel"`,
		"the walker's key_param should become the dependent's ParentIDParam")

	requireGeneratedCompiles(t, outputDir)
}

// A walker whose child path DOES carry a placeholder must be unaffected: the key
// belongs in the path, not the query string. This is the regression guard for
// every hierarchical API already relying on the path-substitution behavior.
func TestWalkerWithPathPlaceholderStillSubstitutesIntoPath(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("walker-path")
	apiSpec.Resources = map[string]spec.Resource{
		"games": {
			Description: "Games",
			Endpoints: map[string]spec.Endpoint{
				"list": {
					Method:      "GET",
					Path:        "/games",
					Description: "List games",
					IDField:     "game_key",
				},
			},
		},
		"leagues": {
			Description: "Leagues",
			Endpoints: map[string]spec.Endpoint{
				"list": {
					Method:      "GET",
					Path:        "/games/{game_key}/leagues",
					Description: "List leagues for a game",
					Walker: &spec.WalkerConfig{
						Parent:   "games",
						KeyField: "game_key",
						KeyParam: "game_key",
					},
				},
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), "walker-path-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	syncSrc := readGeneratedFile(t, outputDir, "internal", "cli", "sync.go")
	require.Contains(t, syncSrc, "path = replacePathParam(path, pathParam.Param",
		"a placeholder-carrying walker path must still be substituted, not diverted to the query string")
	require.Contains(t, syncSrc, `PathTemplate: "/games/{game_key}/leagues"`,
		"the dependent should keep its templated path")

	requireGeneratedCompiles(t, outputDir)
}

func indexOfGenerated(t *testing.T, haystack, needle string) int {
	t.Helper()
	idx := strings.Index(haystack, needle)
	require.GreaterOrEqual(t, idx, 0, "expected generated source to contain %q", needle)
	return idx
}
