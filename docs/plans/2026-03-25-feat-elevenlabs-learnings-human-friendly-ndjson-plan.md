---
title: "ElevenLabs Learnings - --human-friendly Flag and NDJSON Event Streaming"
type: feat
status: active
date: 2026-03-25
---

# ElevenLabs Learnings

ElevenLabs just shipped their CLI as agent-first (v0.5.0, March 25 2026). Two patterns worth stealing.

## What They Did That We Should Copy

### 1. `--human-friendly` flag (color opt-in instead of auto-detect)

**ElevenLabs pattern:** Default output is plain text. `--human-friendly` enables the rich Ink/React TUI.

**Our current behavior:** Colors are auto-detected via `isatty`. If an agent runs in a PTY (which Claude Code and many agent frameworks do), it gets color escape codes it doesn't want.

**Fix:** Add `--human-friendly` flag. When NOT set: suppress colors entirely (override colorEnabled to false). When set: use the current isatty detection. This makes the default truly agent-safe even in PTY environments.

### 2. NDJSON event streaming in --json mode

**ElevenLabs gap (they don't have this but should):** Their agent mode has no progress indication. Long-running commands produce nothing until they finish.

**Our opportunity:** When `--json` is set on long-running commands (pagination, bulk operations), emit one JSON line per event as it happens:

```json
{"event":"page_fetch","page":1,"count":50}
{"event":"page_fetch","page":2,"count":50}
{"event":"complete","total":100,"duration_ms":2340}
```

Agents can process events in real-time instead of waiting for the final blob.

## Implementation Units

### Unit 1: Add --human-friendly flag

**File:** `internal/generator/templates/root.go.tmpl`

Add flag:
```go
rootCmd.PersistentFlags().BoolVar(&flags.humanFriendly, "human-friendly", false, "Enable colored output and rich formatting")
```

Add to rootFlags struct:
```go
humanFriendly bool
```

**File:** `internal/generator/templates/helpers.go.tmpl`

Change `colorEnabled()` to check the flag:
```go
func colorEnabled() bool {
    if noColor {
        return false
    }
    if !humanFriendly {
        return false  // default: no color (agent-safe)
    }
    // human-friendly mode: use isatty detection
    if os.Getenv("NO_COLOR") != "" {
        return false
    }
    if os.Getenv("TERM") == "dumb" {
        return false
    }
    return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}
```

Need to make `humanFriendly` accessible to `colorEnabled()`. Simplest: make it a package-level var like `noColor`:
```go
var humanFriendly bool
```

And register it alongside `noColor` in root.go.tmpl.

**Verification:** Generate petstore CLI. Default `--help` output has no ANSI escape codes. With `--human-friendly`, colored output appears on TTYs.

### Unit 2: NDJSON event streaming for pagination

**File:** `internal/generator/templates/helpers.go.tmpl`

In `paginatedGet`, when fetching multiple pages and `--json` is set, emit NDJSON progress events to stderr (not stdout - stdout is for the final result):

```go
if asJSON {
    fmt.Fprintf(os.Stderr, `{"event":"page_fetch","page":%d,"items":%d}`+"\n", page, len(items))
}
```

At completion:
```go
if asJSON {
    fmt.Fprintf(os.Stderr, `{"event":"complete","total":%d,"pages":%d}`+"\n", len(allItems), page)
}
```

Events go to stderr so they don't corrupt the JSON array on stdout. Agents that want progress events read stderr. Agents that just want data read stdout.

**Verification:** Generate petstore CLI. Run a paginated command with --json. Stderr shows NDJSON events. Stdout has clean JSON array.

### Unit 3: Update README template

**File:** `internal/generator/templates/readme.md.tmpl`

Add to the Agent Usage section:
```markdown
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set
- **Progress events** - paginated commands emit NDJSON events to stderr in `--json` mode
```

### Unit 4: Update scorecard

**File:** `internal/pipeline/scorecard.go`

In `scoreAgentNative`, add a check for `--human-friendly`:
```go
if strings.Contains(combined, "human-friendly") || strings.Contains(combined, "humanFriendly") {
    score += 1
}
```

Reduce another dimension by 1 point to keep the cap at 10.

## Acceptance Criteria

- [ ] Default output has zero ANSI escape codes (no color, no bold)
- [ ] `--human-friendly` enables color on TTYs
- [ ] Paginated --json commands emit NDJSON events to stderr
- [ ] Stdout in --json mode is clean JSON (no events mixed in)
- [ ] README documents both patterns
- [ ] `go test ./...` passes
- [ ] 10-API gauntlet still passes 10/10

## Scope Boundaries

- Do NOT add a full TUI (bubbletea/lipgloss) - just control color
- Do NOT change the table output format - tables stay plain text
- NDJSON is stderr only - never stdout
- Keep `--no-color` as a backwards-compatible alias
