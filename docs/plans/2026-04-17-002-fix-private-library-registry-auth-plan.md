---
title: "fix: Private library registry auth for mega MCP"
type: fix
status: active
date: 2026-04-17
---

# fix: Private library registry auth for mega MCP

## Overview

The `printing-press-mcp` binary fetches `registry.json` and per-API `tools-manifest.json` files from `raw.githubusercontent.com/mvanhorn/printing-press-library/main/...` via unauthenticated HTTP. The library repo is private, so every fetch returns 404 and the mega MCP ships with an empty tool catalog. Fix: attach `Authorization: token $GITHUB_TOKEN` when the env var is set. Same URL, same response shape, just with auth.

## Problem Frame

Two `http.Get` calls need the header:

- `FetchRegistry` in `internal/megamcp/registry.go:18`
- `fetchManifestData` in `internal/megamcp/manifest.go:283`

Single user (the maintainer), single private repo, token read from `GITHUB_TOKEN` the same way `internal/crowdsniff/github.go` already does it.

## Scope Boundaries

- **In scope:** Add `Authorization` header to the two fetch calls. One test that the header gets set when `GITHUB_TOKEN` is present.
- **Out of scope:** Shared auth helper, upgraded error diagnostics, doc reconciliation, tests that the token never appears in error strings, a `library_status` meta-tool, flipping repo visibility.

## Key Technical Decisions

- **Env var is `GITHUB_TOKEN`.** Matches `internal/crowdsniff/github.go:19,43` convention. No new env var.
- **Inline at each call site.** Two fetches, ~5 lines each. A shared helper adds more surface than it saves.
- **Keep existing error shapes.** `HTTP 404` on 404 is fine. Maintainer reading their own repo knows what that means.

## Implementation Units

- [ ] **Unit 1: Attach GITHUB_TOKEN auth header to registry and manifest fetches**

**Goal:** Both outbound GETs carry `Authorization: token <value>` when `GITHUB_TOKEN` is set. Nothing else changes.

**Files:**
- Modify: `internal/megamcp/registry.go`
- Modify: `internal/megamcp/manifest.go`
- Modify: `internal/megamcp/registry_test.go` (add one auth-header assertion)

**Approach:**
- In both files, replace `http.Get(url)` with a `http.NewRequest("GET", url, nil)` + `http.DefaultClient.Do(req)` pair, and set the `Authorization` header when `os.Getenv("GITHUB_TOKEN")` is non-empty.
- Use `token <value>` format (matches the crowdsniff package, not `Bearer`).
- Don't wrap in a helper - two call sites isn't enough to justify the indirection.

**Patterns to follow:**
- `internal/crowdsniff/github.go` for the token-read + header-set pattern.

**Test scenarios:**
- Happy path: with `GITHUB_TOKEN` set via `t.Setenv`, the `httptest.NewServer` handler observes `Authorization: token <value>` on the inbound request. (Extend an existing `registry_test.go` case - don't add a new file.)

**Verification:**
- `go test ./internal/megamcp/...` passes.
- Rebuild (`go install ./cmd/printing-press-mcp`), restart Claude Code, `library_info` returns a non-empty API list.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Token missing or scope-insufficient still returns 404, with no hint why | Accepted. Maintainer context - you know what 404 means on a private repo read. |
| Future contributor adds a third fetch and forgets the header | Accepted for now. Revisit if a third fetch path lands. |

## Sources & References

- `internal/megamcp/registry.go:15-39`
- `internal/megamcp/manifest.go:282-299`
- `internal/crowdsniff/github.go:17-44`
- Repo visibility: `gh repo view mvanhorn/printing-press-library --json visibility` -> `PRIVATE` as of 2026-04-17.
