---
title: "feat: Per-endpoint version header routing and auth inference from Authorization header params"
type: feat
status: active
date: 2026-04-05
origin: docs/retros/2026-04-05-cal-com-run2-retro.md (findings F1, F4)
---

# feat: Per-endpoint version header routing and auth inference from Authorization header params

## Overview

Two OpenAPI parser improvements that address the same root cause: the parser silently discards information from individual operations that it should preserve. For headers, it promotes one global value and drops per-endpoint variations. For auth, it never scans operation-level header parameters at all. Both cause runtime failures (404s, missing auth) on APIs like Cal.com that declare these at the operation level rather than globally.

## Problem Frame

**Per-endpoint headers (F1):** Cal.com uses `cal-api-version` with DIFFERENT values per endpoint group: bookings gets `2024-08-13`, event-types gets `2024-06-14`, schedules gets `2024-04-15`. The parser's `detectRequiredHeaders()` promotes the majority value globally, causing 404s on ~20% of endpoints. This affects any API with per-resource versioning (Cal.com, Twilio, enterprise APIs). ~10% of catalog.

**Auth from header params (F4):** Cal.com's spec has no `securitySchemes`, no auth-like query params, and a minimal `info.description` ("Cal.com v2 API") with no auth keywords. All three existing auth inference tiers fail. But the spec declares `Authorization` as a required header parameter on individual endpoints. The parser never scans operation-level header params for auth signals. ~20% of specs without formal security sections.

## Requirements Trace

- R1. When a required header has multiple distinct values across endpoint groups, each generated command sends the correct value for its endpoint
- R2. The global `RequiredHeaders` on `APISpec` continues to work for single-value headers (no regression)
- R3. When all three auth inference tiers fail, the parser scans operations for required `Authorization` header parameters and infers Bearer auth
- R4. Inferred auth from header params sets `Auth.Type`, `Auth.Header`, `Auth.EnvVars`, `Auth.In`, and `Auth.Inferred = true`
- R5. Explicit auth (`securitySchemes` present) always takes precedence — inference tiers only run when prior tiers return `Type: "none"`
- R6. The scorer recognizes auth inferred from header params (existing `Auth.Inferred` infrastructure)

## Scope Boundaries

- **In scope:** Per-endpoint header value overrides; auth inference from Authorization header params
- **Not in scope:** Inferring OAuth2 flows from header params; per-endpoint headers for non-required params; changes to the 30% frequency threshold
- **Not in scope:** Changes to `inferQueryParamAuth` or `inferDescriptionAuth` — these are stable

## Context & Research

### Relevant Code and Patterns

- `internal/openapi/parser.go:445` — `detectRequiredHeaders()`: scans operations, promotes headers at >80% frequency. Currently stores single `defaultValue` per header name
- `internal/openapi/parser.go:272` — `mapAuth()`: chains `selectSecurityScheme` → `inferQueryParamAuth` → `inferDescriptionAuth`. Fourth tier slots after description
- `internal/openapi/parser.go:383` — `inferQueryParamAuth()`: the exact pattern to follow for the new tier. Same iteration, same frequency threshold, same `AuthConfig` construction
- `internal/openapi/parser.go:553` — `inferDescriptionAuth()`: the most recent tier added. Returns `Inferred: true`
- `internal/openapi/parser.go:1222` — `mapParameters()`: filters out header params entirely (`In != path && In != query → continue`). Auth tier 4 must scan independently, not rely on mapParameters
- `internal/spec/spec.go:29` — `RequiredHeader{Name, Value}`: needs extension for per-endpoint overrides
- `internal/spec/spec.go:65` — `Endpoint` struct: has `Meta map[string]string` for ad-hoc metadata. Per-endpoint header overrides are better as a typed field
- `internal/generator/templates/client.go.tmpl:346` — global header emission in `do()` method
- `internal/generator/templates/command_endpoint.go.tmpl` — per-endpoint command; currently cannot override headers

### Institutional Learnings

- Steam retro #5: `inferQueryParamAuth` precedent validates the "fallback tier" pattern and 30% threshold
- Steam run 4 retro: "don't weaken existing thresholds for edge cases" — add new tiers instead
- Steam run 7 retro: `Auth.EnvVars` MUST be populated for verify to pass env vars to CLI subprocess
- Required header detection plan (004-002): explicitly scoped per-endpoint routing as future work
- Auth inference plan (004-003): explicitly scoped operation-level param scanning as future work

## Key Technical Decisions

- **Per-endpoint overrides on Endpoint struct, not on RequiredHeader:** The retro WU-1 suggested not changing the spec data model, but the current `RequiredHeader{Name, Value}` struct cannot express per-endpoint variations. Adding `HeaderOverrides []RequiredHeader` to `spec.Endpoint` is the cleanest path — it's typed, per-endpoint, and doesn't break the global `APISpec.RequiredHeaders` contract.

- **Client passes overrides via params map extension, not method signature change:** Rather than changing `do()` to accept a headers parameter (which touches every caller), the command template sets per-endpoint headers directly on the `http.Request` after calling `http.NewRequest` but before `client.Do`. This requires the command template to have access to the raw request, which means either: (a) add an optional `headers map[string]string` parameter to the client methods, or (b) use a `WithHeaders` option pattern. Option (a) is simpler given the existing codebase patterns.

- **Fourth-tier auth uses same iteration as `inferQueryParamAuth`:** Walk `doc.Paths`, call `mergeParameters()`, filter for `In == "header"` and name matches `authorization` (case-insensitive). Count occurrences, apply >30% threshold. This is consistent with the established pattern and avoids special-casing.

- **Parser detects per-endpoint values during `detectRequiredHeaders`, stores them on Endpoint during `mapResources`:** The detection and storage happen in two phases. First, `detectRequiredHeaders` is extended to also return a per-path value map when multiple values exist for the same header. Then, during `mapResources` (which builds endpoints), each endpoint gets its header overrides from this map. This keeps the parser's two-phase structure (headers detected globally, then applied per-endpoint).

## Open Questions

### Resolved During Planning

- **Q: Should per-endpoint header overrides be on the Endpoint struct or as a separate map on APISpec?** On Endpoint — it's per-endpoint data, so it belongs on the Endpoint struct. A top-level map keyed by path would duplicate what the Endpoint already represents.

- **Q: How does the command template pass per-endpoint headers to the client?** Add an optional `headers map[string]string` parameter to `Get`, `Post`, `Put`, `Patch`, `Delete` methods on the client. When non-nil, these are set on the request alongside the global RequiredHeaders. The global headers serve as defaults; per-endpoint headers override them for matching names.

- **Q: What if the Authorization header param has a description hinting at the format?** Scan the description for "Bearer" or "Basic" keywords (case-insensitive) to refine the inferred auth type. Default to `bearer_token` if no format hint is present (Bearer is the most common API auth scheme).

### Deferred to Implementation

- **Q: Exact header override storage format?** Whether `[]RequiredHeader` or `map[string]string` on Endpoint — implementation will determine which is cleaner with the template syntax.

## High-Level Technical Design

> *This illustrates the intended approach and is directional guidance for review, not implementation specification.*

### Per-Endpoint Header Routing

```
detectRequiredHeaders(doc, auth)
  │
  ├── Existing: count per-header frequency across all operations
  ├── NEW: also track per-path values when values differ
  │         headerValues[headerName][path] = value
  │
  ├── For headers above 80% threshold:
  │   ├── If all values identical → single RequiredHeader (existing behavior)
  │   └── If multiple values → RequiredHeader with majority value as global
  │         + return perEndpointHeaders map
  │
  └── Return ([]RequiredHeader, map[string]map[string]string)
                                ↑ headerName → path → value

mapResources(doc, out, basePath)
  │
  ├── Calls detectRequiredHeaders (now returns per-endpoint map)
  ├── For each endpoint being built:
  │   └── Check if this endpoint's path has a header override
  │       → Set endpoint.HeaderOverrides = [{Name, Value}]
  │
  └── Endpoints with no override use the global RequiredHeaders automatically
```

### Auth Inference Fourth Tier

```
mapAuth(doc, name)
  │
  ├── Tier 1: selectSecurityScheme(doc) → if found, return
  ├── Tier 2: inferQueryParamAuth(doc, name, fallback) → if not "none", return
  ├── Tier 3: inferDescriptionAuth(doc, name, fallback) → if not "none", return
  └── Tier 4: inferAuthHeaderParam(doc, name, fallback)    ← NEW
        │
        ├── Walk doc.Paths, mergeParameters(pathItem, op)
        ├── Filter: In == "header", name matches "authorization" (case-insensitive)
        ├── Count operations with Authorization header param
        │
        ├── If count / totalOps > 0.30:
        │   ├── Check param description for "Bearer" / "Basic" hints
        │   ├── Default: Type="bearer_token", Header="Authorization"
        │   ├── EnvVars=[PREFIX_TOKEN], In="header", Inferred=true
        │   └── Return AuthConfig
        │
        └── Else: return fallback (Type: "none")
```

## Implementation Units

- [ ] **Unit 1: Extend Endpoint struct with HeaderOverrides**

**Goal:** Add typed per-endpoint header override support to the spec data model

**Requirements:** R1, R2

**Dependencies:** None

**Files:**
- Modify: `internal/spec/spec.go`
- Test: `internal/spec/spec_test.go`

**Approach:**
- Add `HeaderOverrides []RequiredHeader` field to `Endpoint` with `yaml:"header_overrides,omitempty" json:"header_overrides,omitempty"` tags
- Existing global `RequiredHeaders` on `APISpec` is unchanged — per-endpoint overrides only exist on `Endpoint`

**Patterns to follow:**
- `RequiredHeaders` field pattern on `APISpec` (line 23)
- `Inferred bool` field addition pattern from auth inference plan

**Test scenarios:**
- Happy path: Endpoint with HeaderOverrides round-trips through YAML/JSON
- Happy path: Endpoint with empty HeaderOverrides omits the field (omitempty)

**Verification:**
- `go test ./internal/spec/...` passes

---

- [ ] **Unit 2: Extend detectRequiredHeaders to track per-endpoint values**

**Goal:** When a required header has multiple distinct values across endpoints, return both the global default and a per-path value map

**Requirements:** R1, R2

**Dependencies:** Unit 1

**Files:**
- Modify: `internal/openapi/parser.go`
- Test: `internal/openapi/parser_test.go`
- Create: `testdata/openapi/multi-version-header.yaml`

**Approach:**
- Change `detectRequiredHeaders` return type from `[]RequiredHeader` to `([]RequiredHeader, map[string]map[string]string)` — the second value maps `headerName → apiPath → value`
- During iteration, when a header name is already known with a different value, record it in the per-path map instead of ignoring it
- The global `RequiredHeader.Value` uses the majority value (most frequent across operations)
- When all values are identical, the per-path map is empty (existing behavior preserved)
- Update the call site in `parse()` to capture both returns

**Patterns to follow:**
- `inferQueryParamAuth` iteration pattern: `doc.Paths.InMatchingOrder()`, pathItem.Operations(), `mergeParameters()`
- Existing `headerInfo` struct (line 464) — extend with `values map[string]int` to count per-value frequency

**Test scenarios:**
- Happy path: Spec with uniform header value → single RequiredHeader, empty per-path map
- Happy path: Spec with 3 endpoint groups using different values → RequiredHeader with majority value, per-path map with deviations
- Edge case: Header with 2 values at 50/50 split → either is acceptable as global default; both are in per-path map
- Edge case: Header with value on some endpoints and no value on others → global uses majority; endpoints without value not in per-path map
- Integration: Existing versioned-api.yaml fixture → unchanged behavior (all same value)
- Integration: Existing petstore.yaml → still no required headers

**Verification:**
- `go test ./internal/openapi/...` passes
- New fixture exercises the per-endpoint path

---

- [ ] **Unit 3: Populate Endpoint.HeaderOverrides during mapResources**

**Goal:** Wire per-path header values into each endpoint's HeaderOverrides field during resource mapping

**Requirements:** R1

**Dependencies:** Unit 2

**Files:**
- Modify: `internal/openapi/parser.go` (mapResources and parse functions)
- Test: `internal/openapi/parser_test.go`

**Approach:**
- In `parse()`, capture the per-path map from `detectRequiredHeaders` and pass it to `mapResources`
- In `mapResources`, after building each endpoint, check if the endpoint's API path has an override in the per-path map. If the override differs from the global `RequiredHeaders` value, set `endpoint.HeaderOverrides`
- Only set overrides when the value DIFFERS from global — don't redundantly store the global value

**Patterns to follow:**
- The existing flow where `result.RequiredHeaders` is set in `parse()` before `mapResources` is called

**Test scenarios:**
- Happy path: Parse multi-version-header.yaml → bookings endpoints have no overrides (they match global), event-types endpoints have HeaderOverrides with the different value
- Edge case: Endpoint with no matching override → HeaderOverrides is nil/empty
- Integration: Parse petstore.yaml → no endpoints have HeaderOverrides

**Verification:**
- `go test ./internal/openapi/...` passes
- Full parse of multi-version fixture produces correct per-endpoint overrides

---

- [ ] **Unit 4: Extend client methods to accept per-request header overrides**

**Goal:** Client's HTTP methods accept optional headers that override the global RequiredHeaders for that request

**Requirements:** R1

**Dependencies:** Unit 1

**Files:**
- Modify: `internal/generator/templates/client.go.tmpl`
- Test: `internal/generator/generator_test.go`

**Approach:**
- Add a `headers map[string]string` parameter to the `do()` method. When non-nil, set these on the request AFTER global RequiredHeaders, overriding any matching names
- Update `Get`, `Post`, `Put`, `Patch`, `Delete` method signatures to accept and pass through optional headers. Use a nil-safe pattern so existing callers (which don't need overrides) pass nil
- The global RequiredHeaders in client.go.tmpl remain as-is — they set defaults. Per-endpoint overrides take priority

**Patterns to follow:**
- The existing `params map[string]string` parameter pattern on client methods

**Test scenarios:**
- Happy path: Generated client with per-endpoint headers compiles and runs
- Happy path: Override header value supersedes global RequiredHeader for that request
- Edge case: nil headers parameter → global RequiredHeaders applied (no change from current behavior)

**Verification:**
- Generated CLI compiles (`go build ./...`)

---

- [ ] **Unit 5: Command templates emit per-endpoint header overrides**

**Goal:** Command endpoint and promoted templates pass HeaderOverrides to client methods

**Requirements:** R1

**Dependencies:** Units 3, 4

**Files:**
- Modify: `internal/generator/templates/command_endpoint.go.tmpl`
- Modify: `internal/generator/templates/command_promoted.go.tmpl`
- Test: `internal/generator/generator_test.go`

**Approach:**
- In the command template, check if `.Endpoint.HeaderOverrides` is non-empty
- If so, build a `map[string]string` from the overrides and pass it to the client method call
- If empty, pass nil (existing behavior)
- Template conditional: `{{- if .Endpoint.HeaderOverrides}} headers := map[string]string{...} {{- end}}`

**Patterns to follow:**
- Existing `{{- if .HasStore}}` conditional blocks in command templates

**Test scenarios:**
- Happy path: Generate from multi-version-header fixture → event-types command passes override headers
- Happy path: Generate from petstore.yaml → commands pass nil headers (no overrides)
- Integration: Full generation + build from multi-version fixture succeeds

**Verification:**
- Generated CLI compiles and the correct header value appears in dry-run output

---

- [ ] **Unit 6: Implement inferAuthHeaderParam as fourth auth tier**

**Goal:** Detect Bearer auth from required Authorization header parameters when all three prior tiers fail

**Requirements:** R3, R4, R5, R6

**Dependencies:** None (independent of Units 1-5)

**Files:**
- Modify: `internal/openapi/parser.go`
- Test: `internal/openapi/parser_test.go`
- Create: `testdata/openapi/auth-header-param.yaml`

**Approach:**
- New function `inferAuthHeaderParam(doc *openapi3.T, name string, fallback spec.AuthConfig) spec.AuthConfig`
- Guard for `doc == nil` or `doc.Paths == nil` → return fallback
- Walk all operations via `doc.Paths.InMatchingOrder()`, use `mergeParameters()` to get all params
- Count operations where a parameter has `In == "header"` and `Name` matches `authorization` (case-insensitive) and `Required == true`
- If count/totalOps > 0.30: infer Bearer auth
  - Check param description for "Bearer" → `Type: "bearer_token"`, "Basic" → `Type: "api_key"` with Basic format
  - Default to `Type: "bearer_token"` if no hint
  - Set `Header: "Authorization"`, `In: "header"`, `EnvVars: [PREFIX_TOKEN]`, `Inferred: true`
- Wire into `mapAuth`: after `inferDescriptionAuth` returns "none", call `inferAuthHeaderParam`

**Patterns to follow:**
- `inferQueryParamAuth` for function signature, iteration, threshold, and AuthConfig construction
- `inferDescriptionAuth` for the `Inferred: true` pattern and negation guards
- `commonAuthQueryParams` map for keyword matching style

**Test scenarios:**
- Happy path: Spec with required Authorization header param on >30% of ops → Type="bearer_token", Header="Authorization", Inferred=true, EnvVars populated
- Happy path: Spec with Authorization param description containing "Bearer" → Type="bearer_token"
- Happy path: Spec with Authorization param description containing "Basic" → Type="api_key" with Basic format
- Edge case: Spec WITH securitySchemes AND Authorization header params → explicit auth wins, tier 4 never runs
- Edge case: Spec with Authorization header params on <30% of ops → fallback returned
- Edge case: Spec with no header params at all → fallback returned
- Edge case: Optional (not required) Authorization header param → not counted
- Integration: Existing petstore.yaml → auth unchanged (has securitySchemes)
- Integration: Existing versioned-api.yaml → auth stays "none" (no Authorization params)

**Verification:**
- `go test ./internal/openapi/...` passes
- New fixture exercises the fourth tier

## System-Wide Impact

- **Interaction graph (headers):** `detectRequiredHeaders` → `RequiredHeader` → `Endpoint.HeaderOverrides` → command template → client `do()` → HTTP request. Global RequiredHeaders in client.go.tmpl are defaults; per-endpoint overrides supersede.
- **Interaction graph (auth):** `mapAuth` → `inferAuthHeaderParam` → `AuthConfig` → config template + client template + doctor template + auth template + README template. All downstream templates use the same `Auth.*` fields — no new plumbing needed.
- **Unchanged invariants:** `selectSecurityScheme`, `inferQueryParamAuth`, `inferDescriptionAuth` are not modified. The 30% query-param threshold is unchanged. Specs with explicit security sections produce identical output.
- **Verify integration:** `Auth.EnvVars` must be populated for verify to pass correct env vars (Steam run 7 learning). The new tier must set this field.
- **Scorer interaction:** The `auth_protocol` scorecard dimension already handles `Auth.Inferred` (from the description inference work). No scorer changes needed.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Client method signature change breaks generated CLI compilation | All callers updated by template; nil parameter preserves current behavior |
| Per-endpoint header map grows large for specs with many paths | Only stored when values DIFFER from global — most endpoints use global default |
| False positive auth inference from non-auth Authorization params | >30% threshold + required=true filter minimizes risk; Inferred flag tells user to verify |
| Breaking existing CLIs | Per-endpoint overrides are additive; existing global headers unchanged; auth tiers only fire when prior tiers return "none" |

## Sources & References

- **Origin document:** [docs/retros/2026-04-05-cal-com-run2-retro.md](docs/retros/2026-04-05-cal-com-run2-retro.md) — findings F1, F4
- **Prior art (headers):** [docs/plans/2026-04-04-002-feat-required-api-header-detection-plan.md](docs/plans/2026-04-04-002-feat-required-api-header-detection-plan.md) — completed plan that implemented the initial RequiredHeader feature
- **Prior art (auth):** [docs/plans/2026-04-04-003-feat-auth-inference-from-description-plan.md](docs/plans/2026-04-04-003-feat-auth-inference-from-description-plan.md) — completed plan for tiers 1-3
- Related retros: Cal.com retro finding #5 (per-endpoint versioning), Steam retro #5 (query-param auth precedent), Steam run 7 retro (EnvVars must be populated)
- Related code: `internal/openapi/parser.go` — `detectRequiredHeaders` (line 445), `mapAuth` (line 272), `inferQueryParamAuth` (line 383), `inferDescriptionAuth` (line 553)
