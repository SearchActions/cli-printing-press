---
title: "feat: Route generated CLI output to shelf/ directory"
type: feat
status: completed
date: 2026-03-27
---

# feat: Route generated CLI output to shelf/ directory

## Overview

Change the default output directory for generated CLIs from the repo root (e.g., `./stripe-cli/`) to a `shelf/` subdirectory (e.g., `./shelf/stripe-cli/`). Create `shelf/` if it doesn't exist. This namespaces generated output, keeps the repo root clean, and establishes a consistent location for CLIs that may later be submitted via PR or published elsewhere.

## Problem Frame

Generated CLIs currently land in the repo root, cluttering it alongside source code, docs, and the existing `catalog/` metadata directory. There's no single place to find all generated CLIs. The `shelf/` directory provides a printing-press-aligned namespace — CLIs are "placed on the shelf" after being printed.

## Requirements Trace

- R1. Default output directory changes from `./<name>-cli` to `./shelf/<name>-cli` across all commands
- R2. `shelf/` directory is created automatically if it doesn't exist
- R3. Explicit `--output` flag still overrides the default (no behavior change for explicit paths)
- R4. Flag help text and CLI examples reflect the new default
- R5. Skill documentation reflects the new default path

## Scope Boundaries

- The `catalog/` directory (API metadata YAML) is unaffected
- No changes to the generator itself — it receives an absolute path and is path-agnostic
- No migration of previously generated CLIs
- Pipeline state/plan directories (`docs/plans/<name>-pipeline/`) are unaffected
- Test helpers that use `t.TempDir()` are unaffected (they use explicit paths, not the default)

## Context & Research

### Relevant Code and Patterns

There are **four locations** where the default output path `<name>-cli` is constructed:

1. **`generate` command** — `internal/cli/root.go:213-214` — `outputDir = apiSpec.Name + "-cli"`
2. **`generate --docs` path** — `internal/cli/root.go:114-115` — `outputDir = parsed.Name + "-cli"`
3. **`print` command** — `internal/pipeline/pipeline.go:25-26` — `outputDir = "./" + apiName + "-cli"`
4. **`vision` command** — `internal/cli/vision.go:33-34` — `outputDir = apiName + "-cli"`

All four follow the same pattern: check if `outputDir` is empty, if so set it to `<name>-cli`. The fix is to prepend `shelf/` to each default.

After the default is set, each location calls `filepath.Abs(outputDir)` which will resolve `shelf/<name>-cli` relative to cwd — no additional changes needed for path resolution.

### Documentation References

- **Skill file** — `skills/printing-press-catalog/SKILL.md:93` — `--output ./<name>-cli`
- **Flag help text** — `root.go:279`, `root.go:472`, `vision.go:93`

## Key Technical Decisions

- **Use `filepath.Join("shelf", name+"-cli")` instead of string concatenation** — Ensures correct path separators on all platforms and is consistent with Go conventions.
- **Create `shelf/` via `os.MkdirAll` on the resolved absolute path** — `os.MkdirAll` on the full output path (e.g., `/abs/path/shelf/stripe-cli/`) will create both `shelf/` and the CLI subdirectory as needed. The generator's `Generate()` method already calls `os.MkdirAll` for its internal subdirectories, so `shelf/` creation happens naturally. No separate mkdir step is needed.
- **Add `shelf/` to `.gitignore`** — Generated CLIs are build artifacts, not source. Keeping them gitignored prevents accidental commits of large generated trees while still allowing the directory to exist locally.

## Open Questions

### Resolved During Planning

- **Should `shelf/` be gitignored?** Yes — generated CLIs are ephemeral build output. Users who want to commit a CLI can use `git add -f` or move it out. This matches the pattern where `printing-press` (the binary) is already gitignored.

### Deferred to Implementation

- None

## Implementation Units

- [ ] **Unit 1: Change default output paths to shelf/ prefix**

  **Goal:** All four default-path locations prepend `shelf/` so generated CLIs land in `shelf/<name>-cli/` by default.

  **Requirements:** R1, R2

  **Dependencies:** None

  **Files:**
  - Modify: `internal/cli/root.go`
  - Modify: `internal/cli/vision.go`
  - Modify: `internal/pipeline/pipeline.go`

  **Approach:**
  - In each of the four default-path assignments, change from `<name> + "-cli"` to `filepath.Join("shelf", <name>+"-cli")`.
  - `root.go:214`: `outputDir = filepath.Join("shelf", apiSpec.Name+"-cli")`
  - `root.go:115`: `outputDir = filepath.Join("shelf", parsed.Name+"-cli")`
  - `pipeline.go:26`: `outputDir = filepath.Join("shelf", apiName+"-cli")`
  - `vision.go:34`: `outputDir = filepath.Join("shelf", apiName+"-cli")`
  - The subsequent `filepath.Abs()` call handles resolution to an absolute path. `os.MkdirAll` in the generator handles directory creation. No other code changes needed for path mechanics.

  **Patterns to follow:**
  - Existing `filepath.Abs()` resolution pattern already in place at each location
  - `os.MkdirAll` usage in `generator.Generate()` for subdirectory creation

  **Test scenarios:**
  - `go build` succeeds
  - `go vet` passes
  - Existing tests pass (`go test ./...`) — tests use explicit `t.TempDir()` paths, not defaults

  **Verification:**
  - All four default paths produce `shelf/<name>-cli` when `--output` is not specified
  - `--output /custom/path` still works unchanged
  - `go test ./...` passes

- [ ] **Unit 2: Update flag help text and examples**

  **Goal:** CLI help output reflects the new `shelf/` default so users aren't surprised.

  **Requirements:** R4

  **Dependencies:** Unit 1

  **Files:**
  - Modify: `internal/cli/root.go`
  - Modify: `internal/cli/vision.go`

  **Approach:**
  - `root.go:279`: Change flag description to `"Output directory (default: shelf/<name>-cli)"`
  - `root.go:472`: Change to `"Output directory (default: shelf/<api-name>-cli)"`
  - `vision.go:93`: Change to `"Output directory (default: shelf/<api>-cli)"`

  **Patterns to follow:**
  - Existing flag description format

  **Test scenarios:**
  - `printing-press generate --help` shows updated default
  - `printing-press print --help` shows updated default

  **Verification:**
  - All three `--output` flag descriptions reference `shelf/`

- [ ] **Unit 3: Add shelf/ to .gitignore and update skill docs**

  **Goal:** Generated CLIs are gitignored; skill documentation reflects the new path.

  **Requirements:** R5

  **Dependencies:** Unit 1

  **Files:**
  - Modify: `.gitignore`
  - Modify: `skills/printing-press-catalog/SKILL.md`

  **Approach:**
  - Add `shelf/` line to `.gitignore`
  - In `SKILL.md`, update the `--output ./<name>-cli` example to `--output ./shelf/<name>-cli` and update the "Try it" section paths from `cd <name>-cli` to `cd shelf/<name>-cli`

  **Patterns to follow:**
  - Existing `.gitignore` entries (`printing-press`, `.cache/`)

  **Test scenarios:**
  - `git status` does not show generated CLI directories inside `shelf/`

  **Verification:**
  - `.gitignore` contains `shelf/`
  - Skill file references `shelf/` in generation and try-it instructions

## System-Wide Impact

- **Interaction graph:** The `generate`, `print`, and `vision` commands are the only entry points. The generator package is path-agnostic (receives absolute path). Pipeline state stores the absolute path, so existing pipeline checkpoints will use whatever path was set at init time.
- **Error propagation:** No change — the existing `--force`/existence checks operate on the resolved absolute path.
- **State lifecycle risks:** Existing pipeline state files (`docs/plans/<name>-pipeline/state.json`) store `output_dir` as an absolute path. Pipelines created before this change will continue to use their stored path. No migration needed.
- **API surface parity:** The `--output` flag override is unchanged. Only the default changes.

## Risks & Dependencies

- **Low risk:** Users with muscle memory for `cd <name>-cli` after generation will need to use `cd shelf/<name>-cli`. Mitigated by updating the post-generation output message (which already prints the output path).

## Sources & References

- Related code: `internal/cli/root.go`, `internal/cli/vision.go`, `internal/pipeline/pipeline.go`
- Skill docs: `skills/printing-press-catalog/SKILL.md`
