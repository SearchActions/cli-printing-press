---
title: "Pipeline Nightnight-Style Chaining - Autonomous Phase Execution"
type: feat
status: active
date: 2026-03-24
---

# Pipeline Nightnight-Style Chaining

## Overview

Update the printing-press SKILL.md with Workflow 4 (autonomous pipeline) and Workflow 5 (resume). When the user says `/printing-press print gmail`, the skill initializes the pipeline, then chains through 6 phases - each one running ce:plan to expand the seed into a full plan, then ce:work to execute it, then CronCreate to chain to the next phase with fresh context.

Modeled directly on osc-nightnight's proven patterns: checkpoint-based resumption, heartbeat safety net, mechanical budget gate, JSONL event logging.

## Acceptance Criteria

- [ ] SKILL.md has Workflow 4 (autonomous pipeline) with full phase loop
- [ ] SKILL.md has Workflow 5 (resume pipeline)
- [ ] Skill allowed-tools includes Skill, Agent, CronCreate, CronList, CronDelete, Edit
- [ ] Heartbeat CronCreate scheduled at pipeline start (45 min safety net)
- [ ] Phase chain CronCreate scheduled after each phase completes (30 sec)
- [ ] Budget gate checks elapsed time between phases (3h max)
- [ ] Resume mode loads state.json and runs budget gate FIRST
- [ ] Morning report written on completion or budget exhaustion
- [ ] Error handling: retry once per phase, skip after 2 failures, never die silently
- [ ] Version bumped to 0.4.0

## Implementation Units

### Unit 1: Update SKILL.md Frontmatter

**Files:** `skills/printing-press/SKILL.md`

**Approach:**
- Bump version from 0.3.0 to 0.4.0
- Add to allowed-tools: `Skill`, `Agent`, `Edit`, `CronCreate`, `CronList`, `CronDelete`
- Update description to mention autonomous pipeline

### Unit 2: Write Workflow 4 - Autonomous Pipeline

**Files:** `skills/printing-press/SKILL.md`

**Approach:** Insert after Workflow 3 (Submit to Catalog), before Safety Gates section.

Workflow 4 structure:

```
When user says "print <api-name>":

Step 1: Initialize
  - Build press binary: go build -o ./printing-press ./cmd/printing-press
  - Run: ./printing-press print <api-name> [--output] [--force]
  - This creates docs/plans/<api-name>-pipeline/ with 6 plan seeds + state.json

Step 2: Heartbeat safety net
  - CronCreate 45 min from now: "/printing-press print <api-name> --resume"
  - Uses: date -v+45M '+%M %H %d %m' for cron time

Step 3: Phase execution loop
  For each phase in state.json where status != "completed":
    a. Read the plan seed at state.phases[phase].plan_path
    b. Run Skill("compound-engineering:ce:plan", plan_path)
       - ce:plan expands the seed with parallel research agents
    c. Run Skill("compound-engineering:ce:work", plan_path)
       - ce:work implements, tests, checks off criteria
    d. Update state.json: mark phase "completed"
    e. Run budget gate (Step 4)
    f. If more phases remain:
       - CronCreate 30 sec from now: "/printing-press print <api-name> --resume"
       - Print: "[phase] Complete. Chaining to [next] in 30s..."
       - END SESSION (fresh context for next phase)

Step 4: Budget gate (between every phase)
  python3 -c "
  import json, datetime
  s = json.load(open('docs/plans/<api-name>-pipeline/state.json'))
  started = datetime.datetime.fromisoformat(s['started_at'])
  elapsed = (datetime.datetime.now() - started).total_seconds() / 3600
  if elapsed > 3:
      print('STOP')
  else:
      print('CONTINUE')
  "
  If STOP: go to Step 5 (morning report), do NOT chain.

Step 5: Morning report
  Write docs/plans/<api-name>-pipeline/report.md:
  - API name, spec source, spec URL
  - Phases completed vs total
  - Resources and endpoint count (from scaffold output)
  - Enrichments applied (from overlay.yaml if exists)
  - Quality score (from review phase if reached)
  - Total elapsed time
  - Next steps for the human
```

**Patterns to follow:** osc-nightnight SKILL.md - CronCreate syntax, budget gate python3 command, heartbeat scheduling, resume logic

### Unit 3: Write Workflow 5 - Resume Pipeline

**Files:** `skills/printing-press/SKILL.md`

**Approach:**

```
When user says "resume <api-name>" or "--resume" flag:

1. Load docs/plans/<api-name>-pipeline/state.json
2. MANDATORY: Run budget gate FIRST (before any work)
3. If STOP: write morning report if not already written, exit
4. If CONTINUE:
   - Show status: which phases done, which is next
   - Delete heartbeat cron if stale
   - Schedule new heartbeat (45 min)
   - Go to Workflow 4 Step 3 (phase loop)
```

### Unit 4: Error Handling Section

**Files:** `skills/printing-press/SKILL.md`

**Approach:** Add error handling rules to Workflow 4:

- If ce:plan fails on a phase: log error to state.json errors array, retry once. If still fails, mark phase "failed", skip to next phase.
- If ce:work fails: retry once with same plan. If still fails, mark "failed", skip to next.
- If 2+ consecutive phases fail: write morning report and STOP. Do not chain.
- NEVER die silently. Always update state.json before ending a session.
- If CronCreate fails: print manual resume command as fallback.

### Unit 5: Update Limitations Section

**Files:** `skills/printing-press/SKILL.md`

**Approach:** Update limitations to reflect current capabilities:
- Remove "OAuth2 flows are simplified to bearer_token" (now fully supported)
- Remove "Generated CLIs do not include retry/rate-limiting" (now included)
- Update endpoint cap from 20 to 50
- Add: "Pipeline mode requires Compound Engineering plugin"
- Add: "Pipeline budget gate is 3 hours max"

**Verification:** Read the final SKILL.md, verify all 5 workflows are present, frontmatter is correct, and the phase loop matches nightnight's pattern.

## Scope Boundaries

- Don't modify Go code - this plan is SKILL.md only
- Don't implement the actual phase logic in Go - the skill orchestrates via Bash/Skill calls
- Don't add new plan seed templates - those already exist in seeds.go

## Sources

- osc-nightnight: `~/.claude/skills/osc-nightnight/SKILL.md` - CronCreate syntax, budget gate, heartbeat, resume logic
- Current skill: `skills/printing-press/SKILL.md` - existing 4 workflows to extend
- Pipeline state: `internal/pipeline/state.go` - PipelineState, PhaseOrder, status constants
- Pipeline init: `internal/pipeline/pipeline.go` - Init() creates dir + seeds
