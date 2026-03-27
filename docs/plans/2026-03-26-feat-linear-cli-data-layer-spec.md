---
title: "Data Layer Specification: Linear CLI"
type: feat
status: active
date: 2026-03-26
phase: "0.7"
api: "linear"
---

# Data Layer Specification: Linear CLI

## Overview

This document specifies the SQLite data layer for linear-cli based on Linear's GraphQL schema analysis, social signal mining, and data gravity scoring. The data layer enables instant local queries, full-text search, and cross-entity workflow commands that would be impossible or rate-limit-expensive via the live API.

## Entity Classification

| Entity | Type | Estimated Volume | Update Frequency | Temporal Field |
|--------|------|-----------------|------------------|---------------|
| **Issue** | Accumulating | 1k-50k per workspace | High (daily state changes) | `updatedAt` |
| **Comment** | Accumulating | 2x-5x issue count | Medium | `updatedAt` |
| **Project** | Reference | 10-200 | Low (weekly) | `updatedAt` |
| **Cycle** | Reference | 10-100 per team | Low (bi-weekly) | `updatedAt` |
| **Team** | Reference | 5-50 | Rare | `updatedAt` |
| **User** | Reference | 10-500 | Rare | `updatedAt` |
| **IssueLabel** | Reference | 20-200 | Rare | `updatedAt` |
| **WorkflowState** | Reference | 5-15 per team | Rare | `updatedAt` |
| **Document** | Accumulating | 50-500 | Medium | `updatedAt` |
| **Initiative** | Reference | 5-50 | Low | `updatedAt` |
| **Milestone** | Reference | 5-30 | Low | `updatedAt` |
| **Attachment** | Accumulating | 1x-2x issue count | Low | `updatedAt` |
| **IssueRelation** | Accumulating | 0.1x-0.5x issue count | Low | `updatedAt` |
| **CustomView** | Reference | 10-100 | Low | `updatedAt` |
| **IssueHistory** | Append-only | 3x-10x issue count | N/A (immutable) | `createdAt` |
| **Notification** | Append-only | High | N/A | `createdAt` |

## Social Signal Mining Results

| # | Signal | Evidence | Score |
|---|--------|----------|-------|
| 1 | Users want offline/local issue search (no existing tool provides this) | 4 CLIs exist, none with local DB. Phase 0 confirmed "discrawl gap." | 8/10 |
| 2 | Sprint velocity/burndown analytics from local data | Screenful, Count.co sell this as SaaS. CLI alternative has demand. | 7/10 |
| 3 | Stale issue detection as a common grooming task | Linear added auto-close. Morgen guide recommends monthly stale review. | 7/10 |
| 4 | Cross-entity queries (issues by label + team + state + date) | All 4 competing CLIs support filters but hit API each time. Local enables instant. | 8/10 |
| 5 | Issue export/backup to local storage | Linear has CSV export. GitHub has linear/linear/packages/import. Users want programmatic access. | 6/10 |
| 6 | Duplicate detection via text similarity | Linear Backlog Grooming Agent (Cotera) addresses this. No CLI tool. | 6/10 |
| 7 | Activity timeline across teams | No existing tool provides cross-team activity feed via CLI. | 6/10 |

## Data Gravity Scoring

| Entity | Volume | QueryFreq | JoinDemand | SearchNeed | TemporalValue | **Total** | Classification |
|--------|--------|-----------|------------|------------|---------------|-----------|---------------|
| **Issue** | 3 | 3 | 3 | 3 | 3 | **15/12** | Primary |
| **Comment** | 2 | 2 | 2 | 3 | 1 | **10/12** | Primary |
| **Project** | 1 | 2 | 3 | 1 | 1 | **8/12** | Primary |
| **Cycle** | 1 | 2 | 2 | 0 | 2 | **7/12** | Support |
| **Team** | 0 | 2 | 3 | 0 | 0 | **5/12** | Support |
| **User** | 1 | 2 | 3 | 1 | 0 | **7/12** | Support |
| **IssueLabel** | 1 | 2 | 2 | 1 | 0 | **6/12** | Support |
| **WorkflowState** | 0 | 2 | 3 | 0 | 0 | **5/12** | Support |
| **IssueRelation** | 1 | 2 | 2 | 0 | 1 | **6/12** | Support |
| **Document** | 1 | 1 | 1 | 2 | 1 | **6/12** | Support |
| **Initiative** | 0 | 1 | 1 | 1 | 1 | **4/12** | API-only |
| **Milestone** | 0 | 1 | 1 | 0 | 1 | **3/12** | API-only |
| **Attachment** | 1 | 1 | 1 | 0 | 0 | **3/12** | API-only |
| **CustomView** | 0 | 1 | 0 | 0 | 0 | **1/12** | API-only |

**Primary entities (score >= 8):** Issue (15), Comment (10), Project (8)
**Support entities (score 5-7):** Cycle, Team, User, IssueLabel, WorkflowState, IssueRelation, Document

## SQLite Schema

```sql
-- Primary Tables

CREATE TABLE issues (
    id TEXT PRIMARY KEY,
    identifier TEXT NOT NULL,          -- e.g. "ENG-123"
    title TEXT NOT NULL,
    description TEXT,
    priority INTEGER,                   -- 0=none, 1=urgent, 2=high, 3=medium, 4=low
    estimate REAL,
    due_date TEXT,                      -- ISO 8601
    sort_order REAL,
    state_id TEXT REFERENCES workflow_states(id),
    team_id TEXT REFERENCES teams(id),
    assignee_id TEXT REFERENCES users(id),
    creator_id TEXT REFERENCES users(id),
    project_id TEXT REFERENCES projects(id),
    cycle_id TEXT REFERENCES cycles(id),
    parent_id TEXT REFERENCES issues(id),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    archived_at TEXT,
    canceled_at TEXT,
    completed_at TEXT,
    started_at TEXT,
    data JSON NOT NULL,                 -- full API response
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_issues_team ON issues(team_id);
CREATE INDEX idx_issues_assignee ON issues(assignee_id);
CREATE INDEX idx_issues_project ON issues(project_id);
CREATE INDEX idx_issues_cycle ON issues(cycle_id);
CREATE INDEX idx_issues_state ON issues(state_id);
CREATE INDEX idx_issues_priority ON issues(priority);
CREATE INDEX idx_issues_updated ON issues(updated_at);
CREATE INDEX idx_issues_identifier ON issues(identifier);

CREATE VIRTUAL TABLE issues_fts USING fts5(
    title,
    description,
    identifier,
    content='issues',
    content_rowid='rowid'
);

CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    body TEXT NOT NULL,
    issue_id TEXT REFERENCES issues(id),
    user_id TEXT REFERENCES users(id),
    parent_id TEXT REFERENCES comments(id),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_comments_issue ON comments(issue_id);
CREATE INDEX idx_comments_user ON comments(user_id);
CREATE INDEX idx_comments_updated ON comments(updated_at);

CREATE VIRTUAL TABLE comments_fts USING fts5(
    body,
    content='comments',
    content_rowid='rowid'
);

CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    state TEXT,                         -- planned, started, paused, completed, canceled
    icon TEXT,
    color TEXT,
    lead_id TEXT REFERENCES users(id),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_projects_lead ON projects(lead_id);
CREATE INDEX idx_projects_state ON projects(state);

CREATE VIRTUAL TABLE projects_fts USING fts5(
    name,
    description,
    content='projects',
    content_rowid='rowid'
);

-- Support Tables

CREATE TABLE teams (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    key TEXT NOT NULL,                  -- e.g. "ENG"
    description TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    email TEXT,
    avatar_url TEXT,
    is_me INTEGER DEFAULT 0,           -- flag for the authenticated user
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE workflow_states (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,                 -- triage, backlog, unstarted, started, completed, canceled
    team_id TEXT REFERENCES teams(id),
    position REAL,
    color TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_workflow_states_team ON workflow_states(team_id);
CREATE INDEX idx_workflow_states_type ON workflow_states(type);

CREATE TABLE issue_labels (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT,
    team_id TEXT REFERENCES teams(id),
    parent_id TEXT REFERENCES issue_labels(id),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE issue_label_assignments (
    issue_id TEXT REFERENCES issues(id),
    label_id TEXT REFERENCES issue_labels(id),
    PRIMARY KEY (issue_id, label_id)
);

CREATE TABLE cycles (
    id TEXT PRIMARY KEY,
    name TEXT,
    number INTEGER,
    team_id TEXT REFERENCES teams(id),
    starts_at TEXT,
    ends_at TEXT,
    completed_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_cycles_team ON cycles(team_id);

CREATE TABLE issue_relations (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,                 -- blocks, isBlockedBy, duplicate, related
    issue_id TEXT REFERENCES issues(id),
    related_issue_id TEXT REFERENCES issues(id),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_issue_relations_issue ON issue_relations(issue_id);
CREATE INDEX idx_issue_relations_related ON issue_relations(related_issue_id);
CREATE INDEX idx_issue_relations_type ON issue_relations(type);

CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    project_id TEXT REFERENCES projects(id),
    creator_id TEXT REFERENCES users(id),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Sync Metadata

CREATE TABLE sync_cursors (
    entity_type TEXT PRIMARY KEY,
    last_cursor TEXT,
    last_updated_at TEXT,
    last_sync_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

## Sync Strategy

**Cursor field:** `updatedAt` (DateTime, supported as filter and orderBy on all paginated queries)

**Validated:** Linear's GraphQL API supports `filter: { updatedAt: { gte: "2026-03-01T00:00:00Z" } }` on issues, comments, projects, and all entity types. Ordering by `updatedAt` confirmed in API docs.

**Pagination:** Relay-style cursor with `first` (max 250 per the API) and `after` (endCursor from pageInfo).

**Sync algorithm per entity type:**

```
1. Read sync_cursors for entity_type
2. If last_updated_at exists:
     Query with filter: { updatedAt: { gte: last_updated_at } }, orderBy: updatedAt
   Else:
     Full sync: Query all, orderBy: updatedAt
3. For each page (first: 100, after: endCursor):
     Upsert each record into SQLite (INSERT OR REPLACE)
     Update FTS5 triggers
4. Update sync_cursors with max(updatedAt) from results
```

**Batch size:** 100 records per page (conservative to stay within complexity limits). Max 250 allowed but 100 is safer with nested fields.

**Deletion handling:** Linear uses `archivedAt` soft deletes. Sync checks `archivedAt` field and marks records locally. Hard deletes (rare) detected by periodic full reconciliation.

## Domain-Specific Search Filters

| CLI Flag | SQL WHERE Clause | Entity |
|----------|-----------------|--------|
| `--team <key>` | `WHERE team_id = (SELECT id FROM teams WHERE key = ?)` | issues, workflow_states, cycles |
| `--assignee <name>` | `WHERE assignee_id = (SELECT id FROM users WHERE display_name LIKE ?)` | issues |
| `--project <name>` | `WHERE project_id = (SELECT id FROM projects WHERE name LIKE ?)` | issues |
| `--state <name>` | `WHERE state_id = (SELECT id FROM workflow_states WHERE name = ?)` | issues |
| `--priority <level>` | `WHERE priority = ?` (0-4) | issues |
| `--label <name>` | `WHERE id IN (SELECT issue_id FROM issue_label_assignments WHERE label_id = (SELECT id FROM issue_labels WHERE name = ?))` | issues |
| `--days <n>` | `WHERE updated_at >= datetime('now', '-N days')` | issues, comments |
| `--since <date>` | `WHERE updated_at >= ?` | issues, comments |
| `--cycle <name>` | `WHERE cycle_id = (SELECT id FROM cycles WHERE name = ? OR number = ?)` | issues |
| `--unassigned` | `WHERE assignee_id IS NULL` | issues |
| `--overdue` | `WHERE due_date < date('now') AND completed_at IS NULL AND canceled_at IS NULL` | issues |

## Compound Cross-Entity Queries

### 1. Stale issues by team with state context
```sql
SELECT i.identifier, i.title, i.priority, ws.name as state, u.display_name as assignee,
       julianday('now') - julianday(i.updated_at) as days_stale
FROM issues i
JOIN workflow_states ws ON i.state_id = ws.id
LEFT JOIN users u ON i.assignee_id = u.id
JOIN teams t ON i.team_id = t.id
WHERE t.key = ?
  AND i.updated_at < datetime('now', '-30 days')
  AND ws.type NOT IN ('completed', 'canceled')
  AND i.archived_at IS NULL
ORDER BY days_stale DESC;
```

### 2. Cycle velocity (completion rate per cycle)
```sql
SELECT c.name as cycle, c.number,
       COUNT(*) as total_issues,
       SUM(CASE WHEN ws.type = 'completed' THEN 1 ELSE 0 END) as completed,
       ROUND(100.0 * SUM(CASE WHEN ws.type = 'completed' THEN 1 ELSE 0 END) / COUNT(*), 1) as pct
FROM issues i
JOIN cycles c ON i.cycle_id = c.id
JOIN workflow_states ws ON i.state_id = ws.id
JOIN teams t ON i.team_id = t.id
WHERE t.key = ?
GROUP BY c.id
ORDER BY c.starts_at DESC
LIMIT 5;
```

### 3. Orphan issues (no project, no cycle, unassigned)
```sql
SELECT i.identifier, i.title, i.priority, ws.name as state, t.key as team,
       julianday('now') - julianday(i.created_at) as days_old
FROM issues i
JOIN workflow_states ws ON i.state_id = ws.id
JOIN teams t ON i.team_id = t.id
WHERE i.project_id IS NULL
  AND i.cycle_id IS NULL
  AND i.assignee_id IS NULL
  AND ws.type NOT IN ('completed', 'canceled')
  AND i.archived_at IS NULL
ORDER BY i.priority ASC, days_old DESC;
```

### 4. Blocked issues with blocker context
```sql
SELECT i.identifier, i.title,
       bi.identifier as blocked_by_identifier, bi.title as blocked_by_title,
       u.display_name as blocked_by_assignee
FROM issue_relations ir
JOIN issues i ON ir.issue_id = i.id
JOIN issues bi ON ir.related_issue_id = bi.id
LEFT JOIN users u ON bi.assignee_id = u.id
JOIN workflow_states ws ON i.state_id = ws.id
WHERE ir.type = 'isBlockedBy'
  AND ws.type NOT IN ('completed', 'canceled')
ORDER BY i.priority ASC;
```

### 5. Full-text search with domain filters
```sql
SELECT i.identifier, i.title, i.priority, ws.name as state,
       u.display_name as assignee, t.key as team
FROM issues_fts fts
JOIN issues i ON i.rowid = fts.rowid
JOIN workflow_states ws ON i.state_id = ws.id
LEFT JOIN users u ON i.assignee_id = u.id
JOIN teams t ON i.team_id = t.id
WHERE issues_fts MATCH ?
  AND t.key = ?
  AND ws.type NOT IN ('completed', 'canceled')
ORDER BY rank;
```

## Tail Strategy

| Method | Available? | Decision |
|--------|-----------|----------|
| WebSocket/Gateway | No | N/A |
| SSE | No | N/A |
| **REST Polling** | **Yes** | Primary method |

**Decision:** REST polling with `updatedAt` cursor. Linear has no WebSocket or SSE endpoint. Webhooks require a server (not suitable for CLI). The `tail` command will poll with `filter: { updatedAt: { gte: lastSeen } }` on a configurable interval (default: 30s).

## Phase 4 Priority 0 Commands (Data Layer)

1. `sync` - Incremental sync of all entities to SQLite
2. `search` - FTS5 search with domain filters
3. `sql` - Raw read-only SQL queries
4. `issues` - Query local DB with rich filters (replaces API-dependent `issue list`)
5. `tail` - Poll for new/updated issues and stream to stdout
