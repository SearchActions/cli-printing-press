package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateDeduplicatesCamelCollidingParams covers Case A from issue #275 F-2.
// Twilio's spec lists StartTime, StartTime>, and StartTime< as distinct query
// params for date-range filtering. toCamel strips '>' and '<' as non-alphanumeric,
// so all three would yield Go identifier "StartTime" and the template would emit
// three `var flagStartTime` declarations in one function — illegal redeclaration.
func TestGenerateDeduplicatesCamelCollidingParams(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("collide-camel")
	// Two endpoints so `list` renders to its own file rather than being
	// consolidated by the single-endpoint promotion path.
	apiSpec.Resources["calls"] = spec.Resource{
		Description: "Calls",
		Endpoints: map[string]spec.Endpoint{
			"list": {
				Method:      "GET",
				Path:        "/calls",
				Description: "List calls within a date range",
				Params: []spec.Param{
					{Name: "StartTime", Type: "string", Description: "Exact timestamp"},
					{Name: "StartTime>", Type: "string", Description: "After timestamp"},
					{Name: "StartTime<", Type: "string", Description: "Before timestamp"},
				},
			},
			"get": {
				Method:      "GET",
				Path:        "/calls/{id}",
				Description: "Get one call",
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), "collide-camel-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	flagVars, flagBindings := parseFlagDeclarations(t,
		filepath.Join(outputDir, "internal", "cli", "calls_list.go"))

	assertNoDuplicates(t, flagVars,
		"each param must produce a distinct Go identifier")
	assertNoDuplicates(t, flagBindings,
		"each param must register a distinct cobra flag name")
	require.Len(t, flagVars, 3,
		"all three params must still be represented after dedup")
}

func TestGenerateRejectsAuthoredPublicFlagCollisions(t *testing.T) {
	cases := []struct {
		name    string
		params  []spec.Param
		body    []spec.Param
		paging  *spec.Pagination
		method  string
		wantErr string
	}{
		{
			name: "alias equals fallback public name",
			params: []spec.Param{
				{Name: "s", Type: "string", FlagName: "address", Aliases: []string{"city"}},
				{Name: "city", Type: "string"},
			},
			method:  "GET",
			wantErr: `public name "city" collides with param "s" alias`,
		},
		{
			name: "duplicate authored flag name",
			params: []spec.Param{
				{Name: "s", Type: "string", FlagName: "address"},
				{Name: "street", Type: "string", FlagName: "address"},
			},
			method:  "GET",
			wantErr: `public name "address" collides with param "s" public name`,
		},
		{
			name: "body collision",
			params: []spec.Param{
				{Name: "s", Type: "string", FlagName: "address"},
			},
			body: []spec.Param{
				{Name: "address", Type: "string"},
			},
			method:  "POST",
			wantErr: `body "address" public name "address" collides with param "s" public name`,
		},
		{
			name: "reserved pagination flag",
			params: []spec.Param{
				{Name: "includeAll", Type: "boolean", FlagName: "all"},
			},
			paging:  &spec.Pagination{Type: "cursor"},
			method:  "GET",
			wantErr: `collides with reserved flag --all`,
		},
		{
			name: "reserved mutating stdin flag",
			params: []spec.Param{
				{Name: "source", Type: "string", FlagName: "stdin"},
			},
			method:  "POST",
			wantErr: `collides with reserved flag --stdin`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec := minimalSpec("public-collisions")
			apiSpec.Resources["stores"] = spec.Resource{
				Description: "Stores",
				Endpoints: map[string]spec.Endpoint{
					"find": {
						Method:      tt.method,
						Path:        "/stores",
						Description: "Find stores",
						Params:      tt.params,
						Body:        tt.body,
						Pagination:  tt.paging,
					},
				},
			}

			err := New(apiSpec, filepath.Join(t.TempDir(), "out")).Generate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// TestGenerateRenamesParamCollidingWithPaginationAll covers Case B from issue #275 F-2.
// GitHub's spec has a `repos_notifications_activity-list-repo-for-authenticated-user`
// endpoint that takes an `all` param and is paginated. The endpoint template emits
// `var flagAll` once for the user-defined `all` param and again for pagination's
// "fetch all pages" flag — illegal redeclaration.
func TestGenerateRenamesParamCollidingWithPaginationAll(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("collide-all")
	apiSpec.Resources["notifications"] = spec.Resource{
		Description: "Notifications",
		Endpoints: map[string]spec.Endpoint{
			"list": {
				Method:      "GET",
				Path:        "/notifications",
				Description: "List notifications",
				Params: []spec.Param{
					{Name: "all", Type: "bool", Description: "Include read notifications"},
				},
				Pagination: &spec.Pagination{
					Type:           "page_token",
					LimitParam:     "per_page",
					CursorParam:    "page",
					NextCursorPath: "next",
					HasMoreField:   "has_more",
				},
			},
			"get": {
				Method:      "GET",
				Path:        "/notifications/{id}",
				Description: "Get one notification",
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), "collide-all-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	flagVars, flagBindings := parseFlagDeclarations(t,
		filepath.Join(outputDir, "internal", "cli", "notifications_list.go"))

	assertNoDuplicates(t, flagVars,
		"pagination's reserved flagAll must not collide with a user param named 'all'")
	assertNoDuplicates(t, flagBindings,
		"--all from pagination must not collide with --all from a user param")
	assert.Contains(t, flagVars, "flagAll",
		"pagination's flagAll keeps the canonical name")
}

func TestGenerateRenamesParamCollidingWithBrowserTLSFlag(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("collide-browser-insecure")
	apiSpec.HTTPTransport = spec.HTTPTransportBrowserChrome
	endpoint := apiSpec.Resources["items"].Endpoints["list"]
	endpoint.Params = append(endpoint.Params, spec.Param{
		Name:        "insecure",
		Type:        "boolean",
		Description: "Filter items by upstream security classification",
	})
	apiSpec.Resources["items"].Endpoints["list"] = endpoint
	apiSpec.Resources["items"].Endpoints["get"] = spec.Endpoint{
		Method:      "GET",
		Path:        "/items/{id}",
		Description: "Get an item",
	}

	outputDir := filepath.Join(t.TempDir(), "collide-browser-insecure-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	commandPath := filepath.Join(outputDir, "internal", "cli", "items_list.go")
	flagVars, flagBindings := parseFlagDeclarations(t, commandPath)
	assert.Contains(t, flagVars, "flagInsecure2")
	assert.Contains(t, flagBindings, "insecure-2")

	commandSrc, err := os.ReadFile(commandPath)
	require.NoError(t, err)
	assert.Contains(t, string(commandSrc), `params["insecure"] = formatCLIParamValue(flagInsecure2)`)

	runtimeTest := `package cli

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGeneratedBrowserTLSFlagDoesNotShadowEndpointParam(t *testing.T) {
	queryValues := make(chan string, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryValues <- r.URL.Query().Get("insecure")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, "[]")
	}))
	defer server.Close()

	t.Setenv("MYAPI_TOKEN", "test-token")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte(fmt.Sprintf("base_url = %q\n", server.URL)), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var flags rootFlags
	cmd := newRootCmd(&flags)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--config", configPath, "--insecure", "items", "list", "--insecure-2", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute generated command: %v", err)
	}
	select {
	case got := <-queryValues:
		if got != "true" {
			t.Fatalf("wire query insecure = %q, want true", got)
		}
	case <-time.After(time.Second):
		t.Fatal("generated command did not call the TLS test server")
	}
	if !flags.insecure {
		t.Fatal("root --insecure flag was not parsed independently")
	}
}
`
	require.NoError(t, os.WriteFile(
		filepath.Join(outputDir, "internal", "cli", "browser_tls_flag_runtime_test.go"),
		[]byte(runtimeTest), 0o600))
	runGoCommandRequired(t, outputDir, "test", "./internal/cli", "-run", "^TestGeneratedBrowserTLSFlagDoesNotShadowEndpointParam$", "-count=1")
	requireGeneratedCompiles(t, outputDir)
}

// TestGenerateRenamesParamCollidingWithAsyncWait covers the async-reserved-name
// path. Async-job endpoints emit `var flagWait`, `var flagWaitTimeout`, and
// `var flagWaitInterval` from the IsAsync branch in command_endpoint.go.tmpl;
// a user param literally named `wait` (or `wait_timeout`, `wait_interval`)
// would otherwise produce a duplicate `var flagWait` in the same function.
//
// Async detection requires a job-id-shaped response field plus a sibling status
// endpoint, so the spec mirrors that contract.
func TestGenerateRenamesParamCollidingWithAsyncWait(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("collide-async")
	apiSpec.Types = map[string]spec.TypeDef{
		"JobResp": {Fields: []spec.TypeField{
			{Name: "job_id", Type: "string"},
			{Name: "status", Type: "string"},
		}},
	}
	apiSpec.Resources["videos"] = spec.Resource{
		Description: "Videos",
		Endpoints: map[string]spec.Endpoint{
			"create": {
				Method:      "POST",
				Path:        "/videos",
				Description: "Create a video render job",
				Response:    spec.ResponseDef{Type: "object", Item: "JobResp"},
				Params: []spec.Param{
					{Name: "wait", Type: "string", Description: "Watermark text on the rendered video"},
				},
			},
			"get": {
				Method:      "GET",
				Path:        "/videos/{id}",
				Description: "Get one video",
				Response:    spec.ResponseDef{Type: "object", Item: "JobResp"},
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), "collide-async-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	flagVars, flagBindings := parseFlagDeclarations(t,
		filepath.Join(outputDir, "internal", "cli", "videos_create.go"))

	assertNoDuplicates(t, flagVars,
		"async's reserved flagWait must not collide with a user param named 'wait'")
	assertNoDuplicates(t, flagBindings,
		"--wait from async must not collide with --wait from a user param")
	assert.Contains(t, flagVars, "flagWait",
		"async's flagWait keeps the canonical name")
}

// TestGlobalReservedFlagsMatchTemplate is the anti-drift guard: every cobra
// flag name registered as a root PersistentFlag in root.go.tmpl must be in
// globalPersistentFlagNames, so a future global flag addition cannot silently
// reintroduce the shadow bypass (INC-2026-166). root.go.tmpl is a Go template,
// not valid Go, so it is scanned as raw bytes rather than AST-parsed.
func TestGlobalReservedFlagsMatchTemplate(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile(filepath.Join("templates", "root.go.tmpl"))
	require.NoError(t, err, "read root.go.tmpl")

	re := regexp.MustCompile(`PersistentFlags\(\)\.\w+Var\(&[^,]+,\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(src), -1)
	require.NotEmpty(t, matches, "expected to find PersistentFlags registrations in template")

	templateFlags := map[string]struct{}{}
	for _, m := range matches {
		name := m[1]
		// Skip the dynamic per-spec path-template flag ("{{kebab .}}"); it is a
		// separate shadow class tracked in TODO.md, not a fixed global name.
		if strings.Contains(name, "{{") {
			continue
		}
		templateFlags[name] = struct{}{}
	}

	for name := range templateFlags {
		_, ok := globalPersistentFlagNames[name]
		assert.True(t, ok, "global flag --%s is registered in root.go.tmpl but not reserved in globalPersistentFlagNames; add it so local flags cannot shadow it", name)
	}
	for name := range globalPersistentFlagNames {
		_, ok := templateFlags[name]
		assert.True(t, ok, "globalPersistentFlagNames reserves --%s but root.go.tmpl no longer registers it; remove the stale reservation", name)
	}
}

// TestGenerateRenamesBodyFieldCollidingWithGlobalDryRun proves the INC-2026-166
// bypass is closed: a body field named `dry_run` must NOT register a local
// --dry-run flag (which cobra would resolve in preference to the inherited
// global preview flag, silently sending instead of previewing). The derived
// flag auto-renames; the wire field name is preserved on the JSON body.
func TestGenerateRenamesBodyFieldCollidingWithGlobalDryRun(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("collide-global")
	apiSpec.Resources["campaigns"] = spec.Resource{
		Description: "Campaigns",
		Endpoints: map[string]spec.Endpoint{
			"create": {
				Method:      "POST",
				Path:        "/campaigns",
				Description: "Create a campaign",
				Body: []spec.Param{
					{Name: "name", Type: "string", Description: "Campaign name"},
					{Name: "dry_run", Type: "boolean", Description: "Spec field that must not shadow the global --dry-run"},
				},
			},
			"get": {
				Method:      "GET",
				Path:        "/campaigns/{id}",
				Description: "Get one campaign",
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), "collide-global-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	_, flagBindings := parseFlagDeclarations(t,
		filepath.Join(outputDir, "internal", "cli", "campaigns_create.go"))

	assert.NotContains(t, flagBindings, "dry-run",
		"body field dry_run must not register a local --dry-run shadowing the global preview flag")
	assert.Contains(t, flagBindings, "dry-run-2",
		"the body field dry_run must auto-rename to the non-shadowing --dry-run-2")

	src, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "campaigns_create.go"))
	require.NoError(t, err)
	assert.Contains(t, string(src), `body["dry_run"]`,
		"the wire-side body key must remain dry_run in the body map even though the public flag is renamed")
}

// TestGenerateRejectsAuthoredFlagCollidingWithGlobal covers the explicit
// flag_name path: authoring flag_name: dry-run on a param is unambiguous intent
// to claim a reserved global name and must hard-error rather than silently
// rename.
func TestGenerateRejectsAuthoredFlagCollidingWithGlobal(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("authored-global")
	apiSpec.Resources["stores"] = spec.Resource{
		Description: "Stores",
		Endpoints: map[string]spec.Endpoint{
			"find": {
				Method:      "GET",
				Path:        "/stores",
				Description: "Find stores",
				Params: []spec.Param{
					{Name: "preview", Type: "boolean", FlagName: "dry-run"},
				},
			},
		},
	}

	err := New(apiSpec, filepath.Join(t.TempDir(), "out")).Generate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collides with reserved flag --dry-run")
}

// parseFlagDeclarations returns the names of all `var flagXxx` declarations and
// the literal flag names passed to cobra's *Var registrations.
func parseFlagDeclarations(t *testing.T, path string) (vars, bindings []string) {
	t.Helper()
	src, err := os.ReadFile(path)
	require.NoError(t, err, "read generated file")

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	require.NoError(t, err, "generated file must parse as Go")

	ast.Inspect(file, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl:
			if decl.Tok != token.VAR {
				return true
			}
			for _, sp := range decl.Specs {
				vs, ok := sp.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					if strings.HasPrefix(name.Name, "flag") {
						vars = append(vars, name.Name)
					}
				}
			}
		case *ast.CallExpr:
			// cobra registrations: cmd.Flags().StringVar(&flagX, "name", ...)
			sel, ok := decl.Fun.(*ast.SelectorExpr)
			if !ok || !strings.HasSuffix(sel.Sel.Name, "Var") {
				return true
			}
			if len(decl.Args) < 2 {
				return true
			}
			lit, ok := decl.Args[1].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			bindings = append(bindings, strings.Trim(lit.Value, `"`))
		}
		return true
	})
	return vars, bindings
}

func assertNoDuplicates(t *testing.T, names []string, msg string) {
	t.Helper()
	seen := map[string]int{}
	for _, n := range names {
		seen[n]++
	}
	for n, count := range seen {
		assert.Equal(t, 1, count, "%s: %q appears %d times", msg, n, count)
	}
}
