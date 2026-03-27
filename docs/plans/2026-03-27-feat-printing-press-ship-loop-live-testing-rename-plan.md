---
title: "feat: Ship Loop, Live API Testing, and Scoring Rename"
type: feat
status: active
date: 2026-03-27
---

# feat: Ship Loop, Live API Testing, and Scoring Rename

## Overview

Three improvements to the printing press pipeline based on real usage feedback from shipping Discord, Notion, and Linear CLIs:

1. **"Is this shippable?" triggers a fix loop** - not just thoughts. When the user asks if it's shippable and the answer is "no, these things are broken", the pipeline should make a plan and fix them, then re-score.
2. **API key prompt at the beginning, live testing at the end** - ask "do you want to give me an API key so I can test this aggressively at the end?" in Phase 0. If yes, add a Phase 5.5: Live API Testing with safety rails (read-only operations only, no posting/creating/deleting).
3. **Rename Steinberger scoring** - the two audit passes should have more descriptive names: "Non-Obvious Insight Review" for the first (gap-finding) pass, and a different name for the second (improvement-verification) pass.

## Problem Statement

### 1. Ship Check is a Dead End

Currently when the user asks "is this shippable?", the pipeline gives an honest assessment of problems but stops there. The user has to manually decide what to fix and ask for each fix. The pattern from last night:

```
User: "is it shippable?"
Claude: "No - sync doesn't paginate, FTS indexes raw JSON,
         dead flags, auth sends Bearer instead of Bot..."
User: "ok fix those"
Claude: [fixes some things]
User: "is it shippable now?"
Claude: "Better but still has ghost tables and..."
```

This should be a single loop: identify problems -> make a fix plan -> fix them -> re-score -> repeat until PASS or user says stop.

### 2. No Live API Testing

All 3 CLIs shipped with bugs that would have been caught by a single real API call:
- Discord CLI: Auth sends `Bearer` instead of `Bot` - every request 401s
- Notion CLI: sync.go calls `c.Post("/v1/search", nil, bodyBytes)` but the client's Post method signature is `Post(path, body)` - would crash at runtime
- Linear CLI: GraphQL field names validated against schema but actual API responses have nullable fields that cause nil pointer panics

The fix: ask for the API key upfront (Phase 0), store it safely, and add a Phase 5.5 that makes **read-only** API calls to validate the CLI actually works.

**Safety rails** (the user mentioned "it went rogue last time with an API key"):
- ONLY execute GET/read/list/search operations
- NEVER execute POST/PUT/PATCH/DELETE (create, update, delete)
- NEVER send messages, post comments, modify data
- Timeout every call at 10 seconds
- Stop on first auth failure (don't burn rate limits)
- Print every API call to stderr before executing it

### 3. "Steinberger" Naming is Opaque

"Steinberger Score" and "Steinberger Audit" are insider references to Peter Steinberger's gogcli being the quality bar. New users have no idea what this means. The two scoring passes need clearer names:

| Current Name | Proposed Name | Why |
|---|---|---|
| "First Steinberger Analysis" (Phase 3) | **"Non-Obvious Insight Review"** | This pass finds gaps that aren't visible from just reading the code - hallucinated paths, dead code, broken data pipelines |
| "Final Steinberger" (Phase 5) | **"Ship Readiness Assessment"** | This pass proves the fixes worked and gives a go/no-go verdict |

The underlying scoring system (0-100, Grade A-F) stays the same. Only the human-facing names change. The code struct can keep `SteinerScore` internally since it's not user-visible.

## Proposed Solution

### Change 1: Ship Loop in SKILL.md

Add to the end of Phase 5 in SKILL.md:

```markdown
## Phase 5.7: SHIP LOOP

After presenting the Final Report, if the verdict is not PASS (score < 65 or
critical issues remain):

1. Extract the top 3 highest-impact issues from the report
2. Write a targeted fix plan (not a full Phase 4 rerun - just the specific issues)
3. Apply fixes
4. Re-run Proof of Behavior verification (Phase 4.7)
5. Re-run scorecard
6. Present updated score with delta

If the user asks "is this shippable?" at any point:
1. Run the scorecard
2. Run Proof of Behavior verification
3. If PASS: "Yes, ship it. Score: X/100, 0 critical issues."
4. If WARN: "Shippable with caveats: [list]. Score: X/100."
5. If FAIL: "Not yet. Top issues: [list]. Want me to fix these and re-score?"
   - If user says yes: enter the fix loop above
   - If user says no: present the issues for manual review

Max 3 fix-loop iterations. After 3, report remaining issues and stop.
```

### Change 2: API Key Prompt + Live Testing Phase

**Phase 0 addition** (beginning of pipeline):

```markdown
## Phase 0.1: API KEY PROMPT

Before any research or generation, ask:

"Do you want to provide an API key so I can test the generated CLI
against the real API at the end? This is optional but catches auth
mismatches, wrong endpoint paths, and response parsing bugs.

Safety: I will ONLY run read-only operations (list, get, search).
I will NEVER create, update, delete, or post anything."

If yes: Store the key in a session variable (not written to disk).
If no: Skip Phase 5.5, rely on dry-run validation only.
```

**Phase 5.5: LIVE API TESTING** (after Proof of Behavior, before Final Report):

```markdown
## Phase 5.5: LIVE API TESTING (requires API key from Phase 0.1)

Skip this phase if no API key was provided.

### Safety Rules (NON-NEGOTIABLE)

1. ONLY execute these HTTP methods: GET
2. ONLY call these command types:
   - `doctor` (validates auth)
   - `<resource> list --limit 1` (validates path + response parsing)
   - `<resource> get <first-id-from-list>` (validates single-resource fetch)
   - `search "<common-term>" --limit 1` (validates search)
   - `sync --max-pages 5` (validates sync with tiny scope)
3. NEVER execute: create, update, delete, post, patch, put, send, execute
4. NEVER pass --stdin with body content
5. NEVER call webhook execute, message create, or any mutation endpoint
6. Timeout: 10 seconds per call, 2 minutes total for all testing
7. Stop immediately on 401/403 (don't burn rate limits on bad auth)
8. Print every command to stderr BEFORE executing

### Test Sequence

1. `<cli> doctor` - validates auth works
2. Pick 3 list endpoints, run each with `--limit 1 --json`
3. From list results, extract one ID
4. Run `<cli> <resource> get <id> --json`
5. If data layer exists: `<cli> sync --max-pages 5`
6. If search exists: `<cli> search "a" --limit 1`
7. Report: N/M calls succeeded, any parsing errors, auth status

### Output

```
LIVE API TEST RESULTS
=====================
Auth:     PASS (200 OK on doctor)
List:     3/3 passed (users, channels, guilds)
Get:      1/1 passed (user abc123)
Sync:     PASS (5 pages synced, 12 blocks)
Search:   PASS (3 results for "a")
Parsing:  0 errors

Verdict:  PASS - CLI works against real API
```
```

### Change 3: Rename Scoring Passes

**In SKILL.md**, rename:
- "PHASE 3: STEINBERGER AUDIT" -> "PHASE 3: NON-OBVIOUS INSIGHT REVIEW"
- "PHASE 5: FINAL STEINBERGER" -> "PHASE 5: SHIP READINESS ASSESSMENT"
- "Steinberger Score" in output -> "Quality Score"
- "Steinberger Grade" -> "Ship Grade"
- Keep the 0-100 scoring, A-F grading, same dimensions

**In scorecard.go**, rename:
- Printed header: "Steinberger Scorecard" -> "Quality Scorecard"
- `computeGrade` output labels: keep the same letter grades
- Struct name `SteinerScore` -> keep internal (not user-facing)

**In SKILL.md description line**:
- "5-phase loop with dual Steinberger analysis" -> "5-phase loop with Non-Obvious Insight Review and Ship Readiness Assessment"

**Do NOT rename** in plan artifact filenames - those are already written and referenced.

## Acceptance Criteria

### Ship Loop
- [ ] SKILL.md has Phase 5.7 Ship Loop section
- [ ] "Is this shippable?" triggers scorecard + verification + fix-plan-if-needed
- [ ] Max 3 fix-loop iterations with circuit breaker
- [ ] Fix loop only addresses top 3 highest-impact issues per iteration

### Live API Testing
- [ ] Phase 0.1 asks for API key (optional)
- [ ] Phase 5.5 runs read-only tests against real API
- [ ] Safety rules block ALL mutation operations (POST/PUT/PATCH/DELETE)
- [ ] `--limit 1` on all list calls (minimize API usage)
- [ ] `--max-pages 5` on sync (tiny scope)
- [ ] 10-second timeout per call
- [ ] Stops on auth failure (401/403)
- [ ] Prints every command before executing
- [ ] Produces structured pass/fail report

### Rename
- [ ] SKILL.md Phase 3 renamed to "Non-Obvious Insight Review"
- [ ] SKILL.md Phase 5 renamed to "Ship Readiness Assessment"
- [ ] scorecard.go printed header says "Quality Scorecard" not "Steinberger Scorecard"
- [ ] SKILL.md description updated
- [ ] No changes to internal Go struct names or plan artifact filenames

## Files to Change

| File | Change |
|------|--------|
| `skills/printing-press/SKILL.md` | Add Phase 0.1 (API key prompt), Phase 5.5 (live testing), Phase 5.7 (ship loop), rename Phase 3 and 5 headings |
| `internal/pipeline/scorecard.go` | Change printed header from "Steinberger Scorecard" to "Quality Scorecard" |
| `internal/pipeline/fullrun.go` | Add live testing step between verification and scorecard (conditional on API key) |

## Dependencies & Risks

- **Risk**: Live API testing could accidentally hit a mutation endpoint if the safety check regex fails. Mitigation: whitelist approach (only allow explicit safe patterns), not blacklist.
- **Risk**: API keys could leak into logs or plan artifacts. Mitigation: never write the key to disk, mask in stderr output, clear from memory after Phase 5.5.
- **Risk**: The ship loop could run forever if issues are systemic (e.g., wrong spec). Mitigation: max 3 iterations, each targeting only top 3 issues.
- **Dependency**: Proof-of-Behavior verification (already shipped in `efaec84b`) is used by the ship loop for re-verification after fixes.

## Sources & References

### Internal
- SKILL.md: `skills/printing-press/SKILL.md` - main pipeline definition
- Scorecard: `internal/pipeline/scorecard.go` - scoring implementation
- Verification: `internal/pipeline/verify.go` - Proof-of-Behavior (just shipped)
- Fullrun: `internal/pipeline/fullrun.go` - pipeline orchestration
- Discord CLI auth bug: Bearer vs Bot mismatch found only after "is it shippable?" question
- Notion CLI runtime crash: Post method signature mismatch found only when user asked to test
- User feedback: "it went rogue last time I gave you an API key" - safety rails are critical
