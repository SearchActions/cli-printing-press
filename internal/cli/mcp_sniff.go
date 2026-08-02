package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/cli-printing-press/v4/internal/mcpspec"
)

type mcpSniffOptions struct {
	stdout io.Writer
	stderr io.Writer
	// client lets tests substitute a transport without reaching the network.
	client *http.Client
}

func newMCPSniffCmd() *cobra.Command {
	return newMCPSniffCmdWithOptions(mcpSniffOptions{})
}

func newMCPSniffCmdWithOptions(opts mcpSniffOptions) *cobra.Command {
	var serverURL string
	var outputPath string
	var token string
	var headers []string
	var apiName string

	cmd := &cobra.Command{
		Use:   "mcp-sniff",
		Short: "Fetch an MCP server's tool catalog and emit a spec-ready capture",
		Long: `Connect to a remote MCP server, negotiate initialize, call tools/list, and
write the advertised tool catalog to disk.

The capture is consumed directly by 'cli-printing-press generate', which turns
each MCP tool into a command. Every tool call becomes a JSON-RPC POST to the
server's single endpoint; read tools keep semantic GET methods so the generated
MCP surface gets correct safety annotations.

Auth: most remote MCP servers are OAuth bearer-protected. Pass --token, or set
the bearer via --header "Authorization: Bearer <token>". Servers behind an
interactive OAuth flow cannot be captured here; obtain a token with that
server's own login flow first.`,
		Example: `  # Capture a bearer-protected server's catalog
  cli-printing-press mcp-sniff --url https://mcp.example.com/mcp --token "$MCP_TOKEN" -o example-mcp.json

  # Generate a CLI from the capture
  cli-printing-press generate example-mcp.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPSniff(cmd.Context(), mcpSniffRequest{
				ServerURL:  serverURL,
				OutputPath: outputPath,
				Token:      token,
				Headers:    headers,
				APIName:    apiName,
			}, opts)
		},
	}

	cmd.Flags().StringVar(&serverURL, "url", "", "MCP server endpoint (e.g. https://mcp.example.com/mcp)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Path to write the tool catalog (default: <name>-mcp.json)")
	cmd.Flags().StringVar(&token, "token", "", "Bearer token for the MCP server")
	cmd.Flags().StringArrayVar(&headers, "header", nil, "Extra request header as 'Name: value' (repeatable)")
	cmd.Flags().StringVar(&apiName, "name", "", "Override the derived API slug")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

type mcpSniffRequest struct {
	ServerURL  string
	OutputPath string
	Token      string
	Headers    []string
	APIName    string
}

// mcpSniffCapture is the on-disk envelope. It records the server URL alongside
// the catalog so `generate` can resolve the base URL and endpoint path without
// the operator re-supplying them.
type mcpSniffCapture struct {
	ServerURL  string          `json:"server_url"`
	ServerInfo json.RawMessage `json:"server_info,omitempty"`
	Tools      []any           `json:"tools"`
}

func runMCPSniff(ctx context.Context, req mcpSniffRequest, opts mcpSniffOptions) error {
	stdout := opts.stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	if err := validateMCPServerURL(req.ServerURL); err != nil {
		return &ExitError{Code: ExitSpecError, Err: err}
	}
	extraHeaders, err := parseHeaderFlags(req.Headers)
	if err != nil {
		return &ExitError{Code: ExitSpecError, Err: err}
	}
	if req.Token != "" {
		if _, taken := extraHeaders["Authorization"]; taken {
			return &ExitError{Code: ExitSpecError, Err: fmt.Errorf("--token and an Authorization --header are mutually exclusive; pass one")}
		}
		extraHeaders["Authorization"] = "Bearer " + req.Token
	}

	httpClient := opts.client
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	sniffer := &mcpSniffer{
		url:     req.ServerURL,
		headers: extraHeaders,
		client:  httpClient,
	}

	// initialize is advisory: stateless servers answer tools/list directly, so
	// a handshake failure must not abort the capture.
	serverInfo, err := sniffer.initialize(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "warning: initialize failed (%v); attempting tools/list anyway\n", err)
	}

	tools, err := sniffer.listTools(ctx)
	if err != nil {
		return &ExitError{Code: ExitInputError, Err: err}
	}
	if len(tools) == 0 {
		return &ExitError{Code: ExitInputError, Err: fmt.Errorf("%s advertises no tools; nothing to generate from", req.ServerURL)}
	}

	capture := mcpSniffCapture{
		ServerURL:  req.ServerURL,
		ServerInfo: serverInfo,
		Tools:      tools,
	}
	encoded, err := json.MarshalIndent(capture, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding capture: %w", err)
	}
	encoded = append(encoded, '\n')

	// Re-parse through the real parser so a capture that cannot produce a spec
	// fails here, at capture time, rather than surfacing later as a confusing
	// generate error.
	parsed, err := mcpspec.Parse(req.ServerURL, encoded, mcpspec.ParseOptions{
		ServerURL: req.ServerURL,
		Name:      req.APIName,
	})
	if err != nil {
		return &ExitError{Code: ExitSpecError, Err: fmt.Errorf("captured catalog is not generatable: %w", err)}
	}

	outputPath := req.OutputPath
	if outputPath == "" {
		outputPath = parsed.Name + "-mcp.json"
	}
	if err := os.WriteFile(outputPath, encoded, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outputPath, err)
	}

	resourceCount := len(parsed.Resources)
	endpointCount := 0
	for _, res := range parsed.Resources {
		endpointCount += len(res.Endpoints)
	}
	fmt.Fprintf(stdout, "Captured %d tools from %s\n", len(tools), req.ServerURL)
	fmt.Fprintf(stdout, "  spec name:  %s\n", parsed.Name)
	fmt.Fprintf(stdout, "  base URL:   %s\n", parsed.BaseURL)
	fmt.Fprintf(stdout, "  MCP path:   %s\n", parsed.MCPEndpointPath)
	fmt.Fprintf(stdout, "  maps to:    %d resources, %d endpoints\n", resourceCount, endpointCount)
	fmt.Fprintf(stdout, "  wrote:      %s\n", outputPath)
	fmt.Fprintf(stdout, "\nNext: cli-printing-press generate %s\n", outputPath)
	return nil
}

func validateMCPServerURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("--url is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parsing --url %q: %w", raw, err)
	}
	if u.Scheme != "https" && !isLoopbackHost(u.Hostname()) {
		return fmt.Errorf("--url must be https (got %q); plaintext would send the bearer token in the clear", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("--url %q has no host", raw)
	}
	return nil
}

// isLoopbackHost allows http:// against a local MCP server under development,
// where there is no network path for a token to leak onto.
func isLoopbackHost(host string) bool {
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1", "[::1]":
		return true
	}
	return false
}

// parseHeaderFlags splits repeated `Name: value` flags into a header map.
func parseHeaderFlags(raw []string) (map[string]string, error) {
	out := map[string]string{}
	for _, h := range raw {
		name, value, found := strings.Cut(h, ":")
		if !found {
			return nil, fmt.Errorf("--header %q must be in 'Name: value' form", h)
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" {
			return nil, fmt.Errorf("--header %q has an empty name", h)
		}
		out[http.CanonicalHeaderKey(name)] = value
	}
	return out, nil
}

type mcpSniffer struct {
	url     string
	headers map[string]string
	client  *http.Client

	sessionID string
	nextID    int64
}

func (s *mcpSniffer) initialize(ctx context.Context) (json.RawMessage, error) {
	result, err := s.call(ctx, "initialize", map[string]any{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "cli-printing-press", "version": "mcp-sniff"},
	})
	if err != nil {
		return nil, err
	}
	var info struct {
		ServerInfo json.RawMessage `json:"serverInfo"`
	}
	if err := json.Unmarshal(result, &info); err != nil {
		return nil, nil
	}
	return info.ServerInfo, nil
}

func (s *mcpSniffer) listTools(ctx context.Context) ([]any, error) {
	var tools []any
	cursor := ""
	// Follow nextCursor so a server that pages its catalog is captured whole
	// rather than silently truncated to the first page.
	for range 100 {
		params := map[string]any{}
		if cursor != "" {
			params["cursor"] = cursor
		}
		result, err := s.call(ctx, "tools/list", params)
		if err != nil {
			return nil, err
		}
		var body struct {
			Tools      []any  `json:"tools"`
			NextCursor string `json:"nextCursor"`
		}
		if err := json.Unmarshal(result, &body); err != nil {
			return nil, fmt.Errorf("decoding tools/list result: %w", err)
		}
		tools = append(tools, body.Tools...)
		if body.NextCursor == "" {
			return tools, nil
		}
		cursor = body.NextCursor
	}
	return tools, nil
}

func (s *mcpSniffer) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	s.nextID++
	payload, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      s.nextID,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return nil, fmt.Errorf("encoding %s request: %w", method, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building %s request: %w", method, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("MCP-Protocol-Version", "2025-06-18")
	if s.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", s.sessionID)
	}
	for name, value := range s.headers {
		httpReq.Header.Set(name, value)
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", method, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if id := resp.Header.Get("Mcp-Session-Id"); id != "" {
		s.sessionID = id
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, fmt.Errorf("reading %s response: %w", method, err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("%s: HTTP %d — the server requires authorization. Pass --token or --header \"Authorization: Bearer <token>\"%s",
			method, resp.StatusCode, authHintFromChallenge(resp.Header.Get("WWW-Authenticate")))
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: HTTP %d: %s", method, resp.StatusCode, truncateForError(string(body)))
	}

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(extractSSEPayload(body), &envelope); err != nil {
		return nil, fmt.Errorf("decoding %s response: %w (body: %s)", method, err, truncateForError(string(body)))
	}
	if envelope.Error != nil {
		return nil, fmt.Errorf("%s: mcp error %d: %s", method, envelope.Error.Code, envelope.Error.Message)
	}
	return envelope.Result, nil
}

// authHintFromChallenge surfaces the scopes a WWW-Authenticate challenge names,
// so an operator sees which grant to obtain rather than a bare 401.
func authHintFromChallenge(challenge string) string {
	if challenge == "" {
		return ""
	}
	return fmt.Sprintf(" (server challenge: %s)", truncateForError(challenge))
}

// extractSSEPayload pulls JSON out of an SSE frame; already-JSON input passes
// through. Streamable-HTTP MCP servers may answer a single call as one event.
func extractSSEPayload(body []byte) []byte {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return body
	}
	var out strings.Builder
	for line := range strings.SplitSeq(trimmed, "\n") {
		if data, ok := strings.CutPrefix(strings.TrimRight(line, "\r"), "data:"); ok {
			out.WriteString(strings.TrimSpace(data))
		}
	}
	if out.Len() == 0 {
		return body
	}
	return []byte(out.String())
}

func truncateForError(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) <= 200 {
		return s
	}
	return string(r[:200]) + "…"
}
