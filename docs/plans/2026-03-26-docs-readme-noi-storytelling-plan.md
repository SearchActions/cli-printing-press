---
title: "docs: Rewrite README around the NOI as the hero, with discrawl and gogcli stories"
type: docs
status: completed
date: 2026-03-26
---

# docs: Rewrite README around the NOI as the hero

## The Narrative Problem

The README currently has two competing heroes: the agent/token economics section and the NOI section. The agent stuff is the *obvious* insight - everyone knows CLIs are better for agents. The NOI is the *non-obvious* insight about the printing press itself. The README should practice what it preaches.

The user's key insight: "Just making a CLI is not hard. Making a Non-Obvious power-user thoughtful CLI is extremely hard, and that's the magic of this."

## The Story to Tell

### Act 1: The discrawl revelation

Discord has 300+ API endpoints. The printing press can generate a CLI that wraps all 323 of them. But discrawl - Peter Steinberger's Discord tool - has 12 commands and 551 stars. Why?

Because discrawl saw something in Discord's data that Discord itself didn't design for: conversations are institutional knowledge. Every message thread is a document that should be archived, indexed, and searched. The 12 commands that embody that insight are worth more than 323 that don't.

That's the Non-Obvious Insight. And until now, discovering it required Peter Steinberger.

### Act 2: The gogcli vs Google surprise

Matt Van Horn (the author) was deciding which Google Workspace CLI to use. He ran [/last30days](https://github.com/mvanhorn/last30days-skill) - his recency research skill - to compare gogcli (Steinberger's, 6.5K stars) vs Google's official Workspace CLI (10K+ stars in a week). The community data said: use gogcli. The newer, official one with 10x the API coverage lost to the older third-party one.

Why? Because gogcli understands what users actually DO with Google Workspace. Human-friendly flags. Sane defaults. Workflow-oriented commands. The official one wraps every API endpoint but doesn't understand the user.

That's the NOI again. Breadth doesn't beat depth. Understanding the user beats understanding the API.

### Act 3: The printing press automates the insight

The printing press doesn't just wrap APIs. It discovers the NOI automatically:

1. **Profiler** classifies the API's domain archetype (PM, communication, payments...)
2. **Phase 0** forces the LLM to write the NOI before proceeding
3. **Entity mapper** identifies which resources are primary vs support
4. **Data gravity scoring** determines what deserves local SQLite tables vs API-only
5. **Workflow templates** generate domain-specific commands from the archetype
6. **Insight templates** generate behavioral analysis commands

The result: a CLI that thinks like discrawl, not like a spec compiler.

## Proposed README Structure

### 1. Rewrite opening tagline

Something that captures: this is the NOI engine, not just a CLI generator.

### 2. Restructure to lead with the story, not the mechanics

Current order:
1. Tagline
2. Why CLIs Matter (agent economics)
3. NOI section
4. Creativity Ladder
5. Domain Archetypes
6. How It Works (phases)

New order:
1. Tagline (NOI-focused)
2. **The discrawl story** (323 commands < 12 commands)
3. **The NOI concept** (the formula, the examples)
4. **The gogcli surprise** (Matt's personal story with /last30days)
5. Why CLIs Matter (agent economics - shorter, supporting role)
6. Creativity Ladder + Domain Archetypes
7. How It Works (phases)

### 3. Give credit to last30days

Include that the gogcli comparison was done with [last30days](https://github.com/mvanhorn/last30days-skill) - Matt Van Horn's recency research skill that searches Reddit, X, YouTube, and more.

### 4. The discrawl numbers

| | Discord API Wrapper | discrawl |
|---|---|---|
| Commands | 323 | 12 |
| What it wraps | Every endpoint | None (custom) |
| Local persistence | Generic JSON blobs | Domain-specific SQLite + FTS5 |
| Stars | ~0 | 551 |
| The insight | "Discord has channels" | "Discord is a knowledge base" |

The press generates the 323-command wrapper in Phase 2. Then it generates the 12 discrawl-style commands in Phase 4. That's the difference between Rung 1 and Rung 5.

## Acceptance Criteria

- [ ] README leads with the discrawl story (323 vs 12 commands narrative)
- [ ] NOI section comes immediately after the story, as the explanation
- [ ] gogcli vs Google Workspace CLI story included with /last30days credit
- [ ] Agent/token economics section shortened to supporting role (not the lead)
- [ ] Credit to Peter Steinberger for discrawl and gogcli
- [ ] Credit to Matt Van Horn and /last30days
- [ ] The printing press framed as "automates the insight Peter Steinberger does manually"

## Files

- `README.md`
