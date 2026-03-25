---
title: "Fix Printing Press v1.1: Steinberger Analysis, Deeper Research, Complex Body Fields"
type: fix
status: active
date: 2026-03-25
---

# Fix: Printing Press v1.1 - Steinberger Loop + Depth

## Problem Statement

The v1.0.0 SKILL.md produced a 6m16s Notion run (vs 1m42s before). It followed the 5 phases and produced artifacts. But the artifacts are shallow:

1. **Research artifact** is surface-level - lists competitors but doesn't analyze their quality, maintenance status, or user experience. No strategic justification for why this CLI should exist.

2. **Audit artifact** fixes cosmetics (command names, help text) but ignores the strategic gap it identifies - 22 commands vs competitor's 39. It renamed commands but didn't add any.

3. **No Peter Steinberger analysis** - The scorecard at the end is a self-graded checklist, not a rigorous comparison against the best CLI in the world (gogcli). No dimension-by-dimension teardown of what a 10/10 looks like.

4. **Complex body fields punted** - The generator says "skipping body field: complex type not supported" and the skill just accepts it. The LLM brain should handle these - read the spec, understand what fields like `children`, `properties`, `filter` look like, and generate `--json-body` or `--stdin` patterns.

5. **Steinberger report only at the end** - Should happen TWICE: once during audit (to find gaps) and again after fixes (to prove improvement). The delta is the proof of work.

## What the User Wants

The printing press should produce the GOAT CLI for any API. Not just "it compiles" or "the help text is nice." The best CLI that has ever existed for that API. This means:

- Deep competitive research (not just star counts)
- Understanding why users would choose this CLI
- Fixing real gaps (missing commands, not just renames)
- Two Steinberger analyses showing before/after improvement
- Complex body fields handled (not punted)

## Root Causes

### RC1: Research phase lacks depth requirements

The skill says "WebSearch for competitors" but doesn't require:
- Reading competitor source code (not just README)
- Checking maintenance status (last commit, open issues)
- Analyzing actual user complaints about existing tools
- Understanding what makes the best CLI in the world (gogcli) the best

### RC2: Audit phase lacks strategic thinking

The skill says "compare command count vs competitors" but doesn't require:
- Planning how to CLOSE the gap (not just identify it)
- Prioritizing fixes by impact (agent-native features > cosmetic renames)
- Generating a Steinberger-grade analysis with specific scores and "what would make this a 10"

### RC3: No "make it the GOAT" phase

After fixing cosmetics, there should be a phase that asks: "What would make this THE BEST CLI for this API, period?" This is where:
- Missing high-value commands get added
- Complex body fields get `--json-body` support
- Convenience features (open in browser, URL-or-ID input) get planned
- The CLI differentiates itself beyond just "it has --json"

### RC4: Steinberger analysis is manual self-grading

The skill says "score 0-10 on each dimension" but Claude just picks numbers. There's no:
- Comparison against a specific reference CLI (gogcli)
- "What would a 10 look like?" for each dimension
- Action items to get from current score to 10
- Before/after delta showing improvement

### RC5: Complex body fields unconditionally punted

The Go generator says "complex type not supported as CLI flag" and the skill accepts it. But the LLM brain can:
- Add `--json-body` flag that accepts raw JSON
- Add `--stdin` support (already exists but not documented per-command)
- Generate example JSON for common operations
- Add convenience wrappers for the most common complex bodies

## Proposed Solution

### Fix 1: Deeper research with nightnight-style gating

Research phase must produce:
- Competitor maintenance status (last commit date, open issue count)
- Actual user complaints about existing CLIs (from GitHub issues, Reddit)
- gogcli feature checklist comparison (what does the 10/10 CLI have?)
- Strategic differentiator beyond "agent-native"

**Gate:** Research artifact must have >= 5 specific competitor features analyzed, >= 2 user complaint quotes, and a "why this CLI should exist" section.

### Fix 2: Steinberger analysis runs TWICE

**First Steinberger (Phase 3: AUDIT):**
- Score each of 8 dimensions 0-10
- For each dimension: "What does a 10 look like?" (reference gogcli or best-in-class)
- For each dimension: "What specific changes would get us to 10?"
- Total score = baseline

**Second Steinberger (Phase 5: SCORE):**
- Re-score after fixes
- Show delta: "Before: 52/80. After: 68/80. +16 points."
- For each dimension that improved: "Changed from X to Y because we did Z"
- For dimensions still low: "Remaining gap: [what's needed in a future pass]"

### Fix 3: "Make it the GOAT" phase between audit and fix

New Phase 3.5: GOAT PLANNING
- Read the Steinberger analysis from Phase 3
- For every dimension scoring < 8: write a specific improvement plan
- For the Breadth dimension: identify top 5 commands to ADD (not rename)
- For complex body fields: plan `--json-body` support for top endpoints
- Prioritize: what 5 changes would have the biggest impact on the score?

### Fix 4: Complex body field handling

When the generator warns "skipping body field: complex type":
1. Note which fields were skipped and for which endpoints
2. In the audit, check if these fields are critical for the endpoint's purpose
3. In the fix phase, add `--json-body` flag support where needed
4. Add example JSON in the command's Example field showing what to pipe via `--stdin`

### Fix 5: Nightnight-style anti-shortcut patterns

Adopt these specific patterns from osc-nightnight:

1. **Executable gates, not prose gates** - Each phase gate runs a bash check, not just "verify the artifact exists"
2. **Banned phrases** - If the skill output contains "This is a limitation of the generator" or "Complex types not supported" without attempting to fix it, the phase fails
3. **Observable impact verification** - Before presenting the final report, diff the generated CLI against its pre-fix state. If fewer than 10 files changed, the fix phase was too shallow.

## Acceptance Criteria

- [ ] Research artifact has competitor maintenance status (last commit, issues)
- [ ] Research artifact has >= 2 user complaint quotes from GitHub issues or Reddit
- [ ] Research artifact has "why this CLI should exist" strategic section
- [ ] Steinberger analysis runs in Phase 3 (AUDIT) with per-dimension "what would 10 look like"
- [ ] Phase 3.5 (GOAT PLANNING) identifies top 5 commands to ADD and 5 highest-impact improvements
- [ ] Complex body fields have `--json-body` or `--stdin` examples (not just "skipped")
- [ ] Phase 4 (FIX) adds at least some new commands, not just renames
- [ ] Steinberger analysis runs again in Phase 5 (SCORE) with before/after delta
- [ ] Final report shows improvement: "Before: X/80. After: Y/80. +Z points."
- [ ] A `/printing-press Notion` run takes 12-20 minutes
- [ ] Artifacts are saved to `docs/plans/` for later review

## Implementation: SKILL.md v1.1.0 Changes

### Phase 1: RESEARCH - Add depth requirements

Add to the research phase:

```markdown
### Step 1.2b: Deep competitor analysis

For the TOP 2 competitors (by stars):

1. **WebFetch** their GitHub repo page - check:
   - Last commit date
   - Open issue count
   - Number of contributors
   - Is it actively maintained?

2. **WebFetch** their GitHub issues page - look for:
   - User complaints about missing features
   - Requests for specific functionality
   - Pain points that users report

3. Record at least 2 specific user quotes/complaints.

### Step 1.3b: Strategic justification

Answer: "Why should this CLI exist when [best competitor] already has [N] stars?"

The answer must be SPECIFIC, not just "agent-native." Examples:
- "4ier/notion-cli hasn't been updated in 6 months and doesn't support the 2025-09-03 API"
- "No existing CLI supports --json + --select + --dry-run for agent workflows"
- "Users on HN are asking for X which no CLI provides"
```

### Phase 3: AUDIT - Add Steinberger analysis

Replace the current audit with a two-part structure:

```markdown
### Step 3.7: First Steinberger Analysis

Score against gogcli (the 10/10 reference). For each dimension:

| Dimension | Our Score | gogcli Score | What Would 10 Look Like | Gap |
|-----------|-----------|-------------|------------------------|-----|
| Output modes | X/10 | 10/10 | [specific features] | [what we're missing] |
| Auth | X/10 | 9/10 | [specific features] | [what we're missing] |
...

Total baseline: X/80

### Step 3.8: Write GOAT improvement plan

For every dimension scoring < 8:
1. What specific changes would raise it to 8+?
2. Which are achievable in Phase 4 (fix) vs future work?
3. Prioritize the top 5 highest-impact improvements

For the Breadth dimension specifically:
- List the top 5 commands to ADD (not rename)
- For each: which endpoint, what it does, why it matters

For complex body fields:
- List the top 3 endpoints where --stdin examples would be most valuable
- Write the example JSON that users would pipe in
```

### Phase 4: FIX - Add GOAT improvements

```markdown
### Step 4.1b: Add new commands (from GOAT plan)

For each new command identified in Step 3.8:
1. Add a new command file or entry
2. Include --stdin support with example JSON
3. Include realistic examples

### Step 4.1c: Add --stdin examples for complex body fields

For the top 3 endpoints with complex bodies:
1. Write example JSON showing what to pipe
2. Add the example to the command's Example field:
   `echo '{"children":[...]}' | notion-cli blocks children append <id> --stdin`
```

### Phase 5: SCORE - Add before/after delta

```markdown
### Step 5.1: Second Steinberger Analysis

Re-score all 8 dimensions after fixes. Show the delta:

| Dimension | Before | After | Delta | What Changed |
|-----------|--------|-------|-------|-------------|
| Output modes | X/10 | Y/10 | +Z | Added [feature] |
...

Total: Before X/80 -> After Y/80 (+Z points)

### Step 5.2: Remaining gaps

For dimensions still < 8:
- What would it take to reach 8+?
- Is this a generator limitation or a fixable gap?
- Tag as "future work" with specific plan
```

## Sources

- Notion run transcript from 2026-03-25 (6m16s, 21 commands, 66/80 manual score)
- Research artifact: `docs/plans/2026-03-25-feat-notion-cli-research.md` (shallow)
- Audit artifact: `docs/plans/2026-03-25-fix-notion-cli-audit.md` (cosmetic fixes only)
- OSC nightnight skill patterns: executable gates, banned phrases, checkpoint-first recovery
- gogcli as the 10/10 reference CLI: github.com/grafana/grafana-openapi-client-go (6.5k stars)
