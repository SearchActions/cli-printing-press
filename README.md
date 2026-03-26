# CLI Printing Press

Give it an API name. Get back the best CLI that has ever existed for it.

```bash
/printing-press Discord
/printing-press Stripe
/printing-press Linear
```

One command. 8 phases. ~1 hour. Produces a production-ready Go CLI binary + 7 deep analysis documents. REST or GraphQL - it figures it out.

## Why CLIs Matter Now

CLIs are the native interface for AI agents. When an agent needs to interact with an API, it has two choices: import an SDK and write 15 lines of code (~500 tokens), or run a CLI command (~50 tokens). The CLI wins on every dimension that matters - deterministic, structured JSON output, composable with pipes, zero dependency management.

Every API that gets a CLI becomes instantly accessible to every agent framework - Claude Code, Codex, Gemini CLI, open source agents. No SDK integration. No code generation. Just `--json --select` and pipe.

The printing press is an agent-infrastructure factory. One hour per API, and every agent in the world can use it.

## How It Works

The press runs 8 mandatory phases. Each phase writes a comprehensive plan document. The artifacts are the product - the CLI is a side effect.

```
Phase 0    Visionary Research       (3-5 min)    Who uses this API? What do they build?
Phase 0.5  Power User Workflows     (2-3 min)    What compound commands would power users want?
Phase 0.7  Prediction Engine        (15-25 min)  What local data layer would make this a 500-star tool?
Phase 1    Deep Research            (5-8 min)    What competitors exist? Why should this CLI exist?
Phase 2    Generate                 (1-2 min)    Produce the Go CLI from the API spec
Phase 3    Steinberger Audit        (5-8 min)    Score against the quality bar, plan improvements
Phase 4    GOAT Build               (5-10 min)   Build data layer + workflow commands + fixes
Phase 4.5  Dogfood Emulation        (10-20 min)  Test every command against spec-derived mocks
Phase 5    Final Steinberger        (2-3 min)    Before/after scoring delta + final report
```

### Phase 0: Visionary Research

Before generating anything, the press understands the API's domain, users, and ecosystem. It searches GitHub, Reddit, Hacker News, and Stack Overflow for usage patterns, pain points, and existing tools. It classifies every tool it finds - API wrappers, data tools, workflow tools, environment tools - and identifies what's missing.

### Phase 0.5: Power User Workflows

Predicts what compound commands a power user would want - the features that make a tool worth starring. "Find stale channels." "Audit who banned whom." "Preview a member prune without executing." These workflow commands combine 2+ API calls into single operations.

### Phase 0.7: Prediction Engine

This is the differentiator. The press predicts what a developer would build on TOP of the API - without looking at competitors. It classifies every entity (accumulating, reference, append-only, ephemeral), scores them on data gravity (volume, query frequency, join demand, search need, temporal value), and produces a concrete data layer specification: domain-specific SQLite schema, incremental sync strategy with validated cursors, FTS5 full-text search with domain filters, and compound queries.

This is how you build discrawl (12 commands, 540 stars) without knowing discrawl exists.

### Phase 1: Deep Research

Finds the OpenAPI or GraphQL spec, analyzes the top 2 competitors in depth (READMEs, star counts, open issues, user complaints), and answers: "Why should this CLI exist when [competitor] already has [N] stars?"

### Phase 2: Generate

Runs the Go generator against the spec. Produces hundreds of commands with full agent-native features. For GraphQL APIs, produces scaffolding + a GraphQL client wrapper - commands get hand-written in Phase 4.

### Phase 3: Steinberger Audit

Scores the generated CLI against the Steinberger bar (11 dimensions, 110 max). Peter Steinberger's gogcli is the 10/10 reference. Each dimension gets: current score, what 10/10 looks like, and specific changes to get there.

### Phase 4: GOAT Build

Builds the product. Priority 0 is the data layer (domain SQLite tables, sync, search, sql command). Priority 1 is workflow commands powered by the local database. Priority 2 is scorecard gap fixes. Priority 3 is polish.

### Phase 4.5: Dogfood Emulation

Tests every command against spec-derived mock responses - no API keys needed. Scores each command on 5 dimensions (request construction, response parsing, schema fidelity, example quality, workflow integrity). Auto-fixes issues, re-scores, and writes a report: "Here's what I learned. Here's what to fix. Here's what to make."

Inspired by [Vercel's emulate](https://github.com/vercel-labs/emulate) - production-fidelity API simulation, zero config.

### Phase 5: Final Steinberger

Before/after scoring delta. The proof of work.

## What Gets Generated

Every CLI ships with:

| Feature | Flag | What It Does |
|---------|------|-------------|
| JSON output | `--json` | Machine-readable, pipeable to jq |
| Field filtering | `--select id,name` | Only the fields you want |
| Dry run | `--dry-run` | Shows the exact API request without sending |
| Stdin input | `--stdin` | Pipe JSON body: `echo '{}' \| mycli create --stdin` |
| Response cache | `--no-cache` | GET responses cached 5 min, bypass with flag |
| Skip confirmation | `--yes` | For agents and scripts on destructive actions |
| Plain output | `--plain` | Tab-separated for awk/cut |
| CSV output | `--csv` | For spreadsheets and data tools |
| Quiet mode | `--quiet` | Bare output, one value per line |
| Color control | `--no-color` | Respects NO_COLOR, TERM=dumb |

Plus: doctor health check, TOML config, OAuth2 browser flow, retry with backoff, rate limit detection (exit code 7), typed exit codes with actionable hints, idempotent creates/deletes, progress events as NDJSON to stderr.

**Data layer** (when prediction engine identifies high-gravity entities):
- Domain-specific SQLite tables with proper columns (not JSON blobs)
- FTS5 full-text search with domain filters (`--channel`, `--author`, `--team`)
- Incremental sync with validated cursors
- `sql` command for raw read-only queries
- `sync`, `search`, `tail`, `export`, `analytics` commands

**Exit codes:** `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## 7 Plan Artifacts Per Run

Every run produces 7 comprehensive analysis documents in `docs/plans/`:

```
Phase 0   -> <api>-cli-visionary-research.md      API identity, usage patterns, tool landscape
Phase 0.5 -> <api>-cli-power-user-workflows.md    Workflow ideas, scoring, top 7 selected
Phase 0.7 -> <api>-cli-data-layer-spec.md         SQLite schema, sync strategy, search filters
Phase 1   -> <api>-cli-research.md                Competitors, strategic justification
Phase 3   -> <api>-cli-audit.md                   Steinberger scores, improvement plan
Phase 4   -> <api>-cli-goat-build-log.md          What was built, what was fixed
Phase 4.5 -> <api>-cli-dogfood-report.md          Per-command scores, hallucination detection
```

Each artifact chains into the next. The prediction engine's data layer spec informs Phase 4's implementation. The dogfood report's fix recommendations feed back into the code.

## Works With Any API

**REST APIs** (OpenAPI/Swagger): Full pipeline - generator produces commands, prediction engine adds data layer, dogfood validates everything.

**GraphQL APIs** (Linear, Shopify, GitHub GraphQL): Warns about generator limitations, produces scaffolding + GraphQL client wrapper, hand-writes commands in Phase 4. All research, prediction, data layer, and dogfood phases run normally.

**No spec available**: Reads the API docs, writes a spec, generates from it.

The press dynamically detects API type from the spec content - not a hardcoded list.

## The Steinberger Bar

11 dimensions, 110 points max. Grade A = 80%+.

| Dimension | What 10/10 Looks Like |
|-----------|----------------------|
| Output Modes | --json, --csv, --plain, --select, --quiet, --template |
| Auth | OAuth browser flow, token storage, multiple profiles, doctor validates |
| Error Handling | Typed exits, retry with backoff, "did you mean?" suggestions |
| Terminal UX | Progress spinners, color themes, pager for long output |
| README | Install, quickstart, every command with example, cookbook, FAQ |
| Doctor | Validates auth, API version, rate limits, config health |
| Agent-Native | --json, --select, --dry-run, --stdin, idempotent, typed exits |
| Local Cache | SQLite + FTS5, --no-cache bypass, cache clear |
| Breadth | Every API endpoint covered + convenience wrappers |
| Vision | Sync + search + tail + export + domain workflows |
| Workflows | Compound commands combining 2+ API calls |

## Quick Start

```bash
git clone https://github.com/mvanhorn/cli-printing-press.git
cd cli-printing-press
go build -o ./printing-press ./cmd/printing-press
```

Then use it as a Claude Code skill:

```bash
# In Claude Code
/printing-press Discord
/printing-press Stripe
/printing-press --spec ./openapi.yaml
```

Or run the generator directly:

```bash
# Generate from an OpenAPI spec
./printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json --output ./petstore-cli --force --lenient --validate

# Score a generated CLI
./printing-press scorecard --dir ./petstore-cli
```

## Project Structure

```
cmd/printing-press/         CLI entry point
internal/
  catalog/                  Catalog schema validator
  cli/                      CLI commands (generate, scorecard, version)
  docspec/                  Doc-to-spec generator
  generator/                Template engine + quality gates
    templates/              14 Go templates (root, command, store, sync, search, etc.)
  llm/                      LLM runner (claude/codex CLI)
  llmpolish/                LLM polish pass (help, examples, README)
  openapi/                  OpenAPI 3.0+ parser (strict + lenient modes)
  pipeline/                 Intelligence engine (research, scorecard, dogfood, planner)
  profiler/                 API shape analysis (volume, search need, realtime, etc.)
  spec/                     Internal YAML spec parser
catalog/                    Known API specs with verified URLs
skills/printing-press/      Claude Code skill definition (8-phase pipeline)
docs/plans/                 Generated plan artifacts
```

## Development

```bash
go build -o ./printing-press ./cmd/printing-press
go test ./...
```

## License

MIT
