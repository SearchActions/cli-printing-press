// Copyright 2026 mvanhorn. Licensed under Apache-2.0. See LICENSE.

package generator

import (
	"path/filepath"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pathResolutionSpec(name string) *spec.APISpec {
	apiSpec := minimalSpec(name)
	apiSpec.BaseURL = "https://www.example.com"
	apiSpec.Auth = spec.AuthConfig{
		Type:         "cookie",
		Header:       "Cookie",
		In:           "cookie",
		CookieDomain: ".example.com",
		Cookies:      []string{"session_id"},
		EnvVars:      []string{name + "_SESSION"},
	}
	return apiSpec
}

// The cookie-auth flow reads the browser's cookie database, so anything it
// resolves through PATH executes in the user's context on the one code path
// that handles credentials — and does so before a single cookie is read.
// Every printed CLI already links modernc.org/sqlite for its own store, so
// shelling out to an external sqlite3 buys nothing and costs a hijack surface.
func TestCookieAuthReadsCookieDBThroughLinkedDriver(t *testing.T) {
	t.Parallel()

	outputDir := filepath.Join(t.TempDir(), "pathres-pp-cli")
	require.NoError(t, New(pathResolutionSpec("pathres"), outputDir).Generate())

	authGo := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")

	assert.NotContains(t, authGo, `exec.Command("sqlite3"`,
		"cookie DB must be read through the linked driver, not an external sqlite3")
	assert.NotContains(t, authGo, `exec.LookPath("sqlite3")`,
		"profile discovery must not gate on an external sqlite3 being installed")

	assert.Contains(t, authGo, `sql.Open("sqlite", tmpPath)`)
	assert.Contains(t, authGo, "host_key LIKE ?",
		"cookie queries must bind parameters rather than interpolate them")
	// Retired with the string-interpolated queries it existed to escape; a
	// leftover copy would be dead code inviting a new interpolated call site.
	assert.NotContains(t, authGo, "func sqlQuoteLiteral(")

	requireGeneratedCompiles(t, outputDir)
}

// Running a bare PATH-resolved name to learn whether it exists is a
// code-execution primitive wearing a probe's clothes: `cookies` and
// `cookie-scoop` are short, unowned names, so anything shadowing them on a
// writable PATH entry runs from `auth login` alone. exec.LookPath answers the
// same question without executing anything.
func TestCookieToolDetectionResolvesWithoutExecuting(t *testing.T) {
	t.Parallel()

	outputDir := filepath.Join(t.TempDir(), "toolprobe-pp-cli")
	require.NoError(t, New(pathResolutionSpec("toolprobe"), outputDir).Generate())

	for _, file := range []string{"auth.go", "doctor.go"} {
		emitted := readGeneratedFile(t, outputDir, "internal", "cli", file)
		assert.NotContainsf(t, emitted, `"--help").Run()`,
			"%s must not execute a binary to test for its existence", file)
		assert.NotContainsf(t, emitted, `exec.Command("cookies"`+", \"--help\"", file)
		assert.NotContainsf(t, emitted, `exec.Command("cookie-scoop"`+", \"--help\"", file)
	}

	authGo := readGeneratedFile(t, outputDir, "internal", "cli", "auth.go")
	assert.Contains(t, authGo, `exec.LookPath("cookies")`)
	assert.Contains(t, authGo, `exec.LookPath("cookie-scoop")`)

	doctorGo := readGeneratedFile(t, outputDir, "internal", "cli", "doctor.go")
	assert.Contains(t, doctorGo, `exec.LookPath(name)`)
	// Module importability has no test but running it, so this one probe stays
	// an exec — python3 is a long, conventionally-owned name, unlike `cookies`.
	assert.Contains(t, doctorGo, `exec.Command("python3", "-c", "import pycookiecheat")`)

	requireGeneratedCompiles(t, outputDir)
}
