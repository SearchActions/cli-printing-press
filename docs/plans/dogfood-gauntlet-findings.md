# Dogfood Gauntlet Findings - 2026-03-24

## Scorecard: 3/10 Pass All 7 Gates

| # | API | Paths | Result | Failure Point | Root Cause |
|---|-----|-------|--------|---------------|------------|
| 1 | Fly.io | 51 | FAIL | go vet | types.go: generated types have syntax errors (`.` in type names from complex $ref schemas) |
| 2 | Spotify | 68 | FAIL | binary crash | CLI name derived from spec title is absurdly long ("spotify-web-with-fixes-and-improvements-from-sonallux-cli"), causes panic in command registration |
| 3 | Telegram | 74 | PASS | - | 7/7 gates pass. 2 endpoints skipped (resource limit 50). Flat structure (no sub-resources). |
| 4 | Vercel | 85 | FAIL | go vet | `$schema` JSON reference keys leak into generated Go code as `$schema` variable names (invalid Go identifier) |
| 5 | Supabase | 105 | FAIL | parse | `base_url is required` - spec has no `servers` field, parser requires it |
| 6 | Sentry | 126 | PASS | - | 7/7 gates pass. Flat "api" resource. Doctor shows templated URL bug: `https://{region}.sentry.io` has `{region}` literal in host. |
| 7 | LaunchDarkly | 221 | PASS | - | 7/7 gates pass. Flat "api" resource. Doctor shows API reachable (200). Clean output. |
| 8 | Trello | 264 | FAIL | go build | Syntax errors from `/` characters in generated enum type names (Trello uses `modelTypes/card` style enums) |
| 9 | Jira | 317 | FAIL | go vet | types.go: malformed type definitions from deeply nested $ref schemas with numeric keys |
| 10 | Cloudflare | 1716 | FAIL | parse | `name is required` - spec missing required `info.title` field (Cloudflare uses `x-api-name` extension) |

## Grade: D (3/10 pass)

Below the 7/10 target. Generator needs targeted fixes for common patterns.

## Bugs Found

### Bug 1: `$ref` / `$schema` keys leak into Go identifiers
- **Affected:** Fly.io, Vercel, Jira
- **Root cause:** JSON Schema `$ref` and `$schema` keys are passed through to Go type/variable names without sanitizing the `$` prefix
- **Fix needed in:** `internal/openapi/parser.go` (type name sanitization) and `internal/generator/templates/types.go.tmpl`
- **Severity:** High - affects any spec with complex schemas

### Bug 2: Spec title used as CLI name without truncation
- **Affected:** Spotify ("Spotify Web API with fixes and improvements from sonallux")
- **Root cause:** `--name` defaults to kebab-cased spec title, no length limit
- **Fix needed in:** `internal/cli/root.go` (name derivation) or `internal/openapi/parser.go`
- **Severity:** Medium - workaround is `--name spotify`

### Bug 3: Missing `servers` field causes parse failure
- **Affected:** Supabase
- **Root cause:** Parser requires `base_url` but some specs have no `servers` array
- **Fix needed in:** `internal/openapi/parser.go` (fallback to empty string or host from spec URL)
- **Severity:** Medium - affects specs that rely on relative URLs

### Bug 4: Cloudflare spec uses `x-api-name` instead of `info.title`
- **Affected:** Cloudflare
- **Root cause:** Parser requires `info.title` but Cloudflare uses a custom extension
- **Fix needed in:** `internal/openapi/parser.go` or `internal/spec/spec.go` (fallback to x-api-name)
- **Severity:** Low - Cloudflare-specific

### Bug 5: Enum values with `/` generate invalid Go type names
- **Affected:** Trello
- **Root cause:** Trello uses `modelTypes/card` style enum values that become Go identifiers
- **Fix needed in:** `internal/openapi/parser.go` (enum value sanitization)
- **Severity:** Medium - affects specs with path-like enum values

### Bug 6: Templated base URL not resolved
- **Affected:** Sentry (`https://{region}.sentry.io`)
- **Root cause:** Server URL templates with `{variable}` placeholders are used verbatim instead of resolved
- **Fix needed in:** `internal/openapi/parser.go` (resolve server URL templates using default values)
- **Severity:** Low - CLI still compiles, doctor just shows unreachable

## Dogfood Observations

### Telegram (PASS)
- 50+ commands at top level (flat, no sub-resources)
- All commands are verb-noun style: `send-message`, `answer-callback-query`
- Doctor: base_url has `{token}` placeholder (same templated URL issue as Sentry)
- Descriptions say "Manage X" - generic, not helpful

### Sentry (PASS)
- Single "api" top-level resource (flat hierarchy)
- All 50 endpoints under `api` - the sub-resource detection doesn't help because all paths start with `/api/0/`
- Doctor: templated URL `{region}` literal in host
- CLI name: "reference-cli" (from spec title "Sentry Public API" being overridden by... something)

### LaunchDarkly (PASS)
- Single "api" top-level resource (same flat problem)
- Doctor: API reachable, 200 response
- Clean help output, good descriptions
- Truncated spec description in `--help` banner (too long)

## Priority Fix Order

1. **$ref/$schema sanitization** (fixes 3 APIs: Fly.io, Vercel, Jira)
2. **Spec title truncation for CLI name** (fixes Spotify)
3. **Missing servers fallback** (fixes Supabase)
4. **Enum value sanitization** (fixes Trello)
5. **Server URL template resolution** (fixes Sentry/Telegram doctor)
6. **x-api-name fallback** (fixes Cloudflare)

Fixes 1-4 would bring the pass rate from 3/10 to 8/10.

## Recommendations

- The "flat api resource" problem (Sentry, LaunchDarkly) is a parser issue: all paths start with `/api/v2/` or `/api/0/` and the first segment becomes the only resource. The parser should strip common API path prefixes before grouping.
- Consider bumping the endpoint-per-resource limit from 50 to 100 for APIs with flat structures.
- The spec title to CLI name derivation needs a hard cap (e.g., 30 chars) and a smart truncation that picks the most meaningful words.
