---
title: "CLI Printing Press Phase 2: OpenAPI Parser"
type: feat
status: active
date: 2026-03-23
origin: docs/plans/2026-03-23-feat-cli-printing-press-plan.md
---

# CLI Printing Press Phase 2: OpenAPI Parser

## Goal

Add `printing-press generate --spec openapi.yaml` that accepts OpenAPI 3.0+ specs and maps them to our internal YAML format, then runs the existing generator. The user should be able to take any published OpenAPI spec (Stytch, Clerk, ClickUp, Discord) and get a production CLI.

By the end of Phase 2: `printing-press generate --spec https://raw.githubusercontent.com/stytchauth/stytch-openapi/main/openapi.yaml` produces a working CLI directly from the official spec.

## Learnings from Phase 1 (Applied Here)

Building Phase 1 revealed several things that directly shape Phase 2:

### 1. Template bugs surface late

The config template had a field/method name collision (`AuthHeader` as both a struct field and method) that only showed up when the generated code compiled. The command template had a `replacePathParam` function duplicated across per-resource files. These were invisible until `go build` ran on the output.

**Applied to Phase 2**: The OpenAPI-to-internal-spec mapping must be validated BEFORE passing to the generator. Don't just map and hope - parse the internal spec back through `spec.Validate()` to catch issues early.

### 2. The internal YAML format is good but has gaps

Writing the Stytch/Clerk/Loops specs by hand revealed:
- No way to express **nested body objects** cleanly (Param.Fields exists but the generator doesn't handle it well)
- No way to express **response pagination patterns** that differ from query pagination
- No way to express **enum values** for parameters (the `enum` type in Cobra flags was hand-coded in the template)
- No way to express **header-based auth** vs **query-based auth** (Stytch uses Basic auth in headers, some APIs use API keys in query params)
- The `auth.format` field is a string template (`"Basic {project_id}:{secret}"`) which is fragile

**Applied to Phase 2**: The OpenAPI parser should handle these natively and produce an internal spec that includes: nested body schemas, response envelope paths, enum constraints, and auth scheme details. Extend the internal spec format where needed.

### 3. Template functions are the critical integration point

The generator has 12 custom template functions (`goType`, `cobraFlagFunc`, `defaultVal`, `envVarField`, etc.). Every spec field that needs to become Go code flows through these functions. When the OpenAPI parser adds new types (e.g., `array of strings`, `object with nested fields`, `number with enum constraints`), these functions must handle them.

**Applied to Phase 2**: Add comprehensive type mapping tests. Every OpenAPI type combination (string, integer, boolean, array, object, string+enum, string+format:date-time, etc.) must have a test showing what Go type and Cobra flag type it maps to.

### 4. Quality gates catch real problems

The validate.go gates (go mod tidy, go vet, go build, --help, version, doctor) caught issues that would have been silent in Phase 1. The integration test running all 3 specs took 25 seconds - fast enough to run on every change.

**Applied to Phase 2**: Every OpenAPI spec we test against must pass all quality gates. If a generated CLI doesn't compile, the spec mapping is wrong. Add the 5 real-world specs (Stytch, Clerk, ClickUp, Discord, Asana) as test fixtures.

### 5. Codex delegation worked well for mechanical tasks

Codex built validate.go, the test specs, and the integration tests in one shot. The prompt worked best when it was: "here's what exists, here's exactly what to build, here are the rules." The structured prompt format (context -> task -> rules) produced clean results.

**Applied to Phase 2**: The OpenAPI mapping logic is mechanical (schema -> type, path -> endpoint, security -> auth). Good candidate for Codex delegation. The edge-case handling (weird OpenAPI patterns) is better for Claude.

## What Gets Built

### New file: `internal/openapi/parser.go`

Parses OpenAPI 3.0+ YAML/JSON specs into our `spec.APISpec` internal format.

### New file: `internal/openapi/parser_test.go`

Tests with real-world OpenAPI specs from GitHub.

### Modified: `internal/cli/root.go`

The `generate` command auto-detects whether `--spec` points to an OpenAPI spec or our internal format. Detection: check for `openapi:` key at the top level of the YAML.

### Modified: `internal/spec/spec.go`

Extend the internal format with:
- `Param.Enum []string` - enum constraints
- `Param.Format string` - OpenAPI format hints (date-time, email, uri, etc.)
- `Endpoint.ResponsePath string` - path to extract the data array from response (e.g., `"data"`, `"results.items"`)
- `AuthConfig.Scheme string` - OpenAPI security scheme name
- `AuthConfig.In string` - header, query, cookie

### New: `testdata/openapi/` directory

Real OpenAPI specs downloaded from:
- `stytch-openapi.yaml` (github.com/stytchauth/stytch-openapi)
- `clerk-openapi.yaml` (clerk.com/docs/reference)
- `clickup-openapi.json` (developer.clickup.com)

## Implementation Steps

### Step 1: Extend internal spec format

Add the new fields to `spec.go`. Update validation. Update existing tests.

Files: `internal/spec/spec.go`, `internal/spec/spec_test.go`

### Step 2: Build OpenAPI parser

Use `github.com/getkin/kin-openapi/openapi3` (the most popular Go OpenAPI 3.0 parser, 2.1K stars).

Map OpenAPI concepts to internal spec:
- `info.title` -> `name` (kebab-cased)
- `info.description` -> `description`
- `info.version` -> `version`
- `servers[0].url` -> `base_url`
- `security` + `components.securitySchemes` -> `auth`
- `paths` -> `resources` (group by first path segment, e.g., `/users/{id}` -> resource `users`)
- `paths.{path}.{method}` -> `endpoints` (name from operationId or path+method)
- `parameters` -> `params` (path params are positional, query params are flags)
- `requestBody` -> `body`
- `responses.200.content.application/json.schema` -> `response`
- `components.schemas` -> `types`

Files: `internal/openapi/parser.go`

### Step 3: Auto-detection in CLI

When `--spec` is provided, peek at the first few bytes. If it contains `openapi:` or `"openapi"`, parse as OpenAPI. Otherwise, parse as internal format.

Files: `internal/cli/root.go`

### Step 4: Download real OpenAPI specs for testing

```bash
# Stytch
curl -o testdata/openapi/stytch.yaml https://raw.githubusercontent.com/stytchauth/stytch-openapi/main/openapi/stytch_api.yaml

# Clerk
curl -o testdata/openapi/clerk.json https://clerk.com/docs/reference/backend-api/openapi.json

# Discord (large, for stress testing)
curl -o testdata/openapi/discord.json https://raw.githubusercontent.com/discord/discord-api-spec/main/specs/openapi.json
```

Files: `testdata/openapi/`

### Step 5: Integration tests with real specs

Generate CLIs from all 3 real OpenAPI specs. Run quality gates. Verify compilation.

Files: `internal/openapi/parser_test.go`

### Step 6: Handle OpenAPI edge cases

Real OpenAPI specs have:
- `$ref` references (must resolve)
- `allOf`/`oneOf`/`anyOf` (common in request bodies)
- Deeply nested schemas
- Parameters defined at path level (inherited by all operations)
- Multiple security schemes (pick the simplest)
- Response schemas with wrappers (data is nested inside `{"data": [...]}`)
- Pagination via headers (Link header) or response fields

The kin-openapi library handles `$ref` resolution. For `allOf`, merge properties. For `oneOf`/`anyOf`, use `any` type. For nested schemas, flatten to max 2 levels.

## Acceptance Criteria

- [ ] `printing-press generate --spec stytch-openapi.yaml` produces a compilable CLI
- [ ] `printing-press generate --spec clerk-openapi.json` produces a compilable CLI
- [ ] Auto-detection: OpenAPI specs and internal specs both work with `--spec`
- [ ] `$ref` resolution works (schemas reference other schemas)
- [ ] `allOf` merging works (common in request bodies)
- [ ] Path parameters become positional args
- [ ] Query parameters become flags
- [ ] Request body fields become flags
- [ ] Security schemes map to auth config
- [ ] Response schemas map to types
- [ ] Enum constraints are preserved
- [ ] All quality gates pass on generated CLIs
- [ ] Existing internal-format specs still work (no regressions)

## What Phase 2 Does NOT Include

- URL-based spec fetching (`--spec https://...`) - nice-to-have, not required
- OpenAPI 2.0 (Swagger) support - only OpenAPI 3.0+
- OAuth2 flow generation (just maps to bearer_token for now)
- Webhook handling
- Server-sent events / streaming endpoints

## Technical Decisions

- **OpenAPI library**: `github.com/getkin/kin-openapi/openapi3` - handles parsing, validation, and `$ref` resolution. Used by Kubernetes, CoreDNS, and many production tools.
- **Resource grouping**: Group endpoints by first path segment. `/users/{id}` and `/users` both go under `users` resource. `/users/{id}/sessions` goes under `users` with nested path.
- **Operation naming**: Use `operationId` if present, otherwise `method + last path segment` (e.g., `GET /users/{id}` -> `get`, `POST /users` -> `create`, `DELETE /users/{id}` -> `delete`).
- **Type mapping**: OpenAPI `string` -> Go `string`, `integer` -> `int`, `boolean` -> `bool`, `number` -> `float64`, `array` -> use item type, `object` -> generate struct in types.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Real OpenAPI specs are messy/incomplete | Generated CLI is missing endpoints | Log warnings for skipped endpoints. Generate what we can, skip what we can't. |
| kin-openapi library has quirks | Parser bugs | Pin version, write integration tests against real specs |
| Type mapping gets complex (nested objects, arrays of objects) | Generated Go code won't compile | Start with flat types, add nesting incrementally. Flatten to `json.RawMessage` for anything too complex. |
| allOf/oneOf/anyOf handling | Complex schemas produce bad Go types | allOf: merge properties. oneOf/anyOf: use `any` type. Good enough for v1. |

## Origin

Phase 2 of: docs/plans/2026-03-23-feat-cli-printing-press-plan.md
Builds on: docs/plans/2026-03-23-feat-cli-printing-press-phase1-template-engine-plan.md (completed)
