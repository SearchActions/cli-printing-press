---
title: "Steinberger Audit: Notion CLI"
type: fix
status: active
date: 2026-03-26
phase: "3"
api: "notion"
---

# Steinberger Audit: Notion CLI

## Automated Scorecard Baseline

Total: **63/100 (Grade C)**

### Dimension Breakdown

| Dimension | Score | What 10 Looks Like | How to Get There |
|-----------|-------|-------------------|-----------------|
| Output Modes | 10/10 | gogcli: --json, --csv, --plain, --quiet, --compact, --select | Already at 10 - all modes present |
| Auth | 8/10 | gogcli: OAuth flow, token storage, profiles, doctor validates | Add profile switching, improve doctor to test actual API call |
| Error Handling | 10/10 | gogcli: typed exits, retry with backoff, helpful hints | Already at 10 - classifyAPIError with hints |
| Terminal UX | 9/10 | gogcli: progress spinners, color themes, pager | Add pager for long output |
| README | 5/10 | gogcli: install, quickstart, every command with example, cookbook, FAQ | Add cookbook with data layer examples, FAQ, all commands documented |
| Doctor | 10/10 | gogcli: validates auth, API version, rate limits, config health | Already at 10 |
| Agent Native | 8/10 | gogcli: --json, --select, --dry-run, --stdin, --yes, --no-input, typed exits | Add --stdin support to create/update commands |
| Local Cache | 10/10 | gogcli: SQLite + FTS5, --no-cache bypass, sync state | Already at 10 (structure exists) |
| Breadth | 5/10 | 45+ commands covering every API endpoint + workflows | Missing file uploads, views, markdown endpoints |
| Vision | 8/10 | discrawl: domain-specific SQLite + FTS5 + sync + search + workflows | Data layer exists but generic. Needs domain-specific tables |
| Workflows | 4/10 | 5+ compound workflow commands | Only archive + status. Need stale, orphans, stats, sql, export, tail |
| Insight | 0/10 | health, stale, orphans, stats, velocity | Zero insight commands exist |

### Domain Correctness Breakdown

| Dimension | Score | Notes |
|-----------|-------|-------|
| Path Validity | 5/10 | Notion API uses /v1/ prefix but sync hits wrong paths |
| Auth Protocol | 5/10 | Bearer token correct, but Notion-Version header may be missing |
| Data Pipeline Integrity | 7/10 | Store exists, generic upsert works, but domain tables are JSON blobs |
| Sync Correctness | 8/10 | Incremental sync logic is sound, but targets wrong resources |
| Type Fidelity | 2/5 | Many complex body fields skipped |
| Dead Code | 0/5 | Ghost tables (move, search), likely dead functions |

## Critical Issues

### 1. Store is Generic JSON Blobs (Priority 0)
The `resources` table stores everything as undifferentiated JSON blobs. No domain-specific columns for pages (title, parent_id, last_edited_time), blocks (type, plain_text, page_id), or users (name, email). FTS5 indexes raw JSON instead of extracted text.

**Fix:** Replace with Phase 0.7 domain-specific schema.

### 2. Sync Targets Wrong Resources (Priority 0)
`defaultSyncResources()` returns `["children", "comments", "properties", "templates", "users"]` - these are sub-resources, not top-level. Should sync pages via POST /v1/search, then blocks per page.

**Fix:** Rewrite sync to use POST /v1/search for page discovery, GET /v1/blocks/{id}/children for block retrieval.

### 3. Zero Insight/Workflow Commands (Priority 1)
Only `workflow archive` and `workflow status` exist. Missing: stale, orphans, stats, sql, export-md, tail, search (local FTS5).

**Fix:** Implement all 7 Phase 0.5 workflow commands.

### 4. Ghost Tables (Priority 2)
Tables `move`, `search`, `query` are created but likely never populated via sync.

**Fix:** Remove or wire to real sync paths.

### 5. README Incomplete (Priority 3)
5/10 - missing cookbook, FAQ, data layer examples.

## GOAT Improvement Plan

1. **Replace store.go** with domain-specific schema from Phase 0.7 (pages, blocks, database_items, users, databases, comments tables with proper columns + FTS5)
2. **Rewrite sync.go** to use Notion's search API for page discovery + recursive block fetching
3. **Add `search` command** - FTS5 full-text search across pages + block content
4. **Add `sql` command** - raw read-only SQL against local SQLite
5. **Add `stale` command** - find pages not edited in N days
6. **Add `orphans` command** - find pages with no children
7. **Add `stats` command** - workspace statistics from local data
8. **Add `tail` command** - watch for workspace changes via polling
9. **Add `export` command** - page-to-markdown export
10. **Expand README** with cookbook and FAQ
