---
title: "feat: Add standalone /printing-press-score skill"
type: feat
status: completed
date: 2026-03-27
origin: docs/brainstorms/2026-03-27-score-command-requirements.md
---

# feat: Add standalone /printing-press-score skill

## Overview

Add a new Claude Code skill (`printing-press-score`) that provides on-demand CLI scoring — rescore the current CLI, score any CLI by name or path, or compare two CLIs side-by-side. A small Go change copies the OpenAPI spec into the output directory during generation to make scoring self-contained.

## Problem Frame

The printing-press pipeline produces a scorecard during CLI generation, but there's no way to re-score an existing CLI, score a CLI by name from the library, or compare two CLIs without re-running the full pipeline. Scoring should be portable and on-demand. (see origin: `docs/brainstorms/2026-03-27-score-command-requirements.md`)

## Requirements Trace

- R1. Score the current CLI with no arguments (`/printing-press-score`)
- R2. Score any CLI by name (`/printing-press-score notion-pp-cli-4`) or path (`/printing-press-score ~/my-cli`)
- R3. Compare two CLIs side-by-side with deltas (`/printing-press-score notion-pp-cli-4 vs notion-pp-cli-2`)
- R4. Free-form argument parsing — interpret intent, don't enforce syntax
- R5. Parallel scorecard execution for compare mode
- R6. Spec resolution chain: `<cli-dir>/spec.json` → pipeline state → Tier 1 only
- R7. Copy spec into output dir during pipeline generation (enables R6)
- R8. Rich table rendering from JSON output (single and compare modes)

## Scope Boundaries

- No changes to the scoring algorithm or dimensions
- No changes to the existing `printing-press` skill
- No changes to the `scorecard` CLI subcommand interface
- Compare limited to two CLIs (more can be added later)

## Context & Research

### Relevant Code and Patterns

- `skills/printing-press/SKILL.md` — skill frontmatter format, tool permissions, how to shell out to binary
- `skills/printing-press-catalog/SKILL.md` — simpler read-only skill, closest structural template
- `internal/pipeline/fullrun.go:60` — `MakeBestCLI()` function, insertion point for spec copy after line ~107
- `internal/cli/scorecard.go` — existing CLI command: `--dir`, `--spec`, `--json` flags
- `internal/pipeline/scorecard.go` — `RunScorecard()`, `Scorecard` struct with 18 dimensions
- `internal/pipeline/state.go` — `PipelineState` struct with `SpecPath`, `SpecURL`, `OutputDir`; `PipelineDir()` returns `docs/plans/<apiName>-pipeline`

### Institutional Learnings

- The scorecard has 18 dimensions (12 Tier 1 at /10 each + 6 Tier 2 with varying max), not the 9 documented in the stale `scorecard-patterns.md`. The skill must render TypeFidelity and DeadCode as /5.
- `specFlag` distinguishes `--spec` (local file path, copyable) from `--docs` (URL, no raw spec to copy). Only `--spec` mode should trigger the spec copy.
- Always use `--json` output and parse the `Scorecard` struct. Never scrape human-readable output.
- Skills build the binary first: `go build -o ./printing-press ./cmd/printing-press` with relative paths per AGENTS.md.

## Key Technical Decisions

- **Skill as orchestrator, binary as scorer:** The skill handles argument parsing, name resolution, menu presentation, and table rendering. The Go binary does the math. This follows the established "skill is the brain, binary is the tool" architecture. (see origin)
- **Spec copy only for `--spec` mode:** When `specFlag == "--docs"`, the spec is generated from documentation scraping and doesn't exist as a standalone file. Only `--spec` mode has a raw spec file to copy.
- **`library/` resolution with state.json fallback:** `library/<name>/` may not exist yet for older CLIs. The skill also scans `docs/plans/*-pipeline/state.json` to resolve names to `output_dir` paths.
- **Parallel scoring via concurrent Bash calls:** For compare mode, the skill invokes both `./printing-press scorecard --json` commands simultaneously using parallel Bash tool calls.

## Open Questions

### Resolved During Planning

- **Where to copy the spec?** → `<output-dir>/spec.json` at the root, alongside `go.mod`. Simple and discoverable.
- **How to handle URL specs?** → Only copy when `specFlag == "--spec"` (local file). `--docs` mode has no raw spec to copy.
- **How to find "current" CLI?** → Scan `docs/plans/*-pipeline/state.json` files. If exactly one, use it. If multiple, present a menu via AskUserQuestion.

### Deferred to Implementation

- Exact skill wording for menus and error messages
- Whether `scorecard-patterns.md` should be updated (out of scope, separate concern)

## Implementation Units

- [x] **Unit 1: Copy spec into output dir during generation**

  **Goal:** After successful CLI generation in `MakeBestCLI`, copy the source spec file into `<output-dir>/spec.json` so scoring can find it later.

  **Requirements:** R6, R7

  **Dependencies:** None

  **Files:**
  - Modify: `internal/pipeline/fullrun.go`
  - Test: `internal/pipeline/fullrun_test.go`

  **Approach:**
  - Insert spec copy logic after generation succeeds (around line 107) and before LLM polish (line 110)
  - Only copy when `specFlag == "--spec"` — this means `specURL` is a local file path
  - Use `os.ReadFile` + `os.WriteFile` with `0o644` permissions (matching existing codebase pattern)
  - Log but do not fail if the copy errors — append to `result.Errors` and continue, matching the error handling pattern used throughout `MakeBestCLI`
  - Extract the copy logic into a small helper function (e.g., `copySpecToOutput`) for testability

  **Patterns to follow:**
  - Error handling in `MakeBestCLI`: append to `result.Errors`, don't return early (except on generation failure)
  - File I/O: `os.ReadFile`/`os.WriteFile` with `0o` prefix permissions

  **Test scenarios:**
  - Happy path: spec file exists, `specFlag == "--spec"` → `spec.json` appears in output dir with identical content
  - Edge case: `specFlag == "--docs"` → no `spec.json` created, no error
  - Error path: spec file path doesn't exist → error appended to result, pipeline continues
  - Edge case: output dir doesn't exist yet → function handles gracefully (output dir is created by generation step)

  **Verification:**
  - `go test ./internal/pipeline/... -run TestCopySpec` passes
  - `go vet ./internal/pipeline/...` clean

- [x] **Unit 2: Create the printing-press-score skill**

  **Goal:** New skill at `skills/printing-press-score/SKILL.md` that handles all three modes: rescore current, score by name/path, compare two CLIs.

  **Requirements:** R1, R2, R3, R4, R5, R6, R8

  **Dependencies:** Unit 1 (spec.json convention must be established, though the skill degrades gracefully without it)

  **Files:**
  - Create: `skills/printing-press-score/SKILL.md`

  **Approach:**

  The skill is structured as a Claude Code skill markdown file with YAML frontmatter and natural-language instructions. It instructs the LLM to:

  1. **Parse arguments** — interpret the user's free-form input to determine mode:
     - No args → rescore current (scan state files, present menu if ambiguous)
     - Single token → score by name or path
     - Two+ tokens (after stripping `vs`, `and`, `compare`, `against`) → compare mode

  2. **Resolve CLI directories** — for each CLI identifier:
     - Contains `/` or `.` → treat as path
     - Otherwise → try `library/<name>/`, then `library/<name>-cli/`, then scan `docs/plans/*-pipeline/state.json` for matching `output_dir`

  3. **Find spec** — for each resolved CLI dir:
     - Check `<cli-dir>/spec.json`
     - If not found, glob `docs/plans/*-pipeline/state.json`, parse for `spec_path`, check if file exists
     - If nothing found, omit `--spec` flag (Tier 2 scores will be 0, note this to user)

  4. **Build binary** — `go build -o ./printing-press ./cmd/printing-press`

  5. **Run scorecard** — `./printing-press scorecard --dir <path> --spec <spec> --json`
     - For compare mode, run both in parallel via concurrent Bash tool calls

  6. **Render output** — parse JSON, render rich markdown tables:
     - Single: two-section table (Tier 1 /10 dimensions, Tier 2 with /10 and /5 dimensions, total, grade, gaps)
     - Compare: side-by-side table with delta column, same section structure

  **Patterns to follow:**
  - `skills/printing-press-catalog/SKILL.md` — frontmatter format, simpler skill structure
  - `skills/printing-press/SKILL.md` — AskUserQuestion usage pattern for menus
  - `internal/cli/scorecard.go` — the JSON output contract the skill parses

  **Test scenarios:**
  - (Skill files are not unit-tested. Verification is manual invocation.)

  **Verification:**
  - `/printing-press-score` with a generated CLI in `docs/plans/*-pipeline/` correctly discovers and scores it
  - `/printing-press-score <name>` resolves a library CLI and produces a score table
  - `/printing-press-score <path>` scores a CLI at an arbitrary path
  - `/printing-press-score <name1> vs <name2>` runs both scorecards and produces a comparison table
  - When no spec is found, Tier 2 is noted as skipped
  - When multiple current CLIs exist, a selection menu is presented

## System-Wide Impact

- **Interaction graph:** The skill shells out to the existing `scorecard` CLI subcommand. No new code paths in the Go binary beyond the spec copy. The scorecard engine is read-only against generated CLI files.
- **Error propagation:** Spec copy failure is non-fatal (logged, pipeline continues). Scorecard failures in the skill are reported to the user with the error message.
- **State lifecycle risks:** None — the skill is read-only against pipeline state. The spec copy is a one-time write during generation.
- **API surface parity:** The `scorecard` CLI subcommand interface is unchanged. The skill is a new consumer of its `--json` output.
- **Unchanged invariants:** The scoring algorithm, grade thresholds, and dimension weights are untouched. The existing `printing-press` skill's pipeline flow is unmodified.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| `library/` dir may not exist for a given CLI | Name resolution falls back to pipeline state files |
| Spec file at `spec_path` in state.json may have been moved/deleted | Graceful degradation to Tier 1 only with user notification |
| Stale `scorecard-patterns.md` could confuse skill | Skill uses JSON output struct as source of truth, not the reference file |

## Sources & References

- **Origin document:** [docs/brainstorms/2026-03-27-score-command-requirements.md](docs/brainstorms/2026-03-27-score-command-requirements.md)
- Related code: `internal/pipeline/scorecard.go` (scoring engine), `internal/cli/scorecard.go` (CLI wrapper), `internal/pipeline/fullrun.go` (spec copy target)
- Existing skills: `skills/printing-press/SKILL.md`, `skills/printing-press-catalog/SKILL.md`
