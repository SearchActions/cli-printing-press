---
title: "Agent-Native CLI Audit - Embed Best Practices into the Press"
type: feat
status: active
date: 2026-03-25
---

# Agent-Native CLI Audit - Embed Best Practices into the Press

## Overview

Audit every template in the printing press against the "Building CLIs for agents" checklist. What we have, what's missing, and what template changes make every future generated CLI agent-native by default.

## The Audit

### 1. Non-interactive (no prompts that block agents)

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| No interactive prompts in main commands | command.go.tmpl | PASS | No bufio, survey, or readline imports. All inputs are flags. |
| OAuth flow has browser redirect, not interactive | auth.go.tmpl | PASS | Uses browser redirect with HTTP callback server, no terminal prompts |
| Missing flags error immediately, don't prompt | command.go.tmpl | PASS | Cobra handles missing required flags with usage error |

**Verdict: PASS.** Generated CLIs are already non-interactive. No template changes needed.

### 2. Progressive help discovery (don't dump all docs)

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| Root --help shows only subcommands | root.go.tmpl | PASS | Cobra's default behavior - lists subcommands with one-line descriptions |
| Each subcommand has its own --help | command.go.tmpl | PASS | Cobra auto-generates --help for every command |
| Subcommand --help is concise | command.go.tmpl | PARTIAL | Short description exists but Long description is missing |

**Verdict: PASS.** Cobra handles this naturally. No changes needed.

### 3. --help includes examples

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| Every subcommand --help has examples | command.go.tmpl | PASS | `Example: "{{exampleLine ...}}"` on every endpoint command |
| Examples show real flag usage | generator.go exampleLine() | PARTIAL | Shows positional args and some flags, but doesn't show realistic values |
| Examples show common workflows | readme.md.tmpl | PASS | README has Output Formats and Agent Usage sections |

**Gap: Examples use placeholder values.** `exampleLine` generates `mycli users get <id>` but not `mycli users get usr_123`. Agents would benefit from realistic-looking example values.

**Fix:** Enhance `exampleLine()` in generator.go to use type-appropriate example values (e.g. `"usr_123"` for ID params, `"2026-01-01"` for date params, `"active"` for status params).

### 4. Accept flags and stdin for everything

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| All inputs are flags | command.go.tmpl | PASS | Every parameter becomes a Cobra flag |
| Supports stdin piping | command.go.tmpl | FAIL | No --stdin flag, no os.Stdin reading |
| Output is pipeable | helpers.go.tmpl | PASS | --json output to stdout, errors to stderr |

**Gap: No --stdin support.** Agents can't pipe data into commands. `cat body.json | mycli create --stdin` doesn't work.

**Fix:** Add `--stdin` flag to POST/PUT/PATCH commands in command.go.tmpl that reads JSON body from stdin when present.

### 5. Fail fast with actionable errors

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| Missing flags error immediately | command.go.tmpl | PASS | Cobra errors on missing required flags |
| Error messages include correct invocation | helpers.go.tmpl | PASS | Error hints added: "run doctor to verify auth", "run list to see items" |
| Error messages suggest next command | helpers.go.tmpl | PASS | 401 -> "run doctor", 404 -> "run list" |
| Typed exit codes | helpers.go.tmpl | PASS | 6 distinct exit codes (2-10) |

**Verdict: PASS.** This is one of our strongest areas (scored 10/10 on Steinberger).

### 6. Idempotent commands

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| Create commands handle "already exists" | command.go.tmpl | FAIL | No idempotency check - will create duplicates |
| Update commands are naturally idempotent | command.go.tmpl | PASS | PUT/PATCH are idempotent by HTTP semantics |
| Delete commands handle "already deleted" | command.go.tmpl | FAIL | Will error on 404 instead of returning "already deleted, no-op" |

**Gap: No idempotency handling.** POST commands will create duplicates on retry. DELETE commands will error on already-deleted resources.

**Fix:** In the error handling for POST (409 Conflict -> "already exists, no-op") and DELETE (404 -> "already deleted, no-op"). Add these to `classifyAPIError` in helpers.go.tmpl.

### 7. --dry-run for destructive actions

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| --dry-run flag exists | root.go.tmpl | PASS | Global `--dry-run` flag on all commands |
| --dry-run shows what would happen | client.go.tmpl | PASS | Shows method, URL, headers (masked), body |
| --dry-run prevents execution | client.go.tmpl | PASS | Returns before making HTTP call |

**Verdict: PASS.** --dry-run is excellent (scored 9/10 on Steinberger agent-native dimension).

### 8. --yes / --force to skip confirmations

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| Destructive commands ask for confirmation | command.go.tmpl | FAIL | No confirmation on DELETE commands |
| --yes/--force bypasses confirmation | root.go.tmpl | FAIL | No --yes or --force flag |

**Gap: No confirmation on destructive actions, no --force flag.** This is a non-issue for agents (they WANT no confirmation) but bad for humans. The fix is: add confirmation for DELETE commands that agents bypass with --yes.

**Fix:** Add `--yes` flag to root.go.tmpl. In command.go.tmpl, DELETE commands check `!flags.yes` and print "are you sure? use --yes to skip" if stdin is a TTY.

### 9. Predictable command structure (resource + verb)

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| Pattern: `cli resource verb` | command.go.tmpl | PASS | `plaid-cli transactions list`, `plaid-cli accounts get` |
| Consistent across resources | generator.go | PASS | Every resource gets the same verb pattern from its endpoints |
| CRUD aliases | command.go.tmpl | PASS | CRUD aliases were added (commit 5b3d281) |

**Verdict: PASS.** This is structural to how the generator works.

### 10. Return structured data on success

| Principle | Template | Status | Evidence |
|-----------|----------|--------|----------|
| --json returns structured data | helpers.go.tmpl | PASS | printOutput with --json flag |
| --select filters fields | helpers.go.tmpl | PASS | filterFields with --select flag |
| Success output includes IDs/URLs | helpers.go.tmpl | PARTIAL | Returns whatever the API returns, doesn't add metadata |
| No emoji/decoration in machine output | helpers.go.tmpl | PASS | --json mode has zero decoration |

**Gap: No metadata enrichment.** When you create a resource, the CLI returns the API response as-is. It doesn't add the request URL, duration, or other context that agents find useful.

**Low priority** - the API response usually contains everything needed. Could add `--verbose` later.

## Summary Scorecard

| Principle | Status | Priority |
|-----------|--------|----------|
| Non-interactive | PASS | - |
| Progressive help | PASS | - |
| Examples in --help | PARTIAL | Medium - improve example values |
| Stdin support | FAIL | High - agents need pipelines |
| Fail fast + actionable errors | PASS | - |
| Idempotent commands | FAIL | Medium - retry safety |
| --dry-run | PASS | - |
| --yes/--force | FAIL | Medium - human safety + agent bypass |
| Predictable structure | PASS | - |
| Structured output | PASS | - |

**Score: 7/10 pass, 3 gaps to fix.**

## Implementation Units

### Unit 1: Add --stdin Support to Write Commands

**File:** `internal/generator/templates/command.go.tmpl`

Add a `--stdin` flag to POST/PUT/PATCH commands. When set, read the request body from stdin as JSON instead of assembling it from individual flags.

```go
// In the command template for POST/PUT/PATCH methods
if stdinFlag {
    body, err := io.ReadAll(os.Stdin)
    if err != nil {
        return fmt.Errorf("reading stdin: %w", err)
    }
    var jsonBody map[string]any
    if err := json.Unmarshal(body, &jsonBody); err != nil {
        return fmt.Errorf("parsing stdin JSON: %w", err)
    }
    // Use jsonBody instead of flag-assembled body
}
```

**Also add to root.go.tmpl:** Import `"io"` conditionally.

**Verification:** Generate petstore CLI, verify `echo '{"name":"Fido","status":"available"}' | petstore-cli pet create --stdin` works.

### Unit 2: Add Idempotency Handling

**File:** `internal/generator/templates/helpers.go.tmpl`

Enhance `classifyAPIError` to handle idempotency cases:

```go
case strings.Contains(msg, "HTTP 409"):
    return fmt.Errorf("already exists (no-op)")  // idempotent success
case strings.Contains(msg, "HTTP 404") && isDeleteCommand:
    return fmt.Errorf("already deleted (no-op)")  // idempotent success
```

For 409 Conflict: return exit code 0 (success) with a message. The resource already exists - the agent's intent was fulfilled.

For 404 on DELETE: same - the resource is already gone.

**Verification:** Generate CLI, observe that 409 on create returns success, 404 on delete returns success.

### Unit 3: Add --yes Flag for Destructive Commands

**Files:** `root.go.tmpl`, `command.go.tmpl`

Add `--yes` bool flag to rootFlags. For DELETE commands in command.go.tmpl, add a confirmation gate:

```go
if !flags.yes && isatty.IsTerminal(os.Stdin.Fd()) {
    fmt.Fprintf(os.Stderr, "Delete %s? Use --yes to skip confirmation.\n", resourceID)
    return fmt.Errorf("confirmation required (use --yes to bypass)")
}
```

Agents pass `--yes`. Humans get a safety prompt. Non-TTY (piped) environments skip confirmation automatically.

**Verification:** Generate CLI, verify DELETE without --yes errors with hint, DELETE with --yes proceeds.

### Unit 4: Improve Example Values in --help

**File:** `internal/generator/generator.go` (exampleLine function)

Enhance `exampleLine()` to generate realistic example values based on parameter types:

| Param Type | Param Name Pattern | Example Value |
|-----------|-------------------|---------------|
| string | *_id, *Id | `"usr_abc123"` |
| string | email | `"user@example.com"` |
| string | name | `"My Resource"` |
| string | status | `"active"` |
| string | date, *_at | `"2026-01-01"` |
| integer | limit | `25` |
| integer | page | `1` |
| boolean | * | (omit from example - default is fine) |

Current: `plaid-cli transactions list`
After: `plaid-cli transactions list --start-date 2026-01-01 --end-date 2026-03-25 --count 100`

**Verification:** Generate petstore CLI, check `petstore-cli pet get --help` shows a realistic example.

### Unit 5: Add Agent-Native Scorecard Dimension

**File:** `internal/pipeline/scorecard.go`

Update `scoreAgentNative` to also check for the new agent patterns:

- +1 for --stdin support (grep for "stdin" in command files)
- +1 for --yes flag (grep for "yes" in root.go)
- +1 for idempotency handling (grep for "409" or "already exists" in helpers.go)

This makes the scorecard actually measure agent-nativeness, not just the presence of --json and --dry-run.

### Unit 6: Update README Template with Agent Section

**File:** `internal/generator/templates/readme.md.tmpl`

The Agent Usage section already exists but is minimal. Expand it with the patterns from the checklist:

```markdown
## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Retryable** - `--yes` skips confirmations, idempotent creates/deletes
- **Piped input** - `echo '{}' | mycli create --stdin` for complex bodies

Exit codes: 0 success, 2 usage, 3 not-found, 4 auth, 5 api, 7 rate-limit, 10 config
```

## Acceptance Criteria

- [ ] --stdin reads JSON body from stdin for POST/PUT/PATCH commands
- [ ] 409 Conflict on create returns "already exists (no-op)" with exit 0
- [ ] 404 on delete returns "already deleted (no-op)" with exit 0
- [ ] --yes flag bypasses confirmation on DELETE commands
- [ ] DELETE without --yes on a TTY shows confirmation hint
- [ ] Example values in --help are realistic (IDs, dates, emails) not just "value"
- [ ] Scorecard agent-native dimension checks for stdin, --yes, idempotency
- [ ] README Agent Usage section documents all patterns
- [ ] `go test ./...` passes
- [ ] Full run scorecard shows improvement on agent-native dimension

## Scope Boundaries

- Do NOT add interactive prompts (that defeats the purpose)
- Do NOT add --watch or long-running modes (agents don't watch)
- Do NOT add --verbose metadata enrichment (low priority, API responses are enough)
- Idempotency is handled at the error classification level, NOT at the HTTP level (no conditional requests)
