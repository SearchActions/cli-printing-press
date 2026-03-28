---
title: "fix: Require interactive API key consent in printing-press skill"
type: fix
status: completed
date: 2026-03-28
---

# fix: Require interactive API key consent in printing-press skill

## Overview

The printing-press skill's Phase 0 silently proceeds when no API key is found, and never explains how a found key will be used. This must become an interactive gate: always stop, always explain, always get consent.

## Problem Frame

During a Linear CLI build, the skill detected `LINEAR_API_KEY` was empty and silently continued. The user had to interrupt and manually provide it. When a key *is* found, the skill should explain its usage (read-only smoke testing) and confirm before proceeding. The current instruction "Do not block the build on token collection" directly causes this.

## Requirements Trace

- R1. When no API key is detected, the skill MUST stop and ask the user to provide one
- R2. When an API key IS detected, the skill MUST explain how it will be used (read-only live smoke testing only) and ask for confirmation before proceeding
- R3. If the user cannot provide a key (R1) or declines usage (R2), warn that live smoke testing will be skipped but allow the build to continue
- R4. The key must never be used for write operations — this should be stated in the consent explanation

## Scope Boundaries

- Only the SKILL.md Phase 0 token handling section changes
- No Go code changes — the runtime already treats API keys as optional
- No changes to Phase 5 (Optional Live Smoke) — it already gates on token availability

## Context & Research

### Relevant Code and Patterns

- `skills/printing-press/SKILL.md` lines 182-192: Current Phase 0 token detection block
- `internal/pipeline/runtime.go` line 24-25: `APIKey` is already optional in `VerifyConfig`
- `internal/pipeline/runtime.go` line 81-84: Runtime already handles missing key gracefully (mock mode)
- `internal/pipeline/runtime.go` line 342-346: Write commands are never tested live

## Key Technical Decisions

- **Interactive gate, not hard block**: The user confirmed that if they can't provide a key, the build should warn and continue rather than refuse entirely. This preserves the ability to generate CLIs without API access while making the consent explicit.
- **Use AskUserQuestion**: The skill already has `AskUserQuestion` in its allowed-tools list. The consent flow should use it for a clean interactive experience.

## Implementation Units

- [x] **Unit 1: Rewrite Phase 0 token handling in SKILL.md**

**Goal:** Replace the non-blocking token detection with an interactive consent flow

**Requirements:** R1, R2, R3, R4

**Dependencies:** None

**Files:**
- Modify: `skills/printing-press/SKILL.md`

**Approach:**
Replace lines 182-192 (the token detection block through "Do not block the build on token collection") with new instructions that define three paths:
1. **Key found** → Explain to user: "Found `<ENV_VAR>` in your environment. This will be used for read-only live smoke testing in Phase 5 only — no write operations. OK to use it?" → If yes, proceed with key. If no, proceed without key + warning.
2. **Key not found** → Ask user: "No API key detected for `<API>`. Provide one now for live smoke testing, or continue without it?" → If provided, proceed with key. If skipped, proceed without key + warning.
3. **Warning path** → When continuing without a key, state clearly: "Live smoke testing (Phase 5) will be skipped. The CLI will still be generated and verified against mock responses."

Remove the line "Do not block the build on token collection" and replace with "Always stop to ask about the API key before proceeding to Phase 1."

**Patterns to follow:**
- Phase 1.5 already has a mandatory stop gate ("REQUIRES USER APPROVAL before proceeding to Phase 2") — mirror that pattern
- The skill already uses `AskUserQuestion` in its allowed-tools

**Test scenarios:**
- Happy path: Key found in env → skill explains usage and asks consent → user approves → build proceeds with key available for Phase 5
- Happy path: Key not found → skill asks user to provide one → user provides it → build proceeds with key
- Edge case: Key found → user declines usage → build proceeds without key, warning displayed, Phase 5 skipped
- Edge case: Key not found → user says they don't have one → build proceeds without key, warning displayed, Phase 5 skipped
- Edge case: Key not found → user provides an invalid-looking key (no validation needed, just accept it — Phase 5 will catch failures)

**Verification:**
- Reading the updated SKILL.md, the instructions clearly require stopping at Phase 0 for API key consent
- The phrase "Do not block the build on token collection" no longer appears
- All three paths (found+consent, not-found+provided, continue-without) are explicitly documented

## System-Wide Impact

- **Phase 5 dependency**: Phase 5 already checks for token availability, so no change needed there
- **Codex delegation mode**: The consent flow happens before any delegation, so Codex mode is unaffected
- **Emboss mode**: Emboss also runs Phase 0, so it will also get the consent flow — this is correct behavior

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| User pastes API key in plaintext into chat (as happened in the session log) | The skill should instruct: "You can set it with `export ENV_VAR=...` or paste it here" — accept either path. Note: the session log key is already exposed; user should rotate it. |
