---
name: printing-press
description: Generate a ship-ready CLI for an API with a lean research -> generate -> build -> shipcheck loop.
version: 2.0.0
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

Generate the best useful CLI for an API without burning an hour on phase theater.

```bash
/printing-press Notion
/printing-press Discord codex
/printing-press --spec ./openapi.yaml
/printing-press emboss ./library/notion-cli
```

## What Changed In v2

The old skill inflated the path to ship:
- too many mandatory research documents before code existed
- too many separate late-stage validation phases after code existed
- too many chances to discover obvious failures late

This version uses one lean loop:
1. Resolve the spec and write one research brief
2. Generate
3. Build the highest-value gaps
4. Run one shipcheck block
5. Optionally run live API smoke tests

Artifacts are still written, but only the ones that materially help the next step.

## Modes

### Default

Normal mode. Claude does research, generation orchestration, implementation, and verification.

### Codex Mode

If the arguments include `codex` or `--codex`, offload pure code-writing tasks to Codex CLI.

Use Codex for:
- writing store/data-layer code
- writing workflow commands
- fixing dead flags / dead code / path issues
- README cookbook edits

Keep on Claude:
- research and product positioning
- choosing which gaps matter
- verification results and ship decisions

If Codex fails 3 times in a row, stop delegating and finish locally.

### Emboss Mode

If the arguments start with `emboss`, this is a second-pass improvement cycle for an existing generated CLI.

```bash
/printing-press emboss ./library/notion-cli
```

Use the built-in audit command:

```bash
cd ~/cli-printing-press && ./printing-press emboss --dir <cli-dir> --spec <spec-path> --audit-only
```

Emboss is:
1. audit baseline
2. quick re-research
3. top-5 gap analysis
4. implement improvements
5. re-audit and report delta

Do not run emboss automatically.

## Rules

- Optimize for time-to-ship, not time-to-document.
- Reuse prior research whenever it is already good enough.
- Do not split one idea across multiple mandatory artifacts.
- Do not create a separate narrative phase for dogfood, dead-code audit, runtime verification, and final score. Treat them as one shipcheck block.
- Run cheap, high-signal checks early.
- Fix blockers and high-leverage failures first.
- Reuse the same spec path across `generate`, `dogfood`, `verify`, and `scorecard`.
- YAML, JSON, local paths, and URLs are all valid spec inputs for the verification tools.
- Maximum 2 verification fix loops unless the user explicitly asks for more.

## Outputs

Every run writes up to 4 concise artifacts in `~/cli-printing-press/docs/plans/`:

1. `<date>-feat-<api>-cli-brief.md`
2. `<date>-fix-<api>-cli-build-log.md`
3. `<date>-fix-<api>-cli-shipcheck.md`
4. `<date>-fix-<api>-cli-live-smoke.md` (only if live testing runs)

These do not need to be 200+ lines. Keep them dense, evidence-backed, and directly useful.

## Phase 0: Resolve And Reuse

Before new research:

1. Resolve the spec source.
2. Check for prior research in:
   - `~/cli-printing-press/docs/plans/*<api>*`
   - `~/docs/plans/*<api>*`
3. Reuse good prior work instead of redoing it.
4. Detect whether an API key is already available.

Token detection:
- GitHub: `GITHUB_TOKEN`, `GH_TOKEN`, or `gh auth token`
- Discord: `DISCORD_TOKEN`, `DISCORD_BOT_TOKEN`
- Linear: `LINEAR_API_KEY`
- Notion: `NOTION_TOKEN`
- Stripe: `STRIPE_SECRET_KEY`
- Generic: `API_KEY`, `API_TOKEN`

If a token is available, ask once whether to use it for read-only live testing at the end. Do not block the build on token collection.

## Phase 1: Research Brief

Write one build-driving brief, not a stack of phase essays.

The brief must answer:

1. What is this API actually used for?
2. What are the top 3-5 power-user workflows?
3. What are the top table-stakes competitor features?
4. What data deserves a local store?
5. Why would someone install this CLI instead of the incumbent?
6. What is the product name and thesis?

Research checklist:
- Find the spec or docs source
- Find the top 1-2 competitors
- Find 2-3 concrete user pain points
- Identify the highest-gravity entities
- Pick the top 3-5 commands that matter most

Do not produce separate mandatory documents for:
- workflow ideation
- parity audit
- data-layer prediction
- product thesis

Put them in the one brief.

Write:

`~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-brief.md`

Suggested shape:

```markdown
# <API> CLI Brief

## API Identity
- Domain:
- Users:
- Data profile:

## Top Workflows
1. ...

## Table Stakes
- ...

## Data Layer
- Primary entities:
- Sync cursor:
- FTS/search:

## Product Thesis
- Name:
- Why it should exist:

## Build Priorities
1. ...
2. ...
3. ...
```

## Phase 1.5: Ecosystem Absorb Gate

THIS IS A MANDATORY STOP GATE. Do not generate until this is complete and approved.

The GOAT CLI doesn't "find gaps." It absorbs EVERY feature from EVERY tool and then transcends with compound use cases nobody thought of. This phase builds the absorb manifest.

### Step 1.5a: Search for every tool that touches this API

Run these searches in parallel:

1. **WebSearch**: `"<API name>" Claude Code plugin site:github.com`
2. **WebSearch**: `"<API name>" MCP server model context protocol`
3. **WebSearch**: `"<API name>" Claude skill SKILL.md site:github.com`
4. **WebSearch**: `"<API name>" CLI tool site:github.com` (competing CLIs)
5. **WebSearch**: `"<API name>" CLI site:npmjs.com` (npm packages)
6. **WebFetch**: Check `github.com/anthropics/claude-plugins-official/tree/main/external_plugins` for official plugin
7. **WebSearch**: `"<API name>" MCP site:lobehub.com OR site:mcpmarket.com OR site:fastmcp.me`
8. **WebSearch**: `"<API name>" automation script workflow site:github.com`

### Step 1.5b: Catalog every feature into the absorb manifest

For EACH tool found, list EVERY feature/tool/command it provides. Then define how our CLI matches AND beats it:

```markdown
## Absorb Manifest

### Absorbed (match or beat everything that exists)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | Search issues by text | Linear MCP search_issues | FTS5 offline search | Works offline, regex, SQL composable |
| 2 | Create issue | Linear MCP create_issue | --stdin batch, --dry-run | Agent-native, scriptable, idempotent |
| 3 | Sprint board view | jira-cli sprint view | SQLite-backed sprint query | Historical velocity, offline |
```

Every row = a feature we MUST build. No exceptions. If someone else has it, we have it AND it works offline, with --json, --dry-run, typed exit codes, and SQLite persistence.

### Step 1.5c: Identify transcendence features

What compound use cases become possible ONLY when ALL absorbed features live in SQLite together?

```markdown
### Transcendence (only possible with our local data layer)
| # | Feature | Command | Why Only We Can Do This |
|---|---------|---------|------------------------|
| 1 | Bottleneck detection | bottleneck | Requires local join across issues + assignees + cycle data |
| 2 | Velocity trends | velocity --weeks 4 | Requires historical cycle snapshots in SQLite |
| 3 | Duplicate detection | similar "login bug" | Requires FTS5 across ALL issue text + comments |
```

Minimum 5 transcendence features. These are the NOI commands.

### Step 1.5d: Write the manifest artifact

Write to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-absorb-manifest.md`

### Phase Gate 1.5

**STOP.** Present the absorb manifest to the user:

"Found [N] features across [X] tools (MCPs, skills, CLIs, scripts). Our CLI will absorb all [N] and add [M] transcendence features. Total: [N+M] features. This is [Z]% more than the best existing tool. Approve to proceed to generation."

Use AskUserQuestion. WAIT for approval. Do NOT generate until approved.

---

## Phase 2: Generate

Use the resolved spec source and generate immediately.

OpenAPI / internal YAML:

```bash
cd ~/cli-printing-press && ./printing-press generate \
  --spec <spec-path-or-url> \
  --output ./library/<api>-cli \
  --force --lenient --validate
```

Docs-only:

```bash
cd ~/cli-printing-press && ./printing-press generate \
  --docs <docs-url> \
  --name <api> \
  --output ./library/<api>-cli \
  --force --validate
```

GraphQL-only APIs:
- Generate scaffolding only in Phase 2
- Build real commands in Phase 3 using a GraphQL client wrapper

After generation:
- note skipped complex body fields
- fix only blocking generation failures here
- do not start broad polish work yet

If generation fails:
- fix the specific blocker
- retry at most 2 times
- prefer generator fixes over manual generated-code surgery when the failure is systemic

## Phase 3: Build The GOAT

Build comprehensively. The absorb manifest from Phase 1.5 IS the feature list.

Priority 0 (foundation):
- data layer for ALL primary entities from the manifest
- sync/search/SQL path - this is what makes transcendence possible

Priority 1 (absorb - match everything):
- ALL absorbed features from the Phase 1.5 manifest
- Every feature from every competing tool, matched and beaten with agent-native output
- This is NOT "top 3-5" - it is the FULL manifest

Priority 2 (transcend - build what nobody else has):
- ALL transcendence features from Phase 1.5
- The NOI commands that only work because everything is in SQLite
- These are the commands that make someone say "I need this"

Priority 3 (polish):
- skipped complex request bodies that block important commands
- naming cleanup for ugly operationId-derived commands
- tests for non-trivial store/workflow logic

### Agent Build Checklist (per command)

After building each command in Priority 1 and Priority 2, verify these 7 principles are met. These map 1:1 to what Phase 4.9's agent readiness reviewer will check - apply them now so the review becomes a confirmation, not a catch-all.

1. **Non-interactive**: No TTY prompts, no `bufio.Scanner(os.Stdin)`, works in CI without a terminal
2. **Structured output**: `--json` produces valid JSON, `--select` filters fields correctly
3. **Progressive help**: `--help` shows realistic examples with domain-specific values (not "abc123")
4. **Actionable errors**: Error messages name the specific flag/arg that's wrong and the correct usage
5. **Safe retries**: Mutation commands support `--dry-run`, idempotent where possible
6. **Composability**: Exit codes are typed (0/2/3/4/5/7), output pipes to `jq` cleanly
7. **Bounded responses**: `--compact` returns only high-gravity fields, list commands have `--limit`

### Priority 1 Review Gate

After completing ALL Priority 1 (absorbed) features, BEFORE starting Priority 2 (transcendence):

Pick 3 random commands from Priority 1. Run each with:
```bash
<cli> <command> --help          # Does it show realistic examples?
<cli> <command> --dry-run       # Does it show the request without sending?
<cli> <command> --json          # Does it produce valid JSON?
```

If any of the 3 fail, there's a systemic issue. Fix it across all commands before proceeding. This catches problems like "--dry-run not wired" or "--json outputs table instead of JSON" early, when they're cheap to fix.

Get Priority 0 and 1 working first (the foundation and absorbed features), pass the review gate, then build Priority 2 (transcendence), then verify.

Write:

`~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-build-log.md`

Include:
- what was built
- what was intentionally deferred
- skipped body fields that remain
- any generator limitations found

## Phase 4: Shipcheck

Run one combined verification block.

```bash
cd ~/cli-printing-press
./printing-press dogfood   --dir ./library/<api>-cli --spec <same-spec>
./printing-press verify    --dir ./library/<api>-cli --spec <same-spec> --fix
./printing-press scorecard --dir ./library/<api>-cli --spec <same-spec>
```

Interpretation:
- `dogfood` catches dead flags, dead helpers, invalid paths, example drift, and broken data wiring
- `verify` catches runtime breakage and runs the auto-fix loop for common failures
- `scorecard` is the structural quality snapshot, not the source of truth by itself

Fix order:
1. generation blockers or build breaks
2. invalid paths and auth mismatches
3. dead flags / dead functions / ghost tables
4. broken dry-run and runtime command failures
5. scorecard-only polish gaps

Ship threshold:
- `verify` verdict is `PASS` or high `WARN` with 0 critical failures
- `dogfood` no longer fails because of spec parsing, binary path, or skipped examples
- `scorecard` is at least 65, or meaningfully improved and no core behavior is broken

Maximum 2 shipcheck loops by default.

Write:

`~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-shipcheck.md`

Include:
- command outputs and scores
- top blockers found
- fixes applied
- before/after verify pass rate
- before/after scorecard total
- final ship recommendation: `ship`, `ship-with-gaps`, or `hold`

## Phase 5: Optional Live Smoke

Only run this if a token is available and the user agreed.

Use read-only smoke tests:
- `--help`
- one or two representative GET/list commands
- sync/search/health path if a local data layer exists

If live smoke finds bugs:
- fix only the real bug
- re-run the shipcheck block once

Write:

`~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-live-smoke.md`

## Fast Guidance

### When to use `printing-press print`

Use `./printing-press print <api>` only when the user explicitly wants a resumable on-disk pipeline with phase seeds. It is optional.

The fast path for `/printing-press <API>` is:
- brief
- generate
- build
- shipcheck

### When to stop researching

Stop when you can answer:
- what to build first
- what data to persist
- what incumbent features cannot be missing

If the next research step does not change those answers, stop and generate.

### What not to do

Do not:
- write 5 separate mandatory research documents
- defer all workflows to “future work”
- skip verification because the CLI compiles
- treat scorecard alone as ship proof
- discover YAML/URL spec incompatibility late and manually convert specs if the tools can already consume them
- rerun the whole late-phase gauntlet for cosmetic README polish
- skip features because “the MCP already handles that” (absorb everything, beat it with offline + agent-native)
- build only “top 3-5 workflows” when the absorb manifest has 15+ (build them ALL, then transcend)
- generate before the Phase 1.5 Ecosystem Absorb Gate is approved
- call a CLI “GOAT” without matching every feature the best competitor has

### What counts as success

Success is:
- a generated CLI that gets to shipcheck without generator blockers
- verification tools working against the same spec the user generated from
- one or two fix loops, not a maze of re-entry phases
- a CLI that is plausibly shippable today, not a perfect design memo
