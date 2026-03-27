---
title: "Power User Workflows: Notion CLI"
type: feat
status: active
date: 2026-03-26
phase: "0.5"
api: "Notion"
---

# Power User Workflows: Notion CLI

## Overview

Notion maps to the **Content** API archetype (documents, pages, blocks, media) with strong **Project Management** overlap (databases as task trackers). The API has 55+ REST endpoints but no webhooks or real-time capabilities, making local persistence and offline querying the key differentiator for a power-user CLI.

The 4ier/notion-cli (90 stars, 39 commands) is the current market leader but is purely an API wrapper - it cannot search content offline, detect stale pages, diff database changes, or run cross-database queries. These workflow commands, powered by a local SQLite data layer, are what transform an API wrapper into a productivity tool.

## API Archetype Classification

| Archetype | Match Level | Signals |
|-----------|-------------|---------|
| **Content** | PRIMARY | Pages, blocks, databases, markdown endpoints, file uploads |
| **Project Management** | SECONDARY | Database views as task boards, filters, sorts, status properties |

Expected workflow categories:
- Backup to local files, diff, template management, publish workflows (Content)
- Stale detection, orphan pages, activity analytics, bulk state changes (PM)

## All Workflow Ideas (13)

### 1. `search` - Offline Full-Text Search (Score: 12/12)

**What it does:** Search across ALL synced page titles, block content, and database entries using local FTS5 index.

**API calls:** NONE - queries local SQLite only. Requires prior `sync`.

**Why it matters:** Notion's API search (`POST /search`) only searches page/database titles. It cannot search block content. Local FTS5 unlocks true full-text search across everything.

**CLI signature:**
```bash
notion-cli search "quarterly review" --type page --parent <db-id> --days 30
notion-cli search "TODO" --type block --json --select id,content,page_id
```

**Scoring:** Frequency=3 (daily), Pain=3 (API search is title-only), Feasibility=3 (SQLite FTS5 is trivial), Uniqueness=3 (no tool does this)

---

### 2. `sync` - Incremental Workspace Sync (Score: 11/12)

**What it does:** Sync all pages, databases, and blocks to local SQLite. Incremental via `last_edited_time` cursor.

**API calls:**
1. `POST /search` with filter to get recently modified pages
2. `GET /pages/{id}` for each changed page
3. `GET /blocks/{id}/children` recursively for each page
4. `GET /databases/{id}` for database schemas

**Validation:** The search endpoint supports `filter.timestamp` with `last_edited_time` for incremental queries. Pagination via `start_cursor`. Rate limit: 3 req/sec means ~10,800 pages/hour metadata, slower for full block content.

**CLI signature:**
```bash
notion-cli sync                           # full sync
notion-cli sync --since "2 days ago"      # incremental
notion-cli sync --database <id>           # sync specific database
notion-cli sync --depth 2                 # limit block depth
```

**Scoring:** Frequency=3, Pain=3, Feasibility=2 (rate limiting complicates), Uniqueness=3

---

### 3. `stale` - Find Stale Content (Score: 11/12)

**What it does:** Find pages not edited in N days, grouped by parent database/page.

**API calls:** NONE - queries local SQLite. Requires prior `sync`.

**CLI signature:**
```bash
notion-cli stale --days 90                    # all stale pages
notion-cli stale --days 30 --database <id>    # stale in specific DB
notion-cli stale --json --select title,last_edited_time,parent_id
```

**Scoring:** Frequency=2, Pain=2, Feasibility=3, Uniqueness=3 (no tool does this). Note: slightly lower frequency/pain but very high uniqueness.

---

### 4. `export` - Export to Markdown Files (Score: 11/12)

**What it does:** Export pages to local markdown files preserving hierarchy.

**API calls:**
1. `GET /pages/{id}/markdown` for each page
2. `GET /blocks/{id}/children` to discover child pages
3. Write `.md` files to disk with directory structure matching page hierarchy

**Validation:** The new markdown endpoint (`GET /pages/{id}/markdown`) is available in API version 2026-03-11. This is a first-class API feature.

**CLI signature:**
```bash
notion-cli export <page-id> -o ./docs/          # single page
notion-cli export --database <id> -o ./tasks/    # all pages in a DB
notion-cli export --all -o ./backup/             # full workspace
notion-cli export <page-id> --format html        # HTML export
```

**Scoring:** Frequency=2, Pain=3, Feasibility=3, Uniqueness=2 (notion2md exists but doesn't use new API)

---

### 5. `diff` - Database Change Detection (Score: 11/12)

**What it does:** Compare current database state against last sync snapshot. Show added, removed, and modified entries.

**API calls:**
1. `POST /databases/{id}/query` to get current state
2. Compare against local SQLite snapshot from last sync

**CLI signature:**
```bash
notion-cli diff <database-id>                    # show all changes
notion-cli diff <database-id> --since "1 week"   # changes in timeframe
notion-cli diff <database-id> --json             # machine-readable
```

**Scoring:** Frequency=2, Pain=3, Feasibility=2 (snapshot comparison logic), Uniqueness=3

---

### 6. `stats` - Workspace Analytics (Score: 11/12)

**What it does:** Show workspace statistics - page count, database count, block count, most active pages, user activity distribution, content age distribution.

**API calls:** NONE - queries local SQLite. Requires prior `sync`.

**CLI signature:**
```bash
notion-cli stats                         # workspace overview
notion-cli stats --database <id>         # database-specific stats
notion-cli stats --json                  # machine-readable
notion-cli stats --top 10               # top 10 most active pages
```

**Scoring:** Frequency=2, Pain=2, Feasibility=3, Uniqueness=3

---

### 7. `import` - Push Markdown to Notion (Score: 10/12)

**What it does:** Push local markdown files back to Notion pages, creating or updating.

**API calls:**
1. `PATCH /pages/{id}/markdown` to update existing pages
2. `POST /pages` + `PATCH /pages/{id}/markdown` to create new pages

**Validation:** The PATCH markdown endpoint is available. Creating new pages requires a parent (database or page).

**CLI signature:**
```bash
notion-cli import ./docs/meeting-notes.md --parent <page-id>
notion-cli import ./blog/ --database <id>    # bulk import directory
notion-cli import --update <page-id> ./updated.md
```

**Scoring:** Frequency=2, Pain=3, Feasibility=2, Uniqueness=2

---

### 8. `orphans` - Find Unlinked Pages (Score: 10/12)

**What it does:** Find pages that aren't linked from any other page or database - content islands that may be abandoned.

**API calls:** NONE - queries local SQLite for pages with no incoming references.

**CLI signature:**
```bash
notion-cli orphans                       # all orphan pages
notion-cli orphans --json --select title,created_time
```

**Scoring:** Frequency=1, Pain=2, Feasibility=2, Uniqueness=3

---

### 9. `tree` - Visual Page Hierarchy (Score: 10/12)

**What it does:** Display the workspace page hierarchy as a tree, showing nesting depth.

**API calls:** NONE - queries local SQLite.

**CLI signature:**
```bash
notion-cli tree                          # full workspace tree
notion-cli tree <page-id> --depth 3     # subtree from specific page
notion-cli tree --databases-only         # just databases
```

**Scoring:** Frequency=2, Pain=2, Feasibility=3, Uniqueness=2

---

### 10. `watch` - Change Polling Stream (Score: 10/12)

**What it does:** Poll for changes and emit NDJSON stream. Poor man's webhooks.

**API calls:** Loop: `POST /search` with `filter.timestamp.last_edited_time.after` cursor.

**CLI signature:**
```bash
notion-cli watch --interval 30s          # poll every 30s
notion-cli watch --database <id>         # watch specific DB
notion-cli watch | jq '.title'           # pipe to processors
```

**Scoring:** Frequency=3, Pain=3, Feasibility=1 (rate limit makes this expensive), Uniqueness=3

---

### 11. `archive` - Bulk Archive Pages (Score: 9/12)

**What it does:** Archive pages matching a filter (e.g., all completed tasks older than 30 days).

**API calls:** `PATCH /pages/{id}` with `archived: true` for each matching page.

**CLI signature:**
```bash
notion-cli archive --database <id> --filter "Status=Done" --older-than 30d
notion-cli archive --dry-run             # preview without archiving
```

**Scoring:** Frequency=1, Pain=2, Feasibility=2, Uniqueness=2. Note: `--dry-run` is critical for safety.

---

### 12. `template` - Create from Template (Score: 8/12)

**What it does:** Create a new page pre-populated from a template page's blocks.

**API calls:**
1. `GET /blocks/{template-id}/children` recursively
2. `POST /pages` with parent
3. `PATCH /blocks/{new-page-id}/children` with template blocks

**CLI signature:**
```bash
notion-cli template <template-id> --parent <db-id> --title "Weekly Standup 03-26"
```

**Scoring:** Frequency=2, Pain=2, Feasibility=2, Uniqueness=1

---

### 13. `migrate` - Cross-Database Migration (Score: 7/12)

**What it does:** Copy/move entries from one database to another, mapping properties.

**Scoring:** Frequency=1, Pain=1, Feasibility=1, Uniqueness=2. Too complex for v1, skip.

## Selected Top 7 for Phase 4 Implementation

| Priority | Command | Score | Key Value |
|----------|---------|-------|-----------|
| 1 | `sync` | 11 | Foundation for everything else |
| 2 | `search` | 12 | True full-text search (API can't do this) |
| 3 | `export` | 11 | Markdown export using new API endpoints |
| 4 | `stale` | 11 | Knowledge base hygiene |
| 5 | `diff` | 11 | Database change tracking |
| 6 | `stats` | 11 | Workspace analytics |
| 7 | `import` | 10 | Markdown round-trip completion |

**Note:** `sync` must be Priority 1 because `search`, `stale`, `diff`, and `stats` all depend on local data.

## Implementation Dependencies

```
sync ─────┬──> search (FTS5 queries)
          ├──> stale (last_edited_time queries)
          ├──> diff (snapshot comparison)
          ├──> stats (aggregate queries)
          └──> orphans, tree (relationship queries)

export ──────> standalone (uses API directly)
import ──────> standalone (uses API directly)
```

## Sources

- https://developers.notion.com/llms.txt (full endpoint list)
- https://developers.notion.com/reference/post-search
- https://developers.notion.com/reference/request-limits
- Phase 0 visionary research artifact
