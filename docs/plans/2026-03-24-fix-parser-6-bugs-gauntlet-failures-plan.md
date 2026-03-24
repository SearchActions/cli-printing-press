---
title: "Fix 6 Parser Bugs from Dogfood Gauntlet - 3/10 to 8/10"
type: fix
status: completed
date: 2026-03-24
origin: docs/plans/dogfood-gauntlet-findings.md
---

# Fix 6 Parser Bugs from Dogfood Gauntlet

## Overview

The dogfood gauntlet ran 10 new API specs through the generator. 3 passed, 7 failed. Six distinct bugs in the parser and type generator cause all 7 failures. Fixing them should bring the pass rate from 3/10 to at least 8/10.

## The Bugs (Priority Order)

| Bug | Affects | Root Cause | Fix Location |
|-----|---------|------------|--------------|
| 1. `$ref`/`$schema` in type names | Fly.io, Vercel, Jira | Schema names with `$` prefix pass through to Go identifiers | `parser.go:mapTypes`, `parser.go:874` |
| 2. Spec title too long for CLI name | Spotify | `cleanSpecName` has no length cap | `parser.go:cleanSpecName` (line 1526) |
| 3. Missing `servers` field | Supabase | Parser requires base_url but spec has no servers array | `parser.go:42-56`, `spec.go:Validate` |
| 4. `/` in enum values | Trello | Enum values like `modelTypes/card` become Go identifiers | `parser.go:mapTypes` field names, `generator/templates/types.go.tmpl` |
| 5. Server URL templates unresolved | Sentry, Telegram | `{region}` in server URL used verbatim | `parser.go:44-46` |
| 6. Missing `info.title` | Cloudflare | Spec uses `x-api-name` extension, no `info.title` | `parser.go:34-40` |

## Acceptance Criteria

- [ ] Fly.io (51 paths) passes all 7 gates
- [ ] Spotify (68 paths) passes all 7 gates with auto-derived name
- [ ] Vercel (85 paths) passes all 7 gates
- [ ] Supabase (105 paths) passes all 7 gates
- [ ] Trello (264 paths) passes all 7 gates
- [ ] Jira (317 paths) passes all 7 gates
- [ ] Sentry doctor shows resolved URL (not `{region}` literal)
- [ ] Telegram doctor shows resolved URL (not `{token}` literal)
- [ ] Cloudflare (1716 paths) at minimum parses without crash
- [ ] `go test ./...` passes after every fix
- [ ] Existing specs (petstore, stytch, discord, gmail) still pass all 7 gates

## Implementation Units

### Unit 1: Sanitize `$` from Schema Names and Field Names

**Files:** `internal/openapi/parser.go`

**Root cause:** `mapTypes` at line 874 uses the raw schema name (`name` from `doc.Components.Schemas`) as the Go type name. Schema names like `$ref`, `$schema`, `components.schemas.something` contain characters that are invalid in Go identifiers.

Similarly, `collectTypeProperties` at line 920 uses raw field names that may contain `$`.

**Fix:**
1. In `mapTypes` (line 874), sanitize the schema name before using it:
```go
for _, name := range names {
    goName := sanitizeTypeName(name)
    if goName == "" {
        continue
    }
    // ... rest uses goName instead of name for the type key
```

2. In `collectTypeProperties` (line 920), sanitize field names:
```go
for name, prop := range schema.Properties {
    sanitized := sanitizeFieldName(name)
    if sanitized == "" || strings.HasPrefix(sanitized, "_") {
        continue
    }
    properties[sanitized] = prop
}
```

3. Add `sanitizeTypeName` function:
```go
func sanitizeTypeName(name string) string {
    // Remove $ prefix (JSON Schema $ref, $schema, $id)
    name = strings.TrimLeft(name, "$")
    // Remove dots, slashes, backslashes
    name = strings.NewReplacer(
        ".", "_",
        "/", "_",
        "\\", "_",
        "-", "_",
        " ", "_",
    ).Replace(name)
    // Remove any remaining non-alphanumeric/underscore chars
    var b strings.Builder
    for _, r := range name {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
            b.WriteRune(r)
        }
    }
    result := b.String()
    // Must start with letter
    if len(result) > 0 && !unicode.IsLetter(rune(result[0])) {
        result = "T" + result
    }
    return result
}
```

4. Add `sanitizeFieldName` (same logic but lowercase-first):
```go
func sanitizeFieldName(name string) string {
    name = strings.TrimLeft(name, "$")
    name = strings.NewReplacer(".", "_", "/", "_", "\\", "_").Replace(name)
    var b strings.Builder
    for _, r := range name {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
            b.WriteRune(r)
        }
    }
    return b.String()
}
```

**Verification:** Fly.io, Vercel, and Jira pass `go vet` gate.

**Execution note:** Fix-first. This unlocks 3 APIs at once. Run all 3 after the fix to verify.

### Unit 2: Cap Spec Title Length for CLI Name

**Files:** `internal/openapi/parser.go`

**Root cause:** `cleanSpecName` (line 1526) filters noise words but has no length cap. Spotify's title "Spotify Web API with fixes and improvements from sonallux" produces a 60+ char CLI name that panics during command registration.

**Fix:** After the filtered tokens are joined (line 1575), cap at 3 tokens max:

```go
// Cap at 3 meaningful tokens for CLI name
if len(filtered) > 3 {
    filtered = filtered[:3]
}
name := toKebabCase(strings.Join(filtered, " "))
```

Also add "with", "and", "from", "for", "the", "by", "of", "in", "on", "to" to the `noiseWords` map so they're stripped before the cap.

**Verification:** Spotify generates as `spotify-web` or `spotify` (not the 60+ char monstrosity). Passes all 7 gates.

### Unit 3: Handle Missing `servers` Field

**Files:** `internal/openapi/parser.go`, `internal/spec/spec.go`

**Root cause:** When `doc.Servers` is empty or nil, `baseURL` stays empty string. Then `spec.Validate()` rejects it with "base_url is required".

**Fix:**
1. In `parser.go` after the servers block (around line 56), if both `baseURL` and `basePath` are empty, set a placeholder:
```go
if baseURL == "" && basePath == "" {
    warnf("no servers defined in spec; generated CLI will require base_url in config")
    // Use a placeholder that doctor will report as "not configured"
    baseURL = "https://api.example.com"
}
```

2. In `spec.go:Validate()`, relax the base_url check - allow empty string but warn (or remove the requirement entirely since doctor already checks it at runtime).

**Verification:** Supabase parses without error and passes all 7 gates. Doctor shows the placeholder URL.

### Unit 4: Sanitize Enum Values with Slashes

**Files:** `internal/openapi/parser.go`

**Root cause:** Trello uses enum values like `modelTypes/card`, `modelTypes/board`. These flow through to the `types.go.tmpl` template where they become Go identifiers, causing syntax errors from `/`.

The fix overlaps with Unit 1 - the `sanitizeFieldName` function should also be applied to enum values. But the enum values flow through the `mapSchemaType` function and the spec's `TypeField.Type` string.

**Fix:** The enum values are used in the generated types template. Check where enum types are collected in `mapTypes` / `mapSchemaType` and apply sanitization:

```go
// In mapSchemaType or wherever enum values are extracted:
for i, v := range enumValues {
    enumValues[i] = sanitizeEnumValue(v)
}
```

```go
func sanitizeEnumValue(val string) string {
    return strings.NewReplacer("/", "_", ".", "_", " ", "_", "-", "_").Replace(val)
}
```

**Verification:** Trello passes `go build` gate.

### Unit 5: Resolve Server URL Templates

**Files:** `internal/openapi/parser.go`

**Root cause:** OpenAPI 3.0 server URLs can contain template variables like `https://{region}.sentry.io`. The `kin-openapi` library parses `doc.Servers[0].Variables` which contains default values. The parser ignores these and uses the raw URL with `{variable}` literals.

**Fix:** After extracting the server URL (line 45), resolve any template variables:

```go
serverURL := strings.TrimRight(strings.TrimSpace(doc.Servers[0].URL), "/")
// Resolve server URL template variables
if strings.Contains(serverURL, "{") && doc.Servers[0].Variables != nil {
    for varName, variable := range doc.Servers[0].Variables {
        if variable != nil && variable.Default != "" {
            serverURL = strings.ReplaceAll(serverURL, "{"+varName+"}", variable.Default)
        }
    }
}
// If any unresolved variables remain, strip them with a sensible fallback
if strings.Contains(serverURL, "{") {
    // Remove the template variable and surrounding dots/slashes
    // e.g., "{token}" in Telegram's "https://api.telegram.org/bot{token}" -> just strip it
    for {
        start := strings.Index(serverURL, "{")
        if start == -1 {
            break
        }
        end := strings.Index(serverURL, "}")
        if end == -1 {
            break
        }
        serverURL = serverURL[:start] + serverURL[end+1:]
    }
    serverURL = strings.ReplaceAll(serverURL, "//", "/")
    serverURL = strings.TrimRight(serverURL, "/")
}
```

**Verification:** Sentry doctor shows `https://us.sentry.io` (using default region). Telegram doctor shows resolved URL without `{token}`.

### Unit 6: Fallback to `x-api-name` for Missing `info.title`

**Files:** `internal/openapi/parser.go`

**Root cause:** Cloudflare's spec has no `info.title` but uses `x-api-name` extension. The parser falls back to "api" as the name, then `spec.Validate()` fails with "name is required" because the spec also lacks other required fields.

**Fix:** In the `Parse` function (lines 31-40), check for common extensions:

```go
name := "api"
description := ""
version := ""
if doc.Info != nil {
    if v := cleanSpecName(doc.Info.Title); v != "" && v != "api" {
        name = v
    } else if ext, ok := doc.Info.Extensions["x-api-name"]; ok {
        if s, ok := ext.(string); ok {
            name = cleanSpecName(s)
        }
    }
    description = strings.TrimSpace(doc.Info.Description)
    version = strings.TrimSpace(doc.Info.Version)
}
```

**Verification:** Cloudflare parses without "name is required" error. May still fail on other gates due to spec size (1716 paths) but parsing should succeed.

## Execution Loop

After each unit, run the dogfood loop:

```bash
go test ./...
go build -o ./printing-press ./cmd/printing-press

# Re-run the failing APIs
for api in flyio spotify vercel supabase trello jira cloudflare; do
    rm -rf /tmp/dogfood-${api}-cli 2>/dev/null
    ./printing-press generate --spec <url> --output /tmp/dogfood-${api}-cli 2>&1 | tail -3
done
```

Track which APIs flip from FAIL to PASS after each fix.

## Scope Boundaries

- Only modify parser.go and spec.go - don't change templates unless enum sanitization requires it
- Don't fix the "flat api resource" problem (Sentry/LaunchDarkly) - that's a separate enhancement
- Don't fix the truncated description in --help banner - cosmetic, separate fix
- Don't change the resource/endpoint limits (50/50)
- Every fix must not break existing passing specs (petstore, stytch, discord, gmail, telegram, sentry, launchdarkly)

## Sources

- Dogfood gauntlet findings: `docs/plans/dogfood-gauntlet-findings.md`
- Parser: `internal/openapi/parser.go` (1900+ lines)
- Spec types: `internal/spec/spec.go`
- Type template: `internal/generator/templates/types.go.tmpl`
- Existing sanitizers: `sanitizeResourceName` (line 1495), `cleanSpecName` (line 1526)
- kin-openapi server variables: `doc.Servers[0].Variables` (map of `*openapi3.ServerVariable`)
