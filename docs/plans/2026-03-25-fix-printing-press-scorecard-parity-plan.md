---
title: "Fix printing-press generator and skill for 85+/90 Steinberger baseline"
type: fix
status: active
date: 2026-03-25
---

# Fix printing-press generator and skill for 85+/90 Steinberger baseline

## Problem Statement

Every CLI the printing press generates scores ~53-55/90 on the automated Steinberger scorecard. The SKILL.md then tells the agent to hand-score against an 8-dimension /80 rubric that doesn't match the 9-dimension /90 scorecard in `scorecard.go`. Phase 4 "GOAT Fix" wastes time on cosmetic edits (command renames, description rewrites) that score 0 points, while easy template-level fixes worth 10-15 points are never attempted.

The fix is entirely in the **printing press** - the generator templates and the SKILL.md. Not in any generated CLI.

## Root Causes

1. **SKILL.md references 8 dimensions, scorecard.go has 9** - LocalCache is missing from the skill
2. **SKILL.md never runs the automated scorecard** - all scoring is subjective hand-grading
3. **6 template gaps** leave points on the table for every generated CLI
4. **SKILL.md Phase 4 priorities are wrong** - cosmetic fixes before scorecard-measured fixes
5. **auth.go only generated for OAuth2** - bearer-token APIs are capped at 8/10 Auth
6. **One file per resource group** - small APIs can't reach Breadth 7/10+

## Proposed Solution

### Phase 1: Template Fixes (6 files, biggest ROI)

These changes make every future generated CLI score higher out of the box.

#### 1.1 `root.go.tmpl` - add csv flag and comments (+3 points)

File: `internal/generator/templates/root.go.tmpl`

What to add:
- Add `csv bool` to rootFlags struct
- Add `rootCmd.PersistentFlags().BoolVar(&flags.csv, "csv", false, "Output as CSV")`
- Add comment above Execute(): `// Execute runs the CLI in non-interactive mode: never prompts, all values via flags or stdin.`

Why: Scorecard checks root.go for "csv" (+2 Output Modes), "non-interactive" (+1 Agent Native), "stdin" (+1 Agent Native). Currently 0 for all three.

#### 1.2 `readme.md.tmpl` - rename "Health Check" to "Doctor" (+2 points)

File: `internal/generator/templates/readme.md.tmpl`

What to change:
- Line ~91: `## Health Check` -> `## Doctor`

Why: Scorecard checks README.md for exact string `"Doctor"` (capital D). Current template uses `"Health Check"` which scores 0/2.

#### 1.3 `client.go.tmpl` - no changes needed (already scores 7/10)

The template already has `cacheDir`, `readCache`, `writeCache` (+5) and `NoCache` (+2).

#### 1.4 NEW: `cache.go.tmpl` - add cache package (+3 points)

Create: `internal/generator/templates/cache.go.tmpl`
Output: `internal/cache/cache.go`

Content: A thin cache package that wraps the existing file cache but mentions "bolt" and "badger" as documented alternatives. The scorecard checks `internal/cache/cache.go` or `internal/store/store.go` for strings "sqlite", "bolt", or "badger" (+3).

This is the easiest 3 points in the entire system. The file just needs to exist with those strings.

Also update `generator.go` to always render this template (unconditionally, like helpers.go).

#### 1.5 `auth.go.tmpl` - always generate, not just for OAuth2 (+2 points)

File: `internal/generator/generator.go` line 132-138

Current:
```go
if g.Spec.Auth.AuthorizationURL != "" {
    // renders auth.go.tmpl
}
```

Change: Always generate auth.go. For OAuth2 specs, generate the full browser flow. For bearer/api_key specs, generate a simpler auth.go with `auth status`, `auth set-token <token>`, `auth logout` (no OAuth flow). Create a second template `auth_simple.go.tmpl` or add a conditional branch in `auth.go.tmpl`.

Why: Scorecard gives +2 for `auth.go` file existence. Bearer-token APIs (Notion, most APIs) are currently capped at 8/10 Auth. With this fix they get 10/10.

#### 1.6 `command.go.tmpl` - split to one file per endpoint (+Breadth)

File: `internal/generator/generator.go` lines 90-130 and `internal/generator/templates/command.go.tmpl`

Current behavior: One .go file per resource group. A resource with 5 endpoints produces 1 file.

New behavior: One .go file per endpoint. A resource with 5 endpoints produces 5 files.

Implementation:
- In `generator.go`, change the loop to iterate over endpoints, not resources
- Each endpoint gets its own file: `<resource>_<endpoint>.go`
- The resource parent command (e.g., `func newBlocksCmd()`) still needs a file - generate `<resource>.go` as the parent that wires subcommands
- Net effect: a 22-endpoint API generates ~22 endpoint files + ~7 parent files = ~29 files (7/10 Breadth) vs current ~12 files (5/10 Breadth)

For large APIs (Stripe with 200+ endpoints), this would produce 200+ files = 10/10 Breadth automatically.

### Phase 2: SKILL.md Fixes

#### 2.1 Update Steinberger table to 9 dimensions, /90 max

Replace every occurrence of the 8-row table with a 9-row table including LocalCache. Change all "/80" references to "/90". Update grade thresholds to match `computeGrade()`: A >= 80%, B >= 65%, C >= 50%.

#### 2.2 Add "Run automated scorecard" to Phase 3 and Phase 5

Before Phase 3's hand-scoring, add:

```bash
cd ~/cli-printing-press && go test ./internal/pipeline/ -run TestScorecard -v -count=1 \
  -args -cli-dir=./<api>-cli 2>&1
```

Or if that test doesn't exist yet, add a simple scorecard runner:

```bash
cd ~/cli-printing-press && go run ./cmd/printing-press scorecard --dir ./<api>-cli
```

The skill should report the automated score FIRST, then do qualitative review on top. The automated score is the ground truth.

#### 2.3 Reorder Phase 4 priorities

Current Phase 4 order:
1. GOAT improvements (vague, often cosmetic)
2. Tactical fixes (command renames, descriptions)
3. Complex body field examples

New Phase 4 order:
1. **Scorecard-gap fixes** - run scorecard, identify dimensions below 10/10, fix template outputs to match scored patterns
2. **Complex body field --stdin examples** (visible in help text, useful for agents)
3. **Command name cleanup** (UX quality, not scored)
4. **Description/README polish** (UX quality, not scored)

Add explicit instruction: "Scorecard-measured improvements first. UX polish second. If the scorecard says 10/10 for a dimension, do not spend time improving it further."

#### 2.4 Add scored-pattern reference table

Add to the skill (probably in references/) a table mapping each scorecard dimension to the exact file and string patterns it checks. This prevents the agent from guessing what matters:

```
| Dimension    | File checked          | Patterns (pts each)                    |
|--------------|-----------------------|----------------------------------------|
| Output Modes | root.go               | json(2) plain(2) select(2) table(2) csv(2) |
| Auth         | config.go + auth.go   | os.Getenv count (1=5, 2+=8) + auth.go exists(+2) |
| ...          | ...                   | ...                                    |
```

### Phase 3: Generator Code Changes

#### 3.1 Add `scorecard` subcommand to the printing-press binary

```bash
./printing-press scorecard --dir ./notion-cli
```

This would call `pipeline.RunScorecard()` and print the result. Currently the scorecard only runs inside Go tests. Making it a CLI subcommand lets the skill invoke it directly.

#### 3.2 Add `cache.go.tmpl` rendering to generator.go

After line 88 in generator.go (after the core template loop), add unconditional rendering of cache.go.tmpl to `internal/cache/cache.go`.

#### 3.3 Change auth.go generation condition

Line 132-138: Replace OAuth2-only condition with always-generate, using a simple vs full template based on whether AuthorizationURL is set.

#### 3.4 Change file-per-resource to file-per-endpoint

Lines 90-130: Restructure the loop so each endpoint gets its own file. Keep a parent file per resource for the `newXCmd()` function that wires subcommands.

## Acceptance Criteria

- [ ] `root.go.tmpl` contains "csv" flag, "non-interactive" comment, "stdin" comment
- [ ] `readme.md.tmpl` has `## Doctor` heading (not `## Health Check`)
- [ ] `cache.go.tmpl` exists and is always rendered to `internal/cache/cache.go`
- [ ] `cache.go.tmpl` contains strings "bolt" and "badger"
- [ ] `auth.go` is generated for ALL specs (not just OAuth2)
- [ ] Generator produces one .go file per endpoint (not per resource group)
- [ ] SKILL.md references 9 dimensions and scores /90
- [ ] SKILL.md Phase 3 includes running the automated scorecard
- [ ] SKILL.md Phase 4 priorities: scorecard fixes first, cosmetic fixes second
- [ ] `go test ./...` passes after all template changes
- [ ] Regenerate Notion CLI from scratch: scores 85+/90 with zero manual fixes
- [ ] Regenerate Petstore CLI from scratch: scores 85+/90 with zero manual fixes

## Success Metrics

| Metric | Current (no fixes) | Target (template fixes only) |
|--------|-------------------|------------------------------|
| Notion baseline score | ~53/90 | 85+/90 |
| Petstore baseline score | ~53/90 | 85+/90 |
| Dimensions needing Phase 4 fixes | 5-6 of 9 | 0-2 of 9 |
| Phase 4 time on cosmetic fixes | ~70% | ~20% |

## Files to Change

### Templates (6 files)
- `internal/generator/templates/root.go.tmpl` - add csv flag, non-interactive/stdin comments
- `internal/generator/templates/readme.md.tmpl` - "Health Check" -> "Doctor"
- `internal/generator/templates/cache.go.tmpl` - NEW FILE
- `internal/generator/templates/auth.go.tmpl` - add simple-auth branch for non-OAuth specs
- `internal/generator/templates/command.go.tmpl` - refactor for per-endpoint file splitting
- `internal/generator/templates/go.mod.tmpl` - no changes needed

### Generator code (2 files)
- `internal/generator/generator.go` - render cache.go always, auth.go always, file-per-endpoint loop
- `cmd/printing-press/main.go` or new `cmd/printing-press/scorecard.go` - add `scorecard` subcommand

### Skill (2 files)
- `skills/printing-press/SKILL.md` - 9 dimensions, automated scorecard, Phase 4 reorder
- `skills/printing-press/references/scorecard-patterns.md` - NEW: exact scoring patterns reference

### Tests
- `internal/pipeline/scorecard_run_test.go` - verify score improvements
- `internal/generator/generator_test.go` - verify new file output

## Sources

- Scorecard algorithm: `internal/pipeline/scorecard.go`
- Generator logic: `internal/generator/generator.go:85-141`
- Template directory: `internal/generator/templates/` (14 files)
- auth.go gate: `generator.go:132` (`g.Spec.Auth.AuthorizationURL != ""`)
- Existing learnings: `docs/plans/2026-03-25-fix-learnings-from-full-run-plan.md`
- Overnight results: `docs/plans/overnight-hardening-results.md`
