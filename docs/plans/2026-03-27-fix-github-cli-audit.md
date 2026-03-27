---
title: "Non-Obvious Insight Review: GitHub CLI"
type: fix
status: active
date: 2026-03-27
phase: "3"
api: "github"
---

# Non-Obvious Insight Review: GitHub CLI

## Automated Scorecard Baseline

| Dimension | Score | Notes |
|-----------|-------|-------|
| Output Modes | 10/10 | --json, --csv, --plain, --quiet, --select, --compact |
| Auth | 8/10 | GITHUB_TOKEN env var, config file, doctor validates |
| Error Handling | 10/10 | Typed exits, retry patterns |
| Terminal UX | 9/10 | tabwriter, color support, --no-color |
| README | 7/10 | Missing cookbook, FAQ, workflow examples |
| Doctor | 10/10 | Auth validation, API connectivity check |
| Agent Native | 8/10 | --json, --select, --dry-run, --yes, --no-input |
| Local Cache | 10/10 | SQLite store, --no-cache bypass |
| Breadth | 10/10 | 117 commands from 51 paths |
| Vision | 9/10 | sync, tail, search, analytics, export, import |
| Workflows | 4/10 | Generic workflow/analytics stubs, not domain-specific |
| Insight | 0/10 | No insight commands (health, trends, stale, etc.) |
| Path Validity | 5/10 | Needs domain validation |
| Auth Protocol | 5/10 | Needs domain validation |
| Data Pipeline | 10/10 | Store + sync + search connected |
| Sync Correctness | 8/10 | Generic sync, not domain-aware |
| Type Fidelity | 1/5 | Reserved word types fixed, but schema is generic |
| Dead Code | 0/5 | Likely dead flags and functions |

**Baseline Total: 68/100 (Grade B)**

## GOAT Improvement Plan

### Priority 0: Data Layer Foundation (from Phase 0.7) [+15-20 points expected]
1. **Replace store.go** - Current store uses generic `resources` table. Replace with domain-specific tables from Phase 0.7 (issues, pull_requests, commits, comments, repos, users, workflow_runs, events, reviews)
2. **Rewrite sync.go** - Use `since` cursor for issues/commits, `created` for workflow_runs. Add --repo and --org scoping flags.
3. **Add domain search** - FTS5 on issue titles, PR titles, commit messages, comment bodies
4. **Add sql command** - Raw read-only SQL against local DB
5. **Add domain list commands** - issues, prs, commits queries against local DB with filters

### Priority 1: Workflow Commands (from Phase 0.5) [+10-15 points expected]
Build the 7 workflow commands: pr-triage, stale, actions-health, changelog, security, activity, contributors

### Priority 2: Scorecard Fixes [+5-10 points expected]
- Fix README with cookbook section showcasing workflows
- Fix placeholder examples ("example-value" -> realistic values)
- Remove dead flags and functions
- Wire all flags to actual command logic

### Priority 3: Polish
- README FAQ section
- Domain-specific help text improvements

## Target: 85+/100 (Grade A)
