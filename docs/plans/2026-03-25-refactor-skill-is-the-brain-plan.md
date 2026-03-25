---
title: "The Skill IS the Brain - Claude Code Drives Everything"
type: refactor
status: active
date: 2026-03-25
---

# The Skill IS the Brain

## The Mistake

I put the LLM intelligence inside the Go binary (`internal/llm/`, `internal/llmpolish/`). The Go binary shells out to `claude` CLI to get LLM answers. This is insane because:

1. Claude Code can't call itself (nesting)
2. The skill runs INSIDE Claude Code - Claude Code IS the LLM
3. Shelling out to an LLM from a Go binary is solving a problem that doesn't exist

## The Correct Architecture

```
User: /printing-press Notion

Claude Code (the brain):
  1. WebFetch Notion API docs
  2. Read the docs, understand every endpoint
  3. Write a YAML spec file (Claude Code writes this, not regex)
  4. Run: printing-press generate --spec <spec-I-wrote>
  5. Read the generated code
  6. Edit the help text, examples, README (Claude Code edits files directly)
  7. Run: go test ./... to verify
  8. Run scorecard
  9. Report results

Go binary (the template engine):
  - Takes a YAML/OpenAPI spec
  - Renders Go templates
  - Runs quality gates
  - That's it. No intelligence.
```

Claude Code is already running. It already has WebFetch, Read, Write, Edit, Bash, Agent tools. It doesn't need a Go package to shell out to itself.

## What Changes

### The Skill Does the Thinking

The SKILL.md file (`skills/printing-press/SKILL.md`) becomes the orchestrator. It tells Claude Code exactly what to do at each step. The skill IS the product.

**Before (broken):** Go binary tries to be smart by calling `claude` CLI
**After (correct):** Skill tells Claude Code to be smart, Go binary just renders templates

### The Go Binary Stays Dumb

The Go binary (`printing-press generate`) keeps doing what it does well:
- Parse OpenAPI specs
- Parse internal YAML specs
- Render 14 Go templates
- Run 7 quality gates
- That's it

No `internal/llm/`. No `internal/llmpolish/`. No `GenerateFromDocsLLM`. The Go binary is a tool that Claude Code uses, not a brain.

### The Workflow

When someone says `/printing-press Notion`:

**Step 1: Research (Claude Code does this)**
- WebFetch the Notion API docs page
- Search GitHub for competing CLIs (GitHub API via WebFetch or Bash curl)
- Read competitor READMEs
- Understand the API - every endpoint, auth method, base URL
- Write a ce:plan file: `docs/plans/<api>-research-plan.md`

**Step 2: Write Spec (Claude Code does this)**
- Based on research, Claude Code writes a YAML spec file
- Uses the internal YAML format that `spec.ParseBytes` understands
- Write to a temp file or to the pipeline directory
- Claude Code knows the API because it just READ the docs

**Step 3: Generate (Go binary does this)**
- `printing-press generate --spec <spec-claude-wrote> --output <dir> --force`
- Template engine renders Go code
- Quality gates verify it compiles
- Claude Code reads the output

**Step 4: Polish (Claude Code does this)**
- Read the generated help descriptions and rewrite them
- Read the generated examples and improve them
- Read the generated README and rewrite it to sell the tool
- Use Edit tool to modify files directly - no Go package needed

**Step 5: Score (Go binary does this)**
- Run the scorecard: `go test ./internal/pipeline/ -run TestScorecardOnRealCLI`
- Claude Code reads the scorecard output
- If score is low, Claude Code writes a fix plan and iterates

**Step 6: Report (Claude Code does this)**
- Print the scorecard to the user
- Show before/after (regex spec vs Claude-written spec)
- Suggest next steps

## Implementation Units

### Unit 1: Rewrite the Skill

**File:** `skills/printing-press/SKILL.md`

The skill needs complete rewrite. It should have these workflows:

**Workflow 0: Natural Language** (`/printing-press Notion`)
```
1. Is there an OpenAPI spec? (check KnownSpecs, apis-guru, web search)
   YES -> go to Workflow 1
   NO  -> go to step 2

2. WebFetch the API docs URL
3. Read the docs. Write a YAML spec based on what you find.
   - List every endpoint (method, path, description)
   - Identify auth (bearer, api_key, oauth)
   - Find the base URL
   - Group by resource
   - Write spec to /tmp/<api>-spec.yaml

4. Run: printing-press generate --spec /tmp/<api>-spec.yaml --output ./<api>-cli --force
5. If quality gates fail, read the errors, fix the spec, regenerate
6. Read the generated --help output, rewrite any bad descriptions using Edit tool
7. Read the generated README.md, rewrite to sell the tool using Edit tool
8. Run the scorecard, report results
```

**Workflow 1: From Spec** (`/printing-press --spec <url>`)
```
1. Run: printing-press generate --spec <url> --output ./<name>-cli --force
2. Read generated code
3. Polish with Edit tool if --polish flag
4. Run scorecard, report
```

**Workflow 2: From Docs** (`/printing-press --docs <url>`)
```
1. WebFetch the docs URL
2. Write YAML spec (Claude Code does this, not regex)
3. Run: printing-press generate --spec <spec> --output ./<name>-cli --force
4. Polish, score, report
```

**Workflow 5: Scorecard** (`/printing-press score <dir>`)
```
1. Run: SCORECARD_CLI_DIR=<dir> go test ./internal/pipeline/ -run TestScorecardOnRealCLI -v
2. Read output, present to user
```

### Unit 2: Remove Go LLM Packages (or Mark as Optional Fallback)

The `internal/llm/`, `internal/llmpolish/` packages aren't needed when running from Claude Code. Two options:

**Option A: Delete them.** The skill does everything. The Go binary is pure templates.

**Option B: Keep as CLI-only fallback.** When someone runs `printing-press generate --polish` from their terminal (without Claude Code), the Go binary can still try to shell out to claude/codex for polish. But this is a fallback, not the primary path.

**Recommendation: Option B.** Keep the code but make it clear in docs that the primary path is through the skill. The Go packages are for headless/CI use when Claude Code isn't available.

### Unit 3: Test End-to-End from Claude Code

After rewriting the skill, test it by invoking it:

```
/printing-press Notion
```

This should:
1. Claude Code WebFetches Notion docs
2. Claude Code writes a YAML spec with 20+ endpoints
3. Go binary generates the CLI
4. Claude Code polishes the output
5. Scorecard runs
6. All within this Claude Code session - no nesting, no external calls

**This is the test that proves the architecture works.**

## Acceptance Criteria

- [ ] Skill SKILL.md rewritten with Claude Code as the brain
- [ ] `/printing-press Notion` works end-to-end in Claude Code
- [ ] Claude Code writes the YAML spec (not regex, not shelled-out LLM)
- [ ] Generated Notion CLI has 15+ commands (Claude Code reads docs properly)
- [ ] Claude Code polishes help text by editing files directly
- [ ] Scorecard runs and reports results
- [ ] No `claude` CLI nesting required
- [ ] Go binary stays dumb - just templates + quality gates

## Scope Boundaries

- Do NOT delete internal/llm/ or internal/llmpolish/ (keep as CLI fallback)
- Do NOT change the Go template engine
- Do NOT change the scorecard
- The skill is the only thing that changes
- --polish flag on the Go binary still works for CLI-only users
