# Dogfood Gauntlet Findings - 2026-03-24

## Scorecard: 10/10 Pass All 7 Gates (was 4/10 before Round 2 fixes)

| # | API | Paths | Result | Notes |
|---|-----|-------|--------|-------|
| 1 | Fly.io | 51 | PASS | 7/7 gates. Complex body fields skipped (config, metadata). |
| 2 | Spotify | 68 | PASS | 7/7 gates. CLI name: spotify-web-sonallux-cli. Array body fields skipped (ids, uris). |
| 3 | Telegram | 74 | PASS | 7/7 gates. 22 endpoints skipped (resource limit 50). Flat structure. |
| 4 | Vercel | 85 | PASS | 7/7 gates. Global param `teamId` filtered (93/113 endpoints). oneOf bodies skipped. |
| 5 | Supabase | 105 | PASS | 7/7 gates. No servers in spec - fallback placeholder URL used. |
| 6 | Sentry | 126 | PASS | 7/7 gates. Flat "api" resource. Endpoints truncated at limit 50. |
| 7 | LaunchDarkly | 221 | PASS | 7/7 gates. Flat "api" resource. Endpoints truncated at limit 50. |
| 8 | Trello | 264 | PASS | 7/7 gates. Global params `key`/`token` filtered (324/324 endpoints). |
| 9 | Jira | 317 | PASS | 7/7 gates. Flat "rest" resource. Endpoints truncated at limit 50. |
| 10 | Cloudflare | 1716 | PASS | 7/7 gates. Largest spec tested. Complex body fields skipped. |

## Grade: A (10/10 pass, up from 4/10)

All 6 bugs from Round 1 findings are fixed. The generator handles diverse real-world specs.

## Bugs Fixed (Round 2 - commit 1eda222)

### Bug 1: `$ref`/`$schema` keys leak into Go identifiers - FIXED
- **Was:** Fly.io, Vercel, Jira failed go vet
- **Fix:** `toCamel` rewritten to split on all non-letter/non-digit chars. `flagName` rewritten with full sanitization.

### Bug 2: Spec title too long for CLI name - FIXED
- **Was:** Spotify panic from 60+ char CLI name
- **Fix:** `cleanSpecName` noise word filtering already handled this (prior commit). Name now derives cleanly.

### Bug 3: Missing `servers` field causes parse failure - FIXED
- **Was:** Supabase failed with "base_url is required"
- **Fix:** Parser falls back to placeholder URL when no servers defined (prior commit).

### Bug 4: Enum values with `/` generate invalid Go type names - FIXED
- **Was:** Trello failed go build
- **Fix:** `toCamel` now strips `/` and all non-alphanumeric chars. `flagName` does the same.

### Bug 5: `defaultVal` type mismatch - FIXED (new bug found in Round 2)
- **Was:** Supabase had string "true" used as bool flag default
- **Fix:** `defaultVal` now coerces defaults by declared param type instead of Go type-switching on the raw value.

### Bug 6: Duplicate body/query param flag names - FIXED (new bug found in Round 2)
- **Was:** Some specs had body fields that collided with query/path params
- **Fix:** Parser deduplicates body params against query/path params by kebab-case flag name. Also deduplicates within `mapRequestBody` and `mapTypes` by camelCase name.

### Bug 7: Cloudflare parse failure - FIXED
- **Was:** "name is required" because spec had no info.title
- **Fix:** Parser fallback to filename-derived name (prior commit).

## Remaining Observations (Not Bugs)

### Flat Resource Problem
- Sentry, LaunchDarkly, Jira all produce a single flat resource ("api", "rest") because all paths share a common prefix (`/api/v2/`, `/api/0/`, `/rest/api/3/`).
- Enhancement opportunity: strip common API path prefixes before grouping into resources.

### Endpoint Truncation
- Telegram (22 skipped), Sentry, LaunchDarkly, Jira all hit the 50-endpoint-per-resource limit.
- Consider bumping to 100 for flat-structure APIs, or implementing smarter sub-resource detection.

### Global Param Filtering
- Vercel's `teamId` and Trello's `key`/`token` correctly auto-detected and filtered as global params.
- This is working well - the heuristic (present on >80% of endpoints) catches real global params.

## Fix History

| Round | Commit | Pass Rate | Key Fixes |
|-------|--------|-----------|-----------|
| Initial | - | 3/10 | Baseline |
| Round 1 | 3f096bc | 4/10 | Parser type name sanitization, unlocked Jira |
| Round 1.5 | d8a93e0 | 4/10 | Template sanitization, toCamel/flagName/types hardening |
| Round 2 | 651d11e | 4/10 | Doctor health endpoint checking |
| Round 2.5 | 1eda222 | 10/10 | toCamel/flagName full rewrite, defaultVal coercion, body dedup |
