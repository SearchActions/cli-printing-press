---
title: "Research: Linear CLI"
type: feat
status: active
date: 2026-03-25
---

# Research: Linear CLI

## Spec Discovery
- Official OpenAPI spec: **None** - Linear is GraphQL-only
- GraphQL schema: https://github.com/linear/linear/blob/master/packages/sdk/src/schema.graphql
- API endpoint: `https://api.linear.app/graphql`
- Format: GraphQL (supports introspection, browsable via Apollo Studio)
- Entity count: Issues, Teams, Projects, Cycles, WorkflowStates, Users, Comments, Labels, Documents, ProjectUpdates, Initiatives, Roadmaps, Releases + more

## Competitors (Deep Analysis)

### schpet/linear-cli (518 stars) - MARKET LEADER
- Repo: https://github.com/schpet/linear-cli
- Language: TypeScript/Deno
- Commands: Issues (CRUD, list, start, comment), Teams, Projects, Milestones, Documents, Config, Shell completions
- Last commit: 2026-03-25 (today - very active, 411+ commits)
- Open issues: 33
- Maintained: Yes (solo maintainer)
- Install: Homebrew, JSR/Deno, npm/bun/pnpm, GitHub release binaries
- Notable features: Git + Jujutsu aware, auto-creates branches, .linear.toml config, Claude Code skill
- Weaknesses:
  - No `--json` output on most commands (#179, #188) - critical for scripting/agents
  - No `--label` filter (#180)
  - No native search command (#143)
  - Missing assignee/priority in issue view (#190)
  - No date range filters (#191)
  - ReDoS vulnerability in minimatch (#194)
  - Issue themes: branch ergonomics, credential safety, agent support gaps

### joa23/linear-cli (113 stars) - AGENT-NATIVE CONTENDER
- Repo: https://github.com/joa23/linear-cli
- Language: Go
- Commands: Issues (CRUD, search), Dependencies, Projects, Cycles (velocity/analytics), Teams, Users, Task export
- Last commit: 2026-02-26 (~1 month ago)
- Open issues: 5
- Maintained: New (2 months old), uncertain
- Notable features: Token-efficient output (~50 tokens minimal), 3 verbosity levels, OAuth agent mode, cycle analytics, JSON+jq, dependency graphs
- Weaknesses:
  - OAuth token refresh corrupts auth scope (#42)
  - No multi-workspace support (#2, #43)
  - Only 2 months old, longevity unproven

### Finesssee/linear-cli (58 stars) - FEATURE-RICH
- Repo: https://github.com/Finesssee/linear-cli
- Language: Rust
- Commands: 50+ (issues, projects, cycles, sprints, teams, docs, labels, webhooks, notifications, templates, milestones, roadmaps, initiatives, bulk ops)
- Last commit: 2026-03-24 (yesterday)
- Open issues: 0
- Maintained: Yes, but low community engagement
- Notable features: Burndown charts, velocity tracking, webhook listener with HMAC, self-update, watch mode, template system, saved filters
- Weaknesses: No Homebrew, no agent features, Rust install barrier, 58 stars despite broadest features

## User Pain Points
> "doing the same thing twice in two different places" - users quantify ~1 minute of friction per task for: find issue, assign, mark in-progress, copy branch name, switch to terminal, checkout (Reddit)
> "Linear's official MCP responses are too token-heavy for agents" - AI agent developers (multiple sources)
> "No good way to export Linear issues into agent task formats" - Claude/Cursor/Devin users
> "GraphQL complexity/rate-limit management is a DX tax" - developers hitting 1,500 req/hr limit

## Auth Method
- Type: Personal API key (Bearer token) + OAuth 2.0 (Authorization Code + PKCE)
- Env var convention: `LINEAR_API_KEY` (most tools) or `LINEAR_TOKEN`
- Rate limits: 1,500 req/hr (API key), 250,000 complexity points/hr, max 10,000 points per query

## Demand Signals
- 8+ independent people have built Linear CLIs since 2024 - classic unmet-demand signal
- schpet/linear-cli has 33 open issues (mostly feature requests) - solo maintainer can't keep up
- Linear launched official MCP server (May 2025) and "Linear for Agents" - validates the market
- AI agent integration is the fastest-growing usage pattern (MCP, Claude, Cursor, Devin)

## Strategic Justification
**Why this CLI should exist:** schpet/linear-cli (market leader, 518 stars) is designed for human terminal use - it lacks `--json` output on most commands, has no workflow commands (stale, velocity, standup), and no bulk operations. joa23/linear-cli (113 stars, Go) has token efficiency but is 2 months old with auth bugs. No existing CLI has compound workflow commands that combine multiple API calls. Our CLI fills the gap: Go binary + GraphQL client + 7 workflow commands + full `--json` output + agent-native design. We compete with the *combined* feature set of all three leaders while being the only one written in Go with workflow intelligence.

## Target
- Command count: ~25-30 (CRUD commands + 7 workflow commands + doctor + config)
- Key differentiator: Workflow commands (stale, velocity, standup, triage, workload, release-notes, my-day) that no competitor has
- Quality bar: Steinberger Grade A (80+/100)
