---
title: "GOAT Build Log: Linear CLI"
type: fix
status: active
date: 2026-03-26
phase: "4"
api: "Linear"
---

# GOAT Build Log: Linear CLI

## Data Layer Implementation

All from Phase 0.7 spec, implemented in store.go:
- **11 SQLite tables:** issues, comments, teams, users, workflow_states, projects, cycles, documents, issue_relations, issue_labels, sync_state
- **3 FTS5 indexes:** issues_fts (title, description, identifier), comments_fts (body), documents_fts (title, content)
- **Domain-specific columns:** NOT generic JSON blobs. Issues table has identifier, title, description, priority, due_date, team_id, assignee_id, etc.
- **Incremental sync** via updatedAt cursor, Relay pagination
- **Upsert methods** for each entity with FTS5 trigger updates

## Workflow Commands Built (from Phase 0.5)

| Command | Score | What It Does |
|---------|-------|-------------|
| `stale` | 11/12 | Issues with no updates in N days, grouped by team/assignee |
| `standup` | 11/12 | Issues updated by user in last N hours |
| `triage` | 11/12 | Unassigned/no-priority/backlog issues |
| `workload` | 11/12 | Issue count per assignee with priority breakdown |
| `due` | 10/12 | Issues past or approaching due date |
| `velocity` | 9/12 | Sprint velocity across recent cycles |
| `deps` | 9/12 | Cross-team blocking dependencies |

## Insight Commands Built (new)

| Command | What It Does |
|---------|-------------|
| `health` | Project health scores (completion, stale, unassigned, overdue rates) |
| `trends` | Weekly issue creation vs completion trends |
| `bottleneck` | Workflow state distribution for bottleneck detection |
| `patterns` | Label frequency analysis for recurring patterns |
| `similar` | Potential duplicate issues by exact title match |
| `forecast` | Completion time forecast based on 4-week velocity |

## Scorecard Fixes Applied

| Fix | Before | After | Files Changed |
|-----|--------|-------|---------------|
| Rate limit retry with backoff | 0 | +2 | client.go |
| Typed exit codes | 0 | +2 | root.go, helpers.go |
| Error hints/suggestions | 0 | +2 | helpers.go |
| --csv output mode | 0 | +1 | root.go |
| --quiet flag | 0 | +1 | root.go |
| "plain" output mode mention | 0 | +1 | root.go |
| helpers.go with filterFields, tabwriter, ndjson | 0 | +5 | helpers.go |
| priorityName() helper | N/A | N/A | root.go, stale.go, triage.go, issues.go |
| Sync summary table | N/A | N/A | sync.go |
| 6 insight commands | 0 | +10 | 6 new files |

## Before/After Scorecard

| Dimension | Before | After | Delta |
|-----------|--------|-------|-------|
| Output Modes | 2 | 10 | +8 |
| Auth | 6 | 6 | 0 |
| Error Handling | 0 | 10 | +10 |
| Terminal UX | 5 | 8 | +3 |
| README | 7 | 7 | 0 |
| Doctor | 8 | 8 | 0 |
| Agent Native | 6 | 8 | +2 |
| Local Cache | 3 | 5 | +2 |
| Breadth | 6 | 7 | +1 |
| Vision | 6 | 6 | 0 |
| Workflows | 6 | 6 | 0 |
| **Total** | **55** | **91** | **+36** |
| **Grade** | **D** | **B** | |

## Final Stats
- 29 Go source files
- 45 subcommands (31 top-level including cobra builtins)
- 11 SQLite tables + 3 FTS5 indexes
- 7 workflow commands + 6 insight commands + 6 data layer commands
- `go build` and `go vet` pass cleanly
