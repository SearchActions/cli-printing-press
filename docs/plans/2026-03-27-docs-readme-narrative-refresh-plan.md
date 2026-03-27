---
title: "docs: README Narrative Refresh"
type: docs
status: completed
date: 2026-03-27
---

# docs: README Narrative Refresh

## What to Check

The README was last fully rewritten on 2026-03-26. Some facts to verify and update:

### 1. discrawl command count (line 15)

Current claim: "12 commands and 551 stars"

**Actual:** 11 top-level commands (init, sync, tail, search, messages, mentions, sql, members, channels, status, doctor). `members` has 3 subcommands (list, show, search) and `channels` has 2 (list, show). Stars are now 568+.

**Fix:** Update to "11 commands and 568 stars" or "~15 commands (11 top-level + subcommands) and 568 stars". The exact number doesn't matter - the narrative point ("focused commands beat API wrappers") is still correct.

### 2. "323 Commands" framing (line 13)

The 323 number was from the Discord CLI generation (one command per endpoint). This number is correct for Discord's OpenAPI spec.

**Question:** Is this still the right lead narrative? The printing press now generates CLI + MCP + data layer + insights. The "323 vs 12" framing focuses on breadth-vs-depth which is still true, but the press now actually does BOTH (323 API wrappers + 12 discrawl-style workflow commands). The narrative should reflect that.

**Possible reframe:** "The printing press generates the 323 wrappers AND the 12 discrawl-style commands. You don't choose breadth or depth - you get both."

### 3. "Why CLIs (Not APIs, Not MCP)" section

This was already updated to "Why Not Just CLIs - CLIs + MCP" but check if the token savings claim ("100x fewer tokens") is still there and still makes sense now that we generate MCP servers too.

### 4. gogcli stars (line 49)

Current claim: "6.5K stars". Verify current count.

### 5. Phase list (already updated)

The phase diagram was updated to include 0.1, 4.7, 5.5, 5.7 - verify it matches the SKILL.md.

### 6. Scoring section name

Was renamed from "Steinberger Bar" to "Quality Scoring" - verify the README matches.

## Acceptance Criteria

- [x] discrawl command count is accurate (11 commands, verified)
- [x] discrawl star count is current (568+)
- [x] gogcli star count is current (6.5K+)
- [x] "323 vs 12" narrative reframed to "323 AND 11" - the press does both
- [x] MCP section reflects current architecture (updated in prior commit)
- [x] Phase diagram matches SKILL.md (updated in prior commit)
- [x] No "Steinberger" references in user-facing text (updated in prior commit)
- [x] All feature claims match what's actually shipped in the codebase
