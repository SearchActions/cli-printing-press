---
name: printing-press
description: Generate production Go CLIs from any API. Claude Code is the brain - it researches, writes specs, generates, polishes, and scores. Say an API name and get a complete CLI.
version: 0.5.0
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

Generate a production Go CLI from any API. Claude Code does the thinking. The Go binary does the template rendering.

## Architecture

```
Claude Code (brain)  ->  printing-press binary (template engine)  ->  Claude Code (polish)
  researches API           renders Go templates                       improves help text
  writes YAML spec         runs quality gates                         rewrites README
  analyzes competitors     deterministic, fast                        scores output
```

## Quick Start

```
/printing-press Notion
/printing-press Plaid payments API
/printing-press --spec ./openapi.yaml
/printing-press --docs https://developers.notion.com/reference --name notion
```

## Prerequisites

- Go 1.21+ installed
- The printing-press repo at `~/cli-printing-press`
- printing-press binary built: `cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press`

## Workflows

### Workflow 0: Natural Language (Primary)

When the user provides an API name or description:

**Step 1: Parse intent and find the spec**

Extract the API name. Then search for an OpenAPI spec:

1. Check KnownSpecs in `~/cli-printing-press/internal/pipeline/discover.go`
2. Check catalog: `ls ~/cli-printing-press/catalog/<name>.yaml`
3. WebSearch: `"<api-name>" openapi spec site:github.com`
4. Try common URLs: `https://raw.githubusercontent.com/<org>/openapi/main/openapi.yaml`

If OpenAPI spec found: go to Step 3 (generate from spec).
If no spec found: go to Step 2 (research and write spec).

**Step 2: Research the API and write a spec (Claude Code does this)**

This is where Claude Code IS the brain. No regex. No shelling out to another LLM.

a. **Fetch the API docs:**
```
WebFetch the API documentation URL (e.g., developers.notion.com/reference)
```

b. **Research competitors:**
```
WebSearch: "<api-name> cli" site:github.com
```
For each competitor found, note: name, stars, language, last updated.

c. **Read the docs and identify EVERY endpoint:**
Read the fetched docs content. List every API endpoint you find:
- HTTP method (GET, POST, PUT, PATCH, DELETE)
- Path (/v1/databases, /v1/pages/{id})
- Description
- Parameters (path params, query params, body fields)
- Auth method (Bearer, API key, OAuth)
- Base URL

d. **Write the YAML spec:**
Read the spec format reference: `cat ~/cli-printing-press/skills/printing-press/references/spec-format.md`

Write a complete YAML spec to `/tmp/<name>-spec.yaml` with ALL endpoints found. Include:
- name (kebab-case)
- description
- base_url
- auth (type, header, env_vars)
- resources (group endpoints by resource)
- endpoints (method, path, params, body)

e. **Generate:**
```bash
cd ~/cli-printing-press && ./printing-press generate \
  --spec /tmp/<name>-spec.yaml \
  --output ./<name>-cli \
  --force
```

f. **If quality gates fail:** Read the error. Fix the YAML spec. Regenerate. Max 3 retries.

g. Go to Step 4 (polish).

**Step 3: Generate from OpenAPI spec**

```bash
cd ~/cli-printing-press && ./printing-press generate \
  --spec "<spec-url>" \
  --output ./<name>-cli \
  --force --lenient
```

Use `--lenient` for specs with broken $refs (PagerDuty, Intercom).

If all 7 quality gates pass: go to Step 4 (polish).
If gates fail: read error, fix, retry (max 3).

**Step 4: Polish (Claude Code does this)**

Read the generated code and improve it directly:

a. **Improve help descriptions:**
Read each command file in `<output>/internal/cli/*.go`. Find `Short:` strings. If any are jargon-heavy spec descriptions, use Edit to rewrite them to be developer-friendly (under 80 chars, starts with a verb).

b. **Improve examples:**
Read each command's `Example:` string. If it uses generic values like "value" or "<id>", use Edit to replace with realistic values (e.g., "usr_abc123", "2026-01-01", "user@example.com").

c. **Improve README:**
Read `<output>/README.md`. If the description is generic spec text, use Edit to rewrite:
- Add a one-line hook that makes developers want to install it
- Add "Why This Exists" if there's no official CLI for this API
- Ensure Quick Start has real 3-command workflow

d. **Note: this step is optional.** If the user doesn't want polish, skip it. If the generated output is already good, skip it. Use judgment.

**Step 5: Score (optional)**

Run the Steinberger scorecard:
```bash
cd ~/cli-printing-press && SCORECARD_CLI_DIR=./<name>-cli SCORECARD_PIPELINE_DIR=/tmp/<name>-pipeline \
  go test ./internal/pipeline/ -run TestScorecardOnRealCLI -v 2>&1 | tail -20
```

Report the score to the user.

**Step 6: Present result**

Show:
1. What was generated (directory, resources, commands)
2. Example commands to try
3. How to install: `cd <name>-cli && go install ./cmd/<name>-cli`
4. Steinberger score if computed
5. Competitors found (if any)
6. Note if spec was auto-written vs from OpenAPI

### Workflow 1: From Spec File

When `--spec <local-path>`:
1. Verify file exists
2. `cd ~/cli-printing-press && ./printing-press generate --spec <path> [--output <dir>] --lenient`
3. Optional: polish (Step 4 above)
4. Present result

### Workflow 2: From Docs URL

When `--docs <url>`:
1. WebFetch the docs URL
2. Claude Code reads the docs and writes a YAML spec (Step 2 above)
3. Generate from the spec
4. Polish
5. Present result

### Workflow 3: Submit to Catalog

When `submit <name>`:
1. Gather metadata (ask user for display name, description, category, homepage)
2. Write `~/cli-printing-press/catalog/<name>.yaml`
3. `git checkout -b catalog/<name> && git add catalog/<name>.yaml && git commit && git push && gh pr create`
4. Present PR URL

### Workflow 4: Autonomous Pipeline

When `print <api-name>`:

Uses the multi-phase pipeline with ce:plan -> ce:work loops:

| Phase | What Claude Code Does |
|-------|----------------------|
| Preflight | Verify Go, download spec, cache conventions |
| Research | WebSearch competitors, WebFetch their READMEs, analyze |
| Scaffold | Write spec (if needed), run `printing-press generate` |
| Enrich | Read generated output, identify missing endpoints, improve spec |
| Regenerate | Re-run generator with enriched spec |
| Review | Run scorecard, dogfood Tier 1, polish output |
| Comparative | Score vs competitors |
| Ship | Git init, write report |

Each phase: read the plan seed -> expand with ce:plan -> execute with ce:work -> write next phase's plan -> chain to next session.

**Step 1: Initialize**
```bash
cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press
./printing-press print <api-name> [--output <dir>] [--force]
```

**Step 2: Phase execution loop**
For each phase where `plan_status` is not "completed":
a. Read plan file
b. If `status: seed`: run `Skill("compound-engineering:ce:plan", plan_path)` to expand
c. If `status: active`: run `Skill("compound-engineering:ce:work", plan_path)` to execute
d. Update state.json
e. Chain to next session via CronCreate (30s)

Budget gate: stop after 3 hours. Write morning report.

### Workflow 5: Resume Pipeline

When `resume <api-name>`:
1. Load state.json
2. Budget gate first
3. Continue from next incomplete phase

### Workflow 6: Scorecard

When `score <dir>`:
```bash
cd ~/cli-printing-press && SCORECARD_CLI_DIR=<dir> SCORECARD_PIPELINE_DIR=/tmp/score-pipeline \
  go test ./internal/pipeline/ -run TestScorecardOnRealCLI -v
```

### Workflow 7: Full Test Run

When `test` or `fullrun`:
```bash
cd ~/cli-printing-press && FULL_RUN=1 go test ./internal/pipeline/ -run TestFullRun -v -timeout 10m
```

## Key Principle

**Claude Code IS the LLM brain.** The printing-press Go binary is a template engine. When the skill says "read the API docs and write a spec," that means YOU (Claude Code) use WebFetch to read the docs and Write to create the spec file. You don't shell out to another LLM. You ARE the LLM.

The Go binary's `internal/llm/` and `internal/llmpolish/` packages exist as a fallback for when someone runs `printing-press generate --polish` from their terminal without Claude Code. But inside this skill, YOU do the thinking.

## Safety Gates

- Preview before generating: show API name, base URL, estimated resources
- Output directory conflict: ask before overwriting
- Untrusted specs: note if spec is from a URL not in known-specs registry

## Limitations

- Only generates Go CLIs
- OpenAPI 3.0+ and Swagger 2.0 supported
- Large APIs truncated to 50 resources / 50 endpoints per resource
- No GraphQL support (but Claude Code can write a YAML spec that wraps GraphQL, as we did for Linear)
