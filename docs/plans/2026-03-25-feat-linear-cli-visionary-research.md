---
title: "Visionary Research: Linear CLI"
type: feat
status: active
date: 2026-03-25
---

# Visionary Research: Linear CLI

## API Identity
- **Domain:** Project management / issue tracking for software development
- **Primary users:** Software engineers, engineering managers, product managers, AI coding agents
- **Core entities:** Issues, Teams, Projects, Cycles, Workflow States, Users, Comments, Labels, Documents, Project Updates, SLAs
- **API type:** GraphQL-only (no REST API) at `https://api.linear.app/graphql`
- **Data profile:**
  - Write pattern: Full CRUD via GraphQL mutations
  - Volume: Rate limited (1,500 req/hr API key, 250,000 complexity points/hr)
  - Real-time: Webhooks for all entity types; no public WebSocket API
  - Search need: HIGH - users constantly search/filter issues by team, state, assignee, label, priority
  - Pagination: Relay-style cursor-based, default 50 per page

## Usage Patterns (Top 5 by Evidence)

| Rank | Pattern | Evidence Score | What it needs |
|------|---------|---------------|---------------|
| 1 | GitHub sync (issues, PRs, status) | 9/10 | Bidirectional issue-PR linking, auto-status updates |
| 2 | AI agent integration (MCP/Claude/Cursor) | 8/10 | Token-efficient output, task export, structured JSON |
| 3 | Migration/import (Jira, Asana, Pivotal) | 7/10 | Bulk create with preserved metadata |
| 4 | Release notes / changelog generation | 5/10 | Fetch completed issues by cycle/project, group by label |
| 5 | CLI issue management (create, list, filter, update) | 5/10 | Terminal CRUD without context-switching |

## Tool Landscape (Beyond API Wrappers)

### CLIs (Environment Tools)
- **schpet/linear-cli** (518 stars, TS/Deno): Market leader. Git/jj-aware, agent skill built-in. Lacks --json on most commands.
- **joa23/linear-cli** (113 stars, Go): Agent-native, token-efficient (~50 tokens minimal mode). 2 months old, auth bugs.
- **czottmann/linearis** (163 stars, TS): JSON-structured output, LLM-optimized. Simpler feature set.
- **Finesssee/linear-cli** (58 stars, Rust): Broadest features (50+ commands, burndown charts, webhooks). No agent features.
- **evangodon/linear-cli** (92 stars, Go): Unmaintained (README warns "should not be used").

### Data Tools
- **wzhudev/reverse-linear-sync-engine** (1,922 stars): Educational reverse engineering of Linear's sync. Not a usable tool.
- **terrastruct/byelinear** (14 stars, Go): Export Linear issues to GitHub issues.
- **nverges/linear-importer** (14 stars, JS): Import from Pivotal Tracker.

### Workflow Tools
- **jtormey/linear-sync** (77 stars, Elixir): Bidirectional Linear-GitHub sync via webhooks.
- **linear/linear-release-action** (3 stars, official): Scans commits for Linear issue IDs in CI.

### Integration Tools
- **jerhadf/linear-mcp-server** (346 stars): Deprecated in favor of official `mcp.linear.app/sse`.
- **tacticlaunch/mcp-linear** (134 stars): Community MCP server.
- **casals/obsidian-linear-integration-plugin** (21 stars): Obsidian bidirectional sync.

## Workflows

| # | Workflow | Steps | Proposed CLI Command |
|---|----------|-------|---------------------|
| 1 | Stale issue triage | Query issues by team, filter by updatedAt, group by priority | `linear-cli stale --days 30 --team ENG` |
| 2 | Cycle velocity report | Fetch cycle issues, count completed/cancelled/carried, compute velocity | `linear-cli velocity --team ENG --cycles 3` |
| 3 | Sprint standup | List in-progress issues by assignee, show blockers | `linear-cli standup --team ENG` |
| 4 | Release notes | Fetch completed issues from cycle, group by label, format as markdown | `linear-cli release-notes --cycle current` |
| 5 | Issue triage | List untriaged issues (no priority/assignee), bulk update | `linear-cli triage --team ENG` |
| 6 | PR dashboard | List issues with linked PRs, show PR status | `linear-cli pr-status --team ENG` |

## Architecture Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **API client** | Hand-written GraphQL client in Go | GraphQL-only API; Go for fast binary, easy distribution |
| **Persistence** | None (stateless API calls) | Linear's search/filter is good enough; no need for local SQLite |
| **Real-time** | Not in v1 | Webhooks require public endpoints; focus on polling-based workflows |
| **Search** | GraphQL filters + local post-processing | Linear supports rich filtering on issues |
| **Bulk** | Sequential mutations with rate limiting | No bulk mutation API; must loop with backoff |
| **Cache** | Optional in-memory cache for team/user lookups | Teams and users change rarely; avoid redundant queries |
| **Auth** | API key (primary) + OAuth 2.0 (future) | API key is simplest for CLI; OAuth for enterprise later |

## Top 5 Features for the World

| Rank | Feature | Score | Description |
|------|---------|-------|-------------|
| 1 | Stale issue detection | 14/16 | Find issues with no updates in N days, grouped by team/priority. No existing tool does this well. |
| 2 | Cycle velocity analytics | 13/16 | Completion rate, carry-over rate, velocity trend across N cycles. Only Finesssee has burndown. |
| 3 | Standup report | 12/16 | What's in-progress, what's blocked, what was completed yesterday - per team or person. |
| 4 | Release notes generation | 11/16 | Completed issues from a cycle/project, grouped by label, formatted as markdown changelog. |
| 5 | Agent-native issue export | 10/16 | Export issues as structured task lists optimized for AI agent consumption (minimal tokens). |
