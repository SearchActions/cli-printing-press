---
title: "Research: Discord CLI v2"
type: feat
status: active
date: 2026-03-26
phase: "1"
api: "discord"
---

# Research: Discord CLI v2

## Spec Discovery
- **Official OpenAPI spec:** https://raw.githubusercontent.com/discord/discord-api-spec/main/specs/openapi.json
- **Source:** discord/discord-api-spec GitHub repository (official, maintained by Discord)
- **Format:** OpenAPI 3.1.0 JSON
- **Endpoint count:** 140 endpoints across 16 resource categories
- **Size:** 1.07 MB
- **Key categories:** guilds (46), channels (31), applications (16), lobbies (10), users (9)

## Competitors (Deep Analysis)

### discrawl (569 stars) - THE BENCHMARK
- **Repo:** https://github.com/steipete/discrawl
- **Language:** Go 1.26+
- **Commands:** 12 (init, sync, tail, search, messages, mentions, sql, members, channels, status, doctor)
- **Last commit:** March 8, 2026 (51 commits)
- **Open issues:** 4
- **Maintained:** YES - actively developed
- **Notable features:**
  - SQLite + FTS5 for local search
  - Gateway WebSocket for real-time tail
  - Bot-token only (no user-token hacks)
  - Multi-guild schema
  - Structured mention tracking
  - Attachment text extraction into search index
  - Periodic repair syncs during tail
- **Weaknesses:**
  - Only 12 commands - no REST API coverage
  - No audit log support
  - No server health/analytics commands
  - No export functionality
  - No --json/--select/--dry-run agent-native flags
  - 4 open issues including "AI agent integration" request (#15)
  - No schema migrations (#8)

### DiscordChatExporter (10,700 stars)
- **Repo:** https://github.com/Tyrrrz/DiscordChatExporter
- **Language:** C# (.NET)
- **Commands:** GUI + CLI with export commands
- **Last commit:** March 21, 2026 (v2.47.1)
- **Open issues:** 10
- **Contributors:** 63
- **Maintained:** YES - very active
- **Notable features:**
  - Export to HTML (dark/light), TXT, CSV, JSON
  - User or bot token auth
  - File partitioning and date range filtering
  - Cross-platform
  - Offline-capable exports
- **Weaknesses:**
  - Export-ONLY tool - no search, no sync, no real-time
  - Requires .NET runtime
  - Hits API on every export (no incremental/local cache)
  - No SQLite storage
  - No agent-native features (--json, --select, etc.)

### discli (7 stars)
- **Repo:** https://github.com/ibbybuilds/discli
- **Language:** TypeScript (Node.js)
- **Commands:** 40+ across 9 categories
- **Maintained:** Active (59 commits)
- **Notable features:**
  - YAML output (claimed "5x fewer tokens than JSON" for AI agents)
  - --dry-run support
  - --confirm for destructive operations
  - SOUL.md personality file
  - "One command = one API call" philosophy
- **Weaknesses:**
  - Only 7 stars - minimal adoption
  - No local storage/SQLite
  - No search, no sync, no real-time
  - No FTS5
  - Requires Node.js
  - Pure API wrapper with no compound commands

### jackwener/discord-cli (78 stars)
- **Repo:** https://github.com/jackwener/discord-cli
- **Language:** Python
- **Commands:** 16 across 4 categories
- **Notable features:**
  - Local-first SQLite storage
  - Full-text search
  - YAML/JSON structured output
  - Export in multiple formats
  - Analytics/timeline generation
- **Weaknesses:**
  - Python (slower than Go)
  - 78 stars - moderate adoption
  - Unclear maintenance status

## User Pain Points

> "Using discrawl as memory augmentation for AI agents" - discrawl issue #15 (codexGW, March 2026)

> "Storage: schema migrations & versioning" - discrawl issue #8 (thenotespublisher, March 2026)

> "Search/FTS: injection & tokenizer configuration" - discrawl issue #9 (thenotespublisher, March 2026)

> Rate limits on the role create endpoint are "extreme" with a 24-hour cooldown - Discord support forum

> Discord's built-in search is slow, incomplete, and can't do cross-channel queries - widespread community complaint

> DiscordChatExporter requires .NET and hits the API fresh on every export - no local cache - implied pain from architecture

## Auth Method
- **Type:** Bot token (primary), OAuth2 (for user-facing apps)
- **Env var convention:** `DISCORD_TOKEN` (discrawl), `DISCORD_BOT_TOKEN` (discli)
- **Our choice:** `DISCORD_TOKEN` (matches the market leader)

## Demand Signals
- discordo (Discord TUI client) - HN discussion August 2022, showing demand for terminal Discord access
- Discoding (AI CLI to Discord bridge) - HN Show HN February 2026, proving CLI + Discord integration demand
- Remote-OpenCode (Discord bot for AI coding) - HN Show HN February 2026, AI agent + Discord demand
- discrawl issue #15 explicitly asks for AI agent memory integration
- AnswerOverflow (500+ stars) makes Discord searchable on web - proves search is the killer need

## Strategic Justification

**Why this CLI should exist when discrawl has 569 stars:**

1. **discrawl is depth-only, we're depth + breadth.** discrawl has 12 commands focused on messages/search. It has ZERO coverage of the other 120+ Discord API endpoints. Server admins can't manage roles, channels, invites, bans, or webhooks from discrawl.

2. **No audit log tool exists as a CLI.** Discord retains audit logs for 45 days. Our CLI persists them in SQLite forever and adds forensic analysis commands (filter by action, moderator, target, cross-reference with members).

3. **Agent-native gap.** discrawl's issue #15 asks for AI agent integration. Our CLI ships with --json, --select, --dry-run, --stdin, --yes, --no-cache from day one. It's built for the age of AI agents.

4. **discli proves breadth demand but has zero traction (7 stars).** The API wrapper approach alone doesn't resonate. Users want breadth PLUS data intelligence. We provide both.

5. **Go binary, zero dependencies.** Unlike DiscordChatExporter (.NET) or jackwener/discord-cli (Python) or discli (Node.js), our CLI is a single Go binary. Install and run.

## Target
- **Command count:** 60+ (12 data layer + 7 workflows + 40+ API endpoints)
- **Key differentiator:** discrawl-level data intelligence PLUS full REST API coverage PLUS agent-native design
- **Quality bar:** Steinberger Grade A (80+/100)
- **Competitive positioning:** "discrawl + discli + audit logs + analytics, in one Go binary"
