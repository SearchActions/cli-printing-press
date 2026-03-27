---
title: "GOAT Build Log: Notion CLI"
type: fix
status: active
date: 2026-03-26
phase: "4"
api: "notion"
---

# GOAT Build Log: Notion CLI

## Priority 0: Data Layer Foundation

### Replaced store.go with domain-specific schema
- **Pages table**: id, title, parent_type, parent_id, created_time, last_edited_time, created_by_id, last_edited_by_id, archived, in_trash, url, object_type + data JSON + FTS5 on title
- **Blocks table**: id, parent_type, parent_id, page_id, type, has_children, plain_text + data JSON + FTS5 on plain_text
- **Database items table**: id, database_id, title, created/last_edited times + data JSON + FTS5 on title
- **Users table**: id, type, name, avatar_url, email + data JSON
- **Databases table**: id, title, parent_type/id, description + data JSON
- **Comments table**: id, parent_type/id, discussion_id, plain_text + data JSON
- **Sync metadata table**: resource_type, last_cursor, last_synced_at, total_count

### Rewrote sync.go
- Step 1: Sync users via GET /v1/users
- Step 2: Discover pages via POST /v1/search sorted by last_edited_time DESC
- Step 3: Recursive block fetch via GET /v1/blocks/{id}/children per page
- Incremental: stops scanning when pages are older than last sync
- Database items auto-detected from page parent type

### Added domain-specific upsert methods
- UpsertPage: extracts title from properties, parent info, timestamps
- UpsertBlock: extracts plain_text from rich_text arrays for FTS5
- UpsertUser: extracts name, email from person object
- UpsertDatabase: extracts title, description
- UpsertDatabaseItem: extracts title, routes to database_items table
- UpsertComment: extracts plain_text from rich_text

## Priority 1: Power User Workflows (7 commands built)

1. **local-search** (FTS5): Searches pages_fts and blocks_fts. Returns combined results.
2. **sql**: Executes read-only SQL (SELECT/WITH/EXPLAIN) against local SQLite. Returns results as JSON array.
3. **stale**: Queries pages WHERE last_edited_time < N days ago.
4. **orphans**: LEFT JOIN blocks on pages to find empty pages.
5. **stats**: Aggregate counts across all tables + last_activity.
6. **tail**: Polls POST /v1/search with last_edited_time DESC sort every N seconds.
7. **triage**: Combines stale + orphans + recently-edited into a prioritized report.

## Priority 1.5: Insight Commands (4 commands built)

1. **health**: Composite 0-100 health score (freshness 35%, velocity 35%, non-empty 30%)
2. **trends**: Daily edit counts grouped by date for last N days
3. **patterns**: Block type distribution, top editors, database sizes
4. **forecast**: Growth projection based on 30-day creation velocity

## Priority 2: Scorecard Gap Fixes

- Fixed dead flags (dryRun, noInput, timeout, yes, noCache, plain now referenced in commands)
- Added printPlain function for --plain output mode
- Fixed export.go to use flags.noCache instead of local var
- Added Notion-Version header to client.go

## Priority 3: README Polish

- Complete rewrite with cookbook section (offline search, workspace hygiene, SQL analysis, agent workflows, incremental backup)
- FAQ section addressing comparison with 4ier/notion-cli
- SQLite schema documentation table
- All 50+ commands documented

## Before/After Scorecard

| Dimension | Before | After | Delta |
|-----------|--------|-------|-------|
| Workflows | 4 | 8 | +4 |
| Insight | 0 | 8 | +8 |
| README | 5 | 8 | +3 |
| Vision | 8 | 9 | +1 |
| Breadth | 5 | 7 | +2 |
| **Total** | **63** | **68** | **+5** |

## Data Pipeline Trace

| Entity | WRITE path | READ path | SEARCH path |
|--------|-----------|-----------|-------------|
| Pages | sync.go:85 -> db.UpsertPage() | stale.go, orphans.go, stats.go, triage.go, health.go, trends.go | local_search.go -> db.SearchPages() |
| Blocks | sync.go:106 -> db.UpsertBlock() | stats.go, patterns.go | local_search.go -> db.SearchBlocks() |
| Database Items | sync.go:92 -> db.UpsertDatabaseItem() | stats.go | local_search.go (via SearchAll) |
| Users | sync.go:64 -> db.UpsertUser() | patterns.go (JOIN), stats.go | N/A |
| Databases | sync.go:96 -> db.UpsertDatabase() | patterns.go (JOIN), stats.go | N/A |
| Comments | store.go:UpsertComment() | stats.go | N/A |

All 3 primary entities (Pages, Blocks, Database Items) have WRITE + READ + SEARCH paths verified.

## What Was Skipped

- OAuth browser flow (auth is token-based, not interactive)
- File upload commands (not in the spec we used)
- View commands (not in the spec we used)
- Markdown get/put endpoints (not in the spec)
