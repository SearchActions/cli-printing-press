---
name: printing-press
description: Generate production Go CLIs from any API. Uses ce:plan for deep research before generating, then ce:plan again to audit the output, then ce:work to fix it. Plan-execute-plan-execute loop.
version: 0.6.0
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - WebFetch
  - WebSearch
  - AskUserQuestion
  - Skill
  - Agent
  - Edit
  - CronCreate
  - CronList
  - CronDelete
---

# /printing-press

Generate a production Go CLI from any API. Claude Code is the brain. The Go binary is the template engine.

**The loop:** ce:plan (research) -> ce:work (generate) -> ce:plan (audit) -> ce:work (fix) -> report

## Quick Start

```
/printing-press Notion
/printing-press Plaid payments API
/printing-press --spec ./openapi.yaml
```

## Prerequisites

- Go 1.21+ installed
- The printing-press repo at `~/cli-printing-press`
- Build: `cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press`

## Workflow 0: Natural Language (Primary)

When the user provides an API name:

### Phase 1: RESEARCH (ce:plan does this)

Write a research plan file, then let ce:plan expand it with deep research.

**Step 1: Write the research plan seed**

```bash
mkdir -p ~/cli-printing-press/docs/plans
```

Write to `docs/plans/<date>-feat-<api>-cli-research-plan.md`:
```markdown
---
title: "Research for <API> CLI generation"
type: feat
status: active
date: <today>
---

# Research for <API> CLI

## What to Find
1. OpenAPI spec (search GitHub, apis-guru, common URLs)
2. Competing CLIs (GitHub search: "<api> cli", note stars, language, commands)
3. API documentation URL
4. Auth method
5. Community demand signals (Reddit, HN mentions of needing a CLI)

## What to Produce
- Spec URL (or determination that we need to write one from docs)
- List of competitors with command counts
- Recommendation: use OpenAPI spec vs write from docs
```

**Step 2: Run ce:plan to expand the research**

```
Skill("compound-engineering:ce:plan", "docs/plans/<date>-feat-<api>-cli-research-plan.md")
```

ce:plan will WebSearch, WebFetch competitor READMEs, analyze the landscape, and produce a comprehensive research plan with findings.

**Step 3: Execute the research (manual - read ce:plan output)**

Read the expanded plan. Extract:
- The spec URL (OpenAPI or docs)
- Competitor list with command counts
- Auth method and base URL

### Phase 2: GENERATE (ce:work does this)

**Step 4: Generate the CLI**

If OpenAPI spec was found:
```bash
cd ~/cli-printing-press && ./printing-press generate \
  --spec "<spec-url>" \
  --output ./<api>-cli \
  --force --lenient
```

If no spec (docs only): Claude Code reads the docs (WebFetch), writes a YAML spec, and generates from it. See "Writing Specs from Docs" section below.

If quality gates fail: read error, fix spec, retry (max 3).

### Phase 3: AUDIT (ce:plan does this)

After the CLI is generated, write an audit plan and run ce:plan to expand it into a full code review.

**Step 5: Write the audit plan seed**

Write to `docs/plans/<date>-fix-<api>-cli-audit-plan.md`:
```markdown
---
title: "Audit <API> CLI against competitors and quality bar"
type: fix
status: active
date: <today>
---

# Audit <API> CLI

## Generated CLI Location
<output-dir>

## Competitors Found
<list from Phase 1 research>

## Audit Checklist
1. Command count: do we match or beat competitors?
2. Help descriptions: are they developer-friendly or spec jargon?
3. Examples: are they realistic or placeholder values?
4. README: does it sell the tool or just describe it?
5. Agent-native: --json, --select, --dry-run, --stdin, --yes all present?
6. Missing endpoints: any important API operations we missed?
7. Auth: does doctor validate credentials correctly?

## What to Produce
- List of specific fixes needed (file path + what to change)
- Honest assessment: is this CLI better than the competitors?
```

**Step 6: Run ce:plan to do the audit**

```
Skill("compound-engineering:ce:plan", "docs/plans/<date>-fix-<api>-cli-audit-plan.md")
```

ce:plan will read the generated code, compare against competitors, check every item on the audit checklist, and write specific fix instructions.

### Phase 4: FIX (ce:work does this)

**Step 7: Execute the fixes**

```
Skill("compound-engineering:ce:work", "docs/plans/<date>-fix-<api>-cli-audit-plan.md")
```

ce:work will:
- Edit help descriptions to be developer-friendly
- Add realistic examples
- Rewrite README to sell the tool
- Add any missing endpoints noted in the audit
- Verify fixes compile: `cd <output> && go build ./... && go vet ./...`

### Phase 5: SCORE + REPORT

**Step 8: Run the scorecard**

```bash
cd ~/cli-printing-press && SCORECARD_CLI_DIR=./<api>-cli SCORECARD_PIPELINE_DIR=/tmp/<api>-score \
  go test ./internal/pipeline/ -run TestScorecardOnRealCLI -v 2>&1 | tail -20
```

**Step 9: Present the final result**

Show ALL of these:
1. Resources and commands (table)
2. Steinberger score and grade
3. Competitor comparison:
   - "Found N competing CLIs"
   - "Best competitor: X (Y stars, Z commands)"
   - "We beat them on: ..."
   - "We're missing: ..."
4. Example commands with realistic values
5. How to install
6. Spec source
7. Any limitations (skipped complex body fields, etc.)

## Workflow 1: From Spec File

`/printing-press --spec <path>`

Skip Phase 1 research (spec is provided). Run Phases 2-5.

## Workflow 2: From URL

`/printing-press --spec <url>`

Skip Phase 1 research. Run Phases 2-5.

## Workflow 3: Submit to Catalog

`/printing-press submit <name>`

1. Gather metadata
2. Write `catalog/<name>.yaml`
3. `git checkout -b catalog/<name> && git add && git commit && gh pr create`

## Workflow 4: Autonomous Pipeline

`/printing-press print <api-name>`

Full 8-phase pipeline with session chaining:
preflight -> research -> scaffold -> enrich -> regenerate -> review -> comparative -> ship

Each phase: read seed plan -> ce:plan expands it -> ce:work executes -> chain to next session.

Budget gate: 3 hours max. Morning report on completion.

See state.json and pipeline directory for phase tracking.

## Writing Specs from Docs

When no OpenAPI spec exists and Claude Code needs to write one:

1. WebFetch the API documentation URL
2. Read `~/cli-printing-press/skills/printing-press/references/spec-format.md`
3. Claude Code reads the docs and identifies EVERY endpoint:
   - Method, path, description, params, body fields, auth
4. Write YAML spec to `/tmp/<api>-spec.yaml`
5. Generate from it

This is where Claude Code IS the brain. No regex. No shelling out.

## Key Principle

**Plan-execute-plan-execute.** Never just generate and present. Always:
1. Plan (research the API and competitors)
2. Execute (generate the CLI)
3. Plan again (audit the output, find problems)
4. Execute again (fix the problems)
5. Score and report

This is what makes the press smart, not just fast.

## Safety Gates

- Preview before generating
- Output directory conflict: ask before overwrite
- Untrusted specs: note if not in known-specs registry
- Max 3 retries on quality gate failure

## Limitations

- Go CLIs only
- OpenAPI 3.0+ and Swagger 2.0
- 50 resources / 50 endpoints per resource limit
- No GraphQL (but can wrap GraphQL in YAML spec, like Linear)
- ce:plan and ce:work require Compound Engineering plugin
