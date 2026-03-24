package pipeline

import (
	"fmt"
	"strings"
)

// KnownSpecs maps common API names to their OpenAPI spec URLs.
var KnownSpecs = map[string]string{
	"petstore":     "https://petstore3.swagger.io/api/v3/openapi.json",
	"gmail":        "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/gmail/v1/openapi.yaml",
	"calendar":     "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/calendar/v3/openapi.yaml",
	"drive":        "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/drive/v3/openapi.yaml",
	"sheets":       "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/sheets/v4/openapi.yaml",
	"youtube":      "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/googleapis.com/youtube/v3/openapi.yaml",
	"stripe":       "https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json",
	"twilio":       "https://raw.githubusercontent.com/twilio/twilio-oai/main/spec/json/twilio_api_v2010.json",
	"sendgrid":     "https://raw.githubusercontent.com/sendgrid/sendgrid-oai/main/oai_stoplight.json",
	"github":       "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json",
	"discord":      "https://raw.githubusercontent.com/discord/discord-api-spec/main/specs/openapi.json",
	"digitalocean": "https://api-engineering.nyc3.cdn.digitaloceanspaces.com/spec-ci/DigitalOcean-public.v2.yaml",
	"slack":        "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/slack.com/1.7.0/openapi.yaml",
	"asana":        "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/asana.com/1.0/openapi.yaml",
	"hubspot":      "https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/hubspot.com/crm/v3/openapi.yaml",
	"openai":       "https://raw.githubusercontent.com/openai/openai-openapi/master/openapi.yaml",
	"anthropic":    "https://raw.githubusercontent.com/anthropics/anthropic-cookbook/main/misc/anthropic.openapi.yaml",
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
	if url, ok := KnownSpecs[normalized]; ok {
		return url, "known-specs registry", nil
	}

	// Try apis-guru with common version patterns
	for _, version := range []string{"v1", "v2", "v3", "1.0", "2.0"} {
		url := ApisGuruPattern(normalized+".com", version)
		return url, "apis-guru (unverified, needs fetch validation)", nil
	}

	return "", "", fmt.Errorf("could not find OpenAPI spec for %q - try providing a URL with --spec", apiName)
}
