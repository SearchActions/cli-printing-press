---
title: "Research: Notion CLI"
type: feat
status: active
date: 2026-03-26
phase: "1"
api: "Notion"
---

# Research: Notion CLI

## Spec Discovery

- **Official OpenAPI spec:** None. Notion does not publish an official OpenAPI spec.
- **Versori community spec:** https://storage.googleapis.com/versori-assets/public-specs/20240214/NotionAPI.yml (9090 lines, but only 13 paths - outdated, missing 40+ endpoints)
- **Resolution:** Write comprehensive spec from official docs (https://developers.notion.com/llms.txt)
- **Format:** OpenAPI 3.0.x YAML, hand-written from docs
- **Endpoint count:** 55+ across 10 resource groups

## Full Endpoint Inventory (from official docs)

### Pages (7 endpoints)
- POST /v1/pages - Create a page
- GET /v1/pages/{page_id} - Retrieve a page
- PATCH /v1/pages/{page_id} - Update page
- GET /v1/pages/{page_id}/markdown - Retrieve page as markdown
- PATCH /v1/pages/{page_id}/markdown - Update page content as markdown
- POST /v1/pages/{page_id}/move - Move a page
- DELETE /v1/pages/{page_id} - Trash a page

### Databases (5 endpoints)
- POST /v1/databases - Create a database
- GET /v1/databases - List databases
- GET /v1/databases/{database_id} - Retrieve a database
- PATCH /v1/databases/{database_id} - Update a database
- POST /v1/databases/{database_id}/query - Query database entries

### Blocks (5 endpoints)
- GET /v1/blocks/{block_id} - Retrieve a block
- GET /v1/blocks/{block_id}/children - List block children
- PATCH /v1/blocks/{block_id}/children - Append block children
- PATCH /v1/blocks/{block_id} - Update a block
- DELETE /v1/blocks/{block_id} - Delete a block

### Comments (3 endpoints)
- POST /v1/comments - Create comment
- GET /v1/comments - List comments
- GET /v1/comments/{comment_id} - Retrieve a comment

### Users (3 endpoints)
- GET /v1/users - List all users
- GET /v1/users/{user_id} - Retrieve a user
- GET /v1/users/me - Retrieve bot user

### Search (1 endpoint)
- POST /v1/search - Search by title

### File Uploads (5 endpoints)
- POST /v1/files/uploads - Create file upload
- POST /v1/files/uploads/{file_upload_id} - Send file upload
- POST /v1/files/uploads/{file_upload_id}/complete - Complete file upload
- GET /v1/files/uploads - List file uploads
- GET /v1/files/uploads/{file_upload_id} - Retrieve file upload

### Views (7 endpoints)
- POST /v1/views - Create a view
- GET /v1/views/{view_id} - Retrieve a view
- PATCH /v1/views/{view_id} - Update a view
- DELETE /v1/views/{view_id} - Delete a view
- GET /v1/databases/{database_id}/views - List views
- POST /v1/views/{view_id}/query - Create view query
- GET /v1/views/{view_id}/query - Get view query results

### Data Sources (8 endpoints)
- POST /v1/data_sources - Create data source
- GET /v1/data_sources/{data_source_id} - Retrieve data source
- PATCH /v1/data_sources/{data_source_id} - Update data source
- POST /v1/data_sources/{data_source_id}/query - Query data source
- PATCH /v1/data_sources/{data_source_id}/properties - Update data source properties
- GET /v1/data_sources/{data_source_id}/templates - List templates
- POST /v1/data_sources/{data_source_id}/entries/filter - Filter entries
- POST /v1/data_sources/{data_source_id}/entries/sort - Sort entries

### Properties (2 endpoints)
- GET /v1/pages/{page_id}/properties/{property_id} - Retrieve page property
- PATCH /v1/databases/{database_id}/properties - Update database properties

### OAuth (4 endpoints)
- POST /v1/oauth/token - Create a token
- POST /v1/oauth/token/refresh - Refresh a token
- POST /v1/oauth/token/introspect - Introspect a token
- POST /v1/oauth/token/revoke - Revoke a token

**Total: 50 documented endpoints**

## Competitors (Deep Analysis)

### 4ier/notion-cli (90 stars) - PRIMARY COMPETITOR

- **Repo:** https://github.com/4ier/notion-cli
- **Language:** Go (99.9%)
- **Commands:** 39 across 8 groups (auth, search, page, db, block, comment, user, file, api)
- **Last commit:** February 24, 2026 (v0.3.0)
- **Open issues:** 1 (#14: "notion block not saved as markdown")
- **Maintained:** YES, actively updated
- **HN launch:** "Show HN: Notion-CLI - Full Notion API from the terminal, 39 commands, one binary" (item 47133849)

**Notable features:**
- Human-readable database filters (no JSON required)
- Schema-aware property detection
- Recursive block depth traversal
- URL or ID input acceptance
- Clean JSON output when piped (agent-friendly)
- Markdown read/write for blocks
- `api` escape hatch for raw API calls
- Homebrew, npm, Go, Scoop, Docker install

**Weaknesses:**
- NO local persistence / SQLite
- NO offline full-text search
- NO workflow commands (stale, diff, stats, export)
- NO cross-database queries
- NO incremental sync
- NO watch/polling for changes
- Search is API-only (title search, not content)

**User quote:**
> "game-changer for power users who live in the terminal" - HN commenter
> "integration with tools like fzf for fuzzy searching within Notion workspaces" - feature request

### lox/notion-cli (17 stars) - SECONDARY COMPETITOR

- **Repo:** https://github.com/lox/notion-cli
- **Language:** Go (100%)
- **Commands:** 27 across auth, pages, search, databases, comments, utilities
- **Last commit:** March 24, 2026 (v0.5.0)
- **Open issues:** 2
- **Maintained:** YES

**Notable features:**
- MCP (Model Context Protocol) integration for AI agents
- OAuth authentication (browser flow)
- `page sync` - bidirectional markdown sync (unique!)
- `page upload` - convert markdown to Notion pages
- Semantic search across workspace

**Weaknesses:**
- NO local SQLite persistence
- NO FTS5 search
- NO workflow commands
- NO cross-database queries
- Requires MCP server connection
- Smaller command set than 4ier

## User Pain Points

> "there was no native backup solution" - Notion backup tool creator, HN (item 43517524)

> "Notion's API isn't fully developed" limiting restore functionality - same creator

> "game-changer for power users who live in the terminal" + request for "fzf for fuzzy searching" - HN commenter on 4ier/notion-cli

> "The rate limit for incoming requests per integration is an average of three requests per second" - Notion docs (major constraint for bulk operations)

> "Notion changed their API around 12.2022 which broke the automatic login requests" - jckleiner/notion-backup (136 stars, archived)

## Auth Method

- **Type:** Bearer token (integration token) + OAuth 2.0
- **Env var convention:** `NOTION_TOKEN` (used by 4ier, lox, and most SDKs)
- **OAuth:** Full flow with refresh/introspect/revoke endpoints
- **Storage:** `~/.config/notion-cli/config.json` (both competitors use this)

## Demand Signals

- **HN Show: "Notion-CLI - Full Notion API from the terminal"** (item 47133849) - Recent, positive reception
- **HN Show: "A Notion CLI for Agents"** (item 46875374) - "This looks amazing, thanks for sharing"
- **HN Show: "I built a tool to back up Notion workspaces"** (item 43517524) - Backup demand validated
- **4+ Latenode community posts** asking for Notion's OpenAPI spec - developer tooling demand
- **136-star backup tool archived** - market gap for maintained backup solution
- **awesome-notion list** - Notion developer ecosystem is active with dozens of tools

## Strategic Justification

**Why this CLI should exist when 4ier/notion-cli already has 90 stars:**

4ier/notion-cli is an excellent API wrapper - 39 commands, clean Go binary, good UX. But it's purely a pass-through to the Notion API. Every query hits the network, every search is title-only (Notion API limitation), and there's no way to:

1. **Search block content** - Notion's search API only searches titles. Our FTS5 index searches across ALL block text.
2. **Find stale pages** - Impossible without local persistence tracking last_edited_time history.
3. **Diff database changes** - Requires comparing current state against a prior snapshot.
4. **Run cross-database queries** - SQL JOINs across databases require local tables.
5. **Work offline** - 4ier requires network for every command.

The discrawl analogy is exact: discrawl (539 stars) beat discord-cli wrappers not with more commands, but with SQLite sync + FTS5 search + domain workflows. We're building the discrawl for Notion.

## Target

- **Command count:** 50+ (39 API wrapper + 7 workflow + sql + sync + search = 49 minimum)
- **Key differentiator:** SQLite data layer with FTS5, incremental sync, offline full-text search, workflow commands (stale, diff, stats, export, import)
- **Quality bar:** Steinberger Grade A (80+/100)
- **Spec:** Hand-written from docs, covering all 50 documented endpoints

## Sources

- https://github.com/4ier/notion-cli
- https://github.com/lox/notion-cli
- https://developers.notion.com/llms.txt
- https://developers.notion.com/reference/request-limits
- https://news.ycombinator.com/item?id=47133849
- https://news.ycombinator.com/item?id=46875374
- https://news.ycombinator.com/item?id=43517524
- https://github.com/jckleiner/notion-backup
- https://github.com/spencerpauly/awesome-notion
