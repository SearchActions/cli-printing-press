---
title: "Fix Skill: Always Research, Polish, and Score"
type: fix
status: active
date: 2026-03-25
---

# Fix: Always Research, Polish, and Score

## The Problem

`/printing-press Notion` found an OpenAPI spec and jumped straight to generate. No research, no competitor analysis, no polish, no scorecard. The skill only does the brain work when there's NO spec. But the brain work should happen EVERY time.

## The Fix

Change the skill flow from:

```
Found spec? -> YES -> generate -> done (dumb)
            -> NO  -> research -> write spec -> generate -> polish -> score (smart)
```

To:

```
Always: research competitors -> find/write spec -> generate -> polish -> score -> report
```

Every run gets the full brain treatment regardless of how we got the spec.

## What to Change in SKILL.md

### Step 1 stays: Parse intent and find spec

No change. Still search for OpenAPI spec first.

### NEW Step 1.5: Research competitors (ALWAYS runs)

Even when we have an OpenAPI spec, Claude Code should:

```
WebSearch: "<api-name> cli" site:github.com
```

For each competitor found:
- Note name, stars, language, last updated
- WebFetch their README if < 5 competitors
- Count their commands
- Identify what they have that we should match

Report to user: "Found N competing CLIs. Best: X (Y stars). They have Z commands."

### Step 2-3 stay: Write spec or use OpenAPI

No change to spec acquisition.

### Step 4 becomes MANDATORY (was optional): Polish

After generating, Claude Code ALWAYS:

a. Reads the generated `--help` output and checks command count
b. If help descriptions are jargon-heavy, rewrites them via Edit
c. If README description is generic spec text, rewrites it
d. If competitor has features we're missing, notes it in the report

### Step 5 becomes MANDATORY: Score

Always run the scorecard. Always report the grade.

### Step 6 enhanced: Report includes comparison

```
Generated notion-cli: 7 resources, 20 commands
Steinberger score: 65/90 (72%, Grade B)

Competitors:
- 4ier/notion-cli (87 stars, Go) - 15 commands
- notion-cli (bash, 2025 API) - 12 commands

We beat competitors on: commands (20 vs 15), auth, output formats
We're missing: database query support (complex body params were skipped)
```

## Implementation

One file change: `skills/printing-press/SKILL.md`

Move Steps 1.5 (research), 4 (polish), and 5 (score) from conditional to mandatory in ALL workflows.
