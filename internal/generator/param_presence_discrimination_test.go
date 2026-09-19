package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParamPresenceWireProof_DiscriminationCheck proves that cases (a), (d)
// and (e) in TestParamPresenceWireProof (param_presence_runtime_test.go)
// genuinely discriminate old vs. new generator behavior, not just old vs.
// new emitted text: it copies that sibling file verbatim into a clean git
// worktree checked out at the commit this branch forked from, builds and
// runs the SAME test there against the OLD generator package, and asserts
// those three subtests fail there. A test that passes on both old and new
// code proves nothing (positive-control rule) — this is what makes
// TestParamPresenceWireProof an actual regression test rather than a test
// of the new code's own logic.
func TestParamPresenceWireProof_DiscriminationCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("generated CLI compile/build tests run in the full generated-test CI lane")
	}

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed to resolve this file's path")
	generatorDir := filepath.Dir(thisFile)
	repoRoot := filepath.Dir(filepath.Dir(generatorDir))
	if _, statErr := os.Stat(filepath.Join(repoRoot, "go.mod")); statErr != nil {
		t.Skipf("could not locate repo root two levels above %s: %v", generatorDir, statErr)
	}

	// The presence-gate fix lands as uncommitted working-tree changes on
	// this branch (see git status: generator.go and the three templates are
	// all "M", not part of any commit yet), so HEAD itself — not
	// merge-base(HEAD, origin/main) — is the pre-change baseline. A clean
	// `git worktree add` of HEAD naturally excludes uncommitted changes.
	// Guard against the case where the fix has since been committed (no
	// uncommitted diff left to discriminate against) so this degrades to a
	// skip instead of a false pass.
	dirty := strings.TrimSpace(runCapture(t, repoRoot, "git", "status", "--porcelain",
		"internal/generator/generator.go",
		"internal/generator/templates/command_endpoint.go.tmpl",
		"internal/generator/templates/command_promoted.go.tmpl",
		"internal/generator/templates/helpers.go.tmpl"))
	if dirty == "" {
		t.Skip("no uncommitted changes to the presence-gate sources; HEAD would check out the current (fixed) code, so there is nothing to discriminate against here — rely on the commit history instead")
	}

	worktreeDir := filepath.Join(t.TempDir(), "old-generator")
	runOK(t, repoRoot, "git", "worktree", "add", "--detach", "--force", worktreeDir, "HEAD")
	t.Cleanup(func() {
		runOK(t, repoRoot, "git", "worktree", "remove", "--force", worktreeDir)
	})

	// Copy the wire-proof sibling file verbatim so the OLD generator.New /
	// bodyLeafPresenceExpr / templates render the identical fixture spec and
	// run the identical assertions (K-1046: one derivation, not a second
	// hand-maintained copy that can drift from the one already proven
	// meaningful in TestParamPresenceWireProof).
	wireProofSrc, err := os.ReadFile(filepath.Join(generatorDir, "param_presence_runtime_test.go"))
	require.NoError(t, err)
	oldTestPath := filepath.Join(worktreeDir, "internal", "generator", "param_presence_runtime_test.go")
	require.NoError(t, os.WriteFile(oldTestPath, wireProofSrc, 0o644))

	cmd := exec.Command("go", "test", "-mod=mod", "-run", "TestParamPresenceWireProof$", "-v", "./internal/generator/...")
	cmd.Dir = worktreeDir
	out, runErr := cmd.CombinedOutput()
	output := string(out)

	require.Error(t, runErr, "expected the pre-change generator to fail the wire proof; full output:\n%s", output)
	for _, subtest := range []string{
		"--- FAIL: TestParamPresenceWireProof/(a)",
		"--- FAIL: TestParamPresenceWireProof/(d)",
		"--- FAIL: TestParamPresenceWireProof/(e)",
	} {
		require.Contains(t, output, subtest, "expected this subtest to fail against the pre-change generator; full output:\n%s", output)
	}
}

func runCapture(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func runOK(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}
