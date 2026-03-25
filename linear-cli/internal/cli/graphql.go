package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

// jsonRaw is an alias so interface signatures are cleaner.
type jsonRaw = json.RawMessage

func jsonUnmarshal(data json.RawMessage, v any) error {
	return json.Unmarshal(data, v)
}

// gql builds a GraphQL request body from a query string and optional variables.
func gql(query string, vars map[string]any) map[string]any {
	body := map[string]any{"query": query}
	if len(vars) > 0 {
		body["variables"] = vars
	}
	return body
}

// extractData navigates a GraphQL response to the inner data payload.
// path is dot-separated: "data.issues.nodes" extracts response.data.issues.nodes
func extractData(raw json.RawMessage, path string) (json.RawMessage, error) {
	parts := strings.Split(path, ".")
	current := raw
	for _, part := range parts {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(current, &obj); err != nil {
			return current, nil // not an object, return as-is
		}
		next, ok := obj[part]
		if !ok {
			return current, nil
		}
		current = next
	}
	return current, nil
}

// extractErrors checks for GraphQL errors in the response.
func extractErrors(raw json.RawMessage) error {
	var resp struct {
		Errors []struct {
			Message    string `json:"message"`
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil
	}
	if len(resp.Errors) == 0 {
		return nil
	}
	msgs := make([]string, len(resp.Errors))
	for i, e := range resp.Errors {
		msgs[i] = e.Message
		if e.Extensions.Code == "RATELIMITED" {
			return rateLimitErr(fmt.Errorf("rate limited: %s", e.Message))
		}
	}
	return apiErr(fmt.Errorf("GraphQL errors: %s", strings.Join(msgs, "; ")))
}

// buildFilter constructs a Linear GraphQL filter object from flag values.
func buildFilter(state, assignee, team, project, label, priority string) string {
	var filters []string

	if state != "" {
		filters = append(filters, fmt.Sprintf(`state: { name: { eqIgnoreCase: %q } }`, state))
	}
	if assignee != "" {
		if assignee == "me" {
			filters = append(filters, `assignee: { isMe: { eq: true } }`)
		} else {
			filters = append(filters, fmt.Sprintf(`assignee: { name: { eqIgnoreCase: %q } }`, assignee))
		}
	}
	if team != "" {
		filters = append(filters, fmt.Sprintf(`team: { key: { eqIgnoreCase: %q } }`, team))
	}
	if project != "" {
		filters = append(filters, fmt.Sprintf(`project: { name: { containsIgnoreCase: %q } }`, project))
	}
	if label != "" {
		filters = append(filters, fmt.Sprintf(`labels: { name: { eqIgnoreCase: %q } }`, label))
	}
	if priority != "" {
		p := priorityNumber(priority)
		if p >= 0 {
			filters = append(filters, fmt.Sprintf(`priority: { eq: %d }`, p))
		}
	}

	if len(filters) == 0 {
		return ""
	}
	return "filter: { " + strings.Join(filters, ", ") + " }"
}

func priorityNumber(s string) int {
	switch strings.ToLower(s) {
	case "none", "0":
		return 0
	case "urgent", "1":
		return 1
	case "high", "2":
		return 2
	case "medium", "3":
		return 3
	case "low", "4":
		return 4
	default:
		return -1
	}
}
