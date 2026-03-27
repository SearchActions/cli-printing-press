---
title: "Honest Analysis: Generated discord-cli vs discrawl"
type: feat
status: active
date: 2026-03-26
---

# Honest Analysis: Generated discord-cli vs discrawl

## Overview

The printing-press generated a 323-command Discord CLI in ~90 minutes and scored it Grade A (96/110) on its own scorecard. discrawl has 11 commands and 551 stars. This analysis asks the uncomfortable question: which one actually works?

**Verdict: discrawl is a real tool. Our discord-cli is a compilation-verified hallucination.**

## The Uncomfortable Truth, Dimension by Dimension

### 1. Sync: The Core Value Proposition

**discrawl:**
- Guild-scoped: iterates `guilds -> channels -> messages` with the correct nested API paths (`/guilds/{id}/channels`, `/channels/{id}/messages`)
- Snowflake cursor tracking per channel in `sync_state` table
- Dual-direction pagination (forward for catch-up, backward for backfill)
- Configurable concurrency (8-32 workers) with auto-sizing
- Incomplete backfill detection and automatic re-batching
- Channel skip logic: skips already-complete channels, saving API calls
- Member refresh on a separate 24-hour cycle
- Rate limit handling delegated to discordgo's battle-tested implementation
- Error classification: retryable (timeout, 429, 5xx) vs permanently unavailable (403, 404)

**Ours:**
- `defaultSyncResources()` returns `[]string{}`. Running `discord-cli sync` with no flags syncs zero things and reports success.
- Hits `GET /channels`, `GET /messages` - flat paths that don't exist in Discord's API. Discord requires guild-scoped routes.
- No guild context at all in the sync flow. The "sync" is a template that assumes flat REST resources.
- The nice domain-specific `UpsertMessage`/`UpsertChannel` methods in our store? **Nothing calls them.** The sync uses the generic `Upsert` which writes to the `resources` table, bypassing our entire domain schema.

**Rating: discrawl 10/10, ours 1/10.** Our sync compiles. It does not sync.

### 2. Search: Why Anyone Would Install This

**discrawl:**
- FTS5 on `message_fts` virtual table indexing `author_name`, `channel_name`, and `content`
- BM25 ranking with secondary sort by `created_at DESC`
- `normalized_content` includes embed titles, attachment filenames, poll questions, reply text - not just raw message text
- Domain filters: `--channel`, `--author`, `--guild` mapped to SQL WHERE clauses on unindexed columns
- Fallback: if FTS5 query fails, transparently degrades to LIKE queries
- FTS versioning: when schema changes, drops and rebuilds the index automatically

**Ours:**
- `search.go` calls `db.Search()` which queries `resources_fts` - the generic table
- Our beautifully designed `messages_fts` table and `SearchMessages` method exist in the store but are never called
- No domain filters (no `--channel`, `--author` flags on the search command)
- The search would return results from the generic `resources` table, which is never populated by our broken sync

**Rating: discrawl 10/10, ours 1/10.** Our search command exists. It searches an empty table.

### 3. Tail: Real-time Monitoring

**discrawl:**
- Real Discord Gateway WebSocket via `discordgo.Session.Open()`
- Handles 6 event types: MessageCreate, MessageUpdate, MessageDelete, ChannelCreate/Update, GuildMemberAdd/Update/Remove
- 4-16 worker pool with buffered queue, 30-second handler timeout, panic recovery
- Periodic repair syncs every 6 hours to catch missed events
- Appends to `message_events` audit log for edit/delete tracking
- Clean shutdown on SIGINT/SIGTERM

**Ours:**
- REST polling on a timer hitting `GET /<resource>` - invalid paths
- No deduplication - re-emits all data every tick
- No Gateway WebSocket connection at all
- Would exhaust rate limits within minutes on any real server

**Rating: discrawl 10/10, ours 0/10.** Ours is architectural fiction.

### 4. Auth: Getting Past the Front Door

**discrawl:**
- `Bot <token>` prefix for bot tokens (Discord's requirement)
- OpenClaw integration for token reuse
- Token from env var (`DISCORD_TOKEN`) or config file
- File permissions locked to 0600

**Ours:**
- `Bearer <token>` prefix always. Discord bots require `Bot <token>`.
- Every API request would return 401 Unauthorized with a bot token.
- We have a full OAuth2 PKCE flow that's impressive... but Discord bots don't use OAuth2 for API access.

**Rating: discrawl 10/10, ours 2/10.** Our auth is correct for the wrong protocol.

### 5. Data Layer: What's Actually in SQLite

**discrawl:**
- 9 domain tables with typed columns, NOT JSON blobs
- `normalized_content` on messages = raw content + embed titles + attachment filenames + poll questions + reply text
- `mention_events` = structural mention tracking (who mentioned whom, target type, timestamp)
- `message_events` = append-only audit log preserving edit/delete history
- `message_attachments` = metadata + extracted text content (for text files up to 256KB)
- Member profile extraction: recursively walks JSON to find bio, pronouns, location, social links
- 10 focused indexes on foreign keys and common query patterns

**Ours:**
- 8 domain tables with typed columns - structurally similar to discrawl
- BUT: nothing populates them. The sync writes to the `resources` generic table.
- `UpsertMessage` correctly decomposes nested JSON and updates FTS. It's good code. Nobody calls it.
- No normalized content. No mention extraction. No attachment text extraction. No message events audit log.
- The store is a beautiful ghost town.

**Rating: discrawl 10/10, ours 4/10.** Our schema is well-designed. It's just empty.

### 6. Workflow Commands: The "Product"

**discrawl (11 commands, all functional):**
- `messages` has `--sync` flag for just-in-time sync before query. Eliminates the two-step workflow.
- `members search` uses a separate `member_fts` table with bio/social profile text
- `members show` provides a full profile with message stats, first/last message, and recent messages
- `mentions` queries a dedicated structural `mention_events` table
- `sql --unsafe --confirm` enables writes with explicit safety gate

**Ours (5 workflow commands, partially functional):**
- `activity`, `stale`, `mentions`, `sql` query local SQLite with correct SQL... against empty tables
- `audit` is the one workflow command that makes a live API call and would actually work (if auth wasn't broken)
- The SQL in our commands is correct. The joins are valid. The queries would return real results IF the tables had data. They don't because sync is broken.

**Rating: discrawl 10/10, ours 3/10.** Our SQL is correct. The precondition (data in tables) is never met.

### 7. Code Quality & Testing

**discrawl:**
- 80% test coverage floor enforced by CI
- golangci-lint + staticcheck + gofumpt + gosec in CI
- govulncheck for dependency vulnerabilities
- gitleaks for secret scanning
- Interface-based dependency injection throughout (mockable time, mockable clients, factory injection)
- 51 commits, clean git history, semantic versioning

**Ours:**
- Zero tests
- Module path is `github.com/USER/discord-cli` - literally the string "USER"
- Dummy import guards on every file (`var _ = strings.ReplaceAll // ensure import`)
- No CI, no linting, no coverage
- `go build` and `go vet` pass. That's the entire quality assurance.

**Rating: discrawl 9/10, ours 2/10.**

## The Scorecard Gamed Itself

Our Steinberger scorecard gave us 96/110 (Grade A). Here's how each dimension maps to reality:

| Dimension | Score | Reality |
|---|---|---|
| Output Modes 10/10 | Correct | --json, --csv, --select, etc. are real and functional |
| Auth 10/10 | **Wrong** | Auth sends `Bearer` instead of `Bot`. Every request would 401. |
| Error Handling 10/10 | Partially true | classifyAPIError is good. But errors from non-existent endpoints are never classified. |
| Terminal UX 10/10 | Correct | Color, progress, formatting work |
| README 5/10 | **Misleading** | Documents commands that don't work. The cookbook shows workflows that would fail. |
| Doctor 10/10 | Correct | Actually checks config, auth, connectivity |
| Agent Native 8/10 | Correct | --json, --select, --dry-run, --yes all work |
| Local Cache 10/10 | **Wrong** | Store exists but is never populated by any working code path |
| Breadth 6/10 | **Superficial** | 323 commands exist. Most hit non-existent API paths. |
| Vision 9/10 | **Wrong** | The vision (SQLite + FTS5 + sync + search + tail) exists in code but none of it works end-to-end |
| Workflows 8/10 | **Wrong** | Workflow commands have correct SQL that queries empty tables |

**Honest score:** ~35/110 (32%). The infrastructure (output, config, error types, CLI framework) is real. Everything that touches the Discord API or local data is broken.

## What the Printing Press Actually Produced

It's not nothing. And it's not what the scorecard claims. Here's what's real:

### Genuinely Good (would survive code review)
- **store.go**: Domain-specific SQLite schema with proper types, FTS5, upsert methods, transaction handling
- **client.go**: HTTP client with rate limit handling, caching, dry-run mode
- **config.go**: TOML config with env var overrides, token persistence
- **helpers.go**: Output formatting with auto-table, field selection, paginated fetching
- **Error types**: Structured exit codes, API error classification
- **CLI framework**: Root command with all agent-native flags properly wired

### Broken (would fail on first real use)
- **sync.go**: Empty resource list, flat API paths, doesn't call domain-specific upsert methods
- **search.go**: Queries wrong table (generic `resources_fts` instead of `messages_fts`)
- **tail.go**: REST polling instead of Gateway WebSocket, hits non-existent paths
- **auth**: `Bearer` prefix instead of `Bot` prefix
- **Module path**: `github.com/USER/discord-cli`
- **Generated API commands**: Snowflake IDs as `int` instead of `string`, missing complex body flags
- **Data pipeline**: Domain tables exist but nothing populates them

### The Gap

The printing-press is good at generating **infrastructure** (HTTP clients, CLI frameworks, SQLite stores, config layers, output formatting). It is bad at generating **domain logic** (API topology, auth protocols, data pipelines, real-time connections).

discrawl's 11 commands represent ~3,000 lines of carefully written domain logic that understands Discord. Our 323 commands represent ~18,000 lines of infrastructure templates with a thin veneer of Discord awareness.

## What Would It Take to Make discord-cli Real?

### Phase 1: Fix the Pipeline (make data flow)
1. Fix auth to use `Bot` prefix for bot tokens
2. Rewrite sync to be guild-scoped (iterate guilds -> channels -> messages using correct Discord API paths)
3. Wire sync to call `UpsertMessage`, `UpsertChannel`, `UpsertMember` instead of generic `Upsert`
4. Fix search to call `SearchMessages` instead of generic `Search`
5. Fix module path from `USER` to a real GitHub org

### Phase 2: Fix the Types
1. Change all snowflake ID parameters from `int` to `string`
2. Add `--guild-id` as a persistent config option so every command doesn't need it

### Phase 3: Add Real-time
1. Add `discordgo` dependency
2. Rewrite tail to use Gateway WebSocket with proper intents

### Phase 4: Add Domain Intelligence
1. Normalized content (embeds + attachments + polls in search text)
2. Mention extraction to structured table
3. Message event audit log (edit/delete tracking)
4. Member profile extraction
5. Attachment text extraction

This is roughly equivalent to writing discrawl from scratch - which makes sense, because discrawl IS what a working Discord CLI looks like.

## The Real Lesson

The printing-press pipeline spent 90 minutes on research, planning, generation, auditing, and scoring. It produced 7 plan artifacts and a 96/110 score. discrawl spent ~3 weeks of focused development by an experienced Go developer who understands Discord's API deeply.

The printing-press is valuable as a **scaffolding accelerator** - it produces a real CLI skeleton with good infrastructure in minutes. But the scorecard's self-assessment creates a dangerous illusion of completeness. A CLI that compiles, has --help text, and passes `go vet` is not the same as a CLI that works.

**discrawl's 11 commands that work > our 323 commands that compile.**

## Acceptance Criteria

- [ ] This analysis is read by the user and informs printing-press improvements
- [ ] The scorecard is updated to test actual API connectivity, not just code structure
- [ ] Future runs distinguish "infrastructure quality" from "domain correctness"

## Sources

- [discrawl](https://github.com/steipete/discrawl) - 551 stars, v0.2.0, live DB at ~/.discrawl/discrawl.db
- Generated discord-cli at ~/cli-printing-press/discord-cli/
- Discord API documentation at docs.discord.com
- discrawl issues: FTS injection, schema migrations, AI agent memory use case
