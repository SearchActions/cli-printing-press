package generator

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGeneratedOutputIsGofmtClean is a regression test for the bug where
// renderTemplate wrote rendered Go to disk without running format.Source,
// causing every printed CLI's tree to need ~210 of 229 .go files
// reformatted by an out-of-band `go fmt ./...` pass.
//
// The contract: after Generate() returns, every .go file under the output
// directory must be byte-identical to format.Source(content). If a future
// edit to renderTemplate or one of its callers reintroduces an unformatted
// write path, this test fails before any printed CLI ships.
func TestGeneratedOutputIsGofmtClean(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("gofmt-clean")
	outputDir := filepath.Join(t.TempDir(), "gofmt-clean-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	var unformatted []string
	walkErr := filepath.WalkDir(outputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		formatted, fmtErr := format.Source(content)
		if fmtErr != nil {
			// Unparseable Go is a separate failure class — surface it here
			// rather than letting it look like a formatting drift.
			rel, _ := filepath.Rel(outputDir, path)
			unformatted = append(unformatted, rel+" (unparseable: "+fmtErr.Error()+")")
			return nil
		}
		if string(formatted) != string(content) {
			rel, _ := filepath.Rel(outputDir, path)
			unformatted = append(unformatted, rel)
		}
		return nil
	})
	require.NoError(t, walkErr)

	if len(unformatted) > 0 {
		t.Errorf("generated tree has %d Go files that are not gofmt-clean:\n  %s",
			len(unformatted), strings.Join(unformatted, "\n  "))
	}
}
