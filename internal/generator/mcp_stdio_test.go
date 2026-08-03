package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mvanhorn/cli-printing-press/v4/internal/mcpspec"
	"github.com/mvanhorn/cli-printing-press/v4/internal/naming"
)

// stdioCatalog is a two-tool catalog whose server is launched as a subprocess.
const stdioCatalog = `{
  "stdio": {"command": "uvx", "args": ["--from", "example-mcp", "example-mcp"], "env": ["EXAMPLE_CLI_PATH"], "ready_tool": "check_env"},
  "server_info": {"name": "example-stdio", "version": "0.1.0"},
  "tools": [
    {
      "name": "list_records",
      "description": "List stored records.",
      "annotations": {"readOnlyHint": true},
      "inputSchema": {"type": "object", "properties": {"limit": {"type": "integer"}}}
    },
    {
      "name": "delete_record",
      "description": "Delete a stored record.",
      "annotations": {"destructiveHint": true},
      "inputSchema": {"type": "object", "properties": {"id": {"type": "string"}}, "required": ["id"]}
    }
  ]
}`

func generateStdioMCPCLI(t *testing.T) (string, string) {
	t.Helper()
	apiSpec, err := mcpspec.Parse("example-stdio", []byte(stdioCatalog), mcpspec.ParseOptions{})
	require.NoError(t, err)
	require.NoError(t, apiSpec.Validate())

	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	gen := New(apiSpec, outputDir)
	require.NoError(t, gen.Generate())

	stdioSrc, err := os.ReadFile(filepath.Join(outputDir, "internal", "client", "mcp_stdio.go"))
	require.NoError(t, err, "a stdio-backed spec must emit the subprocess transport")
	return outputDir, string(stdioSrc)
}

// The emitted transport is new code on a new path; asserting on template text
// would not catch an emitted call site that does not compile.
func TestGenerateMCPStdioTransportCompiles(t *testing.T) {
	t.Parallel()
	outputDir, _ := generateStdioMCPCLI(t)
	requireGeneratedCompiles(t, outputDir)
}

// The launch command is per-server data. Hardcoding or dropping it leaves a CLI
// with nothing to start.
func TestGenerateMCPStdioRecordsTheLaunchCommand(t *testing.T) {
	t.Parallel()
	_, src := generateStdioMCPCLI(t)

	assert.Contains(t, src, `mcpStdioCommand = "uvx"`)
	assert.Contains(t, src, `"--from"`)
	assert.Contains(t, src, `"example-mcp"`)
	assert.Contains(t, src, `"EXAMPLE_CLI_PATH"`, "forwarded env var names are recorded")
	assert.Contains(t, src, `MCPStdioReadyTool = "check_env"`)
}

// exec.Command with an argument vector, never a shell: a recorded value
// carrying a semicolon must stay one argument instead of becoming a command.
func TestGenerateMCPStdioNeverShellsOut(t *testing.T) {
	t.Parallel()
	_, src := generateStdioMCPCLI(t)

	assert.Contains(t, src, "exec.Command(name, args...)")
	for _, shell := range []string{`exec.Command("sh"`, `exec.Command("cmd"`, `exec.Command("bash"`, "/bin/sh", "-c\""} {
		assert.NotContains(t, src, shell, "the launch must not route through a shell")
	}
}

// The verify-mode mutation gate lives in the HTTP client's doInternal, which
// this transport bypasses. Without a gate here, a verify run would execute real
// mutations against the operator's own machine.
func TestGenerateMCPStdioGatesMutationsUnderVerify(t *testing.T) {
	t.Parallel()
	_, src := generateStdioMCPCLI(t)

	assert.Contains(t, src, "cliutil.IsVerifyEnv()")
	assert.Contains(t, src, "cliutil.VerifyLiveHTTPEnvVar")
	assert.Contains(t, src, `"verify_noop":true`)
}

// Both transports define mcpRoundTrip. Emitting both files would be a
// duplicate-symbol compile error, so the HTTP one must be conditional.
func TestGenerateMCPStdioOmitsTheHTTPRoundTrip(t *testing.T) {
	t.Parallel()
	outputDir, _ := generateStdioMCPCLI(t)

	shared, err := os.ReadFile(filepath.Join(outputDir, "internal", "client", "mcp.go"))
	require.NoError(t, err)
	assert.NotContains(t, string(shared), `c.doInternal(ctx, "POST", mcpEndpointPath`,
		"the HTTP round trip must not be emitted alongside the stdio one")
	assert.Contains(t, string(shared), "c.mcpRoundTrip(",
		"the shared lifecycle still routes every call through the seam")
}

// An HTTP-backed MCP spec must be untouched by the stdio work.
func TestGenerateMCPHTTPStillEmitsTheHTTPRoundTrip(t *testing.T) {
	t.Parallel()
	apiSpec, err := mcpspec.Parse("example-http", []byte(`{
  "server_url": "https://mcp.example.com/mcp",
  "tools": [{"name": "list_records", "description": "List.", "inputSchema": {"type": "object", "properties": {}}}]
}`), mcpspec.ParseOptions{})
	require.NoError(t, err)
	require.NoError(t, apiSpec.Validate())

	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())

	_, statErr := os.Stat(filepath.Join(outputDir, "internal", "client", "mcp_stdio.go"))
	assert.True(t, os.IsNotExist(statErr), "an HTTP server must not get the subprocess transport")

	shared, err := os.ReadFile(filepath.Join(outputDir, "internal", "client", "mcp.go"))
	require.NoError(t, err)
	assert.Contains(t, string(shared), `c.doInternal(ctx, "POST", mcpEndpointPath`)
}

// A stdio server has no URL, so the HTTP doctor's headline check told the
// operator to set a base_url that must not exist. The only real health signal
// is whether the server starts and answers.
func TestGenerateMCPStdioDoctorProbesTheServerNotABaseURL(t *testing.T) {
	t.Parallel()
	outputDir, _ := generateStdioMCPCLI(t)

	doctorSrc, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "doctor.go"))
	require.NoError(t, err)
	doctor := string(doctorSrc)

	assert.NotContains(t, doctor, "not configured (set base_url in config file)",
		"a subprocess has no base URL to configure")
	assert.Contains(t, doctor, "client.MCPStdioReadyTool",
		"doctor must call the declared read-only tool")
	assert.Contains(t, doctor, "MCP server unreachable")
	assert.Contains(t, doctor, "client.MCPStdioForwardEnv",
		"unset server env vars explain a server-side error that looks like a CLI bug")
}

// The HTTP doctor must keep its reachability and credential checks.
func TestGenerateMCPHTTPDoctorKeepsTheBaseURLCheck(t *testing.T) {
	t.Parallel()
	apiSpec, err := mcpspec.Parse("example-http-doctor", []byte(`{
  "server_url": "https://mcp.example.com/mcp",
  "tools": [{"name": "list_records", "description": "List.", "inputSchema": {"type": "object", "properties": {}}}]
}`), mcpspec.ParseOptions{})
	require.NoError(t, err)

	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())

	doctorSrc, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "doctor.go"))
	require.NoError(t, err)
	assert.Contains(t, string(doctorSrc), "not configured (set base_url in config file)")
}
