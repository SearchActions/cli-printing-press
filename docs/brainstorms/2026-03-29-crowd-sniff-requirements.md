---
date: 2026-03-29
topic: crowd-sniff
---

# Crowd Sniff: API Discovery from Community Signals

## Problem Frame

Printing-press generates CLIs from API specs, but many public APIs lack published specs. The existing `sniff` command solves this by observing live web traffic -- but it only captures what one browsing session happens to hit. It can't tell you which endpoints developers actually use, which params matter, or how popular each operation is.

Meanwhile, thousands of developers have already mapped these APIs in npm packages, GitHub code, and Postman collections. This crowd knowledge is structured, tested, and popularity-weighted -- but nobody systematically extracts it for CLI generation.

Crowd sniff mines these community signals to either discover API endpoints (when no spec exists) or enrich existing specs with real-world usage data (which endpoints matter most, which params people actually pass, which auth patterns work).

## How It Fits

```
Phase 1 Research
       |
       v
Phase 1.7: Sniff Gate (existing)
       |
       v
Phase 1.8: Crowd Sniff Gate (NEW)
       |
       v
  +----+----+
  |         |
 Has spec   No spec
  |         |
  v         v
Enrichment  Primary Discovery
mode        mode
  |         |
  v         v
crowd-sniff-spec.yaml
  |
  v
Phase 2: Generate
(--spec original --spec crowd)
```

Crowd sniff is a **new CLI command** (`printing-press crowd-sniff`) that outputs spec YAML -- same integration surface as `sniff`. The skill invokes it during a new Phase 1.8 gate when appropriate.

## Requirements

**Sources & Search**

- R1. Search npm registry for SDK packages matching the target API (by name, keywords, scope). For each candidate package (capped at 10), download the tarball from the registry, decompress to a temp directory, grep source for endpoint patterns (URL paths, HTTP method calls, TypeScript interfaces), extract findings, and clean up. Skip tarballs over 10MB. No AST parsing in v1.
- R2. Search GitHub code for usage patterns (`fetch("https://api.example.com/`, `requests.get(`, client library method calls). Aggregate across repos to extract endpoint URLs, common query parameters, and per-endpoint frequency counts (number of distinct repos using that endpoint). Use GitHub's code search API. GitHub code search requires authentication (`GITHUB_TOKEN` or `gh` CLI) and has a rate limit of 10 requests/minute -- use minimum 6-second intervals between requests.
- R3. Search Postman's public API network for collections matching the target API. Parse collection JSON to extract endpoints, request/response examples, auth configurations, and environment variables. Programmatic access to Postman Explore is an open question -- a printing-press CLI for the Postman Explore site is in progress, which will clarify the API surface.
- R4. Each source is optional and independently useful. When a source is unavailable (no auth token, API unreachable, zero results), it silently returns zero endpoints and the other sources still run.

**Recency**

- R5. Strict 6-month recency cutoff on all sources. npm packages must have been published/updated within 6 months. GitHub code must come from repos pushed within 6 months. Postman collections must have been updated within 6 months. Sources outside this window are excluded entirely.
- R6. Recency is a binary filter only (in or out). It does not affect confidence scoring within the window.

**Confidence**

- R7. Each discovered endpoint carries two metadata fields:
  - `source_tier`: one of `official-sdk` (published by the API vendor), `community-sdk` (third-party npm package), `code-search` (GitHub code), or `postman`.
  - `source_count`: how many independent sources found this endpoint.
- R8. Source tier reflects authority: official SDK > community SDK > code search > Postman. Cross-source agreement (higher `source_count`) is an additional quality signal. The command outputs these fields; the skill/generator decides how to use them.
- R9. Confidence metadata is stored in a `Meta map[string]string` field on `spec.Endpoint` (with `yaml:"meta,omitempty"`). Crowd sniff sets `source_tier` and `source_count` as map entries. Existing specs, the generator, and `mergeSpecs` are unaffected because the field is omitempty and templates don't reference it.

**Input & Output**

- R10. Output is a valid printing-press spec YAML file, same format as `sniff` output. Can be passed to `printing-press generate --spec <path>` or merged with other specs via multi-`--spec`.
- R11. The command is `printing-press crowd-sniff --api <name-or-url>`. Accepts an API name ("notion", "stripe") or a base URL. Outputs spec YAML to a default cache path or `--output <path>`.
- R12. Report summary on completion: how many endpoints discovered, from which sources, source tier distribution. Similar to sniff's "N endpoints across M resources" output.
- R13. When `--api` is ambiguous (e.g., "cal" could match cal.com or Google Calendar), display candidates with context (npm download counts, GitHub stars) and prompt the user to select. Do not silently pick one.
- R14. Base URL resolution: prefer base URL from official SDK configuration/constants, fall back to most frequently observed domain prefix from GitHub code search results, fail with error asking user to provide `--base-url` if no URL can be inferred.

**Skill Integration**

- R15. The printing-press skill adds a Phase 1.8 "Crowd Sniff Gate" after the existing Phase 1.7 Sniff Gate. Decision matrix mirrors sniff's: offer crowd sniff when spec has gaps or no spec exists. Skip when spec appears complete.
- R16. Crowd sniff can run independently of sniff. They are complementary -- sniff discovers from live traffic, crowd sniff discovers from community usage. Both output spec YAML that merges via `--spec`.
- R17. For npm SDK discovery, prefer official SDKs (scope matches API vendor, e.g., `@notionhq/client`) over community packages. Use npm registry search API with keyword and scope filters.

## Success Criteria

- The source tier correctly ranks official SDK endpoints above code-search-only endpoints in the output metadata.
- End-to-end: `printing-press crowd-sniff --api notion` produces a spec YAML that `printing-press generate` can consume without errors.
- Recency filtering successfully excludes deprecated endpoints from abandoned packages/collections.
- Manual validation against 3 popular APIs (e.g., Notion, Discord, Stripe) shows crowd sniff discovers a meaningful subset of known endpoints with high accuracy (discovered endpoints actually exist in the current API).

## Scope Boundaries

- **No AST parsing in v1.** Heuristic grep patterns only. AST parsing for JS/TS is a potential v2 enhancement.
- **No PyPI/RubyGems in v1.** npm only for SDK analysis. Other registries are a natural extension but not initial scope.
- **No automatic threshold decisions.** The command outputs all discovered endpoints with confidence metadata. The skill decides the threshold, not the command.
- **No live API probing.** Crowd sniff is passive -- it reads what others have published. Active endpoint probing (hitting the API to check 200 vs 404) is a separate concern.
- **GraphQL introspection is out of scope.** Valuable but mechanically different from crowd signal mining. Could be a separate command.
- **Postman access is an open question.** A printing-press CLI for the Postman Explore site is in progress separately. Its outcome will determine how crowd sniff integrates with Postman. If no programmatic access materializes, Postman is dropped from v1.

## Key Decisions

- **Standalone CLI command** over skill-only integration: Enables manual invocation, testability, and clean separation from orchestration logic. Same pattern as `sniff`.
- **Strict 6-month cutoff** over adaptive scoring: Simpler, avoids false confidence from stale sources. We'd rather miss a deprecated endpoint than include one.
- **source_tier + source_count** over weighted numerical scoring: Achieves ranking without specifying a formula nobody consumes in v1. The skill/generator can interpret tiers however it wants. Weighted scoring is a v2 upgrade if needed.
- **Meta map on Endpoint** over sidecar file: `Meta map[string]string` with `yaml:"meta,omitempty"` on `spec.Endpoint`. Zero impact on existing specs (omitempty), generator (templates don't reference it), and mergeSpecs (copies the struct wholesale). Same pattern as existing optional fields like `ResponsePath`.
- **Heuristic grep** over AST parsing: Language-agnostic, fast, proven by OSC's convention analysis patterns. Good enough for v1; AST parsing is a v2 upgrade path.
- **Phase 1.8 (after sniff)** rather than replacing sniff: They discover different things. Sniff finds what the web app does; crowd sniff finds what developers need. Complementary, not competing.

## Dependencies / Assumptions

- GitHub code search API requires authentication (`GITHUB_TOKEN` env var or `gh` CLI auth). Rate limit: 10 requests/minute. When no token is available, the GitHub source silently returns zero results (per R4).
- npm registry search API is public and unauthenticated.
- The existing `mergeSpecs()` in `root.go` handles combining crowd-sniff output with other spec sources. Note: `mergeSpecs` does not currently preserve per-endpoint metadata -- this may need updating if confidence fields are embedded in the spec struct.

## Outstanding Questions

### Deferred to Planning

- [Affects R2][Needs research] GitHub code search API has known limitations (only indexed repos, max 1000 results per query). What's the practical endpoint discovery ceiling for a mid-popularity API?
- [Affects R3][Blocked on other work] Postman Explore API access depends on the in-progress printing-press CLI for Postman Explore. That work will reveal the API surface. Crowd sniff's Postman source can be built once that CLI ships.
- [Affects R1][Technical] What grep patterns reliably extract endpoint URLs from SDK source across different coding styles? Prototype against 3-5 real SDKs (Notion, Linear, Discord, Stripe, Twilio).
- [Affects R1][Technical] npm tarball workflow: search registry, download tarballs, decompress, grep, clean up. Prototype the end-to-end path in Go including temp directory management.
- [Affects R14][Technical] Base URL extraction from SDK source -- what patterns identify the base URL constant in common SDK styles?
- [Affects R7][Needs research] What npm download count threshold reliably separates maintained community SDKs from abandoned ones? Proposed: 100/week. Validate against real data for mid-tier APIs.
- [Affects performance] Expected wall-clock time given GitHub rate limits (~6s between requests). Should sources be queried in parallel (npm + GitHub concurrently)?

## Next Steps

-> `/ce:plan` for structured implementation planning. No blocking questions remain.
