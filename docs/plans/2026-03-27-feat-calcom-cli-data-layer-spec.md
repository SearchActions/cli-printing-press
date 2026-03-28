---
title: "Data Layer Specification: Cal.com CLI"
type: feat
status: active
date: 2026-03-27
phase: "0.7"
api: "cal.com"
---

# Data Layer Specification: Cal.com CLI

## Entity Classification

| Entity | Type | Est. Volume | Update Freq | Temporal Field | Persistence Need |
|--------|------|-------------|-------------|----------------|-----------------|
| **Bookings** | Accumulating | 100-10,000/account | Daily | `createdAt`, `updatedAt` | SQLite + incremental sync |
| **Event Types** | Reference | 5-50/account | Weekly | N/A | SQLite + periodic refresh |
| **Attendees** | Accumulating (nested) | Same as bookings | With parent booking | via booking `updatedAt` | Extracted from booking data |
| **Schedules** | Reference | 1-5/account | Monthly | N/A | SQLite + periodic refresh |
| **Teams** | Reference | 1-10/account | Rarely | N/A | SQLite + periodic refresh |
| **Calendars** | Ephemeral | 2-5 connections | Rarely | N/A | API-only (live check) |
| **Webhooks** | Reference | 1-20/account | Rarely | N/A | API-only |
| **Slots** | Ephemeral | Generated per query | N/A | N/A | API-only |

## Data Gravity Scoring

| Entity | Volume (0-3) | QueryFreq (0-3) | JoinDemand (0-2) | SearchNeed (0-2) | TemporalValue (0-2) | **Total** |
|--------|-------------|-----------------|-----------------|-----------------|--------------------|-----------|
| **Bookings** | 2 | 3 (daily: agenda, search) | 2 (event_types, attendees) | 2 (title, description) | 2 (trends, analytics) | **11** |
| **Attendees** | 2 | 3 (daily: search by name/email) | 2 (linked to bookings) | 2 (name, email) | 1 | **10** |
| **Event Types** | 1 | 2 (weekly: stale, stats) | 2 (referenced by bookings) | 1 (title, slug) | 0 | **6** |
| **Schedules** | 0 | 1 | 1 | 0 | 0 | **2** |
| **Teams** | 0 | 1 | 1 | 0 | 0 | **2** |

**Primary entities (>= 8):** Bookings (11), Attendees (10)
**Support entities (5-7):** Event Types (6)
**API-only (< 5):** Schedules, Teams, Calendars, Webhooks, Slots

## Social Signal Mining

Evidence for local data persistence:
1. No Cal.com backup/export tools exist (GitHub search confirmed)
2. Calendar sync issues are #1 complaint — local copy is insurance
3. callytics web dashboard (0 stars) tried analytics but no CLI
4. n8n/Zapier integrations show demand for data out of Cal.com
5. gcalcli (3k stars) proves calendar CLIs succeed by enabling offline workflows

## SQLite Schema

```sql
CREATE TABLE IF NOT EXISTS bookings (
    id          INTEGER PRIMARY KEY,
    uid         TEXT NOT NULL UNIQUE,
    title       TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT '',
    start_time  TEXT NOT NULL DEFAULT '',
    end_time    TEXT NOT NULL DEFAULT '',
    duration    INTEGER NOT NULL DEFAULT 0,
    location    TEXT NOT NULL DEFAULT '',
    meeting_url TEXT NOT NULL DEFAULT '',
    event_type_id    INTEGER NOT NULL DEFAULT 0,
    event_type_slug  TEXT NOT NULL DEFAULT '',
    cancellation_reason TEXT NOT NULL DEFAULT '',
    rescheduling_reason TEXT NOT NULL DEFAULT '',
    rating      REAL NOT NULL DEFAULT 0,
    absent_host INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT '',
    updated_at  TEXT NOT NULL DEFAULT '',
    data        TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_start ON bookings(start_time);
CREATE INDEX IF NOT EXISTS idx_bookings_updated ON bookings(updated_at);
CREATE INDEX IF NOT EXISTS idx_bookings_event_type ON bookings(event_type_id);

CREATE TABLE IF NOT EXISTS attendees (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    booking_id  INTEGER NOT NULL REFERENCES bookings(id),
    booking_uid TEXT NOT NULL DEFAULT '',
    name        TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',
    phone       TEXT NOT NULL DEFAULT '',
    timezone    TEXT NOT NULL DEFAULT '',
    absent      INTEGER NOT NULL DEFAULT 0,
    language    TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_attendees_booking ON attendees(booking_id);
CREATE INDEX IF NOT EXISTS idx_attendees_email ON attendees(email);

CREATE TABLE IF NOT EXISTS event_types (
    id              INTEGER PRIMARY KEY,
    title           TEXT NOT NULL DEFAULT '',
    slug            TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    length_minutes  INTEGER NOT NULL DEFAULT 0,
    hidden          INTEGER NOT NULL DEFAULT 0,
    price           REAL NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT '',
    owner_id        INTEGER NOT NULL DEFAULT 0,
    schedule_id     INTEGER NOT NULL DEFAULT 0,
    booking_url     TEXT NOT NULL DEFAULT '',
    data            TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_event_types_slug ON event_types(slug);
CREATE INDEX IF NOT EXISTS idx_event_types_owner ON event_types(owner_id);

-- FTS5 on bookings
CREATE VIRTUAL TABLE IF NOT EXISTS bookings_fts USING fts5(
    title, description,
    content='bookings', content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS bookings_ai AFTER INSERT ON bookings BEGIN
    INSERT INTO bookings_fts(rowid, title, description) VALUES (new.id, new.title, new.description);
END;
CREATE TRIGGER IF NOT EXISTS bookings_ad AFTER DELETE ON bookings BEGIN
    INSERT INTO bookings_fts(bookings_fts, rowid, title, description) VALUES ('delete', old.id, old.title, old.description);
END;
CREATE TRIGGER IF NOT EXISTS bookings_au AFTER UPDATE ON bookings BEGIN
    INSERT INTO bookings_fts(bookings_fts, rowid, title, description) VALUES ('delete', old.id, old.title, old.description);
    INSERT INTO bookings_fts(rowid, title, description) VALUES (new.id, new.title, new.description);
END;

-- FTS5 on attendees
CREATE VIRTUAL TABLE IF NOT EXISTS attendees_fts USING fts5(
    name, email,
    content='attendees', content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS attendees_ai AFTER INSERT ON attendees BEGIN
    INSERT INTO attendees_fts(rowid, name, email) VALUES (new.id, new.name, new.email);
END;
CREATE TRIGGER IF NOT EXISTS attendees_ad AFTER DELETE ON attendees BEGIN
    INSERT INTO attendees_fts(attendees_fts, rowid, name, email) VALUES ('delete', old.id, old.name, old.email);
END;
CREATE TRIGGER IF NOT EXISTS attendees_au AFTER UPDATE ON attendees BEGIN
    INSERT INTO attendees_fts(attendees_fts, rowid, name, email) VALUES ('delete', old.id, old.name, old.email);
    INSERT INTO attendees_fts(rowid, name, email) VALUES (new.id, new.name, new.email);
END;

-- Sync state tracking
CREATE TABLE IF NOT EXISTS sync_state (
    entity      TEXT PRIMARY KEY,
    last_synced TEXT NOT NULL DEFAULT '',
    cursor      TEXT NOT NULL DEFAULT '',
    total_items INTEGER NOT NULL DEFAULT 0
);
```

## Sync Strategy

### Bookings — Incremental via `afterUpdatedAt`

**VALIDATED:** `afterUpdatedAt` query param confirmed in OpenAPI spec for GET /v2/bookings.

1. Read `sync_state` for entity='bookings' to get last cursor
2. GET /v2/bookings?afterUpdatedAt={cursor}&take=100&sortUpdatedAt=asc
3. For each booking: upsert into `bookings`, delete+reinsert `attendees`
4. Update sync_state with latest updatedAt
5. Paginate with skip if hasNextPage
6. Also sync cancelled separately (afterUpdatedAt + status=cancelled)

**Batch size:** 100 (API max), **Rate budget:** 100 pages = ~50s at 120 req/min

### Event Types — Full Refresh
GET /v2/event-types → upsert all → update sync_state

### Scoping Flags
- `--team <id>` → adds teamId filter
- `--since <date>` → overrides cursor
- `--status <s>` → filters by status
- `--max-pages N` → limits pagination (default: unlimited)

## Search Filters

| CLI Flag | SQL WHERE | Description |
|----------|----------|-------------|
| `--status` | `WHERE status = ?` | Filter by booking status |
| `--attendee` | `JOIN attendees ... WHERE name LIKE ?` | Attendee name search |
| `--email` | `JOIN attendees ... WHERE email = ?` | Exact email match |
| `--event-type` | `WHERE event_type_id = ?` | Filter by event type |
| `--after` | `WHERE start_time >= ?` | Start date |
| `--before` | `WHERE start_time <= ?` | End date |
| `--days N` | `WHERE start_time >= datetime('now', '-N days')` | Relative window |
| (positional) | FTS5 MATCH on bookings_fts + attendees_fts | Free text search |

## Compound Queries

### 1. Search by attendee
```sql
SELECT b.*, a.name, a.email FROM bookings b
JOIN attendees a ON b.id = a.booking_id
JOIN attendees_fts ON attendees_fts.rowid = a.id
WHERE attendees_fts MATCH ? ORDER BY b.start_time DESC LIMIT ?
```

### 2. Stats by event type
```sql
SELECT et.title, COUNT(*) as total,
  SUM(CASE WHEN b.status='cancelled' THEN 1 ELSE 0 END) as cancelled,
  ROUND(100.0*SUM(CASE WHEN b.status='past' THEN 1 ELSE 0 END)/COUNT(*),1) as show_rate
FROM bookings b LEFT JOIN event_types et ON b.event_type_id = et.id
WHERE b.start_time >= ? GROUP BY et.id ORDER BY total DESC
```

### 3. Stale event types
```sql
SELECT et.id, et.title, et.slug, COUNT(b.id) as cnt, MAX(b.start_time) as last
FROM event_types et LEFT JOIN bookings b ON et.id = b.event_type_id
  AND b.start_time >= datetime('now', '-'||?||' days')
GROUP BY et.id HAVING cnt = 0 ORDER BY et.title
```

### 4. Today's agenda
```sql
SELECT b.*, GROUP_CONCAT(a.name, ', ') as attendees
FROM bookings b LEFT JOIN attendees a ON b.id = a.booking_id
WHERE b.start_time >= ? AND b.start_time < ? AND b.status != 'cancelled'
GROUP BY b.id ORDER BY b.start_time ASC
```

### 5. Conflicts (overlapping bookings)
```sql
SELECT b1.title, b1.start_time, b1.end_time, b2.title as conflict, b2.start_time as c_start
FROM bookings b1 JOIN bookings b2 ON b1.id < b2.id
  AND b1.start_time < b2.end_time AND b1.end_time > b2.start_time
WHERE b1.status != 'cancelled' AND b2.status != 'cancelled' AND b1.start_time >= ?
```

## Tail Strategy

**Decision: REST polling** — Cal.com has no WebSocket/SSE for clients. Webhooks are push (server-to-server). Use afterUpdatedAt polling every 30-60s for a `tail` command.

## Phase 4 Priority 0 Commands

| Command | Data Path |
|---------|-----------|
| sync | WRITE: UpsertBooking + UpsertAttendees + UpsertEventType |
| search | READ: bookings_fts + attendees_fts |
| sql | READ: any table |
| agenda | READ: bookings + attendees (compound query 4) |
| stats | READ: bookings + event_types (compound query 2) |
| stale | READ: event_types LEFT JOIN bookings (compound query 3) |
| conflicts | READ: bookings self-join (compound query 5) |
