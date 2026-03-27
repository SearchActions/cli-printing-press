---
title: "feat: Apply discrawl v0.2.0 Learnings to Printing Press Templates"
type: feat
status: completed
date: 2026-03-26
---

# feat: Apply discrawl v0.2.0 Learnings to Printing Press Templates

## Overview

Peter Steinberger shipped discrawl v0.2.0 on 2026-03-26 with 7 sync performance commits that made it a daily-driver tool for Discord archive analysis. Matt uses it every day on the OpenClaw Discord. This plan extracts the specific architectural patterns that make discrawl fast and reliable, and applies them to the printing press generator templates so every generated CLI inherits these patterns automatically.

The gap is stark: discrawl's sync is concurrent, resumable, rate-limit-aware, and incrementally smart. The printing press sync template does a single GET per resource with no pagination, no rate limiting, no cursor tracking, and no incremental logic. Hand-written CLIs (discord-cli, notion-cli, linear-cli) each independently reinvented the same patterns discrawl ships out of the box.

## Problem Statement / Motivation

The printing press `sync.go.tmpl` generates a sync command that:
1. Does a **single GET per resource** with no pagination loop
2. Has a `--full` flag that is **dead code** (never used in the template logic)
3. Stores empty cursor `""` in sync_state (no incremental sync)
4. Has **no rate limiting** (no sleep, no backoff, no Retry-After handling)
5. Has **no concurrency** (serial fetch, one resource at a time)
6. Has **no progress reporting** (large syncs appear hung)
7. Uses generic `Upsert()` per row (no batch transactions)

Every hand-written CLI has independently fixed these issues, proving the patterns are universally needed, not domain-specific.

## What discrawl v0.2.0 Teaches Us

### 7 Specific Improvements to Extract

| # | discrawl Pattern | Template Gap | Priority |
|---|-----------------|-------------|----------|
| 1 | **Cursor-based pagination loop** with resume from `sync_state.BackfillCursor` | Template does single GET, no pagination | P0 - without this, sync is useless for any API with >1 page of data |
| 2 | **Batched SQLite transactions** - one `BeginTx/Commit` per batch, not per row | Template calls `Upsert()` per individual record | P0 - 10-100x write performance on bulk sync |
| 3 | **Rate limit handling** - exponential backoff with Retry-After header respect | Template has no rate limit handling at all | P0 - without this, bulk syncs hit 429s and fail |
| 4 | **Incremental sync** - `--since` flag that filters by `updatedAt > last_sync_at` | Template's `--full` flag is dead code | P1 - makes daily syncs fast instead of re-downloading everything |
| 5 | **Concurrent channel workers** - configurable worker pool for parallel resource sync | Template syncs resources serially | P1 - 4-8x faster on APIs with many independent resources |
| 6 | **Skip-logic for complete resources** - check sync_state before making API calls | Template always fetches everything | P1 - zero API calls for already-synced resources |
| 7 | **Progress reporting** - periodic stderr updates during long syncs | Template is silent during sync | P2 - UX improvement for daily-driver usage |

### 3 Architectural Patterns to Adopt

**A. Separate read and write DB connections**
discrawl opens one writer connection (`MaxOpenConns(1)`) and a separate read-only connection (`mode=ro&_pragma=query_only(1)`). This means `search` and `sql` commands don't block during `sync` or `tail`. The template currently uses a single connection.

**B. FTS rowid optimization**
discrawl converts string IDs to int64 rowids via `strconv.ParseInt` (for numeric IDs) or FNV-64a hashing (for UUIDs). This enables deterministic `DELETE FROM fts WHERE rowid = ?` followed by `INSERT` for updates, avoiding the common FTS5 duplicate-row bug. The template uses auto-assigned rowids via `content='table'` triggers which can desync on updates.

**C. Tuned SQLite PRAGMAs**
discrawl uses `synchronous(NORMAL) + temp_store(MEMORY) + mmap_size(268435456)` in addition to WAL. The template only sets WAL + busy_timeout. The additional pragmas give 2-3x write throughput improvement.

## Proposed Solution

Update 4 generator templates to incorporate discrawl's patterns:

### 1. `sync.go.tmpl` - Complete rewrite

Replace the single-GET loop with a proper pagination engine:

```go
// Pagination loop with cursor tracking
func syncResource(c *client.Client, db *store.Store, resource, path string, opts syncOpts) error {
    cursor := ""
    if !opts.full {
        cursor = db.GetSyncCursor(resource)
    }

    total := 0
    for {
        params := map[string]string{"limit": opts.pageSize}
        if cursor != "" {
            params[opts.cursorParam] = cursor  // "after", "start_cursor", "offset", etc.
        }
        if opts.since != "" {
            params[opts.sinceParam] = opts.since  // "updated_after", "since", etc.
        }

        data, err := c.Get(path, params)
        if err != nil {
            return fmt.Errorf("fetching %s: %w", resource, err)
        }

        items := extractItems(data, opts.itemsKey)  // "data", "results", root array
        if len(items) == 0 {
            break
        }

        // Batch upsert in single transaction
        if err := db.UpsertBatch(resource, items); err != nil {
            return fmt.Errorf("storing %s: %w", resource, err)
        }

        total += len(items)
        fmt.Fprintf(os.Stderr, "\r  %s: %d synced", resource, total)

        // Update cursor for resume
        cursor = extractCursor(data, items, opts)
        db.SaveSyncCursor(resource, cursor)

        if !hasMorePages(data, items, opts) {
            break
        }

        // Rate limit: respect API-specific delays
        time.Sleep(opts.requestDelay)
    }

    fmt.Fprintf(os.Stderr, "\r  %s: %d synced (done)\n", resource, total)
    return nil
}
```

Key additions:
- `--since` flag with human-friendly durations (`7d`, `24h`, `1w`) parsed to RFC3339
- `--concurrency` flag (default: 4 workers for independent resources)
- `--full` flag that actually works (clears sync cursors before starting)
- Progress reporting to stderr
- Configurable page size per API (from spec's `x-pagination` or profiler heuristics)

### 2. `store.go.tmpl` - Batch upsert + tuned pragmas

Add to the store template:

```go
// Tuned pragmas (from discrawl)
pragmas := []string{
    "PRAGMA journal_mode=WAL",
    "PRAGMA synchronous=NORMAL",
    "PRAGMA temp_store=MEMORY",
    "PRAGMA mmap_size=268435456",
    "PRAGMA busy_timeout=5000",
    "PRAGMA foreign_keys=ON",
}

// Batch upsert method
func (s *Store) UpsertBatch(resource string, items []json.RawMessage) error {
    tx, err := s.db.BeginTx(context.Background(), nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for _, item := range items {
        // ... domain-specific upsert logic per table ...
    }

    return tx.Commit()
}
```

### 3. `client.go.tmpl` (in `internal/generator/templates/`) - Rate limit handling

Add exponential backoff to the HTTP client:

```go
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
    for attempt := 0; attempt <= maxRetries; attempt++ {
        resp, err := c.http.Do(req)
        if err != nil {
            return nil, err
        }
        if resp.StatusCode != 429 {
            return resp, nil
        }

        // Respect Retry-After header
        wait := time.Duration(1<<uint(attempt)) * time.Second
        if ra := resp.Header.Get("Retry-After"); ra != "" {
            if secs, err := strconv.Atoi(ra); err == nil {
                wait = time.Duration(secs) * time.Second
            }
        }
        resp.Body.Close()

        if attempt == maxRetries {
            return nil, fmt.Errorf("rate limited after %d retries", maxRetries)
        }
        fmt.Fprintf(os.Stderr, "rate limited, waiting %s...\n", wait)
        time.Sleep(wait)
    }
    return nil, fmt.Errorf("unreachable")
}
```

### 4. Profiler enhancements

Teach the profiler to detect pagination patterns in the spec:

- **Cursor param detection**: Scan query params for `after`, `cursor`, `start_cursor`, `page_token`, `offset`
- **Page size detection**: Scan for `limit`, `per_page`, `page_size`, `first`, `take`
- **Since param detection**: Scan for `since`, `updated_after`, `modified_since`, `updatedAt`
- **Items key detection**: Check response schema for arrays at `.data`, `.results`, `.items`, or root-level array

Store these in the `APIProfile` so templates can use them:

```go
type PaginationProfile struct {
    CursorParam string // "after", "cursor", etc.
    PageSizeParam string // "limit", "per_page", etc.
    SinceParam string // "since", "updated_after", etc.
    ItemsKey string // "data", "results", "" (root array)
    DefaultPageSize int // detected from spec or default 100
}
```

## Technical Considerations

- **Backward compatibility**: Existing generated CLIs should still compile. The template changes add new functionality without breaking the existing API.
- **Performance**: Batched upserts + tuned pragmas should give 10-100x write performance. Concurrent workers should give 4-8x sync speed on multi-resource APIs.
- **Cross-API generality**: The pagination patterns must work for REST (cursor, offset, page), GraphQL (Relay cursors), and Notion-style (start_cursor/has_more). The profiler detects which pattern applies.
- **Testing**: Run the updated templates against Discord, Linear, and Notion specs to verify generated sync actually paginates correctly.

## Acceptance Criteria

- [ ] `sync.go.tmpl` generates a pagination loop with cursor tracking (not single GET)
- [ ] `--since` flag works for incremental sync (filters by last sync time)
- [ ] `--full` flag clears cursors and re-syncs from scratch
- [ ] `--concurrency` flag controls parallel resource sync workers
- [ ] `store.go.tmpl` generates batch upsert in single transaction
- [ ] `store.go.tmpl` includes tuned SQLite pragmas (synchronous=NORMAL, temp_store=MEMORY, mmap_size=256MB)
- [ ] `client.go.tmpl` includes exponential backoff with Retry-After header respect
- [ ] Progress reporting to stderr during sync
- [ ] Profiler detects pagination params from spec (cursor, page_size, since, items_key)
- [ ] Generated discord-cli sync actually paginates through messages (not single page)
- [ ] Proof-of-Behavior verification passes on generated CLIs (sync calls domain-specific UpsertBatch)

## Success Metrics

| Metric | Before | Target |
|--------|--------|--------|
| Sync template pagination | Single GET (1 page) | Full cursor-based pagination |
| Bulk sync write speed | 1 Upsert/row (slow) | 1 transaction/batch (10-100x faster) |
| Rate limit handling | None (429 = crash) | Exponential backoff + Retry-After |
| Incremental sync | Not functional | --since flag with cursor tracking |
| Sync resume after interrupt | Start over | Resume from last cursor |
| Daily sync cost | Full re-download | Only changes since last sync |

## Dependencies & Risks

- **Risk**: Different APIs have different pagination patterns. Mitigation: Profiler detects the pattern; templates use the detected params with sensible defaults.
- **Risk**: Concurrent sync workers could hit global rate limits faster. Mitigation: Default concurrency is conservative (4), users can tune with `--concurrency`.
- **Risk**: Batch transactions could exceed SQLite's memory on very large pages. Mitigation: Batch size is bounded by API page size (typically 50-100 items).
- **Dependency**: Proof-of-Behavior verification (just shipped in `efaec84b`) should validate the new sync patterns.

## Sources & References

### External
- [discrawl v0.2.0 release](https://github.com/steipete/discrawl/releases/tag/v0.2.0) - 7 sync performance commits
- [discrawl repository](https://github.com/steipete/discrawl) - 568 stars, daily-driver Discord archive tool
- discrawl SQLite pragmas: `WAL + synchronous(NORMAL) + temp_store(MEMORY) + mmap_size(268435456)`
- discrawl concurrent sync: configurable worker pool, default `min(32, max(8, GOMAXPROCS*2))`
- discrawl rate limiting: delegated to bwmarrin/discordgo + 45s request timeout + error skipping

### Internal
- Sync template: `internal/generator/templates/sync.go.tmpl` - current single-GET pattern
- Store template: `internal/generator/templates/store.go.tmpl` - WAL + per-row upsert
- Client template: embedded in `internal/generator/generator.go` - no retry logic
- Profiler: `internal/profiler/profiler.go` - HighVolume/NeedsSearch detection
- Discord CLI sync: `discord-cli/internal/cli/sync.go` - hand-written cursor pagination
- Notion CLI sync: `notion-cli/internal/cli/sync.go` - hand-written POST /search pagination
- Linear CLI sync: `linear-cli/internal/cli/sync.go` - hand-written GraphQL Relay pagination
- Proof-of-Behavior: `internal/pipeline/verify.go` - validates sync calls domain-specific UpsertBatch
