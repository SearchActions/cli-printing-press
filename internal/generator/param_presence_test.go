package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/naming"
	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/require"
)

// TestParamPresenceExpr pins paramPresenceExpr's three-way shape selection
// (VALUE / CHANGED / SUPERSET) in isolation from flag-name derivation: the
// changed/value expressions are opaque placeholders so each case tests only
// the selection logic, not naming.
func TestParamPresenceExpr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		p    spec.Param
		site presenceSite
		want string
	}{
		{
			name: "optional string at query site is SUPERSET with an empty-string comparator",
			p:    spec.Param{Name: "name", Type: "string"},
			site: presenceSiteQuery,
			want: `(CHANGED || VALUE != "")`,
		},
		{
			name: "optional int at query site is SUPERSET with a zero comparator",
			p:    spec.Param{Name: "count", Type: "int"},
			site: presenceSiteQuery,
			want: `(CHANGED || VALUE != 0)`,
		},
		{
			name: "int with a default keeps the zero comparator",
			p:    spec.Param{Name: "limit", Type: "int", Default: 50},
			site: presenceSiteQuery,
			want: `(CHANGED || VALUE != 0)`,
		},
		{
			name: "ID-like int is string-typed so the comparator is empty-string",
			p:    spec.Param{Name: "user_id", Type: "int"},
			site: presenceSiteQuery,
			want: `(CHANGED || VALUE != "")`,
		},
		{
			name: "cursor int is string-typed so the comparator is empty-string",
			p:    spec.Param{Name: "offset", Type: "int"},
			site: presenceSiteQuery,
			want: `(CHANGED || VALUE != "")`,
		},
		{
			name: "required non-defaulted bool at the body site is SUPERSET (string-backed)",
			p:    spec.Param{Name: "all_day", Type: "boolean", Required: true},
			site: presenceSiteBody,
			want: `(CHANGED || VALUE != "")`,
		},
		{
			name: "optional bool at the body site is pure CHANGED",
			p:    spec.Param{Name: "private", Type: "boolean"},
			site: presenceSiteBody,
			want: `CHANGED`,
		},
		{
			name: "defaulted (even if required) bool at the body site is pure CHANGED",
			p:    spec.Param{Name: "enabled", Type: "boolean", Required: true, Default: true},
			site: presenceSiteBody,
			want: `CHANGED`,
		},
		{
			name: "optional bool at the query site stays SUPERSET so an untouched default still sends",
			p:    spec.Param{Name: "verbose", Type: "boolean"},
			site: presenceSiteQuery,
			want: `(CHANGED || VALUE != false)`,
		},
		{
			name: "optional bool at the query site stays SUPERSET even at depth zero (HTML/query sites never use CHANGED)",
			p:    spec.Param{Name: "verbose", Type: "bool"},
			site: presenceSiteQuery,
			want: `(CHANGED || VALUE != false)`,
		},
		{
			name: "env-default global-scope param is bare VALUE, ignoring CHANGED entirely",
			p:    spec.Param{Name: "region", Type: "string", GlobalScope: true},
			site: presenceSiteQuery,
			want: `VALUE != ""`,
		},
		{
			name: "env-default global-scope bool is not eligible (env defaults are string-only) so it still gates on VALUE",
			p:    spec.Param{Name: "region", Type: "string", GlobalScope: true},
			site: presenceSiteBody,
			want: `VALUE != ""`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := paramPresenceExpr(tc.p, "VALUE", "CHANGED", tc.site)
			require.Equal(t, tc.want, got)
		})
	}
}

// TestPresenceChangedExpr pins the alias-OR vs. single-Changed split: a
// top-level flag ORs across its hidden aliases (K-1046 shares this with
// renderFlatBodyRequiredCheck), while a nested body field has no alias flag
// registered (renderFlatBodyFlagReg guards on topLevel) so it tests a single
// Changed() on the joined flag name.
func TestPresenceChangedExpr(t *testing.T) {
	t.Parallel()

	t.Run("top level ORs across hidden aliases, matching flagChangedExpr exactly", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "query", Aliases: []string{"q"}}
		got := presenceChangedExpr(p, "ignored-when-top-level", true)
		require.Equal(t, flagChangedExpr(p), got)
		require.Contains(t, got, `cmd.Flags().Changed("q")`)
	})

	t.Run("nested field tests a single Changed on the joined flag name, never the alias", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "dateTime", Aliases: []string{"dt"}}
		got := presenceChangedExpr(p, "start-date-time", false)
		require.Equal(t, `cmd.Flags().Changed("start-date-time")`, got)
		require.NotContains(t, got, "dt")
	})
}

// TestBodyLeafPresenceExpr pins bodyLeafPresenceExpr's own branch: complex
// (object/array) and JSON-string leaves stay VALUE unconditionally (D4),
// because an empty string has no wire meaning for them and SUPERSET would
// turn an explicit empty into a JSON parse error instead of an omit.
// json_or_scalar is neither complex nor a JSON-string param (its Format is
// "json_or_scalar", not "json"), so it must fall through to the general
// SUPERSET path rather than being swept into the VALUE-only carve-out.
func TestBodyLeafPresenceExpr(t *testing.T) {
	t.Parallel()

	t.Run("complex object leaf stays VALUE", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "metadata", Type: "object"}
		got := bodyLeafPresenceExpr(p, "Metadata", "metadata", true)
		require.Equal(t, `bodyMetadata != ""`, got)
	})

	t.Run("array leaf stays VALUE", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "tags", Type: "array"}
		got := bodyLeafPresenceExpr(p, "Tags", "tags", true)
		require.Equal(t, `bodyTags != ""`, got)
	})

	t.Run("JSON-string leaf (format=json) stays VALUE", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "config", Type: "string", Format: "json"}
		got := bodyLeafPresenceExpr(p, "Config", "config", true)
		require.Equal(t, `bodyConfig != ""`, got)
	})

	t.Run("json_or_scalar leaf is SUPERSET, not VALUE", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "response_engine", Type: "string", Format: "json_or_scalar"}
		got := bodyLeafPresenceExpr(p, "ResponseEngine", "response-engine", true)
		require.Equal(t, `(cmd.Flags().Changed("response-engine") || bodyResponseEngine != "")`, got)
	})

	t.Run("scalar string leaf is SUPERSET", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "name", Type: "string"}
		got := bodyLeafPresenceExpr(p, "Name", "name", true)
		require.Equal(t, `(cmd.Flags().Changed("name") || bodyName != "")`, got)
	})

	t.Run("optional bool leaf is pure CHANGED, matching bodyMap's own pin", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "private", Type: "boolean"}
		got := bodyLeafPresenceExpr(p, "Private", "private", true)
		require.Equal(t, `cmd.Flags().Changed("private")`, got)
	})

	t.Run("nested (non-top-level) scalar leaf uses the joined flag name, no alias", func(t *testing.T) {
		t.Parallel()
		p := spec.Param{Name: "dateTime", Aliases: []string{"dt"}, Type: "string"}
		got := bodyLeafPresenceExpr(p, "StartDateTime", "start-date-time", false)
		require.Equal(t, `(cmd.Flags().Changed("start-date-time") || bodyStartDateTime != "")`, got)
	})
}

// TestGeneratedOutput_ParamPresenceGates is the compile-level companion to
// the unit tests above: it generates one fixture spec covering every site
// class the plan brought into scope (JSON body, nested body, multipart,
// form, GET query, paginated GET, promoted query, HTML) and asserts the
// post-format.Source emitted text at each site, plus the two explicitly
// unchanged sites (D3 client-side filter, the env-default gate).
func TestGeneratedOutput_ParamPresenceGates(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("presence-gates")
	// SpecSource "docs" plus a registered response Type is what makes the
	// /widgets/batch endpoint below eligible for the D3 client-side filter
	// path (endpointClientSideFilters), which the plan keeps unchanged.
	apiSpec.SpecSource = "docs"
	apiSpec.Types = map[string]spec.TypeDef{
		"WidgetItem": {Fields: []spec.TypeField{{Name: "status", Type: "string"}}},
	}
	apiSpec.Resources = map[string]spec.Resource{
		"widgets": {
			Description: "Manage widgets",
			Endpoints: map[string]spec.Endpoint{
				"create": {
					Method:      "POST",
					Path:        "/widgets",
					Description: "Create a widget",
					Body: []spec.Param{
						{Name: "name", Type: "string", Required: true},
						{Name: "count", Type: "int"},
						{
							Name: "dimensions",
							Type: "object",
							Fields: []spec.Param{
								{Name: "width", Type: "int"},
								{Name: "height", Type: "int"},
							},
						},
					},
				},
				"list": {
					Method:      "GET",
					Path:        "/widgets",
					Description: "List widgets",
					Params: []spec.Param{
						{Name: "verbose", Type: "boolean"},
						{Name: "region", Type: "string", GlobalScope: true},
					},
				},
				"paged": {
					Method:      "GET",
					Path:        "/widgets/paged",
					Description: "List widgets with pagination",
					Pagination: &spec.Pagination{
						Type:           "cursor",
						LimitParam:     "limit",
						CursorParam:    "after",
						NextCursorPath: "next_cursor",
						HasMoreField:   "has_more",
					},
					Params: []spec.Param{
						{Name: "after", Type: "string", Description: "Cursor"},
						{Name: "limit", Type: "int", Description: "Page size"},
						{Name: "active", Type: "boolean", Description: "Filter to active widgets only"},
					},
				},
				"filter": {
					Method:      "GET",
					Path:        "/widgets/batch",
					Description: "Client-side filtered read (D3: unchanged)",
					Params: []spec.Param{
						{Name: "status", Type: "string"},
					},
					Response: spec.ResponseDef{Type: "array", Item: "WidgetItem"},
				},
			},
		},
		"uploads": {
			Description: "Upload assets",
			Endpoints: map[string]spec.Endpoint{
				// A second endpoint keeps this resource from single-endpoint
				// promotion (buildPromotedCommands), so "create" exercises the
				// plain command_endpoint.go.tmpl multipart sites, not the
				// command_promoted.go.tmpl ones already covered by "promo" below.
				"list": {
					Method:      "GET",
					Path:        "/uploads",
					Description: "List uploads",
				},
				"create": {
					Method:             "POST",
					Path:               "/uploads",
					Description:        "Upload an asset via multipart",
					RequestContentType: "multipart/form-data",
					Body: []spec.Param{
						{Name: "label", Type: "string", Required: true},
						{Name: "priority", Type: "int"},
					},
				},
			},
		},
		"forms": {
			Description: "Form-encoded resource",
			Endpoints: map[string]spec.Endpoint{
				"list": {
					Method:      "GET",
					Path:        "/forms",
					Description: "List forms",
				},
				"create": {
					Method:             "POST",
					Path:               "/forms",
					Description:        "Create via form-encoding",
					RequestContentType: "application/x-www-form-urlencoded",
					Body: []spec.Param{
						{Name: "title", Type: "string", Required: true},
						{Name: "weight", Type: "int"},
					},
				},
			},
		},
		"promo": {
			Description: "Single-endpoint resource that promotes to the top level",
			Endpoints: map[string]spec.Endpoint{
				"list": {
					Method:      "GET",
					Path:        "/promo",
					Description: "List promo (promoted: only endpoint in its resource)",
					Params: []spec.Param{
						{Name: "score", Type: "int"},
					},
				},
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())

	// Site: JSON body (widgets create) — scalar and nested leaves are SUPERSET.
	// format.Source strips the outer parens because each if-condition here is
	// one whole wrapped expression (unlike a multi-leaf join, which keeps its
	// inner parens — see body_nested_object_test.go's required-check pin).
	create := readGeneratedFile(t, outputDir, "internal", "cli", "widgets_create.go")
	require.Contains(t, create, `if cmd.Flags().Changed("name") || bodyName != "" {`)
	require.Contains(t, create, `if cmd.Flags().Changed("count") || bodyCount != 0 {`)
	require.Contains(t, create, `if cmd.Flags().Changed("dimensions-width") || bodyDimensionsWidth != 0 {`)
	require.Contains(t, create, `if cmd.Flags().Changed("dimensions-height") || bodyDimensionsHeight != 0 {`)

	// Site: GET query — optional bool stays SUPERSET (not CHANGED; only body
	// bools get the CHANGED carve-out), env-default global scope stays bare VALUE.
	list := readGeneratedFile(t, outputDir, "internal", "cli", "widgets_list.go")
	require.Contains(t, list, `if cmd.Flags().Changed("verbose") || flagVerbose != false {`)
	require.Contains(t, list, `if flagRegion != "" {`)
	require.NotContains(t, list, `if cmd.Flags().Changed("region")`)

	// Site: paginated GET — cursor param unconditional, others gated; the
	// paginatedGet strip in helpers.go no longer contains a "0"/"false" filter.
	listPaged := readGeneratedFile(t, outputDir, "internal", "cli", "widgets_paged.go")
	require.Contains(t, listPaged, `paginatedParams := map[string]string{}`)
	require.Contains(t, listPaged, `paginatedParams["after"] = formatCLIParamValue(flagAfter)`)
	require.Contains(t, listPaged, `if cmd.Flags().Changed("limit") || flagLimit != 0 {`)
	require.Contains(t, listPaged, `if cmd.Flags().Changed("active") || flagActive != false {`)
	require.NotContains(t, listPaged, `if flagAfter != "" {`, "cursor param must be unconditional, not gated")

	// Scoped to the paginatedGet strip itself (not the whole file): an
	// unrelated helper elsewhere in helpers.go legitimately still checks
	// for "false"/"0" as part of a broader "looks empty" heuristic, and a
	// file-wide NotContains would false-positive on that.
	helpers := readGeneratedFile(t, outputDir, "internal", "cli", "helpers.go")
	paginatedGetIdx := strings.Index(helpers, "func paginatedGet(")
	require.GreaterOrEqual(t, paginatedGetIdx, 0, "expected a paginatedGet function in helpers.go")
	paginatedGetBody := helpers[paginatedGetIdx:]
	if end := strings.Index(paginatedGetBody, "\nfunc "); end >= 0 {
		paginatedGetBody = paginatedGetBody[:end]
	}
	require.Contains(t, paginatedGetBody, `if v == "" {`)
	require.NotContains(t, paginatedGetBody, `v == "0"`)
	require.NotContains(t, paginatedGetBody, `v == "false"`)

	// Site: D3 — client-side response filter stays the OLD value-only gate.
	// The same param also drives the request-side query gate (now SUPERSET,
	// `cmd.Flags().Changed("status") || flagStatus != ""`), so this asserts
	// the exact filterJSONByFieldValues block rather than a blanket
	// NotContains on "Changed", which the query-side gate would trip.
	filter := readGeneratedFile(t, outputDir, "internal", "cli", "widgets_filter.go")
	require.Contains(t, filter, "if flagStatus != \"\" {\n\t\t\t\tdata = filterJSONByFieldValues(data, \"status\", formatCLIParamValue(flagStatus))\n\t\t\t}")

	// Site: multipart — scalar string and int branches are SUPERSET.
	upload := readGeneratedFile(t, outputDir, "internal", "cli", "uploads_create.go")
	require.Contains(t, upload, `if cmd.Flags().Changed("label") || bodyLabel != "" {`)
	require.Contains(t, upload, `if cmd.Flags().Changed("priority") || bodyPriority != 0 {`)

	// Site: form — same shape as multipart.
	formSrc := readGeneratedFile(t, outputDir, "internal", "cli", "forms_create.go")
	require.Contains(t, formSrc, `if cmd.Flags().Changed("title") || bodyTitle != "" {`)
	require.Contains(t, formSrc, `if cmd.Flags().Changed("weight") || bodyWeight != 0 {`)

	// Site: promoted query param.
	promoted := readGeneratedFile(t, outputDir, "internal", "cli", "promoted_promo.go")
	require.Contains(t, promoted, `if cmd.Flags().Changed("score") || flagScore != 0 {`)

	requireGeneratedCompiles(t, outputDir)
}
