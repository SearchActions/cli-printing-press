# Absorb Scoring: Auto-Suggest Novel Features

> **When to read:** This file is referenced by Phase 1.5 of the printing-press skill.
> Read it during Step 1.5c.5 to run gap analysis, score candidates, and populate the transcendence table.

### Step 1.5c.5: Auto-Suggest Novel Features

**This step runs automatically.** No user interaction. Synthesize ALL research gathered so far (Phase 1 brief + Phase 1.5a ecosystem search + Phase 1.5b absorb manifest) into evidence-backed feature recommendations.

#### User-First Feature Discovery

Before generating features from technical capabilities, think about the humans
who will use this CLI. The best transcendence features come from understanding
user rituals and service-specific content patterns, not from asking "what can
SQLite do?"

##### Step 1: Identify specific user personas (2-4 personas)

Don't say "developers" or "users." Name specific people with specific habits:

- "Someone who checks HN every morning before standup"
- "A hiring manager scanning Who's Hiring threads monthly"
- "A movie buff deciding what to watch tonight"
- "A developer about to post their Show HN launch"

Draw these from the Phase 1 brief's "Users" and "Top Workflows" sections. Each
persona represents a feature surface.

##### Step 2: Map each persona's rituals and frustrations

For each persona, answer:
- **What do they do repeatedly?** (daily/weekly/monthly rituals with this service)
- **What question do they wish they could answer but can't?** (This IS the feature.)
- **What's tedious about their current workflow?** (This IS the automation opportunity.)

Example (HN):
- Persona: "Morning HN checker"
- Ritual: Opens HN, scans top stories, opens a few
- Question they can't answer: "What hit the front page while I was coding?"
- Feature: `hn since 2h` — one command, no setup

##### Step 3: Identify service-specific content patterns

Every service has unique content types, categories, or workflows that define its
identity. These are feature surfaces that generic "CRUD + analytics" thinking misses:

- HN: Show HN, Ask HN, Who's Hiring, Who's Looking (each is a feature surface)
- Spotify: Playlists, Discover Weekly, Wrapped/year-end stats
- GitHub: PRs, Issues, Actions, Discussions (each has its own workflows)
- TMDb: Collections/franchises, Watch providers, Trending

For each content pattern, ask: "What would the power user of THIS specific
feature want that no existing tool provides?"

##### Step 4: Validate each feature is mechanical

For every candidate feature, verify it can be built without an LLM or external
service. If it needs AI summarization, reframe it as a mechanical analysis
(top-rated comments, author stats, frequency counts) with pipe-friendly output
that users can send to an LLM themselves.

#### Gap Analysis

After the user-first discovery, run these technical analyses to find anything
the persona work missed:

1. **Cross-entity queries** — What joins across synced tables produce insights no single API call can?

2. **User pain points** — From Phase 1 research: npm README "limitations" sections, GitHub issues on competitor repos, community docs mentioning workarounds

3. **Competitor edges** — From the absorb manifest: what does the BEST competitor tool uniquely offer? Can we beat it?

4. **Agent workflow gaps** — What would an AI agent using this CLI wish it could do in one command instead of multiple?

5. **Self-brainstorm** — Answer using research context:
   - What workflows do power users do that aren't covered in the absorbed features?
   - What are the most annoying limitations that a CLI with local data could fix?
   - What single "killer feature" would make a power user install this over any alternative?
   - (Only when `USER_BRIEFING_CONTEXT` is non-empty) What features directly serve the user's stated goals?

#### Generate and Score Candidates

Generate 3-8 novel feature ideas (across all 6 categories). For each, score on 4 dimensions:

| Dimension | Points | Scoring |
|-----------|--------|---------|
| **Domain Fit** | 0-3 | 3=core to this API's power users, 2=useful but niche, 1=tangential, 0=wrong domain |
| **User Pain** | 0-3 | 3=research surfaced explicit demand (community complaints, competitor gap), 2=implied need, 1=speculative, 0=no evidence |
| **Build Feasibility** | 0-2 | 2=SQLite store + existing sync covers it, 1=needs minor data model additions, 0=requires new infrastructure |
| **Research Backing** | 0-2 | 2=evidence from 2+ sources in Phase 1/1.5 research, 1=evidence from 1 source, 0=invented |

**Normalize:** `score_10 = round(raw / 10 * 10)`. Include features scoring >= 5/10.

#### Add to Transcendence Table

Add each qualifying feature as a new row in the transcendence table:

```markdown
| # | Feature | Command | Why Only We Can Do This | Score | Evidence |
|---|---------|---------|------------------------|-------|----------|
| N | Player comparison | compare "LeBron" "Curry" | Requires local join across player stats + team + season data | 8/10 | ESPN community requests, espn_scraper lacks cross-player queries |
```

The "Evidence" column MUST cite specific findings from Phase 1 or Phase 1.5 research. No unsupported assertions.
