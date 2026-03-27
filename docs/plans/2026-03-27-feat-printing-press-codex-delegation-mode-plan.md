---
title: "feat: Codex Delegation Mode for Printing Press"
type: feat
status: completed
date: 2026-03-27
---

# feat: Codex Delegation Mode for Printing Press

## Overview

Add a `--codex` flag to `/printing-press` that offloads code generation to Codex CLI, keeping Claude Code for research, planning, scoring, and review. This cuts Opus token usage by ~60% per run while maintaining the same quality output.

The pattern already exists in two of Matt's skills - osc-nightnight and ce-work-beta. This plan applies the same proven delegation pattern to the printing press pipeline.

## Problem Statement

A full printing press run burns 45-85 minutes of Opus tokens across 8 phases. But only 3 phases actually generate code. The rest are research, analysis, and scoring - work that requires Opus-level reasoning. The code generation phases (Phase 4 GOAT Build, Phase 4.5 Dogfood fixes, Phase 5.7 Ship Loop fixes) are 60%+ of the token cost but are mostly mechanical: "write this Go function that does X, following this pattern from Y."

Codex CLI is free (or much cheaper) and handles well-scoped code generation tasks with clear acceptance criteria. The osc-nightnight skill already proves this works at scale - it generates 20+ PRs per night by delegating implementation to Codex while Claude handles planning and review.

## Proposed Solution

### Token Budget by Phase (Current vs Codex Mode)

| Phase | Current (Opus) | Codex Mode (Opus) | Codex Mode (Codex) | Savings |
|-------|---------------|-------------------|---------------------|---------|
| 0: Visionary Research | Opus | Opus | - | 0% |
| 0.5: Workflows | Opus | Opus | - | 0% |
| 0.7: Prediction Engine | Opus | Opus | - | 0% |
| 1: Deep Research | Opus | Opus | - | 0% |
| 2: Generate | Opus (1-2m) | Opus (cli call) | - | 0% (already fast) |
| 3: Non-Obvious Insight Review | Opus | Opus | - | 0% |
| **4: GOAT Build** | **Opus (5-10m)** | Opus (plan only) | **Codex (implement)** | **~80%** |
| **4.5: Dogfood fixes** | **Opus (10-20m)** | Opus (identify bugs) | **Codex (fix code)** | **~70%** |
| 4.7: Proof of Behavior | Opus | Opus | - | 0% |
| 5: Ship Readiness | Opus | Opus | - | 0% |
| **5.5: Live API Testing** | Opus | Opus (run tests) | - | 0% |
| **5.7: Ship Loop fixes** | **Opus (5-10m)** | Opus (plan fixes) | **Codex (apply fixes)** | **~80%** |

**Net savings: ~60% of total Opus token usage per run.**

### How It Works

**Claude Code stays the brain.** It does all research, planning, scoring, and review. It decides WHAT to build and WHERE. It writes the fix plans. It reviews the diffs.

**Codex does the hands.** It receives a focused prompt with: the specific task, exact files to modify, current code context, expected change in plain English, and conventions to follow. It writes the code, runs lint, and exits.

### Codex Invocation Pattern (from osc-nightnight)

```bash
# Environment: pass auth token for APIs that need it
export GH_TOKEN="$CODEX_GH_TOKEN"

# Build prompt with full context
CODEX_PROMPT="TASK: Write the sync.go workflow command for discord-cli.

FILES TO MODIFY:
- discord-cli/internal/cli/sync.go (create new)

CURRENT CODE (store.go UpsertMessage signature):
$(head -20 discord-cli/internal/store/store.go)

EXPECTED CHANGE:
Create a sync command that:
1. Lists guild channels via GET /guilds/{id}/channels
2. For each text channel, paginates messages with after= cursor
3. Calls db.UpsertMessage for each message
4. Tracks sync state per channel
5. Reports progress to stderr

CONVENTIONS:
- Package: cli
- Use cobra.Command pattern (see root.go)
- Error handling: return fmt.Errorf with context
- Progress: fmt.Fprintf(os.Stderr, ...)

CONSTRAINTS:
- Do NOT run git commit, git push, or git add
- Do NOT modify files outside discord-cli/internal/cli/sync.go
- Keep changes under 200 lines
- Run: go build ./... && go vet ./...

VERIFY: After making changes, run:
  cd discord-cli && go build ./... && go vet ./..."

# Delegate to Codex
cd ~/cli-printing-press && echo "$CODEX_PROMPT" | codex exec \
  --yolo \
  -c 'model_reasoning_effort="medium"' \
  -
```

### Phase-by-Phase Delegation

#### Phase 4: GOAT Build (biggest savings)

**Claude does:**
1. Reads the Phase 3 audit and identifies all fixes needed
2. Prioritizes: data layer first, then workflow commands, then scorecard fixes
3. For each task, assembles a Codex prompt with:
   - The specific file to create/modify
   - Current code context (relevant functions from store.go, helpers.go, root.go)
   - Expected behavior in plain English
   - Conventions from the existing codebase
4. Reviews the diff after Codex finishes
5. Runs Proof of Behavior verification
6. Runs scorecard to measure improvement

**Codex does:**
- Write the store.go domain tables (given the Phase 0.7 schema)
- Write each workflow command (sync, search, sql, stale, health, etc.)
- Write the README cookbook section
- Apply scorecard fixes (wiring dead flags, removing dead code)

**One Codex call per task, not one giant call.** Each task is scoped to 1-2 files, <200 lines. If Codex fails on a task, Claude falls back to writing it directly.

#### Phase 4.5/5.7: Fix Cycles

**Claude does:**
1. Identifies the bugs (from dogfood, live tests, or scorecard)
2. For each bug, writes a Codex prompt with: the bug, the file, the current code, the expected fix
3. Reviews the fix diff
4. Re-runs verification

**Codex does:**
- Apply each fix (typically 5-50 lines per fix)

### Mode Detection (Opt-In Only)

Codex mode is **off by default**. You must explicitly request it. This ensures quality is never accidentally degraded and you always know when Codex is being used.

```bash
# In SKILL.md: Codex mode is OPT-IN ONLY
# User must type "codex" in the arguments to enable
if echo "$ARGUMENTS" | grep -qiE '(^| )(--?codex|codex)( |$)'; then
  CODEX_MODE=true
  # Verify codex is actually installed
  if ! command -v codex >/dev/null 2>&1; then
    echo "Codex CLI not found - running in standard Opus mode."
    CODEX_MODE=false
  fi
else
  CODEX_MODE=false
fi
```

Usage:
```
/printing-press Discord             # standard Opus mode (default)
/printing-press Discord codex       # enable Codex delegation
/printing-press Discord --codex     # also works with flag syntax
```

### Environment Guard

Before delegating, check we're not already inside a Codex sandbox:

```bash
if [ -n "$CODEX_SANDBOX" ] || [ -n "$CODEX_SESSION_ID" ]; then
  echo "Already inside Codex sandbox - using standard mode."
  CODEX_MODE=false
fi
```

### Failure Handling

From ce-work-beta's proven pattern:
- On any Codex failure (rate limit, error, empty diff): fall back to Claude for that task
- Track consecutive failures
- After 3 consecutive Codex failures: disable delegation for remaining tasks
- Print: "Codex disabled after 3 consecutive failures - completing in standard mode."
- Never skip a task just because Codex failed

### Post-Codex Validation (from osc-nightnight)

After every Codex delegation:
1. `go build ./...` - must compile
2. `go vet ./...` - must pass
3. If format/lint available: auto-fix and commit "style: auto-format"
4. `git diff --stat` - must have non-empty diff
5. If any check fails: revert Codex changes, fall back to Claude

## What Does NOT Get Delegated

These phases require Opus-level reasoning and stay on Claude:

| Phase | Why Not Delegatable |
|-------|-------------------|
| 0: Visionary Research | Requires web search, competitor analysis, architectural judgment |
| 0.5: Workflows | Requires domain understanding to choose the right commands |
| 0.7: Prediction Engine | Core architectural decisions - which entities get tables, what indexes |
| 1: Deep Research | Web fetching, spec analysis, strategic positioning |
| 3: Non-Obvious Insight Review | Requires reading generated code and judging quality |
| 4.7: Proof of Behavior | Verification logic, not generation |
| 5: Ship Readiness Assessment | Scoring, reporting, judgment |
| 5.5: Live API Testing | Running and interpreting real API responses |

## Acceptance Criteria

- [ ] `--codex` flag enables Codex delegation in SKILL.md
- [ ] `--no-codex` flag forces Opus-only mode
- [ ] Auto-detection: if `codex` binary exists, enable by default
- [ ] Phase 4 GOAT Build delegates each task to Codex with scoped prompts
- [ ] Phase 4.5/5.7 fix cycles delegate each fix to Codex
- [ ] Post-Codex validation: go build + go vet + non-empty diff after every delegation
- [ ] Fallback: any Codex failure -> Claude handles that task
- [ ] Circuit breaker: 3 consecutive failures -> disable delegation
- [ ] Environment guard: skip delegation if already in Codex sandbox
- [ ] Quality: generated CLIs score the same Grade with or without --codex

## Success Metrics

| Metric | Target |
|--------|--------|
| Opus tokens per full run | -60% vs current |
| Quality Score (same API, codex vs no-codex) | Within 5 points |
| Phase 4 wall-clock time | Similar (Codex is fast) |
| Codex fallback rate | <20% of tasks |

## Files to Change

| File | Change |
|------|--------|
| `skills/printing-press/SKILL.md` | Add --codex/--no-codex mode detection, delegation logic in Phase 4/4.5/5.7 |

That's it - one file. The delegation is all in the skill instructions, not Go code. Codex is invoked via Bash, not a Go API.

## Dependencies & Risks

- **Dependency**: Codex CLI installed (`/opt/homebrew/bin/codex` - already on Matt's machine)
- **Risk**: Codex generates code that compiles but is subtly wrong (e.g., wrong API path). Mitigation: Proof of Behavior verification catches this.
- **Risk**: Codex prompt too vague -> empty diff or wrong files. Mitigation: paste actual current code in prompt, not descriptions.
- **Risk**: Codex model changes break the pattern. Mitigation: omit `-m` flag (let codex pick its default model).

## Sources & References

### Existing Codex Delegation Patterns
- osc-nightnight: `~/.claude/skills/osc-nightnight/SKILL.md` lines 1212-1264 - working production pattern with `codex exec --yolo`, prompt assembly, format/lint auto-fix, fallback to Claude
- ce-work-beta: `~/.claude/skills/ce-work-beta/SKILL.md` lines 422-483 - abstract specification with environment guards, temp file pattern, 3-failure circuit breaker
- Key difference: osc-nightnight pipes via `echo | codex exec -`, ce-work-beta writes to temp file. Either works; pipe is simpler.

### Internal
- Printing press SKILL.md: `~/cli-printing-press/skills/printing-press/SKILL.md` - Phase 4 is 5-10 min of Opus tokens, Phase 4.5 is 10-20 min
- Phase 4 code: lines 905-1033 (hand-writing workflow commands, data layer, scorecard fixes)
