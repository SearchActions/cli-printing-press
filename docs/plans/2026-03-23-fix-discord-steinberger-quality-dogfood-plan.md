---
title: "Dogfood Discord CLI to Steinberger Quality with Live API Testing"
type: fix
status: completed
date: 2026-03-23
---

# Dogfood Discord CLI to Steinberger Quality with Live API Testing

## Overview

The printing press generates a functional Discord CLI, but it's not Steinberger quality. 230 operations in the spec, zero summaries, zero descriptions, zero tags - Discord is the hardest test. The current output has 5 critical quality gaps that make it feel like a toy instead of a tool. This plan is a tight dogfood loop: generate, use against the real Discord API with a real bot token, find what's broken, fix the press, repeat.

## Problem Statement

Current Discord CLI baseline (generated 2026-03-23):

| Quality dimension | Current state | Steinberger target |
|---|---|---|
| **Endpoint coverage** | 60/230 (20-per-resource cap) | 230/230 (smart sub-resources) |
| **Descriptions** | Auto-generated from name ("Get", "Create message") | Real descriptions or smart context-aware generation |
| **Auth** | "not configured" despite BotToken scheme in spec | `DISCORD_BOT_TOKEN` env var, `Authorization: Bot {token}` |
| **Flag types** | `--tts string`, `--flags string` (wrong types) | `--tts` (bool), `--flags int` (correct types) |
| **Sub-resource grouping** | Flat list of 20 commands under `guilds` | `guilds members list`, `guilds roles create`, `guilds emojis get` |
| **Usage examples** | None | `discord-cli channels create-message 123 --content "Hello"` |
| **Live API test** | Never tested | Actually sends messages, lists guilds, creates channels |

## Execution Loop

Same loop as the previous plan, but targeting Discord specifically and testing against the live API:

```
while (not steinberger quality):
  1. Build: go build -o ./printing-press ./cmd/printing-press
  2. Clean: rm -rf discord-cli/
  3. Generate: ./printing-press generate --spec testdata/openapi/discord.json --output ./discord-cli
  4. Build CLI: cd discord-cli && go build -o ./discord ./cmd/discord-cli
  5. USE IT against real Discord API:
     - discord-cli doctor (auth configured, API reachable 200)
     - discord-cli users get-my (get bot's own user)
     - discord-cli guilds list-my-guilds... wait, that's under users
     - discord-cli users list-my-guilds (list bot's guilds)
     - discord-cli guilds get <guild_id> (get guild details)
     - discord-cli channels list-messages <channel_id> (read messages)
     - discord-cli channels create-message <channel_id> --content "test" (send a message!)
  6. Write down what's wrong
  7. Fix the PRESS (parser.go, templates, generator.go)
  8. go test ./...
  9. Repeat
```

## Phase 1: Sub-Resource Grouping (Biggest Impact)

**Problem:** Discord's `guilds` resource has 90+ endpoints covering members, roles, emojis, bans, auto-mod, scheduled events, soundboard, stickers, templates, voice states, webhooks, and more. The current 20-endpoint cap loses 70+ endpoints.

**Fix:** Detect sub-resources from path segments. `/guilds/{guild_id}/members` becomes `guilds members list` instead of `guilds list-members`.

### Changes to `internal/openapi/parser.go`:

1. **Add sub-resource detection to `mapResources`:**
   - After extracting the primary resource name from the first path segment, check if there's a second non-param segment
   - If yes, create a nested resource: `guilds` -> `members`, `guilds` -> `roles`, `guilds` -> `emojis`
   - Store as `spec.Resource` with a `SubResources map[string]Resource` field

2. **Update `resourceNameFromPath` to return both primary and sub-resource:**
   - `/guilds/{guild_id}/members/{user_id}` -> primary: `guilds`, sub: `members`
   - `/guilds/{guild_id}/auto-moderation/rules` -> primary: `guilds`, sub: `auto-moderation`
   - `/guilds/{guild_id}` (no sub-path after param) -> primary: `guilds`, sub: "" (direct)

3. **Raise or remove the endpoint limit for sub-resources:**
   - Each sub-resource gets its own 20-endpoint limit
   - The parent resource only holds "direct" endpoints (get, update, delete the guild itself)

### Changes to `internal/spec/spec.go`:

4. **Add SubResources field:**
   ```go
   type Resource struct {
       Description  string                 `yaml:"description"`
       Endpoints    map[string]Endpoint    `yaml:"endpoints"`
       SubResources map[string]Resource    `yaml:"sub_resources,omitempty"`
   }
   ```

### Changes to `internal/generator/templates/`:

5. **Update `command.go.tmpl` to generate nested Cobra commands:**
   - Parent command (`guilds`) adds sub-resource commands as children
   - Sub-resource command (`members`) has its own endpoints as children
   - Result: `discord-cli guilds members list <guild_id>`

6. **Update `root.go.tmpl` to handle sub-resources recursively**

### Acceptance:

- [ ] `discord-cli guilds --help` shows sub-resources: members, roles, emojis, bans, auto-mod, etc.
- [ ] `discord-cli guilds members --help` shows: list, get, update, delete, search
- [ ] `discord-cli guilds roles --help` shows: list, create, update, delete
- [ ] No endpoints skipped due to limit (or minimal skipping)
- [ ] `go test ./...` passes

## Phase 2: Auth Mapping for Discord

**Problem:** Discord uses `Authorization: Bot {token}` header. The spec has `BotToken` security scheme type `apiKey` in header `Authorization`. The parser maps this to `api_key` with `DISCORD_API_KEY` env var, but Discord bots use `Bot ` prefix.

### Changes to `internal/openapi/parser.go`:

1. **Detect Discord-style bot token auth:**
   - When scheme type is `apiKey`, in=header, name=Authorization
   - Check if scheme name contains "bot" (case-insensitive)
   - Set `auth.Format = "Bot {token}"` and `auth.EnvVars = ["DISCORD_BOT_TOKEN"]`

2. **Or more generically:** map the scheme name into the env var name:
   - `BotToken` -> `DISCORD_BOT_TOKEN`
   - `OAuth2` -> `DISCORD_TOKEN`

### Acceptance:

- [ ] `discord-cli doctor` with `DISCORD_BOT_TOKEN=xxx` shows auth=configured
- [ ] `discord-cli users get-my` with real token returns bot user JSON
- [ ] Auth header is `Authorization: Bot {token}` (not `Bearer`)

## Phase 3: Flag Type Accuracy

**Problem:** `--tts string` should be `--tts` (bool). `--flags string` should be `--flags int`. The parser falls through to `string` for any type it doesn't recognize, but the Discord spec uses proper types in the schema.

### Investigation:

1. Check why `tts` (boolean in spec) renders as string flag
2. Check if `mapSchemaType` correctly handles the `boolean` type from the request body schema
3. The issue might be in `mapRequestBody` -> schema property extraction

### Changes:

4. Fix any type mapping bugs found
5. Ensure `bool` params render as `--flag` (BoolVar) not `--flag string`

### Acceptance:

- [ ] `discord-cli channels create-message --help` shows `--tts` as a boolean flag (no value needed)
- [ ] `--flags` is an int flag
- [ ] Complex object flags are correctly skipped (not shown as `--embeds string`)

## Phase 4: Usage Examples in Help

**Problem:** No `--help` output shows usage examples. Steinberger CLIs always show examples.

### Changes to `internal/generator/templates/command.go.tmpl`:

1. **Add `Example` field to Cobra commands:**
   ```go
   cmd := &cobra.Command{
       Use:     "create-message <channel_id>",
       Short:   "Create message",
       Example: "  discord-cli channels create-message 123456789 --content \"Hello world\"",
   }
   ```

2. **Auto-generate examples from endpoint metadata:**
   - For POST/PUT: show required flags with placeholder values
   - For GET with positional args: show the positional arg
   - For DELETE: show the positional arg
   - Include the full command path: `{cli-name} {resource} {command} {args} {flags}`

### Changes to `internal/generator/generator.go`:

3. **Add `example` template function** that generates example strings

### Acceptance:

- [ ] `discord-cli channels create-message --help` shows an Example section
- [ ] `discord-cli guilds get --help` shows `discord-cli guilds get 123456789`
- [ ] Examples use realistic placeholder values (not empty strings)

## Phase 5: Live API Dogfood (YOLO Mode)

**Setup:**
1. User provides Discord bot token
2. Configure: `DISCORD_BOT_TOKEN=xxx discord-cli doctor`
3. Or write to `~/.config/discord-cli/config.toml`

**Test matrix (all must work with real API):**

- [ ] `discord-cli doctor` - all green
- [ ] `discord-cli users get-my` - returns bot user JSON with username, id, discriminator
- [ ] `discord-cli users list-my-guilds` - returns list of guilds the bot is in
- [ ] `discord-cli guilds get <guild_id>` - returns full guild object
- [ ] `discord-cli guilds list-channels <guild_id>` - returns channel list
- [ ] `discord-cli channels get <channel_id>` - returns channel details
- [ ] `discord-cli channels list-messages <channel_id>` - returns recent messages
- [ ] `discord-cli channels create-message <channel_id> --content "Hello from the printing press"` - sends a real message
- [ ] `discord-cli guilds list-emojis <guild_id>` - returns emoji list (tests sub-resources)
- [ ] `discord-cli guilds members list <guild_id>` - returns member list (tests sub-resources)

## Scope

Fix the press to make Discord output Steinberger-quality. Nothing else:
- Only modify files in `internal/` and `internal/generator/templates/`
- Never edit generated `*-cli/` output
- Don't break Petstore or Stytch (run all 3 specs in test)
- No README, no docs, no catalog changes

## Phases in Priority Order

1. **Sub-resource grouping** - Unlocks 170 missing endpoints. Biggest bang for buck.
2. **Auth mapping** - Required for live API testing. Blocks Phase 5.
3. **Flag types** - Polish. Wrong types look amateur.
4. **Usage examples** - Polish. Missing examples feel incomplete.
5. **Live API dogfood** - Validation. Proves everything works end-to-end.

## How to Run

```bash
cd ~/cli-printing-press
# Phase 1-4: fix the press
# Phase 5: set up Discord bot token and test live
export DISCORD_BOT_TOKEN="your-bot-token-here"
./printing-press generate --spec testdata/openapi/discord.json --output ./discord-cli
cd discord-cli && go build -o ./discord ./cmd/discord-cli
./discord doctor
./discord users get-my
./discord channels create-message <channel_id> --content "It works."
```
