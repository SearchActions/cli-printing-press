---
title: "Visionary Research: Discord CLI"
type: feat
status: active
date: 2026-03-25
---

## Visionary Research: Discord CLI

### API Identity
- Domain: Communication (messaging, voice, communities)
- Primary users: Bot developers, server admins, community managers, data analysts
- Core entities: Guilds, Channels, Messages, Users, Roles, Emojis, Webhooks, Invites, Audit Log, Auto Moderation, Stage Instances, Scheduled Events, Polls, Stickers, Soundboard, SKUs, Entitlements, Subscriptions, Guild Templates, Lobbies
- Data profile:
  - Write pattern: Append-only (messages), mutable (guilds, channels, roles, members)
  - Volume: HIGH (millions of messages across servers)
  - Real-time: YES (Gateway/WebSocket for events, webhooks for notifications)
  - Search need: HIGH (users constantly need to find messages, members, audit events)
- Auth: Bot token (primary), OAuth2 (user auth)
- Base URL: https://discord.com/api/v10
- Rate limits: Per-route + global (10,000 req/10min = 429 ban for 1hr+)

### Usage Patterns (Top 5 by Evidence)

1. **Message export/archive** (Evidence: 10/10) - DiscordChatExporter 10.7k stars, multiple archive tools
2. **Server backup/restore** (Evidence: 7/10) - discord-backup, discopy, discord-server-backup
3. **Message search/analytics** (Evidence: 7/10) - AnswerOverflow, discord-analytics, discord-cli FTS5
4. **Real-time channel monitoring** (Evidence: 5/10) - discord-dl, discord-cli tail, keyword alerting
5. **Server management automation** (Evidence: 5/10) - Bot management, role CRUD, audit log review

### Tool Landscape (Beyond API Wrappers)

| Tool | Stars | Type | What it does |
|------|-------|------|-------------|
| DiscordChatExporter | 10,700 | Data Tool | Export messages to HTML/JSON/CSV/TXT |
| discordo | 5,400 | Environment Tool | TUI Discord client |
| jackwener/discord-cli | 78 | Data Tool | SQLite sync + FTS5 search + tail + export |
| discord-dl | 26 | Data Tool | Archive to SQLite with web UI |
| discord-migrate | 0 | Data Tool | Export to SQLite for Matrix migration |
| Copycord | N/A | Integration Tool | Real-time server mirroring |
| discopy | N/A | Data Tool | Server config backup/restore |

### Workflows

1. **Archive and Search**: sync messages -> SQLite -> FTS5 search -> export
2. **Server Audit**: fetch audit log -> filter by action/user/date -> export report
3. **Channel Health**: list channels -> message volume -> stale detection -> report
4. **Member Analytics**: list members -> role grouping -> activity analysis -> export
5. **Bulk Role Management**: list roles -> compare template -> bulk assign -> dry-run

### Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Persistence | SQLite + FTS5 | High volume + high search need. Validated by discord-cli. |
| Real-time | REST polling (tail) | Simpler than Gateway WebSocket for CLI use. |
| Search | FTS5 local | Offline search, zero rate limit exposure. |
| Bulk | Pagination + rate-limit | Transparent handling of Discord's per-route limits. |
| Cache | SQLite IS the cache | Sync once, query many. --no-cache bypasses. |

### Top 5 Features for the World

| # | Feature | Score | Description |
|---|---------|-------|-------------|
| 1 | Local sync + FTS5 search | 14/16 | Sync to SQLite, search offline. |
| 2 | Channel health detection | 12/16 | Inactive channels, volume trends. |
| 3 | Audit log analysis + export | 12/16 | Query by action/user/date, export JSON. |
| 4 | Server config backup/restore | 11/16 | Roles, channels, permissions to JSON. |
| 5 | Message tail (real-time follow) | 10/16 | tail -f for Discord channels. |
