---
title: "Visionary Research: Discord CLI"
type: feat
status: active
date: 2026-03-26
phase: "0"
api: "discord"
---

# Visionary Research: Discord CLI

## Overview

Discord's API serves a massive developer ecosystem of bot builders, community managers, and platform integrators. The API is REST-based (v10) with a real-time Gateway (WebSocket) for live events, covering guilds (servers), channels, messages, members, roles, emojis, interactions, webhooks, and more. The official OpenAPI spec (3.1.0) is publicly available at discord/discord-api-spec on GitHub.

The CLI landscape for Discord is fragmented: DiscordChatExporter (10.7k stars) dominates export but offers no local persistence or search. discrawl (564 stars) pioneered SQLite+FTS5+Gateway sync for Discord in Go with 12 commands. discordo (5.4k stars) is a TUI client, not a management CLI. discli (6 stars) targets agent-native workflows but has no data layer. jackwener/discord-cli (78 stars) combines SQLite sync with search but is Python-only and uses user tokens (TOS risk). No single tool combines bot-token REST management + local SQLite data layer + FTS5 search + Gateway tail + agent-native output modes in Go.

The printing-press Discord CLI should be a discrawl-class data tool with full API coverage - not just an API wrapper, but a local-first intelligence layer over Discord's data.

## API Identity

- **Domain:** Communication (messaging, channels, threads, voice, communities)
- **Primary users:** Bot developers, server administrators, community managers, AI agent builders, moderation teams
- **Core entities:** Guilds, Channels (text/voice/forum/stage), Messages, Users, Members, Roles, Emojis, Stickers, Webhooks, Audit Logs, Threads, Scheduled Events, Invites, Auto Moderation Rules, Polls, Soundboards, Stage Instances
- **API base URL:** `https://discord.com/api/v10`
- **Auth:** Bot Token (primary, via `DISCORD_BOT_TOKEN` env var), OAuth2 (for user-facing apps)

### Data Profile

| Dimension | Assessment |
|---|---|
| **Write pattern** | Append-only for messages/events/audit logs, mutable for guilds/channels/roles/members |
| **Volume** | Very high - large guilds have millions of messages, thousands of members |
| **Real-time** | Yes - Gateway WebSocket with intents, event-driven architecture |
| **Search need** | Very high - finding messages, members, mentions, content across channels is a core workflow |

## Usage Patterns (Top 5 by Evidence)

| Rank | Pattern | Evidence Score | Key Sources |
|------|---------|---------------|-------------|
| 1 | **Message archive & search** | 10/10 | DiscordChatExporter (10.7k stars), discrawl (564), jackwener/discord-cli (78), discord-dl, discord-server-backup, multiple Reddit/HN threads |
| 2 | **Local search over history** | 8/10 | discrawl FTS5 search, AnswerOverflow (Discord indexing), Comly.app (HN post about Discord SEO), jackwener/discord-cli search |
| 3 | **Server management** | 7/10 | discli (agent-native mgmt), discord-server-mirror, Copycord, backup/restore tools |
| 4 | **Analytics & activity** | 6/10 | discord-analytics (Tarasa24), Cially dashboard, james-long/discord-analytics, discord-bot-analytics |
| 5 | **Live monitoring & alerting** | 6/10 | discord-user-monitor, web-watcher, voice-monitor, quest-watcher, discrawl tail |

## Tool Landscape (Beyond API Wrappers)

| Tool | Stars | Type | What It Does | Lang |
|------|-------|------|-------------|------|
| DiscordChatExporter | 10,700 | Data Tool | Export messages to HTML/TXT/CSV/JSON with rich media. 63 contributors. Active (v2.47.1, 2026-03-21). | C# |
| discordo | 5,400 | Environment Tool | Full TUI Discord client. 45 open issues. Go. | Go |
| discrawl | 564 | Data Tool | SQLite+FTS5 sync/search/tail/members/sql. 12 commands. 4 open issues. THE benchmark. | Go |
| AnswerOverflow | ~2,000 | Integration Tool | Makes Discord threads indexable on Google/AI | TS |
| Copycord | ~200 | Workflow Tool | Clone/mirror entire servers in real-time | Python |
| jackwener/discord-cli | 78 | Data Tool | SQLite-based sync/search/export with AI analysis | Python |
| discli | 6 | API Wrapper | Agent-native server mgmt (YAML output, --dry-run) | TS |
| discord-migrate | ~50 | Data Tool | Discord to SQLite for Matrix migration | Python |

**Key insight:** The highest-starred tools are Data Tools (DiscordChatExporter, discrawl), not API wrappers. Users want to OWN their Discord data locally. discrawl is the Steinberger benchmark - 12 focused commands, SQLite+FTS5, Gateway sync, and a `sql` command for ad-hoc queries.

## Workflows

| # | Name | Steps | Pain Point | Proposed CLI Command |
|---|------|-------|-----------|---------------------|
| 1 | Archive & Search | Sync messages -> store in SQLite -> FTS5 search with filters | Manual pagination, rate limits, thread discovery | `discord-cli sync --guild <id>` + `discord-cli search "keyword" --channel general --author user` |
| 2 | Member Activity | Fetch members -> query message counts -> rank by activity -> filter by role | No single API call gives activity; requires message+member correlation | `discord-cli activity --guild <id> --days 30 --role moderator` |
| 3 | Audit Investigation | Fetch audit log -> filter by action type -> show timeline | Raw events, no correlation or timeline view | `discord-cli audit --guild <id> --action member_kick --days 7` |
| 4 | Keyword Monitoring | Connect Gateway -> filter message_create -> match keywords -> alert | Requires writing a bot; no CLI does real-time keyword monitoring | `discord-cli tail --guild <id> --match "urgent|bug|down"` |
| 5 | Compliance Export | Sync all channels -> export JSON/CSV -> date range filter | DiscordChatExporter needs manual channel selection | `discord-cli export --guild <id> --format json --since 2026-01-01` |

## Architecture Decisions

| Area | Need | Decision | Rationale |
|------|------|----------|-----------|
| **Persistence** | HIGH | SQLite with domain-specific tables + FTS5 | Millions of messages, high search need. discrawl proves this. Proper columns enable joins (messages x members x channels). |
| **Real-time** | HIGH | Gateway WebSocket for `tail` command | Discord Gateway is canonical. REST polling misses events and burns rate limit. discrawl uses Gateway. |
| **Search** | HIGH | FTS5 on message content, channel names, member display names | Local FTS5 is instant vs Discord's rate-limited API search. Domain filters (--channel, --author) map to SQL WHERE. |
| **Bulk** | MEDIUM | Incremental sync with snowflake ID cursors | Discord message IDs are snowflakes (timestamp-encoded). GET /channels/{id}/messages supports `after` param. |
| **Cache** | MEDIUM | SQLite IS the cache. `--no-cache` bypasses for live API calls. `--sync` flag triggers API fetch before query. | No separate cache layer needed. |

## Top 5 Features for the World

| Rank | Feature | Score | Evidence | Impact | Feasibility | Uniqueness | Composability | Data Fit | Maintain | Moat |
|------|---------|-------|----------|--------|-------------|------------|---------------|----------|----------|------|
| 1 | SQLite sync + FTS5 search | **15/16** | 3 | 3 | 2 | 1 | 2 | 2 | 1 | 1 |
| 2 | Member activity analytics | **13/16** | 2 | 3 | 2 | 2 | 2 | 2 | 0 | 0 |
| 3 | Audit log investigation | **13/16** | 2 | 2 | 2 | 2 | 2 | 2 | 0 | 1 |
| 4 | Gateway tail (live stream) | **12/16** | 3 | 2 | 1 | 1 | 2 | 2 | 1 | 0 |
| 5 | Agent-native server mgmt | **12/16** | 2 | 2 | 2 | 1 | 2 | 1 | 1 | 1 |

All 5 score >=12: **Must-have.**

## Sources

- [DiscordChatExporter](https://github.com/Tyrrrz/DiscordChatExporter) - 10.7k stars, C#
- [discordo](https://github.com/ayn2op/discordo) - 5.4k stars, Go TUI
- [discrawl](https://github.com/steipete/discrawl) - 564 stars, Go, SQLite+FTS5 (Steinberger benchmark)
- [jackwener/discord-cli](https://github.com/jackwener/discord-cli) - 78 stars, Python
- [discli](https://github.com/ibbybuilds/discli) - 6 stars, TS, agent-native
- [AnswerOverflow](https://github.com/AnswerOverflow/AnswerOverflow) - Discord indexing
- [Discord API Docs](https://docs.discord.com/developers/intro)
- [Discord OpenAPI Spec](https://github.com/discord/discord-api-spec)
- [HN: Discord indexing tool](https://news.ycombinator.com/item?id=46898145)
- [HN: discordo](https://news.ycombinator.com/item?id=32474133)
- [HN: Discoding - AI CLIs via Discord](https://news.ycombinator.com/item?id=47048164)
