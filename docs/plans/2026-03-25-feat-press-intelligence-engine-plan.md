---
title: "Press Intelligence Engine - The Press Plans Its Own Work"
type: feat
status: active
date: 2026-03-25
---

# Press Intelligence Engine - The Press Plans Its Own Work

## The Insight

OSC is powerful because it constantly writes ce:plan files before doing anything. The printing press pipeline has phases but they use static seed templates. The press should think like OSC: research, plan, execute, plan again, execute again. Each phase writes a plan informed by what the previous phase learned.

This plan makes the press SMARTER, not just wider. No CLI shipping until the press itself is excellent.

## What Makes the Press Dumb Today

1. **Static seed templates** - Every scaffold plan says the same thing regardless of what research found
2. **No competitor awareness** - The press doesn't know that competing CLIs exist, let alone what they're missing
3. **No doc-to-spec** - If there's no OpenAPI spec, the press gives up and asks a human to write YAML
4. **No plan chaining** - Research results don't flow into scaffold plans. Dogfood results don't flow into ship plans.
5. **No learning** - The press makes the same mistakes on every API. Flat resource problem? Hits it every time.

## What Makes the Press Smart After This Plan

1. **Dynamic plan generation** - Each phase writes the NEXT phase's plan using what it learned
2. **Competitor intelligence** - Reads competing CLIs' GitHub issues, README, and PRs to know what users want
3. **Doc-to-spec** - Crawls API docs and auto-generates specs. No hand-writing ever.
4. **Plan chaining** - research.json feeds scaffold plan. dogfood-results.json feeds ship plan. Everything connects.
5. **Learning from past runs** - The press remembers what went wrong (flat resources, empty auth, bad refs) and avoids it

## Implementation Units

### Unit 1: Dynamic Plan Generation (Plan Chaining)

**Goal:** Each pipeline phase writes the next phase's plan dynamically, incorporating what it learned.

**Files:**
- `internal/pipeline/seeds.go` - replace static templates with dynamic plan generators
- `internal/pipeline/planner.go` - new file: dynamic plan writer that reads prior phase outputs

**How it works today:**
```
Research phase completes → writes research.json
Scaffold phase starts → reads STATIC seed template (ignores research.json)
```

**How it works after:**
```
Research phase completes → writes research.json
Research phase → calls generateScaffoldPlan(research.json)
  → writes scaffold plan that says:
    "This API has 3 competing CLIs. The best has 40 commands.
     Generate at least 45 commands. Competitor X is missing
     transaction sync - include it. Use --lenient because
     similar specs had broken refs."
Scaffold phase starts → reads DYNAMIC plan (informed by research)
```

**Implementation:**

```go
// internal/pipeline/planner.go
package pipeline

// GenerateNextPlan writes a ce:plan-style plan file for the next phase,
// incorporating outputs from all completed prior phases.
func GenerateNextPlan(state *PipelineState, nextPhase string) (string, error) {
    var context PlanContext

    // Load all available prior phase outputs
    if research, err := LoadResearch(PipelineDir(state.APIName)); err == nil {
        context.Research = research
    }
    if dogfood, err := LoadDogfoodResults(PipelineDir(state.APIName)); err == nil {
        context.Dogfood = dogfood
    }
    if comparative, err := LoadComparativeResult(PipelineDir(state.APIName)); err == nil {
        context.Comparative = comparative
    }

    // Generate phase-specific plan using context
    switch nextPhase {
    case PhaseScaffold:
        return generateScaffoldPlan(state, context)
    case PhaseEnrich:
        return generateEnrichPlan(state, context)
    case PhaseReview:
        return generateReviewPlan(state, context)
    case PhaseShip:
        return generateShipPlan(state, context)
    default:
        return RenderSeed(nextPhase, seedDataFromState(state))
    }
}
```

Each generated plan includes:
- What prior phases discovered (competitor count, novelty score, demand signals)
- Specific instructions based on findings ("use --lenient", "target 45+ commands", "add transaction-sync endpoint")
- Known pitfalls from the press's learning system (see Unit 5)

**Verification:** `printing-press print petstore` generates dynamic scaffold plan that references research results, not the static seed template.

### Unit 2: Competitor Intelligence Engine

**Goal:** Research phase analyzes competing CLIs' GitHub repos to produce actionable insights.

**File:** `internal/pipeline/research.go` (extend existing)

**New functions:**

```go
func analyzeCompetitorRepo(owner, repo string) (*CompetitorAnalysis, error)
```

For each competing CLI discovered:
1. **Fetch open issues labeled enhancement/feature** via GitHub API
2. **Fetch README** and parse for command listings
3. **Fetch recent closed PRs** (unmerged = abandoned features = demand signals)
4. **Search issues for pain points** ("bug", "broken", "doesn't work")

**New types:**

```go
type CompetitorAnalysis struct {
    RepoURL          string   `json:"repo_url"`
    CommandsFound    []string `json:"commands_found"`
    FeatureRequests  []string `json:"feature_requests"`
    AbandonedPRs     []string `json:"abandoned_prs"`
    PainPoints       []string `json:"pain_points"`
    CommandCount     int      `json:"command_count"`
}

// Added to ResearchResult
type CompetitorInsights struct {
    Analyses         []CompetitorAnalysis `json:"analyses"`
    CommandTarget    int                  `json:"command_target"`     // max(competitor commands) * 1.2
    UnmetFeatures    []string             `json:"unmet_features"`     // features no competitor has
    PainPointsToAvoid []string            `json:"pain_points_to_avoid"`
}
```

**How this feeds into planning:** The dynamic scaffold plan (Unit 1) reads `CompetitorInsights.CommandTarget` and includes it as a goal: "Generate at least N commands." It lists `UnmetFeatures` as endpoints to prioritize in the spec.

**Verification:** `RunResearch("plaid", ...)` produces competitor insights for `landakram/plaid-cli` with feature requests and command count.

### Unit 3: Doc-to-Spec Generator

**Goal:** `printing-press generate --docs <url>` auto-generates a YAML spec from API documentation.

**File:** `internal/docspec/docspec.go` (new package)

**What it does:**

1. HTTP GET the docs page
2. Regex scan for endpoint patterns: `(GET|POST|PUT|PATCH|DELETE)\s+/[a-zA-Z0-9/{}_.-]+`
3. Extract parameter tables from HTML (detect `<table>` with parameter/type/required headers)
4. Extract JSON examples from `<pre>` and `<code>` blocks
5. Find auth method (scan for "Bearer", "API key", "OAuth", "Authorization")
6. Find base URL (scan for `https://api.`)
7. Generate internal YAML spec format

**Design principle:** Good enough to compile, not perfect. The spec is a starting point. Missing fields get sensible defaults (`type: string`, `required: false`). The goal is 7/7 quality gates, not 100% API coverage.

**File:** `internal/cli/root.go` - add `--docs` flag to generate command

**Verification:** `printing-press generate --docs "https://developers.notion.com/reference" --name notion --output /tmp/notion-test` produces a CLI that passes 7/7 gates.

### Unit 4: Press Learning System

**Goal:** The press remembers what went wrong on past runs and avoids repeating mistakes.

**File:** `internal/pipeline/learnings.go` (new)

**How it works:**

After each pipeline run, the press writes a learnings file:

```go
type PressLearning struct {
    APIName     string    `json:"api_name"`
    Date        time.Time `json:"date"`
    SpecType    string    `json:"spec_type"`    // "openapi", "docs", "yaml"
    Issues      []Issue   `json:"issues"`
    Fixes       []Fix     `json:"fixes"`
}

type Issue struct {
    Phase    string `json:"phase"`
    Gate     string `json:"gate"`       // which quality gate failed
    Error    string `json:"error"`
    Pattern  string `json:"pattern"`    // categorized: "empty-auth", "broken-ref", "flat-resource"
}

type Fix struct {
    Pattern  string `json:"pattern"`
    Solution string `json:"solution"`   // "--lenient", "guard empty EnvVars", "strip common prefix"
}
```

Stored at `~/.cache/printing-press/learnings.json`. Before each run, the press loads learnings and:
- If pattern "broken-ref" seen before: auto-enable `--lenient`
- If pattern "flat-resource" seen for similar spec size: warn in the scaffold plan
- If pattern "empty-auth" seen: the template already guards against this (learned from Fly.io/Telegram)

**How this feeds into planning:** The dynamic scaffold plan (Unit 1) includes a "Known Pitfalls" section generated from learnings: "APIs of this size (100+ paths) often produce flat resources. The enrich phase should check for common path prefix stripping."

**Verification:** After generating a CLI that required --lenient, the learning is saved. On the next run of a similar spec, --lenient is auto-suggested in the plan.

### Unit 5: Wire Pipeline Phases to Use Dynamic Plans

**Goal:** The `print` command's pipeline actually uses dynamic plans instead of static seeds.

**Files:**
- `internal/pipeline/pipeline.go` - update `InitPipeline` to call `GenerateNextPlan` after each phase completes
- `internal/pipeline/state.go` - add `LearningsPath` field to track learning file location

**Current flow in `InitPipeline`:**
```go
for _, phase := range PhaseOrder {
    seed, _ := RenderSeed(phase, seedData)
    os.WriteFile(state.PlanPath(phase), seed, 0644)
    state.MarkSeedWritten(phase)
}
```

**New flow:**
```go
// Write only the first phase's plan (preflight is always static)
seed, _ := RenderSeed(PhasePreflight, seedData)
os.WriteFile(state.PlanPath(PhasePreflight), seed, 0644)

// Subsequent phase plans are generated AFTER the prior phase completes
// (in the pipeline runner, not at init time)
```

Then in the phase completion handler:
```go
func (s *PipelineState) CompleteAndPlanNext(phase string) error {
    s.Complete(phase)
    nextPhase := s.NextPhase()
    if nextPhase == "" {
        return nil // all done
    }
    plan, err := GenerateNextPlan(s, nextPhase)
    if err != nil {
        // Fall back to static seed
        plan, _ = RenderSeed(nextPhase, seedDataFromState(s))
    }
    return os.WriteFile(s.PlanPath(nextPhase), []byte(plan), 0644)
}
```

**Verification:** `printing-press print petstore` writes Phase 0 plan at init, then dynamically generates Phase 1 plan only after Phase 0 completes, incorporating Phase 0 outputs.

### Unit 6: Automated Steinberger Scorecard

**Goal:** After every CLI generation, the press automatically scores the output against the Steinberger 10/10 bar AND against discovered competitors. Produces a report card.

**File:** `internal/pipeline/scorecard.go` (new)

**The scorecard runs automatically in the Review phase.** It reads the generated CLI and scores it:

```go
type Scorecard struct {
    APIName           string         `json:"api_name"`
    SteinbergerScore  SteinerScore   `json:"steinberger_score"`
    CompetitorScores  []CompScore    `json:"competitor_scores"`
    OverallGrade      string         `json:"overall_grade"`      // A/B/C/D/F
    GapReport         []string       `json:"gap_report"`
}

type SteinerScore struct {
    OutputModes     int `json:"output_modes"`      // 0-10 (--json, --plain, --select)
    Auth            int `json:"auth"`              // 0-10 (env var, config, OAuth)
    ErrorHandling   int `json:"error_handling"`    // 0-10 (typed errors, hints)
    TerminalUX      int `json:"terminal_ux"`       // 0-10 (color, NO_COLOR)
    README          int `json:"readme"`            // 0-10 (quickstart, examples)
    Doctor          int `json:"doctor"`            // 0-10 (auth, connectivity)
    AgentNative     int `json:"agent_native"`      // 0-10 (--json, --select, --dry-run)
    LocalCache      int `json:"local_cache"`       // 0-10 (SQLite, offline)
    Total           int `json:"total"`             // 0-80
    Percentage      int `json:"percentage"`        // total/80 * 100
}

type CompScore struct {
    CompetitorName  string `json:"competitor_name"`
    OurCommands     int    `json:"our_commands"`
    TheirCommands   int    `json:"their_commands"`
    OurScore        int    `json:"our_score"`       // 0-100 from comparative
    TheirScore      int    `json:"their_score"`     // 0-100 estimated
    WeWin           bool   `json:"we_win"`
}
```

**How it measures each dimension:**

| Dimension | How Measured (automated, no manual inspection) |
|-----------|----------------------------------------------|
| Output modes | grep generated root.go for "json", "plain", "select" flags |
| Auth | grep config.go for env var count + check auth.go exists |
| Error handling | grep helpers.go for "hint:" messages + count distinct exit codes |
| Terminal UX | grep helpers.go for "colorEnabled", "NO_COLOR" |
| README | count sections in README.md (Quick Start, Output Formats, Agent Usage, Troubleshooting) |
| Doctor | grep doctor.go for health check count |
| Agent native | grep root.go for "json", "select", "dry-run" flags |
| Local cache | check if any SQLite or cache package exists (currently always 0) |

**vs Competitors:** Uses the 6-dimension comparative scoring already built in `comparative.go`, but now runs automatically and produces a side-by-side report.

**Output:** `scorecard.md` in the pipeline directory:

```
# Plaid CLI Scorecard

## Steinberger Score: 69% (55/80)
| Dimension      | Score | Steinberger 10/10 | Gap |
|----------------|-------|-------------------|-----|
| Output modes   | 8     | --json + --select | --  |
| Auth           | 7     | env + config      | no keyring |
| Error handling | 8     | hints on 401/404  | -- |
| Terminal UX    | 6     | color + NO_COLOR  | no spinner |
| README         | 7     | quickstart + examples | no screenshots |
| Doctor         | 7     | auth + connectivity | -- |
| Agent native   | 8     | json + select + dry-run | -- |
| Local cache    | 0     | none              | no caching |

## vs Competitors
| Dimension | plaid-cli (ours) | landakram/plaid-cli | We win? |
|-----------|-----------------|---------------------|---------|
| Commands  | 51              | 8                   | YES |
| Install   | go install      | go install          | TIE |
| Auth      | env + config    | env only            | YES |
| JSON out  | yes             | yes                 | TIE |
| Dry run   | yes             | no                  | YES |
| Active    | 2026-03-25      | 2021-08-01          | YES |

## Grade: B+ (69% Steinberger, 4/6 dimensions beat competitor)
```

**How this feeds into planning:** The dynamic ship plan (Unit 1) reads the scorecard. If grade < B, the ship plan says "DO NOT SHIP - fix these gaps first." If grade >= B, it includes the scorecard in the README as proof of quality.

**Verification:** Generate petstore CLI, run scorecard, verify it produces a valid markdown report with percentage scores.

### Unit 7: Dogfood-Driven Press Improvement Loop

**Goal:** After scoring, the press identifies its own weaknesses and creates a ce:plan to fix them.

**File:** `internal/pipeline/selfimprove.go` (new)

**How it works:**

After the scorecard runs, if any Steinberger dimension scores < 5/10:

1. Identify the template file responsible for that dimension
2. Write a `fix-<dimension>-plan.md` in the pipeline directory
3. The plan describes what template change would improve the score

Example: if README scores 4/10 because the generated README is missing a "Troubleshooting" section:
```markdown
# Fix: README Template Missing Troubleshooting Section

## Problem
Generated README for plaid-cli scored 4/10 on the Steinberger README dimension.
Missing sections: Troubleshooting, detailed examples per command.

## Fix
Edit internal/generator/templates/readme.md.tmpl:
- Add Troubleshooting section with common error codes
- Add per-command examples using exampleLine template function

## Acceptance
Re-run scorecard. README dimension should score >= 7/10.
```

**This is the press improving itself.** Each CLI it generates is a test. Each test produces a scorecard. Each scorecard gap produces a plan. Each plan, when executed, makes the NEXT CLI better.

**Verification:** Generate a CLI, get a scorecard with a gap, verify a fix plan is auto-generated.

## Acceptance Criteria

### Press Intelligence (Units 1-5)
- [ ] Dynamic plan generation: each phase writes the next phase's plan using prior outputs
- [ ] Competitor intelligence: research phase fetches issues/README/PRs from competing CLIs
- [ ] Doc-to-spec: `--docs` flag generates YAML spec from API documentation
- [ ] Learning system: press saves issues/fixes and loads them for future runs
- [ ] Pipeline wiring: `print` command uses dynamic plans, not static seeds

### Automated Scoring (Units 6-7)
- [ ] Steinberger scorecard runs automatically after every generation
- [ ] Scorecard produces percentage score (0-100%) on 8 dimensions
- [ ] Competitor comparison produces side-by-side table
- [ ] Overall grade (A/B/C/D/F) determined automatically
- [ ] Gaps below 5/10 auto-generate fix plans for the press templates
- [ ] Fix plans are valid ce:plan files that can be executed with ce:work

### Dogfood Validation (generate CLIs to test the press)
- [ ] Generate Plaid CLI with competitor intelligence - scorecard shows improvement over raw generation
- [ ] Generate PagerDuty CLI with --lenient + dynamic plans - score > 60% Steinberger
- [ ] Generate at least 1 CLI from docs (Notion or Airtable) using doc-to-spec
- [ ] Each generated CLI is a TEST of the press, not a product
- [ ] `go test ./...` passes with all new packages

## Scope Boundaries

- CLIs are dogfood tests of the press, NOT products to ship
- Do NOT create Homebrew tap or push CLIs to repos yet
- Do NOT write the epic README (press capabilities aren't done yet)
- Social demand signals are out of scope (nice-to-have for later)
- The self-improvement loop (Unit 7) generates plans but does NOT auto-execute them

## Sources

- OSC plan-execute-plan pattern: `~/.claude/skills/osc-work/SKILL.md`, `osc-plan/SKILL.md`
- Current static seeds: `internal/pipeline/seeds.go`
- Current research phase: `internal/pipeline/research.go`
- Overnight learnings: `docs/plans/overnight-hardening-results.md`
- Steinberger gap analysis: generated CLIs score 6.9/10 - dynamic planning is part of closing to 8+
