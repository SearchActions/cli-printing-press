---
title: "fix: Scorecard is trivially easy - make it measure real quality"
type: fix
status: active
date: 2026-03-25
---

# Scorecard is Trivially Easy - Make It Measure Real Quality

## Problem

Every generated CLI scores 90/100 (or 90/90 before Vision). This is absurd. The scorecard checks for the *presence* of patterns that the generator *always* includes. It's grading its own homework.

### Why Every CLI Gets 10/10 on Everything

| Dimension | What It Checks | Why It's Always 10/10 |
|---|---|---|
| Output Modes | Does root.go contain "json", "plain", "select"? | Template always includes these flags |
| Auth | Does config.go have `os.Getenv`? | Template always uses env vars |
| Error Handling | Does helpers.go have "hint:" and "code:" strings? | Template always includes hints + codes |
| Terminal UX | Does helpers.go have "colorEnabled", "NO_COLOR", "isatty"? | Template always includes these |
| README | Does README.md have "Quick Start", "Agent Usage", etc.? | Template always generates these sections |
| Doctor | Does doctor.go have HTTP health check patterns? | Template always includes health checks |
| Agent Native | Does root.go have "json", "select", "dry-run", "stdin", "yes"? | Template always includes all flags |
| Local Cache | Does client.go have "cacheDir", "readCache", "no-cache"? | Template always includes caching |
| Breadth | How many .go files in internal/cli/? | Any real API generates 60+ files |

**Result:** The scorecard measures "did the template render correctly?" not "is this CLI any good?"

### What a Real Scorecard Should Measure

A 10/10 should be **hard to achieve**. It should require:
- Quality, not just presence
- Correct behavior, not just correct strings
- Real developer experience polish, not template boilerplate
- Comparison to the best tools in the category, not absolute presence

---

## Proposed Solution: Quality-Based Scoring

Replace presence checks with quality checks. Each dimension gets 3 tiers:

| Points | Tier | Meaning |
|---|---|---|
| 0-3 | **Present** | The feature exists (what current scoring gives 10 for) |
| 4-6 | **Functional** | The feature works correctly and handles edge cases |
| 7-10 | **Excellent** | The feature matches or exceeds best-in-class tools |

### Dimension-by-Dimension Redesign

#### Output Modes (0-10)

**Current (always 10):** Checks if "json", "plain", "select", "table", "csv" strings exist in root.go.

**Proposed:**
| Points | Criteria | How to Check |
|---|---|---|
| +2 | --json flag exists | `strings.Contains(root.go, `"json"`)` |
| +1 | --plain flag exists | `strings.Contains(root.go, "plain")` |
| +1 | --select flag exists | `strings.Contains(root.go, "select")` |
| +1 | --csv flag exists | `strings.Contains(root.go, "csv")` |
| +1 | --quiet flag exists | `strings.Contains(root.go, "quiet")` |
| +2 | Output formatting is field-aware (select actually filters fields, not just string matching) | Check helpers.go for `filterFields` function with real JSON parsing logic (json.Unmarshal + field iteration, not just string ops) |
| +2 | Pagination outputs NDJSON progress events | Check helpers.go for "event" + "page_fetch" (progress event pattern) |

**Expected baseline:** 5-6/10 (template gets presence, quality takes work)

#### Auth (0-10)

**Current (always 10):** Counts `os.Getenv` in config.go + checks auth.go exists.

**Proposed:**
| Points | Criteria | How to Check |
|---|---|---|
| +2 | At least one env var for auth | `os.Getenv` count >= 1 in config.go |
| +1 | Auth file exists | auth.go present |
| +2 | Token stored securely (config file has restricted permissions) | Check config.go for `0o600` or `0600` permission bits |
| +2 | Token masking in output | Check client.go or helpers.go for token masking pattern (e.g., showing only last 4 chars) |
| +1 | Multiple auth methods (env var + config file + flag) | Count distinct auth source patterns >= 2 |
| +2 | OAuth2 browser flow with refresh | Check auth.go for "oauth2" + "refresh" + "browser" patterns |

**Expected baseline:** 3-5/10 (simple API key gets 3, full OAuth2 gets 8+)

#### Error Handling (0-10)

**Current (always 10):** Checks for "hint:" and counts "code:" occurrences.

**Proposed:**
| Points | Criteria | How to Check |
|---|---|---|
| +2 | At least 3 distinct exit codes | Count unique `code:` values >= 3 |
| +1 | Error messages include hints | "hint:" or "Hint:" present |
| +2 | Rate limit handling (429 detection + retry) | Check client.go for "429" AND ("Retry-After" or "backoff") |
| +2 | Idempotency handling (409 = success) | Check helpers.go for "409" AND "already exists" with exit 0 |
| +1 | Not-found returns specific exit code | Check for "404" AND separate exit code |
| +2 | Error messages include actionable suggestions (not just "failed") | Check for "Run" + "doctor" or "try" in error messages (actionable patterns, not just status codes) |

**Expected baseline:** 4-6/10

#### Terminal UX (0-10)

**Current (always 10):** Checks for "colorEnabled", "NO_COLOR", "isatty".

**Proposed:**
| Points | Criteria | How to Check |
|---|---|---|
| +2 | NO_COLOR support | "NO_COLOR" in helpers.go |
| +1 | TTY detection | "isatty" in helpers.go |
| +1 | Color toggle flag | "no-color" in root.go |
| +2 | Table output uses aligned columns (tabwriter or similar) | Check helpers.go for "tabwriter" |
| +2 | Help text descriptions are meaningful (not raw spec jargon) | Sample 5 random command files, check Short descriptions are > 10 chars and don't just repeat the command name |
| +2 | Example values are realistic (not "abc123" or "value") | Sample 5 random command files, check Example lines don't contain "abc123" or bare "value" |

**Expected baseline:** 4-6/10 (template gets basics, description quality requires polish)

#### README (0-10)

**Current (always 10):** Checks for section header strings.

**Proposed:**
| Points | Criteria | How to Check |
|---|---|---|
| +1 | Has "Quick Start" section | String check |
| +1 | Has "Agent Usage" section | String check |
| +1 | Has "Doctor" section | String check |
| +1 | Has "Troubleshooting" section | String check |
| +2 | Quick Start has realistic example (not placeholder values) | Check Quick Start section doesn't contain "abc123", "your-key-here", "USER/tap" |
| +2 | Has Cookbook or Recipes section with 3+ examples | Check for "Cookbook" or "Recipes" AND count code blocks >= 3 |
| +2 | Commands section lists actual command count matching generated files | Compare README command count to `ls internal/cli/*.go | wc -l` - within 20% |

**Expected baseline:** 4-6/10

#### Doctor (0-10)

**Current (always 10):** Counts HTTP patterns.

**Proposed:**
| Points | Criteria | How to Check |
|---|---|---|
| +2 | Doctor command exists | doctor.go present |
| +2 | Checks auth validity | "auth" or "token" in doctor.go check logic |
| +2 | Checks API connectivity | HTTP request pattern in doctor.go |
| +2 | Checks config file health | "config" check in doctor.go |
| +2 | Checks version compatibility | "version" in doctor.go check logic OR HEAD request to API endpoint |

**Expected baseline:** 4-6/10

#### Agent Native (0-10)

**Current (always 10):** Checks for flag strings.

**Proposed:**
| Points | Criteria | How to Check |
|---|---|---|
| +1 | --json flag | Present |
| +1 | --select flag | Present |
| +1 | --dry-run flag | Present |
| +1 | --stdin flag | Present |
| +1 | --yes flag | Present |
| +1 | Non-interactive (no prompts in command logic) | Absence of "Scan" or "ReadLine" or "Prompt" in command files |
| +2 | Typed exit codes (at least 5 distinct) | Count unique exit codes >= 5 |
| +2 | --stdin examples in at least 3 commands | Count command files containing "--stdin" in Example >= 3 |

**Expected baseline:** 5-7/10 (flags are easy, stdin examples require work)

#### Local Cache (0-10)

**Current (always 10 for templates, 7 without sqlite):** Checks for cache patterns.

**Proposed:**
| Points | Criteria | How to Check |
|---|---|---|
| +2 | GET response caching | "readCache" or "writeCache" in client.go |
| +1 | --no-cache bypass flag | "no-cache" or "NoCache" |
| +2 | Cache has TTL (not infinite) | Time duration or expiry check near cache logic |
| +2 | Cache directory is configurable or uses XDG | Check for ".cache" or "XDG_CACHE_HOME" |
| +3 | SQLite or embedded DB backend (not just file cache) | "sqlite" or "bolt" or "badger" in any Go file |

**Expected baseline:** 5-7/10 (file cache = 5, sqlite = 10)

#### Breadth (0-10)

**Current scoring is reasonable.** Keep it but add a quality gate:

**Proposed addition:**
| Points | Criteria | How to Check |
|---|---|---|
| (existing) | Command file count tiers | Same as current |
| -2 penalty | More than 50% of commands have identical 1-word Short descriptions (e.g., "Get", "Create", "Delete") | Sample Short: fields, count ones that are <= 1 word |

**Expected baseline:** 7-9/10 (still easy for large APIs, penalty catches lazy descriptions)

#### Vision (0-10)

**Already redesigned in previous commit.** Keep as-is - it's the only dimension that was already hard (pure API wrappers correctly score 0).

---

## Implementation

### File: `internal/pipeline/scorecard.go`

Rewrite each `score*` function with the new criteria. Keep the same function signatures so no callers change.

### Key implementation details:

1. **Sampling for quality checks:** For dimensions that check "quality" across many files (Terminal UX description check, Agent Native stdin examples), sample 5 random command files rather than checking all. Use deterministic seed (hash of API name) for reproducibility.

2. **The "abc123" detector:** Create a helper `hasPlaceholderValues(content string) bool` that checks for common placeholder patterns: "abc123", "my-resource", "your-key-here", "USER/tap", bare "value" as positional arg.

3. **Description quality check:** `isQualityDescription(desc string) bool` returns true if description is > 10 chars, doesn't just repeat the command verb, and contains at least one space (multi-word).

### Expected Score Distribution After Fix

| CLI Quality | Old Score | New Score |
|---|---|---|
| Fresh from generator (no polish) | 90-100/100 | 45-55/100 (Grade C) |
| Generator + LLM polish pass | 90-100/100 | 60-70/100 (Grade B) |
| Generator + polish + manual GOAT fixes | 90-100/100 | 75-85/100 (Grade A) |
| Handcrafted gogcli-level quality | N/A | 90-100/100 |
| discrawl-level (visionary + quality) | N/A | 85-95/100 |

**Key insight:** Grade A should require WORK. Getting there should mean the CLI is genuinely good, not that the template rendered.

---

## Acceptance Criteria

- [ ] Fresh-from-generator Petstore CLI scores 40-55/100 (was 90+)
- [ ] Discord CLI (with our GOAT fixes) scores 55-70/100 (was 90+)
- [ ] No dimension auto-scores 10/10 from template output alone
- [ ] At least 3 dimensions require quality checks that sampling validates
- [ ] Placeholder value detection catches "abc123", "my-resource", "your-key-here"
- [ ] Description quality check distinguishes "Get" from "Get a channel by ID"
- [ ] `go build ./...` and `go test ./...` pass
- [ ] Scorecard output format unchanged (same table, same grades)

---

## Sources

- Current scorecard: `internal/pipeline/scorecard.go`
- Current scorecard CLI: `internal/cli/scorecard.go`
- Generator templates: `internal/generator/templates/*.tmpl`
- Previous Vision plan: `docs/plans/2026-03-25-feat-visionary-research-phase-plan.md`
