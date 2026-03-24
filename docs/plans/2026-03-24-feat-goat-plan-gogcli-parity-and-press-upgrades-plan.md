---
title: "The GOAT Plan - Print gogcli-Quality CLIs and Close Every Gap"
type: feat
status: completed
date: 2026-03-24
---

# The GOAT Plan - Print gogcli-Quality CLIs and Close Every Gap

## Context

gogcli is Peter Steinberger's Google Workspace CLI: 80 commands across 8 services, OAuth2 with keyring, FTS5 search, multi-account. If the printing press can match gogcli quality, it can print anything.

## What's Already Done

- Deep sub-resource detection with common prefix collapse (Gmail now generates `messages list` not `gmail users-messages-list`)
- Global param filtering (Google's 10 API-wide params stripped from every endpoint)
- Pagination detection and `--all` flag generation (parser detects pageToken/nextPageToken patterns, templates generate pagination loop)
- Retry with exponential backoff
- Dry-run mode
- Table output with column auto-detection
- Structured exit codes
- Usage examples generation

## Scope Boundaries

- No Discovery-to-OpenAPI converter (use pre-converted specs from APIs-guru)
- No file upload/download (binary content handling is a separate plan)
- No Apps Script execution
- Templates only - no runtime library

## Implementation Units

### Unit 1: Color and TTY Detection

**Goal:** Generated CLIs detect terminal vs pipe and apply color/bold formatting. Doctor output gets colored checkmarks. Tables get bold headers.

**Files:**
- `internal/generator/templates/helpers.go.tmpl` - add `isTerminal()`, color helpers
- `internal/generator/templates/doctor.go.tmpl` - colored output
- `internal/generator/templates/go.mod.tmpl` - add `mattn/go-isatty` dependency

**Approach:**
1. Add `mattn/go-isatty v0.0.20` to go.mod.tmpl
2. In helpers.go.tmpl, add:
   - `isTerminal()` that checks `os.Stdout` via isatty
   - `colorEnabled()` that returns false if `NO_COLOR` env is set, `TERM=dumb`, or not a terminal
   - `bold(s string)` and `green(s string)` and `red(s string)` ANSI wrappers that no-op when color disabled
3. In doctor.go.tmpl, replace "PASS"/"FAIL" strings with `green("PASS")` and `red("FAIL")`
4. In helpers.go.tmpl, update `renderTable` to use bold headers when terminal
5. Add `--no-color` persistent flag to root.go.tmpl

**Patterns to follow:** Existing `helpers.go.tmpl` render functions, `root.go.tmpl` persistent flags pattern

**Execution note:** Straightforward, no test-first needed

**Test scenarios:**
- All 4 specs (Petstore, Stytch, Discord, Gmail) still compile after template changes
- Generated CLI `--help` shows `--no-color` flag
- Doctor output includes ANSI codes when run in terminal

**Verification:** `go test ./...` passes, regenerate all 4 specs and `go build` each

---

### Unit 2: URL-Based Spec Input

**Goal:** `printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json` fetches the spec from URL, caches it, and generates.

**Files:**
- `internal/cli/generate.go` (or wherever the generate command lives) - URL detection and fetching
- New file: `internal/cli/fetch.go` - HTTP fetch + cache logic

**Approach:**
1. In the generate command, before reading the spec file, check if `--spec` value starts with `http://` or `https://`
2. If URL: fetch with `net/http`, cache to `~/.cache/printing-press/specs/<sha256-of-url>.json`
3. If cached file exists and is <24h old, use cache. `--refresh` flag forces re-download
4. Auto-detect format from Content-Type header or file extension in URL
5. Pass fetched bytes to existing parse pipeline

**Patterns to follow:** Existing `internal/cli/` command structure

**Execution note:** Find the generate command first - `grep -rn "generate" internal/cli/`

**Test scenarios:**
- `printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json --output /tmp/pet-url-test` works
- Second run uses cache (faster)
- `--refresh` forces re-download
- Invalid URL gives clear error

**Verification:** `go test ./...` passes, manual test with Petstore URL

---

### Unit 3: OAuth2 Auth Flow

**Goal:** Generated CLIs support `auth login` (browser-based OAuth2), `auth status`, `auth logout`. Tokens stored in config file. Auto-refresh on expiry.

**Files:**
- New template: `internal/generator/templates/auth.go.tmpl` - auth login/logout/status commands
- `internal/generator/templates/config.go.tmpl` - add token storage fields
- `internal/generator/templates/client.go.tmpl` - auto-refresh before API calls
- `internal/generator/templates/root.go.tmpl` - register auth command group
- `internal/generator/generator.go` - conditionally render auth.go.tmpl when auth type is oauth2
- `internal/generator/templates/go.mod.tmpl` - add `golang.org/x/oauth2` dependency

**Approach:**
1. Parser already detects OAuth2 (`auth.Type = "bearer_token"` when scheme type is oauth2). Need to also capture the OAuth2 flow URLs (authorization endpoint, token endpoint) from the security scheme.
2. Update `spec.AuthConfig` to include `AuthorizationURL`, `TokenURL`, `Scopes` fields
3. In `parser.go`, extract OAuth2 flow details when `scheme.Type == "oauth2"` and `scheme.Flows.AuthorizationCode` exists
4. New `auth.go.tmpl`:
   - `auth login`: starts local HTTP server on `localhost:8085/callback`, opens browser to authorization URL with redirect_uri, exchanges code for tokens, stores in config.toml
   - `auth status`: reads stored token, shows expiry, account info
   - `auth logout`: removes stored tokens
5. In `config.go.tmpl`, add `AccessToken`, `RefreshToken`, `TokenExpiry`, `ClientID`, `ClientSecret` fields. Load/save to config.toml
6. In `client.go.tmpl`, before each request check if token is expired, if so use refresh_token to get new one, update config
7. In `generator.go`, only render `auth.go.tmpl` when `spec.Auth.Type == "bearer_token"` and `spec.Auth.AuthorizationURL != ""`
8. In `root.go.tmpl`, conditionally register auth subcommand

**Patterns to follow:** Existing `doctor.go.tmpl` command structure, `config.go.tmpl` for file storage

**Execution note:** This is the largest unit. Build auth.go.tmpl first, then wire up config and client changes. Test with Gmail spec.

**Test scenarios:**
- Gmail CLI generates with `auth login`, `auth status`, `auth logout` commands
- Petstore CLI (no OAuth2) does NOT generate auth commands
- Discord CLI (apiKey auth) does NOT generate auth commands
- `gmail-cli auth login --help` shows expected flags
- All 4 specs compile after template changes

**Verification:** `go test ./...` passes, regenerate all 4 specs and `go build` each. `gmail-cli auth --help` shows login/status/logout.

---

### Unit 4: Multi-Spec Composition

**Goal:** `printing-press generate --spec gmail.yaml --spec calendar.yaml --output google-cli` merges multiple API specs into one CLI with each spec as a top-level resource group.

**Files:**
- `internal/cli/generate.go` - accept multiple `--spec` flags
- `internal/spec/spec.go` - add `MergeSpecs()` function
- `internal/generator/generator.go` - minor updates for merged spec

**Approach:**
1. Change `--spec` flag from `StringVar` to `StringSliceVar` (cobra supports this)
2. Parse each spec independently
3. New `MergeSpecs(specs []*APISpec) *APISpec` function:
   - First spec's name becomes the CLI name, or use `--name` flag
   - Each spec's resources get prefixed with the spec name if there are conflicts
   - Auth configs merged (prefer OAuth2 if any spec uses it)
   - Description concatenated
4. Pass merged spec to generator

**Patterns to follow:** Existing spec parsing pipeline

**Execution note:** Depends on having multiple Google API specs in testdata. Start by duplicating/modifying Petstore to create a second test spec, or download Calendar spec.

**Test scenarios:**
- `printing-press generate --spec petstore.yaml --spec stytch.yaml --output multi-test` generates a combined CLI
- Both specs' resources appear as commands
- Single doctor command checks both
- Single `--spec` still works (backwards compatible)

**Verification:** `go test ./...` passes, manual test with two specs

---

### Unit 5: Dogfood Gmail CLI End-to-End

**Goal:** Generate Gmail CLI, verify all commands look right, ensure help output is clean.

**Files:** No press changes - this is pure verification.

**Approach:**
1. Regenerate Gmail CLI with latest press
2. Build it
3. Check every top-level resource has clean commands
4. Spot-check 5+ endpoints for clean flags (no global params, good names)
5. Check `doctor` output (should show OAuth2 auth if Unit 3 landed)
6. Document any remaining gaps as follow-up issues

**Verification:** Gmail CLI compiles, all resources have clean command names, no endpoint limit warnings, no global params on individual commands.

## Dependencies

```
Unit 1 (Color)     - independent
Unit 2 (URL)       - independent
Unit 3 (OAuth2)    - independent
Unit 4 (Multi-Spec)- independent
Unit 5 (Dogfood)   - depends on Units 1-4
```

Units 1-4 can run in parallel (they touch different files). Unit 5 is final verification.

## Execution Loop

```
for each unit:
  1. Edit templates/parser
  2. go test ./...
  3. Regenerate all 4 specs
  4. go build each generated CLI
  5. Check help output
  6. Fix what's wrong
  7. Commit when green
```

## What "Done" Looks Like

```bash
# Print a Gmail CLI from a URL
printing-press generate --spec https://gmail.googleapis.com/$discovery/rest?version=v1 --output gmail-cli

# Use it
cd gmail-cli && go build -o ./gmail ./cmd/gmail-cli
./gmail auth login                    # opens browser, OAuth2 flow
./gmail messages list me --limit 5    # table output, paginated, colored
./gmail messages list me --all        # fetches everything with progress
./gmail labels list me                # formatted table with bold headers
./gmail doctor                        # colored checkmarks
```
