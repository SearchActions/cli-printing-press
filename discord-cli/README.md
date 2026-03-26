# discord-cli

Manage Discord servers, channels, messages, roles, and more from your terminal. Built for bot developers, server admins, and AI agents.

## Install

### Homebrew

```
brew install USER/tap/discord-cli
```

### Go

```
go install github.com/USER/discord-cli/cmd/discord-cli@latest
```

### Binary

Download from [Releases](https://github.com/USER/discord-cli/releases).

## Quick Start

```bash
# 1. Set your bot token
export DISCORD_TOKEN="Bot YOUR_BOT_TOKEN_HERE"

# 2. Verify everything works
discord-cli doctor

# 3. List your guilds (servers)
discord-cli users list-my-guilds --json

# 4. List channels in a guild
discord-cli guilds channels list-guild 123456789012345678 --json
```

## Workflow Commands

These compound commands combine multiple API calls into single operations. They're the reason this CLI exists - not just an API wrapper.

### channel-health

Detect stale channels and show activity trends across a guild.

```bash
# Find inactive channels (no messages in 30 days)
discord-cli workflow channel-health --guild 123456789012345678

# Custom staleness threshold
discord-cli workflow channel-health --guild 123456789012345678 --days 14 --json
```

### audit-report

Analyze audit log entries grouped by action type and user.

```bash
# Full audit report
discord-cli workflow audit-report --guild 123456789012345678

# Filter by action type (22 = MEMBER_BAN_ADD)
discord-cli workflow audit-report --guild 123456789012345678 --action 22 --json
```

### member-report

Analyze member roles, join dates, and composition across a guild.

```bash
# Member report with role distribution
discord-cli workflow member-report --guild 123456789012345678

# Fetch more members
discord-cli workflow member-report --guild 123456789012345678 --max 500 --json
```

### server-snapshot

Backup guild configuration (roles, channels, emojis) to JSON.

```bash
# Backup to file
discord-cli workflow server-snapshot --guild 123456789012345678 -o backup.json

# Backup to stdout
discord-cli workflow server-snapshot --guild 123456789012345678 --json
```

### message-stats

Show message volume and top contributors for a channel or guild.

```bash
# Stats for a specific channel
discord-cli workflow message-stats --channel 123456789012345678 --limit 100

# Stats across all channels in a guild
discord-cli workflow message-stats --guild 123456789012345678 --json
```

### prune-preview

Preview how many members would be pruned without executing.

```bash
# Preview 7-day prune (never modifies the guild)
discord-cli workflow prune-preview --guild 123456789012345678 --days 7
```

### webhook-test

Send a test payload to a Discord webhook.

```bash
# Simple test message
discord-cli workflow webhook-test --id 123456789012345678 --token abcdef123456 --message "Hello from CLI!"

# Rich embed via stdin
echo '{"content":"Deploy alert","embeds":[{"title":"Build #42","color":3066993}]}' | discord-cli workflow webhook-test --id 123456789012345678 --token abcdef123456 --stdin
```

## API Commands

### channels

Manage Discord channels, messages, threads, pins, and permissions.

```bash
# Get channel details
discord-cli channels get 123456789012345678

# List messages
discord-cli channels messages list 123456789012345678 --limit 25 --json

# Send a message
discord-cli channels messages create 123456789012345678 --content "Hello, world!"

# Send a rich message with embeds
echo '{"content":"Announcement","embeds":[{"title":"v2.0 Released","description":"Check out the new features","color":5814783}]}' | discord-cli channels messages create 123456789012345678 --stdin

# Pin a message
discord-cli channels pins pin 123456789012345678 987654321098765432

# Create a thread
echo '{"name":"Bug Discussion","auto_archive_duration":1440}' | discord-cli channels threads create 123456789012345678 --stdin
```

### guilds

Manage guilds (servers) - members, roles, bans, channels, audit logs, and more.

```bash
# Get guild details
discord-cli guilds get 123456789012345678 --json

# List members
discord-cli guilds members list 123456789012345678 --limit 100 --json

# List roles
discord-cli guilds roles list-guild 123456789012345678 --json

# Ban a user
discord-cli guilds bans user-from-guild 123456789012345678 987654321098765432

# List audit log
discord-cli guilds audit-logs list-guild-entries 123456789012345678 --json

# Create a channel
echo '{"name":"new-channel","type":0,"permission_overwrites":[{"id":"ROLE_ID","type":0,"allow":"1024"}]}' | discord-cli guilds channels create-guild 123456789012345678 --stdin
```

### webhooks

```bash
# Get webhook details
discord-cli webhooks get 123456789012345678

# Execute a webhook
echo '{"content":"Build passed!","embeds":[{"title":"CI/CD","color":3066993}]}' | discord-cli webhooks execute 123456789012345678 TOKEN --stdin
```

### users

```bash
# Get current user
discord-cli users get-my --json

# List guilds you're in
discord-cli users list-my-guilds --json

# Get a user by ID
discord-cli users get 123456789012345678 --json
```

## Local Sync and Search

Sync messages to a local SQLite database for instant full-text search without API calls.

```bash
# Sync all resources
discord-cli sync

# Full-text search across synced data
discord-cli search "error timeout" --json

# Follow new messages (tail -f for Discord)
discord-cli tail messages --interval 5s

# Export messages to JSONL
discord-cli export messages --format jsonl --output messages.jsonl

# Analytics on synced data
discord-cli analytics --type messages --group-by channel_id
```

## Output Formats

```bash
# Human-readable table (default)
discord-cli guilds roles list-guild 123456789012345678

# JSON for scripting and agents
discord-cli guilds roles list-guild 123456789012345678 --json

# Filter specific fields
discord-cli guilds roles list-guild 123456789012345678 --json --select id,name,color

# Plain tab-separated for piping
discord-cli guilds members list 123456789012345678 --plain

# Dry run (show request without sending)
discord-cli guilds bans user-from-guild 123456789012345678 USER_ID --dry-run
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Retryable** - creates return "already exists" on retry, deletes return "already deleted"
- **Confirmable** - `--yes` for explicit confirmation of destructive actions
- **Piped input** - `echo '{"key":"value"}' | discord-cli <resource> create --stdin`
- **Cacheable** - GET responses cached for 5 minutes, bypass with `--no-cache`
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set
- **Progress events** - paginated commands emit NDJSON events to stderr in default mode

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Doctor

```bash
discord-cli doctor
```

Validates your bot token, API connectivity, and configuration.

## Configuration

Config file: `~/.config/discord-cli/config.toml`

Environment variables:
- `DISCORD_TOKEN` - Bot token (format: `Bot YOUR_TOKEN`)

## FAQ

**How do I get a bot token?**
1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create an application
3. Go to Bot section, click "Reset Token"
4. Copy the token and set `DISCORD_TOKEN="Bot YOUR_TOKEN"`

**What permissions does my bot need?**
It depends on what you want to do. For read-only operations (list channels, messages), the bot needs basic permissions. For management (bans, role changes), it needs admin permissions. Check Discord's [permission calculator](https://discord.com/developers/docs/topics/permissions).

**Can I use a user token?**
This CLI is designed for bot tokens (TOS-compliant). User tokens may work but violate Discord's Terms of Service and risk account suspension.

**How does rate limiting work?**
The CLI automatically handles Discord's per-route rate limits with retry + exponential backoff. If you see `exit code 7`, you're being rate-limited - wait a few minutes.

**Where is the local database stored?**
`~/.local/share/discord-cli/data.db` (SQLite). Use `--db` flag to specify a custom path.

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `discord-cli doctor` to check credentials
- Verify the environment variable: `echo $DISCORD_TOKEN`
- Make sure the format is `Bot YOUR_TOKEN` (with the "Bot " prefix)

**Not found errors (exit code 3)**
- Check the resource ID is correct (use Developer Mode in Discord to copy IDs)
- Run the `list` command to see available items

**Rate limit errors (exit code 7)**
- The CLI auto-retries with exponential backoff
- If persistent, wait a few minutes and try again

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
