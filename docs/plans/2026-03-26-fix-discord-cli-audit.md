---
title: "Steinberger Audit: Discord CLI"
type: fix
status: active
date: 2026-03-26
phase: "3"
api: "discord"
---

# Steinberger Audit: Discord CLI

## Automated Scorecard Baseline

Total: **63/100 (Grade C)**

| Dimension | Score | Notes |
|-----------|-------|-------|
| Output Modes | 10/10 | --json, --csv, --plain, --quiet, --select all present |
| Auth | 10/10 | DISCORD_TOKEN env var, config file, doctor validates |
| Error Handling | 10/10 | Typed exits (2=usage, 3=404, 4=auth, 5=api, 7=ratelimit), hints |
| Terminal UX | 10/10 | Color, tabwriter, pagination progress |
| README | 3/10 | Generic description, no cookbook, no workflow examples |
| Doctor | 10/10 | Validates auth, API connectivity |
| Agent Native | 8/10 | Missing --stdin flag |
| Local Cache | 10/10 | SQLite store present |
| Breadth | 6/10 | 316 commands but many have lazy descriptions |
| Vision | 9/10 | sync/search/tail/export/analytics scaffolding present |
| Workflows | 4/10 | Workflow templates failed to generate (stale/orphans/load missing) |
| Insight | 0/10 | No insight/health/status/activity commands |
| Path Validity | 5/10 | Some paths use UUID format instead of Discord snowflake IDs in examples |
| Auth Protocol | 5/10 | Need to verify Bot prefix in Authorization header |
| Data Pipeline Integrity | 10/10 | Store schema present |
| Sync Correctness | 4/10 | Generic sync, not domain-aware |
| Type Fidelity | 2/5 | `after` param is int instead of string (snowflake) |
| Dead Code | 0/5 | Multiple dead flags and functions likely |

## GOAT Improvement Plan

### Priority 0: Data Layer Foundation (Phase 4)
1. Replace generic store.go with domain-specific schema from Phase 0.7 (messages, members, channels, audit_log, sync_state tables)
2. Rewrite sync command to use snowflake ID cursors, per-channel sync, thread discovery
3. Add FTS5 search with domain filters (--channel, --author, --before, --after, --days)
4. Add `sql` command for raw read-only queries
5. Add `activity` command (cross-entity query: messages x members)
6. Add `audit` command (audit log with action_type filtering)

### Priority 1: Workflow Commands
7. Build `members` command that queries local SQLite
8. Build `channels` command with message count stats
9. Build `stale` command for dead channel detection
10. Build `status` command for archive health

### Priority 2: Scorecard Gap Fixes
11. Fix README: add cookbook section with data layer + workflow examples
12. Fix example IDs: replace UUID format with Discord snowflake format (e.g., "1234567890123456789")
13. Fix `after` param type from int to string (snowflakes are strings)
14. Fix lazy Short descriptions ("List", "Get", "Delete" -> domain-specific)
15. Add --stdin support for complex body fields (embeds, components)
16. Add insight commands (activity, health, status)

### Complex Body Fields Plan
Top 3 endpoints where --stdin matters most:
1. **channels messages create** - needs embeds, components, attachments, sticker_ids
2. **webhooks execute** - needs embeds, components, attachments
3. **guilds channels create** - needs permission_overwrites, available_tags

Example --stdin usage:
```bash
echo '{"content":"Hello","embeds":[{"title":"Test","description":"Example embed"}]}' | discord-cli channels messages create 1234567890123456789 --stdin
```
