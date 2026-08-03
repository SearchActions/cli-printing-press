package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mvanhorn/cli-printing-press/v4/internal/mcpspec"
)

func TestValidateMCPServerURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr string
	}{
		{name: "https accepted", url: "https://mcp.example.com/mcp"},
		{name: "loopback http accepted", url: "http://127.0.0.1:8791/mcp"},
		{name: "localhost http accepted", url: "http://localhost:3000/mcp"},
		{name: "plaintext remote rejected", url: "http://mcp.example.com/mcp", wantErr: "must be https"},
		{name: "empty rejected", url: "", wantErr: "--url is required"},
		// The capture file records the server URL, so embedded credentials
		// would be written to disk at 0644.
		{name: "embedded credentials rejected", url: "https://alice:hunter2@mcp.example.com/mcp", wantErr: "must not embed credentials"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMCPServerURL(tc.url)
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestParseHeaderFlags(t *testing.T) {
	got, err := parseHeaderFlags([]string{"Authorization: Bearer abc", "x-tenant:  acme  "})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"Authorization": "Bearer abc",
		"X-Tenant":      "acme",
	}, got, "names canonicalize and values trim")

	_, err = parseHeaderFlags([]string{"NoColonHere"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name: value")

	_, err = parseHeaderFlags([]string{": novalue"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty name")
}

func TestExtractSSEPayload(t *testing.T) {
	// A streamable-HTTP server may answer a single JSON-RPC call as an SSE
	// event; plain JSON must pass through untouched.
	sse := "event: message\r\ndata: {\"jsonrpc\":\"2.0\",\"id\":1}\r\n\r\n"
	assert.JSONEq(t, `{"jsonrpc":"2.0","id":1}`, string(extractSSEPayload([]byte(sse))))

	plain := []byte(`{"jsonrpc":"2.0","id":2}`)
	assert.Equal(t, plain, extractSSEPayload(plain))

	garbage := []byte("not json and not sse")
	assert.Equal(t, garbage, extractSSEPayload(garbage))
}

// newFakeMCPServer serves initialize and a paginated tools/list.
func newFakeMCPServer(t *testing.T, requireAuth bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requireAuth && r.Header.Get("Authorization") == "" {
			w.Header().Set("WWW-Authenticate", `Bearer scope="tools:read"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var req struct {
			ID     int64           `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		// require calls runtime.Goexit on failure, which from a handler
		// goroutine leaves the test hanging rather than failing. Report and
		// return instead.
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		switch req.Method {
		case "initialize":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0", "id": req.ID,
				"result": map[string]any{
					"serverInfo": map[string]any{"name": "fake-server", "title": "Fake Server"},
				},
			})
		case "tools/list":
			var p struct {
				Cursor string `json:"cursor"`
			}
			_ = json.Unmarshal(req.Params, &p)
			// First page returns a cursor so the pagination loop is exercised.
			if p.Cursor == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"jsonrpc": "2.0", "id": req.ID,
					"result": map[string]any{
						"tools": []map[string]any{{
							"name":        "list_widgets",
							"description": "List widgets.",
							"annotations": map[string]any{"readOnlyHint": true},
							"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
						}},
						"nextCursor": "page2",
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0", "id": req.ID,
				"result": map[string]any{
					"tools": []map[string]any{{
						"name":        "delete_widget",
						"description": "Delete a widget.",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{"id": map[string]any{"type": "string"}},
							"required":   []string{"id"},
						},
					}},
				},
			})
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func TestRunMCPSniffCapturesAllPages(t *testing.T) {
	srv := newFakeMCPServer(t, false)
	defer srv.Close()

	out := filepath.Join(t.TempDir(), "capture.json")
	var stdout, stderr bytes.Buffer
	err := runMCPSniff(context.Background(), mcpSniffRequest{
		ServerURL:  srv.URL + "/mcp",
		OutputPath: out,
	}, mcpSniffOptions{stdout: &stdout, stderr: &stderr, client: srv.Client()})
	require.NoError(t, err)

	data, err := os.ReadFile(out)
	require.NoError(t, err)

	var capture struct {
		ServerURL string           `json:"server_url"`
		Tools     []map[string]any `json:"tools"`
	}
	require.NoError(t, json.Unmarshal(data, &capture))
	assert.Equal(t, srv.URL+"/mcp", capture.ServerURL)
	require.Len(t, capture.Tools, 2, "both pages of tools/list are captured, not just the first")
	assert.Equal(t, "list_widgets", capture.Tools[0]["name"])
	assert.Equal(t, "delete_widget", capture.Tools[1]["name"])

	// The capture must be directly generatable.
	assert.True(t, mcpspec.IsMCPToolsList(data), "capture must be recognized as an MCP catalog")
	assert.Contains(t, stdout.String(), "Captured 2 tools")
}

func TestRunMCPSniffSurfacesAuthChallenge(t *testing.T) {
	srv := newFakeMCPServer(t, true)
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	err := runMCPSniff(context.Background(), mcpSniffRequest{
		ServerURL:  srv.URL + "/mcp",
		OutputPath: filepath.Join(t.TempDir(), "capture.json"),
	}, mcpSniffOptions{stdout: &stdout, stderr: &stderr, client: srv.Client()})

	require.Error(t, err)
	// The operator needs to know it is an auth problem and what scope is
	// wanted, not a bare HTTP 401.
	assert.Contains(t, err.Error(), "requires authorization")
	assert.Contains(t, err.Error(), "tools:read")
}

func TestRunMCPSniffTokenReachesServer(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var req struct {
			ID     int64  `json:"id"`
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0", "id": req.ID,
			"result": map[string]any{"tools": []map[string]any{{
				"name":        "get_thing",
				"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
			}}},
		})
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	err := runMCPSniff(context.Background(), mcpSniffRequest{
		ServerURL:  srv.URL + "/mcp",
		OutputPath: filepath.Join(t.TempDir(), "c.json"),
		Token:      "secret-token",
	}, mcpSniffOptions{stdout: &stdout, stderr: &stderr, client: srv.Client()})
	require.NoError(t, err)
	assert.Equal(t, "Bearer secret-token", gotAuth)
}

func TestRunMCPSniffRejectsConflictingAuthFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runMCPSniff(context.Background(), mcpSniffRequest{
		ServerURL: "https://mcp.example.com/mcp",
		Token:     "abc",
		Headers:   []string{"Authorization: Bearer def"},
	}, mcpSniffOptions{stdout: &stdout, stderr: &stderr})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestRunMCPSniffRejectsPlaintextRemote(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runMCPSniff(context.Background(), mcpSniffRequest{
		ServerURL: "http://mcp.example.com/mcp",
	}, mcpSniffOptions{stdout: &stdout, stderr: &stderr})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be https")
}
