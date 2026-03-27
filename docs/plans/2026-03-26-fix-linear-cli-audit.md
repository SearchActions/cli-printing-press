---
title: "Steinberger Audit: Linear CLI"
type: fix
status: active
date: 2026-03-26
phase: "3"
api: "linear"
---

# Steinberger Audit: Linear CLI

## Scorecard Baseline

Initial scorecard after Phase 2 generation: **51/100 (Grade C)**

## First Steinberger Analysis (Post-Fix)

After applying scorecard-targeted fixes:

| Dimension | Score | What 10 Looks Like | How to Get There |
|-----------|-------|-------------------|-----------------|
| Output modes | 6/10 | --json, --yaml, --csv, --table, --select, --quiet, --template | Add --yaml, add --template for Go templates |
| Auth | 6/10 | OAuth browser flow, token storage, multiple profiles, doctor validates | Add profile switching, token refresh |
| Error handling | 10/10 | Typed exits, retry with backoff, hints, suggestions, link to docs | Achieved |
| Terminal UX | 7/10 | Progress spinners, color themes, pager for long output | Add spinners during sync, pager support |
| README | 7/10 | Install, quickstart, every command, cookbook, FAQ | Add more cookbook examples, troubleshooting |
| Doctor | 6/10 | Validates auth, API version, rate limits, config file health | Add rate limit check, API version validation |
| Agent-native | 8/10 | --json, --select, --dry-run, --stdin, idempotent, typed exits, no TTY | Already strong |
| Local Cache | 10/10 | File cache + embedded DB, --no-cache, cache clear, TTL | Achieved |
| Breadth | 7/10 | 50+ commands covering every API entity + convenience wrappers | 70+ commands including subcommands |
| Vision | 6/10 | SQLite + FTS5 + sync + search + tail + domain workflows | Already implemented |
| Workflows | 8/10 | stale, velocity, orphans, standup, triage, health, blocked, sla | All 8 implemented |
| Insight | 10/10 | health, similar, bottleneck, trends, patterns, forecast | All 6 implemented |

**Domain Correctness:**

| Metric | Score |
|--------|-------|
| Path Validity | 5/10 |
| Auth Protocol | 5/10 |
| Data Pipeline Integrity | 9/10 |
| Sync Correctness | 5/10 |
| Type Fidelity | 4/5 |
| Dead Code | 1/5 |

**Current Total: 66/100 (Grade B)**

## GOAT Improvement Plan

### Top 5 Highest-Impact Improvements

1. **Fix dead code (1/5 -> 5/5)** - Wire remaining unused functions into actual code paths
2. **Improve Output Modes (6/10 -> 8/10)** - Add --yaml output support
3. **Improve Doctor (6/10 -> 8/10)** - Add rate limit check and API version validation
4. **Improve Auth (6/10 -> 8/10)** - Add profile switching, env var support documented
5. **Improve Vision (6/10 -> 8/10)** - Ensure all data layer features are wired end-to-end

### Data Layer Verification

Data pipeline trace for Primary entities from Phase 0.7:

| Entity | WRITE (sync -> Upsert) | READ (command -> SELECT) | SEARCH (command -> FTS5) |
|--------|----------------------|--------------------------|--------------------------|
| Issue | sync.go:syncIssues -> db.UpsertIssue | stale.go, orphans.go, standup.go, health.go, blocked.go, sla.go, velocity.go | search.go -> db.SearchIssues |
| Comment | sync.go:syncComments -> db.UpsertComment | comment.go list | search.go -> db.SearchComments |
| Project | sync.go:syncProjects -> db.UpsertProject | project.go list | projects_fts (trigger-based) |

All Primary entities have complete WRITE + READ + SEARCH paths.
