---
title: "CLI Printing Press: Describe Your API, Get a Production CLI"
type: feat
status: active
date: 2026-03-23
---

# CLI Printing Press

## Describe your API. Get a production CLI.

Steinberger writes the same fetch-parse-output pattern across four languages. gogcli (Go, 6.5K stars), wacli (Go, 677 stars), oracle (TypeScript, 1.7K stars), imsg (Swift, 914 stars) - each one requires import statements, an argument parsing library, an HTTP client, a JSON parser, output formatting, error handling boilerplate, and build configuration. Measured in wacli's actual code: 79% boilerplate, 21% business logic.

Every developer building agent-accessible tools hits this. Every company wrapping an API in a CLI rewrites the same scaffolding. Stainless proved the model works at $30K/year generating CLIs from OpenAPI specs - but they sell to API providers (Stripe, OpenAI), not to the developers who want to USE those APIs from the terminal.

CLI Printing Press fills that gap. Describe what API you want to wrap - in English, or with an OpenAPI spec - and get a production Go CLI with Steinberger-quality patterns. Single binary. `--json` on every command. Auth management. Structured output. Proper exit codes. Ready to `brew install`.

## The Market Gap

| Layer | Who it serves | Examples | Gap |
|-------|--------------|---------|-----|
| Enterprise SDK generators | API providers ($30K/yr) | Stainless ($25M raised), Speakeasy, Fern (acquired by Postman) | Too expensive for individual developers |
| Abandoned OSS generators | Nobody (unmaintained) | danielgtaylor/openapi-cli-generator (195 stars) | Dead projects |
| CLI frameworks | Developers who write CLIs by hand | Cobra (43.5K), Typer (19K), Clap (16.3K) | Still requires writing all the business logic |
| **API consumer CLI generator** | **Developers who want to USE APIs** | **Nobody** | **This is the gap** |

## What It Produces

Given: "I want a CLI for the GitHub API that can list repos, create issues, and manage PRs"

CLI Printing Press generates a complete Go project:

```
github-cli/
  main.go              # Cobra root command, version, help
  cmd/
    repos_list.go      # gh-cli repos list --user steipete --limit 10
    repos_get.go       # gh-cli repos get --owner steipete --name wacli
    issues_create.go   # gh-cli issues create --repo steipete/wacli --title "Bug"
    issues_list.go     # gh-cli issues list --repo steipete/wacli --state open
    prs_list.go        # gh-cli prs list --repo steipete/wacli
    prs_merge.go       # gh-cli prs merge --repo steipete/wacli --number 42
  internal/
    client.go          # HTTP client with auth, retries, rate limiting
    output.go          # --json, --table, --csv output formatting
    auth.go            # Bearer token, OAuth, cookie auth management
    config.go          # Config file (~/.github-cli.yaml)
  go.mod
  go.sum
  Makefile             # build, test, install, brew-formula
  README.md            # Generated from API docs
```

Every generated command follows the Steinberger pattern:
- `--json` flag on every command (machine-readable for agents)
- Human-readable table output by default (for humans)
- Proper exit codes (0 success, 1 user error, 2 API error)
- `--help` with examples generated from API docs
- Auth from env var, config file, or keychain
- Timeout, retry, and rate limiting built in

## The Steinberger Patterns (Extracted from 7 Go Repos)

Analyzed every Go CLI Steinberger has built: wacli, gogcli, discrawl, sag, brabble, sonoscli, camsnap. These patterns repeat across all of them.

### Framework Choice
- **5/7 use Cobra** (wacli, brabble, sonoscli, camsnap, sag)
- **1 uses Kong** (gogcli - the most complex one, ~80 commands)
- **1 uses stdlib `flag`** (discrawl - the simplest approach)
- **Template default: Cobra** (covers 90% of use cases)

### Project Structure (universal across all repos)
```
cmd/<binary>/main.go          # Thin: ~15 lines, calls Execute() + os.Exit()
internal/
  cli/
    root.go                   # Root command, persistent flags (--json, --timeout, --config)
    <feature>.go              # One file per command group
    output.go                 # Tri-mode formatting (human/json/tsv)
    doctor.go                 # Health checks (present in 5/7 repos)
    version.go                # Version command
  config/config.go            # Config loading with defaults
  <domain>/                   # Business logic packages
.goreleaser.yaml              # GoReleaser v2 for builds
.golangci.yml                 # Linter config
.github/workflows/
  ci.yml                      # CI pipeline
  release.yml                 # Auto-release
```

### Tri-Mode Output (every CLI)
| Mode | Flag | Format | Use case |
|------|------|--------|----------|
| Human | default | Colored tables via `text/tabwriter` | Interactive terminal |
| JSON | `--json` (persistent/global) | Pretty-printed, sometimes wrapped in `{success, data, error}` envelope | Agents, piping |
| Plain/TSV | `--plain` or `--format tsv` | Tab-separated, no headers | awk/cut/grep piping |

sonoscli evolved this to `--format plain|json|tsv` with `--json` deprecated. gogcli adds `--results-only` to strip the envelope and `--select` for field selection.

### Auth Hierarchy (4 levels, template generates the right one)
| Level | Method | Example repo | When to use |
|-------|--------|-------------|-------------|
| 0 | None | brabble (local daemon) | Local-only tools |
| 1 | Env var + flag | sag (`ELEVENLABS_API_KEY`) | Simple API key auth |
| 2 | Config file token | discrawl (TOML), camsnap (YAML) | Token from multiple sources |
| 3 | QR/device auth | wacli (WhatsApp linked devices) | Messaging platforms |
| 4 | OAuth2 + keyring | gogcli (`99designs/keyring`, multi-account) | Google/Microsoft/enterprise APIs |

### Structured Exit Codes (gogcli is the gold standard)
```
0  success
1  generic error
2  usage/parse error
3  empty results
4  auth required
5  not found
6  permission denied
7  rate limited
8  retryable (timeout, server error)
10 config error
130 cancelled (SIGINT)
```

### Standard Subcommands (appear across multiple repos)
- `auth` - authentication flow (gogcli, sonoscli)
- `doctor` - health checks (5/7 repos)
- `version` - version info (all repos)
- `status` - current state (discrawl, brabble, sonoscli)
- `sync` - data synchronization (wacli, discrawl)
- `watch`/`tail` - live monitoring (sonoscli, camsnap, discrawl)

### Config Pattern
- XDG-style: `~/.config/<tool>/config.{toml,yaml}`
- Precedence: `--config` flag > env var (`<TOOL>_CONFIG`) > default path
- Formats: TOML (discrawl, brabble) or YAML (camsnap, sonoscli)
- Template default: TOML

### Build/Release
- GoReleaser v2 for all repos
- `ldflags: -s -w -X main.version={{.Version}}`
- Homebrew taps via GoReleaser auto-generation
- Universal macOS binaries (amd64+arm64)
- golangci-lint in 6/7 repos

## CLI Best Practices Checklist (from clig.dev, Atlassian, 12 Factor CLI, agent research)

The template generates CLIs that follow these standards by default. Users don't need to know the rules - the template encodes them.

### Non-Negotiable (every generated CLI has these)
- stdout for data, stderr for messages/progress/errors
- `--help`/`-h` with examples and required/optional markers
- `--version`/`-V` and `version` subcommand
- Exit codes: 0 success, 1 error, 2 usage, 3 not found, 4 auth, 5 conflict
- `--json` persistent flag on every command
- `--quiet`/`-q` for bare output (one value per line)
- `--no-color` + respect `NO_COLOR` env + `TERM=dumb` + TTY detection
- `--force`/`--yes` to skip interactive prompts
- Noun-verb subcommand hierarchy
- XDG config paths (`~/.config/<tool>/`)
- Config precedence: flags > env vars > project config > user config

### Agent-Ready (generated when `--agent-friendly` flag is used)
- JSON to stdout only, never mixed with human text
- Flat JSON objects (avoid deep nesting)
- Consistent types (ISO 8601 dates, seconds for durations)
- NDJSON (one JSON per line) for streaming output
- Structured error objects: `{"error": "code", "message": "...", "suggestion": "..."}`
- `--dry-run` for all mutating operations
- Idempotent verbs: `ensure`/`apply`/`sync` over `create`/`delete`
- Auto-detect non-TTY and output JSON by default in headless mode
- Shell completions (bash, zsh, fish)

### Delightful (generated for human-facing CLIs)
- Spinners for tasks with unknown duration (charmbracelet/huh)
- Progress bars for measurable multi-step tasks
- Suggest next command after each action
- Prompt for missing required options (when TTY)
- Sub-500ms startup time
- `doctor` subcommand for health checks

## Three Input Modes

### Mode 1: Natural Language
```
$ printing-press generate "Stripe API - manage customers, charges, and subscriptions"
```
Agent reads Stripe's API docs, generates OpenAPI spec, produces Go CLI.

### Mode 2: OpenAPI Spec
```
$ printing-press generate --spec stripe-openapi.yaml --name stripe-cli
```
Direct from spec. Most precise. Supports OpenAPI 3.0+.

### Mode 3: Example Requests
```
$ printing-press generate --from-curl examples.sh --name myapi-cli
```
Feed it a collection of curl commands. It infers the API structure and generates a CLI.

## Architecture: A Claude Code Plugin

CLI Printing Press is a Claude Code plugin - the same model as Compound Engineering (EveryInc/compound-engineering-plugin, the most popular CE plugin). The plugin contains:

```
cli-printing-press/
  .claude-plugin/
    plugin.json          # Plugin manifest (name, version, repo)
  CLAUDE.md              # Plugin conventions
  skills/
    printing-press/
      SKILL.md           # /printing-press <description> - generates a CLI
    printing-press-catalog/
      SKILL.md           # /printing-press-catalog - browse pre-built CLIs
  templates/
    go-cobra/            # Go + Cobra template (Steinberger patterns)
      cmd.go.tmpl
      client.go.tmpl
      output.go.tmpl
      auth.go.tmpl
      config.go.tmpl
      Makefile.tmpl
      go.mod.tmpl
    README.md
  catalog/               # Pre-built CLI definitions (the "library")
    github.yaml          # GitHub API CLI definition
    stripe.yaml          # Stripe API CLI definition
    slack.yaml           # Slack API CLI definition
    ... (community-contributed)
  scripts/
    generate.sh          # Template engine
    validate.sh          # Security validation for submissions
```

**Installation:**
```
/plugin marketplace add mvanhorn/cli-printing-press
/plugin install printing-press@cli-printing-press
```

**Usage:**
```
/printing-press "Stripe API - manage customers, charges, and subscriptions"
/printing-press --spec stripe-openapi.yaml --name stripe-cli
/printing-press-catalog                    # browse pre-built CLIs
/printing-press-catalog install github     # install pre-built GitHub CLI
```

## The Community Catalog (and the Security Question)

The catalog is where this gets interesting - and dangerous.

### The Vision

Anyone can contribute a CLI definition to the catalog. You run `/printing-press "Twilio API"`, it generates a great CLI, you submit a PR to add `catalog/twilio.yaml` to the repo. Next person who wants a Twilio CLI just runs `/printing-press-catalog install twilio` instead of generating from scratch.

Like Homebrew formulae. Like ClawHub skills. But for CLI tool definitions.

### The Security Model: Curated, Not Open

The key decision: **catalog entries are curated by maintainers, not auto-published.** Here's why:

**The injection risk is real.** A CLI definition includes:
- API endpoints (could point to a malicious proxy instead of the real API)
- Auth configuration (could exfiltrate tokens to an attacker's server)
- Shell commands in Makefiles (could run arbitrary code during build)
- Generated Go code patterns (could include backdoors in the templates)

A community member submitting `catalog/aws.yaml` that points auth to `auth.totally-not-evil.com` would compromise every user who installs it.

**Why NOT ClawHub/open marketplace:**
- ClawHub has 13.7K skills with VirusTotal scanning, but scanning generated Go source for subtle backdoors (a token exfiltration hidden in error handling) is much harder than scanning for malware binaries
- The attack surface is the API definition, not the binary - a malicious definition produces a "clean" binary that does bad things by design
- Trust needs to be at the definition level, not the binary level

**Why NOT fully open PRs:**
- Even with review, subtle endpoint substitution (`api.stripe.com` vs `api.str1pe.com`) is easy to miss in a YAML file
- The reviewing burden grows linearly with submissions

**The model: three tiers.**

| Tier | Who | What | Trust |
|------|-----|------|-------|
| **Official** | Printing Press maintainers | Top 20 APIs (GitHub, Stripe, Slack, etc.) | Verified against official API docs. Signed. |
| **Verified** | Community PRs, maintainer-reviewed | Any API where the submitter demonstrates it works against the real API | Reviewed for endpoint authenticity. Signed after merge. |
| **Unverified** | Community PRs, auto-merged if CI passes | Any API definition that passes validation | NOT reviewed for authenticity. Flagged with warning. User accepts risk. |

**Security validation (automated, runs in CI on every PR):**
1. All API endpoints must use HTTPS
2. All endpoints must resolve to IPs that match the API provider's known ranges (where available)
3. No shell commands in the definition (Makefile is generated from template, not from definition)
4. Auth endpoints must match the API provider's documented auth URLs
5. Generated Go code is `go vet` clean
6. No `replace` directives in go.mod (prevents dependency hijacking)
7. Binary is built and smoke-tested against the real API (with a read-only test account where possible)

**What the user sees:**
```
$ /printing-press-catalog install stripe
Installing: stripe (Official - verified by CLI Printing Press maintainers)
Source: catalog/stripe.yaml (signed: abc123)

$ /printing-press-catalog install obscure-api
Installing: obscure-api (Unverified - community-contributed, not reviewed)
WARNING: This CLI definition has not been verified against the real API.
         Endpoints and auth configuration have not been audited.
         Proceed? [y/N]
```

### The PR Submission Flow

```
1. User runs:  /printing-press "Twilio API"
2. Agent generates the CLI project + catalog/twilio.yaml definition
3. User tests it, confirms it works
4. User runs:  /printing-press submit twilio
5. Agent:
   - Forks mvanhorn/cli-printing-press
   - Adds catalog/twilio.yaml
   - Runs validation suite locally
   - Opens PR with:
     - The YAML definition
     - Test evidence (screenshots of CLI output against real API)
     - Attestation: "I tested this against the real Twilio API"
6. CI runs security validation
7. Maintainer reviews (for Verified tier) or auto-merges (for Unverified tier)
```

## What APIs Look Like (Patterns from Steinberger's 10 CLIs)

Analysis of the actual APIs behind all 10 Steinberger CLIs reveals three tiers:

**Tier 1: Cloud API wrappers (CLI Printing Press sweet spot)**
- sag (ElevenLabs), oracle (multi-LLM), summarize (content+LLM), gogcli (Google Suite)
- REST/JSON, API keys or OAuth2, cursor pagination
- This is what the template generator optimizes for

**Tier 2: Platform API wrappers (possible with client libraries)**
- discrawl (Discord REST + WebSocket Gateway), wacli (WhatsApp via whatsmeow)
- The CLI wraps a Go client library, not raw HTTP
- Template can scaffold the Cobra structure; domain logic is manual

**Tier 3: Local network/device tools (not templateable)**
- sonoscli (UPnP/SSDP), camsnap (RTSP), brabble (local whisper.cpp)
- No cloud API. Discovery-based, protocol-specific. Template adds no value.

### What the Template MUST Support

| Pattern | Appears in | Implementation |
|---------|-----------|---------------|
| API key auth (env var + flag + file) | 6/10 CLIs | Triple priority chain with provenance tracking |
| REST/JSON | 7/10 CLIs | net/http client with configurable base URL |
| `--json` output | 10/10 CLIs | Tri-mode: human/json/plain via type-switch dispatch |
| Context-aware timeouts | 10/10 Go CLIs | `context.WithTimeout` + signal handling |
| Config file | 8/10 CLIs | TOML (default) at `~/.config/<tool>/config.toml` |
| Cursor pagination | 4/10 CLIs | Generic `nextPageToken` / `before`/`after` patterns |

### What the Template SHOULD Support (optional modules)

| Pattern | When needed | Implementation |
|---------|------------|---------------|
| OAuth2 flow | Google, Microsoft, enterprise APIs | `golang.org/x/oauth2` + keyring storage |
| Streaming (SSE/chunked) | LLM APIs, real-time data | `bufio.Scanner` on response body |
| Retry with backoff | Rate-limited APIs | Exponential backoff with jitter |
| Local SQLite cache | Tools that sync data locally | `modernc.org/sqlite` (pure Go, no CGO) |
| WebSocket tailing | Real-time event APIs | `gorilla/websocket` + worker pool |

## First Test Cases (Verified: Have APIs, No Official CLIs)

### Tier 1: First 3 to build (simplest, cleanest specs)

| Product | Category | OpenAPI Spec | ~Commands | Why first |
|---------|----------|-------------|----------|-----------|
| **Stytch** | Auth | github.com/stytchauth/stytch-openapi | ~15 | Smallest surface. Clean spec. Perfect validation. |
| **Clerk** | Auth | clerk.com/docs/reference | ~15-20 | Popular. On their CLI roadmap. Clean CRUD. |
| **Loops** | Email | loops.so/docs | ~8-10 | Tiny API. Contacts + events + emails. Fastest to validate. |

### Tier 2: Showcase candidates (prove it scales)

| Product | Category | OpenAPI Spec | ~Commands | Why |
|---------|----------|-------------|----------|-----|
| **ClickUp** | PM | developer.clickup.com/docs/open-api-spec | ~25-30 | 5+ community CLIs prove demand. No official. Showcase. |
| **Intercom** | Support | github.com/intercom/Intercom-OpenAPI | ~20-25 | Major platform. Zero CLI presence. |
| **Front** | Messaging | github.com/frontapp/front-api-specs | ~20 | Clean REST. Official specs. |
| **Shortcut** | PM | developer.shortcut.com/api/rest/v3 | ~15-20 | Simpler than ClickUp. Good mid-range test. |

### Tier 3: Large-scale validation

| Product | Category | OpenAPI Spec | ~Commands | Why |
|---------|----------|-------------|----------|-----|
| **Asana** | PM | github.com/Asana/openapi | ~30-35 | Major tool, official spec, many community CLIs |
| **Square** | Payments | github.com/square/connect-api-specification | ~40+ | Tests complex API with many resources |
| **Discord** | Community | github.com/discord/discord-api-spec | ~50+ | Massive. Tests generator at scale. |

### Eliminated (already have official CLIs)
Stripe, Vercel, Netlify, Railway, Render, Fly.io, Supabase, PlanetScale, Neon, Turso, Cloudflare, Fastly, Sentry, HubSpot, Auth0, Slack, Resend, Replicate, Modal, Datadog, W&B, PostHog, Grafana (34 products verified).

## Architectural Learnings from Production Agent Fleet Management

Patterns extracted from a mature 14-skill autonomous agent system (15,000+ lines, 179 calibrated predictions, overnight mode submitting 21 PRs while sleeping). Applied to CLI Printing Press:

### Quality Gates (adapt for CLI generation)

The agent fleet uses 6 mandatory gates before any external action. For CLI Printing Press:

1. **Spec validation** - parse OpenAPI spec, flag incomplete endpoints, missing auth docs
2. **Generation validation** - `go vet` + `go build` on generated code
3. **Wiring verification** - grep for call sites of every generated Cobra command (no orphan commands)
4. **Test validation** - `go test` passes on generated test stubs
5. **Output verification** - each generated command produces valid JSON with `--json`
6. **Doctor validation** - `<tool> doctor` runs successfully

Gate tracker file pattern: accumulate passes in `/tmp/printing-press-gates-$$.txt`, block if any missing.

### Confidence Scoring (adapt for generation quality)

Track per-API-spec generation success:
- Did it compile on first try? (binary: yes/no)
- How many manual fixes needed? (0 = perfect, 1-3 = good, 4+ = bad template)
- Which template modules failed? (auth? pagination? output formatting?)
- Calibrate: adjust template weights based on outcomes

### Registry Pattern

Central JSON registry at `~/.printing-press/registry.json`:
```json
{
  "generated_tools": [
    {
      "name": "stytch-cli",
      "api": "stytch",
      "spec_source": "github.com/stytchauth/stytch-openapi",
      "generated_at": "2026-03-23T...",
      "status": "compiled|tested|published",
      "commands": 15,
      "quality_score": 0.95,
      "last_regenerated": "2026-03-23T..."
    }
  ]
}
```

With `.backup` always alongside.

### Skip Cache

`~/.printing-press/skip-specs.json` for known-bad API specs:
```json
{
  "broken_spec": ["some-api/v1 - missing auth schema"],
  "rate_limited": ["api-that-blocks-generators"],
  "deprecated": ["old-api/v2 - sunset 2026-01"]
}
```

Checked BEFORE any generation attempt.

### Anti-Slop Patterns for Generated Docs

Generated READMEs must not contain: "leverages", "utilizes", "comprehensive", "elegant solution", "not just X, it's Y", em dashes for dramatic effect. Every claim must reference a specific command or flag.

## Implementation

### Phase 1: Template Engine (Weeks 1-2)

Build the Go project generator from a structured API description (not OpenAPI yet - a simpler internal format).

**Deliverables:**
- Go template system that produces Cobra CLI projects
- `client.go` template with auth, retries, rate limiting
- `output.go` template with --json/--table/--csv
- Command template that generates one command file per endpoint
- `Makefile` with build, test, install, brew-formula targets
- 3 example CLIs generated: GitHub API, Stripe API, Slack API

**Gate:** Generated CLIs compile, run, and produce correct output against real APIs.

### Phase 2: OpenAPI Parser (Weeks 3-4)

Parse OpenAPI 3.0+ specs and map them to the internal format.

**Deliverables:**
- OpenAPI 3.0 parser (use kin-openapi or libopenapi in Go)
- Mapping: OpenAPI paths -> Cobra subcommands, parameters -> flags, schemas -> Go structs
- Auth scheme detection (bearer, OAuth2, API key) -> appropriate auth template
- 5 real-world OpenAPI specs tested: GitHub, Stripe, Twilio, Slack, OpenAI

**Gate:** `printing-press generate --spec github-openapi.yaml` produces a working CLI.

### Phase 3: Natural Language Mode (Weeks 5-6)

Agent reads API docs and generates the internal API description.

**Deliverables:**
- Claude Code skill: `/printing-press <description>`
- Agent fetches API documentation, extracts endpoints, generates internal spec
- Fallback: if docs are unclear, generate a skeleton and ask the user to fill in details
- OpenClaw skill published on ClawHub

**Gate:** `/printing-press "Stripe API - customers and charges"` produces a working CLI from just that sentence.

### Phase 4: Catalog and Community (Weeks 7-10)

**Deliverables:**
- Claude Code plugin published: `/plugin marketplace add mvanhorn/cli-printing-press`
- 10 Official catalog entries: GitHub, Stripe, Slack, Twilio, OpenAI, Linear, Vercel, Cloudflare, Supabase, Discord
- `/printing-press-catalog` browser and installer skill
- `/printing-press submit` for community contributions
- Security validation CI pipeline (endpoint verification, HTTPS enforcement, dependency scanning)
- Three-tier trust model (Official / Verified / Unverified) with clear user-facing warnings
- GitHub Action for downstream projects (regenerate CLI when OpenAPI spec changes)
- README with benchmarks and contribution guide

## Benchmark Targets

| Metric | Hand-written Go + Cobra | CLI Printing Press |
|--------|------------------------|-------------------|
| Time to working CLI | Hours to days | Minutes |
| Lines of code written | 200-2000 (depending on API) | 0 (generated) |
| Boilerplate ratio | 79% (measured in wacli) | 0% (all generated) |
| Agent token cost | 500-2000 tokens to write | ~50 tokens to invoke |
| Binary quality | Production (if you're good) | Production (patterns are codified) |
| Auth management | Manual | Built-in |
| Output formatting | Manual | Built-in |

## Why Not Just Ask Claude to Write the Go Code?

You can. And agents do. But:

1. **Inconsistency.** Every generation produces slightly different patterns. CLI Printing Press generates from templates - same patterns every time.
2. **The 79% problem.** Even when agents write correct Go, 79% is boilerplate. CLI Printing Press eliminates it at the template level, not the generation level.
3. **Maintenance.** When the Cobra best practices change, you update the template once. Every generated CLI gets the improvement.
4. **Auth is hard.** OAuth2 flows, token refresh, keychain integration, multi-account support - these are complex and error-prone. Codified once in the template, correct forever.
5. **Discoverability.** Pre-generated CLIs for popular APIs are immediately useful. No agent invocation needed - just `brew install stripe-printing-press`.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| OpenAPI specs are often incomplete or wrong | Generated CLIs miss endpoints or have wrong types | Validate against real API calls. Flag gaps for manual review. |
| api2cli already exists (Node.js, new) | Direct competitor | CLI Printing Press generates Go (single binary), not Node.js. Different target audience. |
| LLMs get good enough that templates aren't needed | Product becomes redundant | Templates provide consistency and maintenance benefits that raw generation can't. If LLMs do make templates unnecessary, the pre-generated CLI catalog still has value. |
| Generated CLIs feel generic | Users don't adopt over hand-written | Make templates customizable. Allow post-generation customization. The goal is 90% done, not 100%. |
| Malicious catalog submissions | Token exfiltration, endpoint hijacking, supply chain attacks | Three-tier trust model. Automated endpoint verification. No shell in definitions. Signed Official entries. User warnings on Unverified entries. |
| Catalog grows faster than review capacity | Unreviewed entries accumulate, trust erodes | Auto-merge Unverified tier with clear warnings. Focus review effort on Verified tier. Community can flag suspicious entries. |

## Relationship to Nullhuman

CLI Printing Press is the practical product. It ships in weeks, generates Go that agents already write well, and fills a market gap between abandoned OSS and $30K/year enterprise.

If Nullhuman (the language) ever ships, CLI Printing Press could generate Nullhuman code instead of Go. But CLI Printing Press doesn't depend on Nullhuman. It's independently valuable.

The two projects share research:
- Steinberger's CLI patterns (codified from wacli/gogcli source code)
- The "CLI > MCP" thesis (32x fewer tokens, 100% vs 72% reliability)
- The four-layer agent stack (Markdown for instructions, typed tools for execution)
- Capability-based security (could be added as a runtime wrapper around generated Go binaries)

## Sources

- steipete/wacli (Go, 677 stars) - WhatsApp CLI, 79% boilerplate measured
- steipete/gogcli (Go, 6.5K stars) - Google Suite CLI
- Stainless ($25M raised, ~$1M ARR) - Enterprise SDK/CLI generation
- Speakeasy - Freemium SDK/CLI generation
- danielgtaylor/openapi-cli-generator (195 stars) - Abandoned OSS
- api2cli (new, Node.js) - "Turn any REST API into an agent-ready CLI"
- ClawHub (13.7K skills) - Distribution channel
- "CLI > MCP" benchmarks: 32x fewer tokens, 100% vs 72% reliability (rentierdigital.xyz, scalekit.com)

### Origin

Split from: docs/plans/2026-03-23-feat-nullhuman-agent-programming-language-plan.md
