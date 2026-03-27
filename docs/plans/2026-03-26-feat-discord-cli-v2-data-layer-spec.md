---
title: "Data Layer Specification: Discord CLI v2"
type: feat
status: active
date: 2026-03-26
phase: "0.7"
api: "discord"
---

# Data Layer Specification: Discord CLI v2

## Entity Classification

| Entity | Type | Est. Volume | Update Freq | Temporal Field | Persistence Need |
|--------|------|-------------|-------------|----------------|-----------------|
| **Messages** | Accumulating | Millions | Append-only (edits via events) | `id` (snowflake) | SQLite + FTS5 |
| **Members** | Reference | 100-100k per guild | Moderate (joins/leaves/role changes) | `joined_at` | SQLite table + periodic refresh |
| **Channels** | Reference | 10-500 per guild | Low (created/deleted/renamed) | `id` (snowflake) | SQLite table + periodic refresh |
| **Threads** | Accumulating | 100-10k per guild | Moderate (created/archived) | `id` (snowflake) | SQLite table (stored as channels) |
| **Roles** | Reference | 10-250 per guild | Low | `id` (snowflake) | SQLite table |
| **Guilds** | Reference | 1-100 per bot | Very low | `id` (snowflake) | SQLite table |
| **Audit Log Entries** | Append-only | 100s-1000s per week | Append-only (45-day retention) | `id` (snowflake) | SQLite table |
| **Reactions** | Accumulating | High per popular msg | Append/remove events | N/A (per-message) | SQLite table |
| **Emojis** | Reference | 10-500 per guild | Low | `id` | SQLite table |
| **Invites** | Ephemeral-ish | 10-100 per guild | Low | `created_at` | API-only (small cardinality) |
| **Webhooks** | Reference | 1-50 per guild | Very low | N/A | API-only |
| **Bans** | Reference | 0-10k per guild | Low | N/A | API-only |
| **Scheduled Events** | Ephemeral | 0-20 per guild | Low | `scheduled_start_time` | API-only |
| **Voice States** | Ephemeral | Real-time only | Continuous via Gateway | N/A | No persistence |
| **Gateway info** | Ephemeral | 1 record | Never | N/A | No persistence |

## Social Signal Evidence

| Signal | Evidence Score | Source | What It Tells Us |
|--------|---------------|--------|-----------------|
| Local message search | 8/10 | discrawl (569 stars), jackwener/discord-cli (78 stars) | Messages MUST be in SQLite with FTS5 |
| Member directory sync | 7/10 | discrawl stores member snapshots, discord-analytics tracks members | Members need proper columns for join analysis |
| Audit log persistence | 6/10 | MEE6 audit logging, Discord's 45-day retention limit | Users want audit logs stored beyond 45 days |
| Channel metadata | 6/10 | AnswerOverflow indexes channels, analytics tools track per-channel | Channels need local storage for cross-referencing |
| Reaction tracking | 4/10 | Emoji usage analysis tools, community engagement metrics | Nice-to-have, not primary |

## Data Gravity Scores

| Entity | Volume | QueryFreq | JoinDemand | SearchNeed | TemporalValue | **Total** | Classification |
|--------|--------|-----------|------------|------------|---------------|-----------|---------------|
| **Messages** | 3 | 3 | 3 (channel, author, guild) | 3 (primary text content) | 2 (timestamps + trends) | **14** | PRIMARY |
| **Members** | 2 | 2 | 3 (messages, roles, audit) | 1 (username/nick) | 1 (joined_at) | **9** | PRIMARY |
| **Channels** | 1 | 2 | 3 (messages, threads, perms) | 1 (name, topic) | 0 | **7** | SUPPORT |
| **Audit Logs** | 2 | 2 | 2 (user, target) | 1 (reason text) | 3 (core time-series) | **10** | PRIMARY |
| **Roles** | 0 | 1 | 3 (members, audit, perms) | 0 | 0 | **4** | SUPPORT (via members) |
| **Threads** | 1 | 1 | 2 (messages, channels) | 1 (name) | 1 | **6** | SUPPORT |
| **Guilds** | 0 | 1 | 3 (everything) | 1 (name, desc) | 0 | **5** | SUPPORT |
| **Emojis** | 0 | 1 | 1 (reactions) | 0 | 0 | **2** | API-only |
| **Reactions** | 2 | 1 | 2 (messages, emojis) | 0 | 1 | **6** | SUPPORT |

**Primary entities (score >= 8):** Messages (14), Audit Logs (10), Members (9)
**Support entities (score 5-7):** Channels (7), Threads (6), Reactions (6), Guilds (5)

## SQLite Schema

```sql
-- Guild metadata
CREATE TABLE guilds (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    icon TEXT,
    owner_id TEXT,
    description TEXT,
    member_count INTEGER,
    features TEXT,  -- JSON array
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- Channels (includes threads - Discord stores threads as channels)
CREATE TABLE channels (
    id TEXT PRIMARY KEY,
    guild_id TEXT NOT NULL,
    parent_id TEXT,
    name TEXT NOT NULL,
    type INTEGER NOT NULL,  -- 0=text, 2=voice, 4=category, 5=announcement, 10/11/12=thread, 13=stage, 15=forum
    topic TEXT,
    position INTEGER,
    nsfw INTEGER DEFAULT 0,
    last_message_id TEXT,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);
CREATE INDEX idx_channels_guild ON channels(guild_id);
CREATE INDEX idx_channels_parent ON channels(parent_id);
CREATE INDEX idx_channels_type ON channels(type);

-- Members
CREATE TABLE members (
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    username TEXT NOT NULL,
    display_name TEXT,  -- global_name or nick
    nick TEXT,
    avatar TEXT,
    joined_at TEXT,
    roles TEXT,  -- JSON array of role IDs
    bot INTEGER DEFAULT 0,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    PRIMARY KEY (guild_id, user_id)
);
CREATE INDEX idx_members_guild ON members(guild_id);
CREATE INDEX idx_members_user ON members(user_id);
CREATE INDEX idx_members_joined ON members(joined_at);

-- Messages (PRIMARY entity - score 14)
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    guild_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    type INTEGER NOT NULL DEFAULT 0,
    timestamp TEXT NOT NULL,
    edited_timestamp TEXT,
    flags INTEGER DEFAULT 0,
    pinned INTEGER DEFAULT 0,
    mention_everyone INTEGER DEFAULT 0,
    attachment_count INTEGER DEFAULT 0,
    embed_count INTEGER DEFAULT 0,
    reference_message_id TEXT,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);
CREATE INDEX idx_messages_channel ON messages(channel_id);
CREATE INDEX idx_messages_guild ON messages(guild_id);
CREATE INDEX idx_messages_author ON messages(author_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);
CREATE INDEX idx_messages_channel_id ON messages(channel_id, id);  -- for cursor pagination

-- FTS5 for message full-text search
CREATE VIRTUAL TABLE messages_fts USING fts5(
    content,
    content=messages,
    content_rowid=rowid,
    tokenize='porter unicode61'
);

-- Triggers to keep FTS in sync
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

-- Audit Log Entries (PRIMARY entity - score 10)
CREATE TABLE audit_log_entries (
    id TEXT PRIMARY KEY,
    guild_id TEXT NOT NULL,
    user_id TEXT,           -- who performed the action
    target_id TEXT,         -- who/what was affected
    action_type INTEGER NOT NULL,
    reason TEXT,
    changes TEXT,           -- JSON array of changes
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);
CREATE INDEX idx_audit_guild ON audit_log_entries(guild_id);
CREATE INDEX idx_audit_user ON audit_log_entries(user_id);
CREATE INDEX idx_audit_target ON audit_log_entries(target_id);
CREATE INDEX idx_audit_action ON audit_log_entries(action_type);
CREATE INDEX idx_audit_guild_id ON audit_log_entries(guild_id, id);  -- cursor pagination

-- Roles (SUPPORT entity)
CREATE TABLE roles (
    id TEXT PRIMARY KEY,
    guild_id TEXT NOT NULL,
    name TEXT NOT NULL,
    color INTEGER DEFAULT 0,
    position INTEGER DEFAULT 0,
    permissions TEXT,
    mentionable INTEGER DEFAULT 0,
    hoist INTEGER DEFAULT 0,
    managed INTEGER DEFAULT 0,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);
CREATE INDEX idx_roles_guild ON roles(guild_id);

-- Reactions (SUPPORT entity)
CREATE TABLE reactions (
    message_id TEXT NOT NULL,
    emoji_name TEXT NOT NULL,
    emoji_id TEXT,
    count INTEGER DEFAULT 1,
    data JSON NOT NULL,
    PRIMARY KEY (message_id, emoji_name, emoji_id),
    FOREIGN KEY (message_id) REFERENCES messages(id)
);
CREATE INDEX idx_reactions_message ON reactions(message_id);

-- Mentions (structured, for query efficiency)
CREATE TABLE mentions (
    message_id TEXT NOT NULL,
    guild_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    mentioned_user_id TEXT,
    mentioned_role_id TEXT,
    mention_type TEXT NOT NULL,  -- 'user', 'role', 'everyone'
    FOREIGN KEY (message_id) REFERENCES messages(id)
);
CREATE INDEX idx_mentions_user ON mentions(mentioned_user_id);
CREATE INDEX idx_mentions_role ON mentions(mentioned_role_id);
CREATE INDEX idx_mentions_guild ON mentions(guild_id);

-- Sync state tracking
CREATE TABLE sync_state (
    guild_id TEXT NOT NULL,
    channel_id TEXT,
    entity_type TEXT NOT NULL,  -- 'messages', 'members', 'audit_logs', 'channels'
    last_id TEXT,               -- snowflake ID cursor
    last_synced_at TEXT NOT NULL,
    message_count INTEGER DEFAULT 0,
    PRIMARY KEY (guild_id, COALESCE(channel_id, ''), entity_type)
);
```

## Sync Strategy

### Messages (Primary - snowflake cursor)
- **Cursor field:** `id` (snowflake ID - encodes timestamp)
- **API endpoint:** `GET /channels/{channel_id}/messages?after={last_id}&limit=100`
- **VALIDATED:** Discord API supports `after` query param for messages (confirmed in spec)
- **Batch size:** 100 (API max)
- **Strategy:** For each channel, store `last_id` in `sync_state`. Resume from there on next sync.
- **Full sync:** Paginate backwards with `before` param from newest to oldest.

### Members (Primary - snowflake cursor)
- **Cursor field:** `user_id` (snowflake ID)
- **API endpoint:** `GET /guilds/{guild_id}/members?after={last_user_id}&limit=1000`
- **VALIDATED:** Discord API supports `after` query param for members (confirmed in spec)
- **Batch size:** 1000 (API max)
- **Strategy:** Full refresh on each sync (members don't have incremental endpoints)

### Audit Logs (Primary - snowflake cursor)
- **Cursor field:** `id` (snowflake ID)
- **API endpoint:** `GET /guilds/{guild_id}/audit-logs?after={last_id}&limit=100`
- **VALIDATED:** Discord API supports `after` AND `before` params for audit logs (confirmed in spec)
- **Batch size:** 100 (API max)
- **Strategy:** Incremental sync using `after` param. Store `last_id` per guild.
- **Note:** Discord retains audit logs for 45 days. Our local copy preserves them forever.

### Channels (Support - full refresh)
- **API endpoint:** `GET /guilds/{guild_id}/channels`
- **Strategy:** Full refresh (small cardinality, no pagination needed)

### Roles (Support - full refresh)
- **API endpoint:** `GET /guilds/{guild_id}/roles`
- **Strategy:** Full refresh (small cardinality)

## Domain-Specific Search Filters

| CLI Flag | SQL WHERE Clause | Validated |
|----------|-----------------|-----------|
| `--channel <name-or-id>` | `WHERE channel_id = ? OR channel_id IN (SELECT id FROM channels WHERE name LIKE ?)` | Yes |
| `--author <name-or-id>` | `WHERE author_id = ? OR author_id IN (SELECT user_id FROM members WHERE username LIKE ? OR nick LIKE ?)` | Yes |
| `--guild <name-or-id>` | `WHERE guild_id = ? OR guild_id IN (SELECT id FROM guilds WHERE name LIKE ?)` | Yes |
| `--days <N>` | `WHERE id > snowflake_from_timestamp(now - N days)` | Yes (snowflake encodes time) |
| `--since <date>` | `WHERE id > snowflake_from_timestamp(date)` | Yes |
| `--until <date>` | `WHERE id < snowflake_from_timestamp(date)` | Yes |
| `--pinned` | `WHERE pinned = 1` | Yes |
| `--has-attachment` | `WHERE attachment_count > 0` | Yes |
| `--has-embed` | `WHERE embed_count > 0` | Yes |
| `--type <msg-type>` | `WHERE type = ?` | Yes |
| `--action <audit-type>` | `WHERE action_type = ?` (for audit logs) | Yes |

## Compound Queries (Cross-Entity)

### 1. Messages by author in channel in last N days
```sql
SELECT m.id, m.content, m.timestamp, mb.username, c.name AS channel_name
FROM messages m
JOIN members mb ON m.author_id = mb.user_id AND m.guild_id = mb.guild_id
JOIN channels c ON m.channel_id = c.id
WHERE c.name = ? AND mb.username = ? AND m.id > ?
ORDER BY m.id DESC LIMIT 50;
```

### 2. Top authors by message count per channel
```sql
SELECT c.name AS channel, mb.username, COUNT(*) AS msg_count
FROM messages m
JOIN channels c ON m.channel_id = c.id
JOIN members mb ON m.author_id = mb.user_id AND m.guild_id = mb.guild_id
WHERE m.guild_id = ? AND m.id > ?
GROUP BY c.name, mb.username
ORDER BY msg_count DESC LIMIT 20;
```

### 3. Audit log actions by moderator with target names
```sql
SELECT a.id, a.action_type, mod.username AS moderator, target.username AS target, a.reason
FROM audit_log_entries a
LEFT JOIN members mod ON a.user_id = mod.user_id AND a.guild_id = mod.guild_id
LEFT JOIN members target ON a.target_id = target.user_id AND a.guild_id = target.guild_id
WHERE a.guild_id = ? AND a.action_type IN (22, 23, 24, 25)  -- ban/kick/member_update
ORDER BY a.id DESC LIMIT 50;
```

### 4. Stale channels (no messages in N days)
```sql
SELECT c.id, c.name, c.type, c.topic,
       MAX(m.id) AS last_message_snowflake,
       COUNT(m.id) AS total_messages
FROM channels c
LEFT JOIN messages m ON c.id = m.channel_id
WHERE c.guild_id = ? AND c.type IN (0, 5, 15)  -- text, announcement, forum
GROUP BY c.id
HAVING last_message_snowflake IS NULL OR last_message_snowflake < ?
ORDER BY total_messages ASC;
```

### 5. Member join timeline with role analysis
```sql
SELECT mb.username, mb.display_name, mb.joined_at, mb.bot,
       GROUP_CONCAT(r.name, ', ') AS roles
FROM members mb
LEFT JOIN roles r ON mb.guild_id = r.guild_id AND r.id IN (
    SELECT value FROM json_each(mb.roles)
)
WHERE mb.guild_id = ?
GROUP BY mb.user_id
ORDER BY mb.joined_at DESC LIMIT 50;
```

## Tail Strategy

| Method | Available? | Decision |
|--------|-----------|----------|
| **Gateway WebSocket** | YES - Discord Gateway is the primary real-time channel | **USE THIS** |
| SSE | No | N/A |
| REST Polling | Fallback only | Use for audit logs (no Gateway event for historical) |

**Decision: Gateway WebSocket** for real-time message/member/reaction events. Discord's Gateway provides:
- `MESSAGE_CREATE`, `MESSAGE_UPDATE`, `MESSAGE_DELETE` for messages
- `GUILD_MEMBER_ADD`, `GUILD_MEMBER_REMOVE`, `GUILD_MEMBER_UPDATE` for members
- `GUILD_AUDIT_LOG_ENTRY_CREATE` for audit logs (requires GuildModeration intent)
- `MESSAGE_REACTION_ADD`, `MESSAGE_REACTION_REMOVE` for reactions

The tail command should connect to Gateway, subscribe to relevant intents, and write events to SQLite in real-time.

**REST polling fallback:** For initial sync and bulk historical data, use REST endpoints with cursor pagination.

## Commands to Build in Phase 4 Priority 0

| Command | Purpose | Data Source |
|---------|---------|------------|
| `sync` | Incremental sync of messages, members, channels, audit logs, roles | REST API -> SQLite |
| `search` | Full-text search with domain filters | FTS5 + SQLite |
| `messages` | List/query messages from local DB | SQLite |
| `members` | List/query members from local DB | SQLite |
| `sql` | Raw read-only SQL queries | SQLite |
| `tail` | Real-time Gateway event stream | Gateway WebSocket -> SQLite |
| `audit` | Query audit log entries with filters | SQLite (after sync) |
| `status` | Show sync state, DB size, entity counts | sync_state table |
