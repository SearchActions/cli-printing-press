---
title: "Research, Dogfood, and Distribution Pipeline for CLI Printing Press"
type: feat
status: completed
date: 2026-03-24
---

# Research, Dogfood, and Distribution Pipeline for CLI Printing Press

## Overview

Transform the Printing Press from a CLI generator into a CLI factory that researches, generates, tests, compares, documents, and distributes - all in one pipeline run. Three new pipeline stages (Research, Dogfood-Enhanced README, Comparative Analysis) plus distribution infrastructure (GoReleaser, Homebrew tap, GitHub Releases) turn `printing-press print <api>` into an end-to-end product launch.

The patterns are transplanted from the Open-Source-Contributor (OSC) project, which has battle-tested logic for dogfooding, competitive gap analysis, evidence capture, anti-AI text filtering, and quality scoring with calibration feedback.

## Problem Statement / Motivation

Today the pipeline generates a CLI that compiles and passes 7 quality gates - but it doesn't know whether:

1. **Someone already built this** - We generated CLIs for GitHub, Stripe, Cloudflare, etc. that already have official CLIs. Wasted effort. A research step would have said "skip this, gh exists and is excellent."
2. **The generated CLI actually works against the real API** - Quality gates prove it compiles, not that it works. The 3-tier dogfood system exists in the Review phase seeds but isn't automated or evidence-producing.
3. **Our CLI is better than alternatives** - The Linear CLI competitive analysis was done manually. This should be automatic.
4. **The README has real examples** - The generated README uses template-derived examples, not captured output from actual runs. Real output is more trustworthy and catches template bugs.
5. **Anyone can install it** - No Homebrew tap, no pre-built binaries, no `curl | sh` installer. Install friction is the #1 gap identified in the competitive analysis.

The Linear CLI session proved the value: research discovered 5+ community CLIs, competitive analysis showed our 45-command breadth advantage, and manual dogfooding caught GraphQL query issues. But it took 15+ minutes of human-in-the-loop work. This plan automates all of it.

## Proposed Solution

### New Pipeline Phases

Expand `PhaseOrder` from 6 to 8 phases:

```
Before:  [preflight, scaffold, enrich, regenerate, review, ship]
After:   [preflight, research, scaffold, enrich, regenerate, review, comparative, ship]
```

| New Phase | Position | What It Does |
|-----------|----------|--------------|
| **Research** | After preflight, before scaffold | Discovers existing CLIs, analyzes gaps, produces `research.json` |
| **Comparative** | After review, before ship | Scores our CLI vs alternatives on 6 dimensions, produces comparison report |

The **Review** phase is enhanced (not replaced) with automated dogfooding evidence capture and README regeneration from real output.

The **Ship** phase is enhanced with GoReleaser execution and Homebrew tap push.

### Phase Details

#### Phase: Research (NEW)

**Trigger:** Runs for every `printing-press print` pipeline. Skippable for `generate` (direct mode).

**Steps:**

1. **Known Alternatives Registry Check** - New `known_alternatives` field in catalog entries. If the API has known CLIs (e.g., `github.yaml` lists `gh`), load them directly.

2. **Automated Discovery** - For APIs without registered alternatives:
   - GitHub search: `<api-name> cli` sorted by stars
   - npm search: `<api-name>-cli`
   - Homebrew search: `brew search <api-name>`
   - PyPI search: `<api-name>-cli`
   - Web search: `"<api-name> cli" OR "<api-name> command line"`

3. **Gap Analysis** - For each discovered alternative:
   - Count commands/endpoints covered (from README or `--help` if installed)
   - Check auth methods supported
   - Check output formats (JSON, table, plain)
   - Note install method and language/runtime
   - Check last commit date (stale = opportunity)

4. **Novelty Assessment** - Score 1-10:
   - 1-3: Official CLI exists and is comprehensive. Flag as "low novelty - consider skipping."
   - 4-6: Community CLIs exist but are incomplete or stale. Good opportunity.
   - 7-10: No CLI exists. High value target.

**Output:** `research.json` in pipeline directory:

```go
// internal/pipeline/research.go
type ResearchResult struct {
    APIName        string        `json:"api_name"`
    NoveltyScore   int           `json:"novelty_score"`  // 1-10
    Alternatives   []Alternative `json:"alternatives"`
    Gaps           []string      `json:"gaps"`           // What alternatives miss
    Patterns       []string      `json:"patterns"`       // What alternatives do well
    Recommendation string        `json:"recommendation"` // "proceed", "proceed-with-gaps", "skip"
}

type Alternative struct {
    Name          string `json:"name"`
    URL           string `json:"url"`
    Language      string `json:"language"`
    InstallMethod string `json:"install_method"` // brew, npm, pip, cargo, binary
    Stars         int    `json:"stars"`
    LastUpdated   string `json:"last_updated"`
    CommandCount  int    `json:"command_count"`
    HasJSON       bool   `json:"has_json_output"`
    HasAuth       bool   `json:"has_auth_support"`
}
```

**Consent Gate:** None. Research is read-only and safe. If novelty score is 1-3, the plan seed for Scaffold includes a warning: "Official CLI exists - consider whether this CLI adds value."

**Borrowed from OSC:** osc-newfeature Phase 1e (community sentiment + competitor analysis) and Phase 1f (gap analysis + idea generation). Simplified for CLI-specific dimensions.

#### Phase: Review Enhancement (EXPANDED)

The existing Review phase's 3-tier dogfood system gets automated evidence capture and README regeneration.

**Tier 1 (No Auth - Always Runs):**
```
<cli> --help                    → capture stdout
<cli> version                   → capture stdout
<cli> doctor                    → capture stdout
<cli> <resource> --help         → capture stdout for each resource
<cli> <resource> <cmd> --dry-run → capture dry-run output for 1 cmd per resource
```

**Tier 2 (Read-Only - If Credentials Available):**
```
<cli> <resource> list           → capture stdout (first list-type cmd per resource)
<cli> <resource> get <id>       → capture stdout (if list returned an ID)
```

Credential detection: check env vars defined in `APISpec.Auth.EnvVars`. If any are set, Tier 2 is available. Consent model: `dogfood_tier` field in the Review plan seed defaults to `1`. User or agent edits to `2` or `3` before execution. Pipeline never escalates beyond declared tier.

**Tier 3 (Sandbox Write - Only if SandboxSafe):**
```
<cli> <resource> create --<required-fields> → capture stdout
<cli> <resource> delete <created-id>        → capture stdout (cleanup)
```

Hard-gated on `SandboxSafe: true` in `KnownSpecs`. No override flag. To add sandbox support for an API, update the catalog entry with `sandbox_endpoint` and set `SandboxSafe: true` in `discover.go`.

**Evidence Capture:**

All captured output goes to `dogfood-evidence/` in the pipeline directory:

```
dogfood-evidence/
  tier1-help.txt
  tier1-version.txt
  tier1-doctor.txt
  tier1-resources/
    issues-help.txt
    issues-list-dry-run.txt
    projects-help.txt
    ...
  tier2-reads/
    issues-list.txt
    issues-get-ENG-123.txt
    ...
  tier3-writes/
    issues-create.txt
    issues-delete.txt
    ...
  dogfood-results.json
```

Results schema:

```go
// internal/pipeline/dogfood.go
type DogfoodResults struct {
    Tier           int             `json:"tier"`          // Highest tier reached
    TotalCommands  int             `json:"total_commands"`
    PassedCommands int             `json:"passed_commands"`
    FailedCommands int             `json:"failed_commands"`
    Commands       []CommandResult `json:"commands"`
    Score          int             `json:"score"`         // 0-50
}

type CommandResult struct {
    Tier     int    `json:"tier"`
    Command  string `json:"command"`
    ExitCode int    `json:"exit_code"`
    Stdout   string `json:"stdout_file"` // path to evidence file
    Stderr   string `json:"stderr"`
    Duration int    `json:"duration_ms"`
    Pass     bool   `json:"pass"`
}
```

**README Regeneration:**

After dogfooding completes, augment the generated README with real output:

1. Read `dogfood-evidence/tier1-help.txt` - replace the `--help` example section
2. Read `dogfood-evidence/tier1-doctor.txt` - add a "Health Check" section with real output
3. Read `dogfood-evidence/tier1-resources/*.txt` - add real command examples per resource
4. If Tier 2 ran: add "Real API Examples" section with actual list/get output
5. Apply anti-AI text filter to all description text (see below)

Implementation: new function `augmentREADME(readmePath string, evidence *DogfoodResults)` in a new `internal/generator/readme_augment.go`. This reads the existing README, finds marker comments (e.g., `<!-- HELP_OUTPUT -->`) inserted by the updated `readme.md.tmpl`, and replaces them with real captured output.

**Anti-AI Text Filter:**

Borrowed from OSC's osc-work Step 5c. A regex-based blocklist applied to README and help text descriptions:

```go
// internal/generator/textfilter.go
var aiSlopPatterns = []string{
    `(?i)\b(comprehensive|robust|seamless|leverage|utilize|facilitate)\b`,
    `(?i)here's a .* that`,
    `(?i)not just .*, it's`,
    `(?i)whether you're .* or .*,`,
    `(?i)\b(streamline|empower|cutting-edge|game-changer)\b`,
}
```

Applied to: README description text, `--help` Short/Long descriptions in generated command files, catalog entry descriptions. NOT applied to: code, variable names, comments, or user-provided spec descriptions.

Action on match: log a warning with the matched pattern and text. Do not auto-replace (too risky for generated code). The warning surfaces in the dogfood score as a deduction.

**Borrowed from OSC:**
- Dogfooding checkpoints with tier escalation (osc-work Step 5e)
- Evidence capture with before/after comparison (osc-newfeature Phase 1b.5)
- Anti-AI text filter blocklist (osc-work Step 5c)
- Gate tracker pattern - session-persistent file recording which gates ran (osc-work)

#### Phase: Comparative Analysis (NEW)

**Trigger:** Runs after Review, before Ship. Requires `research.json` from Research phase.

**Scoring Dimensions (6 total, 100 points max):**

| Dimension | Points | How Measured |
|-----------|--------|-------------|
| **Breadth** | 20 | Command count / max(our commands, best alternative commands) |
| **Install Friction** | 20 | 20 if `go install` or binary download; 15 if requires clone+build; 10 if requires runtime (Node/Python); 5 if complex setup |
| **Auth UX** | 15 | 15 if env var + config file; 10 if env var only; 5 if manual header; 0 if no auth support |
| **Output Formats** | 15 | 5 per format (JSON, table, plain) |
| **Agent Friendliness** | 15 | 5 for `--json`, 5 for `--dry-run`, 5 for non-interactive (no prompts/pagers) |
| **Freshness** | 15 | 15 if last updated < 30 days; 10 if < 90 days; 5 if < 1 year; 0 if > 1 year |

Our CLI always scores: Install Friction 15 (requires `go install` or binary), Auth UX 15, Output Formats 15 (JSON + table + plain), Agent Friendliness 15, Freshness 15. Minimum 75/100 before Breadth. Breadth depends on the API.

**Output:** `comparative-analysis.md` in the pipeline directory. Includes:
- Score table (our CLI vs each alternative)
- Gap summary (what we're missing that alternatives have)
- Advantage summary (what we have that alternatives don't)
- Recommendation: "ship" (we're the best), "ship-with-gaps" (we're competitive but have known gaps), "hold" (an alternative is clearly better)

**Borrowed from OSC:** osc-newfeature Phase 1e competitor analysis and the 8-dimension scoring rubric (simplified to 6 CLI-specific dimensions).

#### Phase: Ship Enhancement (EXPANDED)

**GoReleaser Integration:**

Update `goreleaser.yaml.tmpl` to include:

```yaml
brews:
  - name: "{{ .Name }}"
    repository:
      owner: "{{ .GithubOwner }}"
      name: "homebrew-tap"
    homepage: "{{ .Homepage }}"
    description: "{{ .Description }}"
    install: |
      bin.install "{{ .Name }}"
```

New fields in `APISpec` or a separate `DistConfig`:
- `GithubOwner` - defaults to git remote owner
- `Homepage` - defaults to GitHub repo URL
- `TapRepo` - defaults to `homebrew-tap`

Ship phase steps:
1. `goreleaser check` - validate config
2. If `--release` flag: `goreleaser release --clean`
3. Verify GitHub release was created
4. Verify Homebrew formula was pushed to tap

**Homebrew Tap Setup:**

For the printing press's own distribution AND as a template for generated CLIs:

```
mvanhorn/homebrew-tap/
  Formula/
    printing-press.rb     # The press itself
    linear-cli.rb         # Generated CLIs get their own formula
```

GoReleaser handles formula generation and push. The tap repo must exist and the `GITHUB_TOKEN` must have push access.

**GitHub Releases:**

GoReleaser creates releases with:
- Cross-compiled binaries (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64)
- SHA256 checksums
- Changelog from conventional commits
- Install instructions in release notes

### Pipeline Integration Points

**State Model Changes** (`internal/pipeline/state.go`):

```go
const (
    PhasePreflight    Phase = "preflight"
    PhaseResearch     Phase = "research"      // NEW
    PhaseScaffold     Phase = "scaffold"
    PhaseEnrich       Phase = "enrich"
    PhaseRegenerate   Phase = "regenerate"
    PhaseReview       Phase = "review"
    PhaseComparative  Phase = "comparative"   // NEW
    PhaseShip         Phase = "ship"
)

var PhaseOrder = []Phase{
    PhasePreflight, PhaseResearch, PhaseScaffold, PhaseEnrich,
    PhaseRegenerate, PhaseReview, PhaseComparative, PhaseShip,
}
```

State version field added for migration: `Version int json:"version"`. Old state files (version 0) get Research and Comparative phases auto-inserted with status "skipped" on load.

**Seed Templates** (`internal/pipeline/seeds.go`):

New seed templates for Research and Comparative phases. Each follows the existing pattern: markdown with frontmatter, actionable steps, and expected output files.

**Catalog Schema Extension** (`internal/catalog/catalog.go`):

```go
type Entry struct {
    // ... existing fields ...
    KnownAlternatives []KnownAlt `yaml:"known_alternatives,omitempty"`
    SandboxEndpoint   string     `yaml:"sandbox_endpoint,omitempty"`
}

type KnownAlt struct {
    Name     string `yaml:"name"`
    URL      string `yaml:"url"`
    Language string `yaml:"language"`
}
```

**Bug Fix** (`internal/pipeline/discover.go:142-144`):

The apis-guru fallback loop returns unconditionally on the first iteration. Fix: accumulate results and return the highest version match.

## Technical Considerations

### Architecture

- Research and Comparative phases are independent Go packages (`internal/pipeline/research.go`, `internal/pipeline/comparative.go`) with clear input/output contracts
- Dogfood automation lives in `internal/pipeline/dogfood.go` - orchestrates CLI binary execution and evidence capture
- README augmentation is in `internal/generator/readme_augment.go` - reads marker comments from the template-generated README and replaces with real output
- Anti-AI text filter is in `internal/generator/textfilter.go` - reusable regex matcher
- GoReleaser config template update is a template change only

### Performance

- Research phase adds network latency (GitHub API, npm, web search). Cached with 24h TTL matching spec cache.
- Dogfooding adds CLI execution time. Tier 1: ~5 seconds. Tier 2: ~15 seconds (API latency). Tier 3: ~30 seconds (create + cleanup).
- Per-tier timeout: 60 seconds. Total dogfood timeout: 180 seconds (3 minutes, down from 600).
- Comparative analysis is compute-only (reads research.json + dogfood-results.json). < 1 second.

### Security

- Tier 3 dogfooding is hard-gated on `SandboxSafe` with NO override flag
- Credentials are never logged or stored in evidence files (stdout capture strips Authorization headers via the generated client's `--dry-run` mode for header display)
- GoReleaser requires `GITHUB_TOKEN` - documented but never stored in generated code
- Research phase makes only GET requests to public APIs

## System-Wide Impact

### Interaction Graph

```
printing-press print <api>
  → Preflight (existing)
  → Research (NEW) → writes research.json
  → Scaffold (existing) → reads research.json for README novelty note
  → Enrich (existing)
  → Regenerate (existing)
  → Review (EXPANDED) → reads binary, runs dogfood, writes evidence/, writes dogfood-results.json
                       → augments README with real output
                       → runs anti-AI text filter
  → Comparative (NEW) → reads research.json + dogfood-results.json, writes comparative-analysis.md
  → Ship (EXPANDED) → reads goreleaser.yaml, runs goreleaser, pushes to tap
```

### Error Propagation

- Research failure: non-fatal. Scaffold proceeds with empty `research.json`. Comparative phase produces "no alternatives found" report.
- Dogfood Tier 1 failure: fatal for Review phase (if the CLI can't even show --help, it's broken). Score 0/50.
- Dogfood Tier 2/3 failure: individual command failures are recorded but don't fail the phase. Score reduced proportionally.
- README augmentation failure: non-fatal. Original template README is preserved.
- GoReleaser failure: fatal for Ship phase. No release created. User can retry manually.

### State Lifecycle Risks

- Partial dogfood: `dogfood-results.json` is written atomically after all tiers complete. If the process crashes mid-tier, no results file exists and Review reports "dogfood: incomplete."
- State version migration: old `state.json` files (no version field) default to version 0. Version 1 adds Research and Comparative phases. Migration is additive-only (new phases added with "pending" status).

## Acceptance Criteria

### Functional Requirements

- [ ] `PhaseOrder` expanded to 8 phases with Research and Comparative
- [ ] Research phase discovers alternatives via GitHub API + catalog `known_alternatives`
- [ ] Research phase produces `research.json` with novelty score and alternatives list
- [ ] Dogfood automation executes Tier 1 commands and captures stdout to evidence files
- [ ] Dogfood Tier 2 runs if credentials detected AND `dogfood_tier >= 2` in plan
- [ ] Dogfood Tier 3 runs only if `SandboxSafe: true` AND `dogfood_tier >= 3` in plan
- [ ] README augmented with real `--help`, `version`, `doctor` output after dogfooding
- [ ] Anti-AI text filter warns on sloppy patterns in README descriptions
- [ ] Comparative analysis scores our CLI vs alternatives on 6 dimensions
- [ ] `goreleaser.yaml.tmpl` includes `brews:` section with configurable tap repo
- [ ] Ship phase runs `goreleaser check` to validate config
- [ ] Catalog schema accepts `known_alternatives` and `sandbox_endpoint` fields
- [ ] State migration: old state.json files load correctly with new phases as "skipped"
- [ ] `discover.go` apis-guru fallback loop bug fixed

### Non-Functional Requirements

- [ ] Research phase completes in < 30 seconds (cached) / < 60 seconds (cold)
- [ ] Dogfood Tier 1 completes in < 15 seconds per CLI
- [ ] Total pipeline overhead from new phases: < 2 minutes
- [ ] No external dependencies beyond Go stdlib + existing deps (kin-openapi, cobra, etc.)

### Quality Gates

- [ ] `go test ./...` passes with new packages
- [ ] Run the 10-API gauntlet with Research + Dogfood enabled, all 10 pass
- [ ] Linear CLI (GraphQL/YAML spec) works end-to-end through the expanded pipeline
- [ ] At least 3 catalog entries have `known_alternatives` populated

## Implementation Phases

### Phase 1: Research Stage + Catalog Extension

**Files:**
- `internal/pipeline/research.go` - Research phase logic and output schema
- `internal/pipeline/state.go` - Add `PhaseResearch`, `PhaseComparative`, expand `PhaseOrder`, add state version
- `internal/pipeline/seeds.go` - Add seed templates for Research and Comparative phases
- `internal/catalog/catalog.go` - Add `KnownAlternatives` and `SandboxEndpoint` fields
- `catalog/github.yaml`, `catalog/stripe.yaml`, `catalog/telegram.yaml` - Populate `known_alternatives` for 3 entries
- `internal/pipeline/discover.go` - Fix apis-guru fallback loop bug

**Success criteria:** `printing-press print github` runs Research phase, discovers `gh` as an alternative, produces `research.json` with novelty score 2 (official CLI exists), and recommends "proceed-with-gaps" or "skip".

### Phase 2: Dogfood Automation + Evidence Capture

**Files:**
- `internal/pipeline/dogfood.go` - Dogfood orchestrator, tier escalation, evidence capture
- `internal/generator/readme_augment.go` - README augmentation from evidence files
- `internal/generator/templates/readme.md.tmpl` - Add marker comments for augmentation points
- `internal/generator/textfilter.go` - Anti-AI text filter

**Success criteria:** `printing-press print petstore` runs all 3 tiers (petstore is SandboxSafe), captures evidence files, augments README with real output, and anti-AI filter flags any sloppy descriptions. Dogfood score computed as 0-50.

### Phase 3: Comparative Analysis + Ship Enhancement

**Files:**
- `internal/pipeline/comparative.go` - Comparative analysis scoring and report generation
- `internal/generator/templates/goreleaser.yaml.tmpl` - Add `brews:` section
- `internal/pipeline/seeds.go` - Update Ship seed template with GoReleaser steps

**Success criteria:** After Review, Comparative phase reads `research.json` + `dogfood-results.json`, produces `comparative-analysis.md` with score table. GoReleaser config validates with `goreleaser check`.

### Phase 4: Distribution Infrastructure

**Files/Resources:**
- Create `mvanhorn/homebrew-tap` GitHub repo
- Add `Formula/printing-press.rb` to the tap
- Tag and release printing-press v1.0.0 via GoReleaser
- Document install: `brew install mvanhorn/tap/printing-press`

**Success criteria:** `brew install mvanhorn/tap/printing-press` works on a clean Mac. Binary runs. `printing-press version` shows v1.0.0.

## Alternative Approaches Considered

### A: Extend `generate` command instead of pipeline

**Rejected because:** The `generate` command is intentionally lean (spec in, CLI out). Research and comparative analysis are multi-step processes that benefit from pipeline state tracking, plan seeds, and resumability. Adding them to `generate` would bloat the simple path.

### B: Use OSC's osc-newfeature skill directly

**Rejected because:** osc-newfeature is designed for open-source contribution proposals, not CLI quality assessment. The dimensions are different (maintainer warmth, PR merge likelihood vs. CLI breadth, install friction). Transplanting the patterns and adapting the scoring is better than wrapping the skill.

### C: Actually install and run competing CLIs for comparison

**Rejected because:** Installing npm/pip/cargo packages in an automated pipeline is a sandbox nightmare. Metadata-based comparison (README, `--help` if pre-installed, star count, last commit) is 90% as accurate with 10% of the complexity. Can be added later as an opt-in Tier 2 comparison.

### D: Homebrew-first distribution (skip GoReleaser)

**Rejected because:** GoReleaser handles cross-compilation, checksums, GitHub releases, AND Homebrew formula generation. Doing Homebrew manually would mean maintaining a separate release script. GoReleaser is the standard for Go CLI distribution.

## Dependencies & Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| GitHub API rate limits during Research | Research phase fails for unauthenticated users (60 req/hr) | Cache research results for 24h. Use `GITHUB_TOKEN` if available for 5000 req/hr |
| GoReleaser breaking changes | Ship phase fails on new GoReleaser versions | Pin GoReleaser version in CI. Use `goreleaser check` before `goreleaser release` |
| Anti-AI filter false positives | Legitimate text flagged as AI slop | Filter only warns, never auto-replaces. User reviews warnings |
| Dogfood timeout on slow APIs | Tier 2 reads hang on rate-limited APIs | Per-command timeout of 15 seconds. Total tier timeout of 60 seconds |
| State migration breaks existing pipelines | Users with in-progress pipelines lose progress | Version field + additive migration (new phases = "skipped", don't disturb existing phase states) |

## Sources & References

### Internal References

- Pipeline state model: `internal/pipeline/state.go`
- Pipeline seeds: `internal/pipeline/seeds.go`
- Quality gates: `internal/generator/validate.go`
- Catalog schema: `internal/catalog/catalog.go`
- README template: `internal/generator/templates/readme.md.tmpl`
- GoReleaser template: `internal/generator/templates/goreleaser.yaml.tmpl`
- Discover registry: `internal/pipeline/discover.go`
- Competitive analysis (manual): `docs/plans/2026-03-24-docs-linear-cli-competitive-analysis-plan.md`
- Dogfood gauntlet findings: `docs/plans/dogfood-gauntlet-findings.md`
- Autonomous dogfood phase plan: `docs/plans/2026-03-24-feat-autonomous-dogfood-phase-plan.md`
- Launch readiness plan: `docs/plans/2026-03-24-feat-printing-press-launch-readiness-plan.md`

### OSC Patterns Transplanted

- Dogfooding with tier escalation: `~/.claude/skills/osc-work/SKILL.md` Step 5e
- Evidence capture: `~/.claude/skills/osc-newfeature/SKILL.md` Phase 1b.5
- Anti-AI text filter: `~/.claude/skills/osc-work/SKILL.md` Step 5c
- Competitive gap analysis: `~/.claude/skills/osc-newfeature/SKILL.md` Phase 1e-1f
- Quality scoring with calibration: `~/open-source-contributor/scripts/calibrate.py`
- Gate tracker pattern: `~/.claude/skills/osc-work/SKILL.md` gate tracker file

### External References

- GoReleaser Homebrew tap docs: goreleaser.com/customization/homebrew/
- Homebrew tap creation: docs.brew.sh/How-to-Create-and-Maintain-a-Tap
