---
title: "Scorecard Accuracy Improvements"
type: fix
status: proposed
date: 2026-03-27
---

# Problem: Scorecard Produces Inaccurate Scores for Non-Trivial CLIs

## Summary

The printing-press scorecard (`internal/pipeline/scorecard.go`) uses static file-pattern analysis to score generated CLIs on a 0-100 scale across 18 dimensions. For simple CLIs where the generator's output matches expected file names and patterns, this works. For CLIs that have been through a GOAT build phase — where hand-written workflow commands, renamed files, and domain-specific store rewrites are introduced — the scorecard produces scores that diverge significantly from actual quality.

Concrete example: a CLI scored **57/100 (Grade C)** on the scorecard while simultaneously achieving:
- **91% runtime verify pass rate** (30/33 commands pass help + dry-run + execute)
- **100% live API test pass rate** (auth, bookings, event types, sync, search, SQL, health, stats all verified against the real API)
- **0 agent-readiness blockers** (reviewed by cli-agent-readiness-reviewer)
- **Working FTS5 search** (queries return correct results from synced data)
- **End-to-end data pipeline** (sync → SQLite → FTS5 → agenda/stats/stale/conflicts all working)

The scorecard's score should reflect actual quality. A CLI that passes 91% of runtime tests and all live API tests should not score Grade C.

## Root Causes (6 specific issues)

### Issue 1: `scoreSyncCorrectness()` hardcodes the filename `sync.go`

**File:** `internal/pipeline/scorecard.go:947`
**Code:**
```go
content := readFileContent(filepath.Join(dir, "internal", "cli", "sync.go"))
if content == "" {
    return 0  // <-- returns 0 if sync.go doesn't exist
}
```

**Problem:** The function reads only `internal/cli/sync.go`. If the sync logic lives in any other file (e.g., `channel_workflow.go`, `sync_cmd.go`, or a dedicated `sync/` package), the function returns 0 regardless of whether sync is implemented correctly. In the GOAT build, the generator creates sync logic in `channel_workflow.go` and the emboss pass adds a `sync_cmd.go` — neither matches the hardcoded filename.

**Impact:** 0/10 on sync_correctness even though the CLI has a working incremental sync with `afterUpdatedAt` cursor, `GetSyncState`/`SetSyncState`, and pagination — all verified via live API testing.

**Fix:** Search all `.go` files in `internal/cli/` for sync patterns instead of reading a single hardcoded file:
```go
func scoreSyncCorrectness(dir string) int {
    cliDir := filepath.Join(dir, "internal", "cli")
    allContent := readAllGoFiles(cliDir)  // concatenate all .go files
    if allContent == "" {
        return 0
    }
    // ... rest of scoring logic unchanged
}
```

---

### Issue 2: `scoreDataPipelineIntegrity()` hardcodes the filename `sync.go` for write-path detection

**File:** `internal/pipeline/scorecard.go:915`
**Code:**
```go
syncContent := readFileContent(filepath.Join(dir, "internal", "cli", "sync.go"))
// ...
if domainUpsertRe.MatchString(syncContent) {  // only checks sync.go
    score += 3
}
```

**Problem:** Same filename hardcoding as Issue 1. The function checks for `UpsertBooking`, `UpsertMessage`, etc. in `sync.go` only. If the sync code that calls `UpsertBooking()` lives in `channel_workflow.go` or `sync_cmd.go`, the domain upsert gets 0 points even though it exists and works.

**Impact:** +3 points lost for domain upserts, contributing to the 3/10 score on data_pipeline_integrity.

**Fix:** Same as Issue 1 — search all CLI files for the upsert patterns:
```go
syncContent := readAllGoFiles(filepath.Join(dir, "internal", "cli"))
```

---

### Issue 3: `scoreDeadCode()` over-penalizes with false positives from generated code patterns

**File:** `internal/pipeline/scorecard.go:1041-1057`
**Code:**
```go
// Check if flags defined in root.go are used in other CLI files
flagRe := regexp.MustCompile(`&flags\.(\w+)`)
flagNames := uniqueMatches(flagRe, rootContent)
otherCLI := readOtherGoFiles(cliDir, map[string]bool{"root.go": true})
for _, name := range flagNames {
    if !strings.Contains(otherCLI, "flags."+name) {
        deadFlags++
    }
}

// Check if functions defined in helpers.go are called in other CLI files
funcRe := regexp.MustCompile(`(?m)^func\s+([A-Za-z_]\w*)\s*\(`)
funcNames := uniqueMatches(funcRe, helpersContent)
otherHelpers := readOtherGoFiles(cliDir, map[string]bool{"helpers.go": true})
for _, name := range funcNames {
    if !strings.Contains(otherHelpers, name+"(") {
        deadFunctions++
    }
}

score := 5 - (deadFlags + deadFunctions)
```

**Problem 3a:** The flag usage check searches for the literal string `flags.<name>` in other CLI files. But in the generated code, flags are accessed through the `rootFlags` struct via the closure — many generated commands access flags through helper functions like `printOutput(cmd, flags, data, statusCode)` rather than directly as `flags.asJSON`. The flag IS used, but through a function call that passes the entire `flags` struct.

**Problem 3b:** The function regex `(?m)^func\s+([A-Za-z_]\w*)\s*\(` captures exported helper functions, but it also captures functions that are called by other helpers within the same file (`helpers.go`). A helper that is only called by other helpers in the same file is scored as "dead" even though it's reachable. The check only looks in `otherHelpers` (all files except helpers.go), missing intra-file calls.

**Problem 3c:** The scoring formula `5 - (deadFlags + deadFunctions)` means finding just 5 false positives gives 0/5. With ~15 flags defined in root.go and ~30 functions in helpers.go, even a 10% false-positive rate produces 4-5 false hits, zeroing the score.

**Impact:** 0/5 on dead_code even though `go vet` reports zero issues and the code compiles cleanly.

**Fix:**
1. For flags: check if `flags` is passed as an argument to any function call (e.g., `printOutput(cmd, flags, ...)`) in addition to direct field access. If `flags` appears as a function argument, all fields are reachable.
2. For functions: also check within helpers.go itself for intra-file calls.
3. Use `go vet` output as a floor: if `go vet ./...` passes, score at least 3/5.

```go
// Also count flags as "used" if the flags struct is passed to functions
if strings.Contains(otherCLI, "flags,") || strings.Contains(otherCLI, "flags)") {
    deadFlags = 0  // flags struct is passed around, all fields are reachable
}

// Check helpers.go itself for intra-file calls
allContent := helpersContent + otherHelpers
for _, name := range funcNames {
    if !strings.Contains(allContent, name+"(") || name == funcNames[0] {
        deadFunctions++
    }
}
```

---

### Issue 4: `scoreWorkflows()` uses a narrow prefix list that misses common workflow patterns

**File:** `internal/pipeline/scorecard.go:644`
**Code:**
```go
workflowPrefixes := []string{"stale", "orphan", "triage", "load", "overdue", "standup", "deps", "workflow"}
```

**Problem:** The prefix list is biased toward project-management APIs (Linear, GitHub Issues). For other API domains:
- Scheduling APIs: `agenda`, `free`, `conflicts`, `unconfirmed`, `stats`, `trends` are all workflow commands but match zero prefixes
- Payment APIs: `reconcile`, `disputed`, `revenue` would also miss
- Communication APIs: `archive`, `search`, `export` would also miss

The function has a secondary check (lines 670-689) that counts files with 2+ different API call types (`c.Get` + `c.Post` + `store.`), which partially compensates. But single-method workflows that do complex local computation (like `stats` which does SQL aggregation, or `conflicts` which does a self-join) only call `store.Open` and score 0 on both checks.

**Impact:** The Cal.com CLI has 10 workflow commands (agenda, search, stats, free, stale, conflicts, health, trends, unconfirmed, sync) but only `stale` and `workflow` match the prefix list. The secondary check catches a few more, but the final score is 6/10 instead of 10/10.

**Fix:** Either:
1. Expand the prefix list to be domain-agnostic: add `agenda`, `free`, `conflicts`, `unconfirmed`, `stats`, `trends`, `health`, `reconcile`, `revenue`, `archive`, `search`, `sync`, `busy`, `export`, `noshow`, `reassign`, `clone`.
2. Or better: instead of prefix-matching filenames, count commands that import `store` OR make 2+ API calls OR are registered under a `workflow` parent. Any command that queries the local store IS a workflow command by definition — it's doing compound work (sync → store → query).

```go
// A command that imports store is a workflow command (it uses the data layer)
if strings.Contains(content, "/store") || strings.Contains(content, "store.Open") {
    compoundCommands++
    continue
}
```

---

### Issue 5: `scoreInsight()` uses a narrow prefix list biased toward infrastructure APIs

**File:** `internal/pipeline/scorecard.go:715`
**Code:**
```go
insightPrefixes := []string{"health", "similar", "bottleneck", "trends", "patterns", "forecast"}
```

**Problem:** Only 6 prefixes. The Cal.com CLI has `health.go` and `trends` (inside `health.go`), but `stats.go`, `conflicts.go`, `stale.go` are all insight commands that don't match any prefix. A `stats` command that computes show rates and cancellation rates IS an insight. A `conflicts` command that detects double-bookings IS an insight.

**Impact:** 2/10 on insight (only `health.go` matches) despite having 5 genuine insight commands.

**Fix:** Same approach as Issue 4. Either expand the prefix list (`stats`, `conflicts`, `stale`, `analytics`, `busiest`, `velocity`, `utilization`, `coverage`, `gaps`, `noshow`) or detect insight commands structurally (commands that query the local store and produce aggregated/computed output rather than raw API passthrough).

---

### Issue 6: Verify results are not incorporated into the final score

**File:** `internal/pipeline/scorecard.go:57-129` (RunScorecard function)

**Problem:** The scorecard and the verify command are completely independent. `RunScorecard()` never calls or reads verify results. A CLI can score 100/100 on the scorecard (all patterns present) but fail 100% of runtime tests (patterns are present but broken). Conversely, a CLI can score 50/100 on the scorecard (patterns in unexpected locations) but pass 91% of runtime tests (everything actually works).

The verify command (`internal/pipeline/runtime.go`) builds the CLI, runs every command with `--help`, `--dry-run`, and mock execution, tests the data pipeline end-to-end, and produces a pass rate. This is strictly more authoritative than string pattern matching.

**Impact:** The final score can diverge from actual quality by 30+ points in either direction. This makes the scorecard unreliable as a quality gate — the skill's anti-shortcut rule says "The scorecard measures proxies for quality. Optimize for actual quality." but the scorecard doesn't measure actual quality at all.

**Fix:** Add an optional verify-integration step to RunScorecard. When verify results are available (either passed as input or run inline), use them as a floor/ceiling:

```go
func RunScorecard(outputDir, pipelineDir, specPath string, verifyReport *VerifyReport) (*Scorecard, error) {
    // ... existing scoring logic ...

    // If verify results are available, use them as a calibration signal
    if verifyReport != nil {
        verifyScore := int(verifyReport.PassRate * 100)  // 0-100

        // Verify pass rate sets a floor: if 90% of commands actually work,
        // the score can't be below 65 regardless of pattern matching
        floor := (verifyScore * 80) / 100  // 90% verify → 72 floor
        if sc.Steinberger.Total < floor {
            sc.Steinberger.Total = floor
            sc.Steinberger.CalibrationNote = fmt.Sprintf(
                "Score raised from %d to %d based on %d%% verify pass rate",
                originalTotal, floor, verifyScore)
        }

        // Verify failures cap specific dimensions
        if !verifyReport.DataPipeline {
            // Cap data pipeline at 5 if verify says pipeline fails
            if sc.Steinberger.DataPipelineIntegrity > 5 {
                sc.Steinberger.DataPipelineIntegrity = 5
            }
        }
    }
}
```

This doesn't replace static analysis — it calibrates it. Pattern matching catches structural issues (missing flags, wrong file organization). Verify catches behavioral issues (commands that crash, paths that 404). The two are complementary but verify should be authoritative when they disagree.

## Summary of Changes

| Issue | File | Lines | Severity | Effort |
|-------|------|-------|----------|--------|
| 1. sync_correctness hardcodes `sync.go` | scorecard.go | 947 | High — zeroes the dimension | Low — change to search all files |
| 2. data_pipeline hardcodes `sync.go` | scorecard.go | 915 | High — loses 3+ points | Low — same fix as #1 |
| 3. dead_code false positives | scorecard.go | 1041-1057 | Medium — zeroes the dimension | Medium — improve detection logic |
| 4. workflows prefix list too narrow | scorecard.go | 644 | Medium — underscores by 4+ points | Low — expand list or detect structurally |
| 5. insight prefix list too narrow | scorecard.go | 715 | Medium — underscores by 6+ points | Low — expand list or detect structurally |
| 6. verify results not incorporated | scorecard.go | 57-129 | High — score diverges from reality | Medium — add optional verify integration |

**Expected impact:** Fixing issues 1-5 would raise the example CLI from 57 to ~75-80. Adding issue 6 (verify calibration) would ensure no CLI with 90%+ verify pass rate scores below ~70.

## Testing

Each fix should be validated against the existing test suite:
- `scorecard_tier2_test.go` — update fixtures for issues 1-3
- `scorecard_run_test.go` — run against a real CLI with known verify results
- Add new test: `TestScorecard_VerifyCalibration` — verify that a CLI with high verify pass rate gets a floor score

## Non-Goals

- Changing the 18-dimension structure or the Tier 1/Tier 2 weighting
- Replacing static analysis with verify-only scoring (static analysis catches real issues)
- Changing grade thresholds (the grades are fine; the input scores are wrong)
