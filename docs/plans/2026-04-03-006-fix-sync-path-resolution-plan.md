---
title: "fix: Sync path resolution for non-paginated list endpoints"
type: fix
status: active
date: 2026-04-03
---

# Fix Sync Path Resolution for Non-Paginated List Endpoints

## Overview

The profiler only adds endpoints to `SyncableResources` when `endpoint.Pagination != nil`. APIs like Steam Web that return data arrays without explicit pagination metadata are never added to the syncable map. The sync template's `syncResourcePath()` then falls back to `"/" + resource`, producing 404s. Fix: expand syncable detection to include all list endpoints, not just paginated ones.

## Problem Frame

The sync template already has the correct path mapping infrastructure (`syncResourcePath()` reads from a map populated by `{{range .SyncableResources}}`). The profiler already detects list endpoints via `isListEndpoint()`. But the connection between them is gated on `endpoint.Pagination != nil`:

```
isListEndpoint() → yes → endpoint.Pagination != nil? → yes → add to syncable
                                                      → NO → skip (BUG)
```

This means ~10-15% of APIs (non-REST or no-pagination) have empty syncable maps, breaking sync, transcendence commands, `--data-source local`, and data pipeline scoring.

## Requirements Trace

- R1. List endpoints without explicit pagination metadata are added to `SyncableResources`
- R2. `syncResourcePath()` returns the spec's actual path for all syncable resources
- R3. Sync fetches non-paginated endpoints once (no cursor loop) without error
- R4. Endpoints requiring unfilled path parameters (e.g., `{steamid}`) are excluded from syncable

## Scope Boundaries

- **Not changing the sync template.** The template already handles non-paginated resources correctly — it breaks out of the pagination loop when results < page size.
- **Not adding path parameter substitution.** Endpoints like `/users/{id}/games` that require a known ID remain unsyncable. That's a separate feature.
- **Not changing pagination detection.** The profiler's pagination detection is correct — this fix is about syncable detection, not pagination detection.

## Key Technical Decisions

- **Move `syncable[resourceName] = endpoint.Path` outside the pagination gate**: The simplest fix. List endpoints detected by `isListEndpoint()` should be syncable regardless of whether they have pagination metadata. The sync template's pagination loop already handles single-page responses gracefully.

- **Exclude endpoints with unfilled path parameters**: If `endpoint.Path` contains `{`, the endpoint requires a runtime parameter the sync command can't provide. Skip these.

## Implementation Units

- [ ] **Unit 1: Expand syncable resource detection in the profiler**

**Goal:** Add all list endpoints (not just paginated ones) to `SyncableResources`, excluding those with path parameters.

**Requirements:** R1, R2, R3, R4

**Dependencies:** None

**Files:**
- Modify: `internal/profiler/profiler.go`
- Test: `internal/profiler/profiler_test.go`

**Approach:**
- In the `walk` function, move the `syncable[resourceName] = endpoint.Path` assignment out of the `endpoint.Pagination != nil` block and into the `isListEndpoint()` block
- Preserve the existing shortest-path guard: only set `syncable[resourceName]` if the key is absent or the new path is shorter
- Add a guard: skip if `strings.Contains(endpoint.Path, "{")` — these require path params sync can't provide
- Keep the enum expansion logic inside the pagination block (it depends on pagination for page-through)
- The pagination-specific counter (`p.ListEndpoints++`) stays inside the pagination block — it counts paginated endpoints specifically
- Update the `SyncableResource` doc comment from "paginated list sync" to "list sync (paginated or single-page)"

**Patterns to follow:**
- Existing `isListEndpoint()` detection at `profiler.go`
- Existing `syncable` map population pattern

**Test scenarios:**
- Happy path: spec with paginated list endpoints → syncable resources populated with correct paths (existing behavior preserved)
- Happy path: spec with non-paginated list endpoints (like Steam's `/ISteamApps/GetAppList/v2/`) → added to syncable with correct path
- Edge case: endpoint with path parameter (`/users/{id}/games`) → excluded from syncable
- Edge case: multiple list endpoints for same resource → shortest path wins (existing behavior)
- Integration: generate a CLI from Steam spec → `defaultSyncResources()` includes Steam interfaces, `syncResourcePath()` returns correct paths

**Verification:**
- `go test ./internal/profiler/...` passes
- Generate a CLI from the Steam spec → `syncResourcePath("isteam-apps")` returns `/ISteamApps/GetAppList/v2/` not `/isteam-apps`
- Sync on a Steam CLI fetches from the correct API path

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Non-paginated endpoints may return very large responses | The sync template already has a `limit` param in the request — if the API ignores it, the response may be large but won't loop. Acceptable for a sync operation. |
| Some list endpoints may not be useful to sync (e.g., internal/admin endpoints) | The profiler's `isListEndpoint()` is already conservative. Adding path-param exclusion further narrows the set. |

## Sources & References

- Related code: `internal/profiler/profiler.go` (lines 237-264, syncable population)
- Related code: `internal/generator/templates/sync.go.tmpl` (lines 441-454, syncResourcePath)
- Retro: Steam run 4 retro (sync path issues)
- Retro: Steam run 2 retro (sync correctness findings)
