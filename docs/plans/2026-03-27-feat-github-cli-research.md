---
title: "Research: GitHub CLI"
type: feat
status: active
date: 2026-03-27
phase: "1"
api: "github"
---

# Research: GitHub CLI

## Spec Discovery
- **Official OpenAPI spec:** https://github.com/github/rest-api-description
- **Direct URL:** `https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json`
- **Format:** OpenAPI 3.0.3 (stable, bundled JSON)
- **Endpoint count:** 740 paths, 1,107 operations
- **Status:** Maintained by GitHub, auto-generated, does not accept direct PRs

## Competitors (Deep Analysis)

### gh (cli/cli) - 43,400 stars
- **Repo:** https://github.com/cli/cli
- **Language:** Go (99.2%)
- **Commands:** ~80+ subcommands across pr, issue, repo, run, release, codespace, project, etc.
- **Last commit:** Active (trunk branch, 11,038 commits, 601 contributors)
- **Open issues:** 917
- **Maintained:** YES - official GitHub product
- **Notable features:** Interactive PR creation, issue filing, Actions monitoring, codespace management, `gh api` for raw API access, extension system
- **Weaknesses:**
  - No local persistence or offline capabilities
  - No cross-repo aggregation (can only list PRs/issues per-repo)
  - No analytics or trend detection
  - No bulk data export
  - Users request cross-repo PR listing (#5317), PR update command (#3370), complete comment display (#5788)

### gh-dash (dlvhdr/gh-dash) - 11,200 stars
- **Repo:** https://github.com/dlvhdr/gh-dash
- **Language:** Go (100%)
- **Commands:** TUI application (not a CLI with subcommands)
- **Last commit:** Active (582 commits)
- **Open issues:** 69
- **Maintained:** YES
- **Notable features:** Vim-style keybindings, YAML config, per-repo PR/issue sections, custom actions
- **Weaknesses:**
  - TUI-only (no --json output, not agent-native)
  - No local persistence
  - No offline search
  - No analytics
  - Users request: create PRs from dash (#689), repo-specific view (#179), keybinding customization (#214)

### github-to-sqlite (dogsheep/github-to-sqlite) - 462 stars
- **Repo:** https://github.com/dogsheep/github-to-sqlite
- **Language:** Python
- **Commands:** 12+ commands (issues, pull-requests, commits, releases, starred, repos, etc.)
- **Last commit:** December 2023 (v2.9)
- **Open issues:** 20
- **Maintained:** Slow/dormant (last release Dec 2023)
- **Notable features:** SQLite persistence, Datasette integration, covers most GitHub entities, `get` command for raw API with pagination
- **Weaknesses:**
  - Python dependency (not a standalone binary)
  - Requires Datasette for querying (not self-contained)
  - No FTS5 search built-in
  - No workflow/analytics commands
  - No incremental sync optimization
  - Dormant development

## User Pain Points

> "gh should have a way to list all open issues for the repositories in my account / organization" - [cli/cli#5317](https://github.com/cli/cli/issues/5317)

> Users want "a feature request for showing issues and PRs specific to the repo from which the dash command is run" - [gh-dash#179](https://github.com/dlvhdr/gh-dash/issues/179)

> "While downloading a PR locally via gh pr checkout is simple, there is no command for pushing changes back, particularly for PRs from forks" - [cli/cli#3370](https://github.com/cli/cli/issues/3370)

> "gh pr view --comments should show all comments" including inline review comments - [cli/cli#5788](https://github.com/cli/cli/issues/5788)

> Rate limiting is a major pain: "ETags are per-page, not per collection. If you get a 304 on page 1 of 5, that doesn't mean pages 2-5 are unchanged" - [GitHub Community Discussion](https://github.com/orgs/community/discussions/163553)

## Auth Method
- **Type:** Personal access token (bearer token via Authorization header)
- **Env var convention:** `GITHUB_TOKEN` (used by gh, github-to-sqlite, and most tools)
- **Rate limit:** 5,000 requests/hour (authenticated), 60/hour (unauthenticated)
- **Conditional requests:** ETag/If-Modified-Since headers - 304 responses don't count against limit

## Demand Signals
- gh-dash at 11.2k stars proves massive demand for PR/issue dashboards
- github-to-sqlite at 462 stars proves demand for local GitHub data in SQLite
- Multiple HN Show HN posts about GitHub analytics/monitoring tools
- 917 open issues on gh CLI = active user demand for features gh doesn't have
- "Ask HN: What's the one feature you'd want in a GitHub productivity tool?" - [HN thread](https://news.ycombinator.com/item?id=42287526)

## Strategic Justification

**Why this CLI should exist when gh has 43,400 stars:**

gh is a workflow tool - it helps you create PRs, file issues, and run Actions. It has zero local persistence, zero offline capability, and zero analytics. It is NOT a data tool.

github-to-sqlite is a data tool but it's Python-dependent, requires Datasette for queries, has been dormant since Dec 2023, has no FTS5 search, no workflow commands, and no agent-native output modes.

gh-dash is a beautiful TUI but has no CLI interface (no --json), no local persistence, and no analytics.

**Our CLI fills the gap between all three:**
1. **Go binary** (no Python/Node dependency) like gh
2. **SQLite persistence** with domain-specific tables like github-to-sqlite
3. **FTS5 search** that github-to-sqlite doesn't have
4. **Agent-native** (--json, --select, --dry-run) that gh-dash doesn't have
5. **Workflow commands** (pr-triage, stale, actions-health, changelog, security) that none of them have
6. **Incremental sync** with rate-limit-aware cursors
7. **Raw SQL access** without needing Datasette

This is the "discrawl for GitHub" - a standalone Go binary that syncs GitHub data to SQLite and provides compound workflow commands on top.

## Target
- **Command count:** 40-50 (not 1,107 - depth beats breadth)
  - ~15-20 core API commands (repos, issues, PRs, commits, releases, actions)
  - 6 data layer commands (sync, search, sql, issues, prs, tail)
  - 7 workflow commands (pr-triage, stale, actions-health, changelog, security, activity, contributors)
  - 5-8 utility commands (doctor, auth, config, version, completion)
- **Key differentiator:** Local SQLite + FTS5 + compound workflow commands
- **Quality bar:** Grade A (80+/100)

## Acceptance Criteria
- [x] Research artifact with Spec Discovery section
- [x] 3 competitors analyzed with maintenance status
- [x] 5+ user quotes/pain points documented
- [x] Strategic justification answers "why should this exist?"
- [x] Target command count set (40-50)

## Sources
- https://github.com/github/rest-api-description - OpenAPI spec
- https://github.com/cli/cli - gh CLI (43.4k stars)
- https://github.com/dlvhdr/gh-dash - TUI dashboard (11.2k stars)
- https://github.com/dogsheep/github-to-sqlite - SQLite sync (462 stars)
- https://github.com/cli/cli/issues/5317 - Cross-repo listing request
- https://github.com/cli/cli/issues/3370 - PR update request
- https://github.com/cli/cli/issues/5788 - PR comments request
- https://github.com/orgs/community/discussions/163553 - Rate limit pain
- https://news.ycombinator.com/item?id=42287526 - HN productivity tool ask
