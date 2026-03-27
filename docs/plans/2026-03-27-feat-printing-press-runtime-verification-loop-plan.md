---
title: "Runtime Verification Loop for Printing Press"
type: feat
status: active
date: 2026-03-27
---

# Runtime Verification Loop for Printing Press

## Overview

The printing press generates CLIs that compile but don't work. The scorecard scores file contents, not behavior. The dogfood validates structure, not function. The result: a GitHub CLI scored 73/100 where the core feature (sync -> SQLite -> search) 404s on first call. 3.9% of commands were tested at runtime.

This plan adds a **runtime verification layer** to the press - a `printing-press verify` command that builds the generated binary and tests it against the **real API** when a key is available, falling back to a spec-derived mock server when it's not. Two paths, one bar: if the pass rate is below 80%, it enters an automated fix loop. The skill phase becomes mandatory and un-skippable.

The goal: every CLI that comes off the press actually works against the real API. Not "compiles." Not "passes mock tests." Actually works.

### Two verification paths

| | Has API key | No API key |
|--|------------|------------|
| **What runs** | Real API (read-only GETs) | Spec-derived mock server on localhost |
| **What it proves** | The CLI works in production | The CLI constructs correct requests and parses responses |
| **Confidence level** | High - real auth, real responses, real rate limits | Medium - correct shape but synthetic data |
| **When to use** | Always preferred. The skill already asks for the key in Phase 0.1 | Fallback when no key provided or API requires paid access |

The mock path is not the goal. It's the safety net. The real API path is the truth.

## Problem Statement

### What the press tests today (all static)

| Gate | What it checks | What it misses |
|------|---------------|----------------|
| `go build` | Syntax is valid | Endpoints 404, wrong paths, broken JSON parsing |
| `go vet` | No obvious bugs | Logic errors, wrong HTTP methods, missing params |
| Scorecard | Files contain expected strings | Strings exist but code never executes them |
| Dogfood | Paths in code match spec paths | Paths are correct but request construction is wrong |
| Verify | Dead flags/functions exist | Flags are wired but produce wrong behavior |

### What it should test (runtime)

| Test | What it proves |
|------|---------------|
| `cli --help` exits 0 | Binary runs |
| `cli resource list --dry-run` | Request path + method are correct |
| `cli resource get ID --json` against mock | Response parsing works |
| `cli sync --max-pages 1` against mock | Data pipeline writes to SQLite |
| `cli search "test"` after sync | FTS5 index works end to end |
| `cli sql "SELECT COUNT(*) FROM X"` | SQL command + domain tables work |
| `cli workflow-cmd --json` against mock | Workflow commands produce structured output |

## Proposed Solution

### New command: `printing-press verify`

```
printing-press verify --dir ./github-cli --spec /tmp/github-spec.json [--api-key TOKEN] [--fix] [--threshold 80]
```

If `--api-key` is provided (or the appropriate env var is set), tests run against the **real API** (read-only GETs only). Otherwise, falls back to a spec-derived mock server.

Three phases:
1. **Backend** - Detect API key. If present: real API (read-only). If absent: mock server on localhost.
2. **Run** - Execute every command against the backend, score pass/fail. Write commands get dry-run only on real API.
3. **Fix** (if `--fix` and score < threshold) - Auto-fix failures, re-run, repeat up to 3x

### Architecture

```
┌───────────────────────────────────────────────────────┐
│              printing-press verify                      │
│                                                        │
│  ┌───────────────────────────┐                         │
│  │  API Key provided?        │                         │
│  │  YES → Real API (GETs)    │                         │
│  │  NO  → Mock Server        │                         │
│  └─────────┬─────────────────┘                         │
│            ▼                                           │
│  ┌──────────────┐  ┌─────────────┐                    │
│  │  Runner      │  │  Fix Loop   │                    │
│  │  (tests      │──│  (patches + │                    │
│  │   every cmd) │  │   re-tests) │                    │
│  └──────────────┘  └─────────────┘                    │
│       ▲                    │                           │
│       │   Score < 80%      │                           │
│       └────────────────────┘                           │
└───────────────────────────────────────────────────────┘
```

**Real API path safety rules** (same as SKILL.md Phase 5.5):
- ONLY HTTP GET (list, get, search, doctor)
- NEVER POST, PUT, PATCH, DELETE
- `--limit 1` on all list calls
- `--max-pages 1` on sync
- 10s timeout per call, 2 minutes total
- Stop immediately on 401/403
- Print every command to stderr before executing

## Technical Approach

### Phase 1: Test Backend (Real API or Mock Server)

**File:** `internal/pipeline/testbackend.go`

Two backends, one interface:

```go
type TestBackend interface {
    BaseURL() string       // URL the CLI should hit
    Mode() string          // "live" or "mock"
    RequestLog() []Request // What was sent (mock only)
    Close()
}

func NewTestBackend(apiKey, envVarName string, spec *openapi.Spec) TestBackend {
    if apiKey != "" {
        // Real API - just return the real base URL
        // The CLI will hit the actual API with the real token
        return &LiveBackend{baseURL: spec.Servers[0].URL, apiKey: apiKey, envVar: envVarName}
    }
    // No key - spin up mock server
    return NewMockBackend(spec)
}
```

**Live backend:** Nothing to start. The CLI hits the real API. The runner enforces read-only by only testing GET commands. The `apiKey` is set as the env var the CLI expects (e.g., `GITHUB_TOKEN`).

**Mock backend (fallback):** Generate an `httptest.Server` from the OpenAPI spec:

1. Parse the spec (already done by dogfood/scorecard)
2. For each path + method, register a handler that:
   - Validates the request path matches the spec pattern
   - Returns a 200 with a synthetic JSON response built from the response schema
   - Logs the request for later assertion
3. Return the server URL to override the CLI's base URL

**Response generation heuristics** (from the SKILL.md Phase 4.5 spec):

| Field pattern | Generated value |
|---------------|----------------|
| `id`, `*_id` | Realistic format per API (snowflake, UUID, integer) |
| `name`, `title`, `login` | `"mock-entity-1"`, `"Test Issue Title"` |
| `created_at`, `updated_at` | `"2026-03-27T12:00:00Z"` |
| `state` (enum) | First enum value from spec |
| `url`, `html_url` | `"https://example.com/mock"` |
| Array fields | 2 items with above heuristics |
| Nested objects | Recursive generation |

**Key design decision:** The mock server lives in Go (not a separate process). It's started by `verify`, the generated CLI's base URL is overridden via env var, and it's torn down after tests complete.

```go
func NewMockServer(spec *openapi.Spec) (*httptest.Server, *RequestLog) {
    log := &RequestLog{}
    mux := http.NewServeMux()
    for path, ops := range spec.Paths {
        for method, op := range ops {
            pattern := convertOpenAPIPathToGoPattern(path) // {id} -> {id...}
            mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
                log.Record(r)
                resp := generateMockResponse(op.Responses["200"].Schema)
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(resp)
            })
        }
    }
    return httptest.NewServer(mux), log
}
```

### Phase 2: Command Runner

**File:** `internal/pipeline/runtime.go`

For every command in the generated CLI, run it against the mock server and score it.

**Discovery:** Parse `internal/cli/root.go` to find all registered commands. For each:

```go
type CommandTest struct {
    Name     string   // "repos issues list-for-repo"
    Args     []string // ["owner", "repo"]
    Flags    []string // ["--per-page", "2"]
    Kind     string   // "list", "get", "create", "workflow", "data-layer"
}
```

**Test matrix per command:**

Commands are classified as READ or WRITE. Only READ commands are tested at runtime. WRITE commands get dry-run only.

| Command type | Real API path | Mock path | What's tested |
|-------------|---------------|-----------|---------------|
| **GET** (list, get, search) | Runs against real API with `--limit 1` | Runs against mock | Full execute + JSON parse |
| **POST/PUT/PATCH/DELETE** (create, update, delete) | `--dry-run` ONLY - never sent | Runs against mock (safe - localhost) | Request construction only |
| **Local** (sql, health, trends, patterns) | Runs locally (no API call) | Same | Full execute |
| **Sync** | Real API with `--max-pages 1` (GETs only) | Mock with `--max-pages 1` | Data pipeline end to end |

**Per-command test suite:**

| Test | What runs | Pass criteria |
|------|-----------|--------------|
| Help | `cli <cmd> --help` | Exit 0, output contains "Usage:" |
| Dry-run | `cli <cmd> <args> --dry-run` | Exit 0, output contains HTTP method + path |
| Execute (GET only) | `cli <cmd> <args> --json --limit 1` | Exit 0, valid JSON output |
| Execute (WRITE) | SKIPPED on real API, mock only | Request body matches spec schema |
| Select | `cli <cmd> <args> --json --select id,name` | Output contains only selected fields |
| Error | `cli <cmd> --bad-flag` | Exit non-zero, stderr contains "unknown flag" |

**Data pipeline tests** (the critical path - runs on both paths):

| Test | What runs | Pass criteria |
|------|-----------|--------------|
| Sync | `cli sync --resources X --max-pages 1` (GETs only) | Exit 0, DB file exists, rows > 0 |
| SQL | `cli sql "SELECT COUNT(*) FROM X"` | Exit 0, output contains a number > 0 |
| Search | `cli search --query "test"` | Exit 0, output contains results |
| Health | `cli health` | Exit 0, output contains table names with row counts |

**READ-ONLY guarantee enforced in code:**

```go
func classifyCommand(cmd CommandTest, spec *openapi.Spec) string {
    // Look up the API path in the spec
    for path, ops := range spec.Paths {
        if matchesCommand(cmd, path) {
            if _, hasGet := ops["get"]; hasGet {
                return "read"
            }
            return "write"
        }
    }
    return "local" // sql, health, trends - no API call
}

func (r *Runner) Execute(cmd CommandTest, mode string) TestResult {
    kind := classifyCommand(cmd, r.spec)
    if mode == "live" && kind == "write" {
        // NEVER execute write commands against real API
        return r.DryRunOnly(cmd)
    }
    // ... proceed with execution
}
```

The runner literally cannot send a write request to the real API. It's enforced at the code level, not by convention.

**Scoring:**

```go
type VerifyResult struct {
    Command    string
    HelpPass   bool
    DryRunPass bool
    ExecPass   bool
    SelectPass bool
    ErrorPass  bool
    Score      int // 0-5 per command
}

// Aggregate
type VerifySummary struct {
    Total       int
    Passed      int     // score >= 3/5
    Failed      int     // score < 3/5
    Critical    int     // score 0 (command completely broken)
    PassRate    float64 // Passed / Total
    DataPipeline bool   // sync -> sql -> search all work
}
```

**Thresholds:**
- **PASS:** PassRate >= 80% AND DataPipeline == true AND Critical == 0
- **WARN:** PassRate >= 60% AND Critical <= 3
- **FAIL:** PassRate < 60% OR Critical > 3 OR DataPipeline == false

### Phase 3: Fix Loop

**File:** `internal/pipeline/fixloop.go`

When score < threshold AND `--fix` is set:

```
┌─────────────────────────────────────────┐
│  Iteration 1                            │
│  1. Classify failures by root cause     │
│  2. Generate targeted patches           │
│  3. Apply patches                       │
│  4. go build + go vet                   │
│  5. Re-run verify                       │
│  6. Score >= threshold? → DONE          │
│     Score < threshold? → Iteration 2    │
│                                         │
│  Max 3 iterations, then report failures │
└─────────────────────────────────────────┘
```

**Failure classification:**

| Failure pattern | Root cause | Auto-fix |
|----------------|------------|----------|
| Mock received no request | Wrong base URL or path | Fix URL construction in command file |
| Mock received request to wrong path | Path template wrong | Fix path string in command file |
| JSON parse error on response | Response struct doesn't match schema | Regenerate struct from spec |
| `--dry-run` shows wrong method | HTTP method wrong | Fix method in command file |
| Sync writes 0 rows | Upsert not wired to sync | Wire sync to call domain Upsert |
| SQL returns "no such table" | Migration missing | Add CREATE TABLE to store.go |
| Search returns 0 results | FTS5 not populated by triggers | Fix FTS5 triggers |
| `--select` returns all fields | filterFields not called | Wire selectFields into output path |
| Command exits non-zero | Various | Read stderr, classify, fix |

**Each fix is surgical:** Read the failing command file, identify the specific line, generate a patch. Not a rewrite.

### Integration with existing pipeline

**Update `fullrun.go`** to add verify as the final gate:

```go
// Current pipeline:
// generate -> gates(7) -> dogfood -> verify_static -> scorecard

// New pipeline:
// generate -> gates(7) -> dogfood -> verify_static -> scorecard -> VERIFY_RUNTIME -> [fix loop]
```

The runtime verify runs AFTER the scorecard, so the scorecard score is the "before" number. After the fix loop, re-run scorecard to get the "after" number. The delta is proof of work.

### Integration with the skill

**Update SKILL.md** to add Phase 4.8 after Phase 4.6:

```
# PHASE 4.8: RUNTIME VERIFICATION

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Run the runtime verifier:

\`\`\`bash
cd ~/cli-printing-press && ./printing-press verify \
  --dir ./<api>-cli \
  --spec /tmp/<api>-spec.json \
  --fix \
  --threshold 80
\`\`\`

### PHASE GATE 4.8

**STOP.** The verify command produces a PASS/WARN/FAIL verdict:

- **PASS** (>= 80% commands work, data pipeline works): Proceed to Phase 5.
- **WARN** (60-80%): Review failures. Fix manually if < 3 failures. Re-run.
- **FAIL** (< 60% or data pipeline broken): DO NOT proceed. Fix until PASS.

Tell the user: "Runtime verification: [X]% pass rate ([N]/[M] commands).
Data pipeline: [PASS/FAIL]. Fix loop ran [K] iterations. Proceeding to final report."
```

**Anti-shortcut rule to add:**

```
- "The scorecard is 73 so it's good enough" (The scorecard measures files, not behavior.
  Run `printing-press verify` - that measures behavior. A 73 scorecard with 0% verify
  is a CLI that looks good on paper and crashes on first use.)
```

## Implementation Phases

### Step 1: Test Backend (testbackend.go)

- [ ] `TestBackend` interface: `BaseURL()`, `Mode()`, `RequestLog()`, `Close()`
- [ ] `LiveBackend` - wraps real API URL, sets API key env var, enforces read-only
- [ ] `MockBackend` - httptest.Server from OpenAPI spec
- [ ] Mock response generation from spec schemas (realistic values per field type)
- [ ] Mock pagination (Link header with next page URL)
- [ ] Mock nested response wrappers (e.g., `{"workflow_runs": [...]}`)
- [ ] Request logging for assertion (mock path only)
- [ ] Tests: `testbackend_test.go` with petstore spec fixture

### Step 2: Command Runner (runtime.go)

- [ ] Parse root.go to discover all registered commands
- [ ] Classify each command as READ, WRITE, or LOCAL using the spec
- [ ] Build the generated CLI binary
- [ ] **Live mode:** Run READ commands against real API with `--limit 1`, WRITE commands get `--dry-run` only
- [ ] **Mock mode:** Run all commands against mock server
- [ ] Both modes: help, dry-run, select, error tests for every command
- [ ] Special data-pipeline test sequence (sync -> sql -> search -> health)
- [ ] Set base URL + API key via env vars
- [ ] Score each command, compute aggregate pass rate
- [ ] Output structured JSON result + human-readable table
- [ ] Report which mode was used (live vs mock) - live results are higher confidence
- [ ] Tests: `runtime_test.go` with a pre-generated petstore CLI fixture

### Step 3: Fix Loop (fixloop.go)

- [ ] Classify failures by pattern (path wrong, method wrong, parse error, etc.)
- [ ] Generate surgical patches per failure type
- [ ] Apply patches, rebuild, re-verify
- [ ] Max 3 iterations with diminishing returns detection
- [ ] Track before/after score per iteration
- [ ] Tests: `fixloop_test.go` with deliberately broken CLI fixture

### Step 4: Wire into pipeline (fullrun.go, cli/root.go)

- [ ] Add `verify` subcommand to CLI
- [ ] Add runtime verify as final gate in fullrun pipeline
- [ ] Re-run scorecard after fix loop to capture delta
- [ ] Update `print` (autonomous pipeline) command to include verify

### Step 5: Update SKILL.md

- [ ] Add Phase 4.8: Runtime Verification (mandatory, un-skippable)
- [ ] Add anti-shortcut rule about scorecard vs runtime
- [ ] Update Phase 5 final report template to include verify results
- [ ] Add verify pass rate to the "Tell the user" gate messages

### Step 6: Ensure generated CLIs support base URL override

- [ ] Check `config.go.tmpl` - does it read a `<API>_BASE_URL` env var?
- [ ] If not, add it to the template so the mock server can intercept requests
- [ ] This is a generator fix, not a verify fix - every future CLI gets it for free

## System-Wide Impact

### Interaction Graph
- `fullrun.go` calls `verify.go` after scorecard
- `verify.go` starts mock server, builds binary, runs commands via `exec.Command`
- `fixloop.go` reads verify results, edits generated Go files, triggers rebuild
- Scorecard re-runs after fix loop to capture improved score

### Error Propagation
- Mock server errors (port in use, spec parse failure) -> verify aborts with clear message
- Binary build failure after fix -> fix loop iteration fails, moves to next iteration
- All 3 fix iterations fail -> final FAIL verdict with remaining issues listed

### State Lifecycle Risks
- Mock server must be cleaned up on panic (defer server.Close())
- Fix loop edits files - if interrupted mid-edit, `go build` will catch syntax errors on next run
- SQLite DB created during sync test must be in temp dir, cleaned up after

## Acceptance Criteria

### Functional Requirements
- [ ] `printing-press verify --dir ./X --spec spec.json` produces PASS/WARN/FAIL verdict
- [ ] Mock server returns spec-derived responses for every path in the spec
- [ ] Every registered command is tested (help + dry-run + execute at minimum)
- [ ] Data pipeline test verifies sync -> sql -> search end to end
- [ ] `--fix` flag triggers auto-fix loop when score < threshold
- [ ] Fix loop patches at least: wrong paths, wrong methods, missing upsert wiring
- [ ] Scorecard re-runs after fix loop, delta is reported

### Quality Gates
- [ ] `go test ./internal/pipeline/...` passes with mock server tests
- [ ] Verify works on at least 2 reference CLIs (petstore + one real API)
- [ ] Fix loop successfully raises pass rate on a deliberately broken CLI
- [ ] End-to-end: `printing-press generate + verify --fix` produces a CLI where verify passes

### The Bar
A CLI that comes off the press with `verify --fix` should have:
- **>= 80% command pass rate** (help + dry-run + execute)
- **Data pipeline PASS** (sync populates tables, sql queries them, search finds results)
- **0 critical failures** (no command that completely crashes)
- **Scorecard >= 75** after fix loop

If it can't hit these numbers, the press tells you what's still broken and why. No more "73/100 PASS" when the core feature 404s.

## Success Metrics

| Metric | Before (today) | After (target) |
|--------|----------------|----------------|
| Commands tested at runtime | 3.9% (5/127) | 100% |
| Data pipeline verified | No | Yes (every run) |
| Time to discover "sync is broken" | Never (declared PASS) | < 2 minutes |
| Auto-fix coverage | 0% | 60%+ of common failures |
| False "PASS" rate | High | Near zero |

## Dependencies & Risks

- **Risk:** Mock server response fidelity. If mocks don't match real API response shapes, tests pass on mock but fail on real API. **Mitigation:** Generate mocks strictly from spec schemas, not hand-written.
- **Risk:** Fix loop creates worse code. Patches that fix one test break another. **Mitigation:** Re-run ALL tests after each patch, not just the failing one. Revert patch if net score decreases.
- **Risk:** Slow. Building + running 127 commands takes time. **Mitigation:** Parallel execution (commands are independent). Expect 30-60s for full verify on a typical CLI.
- **Dependency:** Generated CLIs must support base URL override via env var. This is a generator template change (Step 6).

## Sources

### Internal
- `internal/pipeline/scorecard.go` - Current scoring (string matching, 1,488 lines)
- `internal/pipeline/dogfood.go` - Current validation (static analysis, 522 lines)
- `internal/pipeline/verify.go` - Current proof system (static, 541 lines)
- `internal/pipeline/fullrun.go` - Pipeline orchestrator (584 lines)
- `internal/generator/validate.go` - 7 quality gates (123 lines)

### Learnings
- Post-mortem: `docs/plans/2026-03-27-fix-printing-press-post-generation-testing-gaps-plan.md`
- GitHub CLI run exposed: scorecard gaming, compilation-as-testing fallacy, skipped phases
- sync command 404 was the smoking gun - never caught because no runtime test exists
