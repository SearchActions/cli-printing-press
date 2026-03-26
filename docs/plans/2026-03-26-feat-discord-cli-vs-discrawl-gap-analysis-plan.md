---
title: "discord-cli vs discrawl: Gap Analysis and Upgrade Plan"
type: feat
status: active
date: 2026-03-26
---

# discord-cli vs discrawl: Gap Analysis and Upgrade Plan

## Overview

discrawl (540 stars, steipete/discrawl) is a Go CLI that mirrors Discord guild data into local SQLite with FTS5 search. It has 12 commands but each one is deeply purpose-built for Discord. The generated discord-cli has 330+ commands but its data layer is generic. This plan compares them feature-by-feature and identifies what discord-cli needs to steal from discrawl's playbook.

**Bottom line:** discrawl is a far superior data tool. discord-cli is a far superior API wrapper. The ideal CLI combines both.

---

## Feature-by-Feature Comparison

### Data Layer

| Capability | discrawl | discord-cli | Winner |
|---|---|---|---|
| **Schema** | Discord-native: separate tables for messages, members, channels with proper columns | Generic `resources` table with JSON blobs | discrawl by a mile |
| **FTS5 indexing** | Extracts message content, attachment text, embed text, filenames | Indexes raw JSON blobs | discrawl |
| **Sync** | `--guild`, `--channels`, `--since`, `--concurrency` (auto-sized), `--with-embeddings` | Generic `--resources` flag, no Discord-specific filtering | discrawl |
| **Tail** | Gateway WebSocket with repair loops (real-time, misses nothing) | REST polling at intervals (slow, misses events, rate-limited) | discrawl |
| **Search** | `--guild`, `--channel`, `--author`, `--limit`, `--include-empty` | `--type`, `--limit` only (no Discord-specific filters) | discrawl |
| **Messages** | Filter by channel, author, time range (`--days`/`--hours`/`--since`), `--sync` on-demand, `--all`, `--last N` | No equivalent command | discrawl |
| **Mentions** | Structured user/role mention tracking with time filters | No equivalent | discrawl |
| **Members** | list, show (with message history), search (username, bio, pronouns, social handles, URLs) | member-report (role counts only) | discrawl |
| **Channels** | list, show (metadata inspection) | Via API commands only (no local query) | discrawl |
| **SQL** | Raw read-only SQL access to the database | No equivalent | discrawl |
| **Status** | Archive statistics and sync progress | No equivalent | discrawl |
| **Init** | Multi-guild discovery with OpenClaw integration, config file generation | No equivalent | discrawl |
| **Embeddings** | Optional OpenAI semantic search, batch-processed in background | No equivalent | discrawl |

**Score: discrawl 13/13 on data capabilities.**

### API Management

| Capability | discrawl | discord-cli | Winner |
|---|---|---|---|
| **REST API coverage** | 0 commands | 316 commands across 20 resource groups | discord-cli |
| **Guild management** | None (read-only crawler) | Create, update, delete guilds | discord-cli |
| **Channel management** | None | Full CRUD + permissions | discord-cli |
| **Role management** | None | Full CRUD + bulk operations | discord-cli |
| **Ban management** | None | Ban, unban, list bans | discord-cli |
| **Webhook management** | None | Full CRUD + execute | discord-cli |
| **Application commands** | None | Full slash command management | discord-cli |
| **Auto moderation** | None | Full rule management | discord-cli |
| **Scheduled events** | None | Full CRUD | discord-cli |
| **Stage instances** | None | Full management | discord-cli |

**Score: discord-cli 10/10 on API management. discrawl has zero.**

### Workflow Commands

| Capability | discrawl | discord-cli | Winner |
|---|---|---|---|
| **channel-health** | Nothing | Stale channel detection + activity report | discord-cli |
| **audit-report** | Nothing | Audit log analysis by action/user/date | discord-cli |
| **member-report** | Nothing (members search is close) | Role distribution, bot/human counts | discord-cli |
| **server-snapshot** | Nothing | Backup guild config to JSON | discord-cli |
| **message-stats** | Nothing | Message volume, top contributors, hourly activity | discord-cli |
| **webhook-test** | Nothing | Send test payloads to webhooks | discord-cli |
| **prune-preview** | Nothing | Preview prune without executing | discord-cli |

**Score: discord-cli 7/0. discrawl has no workflow commands.**

### Agent-Native Features

| Feature | discrawl | discord-cli | Winner |
|---|---|---|---|
| **--json** | Yes | Yes | Tie |
| **--select** | No | Yes | discord-cli |
| **--dry-run** | No | Yes | discord-cli |
| **--stdin** | No | Yes | discord-cli |
| **--yes** | No | Yes | discord-cli |
| **--no-cache** | No | Yes | discord-cli |
| **--csv** | No | Yes | discord-cli |
| **--plain** | No | Yes | discord-cli |
| **--quiet** | No | Yes | discord-cli |
| **Typed exit codes** | Unknown | 0,2,3,4,5,7,10 | discord-cli |

**Score: discord-cli 8/1.**

---

## The Verdict

| Dimension | discrawl | discord-cli |
|---|---|---|
| Data tool | 13/13 | 0/13 |
| API management | 0/10 | 10/10 |
| Workflow commands | 0/7 | 7/7 |
| Agent-native | 1/9 | 9/9 |
| **Total** | **14/39** | **26/39** |

discord-cli wins on breadth. discrawl wins on depth where it matters most - the data layer.

The printing-press skill itself says it: **"discrawl has 12 commands and 539 stars. Depth beats breadth."** The 316 API wrapper commands are table stakes. The 7 workflow commands are differentiation. But the data layer - sync, search, tail, messages, mentions, members, SQL - is where discrawl demolishes us.

---

## What discord-cli Must Steal from discrawl

### Priority 1: Discord-Native Schema (Critical)

Replace the generic `resources` table with Discord-specific tables:

```sql
-- Messages table with proper columns
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    guild_id TEXT,
    author_id TEXT NOT NULL,
    content TEXT,
    timestamp DATETIME NOT NULL,
    edited_timestamp DATETIME,
    type INTEGER DEFAULT 0,
    data JSON NOT NULL,
    synced_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_messages_channel ON messages(channel_id);
CREATE INDEX idx_messages_author ON messages(author_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);

-- FTS on extracted content (not JSON blobs)
CREATE VIRTUAL TABLE messages_fts USING fts5(
    id, channel_id, author_id, content,
    tokenize='porter unicode61'
);

-- Members table
CREATE TABLE members (
    user_id TEXT NOT NULL,
    guild_id TEXT NOT NULL,
    username TEXT,
    display_name TEXT,
    bio TEXT,
    joined_at DATETIME,
    roles JSON,
    data JSON NOT NULL,
    PRIMARY KEY (user_id, guild_id)
);

-- Mentions table
CREATE TABLE mentions (
    message_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    target_type TEXT NOT NULL, -- 'user' or 'role'
    guild_id TEXT,
    channel_id TEXT,
    timestamp DATETIME,
    FOREIGN KEY (message_id) REFERENCES messages(id)
);
```

**Why:** Without this, search returns JSON blobs. With this, search returns `#channel-name @author: message content` with filters.

### Priority 2: Discord-Aware Sync

Replace the generic sync with Discord-specific backfill:

- `discord-cli sync --guild GUILD_ID` - sync one guild
- `discord-cli sync --channels CHANNEL_ID,CHANNEL_ID` - sync specific channels
- `discord-cli sync --since 2026-01-01` - sync from date
- `discord-cli sync --full` - full historical backfill
- `discord-cli sync --concurrency 16` - auto-sized by default
- Incremental: track last message ID per channel as cursor
- Paginate with `GET /channels/{id}/messages?before=OLDEST_ID&limit=100`

### Priority 3: Messages Command

Add a `messages` command that queries the local database:

```bash
# Messages from a channel in the last 7 days
discord-cli messages --channel general --days 7

# Messages from a specific author
discord-cli messages --author steipete --days 30

# All messages, syncing on-demand if needed
discord-cli messages --channel help --all --sync

# Last 100 messages
discord-cli messages --channel dev --last 100 --json
```

### Priority 4: Search Filtering

Add Discord-specific filters to search:

```bash
# Search within a guild
discord-cli search "error" --guild 123456789012345678

# Search within a channel
discord-cli search "timeout" --channel help

# Search by author
discord-cli search "fix" --author steipete

# Combine filters
discord-cli search "deploy" --guild 123 --channel releases --json
```

### Priority 5: SQL Command

Expose raw read-only SQL access:

```bash
# Run a query
discord-cli sql 'SELECT channel_id, COUNT(*) as cnt FROM messages GROUP BY channel_id ORDER BY cnt DESC LIMIT 10'

# From stdin
echo 'SELECT * FROM members WHERE bio LIKE "%rust%"' | discord-cli sql -
```

### Priority 6: Mentions Command

Track and query structured mentions:

```bash
# Who mentioned me?
discord-cli mentions --target @mybot --days 7

# Role mentions in a channel
discord-cli mentions --channel announcements --type role

# JSON for agent consumption
discord-cli mentions --target USER_ID --json
```

### Priority 7: Gateway Tail (Future)

Replace REST polling with Gateway WebSocket:

- Connect to `wss://gateway.discord.gg/` with bot token
- Subscribe to MESSAGE_CREATE, MESSAGE_UPDATE, MESSAGE_DELETE events
- Write events to SQLite in real-time
- Periodic repair syncs to catch missed events
- Graceful reconnection with resume

This is the hardest change (Gateway protocol is complex) but gives true real-time vs polling delay.

---

## What discord-cli Already Does Better

These are advantages to preserve and highlight:

1. **316 API wrapper commands** - discrawl can't create channels, manage roles, ban users, or execute webhooks
2. **7 workflow commands** - channel-health, audit-report, server-snapshot are unique value
3. **Agent-native features** - --select, --dry-run, --stdin, --yes, --csv, --plain, typed exits
4. **Complex body support** - --stdin for embeds, components, attachments
5. **Doctor with full diagnostics** - Both have this, parity

---

## Implementation Phases

### Phase 1: Schema Migration (2-3 hours)
- Replace generic `resources` table with Discord-specific tables
- Update `store.go` with `UpsertMessage`, `UpsertMember`, `UpsertChannel` methods
- Migrate FTS5 to index extracted content, not JSON blobs
- Add mention extraction during message upsert

### Phase 2: Discord-Aware Sync (3-4 hours)
- Rewrite `sync.go` with Discord-specific logic
- Paginate through channels using `?before=` cursor
- Track per-channel sync state
- Add `--guild`, `--channels`, `--since`, `--concurrency` flags
- Auto-discover guilds from bot token

### Phase 3: Messages + Search Filtering (2-3 hours)
- New `messages` command with local database queries
- Add `--guild`, `--channel`, `--author` filters to search
- Add `--days`, `--hours`, `--since` time filters

### Phase 4: SQL + Mentions (1-2 hours)
- Read-only SQL command
- Mentions command with structured queries

### Phase 5: Gateway Tail (4-6 hours, optional)
- WebSocket client for Discord Gateway
- Event-to-SQLite pipeline
- Repair loops
- Resume on reconnect

**Total estimated work: 8-12 hours for Phases 1-4, plus 4-6 hours for Gateway.**

---

## Acceptance Criteria

- [ ] Discord-specific schema with messages, members, channels, mentions tables
- [ ] FTS5 indexes extracted content (not JSON blobs)
- [ ] Sync supports `--guild`, `--channels`, `--since`, `--concurrency`
- [ ] `messages` command with channel/author/time filters
- [ ] `search` supports `--guild`, `--channel`, `--author` filters
- [ ] `sql` command for raw read-only queries
- [ ] `mentions` command for structured mention queries
- [ ] All existing 316 API commands still work
- [ ] All 7 workflow commands still work
- [ ] `go build ./...` and `go vet ./...` pass

---

## Sources

- [steipete/discrawl](https://github.com/steipete/discrawl) - 540 stars, the reference implementation
- [jackwener/discord-cli](https://github.com/jackwener/discord-cli) - 78 stars, Python alternative with similar approach
- [Tyrrrz/DiscordChatExporter](https://github.com/Tyrrrz/DiscordChatExporter) - 10.7k stars, export-only (no search/sync)
- [Discord API Docs](https://docs.discord.com/developers/intro) - Official REST + Gateway documentation
