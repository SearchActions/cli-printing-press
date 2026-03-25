---
title: "Overnight Press Hardening - Gauntlet, Tests, First Real CLI"
type: feat
status: completed
date: 2026-03-24
---

# Overnight Press Hardening

## Overview

Autonomous overnight task to harden the Printing Press. Three goals: (1) re-run the 10-API gauntlet with the new templates and fix any regressions, (2) add tests for all the new Phase 1-3 code that shipped today, (3) generate the first real CLI (PagerDuty) from a verified OpenAPI spec and push it to a new repo.

This plan is designed to run unattended via `/ce:work`.

## Why These Three Things

Today we shipped:
- Phase 0: `--select` flag, error hints, README rewrite, `{{.Owner}}` variable, generation comments
- Phase 1: Research phase with GitHub API discovery
- Phase 2: Dogfood automation with 3-tier testing
- Phase 3: Comparative analysis with 6-dimension scoring

None of this has been validated end-to-end. The gauntlet hasn't run with the new templates. The new Go packages have zero tests. And we haven't generated a single CLI that we'd actually ship publicly.

## Implementation Units

### Unit 1: Re-run 10-API Gauntlet with New Templates

**Goal:** Verify the template changes (--select, error hints, README, Owner, generation comments) don't break any of the 10 APIs that previously passed.

**Files:** No file changes expected - this is validation only. If failures occur, fix the templates.

**Approach:**
1. Build the printing-press binary: `go build -o ./printing-press ./cmd/printing-press`
2. For each of the 10 gauntlet APIs, run:
   ```
   rm -rf /tmp/dogfood-<name>-cli
   ./printing-press generate --spec "<url>" --output /tmp/dogfood-<name>-cli --force
   ```
3. All 10 must pass 7/7 quality gates
4. If any fail, fix the template that caused the regression and re-run
5. Update `docs/plans/dogfood-gauntlet-findings.md` with the results

**Spec URLs (from gauntlet findings doc):**
- Fly.io: `https://docs.machines.dev/spec/openapi3.json`
- Spotify: `https://api.apis.guru/v2/specs/spotify.com/sonallux/2023.2.27/openapi.json`
- Telegram: `https://api.apis.guru/v2/specs/telegram.org/5.0.0/openapi.json`
- Vercel: `https://api.apis.guru/v2/specs/vercel.com/0.0.1/openapi.json`
- Supabase: `https://api.supabase.com/api/v1-json`
- Sentry: `https://raw.githubusercontent.com/getsentry/sentry-api-schema/main/openapi-derefed.json`
- LaunchDarkly: `https://app.launchdarkly.com/api/v2/openapi.json`
- Trello: `https://api.apis.guru/v2/specs/trello.com/1.0/openapi.json`
- Jira: `https://api.apis.guru/v2/specs/atlassian.com/jira/1001.0.0-SNAPSHOT/openapi.json`
- Cloudflare: `https://raw.githubusercontent.com/cloudflare/api-schemas/main/openapi.json`

**Verification:** All 10 show `PASS` on all 7 gates. Commit the findings update.

### Unit 2: Add Tests for New Packages

**Goal:** Add tests for the code shipped today that has zero test coverage.

**Files to create:**
- `internal/pipeline/research_test.go`
- `internal/pipeline/dogfood_test.go`
- `internal/pipeline/comparative_test.go`
- `internal/generator/textfilter_test.go`
- `internal/generator/readme_augment_test.go`

**Approach per file:**

**research_test.go:**
- Test `scoreNovelty` with 0 alternatives (should return 10), 1 alternative with 10K stars (should return 2), 1 alternative with 50 stars (should return 7)
- Test `deduplicateAlts` removes URL duplicates
- Test `recommend` maps scores to correct strings
- Test `analyzeAlternatives` returns non-empty gaps and patterns
- Test `writeResearchJSON` / `LoadResearch` round-trip

**dogfood_test.go:**
- Test `buildTier1Commands` returns help, version, doctor + per-resource help
- Test `hasCredentials` with empty env vars (false) and set env vars (true)
- Test `computeDogfoodScore` with all pass (40+), half pass (20+), zero pass (0)
- Test `writeDogfoodResults` / `LoadDogfoodResults` round-trip

**comparative_test.go:**
- Test `scoreAlternative` with various install methods and star counts
- Test `compareGapsAndAdvantages` returns our known advantages
- Test `RunComparative` with empty research.json (no alternatives scenario)

**textfilter_test.go:**
- Test `CheckText` catches "comprehensive", "robust", "leverage"
- Test `CheckText` does NOT flag normal technical text ("the API returns JSON")
- Test `FormatWarnings` with 0 warnings returns empty string
- Test `FormatWarnings` with 3 warnings returns formatted output

**readme_augment_test.go:**
- Test `AugmentREADME` with marker comments (replaces them)
- Test `AugmentREADME` without markers (appends section)
- Test `AugmentREADME` with no evidence files (no-op)

**Verification:** `go test ./...` passes with the new test files.

### Unit 3: Generate First Real CLI (PagerDuty)

**Goal:** Generate PagerDuty CLI from the verified OpenAPI spec as the first "real" CLI we'd ship publicly.

**Why PagerDuty (not Notion):** PagerDuty has a verified OpenAPI spec that auto-generates. Notion needs a hand-written YAML spec. PagerDuty is the fastest path to a real CLI we can validate end-to-end. It's also #2 on our launch list (44/50 virality score).

**Files:**
- `pagerduty-cli/` - generated output (in the printing-press repo for now, separate repo later)
- `catalog/pagerduty.yaml` - new catalog entry

**Approach:**
1. Generate: `./printing-press generate --spec "https://raw.githubusercontent.com/PagerDuty/api-schema/main/reference/REST/openapiv3.json" --output ./pagerduty-cli --force`
2. Verify 7/7 quality gates pass
3. Build and run `pagerduty-cli --help` to capture real output
4. Run `pagerduty-cli doctor` to capture health check output
5. Check the generated README has the new format (quickstart, output formats, agent usage, troubleshooting)
6. Verify `--select` flag is present: `pagerduty-cli --help | grep select`
7. Verify generation comments: `grep "Code generated" pagerduty-cli/internal/cli/*.go | wc -l`
8. Add `catalog/pagerduty.yaml` with the spec URL and known_alternatives
9. Add PagerDuty to KnownSpecs in `internal/pipeline/discover.go` (already there, verify)
10. Commit the generated CLI and catalog entry

**Catalog entry:**
```yaml
name: pagerduty
display_name: PagerDuty
description: Incident management, on-call scheduling, and alerting
category: developer-tools
spec_url: https://raw.githubusercontent.com/PagerDuty/api-schema/main/reference/REST/openapiv3.json
spec_format: json
openapi_version: "3.0"
tier: community
verified_date: "2026-03-24"
homepage: https://developer.pagerduty.com
notes: "On-call engineers' best friend. Community CLI (95 stars) abandoned Oct 2024."
known_alternatives:
  - name: pagerduty-cli
    url: https://github.com/martindstone/pagerduty-cli
    language: typescript
```

**Verification:** `pagerduty-cli --help` shows commands, `pagerduty-cli version` works, `pagerduty-cli doctor` runs, README has quickstart section.

### Unit 4: Generate Second Real CLI (Plaid)

**Goal:** Generate Plaid CLI as the second real CLI. This validates multi-spec generation and a different auth pattern.

**Files:**
- `plaid-cli/` - generated output
- `catalog/plaid.yaml` - new catalog entry

**Approach:**
1. Generate: `./printing-press generate --spec "https://raw.githubusercontent.com/plaid/plaid-openapi/master/2020-09-14.yml" --output ./plaid-cli --force`
2. Verify 7/7 quality gates pass
3. Same validation as PagerDuty (help, doctor, README format, --select, generation comments)
4. Add `catalog/plaid.yaml`
5. Commit

**Verification:** Same as Unit 3.

### Unit 5: Generate Third Real CLI (Intercom)

**Goal:** Generate Intercom CLI. This validates yet another OpenAPI spec format.

**Files:**
- `intercom-cli/` - generated output
- `catalog/intercom.yaml` - new catalog entry

**Approach:**
1. Generate: `./printing-press generate --spec "https://raw.githubusercontent.com/intercom/Intercom-OpenAPI/main/descriptions/2.11/api.intercom.io.yaml" --output ./intercom-cli --force`
2. Same validation pipeline
3. Add `catalog/intercom.yaml`
4. Commit

### Unit 6: Final Commit and Summary

**Goal:** Push everything and leave a summary of what was accomplished.

**Approach:**
1. Run `go test ./...` one final time
2. Run `go build -o ./printing-press ./cmd/printing-press` one final time
3. Push to origin: `git push origin main`
4. Write a summary to `docs/plans/overnight-hardening-results.md` with:
   - Gauntlet results (10/10 or N/10)
   - Test count added
   - CLIs generated (PagerDuty, Plaid, Intercom)
   - Any issues found and fixed

## Acceptance Criteria

- [ ] 10-API gauntlet passes 10/10 with new templates
- [ ] 5 new test files with meaningful coverage for research, dogfood, comparative, textfilter, readme_augment
- [ ] `go test ./...` passes
- [ ] PagerDuty CLI generated, 7/7 gates, catalog entry added
- [ ] Plaid CLI generated, 7/7 gates, catalog entry added
- [ ] Intercom CLI generated, 7/7 gates, catalog entry added
- [ ] All generated CLIs have: --select flag, error hints, new README format, generation comments
- [ ] Pushed to origin

## Scope Boundaries (Do NOT Do)

- Do NOT create separate GitHub repos for the CLIs (that's Phase 4)
- Do NOT set up Homebrew tap (that's Phase 4)
- Do NOT hand-write custom YAML specs (Notion, Airtable, Mixpanel are for later)
- Do NOT modify the Linear CLI (it was hand-built, not auto-generated)
- Do NOT add new template features beyond what shipped today
