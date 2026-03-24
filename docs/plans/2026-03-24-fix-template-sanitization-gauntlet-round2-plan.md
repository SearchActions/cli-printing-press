---
title: "Fix Template Sanitization - Gauntlet Round 2 (4/10 to 8/10+)"
type: fix
status: completed
date: 2026-03-24
origin: docs/plans/dogfood-gauntlet-findings.md
---

# Fix Template Sanitization - Gauntlet Round 2

## Overview

Round 1 parser fixes (commit 3f096bc) sanitized type names in `mapTypes` and unlocked Jira. But 6 APIs still fail because the `$`, `/`, and other special characters also flow through to:
- Template helper functions (`toCamel`, `flagName`, `title`) that don't strip special chars
- `types.go.tmpl` which uses `{{$name}}` directly (Go template variable, not the sanitized name)
- `command.go.tmpl` which uses param/body names via `{{camel .Name}}`

The fix: harden the template helper functions so any name that passes through `toCamel`, `flagName`, or `goType` is automatically safe for Go code.

## Remaining Failures

| API | Error | Root Cause |
|-----|-------|------------|
| Vercel | `$` in `deployments.go` flag names | Param names like `$schema` pass through `camel` -> `$Schema` (invalid Go) |
| Trello | `/` in `boards.go` type references | Enum values like `modelTypes/card` become Go identifiers via templates |
| Fly.io | `Version` redeclared in types.go | Two schemas sanitize to same name, causing Go redeclaration |
| Spotify | Panic in command registration | CLI name `spotify-web-sonallux` triggers Cobra panic (deeper issue in command tree) |
| Supabase | `"true"` string as bool flag | Param type mapped as `string` but default value `"true"` used with `BoolVar` |
| Cloudflare | `name is required` | x-api-name extension is nested under `x-api-id` not `x-api-name`, parser didn't find it |

## Acceptance Criteria

- [ ] `toCamel` strips `$`, `.`, `/`, `\` before camelCasing
- [ ] `flagName` strips `$` prefix
- [ ] `types.go.tmpl` uses sanitized type names (not raw Go template vars)
- [ ] Duplicate type names after sanitization are deduplicated (append suffix)
- [ ] Vercel passes all 7 gates
- [ ] Trello passes all 7 gates
- [ ] Fly.io passes all 7 gates
- [ ] Supabase passes all 7 gates
- [ ] Spotify passes all 7 gates (or has a clear, documented limitation)
- [ ] Existing passing APIs still pass (petstore, discord, gmail, telegram, sentry, launchdarkly, jira)
- [ ] `go test ./...` passes

## Implementation Units

### Unit 1: Harden toCamel to Strip Special Characters

**Files:** `internal/generator/generator.go`

**Root cause:** `toCamel` (line 165) only splits on `_`, `-`, ` `. Characters like `$`, `.`, `/` pass through verbatim.

**Fix:** Add these characters to the split function:

```go
func toCamel(s string) string {
    // Strip characters that are invalid in Go identifiers
    s = strings.TrimLeft(s, "$")
    parts := strings.FieldsFunc(s, func(r rune) bool {
        return r == '_' || r == '-' || r == ' ' || r == '.' || r == '/' || r == '\\' || r == '$'
    })
    for i, p := range parts {
        if len(p) > 0 {
            parts[i] = strings.ToUpper(p[:1]) + p[1:]
        }
    }
    result := strings.Join(parts, "")
    // Ensure starts with letter
    if len(result) > 0 && !unicode.IsLetter(rune(result[0])) {
        result = "V" + result
    }
    return result
}
```

**Verification:** `toCamel("$schema")` returns `"Schema"`, not `"$Schema"`. `toCamel("modelTypes/card")` returns `"ModelTypesCard"`.

### Unit 2: Harden flagName to Strip $ Prefix

**Files:** `internal/generator/generator.go`

**Fix:** Update `flagName` (line 336):

```go
func flagName(name string) string {
    name = strings.TrimLeft(name, "$")
    name = strings.ReplaceAll(name, "_", "-")
    name = strings.ReplaceAll(name, "/", "-")
    name = strings.ReplaceAll(name, ".", "-")
    return strings.Trim(name, "-")
}
```

### Unit 3: Fix types.go.tmpl Type Name Sanitization

**Files:** `internal/generator/templates/types.go.tmpl`, `internal/generator/generator.go`

**Root cause:** The template uses `{{$name}}` which is the raw map key. Even though `mapTypes` sanitizes the key in the parser, the template also needs a `safeTypeName` function for the struct declaration.

**Fix:** Add a `safeTypeName` template function in generator.go:

```go
"safeTypeName": safeTypeName,
```

```go
func safeTypeName(name string) string {
    // Strip $ prefix, replace dots/slashes with underscores
    name = strings.TrimLeft(name, "$")
    name = strings.NewReplacer(".", "_", "/", "_", "\\", "_", "-", "_").Replace(name)
    var b strings.Builder
    for _, r := range name {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
            b.WriteRune(r)
        }
    }
    result := b.String()
    if len(result) > 0 && !unicode.IsLetter(rune(result[0])) {
        result = "T" + result
    }
    return result
}
```

Update `types.go.tmpl`:
```
{{range $name, $typeDef := .Types}}
type {{safeTypeName $name}} struct {
{{- range $typeDef.Fields}}
	{{title .Name}} {{goType .Type}} `json:"{{.Name}}"`
{{- end}}
}
{{end}}
```

### Unit 4: Deduplicate Type Names After Sanitization

**Files:** `internal/openapi/parser.go`

**Root cause:** Fly.io has schemas that sanitize to the same Go name (e.g., `fly.Version` and `machines.Version` both become `Version`). This causes "redeclared" errors.

**Fix:** In `mapTypes`, after sanitizing the name, check for duplicates:

```go
usedNames := map[string]int{}
for _, name := range names {
    goName := sanitizeTypeName(name)
    if goName == "" {
        continue
    }
    if count, exists := usedNames[goName]; exists {
        goName = fmt.Sprintf("%s%d", goName, count+1)
        usedNames[goName[:len(goName)-1]] = count + 1
    } else {
        usedNames[goName] = 1
    }
    // ... rest of type mapping uses goName
}
```

### Unit 5: Fix Supabase Bool Flag Type Mismatch

**Files:** `internal/openapi/parser.go` or `internal/generator/generator.go`

**Root cause:** A parameter has type "boolean" in the spec but the parser maps it to "string". Then the template generates `cmd.Flags().BoolVar(&flagX, ...)` but `flagX` is declared as `string`.

**Fix:** Check where boolean types flow through `mapSchemaType` in the parser. Ensure OpenAPI `type: boolean` maps to spec type `bool`, not `string`. Also check that `cobraFlagFunc` and `goType` agree - if the param type is `bool`, the flag must be `BoolVar` and the variable must be `bool`.

### Unit 6: Investigate Spotify Panic

**Files:** investigate only

**Root cause:** Spotify's spec has deeply nested sub-resources that cause Cobra command registration to panic. The name is now shorter (`spotify-web-sonallux`) but the command tree construction still crashes.

**Approach:** Generate Spotify, read the panic stack trace, identify which command.go file has the issue. Likely a circular sub-resource reference or duplicate command name.

If the fix is simple (e.g., dedup command names), fix it. If it's a deeper Cobra issue, document as known limitation.

### Unit 7: Re-run Gauntlet and Update Findings

After all fixes, rebuild and re-run all 10 APIs. Update `docs/plans/dogfood-gauntlet-findings.md` with new scores. Add passing CLIs to catalog.

## Scope Boundaries

- Don't change the 50-resource/50-endpoint limits
- Don't fix the "flat api resource" problem (Sentry/LaunchDarkly) - separate enhancement
- Don't add new template features (table output, pagination) - just fix sanitization
- Don't modify test fixtures - the fixes should be backward compatible
- Every fix must not break existing passing specs

## Sources

- Generator helpers: `internal/generator/generator.go` (toCamel:165, flagName:336, goType:188)
- Types template: `internal/generator/templates/types.go.tmpl`
- Command template: `internal/generator/templates/command.go.tmpl`
- Parser type mapping: `internal/openapi/parser.go:mapTypes` (line 858)
- Gauntlet findings: `docs/plans/dogfood-gauntlet-findings.md`
- Round 1 fixes: commit 3f096bc
