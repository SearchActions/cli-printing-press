---
title: "Data Layer Specification: GitHub CLI"
type: feat
status: active
date: 2026-03-27
phase: "0.7"
api: "github"
---

# Data Layer Specification: GitHub CLI

## Overview

GitHub's API covers 1,107 operations across 25+ resource categories, but only a subset has the data gravity to justify local persistence. This specification identifies 6 primary entities for SQLite tables, defines their schemas with proper domain columns (not JSON blobs), validates sync cursor strategies against actual API parameters, and maps domain-specific search filters to SQL WHERE clauses.

The data layer is what transforms a 1,107-endpoint API wrapper into a useful tool. github-to-sqlite (462 stars) proved the model - we build it natively in Go with FTS5 and compound queries.

## Entity Classification

| Entity | Type | Est. Volume | Update Frequency | Temporal Field | Persistence |
|--------|------|-------------|------------------|---------------|-------------|
| **Issues** | Accumulating | 10k-1M per org | Daily | updated_at | SQLite + FTS5 |
| **Pull Requests** | Accumulating | 1k-100k per org | Daily | updated_at | SQLite + FTS5 |
| **Commits** | Append-only | 10k-10M per org | Daily | commit.author.date | SQLite table |
| **Workflow Runs** | Append-only | 1k-100k per repo | Daily | created_at | SQLite table |
| **Repositories** | Reference | 10-10k per org | Weekly | updated_at | SQLite table |
| **Users** | Reference | 10-10k per org | Monthly | N/A | SQLite table |
| **Events** | Append-only | High volume | Hourly | created_at | SQLite table |
| **Releases** | Append-only | 10-1k per repo | Monthly | created_at | SQLite table |
| **Reviews** | Append-only | 1-100 per PR | Per PR | submitted_at | SQLite table |
| **Comments** | Accumulating | 10k-100k per org | Daily | updated_at | SQLite + FTS5 |
| **Code Scanning Alerts** | Accumulating | 0-10k per repo | Weekly | updated_at | SQLite table |
| **Dependabot Alerts** | Accumulating | 0-1k per repo | Weekly | updated_at | SQLite table |
| **Labels** | Reference | 10-100 per repo | Rarely | N/A | SQLite table |
| **Milestones** | Reference | 1-50 per repo | Weekly | updated_at | SQLite table |
| **Teams** | Reference | 1-100 per org | Monthly | N/A | SQLite table |

## Social Signal Mining Results

### Signal 1: Local issue/PR search is the #1 demand (Evidence: 8/10)
github-to-sqlite (462 stars) + Datasette ecosystem. Users sync issues/PRs to SQLite and query with SQL.

### Signal 2: Cross-repo aggregation for engineering managers (Evidence: 7/10)
gh-dash (11.2k stars) exists as TUI. Reddit/HN posts about needing cross-repo views.

### Signal 3: CI/CD analytics gap (Evidence: 6/10)
github-actions-watcher, BuildBeacon, burndown tools. No CLI aggregates workflow run data for trend analysis.

### Signal 4: Export/backup to structured formats (Evidence: 7/10)
gh2md, export-pull-requests, python-github-backup, ghexport. Users want local copies of GitHub data.

## Data Gravity Scoring

| Entity | Volume (0-3) | QueryFreq (0-3) | JoinDemand (0-2) | SearchNeed (0-2) | TemporalValue (0-2) | **Total** | **Status** |
|--------|-------------|-----------------|------------------|-----------------|--------------------|---------|---------|
| **Issues** | 3 | 3 | 2 | 2 | 2 | **12** | PRIMARY |
| **Pull Requests** | 2 | 3 | 2 | 2 | 2 | **11** | PRIMARY |
| **Commits** | 3 | 2 | 1 | 1 | 2 | **9** | PRIMARY |
| **Workflow Runs** | 2 | 2 | 1 | 0 | 2 | **7** | SUPPORT |
| **Comments** | 3 | 2 | 2 | 2 | 1 | **10** | PRIMARY |
| **Repositories** | 1 | 3 | 2 | 1 | 0 | **7** | SUPPORT |
| **Users** | 1 | 2 | 2 | 1 | 0 | **6** | SUPPORT |
| **Events** | 3 | 1 | 1 | 0 | 2 | **7** | SUPPORT |
| **Releases** | 1 | 1 | 1 | 1 | 1 | **5** | API-ONLY |
| **Reviews** | 1 | 2 | 2 | 0 | 1 | **6** | SUPPORT |
| **Code Scanning Alerts** | 1 | 1 | 1 | 0 | 1 | **4** | API-ONLY |
| **Dependabot Alerts** | 1 | 1 | 1 | 0 | 1 | **4** | API-ONLY |
| **Labels** | 0 | 1 | 2 | 0 | 0 | **3** | API-ONLY |
| **Milestones** | 0 | 1 | 1 | 0 | 0 | **2** | API-ONLY |
| **Teams** | 0 | 1 | 1 | 0 | 0 | **2** | API-ONLY |

**Primary entities (score >= 8):** Issues (12), Pull Requests (11), Comments (10), Commits (9)
**Support entities (score 5-7):** Workflow Runs (7), Repositories (7), Events (7), Users (6), Reviews (6)

## SQLite Schema

### Primary Entity: Issues (Data Gravity: 12)

```sql
CREATE TABLE issues (
    id INTEGER PRIMARY KEY,
    number INTEGER NOT NULL,
    repo_id INTEGER NOT NULL REFERENCES repos(id),
    user_id INTEGER REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT,
    state TEXT NOT NULL DEFAULT 'open',
    state_reason TEXT,
    locked INTEGER NOT NULL DEFAULT 0,
    comments_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    closed_at TEXT,
    is_pull_request INTEGER NOT NULL DEFAULT 0,
    labels TEXT, -- JSON array of label names
    assignees TEXT, -- JSON array of user logins
    milestone_number INTEGER,
    data JSON NOT NULL
);
CREATE INDEX idx_issues_repo ON issues(repo_id);
CREATE INDEX idx_issues_user ON issues(user_id);
CREATE INDEX idx_issues_state ON issues(state);
CREATE INDEX idx_issues_updated ON issues(updated_at);
CREATE INDEX idx_issues_repo_state ON issues(repo_id, state);

CREATE VIRTUAL TABLE issues_fts USING fts5(
    title, body, content='issues', content_rowid='id'
);
```

### Primary Entity: Pull Requests (Data Gravity: 11)

```sql
CREATE TABLE pull_requests (
    id INTEGER PRIMARY KEY,
    number INTEGER NOT NULL,
    repo_id INTEGER NOT NULL REFERENCES repos(id),
    user_id INTEGER REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT,
    state TEXT NOT NULL DEFAULT 'open',
    draft INTEGER NOT NULL DEFAULT 0,
    merged INTEGER NOT NULL DEFAULT 0,
    mergeable TEXT,
    head_ref TEXT,
    base_ref TEXT,
    additions INTEGER,
    deletions INTEGER,
    changed_files INTEGER,
    comments_count INTEGER,
    review_comments_count INTEGER,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    closed_at TEXT,
    merged_at TEXT,
    labels TEXT, -- JSON array of label names
    assignees TEXT, -- JSON array of user logins
    requested_reviewers TEXT, -- JSON array of user logins
    data JSON NOT NULL
);
CREATE INDEX idx_prs_repo ON pull_requests(repo_id);
CREATE INDEX idx_prs_user ON pull_requests(user_id);
CREATE INDEX idx_prs_state ON pull_requests(state);
CREATE INDEX idx_prs_updated ON pull_requests(updated_at);
CREATE INDEX idx_prs_repo_state ON pull_requests(repo_id, state);

CREATE VIRTUAL TABLE pull_requests_fts USING fts5(
    title, body, content='pull_requests', content_rowid='id'
);
```

### Primary Entity: Comments (Data Gravity: 10)

```sql
CREATE TABLE comments (
    id INTEGER PRIMARY KEY,
    issue_id INTEGER REFERENCES issues(id),
    pull_request_id INTEGER REFERENCES pull_requests(id),
    user_id INTEGER REFERENCES users(id),
    body TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    data JSON NOT NULL
);
CREATE INDEX idx_comments_issue ON comments(issue_id);
CREATE INDEX idx_comments_pr ON comments(pull_request_id);
CREATE INDEX idx_comments_user ON comments(user_id);
CREATE INDEX idx_comments_updated ON comments(updated_at);

CREATE VIRTUAL TABLE comments_fts USING fts5(
    body, content='comments', content_rowid='id'
);
```

### Primary Entity: Commits (Data Gravity: 9)

```sql
CREATE TABLE commits (
    sha TEXT PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repos(id),
    author_id INTEGER REFERENCES users(id),
    committer_id INTEGER REFERENCES users(id),
    message TEXT NOT NULL,
    authored_date TEXT NOT NULL,
    committed_date TEXT NOT NULL,
    additions INTEGER,
    deletions INTEGER,
    data JSON NOT NULL
);
CREATE INDEX idx_commits_repo ON commits(repo_id);
CREATE INDEX idx_commits_author ON commits(author_id);
CREATE INDEX idx_commits_date ON commits(authored_date);

CREATE VIRTUAL TABLE commits_fts USING fts5(
    message, content='commits', content_rowid='rowid'
);
```

### Support Entity: Repositories (Data Gravity: 7)

```sql
CREATE TABLE repos (
    id INTEGER PRIMARY KEY,
    owner_id INTEGER REFERENCES users(id),
    name TEXT NOT NULL,
    full_name TEXT NOT NULL UNIQUE,
    description TEXT,
    private INTEGER NOT NULL DEFAULT 0,
    fork INTEGER NOT NULL DEFAULT 0,
    language TEXT,
    stargazers_count INTEGER,
    forks_count INTEGER,
    open_issues_count INTEGER,
    default_branch TEXT,
    created_at TEXT,
    updated_at TEXT,
    pushed_at TEXT,
    archived INTEGER NOT NULL DEFAULT 0,
    data JSON NOT NULL
);
CREATE INDEX idx_repos_owner ON repos(owner_id);
CREATE INDEX idx_repos_full_name ON repos(full_name);
```

### Support Entity: Users (Data Gravity: 6)

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    name TEXT,
    email TEXT,
    avatar_url TEXT,
    type TEXT NOT NULL DEFAULT 'User',
    data JSON NOT NULL
);
CREATE INDEX idx_users_login ON users(login);
```

### Support Entity: Workflow Runs (Data Gravity: 7)

```sql
CREATE TABLE workflow_runs (
    id INTEGER PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repos(id),
    workflow_id INTEGER NOT NULL,
    name TEXT,
    head_branch TEXT,
    head_sha TEXT,
    status TEXT,
    conclusion TEXT,
    run_number INTEGER,
    event TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT,
    run_started_at TEXT,
    data JSON NOT NULL
);
CREATE INDEX idx_wfruns_repo ON workflow_runs(repo_id);
CREATE INDEX idx_wfruns_workflow ON workflow_runs(workflow_id);
CREATE INDEX idx_wfruns_conclusion ON workflow_runs(conclusion);
CREATE INDEX idx_wfruns_created ON workflow_runs(created_at);
```

### Support Entity: Events (Data Gravity: 7)

```sql
CREATE TABLE events (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    actor_id INTEGER REFERENCES users(id),
    repo_id INTEGER REFERENCES repos(id),
    created_at TEXT NOT NULL,
    data JSON NOT NULL
);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_actor ON events(actor_id);
CREATE INDEX idx_events_repo ON events(repo_id);
CREATE INDEX idx_events_created ON events(created_at);
```

### Support Entity: Reviews (Data Gravity: 6)

```sql
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY,
    pull_request_id INTEGER NOT NULL REFERENCES pull_requests(id),
    user_id INTEGER REFERENCES users(id),
    state TEXT NOT NULL,
    body TEXT,
    submitted_at TEXT,
    data JSON NOT NULL
);
CREATE INDEX idx_reviews_pr ON reviews(pull_request_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);
```

## Sync Strategy

### Issues & Pull Requests (Incremental via `since` param)
- **Cursor field:** `updated_at` timestamp
- **API support:** VALIDATED - issues endpoint has `since` as `#/components/parameters/since` ($ref parameter)
- **Strategy:** `GET /repos/{owner}/{repo}/issues?state=all&sort=updated&direction=asc&since={last_sync}&per_page=100`
- **Note:** GitHub's issues endpoint returns BOTH issues and PRs. Filter by `pull_request` field presence.
- **Batch size:** 100 (API max for per_page)
- **Incremental:** Store max(updated_at) per repo as cursor. On next sync, only fetch updated items.

### Commits (Incremental via `since` param)
- **Cursor field:** `since` timestamp parameter
- **API support:** VALIDATED - commits endpoint has explicit `since` and `until` query params
- **Strategy:** `GET /repos/{owner}/{repo}/commits?since={last_sync}&per_page=100`
- **Batch size:** 100

### Workflow Runs (Date range via `created` param)
- **Cursor field:** `created_at`
- **API support:** VALIDATED - workflow runs endpoint has `created` as `#/components/parameters/created`
- **Strategy:** `GET /repos/{owner}/{repo}/actions/runs?created=>={last_sync_date}&per_page=100`
- **Note:** The `created` param supports date range expressions like `>=2026-03-01`
- **Batch size:** 100

### Comments (Incremental via `since` param)
- **Cursor field:** `updated_at`
- **API support:** Issue comments list endpoint supports `since` and `direction` params
- **Strategy:** `GET /repos/{owner}/{repo}/issues/comments?sort=updated&direction=asc&since={last_sync}&per_page=100`
- **Batch size:** 100

### Repositories (Full refresh)
- **Cursor:** None needed - small cardinality
- **Strategy:** `GET /orgs/{org}/repos?per_page=100` or `GET /user/repos?per_page=100`
- **Frequency:** On every sync, refresh full repo list

### Users (Lazy population)
- **Strategy:** Extract user objects from issues, PRs, commits during sync. Upsert on encounter.
- **No dedicated sync pass** - users are populated as side effects of other syncs.

### Events (Tail via REST polling)
- **Cursor field:** Event ID or created_at
- **API support:** Events API returns newest-first, max 300 events, max 90 days
- **Strategy:** Poll `GET /repos/{owner}/{repo}/events?per_page=100`, store all, dedup by ID
- **Tail command:** Poll every 30s, show new events since last poll

### Rate Limit Awareness
- Track `X-RateLimit-Remaining` and `X-RateLimit-Reset` headers
- Pause sync when remaining < 100
- Use conditional requests (ETag/If-Modified-Since) where possible - 304 responses don't count against rate limit
- Display rate limit status in `doctor` output

## Domain-Specific Search Filters

| CLI Flag | SQL WHERE Clause | Applicable To |
|----------|-----------------|---------------|
| `--repo owner/name` | `WHERE repo_id = (SELECT id FROM repos WHERE full_name = ?)` | issues, PRs, commits, workflow_runs |
| `--org name` | `WHERE repo_id IN (SELECT id FROM repos WHERE owner_id = (SELECT id FROM users WHERE login = ?))` | all |
| `--author login` | `WHERE user_id = (SELECT id FROM users WHERE login = ?)` | issues, PRs |
| `--author login` | `WHERE author_id = (SELECT id FROM users WHERE login = ?)` | commits |
| `--state open/closed` | `WHERE state = ?` | issues, PRs |
| `--label name` | `WHERE labels LIKE '%"name"%'` (JSON contains) | issues, PRs |
| `--days N` | `WHERE updated_at >= datetime('now', '-N days')` | issues, PRs, comments |
| `--since date` | `WHERE updated_at >= ?` | issues, PRs, commits, comments |
| `--merged` | `WHERE merged = 1` | PRs |
| `--draft` | `WHERE draft = 1` | PRs |
| `--conclusion success/failure` | `WHERE conclusion = ?` | workflow_runs |
| `--workflow name` | `WHERE name = ?` | workflow_runs |

## Compound Cross-Entity Queries

### 1. "PRs by author with review status" (pr-triage)
```sql
SELECT p.number, p.title, p.state, u.login AS author,
       (SELECT GROUP_CONCAT(u2.login) FROM reviews r JOIN users u2 ON r.user_id = u2.id
        WHERE r.pull_request_id = p.id AND r.state = 'APPROVED') AS approvers,
       p.created_at, p.updated_at
FROM pull_requests p
JOIN repos r2 ON p.repo_id = r2.id
JOIN users u ON p.user_id = u.id
WHERE p.state = 'open'
ORDER BY p.updated_at ASC;
```
**Validated:** pull_requests.user_id -> users.id, reviews.pull_request_id -> pull_requests.id

### 2. "Stale issues across org" (stale)
```sql
SELECT i.number, r.full_name, i.title, i.state, i.updated_at,
       julianday('now') - julianday(i.updated_at) AS days_stale
FROM issues i
JOIN repos r ON i.repo_id = r.id
WHERE i.state = 'open'
  AND i.is_pull_request = 0
  AND i.updated_at < datetime('now', '-30 days')
ORDER BY days_stale DESC;
```
**Validated:** issues.repo_id -> repos.id

### 3. "CI success rate per workflow" (actions-health)
```sql
SELECT w.name, w.repo_id, r.full_name,
       COUNT(*) AS total_runs,
       SUM(CASE WHEN w.conclusion = 'success' THEN 1 ELSE 0 END) AS successes,
       ROUND(100.0 * SUM(CASE WHEN w.conclusion = 'success' THEN 1 ELSE 0 END) / COUNT(*), 1) AS success_rate
FROM workflow_runs w
JOIN repos r ON w.repo_id = r.id
WHERE w.created_at >= datetime('now', '-14 days')
GROUP BY w.workflow_id, w.repo_id
ORDER BY success_rate ASC;
```
**Validated:** workflow_runs.repo_id -> repos.id

### 4. "Top contributors by commit count" (contributors)
```sql
SELECT u.login, u.name, COUNT(c.sha) AS commit_count,
       (SELECT COUNT(*) FROM pull_requests p WHERE p.user_id = u.id AND p.merged = 1) AS merged_prs
FROM commits c
JOIN users u ON c.author_id = u.id
JOIN repos r ON c.repo_id = r.id
WHERE c.authored_date >= datetime('now', '-30 days')
GROUP BY u.id
ORDER BY commit_count DESC
LIMIT 20;
```
**Validated:** commits.author_id -> users.id, commits.repo_id -> repos.id

### 5. "Full-text search across issues, PRs, and comments"
```sql
SELECT 'issue' AS type, i.number, r.full_name, i.title, snippet(issues_fts, 1, '<b>', '</b>', '...', 20) AS match
FROM issues_fts
JOIN issues i ON issues_fts.rowid = i.id
JOIN repos r ON i.repo_id = r.id
WHERE issues_fts MATCH ?
UNION ALL
SELECT 'pr' AS type, p.number, r.full_name, p.title, snippet(pull_requests_fts, 1, '<b>', '</b>', '...', 20) AS match
FROM pull_requests_fts
JOIN pull_requests p ON pull_requests_fts.rowid = p.id
JOIN repos r ON p.repo_id = r.id
WHERE pull_requests_fts MATCH ?
ORDER BY rank
LIMIT 50;
```
**Validated:** FTS5 tables reference correct content tables

## Tail Strategy

**Decision: REST Polling** (only option for GitHub REST API)

GitHub's REST API has no WebSocket or SSE endpoints for general events. The Events API is the closest to real-time but is still REST-based.

| Method | Availability | Decision |
|--------|-------------|----------|
| WebSocket/Gateway | NO - GitHub has no WebSocket API | N/A |
| SSE | NO - GitHub has no SSE endpoints | N/A |
| Webhooks | YES but requires a server | Not CLI-friendly |
| REST Polling | YES - Events API + per-page pagination | **USE THIS** |

**Tail implementation:**
- Poll `/repos/{owner}/{repo}/events?per_page=30` every 30 seconds
- Track last seen event ID to avoid duplicates
- Use `If-None-Match` (ETag) header to avoid rate limit consumption on 304
- Display new events as they arrive with type, actor, timestamp
- Support `--type PushEvent,PullRequestEvent` filtering

## Commands to Build in Phase 4 Priority 0

1. **sync** - Incremental sync of issues, PRs, commits, workflow runs, comments to local SQLite
2. **search** - Full-text search across issues, PRs, comments with domain filters
3. **sql** - Raw read-only SQL access to the local database
4. **issues** - Query local DB for issues with --repo, --author, --state, --days, --label filters
5. **prs** - Query local DB for PRs with --repo, --author, --state, --merged, --draft filters
6. **tail** - Stream new events from a repo via REST polling

## Acceptance Criteria

- [x] Entity classification for 15 API resources
- [x] 4 social signals with evidence scores >= 6
- [x] Data gravity computed for all entities, 4 primary (score >= 8)
- [x] SQLite schema with domain columns for all primary + support entities
- [x] FTS5 on issues (title, body), PRs (title, body), comments (body), commits (message)
- [x] Sync cursors validated: issues (since param), commits (since param), workflow_runs (created param)
- [x] 12 domain-specific search filters mapped to SQL WHERE clauses
- [x] 5 compound queries validated (all joins confirmed)
- [x] Tail strategy decided: REST polling (GitHub has no WS/SSE)

## Sources
- GitHub REST API OpenAPI spec - parameter validation
- github-to-sqlite (462 stars) - prior art for table structure
- GitHub rate limiting docs - conditional request strategy
- gh-dash (11.2k stars) - cross-repo query demand evidence
