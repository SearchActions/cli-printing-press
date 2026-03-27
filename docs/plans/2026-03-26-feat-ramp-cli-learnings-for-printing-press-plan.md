---
title: "Learnings from Ramp CLI for Printing Press"
type: feat
status: active
date: 2026-03-26
---

# Learnings from Ramp CLI for Printing Press

## What Ramp Built

Ramp launched [ramp-cli](https://github.com/ramp-public/ramp-cli) (v0.1.1, March 25, 2026) - a Python CLI for their financial platform designed explicitly for AI agents. 4 stars but it's day 1. MIT licensed.

They also launched [ramp_mcp](https://github.com/ramp-public/ramp_mcp) (31 stars) - an MCP server that loads financial data into an ephemeral SQLite database for LLM analysis.

**Installation:** `curl -fsSL https://agents.ramp.com/install.sh | bash`

## 7 Design Decisions Worth Stealing

### 1. Auto-Detect Agent vs Human Mode (TTY Detection)

Ramp's CLI automatically switches output format based on context:
- **Piped/scripted:** JSON (105 tokens per transaction)
- **Interactive terminal:** Formatted tables (280 tokens)
- Override with `--agent` or `--human` or `--output json|table`

**What printing-press does:** We have `--json` and `--plain` flags but NO auto-detection. The agent has to explicitly pass `--json` every time.

**Lesson:** The `helpers.go.tmpl` `printOutput` function should check `isatty(os.Stdout.Fd())` and default to JSON when piped. This is a one-line template change that eliminates the most common agent friction.

### 2. `--no-input` Global Flag (Non-Interactive Mode)

Ramp has `--no-input` as a global flag that disables ALL interactive prompts. Combined with TTY detection, this means agents never get stuck on a prompt.

**What printing-press does:** We have `--yes` for confirmation bypass but no global `--no-input` that suppresses all prompts.

**Lesson:** Add `--no-input` to root.go.tmpl alongside `--yes`. The template should gate all `bufio.Scanner(os.Stdin)` calls on `!flags.noInput`.

### 3. `resource tool` Command Structure (Not CRUD Verbs)

Ramp uses `ramp transactions list` and `ramp transactions get` but also domain-specific tools like:
- `ramp transactions missing` (find transactions without receipts)
- `ramp transactions flag-missing` (flag them for follow-up)
- `ramp transactions memo-suggestions` (AI-generated memo text)
- `ramp transactions explain` (natural language explanation)
- `ramp transactions trips` (group by travel trip)
- `ramp bills draft` / `ramp bills pending` / `ramp bills approve`

This is NOT generic CRUD. These are **workflow commands** at the resource level.

**What printing-press does:** We generate CRUD commands (list, get, create, update, delete) per resource, then add workflow commands as separate top-level commands (stale, velocity, etc).

**Lesson:** Workflow commands should be subcommands of their resource, not separate top-level commands. `discord-cli messages search` not `discord-cli search`. `linear-cli issues stale` not `linear-cli stale`. This makes progressive help discovery work naturally - agents discover workflow commands when they `--help` on a resource.

### 4. Skills System (Agent Instructions as First-Class)

Ramp has a `skills` command that lets agents browse skill instructions. This is basically built-in agent documentation that's optimized for LLM consumption, not human reading.

**What printing-press does:** We generate --help text but it's written for humans. No skill/instruction system for agents.

**Lesson:** Generated CLIs should include a `skills` or `agent-help` command that outputs concise, token-efficient instructions for each resource. Not the full --help text - a compressed version optimized for agent context windows. This could be a new template.

### 5. `--dry_run` Shows Full Request (Not Just URL)

Ramp's `--dry_run` shows the complete request that would be sent, not just the URL path. For POST/PUT, this includes the full request body.

**What printing-press does:** Our `--dry-run` shows method + URL + body, which is close. But for GraphQL, it should show the full query string.

**Lesson:** We're already doing this well for REST. For GraphQL, the dry-run should show the actual GraphQL query + variables that would be sent.

### 6. MCP Server Uses Ephemeral SQLite (Same Pattern as Our Store)

Ramp's MCP server loads API data into an in-memory SQLite database so Claude can run SQL queries against it. This is exactly our Phase 0.7 data layer pattern.

**What printing-press does:** We generate SQLite stores with domain-specific tables. This validates our approach.

**Lesson:** We're already doing the right thing here. Ramp validates that SQLite-as-intermediary-for-LLM is the correct pattern for financial/business APIs. Our implementation goes further (persistent, FTS5, incremental sync) which is better for CLI use cases.

### 7. Token-Conscious Output Design

Ramp explicitly measures and optimizes output token counts:
- Agent mode: 105 tokens per transaction
- Human mode: 280 tokens per transaction

They design output to minimize context window consumption.

**What printing-press does:** We don't measure or optimize output token counts. Our `--json` output includes the full API response body.

**Lesson:** Add `--compact` or `--brief` output mode that returns only the most important fields (id, name, status, key timestamps). For list commands, show just the summary fields. For get commands, show the full object. This is a new template feature.

## What Ramp Gets Right That We Should Copy

| Feature | Ramp | Printing Press | Priority to Copy |
|---|---|---|---|
| Auto-detect JSON when piped | Yes (TTY check) | No (requires --json) | HIGH - one-line template fix |
| `--no-input` global flag | Yes | No (only --yes) | HIGH - one-line template fix |
| Workflow commands as resource subcommands | Yes (`transactions missing`) | No (separate top-level) | MEDIUM - template restructure |
| Token-counted output modes | Yes (105 vs 280 tokens) | No | MEDIUM - new --compact mode |
| Agent skills/instructions command | Yes (`ramp skills`) | No | LOW - new template |
| `--page_size` and `--next_page_cursor` | Yes (explicit pagination) | Partial (--limit, --all) | LOW - already close |

## What Printing Press Does Better Than Ramp

| Feature | Printing Press | Ramp CLI |
|---|---|---|
| Local SQLite persistence | Domain tables + FTS5 + incremental sync | MCP only (ephemeral, no CLI persistence) |
| Full-text search | FTS5 offline search | No local search |
| Data gravity scoring | Automated from spec | Manual design |
| Workflow analytics | stale, velocity, workload, deps | Limited to approval/receipt workflows |
| Multi-API generation | Any OpenAPI or GraphQL spec | Ramp-only |
| Dogfood validation | Automated `printing-press dogfood` command | None |
| Honest scorecard | Tier 2 semantic validation | None |

## Concrete Changes to Implement

### Quick Wins (1 hour total)

1. **Auto-JSON when piped** - In `helpers.go.tmpl`, change `printOutput` to detect `!isatty.IsTerminal(os.Stdout.Fd())` and default to JSON output. Keep explicit `--json` flag for override. (~15 min)

2. **Add `--no-input` flag** - In `root.go.tmpl`, add `noInput bool` to rootFlags and `--no-input` persistent flag. Gate any interactive prompts on `!flags.noInput`. (~15 min)

3. **Add `--compact` output mode** - New flag in root.go.tmpl. When set, `printOutput` filters JSON to only include fields named: id, name, title, status, state, created_at, updated_at, url, identifier. (~30 min)

### Medium Effort (future session)

4. **Restructure workflow commands as resource subcommands** - Instead of `linear-cli stale`, generate `linear-cli issues stale`. This requires changes to the entity mapper and root.go.tmpl to nest workflow commands under their parent resource. (~2 hours)

5. **Add `skills` command template** - New template that generates a `skills` subcommand outputting token-efficient agent instructions per resource. (~1 hour)

## Non-Obvious Insight

Ramp's biggest insight isn't any single feature - it's the **output token consciousness**. They explicitly measure the token cost of each output mode and optimize for it. This is a mindset shift: when your primary user is an LLM that pays per-token for context window, **every byte of output is a cost center**.

Our printing-press generates CLIs that produce verbose output because the templates were designed for human readability. The `--json` flag helps but even JSON output includes every field. A `--compact` mode that returns only the 5-7 most important fields per entity would cut agent token costs by 60-80%.

This connects to our Phase 0.7 data gravity scoring: entities with high data gravity should have their high-gravity fields identified and used as the `--compact` field set.

## Sources

- [Ramp CLI](https://github.com/ramp-public/ramp-cli) - 4 stars, MIT, v0.1.1
- [Ramp MCP](https://github.com/ramp-public/ramp_mcp) - 31 stars, MIT
- [Ramp Developer Platform](https://ramp.com/developer-tools)
- [Ramp Developer Community Blog](https://ramp.com/blog/introducing-the-ramp-developer-community)
- [agents.ramp.com](https://agents.ramp.com) - CLI landing page
- [dax on Ramp](https://x.com/thdxr/status/2010729452591054949) - "it's insane how much ramp built"
- [Trevin Chow's 7 Principles](https://x.com/trevin) - published same day as Ramp CLI launch
