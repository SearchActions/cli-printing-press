package mcpspec

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"unicode"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
)

// classifiedTool is a tool after its name has been split into a resource and
// an endpoint verb, and its read/write nature resolved.
type classifiedTool struct {
	Tool     mcpTool
	Resource string
	Endpoint string
	Method   string
	ReadOnly bool
}

// readVerbs and writeVerbs classify a tool by the leading token of its name.
// Only used when the server supplies no readOnlyHint annotation.
var readVerbs = map[string]string{
	"list":     "list",
	"get":      "get",
	"read":     "get",
	"fetch":    "get",
	"search":   "search",
	"find":     "search",
	"query":    "search",
	"describe": "get",
	"show":     "get",
	"check":    "check",
	"preview":  "preview",
	"validate": "validate",
	"resolve":  "resolve",
	"detect":   "detect",
	"report":   "report",
	"count":    "count",
	"inspect":  "get",
	"trace":    "trace",
	"analyze":  "analyze",
	"audit":    "audit",
}

var writeVerbs = map[string]string{
	"create":     "create",
	"add":        "create",
	"write":      "write",
	"update":     "update",
	"set":        "update",
	"edit":       "update",
	"patch":      "update",
	"modify":     "update",
	"rename":     "update",
	"move":       "move",
	"delete":     "delete",
	"remove":     "delete",
	"destroy":    "delete",
	"purge":      "delete",
	"revoke":     "revoke",
	"unregister": "unregister",
	"register":   "register",
	"send":       "send",
	"share":      "share",
	"submit":     "submit",
	"execute":    "execute",
	"run":        "run",
	"start":      "start",
	"stop":       "stop",
	"restart":    "restart",
	"cancel":     "cancel",
	"deploy":     "deploy",
	"finalize":   "finalize",
	"complete":   "complete",
	"apply":      "apply",
	"upload":     "upload",
	"import":     "import",
	"process":    "process",
	"assign":     "assign",
	"duplicate":  "duplicate",
	"copy":       "copy",
}

// classify splits a tool name into resource + endpoint and resolves whether
// the call is a read.
//
// A server-supplied readOnlyHint always wins over the name heuristic: the
// server knows its own semantics, and a wrong readOnlyHint on a mutating tool
// is the one annotation error AGENTS.md calls a real bug.
func classify(t mcpTool, stripPrefix string) classifiedTool {
	bare := strings.TrimPrefix(t.Name, stripPrefix)
	words := splitWords(bare)
	if len(words) == 0 {
		words = splitWords(t.Name)
	}

	verb, object, readByName := splitVerbObject(words)

	resource := object
	if resource == "" {
		resource = defaultResourceName(stripPrefix, t.Name)
	}
	resource = toKebab(pluralize(resource))

	readOnly := readByName
	if t.Annotations != nil && t.Annotations.ReadOnlyHint != nil {
		readOnly = *t.Annotations.ReadOnlyHint
	}

	method := methodFor(verb, readOnly, t.Annotations)

	return classifiedTool{
		Tool:     t,
		Resource: resource,
		Endpoint: toKebab(verb),
		Method:   method,
		ReadOnly: readOnly,
	}
}

// splitVerbObject finds the verb in a tool name and returns it with the object
// it acts on, plus whether the verb is a read.
//
// The verb is located at *any* position, not just the head: real catalogs
// qualify tool names with a service or module first (slack_send_message,
// qbo_sales_create_invoice), and treating the leading qualifier as the verb
// yields junk resources like "slacks". Words before the verb are namespace
// context and are discarded; words after it are the object, which becomes the
// resource. Returns an empty object when no known verb appears, leaving the
// caller to fall back to the server namespace.
func splitVerbObject(words []string) (verb, object string, readOnly bool) {
	for i, w := range words {
		head := strings.ToLower(w)
		if v, ok := readVerbs[head]; ok && i+1 < len(words) {
			return v, strings.Join(words[i+1:], "-"), true
		}
		if v, ok := writeVerbs[head]; ok && i+1 < len(words) {
			return v, strings.Join(words[i+1:], "-"), false
		}
	}
	if len(words) == 0 {
		return "call", "tool", false
	}
	// A trailing verb with nothing after it (…_list) still names the action.
	last := strings.ToLower(words[len(words)-1])
	if v, ok := readVerbs[last]; ok && len(words) > 1 {
		return v, strings.Join(words[:len(words)-1], "-"), true
	}
	if v, ok := writeVerbs[last]; ok && len(words) > 1 {
		return v, strings.Join(words[:len(words)-1], "-"), false
	}
	return strings.Join(words, "-"), "", false
}

// methodFor maps a classified tool onto the semantic HTTP verb that drives
// the generated command's MCP safety annotations. The wire call is always a
// JSON-RPC POST; this is deliberately the semantic verb, not the transport's.
func methodFor(verb string, readOnly bool, ann *toolAnnotations) string {
	if readOnly {
		return "GET"
	}
	if ann != nil && ann.DestructiveHint != nil && *ann.DestructiveHint {
		return "DELETE"
	}
	switch verb {
	case "delete", "remove", "destroy", "purge", "revoke", "unregister":
		return "DELETE"
	case "update", "set", "edit", "patch", "modify", "rename":
		return "PATCH"
	default:
		return "POST"
	}
}

// buildEndpoint renders one classified tool as an endpoint. Every tool input
// property becomes a body param: the JSON-RPC envelope carries arguments in
// the body regardless of the semantic method, so there are no path or query
// params to derive.
func buildEndpoint(ct classifiedTool, endpointPath string) spec.Endpoint {
	// Each tool gets a distinct synthetic path (<endpoint>/<tool>) rather than
	// every endpoint sharing the server's single route. The generated
	// transport maps it back to a tools/call POST at the real endpoint; the
	// distinct path is what lets the rest of the generator treat these as
	// ordinary, separately-addressable endpoints.
	ep := spec.Endpoint{
		Method:      ct.Method,
		Path:        endpointPath + "/" + ct.Tool.Name,
		Description: describeTool(ct),
		Response:    spec.ResponseDef{Type: "object"},
		Meta: map[string]string{
			mcpToolAnnotation: ct.Tool.Name,
		},
	}
	if ct.ReadOnly {
		ep.Meta[mcpReadOnlyAnnotation] = "true"
	}

	params := paramsFromSchema(ct.Tool.schema())
	// A read that rides POST on the wire still presents its inputs as flags;
	// keeping them in Params (not Body) is what lets the emitted read command
	// route through doRead() rather than the mutation gate.
	if ct.ReadOnly {
		ep.Params = params
	} else {
		ep.Body = params
	}
	return ep
}

const (
	// mcpToolAnnotation records the upstream MCP tool name so the emitted
	// transport can build the tools/call envelope, and so a regen can map an
	// endpoint back to its source tool.
	mcpToolAnnotation = "pp:mcp-tool"
	// mcpReadOnlyAnnotation marks a read that rides JSON-RPC POST so the
	// transport routes it through doRead() instead of the mutation gate.
	mcpReadOnlyAnnotation = "mcp:read-only"
)

// paramsFromSchema flattens a tool's JSON Schema input into params, sorted by
// name so generation is deterministic. Required properties sort first so the
// generated command's positional/required ordering is stable and readable.
func paramsFromSchema(s *jsonSchema) []spec.Param {
	if s == nil || len(s.Properties) == 0 {
		return nil
	}
	required := map[string]bool{}
	for _, r := range s.Required {
		required[r] = true
	}

	names := make([]string, 0, len(s.Properties))
	for n := range s.Properties {
		names = append(names, n)
	}
	sort.Strings(names)
	sort.SliceStable(names, func(i, j int) bool {
		return required[names[i]] && !required[names[j]]
	})

	params := make([]spec.Param, 0, len(names))
	for _, n := range names {
		params = append(params, paramFromSchema(n, s.Properties[n], required[n]))
	}
	return params
}

func paramFromSchema(name string, s *jsonSchema, required bool) spec.Param {
	p := spec.Param{
		Name:     name,
		Required: required,
		Type:     "string",
	}
	if s == nil {
		return p
	}
	p.Type = mapJSONType(s.Type)
	p.Description = strings.TrimSpace(s.Description)
	p.Format = s.Format
	p.Default = s.Default
	p.Maximum = s.Maximum
	p.Enum = enumStrings(s.Enum)

	switch p.Type {
	case "array":
		p.ItemType = "string"
		if s.Items != nil {
			p.ItemType = mapJSONType(s.Items.Type)
		}
	case "object":
		// Nested object properties become sub-fields so the generator can
		// emit them as dotted flags rather than collapsing to opaque JSON.
		p.Fields = paramsFromSchema(s)
	}
	return p
}

// mapJSONType maps a JSON Schema type to the spec param type vocabulary.
// An absent or union type degrades to string, which every value can carry.
func mapJSONType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "integer":
		return "integer"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "string"
	}
}

// enumStrings renders enum values as strings. Non-scalar enum members are
// dropped rather than rendered as Go-syntax noise in help text.
func enumStrings(vals []any) []string {
	if len(vals) == 0 {
		return nil
	}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		switch tv := v.(type) {
		case string:
			out = append(out, tv)
		case bool:
			out = append(out, fmt.Sprintf("%t", tv))
		case float64:
			// Render integral floats without a trailing .0 so an enum of
			// small ints reads as 1|2|3 in help output.
			if tv == float64(int64(tv)) {
				out = append(out, fmt.Sprintf("%d", int64(tv)))
			} else {
				out = append(out, fmt.Sprintf("%g", tv))
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func describeTool(ct classifiedTool) string {
	if d := strings.TrimSpace(ct.Tool.Description); d != "" {
		return firstSentence(d)
	}
	if t := strings.TrimSpace(ct.Tool.Title); t != "" {
		return t
	}
	return fmt.Sprintf("Call the %s MCP tool", ct.Tool.Name)
}

// firstSentence trims a long MCP tool description down to its lead sentence.
// MCP descriptions are written for model consumption and frequently run to
// paragraphs; a Cobra Short line needs one sentence.
func firstSentence(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if idx := strings.Index(s, ". "); idx > 0 {
		return s[:idx+1]
	}
	return truncateRunes(s, 200)
}

// truncateRunes cuts on a rune boundary so a multi-byte description is never
// split mid-character.
func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return strings.TrimSpace(string(r[:max])) + "…"
}

// commonNamePrefix finds a shared `namespace_` prefix across every tool name.
// Servers commonly namespace all tools (fireflies_get_transcript,
// slack_send_message); that prefix is the server, not a resource, so it is
// stripped before resource derivation. Returns "" unless every tool shares it
// and stripping leaves a non-empty remainder.
func commonNamePrefix(tools []mcpTool) string {
	if len(tools) < 2 {
		return ""
	}
	first := tools[0].Name
	cut := strings.Index(first, "_")
	if cut <= 0 {
		return ""
	}
	prefix := first[:cut+1]
	for _, t := range tools {
		if !strings.HasPrefix(t.Name, prefix) || len(t.Name) == len(prefix) {
			return ""
		}
	}
	return prefix
}

// defaultResourceName supplies a resource for verb-only tool names (a bare
// `search` or `authenticate`), preferring the stripped server namespace.
func defaultResourceName(stripPrefix, toolName string) string {
	if p := strings.Trim(stripPrefix, "_"); p != "" {
		return p
	}
	if words := splitWords(toolName); len(words) > 0 {
		return words[0]
	}
	return "tools"
}

// splitServerURL splits an MCP endpoint URL into an origin base URL and the
// endpoint path the JSON-RPC POST targets.
func splitServerURL(raw string) (baseURL, endpointPath string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", nil
	}
	u, parseErr := url.Parse(raw)
	if parseErr != nil {
		return "", "", fmt.Errorf("parsing MCP server URL %q: %w", raw, parseErr)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", "", fmt.Errorf("MCP server URL %q must be absolute (https://host/path)", raw)
	}
	// The generated client replays only scheme+host+path, so a URL carrying
	// userinfo or a query string would silently lose them and stop working.
	// Reject at parse time rather than dropping them. The error deliberately
	// does not echo the URL, which would print the credential.
	if u.User != nil {
		return "", "", fmt.Errorf("MCP server URL must not embed credentials; pass them with --token or --header instead")
	}
	if u.RawQuery != "" {
		return "", "", fmt.Errorf("MCP server URL %q must not carry a query string; the generated client cannot replay it", u.Scheme+"://"+u.Host+u.Path)
	}
	path := strings.TrimRight(u.Path, "/")
	return u.Scheme + "://" + u.Host, path, nil
}

func deriveName(doc catalog, serverURL, source string) string {
	if doc.ServerInfo != nil {
		if n := strings.TrimSpace(doc.ServerInfo.Name); n != "" {
			return toKebab(n)
		}
	}
	// Prefer a meaningful path segment (…/v1/design/mcp -> design) over the
	// host, which is frequently a shared API gateway.
	if serverURL != "" {
		if u, err := url.Parse(serverURL); err == nil {
			segs := []string{}
			for s := range strings.SplitSeq(u.Path, "/") {
				s = strings.TrimSpace(s)
				if s == "" || s == "mcp" || s == "sse" || isVersionSegment(s) {
					continue
				}
				segs = append(segs, s)
			}
			if len(segs) > 0 {
				return toKebab(segs[len(segs)-1])
			}
			if u.Host != "" {
				return toKebab(strings.TrimPrefix(hostLabel(u.Host), "api."))
			}
		}
	}
	return toKebab(strings.TrimSuffix(baseName(source), ".json"))
}

func isVersionSegment(s string) bool {
	if len(s) < 2 || (s[0] != 'v' && s[0] != 'V') {
		return false
	}
	for _, r := range s[1:] {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// hostLabel reduces a host to its registrable-ish label for slug derivation.
func hostLabel(host string) string {
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return host
}

func baseName(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func describeServer(doc catalog, name string) string {
	if doc.ServerInfo != nil {
		if t := strings.TrimSpace(doc.ServerInfo.Title); t != "" {
			return t
		}
	}
	return fmt.Sprintf("Generated from the %s MCP server tool catalog", name)
}

// defaultAuth assumes a bearer token. Most remote MCP servers are OAuth
// bearer-protected; a spec author retargets this when the server takes an API
// key or the CLI needs the full authorization-code flow.
func defaultAuth(name string) spec.AuthConfig {
	return spec.AuthConfig{
		Type:    "bearer_token",
		Header:  "Authorization",
		EnvVars: []string{strings.ToUpper(strings.ReplaceAll(toKebab(name), "-", "_")) + "_TOKEN"},
	}
}

// uniqueEndpointName resolves two tools in one resource that classify to the
// same verb (list_files and list_file_versions both -> "list" after the
// object becomes the resource) by suffixing an ordinal.
func uniqueEndpointName(base string, used map[string]struct{}) string {
	if _, clash := used[base]; !clash {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if _, clash := used[candidate]; !clash {
			return candidate
		}
	}
}

// splitWords splits snake_case, kebab-case, dotted and camelCase names into
// lowercase words.
func splitWords(s string) []string {
	var words []string
	var cur strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		switch {
		case r == '_' || r == '-' || r == '.' || r == ' ' || r == '/':
			if cur.Len() > 0 {
				words = append(words, strings.ToLower(cur.String()))
				cur.Reset()
			}
		case unicode.IsUpper(r):
			// Break before an uppercase run start, and at the tail of an
			// acronym run (HTTPServer -> http, server).
			prevLower := i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1]))
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if cur.Len() > 0 && (prevLower || nextLower) {
				words = append(words, strings.ToLower(cur.String()))
				cur.Reset()
			}
			cur.WriteRune(unicode.ToLower(r))
		default:
			cur.WriteRune(unicode.ToLower(r))
		}
	}
	if cur.Len() > 0 {
		words = append(words, strings.ToLower(cur.String()))
	}
	return words
}

func toKebab(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	return strings.Join(words, "-")
}

// pluralize renders a resource noun plural so resource names read as
// collections, matching the convention the other parsers follow.
func pluralize(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	switch {
	case strings.HasSuffix(lower, "s"), strings.HasSuffix(lower, "x"),
		strings.HasSuffix(lower, "ch"), strings.HasSuffix(lower, "sh"):
		return s
	case strings.HasSuffix(lower, "y") && len(s) > 1 && !isVowel(rune(lower[len(lower)-2])):
		return s[:len(s)-1] + "ies"
	default:
		return s + "s"
	}
}

func isVowel(r rune) bool {
	return strings.ContainsRune("aeiou", r)
}
