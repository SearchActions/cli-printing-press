package pipeline

import (
	"fmt"
	"strings"
)

// KnownSpec holds metadata about a known API spec.
type KnownSpec struct {
	URL         string
	SandboxSafe bool
}

// KnownSpecs maps common API names to their OpenAPI spec URLs.
var KnownSpecs = map[string]KnownSpec{
	"petstore": {
		URL:         "https://petstore3.swagger.io/api/v3/openapi.json",
		SandboxSafe: true,
	},
	"gmail": {
		URL:         "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/gmail/v1/openapi.yaml",
		SandboxSafe: false,
	},
	"calendar": {
		URL:         "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/calendar/v3/openapi.yaml",
		SandboxSafe: false,
	},
	"drive": {
		URL:         "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/drive/v3/openapi.yaml",
		SandboxSafe: false,
	},
	"sheets": {
		URL:         "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/sheets/v4/openapi.yaml",
		SandboxSafe: false,
	},
	"youtube": {
		URL:         "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/youtube/v3/openapi.yaml",
		SandboxSafe: false,
	},
	"stripe": {
		URL:         "https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json",
		SandboxSafe: false,
	},
	"twilio": {
		URL:         "https://raw.githubusercontent.com/twilio/twilio-oai/main/spec/json/twilio_api_v2010.json",
		SandboxSafe: false,
	},
	"sendgrid": {
		URL:         "https://raw.githubusercontent.com/sendgrid/sendgrid-oai/main/oai_stoplight.json",
		SandboxSafe: false,
	},
	"github": {
		URL:         "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json",
		SandboxSafe: false,
	},
	"discord": {
		URL:         "https://raw.githubusercontent.com/discord/discord-api-spec/main/specs/openapi.json",
		SandboxSafe: false,
	},
	"digitalocean": {
		URL:         "https://api-engineering.nyc3.cdn.digitaloceanspaces.com/spec-ci/DigitalOcean-public.v2.yaml",
		SandboxSafe: false,
	},
	"slack": {
		URL:         "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/slack.com/1.7.0/openapi.yaml",
		SandboxSafe: false,
	},
	"asana": {
		URL:         "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/asana.com/1.0/openapi.yaml",
		SandboxSafe: false,
	},
	"hubspot": {
		URL:         "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/hubspot.com/crm/v3/openapi.yaml",
		SandboxSafe: false,
	},
	"openai": {
		URL:         "https://raw.githubusercontent.com/openai/openai-openapi/master/openapi.yaml",
		SandboxSafe: false,
	},
	"anthropic": {
		URL:         "https://raw.githubusercontent.com/anthropics/anthropic-cookbook/main/misc/anthropic.openapi.yaml",
		SandboxSafe: false,
	},
}

// ApisGuruPattern builds an apis-guru URL for a provider and version.
func ApisGuruPattern(provider, version string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/%s/%s/openapi.yaml", provider, version)
}

// DiscoverSpec finds the OpenAPI spec URL for a given API name.
// Returns the URL and a source description.
func DiscoverSpec(apiName string) (string, string, error) {
	normalized := strings.ToLower(strings.TrimSpace(apiName))

	// Check known specs first
	if spec, ok := KnownSpecs[normalized]; ok {
		return spec.URL, "known-specs registry", nil
	}

	// Try apis-guru with common version patterns
	for _, version := range []string{"v1", "v2", "v3", "1.0", "2.0"} {
		url := ApisGuruPattern(normalized+".com", version)
		return url, "apis-guru (unverified, needs fetch validation)", nil
	}

	return "", "", fmt.Errorf("could not find OpenAPI spec for %q - try providing a URL with --spec", apiName)
}

// IsSandboxSafe returns true if the API is known to have a safe test/sandbox environment.
func IsSandboxSafe(apiName string) bool {
	normalized := strings.ToLower(strings.TrimSpace(apiName))
	if spec, ok := KnownSpecs[normalized]; ok {
		return spec.SandboxSafe
	}
	return false
}
