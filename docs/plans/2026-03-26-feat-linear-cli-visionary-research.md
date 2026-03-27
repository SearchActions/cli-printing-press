---
title: "Visionary Research: Linear CLI"
type: feat
status: active
date: 2026-03-26
phase: "0"
api: "linear"
---

# Visionary Research: Linear CLI

## Overview

Linear is a project management and issue tracking platform built for software engineering teams. It offers a GraphQL-only API with cursor-based (Relay-style) pagination, OAuth 2.0 and API key authentication, webhooks for real-time updates, and ~400+ GraphQL types. The API is well-documented but complex - query complexity scoring limits aggressive data fetching, and rate limiting discourages polling (webhooks preferred).

The competitive landscape reveals 4 existing CLIs (schpet/linear-cli leading at 521 stars) plus a rich ecosystem of sync tools (synclinear.com, linear-sync). However, NO existing tool provides local persistence with full-text search - every CLI hits the live API on every query. This is the primary strategic opportunity.

## API Identity

- **Domain:** Project Management / Issue Tracking
- **Primary users:** Software engineers, engineering managers, DevOps/platform teams
- **Core entities:** Issues, Projects, Cycles, Teams, Users, Comments, Labels, WorkflowStates, Documents, Initiatives, Milestones, CustomViews, Attachments
- **Data profile:**
  - Write pattern: Mutable (issues change states, get assigned, re-prioritized)
  - Volume: Medium-high (thousands to tens of thousands of issues per workspace)
  - Real-time: Webhooks (HTTP push on create/update/delete), no WebSocket/SSE
  - Search need: **HIGH** - users constantly search issues by title, description, assignee, label, project, state
- **API type:** GraphQL only (single endpoint, Relay-style cursor pagination)
- **Auth:** API key (personal) + OAuth 2.0 (applications)
- **Rate limiting:** Complexity-based (0.1 per property, 1 per object, connections multiply by pagination arg)

## Usage Patterns (Top 5 by Evidence)

| Rank | Pattern | Evidence Score | Sources |
|------|---------|---------------|---------|
| 1 | **Issue management from CLI** (list, create, update, triage) | 10/10 | 4 competing CLIs, HN posts, multiple GitHub repos |
| 2 | **Git branch <-> issue linking** (create branch from issue, detect issue from branch) | 8/10 | schpet/linear-cli (521 stars) focuses on this, linear-4-terminal has git integration |
| 3 | **Linear <-> GitHub sync** (bidirectional issue sync) | 7/10 | synclinear.com (calcom), jtormey/linear-sync, spacedriveapp/linear-sync |
| 4 | **Bulk operations** (batch status changes, assignments, label updates) | 6/10 | AdiKsOnDev/linear-cli has bulk ops, multiple GitHub issues requesting this |
| 5 | **Stale issue detection and cleanup** (find old unassigned issues) | 6/10 | Engineering manager workflow, no existing tool provides this as a command |

## Tool Landscape (Beyond API Wrappers)

| Tool | Type | Stars | What It Does |
|------|------|-------|-------------|
| schpet/linear-cli | Workflow Tool | 521 | Git-aware issue management, PR creation, agent-friendly skills |
| synclinear.com (calcom) | Integration Tool | - | Bidirectional Linear <-> GitHub issue sync |
| reverse-linear-sync-engine | Data Tool | - | Reverse engineering of Linear's local-first sync engine |
| Swarmia | Analytics Tool | - | Team productivity insights from Linear data |
| linear-4-terminal | API Wrapper | 8 | Interactive TUI + CLI with 30+ commands (Rust) |
| AdiKsOnDev/linear-cli | API Wrapper | 8 | 40+ commands with bulk ops (Python) |
| iatsiuk/linear-cli | API Wrapper | 0 | 15+ commands (Go) |
| OpenSwarm | Workflow Tool | - | Multi-agent Claude CLI orchestrator for Linear/GitHub |

**Key insight:** No existing tool provides local SQLite persistence with full-text search for Linear issues. Every CLI hits the API on every query. The "discrawl gap" is wide open.

## Workflows

| # | Name | Steps | Proposed CLI Feature |
|---|------|-------|---------------------|
| 1 | **Stale Issue Triage** | Query issues by team -> filter by updatedAt -> group by state -> report | `linear stale --days 30 --team ENG` |
| 2 | **Sprint Velocity** | Get current cycle -> count issues by state -> compute completion % -> trend | `linear velocity --cycle current --team ENG` |
| 3 | **Orphan Detection** | Find issues with no project, no cycle, unassigned -> report | `linear orphans --team ENG` |
| 4 | **Standup Report** | Get issues assigned to me -> filter by recent activity -> format | `linear standup --days 1` |
| 5 | **Label Audit** | List all labels -> find unused labels -> find issues with no labels | `linear label-audit --team ENG` |

## Architecture Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **Persistence** | SQLite with domain-specific tables | Issues are high-volume, mutable, and heavily searched. JSON blobs won't support the joins and filters users need. |
| **Real-time** | REST polling with `updatedAt` cursor (no WebSocket/SSE available) | Linear's API has no WebSocket or SSE. Webhooks are push-only (need a server). Polling with `filter: { updatedAt: { gte: cursor } }` is the viable CLI approach. |
| **Search** | FTS5 on issue title + description + comment body | Users search issues constantly. Local FTS5 is instant vs. API round-trips with complexity cost. |
| **Bulk** | GraphQL mutations with batch variables | The API supports individual mutations; batch via multiple mutations in one request. |
| **Cache** | SQLite IS the cache. `--no-cache` bypasses local DB and hits API directly. | Single persistence layer avoids cache invalidation complexity. |

## Top 5 Features for the World

| Rank | Feature | Score | Breakdown | Description |
|------|---------|-------|-----------|-------------|
| 1 | **Local sync + FTS5 search** | 14/16 | Evidence:3, Impact:3, Feasibility:2, Unique:2, Compose:2, DataFit:2, Maintain:0, Moat:0 | Sync issues to SQLite, search instantly with `linear search "auth bug" --team ENG --state "In Progress"` |
| 2 | **Stale issue detector** | 13/16 | Evidence:2, Impact:3, Feasibility:2, Unique:2, Compose:2, DataFit:2, Maintain:1, Moat:0 | `linear stale --days 30 --team ENG` - no existing tool does this |
| 3 | **Cycle velocity/burndown** | 12/16 | Evidence:2, Impact:2, Feasibility:2, Unique:2, Compose:2, DataFit:2, Maintain:1, Moat:0 | `linear velocity --cycle current` - powered by local DB aggregation |
| 4 | **Standup generator** | 11/16 | Evidence:2, Impact:3, Feasibility:2, Unique:2, Compose:1, DataFit:1, Maintain:0, Moat:0 | `linear standup --days 1` - my recent activity across teams |
| 5 | **Raw SQL access** | 10/16 | Evidence:1, Impact:2, Feasibility:2, Unique:3, Compose:2, DataFit:2, Maintain:0, Moat:0 | `linear sql "SELECT count(*) FROM issues WHERE state = 'Done' AND updated_at > date('now', '-7 days')"` |

## Demand Signals

- HN "A CLI for Using Linear with GitHub" (2025) - creator says it's their "second most used CLI tool after git"
- HN "Show HN: OpenSwarm - Multi-Agent Claude CLI Orchestrator for Linear/GitHub" (2026) - agent orchestration demand
- schpet/linear-cli explicitly designed as "agent friendly" with Claude Code skills
- Linear's own developer docs warn against API polling, suggesting local persistence is the right architectural choice
- 4 competing CLIs exist but none provide local persistence or workflow commands

## Sources

- https://linear.app/developers - Official API docs
- https://github.com/schpet/linear-cli - Leading competitor (521 stars)
- https://github.com/nooesc/linear-4-terminal - Rust CLI (8 stars)
- https://github.com/AdiKsOnDev/linear-cli - Python CLI (8 stars)
- https://github.com/iatsiuk/linear-cli - Go CLI (0 stars)
- https://github.com/dalys/awesome-linear - Ecosystem tools list
- https://github.com/calcom/synclinear.com - Linear<->GitHub sync
- https://news.ycombinator.com/item?id=44222504 - HN: Linear CLI discussion
- https://news.ycombinator.com/item?id=43922108 - HN: Show HN Linear CLI
- https://news.ycombinator.com/item?id=47160980 - HN: OpenSwarm multi-agent
