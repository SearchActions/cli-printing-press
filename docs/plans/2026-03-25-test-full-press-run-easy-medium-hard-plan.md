---
title: "Full Press Run: Easy/Medium/Hard with Steinberger Scorecard and Learnings Loop"
type: test
status: active
date: 2026-03-25
---

# Full Press Run: Easy/Medium/Hard

## What This Is

Run the printing press at full power on 3 APIs. Clean room - fresh /tmp directories, no prior CLIs visible. The press uses every capability it has: research competitors, generate the CLI, dogfood it, score it against Steinberger AND the competitors it found, measure API coverage, then write a learnings plan for how to do it better next time.

## The 3 APIs

| Level | API | Why | Spec Source | Known Competitor |
|-------|-----|-----|-------------|-----------------|
| EASY | Petstore | Simple, 3 resources, demo API | OpenAPI JSON | None |
| MEDIUM | Plaid | Real fintech API, 50+ resources, official spec | OpenAPI YAML | landakram/plaid-cli (57 stars, abandoned) |
| HARD | Notion | No OpenAPI spec, must use doc-to-spec | Docs URL | 4ier/notion-cli (87 stars, active) |

## What "Full Power" Means

For each API, the press runs this sequence:

```
Step 1: RESEARCH
  - Search GitHub for competing CLIs
  - For each competitor: fetch issues, README, PRs
  - Produce research.json with novelty score + competitor insights
  - Measure: how many competitors? what's the command target?

Step 2: GENERATE
  - Easy/Medium: generate from OpenAPI spec
  - Hard: generate from docs URL (doc-to-spec)
  - Run 7 quality gates
  - Measure: did it compile? how many commands? how many resources?

Step 3: API COVERAGE
  - Count endpoints in the source spec (or extracted from docs)
  - Count commands in the generated CLI
  - Calculate: commands / spec_endpoints = % API coverage
  - Measure: what % of the API did we cover?

Step 4: DOGFOOD
  - Run Tier 1: --help, version, doctor, <resource> --help for each resource
  - Capture all output to evidence files
  - Measure: how many commands passed? any crashes?

Step 5: SCORECARD
  - Score on 8 Steinberger dimensions (0-80, shown as %)
  - Compare vs each competitor found in research
  - Assign grade (A/B/C/D/F)
  - Measure: what's our % vs Steinberger? do we beat competitors?

Step 6: LEARNINGS
  - What broke? What scored low? What could the press do better?
  - Write a learnings plan (ce:plan format) for press improvements
  - Measure: how many fix plans generated?
```

## Implementation Units

### Unit 1: Fix Scorecard File Paths (Blocker)

**Problem:** The scorecard (`internal/pipeline/scorecard.go`) looks for files at wrong paths. Every dimension except README scores 0 because it can't find the code.

**File:** `internal/pipeline/scorecard.go`

**Fixes (the scoring functions grep files at these wrong paths):**

| Function | Currently Looks At | Should Look At |
|----------|-------------------|----------------|
| `scoreOutputModes` | `cmd/root.go` | `internal/cli/root.go` |
| `scoreAuth` | `internal/config.go` | `internal/config/config.go` |
| `scoreAuth` | `internal/auth.go` | `internal/cli/auth.go` |
| `scoreErrorHandling` | `internal/helpers.go` | `internal/cli/helpers.go` |
| `scoreErrorHandling` | counts `os.Exit(` | should count `code:` in cliError structs |
| `scoreTerminalUX` | `internal/helpers.go` | `internal/cli/helpers.go` |
| `scoreDoctor` | check path | `internal/cli/doctor.go` |
| `scoreAgentNative` | `cmd/root.go` | `internal/cli/root.go` |

**Verification:** Run scorecard on an existing generated CLI. Output modes should score 6+ (has json, plain, select). Error handling should score 5+ (has hint: messages). Terminal UX should score 8+ (has colorEnabled, NO_COLOR, isatty).

### Unit 2: Add API Coverage Metric

**Problem:** We don't track what % of the API we covered. If Plaid has 200 endpoints and we generated 51 commands, that's ~25% coverage.

**File:** `internal/pipeline/scorecard.go` (extend)

**Add to Scorecard struct:**
```go
type Scorecard struct {
    // ... existing fields ...
    APICoverage APICoverageMetric `json:"api_coverage"`
}

type APICoverageMetric struct {
    SpecEndpoints   int `json:"spec_endpoints"`    // total endpoints in the source spec
    GeneratedCmds   int `json:"generated_commands"` // commands in the CLI
    CoveragePercent int `json:"coverage_percent"`   // generated/spec * 100
}
```

**How to count:**
- Spec endpoints: count paths in the parsed OpenAPI spec (already available from the generator). For doc-to-spec, count extracted endpoints.
- Generated commands: count `.go` files in `internal/cli/` minus helpers.go, root.go, doctor.go, auth.go (those are infrastructure, not API commands). Or grep for `func new.*Cmd` in the generated files.
- Coverage = generated / spec * 100

**Verification:** Petstore should show ~100% (small spec, all endpoints covered). Plaid should show < 50% (large spec, 50-endpoint-per-resource limit truncates).

### Unit 3: Build MakeBestCLI Function

**Goal:** One function that runs the entire sequence. No pipeline state files, no plan files, no agent execution. Just: API name in, graded CLI out.

**File:** `internal/pipeline/fullrun.go` (new)

```go
// FullRunResult holds everything the press produced for one API.
type FullRunResult struct {
    APIName     string
    Level       string // "easy", "medium", "hard"

    // Step 1: Research
    Research    *ResearchResult

    // Step 2: Generate
    OutputDir   string
    GatesPassed int // out of 7
    GatesFailed int
    CommandCount int
    ResourceCount int

    // Step 3: API Coverage
    Coverage    APICoverageMetric

    // Step 4: Dogfood
    Dogfood     *DogfoodResults

    // Step 5: Scorecard
    Scorecard   *Scorecard

    // Step 6: Learnings
    FixPlans    []string // paths to generated fix plans

    // Errors
    Errors      []string
}

func MakeBestCLI(apiName, level, specFlag, specURL, outputDir string) (*FullRunResult, error) {
    result := &FullRunResult{APIName: apiName, Level: level, OutputDir: outputDir}
    pipelineDir := filepath.Join(outputDir, ".pipeline")
    os.MkdirAll(pipelineDir, 0755)

    // Step 1: Research
    result.Research, _ = RunResearch(apiName, "catalog", pipelineDir)

    // Step 2: Generate (shell out to printing-press binary)
    // Use --spec or --docs based on specFlag
    // Parse output for PASS/FAIL gates

    // Step 3: API Coverage
    // Count commands in generated CLI vs endpoints in spec

    // Step 4: Dogfood
    // Find the built binary, run Tier 1

    // Step 5: Scorecard
    result.Scorecard, _ = RunScorecard(outputDir, pipelineDir)

    // Step 6: Learnings
    // If scorecard exists, generate fix plans for gaps

    return result, nil
}
```

**Verification:** `MakeBestCLI("petstore", "easy", "--spec", "https://petstore3.swagger.io/api/v3/openapi.json", "/tmp/fullrun-petstore")` returns a complete FullRunResult with all 6 steps populated.

### Unit 4: Build the Comparison Table Printer

**File:** `internal/pipeline/fullrun.go` (extend)

```go
func PrintComparisonTable(results []*FullRunResult) string
```

Produces the final output:

```
╔═══════════════════════════════════════════════════════════════════════╗
║              PRINTING PRESS FULL RUN RESULTS                        ║
╠═══════════════════════════════════════════════════════════════════════╣
║                     │ Petstore (EASY) │ Plaid (MED)  │ Notion (HARD) ║
╠═════════════════════╪═════════════════╪══════════════╪═══════════════╣
║ GENERATION          │                 │              │               ║
║   Quality Gates     │ 7/7 PASS        │ 7/7 PASS     │ 7/7 PASS      ║
║   Commands          │ 8               │ 51           │ ???           ║
║   Resources         │ 3               │ 51           │ ???           ║
║   API Coverage      │ 100%            │ 25%          │ ???           ║
╠═════════════════════╪═════════════════╪══════════════╪═══════════════╣
║ STEINBERGER (0-80)  │                 │              │               ║
║   Output Modes      │ 8/10            │ 8/10         │ 8/10          ║
║   Auth              │ 0/10            │ 8/10         │ 5/10          ║
║   Error Handling    │ 8/10            │ 8/10         │ 8/10          ║
║   Terminal UX       │ 10/10           │ 10/10        │ 10/10         ║
║   README            │ 8/10            │ 8/10         │ 8/10          ║
║   Doctor            │ 4/10            │ 4/10         │ 4/10          ║
║   Agent Native      │ 9/10            │ 9/10         │ 9/10          ║
║   Local Cache       │ 0/10            │ 0/10         │ 0/10          ║
║   TOTAL             │ 47/80 (59%)     │ 55/80 (69%)  │ 52/80 (65%)   ║
║   GRADE             │ C               │ B            │ B             ║
╠═════════════════════╪═════════════════╪══════════════╪═══════════════╣
║ vs COMPETITORS      │                 │              │               ║
║   Competitors Found │ 0               │ 1            │ 1             ║
║   Best Competitor   │ N/A             │ plaid-cli    │ notion-cli    ║
║   Their Commands    │ N/A             │ 8            │ 87 stars      ║
║   WE WIN?           │ N/A             │ YES (51>8)   │ ???           ║
╠═════════════════════╪═════════════════╪══════════════╪═══════════════╣
║ DOGFOOD             │                 │              │               ║
║   Tier 1 Pass Rate  │ 100%            │ 100%         │ ???           ║
║   Dogfood Score     │ 40/50           │ 40/50        │ ???           ║
╠═════════════════════╪═════════════════╪══════════════╪═══════════════╣
║ FIX PLANS GENERATED │ 2               │ 2            │ 2             ║
╚═══════════════════════════════════════════════════════════════════════╝
```

### Unit 5: Build the Test Runner

**File:** `internal/pipeline/fullrun_test.go` (new)

```go
func TestFullRun(t *testing.T) {
    if os.Getenv("FULL_RUN") == "" {
        t.Skip("Set FULL_RUN=1 to run full press test (takes 2-5 min)")
    }

    // Clean room: use /tmp, no access to prior CLIs
    baseDir := filepath.Join(os.TempDir(), "press-fullrun-"+time.Now().Format("20060102-150405"))
    os.MkdirAll(baseDir, 0755)
    defer os.RemoveAll(baseDir)

    apis := []struct {
        name, level, flag, url string
    }{
        {"petstore", "easy", "--spec", "https://petstore3.swagger.io/api/v3/openapi.json"},
        {"plaid", "medium", "--spec", "https://raw.githubusercontent.com/plaid/plaid-openapi/master/2020-09-14.yml"},
        {"notion", "hard", "--docs", "https://developers.notion.com/reference"},
    }

    var results []*FullRunResult
    for _, api := range apis {
        t.Run(api.name, func(t *testing.T) {
            outputDir := filepath.Join(baseDir, api.name+"-cli")
            result, err := MakeBestCLI(api.name, api.level, api.flag, api.url, outputDir)
            require.NoError(t, err)
            results = append(results, result)

            // Basic assertions
            assert.Equal(t, 7, result.GatesPassed, "all 7 gates should pass")
            assert.True(t, result.CommandCount > 0, "should generate at least 1 command")
            assert.NotNil(t, result.Scorecard, "scorecard should be produced")
        })
    }

    // Print the comparison table
    fmt.Println(PrintComparisonTable(results))

    // Write results to a file for review
    resultsPath := filepath.Join(baseDir, "full-run-results.md")
    os.WriteFile(resultsPath, []byte(PrintComparisonTable(results)), 0644)
    fmt.Printf("\nResults written to: %s\n", resultsPath)
}
```

**How to run:**
```bash
FULL_RUN=1 go test ./internal/pipeline/ -run TestFullRun -v -timeout 10m
```

### Unit 6: Learnings Plan Generator

**Goal:** After the full run, auto-generate a ce:plan for improving the press based on what went wrong.

**File:** `internal/pipeline/fullrun.go` (extend)

```go
func GenerateLearningsPlan(results []*FullRunResult, outputPath string) error
```

Reads all 3 results and produces a markdown plan:

```markdown
---
title: "Learnings from Full Run 2026-03-25"
type: fix
status: active
date: 2026-03-25
---

# Learnings from Full Press Run

## Run Summary
- Easy (Petstore): Grade C, 59% Steinberger, 100% API coverage
- Medium (Plaid): Grade B, 69% Steinberger, 25% API coverage
- Hard (Notion): Grade C, 65% Steinberger, 15% API coverage

## Consistent Gaps (same problem across all 3)
- Local Cache: 0/10 on all 3. Need SQLite caching template.
- Doctor: 4/10 on all 3. Need more health check patterns.

## Level-Specific Issues
- HARD: doc-to-spec only extracted N endpoints from Notion docs.
  Need multi-page crawling.
- MEDIUM: API coverage only 25% due to 50-endpoint limit per resource.
  Need smarter resource grouping.

## Recommended Fixes (prioritized by impact)
1. [fix] Bump endpoint-per-resource limit from 50 to 100
2. [fix] Doc-to-spec: follow links to find all endpoint pages
3. [feat] Add SQLite caching template for local-first CLI pattern
4. [fix] Doctor: add more health check endpoint patterns

## Next Run
After implementing fixes, re-run:
FULL_RUN=1 go test ./internal/pipeline/ -run TestFullRun -v -timeout 10m

Expected improvement: Grade C -> B on easy, Grade B -> B+ on medium.
```

**The loop:** Run full press -> get scores -> generate learnings plan -> execute fixes -> run full press again -> scores improve.

## Acceptance Criteria

- [ ] Scorecard file paths fixed - all 8 dimensions score correctly on real CLIs
- [ ] API coverage metric added - shows % of spec endpoints covered
- [ ] `MakeBestCLI()` runs the full sequence: research, generate, dogfood, score
- [ ] Comparison table prints all 3 APIs side by side with scores
- [ ] Petstore (easy): 7/7 gates, grade C or better, 80%+ API coverage
- [ ] Plaid (medium): 7/7 gates, grade B or better, beats landakram/plaid-cli on commands
- [ ] Notion (hard): 7/7 gates from doc-to-spec, produces a grade (any grade - it compiled from docs)
- [ ] Learnings plan auto-generated identifying consistent gaps and recommended fixes
- [ ] Clean room: all 3 runs in /tmp, no access to prior generated CLIs
- [ ] One command runs everything: `FULL_RUN=1 go test ./internal/pipeline/ -run TestFullRun -v -timeout 10m`

## Scope Boundaries

- Do NOT fix the press based on results (that's the learnings plan's job)
- Do NOT ship any CLIs
- Do NOT create Homebrew tap
- Do NOT run Tier 2/3 dogfood (no API credentials)
- The learnings plan is WRITTEN but not EXECUTED in this plan
