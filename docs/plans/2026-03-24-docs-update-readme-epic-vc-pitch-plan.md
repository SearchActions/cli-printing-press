---
title: "Update README - Epic VC-Pitch Style, Steinberger-Inspired"
type: docs
status: completed
date: 2026-03-24
---

# Update README - Epic VC-Pitch Style, Steinberger-Inspired

## Overview

The current README is 104 lines and reflects maybe 30% of what the printing press actually does. The project has shipped OAuth2, pagination, table output, retry logic, color/TTY, multi-spec composition, URL specs with caching, autonomous 6-phase pipeline, a 12-entry catalog, and a Claude Code plugin. None of that is in the README.

The new README should read like a VC pitch deck meets a technical masterpiece. The core narrative: **"Automagically make Peter Steinberger-quality CLIs for ANY API in the world."** One command. Any API. Production-quality CLI out the other side.

Inspired by gogcli's README (1,200 lines, code-heavy, every command documented, no fluff) but with the added magic of "this is a printing press that PRINTS those CLIs for you."

## Acceptance Criteria

- [ ] README conveys the vision in the first 3 lines: describe any API, get a finished CLI
- [ ] Epic opening section with the "why this matters" narrative (Steinberger builds one CLI per API by hand - the press prints them automatically)
- [ ] Shows real command output (not hypothetical) for at least 3 examples
- [ ] Documents ALL shipped features (see Feature Inventory below)
- [ ] Covers all 3 input modes: natural language, spec file, URL
- [ ] Documents the pipeline mode (`print` command) as the hero feature
- [ ] Catalog table with all 12 entries
- [ ] "What Gets Generated" section that makes people say "wait, it does ALL that?"
- [ ] Claude Code plugin installation and usage
- [ ] Contributing section for catalog entries
- [ ] Development section
- [ ] No badges wall, no fluff - code-heavy like gogcli
- [ ] MIT license

## Feature Inventory (Everything the README Must Cover)

### Core Generator
- Cobra subcommand hierarchy (`<name>-cli <resource> <action>`)
- Sub-resource grouping for nested API paths
- `--json`, `--plain`, `--quiet` output modes
- Auth management: API key, bearer token, basic auth, OAuth2 (authorization_url, token_url, browser flow)
- Config file at `~/.config/<name>-cli/config.toml`
- `doctor` command with formatted health dashboard
- `version` command
- Auto-generated usage examples in command help
- Auto-detect pagination: `--limit`, `--all` flags
- Table output for array responses
- Retry logic with exponential backoff + rate limit detection (429 + Retry-After)
- Structured exit codes (0=ok, 1=general, 3=not-found, 4=auth, 7=rate-limit, 130=interrupt)
- Dry-run support (`--dry-run` for mutation commands)
- Color + TTY detection (respects `NO_COLOR`, `TERM=dumb`)
- `--no-color` flag
- Makefile, .goreleaser.yaml, .golangci.yml, README.md auto-generated
- 7 quality gates (go mod tidy, go vet, go build, binary build, --help, version, doctor)

### Input Formats
- OpenAPI 3.0+ (YAML or JSON)
- Swagger 2.0 (YAML or JSON)
- Internal YAML format (simpler, hand-written)
- Auto-detection: reads first 500 bytes, routes to correct parser
- URL input with local caching (24h TTL at `~/.cache/printing-press/specs/`)
- Multi-spec composition (`--spec` repeatable, merged into one CLI)

### Pipeline Mode (Hero Feature)
- `printing-press print <api-name>` creates autonomous 6-phase pipeline
- Phases: preflight, scaffold, enrich, regenerate, review, ship
- Plan seeds auto-generated per phase (Compound Engineering compatible)
- Spec overlay system (enrich phase produces overlay.yaml, regenerate merges it)
- State tracking with state.json (pending/planned/executing/completed/failed)
- Budget gate (3h max), heartbeat safety net via cron
- Morning report generation
- Session chaining with fresh context per phase (nightnight-style)
- 16 known APIs discoverable by name (apis-guru fallback for unknown)

### Catalog
- 12 pre-built entries: Stripe, Square, Stytch, Discord, Twilio, SendGrid, GitHub, DigitalOcean, Asana, HubSpot, Front, Petstore
- Schema validation (CI validates catalog PRs)
- Categories: auth, payments, email, developer-tools, project-management, communication, crm, example
- Tiers: official, community

### Distribution
- Claude Code plugin (`/plugin marketplace add mvanhorn/cli-printing-press`)
- Two skills: `/printing-press` (generator) and `/printing-press-catalog` (browse/install)
- Natural language input ("Stytch authentication API" - finds spec, generates CLI)

## Implementation Units

### Unit 1: Hero Section + Vision

**Files:** `README.md`

**Approach:**
Write the opening that makes someone stop scrolling. Structure:

```
# CLI Printing Press

One command. Any API. A production CLI walks out the other side.

[2-3 sentence pitch: Peter Steinberger hand-crafts gogcli, wacli, discrawl -
beautiful CLIs that take weeks each. The printing press does it in 60 seconds.
Point it at any OpenAPI spec in the world and get a finished CLI with auth,
pagination, tables, retries, color, doctor, completions. Not a scaffold.
A finished product.]

## The 60-Second Demo

[3 bash blocks showing the magic:
1. printing-press generate --spec <petstore-url> --output ./petstore-cli
2. cd petstore-cli && go build && ./petstore-cli --help (show output)
3. ./petstore-cli pet find-by-status --status available --limit 5 (show table)]
```

The demo should use REAL output captured from actual runs (we validated petstore E2E tonight).

**Verification:** First 30 lines make someone want to star the repo.

### Unit 2: What Gets Generated (The "Wait, It Does ALL That?" Section)

**Files:** `README.md`

**Approach:**
Expand the current 7-bullet list into a comprehensive feature showcase. Group by category:

```
## What Gets Generated

Every printed CLI ships with:

### Commands
- Cobra subcommand tree: `<name>-cli <resource> <action>`
- Sub-resources: nested paths become `resource sub-resource action`
- doctor: formatted health dashboard with checkmarks and color
- version: semantic version from config
- Shell completions: bash, zsh, fish, powershell

### Output
- Table output for list responses (auto-detected from array schemas)
- `--json` raw JSON, `--plain` tab-separated, `--quiet` bare values
- Color and TTY detection (respects NO_COLOR, TERM=dumb, --no-color)

### Auth
- API key (header or query)
- Bearer token
- Basic auth
- OAuth2 with browser-based authorization flow and token refresh

### Reliability
- Retry with exponential backoff on 5xx
- Rate limit detection (429 + Retry-After header)
- Structured exit codes (0/1/3/4/7/130)
- Request timeout with --timeout flag
- --dry-run for mutation commands

### Pagination
- Auto-detected from spec: --limit, --all flags
- Cursor-based and offset-based pagination

### Build Tooling
- go.mod, Makefile, .goreleaser.yaml, .golangci.yml
- Auto-generated README.md for the CLI itself
- Config at ~/.config/<name>-cli/config.toml
```

### Unit 3: Input Formats + Usage Modes

**Files:** `README.md`

**Approach:**
Show all 3 input paths with real examples:

```
## Usage

### From OpenAPI Spec (URL)
./printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json

### From Local Spec File
./printing-press generate --spec ./my-api-spec.yaml

### Multi-Spec Composition
./printing-press generate \
  --spec https://api1.example.com/openapi.json \
  --spec https://api2.example.com/openapi.json \
  --name combined-cli

### Supported Formats
- OpenAPI 3.0+ (YAML/JSON)
- Swagger 2.0 (YAML/JSON)
- Internal YAML format (hand-written, simpler)
- Auto-detection: reads first 500 bytes, routes to correct parser
```

### Unit 4: Pipeline Mode (The Hero Feature)

**Files:** `README.md`

**Approach:**
This is the differentiator. Frame it as magic:

```
## Autonomous Pipeline (The Magic)

For complex APIs, the press doesn't just scaffold - it thinks.

./printing-press print gmail

This kicks off a 6-phase autonomous pipeline:

| Phase | What Happens |
|-------|-------------|
| 0. Preflight | Verify Go, download spec, cache conventions |
| 1. Scaffold | Generate CLI, pass 7 quality gates |
| 2. Enrich | Research API docs, discover missing endpoints, auth flows |
| 3. Regenerate | Merge enrichments back, regenerate with full knowledge |
| 4. Review | Lint, test, benchmark, fix issues |
| 5. Ship | Build, tag, generate release notes |

Each phase runs in a fresh AI session with its own context window.
Sessions chain automatically via cron. Budget gate stops at 3 hours.
Morning report tells you what happened overnight.

16 APIs are pre-registered and discoverable by name:
[list of known specs]
```

### Unit 5: Claude Code Plugin + Catalog

**Files:** `README.md`

**Approach:**
```
## Claude Code Plugin

/plugin marketplace add mvanhorn/cli-printing-press

Then just say what you want:

/printing-press Stripe payments API
/printing-press Discord bot API
/printing-press --spec ./openapi.yaml

## Catalog

12 pre-built CLI definitions. Generate instantly:

/printing-press-catalog install stripe

[table of 12 entries]
```

### Unit 6: Quality Gates + Contributing + Development

**Files:** `README.md`

**Approach:**
```
## Quality Gates

Every generated CLI must pass 7 gates before the press considers it done:

1. go mod tidy (clean dependencies)
2. go vet ./... (static analysis)
3. go build ./... (compilation)
4. Binary build (produces runnable binary)
5. --help (usage renders without crash)
6. version (prints version string)
7. doctor (health check runs)

If any gate fails, the press retries. If it still fails, it tells you exactly what broke.

## Contributing Catalog Entries
[instructions]

## Development
[build, test, generate commands]
```

### Unit 7: Capture Real Output for Examples

**Files:** `README.md`

**Execution note:** Run before writing the final README. Capture actual terminal output from:

**Approach:**
```bash
# 1. Generate petstore and capture --help output
./printing-press generate --spec "https://petstore3.swagger.io/api/v3/openapi.json" --output /tmp/readme-demo
cd /tmp/readme-demo && go build -o petstore-cli ./cmd/petstore-cli
./petstore-cli --help
./petstore-cli pet --help
./petstore-cli doctor
./petstore-cli version

# 2. Capture the generate command output (with quality gates)
# Already captured in E2E test - 7 PASS lines

# 3. Capture print command output
./printing-press print petstore --output /tmp/readme-pipeline-demo
```

Paste real output into the README examples. No fake output.

**Verification:** Every code block in the README contains output that was actually produced by the tool.

## Scope Boundaries

- Don't modify Go code (this is README only)
- Don't add GitHub Actions badges or shields.io badges (keep it clean like gogcli)
- Don't add a table of contents (let the content speak)
- Don't add screenshots or GIFs yet (text-first, code-heavy like gogcli)
- Don't fix the version mismatch (root.go=0.1.0 vs plugin=0.4.0) - that's a separate fix
- Don't document unreleased features from the Steinberger Parity plan

## Tone Guide

- **Epic but technical.** Not "amazing revolutionary tool" - more "one command, any API, production CLI."
- **Show, don't tell.** Every claim backed by a code example or output snippet.
- **Steinberger's shadow.** The implicit message: gogcli took weeks to build. This prints gogcli-quality CLIs in 60 seconds.
- **Code-heavy.** 50-60% of the README should be fenced code blocks with real output.
- **No em dashes.** Use hyphens instead (user preference).
- **No badges wall.** Clean header like gogcli.

## Sources

- Current README: `README.md` (104 lines, needs full rewrite)
- SKILL.md: `skills/printing-press/SKILL.md` (322 lines, comprehensive feature reference)
- gogcli README: github.com/steipete/gogcli (1,200 lines, style inspiration)
- Steinberger Parity plan: `docs/plans/2026-03-24-feat-full-steinberger-parity-plan.md` (quality bar reference)
- GOAT plan: `docs/plans/2026-03-24-feat-goat-plan-gogcli-parity-and-press-upgrades-plan.md` (feature inventory)
- E2E petstore test: validated tonight, all 7 gates pass, CLI compiles and runs
- Pipeline state types: `internal/pipeline/state.go` (6 phases, status machine)
- Known specs: `skills/printing-press/references/known-specs.md` (16 APIs)
- Catalog entries: `catalog/*.yaml` (12 entries)
