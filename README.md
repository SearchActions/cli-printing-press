# CLI Printing Press

An agent infrastructure factory. Give it an API name. Get back the CLI that your agents need.

```bash
/printing-press Discord
/printing-press Stripe
/printing-press Linear
```

One command. 8 phases. ~1 hour. Produces a production-ready Go CLI binary + 7 deep analysis documents. REST or GraphQL - it figures it out.

## Why Every API Needs a CLI

In 2026, most code isn't written by humans - it's written by agents under human direction. Agents need to interact with APIs. They have two choices: import an SDK and write 15 lines of code, or run a single CLI command. The CLI wins on every dimension that matters.

**Token economics.** A single MCP server exposes ~93 tools costing [~55,000 tokens](https://manveerc.substack.com/p/mcp-vs-cli-ai-agents) just to load tool definitions. At scale (10,000 sessions/day), that's $1,600/day on definitions alone. A CLI command with `--help` costs ~200 tokens. Full execution cycle: <500 tokens. That's a [100x reduction](https://manveerc.substack.com/p/mcp-vs-cli-ai-agents).

**Training data.** LLMs were trained on enormous volumes of shell interactions. Unix pipe chains are deeply embedded in model weights. [MCP composition patterns have zero training data and zero production hardening](https://manveerc.substack.com/p/mcp-vs-cli-ai-agents). When an agent sees `mycli list --json | jq '.[] | select(.status == "active")'`, it doesn't need to learn anything - it already knows.

**Delegation, not suggestion.** [IDE agents are designed for suggestion. CLI agents are designed for delegation.](https://www.firecrawl.dev/blog/why-clis-are-better-for-agents) Terminal agents run for hours without supervision, coordinate changes across dozens of files, and self-heal on failure. Exit code 1 means "try again." Exit code 0 means "done." No screenshots, no clicking, no fragile UI automation.

**Composability.** `mycli issues list --json --select id,title | jq -r '.[].id' | xargs -I{} mycli issues close {}` - one line, three tools, zero SDK imports. [CLI is the universal interface for both humans and AI agents](https://github.com/HKUDS/CLI-Anything) because text commands match LLM output format and chain into complex workflows.

Every API that gets a CLI becomes instantly accessible to every agent framework - Claude Code, Codex, Gemini CLI, open source agents. No SDK integration. No dependency management. The printing press is the factory that manufactures this interface layer, one API at a time.

### The Human + Agent Model

```
Power User (architect)  -->  Agent (operator)  -->  CLI (interface)  -->  API
  "Find stale issues"      runs the command       linear-cli stale      GraphQL
  "Who's overloaded?"      parses JSON output     linear-cli load       queries
  "Fix the auth bug"       chains 5 commands      linear-cli issue...   mutations
```

The human sets direction. The agent executes. The CLI is the reliable, structured, token-efficient bridge between them. The printing press builds that bridge for any API.

## The Non-Obvious Insight

Every API has a secret. The data it exposes is useful for something its creators never designed for. The printing press finds that secret and builds a CLI around it.

The **Non-Obvious Insight (NOI)** is a one-sentence reframe of what the API's data actually IS:

```
"[API] isn't just [obvious thing]. It's [non-obvious thing].
 Every [data point] is a signal about [hidden truth]."
```

| API | What they think it is | What it actually is |
|-----|----------------------|---------------------|
| Discord | A chat app | A **searchable knowledge base**. Every message thread is institutional memory. |
| Linear | An issue tracker | A **team behavior observatory**. Every state change is a signal about how your team actually works vs. how they think they work. |
| Stripe | A payment processor | A **business health monitor**. Every failed charge and churn event is a signal about product-market fit. |
| GitHub | A code host | An **engineering culture fingerprint**. Every review turnaround and merge pattern is a signal about how your team ships. |
| Notion | A doc editor | A **knowledge decay detector**. Every stale page and orphaned database is a signal about what your team has forgotten. |
| Slack | Messaging | An **organizational nervous system**. Every response time and channel silence is a signal about team health. |

The NOI drives everything downstream - what the store captures, what workflow commands do, what the README says, and what insight commands detect.

Phase 0 cannot complete without an NOI. If the LLM can't write one, the research wasn't deep enough.

## The Creativity Ladder

Most API CLIs stop at Rung 1. The printing press climbs to Rung 5.

| Rung | What It Is | Auto-Generated? | Example |
|------|-----------|-----------------|---------|
| 1 | API wrapper commands | Yes (from spec) | `issue create --title "..."` |
| 2 | Output formatting | Yes (always) | `--json`, `--select`, `--csv`, `--dry-run` |
| 3 | Local persistence | Yes (conditional) | `sync`, `search`, `export`, `tail` |
| 4 | **Domain analytics** | **Yes (from archetype)** | `stale --days 30`, `orphans`, `load` |
| 5 | **Behavioral insights** | **Yes (from archetype)** | `health` (composite score), `similar` (duplicate detection) |

Rung 4 is where discrawl lives (12 commands, 540 stars). Rung 5 is where nobody else is yet.

## Domain Archetypes

The profiler classifies every API into a domain archetype and auto-generates the right workflow + insight commands:

| Archetype | Detected By | Auto-Generated Commands |
|-----------|------------|------------------------|
| **Project Management** | issue/task/ticket resources, assignee fields, priority levels, due dates | `stale`, `orphans`, `load`, `health`, `similar` |
| **Communication** | message/channel/thread resources, threading fields | `channel-health`, `message-stats`, `audit-report`, `health`, `similar` |
| **Payments** | charge/payment/invoice resources, amount/currency fields | `reconcile`, `revenue`, `health`, `similar` |
| **Infrastructure** | server/deploy/instance resources | `health`, `similar` |
| **Content** | document/page/block resources | `health`, `similar` |
| **CRM** | contact/deal/lead resources | `health`, `similar` |
| **Developer Platform** | repo/PR/commit resources | `health`, `similar` |

The archetype is detected automatically from the spec - no configuration needed. The entity mapper figures out which resource is the "primary entity" (issues for PM, messages for comms, charges for payments) and wires the templates accordingly.

## How It Works

The press runs 8 mandatory phases. Each phase writes a comprehensive plan document. The artifacts are the product - the CLI is a side effect.

```
Phase 0    Visionary Research       (3-5 min)    NOI + domain identity + usage patterns
Phase 0.5  Power User Workflows     (2-3 min)    Compound commands power users want
Phase 0.7  Prediction Engine        (15-25 min)  Local data layer spec (SQLite + FTS5)
Phase 1    Deep Research            (5-8 min)    Competitors, strategic justification
Phase 2    Generate                 (1-2 min)    Produce Go CLI from spec + archetype templates
Phase 3    Steinberger Audit        (5-8 min)    Score against quality bar, plan improvements
Phase 4    GOAT Build               (5-10 min)   Review generated workflows, add NOI-driven insights
Phase 4.5  Dogfood Emulation        (10-20 min)  Test every command against spec-derived mocks
Phase 5    Final Steinberger        (2-3 min)    Before/after scoring delta + final report
```

### Phase 0: Visionary Research (now with NOI gate)

Before generating anything, the press understands the API's domain, users, and ecosystem. It searches GitHub, Reddit, Hacker News, and Stack Overflow for usage patterns, pain points, and existing tools. It classifies every tool it finds - API wrappers, data tools, workflow tools, environment tools - and identifies what's missing.

**New gate:** Phase 0 cannot complete without writing the Non-Obvious Insight. The NOI becomes the creative DNA of the entire CLI.

### Phase 0.5: Power User Workflows

Predicts what compound commands a power user would want - the features that make a tool worth starring. "Find stale issues." "Show workload by team member." "Detect cross-team blockers." These workflow commands combine 2+ API calls into single operations.

### Phase 0.7: Prediction Engine

The differentiator. The press predicts what a developer would build on TOP of the API - without looking at competitors. It classifies every entity (accumulating, reference, append-only, ephemeral), scores them on data gravity (volume, query frequency, join demand, search need, temporal value), and produces a concrete data layer specification: domain-specific SQLite schema, incremental sync strategy with validated cursors, FTS5 full-text search with domain filters, and compound queries.

### Phase 2: Generate (now with archetype templates)

Runs the Go generator against the spec. For APIs with detected archetypes, the generator now automatically produces workflow commands (stale, orphans, load) and insight commands (health, similar) from templates - no Phase 4 hand-writing required.

The **schema builder** uses data gravity scoring to generate domain-specific SQLite tables with extracted columns, FK indexes, and FTS5 triggers. High-gravity entities get proper column extraction instead of generic JSON blobs.

### Phase 3: Steinberger Audit

Scores the generated CLI against the Steinberger bar (12 dimensions, 120 max). Peter Steinberger's gogcli is the 10/10 reference.

### Phase 4: GOAT Build (shifted focus)

Before the Creative Vision Engine, Phase 4 meant hand-writing 7 workflow Go files. Now it means reviewing the auto-generated workflows, customizing SQL queries for unique domain patterns, and writing 2-3 NOI-driven insight commands that embody the Non-Obvious Insight. The LLM's time shifts from mechanical coding to creative domain analysis.

### Phase 4.5: Dogfood Emulation

Tests every command against spec-derived mock responses - no API keys needed. Scores each command on 5 dimensions. Auto-fixes issues, re-scores, and writes a report.

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

**Auto-generated workflow commands** (when domain archetype is detected):

| PM APIs | Communication APIs |
|---------|-------------------|
| `stale` - items with no updates in N days | `channel-health` - activity analysis |
| `orphans` - items missing assignment/project | `message-stats` - volume analytics |
| `load` - workload distribution per assignee | `audit-report` - audit log analysis |

**Auto-generated insight commands** (Rung 5):

| Command | What It Does | When Generated |
|---------|-------------|----------------|
| `health` | Composite workspace score (0-100) from stale ratio, orphan ratio, velocity | All archetypes with store |
| `similar` | FTS5-based duplicate detection across items | All archetypes with FTS5 |

**Exit codes:** `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## 7 Plan Artifacts Per Run

Every run produces 7 comprehensive analysis documents in `docs/plans/`:

```
Phase 0   -> <api>-cli-visionary-research.md      NOI, API identity, usage patterns, tool landscape
Phase 0.5 -> <api>-cli-power-user-workflows.md    Workflow ideas, scoring, top 7 selected
Phase 0.7 -> <api>-cli-data-layer-spec.md         SQLite schema, sync strategy, search filters
Phase 1   -> <api>-cli-research.md                Competitors, strategic justification
Phase 3   -> <api>-cli-audit.md                   Steinberger scores, improvement plan
Phase 4   -> <api>-cli-goat-build-log.md          What was built, what was fixed
Phase 4.5 -> <api>-cli-dogfood-report.md          Per-command scores, hallucination detection
```

## Works With Any API

**REST APIs** (OpenAPI/Swagger): Full pipeline - generator produces commands, archetype templates add workflows, dogfood validates everything.

**GraphQL APIs** (Linear, Shopify, GitHub GraphQL): Produces scaffolding + GraphQL client wrapper, hand-writes commands in Phase 4. All research, prediction, data layer, and dogfood phases run normally.

**No spec available**: Reads the API docs, writes a spec, generates from it.

The press dynamically detects API type from the spec content - not a hardcoded list.

## The Steinberger Bar

12 dimensions, 120 points max. Grade A = 80%+.

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
| **Insight** | **Behavioral commands that see patterns humans miss (health, similar, bottleneck)** |

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
    entity_mapper.go        Entity role detection and field mapping (NEW)
    schema_builder.go       Data gravity scoring + domain table generation (NEW)
    templates/              Go templates
      workflows/            PM workflow templates (stale, orphans, load) (NEW)
      insights/             Behavioral insight templates (health, similar) (NEW)
  llm/                      LLM runner (claude/codex CLI)
  llmpolish/                LLM polish pass (help, examples, README)
  openapi/                  OpenAPI 3.0+ parser (strict + lenient modes)
  pipeline/                 Intelligence engine (research, scorecard, dogfood, planner)
  profiler/                 API shape + domain archetype analysis (EXTENDED)
  spec/                     Internal YAML spec parser
  vision/                   Visionary plan types + NOI system (EXTENDED)
    insight.go              Non-Obvious Insight struct (NEW)
catalog/                    Known API specs with verified URLs
skills/printing-press/      Claude Code skill definition (8-phase pipeline)
  references/
    noi-examples.md         10+ NOI examples across all archetypes (NEW)
docs/plans/                 Generated plan artifacts
```

## Development

```bash
go build -o ./printing-press ./cmd/printing-press
go test ./...
```

## License

MIT
