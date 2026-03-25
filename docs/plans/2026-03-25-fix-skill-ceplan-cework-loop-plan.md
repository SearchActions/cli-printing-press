---
title: "Fix Skill: ce:plan -> ce:work Loop on Every Run"
type: fix
status: active
date: 2026-03-25
---

# Fix: ce:plan -> ce:work Loop on Every Run

## What Got Dumber

The old Workflow 4 (autonomous pipeline) used ce:plan to do deep research before each phase and ce:work to execute. But Workflow 0 (the main path) never used ce:plan at all. It just searched for a spec and generated.

The user's vision: EVERY run should plan-then-execute, not just the autonomous pipeline.

## The Flow Should Be

```
/printing-press Notion

Phase 1: RESEARCH (ce:plan writes this)
  Claude Code runs ce:plan which:
  - WebSearches for OpenAPI specs
  - WebSearches for competing CLIs
  - WebFetches competitor READMEs
  - Analyzes what commands competitors have
  - Writes a research plan file

Phase 2: GENERATE (ce:work executes the plan)
  Claude Code runs ce:work which:
  - Downloads the spec (or writes one from docs)
  - Runs printing-press generate
  - Captures the output

Phase 3: AUDIT (ce:plan writes this)
  Claude Code runs ce:plan which:
  - Reads the generated CLI code
  - Code reviews it against agent-native checklist
  - Compares command count vs competitors
  - Identifies missing endpoints, bad descriptions, weak examples
  - Writes an audit plan with specific fixes

Phase 4: FIX (ce:work executes the audit plan)
  Claude Code runs ce:work which:
  - Edits help descriptions
  - Adds missing examples
  - Rewrites README
  - Verifies fixes compile
  - Runs scorecard

Phase 5: REPORT
  - Presents the final CLI with scores, competitor comparison, and grade
```

## What to Change in SKILL.md

Replace the current Workflow 0 Steps 1-6 with this plan-execute-plan-execute loop. The skill calls ce:plan and ce:work as sub-skills at each phase.

## Learnings from the Notion Run

The Notion run that just happened:
- Found an official OpenAPI spec from makenotion/notion-mcp-server (good!)
- Generated 7 resources, ~20 endpoints
- But MISSED the Databases resource (query endpoint was there via data-sources but the main databases CRUD wasn't prominent)
- Did NOT research competitors (4ier/notion-cli at 87 stars has 15+ commands)
- Did NOT polish any help text
- Did NOT score
- Did NOT compare against competitors
- Took 2 minutes but produced a "dumb" result

With the plan-execute loop:
- Phase 1 (ce:plan) would have found the competitors and noted their commands
- Phase 2 (ce:work) would have generated with the spec
- Phase 3 (ce:plan) would have audited and found: "competitors have databases.query but we're missing it as a proper resource"
- Phase 4 (ce:work) would have fixed it
- Result: smarter CLI that beats competitors, not just compiles
