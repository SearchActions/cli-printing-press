package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	assert.Contains(t, src, `"verify_noop":true`)
	// The live-HTTP opt-out means "verify owns an httptest mock server, so let
	// the real wire path run against it." There is no mock subprocess, and
	// verify sets that var on every mock-mode subprocess — honoring it here
	// makes the gate inert exactly where it is needed.
	assert.Contains(t, src, "if !readOnly && cliutil.IsVerifyEnv() {",
		"the gate must be exactly this condition")
	for _, optOut := range []string{
		"cliutil.IsVerifyEnv() && os.Getenv(cliutil.VerifyLiveHTTPEnvVar)",
		"cliutil.IsVerifyEnv() && !cliutil.IsVerifyLiveHTTPEnv()",
	} {
		assert.NotContains(t, src, optOut,
			"the stdio gate must not honor the live-HTTP opt-out")
	}
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

// Behavioral counterpart to the source-grep test above: run the built binary
// under the exact environment `cli-printing-press verify` sets on every
// mock-mode subprocess, which is BOTH PRINTING_PRESS_VERIFY=1 and
// PRINTING_PRESS_VERIFY_LIVE_HTTP=1 (internal/pipeline/runtime.go).
//
// A gate that also honors the live-HTTP opt-out reads as correct in the source
// and is inert exactly here, sending a real mutation to the operator's own
// machine. The launch command is pointed at a path that does not exist, so if
// the gate ever stops firing the process spawn fails loudly instead of
// silently passing.
func TestGeneratedMCPStdioCLIShortCircuitsMutationsUnderVerify(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds a generated module")
	}
	outputDir, _ := generateStdioMCPCLI(t)
	requireGeneratedCompiles(t, outputDir)

	binary := filepath.Join(outputDir, "stdio-cli"+exeSuffix())
	runGoCommand(t, outputDir, "build", "-o", binary, "./cmd/example-stdio-pp-cli")

	missingServer := filepath.Join(t.TempDir(), "no-such-mcp-server")
	env := append(os.Environ(),
		"PRINTING_PRESS_VERIFY=1",
		"PRINTING_PRESS_VERIFY_LIVE_HTTP=1",
		"EXAMPLE_STDIO_MCP_COMMAND="+missingServer,
	)

	cmd := exec.Command(binary, "records", "delete", "--id", "some-record-id", "--json", "--yes")
	cmd.Env = env
	output, _ := cmd.CombinedOutput()
	got := string(output)

	assert.NotContains(t, got, "starting the MCP server",
		"the gate must fire before the subprocess is spawned; verify sets LIVE_HTTP and there is no mock subprocess to point it at")
	assert.NotContains(t, got, "no-such-mcp-server",
		"nothing should have tried to launch the server")
	assert.Contains(t, got, "verify_noop",
		"a mutating tool call under verify must return the synthetic noop envelope")
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// Windows does not reap the child when the parent exits, so a CLI that never
// shuts the server down leaks one process per invocation.
func TestGenerateMCPStdioShutsTheServerDown(t *testing.T) {
	t.Parallel()
	outputDir, _ := generateStdioMCPCLI(t)

	mainSrc, err := os.ReadFile(filepath.Join(outputDir, "cmd", "example-stdio-pp-cli", "main.go"))
	require.NoError(t, err)
	main := string(mainSrc)

	assert.Contains(t, main, "defer client.ShutdownMCP()")
	// The error path calls os.Exit, which skips deferred calls.
	assert.Contains(t, main, "client.ShutdownMCP()\n\t\tos.Exit(")
}

// An HTTP-backed CLI has no child process and must not import the client
// package just to call a shutdown that does nothing.
func TestGenerateMCPHTTPMainHasNoShutdown(t *testing.T) {
	t.Parallel()
	apiSpec, err := mcpspec.Parse("example-http-main", []byte(`{
  "server_url": "https://mcp.example.com/mcp",
  "tools": [{"name": "list_records", "description": "List.", "inputSchema": {"type": "object", "properties": {}}}]
}`), mcpspec.ParseOptions{})
	require.NoError(t, err)

	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())

	mains, err := filepath.Glob(filepath.Join(outputDir, "cmd", "*", "main.go"))
	require.NoError(t, err)
	require.NotEmpty(t, mains)
	for _, path := range mains {
		mainSrc, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		assert.NotContains(t, string(mainSrc), "ShutdownMCP", "in %s", path)
	}
}
