---
title: "Data Layer Specification: Notion CLI"
type: feat
status: active
date: 2026-03-26
phase: "0.7"
api: "notion"
---

# Data Layer Specification: Notion CLI

## Overview

Notion's API exposes structured (databases/data sources) and unstructured (pages/blocks) content. The data layer needs to handle both. The killer feature is FTS5 full-text search across block content - Notion's API only supports title search, so local full-text search is a genuine capability upgrade.

Offline mode was Notion's #1 all-time feature request. A CLI with local SQLite persistence directly addresses this demand signal for developer workflows.

## Entity Classification

| Entity | Type | Est. Volume | Update Frequency | Temporal Field | Persistence Need |
|--------|------|-------------|-----------------|----------------|-----------------|
| **Pages** | Accumulating | 100-10k per workspace | Weekly edits | `last_edited_time` | SQLite + FTS5 |
| **Blocks** | Accumulating | 1k-100k per workspace | Frequent edits | `last_edited_time` | SQLite + FTS5 |
| **Databases** | Reference | 10-100 per workspace | Rarely changed | `last_edited_time` | SQLite table |
| **Database Items** | Accumulating | 100-50k per database | Varies | `last_edited_time` | SQLite + FTS5 |
| **Users** | Reference | 1-1000 per workspace | Rarely changed | N/A | SQLite table |
| **Comments** | Append-only | 0-1000 per page | Rarely | `created_time` | SQLite table |
| **File Uploads** | Append-only | Varies | Never updated | `created_time` | API-only (files too large for SQLite) |
| **Views** | Reference | 1-10 per database | Rarely changed | N/A | API-only |
| **Search Results** | Ephemeral | N/A | N/A | N/A | No persistence |

**Heuristics applied:**
- Pages: has `created_time`/`last_edited_time` + paginated search -> Accumulating
- Blocks: child of pages, has `last_edited_time`, recursive structure -> Accumulating
- Databases: referenced by pages/items via `parent.database_id` -> Reference
- Users: referenced by pages via `created_by`/`last_edited_by` -> Reference (small cardinality)
- Comments: has `created_time`, no update endpoint -> Append-only

## Social Signal Mining Results

| Signal | Source | Evidence Score | What It Reveals |
|--------|--------|---------------|-----------------|
| notion-into-sqlite (GitHub) | Direct SQLite tool | 6 | Users want database content in SQLite for queries |
| Offline #1 feature request (Notion official) | Notion blog + X post | 8 | Massive demand for local access to workspace data |
| alfred-notion-search (GitHub) | Offline page search | 5 | Users want fast local search without API round-trips |
| notion_data CSV export (GitHub) | Data analysis tool | 4 | Users export to Pandas for analytics - want local queryable data |
| notion-export-kernel (GitHub) | Backup/export tool | 5 | Users want recursive page+block export |
| Notionlytics (SaaS) | Analytics product | 3 | Demand for workspace usage analytics |
| 5+ backup tools (GitHub) | Multiple repos | 7 | Strongest signal: backup/sync is the #1 use case |

**Score >= 6 findings that inform data layer:**
1. SQLite as local store (notion-into-sqlite proves the pattern)
2. Full workspace sync with block content (backup tools)
3. Offline search capability (Notion's own #1 request)

## Data Gravity Scoring

| Entity | Volume | QueryFreq | JoinDemand | SearchNeed | TemporalValue | Total | Verdict |
|--------|--------|-----------|-----------|-----------|---------------|-------|---------|
| **Pages** | 2 (100-10k) | 3 (daily) | 3 (blocks, comments, users ref it) | 2 (title text) | 2 (last_edited + trends) | **12** | PRIMARY |
| **Blocks** | 3 (10k-100k) | 3 (daily search) | 2 (pages, parent blocks) | 3 (primary text content) | 1 (created date) | **12** | PRIMARY |
| **Database Items** | 2 (100-50k) | 3 (daily queries) | 2 (database, users) | 2 (title + properties) | 2 (last_edited + trends) | **11** | PRIMARY |
| **Users** | 0 (<100) | 2 (weekly) | 3 (pages, blocks, comments ref them) | 1 (name) | 0 (no time dimension) | **6** | SUPPORT |
| **Databases** | 0 (<100) | 2 (weekly) | 3 (items, views ref them) | 1 (title) | 1 (created date) | **7** | SUPPORT |
| **Comments** | 1 (100-10k) | 1 (monthly) | 2 (pages, users) | 2 (comment text) | 1 (created date) | **7** | SUPPORT |

## SQLite Schema

### Primary Entity: Pages (Score 12)

```sql
CREATE TABLE pages (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    parent_type TEXT,          -- 'workspace', 'page_id', 'database_id'
    parent_id TEXT,
    created_time TEXT NOT NULL,
    last_edited_time TEXT NOT NULL,
    created_by_id TEXT,
    last_edited_by_id TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    in_trash INTEGER NOT NULL DEFAULT 0,
    url TEXT,
    icon_type TEXT,            -- 'emoji', 'external', 'file'
    icon_value TEXT,
    cover_type TEXT,
    cover_url TEXT,
    data JSON NOT NULL,        -- full API response
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_pages_parent ON pages(parent_type, parent_id);
CREATE INDEX idx_pages_last_edited ON pages(last_edited_time);
CREATE INDEX idx_pages_created_by ON pages(created_by_id);
CREATE INDEX idx_pages_archived ON pages(archived);

CREATE VIRTUAL TABLE pages_fts USING fts5(
    title,
    content='pages',
    content_rowid='rowid'
);
```

### Primary Entity: Blocks (Score 12)

```sql
CREATE TABLE blocks (
    id TEXT PRIMARY KEY,
    parent_type TEXT NOT NULL,  -- 'page_id', 'block_id', 'database_id'
    parent_id TEXT NOT NULL,
    page_id TEXT,               -- denormalized: the page this block belongs to
    type TEXT NOT NULL,         -- 'paragraph', 'heading_1', 'bulleted_list_item', etc.
    has_children INTEGER NOT NULL DEFAULT 0,
    archived INTEGER NOT NULL DEFAULT 0,
    in_trash INTEGER NOT NULL DEFAULT 0,
    created_time TEXT NOT NULL,
    last_edited_time TEXT NOT NULL,
    created_by_id TEXT,
    last_edited_by_id TEXT,
    plain_text TEXT,            -- extracted text content from rich_text arrays
    data JSON NOT NULL,         -- full API response
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_blocks_parent ON blocks(parent_type, parent_id);
CREATE INDEX idx_blocks_page ON blocks(page_id);
CREATE INDEX idx_blocks_type ON blocks(type);
CREATE INDEX idx_blocks_last_edited ON blocks(last_edited_time);

CREATE VIRTUAL TABLE blocks_fts USING fts5(
    plain_text,
    content='blocks',
    content_rowid='rowid'
);
```

### Primary Entity: Database Items (Score 11)

```sql
CREATE TABLE database_items (
    id TEXT PRIMARY KEY,        -- same as page ID (database items ARE pages)
    database_id TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    created_time TEXT NOT NULL,
    last_edited_time TEXT NOT NULL,
    created_by_id TEXT,
    last_edited_by_id TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    url TEXT,
    properties_json JSON,       -- extracted property values for quick filtering
    data JSON NOT NULL,         -- full API response
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_dbitems_database ON database_items(database_id);
CREATE INDEX idx_dbitems_last_edited ON database_items(last_edited_time);
CREATE INDEX idx_dbitems_created_by ON database_items(created_by_id);

CREATE VIRTUAL TABLE database_items_fts USING fts5(
    title,
    content='database_items',
    content_rowid='rowid'
);
```

### Support Entity: Users (Score 6)

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,         -- 'person' or 'bot'
    name TEXT,
    avatar_url TEXT,
    email TEXT,                 -- only for person type
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### Support Entity: Databases (Score 7)

```sql
CREATE TABLE databases (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    parent_type TEXT,
    parent_id TEXT,
    created_time TEXT,
    last_edited_time TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    is_inline INTEGER NOT NULL DEFAULT 0,
    url TEXT,
    description TEXT,
    schema_json JSON,           -- database property schema definitions
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_databases_parent ON databases(parent_type, parent_id);
```

### Support Entity: Comments (Score 7)

```sql
CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    parent_type TEXT NOT NULL,  -- 'page_id' or 'discussion_id'
    parent_id TEXT NOT NULL,
    discussion_id TEXT,
    created_time TEXT NOT NULL,
    created_by_id TEXT,
    plain_text TEXT,            -- extracted from rich_text
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_comments_parent ON comments(parent_type, parent_id);
CREATE INDEX idx_comments_discussion ON comments(discussion_id);
CREATE INDEX idx_comments_created ON comments(created_time);
```

### Sync Metadata

```sql
CREATE TABLE sync_metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
-- Stores: last_sync_time, sync_cursor, workspace_id, etc.
```

## Sync Strategy

### Incremental Sync Cursor

**Cursor field:** `last_edited_time` on pages (ISO 8601 timestamp)

**VALIDATION:** The Notion API's POST /search endpoint supports sorting by `last_edited_time` (direction: ascending/descending). However, it does NOT support filtering by `last_edited_time` directly - there is no `?since=` or `?after=` parameter.

**Sync approach (validated):**
1. POST /search with sort by `last_edited_time` descending
2. Paginate through results
3. Stop when we encounter a page with `last_edited_time` <= our last sync cursor
4. This is "scan until stale" - not true cursor filtering, but works for incremental sync
5. Store the most recent `last_edited_time` from results as the new cursor

**Batch size:** 100 (API max per page)
**Rate limit:** Average 3 req/sec. For a 1000-page workspace:
- Initial sync: ~10 search pages + 1000 page fetches + N block fetches = ~2000+ requests = ~12 minutes
- Incremental sync: Usually 1-5 search pages + changed page fetches = seconds

### Per-Entity Sync Details

| Entity | Sync Method | Cursor | Notes |
|--------|-------------|--------|-------|
| Pages | POST /search (sort by last_edited_time) | last_edited_time | Scan until stale |
| Blocks | GET /blocks/{id}/children per page | Per-page: page's last_edited_time | Only re-sync blocks for pages that changed |
| Database Items | POST /data_sources/{id}/query | last_edited_time | Per-database sync |
| Users | GET /users (full list) | N/A | Small set, always full refresh |
| Databases | Discovered via page parent_type | N/A | Sync when encountered |
| Comments | GET /comments?block_id={page_id} | N/A | Per-page, only when requested |

## Search Specification

### FTS5 Text Fields

| Entity | FTS5 Fields | Source in API Response |
|--------|------------|----------------------|
| Pages | title | `properties.title[0].plain_text` or `properties.Name.title[0].plain_text` |
| Blocks | plain_text | `<type>.rich_text[].plain_text` concatenated |
| Database Items | title | `properties.<title_prop>.title[0].plain_text` |

### Domain-Specific Search Filters (CLI flags -> SQL)

| CLI Flag | SQL WHERE Clause | Description |
|----------|-----------------|-------------|
| `--type page` | `WHERE type = 'page'` | Filter by object type |
| `--type database` | Search databases table | Filter databases |
| `--parent <id>` | `WHERE parent_id = ?` | Filter by parent |
| `--author <name>` | `JOIN users ON created_by_id = users.id WHERE users.name LIKE ?` | Filter by creator |
| `--edited-by <name>` | `JOIN users ON last_edited_by_id = users.id WHERE users.name LIKE ?` | Filter by editor |
| `--since <date>` | `WHERE last_edited_time >= ?` | Filter by recency |
| `--before <date>` | `WHERE last_edited_time < ?` | Filter by date |
| `--archived` | `WHERE archived = 1` | Include archived |
| `--database <id>` | `WHERE database_id = ?` | Filter database items |
| `--block-type <type>` | `WHERE type = ?` | Filter block type (heading, paragraph, etc.) |

## Compound Cross-Entity Queries

### 1. Full-text search across pages and blocks
```sql
SELECT p.id, p.title, p.url, b.plain_text, b.type
FROM blocks b
JOIN blocks_fts ON blocks_fts.rowid = b.rowid
JOIN pages p ON b.page_id = p.id
WHERE blocks_fts MATCH ?
ORDER BY rank
LIMIT ?;
```
**Validation:** `b.page_id` references `pages.id`. Both columns exist in schema.

### 2. Recently edited pages with their editors
```sql
SELECT p.id, p.title, u.name as editor, p.last_edited_time
FROM pages p
LEFT JOIN users u ON p.last_edited_by_id = u.id
WHERE p.last_edited_time >= ?
ORDER BY p.last_edited_time DESC;
```
**Validation:** `p.last_edited_by_id` references `users.id`. Both exist.

### 3. Stale pages (not edited in N days)
```sql
SELECT p.id, p.title, p.last_edited_time, p.url,
       julianday('now') - julianday(p.last_edited_time) as days_stale
FROM pages p
WHERE p.archived = 0
  AND p.last_edited_time < datetime('now', '-' || ? || ' days')
ORDER BY p.last_edited_time ASC;
```
**Validation:** `last_edited_time` is indexed and always populated.

### 4. Database items with cross-reference to parent database
```sql
SELECT di.id, di.title, d.title as database_name, di.last_edited_time
FROM database_items di
JOIN databases d ON di.database_id = d.id
WHERE di.database_id = ?
ORDER BY di.last_edited_time DESC;
```
**Validation:** `di.database_id` references `databases.id`. Both exist.

### 5. Orphan detection (pages with no block children)
```sql
SELECT p.id, p.title, p.created_time, p.url
FROM pages p
LEFT JOIN blocks b ON b.parent_id = p.id AND b.parent_type = 'page_id'
WHERE b.id IS NULL
  AND p.archived = 0;
```
**Validation:** `b.parent_id`/`b.parent_type` reference page IDs. Works.

## Tail Strategy

| Method | Available? | Decision |
|--------|-----------|----------|
| WebSocket/Gateway | No | N/A |
| SSE | No | N/A |
| REST Polling | Yes | **USE THIS** |

**Implementation:** Poll POST /search sorted by last_edited_time DESC every N seconds (default 30s). Compare against last known timestamp. Report new/changed pages.

**Justification:** Notion has no real-time streaming API. Webhooks exist only in database automations (button-triggered, not API-subscribable). REST polling is the only option.

## Phase 4 Priority 0 Commands (from this data layer)

1. `sync` - Incremental workspace sync to SQLite
2. `search` - FTS5 full-text search across pages + blocks
3. `sql` - Raw read-only SQL queries
4. `pages` - List/filter pages from local DB
5. `blocks` - List/filter blocks from local DB for a page
6. `items` - List/filter database items from local DB
7. `tail` - Watch for workspace changes via REST polling

## Sources
- Notion API docs: https://developers.notion.com/reference
- notion-into-sqlite: https://github.com/FujiHaruka/notion-into-sqlite
- Notion offline announcement: https://www.notion.com/blog/how-we-made-notion-available-offline
- alfred-notion-search: https://github.com/svenko99/alfred-notion-search
- notion_data: https://github.com/victoriano/notion_data
