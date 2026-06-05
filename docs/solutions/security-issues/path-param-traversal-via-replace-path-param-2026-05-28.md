---
title: Path-parameter traversal via replacePathParam (per-segment escape)
date: 2026-05-28
category: security-issues
module: generator-templates
problem_type: security_issue
component: helpers.go.tmpl
symptoms:
  - GET properties/../accounts/X resolved to a different endpoint on the same host
  - Per-segment url.PathEscape preserved "." and ".." segments verbatim
  - Original whole-value url.PathEscape encoded "/" to "%2F" and produced 404 for legitimate composite resource names (e.g. Google REST "properties/456314183")
root_cause: missing_validation
resolution_type: code_fix
severity: defence-in-depth
tags:
  - path-traversal
  - url-path-escape
  - generator-template
  - security
  - mcp-agent-input
incidents:
  - INC-2026-147
---

# Path-parameter traversal via replacePathParam (per-segment escape)

## Problem

The `replacePathParam` helper that `helpers.go.tmpl` emits into every generated CLI was vulnerable in two complementary ways:

1. **Whole-value `url.PathEscape(value)`** broke composite resource names. Google REST identifiers like `properties/456314183` legitimately contain literal `/` characters; whole-value escape turned them into `properties%2F456314183` and every templated call returned HTTP 404.
2. **The naive fix — split on `/`, escape per-segment** — preserved structural slashes (fixing the 404) but let `..`, `.`, and empty segments survive into the templated path. A caller could supply `value="properties/../foo:doSomething"` and the URL router on the API host would resolve the `..`, redirecting the authenticated request to a different operation on the same host.

The blast radius is bounded but non-zero. OAuth scope on the credential restricts which APIs respond, and per-API host segregation (e.g. `analyticsdata.googleapis.com` only serves the Data API) prevents pivoting to e.g. Gmail. The realistic exposure is "different operation on the same API with the same credential," which becomes material when an MCP server fronts the CLI on agent-controlled (untrusted) input. SEV:3 (defence-in-depth), not SEV:2.

## Symptoms

- `properties/456314183:runReport` returned `HTTP 404 <HTML page>` because `/` was encoded to `%2F`
- After the naive per-segment fix, `properties/../accounts/<other-account>:runReport` returned real data from the attacker-chosen `accounts/...` endpoint
- The downstream `google-analytics-pp-cli` had an in-tree patch applied (the `// DO NOT EDIT` file said so), but the next regeneration from the unfixed template would have silently reverted the fix

## What Didn't Work

- **No escaping at all** (template state before the bug was discovered): user input with `?`, `#`, spaces, or other path-reserved characters landed raw in URLs
- **Whole-value `url.PathEscape`**: encoded structural `/` in composite resource names → 404 on every templated call
- **Per-segment `url.PathEscape` without traversal rejection**: preserved `/` correctly but allowed `..`/`.`/empty segments to redirect the request

## Solution

`replacePathParam` now splits on `/`, **inspects every segment** for `""`, `"."`, or `".."`, and falls back to whole-value escape when any are present:

```go
func replacePathParam(path, name, value string) string {
    segments := strings.Split(value, "/")
    for _, s := range segments {
        if s == "" || s == "." || s == ".." {
            // Reject traversal: fall back to whole-value escape so structural
            // slashes get encoded too. The API then 404s rather than letting
            // the URL router resolve ".." and redirect an authenticated
            // request to an attacker-chosen endpoint on the same host.
            return strings.ReplaceAll(path, "{"+name+"}", url.PathEscape(value))
        }
    }
    for i, s := range segments {
        segments[i] = url.PathEscape(s)
    }
    return strings.ReplaceAll(path, "{"+name+"}", strings.Join(segments, "/"))
}
```

Legitimate composite resource names (no empty/dot segments) take the per-segment escape path and continue to work. Anything that would let the URL router collapse segments takes the whole-value fallback, which encodes every `/` to `%2F` and causes the API to 404 — far better than silently serving an attacker-chosen URL.

Per-segment `url.PathEscape` also neutralizes `?` and `#` injection inside a segment (encoded to `%3F` and `%23`), so a caller cannot graft a query string or fragment onto the URL via a path parameter.

## Tests

`internal/generator/replace_path_param_test.go` has two complementary tests:

- `TestReplacePathParamBehavior` — table-driven against a byte-identical local copy of the emitted function, covering all 10 cases the original kamakazi finding called out (legitimate resource names, multi-level resources, spaces, `..`/`.` traversal, leading/trailing/empty segments, and `?`/`#` injection)
- `TestReplacePathParamTemplateMatchesLocalCopy` — generates a CLI with a path-param endpoint and asserts the emitted `helpers.go` contains the traversal-rejection substring (`s == "" || s == "." || s == ".."`), the whole-value-fallback substring (`url.PathEscape(value)`), the per-segment-escape substring (`segments[i] = url.PathEscape(s)`), and the conditional `"net/url"` import. This is the bridge that prevents template drift from silently invalidating the table-driven cases.

## Downstream propagation

The buggy helper was templated into every CLI the press has generated to date. Downstream consumers fall in three states:

1. **Hand-patched in-tree** (`D:/projects/cli/google-analytics-cli/internal/cli/helpers.go` as of 2026-05-28). The next regeneration will re-apply the press's version — once the press is on a release containing this fix, regeneration is the safer path; until then, the hand patch must be preserved.
2. **Unpatched, shipped with no escape**: vulnerable to both the 404 (when the API needs slash-preservation) and any path-reserved characters in user input. Regenerate against a press release that contains this fix, or apply the patch manually.
3. **Unpatched, shipped with naive per-segment escape**: same traversal exposure as the naively-fixed google-analytics-pp-cli before this fix landed. Regenerate.

The emitted-helper version that downstream CLIs notice via `.printing-press.json` is `printing_press_version`, which release-please bumps on the next release PR triggered by the `fix:` conventional commit that lands this change. There is no separate helper-version mechanism; the press version is the cache-invalidation token.

## Lessons

- When fixing a path-escape bug, always think about what traversal characters survive the new escape strategy. "Split on `/` and escape per segment" is a common shortcut and a common security regression.
- Generator-templated helpers carry their security properties into every downstream consumer. A fix in one generated CLI is invisible to siblings. Always file the upstream fix and an intake note when patching a templated file; the `// DO NOT EDIT` header is a regen-revert risk.
- Bind a template-rendered behavior to a guardrail test that re-reads the generated source. A local-copy unit test alone permits silent template drift.

## See also

- INC-2026-147 (dev-system incident log) — original SEV classification, blast-radius analysis, and downstream-patch record
- `D:/projects/dev-system/_system/librarian/intake/kamakazi-2026-05-28-printing-press-pathescape.md` — original kamakazi findings and reproduction
- `docs/solutions/security-issues/filepath-join-traversal-with-user-input-2026-03-29.md` — adjacent traversal pattern for `filepath.Join` with untrusted segments
