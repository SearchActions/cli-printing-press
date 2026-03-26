---
name: printing-press
description: Generate the GOAT CLI for any API. 5-phase loop with dual Steinberger analysis, deep competitor research, complex body field handling, and before/after scoring delta.
version: 1.1.0
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - WebSearch
  - AskUserQuestion
  - Agent
---

# /printing-press

Generate the best CLI that has ever existed for any API. Five mandatory phases. Dual Steinberger analysis. No shortcuts.

```
/printing-press Notion
/printing-press Plaid payments API
/printing-press --spec ./openapi.yaml
```

## Prerequisites

- Go 1.21+ installed
- The printing-press repo at `~/cli-printing-press`
- Build binary if missing: `cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press`

## How This Works

Every run produces the GOAT CLI through 6 mandatory phases:

```
PHASE 0: VISIONARY RESEARCH  ->  PHASE 1: DEEP RESEARCH  ->  PHASE 2: GENERATE  ->  PHASE 3: STEINBERGER AUDIT  ->  PHASE 4: GOAT FIX  ->  PHASE 5: FINAL STEINBERGER
        (3-5 min)                      (5-8 min)                  (1-2 min)               (5-8 min)                     (3-5 min)                (2-3 min)
```

Total expected time: 18-30 minutes. If a run completes in under 12 minutes, phases were shortcut.

**The Steinberger bar:** Peter Steinberger's gogcli is the 10/10 reference. Every generated CLI is scored against it TWICE - once during audit to find gaps, once after fixes to prove improvement. The delta is the proof of work.

**Grade thresholds (10 dimensions, 100 max):**
- **Grade A:** 80+/100 (80%)
- **Grade B:** 65-79/100 (65-79%)
- **Grade C:** 50-64/100 (50-64%)

---

## Workflow: `--spec` shortcut

When the user provides `--spec <path-or-url>`, skip Phase 1 spec search (spec is provided). STILL run competitor research (Steps 1.2-1.5). Run all other phases.

## Workflow: Natural Language (Primary)

When the user provides an API name, run ALL five phases.

### Step 0: Parse intent and check known specs

Extract the API name. Check `~/cli-printing-press/skills/printing-press/references/known-specs.md`.

If found in registry: note the URL for Phase 2, but STILL run full Phase 1 research.
If not found: Phase 1 searches for spec too.

---

# PHASE 0: VISIONARY RESEARCH

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Before generating any CLI, understand what a thoughtful developer would build - not just what the OpenAPI spec says.

### Step 0a: API Identity & Domain Understanding

Understand what this API IS:

1. **WebFetch** the API's developer docs landing page
2. **WebSearch**: `"<API name>" developer documentation overview`
3. Extract:
   - **Domain category:** messaging, payments, productivity, infrastructure, analytics
   - **Primary users:** Who uses this API? (e.g., "bot developers", "server admins")
   - **Core entities:** What are the main objects? (e.g., "guilds", "channels", "messages")
   - **Data profile:**
     - Write pattern: append-only, mutable, or event-sourced?
     - Volume: high (millions of records), medium, or low?
     - Real-time: does the API have webhooks/websockets/SSE?
     - Search need: high (users need to find things) or low?

### Step 0b: Usage Pattern Discovery

Discover what people ACTUALLY DO with this API:

**Community Research (run these in parallel):**
1. **WebSearch**: `"<API name>" CLI workflow site:reddit.com`
2. **WebSearch**: `"<API name>" automation script site:github.com`
3. **WebSearch**: `"<API name>" "I built" OR "I made" OR "my tool" site:reddit.com OR site:news.ycombinator.com`
4. **WebSearch**: `"<API name>" tutorial automation workflow 2025 2026`

**Pain Point Research:**
5. **WebSearch**: `"<API name>" API "pain point" OR "limitation" OR "workaround"`
6. **WebSearch**: `site:stackoverflow.com "<API name>" API rate limit OR pagination OR bulk`

From all research, identify the **top 5 usage patterns** ranked by evidence score:

| Source | Weight |
|---|---|
| Existing tool with 100+ stars | 3 points |
| Existing tool with 10-99 stars | 2 points |
| Reddit/HN post with 50+ upvotes | 2 points |
| Reddit/HN post with 10-49 upvotes | 1 point |
| Stack Overflow question with 10+ votes | 1 point |
| Blog post / tutorial | 1 point |
| GitHub issue on competitor | 1 point |
| Cross-platform appearance (same need on 2+ platforms) | +2 bonus |

Score >= 6: Strong evidence. Include in CLI.
Score 3-5: Moderate evidence. Consider as optional.
Score < 3: Weak evidence. Skip.

### Step 0c: Tool Landscape Discovery (The Discrawl Finder)

Find ALL tools for this API, not just API wrappers:

**Tier 1: Direct CLI Search** (existing Phase 1 does this too)
1. **WebSearch**: `"<API name>" CLI tool github`

**Tier 2: Non-Wrapper Tool Search** (CRITICAL - finds discrawl-class tools)
2. **WebSearch**: `"<API name>" sync OR archive OR mirror site:github.com`
3. **WebSearch**: `"<API name>" search engine OR analytics OR dashboard site:github.com`
4. **WebSearch**: `"<API name>" backup OR export OR migration site:github.com`
5. **WebSearch**: `"<API name>" monitor OR watcher OR alerting site:github.com`

**Tier 3: Ecosystem Search**
6. **WebSearch**: `awesome "<API name>" site:github.com`

For each tool found, classify it:

| Type | Description | Example |
|---|---|---|
| **API Wrapper** | Translates HTTP to CLI flags | discli |
| **Data Tool** | Adds local persistence/search | discrawl |
| **Workflow Tool** | Orchestrates multi-step sequences | Stripe fixtures |
| **Environment Tool** | Runs local simulation | Supabase CLI |
| **Integration Tool** | Bridges to other systems | Zapier integration |

**The press should generate CLIs that compete with Data Tools and Workflow Tools, not just API Wrappers.**

### Step 0d: Workflow Analysis

From usage patterns (0b) and tool landscape (0c), identify multi-step workflows:

For each workflow, document:
- **Steps:** The sequence of API calls
- **Frequency:** How often users perform this
- **Pain point:** What makes this hard with the raw API
- **Proposed CLI feature:** What compound command would solve it

### Step 0e: Architecture Planning

Based on data profile and workflows, decide what the CLI needs:

| Data Profile | Architecture |
|---|---|
| High volume + search need | SQLite + FTS5 (Discord, Slack) |
| Transaction data + reconciliation | Local ledger with diff tracking (Stripe, Plaid) |
| Document data + offline editing | Local Markdown/JSON sync (Notion, Confluence) |
| Low volume + simple CRUD | Standard API wrapper is fine (most APIs) |

For each decision area (persistence, real-time, search, bulk, caching), document:
- **Need level:** High / Medium / Low
- **Decision:** What to use
- **Rationale:** Why

### Step 0f: Feature Ideation - "Next 5 Features for the World"

Score each feature idea on 8 dimensions (16-point max):

| Dimension | Weight | Scoring |
|---|---|---|
| **Evidence strength** | 3 | 3=existing tool 100+ stars, 2=Reddit/SO demand, 1=weak, 0=speculation |
| **User impact** | 3 | 3=most users feel this pain, 2=niche, 1=nice-to-have, 0=nobody asked |
| **Implementation feasibility** | 2 | 2=can generate template, 1=needs custom code, 0=major infrastructure |
| **Uniqueness** | 2 | 2=no existing tool, 1=improves on existing, 0=already well-served |
| **Composability** | 2 | 2=great with pipes/agents, 1=somewhat, 0=interactive-only |
| **Data profile fit** | 2 | 2=perfect fit, 1=possible, 0=wrong shape |
| **Maintainability** | 1 | 1=generated code supports it, 0=needs human maintenance |
| **Competitive moat** | 1 | 1=hard to replicate, 0=trivial |

Score >= 12: **Must-have.** Build it.
Score 8-11: **Should-have.** Include as optional.
Score < 8: **Won't-have.** Skip or future work.

### Step 0g: Write the Visionary Research Artifact

**Write** to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-visionary-research.md`:

```markdown
## Visionary Research: <API> CLI

### API Identity
- Domain: <category>
- Primary users: <who>
- Data profile: <write pattern>, <volume>, <realtime>, <search need>

### Usage Patterns (Top 5 by Evidence)
1. <pattern> (Evidence: X/10) - <what it needs>
2. ...

### Tool Landscape (Beyond API Wrappers)
- <tool> (<stars> stars): <what it does>
- ...

### Workflows
1. <name>: <steps> -> Proposed: `<api>-cli <command>`
2. ...

### Architecture Decisions
- Persistence: <decision> because <rationale>
- Real-time: <decision> because <rationale>
- Search: <decision> because <rationale>
- Bulk: <decision> because <rationale>
- Cache: <decision> because <rationale>

### Top 5 Features for the World
1. <feature> (Score: X/16) - <1-line description>
2. ...
```

### PHASE GATE 0

**STOP.** Verify ALL of these before proceeding:
1. API Identity documented with data profile
2. At least 3 usage patterns with evidence scores
3. Tool landscape includes non-wrapper tools (Tier 2 search done)
4. At least 2 workflows with proposed CLI features
5. Architecture decisions match data profile
6. Top 5 features scored and ranked

Tell the user: "Phase 0 complete: Domain: [category]. Data profile: [volume]/[realtime]/[search]. Found [N] non-wrapper tools. Top feature: [name] (score [X]/16). Architecture: [key decision]. Proceeding to deep research."

---

# PHASE 1: DEEP RESEARCH

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Research the API landscape deeply. You need to understand the competitive terrain, user pain points, and strategic opportunity before generating anything.

### Step 1.1: Search for the OpenAPI spec

If not in known-specs registry:

1. **WebSearch**: `"<API name>" openapi spec site:github.com`
2. **WebSearch**: `"<API name>" openapi.yaml OR openapi.json specification`
3. Try common URL patterns
4. If found, **WebFetch** first 500 bytes to verify

If no spec found: plan to write one from docs in Phase 2.

### Step 1.2: Search for competing CLIs

**WebSearch**: `"<API name>" CLI tool github`
**WebSearch**: `"<API name>" command line client`

Also search for non-wrapper tools discovered in Phase 0:
**WebSearch**: `"<API name>" sync OR archive OR export site:github.com`

For each competitor found, note repo URL, star count, language.

### Step 1.3: Deep competitor analysis (TOP 2 competitors)

For the top 2 competitors by stars, do ALL of the following:

1. **WebFetch** their README - count commands, note features, assess quality
2. **WebFetch** their GitHub repo main page - check:
   - Last commit date (is it maintained?)
   - Open issue count
   - Number of contributors
3. **WebSearch**: `site:github.com/<org>/<repo>/issues` - look for:
   - User complaints about missing features
   - Requests for specific functionality
   - Pain points users report
4. Record at least 2 specific user quotes or pain points.

### Step 1.4: Check demand signals

**WebSearch**: `"<API name>" "need a CLI" OR "command line" site:reddit.com OR site:news.ycombinator.com`

### Step 1.5: Strategic justification

Answer this question explicitly: **"Why should this CLI exist when [best competitor] already has [N] stars?"**

The answer must be SPECIFIC. Not just "agent-native." Examples:
- "[Competitor] hasn't been updated since [date] and doesn't support the latest API version"
- "No existing CLI supports --json + --select + --dry-run for agent workflows"
- "Users on [platform] are asking for [specific feature] which no CLI provides"
- "[Competitor] has [N] open issues about [problem] we can solve"

### Step 1.6: Write the research artifact

**Write** to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-research.md`:

```markdown
---
title: "Research: <API> CLI"
type: feat
status: active
date: <today>
---

# Research: <API> CLI

## Spec Discovery
- Official OpenAPI spec: <url or "none found - will write from docs">
- Source: <where found>
- Format: <OpenAPI 3.x / Swagger 2.0 / internal YAML>
- Endpoint count: <N>

## Competitors (Deep Analysis)

### <Competitor 1> (<stars> stars)
- Repo: <url>
- Language: <lang>
- Commands: <count>
- Last commit: <date>
- Open issues: <count>
- Maintained: <yes/no>
- Notable features: <list>
- Weaknesses: <what users complain about>

### <Competitor 2> (<stars> stars)
- [same structure]

## User Pain Points
> "<quote from GitHub issue or Reddit>" - <source>
> "<quote>" - <source>

## Auth Method
- Type: <api_key / oauth2 / bearer_token>
- Env var convention: <what competitors use>

## Demand Signals
- <specific posts with URLs, or "none found">

## Strategic Justification
**Why this CLI should exist:** <specific answer, not just "agent-native">

## Target
- Command count: <N - match or beat best competitor>
- Key differentiator: <specific features we'll have that competitors don't>
- Quality bar: Steinberger Grade A (80+/100)
```

### PHASE GATE 1

**STOP.** Verify ALL of these before proceeding:
1. Research artifact exists with Spec Discovery section
2. At least 2 competitors analyzed with maintenance status
3. At least 2 user quotes or pain points documented
4. Strategic justification answers "why should this exist?"
5. Target command count is set

Tell the user: "Phase 1 complete: Found [spec/no spec], [N] competitors. Best: [name] ([stars] stars, [commands] commands, last commit [date]). Strategic angle: [1-sentence justification]. Proceeding to generation."

---

# PHASE 2: GENERATE

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

### Step 2.1: Get the spec ready

**If OpenAPI spec found:**
```bash
curl -sL -o /tmp/printing-press-spec-<api>.json "<spec-url>" && head -c 200 /tmp/printing-press-spec-<api>.json
```

**If no spec (write from docs):**
1. **WebFetch** the API docs
2. **Read** `~/cli-printing-press/skills/printing-press/references/spec-format.md`
3. Write YAML spec to `/tmp/<api>-spec.yaml` with ALL endpoints
4. Include auth config matching research findings

### Step 2.2: Check for existing output, remove if exists

```bash
cd ~/cli-printing-press && rm -rf <api>-cli 2>/dev/null; echo "CLEAN"
```

### Step 2.3: Run the generator

```bash
cd ~/cli-printing-press && ./printing-press generate \
  --spec /tmp/printing-press-spec-<api>.json \
  --output ./<api>-cli \
  --force --lenient --validate 2>&1
```

### Step 2.4: Note skipped complex body fields

**IMPORTANT:** When the generator outputs "warning: skipping body field X: complex type not supported", note EVERY skipped field. You will handle these in Phase 4.

Run:
```bash
cd ~/cli-printing-press && ./printing-press generate --spec /tmp/printing-press-spec-<api>.json --output ./<api>-cli --force --lenient --validate 2>&1 | grep "skipping body field"
```

Save the list of skipped fields. These are NOT acceptable limitations - they are work items for Phase 4.

### Step 2.5: Handle quality gate failures

Max 3 retries. Read errors carefully and fix spec issues.

### PHASE GATE 2

**STOP.** Verify:
1. CLI directory exists
2. `go build ./...` succeeds
3. List of skipped complex body fields is saved for Phase 3

Tell the user: "Phase 2 complete: Generated <api>-cli with [N] resources, [M] endpoints. [K] complex body fields noted for Phase 4. Proceeding to Steinberger audit."

---

# PHASE 3: STEINBERGER AUDIT

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

This phase has TWO parts: (A) code review for tactical fixes, and (B) Steinberger analysis for strategic assessment. Both are required.

## Part A: Code Review

### Step 3.1: Read the generated code

You MUST **Read** these files (not just check they exist):

- `<api>-cli/internal/cli/root.go`
- `<api>-cli/README.md`
- At least 3 resource command files

### Step 3.2: Count commands and compare to target

```bash
cd ~/cli-printing-press/<api>-cli && grep -r "Use:" internal/cli/*.go | grep -v "root.go" | wc -l
```

Compare against target from Phase 1 research.

### Step 3.3: Check help text quality

- Descriptions: developer-friendly or raw spec jargon?
- Examples: realistic values or placeholder garbage ("string", "0", "abc123")?
- Root command: does it explain what the API does?

### Step 3.4: Check for missing endpoints

Compare spec endpoints against generated commands. Note any gaps.

### Step 3.5: Check agent-native features

Verify in root.go: --json, --select, --dry-run, --stdin, --yes, --no-cache, doctor.

### Step 3.6: Check complex body fields

For each field skipped by the generator (from Phase 2 Step 2.4):
1. Is this field critical for the endpoint's purpose?
2. Can the user work around it with `--stdin`?
3. What example JSON would a user pipe in?

## Part B: First Steinberger Analysis

### Step 3.0: Run automated scorecard

Before hand-scoring, run the automated scorecard to get objective baseline numbers:

```bash
cd ~/cli-printing-press && ./printing-press scorecard --dir ./<api>-cli
```

Use these numbers as the baseline. The hand-scoring in Step 3.7 should explain WHY each dimension got its score, not re-guess the number.

### Step 3.7: Score against the Steinberger bar

Score each dimension 0-10. For EACH dimension, provide THREE things:
1. **Current score** with justification
2. **What 10/10 looks like** (reference gogcli or best-in-class)
3. **What specific changes would raise the score** (actionable items)

```markdown
## First Steinberger Analysis (Baseline)

| Dimension | Score | What 10 Looks Like | How to Get There |
|-----------|-------|-------------------|-----------------|
| Output modes | X/10 | gogcli: --json, --yaml, --csv, --table, --select, --quiet, --template | Add --yaml output, add --template for custom formats |
| Auth | X/10 | gogcli: OAuth browser flow, token storage, multiple profiles, doctor validates | Add OAuth flow, add profile switching |
| Error handling | X/10 | gogcli: typed exits, retry with backoff, helpful suggestions, link to docs | Add "did you mean?" suggestions |
| Terminal UX | X/10 | gogcli: progress spinners, color themes, pager for long output | Add progress spinners for pagination |
| README | X/10 | gogcli: install, quickstart, every command with example, cookbook, FAQ | Add cookbook section, add FAQ |
| Doctor | X/10 | gogcli: validates auth, API version, rate limits, config file health | Add API version check, config health |
| Agent-native | X/10 | gogcli: --json, --select, --dry-run, --stdin, idempotent, typed exits, no TTY | Already strong if all flags present |
| Local Cache | X/10 | gogcli: file cache + optional embedded DB (bolt/badger), --no-cache bypass, cache clear | [what changes would raise score] |
| Breadth | X/10 | gogcli: 100+ commands covering every API endpoint + convenience wrappers | Add missing commands, add convenience wrappers |
| Vision | X/10 | discrawl: SQLite + FTS5 + sync + search + tail + domain workflows | Add export, search, sync commands based on Phase 0 research |

**Baseline Total: X/100 (Grade X)**
```

### Step 3.8: Write the GOAT improvement plan

Based on the Steinberger analysis, identify:

1. **Top 5 highest-impact improvements** (will raise the score the most)
2. **Commands to ADD** (not just rename - new functionality)
3. **Complex body field examples** to add (top 3 endpoints where --stdin matters most)
4. **What's achievable in Phase 4** vs what's future work

### Step 3.9: Write the audit artifact

**Write** to `~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-audit.md`:

Include ALL of:
- Command comparison
- Help text quality assessment
- Agent-native checklist
- Specific fixes needed (file paths + what to change)
- First Steinberger Analysis table (full)
- GOAT improvement plan (top 5 + commands to add)
- Complex body field plan

### PHASE GATE 3

**STOP.** Verify ALL of these:
1. Audit artifact exists with Steinberger analysis table
2. Each Steinberger dimension has: score, "what 10 looks like", and "how to get there"
3. GOAT plan has at least 5 specific improvements
4. Complex body fields have a plan (not just "limitation")
5. Baseline total score is recorded

Tell the user: "Phase 3 complete: Baseline Steinberger Score: [X]/100 (Grade [X]). Found [N] tactical fixes + [M] GOAT improvements. Top improvement: [description]. Proceeding to fixes."

---

# PHASE 4: GOAT FIX

## THIS PHASE IS MANDATORY.

Execute fixes in priority order:

1. **Scorecard-gap fixes** - run scorecard, identify dimensions below 10/10, fix patterns the scorecard measures
2. **Complex body field --stdin examples** (useful for agents, visible in help text)
3. **Command name cleanup** (UX quality, not scored by automated scorecard)
4. **Description/README polish** (UX quality, not scored)

Scorecard-measured improvements first. UX polish second. If the scorecard says 10/10 for a dimension, do not spend time improving it further.

### Step 4.1: Apply scorecard-gap fixes

Run the scorecard and fix dimensions below 10/10:

```bash
cd ~/cli-printing-press && ./printing-press scorecard --dir ./<api>-cli
```

For each dimension below 10/10:
1. **Read** the relevant file
2. **Edit** with surgical changes
3. Focus on changes that RAISE THE SCORECARD NUMBER

### Step 4.2: Add complex body field examples

For the top 3 endpoints with complex body fields (identified in Phase 3 Step 3.6):

1. **Read** the command file
2. **Edit** the Example field to include a `--stdin` example with realistic JSON:

```go
Example: `  # Get a page
  <api>-cli pages get d9824bdc-8445-4327-be8b-5b47f462e1b0

  # Create a page with complex properties (pipe JSON via stdin)
  echo '{"parent":{"database_id":"..."},"properties":{"Name":{"title":[{"text":{"content":"My Page"}}]}}}' | <api>-cli pages create --stdin`,
```

### Step 4.3: Command name cleanup and description/README polish

Only if time remains after scorecard-gap fixes:
1. Fix command names -> clean, intuitive names (get/list/create/update/delete)
2. Fix help text jargon -> developer-friendly descriptions
3. Fix examples -> realistic values (real UUIDs, real API objects, not "abc123")
4. Fix README -> compelling, useful documentation

### Step 4.4: Verify compilation

```bash
cd ~/cli-printing-press/<api>-cli && go build ./... && go vet ./... && echo "ALL FIXES VERIFIED"
```

### PHASE GATE 4

**STOP.** Verify:
1. All GOAT improvements applied
2. All tactical fixes applied
3. Complex body field examples added for at least 2 endpoints
4. `go build ./...` and `go vet ./...` pass
5. Count changed files: `cd <api>-cli && git diff --stat 2>/dev/null | tail -1` (if git tracked) or `find . -newer /tmp/printing-press-spec-<api>.json -name "*.go" | wc -l`

Tell the user: "Phase 4 complete: Applied [N] improvements, [M] tactical fixes, [K] complex body field examples. Compilation verified. Proceeding to final Steinberger scoring."

---

# PHASE 5: FINAL STEINBERGER + REPORT

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

### Step 5.1: Second Steinberger Analysis (Post-Fix)

Run the automated scorecard again to measure improvement:

```bash
cd ~/cli-printing-press && ./printing-press scorecard --dir ./<api>-cli
```

Re-score ALL 10 dimensions. Show the DELTA from the baseline:

```markdown
## Final Steinberger Analysis (Post-Fix)

| Dimension | Before | After | Delta | What Changed |
|-----------|--------|-------|-------|-------------|
| Output modes | X/10 | Y/10 | +Z | [specific change] |
| Auth | X/10 | Y/10 | +Z | [specific change] |
| Error handling | X/10 | Y/10 | +Z | [specific change] |
| Terminal UX | X/10 | Y/10 | +Z | [specific change] |
| README | X/10 | Y/10 | +Z | [specific change] |
| Doctor | X/10 | Y/10 | +Z | [specific change] |
| Agent-native | X/10 | Y/10 | +Z | [specific change] |
| Local Cache | X/10 | Y/10 | +Z | [specific change] |
| Breadth | X/10 | Y/10 | +Z | [specific change] |
| Vision | X/10 | Y/10 | +Z | [specific change] |

**Before: X/100 -> After: Y/100 (+Z points)**
**Grade: [A/B/C]**
```

### Step 5.2: Remaining gaps

For each dimension still < 8/10:
- What would it take to reach 8+?
- Is this a generator limitation or achievable with more fix time?
- Tag as "future work" with a specific next step

### Step 5.3: Present the final report

Show ALL of these sections:

**1. Summary:**
```
Generated <api>-cli with <N> resources and <M> commands.
Resources: <comma-separated list>
```

**2. Steinberger Score (Before/After):**
```
Steinberger Score: Before X/100 -> After Y/100 (+Z points) - Grade [A/B/C]

[Full before/after table from Step 5.1]
```

**3. Competitor Comparison:**
```
Found <N> competing CLIs.
Best competitor: <name> (<stars> stars, <commands> commands)
Strategic advantage: <why ours is better - from Phase 1 research>
We beat them on: <specific features>
Remaining gap: <what they have that we don't, or "none">
```

**4. Example Commands (with complex body examples):**
```bash
cd ~/cli-printing-press/<api>-cli
go install ./cmd/<api>-cli

export <AUTH_ENV_VAR>="..."

# Basic usage
<api>-cli --help
<api>-cli doctor
<api>-cli <resource> list --json
<api>-cli <resource> get <realistic-id>

# Complex body fields (pipe JSON via stdin)
echo '<realistic-json>' | <api>-cli <resource> create --stdin

# Agent workflow
<api>-cli <resource> list --json --select id,name | jq -r '.[].id'
```

**5. Spec source and limitations**

**6. Future work** (from remaining gaps)

---

## Writing Specs from Docs

When no OpenAPI spec exists:

1. **WebFetch** the API docs
2. **Read** `~/cli-printing-press/skills/printing-press/references/spec-format.md`
3. Read the docs and identify EVERY endpoint
4. Write YAML spec to `/tmp/<api>-spec.yaml`
5. Generate from it

You ARE the brain. Read the docs yourself and write the spec.

## Submit to Catalog

`/printing-press submit <name>` - gather metadata, write `catalog/<name>.yaml`, create PR.

## Safety Gates

- Preview before generating
- Output directory conflict: check before overwriting
- Untrusted specs: note if not from known-specs registry
- Max 3 retries on quality gate failure

## Anti-Shortcut Rules

These phrases indicate a phase was shortcut. If you catch yourself writing them, STOP and re-do the phase:

- "This is a limitation of the generator" (fix it, don't accept it)
- "Complex types not supported" (add --stdin examples)
- "We'll skip this for now" (no skipping - do it or explain why it's impossible)
- "The quality is good enough" (score it against Steinberger, prove it's good enough with numbers)
- "Let's wrap up" (are all 5 phases complete with artifacts?)
- "This API doesn't need local persistence" (Did you run Phase 0? Check the data profile. If search need is high, it needs persistence.)
- "This is just an API wrapper" (Run Phase 0 again. What would a thoughtful developer build?)
