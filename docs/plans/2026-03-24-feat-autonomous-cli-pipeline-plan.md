---
title: "Autonomous CLI Pipeline - From API Name to Finished CLI"
type: feat
status: active
date: 2026-03-24
---

# Autonomous CLI Pipeline - From API Name to Finished CLI

## Overview

Transform the printing press from a single-pass generator into a multi-phase autonomous pipeline. The user says `printing-press print gmail`. The press discovers the spec, scaffolds, enriches with deep research, reviews, and ships - each phase writing checkpoint state and chaining to the next. Zero human intervention. Overnight quality.

Inspired by OSC nightnight's proven patterns: checkpoint-based resumption, mechanical budget gates, CronCreate session chaining, JSONL event logging.

## Problem Statement

The press currently does one pass: spec in, CLI out, 70% quality. The missing 30% is information that exists in the spec and public docs but requires multiple analysis passes with different lenses to extract. A single context window can't hold spec parsing + deep doc research + dogfood analysis + enrichment application. The solution is phase chaining - the same pattern that lets OSC run for 8 hours overnight.

## Architecture: The Print Pipeline

```
printing-press print gmail
        |
    [PREFLIGHT] - Go installed? Binary fresh? Disk space?
        |
    [DISCOVER]  - Find OpenAPI spec (known-specs, apis-guru, WebSearch)
        |
    [SCAFFOLD]  - Generate initial CLI (existing generate command)
        |
    [ENRICH]    - Deep-read spec hints + research API docs -> spec overlay
        |
    [REGENERATE]- Re-run generator with enriched spec (SCAFFOLD + overlay)
        |
    [REVIEW]    - Static analysis: help output, name quality, compilation, coverage
        |
    [SHIP]      - git init, README, goreleaser validate, morning report
        |
    Done. CLI at ./gmail-cli/
```

**Key decision: ENRICH modifies the spec, not generated code.** The generator remains the single source of truth for code output. ENRICH produces a YAML overlay that gets merged into the original spec before re-generation. This collapses the original 6-phase plan into 7 tighter phases (PREFLIGHT + DISCOVER + SCAFFOLD + ENRICH + REGENERATE + REVIEW + SHIP) where ENRICH + REGENERATE replace the earlier ENRICH + POLISH split.

## Scope Boundaries

- No runtime library - templates only
- No real API testing in REVIEW (no credential management) - static analysis only
- No GitHub repo creation or release publishing in SHIP - local only
- No Google Discovery format converter - use apis-guru pre-converted specs
- No file upload/download generation (separate plan)
- No batch mode (one API per pipeline invocation)
- Single Claude Code skill with phase routing (not 7 separate skills)

## Implementation Units

### Unit 1: Checkpoint Schema and State Machine

**Goal:** Define the pipeline's backbone - the checkpoint JSON that every phase reads and writes. Model after nightnight's `~/.osc/nightnight-session.json`.

**Files:**
- New: `internal/pipeline/checkpoint.go` - checkpoint types, read/write, phase enum
- New: `internal/pipeline/events.go` - JSONL event logger

**Approach:**

Checkpoint schema:
```go
type Checkpoint struct {
    APIName          string            `json:"api_name"`
    StartedAt        time.Time         `json:"started_at"`
    OutputDir        string            `json:"output_dir"`
    SpecPath         string            `json:"spec_path"`         // discovered spec file
    OverlayPath      string            `json:"overlay_path"`      // enrichment overlay
    ConventionsPath  string            `json:"conventions_path"`  // conventions cache
    EventsLogPath    string            `json:"events_log_path"`   // JSONL log
    Phases           map[string]Phase  `json:"phases"`
    Errors           []string          `json:"errors"`
}

type Phase struct {
    Status    string    `json:"status"`    // pending, in_progress, completed, failed
    StartedAt time.Time `json:"started_at"`
    Duration  string    `json:"duration"`
    Retries   int       `json:"retries"`
}
```

Checkpoint location: `~/.cache/printing-press/pipelines/<api-name>/checkpoint.json`
Events log: `~/.cache/printing-press/pipelines/<api-name>/events.jsonl`
Conventions cache: `~/.cache/printing-press/pipelines/<api-name>/conventions.json`

Phase order: `preflight -> discover -> scaffold -> enrich -> regenerate -> review -> ship`

Mechanical gate between phases: phase must be `completed` to advance. Failed after 2 retries = pipeline parks with morning report.

**Test scenarios:**
- Checkpoint round-trips through JSON marshal/unmarshal
- Phase state machine transitions correctly (pending -> in_progress -> completed)
- Events append to JSONL without corruption

**Verification:** `go test ./internal/pipeline/...`

---

### Unit 2: PREFLIGHT Phase

**Goal:** Validate environment before burning any tokens. Go installed, printing-press binary compiles, output dir clean.

**Files:**
- New: `internal/pipeline/preflight.go`

**Approach:**
1. Check `go version` succeeds
2. Check printing-press binary builds: `go build -o /tmp/pp-check ./cmd/printing-press`
3. Check output dir doesn't exist (or `--force` flag)
4. Check disk space > 500MB free
5. Write checkpoint with preflight=completed

**Test scenarios:**
- Preflight passes on healthy system
- Preflight fails gracefully when Go missing (clear error message)

**Verification:** Unit tests with mock exec

---

### Unit 3: DISCOVER Phase

**Goal:** Given an API name, find a usable OpenAPI spec. Check known-specs registry, then apis-guru, then WebSearch.

**Files:**
- New: `internal/pipeline/discover.go`
- Update: `skills/printing-press/references/known-specs.md` - add apis-guru URL patterns

**Approach:**

Discovery pipeline (ordered by reliability):
1. **Known-specs registry** - exact match on API name. Already has 12 verified URLs.
2. **apis-guru directory** - fetch `https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/<provider>/` and find the latest version. This covers 2,000+ APIs including all Google services.
3. **WebSearch** - `"<api-name> openapi spec" site:github.com OR site:swagger.io` with 3 strategies
4. **Fallback** - prompt the user (in interactive mode) or park (in autonomous mode)

Spec quality gate: after fetching, validate:
- Parses as valid OpenAPI (use existing `openapi.IsOpenAPI()` + `openapi.Parse()`)
- Has at least 3 endpoints
- Has at least 1 resource
- Has a base URL

If quality gate fails, try next discovery source.

**Conventions cache:** After successful parse, write to conventions cache:
- Auth type and schemes detected
- Number of resources/endpoints
- Pagination patterns found
- Global params detected
- Common prefix (if any)

**Test scenarios:**
- Discovery finds Petstore from known-specs
- Discovery finds Gmail from apis-guru
- Discovery falls back to WebSearch for unknown APIs
- Quality gate rejects specs with 0 endpoints

**Verification:** `go test ./internal/pipeline/...`, manual test with `gmail` and `stripe`

---

### Unit 4: SCAFFOLD Phase

**Goal:** Run the existing `printing-press generate` command on the discovered spec. Validate with 7 quality gates.

**Files:**
- New: `internal/pipeline/scaffold.go`
- Existing: `internal/generator/generator.go`, `internal/generator/validate.go`

**Approach:**
1. Read spec from checkpoint's `spec_path`
2. Parse with existing openapi/spec parser
3. Generate with existing generator
4. Run 7 quality gates
5. If gates fail: retry once with relaxed settings (skip lint gate), then park
6. Update checkpoint with scaffold=completed

This is a thin wrapper around the existing `generate` + `validate` pipeline.

**Test scenarios:**
- SCAFFOLD succeeds for Petstore, Gmail, Discord, Stytch specs
- SCAFFOLD retries on quality gate failure
- SCAFFOLD parks after 2 failures

**Verification:** `go test ./...`, regenerate all 4 test specs

---

### Unit 5: ENRICH Phase - The Intelligence Layer

**Goal:** Deep-read the original spec for hints the parser missed. Research the API's public docs. Produce a spec overlay that improves the generated CLI.

**Files:**
- New: `internal/pipeline/enrich.go`
- New: `internal/pipeline/overlay.go` - spec overlay types and merge logic

**Approach:**

The overlay format mirrors `spec.APISpec` but every field is optional. Non-nil fields override the original spec when merged.

```go
type SpecOverlay struct {
    Description  *string                       `yaml:"description,omitempty"`
    Resources    map[string]ResourceOverlay     `yaml:"resources,omitempty"`
}

type ResourceOverlay struct {
    Description *string                        `yaml:"description,omitempty"`
    Endpoints   map[string]EndpointOverlay     `yaml:"endpoints,omitempty"`
}

type EndpointOverlay struct {
    Description *string      `yaml:"description,omitempty"`
    Params      []ParamPatch `yaml:"params,omitempty"`
}

type ParamPatch struct {
    Name    string  `yaml:"name"`
    Default *string `yaml:"default,omitempty"` // e.g., "me" for userId
}
```

Enrichment strategies (each runs independently, results merged):

1. **Description hints** - scan param descriptions for default value hints. Gmail's userId says "The special value `me` can be used" - extract `me` as default.
2. **mediaUpload detection** - scan for `x-]google-*` extensions or multipart content types. Flag endpoints that support file upload (future: generate upload commands).
3. **Sync token patterns** - scan response schemas for `historyId`, `syncToken`, `nextSyncToken` fields. Flag resources that support incremental sync.
4. **Batch endpoint detection** - scan for paths ending in `:batchGet`, `:batchCreate`, or `batch*` with array request bodies.
5. **Better descriptions** - for endpoints with empty or generic descriptions, check if the operationId or path provides a clearer name.
6. **Endpoint grouping hints** - for flat resources with many endpoints (like settings), detect naming patterns that suggest sub-groups (e.g., `cse-*`, `send-as-*`).

Time budget: 15 minutes max. Each strategy gets 2 minutes. Partial results are written if time runs out.

**Test scenarios:**
- Gmail enrichment detects `me` default for userId
- Gmail enrichment detects mediaUpload on messages/send
- Overlay merges correctly with original spec (non-nil fields override)
- Overlay merge preserves unmodified fields

**Verification:** `go test ./internal/pipeline/...`, manual test with Gmail spec

---

### Unit 6: REGENERATE Phase

**Goal:** Merge the spec overlay with the original spec and re-run the generator.

**Files:**
- New: `internal/pipeline/regenerate.go`
- Update: `internal/spec/spec.go` - add `MergeOverlay()` method

**Approach:**
1. Load original spec from checkpoint
2. Load overlay from checkpoint
3. Merge: overlay fields replace original fields where non-nil
4. Re-run generator with merged spec (overwrite existing output)
5. Re-run 7 quality gates
6. If gates fail: fall back to original spec (enrichment broke something), log the error

**Test scenarios:**
- Regenerate with overlay produces different output than without
- Regenerate fallback works when overlay breaks compilation
- Quality gates still pass after regeneration

**Verification:** `go test ./...`

---

### Unit 7: REVIEW Phase - Static Quality Analysis

**Goal:** Automated quality assessment of the generated CLI. No API calls - static analysis only.

**Files:**
- New: `internal/pipeline/review.go`

**Approach:**

Review checks (binary pass/fail):
1. **Help completeness** - run `<cli> --help` and every `<cli> <resource> --help`. Verify exit code 0, non-empty output.
2. **Name quality** - no command name > 40 chars, no raw operationId passthrough (contains dots or underscores), no duplicate command names.
3. **Param coverage** - every GET endpoint has at least 1 non-positional param (if the original spec had query params). Flag endpoints where all params were filtered.
4. **Description quality** - no empty descriptions on top-level resources. No descriptions that are just the endpoint name repeated.
5. **Compilation clean** - `go vet` passes with zero warnings.
6. **Binary size** - warn if > 50MB (something is wrong).
7. **Doctor works** - `<cli> doctor` exits 0.

Review produces a quality score (0-100) and a list of issues. Score > 70 = PASS. Score < 70 = log issues in morning report but still PASS (don't block shipping over cosmetics).

**Test scenarios:**
- Review passes for a well-generated CLI
- Review catches empty descriptions
- Review catches overly long command names

**Verification:** `go test ./internal/pipeline/...`

---

### Unit 8: SHIP Phase

**Goal:** Finalize the generated CLI for human consumption.

**Files:**
- New: `internal/pipeline/ship.go`

**Approach:**
1. Run `git init` in output directory
2. Create initial commit: "Initial CLI generated by printing-press"
3. Validate goreleaser config: `goreleaser check` (if installed) or skip
4. Write morning report to `~/.cache/printing-press/pipelines/<api-name>/report.md`
5. Update checkpoint with ship=completed, total duration
6. Print summary to stderr

Morning report includes:
- API name and spec source
- Resources and endpoint count
- Quality score from REVIEW
- Enrichments applied
- Issues found
- Time per phase
- Next steps for the human (configure auth, test against real API, publish)

**Test scenarios:**
- SHIP creates git repo with initial commit
- Morning report contains all required sections
- Pipeline completion updates checkpoint correctly

**Verification:** `go test ./internal/pipeline/...`

---

### Unit 9: Pipeline Orchestrator and CLI Command

**Goal:** Wire all phases together. Add `printing-press print <api-name>` command. Handle chaining, resume, and error recovery.

**Files:**
- New: `internal/pipeline/pipeline.go` - orchestrator
- Update: `internal/cli/root.go` - add `print` command

**Approach:**

The `print` command:
```
printing-press print gmail [--output ./gmail-cli] [--force] [--resume] [--phase enrich]
```

Flags:
- `--output` - output directory (default: `./<api-name>-cli`)
- `--force` - overwrite existing output
- `--resume` - resume from checkpoint
- `--phase` - run only a specific phase (for debugging)

Orchestrator loop:
```go
func RunPipeline(apiName string, opts Options) error {
    checkpoint := loadOrCreateCheckpoint(apiName, opts)

    phases := []struct {
        name string
        fn   func(*Checkpoint) error
    }{
        {"preflight", runPreflight},
        {"discover", runDiscover},
        {"scaffold", runScaffold},
        {"enrich", runEnrich},
        {"regenerate", runRegenerate},
        {"review", runReview},
        {"ship", runShip},
    }

    for _, phase := range phases {
        if checkpoint.Phases[phase.name].Status == "completed" {
            continue // already done (resume mode)
        }

        logEvent(checkpoint, "phase_start", phase.name)
        checkpoint.Phases[phase.name] = Phase{Status: "in_progress", StartedAt: time.Now()}
        saveCheckpoint(checkpoint)

        var err error
        for attempt := 0; attempt < 3; attempt++ {
            err = phase.fn(checkpoint)
            if err == nil {
                break
            }
            checkpoint.Phases[phase.name] = Phase{Status: "failed", Retries: attempt + 1}
            logEvent(checkpoint, "phase_retry", phase.name, err.Error())
        }

        if err != nil {
            checkpoint.Errors = append(checkpoint.Errors, fmt.Sprintf("%s: %v", phase.name, err))
            saveCheckpoint(checkpoint)
            writeMorningReport(checkpoint) // partial report
            return fmt.Errorf("pipeline failed at %s: %w", phase.name, err)
        }

        checkpoint.Phases[phase.name] = Phase{Status: "completed", Duration: time.Since(...).String()}
        saveCheckpoint(checkpoint)
        logEvent(checkpoint, "phase_completed", phase.name)
    }

    writeMorningReport(checkpoint)
    return nil
}
```

**Test scenarios:**
- Full pipeline completes for Petstore (simplest spec)
- Resume skips completed phases
- Pipeline parks on repeated failure with morning report
- `--phase` runs only the specified phase

**Verification:** `go test ./...`, manual end-to-end: `printing-press print petstore`

---

### Unit 10: Claude Code Skill - The `print` Skill

**Goal:** Update the printing-press Claude Code skill to support the autonomous pipeline. User says "print me a gmail CLI" or `/printing-press gmail` and it runs the full pipeline.

**Files:**
- Update: `skills/printing-press/SKILL.md`

**Approach:**

Add a new workflow to the existing skill:

**Workflow 0 (updated): Autonomous Pipeline**
1. Parse API name from user's message
2. Run `printing-press print <api-name> --output ./<api-name>-cli`
3. Monitor output for phase completion messages
4. On completion: present morning report, show example commands
5. On failure: show error, offer to resume or debug specific phase

For overnight/autonomous mode (when invoked by nightnight-style chaining):
- Read checkpoint, determine current phase
- Execute next phase
- If more phases remain: CronCreate to chain back in 30 seconds
- If pipeline complete: write morning report, stop chaining

**Test scenarios:**
- User says "print me a stripe CLI" - full pipeline runs
- User says "resume gmail" - picks up from checkpoint

**Verification:** Manual test in Claude Code session

## Dependencies

```
Unit 1 (Checkpoint) - foundation, everything depends on this
Unit 2 (Preflight) - depends on Unit 1
Unit 3 (Discover) - depends on Unit 1
Unit 4 (Scaffold) - depends on Unit 1
Unit 5 (Enrich) - depends on Unit 1
Unit 6 (Regenerate) - depends on Unit 5
Unit 7 (Review) - depends on Unit 1
Unit 8 (Ship) - depends on Unit 1
Unit 9 (Orchestrator) - depends on Units 1-8
Unit 10 (Skill) - depends on Unit 9
```

Units 2-5, 7-8 can be built in parallel after Unit 1.
Unit 6 needs Unit 5's overlay types.
Unit 9 wires everything together.
Unit 10 is the skill layer on top.

## What "Done" Looks Like

```bash
# One command, one API name, finished CLI
printing-press print gmail

# Output:
# [preflight] Go 1.23 OK, disk OK
# [discover] Found Gmail spec via apis-guru (79 endpoints, OAuth2)
# [scaffold] Generated gmail-cli (7/7 gates passed)
# [enrich] Applied 12 enrichments (userId default=me, 3 upload hints, 8 better descriptions)
# [regenerate] Re-generated with enrichments (7/7 gates passed)
# [review] Quality score: 87/100 (2 minor issues)
# [ship] Initialized git repo, wrote morning report
#
# gmail-cli ready at ./gmail-cli/
# Resources: messages, labels, drafts, threads, history, settings, profile
# Auth: OAuth2 (run `gmail-cli auth login` to authenticate)
# Total time: 4m 23s

cd gmail-cli && go build -o ./gmail ./cmd/gmail-cli
./gmail auth login --client-id $GOOGLE_CLIENT_ID
./gmail messages list me --limit 5
./gmail doctor
```

That's not a scaffold. That's a finished CLI. From two words.

## Sources & References

### Internal
- OSC nightnight chaining pattern: `~/.claude/skills/osc-nightnight/SKILL.md` - checkpoint JSON, CronCreate chaining, mechanical budget gates, JSONL events
- OSC work execution: `~/.claude/skills/osc-work/SKILL.md` - phased CI execution, conventions cache, post-submission monitoring
- OSC plan discovery: `~/.claude/skills/osc-plan/SKILL.md` - 7-layer discovery pipeline, per-issue scoring
- Press generator: `internal/generator/generator.go` - 14 templates, 7 quality gates
- Press parser: `internal/openapi/parser.go` - 1,900 lines, extracts auth/resources/params/pagination
- Press spec types: `internal/spec/spec.go` - APISpec struct, validation
- Known specs registry: `skills/printing-press/references/known-specs.md` - 12 verified URLs

### External
- apis-guru OpenAPI directory: `https://github.com/APIs-guru/openapi-directory` - 2,000+ pre-converted specs
- Google Discovery -> OpenAPI conversions at `APIs-guru/openapi-directory/APIs/googleapis.com/`
