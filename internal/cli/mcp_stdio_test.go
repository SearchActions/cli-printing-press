package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelperMCPStdioServer is a scripted MCP server, re-executing this test
// binary rather than depending on a Python or Node toolchain being installed.
// It runs only when the parent sets GO_HELPER_MCP_SERVER.
func TestHelperMCPStdioServer(t *testing.T) {
	mode := os.Getenv("GO_HELPER_MCP_SERVER")
	if mode == "" {
		t.Skip("helper process; runs only when re-executed by a test")
	}

	if mode == "die" {
		// A server that fails on startup: the shape of a missing dependency or
		// a bad interpreter, which must surface as more than "broken pipe".
		fmt.Fprintln(os.Stderr, "ModuleNotFoundError: No module named 'mcp.server.fastmcp'")
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64<<10), 8<<20)
	for scanner.Scan() {
		var req struct {
			ID     *int64          `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}
		// A notification carries no id and must never be answered: a reply to
		// one would be read as the answer to the next real request.
		if req.ID == nil {
			continue
		}
		switch req.Method {
		case "initialize":
			// Interleave a log notification, the way a real server does.
			fmt.Println(`{"jsonrpc":"2.0","method":"notifications/message","params":{"level":"info"}}`)
			fmt.Printf(`{"jsonrpc":"2.0","id":%d,"result":{"protocolVersion":"2025-06-18","serverInfo":{"name":"helper-mcp","version":"9.9.9"}}}`+"\n", *req.ID)
		case "tools/list":
			var p struct {
				Cursor string `json:"cursor"`
			}
			_ = json.Unmarshal(req.Params, &p)
			if p.Cursor == "" {
				fmt.Printf(`{"jsonrpc":"2.0","id":%d,"result":{"tools":[{"name":"list_records","description":"List records.","annotations":{"readOnlyHint":true},"inputSchema":{"type":"object","properties":{"limit":{"type":"integer"}}}}],"nextCursor":"page2"}}`+"\n", *req.ID)
			} else {
				fmt.Printf(`{"jsonrpc":"2.0","id":%d,"result":{"tools":[{"name":"delete_record","description":"Delete a record.","annotations":{"destructiveHint":true},"inputSchema":{"type":"object","properties":{"record_id":{"type":"string"}},"required":["record_id"]}}]}}`+"\n", *req.ID)
			}
		default:
			fmt.Printf(`{"jsonrpc":"2.0","id":%d,"result":{}}`+"\n", *req.ID)
		}
	}
}

func helperServerLaunch() (string, []string) {
	return os.Args[0], []string{"-test.run", "^TestHelperMCPStdioServer$", "-test.v=false"}
}

func runSniffWithHelper(t *testing.T, mode, outputPath string) (string, string, error) {
	t.Helper()
	t.Setenv("GO_HELPER_MCP_SERVER", mode)

	command, args := helperServerLaunch()
	var stdout, stderr bytes.Buffer
	req := mcpSniffRequest{
		Command:    command,
		Args:       args,
		OutputPath: outputPath,
		APIName:    "helper-mcp",
		ForwardEnv: []string{"HELPER_TOOL_PATH"},
		ReadyTool:  "list_records",
	}
	err := runMCPSniff(context.Background(), req, mcpSniffOptions{stdout: &stdout, stderr: &stderr})
	return stdout.String(), stderr.String(), err
}

// The load-bearing claim of the stdio path: a server that exists only as a
// local process can be captured at all, and a paginated catalog is captured
// whole rather than truncated to page one.
func TestMCPSniffCapturesAStdioServerWholeCatalog(t *testing.T) {
	out := filepath.Join(t.TempDir(), "helper-mcp.json")
	stdout, _, err := runSniffWithHelper(t, "serve", out)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Captured 2 tools")
	assert.Contains(t, stdout, "transport:  stdio")
	assert.NotContains(t, stdout, "base URL", "a subprocess has no HTTP origin to report")

	raw, err := os.ReadFile(out)
	require.NoError(t, err)

	var capture struct {
		ServerURL string `json:"server_url"`
		Stdio     struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
			Env     []string `json:"env"`
			Ready   string   `json:"ready_tool"`
		} `json:"stdio"`
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	require.NoError(t, json.Unmarshal(raw, &capture))

	assert.Empty(t, capture.ServerURL)
	assert.NotEmpty(t, capture.Stdio.Command, "the launch command is the only way to reach the server again")
	assert.Equal(t, []string{"HELPER_TOOL_PATH"}, capture.Stdio.Env)
	assert.Equal(t, "list_records", capture.Stdio.Ready)
	require.Len(t, capture.Tools, 2, "both catalog pages must be captured")
	assert.Equal(t, "list_records", capture.Tools[0].Name)
	assert.Equal(t, "delete_record", capture.Tools[1].Name)
}

// A server that dies during startup otherwise reports only "broken pipe",
// naming the symptom while the cause sits unread in the child's stderr.
func TestMCPSniffSurfacesAStdioServerThatDiesOnStartup(t *testing.T) {
	out := filepath.Join(t.TempDir(), "dead-mcp.json")
	_, _, err := runSniffWithHelper(t, "die", out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ModuleNotFoundError",
		"the child's own stderr is what tells the operator what to fix")

	_, statErr := os.Stat(out)
	assert.True(t, os.IsNotExist(statErr), "a failed capture must not leave a file behind")
}

// --token and --header describe an HTTP request. Accepting them silently on a
// stdio capture would imply a credential is in play when none is.
func TestMCPSniffRejectsHTTPAuthFlagsOnAStdioCapture(t *testing.T) {
	_, _, err := func() (string, string, error) {
		var stdout, stderr bytes.Buffer
		command, args := helperServerLaunch()
		err := runMCPSniff(context.Background(), mcpSniffRequest{
			Command: command,
			Args:    args,
			Token:   "not-a-real-token",
		}, mcpSniffOptions{stdout: &stdout, stderr: &stderr})
		return stdout.String(), stderr.String(), err
	}()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authenticates with nothing")
}

// --url and --command select different transports; taking both would leave the
// choice to whichever branch happened to be checked first.
func TestMCPSniffFlagsAreMutuallyExclusiveAndOneIsRequired(t *testing.T) {
	cmd := newMCPSniffCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	cmd.SetArgs([]string{"--url", "https://mcp.example.com/mcp", "--command", "uvx"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "none of the others can be")

	cmd = newMCPSniffCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "least one")
}

// The command and its arguments reach exec as a vector. A value carrying a
// semicolon must stay one argument rather than becoming another command.
func TestStdioTransportPassesArgumentsWithoutAShell(t *testing.T) {
	transport, err := startMCPStdio(context.Background(), mcpStdioLaunch{
		Command: os.Args[0],
		Args:    []string{"-test.run", "^TestHelperMCPStdioServer$; echo pwned"},
	})
	require.NoError(t, err)
	defer func() { _ = transport.close() }()

	// The child receives the whole string as one -test.run pattern, matches no
	// test, and exits. What must not happen is `echo pwned` running.
	_, callErr := transport.call(context.Background(), "tools/list", map[string]any{})
	require.Error(t, callErr)
	assert.NotContains(t, callErr.Error(), "pwned")
}

// A launch command that does not exist must fail at start, not hang.
func TestStdioTransportFailsFastOnAMissingExecutable(t *testing.T) {
	_, err := startMCPStdio(context.Background(), mcpStdioLaunch{
		Command: filepath.Join(t.TempDir(), "definitely-not-installed"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "starting")
}

// Guard the helper-process idiom itself: if the test binary stops accepting
// -test.run the scripted server silently becomes a no-op and every stdio test
// above would pass for the wrong reason.
func TestHelperProcessIsReachable(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run", "^TestHelperMCPStdioServer$")
	cmd.Env = append(os.Environ(), "GO_HELPER_MCP_SERVER=die")
	output, err := cmd.CombinedOutput()
	require.Error(t, err, "the die mode must exit non-zero")
	assert.Contains(t, string(output), "ModuleNotFoundError")
}
