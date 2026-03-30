---
title: "feat: Add adaptive rate limiting for sniffed APIs"
type: feat
status: active
date: 2026-03-29
origin: docs/brainstorms/2026-03-29-sniffed-api-rate-limiting-requirements.md
---

# feat: Add Adaptive Rate Limiting for Sniffed APIs

## Overview

Add proactive rate limiting to the printing-press generator and sniff skill. Generated CLIs for sniffed APIs will start at a conservative request rate and adaptively find the optimal speed. The sniff skill will add pacing instructions so Claude doesn't burn through rate limits during endpoint discovery.

## Problem Frame

Sniffed APIs are undocumented endpoints designed for one browser. Our CLIs make 100-500+ sequential requests during sync and 10-30 during sniff discovery, reliably hitting 429 rate limits. The current reactive retry-on-429 approach is insufficient — each retry wastes 5+ seconds, and some APIs escalate to IP bans. (see origin: docs/brainstorms/2026-03-29-sniffed-api-rate-limiting-requirements.md)

## Requirements Trace

- R1-R5. Adaptive rate limiter algorithm (conservative floor, ramp-up on success, halve on 429, per-session ceiling)
- R6-R8. Sniff skill pacing instructions
- R9-R12. Rate limiter in generated client template with `--rate-limit` flag
- R13-R14. Sync uses same limiter, shows effective rate in progress output

## Scope Boundaries

- No distributed rate limiting, no ceiling persistence across sessions, no response-time analysis
- No cache TTL changes, no per-endpoint rate config
- No `golang.org/x/time/rate` dependency — use stdlib for zero added deps in generated CLIs
- (see origin for full list)

## Context & Research

### Relevant Code and Patterns

- `internal/spec/spec.go:10-21` — `APISpec` struct, the template data context. Currently has no `SpecSource` field.
- `internal/generator/generator.go:123` — `client.go.tmpl` receives `*spec.APISpec` directly
- `internal/generator/generator.go:386-396` — `root.go.tmpl` receives anonymous struct wrapping `*spec.APISpec`
- `internal/generator/templates/client.go.tmpl:199` — existing 429 retry logic
- `internal/generator/templates/root.go.tmpl:49-63` — existing flag registration
- `skills/printing-press/SKILL.md:500-570` — sniff gate browser-use/agent-browser sections with basic `sleep` delays

### Key Findings

- Catalog metadata (`SpecSource`, `ClientPattern`) is NOT currently available during generation — `spec.APISpec` doesn't have these fields and the catalog entry isn't passed to the generator
- `golang.org/x/time` is not a current dependency and shouldn't be added to generated CLIs
- The sniff skill has `sleep 4` and `sleep 1` placeholders but no structured pacing

## Key Technical Decisions

- **Stdlib rate limiter, not `x/time/rate`**: Generated CLIs should have minimal dependencies. A 40-line token-bucket limiter using `time.Sleep` and `time.Since` avoids adding `golang.org/x/time` to every generated CLI's `go.mod`. The algorithm is simple enough that stdlib covers it.
- **`SpecSource` field on `APISpec`**: Add a `SpecSource string` field to `spec.APISpec`. The generate CLI command sets it from the catalog entry (when using `catalog show`) or from a new `--spec-source` flag. Templates use `{{.SpecSource}}` to conditionally set defaults.
- **Always include limiter, vary the default**: The rate limiter code is present in every generated CLI. When `SpecSource == "sniffed"`, the default rate is 2 req/s. Otherwise the default is 0 (disabled). Users override with `--rate-limit`. No template conditionals for inclusion — just for default values.
- **Adaptive is always-on**: When rate > 0, the ceiling-finder runs automatically. No separate "auto" mode. `--rate-limit N` sets the starting rate AND enables adaptive behavior. `--rate-limit 0` disables entirely.

## Open Questions

### Resolved During Planning

- **How does the template access `spec_source`?** Add `SpecSource string` to `spec.APISpec`. The CLI sets it before calling the generator. Templates access it as `{{.SpecSource}}`.
- **Compile-time or runtime?** Compile-time default via template `{{if eq .SpecSource "sniffed"}}`. The code is always present, only the default value changes.
- **Should `--rate-limit` accept `auto`?** No. Adaptive is always-on when rate > 0. Simpler mental model.
- **What SKILL.md sections need updating?** Step 2a.2 (page collection loop, lines 500-512), Step 2a.4 (fetch loop, lines 526-533), Step 2b.2 (agent-browser loop, lines 553-560), Step 2b.3 (response body fetching, lines 562-570).

### Deferred to Implementation

- Exact ramp-up thresholds (N=10 successes, 25% increase) may need tuning after testing against postman-explore
- The sniff skill pacing is behavioral guidance — Claude's adherence may vary; verify with a real sniff run

## High-Level Technical Design

> *This illustrates the intended approach and is directional guidance for review, not implementation specification. The implementing agent should treat it as context, not code to reproduce.*

```
Rate Limiter (stdlib, ~40 lines in client.go.tmpl):

  struct adaptiveLimiter {
      rate         float64       // current requests per second
      floor        float64       // starting rate (e.g. 2.0)
      ceiling      float64       // discovered ceiling (0 = unknown)
      successes    int           // consecutive successes since last 429
      rampAfter    int           // successes needed to increase (10)
      lastRequest  time.Time
  }

  Wait():
    delay = 1/rate seconds
    elapsed = time.Since(lastRequest)
    if elapsed < delay: sleep(delay - elapsed)
    lastRequest = now

  OnSuccess():
    successes++
    if successes >= rampAfter:
      newRate = rate * 1.25
      if ceiling > 0: cap at ceiling * 0.9
      rate = newRate
      successes = 0

  OnRateLimit():
    ceiling = rate
    rate = rate / 2
    if rate < 0.5: rate = 0.5  // absolute minimum
    successes = 0

Integration in do() loop:
  if limiter != nil: limiter.Wait()
  resp = httpClient.Do(req)
  if resp.StatusCode == 429: limiter.OnRateLimit()
  else if resp.StatusCode < 400: limiter.OnSuccess()
```

## Implementation Units

- [ ] **Unit 1: Add SpecSource to APISpec and plumb through generator**

**Goal:** Make catalog metadata available to templates at generation time.

**Requirements:** R10 (limiter active for sniffed APIs)

**Dependencies:** None

**Files:**
- Modify: `internal/spec/spec.go`
- Modify: `internal/cli/root.go` (generate command)
- Test: `internal/spec/spec_test.go`
- Test: `internal/generator/generator_test.go`

**Approach:**
- Add `SpecSource string` field to `APISpec` struct with `yaml:"spec_source,omitempty"`
- In the generate CLI command, after parsing the spec, set `apiSpec.SpecSource` from either:
  - The catalog entry's `SpecSource` (when generating from a catalog API)
  - A new `--spec-source` flag (when generating from a raw spec)
  - Default to empty string (limiter disabled)
- Verify templates can access `{{.SpecSource}}`

**Patterns to follow:**
- Existing `APISpec` fields like `Owner` which are optional metadata
- Existing `--force`, `--lenient` flags on the generate command

**Test scenarios:**
- Happy path: APISpec with SpecSource="sniffed" round-trips through YAML marshal/unmarshal
- Happy path: Generated CLI from a catalog entry with spec_source=sniffed has SpecSource set in template context
- Edge case: SpecSource empty string (default) — template conditionals produce disabled-limiter defaults
- Edge case: --spec-source flag overrides catalog value when both present

**Verification:**
- `go test ./internal/spec/... ./internal/generator/...` passes
- A generated CLI from the postman-explore catalog entry includes sniffed-aware defaults

- [ ] **Unit 2: Add adaptive rate limiter to client.go.tmpl**

**Goal:** Every generated CLI includes a stdlib rate limiter. Active by default for sniffed APIs, disabled for official.

**Requirements:** R1-R5, R9-R11

**Dependencies:** Unit 1

**Files:**
- Modify: `internal/generator/templates/client.go.tmpl`
- Test: `internal/generator/generator_test.go`

**Approach:**
- Add `adaptiveLimiter` struct and methods (Wait, OnSuccess, OnRateLimit) to the client template
- Add `limiter *adaptiveLimiter` field to `Client` struct
- In `New()`, initialize limiter based on template conditional: `{{if eq .SpecSource "sniffed"}}` sets floor to 2.0 req/s, else nil (disabled)
- In `do()`, call `limiter.Wait()` before sending the request, `limiter.OnSuccess()` on 2xx, `limiter.OnRateLimit()` on 429 (before the existing retry logic)
- The existing 429 retry loop stays — the limiter adjusts the rate, the retry loop handles the actual wait-and-retry

**Patterns to follow:**
- Existing `Client` struct fields (`DryRun`, `NoCache`, `cacheDir`)
- Existing `retryAfter()` helper

**Test scenarios:**
- Happy path: Generate a CLI from a sniffed spec, verify client.go contains `adaptiveLimiter` initialized with floor=2.0
- Happy path: Generate a CLI from an official spec, verify client.go has nil limiter
- Edge case: Limiter nil (disabled) — all requests proceed without delay
- Integration: The adaptive algorithm — after 10 successes the rate increases, after 429 the rate halves, ceiling is respected at 90%

**Verification:**
- Generated client.go compiles with no new external dependencies
- `go test ./internal/generator/...` passes

- [ ] **Unit 3: Add --rate-limit flag to root.go.tmpl**

**Goal:** Users can override the rate limit via CLI flag.

**Requirements:** R12

**Dependencies:** Unit 2

**Files:**
- Modify: `internal/generator/templates/root.go.tmpl`
- Modify: `internal/generator/templates/client.go.tmpl` (accept rate from flag)
- Test: `internal/generator/generator_test.go`

**Approach:**
- Add `rateLimit float64` to `rootFlags` struct
- Register `--rate-limit` persistent flag with default from template: `{{if eq .SpecSource "sniffed"}}2{{else}}0{{end}}`
- Pass the flag value to `client.New()` or set it on the client after creation
- Value of 0 means disabled (no limiter created)
- Value > 0 creates limiter with that floor rate

**Patterns to follow:**
- Existing `--timeout` flag which is similarly passed to the client

**Test scenarios:**
- Happy path: `--rate-limit 5` sets limiter floor to 5 req/s
- Happy path: `--rate-limit 0` disables limiter entirely
- Edge case: Default for sniffed API is 2, default for official is 0
- Edge case: Negative value treated as 0 (disabled)

**Verification:**
- Generated CLI's `--help` shows `--rate-limit` with correct default
- Flag value correctly propagates to client limiter

- [ ] **Unit 4: Show effective rate in sync progress**

**Goal:** Sync output displays the current request rate for observability.

**Requirements:** R13, R14

**Dependencies:** Unit 2

**Files:**
- Modify: `internal/generator/templates/sync.go.tmpl`
- Modify: `internal/generator/templates/client.go.tmpl` (expose current rate getter)

**Approach:**
- Add `Rate() float64` method to `adaptiveLimiter` that returns current rate (0 if nil)
- Add `RateLimit() float64` method to `Client` that delegates to limiter
- In sync progress output, include `[%.1f req/s]` when rate > 0
- Only show in human-friendly mode; JSON progress events include `rate_rps` field

**Patterns to follow:**
- Existing sync progress format: `{"event":"sync_progress","resource":"...","fetched":N}`

**Test scenarios:**
- Happy path: Sync with limiter active shows rate in progress — human-friendly format includes `[2.0 req/s]`
- Happy path: JSON progress event includes `rate_rps` field
- Edge case: Limiter disabled (rate=0) — no rate shown in output

**Verification:**
- Sync progress includes rate information when limiter is active
- JSON and human-friendly output both include rate data in their respective formats

- [ ] **Unit 5: Update SKILL.md sniff gate with pacing instructions**

**Goal:** Claude paces API probing during sniff discovery using the adaptive algorithm.

**Requirements:** R6-R8

**Dependencies:** None (parallel with Units 1-4)

**Files:**
- Modify: `skills/printing-press/SKILL.md`

**Approach:**
- Add a "Sniff Pacing" subsection after "If user approves sniff" (around line 402)
- Document the adaptive algorithm with sniff-phase defaults: floor=1 req/s, ramp after 5 successes, halve on 429
- Replace the hard-coded `sleep 4` and `sleep 1` in Step 2a.2 with a pacing instruction: "Apply the current sniff delay (starting at 1 second) between eval calls. If the previous call succeeded, decrease delay by 20% (min 0.3s). On 429, double the delay and log the event."
- Add 429 recovery guidance to Step 2a.4 and Step 2b sections
- Add guidance: "If you hit 3 consecutive 429s, pause for 30 seconds before continuing"

**Patterns to follow:**
- Existing skill instruction style — imperative, concise, with code examples

**Test scenarios:**
Test expectation: none — skill instruction changes are behavioral guidance for Claude, not compiled code. Validated through real sniff runs.

**Verification:**
- SKILL.md contains clear pacing instructions in the sniff gate section
- Instructions reference the adaptive algorithm with sniff-phase parameters

## System-Wide Impact

- **Generated CLI dependencies**: No new external dependencies added. Limiter uses stdlib only.
- **Backward compatibility**: Existing generated CLIs are unaffected until regenerated. The `--rate-limit` flag defaults to 0 (disabled) for non-sniffed APIs.
- **Template API surface**: `spec.APISpec` gains one field (`SpecSource`). All existing templates continue to work — the field is optional.
- **Sync behavior change**: Sync for sniffed APIs will be slower by default (2 req/s floor) but more reliable (no 429s). This is the intended tradeoff.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Adaptive algorithm too aggressive — still hits 429s at 2 req/s floor | Floor is conservative. If real-world testing shows issues, lower to 1 req/s. Easy to tune the default. |
| Adaptive algorithm too conservative — sync is painfully slow | Users override with `--rate-limit 5` or higher. Ceiling-finder will also ramp up over time. |
| SKILL.md pacing instructions ignored by Claude | The instructions are guidance, not enforcement. If Claude ignores them, the existing reactive 429 handling is the fallback. |

## Sources & References

- **Origin document:** [docs/brainstorms/2026-03-29-sniffed-api-rate-limiting-requirements.md](docs/brainstorms/2026-03-29-sniffed-api-rate-limiting-requirements.md)
- Related code: `internal/spec/spec.go`, `internal/generator/templates/client.go.tmpl`
- Related PRs: #60 (smart-default output), #61 (catalog schema with spec_source)
