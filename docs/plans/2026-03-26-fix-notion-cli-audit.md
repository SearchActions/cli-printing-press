---
title: "Steinberger Audit: Notion CLI"
type: fix
status: active
date: 2026-03-26
phase: "3"
api: "Notion"
---

# Steinberger Audit: Notion CLI

## Automated Scorecard Baseline: 64/100 (Grade C)

| Dimension | Score |
|-----------|-------|
| Output Modes | 10/10 |
| Auth | 8/10 |
| Error Handling | 10/10 |
| Terminal UX | 10/10 |
| README | 5/10 |
| Doctor | 10/10 |
| Agent Native | 8/10 |
| Local Cache | 10/10 |
| Breadth | 9/10 |
| Vision | 8/10 |
| Workflows | 4/10 |
| Insight | 0/10 |
| Path Validity | 5/10 |
| Auth Protocol | 5/10 |
| Data Pipeline Integrity | 10/10 |
| Sync Correctness | 4/10 |
| Type Fidelity | 2/5 |
| Dead Code | 0/5 |

## Critical Issues Found

### 1. Store Schema Is Wrong (pages, blocks tables)
The pages table has columns from the SPEC body params (page_id, property_id, markdown, query, page_size, start_cursor) instead of the actual API RESPONSE fields (created_time, last_edited_time, created_by_id, last_edited_by_id, url, parent_type, parent_id). The blocks table is a bare JSON blob with no domain columns. No FTS on pages or blocks text.

### 2. Sync Command Won't Work
The sync command calls `GET /pages`, `GET /blocks`, etc. But the Notion API doesn't have a `GET /pages` list endpoint. Pages are discovered via `POST /search`. Blocks require `GET /blocks/{block_id}/children` per page. The sync is fundamentally broken for this API.

### 3. No Workflow or Insight Commands
Workflows scored 4/10 (generic scaffolding only). Insight scored 0/10. Need: stale, diff, stats, orphans, tree commands. These are the PRODUCT.

### 4. Duplicate Search Command
root.go adds `newSearchCmd(&flags)` TWICE (lines 64 and 71), causing the help output to show "search" twice.

### 5. README Missing Cookbook
README scored 5/10. No cookbook section, no data layer examples, no workflow examples.

## GOAT Improvement Plan

### Priority 0: Fix Data Layer (from Phase 0.7 spec)
1. Rewrite store.go with domain-specific tables: pages (with created_time, last_edited_time, url, parent_type, parent_id, title), blocks (with page_id, type, plain_text, parent_type, parent_id), FTS5 on both
2. Rewrite sync.go to use POST /search for page discovery, GET /blocks/{id}/children for block content, extract plain_text from rich_text arrays
3. Add `sql` command for raw read-only queries

### Priority 1: Build Workflow Commands (from Phase 0.5)
1. `search` - rewrite to use FTS5 on local data, not just generic resources_fts
2. `stale` - query pages WHERE last_edited_time < threshold
3. `export` - fetch markdown from API, write to disk files
4. `diff` - compare database state vs local snapshot
5. `stats` - aggregate workspace statistics from local data
6. `import` - push markdown files to Notion pages

### Priority 2: Scorecard Fixes
1. Fix duplicate search command in root.go
2. Add Notion-Version header to client requests
3. README cookbook with data layer + workflow examples
4. Insight commands: health, activity

### Priority 3: Polish
1. Fix help examples with realistic Notion UUIDs
2. Add --stdin examples for pages create, blocks append
