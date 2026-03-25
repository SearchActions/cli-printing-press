---
title: "Fix Printing Press Skill: Enforce the 5-Phase Loop"
type: fix
status: completed
date: 2026-03-25
---

# Fix: Printing Press Skill Loop Enforcement

## Problem Statement

The printing-press skill is supposed to run a 5-phase loop that takes 15-20 minutes:

```
ce:plan (RESEARCH) -> ce:work (GENERATE) -> ce:plan (AUDIT) -> ce:work (FIX) -> SCORE
```

Instead, it completes in 1m42s by finding a spec, running the generator, and presenting results. No research, no audit, no fixes, no scoring. The output is "dumb" - it compiles but hasn't been compared against competitors, polished, or scored.

## Root Cause Analysis

There are **three cascading failures**, not one:

### RC1: Version mismatch (the installed skill is v0.3.0, not v0.6.0)

The installed skill at `~/.claude/skills/printing-press/SKILL.md` is **v0.3.0**. It has zero awareness of the plan-execute-plan-execute loop. It just searches for a spec, runs `printing-press generate`, and presents the output.

The v0.6.0 skill with the loop exists only in the repo at `~/cli-printing-press/skills/printing-press/SKILL.md`. The commit `68f63167` ("restore plan-execute-plan-execute loop") updated the repo copy but **never synced it to the installed location**.

This is the primary reason the Notion run took 1m42s - it literally followed v0.3.0 instructions, which have no loop.

### RC2: Skill-within-skill invocation is fragile

Even v0.6.0 has a structural problem. It instructs Claude to call:
```
Skill("compound-engineering:ce:plan", "docs/plans/<date>-feat-<api>-cli-research-plan.md")
```

This is a skill invoking another skill via the Skill tool. In practice, Claude shortcuts through this because:
- The printing-press skill is already loaded and executing
- Claude treats the `Skill()` call as a suggestion, not a hard gate
- There's no structural enforcement that ce:plan actually ran before proceeding to the next phase
- Claude optimizes for speed and skips phases it thinks are optional

### RC3: No phase gates or timing enforcement

The skill says "always do phases 1-5" but provides no way to verify each phase completed. There are no:
- Phase completion markers (files that must exist before the next phase starts)
- Minimum time expectations (a 2-minute run should be flagged as suspicious)
- Output requirements (each phase must produce a specific artifact)
- Blocking checks (don't start Phase 3 until Phase 2's output directory exists)

## Proposed Solution

### Fix 1: Sync the skill (immediate)

Copy v0.6.0+ from the repo to the installed location:
```bash
cp ~/cli-printing-press/skills/printing-press/SKILL.md ~/.claude/skills/printing-press/SKILL.md
```

Add a sync reminder to CLAUDE.md or create a sync script.

### Fix 2: Inline the loop instead of calling sub-skills

Replace `Skill("compound-engineering:ce:plan", ...)` calls with **inline instructions that do the research and audit directly**. The printing-press skill already has WebSearch, WebFetch, Read, Write, Edit, and Agent in its allowed-tools. It doesn't need to delegate to ce:plan - it can do the research itself.

This eliminates the fragile skill-within-skill invocation entirely.

### Fix 3: Add phase gates with artifact requirements

Each phase must produce a specific artifact file. The next phase cannot start until the previous artifact exists and is non-empty.

| Phase | Artifact | Minimum Content |
|-------|----------|----------------|
| 1: RESEARCH | `docs/plans/<date>-feat-<api>-cli-research.md` | Spec URL or docs URL, competitor list with command counts, auth method |
| 2: GENERATE | `<api>-cli/cmd/<api>-cli/main.go` | Generated CLI directory with passing quality gates |
| 3: AUDIT | `docs/plans/<date>-fix-<api>-cli-audit.md` | Audit checklist with findings, specific fix list with file paths |
| 4: FIX | Compilation passes after edits | `go build ./...` succeeds in the CLI directory |
| 5: SCORE | Score output presented to user | Steinberger dimensions, grade, competitor comparison |

### Fix 4: Make phases explicit and unambiguous in the skill

Rewrite the skill with a rigid phase structure. Each phase has:
1. A **STOP** marker that says "DO NOT PROCEED until this phase is complete"
2. Explicit tool calls to make (not suggestions)
3. Required outputs that must be produced
4. A phase transition check

## Acceptance Criteria

- [ ] Installed skill at `~/.claude/skills/printing-press/SKILL.md` matches repo version
- [ ] Skill does NOT call `Skill("compound-engineering:ce:plan", ...)` - research/audit are inlined
- [ ] Phase 1 (RESEARCH) performs at least 3 WebSearches and produces a research artifact
- [ ] Phase 2 (GENERATE) runs the printing-press binary and produces a compiled CLI
- [ ] Phase 3 (AUDIT) reads generated code, compares against competitors, writes specific fixes
- [ ] Phase 4 (FIX) edits generated code and verifies compilation
- [ ] Phase 5 (SCORE) runs scorecard or presents a structured quality assessment
- [ ] A `/printing-press Notion` run takes 10-20 minutes, not under 5
- [ ] Each phase produces a visible output to the user (not silent)
- [ ] Sync script or instructions exist so installed version doesn't drift from repo again

## Technical Approach

### New SKILL.md Structure (v1.0.0)

The rewritten skill will have this structure:

```
# Phase 1: RESEARCH
## MANDATORY - DO NOT SKIP
### Step 1.1: Search for OpenAPI spec
### Step 1.2: Search for competing CLIs
### Step 1.3: Fetch and analyze competitor READMEs
### Step 1.4: Write research summary
### PHASE GATE: Verify research artifact exists

# Phase 2: GENERATE
## MANDATORY - DO NOT SKIP
### Step 2.1: Download spec or write from docs
### Step 2.2: Run printing-press generate
### Step 2.3: Verify quality gates pass
### PHASE GATE: Verify CLI directory exists and compiles

# Phase 3: AUDIT
## MANDATORY - DO NOT SKIP
### Step 3.1: Read all generated Go files
### Step 3.2: Check command count vs competitors
### Step 3.3: Review help descriptions for jargon
### Step 3.4: Check for missing endpoints
### Step 3.5: Write audit findings
### PHASE GATE: Verify audit artifact has specific fixes

# Phase 4: FIX
## MANDATORY - DO NOT SKIP
### Step 4.1: Execute each fix from the audit
### Step 4.2: Verify compilation after fixes
### PHASE GATE: go build ./... passes

# Phase 5: SCORE + REPORT
## MANDATORY - DO NOT SKIP
### Step 5.1: Run scorecard or manual quality assessment
### Step 5.2: Present final report with all required sections
```

### Key Design Decisions

1. **Inline research, not delegated** - The skill does its own WebSearches and WebFetches directly. No calling ce:plan as a sub-skill. This is more reliable because the instructions execute in the same context.

2. **Agent tool for parallel research** - Phase 1 can use the Agent tool to run parallel research queries (spec search, competitor search, demand signals) for speed without sacrificing thoroughness.

3. **Explicit file reads in audit** - Phase 3 must `Read` the generated files, not just run `go vet`. It needs to actually read `root.go`, `command.go` files, the `readme.md`, and compare against competitor command lists.

4. **Edit tool for fixes** - Phase 4 uses the Edit tool to make targeted improvements to help text, examples, and README. Not regeneration - surgical fixes.

5. **Scorecard is optional but quality assessment is not** - If the Go scorecard test exists and works, run it. If not, do a manual structured assessment against the 8 Steinberger dimensions. Either way, a score must be presented.

### Research Phase Detail (Phase 1)

The research phase should take 3-5 minutes and produce:

```markdown
# Research: <API> CLI

## Spec Discovery
- Official OpenAPI spec: <url> (or "none found")
- Source: <where it was found>
- Version: <API version>
- Endpoints: <count>

## Competitors
| Name | Stars | Language | Commands | Notable Features |
|------|-------|----------|----------|-----------------|
| ... | ... | ... | ... | ... |

## Auth Method
- Type: <api_key/oauth2/bearer_token>
- Header: <header name>
- Env var convention: <what competitors use>

## Demand Signals
- <any Reddit/HN/X posts asking for a CLI for this API>

## Recommendation
- Spec source: <OpenAPI vs write from docs>
- Target command count: <match or beat best competitor>
- Key differentiator: <what our CLI should do that competitors don't>
```

### Audit Phase Detail (Phase 3)

The audit phase should take 3-5 minutes and produce:

```markdown
# Audit: <API> CLI

## Command Comparison
- Our CLI: <N> commands across <M> resources
- Best competitor: <name> with <N> commands
- Gap: <what we're missing>

## Help Text Quality
- [ ] Descriptions are developer-friendly (not raw spec jargon)
- [ ] Examples use realistic values (not "string" or "0")
- [ ] Resource descriptions explain WHAT the resource is

## Agent-Native Checklist
- [ ] --json flag present
- [ ] --select flag present
- [ ] --dry-run flag present
- [ ] --stdin flag present
- [ ] Typed exit codes documented
- [ ] doctor command works

## Specific Fixes Needed
1. File: <path> - Change: <what to fix>
2. File: <path> - Change: <what to fix>
...
```

## Sync Strategy

To prevent installed/repo drift:

**Option A: Symlink** (preferred)
```bash
ln -sf ~/cli-printing-press/skills/printing-press/SKILL.md ~/.claude/skills/printing-press/SKILL.md
```

**Option B: Sync script** (add to existing workflow)
```bash
# In ~/cli-printing-press/scripts/sync-skill.sh
cp skills/printing-press/SKILL.md ~/.claude/skills/printing-press/SKILL.md
echo "Synced printing-press skill v$(grep 'version:' skills/printing-press/SKILL.md | awk '{print $2}')"
```

**Option C: Post-commit hook** (automatic)
Add to `.git/hooks/post-commit`:
```bash
if git diff --name-only HEAD~1 HEAD | grep -q "skills/printing-press/SKILL.md"; then
  cp skills/printing-press/SKILL.md ~/.claude/skills/printing-press/SKILL.md
fi
```

## Implementation Steps

1. **Rewrite SKILL.md to v1.0.0** with inlined phases, phase gates, and explicit instructions
2. **Sync to installed location** using symlink (Option A)
3. **Test with `/printing-press Notion`** - verify it takes 10-20 minutes and produces all 5 phase artifacts
4. **Test with `/printing-press Stripe`** - verify it works with a known-spec API too
5. **Commit and push**

## Success Metrics

- A `/printing-press <API>` run takes 10-20 minutes
- Research artifact is produced with competitor analysis
- Audit artifact is produced with specific fix list
- Generated CLI has polished help text and realistic examples
- Final report includes competitor comparison and quality score
- User can see each phase's progress as it happens (not silent)

## Risk: Over-engineering the enforcement

The skill could become so rigid that it breaks on edge cases (API with no competitors, already-perfect generation, etc.). Mitigation: each phase has a "nothing to fix" fast path that still produces the artifact but notes "no issues found."
