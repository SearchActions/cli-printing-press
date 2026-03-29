---
title: "feat: Offer to install browser-use/agent-browser when sniff is accepted"
type: feat
status: completed
date: 2026-03-29
---

# feat: Offer to install browser-use/agent-browser when sniff is accepted

## Overview

When the sniff gate fires in Phase 1.7 and the user approves sniffing, Step 1 (Detect capture tool) currently falls through silently to "manual HAR" if neither browser-use nor agent-browser is installed. Instead, it should offer to install them right there — the user already said yes to sniffing, so the intent is clear.

## Problem Frame

The current flow has a UX gap:

1. Skill says: "No OpenAPI spec found. Want me to sniff the live site?"
2. User says: "Yes"
3. Skill silently discovers neither tool is installed
4. Skill says: "Ask the user for a manual HAR from browser DevTools"

The user expected the skill to *do* the sniff. The fallback to manual HAR feels like a broken promise. The skill should detect the missing tools and offer to install them before giving up.

This matches the pattern already established in the setup contract (lines 114-124) where the skill detects a missing `printing-press` binary and provides install instructions.

## Requirements Trace

- R1. When sniff is approved and no capture tool is installed, offer to install browser-use (preferred) or agent-browser before falling back to manual HAR
- R2. Installation should be automatic with user confirmation, not just "here's a command, go run it yourself"
- R3. After successful install, proceed directly to the capture step (no re-asking the sniff question)
- R4. If install fails or user declines, fall back to manual HAR with clear messaging
- R5. The sniff offer prompt itself should mention that tools will be installed if needed, so the user knows what they're agreeing to

## Scope Boundaries

- Only changes `skills/printing-press/SKILL.md` — no Go code changes
- Does not change the detection preference order (browser-use > agent-browser > manual)
- Does not change how the capture tools are used once installed
- Does not add browser-use/agent-browser as dependencies of printing-press itself

## Context & Research

### Relevant Code and Patterns

- **Setup contract install pattern** (SKILL.md lines 114-124): Detects missing `printing-press`, provides `go install` command, exits if not found. This is the closest existing pattern but it *blocks* — the skill can't continue without the binary. For capture tools, the skill *can* continue (manual HAR fallback), so the install should be offered, not required.
- **Current sniff gate detection** (SKILL.md lines 347-363): Three-way detection with silent `SNIFF_BACKEND="manual"` fallback.
- **Sniff offer prompts** (SKILL.md lines 328-343): The "Sniff as enrichment" and "Sniff as primary" user prompts. These currently don't mention tool installation.

### Install Commands

- **browser-use**: `pip install browser-use` or `uv pip install browser-use` (Python, heavy deps — all LLM SDKs). Also `uvx browser-use` works without global install.
- **agent-browser**: `npm install -g agent-browser` or `brew install agent-browser` (Rust binary, lightweight)

### Why browser-use is preferred

Per the deep dive (docs/plans/2026-03-29-004): browser-use HAR includes response bodies natively. agent-browser HAR omits them, requiring the enriched capture workaround. browser-use also runs autonomously (no Claude token burn for browsing).

## Key Technical Decisions

- **Offer install inline, don't just print instructions**: The user said yes to sniffing — they want results, not homework. Use `AskUserQuestion` to confirm, then run the install command. This matches the "do the thing" philosophy of the skill.
- **Try browser-use first, offer agent-browser as lighter alternative**: browser-use is preferred (better HAR) but heavy (Python + all LLM SDKs). Some users may prefer the lighter agent-browser. The install prompt should present both options.
- **Update sniff offer prompts to set expectations**: The "Want me to sniff?" prompts should mention that capture tools will be installed if needed, so the user isn't surprised by an install step after saying yes.

## Open Questions

### Resolved During Planning

- **Should the skill auto-install without asking?** No — installing Python packages or npm globals should always get user confirmation. The install is offered, not forced.
- **What if the user has `uv` but not `pip`?** Prefer `uv pip install` when `uv` is available, fall back to `pip install`. For browser-use, `uvx browser-use` also works as a one-shot without global install.

### Deferred to Implementation

- **Exact `browser-use` CLI flag for HAR recording**: The skill already handles this (line 389-397) with a fallback to a Python wrapper. No change needed here.

## Implementation Units

- [ ] **Unit 1: Replace silent fallback with install-offer flow in Step 1**

  **Goal:** When neither browser-use nor agent-browser is detected, offer to install one instead of silently falling back to manual HAR.

  **Requirements:** R1, R2, R3, R4

  **Dependencies:** None

  **Files:**
  - Modify: `skills/printing-press/SKILL.md` (Phase 1.7, Step 1: Detect capture tool, lines ~347-363)

  **Approach:**
  - Replace the `else SNIFF_BACKEND="manual"` branch with an install-offer flow
  - Use `AskUserQuestion` with options:
    1. **Install browser-use (Recommended)** — "Autonomous agent, HAR includes response bodies. Requires Python. Install via `pip install browser-use`"
    2. **Install agent-browser** — "Lighter install, Claude drives the browsing. Install via `npm install -g agent-browser`"
    3. **Skip — I'll provide a HAR manually** — "Export HAR from browser DevTools yourself"
  - If user picks an install option:
    - Detect package manager (`uv` vs `pip` for browser-use, `npm` vs `brew` for agent-browser)
    - Run the install command
    - Re-run the detection check to confirm
    - If install succeeds, set `SNIFF_BACKEND` accordingly and proceed to Step 2
    - If install fails, show the error and fall back to manual HAR
  - If user picks manual, proceed as today

  **Patterns to follow:**
  - Setup contract install pattern (SKILL.md lines 114-124) for the detect-then-install flow
  - Phase 0 API Key Gate pattern (lines 169-195) for the `AskUserQuestion` with options pattern

  **Test scenarios:**
  - Happy path: Neither tool installed, user picks browser-use, `pip install` succeeds, sniff proceeds with browser-use
  - Happy path: Neither tool installed, user picks agent-browser, `npm install -g` succeeds, sniff proceeds with agent-browser
  - Happy path: User picks "skip", manual HAR flow runs as before
  - Edge case: `pip` not found but `uv` is — uses `uv pip install browser-use`
  - Error path: Install command fails — clear error message, falls back to manual HAR option
  - Edge case: One tool already installed but the other isn't — detection picks the installed one before reaching install-offer (no change to existing behavior)

  **Verification:**
  - When sniff is approved and no capture tool exists, the user sees install options instead of silent fallback to manual HAR

- [ ] **Unit 2: Update sniff offer prompts to mention tool installation**

  **Goal:** The "Want me to sniff?" prompts should set expectations that capture tools may need to be installed.

  **Requirements:** R5

  **Dependencies:** Unit 1

  **Files:**
  - Modify: `skills/printing-press/SKILL.md` (Phase 1.7, "Sniff as enrichment" and "Sniff as primary" prompts, lines ~324-343)

  **Approach:**
  - Update the sniff offer prompts to include a note like: "I'll check if you have browser-use or agent-browser installed and set them up if needed."
  - Keep it brief — one line added to the existing prompt, not a paragraph
  - The "Yes — sniff" option descriptions should mention: "I'll install capture tools if needed"

  **Patterns to follow:**
  - Existing prompt style in Phase 1.7 (concise, informative, not overwhelming)

  **Test scenarios:**
  - Happy path: Sniff offered as primary discovery — prompt mentions tool install
  - Happy path: Sniff offered as enrichment — prompt mentions tool install

  **Verification:**
  - Both sniff offer prompts mention that capture tools will be installed if needed

## System-Wide Impact

- **Interaction graph:** Only changes the skill markdown — no Go code, no binary changes, no new dependencies
- **Error propagation:** Install failures are caught and fall back to manual HAR — never blocks the pipeline
- **State lifecycle risks:** None — install happens before any capture state is created
- **API surface parity:** No CLI changes
- **Unchanged invariants:** Detection preference order (browser-use > agent-browser > manual) stays the same. Capture workflows (Steps 2a, 2b) are untouched. The `printing-press sniff` command is untouched.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| `pip install browser-use` pulls heavy deps (all LLM SDKs) and user doesn't expect it | The install option description mentions "Requires Python" and the agent-browser alternative is offered as lighter |
| `npm install -g` requires sudo on some systems | Use `npm install -g` without sudo first; if it fails, suggest the user fix npm permissions or use `brew install agent-browser` |
| Install succeeds but tool doesn't work (e.g., Chrome not available for browser-use) | The capture step itself will fail with a clear error — this is the existing behavior for broken installs |

## Sources & References

- **Current SKILL.md sniff gate:** `skills/printing-press/SKILL.md` lines 309-456
- **Setup contract install pattern:** `skills/printing-press/SKILL.md` lines 114-124
- **browser-use deep dive:** `docs/plans/2026-03-29-004-refactor-browser-use-vs-agent-browser-vs-printing-press-plan.md`
- **browser-use sniff backend plan:** `docs/plans/2026-03-29-005-feat-browser-use-sniff-capture-backend-plan.md`
- browser-use install: `pip install browser-use` or `uv pip install browser-use`
- agent-browser install: `npm install -g agent-browser` or `brew install agent-browser`
