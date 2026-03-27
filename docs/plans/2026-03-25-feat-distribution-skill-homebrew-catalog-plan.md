---
title: "Distribution: Skill Update, Homebrew Tap, and Three-Tier Catalog"
type: feat
status: active
date: 2026-03-25
---

# Distribution: Skill, Homebrew, and Three-Tier Catalog

## Current State

The press can generate CLIs but nobody can install them. Three pieces are missing or broken:

1. **Claude Code skill** - exists but is stale (doesn't know about --polish, --docs, LLM brains, agent-native features)
2. **Homebrew tap** - doesn't exist (the GoReleaser template points at a non-existent repo)
3. **Distribution model** - CLIs sit in the monorepo, not installable by anyone

## The Distribution Model

Three ways to use the press, three tiers of distribution:

### Tier 1: Official Catalog (our curated list)

CLIs we've generated, scored, and verified. Published to `mvanhorn/homebrew-tap`. Installable with one command:

```bash
brew install mvanhorn/tap/plaid-cli
```

These are the ones we stand behind. They passed the scorecard (Grade A), the gauntlet, and dogfood testing. We maintain them - when the API spec changes, we regenerate and release.

**How a CLI gets in:** We generate it, score it, verify it works, push to a separate repo, tag a release, GoReleaser publishes to the tap. Same as how Steinberger manages `steipete/homebrew-tap` with 21 formulas.

### Tier 2: Community PRs (anyone can submit)

Someone uses the press to generate a CLI for their favorite API. It's good. They open a PR to add it to the catalog:

```bash
# They generate it
printing-press generate --spec https://their-api.com/openapi.json --polish

# They submit a catalog entry
printing-press submit their-api
# -> Opens a PR to mvanhorn/cli-printing-press with catalog/their-api.yaml
```

CI validates the entry (schema check, spec URL accessible, generate + compile). If it passes our bar (scorecard Grade B+), we merge it and add it to the tap.

### Tier 3: Self-Hosted (companies use the press)

A company clones the press, points it at their internal API, generates a CLI for their team. They host it on their own tap or distribute the binary internally. The press is the tool, not the distribution.

```bash
# Company generates their internal CLI
printing-press generate --spec https://internal-api.corp.com/openapi.json --polish --name internal-tools

# They publish to their own tap
goreleaser release --clean  # from the generated CLI's directory
```

## Implementation Units

### Unit 1: Update the Claude Code Skill

**File:** `skills/printing-press/SKILL.md`

The skill needs to know about everything we built in this session. Update it with:

- `--docs` flag for APIs without OpenAPI specs (LLM reads docs)
- `--polish` flag for LLM improvement pass
- `--lenient` flag for messy specs
- `--stdin`, `--yes`, `--select`, `--no-cache`, `--human-friendly` in generated CLIs
- Agent-native features (idempotency, typed exit codes, NDJSON events)
- The scorecard (`FULL_RUN=1 go test` for quality verification)
- The three-brain architecture (LLM before + template + LLM after)

The skill's workflows should be:

**Workflow 0 (Natural Language):** "Make me a Plaid CLI"
1. Check KnownSpecs registry
2. If found: `printing-press generate --spec <url> --polish`
3. If not found + LLM available: `printing-press generate --docs <docs-url> --name <name> --polish`
4. If not found + no LLM: search apis-guru, fallback to `--docs` with regex

**Workflow 1 (From Spec):** `--spec <url>`
**Workflow 2 (From Docs):** `--docs <url>` (NEW)
**Workflow 3 (Submit to Catalog):** `submit <name>` - creates PR
**Workflow 4 (Full Pipeline):** `print <name>` - autonomous multi-phase
**Workflow 5 (Scorecard):** `score <output-dir>` - run Steinberger scorecard (NEW)

### Unit 2: Create Homebrew Tap

**Action:** Create `mvanhorn/homebrew-tap` GitHub repo.

```bash
gh repo create mvanhorn/homebrew-tap --public --description "Homebrew tap for CLI Printing Press tools"
```

Initialize with a README listing all available CLIs and install instructions.

No formulas yet - those get added when we push individual CLI repos with GoReleaser.

### Unit 3: Push First CLI to Homebrew (Plaid)

**Goal:** `brew install mvanhorn/tap/plaid-cli` works.

Steps:
1. Create `mvanhorn/plaid-cli` GitHub repo
2. Copy `plaid-cli/` from the monorepo to the new repo
3. Update `go.mod` module path from `USER` to `mvanhorn`
4. Update `.goreleaser.yaml` to point at `mvanhorn/homebrew-tap`
5. Tag `v0.1.0`, push tags
6. Run `goreleaser release --clean` (requires GITHUB_TOKEN)
7. Verify `brew install mvanhorn/tap/plaid-cli` works
8. Verify `plaid-cli --help` runs

### Unit 4: Create `regenerate-all.sh` Script

**File:** `scripts/regenerate-all.sh`

When we update templates (like we did 10+ times this session), we need to regenerate all published CLIs. This script:

1. Iterates over all catalog entries
2. For each: regenerates from spec, runs quality gates
3. If a separate repo exists: copies updated code, commits, tags new version
4. Reports results

```bash
#!/bin/bash
for spec in catalog/*.yaml; do
    name=$(grep "^name:" "$spec" | awk '{print $2}')
    url=$(grep "^spec_url:" "$spec" | awk '{print $2}')
    echo "Regenerating $name from $url..."
    ./printing-press generate --spec "$url" --output "/tmp/regen-$name" --force --polish
done
```

### Unit 5: Update Catalog Validation CI

**File:** `.github/workflows/validate-catalog.yml` (exists, needs update)

Add to the CI validation:
- Run the scorecard on each generated CLI
- Require Grade B+ for community PRs to merge
- Include the scorecard output in the PR comment

### Unit 6: Add `submit` Command to the Press

**File:** `internal/cli/root.go`

New command: `printing-press submit <name>`

1. Validates the catalog entry at `catalog/<name>.yaml`
2. Generates the CLI from the spec URL
3. Runs the scorecard
4. If Grade B+: creates a PR via `gh pr create`
5. Includes scorecard results in the PR body

This lets anyone submit a CLI for review without manually creating catalog YAML.

## The User Experience

### For developers (Tier 1 - install official CLIs)

```bash
brew tap mvanhorn/tap
brew install plaid-cli
plaid-cli transactions list --start-date 2026-01-01 --json
```

### For contributors (Tier 2 - submit a new CLI)

```bash
# In Claude Code:
/printing-press Airtable API
# -> Generates airtable-cli, scores it, shows the result

# Submit to catalog:
/printing-press submit airtable
# -> Opens PR, CI validates, maintainer reviews
```

### For companies (Tier 3 - self-hosted)

```bash
git clone https://github.com/mvanhorn/cli-printing-press
cd cli-printing-press
go build -o printing-press ./cmd/printing-press
./printing-press generate --spec https://internal.corp.com/api/openapi.json --polish
# -> Use internally, publish to company Homebrew tap
```

## Acceptance Criteria

- [ ] Claude Code skill updated with --docs, --polish, --lenient, scorecard, LLM brains
- [ ] `mvanhorn/homebrew-tap` repo created with README
- [ ] Plaid CLI pushed to `mvanhorn/plaid-cli` repo with tagged release
- [ ] `brew install mvanhorn/tap/plaid-cli` works on a clean Mac
- [ ] `regenerate-all.sh` script regenerates all catalog CLIs
- [ ] `printing-press submit <name>` creates a PR with scorecard results
- [ ] README documents all three tiers of distribution

## Scope Boundaries

- Do NOT push all 17 catalog CLIs to separate repos yet (just Plaid as proof)
- Do NOT set up automated regeneration on spec changes (manual for now)
- Do NOT build a web UI for the catalog (GitHub README is enough)
- GoReleaser requires GITHUB_TOKEN with repo + packages scopes
