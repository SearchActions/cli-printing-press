---
title: "Audit: Discord CLI"
type: fix
status: active
date: 2026-03-25
---

# Audit: Discord CLI

## Automated Scorecard Baseline
90/90 (100%) - Grade A. All dimensions at 10/10.

## Manual Code Review Findings

### 1. Example values are placeholder garbage
All examples use "abc123" instead of realistic Discord snowflake IDs (18-digit numbers like "1234567890123456789"). Webhook tokens should also be realistic.

### 2. Short descriptions are too terse
"Create" should be "Send a message to a channel". "Get" should be "Get a channel by ID". The descriptions are just the HTTP verb, not developer-friendly.

### 3. Root command description is raw spec text
"Preview of the Discord v10 HTTP API specification" - should be "Manage Discord servers, channels, messages, and more from the command line"

### 4. Complex body fields lack --stdin examples
79 complex fields were skipped. Key endpoints that NEED --stdin examples:
- `channels messages create` - embeds, components, attachments, sticker_ids
- `webhooks execute` - embeds, components, attachments
- `guilds channels create` - permission_overwrites
- `applications commands create` - options, integration_types

### 5. Flag descriptions are raw spec jargon
"Tts" should be "Send as text-to-speech message". "Enforce nonce" should be "Enforce nonce uniqueness for this message".

## First Steinberger Analysis (Baseline)

| Dimension | Score | What 10 Looks Like | How to Get There |
|-----------|-------|-------------------|-----------------|
| Output modes | 10/10 | gogcli: --json, --csv, --plain, --quiet, --select | Already has all output modes |
| Auth | 10/10 | gogcli: env var, config file, doctor validates | Present |
| Error handling | 10/10 | gogcli: typed exits, classifyAPIError | Present |
| Terminal UX | 10/10 | gogcli: colors, tabwriter, no-color flag | Present but descriptions are terse |
| README | 10/10 | gogcli: install, quickstart, every command, agent usage | Structure good, examples weak |
| Doctor | 10/10 | gogcli: validates auth, version, config | Present |
| Agent-native | 10/10 | gogcli: --json, --select, --dry-run, --stdin, --yes, typed exits | Complete |
| Local Cache | 10/10 | gogcli: 5-min cache, --no-cache bypass | Present |
| Breadth | 10/10 | gogcli: covers full API | 307 commands from full spec |

**Baseline Total: 90/90 (Grade A)**

## GOAT Improvement Plan (UX Polish - Not Scorecard-Measured)

1. Fix root command description to be developer-friendly
2. Fix all example values from "abc123" to realistic Discord snowflake IDs
3. Add --stdin examples for top 3 complex body field endpoints
4. Improve README with realistic examples and a cookbook section
5. Fix terse flag descriptions for the most important commands

## Complex Body Field Plan

### channels messages create
```bash
echo '{"content":"Hello!","embeds":[{"title":"My Embed","description":"Rich content","color":5814783}]}' | discord-cli channels messages create 1234567890123456789 --stdin
```

### webhooks execute
```bash
echo '{"content":"Deploy complete","embeds":[{"title":"Status","color":3066993,"fields":[{"name":"Version","value":"v2.1.0"}]}]}' | discord-cli webhooks execute 1234567890123456789 WEBHOOK_TOKEN --stdin
```

### applications commands create
```bash
echo '{"name":"greet","description":"Say hello","options":[{"name":"user","type":6,"description":"User to greet","required":true}]}' | discord-cli applications commands create 1234567890123456789 --stdin
```
