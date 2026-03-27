---
title: "Visionary Research: GitHub CLI"
type: feat
status: active
date: 2026-03-27
phase: "0"
api: "github"
---

# Visionary Research: GitHub CLI

## Overview

GitHub's REST API is one of the largest public APIs in existence - 740 paths, 1,107 operations across 25+ resource categories. The official `gh` CLI (43.4k stars, 601 contributors) dominates the space but is fundamentally a workflow tool, not a data tool. It excels at interactive developer workflows (PR creation, issue filing, Actions monitoring) but has no local persistence, no offline search, no analytics, and no bulk data operations. This creates a massive gap for the "GitHub power user" archetype - engineering managers, open source maintainers, DevOps engineers, and AI agents who need to query, analyze, and monitor GitHub data programmatically.

The strategic opportunity is not to replace `gh` but to complement it - building the "github-to-sqlite on steroids" that combines the data layer approach of Dogsheep's github-to-sqlite (462 stars, 15 tables, SQLite + Datasette) with the agent-native output modes of a modern CLI (--json, --select, --dry-run, --stdin), plus compound workflow commands that solve real recurring problems.

## API Identity

- **Domain:** Developer Platform (repos, issues, PRs, actions, security, packages)
- **Primary users:** Engineering managers, open source maintainers, DevOps/SRE, CI/CD automation, AI coding agents
- **Core entities:** Repositories, Issues, Pull Requests, Commits, Actions (Workflows/Runs/Jobs), Users, Organizations, Teams, Releases, Code Scanning Alerts, Dependabot Alerts
- **Data profile:**
  - Write pattern: Mutable (issues, PRs get updated) + append-only (commits, events, audit logs)
  - Volume: HIGH - large orgs have millions of issues, PRs, commits, events
  - Real-time: Webhooks (not WebSocket/SSE) - REST polling for tail
  - Search need: HIGH - finding issues, PRs, commits across repos is a core use case

## API Spec Discovery

- **Official OpenAPI spec:** `github/rest-api-description` repository
- **URL:** `https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json`
- **Version:** OpenAPI 3.0.3 (stable, bundled)
- **Scale:** 740 paths, 1,107 operations
- **Top resource categories by endpoint count:**
  - repos: 201, actions: 184, orgs: 108, issues: 55, codespaces: 48
  - users: 47, apps: 37, activity: 32, teams: 32, packages: 27, pulls: 27

## Usage Patterns (Top 5 by Evidence)

### 1. PR Triage & Review Management (Evidence: 9/10)
Users need to see all open PRs across repos, prioritize by staleness/size/author, and batch-process reviews. gh-dash (11.2k stars) exists specifically for this. Multiple gh CLI issues request cross-repo PR listing (#5317, #642).
- **Sources:** gh-dash 11.2k stars, gh CLI issues, Reddit/HN discussions, gh-pr-dashboard HN post

### 2. Issue & PR Export/Archive to Local Storage (Evidence: 8/10)
Users want to backup, export, or archive issues/PRs as structured data (SQLite, CSV, Markdown). github-to-sqlite (462 stars), gh2md, export-pull-requests, python-github-backup all serve this need.
- **Sources:** github-to-sqlite 462 stars, gh2md, python-github-backup, multiple backup tools

### 3. CI/CD Monitoring & Actions Analytics (Evidence: 7/10)
Engineering teams need to monitor workflow runs, identify flaky tests, track build times, and alert on failures. github-actions-watcher by Spatie exists. Multiple burndown/velocity tools exist.
- **Sources:** github-actions-watcher, BuildBeacon HN post, burndown-for-github-projects

### 4. Stale Issue/PR Detection & Hygiene (Evidence: 7/10)
Maintainers need to find stale issues (no updates in N days), orphaned PRs, issues without labels, and PRs without reviewers. This is a recurring manual task that existing tools partially address.
- **Sources:** GitHub community discussions, burndown tools, issue triage workflows

### 5. Cross-Repo Search & Analytics (Evidence: 6/10)
Users want to search issues, PRs, and code across all their repos or an org's repos from the terminal. The gh CLI search is limited. github-to-sqlite + Datasette provides this via SQL.
- **Sources:** github-to-sqlite + Datasette demo, gh-search-cli, Stack Overflow questions

## Tool Landscape (Beyond API Wrappers)

| Tool | Stars | Type | What It Does |
|------|-------|------|-------------|
| **gh** (cli/cli) | 43,400 | Workflow Tool | Official CLI - PRs, issues, actions, interactive workflows |
| **gh-dash** | 11,200 | Data Tool (TUI) | Terminal dashboard for PRs and issues with vim keybindings |
| **github-to-sqlite** | 462 | Data Tool | Sync GitHub data to SQLite for Datasette/SQL queries |
| **gh2md** | ~200 | Data Tool | Export issues/PRs to Markdown files |
| **python-github-backup** | ~800 | Data Tool | Full repo backup including issues, PRs, comments, wikis |
| **export-pull-requests** | ~300 | Data Tool | Export PRs/issues to CSV (GitHub, GitLab, Bitbucket) |
| **ghexport** | ~100 | Data Tool | Export personal GitHub activity data |
| **github-actions-watcher** | ~300 | Monitoring Tool | Real-time Actions workflow status in terminal |
| **burndown-for-github-projects** | ~200 | Analytics Tool | Sprint burndown charts from GitHub Projects |

**Key insight:** The Data Tool category is fragmented - github-to-sqlite does sync+SQL but is Python/Datasette-dependent, not a standalone CLI. No single Go CLI combines sync + search + analytics + workflows.

## Workflows

### 1. PR Triage Report
**Steps:** List all open PRs across N repos -> sort by age/size/review status -> group by reviewer -> output as table or JSON
**Frequency:** Daily for engineering managers
**Pain:** gh CLI can only list PRs per-repo, no cross-repo view
**Proposed:** `github-cli pr-triage --org myorg --stale 7 --json`

### 2. Stale Issue Sweep
**Steps:** Fetch issues across repos -> filter by last-updated -> exclude labeled issues -> optionally comment/close
**Frequency:** Weekly for maintainers
**Pain:** Manual process, no single command
**Proposed:** `github-cli stale --org myorg --days 30 --label needs-triage`

### 3. Actions Health Report
**Steps:** Fetch recent workflow runs -> compute success rate per workflow -> identify flaky tests -> track build duration trends
**Frequency:** Weekly for DevOps
**Pain:** GitHub Actions UI is per-repo, no aggregate view
**Proposed:** `github-cli actions-health --repo myorg/myrepo --days 14`

### 4. Release Changelog Generation
**Steps:** Get commits since last release -> group by type (feat/fix/refactor) -> include PR links -> format as Markdown
**Frequency:** Per release
**Pain:** Manual commit log parsing
**Proposed:** `github-cli changelog --since v1.2.0`

### 5. Contributor Leaderboard
**Steps:** Fetch commits + PRs + reviews per contributor -> score by activity -> rank -> output table
**Frequency:** Monthly for team leads
**Pain:** No single view of contributor activity
**Proposed:** `github-cli contributors --org myorg --days 30 --sort commits`

## Architecture Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **Persistence** | SQLite with domain-specific tables | HIGH volume + HIGH search need. Issues, PRs, commits are primary entities. github-to-sqlite proves the model works (462 stars). |
| **Real-time** | REST polling with `since` cursor | GitHub has no WebSocket/SSE for REST API. Webhooks require a server. REST polling with `?since=` or `?updated_after=` is the only CLI-friendly option. |
| **Search** | FTS5 on issue/PR titles, bodies, comments | Users need to find issues by keyword across repos. FTS5 enables instant offline search. |
| **Bulk** | Paginated sync with `per_page=100` + conditional requests (ETag/If-Modified-Since) | GitHub rate limit is 5,000/hour. Conditional requests don't count. Sync must be rate-limit-aware. |
| **Cache** | SQLite IS the cache - no separate cache layer | The local database serves as both persistent storage and cache. `--no-cache` flag bypasses local DB and hits API directly. |

## Top 5 Features for the World

### 1. Offline Cross-Repo Search (Score: 14/16)
| Dimension | Score | Justification |
|-----------|-------|---------------|
| Evidence | 3 | github-to-sqlite (462 stars) + Datasette proves demand |
| User impact | 3 | Every maintainer searches across repos daily |
| Feasibility | 2 | SQLite + FTS5, proven pattern |
| Uniqueness | 2 | No Go CLI does this - github-to-sqlite is Python + Datasette |
| Composability | 2 | `github-cli search "bug" --repo org/* --json | jq` |
| Data profile fit | 2 | Perfect - high volume text data |
| Maintainability | 0 | Needs sync maintenance |
| Competitive moat | 0 | Concept proven, execution differentiates |

### 2. PR Triage Dashboard (Score: 13/16)
| Dimension | Score | Justification |
|-----------|-------|---------------|
| Evidence | 3 | gh-dash (11.2k stars), multiple gh CLI feature requests |
| User impact | 3 | Engineering managers check daily |
| Feasibility | 2 | Query local DB or live API |
| Uniqueness | 1 | gh-dash exists but is TUI-only, not agent-native |
| Composability | 2 | JSON output for agent consumption |
| Data profile fit | 2 | Cross-repo PR data in SQLite |
| Maintainability | 0 | Relies on API stability |
| Competitive moat | 0 | gh-dash is strong competition |

### 3. Actions Health Analytics (Score: 12/16)
| Dimension | Score | Justification |
|-----------|-------|---------------|
| Evidence | 2 | github-actions-watcher (300 stars), BuildBeacon |
| User impact | 3 | DevOps teams need this weekly |
| Feasibility | 2 | Workflow runs API is well-documented |
| Uniqueness | 2 | No CLI does aggregate actions analytics |
| Composability | 2 | `github-cli actions-health --json` for monitoring |
| Data profile fit | 1 | Workflow runs are append-only, moderate volume |
| Maintainability | 0 | Actions API evolves |
| Competitive moat | 0 | Straightforward to build |

### 4. Stale Issue/PR Sweep (Score: 11/16)
| Dimension | Score | Justification |
|-----------|-------|---------------|
| Evidence | 2 | Multiple triage tools, community discussions |
| User impact | 2 | Maintainers do this weekly |
| Feasibility | 2 | Simple date filtering |
| Uniqueness | 2 | No CLI combines detection + optional action |
| Composability | 2 | Pipe to close/label commands |
| Data profile fit | 1 | Queries local DB for speed |
| Maintainability | 0 | Stable pattern |
| Competitive moat | 0 | Simple concept |

### 5. Raw SQL Access to GitHub Data (Score: 11/16)
| Dimension | Score | Justification |
|-----------|-------|---------------|
| Evidence | 3 | github-to-sqlite + Datasette (462 stars) proves the model |
| User impact | 2 | Power users and data analysts |
| Feasibility | 2 | SQLite is the persistence layer already |
| Uniqueness | 1 | github-to-sqlite does this via Datasette |
| Composability | 2 | SQL is the ultimate composability |
| Data profile fit | 1 | Perfect for structured queries |
| Maintainability | 0 | Schema must match API |
| Competitive moat | 0 | Datasette is powerful competition |

## Acceptance Criteria
- [x] API Identity documented with data profile
- [x] 5 usage patterns with evidence scores >= 6
- [x] Tool landscape includes non-wrapper tools (gh-dash, github-to-sqlite, gh2md, etc.)
- [x] 5 workflows with proposed CLI features
- [x] Architecture decisions match data profile (SQLite + FTS5 for high-volume search)
- [x] Top 5 features scored and ranked

## Sources
- https://github.com/github/rest-api-description - Official OpenAPI spec (740 paths, 1107 operations)
- https://github.com/cli/cli - Official gh CLI (43.4k stars)
- https://github.com/dlvhdr/gh-dash - Terminal dashboard (11.2k stars)
- https://github.com/dogsheep/github-to-sqlite - SQLite sync tool (462 stars)
- https://github.com/mattduck/gh2md - Markdown export
- https://github.com/josegonzalez/python-github-backup - Full backup tool
- https://github.com/sshaw/export-pull-requests - CSV export
- https://github.com/spatie/github-actions-watcher - Actions monitor
- https://github.com/cli/cli/issues/5317 - Cross-repo issue listing request
- https://github.com/cli/cli/issues/3370 - PR update command request
- https://docs.github.com/en/rest - Official REST API docs
