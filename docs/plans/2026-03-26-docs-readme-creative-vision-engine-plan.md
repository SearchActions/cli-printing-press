---
title: "docs: Update README with Creative Vision Engine"
type: docs
status: completed
date: 2026-03-26
---

# docs: Update README with Creative Vision Engine

## Overview

The README accurately describes the 8-phase pipeline but is missing the Creative Vision Engine - the biggest upgrade to the printing press since launch. The NOI concept, domain archetype detection, auto-generated workflow/insight commands, the schema builder, and the 12-dimension scorecard are all shipped but invisible to anyone reading the README.

## What's Missing

1. **Non-Obvious Insight (NOI)** - The creative soul of every generated CLI. The formula. The examples. The Phase 0 gate. This is the headline feature and it's not mentioned anywhere.

2. **Domain Archetype System** - The profiler now classifies APIs into 8 archetypes (communication, PM, payments, infrastructure, content, CRM, developer-platform, generic). This drives automatic workflow generation. Not mentioned.

3. **Auto-Generated Workflow Commands** - PM APIs now get `stale`, `orphans`, `load` automatically from templates. Communication APIs get `channel-health`, `message-stats`. This is the Rung 4 breakthrough - not mentioned.

4. **Behavioral Insight Commands** - `health` (composite score) and `similar` (FTS5 duplicate detection) generate automatically. These are Rung 5 - not mentioned.

5. **Schema Builder** - Data gravity scoring now drives domain-specific SQLite table generation (extracted columns, FK indexes, FTS5 triggers). The store.go.tmpl uses real columns instead of JSON blobs. Not mentioned.

6. **12-Dimension Scorecard** - Added "Insight" dimension. Now 12 dimensions, 120 max. Grade thresholds unchanged. Not updated in README.

7. **The Creativity Ladder** - The conceptual framework (Rung 1-5) that explains why the press is different. Not mentioned.

## Proposed Changes to README.md

### Section 1: Update the tagline and intro

Change "Give it an API name. Get back the best CLI that has ever existed for it." to something that captures the NOI concept. The press doesn't just make CLIs - it sees the non-obvious insight in every API.

### Section 2: Add NOI section after "How It Works"

New section: "The Non-Obvious Insight" - explain the formula, show 4-5 examples, explain why it matters. This is the hook that makes people go "holy shit."

### Section 3: Add "The Creativity Ladder" section

Explain the 5 rungs:
1. API wrapper commands (always generated)
2. Output formatting (always generated)
3. Local persistence (conditional - sync, search, export)
4. Domain analytics (NEW - auto-generated from archetype templates)
5. Behavioral insights (NEW - auto-generated health, similar, etc.)

The press now reaches Rung 5 automatically. Before this update, it stopped at Rung 3.

### Section 4: Update "What Gets Generated" section

Add the new auto-generated commands:

| PM APIs | Communication APIs |
|---|---|
| `stale` - items with no updates in N days | `channel-health` - activity analysis |
| `orphans` - items missing assignment | `message-stats` - volume analytics |
| `load` - workload distribution | `audit-report` - audit log analysis |
| `health` - composite workspace score | `health` - composite score |
| `similar` - FTS5 duplicate detection | `similar` - duplicate detection |

### Section 5: Update Steinberger Bar table

Add 12th dimension: Insight (behavioral insight commands that see patterns humans miss).

### Section 6: Update Project Structure

Add new files:
- `internal/vision/insight.go` - NOI struct
- `internal/generator/entity_mapper.go` - Entity role mapping
- `internal/generator/schema_builder.go` - Data gravity scoring + schema generation
- `internal/generator/templates/workflows/` - PM workflow templates
- `internal/generator/templates/insights/` - Behavioral insight templates
- `skills/printing-press/references/noi-examples.md` - 10+ NOI examples

### Section 7: Add "Domain Archetypes" section

Brief explanation of how the profiler classifies APIs and what templates each archetype gets.

## Acceptance Criteria

- [ ] NOI concept explained with formula and 4+ examples
- [ ] Creativity Ladder (Rungs 1-5) explained
- [ ] Auto-generated workflow commands listed by archetype
- [ ] Behavioral insight commands (health, similar) described
- [ ] Steinberger Bar updated to 12 dimensions
- [ ] Project structure updated with new files/directories
- [ ] Domain archetypes listed with signals

## Files to Modify

- `README.md` - Main update target

## Implementation

Single file edit. Read current README, rewrite sections in place. Keep existing content where it's still accurate (phases, flags, exit codes). Add new sections for NOI, creativity ladder, domain archetypes, auto-generated commands.

### Tone

Match the existing README's energy - direct, confident, specific. Use the NOI examples to create a "holy shit" moment for the reader. The NOI formula should hit like a revelation, not a feature description.
