package mcpsync

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/generator"
	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func writeGoMod(t *testing.T, cliDir, gomod string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(cliDir, "go.mod"), []byte(gomod), 0o644))
}

func readGoMod(t *testing.T, cliDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(cliDir, "go.mod"))
	require.NoError(t, err)
	return string(data)
}

// TestEnsureXSysRequire table — the regenerated cliutil imports
// golang.org/x/sys/windows on every sync, so the require must exist, be
// direct, and sit at or above the floor. A replace directive does not put
// the module in the build graph, so it never removes the need for a
// require - it only decides WHICH version that require must name.
func TestEnsureXSysRequire(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		gomod   string
		changed bool
		assert  func(t *testing.T, after string)
	}{
		{
			name:    "absent added as direct at floor",
			gomod:   "module example.com/foo\n\ngo 1.23.0\n",
			changed: true,
			assert: func(t *testing.T, after string) {
				assert.Contains(t, after, "golang.org/x/sys v0.46.0")
				assert.NotContains(t, after, "golang.org/x/sys v0.46.0 // indirect")
			},
		},
		{
			name:    "indirect at floor becomes direct",
			gomod:   "module example.com/foo\n\nrequire (\n\tgolang.org/x/sys v0.46.0 // indirect\n)\n",
			changed: true,
			assert: func(t *testing.T, after string) {
				assert.Contains(t, after, "golang.org/x/sys v0.46.0")
				assert.NotContains(t, after, "golang.org/x/sys v0.46.0 // indirect")
			},
		},
		{
			name:    "below floor raised to direct",
			gomod:   "module example.com/foo\n\nrequire golang.org/x/sys v0.31.0\n",
			changed: true,
			assert: func(t *testing.T, after string) {
				assert.Contains(t, after, "golang.org/x/sys v0.46.0")
				assert.NotContains(t, after, "golang.org/x/sys v0.31.0")
			},
		},
		{
			name:    "already direct at floor is byte-identical",
			gomod:   "module example.com/foo\n\nrequire golang.org/x/sys v0.46.0\n",
			changed: false,
			assert: func(t *testing.T, after string) {
				assert.Equal(t, "module example.com/foo\n\nrequire golang.org/x/sys v0.46.0\n", after)
			},
		},
		{
			name:    "above floor is left alone, never downgraded",
			gomod:   "module example.com/foo\n\nrequire golang.org/x/sys v0.50.0\n",
			changed: false,
			assert: func(t *testing.T, after string) {
				assert.Contains(t, after, "golang.org/x/sys v0.50.0")
				assert.NotContains(t, after, "v0.46.0", "the floor must never downgrade a higher pin")
			},
		},
		{
			// A replace does not put the module in the build graph, so a
			// require still has to. An UNVERSIONED replace applies to
			// whatever version is required, leaving the floor free to win.
			name: "unversioned replace still gets a direct require at the floor",
			gomod: "module example.com/foo\n\nrequire golang.org/x/sys v0.31.0 // indirect\n\n" +
				"replace golang.org/x/sys => ../local-x-sys\n",
			changed: true,
			assert: func(t *testing.T, after string) {
				assert.Contains(t, after, "replace golang.org/x/sys => ../local-x-sys", "the replacement must survive")
				assert.Contains(t, after, "golang.org/x/sys v0.46.0")
				assert.NotContains(t, after, "// indirect")
			},
		},
		{
			// A VERSION-SPECIFIC replace applies only to that exact
			// version, so requiring the floor instead would silently
			// deactivate the replacement.
			name: "version-specific replace keeps its own version, not the floor",
			gomod: "module example.com/foo\n\nrequire golang.org/x/sys v0.40.0 // indirect\n\n" +
				"replace golang.org/x/sys v0.40.0 => ../local-x-sys\n",
			changed: true,
			assert: func(t *testing.T, after string) {
				assert.Contains(t, after, "replace golang.org/x/sys v0.40.0 => ../local-x-sys")
				assert.Contains(t, after, "golang.org/x/sys v0.40.0")
				assert.NotContains(t, after, "v0.46.0", "moving to the floor would deactivate the replacement")
				assert.NotContains(t, after, "// indirect")
			},
		},
		{
			name: "version-specific replace already direct is byte-identical",
			gomod: "module example.com/foo\n\nrequire golang.org/x/sys v0.40.0\n\n" +
				"replace golang.org/x/sys v0.40.0 => ../local-x-sys\n",
			changed: false,
			assert: func(t *testing.T, after string) {
				assert.Equal(t, "module example.com/foo\n\nrequire golang.org/x/sys v0.40.0\n\n"+
					"replace golang.org/x/sys v0.40.0 => ../local-x-sys\n", after)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cliDir := t.TempDir()
			writeGoMod(t, cliDir, tc.gomod)

			changed, err := ensureXSysRequire(cliDir)
			require.NoError(t, err)
			assert.Equal(t, tc.changed, changed)
			tc.assert(t, readGoMod(t, cliDir))
		})
	}
}

// TestSyncRetrofitsPrivpermsAndXSys — end-to-end regression over the real
// Sync entry point. An old library tree (own currentUserSID token lookup,
// no privperms pair, no x/sys require — the cube-cli shape) must sync into
// a state that builds for GOOS=windows under -mod=readonly: no duplicate
// currentUserSID, createPrivateTemp defined, x/sys direct. Skipped under
// -short: the go.mod repair runs a real go mod tidy, which needs the
// module proxy for go.sum hashes.
func TestSyncRetrofitsPrivpermsAndXSys(t *testing.T) {
	if testing.Short() {
		t.Skip("needs network for go mod tidy after the x/sys require repair")
	}
	t.Parallel()

	apiSpec := &spec.APISpec{
		Name:    "privretro",
		Version: "0.1.0",
		BaseURL: "https://api.example.com",
		Auth: spec.AuthConfig{
			Type:    "api_key",
			Header:  "Authorization",
			Format:  "Bearer {token}",
			EnvVars: []string{"PRIVRETRO_TOKEN"},
		},
		Config: spec.ConfigSpec{Format: "toml", Path: "~/.config/privretro-pp-cli/config.toml"},
		Resources: map[string]spec.Resource{
			"items": {
				Description: "Manage items",
				Endpoints: map[string]spec.Endpoint{
					"list": {Method: "GET", Path: "/items", Description: "List items"},
				},
			},
		},
	}

	cliDir := filepath.Join(t.TempDir(), "privretro")
	require.NoError(t, generator.New(apiSpec, cliDir).Generate())

	specData, err := yaml.Marshal(apiSpec)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(cliDir, "spec.yaml"), specData, 0o644))

	// Age the tree to the pre-privperms shape: no privperms pair existed,
	// the Windows perms guard was an x/sys-free stub (nothing anywhere in
	// the tree imported x/sys — the exact cube-cli shape), and go.mod/go.sum
	// carried no x/sys entries. The stub matters: with a live importer a
	// premature `go mod tidy` keeps the require, and the ordering bug this
	// test pins (tidy running before the emission drops it again) goes
	// unnoticed.
	oldCredsWin := "//go:build windows\n\npackage cliutil\n\n" +
		"func currentUserSID() (string, error) { return \"\", nil }\n\n" +
		"func VerifyCredsPerms(path string) error { return nil }\n"
	require.NoError(t, os.WriteFile(filepath.Join(cliDir, "internal", "cliutil", "creds_perms_windows.go"), []byte(oldCredsWin), 0o644))
	require.NoError(t, os.Remove(filepath.Join(cliDir, "internal", "cliutil", "privperms_windows.go")))
	require.NoError(t, os.Remove(filepath.Join(cliDir, "internal", "cliutil", "privperms_unix.go")))
	for _, name := range []string{"go.mod", "go.sum"} {
		path := filepath.Join(cliDir, name)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		require.NoError(t, err)
		var kept []string
		for line := range strings.SplitSeq(string(data), "\n") {
			if strings.Contains(line, "golang.org/x/sys") {
				continue
			}
			kept = append(kept, line)
		}
		require.NoError(t, os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644))
	}

	// Old tools.go so Sync takes the migration path into GenerateMCPSurface.
	oldTools := "// Generated by CLI Printing Press (https://github.com/mvanhorn/cli-printing-press). DO NOT EDIT.\n" +
		"package mcp\n\nfunc RegisterNovelFeatureTools() { shellOutToCLI(\"items list\") }\n"
	require.NoError(t, os.WriteFile(filepath.Join(cliDir, "internal", "mcp", "tools.go"), []byte(oldTools), 0o644))

	result, err := Sync(cliDir, Options{})
	require.NoError(t, err)
	assert.True(t, result.Changed)

	credSrc, err := os.ReadFile(filepath.Join(cliDir, "internal", "cliutil", "creds_perms_windows.go"))
	require.NoError(t, err)
	winSrc, err := os.ReadFile(filepath.Join(cliDir, "internal", "cliutil", "privperms_windows.go"))
	require.NoError(t, err)
	combined := string(credSrc) + string(winSrc)
	assert.Equal(t, 1, strings.Count(combined, "func currentUserSID"),
		"currentUserSID must have exactly one definition across the two files")
	assert.Contains(t, string(winSrc), "func processUserSID")
	assert.Contains(t, string(winSrc), "func createPrivateTemp")

	gomod := readGoMod(t, cliDir)
	assert.Contains(t, gomod, "golang.org/x/sys v0.46.0")
	assert.NotContains(t, gomod, "golang.org/x/sys v0.46.0 // indirect")

	// -mod=readonly: a missing require or go.sum hash fails here instead of
	// being quietly repaired by -mod=mod. This is the exact build an
	// unsynced cube-cli would have broken on Windows.
	for _, goos := range []string{"windows", runtime.GOOS} {
		cmd := exec.Command("go", "build", "-mod=readonly", "./...")
		cmd.Dir = cliDir
		cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH=amd64", "CGO_ENABLED=0")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "GOOS=%s: %s", goos, out)
	}
}
