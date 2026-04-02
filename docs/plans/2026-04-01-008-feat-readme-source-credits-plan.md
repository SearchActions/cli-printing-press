---
title: "feat: Populate README source credits from skill research"
type: feat
status: active
date: 2026-04-01
origin: docs/retros/2026-04-01-dominos-retro.md
---

# feat: Populate README source credits from skill research

## Overview

The README template has a "Sources & Inspiration" section that credits community projects. The generator has `Sources` and `DiscoveryPages` fields and a `--research-dir` flag. But the printing-press skill never populates these, so every generated CLI ships without crediting the community wrappers, MCP servers, and competing CLIs it learned from.

## Problem Frame

During the Domino's session, the skill discovered 10 tools (apizza, node-dominos-pizza-api, pizzapi, mcpizza, etc.) and sniffed 11 pages. None of this appeared in the generated README. The generator's pipeline exists and works -- the skill just never connects to it. The user wants the most important credits to appear automatically, not every URL that was glanced at.

## Requirements Trace

- R1. Generated READMEs must credit the top community projects that informed the CLI's design (sorted by importance, not exhaustive)
- R2. Sniff discovery pages should appear when the CLI was built from sniffed traffic
- R3. The credit section should appear automatically without skill-instruction changes every generation
- R4. Low-value sources (analytics domains, generic search results, abandoned repos) should be filtered out

## Scope Boundaries

- Does not change the README template (it already has the `Sources & Inspiration` section)
- Does not change the generator's `ReadmeSource` struct or `loadResearchSources()` function
- Does not require the binary's `printing-press research` command to run (the skill does its own research)
- Does not change the absorb manifest format

## Context & Research

### Relevant Code and Patterns

- `internal/generator/templates/readme.md.tmpl` lines 194-213: Conditional `Sources & Inspiration` section, renders `.Sources` and `.DiscoveryPages`
- `internal/generator/generator.go` lines 40-41: `Sources []ReadmeSource` and `DiscoveryPages []string` fields
- `internal/cli/root.go` line 362: `--research-dir` flag on generate command
- `internal/cli/root.go` lines 649-666: `loadResearchSources()` reads `research.json` and `discovery/sniff-report.md`
- `internal/pipeline/research.go` lines 23-63: `ResearchResult` and `Alternative` structs
- `internal/pipeline/research.go` lines 214-234: `SourcesForREADME()` filters by URL presence, sorts by stars descending
- `internal/pipeline/research.go` lines 180-209: `ParseDiscoveryPages()` extracts URLs from sniff-report.md "Pages Visited" section

### Data Flow (current, broken)

```
Skill Phase 1.5 discovers tools → writes absorb-manifest.md (markdown, not JSON)
                                    ↓
                              (data lost here)
                                    ↓
Skill Phase 2 runs `printing-press generate --spec ...` (no --research-dir)
                                    ↓
Generator sees empty Sources → README omits "Sources & Inspiration" section
```

### Data Flow (desired)

```
Skill Phase 1.5 discovers tools → writes absorb-manifest.md AND research.json
                                                                    ↓
Skill Phase 2 runs `printing-press generate --spec ... --research-dir $PIPELINE_DIR`
                                    ↓
Generator reads research.json → populates Sources → README shows credits
```

## Key Technical Decisions

- **Write research.json from the skill, not add a new binary command**: The skill already has all the source data (tool names, GitHub URLs, languages) from Phase 1.5. Writing a JSON file is simpler than piping data through CLI flags. The `research.json` format is already defined by the binary's `ResearchResult` struct.
- **Filter to "important" sources by requiring a GitHub URL and preferring starred repos**: `SourcesForREADME()` already filters out entries with empty URLs and sorts by stars descending. The skill's job is to populate the `Alternatives` array with only the tools that materially informed the CLI (the ones from the absorb manifest), not every search result.
- **Cap at 8 sources**: The README section should credit the most important projects, not list 20. The skill should write only the top tools from the absorb manifest (those with GitHub URLs). `SourcesForREADME()` already sorts by stars; if we want a cap, add it there or in the skill.

## Implementation Units

- [ ] **Unit 1: Add research.json writing to the printing-press skill**

**Goal:** After Phase 1.5 (Absorb Gate) completes, the skill writes a `research.json` file that the generator can consume.

**Requirements:** R1, R4

**Dependencies:** None

**Files:**
- Modify: `skills/printing-press/SKILL.md` (Phase 1.5d section, after writing absorb manifest)

**Approach:**
After writing the absorb manifest in Phase 1.5d, add an instruction block that tells Claude to also write `$PIPELINE_DIR/research.json`. The JSON structure must match `ResearchResult`:

```json
{
  "api_name": "<api>",
  "alternatives": [
    {"name": "apizza", "url": "https://github.com/harrybrwn/apizza", "language": "Go", "stars": 200, "command_count": 5},
    {"name": "node-dominos-pizza-api", "url": "https://github.com/RIAEvangelist/node-dominos-pizza-api", "language": "JavaScript", "stars": 800}
  ],
  "researched_at": "<timestamp>"
}
```

The skill should populate `alternatives` from the absorb manifest's "Sources Cataloged" section, filtering to entries that have a GitHub URL. Cap at 8 entries. Order by relevance (tools with the most absorbed features first, then by stars).

The key filtering rule: only include tools that contributed features to the absorb manifest. If a tool was found during search but contributed nothing to the manifest, skip it.

**Patterns to follow:**
- The skill already writes JSON state files (e.g., `$STATE_FILE` with `api_name`, `working_dir`, etc.)
- The `Alternative` struct fields: `name`, `url`, `language`, `stars`, `command_count`, `has_json_output`, `has_auth_support`

**Test scenarios:**
- Happy path: Run `/printing-press` for an API with 5+ community tools. After Phase 1.5, `$PIPELINE_DIR/research.json` exists and contains 5 alternatives with name, url, language.
- Edge case: API with no community tools (e.g., brand-new API). `research.json` should have an empty `alternatives` array. The README section is omitted (template condition handles this).
- Edge case: API with 15 tools found. Only top 8 (by absorbed feature count, then stars) appear in research.json.

**Verification:**
- After Phase 1.5 completes, `cat $PIPELINE_DIR/research.json` shows valid JSON with alternatives array
- The alternatives have GitHub URLs, not npm/PyPI landing pages

- [ ] **Unit 2: Add --research-dir to all generate invocations in the skill**

**Goal:** The generate command receives the pipeline directory so it can load research.json and discovery data.

**Requirements:** R1, R2, R3

**Dependencies:** Unit 1 (research.json must exist before generate reads it)

**Files:**
- Modify: `skills/printing-press/SKILL.md` (Phase 2 generate commands)

**Approach:**
Add `--research-dir "$PIPELINE_DIR"` to every `printing-press generate` invocation in the Phase 2 section. There are 7 variants (OpenAPI, sniff-enriched, sniff-only, crowd-sniff-enriched, crowd-sniff-only, both, docs-only). Each needs the flag.

Also ensure `$PIPELINE_DIR` is defined. It already is: `PIPELINE_DIR="$API_RUN_DIR/pipeline"` from the setup section. However, research.json needs to be in this directory (Unit 1 writes it there).

For `DiscoveryPages`: the sniff report is written to `$DISCOVERY_DIR/sniff-report.md`. The generator's `ParseDiscoveryPages()` looks for it at `<research-dir>/discovery/sniff-report.md`. So either:
- (a) Pass `--research-dir "$API_RUN_DIR"` (parent of both pipeline/ and discovery/), or
- (b) Copy/symlink sniff-report.md into the pipeline dir

Option (a) is cleaner if research.json is moved to `$API_RUN_DIR/research.json`. But the pipeline dir is already `$API_RUN_DIR/pipeline`. Let me check what `loadResearchSources` expects:

It reads `research.json` from the research-dir root, and `discovery/sniff-report.md` from a `discovery/` subdirectory of research-dir. So if `--research-dir "$API_RUN_DIR"`, it would look for `$API_RUN_DIR/research.json` and `$API_RUN_DIR/discovery/sniff-report.md`. The discovery dir IS at `$API_RUN_DIR/discovery`. So Unit 1 should write research.json to `$API_RUN_DIR/research.json` (not pipeline dir), and the generate flag should be `--research-dir "$API_RUN_DIR"`.

**Patterns to follow:**
- Existing generate invocations in the skill all follow the same structure with `\` line continuations

**Test scenarios:**
- Happy path: Generate a CLI after research.json and sniff-report.md exist. README contains "Sources & Inspiration" section with credited tools.
- Edge case: Generate without research.json (e.g., catalog API with no Phase 1.5). `--research-dir` points to an empty dir. `loadResearchSources` silently skips. README has no Sources section. No error.
- Integration: Full run from `/printing-press <API>` through publish. The published README credits the community projects.

**Verification:**
- Generated README contains `## Sources & Inspiration` with at least one credited project
- Generated README contains discovery pages when sniff was used

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Skill writes research.json with wrong field names | The `Alternative` struct is the contract. Include field names verbatim in the skill instruction. |
| research.json written before generate but in wrong directory | Use `$API_RUN_DIR` consistently. The discovery/ subdirectory already lives there. |
| Stars/language data unavailable during research | These fields are optional. `SourcesForREADME()` works without them (just sorts to bottom). |

## Sources & References

- Dominos retro finding #8: README generated with placeholder examples
- `internal/generator/templates/readme.md.tmpl` lines 194-213
- `internal/pipeline/research.go` `ResearchResult` and `Alternative` structs
- `internal/cli/root.go` `loadResearchSources()` and `--research-dir` flag
