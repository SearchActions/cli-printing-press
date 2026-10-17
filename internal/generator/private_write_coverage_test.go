package generator

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// secretBearingTemplates are the emitted files that persist a credential,
// a session cookie, or a token awaiting exchange. Every write in them must
// go through cliutil.AtomicWritePrivateFile, which attaches an owner-only
// security descriptor at create time.
//
// A bare os.WriteFile with a 0600 literal READS as a permission control and
// is not one on Windows: file-mode bits are inert on NTFS, so the file
// inherits whatever the parent directory carries. On a domain-joined host
// that routinely includes other principals with read access, which hands
// the session to any local account that has it.
var secretBearingTemplates = []string{
	"cookiejar.go.tmpl",
	"session_handshake.go.tmpl",
	"auth_device_code.go.tmpl",
	"cliutil_credentials.go.tmpl",
}

var bareWriteRe = regexp.MustCompile(`\bos\.(WriteFile|Create|OpenFile)\(`)

func TestSecretBearingTemplatesUsePrivateWrites(t *testing.T) {
	t.Parallel()

	for _, name := range secretBearingTemplates {
		path := filepath.Join("templates", name)
		if _, err := os.Stat(path); err != nil {
			// A renamed or retired template must fail loudly rather than
			// silently drop its coverage.
			require.NoError(t, err, "secret-bearing template missing: %s (update secretBearingTemplates if it was renamed)", name)
		}

		src, err := os.ReadFile(path)
		require.NoError(t, err)

		for i, line := range strings.Split(string(src), "\n") {
			if !bareWriteRe.MatchString(line) {
				continue
			}
			require.Failf(t, "secret-bearing template creates a file directly",
				"%s:%d writes a file with %q.\n"+
					"Go file-mode bits are inert on NTFS, so this inherits the parent\n"+
					"directory's access entries. Use cliutil.AtomicWritePrivateFile.",
				name, i+1, strings.TrimSpace(line))
		}
	}
}

// TestPrivateWriteProtectsTheDirectory pins the companion half: protecting
// only the files this package writes leaves every other writer in the same
// directory inheriting the parent's entries, so AtomicWritePrivateFile also
// restricts the directory it just created.
func TestPrivateWriteProtectsTheDirectory(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile(filepath.Join("templates", "cliutil_paths.go.tmpl"))
	require.NoError(t, err)

	fn := string(src)
	start := strings.Index(fn, "func AtomicWritePrivateFile(")
	require.GreaterOrEqual(t, start, 0, "expected AtomicWritePrivateFile in cliutil_paths.go.tmpl")
	body := fn[start:]
	if end := strings.Index(body, "\nfunc "); end >= 0 {
		body = body[:end]
	}

	require.Contains(t, body, "EnsurePrivateTree(dir)",
		"AtomicWritePrivateFile must restrict the directory, not just the file")
}
