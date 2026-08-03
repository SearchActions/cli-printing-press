package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// mcpStdioTransport speaks JSON-RPC to an MCP server running as a child
// process. Most MCP servers ship this way — a pip or npm package the client
// launches — so a catalog fetch that only knows HTTP cannot reach them.
//
// Framing is newline-delimited JSON per the MCP stdio transport: one request
// per line on stdin, one response per line on stdout. stderr is drained
// separately and kept only for error messages, since servers log there.
type mcpStdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader

	mu     sync.Mutex
	nextID int64

	stderrMu  sync.Mutex
	stderrLog strings.Builder
}

// maxStdioLine bounds a single JSON-RPC line. A large tool catalog is
// legitimately hundreds of KB, so the ceiling is generous; without one, a
// server that streams unbounded output would exhaust memory.
const maxStdioLine = 32 << 20

// startMCPStdio launches the server and returns a transport ready for
// initialize. env names are forwarded from the parent environment; their values
// are never recorded anywhere the capture can reach.
func startMCPStdio(ctx context.Context, launch mcpStdioLaunch) (*mcpStdioTransport, error) {
	if strings.TrimSpace(launch.Command) == "" {
		return nil, errors.New("--command is required")
	}
	// exec.Command, not a shell: the command and its arguments are passed as a
	// vector so a value carrying spaces, quotes, or a semicolon cannot become
	// another command.
	cmd := exec.CommandContext(ctx, launch.Command, launch.Args...)
	cmd.Env = os.Environ()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("opening stdin for %s: %w", launch.Command, err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("opening stdout for %s: %w", launch.Command, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("opening stderr for %s: %w", launch.Command, err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting %s: %w", launch.Command, err)
	}

	t := &mcpStdioTransport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReaderSize(stdout, 1<<20),
	}
	// Drain stderr continuously. Left unread, a chatty server fills the pipe
	// buffer and blocks on its next log line, which reads as a hung handshake.
	go t.drainStderr(stderr)
	return t, nil
}

type mcpStdioLaunch struct {
	Command string
	Args    []string
}

func (t *mcpStdioTransport) drainStderr(r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for scanner.Scan() {
		t.stderrMu.Lock()
		// Keep only the tail: a server that logs every request would otherwise
		// grow this without bound over a long capture.
		if t.stderrLog.Len() > 64<<10 {
			t.stderrLog.Reset()
		}
		t.stderrLog.WriteString(scanner.Text())
		t.stderrLog.WriteByte('\n')
		t.stderrMu.Unlock()
	}
}

func (t *mcpStdioTransport) stderrTail() string {
	t.stderrMu.Lock()
	defer t.stderrMu.Unlock()
	return strings.TrimSpace(t.stderrLog.String())
}

func (t *mcpStdioTransport) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.nextID++
	if err := t.write(map[string]any{
		"jsonrpc": "2.0",
		"id":      t.nextID,
		"method":  method,
		"params":  params,
	}); err != nil {
		return nil, fmt.Errorf("%s: %w", method, t.wrapPipeError(err))
	}

	// Skip notifications and responses to other ids: a server may interleave
	// log notifications with the reply we are waiting for.
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line, err := t.readLine()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", method, t.wrapPipeError(err))
		}
		var envelope struct {
			ID     *int64          `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(line, &envelope); err != nil {
			// Not JSON at all: a server that printed a banner to stdout. Skip
			// rather than fail, since the real reply may still follow.
			continue
		}
		if envelope.ID == nil || *envelope.ID != t.nextID {
			continue
		}
		if envelope.Error != nil {
			return nil, fmt.Errorf("%s: mcp error %d: %s", method, envelope.Error.Code, envelope.Error.Message)
		}
		return envelope.Result, nil
	}
}

func (t *mcpStdioTransport) notify(_ context.Context, method string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	_ = t.write(map[string]any{"jsonrpc": "2.0", "method": method})
}

func (t *mcpStdioTransport) write(payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = t.stdin.Write(append(encoded, '\n'))
	return err
}

func (t *mcpStdioTransport) readLine() ([]byte, error) {
	var buf []byte
	for {
		chunk, isPrefix, err := t.stdout.ReadLine()
		if err != nil {
			return nil, err
		}
		buf = append(buf, chunk...)
		if len(buf) > maxStdioLine {
			return nil, fmt.Errorf("response line exceeded %d bytes", maxStdioLine)
		}
		if !isPrefix {
			return buf, nil
		}
	}
}

// wrapPipeError turns a broken pipe or EOF into the actionable failure it
// really is. A server that dies during startup — a missing dependency, a bad
// interpreter — is otherwise reported as "the pipe is being closed", which
// names the symptom and hides the cause sitting in the child's stderr.
func (t *mcpStdioTransport) wrapPipeError(err error) error {
	if tail := t.stderrTail(); tail != "" {
		return fmt.Errorf("%w (server stderr: %s)", err, truncateForError(tail))
	}
	return err
}

func (t *mcpStdioTransport) close() error {
	// Closing stdin is the MCP shutdown signal; a well-behaved server exits on
	// EOF. Kill covers the ones that do not.
	_ = t.stdin.Close()
	if t.cmd.Process != nil {
		_ = t.cmd.Process.Kill()
	}
	_ = t.cmd.Wait()
	return nil
}
