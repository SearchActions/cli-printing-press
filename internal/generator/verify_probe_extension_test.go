// Copyright 2026 mvanhorn. Licensed under Apache-2.0. See LICENSE.

package generator

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mvanhorn/cli-printing-press/v4/internal/naming"
	"github.com/mvanhorn/cli-printing-press/v4/internal/openapi"
	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
)

// TestGeneratedDoctorEscapesVerifyProbeLiterals covers the case where
// VerifyPath/HealthCheckPath can come from a third-party OpenAPI document
// (via x-auth-verify-path), so a quote or backslash in the value must not
// break the emitted doctor.go. Proves the printf "%q" switch in
// doctor.go.tmpl by compiling the generated module, not by string matching
// alone.
func TestGeneratedDoctorEscapesVerifyProbeLiterals(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("verify-escape")
	apiSpec.Auth.VerifyPath = `/me"; import "os/exec` + `\`
	apiSpec.HealthCheckPath = `/health"` + `\` + `x`

	outputDir := filepath.Join(t.TempDir(), "verify-escape-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	doctorGo := readGeneratedFile(t, outputDir, "internal", "cli", "doctor.go")
	assert.Contains(t, doctorGo, `verifyPath := "/me\"; import \"os/exec\\"`,
		"VerifyPath literal must be %q-escaped")
	assert.Contains(t, doctorGo, `healthPath := "/health\"\\x"`,
		"HealthCheckPath literal must be %q-escaped")

	requireGeneratedCompiles(t, outputDir)
}

// TestGeneratedAuth0SPAEscapesHealthCheckPathLiteral covers the third raw
// sink, auth0SPACaptureURL() in auth_browser.go.tmpl, which is gated on
// Auth.Subtype == auth0_spa_in_memory and no golden case reaches it, so
// this is the only compile coverage for that sink.
func TestGeneratedAuth0SPAEscapesHealthCheckPathLiteral(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("verify-escape-auth0")
	apiSpec.Auth = spec.AuthConfig{
		Type:    "bearer_token",
		Subtype: spec.AuthSubtypeAuth0SPAInMemory,
		Header:  "Authorization",
		EnvVars: []string{"VERIFY_ESCAPE_AUTH0_TOKEN"},
	}
	apiSpec.HealthCheckPath = `/me"` + `\` + `q`

	outputDir := filepath.Join(t.TempDir(), "verify-escape-auth0-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	authGo := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")
	assert.Contains(t, authGo, `strings.TrimSpace("/me\"\\q")`,
		"auth0SPACaptureURL's HealthCheckPath literal must be %q-escaped")

	requireGeneratedCompiles(t, outputDir)
}

// TestGeneratedDoctorHonorsOpenAPIVerifyPathOverHeuristic covers an OpenAPI
// spec carrying a scheme-level x-auth-verify-path going through
// openapi.Parse then Generate(): the authored value — not the me-shaped
// heuristic's pick — reaches the emitted doctor.go. The fixture's
// only me-shaped GET is /integrations/me, deliberately earlier/likelier than
// the authored /users/self, so the heuristic and the authored value disagree
// and the test can fail if the parser wiring or the generator guard regress.
func TestGeneratedDoctorHonorsOpenAPIVerifyPathOverHeuristic(t *testing.T) {
	t.Parallel()

	yamlSpec := []byte(`openapi: "3.0.3"
info:
  title: Partner Tier API
  version: "1.0.0"
servers:
  - url: https://api.example.com
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: Authorization
      x-auth-verify-path: /users/self
security:
  - ApiKeyAuth: []
paths:
  /integrations/me:
    get:
      responses:
        "200":
          description: OK
  /items:
    get:
      responses:
        "200":
          description: OK
`)

	parsed, err := openapi.Parse(yamlSpec)
	require.NoError(t, err)
	require.Equal(t, "/users/self", parsed.Auth.VerifyPath,
		"parser should have already mapped the extension onto Auth.VerifyPath")

	outputDir := filepath.Join(t.TempDir(), naming.CLI(parsed.Name))
	require.NoError(t, New(parsed, outputDir).Generate())

	doctorGo := readGeneratedFile(t, outputDir, "internal", "cli", "doctor.go")
	assert.Contains(t, doctorGo, `verifyPath := "/users/self"`,
		"the authored x-auth-verify-path must win over the /integrations/me heuristic pick")
	assert.NotContains(t, doctorGo, `verifyPath := "/integrations/me"`,
		"the heuristic must not override an authored verify path")
}

// TestGeneratedDoctorAuthoredVerifyQuerySurvivesGenerate covers Generate()'s
// empty-check at generator.go, which must derive VerifyPath only when
// VerifyQuery is also empty. A spec with an authored VerifyQuery and a
// me-shaped GET must keep VerifyPath empty after Generate(), and the
// emitted doctor.go must carry the GraphQL probe, not a derived REST one.
func TestGeneratedDoctorAuthoredVerifyQuerySurvivesGenerate(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("verify-query-guard")
	apiSpec.Auth.VerifyQuery = "{ viewer { id } }"
	apiSpec.Resources = map[string]spec.Resource{
		"account": {
			Description: "Account",
			Endpoints: map[string]spec.Endpoint{
				"me": {Method: "GET", Path: "/me", Description: "Get the current account"},
			},
		},
	}

	gen := New(apiSpec, filepath.Join(t.TempDir(), "verify-query-guard-pp-cli"))
	require.NoError(t, gen.Generate())

	assert.Empty(t, gen.Spec.Auth.VerifyPath,
		"an authored VerifyQuery must not be clobbered by a derived VerifyPath")

	doctorGo := readGeneratedFile(t, gen.OutputDir, "internal", "cli", "doctor.go")
	assert.Contains(t, doctorGo, `verifyBody := map[string]string{"query": "{ viewer { id } }"}`,
		"the GraphQL probe must be emitted when VerifyPath was never derived")
	assert.NotContains(t, doctorGo, "verifyPath :=",
		"no REST verify path should be emitted when VerifyQuery is authored and VerifyPath was never set")
}

// TestGeneratedDoctorVerifyPathQueryStringReachesLiveRequest covers the
// embedded "?query" on an authored x-auth-verify-path and the query-param
// auth credential, both of which must reach the live probe request. This
// is a runtime test against a stub server, not a template-content match, so
// it fails if either the query string or the auth query param is dropped
// anywhere between config, client, and doctor.
func TestGeneratedDoctorVerifyPathQueryStringReachesLiveRequest(t *testing.T) {
	t.Parallel()

	var (
		mu       sync.Mutex
		meHits   int
		gotQuery string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			return
		case "/me":
			mu.Lock()
			meHits++
			gotQuery = r.URL.RawQuery
			mu.Unlock()
			if r.URL.Query().Get("fields") == "id" && r.URL.Query().Get("api_key") == "secret-key" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	apiSpec := minimalSpec("verify-rt")
	apiSpec.BaseURL = server.URL
	apiSpec.HealthCheckPath = "/health"
	apiSpec.Auth = spec.AuthConfig{
		Type:    "api_key",
		Header:  "api_key",
		In:      "query",
		EnvVars: []string{"VERIFY_RT_API_KEY"},
	}
	apiSpec.Auth.VerifyPath = "/me?fields=id"

	_, binaryPath := buildGeneratedBinary(t, apiSpec)
	env := append(doctorEnv(t.TempDir(), naming.EnvPrefix(apiSpec.Name)),
		"VERIFY_RT_BASE_URL="+server.URL,
		"VERIFY_RT_API_KEY=secret-key",
	)
	payload, err := runDoctorJSON(t, binaryPath, env)
	require.NoError(t, err)

	assert.Equal(t, "valid", payload["credentials"])

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, meHits, "the /me probe should be hit exactly once")
	assert.Contains(t, gotQuery, "fields=id", "the embedded ?query on x-auth-verify-path must reach the request")
	assert.Contains(t, gotQuery, "api_key=secret-key", "the query-param auth credential must also reach the request")
}
