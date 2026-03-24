---
title: "Fix Gmail CLI Dogfood Gaps - Deep Sub-Resources, Global Params, Endpoint Cap"
type: fix
status: completed
date: 2026-03-24
---

# Fix Gmail CLI Dogfood Gaps

## The 4 Problems

The press prints a Gmail CLI that compiles and passes 7 gates, but it's not usable. Four specific problems found during dogfooding:

### 1. Flat command names (no deep sub-resources)

**Current:** `gmail-cli gmail users-messages-list <userId>`
**Want:** `gmail-cli messages list <userId>`

Gmail's paths are `/gmail/v1/users/{userId}/messages/{id}`. The parser detects `gmail` as the primary resource and `users` as the first sub-resource after the `v1` version segment. But `messages` is a sub-sub-resource (two levels deep) and gets flattened into the endpoint name instead of becoming its own command group.

**Fix:** Make sub-resource detection recursive. Walk the path segments, skip params, and create nested sub-resources for each non-param segment after the primary. For Gmail:
- `/gmail/v1/users/{userId}/messages` -> resource: `gmail`, sub: `users`, sub-sub: `messages`
- Or better: detect that `users/{userId}` is just a required context param (every Gmail endpoint has it) and treat `messages`, `labels`, `drafts`, `threads` as the real top-level resources

**Approach:** When all paths under a resource share the same path prefix pattern (e.g., all start with `/gmail/v1/users/{userId}/`), collapse that prefix into required positional args and use the next segment as the resource name.

**Files:** `internal/openapi/parser.go` (resourceAndSubFromPath, mapResources)

**Acceptance:**
- [ ] `gmail-cli messages list <userId>` instead of `gmail-cli gmail users-messages-list <userId>`
- [ ] `gmail-cli labels list <userId>` instead of `gmail-cli gmail users-labels-list <userId>`
- [ ] `gmail-cli drafts create <userId>` instead of `gmail-cli gmail users-drafts-create <userId>`

### 2. Google global params on every command

**Current:** Every Gmail command shows `--access-token`, `--alt`, `--callback`, `--fields`, `--key`, `--oauth-token`, `--pretty-print`, `--quota-user`, `--upload-protocol`, `--upload-type` - 10 extra flags that clutter the help.

These are Google API-wide parameters defined at the path level or as common parameters, not endpoint-specific.

**Fix:** Detect "global" parameters that appear on every single endpoint in the spec. Move them to persistent flags on the root command (or filter them out entirely since they're rarely used by humans).

**Approach:** After parsing all endpoints, count parameter frequency. Any param that appears on >80% of endpoints is "global" - remove it from individual endpoint params and optionally add it as a persistent root flag.

**Files:** `internal/openapi/parser.go` (new post-processing step after mapResources)

**Acceptance:**
- [ ] `gmail-cli messages list --help` does NOT show `--alt`, `--callback`, `--fields`, etc.
- [ ] Global params are either on root command or omitted entirely
- [ ] Endpoint-specific params (like `--q`, `--maxResults`, `--labelIds`) still show

### 3. Endpoint cap hit (20 per resource)

**Current:** `watch` and `untrash` endpoints skipped because `gmail` resource hit 20-endpoint cap.

**Fix:** With deep sub-resources (fix #1), endpoints spread across `messages`, `labels`, `drafts`, `threads`, `history`, `settings` sub-resources. Each gets its own 20-endpoint limit. This should resolve naturally once fix #1 lands.

**Acceptance:**
- [ ] Zero "endpoint limit reached" warnings for Gmail spec
- [ ] All 49 Gmail paths generate commands

### 4. No OAuth2 auth flow

**Current:** `gmail-cli doctor` shows `auth=not configured`. Gmail requires OAuth2 - there's no API key option.

**This is a bigger feature (from the GOAT plan A2) - not fixing in this plan.** This plan focuses on the structural issues (#1-3) that are press bugs. OAuth2 is a new feature that needs its own plan with browser callback server, token storage, refresh flow.

**Workaround for testing:** User can manually set `--access-token` flag or configure `auth_header` in config.toml with a token obtained from Google's OAuth2 Playground.

## Execution Loop

```
while (not good):
  1. Fix the press (parser.go only for #1-3)
  2. go test ./...
  3. Regenerate all 4 specs (Petstore, Stytch, Discord, Gmail)
  4. Check gmail-cli help output
  5. Find what's still wrong
  6. Go to 1
```

## Scope

- Only modify `internal/openapi/parser.go`
- Don't break Petstore, Stytch, or Discord
- Don't add OAuth2 (separate plan)
- Don't change templates
