---
title: "Audit: Moonbase CLI"
type: fix
status: active
date: 2026-03-25
---

# Audit: Moonbase CLI

## Command Count
- Generated: 79 command registrations (19 resource groups + auth + doctor + version)
- Target: ~50 endpoints covered
- All endpoints from the spec are represented

## Help Text Quality Assessment
- Descriptions: Clean, developer-friendly (not raw spec jargon)
- Examples: Generic placeholders ("abc123", "--direction value") - need realistic values
- Root command: Good description of what the API does

## Agent-Native Checklist
- [x] --json output
- [x] --csv output
- [x] --plain output
- [x] --quiet output
- [x] --select fields filtering
- [x] --dry-run
- [x] --stdin for POST/PUT bodies
- [x] --yes for confirmation skip
- [x] --no-cache
- [x] --no-color
- [x] --human-friendly
- [x] --all for auto-pagination
- [x] Typed exit codes (0, 2, 3, 4, 5, 7, 10)
- [x] Doctor command with health checks
- [x] Auth management (status, set-token, logout)
- [x] Local response caching
- [x] Non-interactive by default

## First Steinberger Analysis (Baseline)

Automated scorecard: 90/90 (100%) - Grade A

| Dimension | Score | What 10 Looks Like | Status |
|-----------|-------|-------------------|--------|
| Output modes | 10/10 | --json, --csv, --plain, --quiet, --select | All present |
| Auth | 10/10 | Token storage, env var, status, set-token, logout | All present |
| Error handling | 10/10 | Typed exit codes, error classification, retry | All present |
| Terminal UX | 10/10 | Colors, tabwriter, no-color, human-friendly | All present |
| README | 10/10 | Install, quickstart, every command, agent usage, troubleshooting | All present |
| Doctor | 10/10 | Config check, auth check, API connectivity, env vars | All present |
| Agent-native | 10/10 | --json, --select, --dry-run, --stdin, --yes, --no-cache, typed exits | All present |
| Local Cache | 10/10 | Response cache with --no-cache bypass | Present |
| Breadth | 10/10 | All 50 endpoints covered across 19 resources | Complete |

**Baseline Total: 90/90 (Grade A)**

## GOAT Improvement Plan (UX Polish)

Since all scorecard dimensions are 10/10, improvements are UX quality (not measured by scorecard):

1. **Realistic examples in help text** - Replace "abc123" with realistic Moonbase values (collection refs like "people", "organizations", real-looking UUIDs)
2. **--stdin examples for complex body commands** - Add realistic JSON examples for items create, calls create, messages create, webhooks create, items search
3. **Command example polish** - Show common workflows (list people, search organizations, send program message)

## Complex Body Field Plan
No fields were skipped by the generator. All body fields are handled via flags or --stdin. However, several commands would benefit from --stdin examples with realistic JSON in their help text:

1. **items create** - needs example showing values map structure
2. **items search** - needs example showing filter JSON structure
3. **calls create** - needs example showing participants array
4. **messages create** - needs example showing to/cc/bcc recipients
5. **webhooks create** - needs example showing subscriptions array
