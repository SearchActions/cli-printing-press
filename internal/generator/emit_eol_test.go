package generator

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mvanhorn/cli-printing-press/v4/internal/naming"
)

// TestNormalizeRenderedCollapsesHostNewlines asserts the emit writer converts
// CRLF and lone CR into LF for non-Go artifacts. go/format.Source already
// enforces LF on .go outputs; markdown/yaml artifacts have no such pass, so
// without this a Windows host with a CRLF-smudged tree prints byte-different
// SKILL.md/README.md/.goreleaser.yaml files.
func TestNormalizeRenderedCollapsesHostNewlines(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		in      string
		want    string
		tmpl    string
		outPath string
	}{
		{
			name:    "crlf markdown",
			in:      "# Title\r\n\r\nBody line one.\r\nBody line two.",
			want:    "# Title\n\nBody line one.\nBody line two.\n",
			tmpl:    "skill.md.tmpl",
			outPath: "SKILL.md",
		},
		{
			name:    "lone carriage returns",
			in:      "one\rtwo\r\nthree",
			want:    "one\ntwo\nthree\n",
			tmpl:    "readme.md.tmpl",
			outPath: "README.md",
		},
		{
			name:    "crlf yaml",
			in:      "version: 2\r\nbuilds:\r\n  - id: cli",
			want:    "version: 2\nbuilds:\n  - id: cli\n",
			tmpl:    "goreleaser.yaml.tmpl",
			outPath: ".goreleaser.yaml",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeRendered([]byte(tc.in), tc.tmpl, tc.outPath)
			require.Equal(t, tc.want, string(got))
			require.NotContains(t, string(got), "\r")
		})
	}
}

// TestEmittedTreeIsCRFree runs a full Generate with CRLF-carrying spec
// content (the content-assembly vector: a spec authored on a CRLF host keeps
// those bytes through yaml parsing into the templates) and asserts no emitted
// file contains a carriage return anywhere in the tree.
func TestEmittedTreeIsCRFree(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("crlf-emit")
	apiSpec.Description = "Line one.\r\nLine two."
	items := apiSpec.Resources["items"]
	items.Description = "Manage items.\r\nMultiline."
	apiSpec.Resources["items"] = items
	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	gen := New(apiSpec, outputDir)
	require.NoError(t, gen.Generate())

	var offenders []string
	require.NoError(t, filepath.WalkDir(outputDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if bytes.ContainsRune(data, '\r') {
			rel, _ := filepath.Rel(outputDir, p)
			offenders = append(offenders, rel)
		}
		return nil
	}))
	require.Empty(t, offenders,
		"emitted artifacts must be LF-only regardless of host newlines; CR found in:\n  %s",
		strings.Join(offenders, "\n  "))
}
