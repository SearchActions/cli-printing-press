---
title: "Data Layer Specification: Notion CLI"
type: feat
status: active
date: 2026-03-26
phase: "0.7"
api: "Notion"
---

# Data Layer Specification: Notion CLI

## Overview

The Notion API exposes ~55 endpoints across 10 resource groups. This specification defines the local SQLite data layer that transforms the CLI from an API wrapper into a knowledge base tool. Notion has NO webhooks, NO websockets, and NO SSE - all sync is via REST polling with `last_edited_time` as the cursor. The 3 req/sec rate limit makes local persistence essential for any query-intensive workflow.

## Entity Classification

| Entity | Type | Est. Volume | Update Freq | Temporal Field | Persistence Need |
|--------|------|-------------|-------------|----------------|-----------------|
| **Pages** | Accumulating | 100-100k per workspace | Frequent | `last_edited_time` | SQLite table + incremental sync |
| **Blocks** | Accumulating | 1k-10M per workspace | Frequent | `last_edited_time` | SQLite table + incremental sync |
| **Databases** | Reference | 10-1000 per workspace | Infrequent | `last_edited_time` | SQLite table + periodic refresh |
| **Users** | Reference | 1-1000 per workspace | Rare | None | SQLite table + periodic refresh |
| **Comments** | Append-only | 0-100k per workspace | Moderate | `created_time` | SQLite table + tail |
| **Views** | Reference | 1-10 per database | Infrequent | None | SQLite table + periodic refresh |
| **Data Sources** | Reference | 1-5 per database | Rare | None | SQLite table + periodic refresh |
| **File Uploads** | Append-only | 0-10k per workspace | Moderate | None | API-only (files too large) |
| **OAuth Tokens** | Ephemeral | 1-5 | Rare | N/A | Config file only |
| **Search Results** | Ephemeral | N/A | N/A | N/A | API-only, no persistence |

### Classification Rationale

- **Pages:** Have `created_time` + `last_edited_time` + paginated list (via search). Core accumulating entity.
- **Blocks:** Have timestamps + paginated list (children endpoint). Highest volume entity. Parent is always a page or another block.
- **Databases:** Referenced by pages (parent), views, data sources. Changes rarely. ~10-1000 per workspace.
- **Users:** Referenced by pages (created_by, last_edited_by), blocks, comments. Very small cardinality.
- **Comments:** Have `created_time`, `last_edited_time`. No UPDATE endpoint (append-only).
- **Views/Data Sources:** New API entities. Small cardinality, referenced by databases.
- **File Uploads:** Binary content - too large to persist locally. Keep metadata only.

## Social Signal Mining Results

### Signal 1: Local search across content (Score: 8/10)
- 4ier/notion-cli users request fzf integration (HN thread)
- Notion's API search is title-only - users can't search block content
- Multiple tools export to markdown specifically to enable grep

### Signal 2: Incremental backup with change tracking (Score: 7/10)
- 136-star backup tool archived, 77-star tool active
- HN: "there was no native backup solution"
- Users want git-style versioning of workspace content

### Signal 3: Cross-database reporting (Score: 6/10)
- notion_data exports to CSV for analytics in Pandas
- Users build scripts joining data from multiple databases
- No existing tool supports SQL queries across databases

### Signal 4: Stale content detection (Score: 6/10)
- notion-auto-archive has users wanting automated cleanup
- Knowledge bases accumulate stale content without hygiene tools

### Signal 5: Offline access to workspace content (Score: 6/10)
- LocalNotion project (offline mirror)
- Multiple export tools designed for offline reading

## Data Gravity Scoring

| Entity | Volume (0-3) | QueryFreq (0-3) | JoinDemand (0-2) | SearchNeed (0-2) | TemporalValue (0-2) | **Total** | **Tier** |
|--------|-------------|-----------------|-------------------|------------------|---------------------|-----------|----------|
| **Pages** | 2 | 3 | 2 | 2 | 2 | **11** | Primary |
| **Blocks** | 3 | 3 | 2 | 2 | 2 | **12** | Primary |
| **Databases** | 1 | 2 | 2 | 1 | 1 | **7** | Support |
| **Users** | 0 | 2 | 2 | 0 | 0 | **4** | API-only* |
| **Comments** | 1 | 1 | 1 | 2 | 1 | **6** | Support |
| **Views** | 0 | 1 | 1 | 0 | 0 | **2** | API-only |
| **Data Sources** | 0 | 1 | 1 | 0 | 0 | **2** | API-only |

*Users promoted to Support tier because they're referenced by every page and block (created_by, last_edited_by).

### Primary Entities (Score >= 8): Pages, Blocks
### Support Entities (Score 5-7): Databases, Users, Comments
### API-only (Score < 5): Views, Data Sources, File Uploads

## SQLite Schema

### Primary Entity: Pages

```sql
CREATE TABLE pages (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    parent_type TEXT NOT NULL, -- 'database_id', 'page_id', 'workspace'
    parent_id TEXT,
    created_time TEXT NOT NULL,
    last_edited_time TEXT NOT NULL,
    created_by_id TEXT,
    last_edited_by_id TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    in_trash INTEGER NOT NULL DEFAULT 0,
    url TEXT,
    public_url TEXT,
    icon_type TEXT, -- 'emoji', 'external', 'file', null
    icon_value TEXT, -- emoji char or URL
    cover_url TEXT,
    data JSON NOT NULL, -- full API response
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_pages_parent ON pages(parent_type, parent_id);
CREATE INDEX idx_pages_last_edited ON pages(last_edited_time);
CREATE INDEX idx_pages_created ON pages(created_time);
CREATE INDEX idx_pages_archived ON pages(archived);
```

### Primary Entity: Blocks

```sql
CREATE TABLE blocks (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL, -- 'paragraph', 'heading_1', 'to_do', etc.
    page_id TEXT NOT NULL, -- root page this block belongs to
    parent_type TEXT NOT NULL, -- 'page_id' or 'block_id'
    parent_id TEXT NOT NULL,
    created_time TEXT NOT NULL,
    last_edited_time TEXT NOT NULL,
    created_by_id TEXT,
    last_edited_by_id TEXT,
    has_children INTEGER NOT NULL DEFAULT 0,
    archived INTEGER NOT NULL DEFAULT 0,
    in_trash INTEGER NOT NULL DEFAULT 0,
    plain_text TEXT NOT NULL DEFAULT '', -- extracted text content for search
    data JSON NOT NULL, -- full API response
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_blocks_page ON blocks(page_id);
CREATE INDEX idx_blocks_parent ON blocks(parent_type, parent_id);
CREATE INDEX idx_blocks_type ON blocks(type);
CREATE INDEX idx_blocks_last_edited ON blocks(last_edited_time);
```

### FTS5 Virtual Tables

```sql
-- Full-text search across page titles
CREATE VIRTUAL TABLE pages_fts USING fts5(
    id UNINDEXED,
    title,
    content='pages',
    content_rowid='rowid'
);

-- Full-text search across block content
CREATE VIRTUAL TABLE blocks_fts USING fts5(
    id UNINDEXED,
    page_id UNINDEXED,
    type UNINDEXED,
    plain_text,
    content='blocks',
    content_rowid='rowid'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER pages_ai AFTER INSERT ON pages BEGIN
    INSERT INTO pages_fts(rowid, id, title) VALUES (new.rowid, new.id, new.title);
END;
CREATE TRIGGER pages_ad AFTER DELETE ON pages BEGIN
    INSERT INTO pages_fts(pages_fts, rowid, id, title) VALUES('delete', old.rowid, old.id, old.title);
END;
CREATE TRIGGER pages_au AFTER UPDATE ON pages BEGIN
    INSERT INTO pages_fts(pages_fts, rowid, id, title) VALUES('delete', old.rowid, old.id, old.title);
    INSERT INTO pages_fts(rowid, id, title) VALUES (new.rowid, new.id, new.title);
END;

CREATE TRIGGER blocks_ai AFTER INSERT ON blocks BEGIN
    INSERT INTO blocks_fts(rowid, id, page_id, type, plain_text) VALUES (new.rowid, new.id, new.page_id, new.type, new.plain_text);
END;
CREATE TRIGGER blocks_ad AFTER DELETE ON blocks BEGIN
    INSERT INTO blocks_fts(blocks_fts, rowid, id, page_id, type, plain_text) VALUES('delete', old.rowid, old.id, old.page_id, old.type, old.plain_text);
END;
CREATE TRIGGER blocks_au AFTER UPDATE ON blocks BEGIN
    INSERT INTO blocks_fts(blocks_fts, rowid, id, page_id, type, plain_text) VALUES('delete', old.rowid, old.id, old.page_id, old.type, old.plain_text);
    INSERT INTO blocks_fts(rowid, id, page_id, type, plain_text) VALUES (new.rowid, new.id, new.page_id, new.type, new.plain_text);
END;
```

### Support Entity: Databases

```sql
CREATE TABLE databases (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    parent_type TEXT NOT NULL,
    parent_id TEXT,
    created_time TEXT NOT NULL,
    last_edited_time TEXT NOT NULL,
    created_by_id TEXT,
    last_edited_by_id TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    in_trash INTEGER NOT NULL DEFAULT 0,
    is_inline INTEGER NOT NULL DEFAULT 0,
    url TEXT,
    public_url TEXT,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_databases_parent ON databases(parent_type, parent_id);
CREATE INDEX idx_databases_last_edited ON databases(last_edited_time);
```

### Support Entity: Users

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL, -- 'person' or 'bot'
    name TEXT NOT NULL DEFAULT '',
    avatar_url TEXT,
    email TEXT,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### Support Entity: Comments

```sql
CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    parent_type TEXT NOT NULL, -- 'page_id' or 'block_id'
    parent_id TEXT NOT NULL,
    discussion_id TEXT NOT NULL,
    created_time TEXT NOT NULL,
    last_edited_time TEXT,
    created_by_id TEXT,
    plain_text TEXT NOT NULL DEFAULT '', -- extracted from rich_text
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_comments_parent ON comments(parent_type, parent_id);
CREATE INDEX idx_comments_discussion ON comments(discussion_id);
CREATE INDEX idx_comments_created ON comments(created_time);

-- FTS on comment text
CREATE VIRTUAL TABLE comments_fts USING fts5(
    id UNINDEXED,
    parent_id UNINDEXED,
    plain_text,
    content='comments',
    content_rowid='rowid'
);
```

### Sync Metadata

```sql
CREATE TABLE sync_state (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
-- Stores: last_sync_time, last_cursor, sync_scope, etc.
```

## Sync Strategy

### Incremental Sync Cursor: `last_edited_time`

**Validation:** The `POST /search` endpoint supports `filter.timestamp.last_edited_time` with `on_or_after` parameter. This allows querying for pages/databases modified since the last sync.

**Sync Algorithm:**

```
1. Read last_sync_time from sync_state table
2. POST /search with filter: { timestamp: { last_edited_time: { on_or_after: last_sync_time }}}
3. For each returned page:
   a. Upsert into pages table
   b. GET /blocks/{page_id}/children recursively
   c. For each block: extract plain_text from rich_text arrays, upsert into blocks table
4. POST /search with filter: { property: "object", value: "database" } for database schema sync
5. GET /users to refresh user table
6. Update sync_state.last_sync_time = now()
```

**Batch size:** API max is 100 per page. Use `start_cursor` for pagination.
**Rate limiting:** 3 req/sec average. Use token bucket with exponential backoff on 429.
**Estimated time:** For a workspace with 1000 pages, ~5 min full sync, ~30s incremental.

### Text Extraction from Blocks

Each block type contains rich_text arrays. Extract plain_text by:
1. Iterating the `rich_text` array
2. Concatenating `plain_text` fields
3. For `child_page` blocks: use `title` field
4. For `code` blocks: include both `rich_text` and `language`
5. For `to_do` blocks: prefix with `[x]` or `[ ]`
6. For `bookmark`/`embed`/`link_preview`: include URL

## Domain-Specific Search Filters

| CLI Flag | SQL WHERE Clause | Description |
|----------|-----------------|-------------|
| `--type page` | `FROM pages_fts` | Search page titles only |
| `--type block` | `FROM blocks_fts` | Search block content only |
| `--type comment` | `FROM comments_fts` | Search comment text only |
| `--parent <id>` | `WHERE page_id = ?` or `WHERE parent_id = ?` | Scope to specific page/DB |
| `--database <id>` | `JOIN pages ON pages.parent_id = ? AND pages.parent_type = 'database_id'` | Scope to database entries |
| `--days N` | `WHERE last_edited_time >= datetime('now', '-N days')` | Time filter |
| `--since <date>` | `WHERE last_edited_time >= ?` | Absolute time filter |
| `--block-type <type>` | `WHERE type = ?` | Filter by block type (paragraph, to_do, code, etc.) |
| `--archived` | `WHERE archived = 1` | Include archived content |
| `--author <user-id>` | `WHERE created_by_id = ?` | Filter by author |

## Compound Cross-Entity Queries

### 1. Full-text search with page context
```sql
SELECT p.title AS page_title, p.url, b.type, b.plain_text,
       highlight(blocks_fts, 3, '<b>', '</b>') AS snippet
FROM blocks_fts
JOIN blocks b ON b.id = blocks_fts.id
JOIN pages p ON p.id = b.page_id
WHERE blocks_fts MATCH ?
ORDER BY rank
LIMIT 20;
```
**Validation:** blocks.page_id joins to pages.id. Both tables populated by sync.

### 2. Stale pages by database
```sql
SELECT d.title AS database_name, p.title, p.last_edited_time, p.url,
       julianday('now') - julianday(p.last_edited_time) AS days_stale
FROM pages p
JOIN databases d ON p.parent_id = d.id AND p.parent_type = 'database_id'
WHERE p.last_edited_time < datetime('now', '-' || ? || ' days')
AND p.archived = 0
ORDER BY days_stale DESC;
```
**Validation:** pages.parent_id + parent_type joins to databases.id.

### 3. Content activity by user
```sql
SELECT u.name, COUNT(DISTINCT p.id) AS pages_edited,
       COUNT(DISTINCT b.id) AS blocks_edited,
       MAX(p.last_edited_time) AS last_active
FROM users u
LEFT JOIN pages p ON p.last_edited_by_id = u.id
LEFT JOIN blocks b ON b.last_edited_by_id = u.id
GROUP BY u.id
ORDER BY pages_edited DESC;
```
**Validation:** pages.last_edited_by_id and blocks.last_edited_by_id join to users.id.

### 4. Database diff (entries changed since timestamp)
```sql
SELECT p.id, p.title, p.last_edited_time,
       CASE WHEN p.synced_at > ? THEN 'modified'
            WHEN p.created_time > ? THEN 'added'
            ELSE 'unchanged' END AS change_type
FROM pages p
WHERE p.parent_type = 'database_id' AND p.parent_id = ?
AND (p.last_edited_time > ? OR p.created_time > ?)
ORDER BY p.last_edited_time DESC;
```
**Validation:** pages filtered by parent_type='database_id' + parent_id.

### 5. Workspace statistics
```sql
SELECT
    (SELECT COUNT(*) FROM pages WHERE archived = 0) AS total_pages,
    (SELECT COUNT(*) FROM blocks WHERE archived = 0) AS total_blocks,
    (SELECT COUNT(*) FROM databases WHERE archived = 0) AS total_databases,
    (SELECT COUNT(*) FROM users) AS total_users,
    (SELECT COUNT(*) FROM comments) AS total_comments,
    (SELECT MAX(last_edited_time) FROM pages) AS last_activity,
    (SELECT COUNT(*) FROM pages WHERE last_edited_time >= datetime('now', '-7 days')) AS active_last_week;
```
**Validation:** All tables populated by sync.

## Tail Strategy

| Method | Available? | Decision |
|--------|-----------|----------|
| WebSocket/Gateway | NO | Not available |
| SSE | NO | Not available |
| REST Polling | YES | **USE THIS** - `POST /search` with `last_edited_time.on_or_after` cursor |

**Implementation:** The `watch` command polls `POST /search` on an interval (default 30s), comparing results against previous cursor. New/changed pages emitted as NDJSON to stdout.

**Rate limit budget:** At 30s interval, uses 2 req/cycle (search + pagination) = 4 req/min = well within 180 req/min budget.

## Phase 4 Priority 0 Commands (Data Layer)

These commands MUST be built before any workflow commands:

1. **`sync`** - Populate all tables via incremental sync
2. **`search`** - FTS5 query against local database
3. **`sql`** - Raw read-only SQL queries against local database
4. **`pages`** (local mode) - Query pages table with filters
5. **`blocks`** (local mode) - Query blocks table with filters

## Sources

- https://developers.notion.com/reference/page
- https://developers.notion.com/reference/block
- https://developers.notion.com/reference/database
- https://developers.notion.com/reference/user
- https://developers.notion.com/reference/comment-object
- https://developers.notion.com/reference/post-search (filter.timestamp validation)
- https://developers.notion.com/reference/request-limits
- Phase 0 and Phase 0.5 artifacts
