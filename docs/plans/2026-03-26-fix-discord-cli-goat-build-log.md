---
title: "GOAT Build Log: Discord CLI"
type: fix
status: active
date: 2026-03-26
phase: "4"
api: "discord"
---

# GOAT Build Log: Discord CLI

## Data Layer Implementation

### Tables Created
| Table | Columns | FTS5 | Indexes |
|-------|---------|------|---------|
| messages | id, channel_id, guild_id, author_id, content, timestamp, edited_at, type, pinned, data | Yes (content) | channel, author, guild, timestamp, channel+ts |
| members | guild_id, user_id, username, display_name, nickname, joined_at, roles, data | Yes (username, display_name, nickname) | username, joined |
| channels | id, guild_id, name, type, parent_id, position, topic, last_message_id, data | No | guild |
| audit_log | id, guild_id, user_id, target_id, action_type, reason, data | No | guild, user, action, target |
| sync_state | guild_id, channel_id, last_message_id, last_synced, message_count | No | (composite PK) |

### Store Methods
- UpsertMessage, UpsertMember, UpsertChannel, UpsertAuditEntry
- UpdateSyncState, GetSyncState
- SearchMessages (FTS5 with guild/channel/author filters)
- QuerySQL (read-only enforcement)
- GetActivity (message counts per author with time window)
- GetStaleChannels (channels with no recent messages)
- Status, SyncStates

## Workflow Commands Built

| Command | File | Purpose | DB Tables Used |
|---------|------|---------|---------------|
| sync | sync.go | Incremental message/member/channel sync | messages, members, channels, sync_state |
| search | search.go | FTS5 search with filters | messages, messages_fts |
| activity | activity.go | Per-member message rankings | messages, members |
| audit | audit.go | Audit log with action filtering | audit_log |
| stale | stale.go | Dead channel detection | channels, messages |
| sql | sqlcmd.go | Raw read-only SQL | All tables |
| status | status.go | Archive health dashboard | messages, members, channels, sync_state |
| health | health.go | Guild health metrics | messages, channels, members |
| similar | similar.go | FTS5 similar message finder | messages, messages_fts |
| bottleneck | bottleneck.go | Channel bottleneck detection | messages |
| trends | trends.go | Daily message volume trends | messages |
| patterns | patterns.go | Activity pattern analysis | messages |
| forecast | forecast.go | Activity forecast (linear regression) | messages |

## Scorecard Progression

| Dimension | Before | After | Delta |
|-----------|--------|-------|-------|
| README | 3 | 9 | +6 |
| Insight | 0 | 10 | +10 |
| Workflows | 4 | 6 | +2 |
| Sync Correctness | 4 | 7 | +3 |
| Dead Code | 0 | 5 | +5 |
| Vision | 9 | 8 | -1 |
| **TOTAL** | **63** | **78** | **+15** |

## Data Pipeline Trace

| Entity | WRITE path | READ path | SEARCH path |
|--------|-----------|-----------|-------------|
| Messages | sync.go:142 UpsertMessage | activity.go, trends.go, patterns.go, health.go, bottleneck.go | search.go:54 SearchMessages, similar.go:72 SearchMessages |
| Members | sync.go:200 UpsertMember | activity.go GetActivity (joins) | N/A |
| Channels | sync.go:86 UpsertChannel | stale.go GetStaleChannels, health.go | N/A |
| Audit Log | audit.go:94 UpsertAuditEntry | audit.go (displays results) | N/A |

## What Was Skipped
- Gateway tail command (needs WebSocket - deferred)
- Export command rewrite (existing DiscordChatExporter is better)
- Members FTS5 search command (activity covers member queries)
