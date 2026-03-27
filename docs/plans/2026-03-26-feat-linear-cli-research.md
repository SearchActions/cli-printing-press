---
title: "Research: Linear CLI"
type: feat
status: active
date: 2026-03-26
phase: "1"
api: "linear"
---

# Research: Linear CLI

## Spec Discovery

- Official GraphQL SDL: https://raw.githubusercontent.com/linear/linear/master/packages/sdk/src/schema.graphql
- Source: Linear's official GitHub repo (linear/linear)
- Format: GraphQL SDL (NOT OpenAPI - Linear is GraphQL-only)
- Type count: ~400+ types, covering entities, inputs, filters, comparators, payloads, enums
- Key query types: issues, projects, cycles, teams, users, comments, documents, milestones, initiatives
- All queries use Relay-style cursor pagination (first/after, last/before)
- Filtering supported on all paginated queries via typed filter inputs

## Competitors (Deep Analysis)

### schpet/linear-cli (521 stars) - THE LEADER
- Repo: https://github.com/schpet/linear-cli
- Language: TypeScript/Deno
- Commands: ~40+ across issues, teams, projects, milestones, documents
- Last commit: January 2025 (cb749d3)
- Open issues: 16
- Maintained: Yes (411 commits, active development)
- Notable features:
  - Git branch-aware issue detection (detect issue from current branch)
  - Jj commit trailer support
  - "Agent friendly" design with Claude Code skills
  - Multi-workspace auth with TOML config
  - PR creation via `gh` CLI integration
- Weaknesses:
  - No --json on projects (issue #127)
  - No env var auth for CI/CD (issue #147)
  - No local persistence or search
  - No workflow commands (stale, velocity, health)
  - No bulk operations
  - No SQLite data layer

### nooesc/linear-4-terminal (8 stars)
- Repo: https://github.com/nooesc/linear-4-terminal
- Language: Rust (99.1%)
- Commands: 30+ across issues, projects, teams, comments, search, git, bulk
- Last commit: July 2025
- Open issues: 0
- Maintained: Lightly
- Notable features: Interactive TUI with dual-panel layout, multi-select bulk operations
- Weaknesses: Low adoption, no local persistence, no workflow commands

### AdiKsOnDev/linear-cli (8 stars)
- Repo: https://github.com/AdiKsOnDev/linear-cli
- Language: Python
- Commands: 40+
- Last commit: Unknown (recent)
- Open issues: 2
- Notable features: Bulk operations, YAML output, interactive mode, shell completions
- Weaknesses: Python dependency overhead, no local persistence, no workflow commands

### iatsiuk/linear-cli (0 stars)
- Repo: https://github.com/iatsiuk/linear-cli
- Language: Go
- Commands: 15+
- Open issues: 0
- Notable: Go-based (same language as ours), covers basics

## User Pain Points

> "The project command is missing several operations that exist in the GraphQL API, forcing a fallback to raw API calls. No way to update a project via CLI... project list outputs a human-readable table with slugs but no UUIDs, with no --json flag." - schpet/linear-cli#127

> "Allow reading the API key from the environment... would require an extra step to push credentials in a file within a GitHub Actions workflow." - schpet/linear-cli#147

> "For users who use Graphite and a stacking workflow, the default checkout command results in an untracked state within the stack." - schpet/linear-cli#60

> Linear itself warns: "Calls to our GraphQL API are rate limited... we discourage polling the API to fetch updates" - suggesting local persistence is the right architecture.

## Auth Method

- Type: API key (personal token) + OAuth 2.0 (application)
- Env var convention: `LINEAR_API_KEY` (used by iatsiuk/linear-cli, recommended in Linear docs)
- GraphQL endpoint: `https://api.linear.app/graphql`
- Auth header: `Authorization: <api-key>` (no Bearer prefix for API keys)

## Demand Signals

- HN "A CLI for Using Linear with GitHub" - creator says "second most used CLI tool after git"
- HN "Show HN: OpenSwarm - Multi-Agent Claude CLI Orchestrator for Linear/GitHub" (March 2026) - multi-agent demand
- 4 independent CLI projects exist, proving persistent demand
- schpet/linear-cli at 521 stars demonstrates significant user base
- Linear's developer docs explicitly support CLI/agent use cases

## Strategic Justification

**Why this CLI should exist when schpet/linear-cli has 521 stars:**

1. **No existing CLI has local persistence.** Every query hits the API. Our SQLite + FTS5 data layer enables instant local search, cross-entity queries, and workflow commands that are impossible with live API calls.

2. **No existing CLI has workflow commands.** `stale`, `velocity`, `orphans`, `standup`, `health`, `blocked`, `sla` - these are what engineering managers actually need. Existing CLIs are CRUD wrappers.

3. **No existing CLI has raw SQL access.** `linear sql "SELECT ..."` is unprecedented for a project management CLI.

4. **Agent-native from the ground up.** While schpet/linear-cli is "agent friendly," our CLI adds --json, --select, --dry-run, --stdin, --yes, --no-cache as first-class global flags. Every command is pipe-ready.

5. **Go binary vs TypeScript/Deno.** Single binary, no runtime dependency, instant startup. The Go ecosystem advantage for CLI distribution.

## Target

- **Command count:** 50+ (beat schpet's 40+ with workflow commands + data layer commands)
- **Key differentiator:** SQLite data layer with FTS5 search + 7 workflow commands
- **Quality bar:** Steinberger Grade A (80+/100)
- **Resources:** Issues (CRUD), Projects (CRUD), Cycles (CRUD), Teams (list/view), Users (list/view), Comments (CRUD), Labels (CRUD), Documents (CRUD), Milestones (CRUD), Relations (CRUD), Workflow States (list), plus sync, search, sql, tail, stale, velocity, orphans, standup, triage, health, blocked, sla, doctor, auth, config
