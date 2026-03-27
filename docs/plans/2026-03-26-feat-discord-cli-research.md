---
title: "Research: Discord CLI"
type: feat
status: active
date: 2026-03-26
phase: "1"
api: "discord"
---

# Research: Discord CLI

## Spec Discovery

- **Official OpenAPI spec:** https://raw.githubusercontent.com/discord/discord-api-spec/main/specs/openapi.json
- **Source:** discord/discord-api-spec GitHub repository (official Discord org)
- **Format:** OpenAPI 3.1.0, JSON
- **Size:** ~1MB+
- **Verified:** Yes (in printing-press known-specs registry)

## Competitors (Deep Analysis)

### DiscordChatExporter (10,700 stars)
- **Repo:** https://github.com/Tyrrrz/DiscordChatExporter
- **Language:** C# (91.3%)
- **Last commit:** 2026-03-21 (v2.47.1) - actively maintained
- **Contributors:** 63
- **Open issues:** 10
- **Notable features:** Export to HTML (dark/light), TXT, CSV, JSON. Rich media support. Date range filtering. Self-contained offline exports. Both GUI and CLI.
- **Weaknesses:**
  - Export-only - no local persistence, no search, no analytics
  - Issue #1210: "Rework CLI commands to be more composable" - users want better CLI UX
  - Issue #1265: Thread exports missing parent message context
  - Issue #1483: Can't export age-restricted channels
  - No --json/--select/--dry-run agent-native features
  - No real-time tail capability
  - Requires manual channel selection for compliance exports

### discrawl (564 stars) - THE BENCHMARK
- **Repo:** https://github.com/steipete/discrawl
- **Language:** Go (100%)
- **Last commit:** Recent (main branch active)
- **Contributors:** Small team
- **Open issues:** 4
- **Commands:** 12 (init, sync, tail, search, messages, mentions, sql, members, channels, status, doctor)
- **Notable features:** SQLite+FTS5, Gateway WebSocket tail, offline member directory, raw SQL access, multi-guild support, structured mention tracking
- **Weaknesses:**
  - Issue #15: "Using discrawl as memory augmentation for AI agents" - no agent-native features (no --json, --select, --dry-run)
  - Issue #9: FTS search injection & tokenizer configuration gaps
  - Issue #8: No schema migrations/versioning
  - No audit log support
  - No activity/analytics commands
  - No webhook management
  - No general-purpose REST API wrapper (only data-focused commands)

### discli (6 stars) - Agent-native reference
- **Repo:** https://github.com/ibbybuilds/discli
- **Language:** TypeScript
- **Notable features:** YAML output (5x fewer tokens than JSON), --dry-run, --confirm for destructive ops, SCHEMA.md for agent parsing, SOUL.md personality file
- **Weaknesses:** No data layer, no SQLite, no search, no analytics. One-command-one-API-call only.

### jackwener/discord-cli (78 stars) - Python SQLite reference
- **Repo:** https://github.com/jackwener/discord-cli
- **Language:** Python
- **Last commit:** 2025-01-10 (stale)
- **Notable features:** SQLite sync, FTS search, export, AI analysis integration
- **Weaknesses:** Uses user tokens (TOS violation risk), Python (not a single binary), 1 contributor, stale

## User Pain Points

> "Rework the CLI commands to be more composable" - DiscordChatExporter Issue #1210 (users want piping, filtering, structured output)

> "Using discrawl as memory augmentation for AI agents" - discrawl Issue #15 (users want agent-native features for LLM integration)

> "Discord's text search isn't exact, which can be a big problem sometimes" - HN item 36748981 (native Discord search is unreliable)

> "Once a server exceeds 1000 members, users will no longer be able to see offline members" - Discord Support (member visibility limitation drives offline member directory demand)

## Auth Method
- **Type:** Bot Token (Bearer token in Authorization header)
- **Env var convention:** `DISCORD_BOT_TOKEN` (discrawl uses `DISCRAWL_TOKEN`, discli uses `DISCORD_TOKEN`)
- **Our choice:** `DISCORD_TOKEN` (shortest, most conventional)

## Demand Signals
- HN: discordo discussion (Aug 2022) - terminal Discord client demand
- HN: Discoding (Feb 2026) - AI CLIs relayed to Discord
- HN: Remote-OpenCode (Feb 2026) - controlling AI coding from Discord
- 9+ GitHub repos storing Discord data in SQLite
- Discord support forums: multiple feature requests for offline member visibility and better search

## Strategic Justification

**Why this CLI should exist when discrawl already has 564 stars:**

1. **Agent-native gap:** discrawl has no --json, --select, --dry-run, --stdin, --yes flags. Every AI agent integration (Issue #15) hits this wall. Our CLI is agent-native from day one.
2. **API breadth gap:** discrawl only covers data commands (12 commands). No guild management, no webhook orchestration, no role management, no channel CRUD. Our CLI wraps the full Discord REST API (200+ endpoints) PLUS the data layer.
3. **Analytics gap:** Neither discrawl nor any other CLI provides per-member activity rankings, channel health reports, or stale channel detection. Our compound queries (validated in Phase 0.7) fill this.
4. **Audit log gap:** No Discord CLI exposes audit logs with filtering. Our audit command fills a real moderation need.
5. **Single binary:** discrawl is Go (good), but DiscordChatExporter is C#, jackwener/discord-cli is Python. We're Go - single binary, zero dependencies, cross-platform.

**In summary:** We combine discrawl's data depth (SQLite+FTS5+Gateway) with full API breadth (200+ endpoints) and agent-native features (--json/--select/--dry-run). No existing tool does all three.

## Target
- **Command count:** 50+ (12 workflow commands + 40+ generated API wrappers)
- **Key differentiator:** Data layer + full API + agent-native in one Go binary
- **Quality bar:** Steinberger Grade A (80+/100)
