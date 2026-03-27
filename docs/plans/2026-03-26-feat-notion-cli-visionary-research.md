---
title: "Visionary Research: Notion CLI"
type: feat
status: active
date: 2026-03-26
phase: "0"
api: "Notion"
---

# Visionary Research: Notion CLI

## Overview

Notion is a productivity/content management platform used by millions as a knowledge base, project tracker, and wiki. Its API (v2026-03-11) exposes ~55 REST endpoints covering Pages, Databases, Blocks, Users, Comments, Views, Data Sources, File Uploads, Search, and OAuth. The API is bearer-token authenticated with a strict 3 req/sec rate limit, 500KB payload cap, and 2000-character rich text limit. Critically, Notion has NO webhooks and NO websockets - all change detection requires polling.

This makes Notion a prime candidate for a CLI with local persistence: users need to search their workspace content offline, sync incrementally to avoid rate limit pain, and run cross-database queries that the web UI can't express. The existing CLI landscape is dominated by 4ier/notion-cli (90 stars, 39 commands) - a well-built API wrapper that lacks local persistence, offline search, and workflow commands. Our strategic angle: build the discrawl-class data tool for Notion, not just another API wrapper.

## API Identity

- **Domain:** Productivity / Content Management
- **Primary users:** Developers building integrations, knowledge workers automating content, bot/AI agent developers
- **Core entities:** Pages, Databases, Blocks, Users, Comments, Views, Data Sources, File Uploads
- **Endpoint count:** ~55 REST endpoints across 10 resource groups
- **Auth:** Bearer token (integration token) + OAuth 2.0 with refresh/introspect/revoke
- **Rate limit:** 3 req/sec average, HTTP 429 with Retry-After header
- **Payload limits:** 500KB max, 1000 blocks per request, 100 items per array, 2000 char rich text

### Data Profile

| Dimension | Value | Notes |
|-----------|-------|-------|
| Write pattern | Mutable | Pages/blocks can be created, updated, deleted |
| Volume | Medium-High | Workspaces with thousands of pages, millions of blocks |
| Real-time | NONE | No webhooks, no websockets, no SSE |
| Search need | HIGH | Notion IS a knowledge base - search is the core use case |
| Temporal | Yes | Pages have created_time, last_edited_time |

## Usage Patterns (Top 5 by Evidence)

### 1. Workspace Backup & Export (Evidence: 10/10)

| Source | Points |
|--------|--------|
| jckleiner/notion-backup (136 stars, archived) | 3 |
| upleveled/notion-backup (77 stars) | 2 |
| HN Show: "I built a tool to back up Notion workspaces" (43517524) | 2 |
| darobin/notion-backup, ivanik7/notion-backup, notion4ever | 2 |
| Cross-platform (GitHub + HN) | +2 |
| **Total** | **11** |

Users desperately want automated, incremental backups. The top backup tool (136 stars) is archived because Notion broke its auth. No existing tool does incremental sync well.

### 2. Database Query & Filter (Evidence: 8/10)

| Source | Points |
|--------|--------|
| 4ier/notion-cli (90 stars) - primary feature | 3 |
| Multiple automation scripts on GitHub (forrest-herman, zackrylangford, sinodine) | 2 |
| HN comment: "game-changer for power users" re: database queries | 1 |
| Cross-platform (GitHub repos + HN) | +2 |
| **Total** | **8** |

Database querying is the #1 developer use case. 4ier's CLI does this well with human-readable filters, but results can't be stored or searched offline.

### 3. Content Sync (Notion <-> Markdown/Obsidian) (Evidence: 7/10)

| Source | Points |
|--------|--------|
| notion2md (converter, popular) | 2 |
| Obsidian-Sync-to-Notion (preserves folder structure) | 2 |
| Mk Notes (markdown sync) | 1 |
| notion-github-sync (public pages to GitHub) | 1 |
| Blog-from-Notion pattern (multiple tools) | 1 |
| **Total** | **7** |

Bidirectional markdown sync is a huge pain point. Notion's new `/pages/{id}/markdown` endpoints (GET + PATCH) make this feasible for the first time.

### 4. Recurring Task Automation (Evidence: 6/10)

| Source | Points |
|--------|--------|
| Multiple GitHub repos (Notion-RepetitiveTask, notion-auto-archive) | 2 |
| pocketvince/Automate-Notion.so-with-shell-script (crontab) | 1 |
| n8n community threads on Notion automation | 1 |
| Cross-platform (GitHub + n8n community) | +2 |
| **Total** | **6** |

Users build cron-based scripts to manage recurring tasks, archive completed items, and sync between databases.

### 5. Cross-Database Analytics (Evidence: 6/10)

| Source | Points |
|--------|--------|
| victoriano/notion_data (CSV export for analytics) | 2 |
| notion_analytics (template analytics) | 1 |
| ivynya/analytics (Notion-integrated KPI tracking) | 1 |
| Cross-platform appearance | +2 |
| **Total** | **6** |

No existing tool lets you JOIN across Notion databases or run aggregate queries. This is a unique opportunity for a SQLite-backed CLI.

## Tool Landscape (Beyond API Wrappers)

### API Wrappers
| Tool | Stars | Language | Commands | Status |
|------|-------|----------|----------|--------|
| 4ier/notion-cli | 90 | Go | 39 | Active (Feb 2026) |
| litencatt/notion-cli | 25 | TypeScript | 6 | Active (Nov 2025) |
| lox/notion-cli | 17 | Go | 20+ | Active (Mar 2026, MCP-based) |

### Data Tools
| Tool | Stars | What It Does |
|------|-------|-------------|
| notion_data | - | Export databases to CSV/Pandas for analytics |
| notion4ever | - | Full export with nested subpages to markdown/HTML |
| notion2md | - | Convert Notion blocks to markdown |

### Workflow Tools
| Tool | Stars | What It Does |
|------|-------|-------------|
| notion-auto-archive | - | Auto-archive completed tasks on boards |
| Notion-RepetitiveTask | - | Manage recurring tasks with date cycling |
| notion-github-sync | - | Sync Notion pages to GitHub Discussions |

### Backup Tools
| Tool | Stars | What It Does | Status |
|------|-------|-------------|--------|
| jckleiner/notion-backup | 136 | Multi-cloud backup | ARCHIVED (Dec 2025) |
| upleveled/notion-backup | 77 | GitHub Actions scheduled export | Active |
| darobin/notion-backup | - | Simple workspace export | Active |
| HermanSchoenfeld/LocalNotion | 0 | Offline mirror + git version control | Active |

## Workflows

### 1. Incremental Workspace Sync
**Steps:** List all pages -> filter by last_edited_time -> fetch changed blocks -> store locally
**Frequency:** Daily (automated)
**Pain point:** Rate limit of 3 req/sec means full sync of large workspace takes hours
**Proposed:** `notion-cli sync --workspace --since "2 days ago"` with SQLite persistence

### 2. Cross-Database Report
**Steps:** Query DB A -> Query DB B -> Join on relation properties -> Aggregate
**Frequency:** Weekly
**Pain point:** Impossible in Notion UI; requires custom scripts
**Proposed:** `notion-cli sql "SELECT a.title, COUNT(b.id) FROM tasks a JOIN subtasks b ON ..."` against local SQLite

### 3. Markdown Round-Trip
**Steps:** Export page to markdown -> Edit locally -> Push changes back
**Frequency:** Daily for blog/docs workflows
**Pain point:** Manual export/import, formatting loss
**Proposed:** `notion-cli pages pull <id> --md` / `notion-cli pages push <id> --md ./file.md`

### 4. Stale Page Detection
**Steps:** Scan all pages -> Filter by last_edited_time > N days -> Group by parent database
**Frequency:** Monthly
**Pain point:** No way to find stale content in Notion UI
**Proposed:** `notion-cli stale --days 90 --json` against local SQLite

### 5. Database Diff
**Steps:** Query database -> Compare with previous snapshot -> Show added/removed/changed entries
**Frequency:** Weekly
**Pain point:** No change tracking in Notion for database entries
**Proposed:** `notion-cli db diff <database-id> --since "1 week ago"`

## Architecture Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **Persistence** | SQLite + FTS5 | HIGH search need (knowledge base), medium-high volume, enables offline queries and cross-DB joins |
| **Real-time** | REST polling with `last_edited_time` cursor | No webhooks/websockets/SSE available. Must use last_edited_time as sync cursor |
| **Search** | FTS5 on page titles + block content + DB entry properties | Notion's built-in search is title-only via API. Local FTS5 enables full-text content search |
| **Bulk** | Incremental sync with 3 req/sec rate limiter + exponential backoff | Rate limit is the #1 constraint. Incremental sync minimizes API calls |
| **Cache** | SQLite is the cache. `--no-cache` flag bypasses for live API calls | Local DB serves as both cache and queryable datastore |
| **Markdown** | Use new `/pages/{id}/markdown` GET/PATCH endpoints | First-class markdown support avoids lossy block-to-markdown conversion |

## Top 5 Features for the World

### 1. Local SQLite Sync with Full-Text Search (Score: 15/16)

| Dimension | Score | Reason |
|-----------|-------|--------|
| Evidence strength | 3 | 136-star backup tool + HN demand + multiple sync tools |
| User impact | 3 | Every Notion power user needs offline search |
| Implementation feasibility | 2 | Incremental sync via last_edited_time is proven |
| Uniqueness | 2 | NO existing CLI has local SQLite + FTS5 |
| Composability | 2 | `notion-cli sql` pipes to any tool |
| Data profile fit | 2 | Perfect - high search need + medium-high volume |
| Maintainability | 1 | SQLite schema is stable |
| Competitive moat | 1 | Hard to replicate without Go + SQLite embedding |
| **Total** | **15** | **Must-have** |

### 2. Markdown Round-Trip (Pull/Push) (Score: 13/16)

| Dimension | Score | Reason |
|-----------|-------|--------|
| Evidence strength | 3 | notion2md, Obsidian-Sync, Mk Notes, blog workflows |
| User impact | 3 | Developers want to edit in their editor, not Notion |
| Implementation feasibility | 2 | New markdown API endpoints make this trivial |
| Uniqueness | 1 | notion2md exists but is one-way only |
| Composability | 2 | Markdown files are universally composable |
| Data profile fit | 2 | Content management = markdown is natural |
| Maintainability | 0 | Markdown API is new, may change |
| Competitive moat | 0 | Easy to replicate |
| **Total** | **13** | **Must-have** |

### 3. Cross-Database SQL Queries (Score: 12/16)

| Dimension | Score | Reason |
|-----------|-------|--------|
| Evidence strength | 2 | notion_data (CSV for analytics), analytics tools |
| User impact | 3 | Impossible in Notion UI, huge unlock |
| Implementation feasibility | 2 | SQLite + synced data = just expose `sql` command |
| Uniqueness | 2 | NO tool does this |
| Composability | 2 | SQL output -> JSON -> any pipeline |
| Data profile fit | 2 | Multiple databases = need for joins |
| Maintainability | 0 | Requires keeping schema in sync |
| Competitive moat | 1 | Requires local persistence layer |
| **Total** | **12** | **Must-have** |

### 4. Incremental Backup with Git Integration (Score: 11/16)

| Dimension | Score | Reason |
|-----------|-------|--------|
| Evidence strength | 3 | 136-star tool archived, 77-star tool active, HN demand |
| User impact | 2 | Important but niche (not every user backs up) |
| Implementation feasibility | 1 | Git integration needs careful design |
| Uniqueness | 1 | upleveled/notion-backup does this (but TypeScript, GitHub Actions only) |
| Composability | 2 | Git repos are universally composable |
| Data profile fit | 2 | Mutable content needs versioning |
| Maintainability | 0 | Git integration complexity |
| Competitive moat | 0 | Multiple backup tools exist |
| **Total** | **11** | **Should-have** |

### 5. Stale Content Detection & Hygiene (Score: 10/16)

| Dimension | Score | Reason |
|-----------|-------|--------|
| Evidence strength | 1 | notion-auto-archive exists, but no dedicated tool |
| User impact | 2 | Knowledge bases rot without hygiene |
| Implementation feasibility | 2 | Simple query on last_edited_time |
| Uniqueness | 2 | No existing tool does this |
| Composability | 2 | JSON output for automation |
| Data profile fit | 2 | Temporal data enables this perfectly |
| Maintainability | 0 | Straightforward |
| Competitive moat | 0 | Trivial to build once you have sync |
| **Total** | **10** | **Should-have** |

## Demand Signals

- **HN Show: "Notion-CLI - Full Notion API from the terminal, 39 commands"** - Recent post about 4ier/notion-cli. Commenter called it a "game-changer for power users" and requested fzf integration.
- **HN Show: "A Notion CLI for Agents"** - lox/notion-cli with MCP integration. Shows demand for agent-native Notion access.
- **HN Show: "I built a tool to back up Notion workspaces"** - Backup tool creator confirmed "there was no native backup solution" and customers wanted restore capability limited by Notion's API.
- **Multiple community OpenAPI spec requests** - At least 4 Latenode community posts asking for Notion's OpenAPI spec, showing developer demand for tooling.
- **136-star backup tool archived** - jckleiner/notion-backup died because Notion broke their auth method. Market opportunity for a maintained alternative.

## Sources

- https://developers.notion.com/reference/intro
- https://developers.notion.com/llms.txt
- https://developers.notion.com/reference/request-limits
- https://github.com/4ier/notion-cli (90 stars)
- https://github.com/litencatt/notion-cli (25 stars)
- https://github.com/lox/notion-cli (17 stars)
- https://github.com/jckleiner/notion-backup (136 stars, archived)
- https://github.com/upleveled/notion-backup (77 stars)
- https://github.com/spencerpauly/awesome-notion
- https://news.ycombinator.com/item?id=47133849
- https://news.ycombinator.com/item?id=46875374
- https://news.ycombinator.com/item?id=43517524
