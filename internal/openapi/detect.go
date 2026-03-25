package openapi

import "bytes"

func IsOpenAPI(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check for JSON-style keys (case-sensitive, covers 99% of specs)
	if bytes.Contains(data, []byte(`"openapi"`)) ||
		bytes.Contains(data, []byte(`"swagger"`)) {
		return true
	}

	// Check for YAML-style keys
	if bytes.Contains(data, []byte("openapi:")) ||
		bytes.Contains(data, []byte("swagger:")) {
		return true
	}

	return false
}
