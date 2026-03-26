---
title: "Research: Discord CLI"
type: feat
status: active
date: 2026-03-25
---

# Research: Discord CLI

## Spec Discovery
- Official OpenAPI spec: https://raw.githubusercontent.com/discord/discord-api-spec/main/specs/openapi.json
- Source: Known-specs registry (verified)
- Format: OpenAPI 3.1.0 (JSON)
- Endpoint count: 200+ (22 resource categories)

## Competitors (Deep Analysis)

### DiscordChatExporter (10,700 stars)
- Repo: https://github.com/Tyrrrz/DiscordChatExporter
- Language: C# (91.3%)
- Commands: ~10 (export-focused: exportdm, exportchannel, exportguild, exportall)
- Last commit: March 21, 2026 (v2.47.1)
- Open issues: 10
- Contributors: 63
- Maintained: YES (actively)
- Notable features: GUI + CLI, HTML/JSON/CSV/TXT export, date range filtering, file partitioning, Docker
- Weaknesses: Export-only (no search, no sync, no management), C# binary (large), no agent-native features (--json/--select/--dry-run), no server management, no audit log analysis

### jackwener/discord-cli (78 stars)
- Repo: https://github.com/jackwener/discord-cli
- Language: Python (100%)
- Commands: ~15 (sync, search, tail, export, stats, timeline, top, recent, today, purge, guilds, channels, members, info)
- Last commit: January 2026
- Open issues: unknown (new repo)
- Maintained: YES (new, active)
- Notable features: SQLite + FTS5 search, incremental sync, tail -f, YAML/JSON output, AI agent integration (SCHEMA.md), activity timeline
- Weaknesses: Python (slow startup), uses user tokens (TOS-violating), no server management (no role/channel CRUD), no audit log, no webhooks, no --dry-run, no doctor command

### discordo (5,400 stars)
- Repo: https://github.com/ayn2op/discordo
- Language: Go (100%)
- Purpose: TUI client (NOT a management CLI - different category)
- Maintained: YES
- Not a direct competitor - it's a chat client, not an API management tool

## User Pain Points
> "Communities are moving en-masse to information blackholes like Discord that cannot be indexed by search engines." - discord-dl README
> "Automated user accounts violate Discord's Terms of Service" - discordo README (applies to user-token tools)
> Rate limiting is the #1 pain point - 429 errors, per-route limits not communicated through headers, IP bans after 10k failed requests

## Auth Method
- Type: Bot token (primary, TOS-compliant)
- Env var convention: DISCORD_TOKEN or DISCORD_BOT_TOKEN
- OAuth2 also supported for user-level access

## Demand Signals
- HN: discordo (5.4k stars) shows strong demand for terminal Discord tools
- HN: Discoding (2026) - AI CLI relay to Discord shows developer workflow integration demand
- AnswerOverflow: Making Discord searchable is a validated business (Y Combinator backed)
- n8n integration: Shows automation/workflow demand for Discord API

## Strategic Justification
**Why this CLI should exist:** No Go CLI exists that combines API management (full REST coverage) + data tool (SQLite sync/search/export) + workflow commands (audit-report, channel-health). DiscordChatExporter has 10.7k stars but is export-only with no search/sync/management. jackwener/discord-cli (78 stars) validates the data-tool approach but uses TOS-violating user tokens and is Python (slow). Our CLI is Go (fast binary), uses bot tokens (TOS-compliant), covers 200+ API endpoints, AND includes discrawl-class workflow commands. It's the first tool that combines all three: API wrapper + data tool + workflow engine.

## Target
- Command count: 50+ (beat jackwener at 15, match API breadth + 7 workflow commands)
- Key differentiator: Go binary + bot-token compliant + SQLite sync/search + audit/health workflow commands
- Quality bar: Steinberger Grade A (80+/100)
