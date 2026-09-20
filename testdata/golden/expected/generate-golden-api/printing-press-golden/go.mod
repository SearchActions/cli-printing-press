module printing-press-golden-pp-cli

go 1.26.6

toolchain go1.26.6

require (
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/pelletier/go-toml/v2 v2.2.4
)
require modernc.org/sqlite v1.37.0
require github.com/mark3labs/mcp-go v0.47.0

// x/sys is a DIRECT dependency on every bundle: the private-file permission
// surfaces in internal/cliutil (privperms_windows.go, and for token-bearing
// bundles creds_perms_windows.go) import golang.org/x/sys/windows. Emitted
// as a direct require (no // indirect) so a freshly generated bundle's
// go.mod is correct out of the box, WITHOUT a manual `go mod tidy`; tidy
// considers all build tags, so the marker survives a tidy under any GOOS.
// The version also floors x/sys above the vulnerable v0.31.0 for CLIs that
// would otherwise pull it only transitively (modernc.org/sqlite,
// golang.org/x/net, ...).
require golang.org/x/sys v0.46.0
