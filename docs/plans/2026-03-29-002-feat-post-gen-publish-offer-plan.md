---
title: "feat: Offer publish after CLI generation completes"
type: feat
status: completed
date: 2026-03-29
---

# Offer Publish After CLI Generation Completes

## Overview

After the `/printing-press` skill finishes generating a CLI (Phase 4 shipcheck or Phase 5 live smoke), offer the user the option to publish it via `/printing-press publish`. Also teach the publish skill to detect existing open PRs so re-generating and re-publishing a CLI updates the existing PR rather than failing on `gh pr create`.

## Problem Frame

PR #54 added a full publish workflow (`/printing-press publish`) that validates, packages, and opens a PR against `printing-press-library`. But the main generation skill (`/printing-press`) never mentions or offers it. Users who just generated a CLI have no prompt to publish.

Worse, if someone re-generates a CLI they previously published (e.g., after emboss or spec update), the publish skill will hit an existing branch `feat/<cli-name>` and — if they overwrite the branch — `gh pr create` fails because a PR already exists for that head ref. There's no path to update an existing PR.

## Requirements Trace

- R1. After a CLI passes shipcheck with "ship" or "ship-with-gaps", the main skill offers to publish
- R2. The offer is optional — never auto-publish
- R3. If the user accepts, invoke `/printing-press publish <cli-name>`
- R4. If the user declines, the skill ends as it does today
- R5. The offer appears after Phase 5 (live smoke) when it runs, since that's the final phase
- R6. The publish skill must detect existing open PRs for the same CLI and offer to update them
- R7. When updating an existing PR: force-push the new content to the same branch and update the PR description
- R8. The main skill's offer should surface whether an existing PR exists, so the user has context

## Scope Boundaries

- Skill changes only — no Go binary changes needed
- No changes to the `fullrun` pipeline (it calls `PublishWorkingCLI` programmatically, which publishes to the local library, not to GitHub)
- No changes to emboss mode (emboss improves an existing CLI; publishing is a separate step)

## Context & Research

### Relevant Code and Patterns

- `skills/printing-press/SKILL.md` — Phase 4 ends with ship/ship-with-gaps/hold. Phase 5 is optional live smoke. Neither mentions publish.
- `skills/printing-press-publish/SKILL.md` — Step 7 checks for existing *branches* but not existing *PRs*. If a branch exists, it asks overwrite vs timestamp. But `gh pr create` will fail if a PR already exists for that branch.
- `gh pr list --repo <repo> --head <branch> --state open --json number,url` — detects existing open PRs for a branch. This is the missing check.
- `gh pr edit <number> --body <body>` — updates an existing PR description without creating a new one.

### Institutional Learnings

- Published CLIs live at `~/printing-press/library/<api>-pp-cli/` (see `docs/solutions/best-practices/checkout-scoped-printing-press-output-layout-2026-03-28.md`)
- Validation must not mutate source directories (see `docs/solutions/best-practices/validation-must-not-mutate-source-directory-2026-03-29.md`)

## Key Technical Decisions

- **Offer placement:** After the final phase (Phase 4 or 5), not embedded within shipcheck. Rationale: publish is a post-generation handoff, not quality verification.
- **Use `/printing-press publish` skill, not CLI commands directly:** The publish skill handles interactive decisions (category, branch conflicts, PR description). The main skill just invokes it with the CLI name.
- **Gate on ship verdict:** Only offer on "ship" or "ship-with-gaps". A "hold" means the CLI isn't ready.
- **Detect existing PRs in the publish skill, not the main skill:** The publish skill already manages the clone and branch operations. Adding PR detection there keeps the logic centralized. The main skill just runs a lightweight `gh pr list` check to surface context in the offer text.
- **Update existing PR via force-push + `gh pr edit`:** When an open PR exists for `feat/<cli-name>`, overwrite the branch (as the skill already supports), force-push, then use `gh pr edit` to refresh the description. This avoids creating duplicate PRs and preserves any review comments on the existing PR.

## Open Questions

### Resolved During Planning

- **Should the main skill check for existing PRs?** Yes — a lightweight `gh pr list` check lets it say "update PR #42" vs "publish new" in the offer text. The check is fast (one API call) and gives the user important context before they commit.
- **What about merged PRs?** If a previous PR was merged (the CLI is already in the library), the user is publishing an update. The publish skill should treat this the same as a fresh publish — create a new branch and PR. The `--state open` filter on `gh pr list` naturally excludes merged PRs.
- **What if `gh` isn't authenticated when checking from the main skill?** Skip the PR check silently and show the generic offer. The publish skill will catch the auth issue in its own Step 1.

### Deferred to Implementation

- Exact wording of `AskUserQuestion` prompts — best tuned during implementation

## Implementation Units

- [x] **Unit 1: Teach the publish skill to detect and update existing PRs**

**Goal:** When an open PR already exists for `feat/<cli-name>`, update it (force-push + edit description) instead of failing on `gh pr create`.

**Requirements:** R6, R7

**Dependencies:** None

**Files:**
- Modify: `skills/printing-press-publish/SKILL.md`

**Approach:**
Add a new step between the current Step 6 (Managed Clone) and Step 7 (Branch, Commit, PR). This step:

1. After freshening the clone (Step 6), check for an existing open PR:
   ```
   gh pr list --repo mvanhorn/printing-press-library --head "feat/<cli-name>" --state open --json number,title,url
   ```
2. Store the result — either `EXISTING_PR_NUMBER` and `EXISTING_PR_URL`, or empty if no PR exists.
3. In Step 7, branch the flow:
   - **Existing PR found:** Always overwrite the branch (skip the "overwrite vs timestamp" question — the intent is clearly to update). After commit and force-push, run `gh pr edit <number> --body "<new description>"` to refresh the PR description. Display "Updated PR #N: <url>".
   - **No existing PR:** Current flow unchanged — ask about branch conflicts if any, then `gh pr create`.

The existing branch-conflict question ("overwrite vs timestamp") should still apply when there's an existing *branch* but *no* open PR (e.g., a previous PR was closed or merged). In that case, the user might want a fresh branch.

**Patterns to follow:**
- The existing Step 7 branch-conflict detection pattern
- The `gh pr create` call at the end of Step 7 — the `gh pr edit` call mirrors its structure

**Test scenarios:**
- Happy path (fresh): No existing branch or PR -> standard flow, `gh pr create`
- Happy path (update): Existing open PR #42 on `feat/notion-pp-cli` -> branch overwritten automatically, force-push, `gh pr edit #42`, displays "Updated PR #42"
- Edge case: Existing branch but no open PR (PR was merged or closed) -> asks overwrite vs timestamp as before, then `gh pr create` for the new PR
- Edge case: Existing branch AND closed PR -> treated as no-open-PR (the closed PR is irrelevant), standard branch-conflict question
- Error path: `gh pr list` fails (network, auth) -> fall back to current behavior (branch-conflict question, `gh pr create`)

**Verification:**
- The publish skill has a PR detection step before Step 7
- When an existing open PR is found, it force-pushes and runs `gh pr edit` instead of `gh pr create`
- When no open PR exists, the current flow is unchanged

---

- [x] **Unit 2: Add Phase 6 (Publish Offer) to the main skill**

**Goal:** After the final phase of generation, offer to publish with context about any existing PR.

**Requirements:** R1, R2, R3, R4, R5, R8

**Dependencies:** Unit 1 (publish skill must handle existing PRs before the main skill starts routing users to it for updates)

**Files:**
- Modify: `skills/printing-press/SKILL.md`

**Approach:**
Add a `## Phase 6: Publish` section after Phase 5. The section should:

1. **Gate on verdict:** If the shipcheck verdict from Phase 4 was "hold", skip this phase entirely. Only proceed for "ship" or "ship-with-gaps".

2. **Check for existing PR** (lightweight, may fail silently):
   ```
   gh pr list --repo mvanhorn/printing-press-library --head "feat/<cli-name>" --state open --json number,url --jq '.[0]'
   ```
   If `gh` isn't authenticated or the call fails, continue without PR context.

3. **Present the offer via `AskUserQuestion`:**

   - If an existing open PR was found:
     > "There's an open publish PR for <cli-name> (#N). Want to update it with this regenerated version?"
     > 1. **Yes — update PR #N** (re-validate, re-package, and push to the existing PR)
     > 2. **No — I'm done**

   - If no existing PR:
     > "<cli-name> passed shipcheck. Want to publish it to the printing-press-library?"
     > 1. **Yes — publish now** (validate, package, and open a PR)
     > 2. **No — I'm done**

   - If verdict was "ship-with-gaps", prepend a note: "Note: shipcheck found minor gaps (see shipcheck report above)."

4. **If accepted:** Invoke `/printing-press publish <cli-name>`. The publish skill handles everything from there — name resolution, category, validation, packaging, git ops, and PR creation/update.

5. **If declined:** End normally.

**Patterns to follow:**
- Phase 1.5's `AskUserQuestion` gate pattern
- The "ship/ship-with-gaps/hold" verdict from Phase 4

**Test scenarios:**
- Happy path (fresh publish): "ship" verdict, no existing PR -> offer "publish now" -> user accepts -> `/printing-press publish <cli-name>` invoked
- Happy path (update): "ship" verdict, existing open PR #42 -> offer "update PR #42" -> user accepts -> `/printing-press publish <cli-name>` invoked (publish skill handles the update)
- Ship-with-gaps: offer includes gap note, user accepts -> publish invoked
- Decline: user says no -> skill ends normally
- Hold verdict: no offer shown, skill ends at Phase 4/5
- After live smoke: Phase 5 runs -> publish offer appears after Phase 5
- `gh` not available: PR check fails silently -> generic "publish now" offer shown (publish skill catches auth in its own Step 1)

**Verification:**
- The main skill has a Phase 6 section with `AskUserQuestion`
- The offer text varies based on whether an existing PR was found
- "hold" verdict skips the phase entirely
- The phase invokes `/printing-press publish <cli-name>`, not CLI commands directly

## System-Wide Impact

- **Interaction graph:** The main skill now has a dependency on the publish skill at the end of its flow. The publish skill remains independently invocable.
- **Unchanged invariants:** The publish skill's standalone invocation (`/printing-press publish`) works identically. The main skill's Phases 0-5 are untouched. The Go binary is not modified.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| `gh` not authenticated during Phase 6 PR check | Fail silently, show generic offer. Publish skill catches auth in Step 1. |
| User runs `/printing-press publish` standalone after this change | Still works — Unit 1's PR detection applies regardless of how the skill is invoked |
| Existing PR has review comments the user wants to preserve | Force-push + `gh pr edit` preserves the PR thread. Only the description and branch content change. |

## Sources & References

- Related PRs: #54 (publish skill)
- Publish skill: `skills/printing-press-publish/SKILL.md`
- Main skill: `skills/printing-press/SKILL.md`
