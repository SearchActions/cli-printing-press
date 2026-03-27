---
title: "Printing Press Launch Readiness - From B+ to GOAT"
type: feat
status: active
date: 2026-03-24
---

# Printing Press Launch Readiness - From B+ to GOAT

## E2E Test Learnings (Tonight's Session)

What we proved tonight:
- Full 6-phase pipeline works end-to-end: init -> preflight -> scaffold -> enrich -> regenerate -> review -> ship
- Petstore scored 85/100 (Grade B) - static 50/50, dogfood 35/50
- Thin seed architecture works - seeds are prompts, ce:plan adds intelligence
- Codex delegation works for phase execution (~250K tokens across 6 phases)
- Budget gate correctly returns CONTINUE (0.2h elapsed for full run)
- state.json + plan_status tracking works correctly through all transitions

What broke or was surprising:
- `--force` flag missing from `generate` command (Codex had to rm -rf + regenerate)
- CLI uses raw operationId verbs (`find-by-status`, `get-by-id`) not friendly aliases (`list`, `get`)
- Tier 3 dogfood (create pet) got HTTP 500 from upstream petstore server
- `doctor` reports HTTP 404 for API reachability (base URL `/api/v3` has no root endpoint)
- Codex needed `exec -` syntax, not `--approval-mode` (CLI changed since docs were written)

What the pipeline doesn't test yet:
- Real ce:plan expansion of thin seeds (we manually executed phases, didn't test seed -> expand -> execute)
- CronCreate chaining between phases (tested manually, not with actual cron)
- Error recovery (what happens when a phase fails mid-execution)
- Complex APIs (Gmail, Stripe, Discord) - only tested petstore

## Current State

| Metric | Value | Target for Launch |
|--------|-------|-------------------|
| Dogfood pass rate | 4/10 APIs (40%) | 7/10 (70%) |
| Quality gate pass rate | Petstore 7/7 | All catalog APIs 7/7 |
| Test coverage | 19% (890 lines) | 30%+ with integration tests |
| Catalog size | 15 APIs | 20+ APIs |
| Pipeline E2E | Petstore only | 3+ APIs tested |
| Documentation | README good, no ARCHITECTURE.md | Complete |
| SKILL.md version | 0.4.0 | 1.0.0 |

## Implementation Phases

### Phase 1: Fix the Failures (dogfood pass rate 40% -> 70%+)

**Goal:** Re-run the dogfood gauntlet after recent template sanitization fixes (commits d8a93e0, 3f096bc) and fix remaining failures.

**Files:**
- `internal/generator/generator.go` - template FuncMap fixes
- `internal/generator/templates/*.tmpl` - template hardening
- `internal/openapi/parser.go` - parser edge cases

**Approach:**
- Re-run dogfood gauntlet against all 10 test APIs: petstore, spotify, fly, vercel, trello, supabase, jira, cloudflare, sentry, telegram
- For each failure: diagnose, fix, verify all 7 gates pass
- Target: 7/10 pass = launch-worthy, 10/10 pass = GOAT

**Verification:** `docs/plans/dogfood-gauntlet-findings.md` updated with new results

### Phase 2: Add --force to generate command

**Goal:** `printing-press generate --force` overwrites existing output directory without manual rm -rf.

**Files:**
- `internal/cli/root.go` - add --force flag to generate command
- `internal/generator/generator.go` - check force flag, rm existing dir

**Verification:** `printing-press generate --spec petstore.json --output /tmp/test --force` works when /tmp/test exists

### Phase 3: Friendly Command Aliases

**Goal:** Add `list`, `get`, `create`, `update`, `delete` aliases alongside generated operationId verbs.

**Files:**
- `internal/generator/templates/command.go.tmpl` - add Aliases field to cobra commands
- `internal/openapi/parser.go` - detect CRUD patterns from HTTP method + path

**Approach:**
- GET /pets -> alias "list" on the resource
- GET /pets/{id} -> alias "get"
- POST /pets -> alias "create"
- PUT/PATCH /pets/{id} -> alias "update"
- DELETE /pets/{id} -> alias "delete"
- Only add alias if no name collision exists

**Verification:** `petstore-cli pet list` works as alias for `petstore-cli pet find-by-status`

### Phase 4: Integration Tests

**Goal:** Add integration tests that run full generation + quality gates on real specs.

**Files:**
- New: `internal/generator/integration_test.go`
- New: `internal/pipeline/pipeline_test.go`

**Approach:**
- Test 1: Generate petstore CLI from URL, verify 7 gates pass
- Test 2: Generate from internal YAML spec (stytch), verify 7 gates pass
- Test 3: Pipeline init creates correct state.json + thin seeds
- Test 4: State transitions work (seed -> expanded -> completed)
- Use `testing.Short()` to skip slow tests in CI

**Verification:** `go test -count=1 ./...` passes including new integration tests

### Phase 5: Doctor Health Check Fix

**Goal:** `doctor` should check a real endpoint, not just base URL.

**Files:**
- `internal/generator/templates/doctor.go.tmpl` - try spec-specific health paths

**Approach:**
- Try base URL first, then try common health paths (/health, /status, /ping)
- For petstore specifically: try GET /pet/findByStatus?status=available
- Report "API: reachable" only if we get a 2xx/3xx response
- Report "API: unreachable (HTTP NNN)" with the actual status code

**Verification:** `petstore-cli doctor` shows "API: reachable" with 200

### Phase 6: Pipeline E2E on Gmail or Stripe

**Goal:** Run the full pipeline on a complex API to validate it works beyond petstore.

**Approach:**
- Pick Gmail (auth-heavy, Google Discovery format) or Stripe (large, well-documented)
- Run `printing-press print gmail` through all 6 phases
- Document any failures as targeted fix plans
- This validates the pipeline handles real-world complexity

**Verification:** Gmail/Stripe CLI generated with 7/7 gates, quality score 60+

### Phase 7: Version 1.0.0 + Launch Docs

**Goal:** Bump to 1.0.0, write ARCHITECTURE.md, finalize README.

**Files:**
- `skills/printing-press/SKILL.md` - bump version to 1.0.0
- New: `ARCHITECTURE.md`
- Update: `README.md` - add dogfood results, pipeline demo

**Approach:**
- ARCHITECTURE.md: package map, data flow diagram, template system explanation
- README update: add "Tested on N APIs" badge, pipeline screenshot/demo
- SKILL.md 1.0.0: reflect all fixes and new features from phases 1-6
- Push to GitHub, make repo public if not already

**Verification:** README renders correctly on GitHub, ARCHITECTURE.md is useful for new contributors

## Scope Boundaries

- Don't add new output languages (Python CLI, Bash) - Go only for 1.0
- Don't add GraphQL support
- Don't build a web UI
- Don't add CI/CD (GitHub Actions) yet - manual testing is fine for 1.0
- Don't pursue 100% test coverage - 30% with integration tests is the target

## Dependencies

```
Phase 1 (fix failures) - independent, highest priority
Phase 2 (--force flag) - independent, quick win
Phase 3 (aliases) - independent, high impact
Phase 4 (integration tests) - depends on Phase 1 (need passing APIs to test)
Phase 5 (doctor fix) - independent, nice-to-have
Phase 6 (complex API E2E) - depends on Phase 1
Phase 7 (1.0.0 launch) - depends on all others
```

Phases 1, 2, 3, 5 can run in parallel.
Phase 4 and 6 after Phase 1.
Phase 7 is the capstone.

## What "GOAT" Looks Like

```bash
printing-press print gmail

# 10 minutes later...
Pipeline Report: gmail
  Phases: 6/6 completed
  Quality Score: 92/100 (Grade A)
  Resources: 12 (messages, threads, labels, drafts, ...)
  Endpoints: 47

cd gmail-cli
./gmail-cli auth login --client-id $GOOGLE_CLIENT_ID
./gmail-cli messages list me --limit 5       # 'me' default from enrichment
./gmail-cli messages get me abc123 --json    # JSON output mode
./gmail-cli labels list me                   # All label CRUD
./gmail-cli doctor                           # All green, API reachable
```

Two words in, finished CLI out. Every command works. Every flag documented. Auth configured. Doctor passes. That's GOAT.

## Sources

- E2E test results: this session (2026-03-24)
- Dogfood gauntlet: `docs/plans/dogfood-gauntlet-findings.md`
- Template fixes: commits d8a93e0, 3f096bc
- Seed refactor: commits fa2459d, 1f36046, 37008dd, 7ca3e75
- Pipeline architecture: `docs/plans/2026-03-24-feat-autonomous-cli-pipeline-plan.md`
- SKILL.md: `skills/printing-press/SKILL.md` (v0.4.0)
