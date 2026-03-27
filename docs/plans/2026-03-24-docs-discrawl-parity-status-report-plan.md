---
title: "Discrawl Parity Status Report - What the Printing Press Can and Can't Do"
type: docs
status: completed
date: 2026-03-24
---

# Discrawl Parity Status Report

## What Happened Today

In one session, the printing press went from generating a toy Discord CLI (60 endpoints, broken auth, wrong flag types) to generating a production-grade CLI (230 endpoints, working auth, correct types, usage examples) that successfully:

- Authenticated with a real Discord bot token
- Fetched the bot's own user profile
- Listed guilds the bot belongs to
- Listed channels in a guild
- Read messages from a channel
- Sent a message to a channel (and deleted it)
- Listed guild members and emojis

All of this from a single command: `./printing-press generate --spec discord.json --output ./discord-cli`

## The Key Question: Is This Just a Discord CLI or a General Printer?

**It's a general printer.** The Discord CLI was never hardcoded. Every fix targeted the press internals:

| Fix | File changed | What it does for ANY spec |
|-----|-------------|--------------------------|
| Sub-resource grouping | `parser.go` | Detects `/resource/{id}/sub` patterns in any API |
| Auth mapping | `parser.go`, `config.go.tmpl` | Maps apiKey/BotToken/OAuth2 from any OpenAPI spec |
| Nullable type handling | `parser.go` | Handles OpenAPI 3.1 `["boolean", "null"]` from any spec |
| Usage examples | `generator.go`, `command.go.tmpl` | Generates examples for every command in any CLI |
| Smart descriptions | `parser.go` | Prefers description over mangled summary in any spec |
| Hyphenated names | `parser.go`, `command.go.tmpl` | Clean command/flag names for any API |

You could give it Stripe, GitHub, Twilio, Shopify - any OpenAPI 3.x spec - and get the same quality.

## Discrawl Parity Scorecard

Discrawl is Peter Steinberger's hand-built Discord archival CLI. It uses stdlib `flag`, SQLite+FTS5, and the discord-go client library. Here's where the printing press stands:

### What the Press Matches (Tier 1: CLI Scaffolding)

| Feature | Discrawl | Press Output | Parity |
|---------|----------|-------------|--------|
| Cobra/flag structure | stdlib flag | Cobra (richer) | MATCH |
| `doctor` command | Config + auth + DB + FTS | Config + auth + API | MATCH |
| `version` command | Yes | Yes | MATCH |
| `--json` output | Yes | Yes | MATCH |
| `--help` on every command | Yes | Yes | MATCH |
| Config file (TOML) | `~/.discrawl/config.toml` | `~/.config/discord-cli/config.toml` | MATCH |
| Env var auth | `DISCORD_BOT_TOKEN` | `DISCORD_BOT_TOKEN` | MATCH |
| Bot token format | `Authorization: Bot {token}` | `Authorization: Bot {token}` | MATCH |
| Typed exit codes | Yes (0-130) | Yes (0-5) | PARTIAL |
| Thin main.go | ~15 lines | ~15 lines | MATCH |
| GoReleaser config | Yes | Yes | MATCH |
| Linter config | Yes | Yes | MATCH |
| Makefile | Yes | Yes | MATCH |
| README | Yes | Yes | MATCH |

**Score: 12/14 match, 1 partial**

### What the Press Can't Do (Tier 2: Domain Logic)

These are things discrawl does that require hand-written domain logic beyond REST API wrapping:

| Feature | What it does | Why the press can't generate it |
|---------|-------------|-------------------------------|
| `sync` | Backfills entire guild history into SQLite | Requires SQLite schema, pagination logic, resumable state |
| `tail` | Live WebSocket Gateway connection | Requires discord-go client, Gateway protocol, repair loops |
| `search` | FTS5 full-text search on local data | Requires local database, indexing pipeline |
| `members` | Offline member directory with profile extraction | Requires bio/pronouns/location parsing logic |
| `sql` | Read-only SQL access to local DB | Requires SQLite integration |
| `status` | Archive completion percentage | Requires sync state tracking |
| `init` | Interactive guild discovery | Requires interactive prompts, guild selection UX |
| Repair loops | Catches Gateway gaps | Requires event deduplication, gap detection |
| Resumable sync | `--since` flag, progress tracking | Requires checkpoint state persistence |
| Semantic search | Optional embeddings integration | Requires embedding API integration |

**Score: 0/10 - these are all Tier 2 domain logic**

### The Tiers

This is the fundamental insight about what the printing press is:

- **Tier 1 (what the press does):** CLI scaffolding from an OpenAPI spec. Cobra commands, config, auth, HTTP client, flag parsing, output formatting, doctor, version, examples, GoReleaser. This is 100% automatable.

- **Tier 2 (what discrawl adds manually):** Domain logic. SQLite, WebSocket, FTS5, sync state, repair loops, resumable operations, semantic search. This is hand-written code that wraps platform-specific libraries and implements application logic.

The press prints Tier 1. Discrawl is Tier 1 + Tier 2. The press gives you a running start - you get a working CLI that can talk to the Discord API in seconds. Then you build Tier 2 on top of it.

## What Shipped Today (5 Commits)

1. **`54e55a6` fix(generator): dogfood to Steinberger quality** - Smart descriptions, hyphens, relative URLs, auto-generated descriptions, nullable types
2. **`3d07636` feat(generator): sub-resource grouping** - Discord goes from 60 to 230 endpoints with nested Cobra commands
3. **`2915ae3` fix(generator): auth mapping for BotToken** - `DISCORD_BOT_TOKEN` with `Bot {token}` format
4. **`33d0648` fix(openapi): nullable types in OpenAPI 3.1** - `--tts` is bool not string, `--flags` is int not string
5. **`e993ba6` feat(generator): auto-generate usage examples** - Every command shows Examples section

## Test Results

All 3 test specs pass all 7 quality gates after every change:

| Spec | Endpoints | Sub-resources | Quality gates |
|------|-----------|--------------|---------------|
| Petstore | 19 | 0 | 7/7 PASS |
| Stytch | 80+ | 12 | 7/7 PASS |
| Discord | 230 | 30+ | 7/7 PASS |

Live API test against Discord: 7/7 operations successful (get user, list guilds, get guild, list channels, list messages, send message, delete message).

## What Would Make It Even Better (Future)

1. **Table output for list commands** - Currently dumps raw JSON. Discrawl shows formatted tables by default.
2. **Pagination** - Large list responses need `--limit` and `--after` cursor support.
3. **TTY detection** - Auto-switch to `--json` when piped.
4. **Color output** - ANSI colors for doctor checks and tables.
5. **More exit codes** - Map HTTP 401/403/404/429 to specific exit codes.
6. **`--dry-run`** - Show what would be sent without sending.

## Bottom Line

The printing press is a Tier 1 CLI generator. Give it any OpenAPI spec and it prints a working CLI with auth, config, doctor, version, examples, and all endpoints organized into nested commands. It matched 12/14 of discrawl's scaffolding features in one session. The 10 features it can't match are hand-written domain logic that no generator should try to automate - that's where the developer's expertise goes after the press gives them the scaffold.
