---
title: "fix: Retro trigger-dev template and scorer fixes"
type: fix
status: active
date: 2026-04-09
origin: docs/retros/2026-04-09-trigger-dev-retro.md
---

# fix: Retro trigger-dev template and scorer fixes

## Overview

Four fixes from the Trigger.dev retro (issue #158): scorer tools can't parse internal YAML specs, config.go template produces unused variable, store template doesn't populate typed columns from spec response fields, and sync template under-selects syncable resources.

## Problem Frame

The Printing Press generates CLIs that consistently hit the same friction points: go vet failures from template bugs, scorer tools that reject valid specs, and data layers that store everything as JSON blobs. These are generator-level issues that affect every future CLI, not just Trigger.dev.

## Requirements Trace

- R1. Dogfood, verify, and scorecard accept internal YAML specs the generator accepts
- R2. Generated config.go passes `go vet` without manual fixes
- R3. Per-entity tables include typed columns from spec response fields
- R4. All resources with simple GET list endpoints are included in default sync

## Scope Boundaries

- No changes to the internal YAML spec format itself
- No changes to the printing-press skill instructions (SKILL.md)
- Typed columns (R3) only for resources whose spec defines response fields - JSON blob fallback preserved
- Sync expansion (R4) only for resources with parameterless list paths - compound-path resources deferred

## Context & Research

### Relevant Code and Patterns

- `internal/spec/spec.go` - `APISpec.Validate()` enforces "at least one resource" (line 282). `Parse()` and `ParseBytes()` are the internal YAML entry points.
- `internal/pipeline/dogfood.go` - `loadDogfoodOpenAPISpec()` only tries OpenAPI parsers (line 259-282). No internal YAML path.
- `internal/pipeline/scorecard.go` - `loadOpenAPISpec()` does raw JSON unmarshal. No YAML support.
- `internal/generator/templates/config.go.tmpl` - `os.ReadFile` call at line 54, data only used inside TOML conditional.
- `internal/generator/templates/store.go.tmpl` - Already has `.Columns` iteration (lines 79-125) and `.FTS5Fields` support. Infrastructure exists; profiler needs to populate it.
- `internal/generator/templates/sync.go.tmpl` - `.SyncableResources` drives both `defaultSyncResources()` and `syncResourcePath()` (lines 462-483).
- `internal/generator/generator.go` - Profiler populates template data including `SyncableResources`, `Tables`, and `Columns`.
- `internal/spec/spec_test.go` - `TestValidation` already tests "no resources" (line 70).
- `internal/generator/generator_test.go` - `TestGenerateProjectsCompile` runs end-to-end compilation.

### Institutional Learnings

- ESPN and Postman retros both found empty sync resources (F4/F5 in ESPN retro, Finding #2 in Postman). This is a recurring pattern, not a one-off.
- Pagliacci retro found auth discovery gaps that better spec parsing in scorer tools would help validate.

## Key Technical Decisions

- **Add internal YAML format detection to scorer spec loading**: Rather than making scorers call `spec.Parse()` directly (which would require them to understand the full `APISpec` model), add a format-detection layer that routes YAML specs through `spec.Parse()` and extracts the paths and auth into the existing `openAPISpec` struct the scorers already use. This keeps the scorer interface unchanged.

- **Config template: conditional ReadFile**: Wrap the `os.ReadFile` call inside the format conditional rather than always reading and sometimes ignoring. This eliminates the unused variable for non-TOML formats without changing behavior.

- **Typed columns via profiler, not template changes**: The store template already iterates `.Columns`. The fix is in the profiler (generator.go) - when building table definitions, extract response fields from the spec and convert them to typed columns. The template needs no changes.

- **Sync resource inclusion: relax path-parameter check**: The profiler currently excludes resources whose list endpoint has path parameters. Relax this to include resources whose list path has no parameters (simple `GET /api/v1/<resource>`) and exclude only resources with compound parameters. Log excluded resources with a comment explaining why.

## Open Questions

### Resolved During Planning

- **Q: Do scorer tools need the full APISpec model?** No. They only need paths and auth. The fix routes internal YAML specs through `spec.Parse()` but then extracts just the fields scorers need into the existing `openAPISpec` struct.

- **Q: Does the store template need changes for typed columns?** No. The template already has `.Columns` iteration. Only the profiler needs to populate the columns from spec response fields.

### Deferred to Implementation

- **Q: What Go types should spec response fields map to?** String fields -> TEXT, integer fields -> INTEGER, boolean -> INTEGER, datetime/timestamp -> DATETIME, everything else -> TEXT. Exact mapping to be determined from the spec's type system.

- **Q: Should per-entity FTS indexes be generated from typed columns?** Likely yes (WU-3 in the retro mentions this), but this can be a follow-up if the typed column work is complex enough on its own.

## Implementation Units

- [ ] **Unit 1: Add internal YAML spec support to scorer tools**

**Goal:** Dogfood, verify, and scorecard accept internal YAML specs

**Requirements:** R1

**Dependencies:** None

**Files:**
- Modify: `internal/pipeline/dogfood.go` (add YAML format detection in `loadDogfoodOpenAPISpec`)
- Modify: `internal/pipeline/scorecard.go` (add YAML format detection in `loadOpenAPISpec`)
- Test: `internal/pipeline/dogfood_test.go` or new `internal/pipeline/spec_loading_test.go`

**Approach:**
- Add a format detection function that checks whether the spec bytes start with `name:` or `resources:` (internal YAML markers) vs `openapi:` or `swagger:` (OpenAPI markers) vs JSON (`{` or `[`)
- When internal YAML is detected, route through `spec.Parse()` to get an `APISpec`, then convert `APISpec.Resources` into the `openAPISpec` struct's `Paths` and `Auth` fields
- This conversion function can be shared across dogfood, verify, and scorecard
- When OpenAPI is detected, use the existing `ParseLenient()` path unchanged

**Patterns to follow:**
- `spec.Parse()` in `internal/spec/spec.go` for how internal YAML is parsed
- `loadDogfoodOpenAPISpec()` in `internal/pipeline/dogfood.go` for the existing fallback pattern
- `collectDogfoodSpecPaths()` for how OpenAPI paths are collected into the scorer format

**Test scenarios:**
- Happy path: Pass an internal YAML spec to dogfood, verify spec loads with correct paths and auth
- Happy path: Pass an OpenAPI spec to dogfood, verify existing behavior unchanged
- Edge case: Pass an empty YAML file, verify graceful error
- Edge case: Pass a YAML file with `name:` but no `resources:`, verify "at least one resource" error propagates clearly
- Integration: Run `printing-press dogfood --dir <test-cli> --spec <internal-yaml>` and verify it produces command-level results

**Verification:**
- `printing-press dogfood --spec <trigger-dev-spec.yaml>` no longer errors with "at least one resource"
- `printing-press verify --spec <trigger-dev-spec.yaml>` produces pass/fail results per command
- `printing-press scorecard --spec <trigger-dev-spec.yaml>` produces a score
- Existing OpenAPI spec flows are unbroken

- [ ] **Unit 2: Fix unused variable in config.go template**

**Goal:** Generated config.go passes `go vet` without manual fixes

**Requirements:** R2

**Dependencies:** None

**Files:**
- Modify: `internal/generator/templates/config.go.tmpl` (wrap ReadFile in format conditional)
- Test: `internal/generator/generator_test.go` (add vet check to compilation tests)

**Approach:**
- Move `data, err := os.ReadFile(path)` inside the TOML format conditional block
- For non-TOML formats (default), use `if _, err := os.Stat(path); err == nil { /* file exists */ }` as a presence check if needed, or simply remove the read entirely since env vars take precedence
- Alternatively, if the intent is to support JSON config reading in the future, add the JSON unmarshal branch now

**Patterns to follow:**
- Existing TOML conditional in config.go.tmpl (line 54-60)
- Other templates' config patterns

**Test scenarios:**
- Happy path: Generate a CLI with bearer_token auth, run `go vet ./...`, verify no unused variable errors
- Happy path: Generate a CLI with TOML config format, verify config file reading still works
- Edge case: Generate with no config format specified, verify `go vet` passes

**Verification:**
- `go vet ./...` passes on freshly generated CLIs without manual intervention
- `TestGenerateProjectsCompile` still passes

- [ ] **Unit 3: Populate typed columns from spec response fields**

**Goal:** Per-entity tables have typed columns beyond just id/data/synced_at

**Requirements:** R3

**Dependencies:** None (template infrastructure already exists)

**Files:**
- Modify: `internal/generator/generator.go` (profiler section that builds table definitions)
- Modify: `internal/generator/templates/store.go.tmpl` (may need minor adjustments to column rendering)
- Test: `internal/generator/generator_test.go` (add typed column assertions)

**Approach:**
- In the profiler, when building table definitions for each resource, inspect the resource's list endpoint response fields
- For each response field, map the spec type to a SQLite type: string -> TEXT, integer -> INTEGER, number -> REAL, boolean -> INTEGER, datetime strings -> DATETIME
- Include the top 8 fields as typed columns (id is already PRIMARY KEY; add status, name, created_at, updated_at, and the highest-gravity domain fields)
- Keep the `data JSON NOT NULL` column as a catch-all for the complete response
- The Upsert method should populate both the typed columns and the JSON blob

**Patterns to follow:**
- `.Columns` iteration in `store.go.tmpl` (lines 79-125) - the template already handles dynamic columns
- `.FTS5Fields` extraction in the profiler - similar pattern of inspecting response schema to extract field metadata
- `internal/spec/spec.go` `ResponseField` struct for field type information

**Test scenarios:**
- Happy path: Generate CLI from Stytch spec (has typed response fields), verify entity tables have typed columns beyond id/data/synced_at
- Happy path: Generate CLI from a spec with no response fields defined, verify tables fall back to id/data/synced_at (JSON-blob-only)
- Edge case: Response field has an unknown type, verify it maps to TEXT
- Edge case: Response has >8 fields, verify only top 8 are extracted as columns
- Integration: Generate, build, and run `sync` on a generated CLI, verify typed columns are populated during sync

**Verification:**
- Generated store.go has typed columns matching spec response fields
- Type Fidelity scorecard dimension improves from 3/5 to 4/5 or higher
- CLIs generated from specs without response fields still compile and work

- [ ] **Unit 4: Expand defaultSyncResources to include all simple-path resources**

**Goal:** All resources with parameterless GET list endpoints are synced by default

**Requirements:** R4

**Dependencies:** None

**Files:**
- Modify: `internal/generator/generator.go` (profiler logic for `.SyncableResources`)
- Test: `internal/generator/generator_test.go` (sync resource count assertions)

**Approach:**
- In the profiler, change the syncable resource selection criteria: include any resource that has a list endpoint whose path contains no `{param}` placeholders
- Resources with compound paths (e.g., `/projects/{projectRef}/envvars/{env}`) are excluded with a generated comment explaining why
- The `query` resource type (execute-only, no list) should also be excluded
- Add a `// Excluded from sync: envvars (requires projectRef), query (execute-only)` comment in the generated sync.go for clarity

**Patterns to follow:**
- Existing `.SyncableResources` population in `generator.go`
- `syncResourcePath()` template pattern in `sync.go.tmpl`

**Test scenarios:**
- Happy path: Generate CLI from a spec with 9 resources (like trigger-dev), verify defaultSyncResources includes all simple-path resources (runs, schedules, queues, waitpoints, deployments, batches)
- Edge case: Spec with only compound-path resources (all have {param} in list path), verify empty sync list with comments
- Edge case: Resource has no list endpoint, verify excluded from sync
- Happy path: Existing specs (stytch.yaml) produce the same or more sync resources than before (no regression)

**Verification:**
- Generated sync.go includes more resources than before for specs with many endpoints
- Resources with compound paths are excluded with explanation
- `TestGenerateProjectsCompile` still passes

## System-Wide Impact

- **Interaction graph:** Units 1-4 are independent (no cross-dependencies). Unit 1 changes how scorer tools load specs. Units 2-4 change what the generator emits.
- **Error propagation:** Unit 1 adds a new code path in spec loading that could produce new error types. Ensure errors from `spec.Parse()` are wrapped clearly so users know which parser failed.
- **Unchanged invariants:** The internal YAML spec format is unchanged. OpenAPI spec handling is unchanged. The generator's core template rendering is unchanged. The skill instructions (SKILL.md) are unchanged.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Unit 1 could break existing OpenAPI spec loading | Format detection routes to existing parsers for non-YAML specs. Existing test fixtures validate. |
| Unit 3 typed columns could break existing sync/upsert | JSON blob column is preserved. Typed columns are additive. |
| Unit 4 could include resources that shouldn't be synced | Path-parameter check is the guard. Only parameterless list paths are included. |

## Sources & References

- **Origin document:** `docs/retros/2026-04-09-trigger-dev-retro.md`
- Related issue: mvanhorn/cli-printing-press#158
- Spec parser: `internal/spec/spec.go`
- Dogfood spec loader: `internal/pipeline/dogfood.go` lines 259-282
- Store template: `internal/generator/templates/store.go.tmpl` lines 79-125
- Sync template: `internal/generator/templates/sync.go.tmpl` lines 462-483
