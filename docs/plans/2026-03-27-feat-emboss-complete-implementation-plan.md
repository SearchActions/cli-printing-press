---
title: "Complete the Emboss Command"
type: feat
status: completed
date: 2026-03-27
origin: docs/plans/2026-03-27-feat-printing-press-repress-mode-plan.md
---

# Complete the Emboss Command

## Overview

The emboss command (`printing-press emboss`) was partially built. The `--audit-only` mode works (runs verify + scorecard, prints baseline). But the full cycle - save baseline, make improvements, re-audit, compute delta, write report - is incomplete. This plan finishes it.

## What Exists (Already Built)

- `internal/cli/emboss.go` - EmbossReport, EmbossSnapshot, EmbossDelta structs. `--audit-only` runs verify + scorecard and prints baseline. Non-audit mode just prints "run the skill."
- `skills/printing-press/SKILL.md` - Emboss Mode section with 6-step cycle instructions.
- `printing-press emboss --dir ./github-cli --spec spec.json --audit-only` - tested and working, produces correct baseline (68/100, 96% verify, pipeline FAIL).

## What's Missing

### M1. Save baseline to disk (`--save-baseline`)

Currently the audit runs and prints to stdout but doesn't persist. The emboss cycle needs to save the "before" snapshot so it can be compared to the "after" snapshot later (possibly in a different session).

**File:** `internal/cli/emboss.go`

**Change:** Add `--save-baseline` flag. When set, write the EmbossReport as JSON to `<cli-dir>/.emboss-baseline.json`. This file persists between sessions.

```go
if saveBaseline {
    data, _ := json.MarshalIndent(report, "", "  ")
    baselinePath := filepath.Join(dir, ".emboss-baseline.json")
    os.WriteFile(baselinePath, data, 0644)
    fmt.Fprintf(os.Stderr, "Baseline saved to %s\n", baselinePath)
}
```

**Verification:** `printing-press emboss --dir ./github-cli --spec spec.json --audit-only --save-baseline` creates `./github-cli/.emboss-baseline.json` with valid JSON.

### M2. Load baseline and compute delta (`--compare`)

After improvements are made (by the skill), run emboss again. It should automatically detect a saved baseline, re-audit, and compute the delta.

**File:** `internal/cli/emboss.go`

**Change:** At the start of RunE, check if `<cli-dir>/.emboss-baseline.json` exists. If it does:
1. Load it as the "before" snapshot
2. Run a fresh audit as the "after" snapshot
3. Compute the delta (after - before for each metric)
4. Report all three: before, after, delta

```go
baselinePath := filepath.Join(dir, ".emboss-baseline.json")
if data, err := os.ReadFile(baselinePath); err == nil {
    var baseline EmbossReport
    json.Unmarshal(data, &baseline)
    report.Before = baseline.Before
    // Current audit becomes "after"
    report.After = &currentSnapshot
    report.Delta = computeDelta(report.Before, *report.After)
}
```

**Verification:** After saving a baseline and making changes, `printing-press emboss --dir ./github-cli --spec spec.json --audit-only` shows before/after/delta.

### M3. Write delta report to docs/plans/

After computing the delta, write a markdown report.

**File:** `internal/cli/emboss.go`

**Change:** When delta is computed, write `docs/plans/<today>-emboss-<api>-cli-delta.md` with:
- Before/after table
- List of improvements (from commit messages since baseline timestamp)
- Remaining gaps

**Verification:** Delta report file exists in docs/plans/ after an emboss with compare.

### M4. Clean up baseline after report

After the delta report is generated, delete `.emboss-baseline.json` so the next emboss cycle starts fresh.

**File:** `internal/cli/emboss.go`

**Verification:** `.emboss-baseline.json` is removed after delta report is written.

### M5. Full mode without `--audit-only`

Currently the non-audit mode prints "run the skill." It should instead:
1. Save baseline automatically
2. Print the baseline
3. Print instructions for the skill-driven steps (2-4)
4. Tell the user to re-run `printing-press emboss --dir ./X --spec spec.json` when done to get the delta

This is a UX improvement, not a logic change. The actual improvements are always skill-driven.

**File:** `internal/cli/emboss.go`

**Verification:** `printing-press emboss --dir ./github-cli --spec spec.json` (without --audit-only) saves baseline and prints clear instructions.

## Implementation Units

### Unit 1: Save and load baseline (M1 + M2)

**Goal:** Persist baseline to disk, detect it on next run, compute delta.

**Files:**
- `internal/cli/emboss.go`

**Approach:**
- Add `computeDelta(before, after EmbossSnapshot) *EmbossDelta` function
- Check for `.emboss-baseline.json` at start of RunE
- If found: load as "before", current audit as "after", compute delta
- Add `--save-baseline` flag
- Remove `--audit-only` requirement for saving (full mode auto-saves)

**Verification:**
- `emboss --audit-only --save-baseline` creates `.emboss-baseline.json`
- Second `emboss --audit-only` detects baseline and shows delta
- Delta math is correct (after - before)

### Unit 2: Delta report + cleanup (M3 + M4)

**Goal:** Write a markdown delta report and clean up the baseline file.

**Files:**
- `internal/cli/emboss.go`

**Approach:**
- After computing delta, write `docs/plans/<today>-emboss-<name>-delta.md`
- Include before/after table, delta values, timestamp
- Delete `.emboss-baseline.json` after report is written
- Add `--keep-baseline` flag to skip cleanup (for repeated emboss cycles)

**Verification:**
- Delta report file appears in docs/plans/ with correct content
- `.emboss-baseline.json` is deleted after report
- `--keep-baseline` preserves it

### Unit 3: Full mode UX (M5)

**Goal:** Make non-audit mode useful.

**Files:**
- `internal/cli/emboss.go`

**Approach:**
- Full mode (no --audit-only): auto-save baseline, print baseline, print step-by-step instructions
- Instructions tell user: "Now run `/printing-press emboss <dir>` in Claude Code for Steps 2-4. When done, run `printing-press emboss --dir <dir> --spec <spec>` again to get the delta."

**Verification:**
- `emboss --dir ./X --spec spec.json` without --audit-only prints clear workflow instructions and saves baseline

## Acceptance Criteria

- [x] `printing-press emboss --dir ./X --spec spec.json` saves baseline and prints instructions
- [x] `printing-press emboss --dir ./X --spec spec.json --audit-only --save-baseline` persists baseline to disk
- [x] Second run detects saved baseline and computes before/after delta
- [x] Delta report written to `docs/plans/<today>-emboss-<name>-delta.md`
- [x] Baseline cleaned up after delta report (unless --keep-baseline)
- [x] `go build && go vet` passes
- [x] Tested against github-cli with real baseline data

## Sources

- Origin plan: `docs/plans/2026-03-27-feat-printing-press-repress-mode-plan.md`
- Existing code: `internal/cli/emboss.go` (175 lines, partially implemented)
- Pattern reference: `internal/cli/verify.go` (same report + JSON output pattern)
