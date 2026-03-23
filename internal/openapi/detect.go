package openapi

import "strings"

func IsOpenAPI(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if len(data) > 500 {
		data = data[:500]
	}

	content := strings.ToLower(string(data))
	return strings.Contains(content, "openapi:") || strings.Contains(content, "\"openapi\"")
}
