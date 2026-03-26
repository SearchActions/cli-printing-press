---
title: "Research: Discord CLI"
type: feat
status: active
date: 2026-03-25
---

# Research: Discord CLI

## Spec Discovery
- Official OpenAPI spec: https://raw.githubusercontent.com/discord/discord-api-spec/main/specs/openapi.json
- Source: Known-specs registry (discord/discord-api-spec GitHub repo)
- Format: JSON, OpenAPI 3.1.0
- Endpoint count: 200+ (Discord's full REST API)

## Competitors (Deep Analysis)

### discli (ibbybuilds/discli) - 6 stars
- Repo: https://github.com/ibbybuilds/discli
- Language: TypeScript/JavaScript (Node.js)
- Commands: ~51 across 9 resource groups (server, invites, channels, roles, members, permissions, messages, emojis, audit)
- Last commit: Recent (2026)
- Open issues: 0
- Maintained: Yes (new project)
- Notable features: YAML output by default (token-efficient for AI), --confirm for destructive ops, SOUL.md personality customization, no WebSocket needed
- Weaknesses: Only 6 stars, limited to subset of Discord API (no webhooks, slash commands, scheduled events, threads, auto-mod, stickers, stage instances, voice regions), TypeScript (requires Node.js runtime)

### discordo (ayn2op/discordo) - 5,400 stars
- Repo: https://github.com/ayn2op/discordo
- Language: Go
- Commands: N/A (TUI chat client, not API management CLI)
- Last commit: Active
- Open issues: 45
- Maintained: Yes, 40 contributors
- Notable features: Terminal Discord chat client, QR code login, TOML config
- Weaknesses: Different category entirely - interactive chat TUI, not an API management CLI. Violates Discord ToS (automated user accounts). Not comparable.

## User Pain Points
> "Using external clients were against Discord's wish and terms of service" - discord-cli README (highlights the need for bot-token-based API CLIs that are ToS-compliant)
> "Official plugin support and third party client" discussion on discord-api-docs shows demand for programmatic Discord access beyond the web UI - Discussion #3857

## Auth Method
- Type: Bot token (Bearer) via Authorization header
- Env var convention: DISCORD_BOT_TOKEN (discli uses this)

## Demand Signals
- discord/discord-api-docs Discussion #3857: users requesting official third-party client support
- discli's "for AI agents and humans" framing shows the agent-native CLI demand is recognized but underserved (6 stars = early, not widely adopted)

## Strategic Justification
**Why this CLI should exist:** discli is the only API management CLI and has just 6 stars with ~51 commands covering a fraction of Discord's 200+ endpoint API. It requires Node.js runtime. Our CLI will be a single Go binary covering the full Discord API spec, with --json/--select/--dry-run/--stdin for agent workflows, and proper doctor/cache/auth patterns. We'll beat discli on breadth (full API coverage vs partial) and distribution (single binary vs npm install).

## Target
- Command count: 80+ (match full API spec coverage, beat discli's 51)
- Key differentiator: Full API coverage in a single Go binary with agent-native features (--json, --select, --dry-run, --stdin, doctor, cache)
- Quality bar: Steinberger Grade A (72+/90)
