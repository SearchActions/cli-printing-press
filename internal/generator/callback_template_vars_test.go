package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callbackTemplateVarSpec models the shape this feature exists for: a provider
// whose every request path is scoped by a tenant ID that the provider hands
// back on the authorization redirect rather than from any callable endpoint.
func callbackTemplateVarSpec(name string) *spec.APISpec {
	apiSpec := minimalSpec(name)
	apiSpec.BaseURL = "https://api.example.com/v3/company/{realm_id}"
	apiSpec.EndpointTemplateVars = []string{"realm_id"}
	apiSpec.Auth = spec.AuthConfig{
		Type:             "oauth2",
		Header:           "Authorization",
		Format:           "Bearer {token}",
		AuthorizationURL: "https://auth.example.com/connect/oauth2",
		TokenURL:         "https://auth.example.com/oauth2/v1/tokens/bearer",
		Scopes:           []string{"com.example.accounting"},
		CallbackTemplateVars: map[string]string{
			"realmId": "realm_id",
		},
	}
	return apiSpec
}

func TestCallbackTemplateVarsCaptureEmitted(t *testing.T) {
	t.Parallel()

	apiSpec := callbackTemplateVarSpec("callback-capture")
	require.NoError(t, apiSpec.Validate())

	outputDir := filepath.Join(t.TempDir(), "callback-capture-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	authSrc := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")
	configSrc := readGeneratedFile(t, outputDir, "internal", "config", "config.go")

	// The capture reads the declared callback param and stores it under the
	// placeholder name, not the wire name — those differ here on purpose.
	assert.Contains(t, authSrc, `r.URL.Query().Get("realmId")`)
	assert.Contains(t, authSrc, `capturedTemplateVars["realm_id"] = v`)
	// Persistence has to happen, or the value dies with the process.
	assert.Contains(t, authSrc, "cfg.SaveTemplateVars(captured)")
	assert.Contains(t, configSrc, "func (c *Config) SaveTemplateVars(vars map[string]string) error {")

	requireGeneratedCompiles(t, outputDir)
}

// A spec that declares no callback capture must emit none of the machinery —
// otherwise every printed CLI carries an unused mutex, an unused import, and a
// dead helper that go-lint rejects.
func TestCallbackTemplateVarsAbsentEmitsNothing(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("no-callback-capture")
	apiSpec.Auth = spec.AuthConfig{
		Type:             "oauth2",
		Header:           "Authorization",
		Format:           "Bearer {token}",
		AuthorizationURL: "https://auth.example.com/authorize",
		TokenURL:         "https://auth.example.com/token",
	}
	require.NoError(t, apiSpec.Validate())

	outputDir := filepath.Join(t.TempDir(), "no-callback-capture-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	authSrc := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")
	configSrc := readGeneratedFile(t, outputDir, "internal", "config", "config.go")

	assert.NotContains(t, authSrc, "capturedTemplateVars")
	assert.NotContains(t, authSrc, "sortedTemplateVarNames")
	assert.NotContains(t, authSrc, "SaveTemplateVars")
	assert.NotContains(t, configSrc, "func (c *Config) SaveTemplateVars")
	// The two imports the capture block needs must not be emitted unused.
	assert.NotContains(t, authSrc, "\n\t\"sync\"\n")

	requireGeneratedCompiles(t, outputDir)
}

// A placeholder that no endpoint template var declares would be written into
// TemplateVars and read by nothing, leaving a login that reports success while
// requests still carry an unresolved {placeholder}. Fail at spec-parse time.
func TestCallbackTemplateVarsRejectUnknownPlaceholder(t *testing.T) {
	t.Parallel()

	apiSpec := callbackTemplateVarSpec("callback-unknown")
	apiSpec.Auth.CallbackTemplateVars = map[string]string{"realmId": "tenant_id"}

	err := apiSpec.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `auth.callback_template_vars["realmId"] targets "tenant_id"`)
	assert.Contains(t, err.Error(), "not in endpoint_template_vars")
}

func TestCallbackTemplateVarsRejectEmptyNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		vars    map[string]string
		wantErr string
	}{
		{
			name:    "empty callback param",
			vars:    map[string]string{"": "realm_id"},
			wantErr: "empty callback parameter name",
		},
		{
			name:    "empty template var",
			vars:    map[string]string{"realmId": "  "},
			wantErr: "empty template-var name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			apiSpec := callbackTemplateVarSpec("callback-empty")
			apiSpec.Auth.CallbackTemplateVars = tt.vars

			err := apiSpec.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// SaveTemplateVars must ignore blanks rather than persist them: a blank would
// overwrite a working default with a value that resolves to nothing.
func TestSaveTemplateVarsSkipsBlankValues(t *testing.T) {
	t.Parallel()

	apiSpec := callbackTemplateVarSpec("callback-blank")
	require.NoError(t, apiSpec.Validate())

	outputDir := filepath.Join(t.TempDir(), "callback-blank-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	configSrc := readGeneratedFile(t, outputDir, "internal", "config", "config.go")
	saveFn := configSrc[strings.Index(configSrc, "func (c *Config) SaveTemplateVars"):]
	assert.Contains(t, saveFn[:strings.Index(saveFn, "\n}")], `if name == "" || value == "" {`)
}
