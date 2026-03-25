package docspec

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mvanhorn/cli-printing-press/internal/spec"
)

var (
	endpointRe = regexp.MustCompile(`(GET|POST|PUT|PATCH|DELETE)\s+(/[a-zA-Z0-9/{}_.\-]+)`)
	baseURLRe  = regexp.MustCompile(`https://api\.[a-zA-Z0-9.\-]+`)
	bearerRe   = regexp.MustCompile(`(?i)(Bearer|Authorization:\s*Bearer)`)
	apiKeyRe   = regexp.MustCompile(`(?i)(API[_ ]key|api_key|X-API-Key)`)
	oauthRe    = regexp.MustCompile(`(?i)OAuth`)
	paramRowRe = regexp.MustCompile(`(?i)<t[dr][^>]*>\s*(\w+)\s*</t[dr]>\s*<t[dr][^>]*>\s*(string|integer|int|boolean|bool|number|float|array|object)\s*</t[dr]>`)
	jsonKeyRe  = regexp.MustCompile(`"(\w+)"\s*:`)
	preBlockRe = regexp.MustCompile(`(?s)<(?:pre|code)[^>]*>(.*?)</(?:pre|code)>`)
)

// GenerateFromDocs fetches an API documentation page and extracts a best-effort
// APISpec by scanning for endpoint patterns, auth hints, and parameters.
func GenerateFromDocs(docsURL, apiName string) (*spec.APISpec, error) {
	body, err := fetchHTML(docsURL)
	if err != nil {
		return nil, fmt.Errorf("fetching docs: %w", err)
	}

	endpoints := extractEndpoints(body)
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints found in %s", docsURL)
	}

	resources := groupByResource(endpoints)
	auth := detectAuth(body)
	baseURL := detectBaseURL(body)
	params := extractParams(body)

	// Attach extracted params as body params to POST/PUT/PATCH endpoints
	for resName, res := range resources {
		for epName, ep := range res.Endpoints {
			if ep.Method == "POST" || ep.Method == "PUT" || ep.Method == "PATCH" {
				ep.Body = params
				res.Endpoints[epName] = ep
			}
		}
		resources[resName] = res
	}

	apiSpec := &spec.APISpec{
		Name:        apiName,
		Description: fmt.Sprintf("CLI for %s (generated from docs)", apiName),
		Version:     "1.0.0",
		BaseURL:     baseURL,
		Auth:        auth,
		Config: spec.ConfigSpec{
			Format: "toml",
			Path:   fmt.Sprintf("~/.config/%s-cli/config.toml", apiName),
		},
		Resources: resources,
	}

	if err := apiSpec.Validate(); err != nil {
		return nil, fmt.Errorf("generated spec failed validation: %w", err)
	}

	return apiSpec, nil
}

func fetchHTML(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type rawEndpoint struct {
	Method string
	Path   string
}

func extractEndpoints(body string) []rawEndpoint {
	seen := map[string]bool{}
	var endpoints []rawEndpoint

	matches := endpointRe.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		key := m[1] + " " + m[2]
		if seen[key] {
			continue
		}
		seen[key] = true
		endpoints = append(endpoints, rawEndpoint{Method: m[1], Path: m[2]})
	}
	return endpoints
}

func groupByResource(endpoints []rawEndpoint) map[string]spec.Resource {
	groups := map[string][]rawEndpoint{}

	for _, ep := range endpoints {
		seg := firstSegment(ep.Path)
		groups[seg] = append(groups[seg], ep)
	}

	resources := map[string]spec.Resource{}
	for seg, eps := range groups {
		res := spec.Resource{
			Description: fmt.Sprintf("Operations on %s", seg),
			Endpoints:   map[string]spec.Endpoint{},
		}
		nameCount := map[string]int{}
		for _, ep := range eps {
			name := endpointName(ep.Method, ep.Path)
			nameCount[name]++
			if nameCount[name] > 1 {
				name = fmt.Sprintf("%s_%d", name, nameCount[name])
			}

			pathParams := extractPathParams(ep.Path)

			res.Endpoints[name] = spec.Endpoint{
				Method:      ep.Method,
				Path:        ep.Path,
				Description: fmt.Sprintf("%s %s", ep.Method, ep.Path),
				Params:      pathParams,
				Response: spec.ResponseDef{
					Type: "object",
				},
			}
		}
		resources[seg] = res
	}

	return resources
}

func firstSegment(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)
	seg := parts[0]
	// Clean version prefixes like v1, v2
	if len(seg) >= 2 && seg[0] == 'v' && seg[1] >= '0' && seg[1] <= '9' {
		if len(parts) > 1 {
			sub := strings.SplitN(parts[1], "/", 2)
			return sub[0]
		}
	}
	return seg
}

func endpointName(method, path string) string {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	last := parts[len(parts)-1]

	// If the last segment is a path param like {id}, use the one before it
	if strings.HasPrefix(last, "{") && len(parts) >= 2 {
		last = parts[len(parts)-2]
	}

	// Clean up the segment
	last = strings.ReplaceAll(last, "{", "")
	last = strings.ReplaceAll(last, "}", "")
	last = strings.ReplaceAll(last, "-", "_")

	prefix := strings.ToLower(method)
	switch prefix {
	case "get":
		prefix = "get"
	case "post":
		prefix = "create"
	case "put":
		prefix = "update"
	case "patch":
		prefix = "update"
	case "delete":
		prefix = "delete"
	}

	return prefix + "_" + last
}

func extractPathParams(path string) []spec.Param {
	re := regexp.MustCompile(`\{(\w+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)
	var params []spec.Param
	for _, m := range matches {
		params = append(params, spec.Param{
			Name:        m[1],
			Type:        "string",
			Required:    true,
			Positional:  true,
			Description: fmt.Sprintf("The %s identifier", m[1]),
		})
	}
	return params
}

func detectAuth(body string) spec.AuthConfig {
	upper := strings.ToUpper(body)
	_ = upper

	if bearerRe.MatchString(body) {
		return spec.AuthConfig{
			Type:   "bearer_token",
			Header: "Authorization",
			Format: "Bearer {token}",
			EnvVars: []string{
				"API_TOKEN",
			},
		}
	}
	if apiKeyRe.MatchString(body) {
		return spec.AuthConfig{
			Type:   "api_key",
			Header: "X-API-Key",
			Format: "{token}",
			EnvVars: []string{
				"API_KEY",
			},
		}
	}
	if oauthRe.MatchString(body) {
		return spec.AuthConfig{
			Type:   "oauth2",
			Header: "Authorization",
			Format: "Bearer {token}",
			EnvVars: []string{
				"API_TOKEN",
			},
		}
	}

	// Default
	return spec.AuthConfig{
		Type:   "bearer_token",
		Header: "Authorization",
		Format: "Bearer {token}",
		EnvVars: []string{
			"API_TOKEN",
		},
	}
}

func detectBaseURL(body string) string {
	matches := baseURLRe.FindAllString(body, -1)
	if len(matches) > 0 {
		// Return the first match, preferring longer ones (more specific)
		best := matches[0]
		for _, m := range matches[1:] {
			if len(m) > len(best) {
				best = m
			}
		}
		return best
	}
	return "https://api.example.com"
}

func extractParams(body string) []spec.Param {
	seen := map[string]bool{}
	var params []spec.Param

	// Strategy 1: HTML table rows with name + type
	rowMatches := paramRowRe.FindAllStringSubmatch(body, -1)
	for _, m := range rowMatches {
		name := m[1]
		typ := normalizeType(m[2])
		if !seen[name] {
			seen[name] = true
			params = append(params, spec.Param{
				Name:        name,
				Type:        typ,
				Description: fmt.Sprintf("The %s parameter", name),
			})
		}
	}

	// Strategy 2: JSON keys from <pre>/<code> blocks
	preMatches := preBlockRe.FindAllStringSubmatch(body, -1)
	for _, pm := range preMatches {
		block := pm[1]
		keyMatches := jsonKeyRe.FindAllStringSubmatch(block, -1)
		for _, km := range keyMatches {
			name := km[1]
			if !seen[name] {
				seen[name] = true
				params = append(params, spec.Param{
					Name:        name,
					Type:        "string",
					Description: fmt.Sprintf("The %s parameter", name),
				})
			}
		}
	}

	return params
}

func normalizeType(t string) string {
	switch strings.ToLower(t) {
	case "integer", "int":
		return "integer"
	case "boolean", "bool":
		return "boolean"
	case "number", "float":
		return "number"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "string"
	}
}
