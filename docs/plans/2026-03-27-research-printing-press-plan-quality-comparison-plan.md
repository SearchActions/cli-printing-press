---
title: "Plan Quality Comparison: Pre-Research vs Printing Press"
type: research
status: completed
date: 2026-03-27
---

# Plan Quality Comparison: Pre-Research vs Printing Press

## What We're Comparing

**Session A (Pre-research, /ce:plan):** You ran two separate plans before touching the printing press:
1. `research-github-cli-competitive-landscape-plan.md` (192 lines) - competitive landscape
2. `feat-printing-press-github-cli-plan.md` (150 lines) - the build plan for ghx

**Session B (Printing press run):** The press generated 5 artifacts across Phases 0-3:
1. `feat-github-cli-visionary-research.md` (Phase 0) - API identity, usage patterns, tool landscape
2. `feat-github-cli-power-user-workflows.md` (Phase 0.5) - 13 workflow ideas, scored
3. `feat-github-cli-data-layer-spec.md` (Phase 0.7) - entity classification, SQLite schema, sync strategy
4. `feat-github-cli-research.md` (Phase 1) - spec discovery, competitor deep dive
5. `fix-github-cli-audit.md` (Phase 3) - scorecard baseline, improvement plan

## Verdict: The Pre-Research Was Better

Your pre-research plans were sharper, more opinionated, and more strategically useful than the printing press's research phases. Here's the breakdown.

---

## Dimension 1: Competitive Understanding

### Session A (Pre-research): 9/10

The competitive landscape doc is genuinely excellent. It does things the press never attempted:

- **Three-lane framework** (Forge CLIs / Workflow Overlays / Git UX) - this is original strategic thinking, not just a list of tools
- **lazygit has 75k stars** - the press never found this. It's the most popular Git TUI and the press didn't even mention it
- **Jujutsu (jj) at 27.4k stars** - the press missed the entire "alternative Git UX" category
- **Graphite + GitButler + Git Town** - the press missed the workflow overlay category entirely
- **"gh is not being replaced, it's being layered over"** - this strategic insight shaped the whole build plan
- **Developer stacking patterns** - the table showing "what tool for what task" is immediately actionable
- **AI/agent shift framing** - GitHub reported 1M+ agent-created PRs, tools optimizing for agents as first-class users
- **Real user quotes with sources** - "ticking time bomb" on HN, specific issue numbers

### Session B (Press Phase 0 + Phase 1): 5/10

The press found the right direct competitors (gh 43.4k, gh-dash 11.2k, github-to-sqlite 462) but:

- **Missed the entire workflow overlay lane** - no mention of Graphite, GitButler, Git Town
- **Missed alternative Git UX** - no lazygit (75k!), no Jujutsu (27.4k), no tig
- **No strategic framework** - just a flat list of tools, not a mental model for the landscape
- **Weaker sentiment analysis** - found some gh CLI issues but missed the HN/Reddit threads about token security, multi-account pain
- **No "developer stacking" insight** - didn't understand that devs use 3-4 tools together

**Why the press was worse:** The skill's research steps are optimized for finding API-specific tools and specs, not for understanding market positioning. It searches for `"GitHub API" CLI tool` but never searches for `"GitHub CLI" alternative` or `git TUI` or `stacked PRs tool`. The pre-research used broader, more strategic search queries because it was answering "what do developers actually use?" not "what competes with the API wrapper?"

---

## Dimension 2: Product Vision

### Session A (Build plan): 8/10

The ghx build plan has a clear product thesis:

- **"Everything gh can't do"** - the comparison table (offline search, SQL queries, stale detection, review load, velocity) is the pitch
- **Naming analysis** - ghx, octo, ghub with trademark considerations
- **Weekend timeline** - Friday generate, Saturday polish, Sunday ship + HN post
- **HN headline already written** - "I built a GitHub CLI that finds stale PRs, review bottlenecks, and lets you SQL query your repos offline"
- **Scope discipline** - "Resist adding features. Ship the generated CLI with minimal hand-tuning."
- **gogcli as proof of model** - 6.5k stars doing this for Google APIs, GitHub audience is 10x larger

### Session B (Press artifacts): 6/10

The press produced more detailed technical specs but weaker vision:

- **No product narrative** - it never articulates WHY someone would use this over gh. Just lists features.
- **No naming** - the press calls it "github-cli" which is a terrible name (confused with gh)
- **No shipping plan** - no goreleaser, no Homebrew, no HN strategy
- **No "who is this for"** - the pre-research identified "engineering managers, open source maintainers" but the press buried it in API Identity metadata
- **Excessive detail on the wrong things** - 400 lines of SQLite schema but no README cookbook plan
- **No success metric** - "what would make someone star this?" is never asked

**Why the press was worse:** The skill is designed to build CLIs, not ship products. It optimizes for Steinberger scorecard points, not for "would a developer install this?" The pre-research was done by a human asking strategic questions.

---

## Dimension 3: Technical Depth

### Session A (Build plan): 4/10

Thin on technical detail by design. It trusts the press to figure out:

- Data layer design
- Sync strategy
- API endpoint validation
- SQLite schema
- FTS5 configuration

### Session B (Press artifacts): 9/10

This is where the press genuinely excels:

- **Phase 0.7 Data Layer Spec** is excellent - entity classification for 15 resources, data gravity scores, validated sync cursors, domain-specific SQLite schema, FTS5 config, 5 compound SQL queries
- **Endpoint validation** - confirmed `since` param exists on issues, `created` on workflow runs
- **Evidence-scored usage patterns** - weighted scoring system for community demand
- **API-validated workflows** - every workflow command traced to specific API endpoints with query params confirmed

**Why the press was better:** The skill is purpose-built for API analysis. It reads the OpenAPI spec, validates params, designs schemas. No human would spend 45 minutes classifying 15 entities by data gravity score. This is where automation shines.

---

## Dimension 4: Workflow Command Design

### Session A: 7/10

The build plan listed 16 workflow commands with clear one-line descriptions. Strong product instincts:

- `standup` - "What happened since yesterday?" (developers understand this instantly)
- `bottleneck` - "What's blocking the most PRs right now?"
- `review-load` - "Who's overloaded with review requests?"
- `similar` - "Find similar issues across repos"

These are user-empathy-driven names. They describe outcomes, not API operations.

### Session B: 6/10

The press produced 13 workflows with rigorous scoring (Frequency/Pain/Feasibility/Uniqueness) and API endpoint validation. But:

- Missed `standup`, `bottleneck`, `review-load`, `similar`, `orphans` - the most compelling workflow ideas from Session A
- Names are more technical: `actions-health` vs `bottleneck`, `contributors` vs `review-load`
- The scoring methodology is good but didn't surface the most compelling commands
- **Validated against the spec** is a genuine advantage - every workflow was confirmed possible

**Mixed verdict:** Session A had better product instincts for naming and user empathy. Session B had better technical validation. The ideal is both.

---

## Dimension 5: Execution Quality

### Session A: N/A (plan only, no execution)

### Session B: 3/10

The actual execution was poor:
- Tested 3.9% of commands at runtime
- Core feature (sync) is broken (404)
- Scorecard gaming instead of runtime testing
- Skipped 2 mandatory phases (4.5, 4.6)
- Declared "PASS" with broken data pipeline

---

## The Scoring Summary

| Dimension | Session A (Pre-research) | Session B (Press) | Winner |
|-----------|------------------------|-------------------|--------|
| Competitive understanding | 9/10 | 5/10 | **A by a mile** |
| Product vision | 8/10 | 6/10 | **A** |
| Technical depth | 4/10 | 9/10 | **B** |
| Workflow design | 7/10 | 6/10 | **A** (product) / **B** (validation) |
| Execution quality | N/A | 3/10 | Neither |
| **Overall** | **7/10** | **6/10** | **A** |

---

## What the Printing Press Should Learn From This

### 1. The press needs a strategic research phase, not just API research

The skill's Phase 0 searches for `"GitHub API" CLI tool` and `"GitHub API" automation`. It should ALSO search for:
- `"GitHub CLI" alternative OR replacement` (finds lazygit, jj, Graphite)
- `"stacked PRs" tool` (finds Graphite, Git Town)
- `git TUI terminal` (finds lazygit 75k, gitui, tig)
- `best git workflow tool 2026` (finds the developer stacking pattern)

**Fix:** Add "market landscape" searches to Phase 0 that go beyond the specific API. The competitive terrain isn't just other API wrappers - it's everything a developer might use instead.

### 2. The press needs a product vision step

Somewhere between Phase 0.7 (data layer) and Phase 2 (generate), there should be a "product pitch" gate:
- Who is this for? (one sentence)
- What's the HN headline?
- What's the comparison table vs the incumbent?
- What's the name?

The press jumps from technical research to code generation without ever articulating why someone would care.

**Fix:** Add Phase 0.8: Product Thesis. 10 minutes. Forces articulation of the pitch before any code is generated.

### 3. The press should consume external research

If the user already did research in a separate session (like you did), the press should READ those documents and incorporate them. Phase 0 should check `docs/plans/` for recent research on the same API.

**Fix:** At the start of Phase 0, glob for `docs/plans/*github*` or `docs/plans/*<api-name>*` and read any found. Feed insights into the research phases instead of starting from scratch.

### 4. Workflow naming should be user-empathy-driven, not API-driven

`standup` is a better name than `activity`. `bottleneck` is better than `actions-health`. `review-load` is better than `contributors`. The press names commands after API resources; a good product names commands after user outcomes.

**Fix:** In Phase 0.5, after generating workflow ideas, add a "naming pass" that asks: "If I were an engineering manager, what would I type to get this?" Map API operations to user intents.

### 5. The pre-research and the press are complementary, not competing

The ideal workflow:
1. **Pre-research** (human-driven, /ce:plan): strategic landscape, product vision, naming, shipping plan
2. **Printing press** (machine-driven): API analysis, schema design, endpoint validation, code generation, runtime testing

The pre-research answers "should we build this and what should it be?" The press answers "here's the code and proof it works."

**Fix:** The skill should explicitly say: "If you have a plan document from prior research, provide it. The press will use your product vision and focus on technical execution."

## Sources

- `docs/plans/2026-03-27-research-github-cli-competitive-landscape-plan.md` (Session A)
- `docs/plans/2026-03-27-feat-printing-press-github-cli-plan.md` (Session A)
- `cli-printing-press/docs/plans/2026-03-27-feat-github-cli-visionary-research.md` (Session B)
- `cli-printing-press/docs/plans/2026-03-27-feat-github-cli-power-user-workflows.md` (Session B)
- `cli-printing-press/docs/plans/2026-03-27-feat-github-cli-data-layer-spec.md` (Session B)
- `cli-printing-press/docs/plans/2026-03-27-feat-github-cli-research.md` (Session B)
