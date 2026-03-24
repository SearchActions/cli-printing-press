---
title: "Fix Template Quality: Steinberger Parity"
type: fix
status: active
date: 2026-03-23
---

# Fix Template Quality: Steinberger Parity

## Overview

The printing press generates CLIs that compile but that Steinberger would never ship. This plan fixes the generator through an iterative dogfooding loop: fix the press, delete the output, regenerate, compare, fix more, repeat. The test subjects are three real-world OpenAPI specs at increasing difficulty: Petstore (trivial), Stytch (medium, 47K lines), Discord (hard, 1MB, currently breaks).

## The Dogfooding Loop

This is the core execution pattern. Every fix goes into the generator (parser or templates), never into generated output. Generated CLIs are disposable test artifacts.

```
while (quality bar not met):
  1. Fix the printing press (parser.go, templates, generator.go)
  2. Run go test ./... (verify generator tests pass)
  3. Delete all generated output (rm -rf *-cli/)
  4. Regenerate from all 3 test specs:
     - ./printing-press generate --spec testdata/openapi/petstore.yaml
     - ./printing-press generate --spec testdata/openapi/stytch.yaml
     - ./printing-press generate --spec testdata/openapi/discord.json
  5. Score each generated CLI:
     - Does it pass all 7 quality gates?
     - Are command names clean? (no api_user_v1_ prefixes)
     - Would Steinberger put his name on it?
  6. If any spec fails or output is ugly: identify the root cause, go to step 1
  7. If all 3 pass and look professional: done
```

## What's Broken (Current State)

### Petstore: compiles, looks okay
- Command names: `add_pet`, `delete_pet`, `get_pet_by_id` - acceptable but could be cleaner (`add`, `delete`, `get`)
- All 7 gates pass

### Stytch: compiles, ugly names
- Command names: `api_user_v1_create`, `api_user_v1_delete_biometric_registration` - unusable
- Root cause: `operationIDToName` in `parser.go:935` copies operationIds verbatim
- All 7 gates pass

### Discord: does not compile
- PATCH endpoints produce broken Go code (undefined `data` variable)
- Unexported struct fields (`_errors`) with JSON tags fail `go vet`
- Root cause: command template has no PATCH case, types template doesn't filter `_` fields

## The Comparison (Why This Matters)

### Official stytch-cli vs our generated stytch-api-cli

| Aspect | Official (archived hackweek, 5 stars) | Ours (generated) |
|--------|--------------------------------------|-----------------|
| Commands | `projects list`, `secrets create` | `api_user_v1_create` |
| Auth | OAuth2 PKCE + keyring | Plaintext config file |
| Interactive | Yes (`promptui`) | No |

They wrap different APIs (management vs consumer) so they don't overlap. But the official one shows the quality bar.

### Steinberger patterns we're missing

| Pattern | Gap | Priority |
|---------|-----|----------|
| Clean noun-verb commands | operationId cleaning | HIGH - fixes Stytch |
| PATCH method | Template case missing | HIGH - fixes Discord |
| Valid struct fields | Skip `_` prefixed props | HIGH - fixes Discord |
| TTY detection + auto-JSON | Not implemented | MEDIUM |
| Color output | Not implemented | LOW |
| Keyring auth | Not implemented | LOW |

## Fixes (In Dogfooding Order)

Each fix is applied to the press, then all 3 specs are regenerated and tested.

### Round 1: Make Discord compile

**Fix 1a: Add PATCH to command template**

Add PATCH case to `command.go.tmpl` mirroring PUT (JSON body). Add `Patch` method to `client.go.tmpl`.

Files: `internal/generator/templates/command.go.tmpl`, `internal/generator/templates/client.go.tmpl`

**Fix 1b: Skip unexported fields in type generation**

In `mapTypes` in `parser.go`, skip schema properties whose names start with `_`. They can't be exported Go struct fields.

Files: `internal/openapi/parser.go`, `internal/openapi/parser_test.go`

**Dogfood checkpoint:** Delete all `*-cli/` dirs. Regenerate all 3. Discord should now pass `go vet` and `go build`. If not, identify next blocker and fix.

### Round 2: Clean up command names

**Fix 2: Smart operationId cleaning**

Replace the naive `operationIDToName` in `parser.go`:

1. Strip `api_` prefix
2. Strip resource name prefix (e.g., `user_` when resource is `users`)
3. Strip version segments (`v1_`, `v2_`)
4. For CamelCase IDs (`listPets`), split on case boundaries and strip resource prefix
5. Map common patterns to CRUD verbs when possible

Examples:
- `api_user_v1_create` (resource: users) -> `create`
- `api_user_v1_delete_biometric_registration` -> `delete-biometric-registration`
- `listPets` (resource: pet) -> `list`
- `GetApplicationCommandPermissions` (resource: applications) -> `get-command-permissions`

Files: `internal/openapi/parser.go`, `internal/openapi/parser_test.go`

**Dogfood checkpoint:** Delete all `*-cli/` dirs. Regenerate all 3. Run `stytch-api-cli users --help` and verify command names are clean. Run `discord-cli --help` and verify command names make sense. If names are still ugly for specific patterns, add more cleaning rules and regenerate.

### Round 3: Polish (if time permits)

**Enhancement 3a: TTY detection**

When stdout is not a terminal, auto-enable `--json`. Every Steinberger CLI does this.

Files: `internal/generator/templates/root.go.tmpl`, `internal/generator/templates/go.mod.tmpl`

**Enhancement 3b: Color output**

ANSI color for table headers and doctor output, only when TTY. Inline ANSI codes (no dependency).

Files: `internal/generator/templates/root.go.tmpl`, `internal/generator/templates/doctor.go.tmpl`

**Dogfood checkpoint:** Delete all `*-cli/` dirs. Regenerate. Run in terminal and piped to see color/no-color behavior. Verify `stytch-api-cli users list 2>&1 | cat` produces no ANSI codes.

## Acceptance Criteria

Measured by regenerating from all 3 specs after every change:

- [ ] Petstore: passes all 7 gates, clean command names
- [ ] Stytch: passes all 7 gates, `users create` not `api_user_v1_create`
- [ ] Discord: passes all 7 gates (currently fails entirely)
- [ ] No regressions on internal-format specs (stytch.yaml, clerk.yaml, loops.yaml)
- [ ] No generated `*-cli/` directories committed to repo (they're test artifacts)

## Scope Boundaries

- Not fixing the official Stytch CLI (archived, not ours)
- Not adding OAuth2 PKCE flows
- Not adding interactive prompts (`promptui`)
- Not adding retry/rate-limiting
- Not consolidating `--json`/`--plain`/`--quiet` into `--format`
- Keyring auth is stretch goal only

## Sources

- Official Stytch CLI: https://github.com/stytchauth/stytch-cli (archived, Go, OAuth2 PKCE + keyring)
- Steinberger patterns: `docs/plans/2026-03-23-feat-cli-printing-press-plan.md` (lines 62-143)
- Discord failure: `go vet` on PATCH endpoints + `_errors` unexported fields
- Stytch ugly names: operationIds copied verbatim from spec
- Command template: `internal/generator/templates/command.go.tmpl:53-81`
- Parser operationId: `internal/openapi/parser.go:935`
