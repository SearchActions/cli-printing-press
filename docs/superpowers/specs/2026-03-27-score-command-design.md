# `/printing-press-score` — Standalone Scoring Skill

**Date:** 2026-03-27
**Status:** Draft

## Problem

The printing-press pipeline produces a scorecard during CLI generation, but there's no way to re-score an existing CLI, score a CLI by name from the library, or compare two CLIs side by side without re-running the full pipeline. Scoring should be a portable, on-demand operation.

## Solution

A new Claude Code skill (`printing-press-score`) that orchestrates the existing `printing-press scorecard` Go binary and renders results. One small Go change copies the spec into the output directory for self-contained scoring.

## Modes

### 1. Rescore Current CLI (`/printing-press-score`)

No arguments. Finds the "current" CLI by scanning `docs/plans/*-pipeline/state.json` files:
- If exactly one exists → use its `output_dir`
- If multiple exist → present a numbered menu for user selection

### 2. Score by Name or Path (`/printing-press-score notion-pp-cli-4`)

Single argument resolved as:
- **Path** (contains `/` or `.`) → use directly as the CLI directory
- **Name** → resolve to `library/<name>/`, falling back to `library/<name>-cli/`

### 3. Compare Two CLIs (`/printing-press-score compare notion-pp-cli-4 notion-pp-cli-2`)

Two or more CLI identifiers. The skill:
- Strips noise words (`vs`, `and`, `compare`, `against`)
- Resolves each remaining token using the same path-or-name logic
- Runs both scorecards **in parallel**
- Renders a side-by-side comparison table with deltas

## Argument Parsing

Free-form, intent-based. All of these are equivalent:
- `/printing-press-score compare notion-pp-cli-4 notion-pp-cli-2`
- `/printing-press-score notion-pp-cli-4 vs notion-pp-cli-2`
- `/printing-press-score notion-pp-cli-4 and notion-pp-cli-2`

The skill interprets the user's intent rather than enforcing strict syntax.

## Spec Resolution

Tier 2 (domain correctness) scoring requires the OpenAPI spec. Resolution order:

1. `<cli-dir>/spec.json` (new convention — copied during pipeline generation)
2. `docs/plans/<api-name>-pipeline/state.json` → `spec_path` field (if file still exists on disk)
3. If no spec found → run Tier 1 only, note "Tier 2 skipped — no spec found"

## Output Formatting

### Single Score

The skill runs `printing-press scorecard --dir <path> --spec <spec> --json`, parses the JSON, and renders:

```
Scorecard: notion-pp-cli-4
┌─────────────────────────────┬───────┐
│ Infrastructure (Tier 1)     │       │
├─────────────────────────────┼───────┤
│ Output Modes                │  8/10 │
│ Auth                        │  7/10 │
│ Error Handling              │  6/10 │
│ Terminal UX                 │  9/10 │
│ README                      │  5/10 │
│ Doctor                      │ 10/10 │
│ Agent Native                │  7/10 │
│ Local Cache                 │  4/10 │
│ Breadth                     │  7/10 │
│ Vision                      │  6/10 │
│ Workflows                   │  3/10 │
│ Insight                     │  5/10 │
├─────────────────────────────┼───────┤
│ Domain Correctness (Tier 2) │       │
├─────────────────────────────┼───────┤
│ Path Validity               │  9/10 │
│ Auth Protocol               │  8/10 │
│ Data Pipeline Integrity     │  7/10 │
│ Sync Correctness            │  6/10 │
│ Type Fidelity               │  4/5  │
│ Dead Code                   │  3/5  │
├─────────────────────────────┼───────┤
│ Total                       │ 72/100│
│ Grade                       │   B   │
└─────────────────────────────┴───────┘
Gaps: [list from gap_report if any]
```

### Compare Mode

Side-by-side table with delta column:

```
Scorecard Comparison
┌─────────────────────────────┬──────────────────┬──────────────────┬───────┐
│ Dimension                   │ notion-pp-cli-4  │ notion-pp-cli-2  │ Delta │
├─────────────────────────────┼──────────────────┼──────────────────┼───────┤
│ Output Modes                │            8/10  │            5/10  │   +3  │
│ Auth                        │            7/10  │            7/10  │    —  │
│ Error Handling              │            6/10  │            3/10  │   +3  │
│ ...                         │                  │                  │       │
├─────────────────────────────┼──────────────────┼──────────────────┼───────┤
│ Total                       │          72/100  │          56/100  │  +16  │
│ Grade                       │              B   │              C   │       │
└─────────────────────────────┴──────────────────┴──────────────────┴───────┘
```

Tables are rendered by the skill (LLM), not the Go binary. The binary provides JSON only.

## Implementation Changes

### 1. New Skill: `skills/printing-press-score/SKILL.md`

Responsibilities:
- Parse free-form user arguments into mode (rescore / single / compare)
- Resolve CLI names to directory paths
- Find specs using the resolution chain
- Shell out to `printing-press scorecard --dir <path> --spec <spec> --json`
- Execute parallel scorecard runs for compare mode
- Render rich tables from JSON output
- Present selection menu when current CLI is ambiguous

### 2. Go Change: `internal/pipeline/fullrun.go`

After generation succeeds, copy the source spec into `<output-dir>/spec.json`:
- If `specURL` is a local file path (no `http` prefix): `os.ReadFile` + `os.WriteFile` copy
- If `specURL` is a URL (`http://` or `https://`): `http.Get` + write response body to `<outputDir>/spec.json`
- Skip silently if the copy fails (scoring still works, just Tier 2 may need manual spec path later)

No changes to:
- The scorecard engine (`internal/pipeline/scorecard.go`)
- The existing `printing-press` skill (`skills/printing-press/SKILL.md`)
- The `scorecard` CLI subcommand (`internal/cli/scorecard.go`)

## Non-Goals

- Modifying the scoring algorithm or dimensions
- Adding new scorecard dimensions
- Changing how the existing pipeline invokes scoring
- Supporting more than two CLIs in a single compare (can be added later)
