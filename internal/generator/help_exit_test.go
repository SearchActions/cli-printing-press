package generator

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGeneratedNestedHelpExitsZeroAndUsageErrorsExitTwo(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("help-exit")
	apiSpec.Resources["items"] = spec.Resource{
		Description: "Manage items",
		Endpoints: map[string]spec.Endpoint{
			"list": {Method: "GET", Path: "/items", Description: "List items"},
			"find": {
				Method:      "GET",
				Path:        "/items/find",
				Description: "Find an item",
				Params: []spec.Param{
					{Name: "query", Type: "string", Required: true},
				},
			},
			"compare": {
				Method:      "GET",
				Path:        "/items/{left_id}/compare/{right_id}",
				Description: "Compare two items",
				Params: []spec.Param{
					{Name: "left_id", Type: "string", Required: true, Positional: true, PathParam: true},
					{Name: "right_id", Type: "string", Required: true, Positional: true, PathParam: true},
				},
			},
		},
	}
	apiSpec.Resources["subscribe"] = spec.Resource{
		Description: "Manage subscriptions",
		Endpoints: map[string]spec.Endpoint{
			"create": {
				Method:      "POST",
				Path:        "/subscribe",
				Description: "Create a subscription",
				Body: []spec.Param{
					{Name: "email", Type: "string", Required: true},
				},
			},
		},
	}
	// Single-endpoint resources are always promoted (site C2: a required
	// positional on a promoted top-level command), so this resource has
	// exactly one endpoint.
	apiSpec.Resources["widget"] = spec.Resource{
		Description: "Manage widgets",
		Endpoints: map[string]spec.Endpoint{
			"get": {
				Method:      "GET",
				Path:        "/widgets/{widget_id}",
				Description: "Get a widget",
				Params: []spec.Param{
					{Name: "widget_id", Type: "string", Required: true, Positional: true, PathParam: true},
				},
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), "help-exit-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())
	requireGeneratedCompiles(t, outputDir)

	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	binPath := filepath.Join(outputDir, "help-exit-pp-cli"+exe)
	runGoCommandRequired(t, outputDir, "build", "-o", "./help-exit-pp-cli"+exe, "./cmd/help-exit-pp-cli")

	assertExitCode(t, 0, binPath, "items", "list", "--help")
	assertExitCode(t, 0, binPath, "auth", "status", "--help")
	assertExitCode(t, 2, binPath, "--bogus-flag")
	assertExitCode(t, 2, binPath, "items", "compare", "left-only", "--json")

	// A missing required positional is a usage error (exit 2) in every output
	// mode, human included — not exit-0 help. Before the fix a bare leaf
	// invocation fell through to cobra help and exited 0, so agents read a
	// usage error as a binary bug (#3632). --json adds a structured envelope on
	// stdout; the exit code is 2 either way.
	humanMissing := assertExitCode(t, 2, binPath, "items", "compare")
	require.Contains(t, humanMissing, "missing required argument")
	jsonMissing := assertExitCode(t, 2, binPath, "items", "compare", "--json")
	require.Contains(t, jsonMissing, `"missing required argument"`)

	for _, args := range [][]string{
		{"items", "find", "--json"},
		{"items", "find", "--agent"},
		{"subscribe", "--json"},
		{"subscribe", "--agent"},
	} {
		missingInput := assertExitCode(t, 2, binPath, args...)
		require.Contains(t, missingInput, `"requires input"`, "args %v output:\n%s", args, missingInput)
	}

	// An unknown/misspelled subcommand on a parent group is a usage error in
	// every output mode. Before the fix human mode fell through to exit-0 help
	// (#2955); machine mode already exited 2.
	humanUnknown := assertExitCode(t, 2, binPath, "items", "bogus")
	require.Contains(t, humanUnknown, `unknown subcommand "bogus"`)
	jsonUnknown := assertExitCode(t, 2, binPath, "items", "bogus", "--json")
	require.Contains(t, jsonUnknown, `"unknown subcommand"`)

	// A genuine bare parent invocation still prints help and exits 0 for humans
	// (only a leftover/typo'd token is an error), preserving the friendly UX.
	assertExitCode(t, 0, binPath, "items")

	// Sites A/C: every machine-output mode (not just --json/--agent) must
	// exit 2 with no help prose on stdout. Before the fix, --quiet/--csv/
	// --plain/--compact/--select got 2.5+ KB of cobra help under exit 0.
	for _, leaf := range [][]string{{"items", "find"}, {"subscribe"}} {
		for _, mode := range []string{"--quiet", "--csv", "--plain", "--compact", "--select=error"} {
			args := append(append([]string{}, leaf...), mode)
			stdout, _, code := runSplit(t, binPath, args...)
			require.Equal(t, 2, code, "args %v", args)
			require.NotContains(t, stdout, "Usage:", "args %v stdout:\n%s", args, stdout)
			if mode == "--quiet" {
				require.Empty(t, stdout, "args %v stdout:\n%s", args, stdout)
			} else {
				require.Contains(t, stdout, "requires input", "args %v stdout:\n%s", args, stdout)
			}
		}
	}

	// Human controls: prove the gate above is conditional, not always true.
	// A bare invocation with no machine-output flag still prints help and
	// exits 0.
	for _, leaf := range [][]string{{"items", "find"}, {"subscribe"}} {
		stdout, _, code := runSplit(t, binPath, leaf...)
		require.Equal(t, 0, code, "args %v", leaf)
		require.Contains(t, stdout, "Usage:", "args %v stdout:\n%s", leaf, stdout)
	}

	// Site B (parent group): every machine-output mode exits 2, and --quiet
	// suppresses the raw-encoder envelope the same way it suppresses help.
	stdout, _, code := runSplit(t, binPath, "items", "--csv")
	require.Equal(t, 2, code)
	require.Contains(t, stdout, "subcommand required")

	stdout, _, code = runSplit(t, binPath, "items", "--quiet")
	require.Equal(t, 2, code)
	require.Empty(t, stdout)

	stdout, _, code = runSplit(t, binPath, "items", "bogus", "--quiet")
	require.Equal(t, 2, code)
	require.Empty(t, stdout)

	stdout, _, code = runSplit(t, binPath, "items", "bogus", "--csv")
	require.Equal(t, 2, code)
	require.Contains(t, stdout, "unknown subcommand")

	// Site C1: missing required positional under a non-json machine mode.
	stdout, _, code = runSplit(t, binPath, "items", "compare", "--csv")
	require.Equal(t, 2, code)
	require.Contains(t, stdout, "missing required argument")

	stdout, _, code = runSplit(t, binPath, "items", "compare", "--quiet")
	require.Equal(t, 2, code)
	require.Empty(t, stdout)

	// Site C2: missing required positional on a promoted (single-endpoint)
	// top-level command.
	stdout, _, code = runSplit(t, binPath, "widget", "--csv")
	require.Equal(t, 2, code)
	require.Contains(t, stdout, "is required")

	stdout, _, code = runSplit(t, binPath, "widget", "--quiet")
	require.Equal(t, 2, code)
	require.Empty(t, stdout)
}

// runSplit runs the generated binary with an isolated home directory (so
// rows that reach flags.newClient() never read the developer's real config)
// and returns stdout/stderr separately, unlike assertExitCode's merged
// CombinedOutput, which can't prove stdout is empty or free of help text.
func runSplit(t *testing.T, binaryPath string, args ...string) (stdout, stderr string, code int) {
	t.Helper()

	home := t.TempDir()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	code = 0
	if err != nil {
		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr, "args %v stderr:\n%s", args, errBuf.String())
		code = exitErr.ExitCode()
	}
	return outBuf.String(), errBuf.String(), code
}

// TestGeneratedRecallLearningsForgetAndSearchExitTwoUnderMachineOutput covers
// sites D1-D3: hand-templated commands (recall, learnings forget, search)
// that printed help under exit 0 in every mode, --json/--agent included,
// before the fix.
func TestGeneratedRecallLearningsForgetAndSearchExitTwoUnderMachineOutput(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("help-exit-learn")
	apiSpec.Learn = spec.LearnConfig{Enabled: true, EnabledSet: true}

	outputDir := filepath.Join(t.TempDir(), "help-exit-learn-pp-cli")
	gen := New(apiSpec, outputDir)
	// Force Search on: the minimal fixture has no search-shaped endpoint for
	// the profiler to score, but the D3 defect is in the hand-templated
	// search.go, not in whether a real spec would enable it.
	gen.VisionSet = VisionTemplateSet{Search: true, Store: true, Sync: true, Export: true, Import: true, MCP: true}
	require.NoError(t, gen.Generate())
	requireGeneratedCompiles(t, outputDir)

	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	binPath := filepath.Join(outputDir, "help-exit-learn-pp-cli"+exe)
	runGoCommandRequired(t, outputDir, "build", "-o", "./help-exit-learn-pp-cli"+exe, "./cmd/help-exit-learn-pp-cli")

	for _, tc := range []struct {
		args []string
	}{
		{[]string{"search", "--csv"}},
		{[]string{"recall", "--agent"}},
	} {
		stdout, _, code := runSplit(t, binPath, tc.args...)
		require.Equal(t, 2, code, "args %v", tc.args)
		require.Contains(t, stdout, "requires input", "args %v stdout:\n%s", tc.args, stdout)
		require.NotContains(t, stdout, "Usage:", "args %v stdout:\n%s", tc.args, stdout)
	}

	stdout, _, code := runSplit(t, binPath, "learnings", "forget", "--quiet")
	require.Equal(t, 2, code)
	require.Empty(t, stdout)

	// Human controls: bare invocation with no machine-output flag still
	// prints help and exits 0.
	for _, leaf := range [][]string{{"search"}, {"recall"}} {
		stdout, _, code := runSplit(t, binPath, leaf...)
		require.Equal(t, 0, code, "args %v", leaf)
		require.Contains(t, stdout, "Usage:", "args %v stdout:\n%s", leaf, stdout)
	}
}

// TestGeneratedNovelFeatureStubWithPositionalExitsTwoUnderMachineOutput
// covers site D4: the novel-feature scaffold's bare-positional branch.
func TestGeneratedNovelFeatureStubWithPositionalExitsTwoUnderMachineOutput(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("help-exit-novel")
	outputDir := filepath.Join(t.TempDir(), "help-exit-novel-pp-cli")
	gen := New(apiSpec, outputDir)
	gen.NovelFeatures = []NovelFeature{
		{
			Name:        "Widget classifier",
			Command:     "classify <id>",
			Description: "Classify a widget by ID.",
			Rationale:   "Agents need a bounded classification lookup.",
			Example:     "help-exit-novel-pp-cli classify widget-123",
		},
	}
	require.NoError(t, gen.Generate())

	stub := readGeneratedFile(t, outputDir, "internal", "cli", "classify.go")
	require.Contains(t, stub, "if wantsMachineOutput(flags) {")
	require.Contains(t, stub, `"error": "requires input"`)
	require.Contains(t, stub, "return usageErr(")

	requireGeneratedCompiles(t, outputDir)
}

func TestGeneratedRootTreatsPflagHelpSentinelAsSuccess(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("help-sentinel")
	outputDir := filepath.Join(t.TempDir(), "help-sentinel-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	rootSrc := readGeneratedFile(t, outputDir, "internal", "cli", "root.go")
	goMod := readGeneratedFile(t, outputDir, "go.mod")
	require.Contains(t, rootSrc, `"errors"`)
	require.Contains(t, rootSrc, `"github.com/spf13/pflag"`)
	require.Contains(t, rootSrc, "errors.Is(err, pflag.ErrHelp)")
	require.Contains(t, goMod, "github.com/spf13/pflag v1.0.6")
	require.Less(t,
		strings.Index(rootSrc, "errors.Is(err, pflag.ErrHelp)"),
		strings.Index(rootSrc, "isCobraUsageError(err)"),
		"help sentinel must be handled before Cobra usage errors are wrapped as exit code 2",
	)
}

func assertExitCode(t *testing.T, want int, binaryPath string, args ...string) string {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	output, err := cmd.CombinedOutput()
	if want == 0 {
		require.NoError(t, err, "args %v output:\n%s", args, string(output))
		return string(output)
	}
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "args %v output:\n%s", args, string(output))
	require.Equal(t, want, exitErr.ExitCode(), "args %v output:\n%s", args, string(output))
	return string(output)
}
