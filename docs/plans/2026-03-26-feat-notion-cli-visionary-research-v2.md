---
title: "Visionary Research: Notion CLI"
type: feat
status: active
date: 2026-03-26
phase: "0"
api: "notion"
---

# Visionary Research: Notion CLI

## Overview

Notion is a content-first productivity platform used by millions of developers, teams, and individuals for wikis, project management, and knowledge bases. Its API exposes pages, databases (now "data sources"), blocks, users, comments, search, file uploads, and views. The API is REST-based with bearer token auth and cursor-based pagination.

The Notion API is uniquely positioned for a data-layer CLI because it manages structured content (databases/data sources) alongside unstructured content (pages/blocks). Users don't just need CRUD - they need backup, offline search, content migration, and workspace hygiene workflows.

## API Identity

- **Domain:** Content/Productivity (hybrid: structured databases + unstructured documents)
- **Primary users:** Developers automating workspace management, teams syncing Notion to other tools, individuals backing up content
- **Core entities:** Pages, Databases/Data Sources, Blocks, Users, Comments, Views, File Uploads
- **Data profile:**
  - Write pattern: Mutable (pages/blocks updated frequently), append for comments
  - Volume: Medium-high (active workspaces have thousands of pages, tens of thousands of blocks)
  - Real-time: No native WebSocket/SSE. Webhooks available via database automations (limited). No gateway.
  - Search need: HIGH - users constantly search for pages/content. API search is title-only, no full-text.

## Usage Patterns (Top 5 by Evidence)

| Rank | Pattern | Evidence Score | What It Needs |
|------|---------|---------------|---------------|
| 1 | **Workspace backup/export** | 10/10 | Show HN post, 5+ backup tools on GitHub (jckleiner, darobin, upleveled, nikhilbadyal, LocalNotion), HN discussion about paid backup SaaS, cross-platform appearance | Incremental sync, multiple export formats (MD, HTML, JSON) |
| 2 | **Database automation/querying** | 8/10 | 10+ automation repos (forrest-herman, zackrylangford, sinodine, adhirajpandey), n8n/Zapier/Make integrations, Thomas Frank guides | Database query from CLI, filter/sort, bulk updates |
| 3 | **Content publishing pipeline** | 7/10 | notion2md, notion-exporter, Notion-as-CMS pattern (astro-notion-blog, Next.js blogs), multiple HN posts about CMS workflow | Page-to-markdown export, recursive block traversal |
| 4 | **Recurring task management** | 5/10 | danhenrik/Notion-RepetitiveTask, kris-hansen/notion-cli (Taskbook-style), automation scripts | Page/database item creation, template instantiation, status updates |
| 5 | **Workspace analytics/hygiene** | 4/10 | notion_analytics tool, notion_data CSV export, team sync templates | Database queries with aggregation, stale page detection, orphan finding |

## Tool Landscape (Beyond API Wrappers)

### Tier 1: Direct CLI Tools
| Tool | Stars | Lang | Type | Notes |
|------|-------|------|------|-------|
| 4ier/notion-cli | 91 | Go | API Wrapper | 39 commands, actively maintained (Feb 2026), schema-aware filters, markdown R/W |
| litencatt/notion-cli | 25 | TypeScript | API Wrapper | 6 commands, interactive mode, multi-format output (CSV, JSON, YAML) |
| kris-hansen/notion-cli | ~10 | Python | Task Tool | Taskbook-style, ToDo-focused |
| MrRichRobinson/notion-cli | 0 | Python | API Wrapper | Full resource coverage, 2025 API model |
| lox/notion-cli | ~5 | Go | MCP Bridge | Uses Notion MCP server with OAuth |

### Tier 2: Data/Workflow Tools
| Tool | Stars | Type | What It Does |
|------|-------|------|-------------|
| HermanSchoenfeld/LocalNotion | 0 | Data Tool | Offline mirror with Git version control + HTML export, 30s sync |
| jckleiner/notion-backup | ~200+ | Data Tool | Auto-backup to GDrive/Dropbox/pCloud/Nextcloud/local |
| upleveled/notion-backup | ~50 | Data Tool | Export pages to GitHub on schedule |
| victoriano/notion_data | ~20 | Data Tool | Convert databases to CSV/Pandas DataFrames |
| notion-helper | ~50 | Integration Tool | Reduces Notion API request boilerplate |

### Tier 3: Ecosystem
- awesome-notion (spencerpauly): 200+ stars, curated list with 50+ tools
- notion-sdk-py: Official Python SDK
- notion-sdk-js: Official JS SDK
- notionapi (Go): Unofficial Go SDK

## Workflows

| # | Name | Steps | Pain Point | Proposed CLI Feature |
|---|------|-------|-----------|---------------------|
| 1 | Full workspace backup | List all pages -> fetch each with blocks -> export to MD/HTML/JSON -> store locally with structure | Manual, requires pagination, block recursion, rate limit handling | `notion-cli backup --format md --output ./backup` |
| 2 | Database query + export | Query database with filters -> format results -> pipe to CSV/JSON | JSON filter syntax is painful, 2k char limit per request | `notion-cli db query <id> --filter "Status=Done" --format csv` |
| 3 | Stale page detection | Search all pages -> check last_edited_time -> report pages untouched for N days | No built-in way to find orphaned/stale content | `notion-cli stale --days 30 --workspace` |
| 4 | Page content diff | Fetch page blocks -> convert to text -> diff against local copy | No version history API, manual comparison | `notion-cli diff <page-id> ./local-copy.md` |
| 5 | Bulk page creation from template | Read template page -> create N pages with variable substitution | Repetitive, rate-limited, template API is new | `notion-cli bulk-create --template <id> --data ./items.csv` |

## Architecture Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **Persistence** | SQLite with domain-specific tables | High search need (API search is title-only, no full-text). FTS5 on page titles + block content enables offline full-text search that beats the API. Data gravity is high for pages and blocks. |
| **Real-time** | REST polling with cursor | No WebSocket/SSE/Gateway available. Use `last_edited_time` as sync cursor for incremental updates. |
| **Search** | FTS5 on pages, blocks, database items | Notion's search API only searches titles. Local FTS5 on block content is a killer feature no competitor has. |
| **Bulk** | Chunked operations with rate limit backoff | API rate limit is 3 req/sec average. Batch operations must respect this with exponential backoff. |
| **Cache** | SQLite IS the cache | No separate cache layer needed. Sync populates SQLite, queries read from it. `--no-cache` forces live API calls. |
| **Export** | Markdown as primary format | Block-to-markdown is the most requested pattern. Also support JSON (raw) and HTML (rendered). |

## Top 5 Features for the World

| Rank | Feature | Score | Description |
|------|---------|-------|-------------|
| 1 | **Offline full-text search** | 14/16 | Sync workspace to SQLite with FTS5. Search page titles AND block content locally. No API rate limits, instant results. Nothing like this exists. Evidence: 3=existing demand for search tools, 3=most users want this, 2=feasible with sync+FTS5, 2=no existing tool, 2=great for pipes, 2=perfect data fit, 0=needs maintenance, 0=replicable |
| 2 | **Incremental workspace backup** | 13/16 | Sync pages/blocks to local SQLite, export to Markdown/HTML/JSON. Incremental via last_edited_time cursor. Evidence: 3=5+ backup tools exist, 3=universal pain point, 2=feasible, 1=improves on existing, 2=composable, 2=perfect fit, 0=maintenance, 0=replicable |
| 3 | **Human-friendly database queries** | 11/16 | Query databases without writing JSON filters. `--filter "Status=Done,Priority=High"`. Evidence: 2=Reddit demand, 3=most users hit this, 2=feasible, 1=4ier has basic version, 1=somewhat composable, 2=good fit, 0=maintenance, 0=replicable |
| 4 | **Stale content detector** | 10/16 | Find pages not edited in N days, orphaned pages, empty databases. Evidence: 1=some demand, 2=niche but painful, 2=feasible with sync data, 2=no existing tool, 2=great for pipes, 1=possible, 0=maintenance, 0=replicable |
| 5 | **Page-to-Markdown pipeline** | 9/16 | Export any page to clean Markdown with recursive block traversal and image handling. Evidence: 2=notion2md exists, 2=CMS use case, 2=feasible, 0=well-served, 1=somewhat composable, 2=good fit, 0=maintenance, 0=replicable |

## Key Insight

The #1 competitor (4ier/notion-cli, 91 stars) is a solid API wrapper but has NO local persistence, NO full-text search, and NO workspace-level workflows. It's "gh for Notion" but not "discrawl for Notion." The opportunity is clear: build the data layer that turns Notion API access into a workspace management tool.

## Sources
- https://developers.notion.com - Official API docs
- https://github.com/4ier/notion-cli - Top competitor (91 stars, Go, 39 commands)
- https://github.com/litencatt/notion-cli - Second competitor (25 stars, TypeScript)
- https://github.com/HermanSchoenfeld/LocalNotion - Offline mirror tool (C#)
- https://github.com/jckleiner/notion-backup - Auto-backup tool
- https://github.com/spencerpauly/awesome-notion - Ecosystem catalog
- https://github.com/makenotion/notion-mcp-server - Official OpenAPI spec source
- https://news.ycombinator.com/item?id=43517524 - Show HN: Notion backup tool
- https://thomasjfrank.com/notion-automations/ - Automation guide
- https://thomasjfrank.com/how-to-handle-notion-api-request-limits/ - Rate limit handling
