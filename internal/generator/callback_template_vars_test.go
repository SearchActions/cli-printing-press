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

// The browser-wait budget has to cover sign-in, MFA, company selection and a
// consent screen. The old two-minute cap lost that race routinely, and the
// failure is expensive: the callback server is gone by the time the user
// approves, so the authorization lands on a dead port.
func TestAuthLoginBrowserWaitIsGenerousAndOverridable(t *testing.T) {
	t.Parallel()

	apiSpec := callbackTemplateVarSpec("auth-timeout")
	require.NoError(t, apiSpec.Validate())

	outputDir := filepath.Join(t.TempDir(), "auth-timeout-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	authSrc := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")

	assert.Contains(t, authSrc, "authTimeout := 15 * time.Minute")
	assert.NotContains(t, authSrc, "time.After(2 * time.Minute)")
	// The override has to be rejected when unusable rather than silently
	// falling back, or a typo produces a login that fails for the wrong reason.
	assert.Contains(t, authSrc, `os.Getenv("AUTH_TIMEOUT_AUTH_TIMEOUT")`)
	assert.Contains(t, authSrc, "invalid AUTH_TIMEOUT_AUTH_TIMEOUT")
	assert.Contains(t, authSrc, "AUTH_TIMEOUT_AUTH_TIMEOUT must be positive")

	requireGeneratedCompiles(t, outputDir)
}

// Whitespace in a declared name used to pass validation (which trims) while
// the untrimmed key reached the emitted code, so the captured value landed
// under a key URL expansion never looks up — a login that reports success and
// requests that still carry {realm_id}.
func TestCallbackTemplateVarsNormalizeWhitespace(t *testing.T) {
	t.Parallel()

	apiSpec := callbackTemplateVarSpec("callback-whitespace")
	apiSpec.Auth.CallbackTemplateVars = map[string]string{"  realmId  ": "  realm_id  "}
	require.NoError(t, apiSpec.Validate())

	// Validation normalizes in place so the check and the emission read the
	// same string.
	assert.Equal(t, map[string]string{"realmId": "realm_id"}, apiSpec.Auth.CallbackTemplateVars)

	outputDir := filepath.Join(t.TempDir(), "callback-whitespace-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	authSrc := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")
	assert.Contains(t, authSrc, `r.URL.Query().Get("realmId")`)
	assert.Contains(t, authSrc, `capturedTemplateVars["realm_id"] = v`)
	assert.NotContains(t, authSrc, `Get("  realmId  ")`)
	assert.NotContains(t, authSrc, `capturedTemplateVars["  realm_id  "]`)

	requireGeneratedCompiles(t, outputDir)
}

// Two spellings of one callback param that disagree on the target placeholder
// resolve by map iteration order — nondeterministic, so reject rather than
// silently pick one.
func TestCallbackTemplateVarsRejectConflictingEntries(t *testing.T) {
	t.Parallel()

	apiSpec := callbackTemplateVarSpec("callback-conflict")
	apiSpec.EndpointTemplateVars = []string{"realm_id", "tenant_id"}
	apiSpec.Auth.CallbackTemplateVars = map[string]string{
		"realmId":   "realm_id",
		" realmId ": "tenant_id",
	}

	err := apiSpec.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicting entries for callback parameter")
}

// Only authorization_code redirects the user back with query params, so a
// mapping under any other grant declares a capture from a request that never
// happens. Fail at parse time rather than emit dead code.
func TestCallbackTemplateVarsRequireAuthorizationCodeGrant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(*spec.APISpec)
		wantErr string
	}{
		{
			name: "client_credentials has no browser callback",
			mutate: func(s *spec.APISpec) {
				s.Auth.OAuth2Grant = spec.OAuth2GrantClientCredentials
			},
			wantErr: "requires auth.oauth2_grant",
		},
		{
			name: "device_code has no browser callback",
			mutate: func(s *spec.APISpec) {
				s.Auth.OAuth2Grant = spec.OAuth2GrantDeviceCode
				s.Auth.DeviceAuthorizationURL = "https://auth.example.com/device"
			},
			wantErr: "requires auth.oauth2_grant",
		},
		{
			name: "non-OAuth auth has no callback at all",
			mutate: func(s *spec.APISpec) {
				s.Auth.Type = "api_key"
			},
			wantErr: "requires auth.type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			apiSpec := callbackTemplateVarSpec("callback-grant")
			tt.mutate(apiSpec)

			err := apiSpec.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// The timeout override has to be resolved before the flow binds a port, opens
// the user's browser and starts a server: discovering a malformed value after
// those side effects costs a stray browser tab and a dangling listener to
// report a plain configuration error.
func TestAuthLoginValidatesTimeoutBeforeSideEffects(t *testing.T) {
	t.Parallel()

	apiSpec := callbackTemplateVarSpec("timeout-order")
	require.NoError(t, apiSpec.Validate())

	outputDir := filepath.Join(t.TempDir(), "timeout-order-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	authSrc := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")

	parseAt := strings.Index(authSrc, "authTimeout := 15 * time.Minute")
	listenAt := strings.Index(authSrc, `net.Listen("tcp"`)
	browserAt := strings.Index(authSrc, "openBrowser(fullURL)")
	require.Positive(t, parseAt)
	require.Positive(t, listenAt)
	require.Positive(t, browserAt)

	assert.Less(t, parseAt, listenAt, "timeout must be parsed before the callback port is bound")
	assert.Less(t, parseAt, browserAt, "timeout must be parsed before the browser is opened")
}

// RFC 8252 §7.3 prescribes the loopback IP literal, but the RFC does not bind
// a provider's registration form: Intuit's portal refuses to accept 127.0.0.1
// as a redirect URI, so a CLI that can only send it cannot authorize at all.
// The failure is silent — the provider accepts the authorize request, the user
// approves, and only then is the redirect refused, so the callback server
// waits out its whole timeout for a request that was never sent.
func TestAuthRedirectHostSpecOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		redirectHost string
		wantHost     string
	}{
		{name: "default is the RFC loopback literal", redirectHost: "", wantHost: "127.0.0.1"},
		{name: "spec can demand localhost", redirectHost: "localhost", wantHost: "localhost"},
		{name: "case is normalized", redirectHost: "LocalHost", wantHost: "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			apiSpec := callbackTemplateVarSpec("redirect-host")
			apiSpec.Auth.RedirectHost = tt.redirectHost
			require.NoError(t, apiSpec.Validate())

			outputDir := filepath.Join(t.TempDir(), "redirect-host-pp-cli")
			require.NoError(t, New(apiSpec, outputDir).Generate())

			authSrc := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")
			assert.Contains(t, authSrc, `redirectHost := "`+tt.wantHost+`"`)
			// The URI is assembled from the resolved host, never a hardcoded one.
			assert.Contains(t, authSrc, `fmt.Sprintf("http://%s:%d/callback", redirectHost,`)
			assert.NotContains(t, authSrc, `"http://127.0.0.1:%d/callback"`)

			requireGeneratedCompiles(t, outputDir)
		})
	}
}

// The host lands in the redirect_uri the provider sends the authorization code
// to, so anything non-loopback would hand that code to another machine.
func TestAuthRedirectHostRejectsNonLoopback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		host    string
		wantErr string
	}{
		{name: "external host", host: "evil.example.com", wantErr: "not a loopback host"},
		{name: "full URL", host: "http://localhost", wantErr: "must be a bare host"},
		{name: "host with port", host: "localhost:8085", wantErr: "must be a bare host"},
		{name: "bare IPv6 literal needs bracketing to be a legal URL host", host: "::1", wantErr: "must be a bare host"},
		{name: "userinfo smuggling", host: "localhost@evil.example.com", wantErr: "must be a bare host"},
		{name: "path smuggling", host: "localhost/../evil", wantErr: "must be a bare host"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			apiSpec := callbackTemplateVarSpec("redirect-host-bad")
			apiSpec.Auth.RedirectHost = tt.host

			err := apiSpec.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// The credential-free probe returns before the flow ever reaches the browser,
// so a resolver placed after it would let a probe report success for a
// redirect host the real login refuses. Ordering is asserted against the
// emitted source so it cannot drift back below either short-circuit.
func TestAuthRedirectHostResolvesBeforeEveryReturn(t *testing.T) {
	t.Parallel()

	apiSpec := callbackTemplateVarSpec("redirect-order")
	require.NoError(t, apiSpec.Validate())

	outputDir := filepath.Join(t.TempDir(), "redirect-order-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	authSrc := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")

	resolveAt := strings.Index(authSrc, "redirectHost := ")
	probeAt := strings.Index(authSrc, `"status":"dry_run"`)
	verifyAt := strings.Index(authSrc, "if cliutil.IsVerifyEnv() {")
	listenAt := strings.Index(authSrc, `net.Listen("tcp"`)
	require.Positive(t, resolveAt)
	require.Positive(t, probeAt)
	require.Positive(t, listenAt)

	assert.Less(t, resolveAt, probeAt, "must resolve before the credential-free probe returns")
	assert.Less(t, resolveAt, verifyAt, "must resolve before the verify short-circuit")
	assert.Less(t, resolveAt, listenAt, "must resolve before the callback port is bound")
	// Exactly one resolver — a second would drift out of sync with the first.
	assert.Equal(t, 1, strings.Count(authSrc, "redirectHost := "))
}
