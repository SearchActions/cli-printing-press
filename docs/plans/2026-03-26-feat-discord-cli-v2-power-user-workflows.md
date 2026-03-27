---
title: "Power User Workflows: Discord CLI v2"
type: feat
status: active
date: 2026-03-26
phase: "0.5"
api: "discord"
---

# Power User Workflows: Discord CLI v2

## API Archetype: Communication

Discord is a **Communication** archetype with strong **Project Management** crossover (communities manage tasks, moderate, and track engagement like a PM tool). Primary workflows: archive, offline search, monitor keywords, export conversations, moderation analysis, community health.

## Workflow Ideas (13 total)

### 1. Server Health Report
**Steps:** Query local messages table for last N days -> Count messages per channel -> Count unique active members -> Calculate message velocity -> Identify stale channels (0 messages) -> Report
**Frequency:** Weekly (3)
**Pain:** Manual counting in Discord UI is impossible at scale (3)
**Feasibility:** Easy - all data in SQLite after sync (3)
**Uniqueness:** No CLI does this (3)
**Score: 12/12**

### 2. Audit Log Forensics
**Steps:** Fetch audit log entries -> Store in local DB -> Filter by action type (ban, kick, delete, role change) -> Cross-reference with member table -> Timeline report
**Frequency:** After incidents + weekly (3)
**Pain:** Discord UI shows raw audit log, no filtering or export (3)
**Feasibility:** API supports action_type filter, 45-day retention (3)
**Uniqueness:** No CLI does this (3)
**Score: 12/12**

### 3. Stale Channel Detection
**Steps:** Query local messages table -> Find channels with 0 messages in N days -> Cross-reference with channel metadata (topic, category) -> Report with last-message timestamps
**Frequency:** Monthly (1)
**Pain:** Manual check across 50+ channels is tedious (2)
**Feasibility:** Easy - JOIN channels and messages tables (3)
**Uniqueness:** No CLI does this (3)
**Score: 9/12**

### 4. Member Activity Leaderboard
**Steps:** Query local messages -> GROUP BY author_id -> COUNT messages -> JOIN with members table for display names -> Rank -> Report
**Frequency:** Weekly (2)
**Pain:** No built-in way to see who's most active (2)
**Feasibility:** Easy with local DB (3)
**Uniqueness:** Cially dashboard does this but no CLI (2)
**Score: 9/12**

### 5. Cross-Channel Keyword Monitor (tail + filter)
**Steps:** Connect to Gateway -> Filter MESSAGE_CREATE events -> Match against keyword list -> Alert (print to stdout, or format for pipe to webhook)
**Frequency:** Continuous (3)
**Pain:** Requires a full bot framework just to watch for keywords (3)
**Feasibility:** Gateway client needed, moderate effort (2)
**Uniqueness:** No CLI does real-time keyword filtering (3)
**Score: 11/12**

### 6. Moderation Summary Report
**Steps:** Query audit log entries for bans, kicks, timeouts, message deletes -> Group by moderator -> Group by day -> Trend analysis -> Report
**Frequency:** Weekly (2)
**Pain:** No way to see moderation trends over time (3)
**Feasibility:** Requires audit log in local DB (2)
**Uniqueness:** No CLI does this (3)
**Score: 10/12**

### 7. Export Pipeline (SQLite -> formats)
**Steps:** Query local messages table with filters (--channel, --author, --since, --until) -> Format as JSON/CSV/HTML -> Write to file
**Frequency:** Monthly (1)
**Pain:** DiscordChatExporter requires .NET, no incremental, no local search first (2)
**Feasibility:** Easy from SQLite (3)
**Uniqueness:** Export from LOCAL data is unique - DiscordChatExporter hits API every time (2)
**Score: 8/12**

### 8. Mention Analysis
**Steps:** Query structured mentions in local DB -> Group by mentioned user/role -> Count frequency -> Identify most-mentioned members/roles -> Cross-reference with message context
**Frequency:** Monthly (1)
**Pain:** Discord shows mentions but can't aggregate them (2)
**Feasibility:** discrawl already stores structured mentions (3)
**Uniqueness:** Partial - discrawl has mentions command (1)
**Score: 7/12**

### 9. Thread Cleanup / Stale Thread Detection
**Steps:** List all threads -> Check last_message_id timestamp -> Find threads with no activity in N days -> Optionally archive them
**Frequency:** Monthly (1)
**Pain:** Threads proliferate and become noise (2)
**Feasibility:** API supports thread listing and archival (3)
**Uniqueness:** No tool does this (3)
**Score: 9/12**

### 10. Role Audit
**Steps:** List all roles -> For each role, list members with that role -> Identify roles with 0 members -> Identify members with excessive roles -> Report
**Frequency:** Monthly (1)
**Pain:** Discord UI makes it hard to see role <-> member mappings at scale (2)
**Feasibility:** Easy with members + roles in local DB (3)
**Uniqueness:** No CLI does this (3)
**Score: 9/12**

### 11. Invite Tracking
**Steps:** List invites -> Track usage counts -> Identify top inviters -> Detect expired/unused invites -> Clean up
**Frequency:** Weekly (2)
**Pain:** Discord shows invites but no analysis (2)
**Feasibility:** API supports invite listing with use counts (3)
**Uniqueness:** Partial - some bots do this (1)
**Score: 8/12**

### 12. Emoji Usage Analysis
**Steps:** Parse message reactions from local DB -> Count usage per emoji -> Identify unused custom emojis -> Report most/least popular
**Frequency:** Monthly (1)
**Pain:** Custom emoji slots are limited, want to prune unused ones (2)
**Feasibility:** Needs reaction data in local DB (2)
**Uniqueness:** No CLI does this (3)
**Score: 8/12**

### 13. Permission Diff
**Steps:** Snapshot current permissions for all channels/roles -> Compare to previous snapshot -> Report changes
**Frequency:** After incidents (2)
**Pain:** Permission overwrites are notoriously complex in Discord (3)
**Feasibility:** API supports permission overwrites per channel (2)
**Uniqueness:** No tool does this (3)
**Score: 10/12**

## Validation Against API

| Workflow | Required Endpoints | Validated? |
|----------|-------------------|-----------|
| Server Health | Local DB only (after sync) | Yes - uses synced messages table |
| Audit Forensics | GET /guilds/{id}/audit-logs | Yes - supports action_type, user_id, before filters |
| Stale Channels | Local DB only | Yes - channels + messages tables |
| Member Leaderboard | Local DB only | Yes - messages GROUP BY author_id |
| Keyword Monitor | Gateway MESSAGE_CREATE | Yes - Gateway intent required |
| Moderation Summary | Local audit_log_entries table | Yes - requires audit log sync |
| Export Pipeline | Local DB only | Yes - SELECT with filters |
| Mention Analysis | Local mentions table | Yes - discrawl stores these |
| Thread Cleanup | GET /guilds/{id}/threads | Yes - includes archived threads |
| Role Audit | Local DB (members + roles) | Yes - after member/role sync |
| Invite Tracking | GET /guilds/{id}/invites | Yes - includes use counts |
| Emoji Analysis | Local DB (reactions) | Partial - need to sync reactions |
| Permission Diff | GET /channels/{id} (permission_overwrites) | Yes - per-channel overwrites |

## Top 7 for Implementation (Phase 4 Priority 1)

| Rank | Workflow | Score | CLI Command |
|------|----------|-------|-------------|
| 1 | **Server Health Report** | 12/12 | `discord-cli health --guild <id> --days 30` |
| 2 | **Audit Log Forensics** | 12/12 | `discord-cli audit --guild <id> --action ban --days 7 --json` |
| 3 | **Keyword Monitor** | 11/12 | `discord-cli watch --guild <id> --keywords "error,outage,bug"` |
| 4 | **Moderation Summary** | 10/12 | `discord-cli modreport --guild <id> --days 30` |
| 5 | **Permission Diff** | 10/12 | `discord-cli permdiff --guild <id> --snapshot` |
| 6 | **Stale Channel Detection** | 9/12 | `discord-cli stale --guild <id> --days 30 --type channels` |
| 7 | **Role Audit** | 9/12 | `discord-cli roleaudit --guild <id> --empty --excessive` |

These 7 become **mandatory Phase 4 work items.** They use the local SQLite database where possible, falling back to live API only when needed (audit logs, permissions).
