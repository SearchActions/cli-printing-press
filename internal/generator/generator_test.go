package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mvanhorn/cli-printing-press/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestGenerateProjectsCompile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		specPath      string
		expectedFiles int
	}{
		{name: "stytch", specPath: filepath.Join("..", "..", "testdata", "stytch.yaml"), expectedFiles: 14},
		{name: "clerk", specPath: filepath.Join("..", "..", "testdata", "clerk.yaml"), expectedFiles: 15},
		{name: "loops", specPath: filepath.Join("..", "..", "testdata", "loops.yaml"), expectedFiles: 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiSpec, err := spec.Parse(tt.specPath)
			require.NoError(t, err)

			outputDir := filepath.Join(t.TempDir(), apiSpec.Name+"-cli")
			gen := New(apiSpec, outputDir)
			require.NoError(t, gen.Generate())

			require.Equal(t, tt.expectedFiles, countFiles(t, outputDir))

			runGoCommand(t, outputDir, "mod", "tidy")
			runGoCommand(t, outputDir, "build", "./...")

			binaryPath := filepath.Join(outputDir, apiSpec.Name+"-cli")
			runGoCommand(t, outputDir, "build", "-o", binaryPath, "./cmd/"+apiSpec.Name+"-cli")

			info, err := os.Stat(binaryPath)
			require.NoError(t, err)
			require.False(t, info.IsDir())
			require.NotZero(t, info.Size())
		})
	}
}

func countFiles(t *testing.T, root string) int {
	t.Helper()

	total := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		require.NoError(t, err)
		if !d.IsDir() {
			total++
		}
		return nil
	})
	require.NoError(t, err)
	return total
}

func runGoCommand(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(dir, ".cache", "go-build"))
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}
