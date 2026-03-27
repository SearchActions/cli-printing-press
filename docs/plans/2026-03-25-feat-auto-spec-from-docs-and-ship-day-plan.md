---
title: "Auto-Spec, Competitor Intelligence, and Ship Day"
type: feat
status: active
date: 2026-03-25
---

# Auto-Spec, Competitor Intelligence, and Ship Day

## The Vision

The Printing Press doesn't just generate CLIs - it generates CLIs that are **better than every existing alternative** because it:

1. **Auto-generates specs from API docs** - no hand-writing anything, ever
2. **Studies competing CLIs on GitHub** - reads their issues, README, PRs to learn what users want
3. **Uses those insights to enrich the generated CLI** - adds commands competitors missed, avoids their mistakes
4. **Ships to Homebrew** - `brew install mvanhorn/tap/plaid-cli` on day one

## Implementation Units

### Unit 1: Competitor Intelligence Engine

**Goal:** Enhance the Research phase to analyze competing CLI repos and produce actionable insights that feed into generation.

**File:** `internal/pipeline/research.go` (extend existing)

**What it does when it finds a competing CLI (e.g. `landakram/plaid-cli`):**

1. **Fetch repo metadata** via GitHub API (already done - stars, language, last updated)
2. **NEW: Fetch open issues** - `GET /repos/{owner}/{repo}/issues?state=open&labels=enhancement,feature`
   - Extract feature requests users wanted but never got
   - These become candidate endpoints to add to our spec
3. **NEW: Fetch README** - `GET /repos/{owner}/{repo}/readme`
   - Parse for command listings (detect `usage:`, `commands:`, code blocks with CLI examples)
   - Count commands to set our breadth target (match or beat)
4. **NEW: Fetch closed PRs** - `GET /repos/{owner}/{repo}/pulls?state=closed&sort=created`
   - Features someone tried to add but got rejected or abandoned
   - These are validated demand signals
5. **NEW: Fetch "why I stopped" issues** - search for "abandoned", "alternative", "migrate"
   - UX pain points to avoid in our CLI

**Output schema extension:**

```go
type Alternative struct {
    // ... existing fields ...
    OpenIssueCount    int      `json:"open_issue_count"`
    FeatureRequests   []string `json:"feature_requests"`   // titles of enhancement issues
    CommandsFound     []string `json:"commands_found"`      // commands parsed from README
    AbandonedPRs      []string `json:"abandoned_prs"`       // titles of closed-unmerged PRs
    PainPoints        []string `json:"pain_points"`         // from "why I stopped" issues
}

type ResearchResult struct {
    // ... existing fields ...
    CompetitorInsights CompetitorInsights `json:"competitor_insights"`
}

type CompetitorInsights struct {
    MissingFeatures  []string `json:"missing_features"`  // features no competitor has
    CommandTarget    int      `json:"command_target"`     // breadth target (max competitor + 20%)
    PainPointsToAvoid []string `json:"pain_points_to_avoid"` // UX mistakes to not repeat
    DemandSignals    []string `json:"demand_signals"`    // validated feature requests
}
```

**How Enrich uses this:** The seed template for the Enrich phase will reference `research.json` competitor insights. When enriching the spec, the agent adds endpoints that map to unmet feature requests and ensures command count meets or exceeds the target.

### Unit 2: Doc-to-Spec Generator

**Goal:** `printing-press generate --docs <url>` crawls API docs and auto-generates a YAML spec.

**File:** `internal/docspec/docspec.go` (new package)

**How it works:**

1. Fetch the API docs page via HTTP GET
2. Parse HTML for endpoint patterns:
   - Regex for `(GET|POST|PUT|PATCH|DELETE)\s+/[a-zA-Z0-9/{}_-]+`
   - Extract parameter tables (HTML `<table>` with "Parameter", "Type", "Required" headers)
   - Extract JSON example bodies from `<code>` and `<pre>` blocks
   - Find auth method from "Authentication" or "Authorization" sections
   - Find base URL from "Base URL" or first `https://api.` reference
3. Generate internal YAML spec with extracted endpoints
4. Write to temp file and feed to existing generator pipeline

**Key design decision:** Don't try to be perfect. The goal is a spec that compiles and passes 7/7 quality gates. Missing endpoints get sensible defaults. The generated CLI is a starting point that can be iterated.

**File:** `internal/cli/root.go` - add `--docs` flag

**Test:** `printing-press generate --docs "https://developers.notion.com/reference" --name notion --output /tmp/notion-cli` produces a CLI that compiles.

### Unit 3: Social Demand Signals in Research Phase

**Goal:** Add Reddit/HN/X search to the Research phase to find "need X cli" demand signals.

**File:** `internal/pipeline/research.go` (extend)

**Approach:** Use WebSearch (available in the pipeline's agent context) to query:
- `"<api-name> cli" site:reddit.com`
- `"<api-name> command line" site:news.ycombinator.com`
- `"need <api-name> cli" OR "wish <api-name> had cli"`

**Output:** Add `DemandScore int` (0-10) to `ResearchResult` based on:
- 0 hits = 0 (no demand signal)
- 1-5 hits = 3-5 (some interest)
- 5+ hits with upvotes = 7-10 (strong demand)

**This is a nice-to-have** - implement only if Units 1-2 complete with time remaining. The GitHub competitor analysis (Unit 1) provides 80% of the value.

### Unit 4: Generate All OpenAPI CLIs (with --lenient)

**Goal:** Generate PagerDuty, Intercom, Square, ClickUp from their OpenAPI specs.

```bash
# PagerDuty (works with --lenient, tested last night)
./printing-press generate --spec "https://raw.githubusercontent.com/PagerDuty/api-schema/main/reference/REST/openapiv3.json" --output ./pagerduty-cli --force --lenient

# Intercom (try with --lenient)
./printing-press generate --spec "https://raw.githubusercontent.com/intercom/Intercom-OpenAPI/main/descriptions/2.11/api.intercom.io.yaml" --output ./intercom-cli --force --lenient

# Square
./printing-press generate --spec "https://raw.githubusercontent.com/square/connect-api-specification/master/api.json" --output ./square-cli --force

# ClickUp (community spec)
./printing-press generate --spec "https://raw.githubusercontent.com/rksilvergreen/clickup_openapi_spec_v2/develop/clickup-api-v2-reference.yaml" --output ./clickup-cli --force
```

Each must pass 7/7 quality gates. Add catalog entries for each.

### Unit 5: Generate Doc-Based CLIs (Notion, Airtable, Mixpanel)

**Goal:** Use doc-to-spec for APIs without OpenAPI specs.

```bash
./printing-press generate --docs "https://developers.notion.com/reference" --name notion --output ./notion-cli
./printing-press generate --docs "https://airtable.com/developers/web/api" --name airtable --output ./airtable-cli
./printing-press generate --docs "https://developer.mixpanel.com/reference" --name mixpanel --output ./mixpanel-cli
```

If doc-to-spec can't extract enough for 7/7 gates, iterate - the press keeps trying until it works.

### Unit 6: Create Homebrew Tap + Push Repos

**Goal:** `brew install mvanhorn/tap/plaid-cli` works.

1. `gh repo create mvanhorn/homebrew-tap --public`
2. For each CLI: create repo, push, tag v0.1.0, GoReleaser release
3. Verify `brew install mvanhorn/tap/plaid-cli` on a clean Mac

### Unit 7: Epic README

**Goal:** The printing press README that sells the tool.

- Hero: "Describe your API. Get a CLI that's better than the competition. 60 seconds."
- Demo: real terminal output of generating a CLI
- "Competitor intelligence built in" section showing how it studies existing CLIs
- Catalog of all generated CLIs with install commands
- Steinberger comparison: honest 6.9/10 scoring

## Acceptance Criteria

- [ ] Competitor intelligence: Research phase fetches issues/README/PRs from competing CLIs
- [ ] Doc-to-spec: `--docs` flag produces YAML spec from API documentation pages
- [ ] Notion CLI generated from docs (not hand-written) - passes 7/7 gates
- [ ] PagerDuty CLI generated with --lenient - passes 7/7 gates
- [ ] Homebrew tap exists, `brew install mvanhorn/tap/plaid-cli` works
- [ ] At least 7 CLIs pushed to separate GitHub repos
- [ ] Research.json includes competitor_insights for CLIs with known alternatives
- [ ] README has competitor intelligence section

## Scope Boundaries

- Do NOT hand-write YAML specs for any API
- Do NOT optimize doc-to-spec for perfect extraction - compile is enough
- Do NOT set up CI for individual CLI repos yet
- Social demand signals (Unit 3) is nice-to-have, skip if behind
