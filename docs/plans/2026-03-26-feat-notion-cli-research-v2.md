---
title: "Research: Notion CLI"
type: feat
status: active
date: 2026-03-26
phase: "1"
api: "notion"
---

# Research: Notion CLI

## Spec Discovery

- **Official OpenAPI spec:** https://raw.githubusercontent.com/makenotion/notion-mcp-server/main/scripts/notion-openapi.json
- **Source:** Official Notion MCP server repository (makenotion org)
- **Format:** OpenAPI 3.1.0 JSON
- **API Version:** 2025-09-03 (Data Source Edition)
- **Endpoint count:** 22 operations across 16 paths
- **Spec gaps:** Missing file uploads, views, markdown endpoints, OAuth, trash, custom emojis. Will augment spec from docs for full coverage.

## Competitors (Deep Analysis)

### 4ier/notion-cli (91 stars) - PRIMARY COMPETITOR
- **Repo:** https://github.com/4ier/notion-cli
- **Language:** Go (99.9%)
- **Commands:** 39 subcommands across 8 groups (auth, search, page, db, block, comment, user, file, api)
- **Last commit:** February 24, 2026 (v0.3.0)
- **Open issues:** 1 (block not saved as markdown)
- **Contributors:** ~1 (solo project)
- **Maintained:** Yes, actively
- **Notable features:**
  - Human-friendly filter syntax (no JSON)
  - Schema-aware property detection
  - Smart output (tables in terminal, JSON when piped)
  - Markdown import/export
  - Recursive block traversal with depth control
  - URL or ID flexibility
  - Agent-native design (--json, exit codes)
  - `api` escape hatch for raw HTTP
  - `doctor` command
- **Weaknesses:**
  - NO local persistence/SQLite
  - NO offline search capability
  - NO workspace-level workflows (backup, sync, stale detection)
  - NO FTS5 full-text search
  - Limited to API wrapper functionality
  - Solo maintainer (bus factor 1)

### litencatt/notion-cli (25 stars)
- **Repo:** https://github.com/litencatt/notion-cli
- **Language:** TypeScript (98.6%)
- **Commands:** 6 main commands (block, page, db, user, search, help)
- **Last commit:** November 1, 2025 (v0.15.6)
- **Open issues:** Unknown
- **Maintained:** Yes (39 releases)
- **Notable features:**
  - Multi-format output (table, CSV, JSON, YAML, raw JSON)
  - Interactive mode for database operations
  - Filter condition building and saving
- **Weaknesses:**
  - Far fewer commands than 4ier (6 vs 39)
  - No Go binary (requires Node.js runtime)
  - No agent-native features
  - No local persistence

### Other notable tools
- **MrRichRobinson/notion-cli** (0 stars): Python, full resource coverage, 2025 API model
- **lox/notion-cli** (~5 stars): Go, uses Notion MCP server with OAuth
- **kris-hansen/notion-cli** (~10 stars): Python, Taskbook-style ToDo manager

## User Pain Points

> "game-changer for power users who live in the terminal... especially for quick database queries or page exports" - edgecasehuman on HN (re: 4ier/notion-cli)

> "3 requests per second is nonsense" when you need to make multiple requests to retrieve data - Notion API rate limit discussion

> "notion block not saved as markdown" - GitHub issue #14 on 4ier/notion-cli (block export bug)

> "Notion's pricing acts as a cap since users resist paying more for an extension than for Notion itself" - Notion Backups SaaS founder on HN

> Feature request: "integrating with fzf for fuzzy searching within Notion workspaces" - HN commenter

## Auth Method
- **Type:** Bearer token (Integration token)
- **Env var convention:** `NOTION_TOKEN` or `NOTION_API_KEY` (competitors use both)
- **Header:** `Authorization: Bearer <token>`, `Notion-Version: 2026-03-11`

## Demand Signals
- Show HN: Notion-CLI (47133849) - 3 points, positive reception
- Show HN: Notion backup tool (43517524) - 1 point, SaaS business built on backup demand
- Offline mode was Notion's #1 all-time feature request (official acknowledgment)
- notion-into-sqlite exists - direct evidence of SQLite persistence demand
- alfred-notion-search exists - direct evidence of offline search demand
- 5+ backup tools on GitHub - strongest demand signal for local data

## Strategic Justification

**Why this CLI should exist when 4ier/notion-cli has 91 stars:**

4ier/notion-cli is an excellent API wrapper - the best one available. But it's ONLY an API wrapper. It has zero local persistence, zero offline capability, and zero workspace-level workflows. Every command hits the live API and is limited by the 3 req/sec rate limit.

Our CLI will be "discrawl for Notion" - it adds a SQLite data layer with:
1. **FTS5 full-text search** across page titles AND block content (Notion's API only searches titles)
2. **Incremental sync** that builds a local workspace mirror
3. **Workspace hygiene commands** (stale, orphans, stats) that query local data instantly
4. **`sql` command** for ad-hoc analysis without API rate limits
5. **Offline access** to synced content

This isn't competing on wrapper quality (4ier already does that well). It's competing on a fundamentally different architecture - local-first with API sync.

## Target
- **Command count:** 45+ (match 4ier's 39 generated commands + 7 workflow commands)
- **Key differentiator:** SQLite data layer with FTS5 search, sync, stale detection, sql command
- **Quality bar:** Steinberger Grade A (80+/100)

## Sources
- https://github.com/4ier/notion-cli - Top competitor
- https://github.com/litencatt/notion-cli - Second competitor
- https://github.com/makenotion/notion-mcp-server - Official OpenAPI spec
- https://news.ycombinator.com/item?id=47133849 - Show HN: notion-cli
- https://news.ycombinator.com/item?id=43517524 - Show HN: Notion backup
- https://github.com/FujiHaruka/notion-into-sqlite - SQLite demand evidence
