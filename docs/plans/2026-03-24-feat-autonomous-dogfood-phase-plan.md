---
title: "Autonomous Dogfood Phase - Test Generated CLIs Against Real APIs"
type: feat
status: completed
date: 2026-03-24
---

# Autonomous Dogfood Phase - Test Generated CLIs Against Real APIs

## Overview

The printing press pipeline generates CLIs that compile and pass 7 quality gates - but it never actually uses them. The Review phase (Phase 4) only does static analysis: help output, name quality, description checks. It never calls an API. It never discovers that auth is misconfigured, pagination doesn't work, or table output chokes on real response shapes.

This plan adds autonomous dogfooding to the pipeline. After the CLI is generated and enriched, the press builds it, runs it against the real API (or a sandboxed version), captures what works and what breaks, and writes a structured dogfood report that feeds the Review phase's quality score.

Modeled on the OSC dogfooding patterns (osc-newfeature pre-plan scoring, osc-work post-build gates, evidence checkpoints) adapted for unattended pipeline execution.

## Problem Statement

Today's pipeline:
```
Preflight -> Scaffold -> Enrich -> Regenerate -> Review (static) -> Ship
```

The Review phase checks syntax, not behavior. It scores names and descriptions but can't answer:
- Does `doctor` actually report the API as reachable?
- Does `list` return a table or crash?
- Does auth setup work?
- Do pagination flags fetch all pages?
- Are exit codes correct for 401/404/429 responses?

These are the bugs that dogfooding always finds (ghcrawl #19, Homebrew #21781, Discord CLI, Gmail CLI). Static analysis can't catch them. Only running the CLI as a real user would catches them.

## Proposed Solution

Add a **Dogfood step** inside the Review phase (Phase 4) that:

1. Builds the generated CLI binary
2. Runs a tiered test suite against the API
3. Captures structured results (pass/fail/skip per test, real output, latency)
4. Writes `dogfood-results.json` and appends findings to `review.md`
5. Feeds a combined quality score (static 0-50 + dogfood 0-50 = total 0-100)

This happens autonomously inside the pipeline - no human interaction needed. The test suite adapts based on what the API supports (public vs auth-required, read-only vs read-write).

## Acceptance Criteria

- [ ] Review phase plan seed includes dogfood units alongside static units
- [ ] Dogfood tier system: Tier 1 (always safe), Tier 2 (read-only with auth), Tier 3 (write with test data)
- [ ] Tier 1 tests run without any API credentials (doctor, help, version, compile)
- [ ] Tier 2 tests run with credentials from env vars or config (list, get on safe resources)
- [ ] Tier 3 tests run only on sandboxed/test APIs (create, update, delete with cleanup)
- [ ] `dogfood-results.json` written with structured per-test results
- [ ] `review.md` includes combined score (static + dogfood)
- [ ] Dogfood results include real terminal output (not fake examples)
- [ ] Failed dogfood tests don't block the pipeline - they lower the score and get reported
- [ ] Budget-aware: dogfood phase has a 10-minute timeout per API

## Implementation Units

### Unit 1: Update Review Phase Seed Template

**Files:** `internal/pipeline/seeds.go`

**Approach:**
Update the `reviewSeedTemplate` to include dogfood units after the existing static units. The seed becomes:

```markdown
## Implementation Units

### Unit 1: Build the Generated CLI
[existing - compile and check binary exists]

### Unit 2: Static Quality Checks
[existing - help, names, descriptions, scoring]

### Unit 3: Dogfood Tier 1 - No Credentials Required
Run tests that need zero configuration:

a. Build the binary:
   cd {{.OutputDir}} && go build -o {{.APIName}}-cli ./cmd/{{.APIName}}-cli

b. Version check:
   ./{{.APIName}}-cli version
   Expected: prints version string, exit code 0

c. Doctor check:
   ./{{.APIName}}-cli doctor
   Expected: runs without crash
   Record: which checks pass/fail, base_url, config_path

d. Help completeness:
   ./{{.APIName}}-cli --help
   For each top-level resource: ./{{.APIName}}-cli <resource> --help
   Expected: non-empty output for every resource

e. Dry-run a mutation:
   Pick the first POST/PUT endpoint found. Run with --dry-run flag.
   Expected: shows request preview, does NOT send

f. Output mode flags:
   Run any list command with --json, --plain, --quiet
   Expected: each produces output, no crashes

Record all results to dogfood-results.json.

### Unit 4: Dogfood Tier 2 - Read-Only API Calls (if credentials available)

Check for API credentials:
  - Env vars matching the CLI's auth config (e.g., PETSTORE_API_KEY, STRIPE_API_KEY)
  - Config file at ~/.config/{{.APIName}}-cli/config.toml

If no credentials found:
  - Print "Tier 2 skipped: no credentials available"
  - Score Tier 2 as N/A (don't penalize)

If credentials found:
a. List command on the first safe resource:
   ./{{.APIName}}-cli <resource> list --limit 5
   Expected: returns data or empty list, not error
   Record: response shape, latency, exit code

b. Get command on a known ID (if list returned results):
   ./{{.APIName}}-cli <resource> get <first-id-from-list>
   Expected: returns single record
   Record: fields present, latency

c. Auth error handling:
   Run with intentionally wrong credentials (set env var to "invalid")
   Expected: exit code 4, not a crash or stack trace

d. Rate limit behavior (if API allows):
   Rapid-fire 5 identical requests
   Record: any 429 responses, retry behavior

### Unit 5: Dogfood Tier 3 - Write Operations (sandboxed APIs only)

Only run on APIs known to have test/sandbox modes:
  - Petstore (public test server, free writes)
  - Stripe (test mode with sk_test_ keys)
  - Stytch (test project)

If API is not in the sandbox-safe list: skip Tier 3 entirely.

If sandboxed:
a. Create a test resource:
   ./{{.APIName}}-cli <resource> create --name "printing-press-dogfood-test" ...
   Expected: returns created resource with ID
   Record: response, latency

b. Read it back:
   ./{{.APIName}}-cli <resource> get <created-id>
   Expected: matches what was created

c. Delete it (cleanup):
   ./{{.APIName}}-cli <resource> delete <created-id>
   Expected: exit code 0

d. Verify deletion:
   ./{{.APIName}}-cli <resource> get <created-id>
   Expected: exit code 3 (not found)

### Unit 6: Write Dogfood Results

Write dogfood-results.json:
{
  "api": "{{.APIName}}",
  "timestamp": "ISO8601",
  "tiers_run": [1, 2],
  "tiers_skipped": [3],
  "results": [
    {"test": "version", "tier": 1, "pass": true, "output": "petstore-cli 1.0.27", "latency_ms": 12},
    {"test": "doctor", "tier": 1, "pass": true, "output": "...", "latency_ms": 45},
    {"test": "list-pets", "tier": 2, "pass": true, "output": "...", "latency_ms": 230},
    {"test": "auth-error", "tier": 2, "pass": false, "output": "panic: ...", "issue": "crashes on invalid auth instead of exit code 4"}
  ],
  "score": {
    "tier1": 30,
    "tier2": 15,
    "tier3": 0,
    "total": 45,
    "max_possible": 50
  }
}

### Unit 7: Combined Review Score

Update review.md scoring to include dogfood results:

Static score (0-50):
  +10: compiles cleanly
  +10: all help commands work
  +10: no name quality issues
  +10: no empty descriptions
  +5: doctor works
  +5: binary < 50MB

Dogfood score (0-50):
  +10: version prints correctly
  +10: doctor reports API status
  +10: list/get returns data (Tier 2)
  +10: create/delete roundtrip (Tier 3)
  +5: --dry-run works
  +5: auth error handling correct

Combined: static + dogfood = total (0-100)

Grade: A (90+), B (75+), C (60+), D (40+), F (<40)
```

**Patterns to follow:** osc-newfeature dogfood scoring (complexity tiers), osc-work post-build checkpoints (gate tracker)

### Unit 2: Add Sandbox-Safe API Registry

**Files:** `internal/pipeline/seeds.go` or `internal/pipeline/discover.go`

**Approach:**
Add a `SandboxSafe` field to the known specs that indicates Tier 3 dogfooding is safe:

```go
type KnownSpec struct {
    URL         string
    Source      string
    SandboxSafe bool   // true = Tier 3 write tests are safe
    TestEnvVar  string // e.g., "STRIPE_TEST_KEY" for sandbox auth
}
```

Known sandbox-safe APIs:
- Petstore (public test server, no auth needed for writes)
- Stripe (test mode with sk_test_ prefixed keys)
- Stytch (test project environment)

All other APIs default to Tier 2 max (read-only dogfooding).

### Unit 3: Add Dogfood Timeout to Pipeline State

**Files:** `internal/pipeline/state.go`

**Approach:**
Add a `DogfoodTimeout` field to `PipelineState` (default: 10 minutes). The dogfood step in Review should respect this timeout:

```go
type PipelineState struct {
    // ... existing fields
    DogfoodTimeout time.Duration `json:"dogfood_timeout,omitempty"` // default 10m
}
```

The budget gate already caps the overall pipeline at 3 hours. The dogfood timeout prevents a single API test from hanging the pipeline (e.g., slow DNS, unreachable servers).

### Unit 4: Update SKILL.md Workflow 4 Documentation

**Files:** `skills/printing-press/SKILL.md`

**Approach:**
Update the pipeline phase table in Workflow 4 to reflect the enhanced Review phase:

```markdown
| Phase | What Happens |
|-------|-------------|
| 0. Preflight | Verify Go, download spec, cache conventions |
| 1. Scaffold | Generate CLI, pass 7 quality gates |
| 2. Enrich | Research API docs, discover missing endpoints, auth flows |
| 3. Regenerate | Merge enrichments, regenerate, re-validate |
| 4. Review | Static quality checks + **autonomous dogfooding against real API** |
| 5. Ship | Build, tag, generate release notes |
```

Add a note under the table:
```
The Review phase dogfoods the generated CLI in three tiers:
- Tier 1 (always): version, doctor, help, dry-run, output modes
- Tier 2 (if credentials available): list, get, auth error handling
- Tier 3 (sandbox APIs only): create/delete roundtrip with cleanup
```

## Scope Boundaries

- Don't add a new pipeline phase (keep 6 phases) - dogfooding enhances Review, doesn't replace it
- Don't require API credentials for Tier 1 - the pipeline must produce useful results with zero config
- Don't make write calls to production APIs - Tier 3 is sandbox-only, explicitly gated
- Don't block the pipeline on dogfood failures - failures lower the score, they don't stop shipping
- Don't change the Go CLI templates - this is about testing what's generated, not changing the generator
- Don't implement the full pipeline execution loop (that's already in the nightnight chaining plan)

## Technical Considerations

### Safety
- Tier 3 writes are limited to known sandbox APIs with explicit opt-in
- Delete/cleanup always runs even if create fails (idempotent cleanup)
- Auth credentials are read from env vars, never stored in state.json or plans
- dogfood-results.json may contain API response data - mark as sensitive in .gitignore

### Budget
- Dogfood timeout (10 min) prevents hanging on slow/unreachable APIs
- Overall pipeline budget gate (3h) still applies
- Tier 2/3 tests are skipped if no credentials - don't penalize the score for missing config

### Reliability
- Network failures during dogfooding are recorded as "skip" not "fail"
- Timeout is per-test (30 seconds default), not per-tier
- If the generated CLI crashes during dogfooding, that's a valid finding (recorded as fail)

## Dependencies

- Nightnight chaining plan (completed) - provides the pipeline execution loop
- Compound Engineering plugin - ce:plan and ce:work run the Review phase
- Known specs registry - provides sandbox-safe metadata per API

## Sources

- OSC dogfooding patterns: `~/.claude/skills/osc-newfeature/SKILL.md` (Phase 1b.5, complexity scoring)
- OSC post-build dogfooding: `~/.claude/skills/osc-work/SKILL.md` (Step 5e, evidence capture)
- Dogfood before PR feedback: `~/.claude/projects/-Users-mvanhorn/memory/feedback_dogfood_before_pr.md`
- Current Review seed: `internal/pipeline/seeds.go` (reviewSeedTemplate)
- Pipeline state: `internal/pipeline/state.go` (PipelineState, PhaseOrder)
- Existing dogfood plans: `docs/plans/2026-03-23-fix-press-dogfood-until-steinberger-quality-plan.md`
- Gmail dogfood gaps: `docs/plans/2026-03-24-fix-gmail-cli-dogfood-gaps-plan.md`
- Agent-browser research: web-only tool, not applicable to CLI testing
