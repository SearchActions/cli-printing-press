---
title: "Steinberger Audit: Discord CLI v2"
type: fix
status: active
date: 2026-03-26
phase: "3"
api: "discord"
---

# Steinberger Audit: Discord CLI v2

## Automated Scorecard Baseline: 67/100 (Grade B)

## First Steinberger Analysis (Baseline)

| Dimension | Score | What 10 Looks Like | How to Get There |
|-----------|-------|-------------------|-----------------|
| Output modes | 10/10 | gogcli: --json, --csv, --plain, --quiet, --compact, --select, --no-color, --human-friendly | All present in root.go |
| Auth | 10/10 | gogcli: token storage, doctor validates, env var | Present: DISCORD_TOKEN env, doctor, auth command |
| Error handling | 10/10 | gogcli: typed exits, classifyAPIError, retry hints | Present: typed exit codes (2-10), classifyAPIError with hints |
| Terminal UX | 9/10 | gogcli: progress, color, pager | Progress in sync. Missing: pager for long output |
| README | 3/10 | gogcli: install, quickstart, every command with example, cookbook, FAQ | Bare minimum. Missing: cookbook, FAQ, workflow examples, --stdin examples |
| Doctor | 10/10 | gogcli: validates auth, API version, config health | Present |
| Agent-native | 8/10 | gogcli: --json, --select, --dry-run, --stdin, typed exits, no TTY | All flags present. --stdin needs examples in README |
| Local Cache | 10/10 | gogcli: SQLite, FTS5, sync, search, --no-cache | Present but generic (JSON blob tables, not domain-specific) |
| Breadth | 6/10 | 100+ commands covering every API endpoint + convenience wrappers | 316 generated but many have lazy descriptions ("Get", "Delete") |
| Vision | 9/10 | discrawl: SQLite + FTS5 + sync + search + tail + workflows | Sync, search, tail, export, analytics present as shells |
| Workflows | 4/10 | Compound commands solving real problems | Only generic analytics/workflow shells, no real implementations |
| Insight | 0/10 | Health, stale, audit, modreport commands with real queries | No insight commands exist yet |

**Baseline: 67/100 (Grade B)**

## GOAT Improvement Plan

### Priority 0: Domain-Specific Data Layer (Phase 0.7 spec)
1. **Replace generic store.go** with domain-specific tables (messages, members, channels, audit_log_entries, roles, mentions, reactions)
2. **Add FTS5 on message content** (not generic resources)
3. **Add domain-aware sync** with snowflake ID cursors per channel
4. **Add domain search filters** (--channel, --author, --guild, --days, --since, --until)

### Priority 1: Workflow Commands (7 from Phase 0.5)
1. `health` - Server health report from local DB
2. `audit` - Audit log forensics with filters
3. `watch` - Keyword monitor via Gateway
4. `modreport` - Moderation summary report
5. `permdiff` - Permission diff/snapshot
6. `stale` - Stale channel detection
7. `roleaudit` - Empty/excessive role detection

### Priority 2: Scorecard Gap Fixes
1. README: Add cookbook, FAQ, workflow examples, --stdin examples
2. Lazy descriptions: Fix "Get", "Delete", "Update" to be descriptive
3. Example values: Replace UUIDs with Discord snowflake IDs (e.g., "1234567890123456789")
4. --stdin examples for top 3 complex body endpoints (create message with embeds, create thread, update channel)

### Priority 3: Dead Code Cleanup
1. Audit helpers.go for uncalled functions
2. Verify all root.go flags are read in at least one RunE
3. Remove unused store methods

## Complex Body Field Plan

Top 3 endpoints needing --stdin examples:
1. **POST /channels/{channel_id}/messages** - embeds, components, attachments, sticker_ids
   ```bash
   echo '{"content":"Hello","embeds":[{"title":"Test","description":"Body"}]}' | discord-cli channels messages create 1234567890123456789 --stdin
   ```
2. **POST /channels/{channel_id}/threads** - name, type, auto_archive_duration
   ```bash
   echo '{"name":"Bug Discussion","type":11,"auto_archive_duration":1440}' | discord-cli channels threads create 1234567890123456789 --stdin
   ```
3. **PATCH /channels/{channel_id}** - permission_overwrites, available_tags
   ```bash
   echo '{"name":"renamed-channel","topic":"New topic","nsfw":false}' | discord-cli channels update 1234567890123456789 --stdin
   ```
