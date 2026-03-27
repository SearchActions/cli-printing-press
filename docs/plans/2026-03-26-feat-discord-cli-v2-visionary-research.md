---
title: "Visionary Research: Discord CLI v2"
type: feat
status: active
date: 2026-03-26
phase: "0"
api: "discord"
---

# Visionary Research: Discord CLI v2

## Overview

Discord is the dominant communication platform for developer communities, gaming, and increasingly for business. Its API is one of the richest in the communication space - REST endpoints for CRUD, a WebSocket Gateway for real-time events, and a data model that spans guilds, channels, threads, messages, members, roles, permissions, webhooks, and audit logs. The Discord CLI opportunity is unique because the market is bifurcated: export tools (DiscordChatExporter, 10.7k stars) serve archival needs, while discrawl (569 stars, 12 commands) proves the data-tool thesis with SQLite + FTS5 + Gateway sync. No single CLI combines comprehensive REST API coverage with local-first data intelligence.

## API Identity

- **Domain:** Communication / Community Platform
- **Primary users:** Bot developers, server admins, community managers, DevRel teams, moderation teams
- **Core entities:** Guilds, Channels, Messages, Users, Members, Roles, Threads, Reactions, Emojis, Webhooks, Audit Logs, Invites, Bans, Scheduled Events, Stage Instances, Auto Moderation
- **Data profile:**
  - Write pattern: Append-heavy (messages), mutable (guilds/channels/roles/members)
  - Volume: HIGH (millions of messages per active server)
  - Real-time: YES - Gateway WebSocket (mandatory for presence/typing/voice state), plus webhooks for integrations
  - Search need: HIGH - Discord's built-in search is limited, users constantly ask for better search

## Usage Patterns (Top 5 by Evidence)

| Rank | Pattern | Evidence Score | Sources |
|------|---------|---------------|---------|
| 1 | **Server archival/export** | 10/10 | DiscordChatExporter (10.7k stars), discord-dl, discord-backup, Copycord, discord-migrate |
| 2 | **Local search across server history** | 8/10 | discrawl (569 stars), jackwener/discord-cli (78 stars), AnswerOverflow, discord-to-sqlite |
| 3 | **Server management/administration** | 7/10 | discli (7 stars, 40+ commands), selfbot scripts, n8n/Zapier integrations |
| 4 | **Analytics/metrics on server activity** | 6/10 | discord-analytics (Tarasa24), Cially dashboard, james-long/discord-analytics |
| 5 | **Real-time monitoring/alerting** | 5/10 | discord-voice-monitor, discord-user-monitor, discord-quest-watcher |

## Tool Landscape (Beyond API Wrappers)

### Tier 1: API Wrappers
| Tool | Stars | Language | Commands | Status |
|------|-------|----------|----------|--------|
| discli | 7 | TypeScript | 40+ | Active, YAML output, agent-native |
| fourjr/discord-cli | ~50 | Python | ~10 | Stale |
| mrousavy/discord-cli | ~20 | JavaScript | ~5 | Stale |

### Tier 2: Data Tools (the real competition)
| Tool | Stars | Language | What It Does | Key Insight |
|------|-------|----------|-------------|-------------|
| **discrawl** | 569 | Go | SQLite + FTS5 + sync + tail + search + sql | **The gold standard.** 12 commands beat 316-endpoint wrappers. |
| **jackwener/discord-cli** | 78 | Python | SQLite + sync + search + export + analytics | Proves the pattern works in Python too |
| **discord-to-sqlite** | ~30 | Python | Import Discord data package to SQLite | Data package import, not live sync |

### Tier 3: Export/Archival Tools
| Tool | Stars | Language | What It Does |
|------|-------|----------|-------------|
| **DiscordChatExporter** | 10,700 | C# | Export to HTML/TXT/CSV/JSON |
| **discord-dl** | ~200 | Go | Archive channels/guilds |
| **Copycord** | ~150 | JavaScript | Clone servers with real-time sync |
| **discord-server-backup** | ~100 | JavaScript | Full server backup & recreation |

### Tier 4: Analytics/Dashboard Tools
| Tool | Stars | Language | What It Does |
|------|-------|----------|-------------|
| **Cially** | ~50 | TypeScript | Real-time analytics dashboard |
| **AnswerOverflow** | 500+ | TypeScript | Makes Discord threads searchable on web |
| **discord-analytics** | ~100 | JavaScript | Server statistics & visualizations |

## Workflows

### 1. Server History Deep Search
**Steps:** Sync messages -> Index with FTS5 -> Search with domain filters (--channel, --author, --before, --after)
**Frequency:** Daily for active community managers
**Pain point:** Discord's built-in search is slow, incomplete, and can't do cross-channel queries
**Proposed:** `discord-cli search "error" --channel general --author bot --days 7`

### 2. Server Health Dashboard
**Steps:** Fetch member list -> Get message activity -> Calculate metrics -> Report
**Frequency:** Weekly for server admins
**Pain point:** No built-in way to see activity trends, stale channels, inactive members
**Proposed:** `discord-cli health --guild <id> --days 30`

### 3. Audit Log Analysis
**Steps:** Fetch audit log -> Filter by action type -> Cross-reference with members -> Report
**Frequency:** After incidents, weekly for moderation
**Pain point:** Discord UI shows audit log but can't filter, search, or export it
**Proposed:** `discord-cli audit --guild <id> --action member_ban --days 7`

### 4. Channel Archival Pipeline
**Steps:** List channels -> Sync messages -> Export to format -> Store locally
**Frequency:** Monthly or on-demand
**Pain point:** DiscordChatExporter requires .NET, no SQLite, no incremental sync
**Proposed:** `discord-cli export --channel <id> --format json --since 2026-01-01`

### 5. Real-time Event Monitoring
**Steps:** Connect to Gateway -> Filter events -> Display or alert
**Frequency:** Continuous for moderation bots
**Pain point:** Requires a full bot framework just to watch events
**Proposed:** `discord-cli tail --guild <id> --events message_create,member_join`

## Architecture Decisions

| Decision Area | Choice | Rationale |
|--------------|--------|-----------|
| **Persistence** | SQLite with domain-specific tables | HIGH volume + HIGH search need. discrawl proves this works. Messages, members, channels, audit logs all need proper columns for joins and filters. |
| **Real-time** | Gateway WebSocket | Discord's Gateway is the primary real-time channel. REST polling misses events and wastes rate limit budget. Bot tokens get Gateway access. |
| **Search** | FTS5 on message content, channel names, member names | Discord's search is the #1 pain point. FTS5 provides instant local search across all synced history. |
| **Bulk operations** | Paginated sync with snowflake ID cursors | Discord uses snowflake IDs for pagination (?after=snowflake_id). Incremental sync is natural. |
| **Cache** | SQLite IS the cache | No separate caching layer needed. Local DB serves as both cache and search index. |
| **Auth** | Bot token via env var | discrawl and discli both use bot tokens. OAuth2 is for user-facing apps, not CLI tools. |

## Top 5 Features for the World

| Rank | Feature | Score | Evidence | Impact | Feasibility | Uniqueness | Composability | Data Fit | Maintainability | Moat |
|------|---------|-------|----------|--------|-------------|------------|---------------|----------|-----------------|------|
| 1 | **Full-text search with domain filters** | 14/16 | 3 (discrawl 569 stars) | 3 (every admin wants this) | 2 (FTS5 template exists) | 1 (discrawl does it) | 2 (great with pipes) | 2 (perfect fit) | 1 (generated) | 0 |
| 2 | **Incremental sync with Gateway tail** | 13/16 | 3 (discrawl proves it) | 3 (foundation for everything) | 2 (Gateway client needed) | 1 (discrawl does it) | 2 (enables all other features) | 2 (perfect fit) | 1 | 0 |
| 3 | **Raw SQL access to synced data** | 12/16 | 3 (discrawl has it) | 2 (power users love it) | 2 (simple to add) | 1 (discrawl has it) | 2 (ultimate composability) | 2 (perfect) | 1 | 0 |
| 4 | **Audit log analysis with filters** | 12/16 | 2 (Reddit/SO demand) | 3 (every admin needs this) | 2 (API supports it) | 2 (no CLI does this well) | 2 (great with pipes/agents) | 2 (fits data layer) | 0 | 0 |
| 5 | **Server health/analytics dashboard** | 11/16 | 2 (Cially, discord-analytics) | 2 (niche but valuable) | 1 (needs custom code) | 2 (no CLI does this) | 2 (JSON output for dashboards) | 2 (uses local DB) | 0 | 1 |

### Feature Scoring Rationale

**Feature 1 - Full-text search:** The strongest signal. discrawl's entire value proposition is "search your Discord like a database." DiscordChatExporter has 10.7k stars but NO search. The gap is enormous.

**Feature 2 - Incremental sync + Gateway:** Without sync, nothing else works. discrawl proves Gateway tail is viable for a CLI. Snowflake IDs make incremental sync natural.

**Feature 3 - Raw SQL:** discrawl's `sql` command is the killer feature for power users. Direct SQLite access means infinite composability.

**Feature 4 - Audit log analysis:** No existing CLI handles audit logs well. Discord's UI is limited. Server admins constantly need "who did what" analysis.

**Feature 5 - Server health:** Cially (50 stars) and discord-analytics prove demand. But they're bots/dashboards, not CLI tools. A `health` command using local data would be instant.

## Competitive Strategy

**Why this CLI should exist when discrawl has 569 stars:**

1. **Breadth + depth:** discrawl has 12 commands focused on messages/search. We'll have 50+ covering the full REST API PLUS discrawl-level data intelligence. discli has 40+ API commands but zero data intelligence.
2. **Agent-native:** Neither discrawl nor discli has --json + --select + --dry-run + --stdin + --yes + --no-cache. Our CLI is built for AI agents from day one.
3. **Audit logs:** Nobody does this well. Discord's audit log API is rich but underserved.
4. **Export formats:** DiscordChatExporter proves massive demand (10.7k stars) for export. We can offer export FROM our SQLite data.

## Sources
- discrawl: https://github.com/steipete/discrawl (569 stars, Go, 12 commands)
- DiscordChatExporter: https://github.com/Tyrrrz/DiscordChatExporter (10.7k stars, C#)
- discli: https://github.com/ibbybuilds/discli (7 stars, TypeScript, 40+ commands)
- jackwener/discord-cli: https://github.com/jackwener/discord-cli (78 stars, Python, 16 commands)
- Discord API docs: https://docs.discord.com/developers/intro
- Discord API spec: https://github.com/discord/discord-api-spec
- AnswerOverflow: https://github.com/AnswerOverflow/AnswerOverflow (500+ stars)
- Cially: https://github.com/cially/cially (~50 stars)
- awesome-discord: https://github.com/jacc/awesome-discord
