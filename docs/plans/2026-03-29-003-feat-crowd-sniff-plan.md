---
title: "feat: Add crowd-sniff command for API discovery from community signals"
type: feat
status: active
date: 2026-03-29
origin: docs/brainstorms/2026-03-29-crowd-sniff-requirements.md
deepened: 2026-03-29
---

# feat: Add crowd-sniff command for API discovery from community signals

## Overview

Add a `printing-press crowd-sniff --api <name>` command that discovers API endpoints from npm SDKs and GitHub code search, producing spec YAML compatible with the existing `generate` pipeline. This complements the existing `sniff` command (which discovers from live web traffic) by mining what developers have already mapped in published packages and code.

## Problem Frame

Many public APIs lack published specs. `sniff` captures what one browsing session hits; crowd-sniff captures what thousands of developers actually use. The two are complementary -- sniff finds what the web app does, crowd-sniff finds what developers need. (see origin: `docs/brainstorms/2026-03-29-crowd-sniff-requirements.md`)

## Requirements Trace

- R1. npm SDK search, tarball download, heuristic grep for endpoints
- R2. GitHub code search for API usage patterns with frequency aggregation
- R3. Postman source (deferred -- blocked on in-progress Postman Explore CLI)
- R4. Graceful degradation: each source is optional and independent
- R5. Strict 6-month recency cutoff on all sources
- R6. Recency is binary filter only (in or out), no confidence weighting within window
- R7/R8. `source_tier` and `source_count` metadata per endpoint
- R9. `Meta map[string]string` field on `spec.Endpoint` with `yaml:"meta,omitempty"`
- R10. Output is valid spec YAML consumable by `printing-press generate`
- R11. `printing-press crowd-sniff --api <name-or-url>` with `--output` and `--base-url` flags
- R12. Report summary on completion: endpoint count, source breakdown, tier distribution
- R13. Disambiguation prompt when API name is ambiguous
- R14. Base URL resolution cascade: `--base-url` flag > SDK constants > GitHub code frequency
- R15. Phase 1.8 Crowd Sniff Gate in the skill
- R16. Crowd sniff runs independently of sniff (complementary, not dependent)
- R17. Prefer official SDKs (vendor-scoped npm packages) over community

## Scope Boundaries

- No AST parsing -- heuristic grep only (see origin)
- No PyPI/RubyGems -- npm only for v1 (see origin)
- No Postman in v1 -- blocked on separate work (see origin)
- No live API probing (see origin)
- No automatic threshold decisions -- output all endpoints with metadata (see origin)
- No GraphQL introspection (see origin)

## Context & Research

### Relevant Code and Patterns

| Purpose | File |
|---------|------|
| Command pattern to mirror | `internal/cli/sniff.go` -- `newSniffCmd()` returning `*cobra.Command` |
| Spec struct (add Meta) | `internal/spec/spec.go:46` -- `Endpoint` struct |
| Reusable spec writer | `internal/websniff/specgen.go` -- `WriteSpec()`, `DefaultCachePath()` |
| HTTP client + GitHub API | `internal/pipeline/research.go` -- `newGitHubRequest()`, error handling patterns |
| Spec caching pattern | `internal/cli/root.go` -- `fetchOrCacheSpec()` with SHA256 cache keys |
| Temp dir pattern | `internal/cli/scorecard.go:32` -- `os.MkdirTemp("", "prefix-*")` with `defer os.RemoveAll` |
| Command registration | `internal/cli/root.go:48` -- `rootCmd.AddCommand(...)` |
| Spec merging | `internal/cli/root.go:314` -- `mergeSpecs()` copies Resources/Types with collision prefixing |
| Test style | `internal/websniff/classifier_test.go` -- table-driven, `testify/assert`, `t.Parallel()` |
| Path sanitization | `internal/openapi/parser.go` -- `sanitizeResourceName()` |

### Institutional Learnings

- **Path traversal**: External identifiers (npm package names, GitHub repo names) must be sanitized before use in `filepath.Join`. Use belt-and-suspenders: validate for `..`/`/`/`\` AND verify resolved path stays within expected root. (from `docs/solutions/security-issues/filepath-join-traversal-with-user-input-2026-03-29.md`)
- **Validation must not mutate source**: Any temp artifacts from tarball extraction must use `os.MkdirTemp` with `defer` cleanup. (from `docs/solutions/best-practices/validation-must-not-mutate-source-directory-2026-03-29.md`)

### External API Constraints

| API | Auth | Rate Limit | Max Results | Recency Filter |
|-----|------|------------|-------------|----------------|
| npm search (`registry.npmjs.org/-/v1/search`) | None | Generous (no known cap) | Unlimited pagination | `package.date` in response |
| npm downloads (`api.npmjs.org/downloads/point/last-week/`) | None | Generous | Bulk: 128 packages | N/A |
| npm package meta (`registry.npmjs.org/<pkg>`) | None | Generous | N/A | N/A |
| GitHub code search (`api.github.com/search/code`) | Required (`GITHUB_TOKEN`) | **10 req/min** | **1000 total** | **Not supported in query** |
| GitHub repo info (`api.github.com/repos/{owner}/{repo}`) | Optional (recommended) | 5000 req/hr | N/A | `pushed_at` in response |

## Key Technical Decisions

- **New package `internal/crowdsniff/`**: Mirrors `internal/websniff/` -- separate package for discovery logic, CLI command in `internal/cli/crowd_sniff.go`. Keeps the new code isolated and testable.

- **Parallel source execution with name-based GitHub fallback**: npm and GitHub run concurrently via goroutines. npm searches by keyword directly. GitHub code search uses the base URL domain when `--base-url` is provided (e.g., `"api.notion.com" language:javascript`). When no base URL is known, GitHub falls back to name-based queries (e.g., `"notion" api fetch language:javascript`) — noisier but still useful, and compatible with parallel execution. npm has no rate limits and typically completes in 10-30s. GitHub is the bottleneck (~1-4 min). Total time is ~max(npm, github).

- **GitHub recency via repo info calls**: Code search can't filter by push date. After collecting unique repo names from code search results (typically 50-200 repos from 1000 max results), batch-check `pushed_at` via the standard repos API (5000 req/hr limit, separate pool from code search). Filter out repos not pushed within 6 months.

- **npm tarball workflow**: Search registry -> filter by `package.date` (6-month cutoff) -> fetch package metadata for tarball URL -> download tarball to `os.MkdirTemp` -> `archive/tar` + `compress/gzip` extraction -> grep source files -> cleanup. Cap at 10 packages, skip tarballs > 10MB.

- **Heuristic grep patterns over AST**: Match URL string literals (`"/v1/users"`, `"/api/projects"`), HTTP method calls (`this.get(`, `this.post(`, `fetch(`), base URL constants (`baseUrl`, `BASE_URL`, `apiBase`), and TypeScript type exports. Language-agnostic regex, no parser dependency.

- **`SourceResult` wrapper instead of bare slice**: Each source returns `SourceResult{Endpoints []DiscoveredEndpoint, BaseURLCandidates []string}` rather than just endpoints. This threads base URL signals (from SDK constants, from code frequency) through the interface so the aggregation layer can resolve R14 without a side channel. Without this, the CLI command would need to reach into source internals to extract base URLs.

- **Parameter syntax normalization**: SDK code uses `:id`, `{user_id}`, `<userId>`, `$id` for path parameters. GitHub code uses `{user_id}`, f-strings, etc. These must all normalize to `{id}` for deduplication to work. This is different from websniff's normalization (which replaces concrete values like `123`). Crowd-sniff needs both: syntax unification AND concrete-value replacement.

- **Testable HTTP clients with separate base URLs per service**: Sources accept configurable base URLs instead of hardcoding. npm needs two: `RegistryBaseURL` (defaults to `registry.npmjs.org`) and `DownloadsBaseURL` (defaults to `api.npmjs.org`) since these are different hosts. GitHub needs one `BaseURL` (defaults to `api.github.com`). Tests use `httptest.NewServer` with injected base URLs. This avoids repeating `research.go`'s untestable pattern.

- **Input sanitization at CLI boundary**: The `--api` value flows into HTTP URLs and file paths. URL-encode with `url.QueryEscape` before embedding in any API request. Validate at CLI boundary: reject newlines, null bytes, path separators, and `..`. Apply the belt-and-suspenders pattern from the learnings doc to the output cache path (validate input AND verify resolved path stays within cache root).

- **Tarball security: zip-slip AND symlink protection**: Skip `tar.TypeSymlink` and `tar.TypeLink` entries during extraction (a malicious symlink to `/etc/passwd` would let grep read arbitrary files). Use `io.LimitReader` capped at 10MB+1 as the size gate (Content-Length is unreliable with chunked encoding). Validate tarball URL is HTTPS before downloading.

- **Base URL candidates must be HTTPS**: All base URL candidates (from SDK constants, GitHub code frequency) must use HTTPS. Reject non-HTTPS candidates. The catalog validation already enforces HTTPS for `spec_url` — apply the same standard here.

- **`errgroup` for parallel execution (no `WithContext` cancellation)**: New `go.mod` dependency (`golang.org/x/sync`). Use `errgroup.Group` (not `errgroup.WithContext`) so that one source's failure does not cancel the other — this matches R4's requirement that each source is independent. Sources must never return errors to the errgroup; instead they log warnings to stderr and return empty `SourceResult`. The errgroup is for synchronization and panic recovery only. This is the first use of `context.Context` in the codebase (for timeouts); it is deliberate and should not be back-propagated to existing commands.

- **Duplicate GitHub request helper, don't import pipeline**: `newGitHubRequest()` from `research.go` is only 8 lines (set Accept header, add Bearer token from env). Duplicating it in `crowdsniff` avoids a coupling from a discovery package to the pipeline package.

- **Use plain string constants for source tiers**: `const TierOfficialSDK = "official-sdk"` etc., matching how `AuthConfig.Type` uses plain strings (`"api_key"`, `"bearer_token"`). No named type.

- **Postman source deferred**: Postman is not stubbed as an interface in v1 -- it's simply not implemented. When the Postman Explore CLI ships, adding a source is one new file implementing the same function signature.

## Open Questions

### Resolved During Planning

- **GitHub recency filtering**: Use separate `/repos/{owner}/{repo}` calls after code search to check `pushed_at`. Different rate limit pool (5000/hr vs 10/min), so not a bottleneck. Unique repo count is bounded by the 1000-result cap.
- **npm download counts**: Use the bulk downloads API (`api.npmjs.org/downloads/point/last-week/pkg1,pkg2,...`) to fetch weekly downloads for up to 128 packages in one call. Use 100/week as the threshold for "popular community SDK" vs low-confidence community package.
- **Wall-clock time**: ~2-4 minutes typical. npm completes in 10-30s (no rate limits). GitHub takes 1-4 min (10 req/min for code search + repo info checks). Sources run in parallel.
- **Output location**: Default to `~/.cache/printing-press/crowd-sniff/<name>-spec.yaml`, mirroring `websniff.DefaultCachePath()` pattern.

- **mergeSpecs preserves Meta via shallow copy**: `mergeSpecs()` copies `Resource` structs by value. The `Meta` map header is copied, preserving access to the same underlying data. This is safe because nothing mutates Meta after construction — Meta must be treated as immutable once set. No mergeSpecs changes needed. Verified by examining `root.go:314-357`.

### Deferred to Implementation

- Exact grep patterns for SDK source analysis -- prototype against Notion, Stripe, Discord SDKs during implementation
- Base URL extraction patterns from SDK constants -- validate during npm source implementation
- Parameter syntax normalization regex coverage -- prototype against real SDK path patterns (`:id`, `{user_id}`, `<userId>`, `$id`)

## High-Level Technical Design

> *This illustrates the intended approach and is directional guidance for review, not implementation specification. The implementing agent should treat it as context, not code to reproduce.*

```
                    crowd-sniff --api "notion"
                            |
                     +--------------+
                     | Resolve API  |  npm search + GitHub search
                     | Identity     |  -> disambiguate if needed
                     +--------------+
                            |
               +------------+------------+
               |                         |
        +------+------+          +------+------+
        | npm Source   |          | GitHub Source|
        | (goroutine)  |          | (goroutine) |
        +------+------+          +------+------+
        | 1. Search    |          | 1. Code     |
        | 2. Filter    |          |    search   |
        |    recency   |          | 2. Collect  |
        | 3. Downloads |          |    repos    |
        |    API       |          | 3. Check    |
        | 4. Tarball   |          |    pushed_at|
        |    + grep    |          | 4. Aggregate|
        +------+------+          +------+------+
               |                         |
               +------------+------------+
                            |
                     +--------------+
                     | Aggregate    |  Deduplicate endpoints,
                     | & Rank       |  compute source_tier,
                     |              |  source_count
                     +--------------+
                            |
                     +--------------+
                     | Build spec   |  Resolve base URL,
                     | APISpec      |  group into resources,
                     |              |  set Meta on endpoints
                     +--------------+
                            |
                     +--------------+
                     | Write YAML   |  websniff.WriteSpec()
                     +--------------+
```

## Implementation Units

- [ ] **Unit 1: Add Meta field to spec.Endpoint**

  **Goal:** Add `Meta map[string]string` with `yaml:"meta,omitempty"` to the `Endpoint` struct so crowd-sniff (and future features) can attach per-endpoint metadata.

  **Requirements:** R9

  **Dependencies:** None

  **Files:**
  - Modify: `internal/spec/spec.go`
  - Test: `internal/spec/spec_test.go`

  **Approach:**
  Add one field to the `Endpoint` struct, placed before `Alias` (keeping YAML-serialized fields grouped, non-YAML fields at bottom). The `omitempty` tag ensures zero impact on existing YAML round-trips. No changes needed to `Validate()`, generator templates, or `mergeSpecs()`.

  **Patterns to follow:**
  - Existing optional fields on `Endpoint`: `ResponsePath string yaml:"response_path,omitempty"`, `Pagination *Pagination yaml:"pagination"`

  **Test scenarios:**
  - Happy path: Parse spec YAML with `meta` field populated -> Endpoint.Meta contains expected key-value pairs
  - Happy path: Parse spec YAML without `meta` field -> Endpoint.Meta is nil (zero value)
  - Happy path: Marshal Endpoint with Meta set -> YAML output contains `meta:` section
  - Happy path: Marshal Endpoint with Meta nil -> YAML output omits `meta:` entirely
  - Integration: Meta survives a `mergeSpecs()` round-trip -- create two specs with Meta-bearing endpoints, merge them, verify Meta is preserved on all endpoints
  - Edge case: Existing `TestVersionConsistencyAcrossFiles` and other spec tests still pass unchanged

  **Verification:**
  - `go test ./internal/spec/...` passes
  - `go test ./...` passes (no regressions from field addition)

---

- [ ] **Unit 2: Create internal/crowdsniff package with types and aggregation**

  **Goal:** Define core types (`DiscoveredEndpoint`, `SourceResult`, source tier constants), the aggregation engine (deduplication, source_tier/source_count), parameter syntax normalization, and the spec builder (group into resources, resolve base URL, produce `*spec.APISpec` with Meta).

  **Requirements:** R4, R7, R8, R10, R14

  **Dependencies:** Unit 1

  **Files:**
  - Create: `internal/crowdsniff/types.go` (DiscoveredEndpoint, SourceResult, tier constants)
  - Create: `internal/crowdsniff/aggregate.go` (dedup, normalization, tier computation)
  - Create: `internal/crowdsniff/specgen.go` (APISpec construction from aggregated endpoints)
  - Test: `internal/crowdsniff/aggregate_test.go`
  - Test: `internal/crowdsniff/specgen_test.go`

  **Approach:**
  - `SourceResult` struct: `Endpoints []DiscoveredEndpoint` + `BaseURLCandidates []string` -- each source returns this so base URL signals are threaded through the interface
  - `DiscoveredEndpoint`: method, path, params (optional), source tier (string constant), source name (e.g., `@notionhq/client`)
  - Source tier constants: `const TierOfficialSDK = "official-sdk"`, `TierCommunitySDK = "community-sdk"`, `TierCodeSearch = "code-search"`, `TierPostman = "postman"`
  - `Aggregate(results []SourceResult) ([]AggregatedEndpoint, []string)`: deduplicate by normalized method+path, compute `source_tier` (highest tier) and `source_count` (distinct source count), collect base URL candidates
  - `BuildSpec(name, baseURL string, endpoints []AggregatedEndpoint) *spec.APISpec`: group into resources (reuse `deriveResourceKey` pattern from `websniff/specgen.go`), set `Meta` on each endpoint
  - **Path normalization (two-step)**: First, unify parameter syntax (`:id`, `{user_id}`, `<userId>`, `$id` all become `{id}`). Second, apply websniff-style normalization (replace concrete UUIDs, numeric IDs, hashes with placeholders). Copy the normalization functions from `websniff/classifier.go` rather than importing to keep packages independent.
  - Base URL resolution: accept `--base-url` flag, then SDK candidates, then GitHub frequency candidates, pick first non-empty

  **Patterns to follow:**
  - `internal/websniff/specgen.go` -- `AnalyzeCapture()` flow: classify -> deduplicate -> build endpoints -> assemble spec
  - `internal/websniff/classifier.go` -- `normalizeEntryPath()`, `deriveResourceKey()`, `deriveEndpointName()` (copy, don't import)
  - File naming: responsibility-based (`types.go`, `aggregate.go`, `specgen.go`) matching websniff convention

  **Test scenarios:**
  - Happy path: Aggregate endpoints from two sources -> source_count=2, source_tier=highest tier
  - Happy path: Aggregate with one official-sdk and one code-search source for same endpoint -> source_tier="official-sdk", source_count=2
  - Happy path: BuildSpec produces valid APISpec that passes `spec.Validate()`
  - Happy path: Meta on each endpoint contains `source_tier` and `source_count` as strings
  - Happy path: Base URL candidates from multiple sources -> first non-empty selected
  - Edge case: Single source provides all endpoints -> source_count=1 for all
  - Edge case: Same endpoint from same source twice -> deduplicated, source_count=1
  - Edge case: Empty endpoint list -> returns error (at least one endpoint required)
  - Edge case: Endpoints with different path normalizations that should merge (e.g., `/users/123` and `/users/456` both become `/users/{id}`)
  - Edge case: Parameter syntax normalization -- `/users/:id`, `/users/{user_id}`, `/users/<id>` all merge into `/users/{id}`
  - Happy path: Resource grouping from paths (e.g., `/v1/users` and `/v1/users/{id}` -> `users` resource)

  **Verification:**
  - `go test ./internal/crowdsniff/...` passes
  - Aggregated output matches expected source_tier/source_count for known inputs

---

- [ ] **Unit 3: npm source implementation**

  **Goal:** Implement the npm `Source` that searches the registry, filters by recency, fetches download counts, downloads/extracts tarballs, and greps for endpoint patterns.

  **Requirements:** R1, R5, R6, R17

  **Dependencies:** Unit 2

  **Files:**
  - Create: `internal/crowdsniff/npm.go`
  - Create: `internal/crowdsniff/patterns.go` (shared grep patterns, also used by Unit 4)
  - Test: `internal/crowdsniff/npm_test.go`
  - Test: `internal/crowdsniff/patterns_test.go`
  - Create: `testdata/crowdsniff/npm-search-response.json`
  - Create: `testdata/crowdsniff/npm-package-meta.json`
  - Create: `testdata/crowdsniff/sdk-sample/` (small extracted SDK file set for pattern testing)

  **Approach:**
  - **Constructor**: `NewNPMSource(opts NPMOptions)` where `NPMOptions` includes optional `BaseURL string` (defaults to `registry.npmjs.org`) and `HTTPClient *http.Client` (defaults to `&http.Client{Timeout: 15 * time.Second}`). Configurable base URL enables testing with `httptest.NewServer`. Use 30-second timeout for tarball downloads (larger payloads than JSON responses).
  - **Search**: `GET <baseURL>/-/v1/search?text=<api-name>&size=25` -> filter by `package.date` (6-month cutoff) -> take top 10
  - **Classify SDK type**: official if `package.scope` matches API vendor name (e.g., `@notionhq` for notion); else community
  - **Downloads**: Bulk call `api.npmjs.org/downloads/point/last-week/pkg1,pkg2,...` to get weekly counts. >100/week = popular community, else low-confidence community
  - **Tarball**: For each package, `GET <baseURL>/<pkg>/<version>` for `dist.tarball` URL. Download to `os.MkdirTemp("", "crowd-sniff-npm-*")`, extract with `archive/tar` + `compress/gzip`. Skip if tarball > 10MB (check Content-Length header first).
  - **Grep**: Apply patterns from `patterns.go` to all `.js`, `.ts`, `.mjs` files in extracted tree. Patterns match URL path literals, HTTP method calls, base URL constants. Sanitize all file paths (learnings: path traversal).
  - **Base URL extraction**: Also grep for base URL constants (`baseUrl`, `BASE_URL`, `this.baseUrl`, constructor defaults). Return as `SourceResult.BaseURLCandidates`.
  - **Cleanup**: `defer os.RemoveAll(tmpDir)` per package

  **Patterns to follow:**
  - `internal/cli/scorecard.go:32` -- `os.MkdirTemp` + `defer os.RemoveAll`
  - `internal/pipeline/research.go` -- HTTP client with explicit timeout, non-fatal error handling

  **Test scenarios:**
  - Happy path: Mock npm search response with 3 packages -> returns endpoints from all 3
  - Happy path: Official SDK (scoped `@notionhq/client`) -> source_tier="official-sdk"
  - Happy path: Popular community SDK (>100 downloads/week) -> source_tier="community-sdk"
  - Happy path: Grep finds `"/v1/users"` with `this.get` call -> DiscoveredEndpoint{Method:"GET", Path:"/v1/users"}
  - Edge case: Package `date` older than 6 months -> excluded from results
  - Edge case: Tarball > 10MB -> skipped with warning to stderr
  - Edge case: npm search returns 0 results -> returns empty slice, no error
  - Edge case: Download API unavailable -> still returns endpoints, just without download count classification
  - Error path: Tarball download fails -> skip package, continue with others
  - Error path: Tarball extraction encounters path traversal attempt -> sanitize and skip the malicious entry
  - Error path: Tarball contains symlink entry -> skipped (tar.TypeSymlink/TypeLink rejected)
  - Error path: Tarball URL is non-HTTPS -> skipped with warning
  - Edge case: Tarball has no Content-Length header -> io.LimitReader caps at 10MB

  **Verification:**
  - `go test ./internal/crowdsniff/...` passes
  - Pattern tests cover at least 3 SDK coding styles (class method, fetch wrapper, axios instance)

---

- [ ] **Unit 4: GitHub code search source implementation**

  **Goal:** Implement the GitHub `Source` that searches for API usage patterns in code, checks repo freshness, aggregates endpoint frequency, and extracts common parameters.

  **Requirements:** R2, R5, R6

  **Dependencies:** Unit 2

  **Files:**
  - Create: `internal/crowdsniff/github.go`
  - Test: `internal/crowdsniff/github_test.go`
  - Create: `testdata/crowdsniff/github-code-search-response.json`
  - Create: `testdata/crowdsniff/github-repo-response.json`

  **Approach:**
  - **Constructor**: `NewGitHubSource(opts GitHubOptions)` where `GitHubOptions` includes optional `BaseURL string` (defaults to `api.github.com`), `HTTPClient *http.Client`, and `Token string` (defaults to `os.Getenv("GITHUB_TOKEN")`). Configurable for testing.
  - **Code search**: When base URL is known (from `--base-url`), use domain-based queries: `"api.notion.com" language:javascript`. When no base URL is known, fall back to name-based queries: `"notion" api fetch language:javascript` (noisier but still useful). Use `Accept: application/vnd.github.text-match+json` for matched fragments. Paginate up to 1000 results (10 pages x 100).
  - **Rate limiting**: Use a plain `time.Ticker` at 6-second intervals for code search requests. No need for a reusable rate limiter type — only two call sites, both in this file. Testability is handled by the configurable base URL + `httptest.NewServer` approach (test server responds instantly).
  - **Repo freshness**: Collect unique `repository.full_name` from code search results. For each, `GET /repos/{owner}/{repo}` (5000 req/hr limit, separate pool) and check `pushed_at`. Discard repos not pushed within 6 months.
  - **Endpoint extraction**: From text matches and file URLs, extract URL path patterns. Use patterns from `patterns.go` (shared with npm source). Aggregate by normalized method+path, count distinct repos per endpoint. Return most frequent domain as `SourceResult.BaseURLCandidates`.
  - **Auth**: If token is empty, return empty `SourceResult` immediately (per R4 graceful degradation). Duplicate the 8-line `newGitHubRequest()` helper locally (set Accept header, add Bearer token) -- do not import `pipeline` package.

  **Patterns to follow:**
  - `internal/pipeline/research.go` -- `newGitHubRequest()` logic to duplicate, error handling patterns
  - `internal/websniff/classifier.go` -- `extractPath()`, `normalizeEntryPath()` patterns (copied to crowdsniff in Unit 2)

  **Test scenarios:**
  - Happy path: Mock code search response with 5 results across 3 repos -> returns aggregated endpoints with frequency counts
  - Happy path: Endpoint found in 10+ repos -> DiscoveredEndpoint with high frequency signal
  - Happy path: `text_matches` contain `/v1/users` and `/v1/projects` -> two distinct endpoints extracted
  - Edge case: No GITHUB_TOKEN set -> returns empty slice immediately, no error
  - Edge case: Repo pushed_at is 8 months ago -> all endpoints from that repo excluded
  - Edge case: All 1000 results are from the same repo -> source_count still 1 for code-search source
  - Edge case: Code search returns 0 results for an obscure API -> returns empty slice
  - Error path: Rate limit hit (429 response) -> wait and retry once, then skip remaining queries
  - Error path: GitHub API returns 5xx -> skip GitHub source, log warning to stderr
  - Integration: Code search and repo freshness checks use different rate limit pools -> both complete without mutual blocking

  **Verification:**
  - `go test ./internal/crowdsniff/...` passes
  - Rate limiter test confirms minimum 6-second spacing between code search requests

---

- [ ] **Unit 5: CLI command and orchestration**

  **Goal:** Add the `printing-press crowd-sniff` cobra command that accepts `--api`, orchestrates sources in parallel, handles disambiguation, and writes the output spec YAML.

  **Requirements:** R4, R10, R11, R12, R13, R14

  **Dependencies:** Units 2, 3, 4

  **Files:**
  - Create: `internal/cli/crowd_sniff.go`
  - Modify: `internal/cli/root.go` (register command)
  - Test: `internal/cli/crowd_sniff_test.go`

  **Approach:**
  - `newCrowdSniffCmd()` returns `*cobra.Command` with flags: `--api` (required), `--output`, `--base-url`, `--json`
  - No `--github-token` flag -- use `os.Getenv("GITHUB_TOKEN")` only, matching `research.go` convention
  - **Disambiguation**: If npm search + GitHub search both return results for multiple distinct APIs (e.g., `--api cal` matches `cal.com` and Google Calendar), present candidates via `fmt.Fprintf` to stderr and prompt user. If running non-interactively (check `os.Stdin` is not a terminal), error with "ambiguous API name, use --api <specific-name> or --base-url".
  - **Parallel execution**: Use `errgroup.Group` (not `WithContext`). Sources never return errors to the group — they log warnings to stderr and return empty `SourceResult`. This matches R4: each source is independent. Adds `golang.org/x/sync` to `go.mod`.
  - **Input sanitization**: URL-encode `--api` with `url.QueryEscape` before any HTTP request. Validate at CLI boundary: reject newlines, null bytes, path separators, `..`. Apply belt-and-suspenders to output cache path per learnings doc.
  - **Aggregation**: Call `crowdsniff.Aggregate(results)` which returns aggregated endpoints and merged base URL candidates. Then `crowdsniff.BuildSpec()`.
  - **Base URL**: First check `--base-url` flag. If not set, use candidates from `Aggregate()` (SDK constants first, then GitHub frequency). Fail with error asking user to provide `--base-url` if no URL resolved.
  - **Output**: Default cache path `~/.cache/printing-press/crowd-sniff/<name>-spec.yaml`. Call `websniff.WriteSpec()` from the CLI layer (crowdsniff package returns `*spec.APISpec` only, doesn't import websniff).
  - **Summary**: Print to stdout: endpoint count, resource count, source breakdown (N from npm, M from GitHub), source_tier distribution. `--json` flag outputs structured JSON instead.

  **Patterns to follow:**
  - `internal/cli/sniff.go` -- command structure, flag declaration, RunE flow, success message format
  - `internal/cli/root.go:48` -- command registration

  **Test scenarios:**
  - Happy path: `crowd-sniff --api notion` with mock sources -> produces valid spec YAML at expected output path
  - Happy path: `--output custom/path.yaml` -> writes to specified path
  - Happy path: `--base-url https://api.example.com` overrides auto-detected base URL
  - Happy path: `--json` flag produces structured JSON output
  - Happy path: Summary output includes correct endpoint and source counts
  - Edge case: Both sources return 0 endpoints -> error message "no endpoints discovered"
  - Edge case: Only npm source returns results (no GITHUB_TOKEN) -> still produces valid spec
  - Edge case: GitHub source returns error, npm returns 5 endpoints -> spec produced, warning logged to stderr about GitHub failure
  - Edge case: `--api` flag not provided -> cobra reports missing required flag
  - Error path: Output directory doesn't exist -> created automatically (same as sniff)
  - Error path: `--api "../../.ssh/evil"` -> rejected at CLI boundary (path traversal)
  - Error path: `--api "notion&size=9999"` -> api name URL-encoded in HTTP requests, no query injection

  **Verification:**
  - `go build -o ./printing-press ./cmd/printing-press` succeeds
  - `./printing-press crowd-sniff --help` shows expected flags and description
  - `go test ./internal/cli/...` passes

---

- [ ] **Unit 6: Skill integration (Phase 1.8 Crowd Sniff Gate)**

  **Goal:** Add Phase 1.8 to the printing-press skill that offers crowd-sniff when the research phase identifies spec gaps or no spec exists.

  **Requirements:** R15, R16

  **Dependencies:** Unit 5

  **Files:**
  - Modify: `skills/printing-press/SKILL.md`

  **Approach:**
  - Insert Phase 1.8 after Phase 1.7 (Sniff Gate). Same structure: decision matrix, user prompt via `AskUserQuestion`, fallback on failure.
  - Decision matrix mirrors sniff: offer crowd-sniff when spec has gaps or no spec found. Skip when spec appears complete or user already provided `--spec`.
  - Gate prompt: "Want me to search npm packages and GitHub code for `<api>` to discover additional endpoints? This typically takes 2-4 minutes."
  - On approval: run `printing-press crowd-sniff --api <api> --output "$RESEARCH_DIR/<api>-crowd-spec.yaml"`
  - On success: report endpoint count and feed into Phase 2 as additional `--spec`
  - On failure: "Crowd sniff found no additional endpoints -- proceeding with existing spec."
  - Time budget: 5 minutes (longer than sniff's 3 minutes due to GitHub rate limits)

  **Test expectation: none** -- skill SKILL.md changes are validated by manual testing and the existing quality gate workflow, not Go tests.

  **Verification:**
  - SKILL.md contains Phase 1.8 section with decision matrix, prompt, and fallback
  - Phase 1.8 is referenced in Phase 2 generate command construction (merge crowd-spec if it exists)

## System-Wide Impact

- **Spec struct change**: Adding `Meta map[string]string` to `spec.Endpoint` is the only change to shared types. With `omitempty`, it is invisible to all existing producers and consumers. No template changes, no validation changes, no mergeSpecs changes.
- **New external API dependencies**: npm registry (unauthenticated, generous limits) and GitHub code search (authenticated, 10 req/min). Both are optional per R4 -- the command works with either or both.
- **New go.mod dependency**: `golang.org/x/sync` for `errgroup`. Well-maintained, officially part of the Go project.
- **Binary size**: New `archive/tar` and `compress/gzip` imports add to binary. These are stdlib packages, minimal impact.
- **Unchanged invariants**: The `generate` command, OpenAPI/GraphQL parsers, existing `sniff` command, catalog system, and all templates are unmodified.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| GitHub code search 10 req/min makes the command slow | Run npm and GitHub in parallel. Set user expectation in CLI output ("Searching GitHub... this takes 1-3 minutes due to rate limits"). |
| Heuristic grep patterns produce false positives (match non-API URLs) | Filter by known API domain. Use path normalization to deduplicate. Cross-source agreement (source_count > 1) naturally filters noise. |
| npm tarballs contain malicious paths (zip slip) | Sanitize every extracted file path. Validate resolved path stays within temp dir root. Existing learnings doc covers this pattern. |
| GitHub code search results are stale (old repos) | Separate repo freshness check via `/repos/` API. 6-month hard cutoff on `pushed_at`. |
| Parameter syntax varies across SDKs (`:id`, `{id}`, `<id>`) | Two-step normalization: unify syntax first, then replace concrete values. Dedup depends on this working correctly. Prototype against 3+ real SDKs. |
| Malicious npm tarballs (zip-slip, symlinks) | Reject symlinks and hard links during extraction. Validate all paths stay within temp dir. Require HTTPS tarball URLs. Use io.LimitReader for size gate. |
| --api value used in HTTP URLs and file paths | URL-encode for HTTP. Validate at CLI boundary for path traversal. Belt-and-suspenders on output path. |
| Base URL candidates from untrusted sources | Require HTTPS. Reject localhost/private IPs. Validate before writing to spec YAML. |
| Postman source not available for v1 | Postman is simply not implemented in v1. When the Postman Explore CLI ships, adding it is one new file with the same function signature. |

## Sources & References

- **Origin document:** [docs/brainstorms/2026-03-29-crowd-sniff-requirements.md](docs/brainstorms/2026-03-29-crowd-sniff-requirements.md)
- npm registry API: `registry.npmjs.org/-/v1/search`, `api.npmjs.org/downloads/point/last-week/`
- GitHub code search API: `api.github.com/search/code`
- Related code: `internal/cli/sniff.go`, `internal/websniff/`, `internal/pipeline/research.go`
- Security learning: `docs/solutions/security-issues/filepath-join-traversal-with-user-input-2026-03-29.md`
