package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeResourceFileStem(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		stem     string
		expected string
		why      string
	}{
		{
			name:     "ordinary single-word stem unchanged",
			stem:     "store",
			expected: "store",
			why:      "no collision; baseline pass-through",
		},
		{
			name:     "ordinary multi-word stem unchanged",
			stem:     "scheduling_window_days",
			expected: "scheduling_window_days",
			why:      "last token is not a GOOS/GOARCH",
		},
		{
			name:     "reserved name not handled here (parse-time validation rejects it)",
			stem:     "feedback",
			expected: "feedback",
			why:      "reserved-name collision is rejected by spec.validateReservedNames; safeResourceFileStem only handles GOOS/GOARCH renames",
		},
		{
			name:     "GOOS suffix triggers rename",
			stem:     "scheduling_windows",
			expected: "scheduling_windows_cmd",
			why:      "Go treats *_windows.go as Windows-only build constraint",
		},
		{
			name:     "GOOS suffix linux",
			stem:     "metrics_linux",
			expected: "metrics_linux_cmd",
			why:      "Linux-only build tag would silently exclude the file",
		},
		{
			name:     "GOARCH suffix triggers rename",
			stem:     "cpu_amd64",
			expected: "cpu_amd64_cmd",
			why:      "Go treats *_amd64.go as amd64-only build constraint",
		},
		{
			name:     "GOOS_GOARCH suffix triggers rename",
			stem:     "build_linux_amd64",
			expected: "build_linux_amd64_cmd",
			why:      "Combined GOOS+GOARCH is also a build-constraint pattern",
		},
		{
			name:     "stem ending in non-token even if matches partial token unchanged",
			stem:     "scheduling_window_days",
			expected: "scheduling_window_days",
			why:      "'days' is not a GOOS/GOARCH; only exact-suffix tokens trigger",
		},
		{
			name:     "embedded GOOS in middle position is fine",
			stem:     "windows_special",
			expected: "windows_special",
			why:      "'windows' is only a build constraint when it's the trailing token",
		},
		{
			name:     "single token GOOS by itself unchanged",
			stem:     "windows",
			expected: "windows",
			why:      "bare 'windows.go' has no underscore prefix, so build-tag rule does not match",
		},
		{
			name:     "GOOS as resource alone produces *_<endpoint>.go pattern handled by caller",
			stem:     "linux_list",
			expected: "linux_list",
			why:      "trailing 'list' is not a GOOS/GOARCH; the caller passes the full <resource>_<endpoint> stem so we evaluate the trailing token",
		},
		{
			name:     "GOOS_GOARCH where penultimate is not actually GOOS unchanged",
			stem:     "store_arm64_endpoint",
			expected: "store_arm64_endpoint",
			why:      "'endpoint' is not a GOARCH/GOOS; only triggers on the last token",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := safeResourceFileStem(tc.stem)
			assert.Equal(t, tc.expected, got, tc.why)
		})
	}
}

func TestSafeResourceFileStem_AllGOOSTokensCovered(t *testing.T) {
	t.Parallel()
	// Every GOOS token must trigger the rename when present as the trailing
	// segment. This keeps the goosTokens map honest if Go adds new platforms.
	for token := range goosTokens {
		got := safeResourceFileStem("res_" + token)
		assert.Equal(t, "res_"+token+"_cmd", got, "GOOS token %q should trigger rename", token)
	}
}

func TestSafeResourceFileStem_AllGOARCHTokensCovered(t *testing.T) {
	t.Parallel()
	for token := range goarchTokens {
		got := safeResourceFileStem("res_" + token)
		assert.Equal(t, "res_"+token+"_cmd", got, "GOARCH token %q should trigger rename", token)
	}
}

// (The reserved-name set lives in internal/spec.ReservedCLIResourceNames, where
// the parser rejects colliding resource names. Tests for it live in
// internal/spec/spec_test.go alongside the rest of validation.)
