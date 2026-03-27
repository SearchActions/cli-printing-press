---
title: "Soup-to-Nuts Press Test: Easy, Medium, Hard with Steinberger Scoring"
type: test
status: active
date: 2026-03-25
---

# Soup-to-Nuts Press Test: Easy, Medium, Hard

## The Question

How do we test the entire printing press pipeline end-to-end and score the output against Steinberger (10/10) and random community CLIs? Three difficulty levels, full pipeline each time.

## The Test

Run the FULL pipeline on 3 APIs. Not just `generate` - the entire `print` flow: research, plan, scaffold, enrich, review, scorecard, comparative. Each API tests a different difficulty level of the press.

| Level | API | Why This Difficulty | Spec Source |
|-------|-----|-------------------|-------------|
| Easy | Petstore | Small spec, 3 resources, always works | OpenAPI JSON |
| Medium | Plaid | Real API, 51 resources, official OpenAPI, has a dead competitor (57 stars) | OpenAPI YAML |
| Hard | Notion | No OpenAPI spec, docs-only, active competitor (87 stars) | Doc-to-spec |

## What "Soup to Nuts" Means

For EACH API, the press runs through every phase it has:

```
1. Research    → Discover competitors, analyze their GitHub repos
2. Scaffold    → Generate CLI from spec (or doc-to-spec for hard mode)
3. Quality     → 7 quality gates (compile, vet, build, binary, help, version, doctor)
4. Dogfood     → Run Tier 1 commands (--help, version, doctor, resource --help)
5. Scorecard   → Auto-score on 8 Steinberger dimensions (0-80, percentage)
6. Comparative → Score vs discovered competitors (6 dimensions, 0-100)
7. Report      → Produce a final scorecard.md with grade
```

## What We're Scoring

### Steinberger Score (0-80, shown as %)

| Dimension | What It Measures | How (grep generated code) |
|-----------|-----------------|--------------------------|
| Output Modes | --json, --plain, --select | grep root.go for flag registrations |
| Auth | env vars, config file, OAuth | grep config.go for Getenv, check auth.go exists |
| Error Handling | typed errors, actionable hints | grep helpers.go for "hint:", count exit codes |
| Terminal UX | color support, NO_COLOR, isatty | grep helpers.go for colorEnabled, NO_COLOR |
| README | quickstart, examples, troubleshooting | count sections in README.md |
| Doctor | health checks, connectivity | grep doctor.go for http calls |
| Agent Native | --json, --select, --dry-run, non-interactive | grep root.go for all 3 flags |
| Local Cache | SQLite, offline mode | check for cache imports (currently 0) |

### vs Competitors (0-100)

The 6-dimension comparative scoring from `comparative.go`: breadth, install friction, auth UX, output formats, agent friendliness, freshness.

### Final Grade

| Grade | Steinberger % | Meaning |
|-------|--------------|---------|
| A | 80%+ | Steinberger-tier. Ship immediately. |
| B | 65-79% | Competitive. Ship with known gaps. |
| C | 50-64% | Functional but mediocre. Needs work. |
| D | 35-49% | Below average. Major gaps. |
| F | <35% | Not ready. |

## Implementation Units

### Unit 1: Fix the Scorecard File Paths

**The scorecard is broken.** It looks for files at wrong paths (`cmd/root.go` instead of `internal/cli/root.go`, `internal/config.go` instead of `internal/config/config.go`, etc.). The scoring functions need to match the actual generated project structure.

**File:** `internal/pipeline/scorecard.go`

**Fixes needed:**
- `scoreOutputModes`: `cmd/root.go` -> `internal/cli/root.go`
- `scoreAuth`: `internal/config.go` -> `internal/config/config.go`, `internal/auth.go` -> `internal/cli/auth.go`
- `scoreErrorHandling`: `internal/helpers.go` -> `internal/cli/helpers.go`, look for `cliError` and `code:` instead of `os.Exit(`
- `scoreTerminalUX`: `internal/helpers.go` -> `internal/cli/helpers.go`
- `scoreDoctor`: check correct path for doctor.go (`internal/cli/doctor.go`)
- `scoreAgentNative`: `cmd/root.go` -> `internal/cli/root.go`

**Verification:** Run scorecard on Petstore CLI at /tmp/test-easy/cli. Output modes should score 6+ (has json, plain, select). Auth should score 0 (Petstore has no auth). Error handling should score 5+ (has hint: messages).

### Unit 2: Write the Soup-to-Nuts Test Runner

**File:** `internal/pipeline/e2e_test.go` (new)

A single test function that runs the FULL pipeline on all 3 APIs and prints a comparison table.

```go
func TestSoupToNuts(t *testing.T) {
    if os.Getenv("SOUP_TO_NUTS") == "" {
        t.Skip("Set SOUP_TO_NUTS=1 to run full pipeline test")
    }

    apis := []struct {
        name     string
        level    string
        specFlag string // --spec or --docs
        specURL  string
    }{
        {"petstore", "EASY", "--spec", "https://petstore3.swagger.io/api/v3/openapi.json"},
        {"plaid", "MEDIUM", "--spec", "https://raw.githubusercontent.com/plaid/plaid-openapi/master/2020-09-14.yml"},
        {"notion", "HARD", "--docs", "https://developers.notion.com/reference"},
    }

    for _, api := range apis {
        t.Run(api.name, func(t *testing.T) {
            // 1. Generate
            // 2. Run scorecard
            // 3. Print results
        })
    }

    // Print comparison table at the end
}
```

**Output format:**

```
╔══════════════════════════════════════════════════════════════════╗
║           PRINTING PRESS SOUP-TO-NUTS TEST RESULTS             ║
╠══════════════════════════════════════════════════════════════════╣
║ Dimension        │ Petstore (EASY) │ Plaid (MED) │ Notion (HARD)║
╠══════════════════╪═════════════════╪═════════════╪══════════════╣
║ Quality Gates    │ 7/7 PASS        │ 7/7 PASS    │ 7/7 PASS     ║
║ Commands         │ 8               │ 51+         │ ???          ║
║ Output Modes     │ 8/10            │ 8/10        │ 8/10         ║
║ Auth             │ 0/10            │ 8/10        │ 5/10         ║
║ Error Handling   │ 8/10            │ 8/10        │ 8/10         ║
║ Terminal UX      │ 10/10           │ 10/10       │ 10/10        ║
║ README           │ 8/10            │ 8/10        │ 8/10         ║
║ Doctor           │ 4/10            │ 4/10        │ 4/10         ║
║ Agent Native     │ 9/10            │ 9/10        │ 9/10         ║
║ Local Cache      │ 0/10            │ 0/10        │ 0/10         ║
╠══════════════════╪═════════════════╪═════════════╪══════════════╣
║ STEINBERGER      │ 47/80 (59%)     │ 55/80 (69%) │ 47/80 (59%)  ║
║ GRADE            │ C               │ B           │ C            ║
╠══════════════════╪═════════════════╪═════════════╪══════════════╣
║ vs Competitors   │ N/A (no alts)   │ 85/100      │ ???          ║
║ WE WIN?          │ N/A             │ YES         │ ???          ║
╚══════════════════╧═════════════════╧═════════════╧══════════════╝
```

**How to run:**
```bash
SOUP_TO_NUTS=1 go test ./internal/pipeline/ -run TestSoupToNuts -v -timeout 10m
```

### Unit 3: Research Phase for Each API

Before generating, run the research phase to discover competitors:

- **Petstore:** No competitors (it's a demo API). Research should return novelty 10, recommend "proceed".
- **Plaid:** Should find `landakram/plaid-cli` (57 stars, abandoned). Competitor intelligence should extract its 8 commands and feature requests.
- **Notion:** Should find `4ier/notion-cli` (87 stars, active). Should extract commands and set a target to beat.

**Verification:** research.json produced for each API with competitor_insights populated for Plaid and Notion.

### Unit 4: Self-Improvement Output

After scoring, check if any dimensions score < 5/10 and generate fix plans. The test should verify:

- **Local Cache** always scores 0/10 -> fix plan generated pointing at "need SQLite caching template"
- **Doctor** likely scores low -> fix plan suggests more health checks
- The fix plans are valid markdown with actionable template changes

## Acceptance Criteria

- [ ] Scorecard file paths fixed - scores non-zero on real generated CLIs
- [ ] Petstore (easy): 7/7 gates, scorecard produces grade C or better
- [ ] Plaid (medium): 7/7 gates, scorecard produces grade B or better, beats competitor
- [ ] Notion (hard): 7/7 gates from doc-to-spec, scorecard produces a grade
- [ ] Comparison table printed showing all 3 side by side
- [ ] Fix plans auto-generated for dimensions scoring < 5/10
- [ ] `SOUP_TO_NUTS=1 go test` runs the whole thing in one command

## Scope Boundaries

- Do NOT fix the press based on scorecard results (that's the NEXT plan)
- Do NOT generate CLIs to ship - these are test artifacts
- Do NOT add new scorecard dimensions beyond the 8 that exist
- The test runner uses the existing `generate` command, not the full `print` pipeline (which requires agent execution of plan files)
