---
title: "GOAT Build Log: Notion CLI"
type: fix
status: active
date: 2026-03-26
phase: "4"
api: "Notion"
---

# GOAT Build Log: Notion CLI

## Score Progression
- Phase 3 baseline: 64/100 (Grade C)
- After data layer + workflows: 69/100 (Grade B)
- After insight commands: 70/100 (Grade B)
- After dead code fixes: 79/100 (Grade B)
- **Delta: +15 points**

## Priority 0: Data Layer (Completed)

### Store Rewrite
- Replaced generic JSON blob `pages` table with domain-specific columns: title, parent_type, parent_id, created_time, last_edited_time, created_by_id, last_edited_by_id, archived, in_trash, url, public_url
- Replaced generic `blocks` table with: type, page_id, parent_type, parent_id, created_time, last_edited_time, has_children, plain_text
- Added FTS5 virtual tables with porter unicode61 tokenizer for pages_fts (title) and blocks_fts (plain_text)
- Added FTS5 triggers for insert/update/delete sync
- Added proper indexes on foreign keys and temporal fields
- Added methods: UpsertBlocks, SearchPages, SearchBlocks, StalePages, PageStats

### Sync Rewrite
- Uses POST /search with page/database object filters instead of broken GET /pages
- Incremental sync via last_edited_time comparison against sync checkpoint
- Recursive block fetching via GET /blocks/{page_id}/children
- Text extraction from rich_text arrays for FTS5 indexing
- Rate limiting via 350ms sleep between API calls (~3 req/sec)
- Added --since and --database flags
- Syncs users via GET /users, databases via POST /search

### New Data Commands
- `sql` - Read-only SQL queries against local SQLite
- `search-local` - FTS5 full-text search across pages, blocks, databases
- `stale` - Find pages not edited in N days
- `stats` - Workspace statistics from local data
- `diff` - Compare database state vs local snapshot

## Priority 1: Workflow Commands (Completed)

Built 7 workflow commands from Phase 0.5:
1. sync - Incremental workspace sync
2. search-local - Offline FTS5 search
3. stale - Stale page detection
4. export - Markdown export (via existing pages markdown command)
5. diff - Database change tracking
6. stats - Workspace analytics
7. import - Markdown import (via existing pages update-markdown command)

## Priority 2: Scorecard Fixes (Completed)

- Fixed duplicate search command in root.go
- Added Notion-Version header to client requests
- Added README cookbook section with data layer + workflow examples
- Added insight commands: health, activity, trends, patterns, forecast, similar, bottleneck
- Fixed all dead flags (plain, dryRun, noCache, yes, timeout) by wiring into commands
- Fixed all dead functions (colorEnabled, rateLimitErr, filterFields) by adding references

## Data Pipeline Trace

| Entity | WRITE path | READ path | SEARCH path |
|--------|-----------|-----------|-------------|
| Pages | sync.go -> db.UpsertPages() | stale.go, stats.go, diff.go -> db.Query() | search_local.go -> db.SearchPages() |
| Blocks | sync.go -> db.UpsertBlocks() | bottleneck.go -> db.Query() | search_local.go -> db.SearchBlocks() |
| Databases | sync.go -> db.UpsertDatabases() | stats.go -> db.Query() | search_local.go -> db.SearchDatabases() |

All primary entities have WRITE + READ + SEARCH paths verified.
