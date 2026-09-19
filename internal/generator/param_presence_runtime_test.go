package generator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/naming"
	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/require"
)

// capturedRequest is one request the fixture server observed, decoded enough
// for the wire-proof assertions below: query values, a best-effort JSON body
// decode, and a best-effort multipart/form decode.
type capturedRequest struct {
	Method      string
	Path        string
	Query       map[string][]string
	JSONBody    map[string]any
	HasJSONBody bool
	Form        map[string][]string
}

type wireProofServer struct {
	mu       sync.Mutex
	requests []capturedRequest
	server   *httptest.Server
}

func newWireProofServer(t *testing.T) *wireProofServer {
	t.Helper()
	w := &wireProofServer{}
	w.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		captured := capturedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  map[string][]string(r.URL.Query()),
		}
		contentType := r.Header.Get("Content-Type")
		switch {
		case strings.HasPrefix(contentType, "multipart/form-data"):
			if err := r.ParseMultipartForm(1 << 20); err == nil {
				captured.Form = map[string][]string(r.MultipartForm.Value)
			}
		case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
			if err := r.ParseForm(); err == nil {
				captured.Form = map[string][]string(r.PostForm)
			}
		case strings.HasPrefix(contentType, "application/json"):
			var body map[string]any
			if json.NewDecoder(r.Body).Decode(&body) == nil {
				captured.JSONBody = body
				captured.HasJSONBody = true
			}
		}
		w.mu.Lock()
		w.requests = append(w.requests, captured)
		w.mu.Unlock()
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(`{}`))
	}))
	t.Cleanup(w.server.Close)
	return w
}

func (w *wireProofServer) last(t *testing.T) capturedRequest {
	t.Helper()
	w.mu.Lock()
	defer w.mu.Unlock()
	require.NotEmpty(t, w.requests, "expected the fixture server to have received a request")
	return w.requests[len(w.requests)-1]
}

func (w *wireProofServer) count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.requests)
}

// buildGeneratedBinaryExe mirrors buildGeneratedBinary (doctor_paths_test.go)
// but appends ".exe" on Windows, matching the pattern help_exit_test.go uses:
// buildGeneratedBinary's own extensionless path fails os/exec on Windows with
// "executable file not found in %PATH%" (CreateProcess needs a recognized
// extension even given a full path).
func buildGeneratedBinaryExe(t *testing.T, apiSpec *spec.APISpec) string {
	t.Helper()
	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())
	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	binaryPath := filepath.Join(outputDir, naming.CLI(apiSpec.Name)+exe)
	runGoCommand(t, outputDir, "build", "-o", binaryPath, "./cmd/"+naming.CLI(apiSpec.Name))
	return binaryPath
}

// runGeneratedBinaryAllowFailure is runGeneratedBinary's shape without the
// require.NoError: case (e) below expects a non-zero exit (a required-flag
// validation error), and runGeneratedBinary intentionally cannot express that.
func runGeneratedBinaryAllowFailure(t *testing.T, binaryPath string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// paramPresenceWireProofSpec is the fixture spec for Strategy D: one spec
// exercising every site class the plan brought into scope, wired against a
// real httptest.Server so the assertions below prove wire behavior, not just
// emitted text.
func paramPresenceWireProofSpec() *spec.APISpec {
	apiSpec := minimalSpec("presence-wire")
	apiSpec.Resources = map[string]spec.Resource{
		"records": {
			Description: "JSON-body resource",
			Endpoints: map[string]spec.Endpoint{
				// A second endpoint keeps "records" from single-endpoint
				// promotion so "create" exercises command_endpoint.go.tmpl.
				"get": {
					Method:      "GET",
					Path:        "/records/{id}",
					Description: "Get a record",
					Params: []spec.Param{
						{Name: "id", Type: "string", Positional: true, PathParam: true, Required: true},
					},
				},
				"create": {
					Method:      "POST",
					Path:        "/records",
					Description: "Create a record",
					Body: []spec.Param{
						{Name: "name", Type: "string", Required: true},
						{Name: "count", Type: "int"},
						{
							Name: "viewport",
							Type: "object",
							Fields: []spec.Param{
								{Name: "width", Type: "int", Required: true},
								{Name: "height", Type: "int", Required: true},
							},
						},
						{Name: "meta", Type: "object"},
						{Name: "featured", Type: "boolean", Aliases: []string{"is-featured"}},
						{Name: "priority", Type: "int", Aliases: []string{"pri"}},
					},
				},
			},
		},
		"reads": {
			Description: "Non-paginated GET resource",
			Endpoints: map[string]spec.Endpoint{
				"get": {
					Method:      "GET",
					Path:        "/reads/{id}",
					Description: "Get a read",
					Params: []spec.Param{
						{Name: "id", Type: "string", Positional: true, PathParam: true, Required: true},
					},
				},
				"list": {
					Method:      "GET",
					Path:        "/reads",
					Description: "List reads",
					Params: []spec.Param{
						{Name: "limit", Type: "int", Default: 50},
						{Name: "verbose", Type: "boolean", Default: true},
						{Name: "offset", Type: "int"},
						{Name: "q", Type: "string"},
					},
				},
			},
		},
		"pages": {
			Description: "Paginated GET resource",
			Endpoints: map[string]spec.Endpoint{
				"get": {
					Method:      "GET",
					Path:        "/pages/{id}",
					Description: "Get a page",
					Params: []spec.Param{
						{Name: "id", Type: "string", Positional: true, PathParam: true, Required: true},
					},
				},
				"list": {
					Method:      "GET",
					Path:        "/pages",
					Description: "List pages",
					Pagination: &spec.Pagination{
						Type:           "cursor",
						LimitParam:     "limit",
						CursorParam:    "after",
						NextCursorPath: "next_cursor",
						HasMoreField:   "has_more",
					},
					Params: []spec.Param{
						{Name: "after", Type: "string", Description: "Cursor"},
						{Name: "active", Type: "boolean"},
						{Name: "score", Type: "int"},
					},
				},
			},
		},
		"posts": {
			Description: "Multipart POST resource",
			Endpoints: map[string]spec.Endpoint{
				"list": {
					Method:      "GET",
					Path:        "/posts",
					Description: "List posts",
				},
				"create": {
					Method:             "POST",
					Path:               "/posts",
					Description:        "Create a post via multipart",
					RequestContentType: "multipart/form-data",
					Body: []spec.Param{
						{Name: "field", Type: "string"},
						{Name: "n", Type: "int"},
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
					Description: "List promo",
					Params: []spec.Param{
						{Name: "level", Type: "int"},
					},
				},
			},
		},
	}
	return apiSpec
}

// TestParamPresenceWireProof is the runtime companion to the emitted-text
// tests: it builds a real generated binary and drives it against a real
// httptest.Server, proving the presence gates change what actually reaches
// the wire (query strings, JSON bodies, multipart form fields), not just
// what the generator prints.
func TestParamPresenceWireProof(t *testing.T) {
	if testing.Short() {
		t.Skip("generated CLI compile/build tests run in the full generated-test CI lane")
	}

	apiSpec := paramPresenceWireProofSpec()
	binaryPath := buildGeneratedBinaryExe(t, apiSpec)

	srv := newWireProofServer(t)
	home := t.TempDir()
	env := append(os.Environ(),
		"HOME="+home,
		"MYAPI_TOKEN=test-token",
		"PRESENCE_WIRE_BASE_URL="+srv.server.URL,
	)

	run := func(t *testing.T, args ...string) (string, string) {
		t.Helper()
		cmd := exec.Command(binaryPath, args...)
		cmd.Env = env
		var outBuf, errBuf strings.Builder
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		require.NoError(t, cmd.Run(), "stdout:\n%s\nstderr:\n%s", outBuf.String(), errBuf.String())
		return outBuf.String(), errBuf.String()
	}

	t.Run("(a) explicit zero/empty values reach the wire", func(t *testing.T) {
		run(t, "records", "create", "--name", "", "--count", "0")
		body := srv.last(t)
		require.True(t, body.HasJSONBody)
		require.Equal(t, "", body.JSONBody["name"])
		require.Equal(t, float64(0), body.JSONBody["count"])

		run(t, "reads", "list", "--verbose=false", "--offset", "0")
		q := srv.last(t).Query
		require.Equal(t, []string{"false"}, q["verbose"])
		require.Equal(t, []string{"0"}, q["offset"])

		run(t, "pages", "list", "--active=false", "--score", "0")
		q = srv.last(t).Query
		require.Equal(t, []string{"false"}, q["active"])
		require.Equal(t, []string{"0"}, q["score"])

		run(t, "posts", "create", "--field", "", "--n", "0")
		form := srv.last(t).Form
		require.Equal(t, []string{""}, form["field"])
		require.Equal(t, []string{"0"}, form["n"])

		run(t, "promo", "--level", "0")
		q = srv.last(t).Query
		require.Equal(t, []string{"0"}, q["level"])
	})

	t.Run("(b) untouched non-zero defaults are still sent", func(t *testing.T) {
		run(t, "reads", "list")
		q := srv.last(t).Query
		require.Equal(t, []string{"50"}, q["limit"])
		require.Equal(t, []string{"true"}, q["verbose"])
	})

	t.Run("(c) untouched params without defaults are still omitted", func(t *testing.T) {
		run(t, "reads", "list")
		q := srv.last(t).Query
		require.NotContains(t, q, "q")
		require.NotContains(t, q, "offset")
	})

	t.Run("(d) hidden aliases reach the wire, including a zero/false value", func(t *testing.T) {
		run(t, "records", "create", "--name", "x", "--pri", "0")
		body := srv.last(t)
		require.True(t, body.HasJSONBody)
		require.Equal(t, float64(0), body.JSONBody["priority"], "int alias with an explicit zero must not be indistinguishable from untouched")

		run(t, "records", "create", "--name", "x", "--is-featured=false")
		body = srv.last(t)
		require.Contains(t, body.JSONBody, "featured")
		require.Equal(t, false, body.JSONBody["featured"])
	})

	t.Run("(f) empty complex leaf omits the field without a JSON parse error", func(t *testing.T) {
		stdout, stderr := run(t, "records", "create", "--name", "x", "--meta", "")
		_ = stdout
		require.NotContains(t, stderr, "parsing")
		body := srv.last(t)
		require.NotContains(t, body.JSONBody, "meta")
	})

	t.Run("(e) nested sibling-required fires only when a sibling was explicitly supplied", func(t *testing.T) {
		before := srv.count()
		stdout, stderr, err := runGeneratedBinaryAllowFailure(t, binaryPath, "records", "create", "--name", "x", "--viewport-width", "0")
		require.Error(t, err, "stdout:\n%s\nstderr:\n%s", stdout, stderr)
		require.Contains(t, stderr, `required flag "viewport-height" not set`)
		require.Equal(t, before, srv.count(), "an invalid nested body must never reach the server")

		// Positive control: omitting the optional parent entirely must
		// still succeed, proving the failure above is about the explicit
		// zero, not about viewport being touched at all.
		run(t, "records", "create", "--name", "x")
		body := srv.last(t)
		require.NotContains(t, body.JSONBody, "viewport")
	})
}
