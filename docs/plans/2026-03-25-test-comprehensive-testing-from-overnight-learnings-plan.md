---
title: "Comprehensive Testing Plan - Learning from Overnight Hardening"
type: test
status: active
date: 2026-03-25
---

# Comprehensive Testing Plan - Learning from Overnight Hardening

## Overview

Last night's overnight hardening run exposed three classes of bugs and revealed significant test coverage gaps. This plan turns those learnings into permanent regression tests so the same bugs never resurface.

## Learnings from Last Night

### Bug 1: README Template Crash on Empty Auth (FIXED, NEEDS REGRESSION TEST)

`readme.md.tmpl` called `index .Auth.EnvVars 0` without checking if `EnvVars` was empty. Crashed on Fly.io and Telegram (APIs with no auth). Fixed with conditional guards, but no test catches this - if someone removes the guard, it silently breaks again.

**Lesson:** Every template that accesses optional fields needs a test with that field empty.

### Bug 2: PagerDuty Spec Schema Resolution Error (UNFIXED)

kin-openapi parser fails on `#/components/requestBodies/OrchestrationCacheVariableDataPutResponse`. Our parser is strict - any bad `$ref` kills the entire parse. This blocked generating PagerDuty CLI.

**Lesson:** We need a "lenient parse" mode that skips unresolvable `$ref`s instead of failing. And we need tests that exercise real-world specs with known issues.

### Bug 3: Intercom Spec Missing Component (UNFIXED)

Similar to PagerDuty - `custom_attributes` component not found. Same class of issue.

**Lesson:** Real-world OpenAPI specs are messy. Our parser needs to handle gracefully, not crash.

## Current Coverage Snapshot

| Package | Coverage | Missing Tests |
|---------|----------|---------------|
| catalog | 78.4% | Good enough |
| generator | 61.1% | `validate.go` has zero tests |
| openapi | 79.3% | `detect.go` has zero tests |
| pipeline | 47.9% | `discover.go`, `overlay.go`, `pipeline.go` have zero tests |
| spec | 63.6% | Decent |
| cli | 0.0% | No test files (CLI commands) |
| cmd | 0.0% | No test files (main entry) |

**Files with zero test coverage:**
- `internal/generator/validate.go` - the 7 quality gates
- `internal/openapi/detect.go` - OpenAPI format detection
- `internal/pipeline/discover.go` - API spec discovery + KnownSpecs
- `internal/pipeline/overlay.go` - spec overlay/merge
- `internal/pipeline/pipeline.go` - pipeline init/resume

## Implementation Units

### Unit 1: Template Regression Tests

**Goal:** Never ship a template that crashes on optional/empty fields again.

**File:** `internal/generator/generator_test.go` (extend existing)

**Tests to add:**

```
TestGenerateWithNoAuth
  - Create an APISpec with Auth.EnvVars = nil (empty)
  - Generate all templates
  - Verify 7/7 quality gates pass
  - This would have caught the readme.md.tmpl crash

TestGenerateWithNoResources
  - Create an APISpec with Resources = nil
  - Verify generation handles gracefully (empty CLI with doctor only)

TestGenerateWithEmptyDescription
  - APISpec with Description = ""
  - Verify README doesn't have empty sections

TestGenerateWithSpecialCharsInName
  - APISpec with Name containing dots, slashes, unicode
  - Verify toCamel/flagName sanitization works end-to-end

TestGenerateWithOwnerField
  - APISpec with Owner = "mvanhorn"
  - Verify go.mod, goreleaser, README all use "mvanhorn" not "USER"

TestGenerateWithEmptyOwner
  - APISpec with Owner = ""
  - Verify it defaults to "USER" (backward compat)
```

**Verification:** `go test ./internal/generator/ -run TestGenerate -v` passes. The NoAuth test generates a CLI that compiles.

### Unit 2: Validate.go Tests (7 Quality Gates)

**Goal:** Test the quality gate runner independently - currently 0% coverage.

**File:** `internal/generator/validate_test.go` (new)

**Tests to add:**

```
TestValidatePassesOnGoodProject
  - Generate petstore CLI to a temp dir
  - Run Validate() on it
  - Assert all 7 gates pass

TestValidateFailsOnBadGoCode
  - Create a temp dir with syntactically invalid Go code
  - Run Validate()
  - Assert it fails at "go vet" or "go build" gate
  - Verify the error message names the failed gate

TestValidateFailsOnMissingBinary
  - Create a temp dir with valid Go but no main.go
  - Run Validate()
  - Assert it fails at "binary build" gate

TestValidateReturnsGateResults
  - Verify the return value includes which gates passed/failed
  - Not just pass/fail but which specific gate
```

**Verification:** `go test ./internal/generator/ -run TestValidate -v` passes.

### Unit 3: Detect.go Tests (Format Detection)

**Goal:** Test OpenAPI format detection - currently 0% coverage.

**File:** `internal/openapi/detect_test.go` (new)

**Tests to add:**

```
TestIsOpenAPI_ValidOpenAPI3JSON
  - Feed it `{"openapi": "3.0.0", "info": {"title": "Test"}, "paths": {}}`
  - Assert returns true

TestIsOpenAPI_ValidSwagger2JSON
  - Feed it `{"swagger": "2.0", "info": {"title": "Test"}, "paths": {}}`
  - Assert returns true

TestIsOpenAPI_ValidOpenAPI3YAML
  - Feed it `openapi: "3.0.0"\ninfo:\n  title: Test`
  - Assert returns true

TestIsOpenAPI_InternalYAML
  - Feed it internal YAML spec format (name:, resources:)
  - Assert returns false

TestIsOpenAPI_RandomJSON
  - Feed it `{"foo": "bar"}`
  - Assert returns false

TestIsOpenAPI_EmptyInput
  - Feed it empty bytes
  - Assert returns false (no panic)

TestIsOpenAPI_BinaryGarbage
  - Feed it random binary data
  - Assert returns false (no panic)
```

**Verification:** `go test ./internal/openapi/ -run TestIsOpenAPI -v` passes.

### Unit 4: Discover.go Tests (Spec Discovery)

**Goal:** Test the spec discovery registry and apis-guru fallback.

**File:** `internal/pipeline/discover_test.go` (new)

**Tests to add:**

```
TestDiscoverSpec_KnownAPI
  - Call DiscoverSpec("petstore")
  - Assert returns a valid URL containing "petstore"

TestDiscoverSpec_UnknownAPI
  - Call DiscoverSpec("zzz-nonexistent-api-zzz")
  - Assert returns an error

TestDiscoverSpec_CaseInsensitive
  - Call DiscoverSpec("GitHub") (capital G)
  - Assert returns same result as DiscoverSpec("github")

TestKnownSpecsRegistry_AllURLsHTTPS
  - Iterate all entries in KnownSpecs
  - Assert every URL starts with "https://"

TestKnownSpecsRegistry_NoDuplicates
  - Iterate all entries in KnownSpecs
  - Assert no duplicate URLs

TestApisGuruPattern
  - Call ApisGuruPattern("stripe.com", "v1")
  - Assert returns a well-formed URL
```

**Verification:** `go test ./internal/pipeline/ -run TestDiscover -v` passes.

### Unit 5: Lenient Parser Mode (Fix PagerDuty/Intercom Spec Issues)

**Goal:** Add a `--lenient` flag to the parser that skips unresolvable `$ref`s instead of crashing. This unblocks PagerDuty and Intercom CLIs.

**File:** `internal/openapi/parser.go` (modify)

**Approach:**

The kin-openapi library has a `loader.IsExternalRefsAllowed` option. But the issue is broken internal refs, not external. The fix is to catch parse errors, log warnings for bad refs, and continue with a partial spec (skipping endpoints that reference broken components).

```go
// In Parse(), wrap the loader call:
doc, err := loader.LoadFromData(data)
if err != nil {
    if lenient {
        // Log the error and try again with validation disabled
        fmt.Fprintf(os.Stderr, "warning: spec has errors, trying lenient mode: %v\n", err)
        loader.IsExternalRefsAllowed = true
        doc, err = loader.LoadFromDataWithOptions(data, &openapi3.ValidationOptions{
            DisableSchemaPatternValidation: true,
        })
    }
    if err != nil {
        return nil, err
    }
}
```

**Tests:**

```
TestParseLenient_PagerDutySpec
  - Fetch the real PagerDuty spec
  - Parse with lenient=true
  - Assert it produces a valid APISpec with >0 resources
  - Assert warnings were logged

TestParseLenient_IntercomSpec
  - Fetch the real Intercom spec
  - Parse with lenient=true
  - Assert it produces a valid APISpec with >0 resources

TestParseStrict_InvalidRef
  - Create a minimal spec with a bad $ref
  - Parse with lenient=false
  - Assert it returns an error

TestParseLenient_InvalidRef
  - Same spec, parse with lenient=true
  - Assert it succeeds with a warning
```

**File:** `internal/openapi/parser_test.go` (extend)

**Verification:** PagerDuty and Intercom specs parse successfully in lenient mode. `go test ./internal/openapi/ -run TestParseLenient -v` passes.

### Unit 6: End-to-End Gauntlet Test

**Goal:** Automate the 10-API gauntlet as a Go test so it runs in CI, not just manually.

**File:** `internal/generator/gauntlet_test.go` (new)

**Approach:**

This is a slow integration test gated by a build tag (`//go:build gauntlet`) so it doesn't run on every `go test ./...` but does run in CI and manually via `go test -tags gauntlet ./internal/generator/`.

```
TestGauntlet_Petstore (always runs - fast, local fixture)
  - Parse testdata/petstore.json
  - Generate to temp dir
  - Validate 7/7 gates

TestGauntlet_10APIs (build tag: gauntlet)
  - For each of the 10 gauntlet spec URLs:
    - Fetch spec
    - Parse (lenient mode if needed)
    - Generate to temp dir
    - Validate 7/7 gates
  - Assert 10/10 pass
  - If any fail, report which gate failed for which API
```

**Verification:** `go test -tags gauntlet ./internal/generator/ -run TestGauntlet -v -timeout 10m` passes 10/10.

### Unit 7: Template Feature Verification Tests

**Goal:** Verify that the Phase 0 template improvements (--select, error hints, generation comments, Owner) actually appear in generated output.

**File:** `internal/generator/generator_test.go` (extend)

**Tests to add:**

```
TestGeneratedOutput_HasSelectFlag
  - Generate petstore CLI
  - Read the generated root.go
  - Assert it contains "select" flag registration

TestGeneratedOutput_HasErrorHints
  - Generate petstore CLI
  - Read the generated helpers.go
  - Assert it contains "hint:" in the 401/404 error handlers

TestGeneratedOutput_HasGenerationComment
  - Generate petstore CLI
  - Read all generated .go files
  - Assert each contains "Code generated by CLI Printing Press"

TestGeneratedOutput_UsesOwnerInImports
  - Generate with Owner = "testowner"
  - Read go.mod
  - Assert it contains "github.com/testowner/"

TestGeneratedOutput_READMEHasQuickStart
  - Generate petstore CLI
  - Read README.md
  - Assert it contains "Quick Start" section
  - Assert it contains "doctor" command reference
  - Assert it contains "Output Formats" section
  - Assert it contains "Agent Usage" section
```

**Verification:** `go test ./internal/generator/ -run TestGeneratedOutput -v` passes.

## Coverage Targets

After this plan is executed:

| Package | Current | Target | Key files covered |
|---------|---------|--------|-------------------|
| catalog | 78.4% | 78% (hold) | Already good |
| generator | 61.1% | **75%+** | validate.go, template regression, feature verification |
| openapi | 79.3% | **85%+** | detect.go, lenient parser |
| pipeline | 47.9% | **60%+** | discover.go |
| spec | 63.6% | 63% (hold) | Already decent |

## Acceptance Criteria

- [ ] Template regression tests catch the readme.md.tmpl empty-auth crash
- [ ] validate.go has 3+ tests covering pass, fail, and gate identification
- [ ] detect.go has 6+ tests covering OpenAPI 3, Swagger 2, YAML, edge cases
- [ ] discover.go has 4+ tests covering known APIs, unknown APIs, registry integrity
- [ ] Lenient parser mode parses PagerDuty and Intercom specs without crashing
- [ ] Gauntlet test exists as build-tagged integration test (10/10 pass)
- [ ] Template feature verification tests confirm --select, hints, comments, Owner, README
- [ ] Overall `go test ./...` passes
- [ ] Generator coverage reaches 75%+
- [ ] OpenAPI coverage reaches 85%+

## Scope Boundaries

- Do NOT add tests for the generated CLI code (plaid-cli, pipedrive-cli internal code)
- Do NOT refactor existing code - this is purely a testing plan
- Do NOT add new template features - test what exists
- The lenient parser (Unit 5) is the ONE code change allowed because it directly unblocks testability

## Sources

- Overnight results: `docs/plans/overnight-hardening-results.md`
- Bug 1 fix commit: `2370b45` (readme.md.tmpl empty EnvVars guard)
- Bug 2 spec: PagerDuty `OrchestrationCacheVariableDataPutResponse` bad ref
- Bug 3 spec: Intercom `custom_attributes` missing component
- Existing tests: `internal/generator/generator_test.go`, `internal/openapi/parser_test.go`
- Coverage data: `go test ./... -cover` run 2026-03-25
