// Copyright 2026 mvanhorn. Licensed under Apache-2.0. See LICENSE.

package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tokenPinSpec(name string) *spec.APISpec {
	apiSpec := minimalSpec(name)
	apiSpec.BaseURL = "https://api.example.com"
	apiSpec.Auth = spec.AuthConfig{
		Type:             "oauth2_refresh",
		Header:           "Authorization",
		Format:           "Bearer {access_token}",
		AuthorizationURL: "https://auth.example.com/authorize",
		TokenURL:         "https://auth.example.com/oauth/token",
		EnvVars:          []string{name + "_CLIENT_ID", name + "_CLIENT_SECRET"},
	}
	return apiSpec
}

// The token exchange carries client_secret and refresh_token — offline,
// re-mintable credentials. resolveConfigPath accepts an arbitrary --config
// path while credentials still load from the pinned credentials file, so an
// unconstrained token_url override turns "point this CLI at a config file"
// into "hand over that credential". Only the authority is pinned; path and
// query stay overridable so provider-specific rewrites keep working.
func TestOAuth2TokenURLOverrideIsHostPinned(t *testing.T) {
	t.Parallel()

	outputDir := filepath.Join(t.TempDir(), "tokenpin-pp-cli")
	require.NoError(t, New(tokenPinSpec("tokenpin"), outputDir).Generate())

	configGo := readGeneratedFile(t, outputDir, "internal", "config", "config.go")
	assert.Contains(t, configGo, "func ResolveTokenURL(override string) (string, error)")
	assert.Contains(t, configGo, `const specTokenURL = "https://auth.example.com/oauth/token"`)
	assert.Contains(t, configGo, `got.Scheme != "https"`)
	assert.Contains(t, configGo, "strings.EqualFold(got.Host, spec.Host)")

	// Every consumer must route through the helper. A site still reading the
	// raw config field would be silently unpinned, which is the whole defect.
	for _, file := range []string{"auth.go", "client.go"} {
		var emitted string
		switch file {
		case "client.go":
			emitted = readGeneratedFile(t, outputDir, "internal", "client", "client.go")
		default:
			emitted = readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")
		}
		assert.Containsf(t, emitted, "config.ResolveTokenURL(", "%s must resolve the token URL through the pin", file)
	}

	requireGeneratedCompiles(t, outputDir)
}

// A compile-level assertion cannot show that a hostile host is actually
// refused, so exercise the emitted helper itself.
func TestGeneratedResolveTokenURLRejectsForeignHosts(t *testing.T) {
	t.Parallel()

	outputDir := filepath.Join(t.TempDir(), "tokenpinrt-pp-cli")
	require.NoError(t, New(tokenPinSpec("tokenpinrt"), outputDir).Generate())

	probe := `package config

import "testing"

func TestResolveTokenURLPin(t *testing.T) {
	const want = "https://auth.example.com/oauth/token"
	if got, err := ResolveTokenURL(""); err != nil || got != want {
		t.Fatalf("empty override: got %q err %v, want %q", got, err, want)
	}
	// Path and query variation stays allowed — tenant and regional rewrites.
	if _, err := ResolveTokenURL("https://auth.example.com/oauth/v2/token?x=1"); err != nil {
		t.Fatalf("same-host path override rejected: %v", err)
	}
	for _, bad := range []string{
		"https://evil.example/token",
		"HTTPS://EVIL.EXAMPLE/token",
		"http://auth.example.com/oauth/token",
		"https://auth.example.com:8443/oauth/token",
		"https://auth.example.com.evil.test/token",
	} {
		if _, err := ResolveTokenURL(bad); err == nil {
			t.Errorf("override %q was accepted; it must be refused", bad)
		}
	}
}
`
	probePath := filepath.Join(outputDir, "internal", "config", "zz_token_pin_probe_test.go")
	require.NoError(t, os.WriteFile(probePath, []byte(probe), 0o644))
	t.Cleanup(func() { _ = os.Remove(probePath) })

	// runGoCommandRequired fails the test on a non-zero exit, so a rejected
	// override that slipped through surfaces here as a real failure.
	runGoCommandRequired(t, outputDir, "test", "./internal/config/", "-run", "TestResolveTokenURLPin")
}
