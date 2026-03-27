---
title: "Power User Workflows: Notion CLI"
type: feat
status: active
date: 2026-03-26
phase: "0.5"
api: "notion"
---

# Power User Workflows: Notion CLI

## API Archetype: Content

Notion is a **Content** platform - documents, pages, blocks, media. The archetype signals: backup to local files, diff, template management, publish workflows, offline search.

Secondary archetype: **Project Management** - databases act as task trackers with status, assignee, dates. Signals: stale detection, velocity, triage.

## 15 Workflow Ideas

### Content Workflows

1. **workspace-backup** - Sync entire workspace to local SQLite + export to Markdown/HTML files
   - Steps: List all pages -> fetch blocks recursively -> store in SQLite -> export to filesystem
   - API calls: search (paginated) + get page + get block children (recursive) per page
   - Pain: Manual, requires handling pagination + recursion + rate limits + 2-level nesting limit
   - Proposed: `notion-cli backup --format md --output ./backup --since "2026-03-01"`

2. **search** - Full-text search across all synced content (titles + block text)
   - Steps: Query local FTS5 index
   - API calls: ZERO (uses local SQLite)
   - Pain: Notion's API search is title-only, slow, and rate-limited
   - Proposed: `notion-cli search "quarterly review" --type page --limit 20`

3. **export** - Export a single page or database to Markdown/JSON/HTML
   - Steps: Get page -> get blocks recursively -> convert to target format
   - API calls: get page + get block children (recursive)
   - Pain: No single-command export. Block-to-markdown conversion is complex.
   - Proposed: `notion-cli export <page-id> --format md --output ./page.md`

4. **diff** - Compare local cached version of a page against current API state
   - Steps: Fetch current blocks -> compare against SQLite cache -> show changes
   - API calls: get page + get block children
   - Pain: No version history in API. Manual comparison.
   - Proposed: `notion-cli diff <page-id>`

5. **tail** - Watch for recent changes across workspace
   - Steps: Poll search endpoint with last_edited_time filter -> show new/changed pages
   - API calls: search (periodic polling)
   - Pain: No webhooks/SSE. Manual polling.
   - Proposed: `notion-cli tail --interval 30s`

### Database/Project Management Workflows

6. **db-query** - Query a database with human-friendly filters (no JSON)
   - Steps: Parse CLI filter syntax -> build Notion filter object -> POST query
   - API calls: POST database query
   - Pain: JSON filter syntax is verbose and error-prone
   - Proposed: `notion-cli db query <id> --filter "Status=Done" --sort "-Created"`

7. **stale** - Find pages/database items not updated in N days
   - Steps: Query local SQLite for pages where last_edited_time < threshold
   - API calls: ZERO (uses local data)
   - Pain: No built-in way to find abandoned content
   - Proposed: `notion-cli stale --days 30 --type page`

8. **orphans** - Find pages with no parent database or that aren't linked from anywhere
   - Steps: Query local page/block tables -> find pages not referenced as children
   - API calls: ZERO (uses local data)
   - Pain: Workspaces accumulate orphaned pages over time
   - Proposed: `notion-cli orphans`

9. **stats** - Workspace statistics: page count, database count, block count, users, last activity
   - Steps: Query local SQLite aggregates
   - API calls: ZERO (uses local data)
   - Pain: No workspace-level overview in Notion
   - Proposed: `notion-cli stats`

10. **bulk-update** - Update a property across multiple database items matching a filter
    - Steps: Query database -> filter results -> PATCH each matching page
    - API calls: POST query + N PATCH calls
    - Pain: No bulk update API. Must loop manually with rate limiting.
    - Proposed: `notion-cli bulk-update <db-id> --filter "Status=In Progress" --set "Status=Done"`

### Publishing/Integration Workflows

11. **publish** - Export a page tree to a static site directory (Markdown + images)
    - Steps: Get page -> get children recursively -> download images -> write files
    - API calls: get page + get blocks + file downloads
    - Pain: Blog/docs publishing from Notion is a multi-step manual process
    - Proposed: `notion-cli publish <page-id> --output ./site --format md`

12. **import** - Create a page from a Markdown file
    - Steps: Parse markdown -> convert to Notion blocks -> POST create page
    - API calls: POST create page + PATCH append blocks (chunked for >100 blocks)
    - Pain: No Markdown import in API. Must convert MD to block JSON.
    - Proposed: `echo "# Hello" | notion-cli import --parent <page-id> --stdin`

### Maintenance Workflows

13. **doctor** - Validate auth, API version, rate limit status, workspace access
    - Steps: GET users/me -> check token validity -> test a search call
    - API calls: GET users/me + POST search
    - Pain: Debugging auth issues is painful
    - Proposed: `notion-cli doctor`

14. **tree** - Show page hierarchy as a tree (like `tree` command)
    - Steps: Get root pages -> recurse block children -> display tree
    - API calls: Multiple get block children calls (or use local data)
    - Proposed: `notion-cli tree <page-id> --depth 3`

15. **sql** - Run raw SQL against local SQLite database
    - Steps: Open SQLite -> execute query -> format output
    - API calls: ZERO
    - Proposed: `notion-cli sql "SELECT title, last_edited FROM pages WHERE type='page' ORDER BY last_edited DESC LIMIT 10"`

## Validation Against API Capabilities

| Workflow | Required Endpoints | Validated? | Notes |
|----------|-------------------|-----------|-------|
| backup | POST /search, GET /pages/{id}, GET /blocks/{id}/children | Yes | Search returns all pages. Block children are paginated. |
| search | Local SQLite only | Yes | No API calls needed |
| export | GET /pages/{id}, GET /blocks/{id}/children | Yes | Recursive block traversal needed |
| diff | GET /pages/{id}, GET /blocks/{id}/children | Yes | Compare against local cache |
| tail | POST /search with filter | Partial | Search supports filter by last_edited_time via sort, not direct filtering. Must sort by last_edited_time DESC and check timestamps. |
| db-query | POST /data_sources/{id}/query | Yes | Supports filter and sort objects |
| stale | Local SQLite only | Yes | Query pages.last_edited_time |
| orphans | Local SQLite only | Yes | Cross-reference parent_id in blocks/pages |
| stats | Local SQLite only | Yes | COUNT/GROUP BY queries |
| bulk-update | POST query + PATCH /pages/{id} | Yes | Rate limit: ~3 req/sec average |
| publish | GET /pages/{id}, GET /blocks/{id}/children | Yes | Need image download too |
| import | POST /pages, PATCH /blocks/{id}/children | Yes | 100 block limit per request, must chunk |
| doctor | GET /users/me, POST /search | Yes | Basic connectivity check |
| tree | GET /blocks/{id}/children (recursive) | Yes | Depth-limited recursion |
| sql | Local SQLite only | Yes | Read-only passthrough |

## Impact Scoring

| Workflow | Frequency | Pain | Feasibility | Uniqueness | Total |
|----------|-----------|------|-------------|-----------|-------|
| backup | 2 (weekly) | 3 (high) | 2 (medium - lots of API calls) | 1 (partial - backup tools exist but no CLI with SQLite) | 8 |
| search | 3 (daily) | 3 (high) | 3 (easy - just FTS5 query) | 3 (no CLI does this) | 12 |
| export | 2 (weekly) | 3 (high) | 2 (medium - block conversion) | 2 (partial - notion2md exists) | 9 |
| diff | 1 (monthly) | 2 (medium) | 2 (medium) | 3 (no tool does this) | 8 |
| tail | 2 (weekly) | 2 (medium) | 2 (medium - polling) | 3 (no tool does this) | 9 |
| db-query | 3 (daily) | 3 (high) | 3 (easy) | 2 (4ier has basic version) | 11 |
| stale | 1 (monthly) | 2 (medium) | 3 (easy - local query) | 3 (no tool does this) | 9 |
| orphans | 1 (monthly) | 2 (medium) | 2 (medium - needs parent tracking) | 3 (no tool does this) | 8 |
| stats | 2 (weekly) | 1 (low) | 3 (easy - local aggregates) | 3 (no tool does this) | 9 |
| bulk-update | 1 (monthly) | 3 (high) | 2 (medium - rate limits) | 3 (no CLI does this) | 9 |
| publish | 1 (monthly) | 2 (medium) | 1 (hard - image handling) | 2 (partial) | 6 |
| import | 1 (monthly) | 2 (medium) | 1 (hard - MD-to-blocks) | 2 (partial) | 6 |
| doctor | 2 (weekly) | 1 (low) | 3 (easy) | 0 (4ier has it) | 6 |
| tree | 2 (weekly) | 1 (low) | 2 (medium) | 3 (no tool does this) | 8 |
| sql | 3 (daily) | 1 (low) | 3 (easy) | 3 (no tool does this) | 10 |

## Top 7 for Implementation (Score >= 8)

1. **search** (12/12) - Full-text search across synced content using FTS5
2. **db-query** (11/12) - Human-friendly database queries without JSON
3. **sql** (10/12) - Raw SQL against local SQLite
4. **backup/sync** (8/12) - Incremental workspace sync to SQLite
5. **export** (9/12) - Page-to-Markdown export with recursive blocks
6. **stale** (9/12) - Find pages untouched for N days
7. **tail** (9/12) - Watch workspace for recent changes

These 7 workflows become mandatory Phase 4 work items.

## Sources
- Notion API docs: https://developers.notion.com/reference
- 4ier/notion-cli commands: https://github.com/4ier/notion-cli
- Thomas Frank automation guide: https://thomasjfrank.com/notion-automations/
