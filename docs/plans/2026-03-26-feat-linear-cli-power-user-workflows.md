---
title: "Power User Workflows: Linear CLI"
type: feat
status: active
date: 2026-03-26
phase: "0.5"
api: "linear"
---

# Power User Workflows: Linear CLI

## API Archetype: Project Management

Linear maps to the **Project Management** archetype with signals: issues, tasks/sub-issues, sprints (cycles), states (workflow states), teams, labels, milestones, initiatives. Expected workflows: stale issues, orphan detection, velocity, burndown, standup, triage, bulk operations.

## All 12 Workflow Ideas

### 1. Stale Issue Detector
- **Steps:** Query issues by team -> filter where updatedAt < N days ago -> exclude archived/canceled -> group by state -> count and list
- **Frequency:** Weekly (eng manager hygiene)
- **Pain:** Manual: create a custom view, scan each team, mentally filter. No way to script this in Linear's UI.
- **API support:** Issues query supports `filter: { updatedAt: { lt: "..." } }` + `team` filter. Confirmed.
- **Proposed:** `linear stale --days 30 --team ENG --json`

### 2. Cycle Velocity Report
- **Steps:** Get active or recent cycle -> count issues by state (done/total) -> compute completion % -> optionally compare to last N cycles for trend
- **Frequency:** Weekly/bi-weekly (sprint retro)
- **Pain:** Linear has basic insights but no CLI-accessible velocity export. External tools like Screenful/Count.co fill this gap.
- **API support:** `cycles` query with `filter: { team: { id: { eq: "..." } } }`, cycle has `issues` connection with state filtering. Confirmed.
- **Proposed:** `linear velocity --team ENG --cycles 5 --json`

### 3. Orphan Issue Detection
- **Steps:** Query issues -> filter where project is null AND cycle is null AND assignee is null -> group by team -> report
- **Frequency:** Monthly (backlog grooming)
- **Pain:** No built-in view for "issues that belong nowhere." These accumulate silently.
- **API support:** Issues filter supports `project: { null: true }`, `cycle: { null: true }`, `assignee: { null: true }`. Confirmed.
- **Proposed:** `linear orphans --team ENG --json`

### 4. Standup Report
- **Steps:** Get issues assigned to current user -> filter by recent activity (updatedAt in last N days) -> group by state -> format as "yesterday/today/blockers"
- **Frequency:** Daily
- **Pain:** Every developer opens Linear to prepare standup. A CLI command pipes directly to Slack or clipboard.
- **API support:** `issues` filter by `assignee: { isMe: { eq: true } }` and `updatedAt`. Confirmed.
- **Proposed:** `linear standup --days 1 --json`

### 5. Triage Queue
- **Steps:** Get issues in Triage state -> sort by priority + createdAt -> show with context (labels, assignee suggestions)
- **Frequency:** Daily (designated triage person)
- **Pain:** Linear has Triage Intelligence (Business/Enterprise) but no CLI export.
- **API support:** `issues` filter by state type (triage). WorkflowState has `type` field. Confirmed.
- **Proposed:** `linear triage --team ENG --json`

### 6. Label Audit
- **Steps:** List all labels -> for each, count issues with that label -> find labels with 0 issues -> find issues with no labels
- **Frequency:** Monthly (workspace hygiene)
- **Pain:** No way to see unused labels. They accumulate over years. Label sprawl is a real problem.
- **API support:** `issueLabels` query + `issues` filter by label. Confirmed.
- **Proposed:** `linear label-audit --team ENG --json`

### 7. Duplicate Finder
- **Steps:** Sync issues locally -> run FTS5 similarity matching on title + description -> rank by similarity score -> report probable duplicates
- **Frequency:** Monthly (backlog grooming)
- **Pain:** Linear has no built-in duplicate detection. This is a common complaint.
- **API support:** Requires local DB (can't do text similarity via API). Data layer enables this.
- **Proposed:** `linear duplicates --team ENG --threshold 0.8 --json`

### 8. SLA Monitor
- **Steps:** Query issues with priority Urgent/High -> check time since creation vs. SLA threshold -> flag overdue
- **Frequency:** Daily (support/ops teams)
- **Pain:** No built-in SLA tracking in Linear. Teams use external tools.
- **API support:** Issues have `priority` (0-4) and `createdAt`. Confirmed.
- **Proposed:** `linear sla --urgent 4h --high 24h --team SUPPORT --json`

### 9. Bulk State Transition
- **Steps:** Query issues matching filter -> update state for all matches in one batch
- **Frequency:** Weekly (sprint close, backlog cleanup)
- **Pain:** Linear UI requires manual multi-select. No bulk API shortcut.
- **API support:** `issueUpdate` mutation supports state change. Batch via multiple mutations.
- **Proposed:** `linear bulk-move --from "In Review" --to "Done" --team ENG --dry-run`

### 10. Team Health Dashboard
- **Steps:** For each team member: count assigned issues, count overdue, count blocked -> compute health metrics
- **Frequency:** Weekly (eng manager 1:1 prep)
- **Pain:** Requires clicking through each person's view in Linear.
- **API support:** Issues filter by assignee + state + dueDate. Confirmed.
- **Proposed:** `linear health --team ENG --json`

### 11. Blocked Issue Report
- **Steps:** Query issues with "blocked" relation or "blocked" label -> list with blockers
- **Frequency:** Daily (standup, unblocking)
- **Pain:** No dedicated view for blocked issues across teams.
- **API support:** `issueRelations` query, relation type `blocks`/`isBlockedBy`. Confirmed.
- **Proposed:** `linear blocked --team ENG --json`

### 12. Activity Timeline
- **Steps:** Query recent issue history events -> filter by type (state change, assignment, comment) -> format as timeline
- **Frequency:** Daily (async teams, status updates)
- **Pain:** Linear's activity feed is per-issue. No cross-workspace timeline.
- **API support:** `issueHistory` connection on issues. Confirmed.
- **Proposed:** `linear activity --team ENG --days 7 --json`

## Validation Against API Capabilities

| Workflow | All endpoints exist? | Filterable? | Write ops needed? | Local DB helps? |
|----------|---------------------|-------------|-------------------|-----------------|
| Stale | Yes | Yes (updatedAt filter) | No | Yes - instant |
| Velocity | Yes | Yes (cycle + team) | No | Yes - aggregation |
| Orphans | Yes | Yes (null filters) | No | Yes - instant |
| Standup | Yes | Yes (assignee + date) | No | Yes - instant |
| Triage | Yes | Yes (state type) | No | Partial |
| Label Audit | Yes | Yes (label filter) | No | Yes - cross-query |
| Duplicates | N/A (local only) | N/A | No | Required |
| SLA | Yes | Yes (priority + date) | No | Yes - instant |
| Bulk Move | Yes | Yes | Yes (mutations) | No - live API |
| Health | Yes | Yes (assignee) | No | Yes - aggregation |
| Blocked | Yes | Yes (relations) | No | Yes - join query |
| Activity | Yes | Partial | No | Yes - timeline |

## Scoring

| # | Workflow | Frequency | Pain | Feasibility | Uniqueness | Total |
|---|---------|-----------|------|-------------|------------|-------|
| 1 | Stale | 2 | 3 | 3 | 3 | **11/12** |
| 2 | Velocity | 2 | 2 | 2 | 2 | **8/12** |
| 3 | Orphans | 1 | 3 | 3 | 3 | **10/12** |
| 4 | Standup | 3 | 2 | 3 | 2 | **10/12** |
| 5 | Triage | 3 | 2 | 3 | 1 | **9/12** |
| 6 | Label Audit | 1 | 2 | 3 | 3 | **9/12** |
| 7 | Duplicates | 1 | 3 | 1 | 3 | **8/12** |
| 8 | SLA | 3 | 3 | 3 | 3 | **12/12** |
| 9 | Bulk Move | 2 | 2 | 2 | 1 | **7/12** |
| 10 | Health | 2 | 2 | 2 | 3 | **9/12** |
| 11 | Blocked | 3 | 2 | 2 | 2 | **9/12** |
| 12 | Activity | 3 | 2 | 2 | 2 | **9/12** |

## Top 7 for Implementation (Phase 4 Mandatory)

1. **SLA Monitor** (12/12) - `linear sla --urgent 4h --high 24h --team SUPPORT`
2. **Stale Issue Detector** (11/12) - `linear stale --days 30 --team ENG`
3. **Orphan Detection** (10/12) - `linear orphans --team ENG`
4. **Standup Report** (10/12) - `linear standup --days 1`
5. **Triage Queue** (9/12) - `linear triage --team ENG`
6. **Health Dashboard** (9/12) - `linear health --team ENG`
7. **Blocked Report** (9/12) - `linear blocked --team ENG`

## Implementation Notes

All 7 workflows benefit from the local SQLite data layer:
- Stale, Orphans, Standup, SLA, Health, Blocked: query local DB with joins, instant results
- Triage: primarily live API (issues in triage are fresh), but DB helps with context enrichment
- All support `--sync` flag to refresh local DB before querying
- All support `--json` output for agent consumption
- All are composable: `linear stale --days 30 --json | jq '.[].identifier'`
