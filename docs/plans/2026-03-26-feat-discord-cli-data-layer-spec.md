---
title: "Data Layer Specification: Discord CLI"
type: feat
status: active
date: 2026-03-26
phase: "0.7"
api: "discord"
---

# Data Layer Specification: Discord CLI

## Entity Classification

| Entity | Type | Est. Volume | Update Freq | Key Temporal Field | Persistence Need |
|--------|------|-------------|-------------|-------------------|-----------------|
| **Messages** | Accumulating | Millions | Continuous | `id` (snowflake) | SQLite + FTS5 |
| **Members** | Reference | Thousands | Daily | `joined_at` | SQLite table |
| **Channels** | Reference | Hundreds | Weekly | `id` | SQLite table |
| **Roles** | Reference | Tens | Monthly | `id` | SQLite table |
| **Audit Log Entries** | Append-only | Thousands/month | Continuous | `id` (snowflake) | SQLite table |
| **Threads** | Accumulating | Hundreds-thousands | Daily | `id` (snowflake) | SQLite table (as channels with thread metadata) |
| **Guilds** | Reference | Single-digit | Rarely | `id` | SQLite table |
| **Emojis** | Reference | Tens-hundreds | Monthly | `id` | API-only (low value) |
| **Stickers** | Reference | Tens | Monthly | `id` | API-only |
| **Webhooks** | Reference | Tens | Monthly | `id` | API-only |
| **Invites** | Ephemeral | Tens | On-demand | `created_at` | API-only |
| **Scheduled Events** | Ephemeral | Tens | Weekly | `id` | API-only |
| **Auto Mod Rules** | Reference | Tens | Monthly | `id` | API-only |
| **Stage Instances** | Ephemeral | Few | Rarely | N/A | API-only |

**Heuristics applied:**
- Messages: has created timestamps (snowflake IDs encode time), paginated list endpoint, no UPDATE -> Accumulating
- Members: referenced by messages via author_id, has joined_at, changes with role updates -> Reference
- Channels: referenced by messages (3+ entities), rarely deleted -> Reference
- Audit log: no UPDATE/PATCH, append-only by design -> Append-only

## Social Signal Mining Results

| # | Finding | Source | Evidence Score |
|---|---------|--------|---------------|
| 1 | 9+ tools store Discord messages in SQLite | GitHub search | **8** (discrawl=3, discord-cli=2, discord-migrate=1, discord-sqlite-exporter=1, Discord-Archiver=1) |
| 2 | FTS5 search over messages is the killer feature | discrawl, jackwener/discord-cli | **7** (discrawl=3, discord-cli=2, HN search complaints=2) |
| 3 | Cross-entity queries (messages x members x channels) | chat-analytics, discord-analytics | **6** (chat-analytics tools=2, discord-analytics=2, cross-platform=2) |
| 4 | Offline member directory (servers >1000 hide offline) | Discord support forums, Reddit | **6** (feature requests=2, discrawl members=2, cross-platform=2) |
| 5 | Audit log persistence for incident investigation | GitHub issues, moderation bots | **4** (moderation bots=2, API docs discussions=2) |
| 6 | Discord search is broken/limited (HN complaint) | HN item 36748981 | **3** (HN post=2, Discorch=1) |
| 7 | Trend/activity analysis over time windows | discord-analytics repos, chat-analytics | **6** (3 analytics tools=3, visualization demand=1, cross-platform=2) |

Signals >= 6 inform the data layer: **messages in SQLite (8), FTS5 search (7), cross-entity queries (6), offline member dir (6), trend analysis (6).**

## Data Gravity Scoring

| Entity | Volume (0-3) | QueryFreq (0-3) | JoinDemand (0-2) | SearchNeed (0-2) | TemporalValue (0-2) | **Total** |
|--------|-------------|-----------------|------------------|-----------------|--------------------:|-----------|
| **Messages** | 3 (>1M) | 3 (daily search) | 2 (channel_id, author_id, guild_id) | 2 (content is primary text) | 2 (core time-series) | **12** |
| **Members** | 2 (10k+) | 2 (weekly lookups) | 2 (referenced by messages, audit) | 1 (username/nickname) | 1 (joined_at) | **8** |
| **Channels** | 1 (100-1000) | 2 (weekly) | 2 (referenced by messages, threads) | 1 (name) | 0 (no time dimension) | **6** |
| **Audit Log** | 2 (10k+/year) | 1 (monthly) | 2 (user_id, target_id) | 0 (no text content) | 2 (core time-series) | **7** |
| **Roles** | 0 (<100) | 1 (monthly) | 2 (referenced by members) | 0 | 0 | **3** |
| **Guilds** | 0 (<10) | 1 (rarely) | 2 (referenced by everything) | 0 | 0 | **3** |

**Thresholds:**
- **Primary (>=8): Messages (12), Members (8)** - full SQLite tables with proper columns + FTS5
- **Support (5-7): Channels (6), Audit Log (7)** - simpler tables
- **API-only (<5): Roles (3), Guilds (3)** - no persistence needed

## SQLite Schema

```sql
-- Primary Entity: Messages (Data Gravity: 12)
CREATE TABLE messages (
    id            TEXT PRIMARY KEY,  -- snowflake ID
    channel_id    TEXT NOT NULL,
    guild_id      TEXT NOT NULL,
    author_id     TEXT NOT NULL,
    content       TEXT NOT NULL DEFAULT '',
    timestamp     TEXT NOT NULL,     -- ISO 8601
    edited_at     TEXT,
    type          INTEGER NOT NULL DEFAULT 0,
    pinned        INTEGER NOT NULL DEFAULT 0,
    mention_everyone INTEGER NOT NULL DEFAULT 0,
    tts           INTEGER NOT NULL DEFAULT 0,
    data          TEXT NOT NULL      -- full JSON response
);

CREATE INDEX idx_messages_channel ON messages(channel_id);
CREATE INDEX idx_messages_author ON messages(author_id);
CREATE INDEX idx_messages_guild ON messages(guild_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);
CREATE INDEX idx_messages_channel_ts ON messages(channel_id, timestamp);

-- FTS5 on message content (primary text field)
CREATE VIRTUAL TABLE messages_fts USING fts5(
    content,
    content='messages',
    content_rowid='rowid'
);

-- Triggers to keep FTS5 in sync
CREATE TRIGGER messages_ai AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(rowid, content) VALUES (new.rowid, new.content);
END;
CREATE TRIGGER messages_ad AFTER DELETE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, content) VALUES ('delete', old.rowid, old.content);
END;
CREATE TRIGGER messages_au AFTER UPDATE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, content) VALUES ('delete', old.rowid, old.content);
    INSERT INTO messages_fts(rowid, content) VALUES (new.rowid, new.content);
END;

-- Primary Entity: Members (Data Gravity: 8)
CREATE TABLE members (
    guild_id      TEXT NOT NULL,
    user_id       TEXT NOT NULL,
    username      TEXT NOT NULL DEFAULT '',
    display_name  TEXT NOT NULL DEFAULT '',
    nickname      TEXT,
    avatar        TEXT,
    joined_at     TEXT,
    roles         TEXT NOT NULL DEFAULT '[]',  -- JSON array of role IDs
    deaf          INTEGER NOT NULL DEFAULT 0,
    mute          INTEGER NOT NULL DEFAULT 0,
    data          TEXT NOT NULL,  -- full JSON response
    PRIMARY KEY (guild_id, user_id)
);

CREATE INDEX idx_members_username ON members(username);
CREATE INDEX idx_members_joined ON members(joined_at);

-- FTS5 on member names
CREATE VIRTUAL TABLE members_fts USING fts5(
    username, display_name, nickname,
    content='members',
    content_rowid='rowid'
);

-- Support Entity: Channels (Data Gravity: 6)
CREATE TABLE channels (
    id            TEXT PRIMARY KEY,
    guild_id      TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    type          INTEGER NOT NULL DEFAULT 0,
    parent_id     TEXT,
    position      INTEGER NOT NULL DEFAULT 0,
    topic         TEXT,
    nsfw          INTEGER NOT NULL DEFAULT 0,
    last_message_id TEXT,
    thread_metadata TEXT,  -- JSON for thread-specific fields
    data          TEXT NOT NULL
);

CREATE INDEX idx_channels_guild ON channels(guild_id);
CREATE INDEX idx_channels_parent ON channels(parent_id);

-- Support Entity: Audit Log (Data Gravity: 7)
CREATE TABLE audit_log (
    id            TEXT PRIMARY KEY,  -- snowflake ID
    guild_id      TEXT NOT NULL,
    user_id       TEXT,
    target_id     TEXT,
    action_type   INTEGER NOT NULL,
    reason        TEXT,
    changes       TEXT,  -- JSON array
    options       TEXT,  -- JSON object
    data          TEXT NOT NULL
);

CREATE INDEX idx_audit_guild ON audit_log(guild_id);
CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_action ON audit_log(action_type);
CREATE INDEX idx_audit_target ON audit_log(target_id);

-- Sync state tracking
CREATE TABLE sync_state (
    guild_id      TEXT NOT NULL,
    channel_id    TEXT NOT NULL,
    last_message_id TEXT,
    last_synced   TEXT NOT NULL,
    message_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (guild_id, channel_id)
);
```

## Sync Strategy

### Messages (Primary - cursor: snowflake ID)
- **Cursor field:** `id` (snowflake - encodes timestamp, natural ordering)
- **API endpoint:** `GET /channels/{channel_id}/messages?after={last_id}&limit=100`
- **VALIDATED:** The `after` parameter is confirmed in Discord API v10 docs. Returns messages with ID greater than the given snowflake.
- **Batch size:** 100 (API maximum per request)
- **Sync flow:**
  1. List guild channels (GET /guilds/{id}/channels)
  2. For each text channel: paginate messages using `after` cursor from sync_state
  3. Discover threads: GET /guilds/{id}/threads/active + GET /channels/{id}/threads/archived/public
  4. For each thread: paginate messages the same way
  5. Update sync_state with last_message_id and timestamp
- **Rate limiting:** Per-route rate limits. Typical: 5 req/s per channel endpoint. Use X-RateLimit headers.

### Members (Primary - cursor: snowflake ID)
- **Cursor field:** `user_id` (snowflake)
- **API endpoint:** `GET /guilds/{guild_id}/members?after={last_user_id}&limit=1000`
- **VALIDATED:** The `after` and `limit` params are confirmed. Max 1000 per request.
- **Note:** Requires GUILD_MEMBERS privileged intent for the bot.

### Channels (Support - full refresh)
- **API endpoint:** `GET /guilds/{guild_id}/channels`
- **Strategy:** Full refresh on each sync (low volume, no cursor needed)

### Audit Log (Support - cursor: snowflake ID)
- **Cursor field:** `id` (snowflake, but API uses `before` for older entries)
- **API endpoint:** `GET /guilds/{guild_id}/audit-logs?before={oldest_id}&limit=100`
- **VALIDATED:** `before` and `after` params confirmed. Max 100 per request.
- **Note:** Requires VIEW_AUDIT_LOG permission.

## Domain-Specific Search Filters

| CLI Flag | SQL WHERE Clause | Entity |
|----------|-----------------|--------|
| `--channel <name-or-id>` | `WHERE channel_id = ?` or `WHERE channel_id IN (SELECT id FROM channels WHERE name LIKE ?)` | messages |
| `--author <name-or-id>` | `WHERE author_id = ?` or `WHERE author_id IN (SELECT user_id FROM members WHERE username LIKE ?)` | messages |
| `--guild <id>` | `WHERE guild_id = ?` | messages, members, channels |
| `--before <date>` | `WHERE timestamp < ?` | messages |
| `--after <date>` | `WHERE timestamp > ?` | messages |
| `--days <N>` | `WHERE timestamp > datetime('now', '-N days')` | messages |
| `--hours <N>` | `WHERE timestamp > datetime('now', '-N hours')` | messages |
| `--type <int>` | `WHERE type = ?` | messages |
| `--pinned` | `WHERE pinned = 1` | messages |
| `--role <name-or-id>` | `WHERE roles LIKE '%"role_id"%'` (JSON contains) | members |
| `--action <type>` | `WHERE action_type = ?` | audit_log |
| `--target <id>` | `WHERE target_id = ?` | audit_log |

## Compound Cross-Entity Queries

### 1. Messages by author in channel in last N days
```sql
SELECT m.content, m.timestamp, mb.username, c.name as channel_name
FROM messages m
JOIN members mb ON m.author_id = mb.user_id AND m.guild_id = mb.guild_id
JOIN channels c ON m.channel_id = c.id
WHERE c.name = ? AND mb.username = ? AND m.timestamp > datetime('now', '-30 days')
ORDER BY m.timestamp DESC;
```
**Validated:** messages.author_id -> members.user_id, messages.channel_id -> channels.id

### 2. Top posters per channel (activity command)
```sql
SELECT mb.username, c.name as channel, COUNT(*) as msg_count
FROM messages m
JOIN members mb ON m.author_id = mb.user_id AND m.guild_id = mb.guild_id
JOIN channels c ON m.channel_id = c.id
WHERE m.guild_id = ? AND m.timestamp > datetime('now', '-30 days')
GROUP BY m.author_id, m.channel_id
ORDER BY msg_count DESC
LIMIT 20;
```
**Validated:** Three-way join, all FK columns exist.

### 3. Stale channels (no recent messages)
```sql
SELECT c.name, c.type, MAX(m.timestamp) as last_message
FROM channels c
LEFT JOIN messages m ON c.id = m.channel_id
WHERE c.guild_id = ?
GROUP BY c.id
HAVING last_message IS NULL OR last_message < datetime('now', '-30 days')
ORDER BY last_message ASC;
```
**Validated:** channels.id -> messages.channel_id LEFT JOIN handles empty channels.

### 4. Audit log timeline with actor names
```sql
SELECT a.id, a.action_type, mb.username as actor, a.target_id, a.reason, a.data
FROM audit_log a
LEFT JOIN members mb ON a.user_id = mb.user_id AND a.guild_id = mb.guild_id
WHERE a.guild_id = ? AND a.action_type IN (20, 22, 25)  -- MEMBER_KICK, MEMBER_BAN_ADD, MEMBER_ROLE_UPDATE
ORDER BY a.id DESC
LIMIT 50;
```
**Validated:** audit_log.user_id -> members.user_id

### 5. FTS5 search with channel/author context
```sql
SELECT m.id, m.content, m.timestamp, mb.username, c.name as channel
FROM messages m
JOIN messages_fts ON messages_fts.rowid = m.rowid
JOIN members mb ON m.author_id = mb.user_id AND m.guild_id = mb.guild_id
JOIN channels c ON m.channel_id = c.id
WHERE messages_fts MATCH ?
AND m.guild_id = ?
ORDER BY rank
LIMIT 25;
```
**Validated:** FTS5 JOIN via rowid, context JOINs via FK columns.

## Tail Strategy

| Method | Available? | Decision |
|--------|-----------|----------|
| **Gateway WebSocket** | YES - Discord Gateway with intents (MESSAGE_CREATE, MEMBER_UPDATE, etc.) | **PRIMARY** |
| SSE | No | N/A |
| REST Polling | Yes (GET messages?after=) | **FALLBACK** |

**Decision: Gateway WebSocket (PRIMARY)**

Discord's Gateway is the canonical real-time transport. It provides:
- MESSAGE_CREATE, MESSAGE_UPDATE, MESSAGE_DELETE events
- GUILD_MEMBER_ADD, GUILD_MEMBER_UPDATE, GUILD_MEMBER_REMOVE
- CHANNEL_CREATE, CHANNEL_UPDATE, CHANNEL_DELETE
- All events come with full payloads, no additional API calls needed

The tail command should:
1. Connect to wss://gateway.discord.gg/?v=10&encoding=json
2. Handle HELLO, send IDENTIFY with bot token + intents
3. Maintain heartbeat
4. Filter events by type/guild/channel/keyword
5. Output filtered events as JSON stream (one object per line for piping)
6. Optionally persist events to SQLite (--persist flag)

REST polling is the fallback for bots without Gateway intents or for simpler use cases.

## Commands to Build in Phase 4 Priority 0

| Command | Purpose | Tables Used |
|---------|---------|-------------|
| `sync` | Populate all tables | messages, members, channels, sync_state |
| `search` | FTS5 query with context | messages, messages_fts, members, channels |
| `messages` | Query message slices | messages, members, channels |
| `members` | Offline member directory | members |
| `channels` | Channel listing with stats | channels, messages (for counts) |
| `sql` | Raw read-only SQL | All tables |
| `activity` | Per-member analytics | messages, members, channels |
| `audit` | Audit log with filtering | audit_log, members |
| `tail` | Live Gateway stream | messages (if --persist) |
| `status` | Archive health check | sync_state, messages (counts) |

## Sources

- Discord API: GET /channels/{id}/messages - `after` param confirmed
- Discord API: GET /guilds/{id}/members - `after` + `limit` params confirmed
- Discord API: GET /guilds/{id}/audit-logs - `before`, `after`, `action_type`, `user_id` confirmed
- Discord API: Gateway WebSocket at wss://gateway.discord.gg
- discrawl: validates SQLite+FTS5+Gateway approach (564 stars)
- 9+ GitHub repos store Discord data in SQLite
- HN item 36748981: Discord search is broken/limited
