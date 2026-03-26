---
title: "Steinberger Audit: Discord CLI"
type: fix
status: active
date: 2026-03-25
---

# Steinberger Audit: Discord CLI

## Command Comparison
- Generated: 316 commands across 16 resource groups
- Target: 50+ (beat jackwener/discord-cli at 15)
- Result: 316 commands - massively exceeds target breadth

## Help Text Quality Assessment
- Root description: "Preview of the Discord v10 HTTP API specification" - NEEDS FIX (raw spec title, should describe what the CLI does)
- Command descriptions: Many are lazy 1-word ("Create", "Get", "Delete") - NEEDS FIX
- Examples: Placeholder values ("abc123") - NEEDS FIX with realistic Discord IDs
- Resource group descriptions: Generic ("Manage channels") - OK but could be better

## Agent-Native Checklist
- [x] --json
- [x] --select
- [x] --dry-run
- [x] --stdin
- [x] --yes
- [x] --no-cache
- [x] doctor
- [x] Typed exit codes (0,2,3,4,5,7,10)
- [x] --csv, --plain, --quiet
- [ ] --stdin example in help text for complex body fields

## Complex Body Fields Plan
Top 3 endpoints needing --stdin examples:
1. **channels messages create** - embeds, components, attachments, sticker_ids skipped. Example: `echo '{"content":"Hello","embeds":[{"title":"Test","color":5814783}]}' | discord-cli channels messages create CHANNEL_ID --stdin`
2. **guilds channels create** - permission_overwrites, available_tags skipped. Example: `echo '{"name":"new-channel","type":0,"permission_overwrites":[{"id":"ROLE_ID","type":0,"allow":"1024"}]}' | discord-cli guilds channels create GUILD_ID --stdin`
3. **webhooks execute** - embeds, components, attachments skipped. Example: `echo '{"content":"Deploy complete","embeds":[{"title":"Build #123","color":3066993}]}' | discord-cli webhooks execute WEBHOOK_ID WEBHOOK_TOKEN --stdin`

## First Steinberger Analysis (Baseline)

| Dimension | Score | What 10 Looks Like | How to Get There |
|-----------|-------|-------------------|-----------------|
| Output modes | 10/10 | --json, --csv, --plain, --quiet, --select | Already at 10 |
| Auth | 10/10 | Token via env var, doctor validates, config file | Already at 10 |
| Error handling | 10/10 | Typed exits, retry with backoff, classifyAPIError | Already at 10 |
| Terminal UX | 8/10 | Progress spinners, color themes, pager for long output | Add progress indicator for sync/tail |
| README | 4/10 | Install, quickstart, every command with example, cookbook, FAQ | Add cookbook section with Discord-specific examples, realistic IDs, complex body examples |
| Doctor | 10/10 | Validates auth, API version, config health | Already at 10 |
| Agent-native | 8/10 | --json, --select, --dry-run, --stdin, idempotent, typed exits | Add --stdin examples in help text |
| Local Cache | 10/10 | SQLite sync + search + FTS5 | Already at 10 |
| Breadth | 6/10 | 100+ commands covering every endpoint + convenience wrappers | 316 commands but descriptions are lazy. Fix Short fields. |
| Vision | 9/10 | Sync + search + tail + export + analytics | Already strong. Add workflow commands to reach 10. |
| Workflows | 4/10 | Compound commands combining 2+ API calls | Build Phase 0.5 workflows |

**Baseline Total: 89/110 (80%) - Grade A**

## GOAT Improvement Plan

### Top 5 Highest-Impact Improvements
1. **Build 7 workflow commands** (Phase 0.5) - channel-health, audit-report, member-report, server-snapshot, prune-preview, webhook-test, message-stats
2. **Fix README** - Add cookbook with Discord-specific examples, complex body --stdin examples, FAQ
3. **Fix root description** - "Discord API CLI - manage servers, channels, messages, roles, and more from your terminal"
4. **Fix lazy Short fields** - "Create" -> "Send a message to a channel", "Get" -> "Get channel details by ID"
5. **Add --stdin examples** - Top 3 complex body endpoints in help text

### Commands to ADD (Workflow commands)
1. `discord-cli channel-health` - Stale channel detection + activity histogram
2. `discord-cli audit-report` - Audit log analysis grouped by action/user/date
3. `discord-cli member-report` - Member activity, role distribution, top contributors
4. `discord-cli server-snapshot` - Backup guild config to JSON
5. `discord-cli prune-preview` - Preview member prune before executing
6. `discord-cli webhook-test` - Send test payloads to webhooks
7. `discord-cli message-stats` - Per-channel message volume and activity trends
