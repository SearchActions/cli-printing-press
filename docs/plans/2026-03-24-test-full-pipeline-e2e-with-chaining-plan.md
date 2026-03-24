---
title: "E2E Test - Full Pipeline with Nightnight Chaining on Petstore"
type: test
status: active
date: 2026-03-24
---

# E2E Test - Full Pipeline with Nightnight Chaining

## Overview

We built the pipeline infrastructure tonight:
- `printing-press print` creates 6 plan seeds + state.json (verified in E2E test)
- SKILL.md Workflow 4 documents the autonomous phase loop with CronCreate chaining
- Review phase seed now includes 3-tier dogfooding
- Budget gate, heartbeat, error handling, morning report all documented

But we've never actually run the full loop. This plan tests the complete pipeline end-to-end: `printing-press print petstore` -> ce:plan phase 0 -> ce:work phase 0 -> chain to phase 1 -> ... -> morning report.

Petstore is the simplest spec (3 resources, 13 endpoints, no OAuth2) - ideal for validating the pipeline machinery without API complexity.

## Acceptance Criteria

- [ ] `printing-press print petstore` creates pipeline directory with 6 seeds + state.json
- [ ] Phase 0 (Preflight): ce:plan expands seed, ce:work executes, state.json updated to "completed"
- [ ] Phase 1 (Scaffold): ce:plan expands seed, ce:work generates CLI, all 7 gates pass
- [ ] Phase 2 (Enrich): ce:plan researches, ce:work writes overlay.yaml (or skips if no enrichments)
- [ ] Phase 3 (Regenerate): ce:plan plans merge, ce:work regenerates with overlay
- [ ] Phase 4 (Review): ce:plan plans review, ce:work runs static checks + dogfood tier 1
- [ ] Phase 5 (Ship): ce:plan plans ship, ce:work creates git repo + morning report
- [ ] Budget gate runs between each phase and says CONTINUE (petstore is fast)
- [ ] CronCreate chains work (or manual chaining if cron not available in test)
- [ ] Morning report written to `docs/plans/petstore-pipeline/report.md`
- [ ] state.json shows all 6 phases as "completed"
- [ ] Generated petstore-cli compiles, runs, doctor works

## Implementation Units

### Unit 1: Initialize Pipeline

```bash
cd ~/cli-printing-press
go build -o ./printing-press ./cmd/printing-press
./printing-press print petstore --output /tmp/petstore-pipeline-test --force
```

Verify: 6 plan seeds + state.json created.

### Unit 2: Execute Phase 0 (Preflight) Manually

Since we can't use CronCreate chaining in a test, execute each phase manually following Workflow 4:

a. Read the plan seed: `docs/plans/petstore-pipeline/00-preflight-plan.md`
b. Run: `Skill("compound-engineering:ce:plan", "docs/plans/petstore-pipeline/00-preflight-plan.md")`
c. Run: `Skill("compound-engineering:ce:work", "docs/plans/petstore-pipeline/00-preflight-plan.md")`
d. Update state.json: mark preflight as "completed"
e. Run budget gate - verify CONTINUE

### Unit 3: Execute Phase 1 (Scaffold)

Same pattern:
a. Read `01-scaffold-plan.md`
b. ce:plan to expand
c. ce:work to execute (should run `printing-press generate`)
d. Verify: CLI generated at output dir, 7 gates pass
e. Update state.json, run budget gate

### Unit 4: Execute Phase 2 (Enrich)

a. Read `02-enrich-plan.md`
b. ce:plan to expand (may use WebSearch for docs)
c. ce:work to execute (should produce overlay.yaml or skip)
d. Update state.json, run budget gate

### Unit 5: Execute Phase 3 (Regenerate)

a. Read `03-regenerate-plan.md`
b. ce:plan to expand
c. ce:work to execute (merge overlay if exists, regenerate)
d. Verify: CLI still compiles, gates still pass
e. Update state.json, run budget gate

### Unit 6: Execute Phase 4 (Review + Dogfood)

a. Read `04-review-plan.md`
b. ce:plan to expand
c. ce:work to execute:
   - Static checks (help, names, descriptions)
   - Dogfood Tier 1 (version, doctor, dry-run)
   - Write dogfood-results.json and review.md with combined score
d. Verify: review.md exists with score, dogfood-results.json has test results
e. Update state.json, run budget gate

### Unit 7: Execute Phase 5 (Ship)

a. Read `05-ship-plan.md`
b. ce:plan to expand
c. ce:work to execute:
   - Git init in output dir
   - Write morning report
d. Verify: report.md exists, state.json shows all 6 phases completed

### Unit 8: Validate Final State

```bash
# Check state.json
python3 -c "
import json
s = json.load(open('docs/plans/petstore-pipeline/state.json'))
for phase in ['preflight','scaffold','enrich','regenerate','review','ship']:
    status = s['phases'][phase]['status']
    print(f'{phase}: {status}')
    assert status == 'completed', f'{phase} not completed!'
print('ALL PHASES COMPLETED')
"

# Check morning report
cat docs/plans/petstore-pipeline/report.md

# Check review score
cat docs/plans/petstore-pipeline/review.md

# Check generated CLI
cd /tmp/petstore-pipeline-test
./petstore-cli --help
./petstore-cli doctor
./petstore-cli version
```

### Unit 9: Cleanup

Remove test artifacts:
- `docs/plans/petstore-pipeline/`
- `/tmp/petstore-pipeline-test/`

## Scope Boundaries

- Use manual phase execution (no CronCreate chaining) - this tests the pipeline logic, not the scheduling
- Don't fix bugs found in the pipeline seeds - document them for follow-up
- Don't test with complex APIs (Gmail, Stripe) - Petstore is sufficient for infrastructure validation
- If a phase fails, document the failure and continue to the next phase manually
- This is a test plan, not an implementation plan - don't write new code

## Dependencies

- Nightnight chaining plan (completed) - provides Workflow 4/5 in SKILL.md
- Autonomous dogfood phase (completed) - provides enhanced Review seed
- Compound Engineering plugin - required for ce:plan and ce:work
- Parser fixes (completed) - Petstore already passes, so this is stable

## Expected Outcome

If everything works: 6/6 phases complete, morning report shows petstore-cli with quality score, generated CLI runs. This validates the entire pipeline vision: one command creates a plan-per-phase pipeline that can run autonomously.

If something fails: we document exactly which phase broke and why, creating targeted fix plans for the next session.

## Sources

- Pipeline init: `internal/pipeline/pipeline.go`
- Plan seeds: `internal/pipeline/seeds.go`
- SKILL.md Workflow 4: `skills/printing-press/SKILL.md` (lines 209-307)
- State machine: `internal/pipeline/state.go`
- Review seed with dogfood: commit aae7801
