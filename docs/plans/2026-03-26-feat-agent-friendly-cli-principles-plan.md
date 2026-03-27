---
title: "Apply Agent-Friendly CLI Principles to Printing Press"
type: feat
status: active
date: 2026-03-26
---

# Apply Agent-Friendly CLI Principles to Printing Press

## Overview

Trevin Chow's "7 Principles for Agent-Friendly CLIs" (@trevin on X, 2026-03-26) identifies the gap between "a CLI that technically works" and "a CLI optimized for agent consumers." The printing-press already generates CLIs with agent-native flags (--json, --select, --dry-run, --stdin, --yes), but the implementation is shallow - flags are declared but often unwired, errors are vague, and output is unbounded.

This plan upgrades the generator templates and scorecard to produce CLIs that are genuinely agent-first, not just agent-decorated.

## Problem Statement

The printing-press generates CLIs that look agent-friendly on the surface but fail Trevin's rubric at the Friction and Blocker levels:

| Principle | Trevin's Severity | Our Status |
|---|---|---|
| Non-interactive by default | Blocker if hanging | Auth flow can hang. No global --no-input. |
| Structured output | Blocker if missing | Flags exist but often unwired (dead flags). |
| Actionable errors | Friction | Generated errors say "missing required arguments" not "--content is required". |
| Safe retries | Blocker for mutations | No idempotence. Retry of create silently duplicates. |
| Progressive help | Friction | No Examples section in generated --help. Placeholder values. |
| Composable structure | Friction | --stdin dead flag. Inconsistent naming not detected. |
| Bounded responses | Friction | No truncation hints. No suggested narrowing. Unbounded default lists. |

## Proposed Solution

Three layers of changes, ordered by impact:

### Layer 1: Template Fixes (highest impact, generates better CLIs)

**1a. Fix error messages in command_endpoint.go.tmpl**

Current: `return usageErr(fmt.Errorf("channel_id is required"))`
After: `return usageErr(fmt.Errorf("--channel-id is required\nUsage: %s %s <channel_id>\nExample: %s %s 550e8400-e29b-41d4-a716-446655440000", rootName, cmd.Use, rootName, cmd.Use))`

The generated error should include: the specific missing flag/arg, the correct usage pattern, and a concrete example.

**1b. Add idempotence hints to mutation commands**

For POST commands, the generated RunE should check the response for "already exists" (HTTP 409) and return success with a message instead of an error. The `classifyAPIError` helper already handles 409 - wire it to print "Already exists, no changes" instead of treating it as failure.

For commands with `--dry-run`, the dry-run output should include the full request that would be sent, not just the URL.

**1c. Add truncation hints to list commands**

When a paginated GET returns results, add a footer line:
```
Showing 50 of 312 results.
To narrow: <cli> <resource> list --limit 10 --after <last_id>
```

This teaches agents how to paginate or filter instead of dumping everything.

**1d. Add Examples section to command_endpoint.go.tmpl**

Currently the `Example:` field in generated cobra commands is a single line with placeholder values. Change to 2-3 examples showing common use patterns:
```go
Example: `  # List with JSON output
  <cli> <resource> list --json

  # Get a specific resource
  <cli> <resource> get 550e8400-e29b-41d4-a716-446655440000

  # Pipe to jq for field selection
  <cli> <resource> list --json --select id,name | jq -r '.[].id'`,
```

**1e. Wire all declared flags**

The dogfood command already catches dead flags. But the fix is in the templates: every flag declared in root.go.tmpl must have a corresponding check in the command templates. Specifically:
- `--quiet`: suppress table output, exit code only
- `--csv`: call outputCSV() in table-rendering commands
- `--stdin`: read from os.Stdin in create/update commands
- `--no-cache`: bypass local SQLite, hit API directly
- `--select`: already wired in helpers.go via filterFields

### Layer 2: Scorecard Additions (measures what Trevin measures)

Add a new scorecard dimension: **AgentReadiness** (0-10)

| Check | Points | Maps to Trevin Principle |
|---|---|---|
| No TTY prompts in non-interactive mode (grep for `bufio.Scanner` gated by `isatty`) | 2 | 1. Non-interactive |
| Error messages include flag name (not just "missing required arguments") | 2 | 3. Actionable errors |
| 409 handling returns success-like message | 1 | 4. Safe retries |
| List commands include truncation hint ("Showing N of M") | 2 | 7. Bounded responses |
| --help includes Examples section with 2+ examples | 2 | 5. Progressive help |
| Consistent flag naming across subcommands (--limit not sometimes --max-results) | 1 | 6. Composable structure |

This dimension replaces or augments the existing AgentNative dimension which only checks flag presence, not usage quality.

### Layer 3: Skill Update (guides Claude to generate agent-first)

Add to SKILL.md Phase 4 instructions:

```markdown
**Agent-Friendly CLI Checklist (apply to every generated command):**

For each command, verify:
1. Can an agent run this without stdin? (no hanging prompts)
2. Does --json produce valid, parseable JSON? (not mixed with stderr)
3. Do errors name the specific flag/arg that's wrong?
4. Do mutation commands handle "already exists" gracefully?
5. Do list commands show "Showing N of M" with narrowing hints?
6. Does --help include 2+ concrete examples?
7. Are flag names consistent across all subcommands?

Reference: Trevin Chow's "7 Principles for Agent-Friendly CLIs" (2026)
```

## Acceptance Criteria

- [ ] Error messages in generated commands include the specific flag name and usage example
- [ ] 409 responses on POST commands return success with "already exists" message
- [ ] List commands include "Showing N of M" footer with pagination hint
- [ ] --help on generated commands includes Examples section with 2+ examples
- [ ] All root.go flags (--quiet, --csv, --stdin) are wired in at least one command template
- [ ] New AgentReadiness scorecard dimension scores generated CLIs
- [ ] Dogfood command catches unwired flags (already done in Phase 4)
- [ ] SKILL.md includes agent-friendly checklist

## Success Metrics

- Regenerated discord-cli passes Trevin's 7-principle rubric with 0 Blockers, <= 2 Friction
- AgentReadiness scorecard dimension >= 7/10 on regenerated CLIs
- Dead flag count: 0 (enforced by dogfood)

## What We Already Have (from the v2 overhaul)

- `--json`, `--select`, `--dry-run`, `--stdin`, `--yes`, `--csv`, `--quiet` flags (Phase 1-3)
- Dead flag detection via `printing-press dogfood` (Phase 4)
- Phase 4.6 hallucination audit that catches unwired flags (Phase 6 skill update)
- Format-aware auth that respects spec securitySchemes (Phase 3)
- Domain-specific error types in helpers.go template (existing)
- TTY detection in helpers.go template (existing)

The gap isn't in having the flags - it's in making them actually work and making the generated output agent-consumable.

## Sources

- Trevin Chow (@trevin): "7 Principles for Agent-Friendly CLIs" (X Article, 2026-03-26)
- Anthropic: "Building effective agents" tool design guidance
- CLI Interface Guidelines: https://clig.dev/
- Printing-press v2 overhaul: commits bba34c42..8c92c19d on feat/honest-scorecard-tier2
- Discrawl comparison: docs/plans/2026-03-26-feat-discord-cli-vs-discrawl-analysis-plan.md
