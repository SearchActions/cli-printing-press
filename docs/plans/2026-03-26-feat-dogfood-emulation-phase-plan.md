---
title: "Dogfood Emulation Phase for Printing Press"
type: feat
status: active
date: 2026-03-26
---

# Dogfood Emulation Phase for Printing Press

## Overview

After the printing-press generates a CLI (Phase 4), there's no way to know if the commands actually work. We can't test with real API keys - we don't have them. But the OpenAPI spec already defines every request shape, response schema, error format, and pagination pattern. Vercel's [emulate](https://github.com/vercel-labs/emulate) project proves you can build production-fidelity API simulation from specs alone - stateful, cascading, zero-config.

This plan adds **Phase 4.5: Dogfood Emulation** between GOAT BUILD and FINAL STEINBERGER. It tests every generated command against spec-derived mock responses, scores each endpoint on a reality scale, writes a dogfood report, and auto-fixes what it finds.

## Problem Statement

Today's validation is:
- `go build ./...` - compiles? Yes/no.
- `go vet ./...` - static analysis passes? Yes/no.
- Steinberger scorecard - measures file presence and patterns, not runtime behavior.

None of these answer: **"If I had a real Discord bot token and ran `discord-cli channels messages list 123456789`, would it actually work?"**

Specific failure modes we can't catch today:
- Path parameter substitution broken (e.g., `{channel_id}` not replaced)
- Query param names wrong (spec says `limit`, code sends `count`)
- Response parsing assumes wrong JSON structure
- Pagination cursor field doesn't exist in response
- --select filter references fields that aren't in the response
- --stdin example JSON doesn't match the request body schema
- Commands that compile but would 400/422 on every real call

## Proposed Solution: Phase 4.5 Dogfood Emulation

### Core Principles (from Vercel emulate)

| Vercel Emulate Principle | Our Adaptation |
|---|---|
| **Zero-config** - works without setup | Generate mock responses from the OpenAPI spec - no API keys needed |
| **Stateful simulation** - changes persist | Not needed for CLI testing - each command is stateless |
| **Production fidelity** - mirrors real behavior | Generate responses that match the spec's response schemas exactly |
| **Cascading operations** - deletes remove dependents | Validate that DELETE commands handle 404 correctly (idempotent) |

### What We Test (5 Dimensions, 50 points max)

For every generated command, score on 5 dimensions:

#### 1. Request Construction (0-10)

**Test:** Run the command with `--dry-run` and inspect the request.

| Check | Points | How |
|---|---|---|
| Path params replaced correctly | 2 | Parse --dry-run output, verify no `{param}` literals remain |
| HTTP method is correct | 2 | Compare to spec's method for this endpoint |
| Required query params present | 2 | Check spec's `required: true` params are in the request |
| Body schema matches spec | 2 | If POST/PUT/PATCH, verify body fields match spec's requestBody |
| Auth header present | 2 | Verify Authorization header in --dry-run output |

#### 2. Response Parsing (0-10)

**Test:** Generate a synthetic response from the spec's response schema, pipe it through the command's output pipeline.

| Check | Points | How |
|---|---|---|
| Can parse 200 response | 3 | Generate JSON from response schema, verify command can render it |
| --json output is valid JSON | 2 | Pipe synthetic response, verify stdout is parseable JSON |
| --select works with real fields | 2 | Use --select with fields from the response schema |
| Table output doesn't crash | 2 | Render synthetic array response as table |
| Error responses handled | 1 | Generate 401/404/429 responses, verify error exit codes |

#### 3. Schema Fidelity (0-10)

**Test:** Compare generated command flags against the spec's parameters and request body.

| Check | Points | How |
|---|---|---|
| All required params are flags | 3 | Every spec `required: true` param has a CLI flag |
| Flag types match spec types | 2 | String params -> string flags, int params -> int flags |
| No hallucinated flags | 3 | Every CLI flag maps to a real spec parameter |
| Help text matches spec description | 2 | Flag --help description comes from spec, not invented |

#### 4. Example Quality (0-10)

**Test:** Validate that every example in --help and README would actually work.

| Check | Points | How |
|---|---|---|
| Example IDs are realistic format | 2 | Discord snowflake = 18 digits, not "abc123" |
| --stdin JSON examples match body schema | 3 | Parse example JSON, validate against spec's requestBody schema |
| Required flags present in examples | 3 | Example doesn't omit required params |
| Example commands parse without error | 2 | Run `<cli> <example> --dry-run` and verify no usage error |

#### 5. Workflow Integrity (0-10)

**Test:** Validate that hand-written workflow commands (Phase 0.5/0.7) make valid API calls.

| Check | Points | How |
|---|---|---|
| All API paths exist in spec | 3 | Every path the workflow command hits is a real spec endpoint |
| Query params are valid | 2 | Params sent to the API match spec's query parameters |
| Response fields accessed exist | 3 | Fields the workflow reads from responses are in the response schema |
| Joins are valid | 2 | Cross-entity references (e.g., message.author_id -> user.id) exist in spec |

### Scoring

**Per-command score:** Sum of 5 dimensions (0-50 max)

**Aggregate scores:**
- **Command pass rate:** Commands scoring >= 35/50 (70%)
- **Critical failure count:** Commands scoring < 25/50 (hallucinated or broken)
- **Overall dogfood score:** Average across all commands (0-50)

**Thresholds:**
- **Pass rate >= 90%** and **0 critical failures** -> Dogfood PASS
- **Pass rate >= 70%** and **<= 3 critical failures** -> Dogfood WARN (auto-fix, then re-score)
- **Pass rate < 70%** or **> 3 critical failures** -> Dogfood FAIL (stop, report issues)

---

## Technical Approach

### Step 4.5a: Generate Synthetic Responses from Spec

Read the OpenAPI spec and for each endpoint's response schema, generate a realistic JSON response:

```go
// For each endpoint in the spec:
// 1. Read the 200/201 response schema
// 2. Generate a realistic JSON object with:
//    - String fields: realistic values based on field name
//      (e.g., "id" -> "123456789012345678", "name" -> "general", "content" -> "Hello world")
//    - Integer fields: realistic values (e.g., "type" -> 0, "position" -> 1)
//    - Boolean fields: true/false
//    - Array fields: 2-3 items
//    - Nested objects: recursively generate
// 3. Save to /tmp/<api>-cli-mocks/<endpoint-path>.json
```

**Field value heuristics** (domain-aware, not random):

| Field name pattern | Generated value |
|---|---|
| `id`, `*_id` | `"123456789012345678"` (Discord snowflake format) |
| `name`, `username` | `"test-user"`, `"general"` |
| `content`, `description`, `body` | `"This is a test message for dogfood validation"` |
| `timestamp`, `created_at` | `"2026-03-26T12:00:00.000Z"` |
| `type` (enum) | First enum value from spec |
| `url`, `avatar_url` | `"https://example.com/test"` |
| `email` | `"test@example.com"` |
| `count`, `position` | `1`, `0` |
| `*` (string, no pattern) | `"test-value"` |

### Step 4.5b: Run Dry-Run Validation on Every Command

For each generated command:

```bash
# 1. Run with --dry-run to verify request construction
<api>-cli <resource> <action> <required-args> --dry-run 2>&1

# 2. Check: no {param} literals in path
# 3. Check: method matches spec
# 4. Check: required params present
# 5. Score Request Construction dimension
```

### Step 4.5c: Feed Synthetic Responses Through Output Pipeline

For commands with list/get endpoints:

```bash
# 1. Generate mock response from spec
# 2. Test JSON output mode
echo '<mock-response>' | <api>-cli <resource> list --json 2>/dev/null

# 3. Test --select with real field names
echo '<mock-response>' | <api>-cli <resource> list --json --select id,name

# 4. Test table output (default mode)
echo '<mock-response>' | <api>-cli <resource> list
```

**Note:** This requires the CLI to support piped input for GET responses - currently it doesn't. The dogfood step should use the `printOutput` function directly by importing the CLI package, or test via a thin Go test harness.

### Step 4.5d: Validate Examples and Help Text

```bash
# For each command with examples in --help:
# 1. Extract example commands from --help output
# 2. Run each example with --dry-run
# 3. Verify no "usage error" exit code
# 4. For --stdin examples, validate the JSON against the spec's requestBody schema
```

### Step 4.5e: Score and Report

Generate the dogfood report with:
- Per-command scores (all 5 dimensions)
- Top failures (commands scoring < 25/50)
- Hallucination detection (flags/fields that don't exist in spec)
- Recommendations (what to fix, what to regenerate, what to hand-write)

### Step 4.5f: Auto-Fix

For each fixable issue found:

| Issue Type | Auto-Fix |
|---|---|
| Placeholder example values ("abc123") | Replace with realistic domain values from spec |
| Missing required flags in examples | Add required flags with mock values |
| --stdin JSON doesn't match body schema | Regenerate from spec's requestBody schema |
| Lazy 1-word descriptions | Fetch description from spec and update |
| Path param not substituted | Fix replacePathParam call |
| Hallucinated flag (not in spec) | Remove the flag |

After auto-fix, re-run the dogfood test and report the before/after scores.

---

## Phase Gate 4.5

**STOP.** Verify ALL of these before proceeding to Phase 5:
1. Every generated command tested with --dry-run (no compilation-only validation)
2. Synthetic responses generated for all list/get endpoints
3. Per-command scores computed on all 5 dimensions
4. Dogfood report written with pass rate, critical failures, recommendations
5. Auto-fixes applied for fixable issues
6. Re-score after fixes shows improvement
7. `go build ./...` still passes after fixes
8. Final dogfood score meets PASS or WARN threshold

**Write Phase 4.5 Artifact:** Run the Artifact Writing plan generator with all dogfood results as input. Write to `~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-dogfood-report.md`. Include:
- Per-command score table (all commands, all 5 dimensions)
- Top failures with root cause
- Hallucination list (flags/fields not in spec)
- Auto-fixes applied with before/after
- Recommendations for manual fixes
- Overall dogfood score and pass/fail status

Tell the user: "Phase 4.5 complete: Dogfood score [X]/50 avg across [N] commands. Pass rate: [Y]%. Critical failures: [Z]. Auto-fixed [K] issues. [PASS/WARN/FAIL]. Proceeding to final Steinberger."

---

## Where This Fits in the Pipeline

```
Phase 0 -> 0.5 -> 0.7 -> 1 -> 2 -> 3 -> 4 -> [4.5 DOGFOOD] -> 5
                                                 ^^^^^^^^^^^
                                                 NEW: Test every
                                                 endpoint against
                                                 spec-derived mocks
```

**7 plan artifacts per run now** (was 6):

```
Phase 0   -> visionary-research.md
Phase 0.5 -> power-user-workflows.md
Phase 0.7 -> data-layer-spec.md
Phase 1   -> research.md
Phase 3   -> audit.md
Phase 4   -> goat-build-log.md
Phase 4.5 -> dogfood-report.md          <- NEW
```

---

## Implementation Plan

### Session 1: Add Phase 4.5 to SKILL.md (30-45 min)

**File:** `~/.claude/skills/printing-press/SKILL.md`

1. Insert Phase 4.5 section between Phase 4 Gate and Phase 5
2. Update phase flow diagram (add Phase 4.5, update time estimate)
3. Add dogfood-report.md to artifact list
4. Add anti-shortcut rules:
   - "The CLI compiles so it works" (Compilation proves syntax, not semantics. Run the dogfood.)
   - "We can't test without API keys" (The spec defines response schemas. Generate mocks. Test against them.)
   - "The dry-run looks right" (Dry-run validates request construction. Feed synthetic responses to validate output parsing too.)

### Session 2: Build the Dogfood Harness (1-2 hours)

The dogfood test is executed by the LLM during the printing-press run, not by Go code. The SKILL.md instructions tell the LLM to:

1. **Read the spec** to get response schemas for each endpoint
2. **Run each command with --dry-run** and parse the output
3. **Generate synthetic responses** using the field value heuristics table
4. **Validate --stdin examples** against the spec's requestBody
5. **Score each dimension** using the rubric
6. **Write the dogfood report** as a plan artifact
7. **Auto-fix** issues and re-score

This is all LLM-driven, not a separate binary. The LLM reads the spec, generates mocks, runs bash commands, and scores results.

### Session 3: Test on Discord CLI (1 hour)

Re-run the dogfood step on the existing discord-cli to validate:
- How many of 330 commands pass --dry-run?
- Do the --stdin examples in README match the spec?
- Are there hallucinated flags?
- What's the overall dogfood score?

---

## Acceptance Criteria

- [ ] Phase 4.5 section added to SKILL.md between Phase 4 Gate and Phase 5
- [ ] 5-dimension scoring rubric documented (Request, Response, Schema, Example, Workflow)
- [ ] Synthetic response generation heuristics documented (field name -> realistic value)
- [ ] Auto-fix table documented (issue type -> fix action)
- [ ] Phase Gate 4.5 with 8 verification items
- [ ] Dogfood report artifact template defined
- [ ] Phase flow diagram updated (8 phases now)
- [ ] Time estimate updated
- [ ] 3 new anti-shortcut rules added
- [ ] Discord CLI dogfood produces a score for 330+ commands

## Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Dogfood takes too long (330 commands) | High | Medium | Sample 50 commands (all workflows + random 43 generated). Full scan optional. |
| Synthetic responses don't match real API | Medium | Low | Generated from the actual spec schema - fidelity is high for structure |
| Auto-fix introduces new bugs | Medium | Medium | Re-run `go build` + `go vet` after every fix |
| --dry-run output format changes | Low | Low | Parse stderr output with flexible regex |
| Some commands can't be tested (e.g., file uploads) | Medium | Low | Mark as "untestable" with reason, exclude from pass rate |

## Sources

- [vercel-labs/emulate](https://github.com/vercel-labs/emulate) - Core principles: zero-config, stateful, production-fidelity
- Printing-press SKILL.md: `~/.claude/skills/printing-press/SKILL.md`
- Discord CLI: `~/cli-printing-press/discord-cli/`
- OpenAPI spec: `/tmp/printing-press-spec-discord.json`
