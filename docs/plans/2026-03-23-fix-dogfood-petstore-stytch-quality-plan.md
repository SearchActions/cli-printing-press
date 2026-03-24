---
title: "Dogfood Petstore and Stytch CLIs to Production Quality"
type: fix
status: active
date: 2026-03-23
---

# Dogfood Petstore and Stytch CLIs to Production Quality

## Overview

The printing press generates CLIs that compile but aren't usable. Dogfooding revealed 6 bugs in the generator that affect every generated CLI. Fix the press, not the output. Petstore and Stytch are the test subjects - get them both to production quality before attempting Discord or other complex specs.

## Dogfooding Findings

### Petstore bugs found

1. **Base URL broken** - Petstore spec has `servers[0].url: /api/v3` (relative path). Parser stores this as `base_url: /api/v3` with no scheme. Every API call fails: `unsupported protocol scheme ""`. Doctor confirms: `api=unreachable`.
2. **Binary name absurd** - `swagger-petstore-openapi-30-cli` derived from spec title "Swagger Petstore - OpenAPI 3.0". Should generate something like `petstore-cli`.
3. **Command descriptions are CamelCase garbage** - `Connectedapps`, `Deletebiometricregistration`, `Exchangeprimaryfactor`. The `Short` field in Cobra commands uses the raw operationId summary with no space insertion.

### Stytch bugs found

4. **Resource descriptions empty** - Most resources show blank descriptions in `--help`. The tag descriptions from the OpenAPI spec aren't being mapped to the right resource names (tag name vs snake_case resource name mismatch).
5. **Descriptions truncated oddly** - `create_user_as_pending` flag description is cut off mid-sentence. The `oneline` function truncates at 120 chars without word boundary awareness.
6. **Complex types as string flags** - `trusted_metadata` and `untrusted_metadata` are JSON objects but rendered as `--trusted_metadata string` flags. Users can't pass JSON objects as string flags usably.

### Both CLIs

7. **No --help examples** - Every Steinberger CLI shows usage examples in help text. Ours show none.

## The Dogfooding Loop

```
while (quality bar not met):
  1. Fix the printing press (parser.go, templates, generator.go)
  2. Run go test ./...
  3. Delete petstore-cli/ and stytch-cli/
  4. Regenerate both
  5. Score:
     - Can you actually call the API? (not just compile)
     - Are command names readable?
     - Are descriptions helpful?
     - Would Steinberger ship this?
  6. If any issue: identify root cause in generator, go to step 1
  7. If both are usable: done, move to Discord
```

## Fixes (In Dogfood Order)

### Round 1: Make Petstore API calls work

**Fix 1: Handle relative base URLs**

The OpenAPI parser takes `servers[0].url` as-is. When it's a relative path like `/api/v3`, the generated client can't make HTTP requests. The parser should detect relative URLs and either:
- Warn and skip (require absolute URL)
- Combine with the spec's host/origin if detectable
- For well-known specs (Petstore), the skill/catalog should provide the full URL

Root cause: `internal/openapi/parser.go` line ~44, `baseURL` assignment.

Files: `internal/openapi/parser.go`

**Fix 2: Smarter binary name derivation**

`swagger-petstore-openapi-30` is derived from kebab-casing the full title. Strip common noise words: "swagger", "openapi", version numbers, "api", "rest". For "Swagger Petstore - OpenAPI 3.0" the result should be `petstore`.

Root cause: `toKebabCase(doc.Info.Title)` in `parser.go` line ~35.

Files: `internal/openapi/parser.go`

### Round 2: Fix descriptions

**Fix 3: Human-readable command descriptions**

The `Short` field on subcommands uses operationId summaries that are often CamelCase or missing spaces. Apply `oneline` to clean them, but also insert spaces before capitals in CamelCase strings: `Deletebiometricregistration` -> `Delete biometric registration`.

Root cause: The descriptions come from `op.Summary` or `op.Description` via `firstNonEmpty`, but some specs have summaries like "Connectedapps" which is the operationId, not a real description.

Files: `internal/openapi/parser.go` (description extraction), `internal/generator/generator.go` (template functions)

**Fix 4: Truncate descriptions at word boundaries**

The `oneline` function in `generator.go` cuts at 120 chars without caring about word boundaries. Change to cut at the last space before 120 chars.

Files: `internal/generator/generator.go`

**Fix 5: Map tag descriptions to resources**

OpenAPI tags have descriptions that should map to resource-level descriptions. The current `mapTagDescriptions` converts tag names with `toSnakeCase` but resource names also go through `resourceNameFromPath` which may produce different results. Align the mapping.

Files: `internal/openapi/parser.go`

### Round 3: Usability polish

**Fix 6: Skip complex object params as flags**

When a request body property is an object type (not string/int/bool/float), skip it as a flag. JSON objects can't be passed as `--flag value` usably. Log a warning instead.

Files: `internal/openapi/parser.go`

**Fix 7: Add usage examples to help text (stretch)**

In the command template, generate a `Example` field on Cobra commands showing one realistic invocation. E.g., `stytch-api-cli users create --email user@example.com`.

Files: `internal/generator/templates/command.go.tmpl`

## Acceptance Criteria

Measured by regenerating from both specs after every change:

- [ ] `petstore doctor` shows `api=reachable` (base URL has scheme)
- [ ] `petstore pet find_by_status --status available` returns JSON from the real API
- [ ] Binary name is `petstore-cli` not `swagger-petstore-openapi-30-cli`
- [ ] Command descriptions are readable English, not CamelCase blobs
- [ ] Flag descriptions don't cut off mid-word
- [ ] Resource descriptions show tag descriptions from the spec
- [ ] Complex object params are skipped (not rendered as useless string flags)
- [ ] No regressions on Stytch or internal-format specs
- [ ] All existing Go tests pass

## Scope Boundaries

- Not tackling Discord yet (fix Petstore and Stytch first)
- Not adding usage examples (stretch goal, not blocking)
- Not fixing auth flow (requires API keys the user configures)
- Not adding color/TTY detection (separate plan)

## Sources

- Petstore dogfood: `unsupported protocol scheme ""` on every API call
- Stytch dogfood: empty resource descriptions, truncated flag text
- Generator code: `internal/openapi/parser.go`, `internal/generator/generator.go`
- Petstore spec: `testdata/openapi/petstore.yaml` (`servers[0].url: /api/v3`)
