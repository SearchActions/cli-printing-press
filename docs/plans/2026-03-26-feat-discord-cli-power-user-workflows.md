---
title: "Power User Workflows: Discord CLI"
type: feat
status: active
date: 2026-03-26
phase: "0.5"
api: "discord"
---

# Power User Workflows: Discord CLI

## API Archetype: Communication

Discord is a Communication platform. Power user workflows center on: archive, offline search, keyword monitoring, member analytics, audit trails, moderation actions, and compliance export.

## All 12 Workflow Ideas

### 1. `sync` - Incremental message archive (11/12)
Paginate messages per channel using snowflake `after` cursor, discover active + archived threads, sync thread messages, store in domain-specific SQLite tables. Endpoints: GET /guilds/{id}/channels, GET /channels/{id}/messages?after=&limit=100, GET /guilds/{id}/threads/active, GET /channels/{id}/threads/archived/public.

### 2. `search` - FTS5 local search (11/12)
Query FTS5 index with domain filters (--channel, --author, --before, --after). No API calls - pure local SQLite. Instant results vs Discord's rate-limited search.

### 3. `activity` - Member activity analytics (11/12)
GROUP BY author_id on local messages, COUNT per time window, JOIN with members for names/roles, rank by activity. No API calls after sync.

### 4. `tail` - Live Gateway event stream (10/12)
Connect to Gateway WebSocket, subscribe with intents, filter by event type/channel/keyword, output JSON stream. Endpoint: wss://gateway.discord.gg.

### 5. `audit` - Audit log investigation (10/12)
GET /guilds/{id}/audit-logs with action_type, user_id, before, after filters. Correlate with member/channel data. Timeline view.

### 6. `sql` - Raw SQL queries (10/12)
Accept read-only SQL, execute against local SQLite, output as table/JSON/CSV. Power user escape hatch.

### 7. `mentions` - Mention tracking (9/12)
Parse @user and @role mentions from local messages, group by entity, show frequency and context.

### 8. `channels` - Channel health (9/12)
List channels with local message counts, detect inactive channels, show last message date, sort by activity.

### 9. `members` - Member directory (9/12)
Offline member list from local DB with role filtering, join date sorting, search. Sync from GET /guilds/{id}/members?limit=1000&after=.

### 10. `stale` - Dead channel detection (9/12)
Find channels with no messages in N days. Cross-reference with channel type and permissions.

### 11. `roles` - Role hierarchy analysis (7/12)
Show role hierarchy with permissions, identify overprivileged roles, member counts per role.

### 12. `export` - Compliance export (6/12)
Filter local data by date range, output JSON/CSV. Deferred - DiscordChatExporter handles this well.

## Scoring Table

| Rank | Workflow | Frequency | Pain | Feasibility | Uniqueness | Total |
|------|----------|-----------|------|-------------|------------|-------|
| 1 | sync | 3 | 3 | 3 | 2 | 11 |
| 1 | search | 3 | 3 | 3 | 2 | 11 |
| 1 | activity | 2 | 3 | 3 | 3 | 11 |
| 4 | tail | 3 | 3 | 2 | 2 | 10 |
| 4 | audit | 1 | 3 | 3 | 3 | 10 |
| 4 | sql | 2 | 3 | 3 | 2 | 10 |
| 7 | mentions | 2 | 2 | 3 | 2 | 9 |
| 7 | channels | 1 | 2 | 3 | 3 | 9 |
| 7 | members | 2 | 2 | 3 | 2 | 9 |
| 7 | stale | 1 | 2 | 3 | 3 | 9 |
| 11 | roles | 1 | 1 | 3 | 2 | 7 |
| 12 | export | 1 | 2 | 3 | 0 | 6 |

## Top 7 for Implementation (Mandatory Phase 4 Work)

1. **`sync`** - Foundation. Without sync, nothing works.
2. **`search`** - Killer feature. FTS5 over synced messages.
3. **`activity`** - Analytics differentiator. Per-member rankings.
4. **`tail`** - Real-time hook. Gateway WebSocket stream.
5. **`audit`** - Moderation tool. Audit log with filtering.
6. **`sql`** - Power user escape hatch.
7. **`members`** - Offline member directory with role filtering.

## API Validation

All workflows validated against Discord API v10:
- Message pagination: `after` snowflake cursor confirmed on GET /channels/{id}/messages
- Audit log filtering: `action_type`, `user_id`, `before`, `after` params confirmed
- Member pagination: `after` snowflake cursor + `limit` (1-1000) confirmed
- Thread discovery: GET /guilds/{id}/threads/active + GET /channels/{id}/threads/archived/* confirmed
- Gateway: wss://gateway.discord.gg with intent-based subscriptions confirmed
