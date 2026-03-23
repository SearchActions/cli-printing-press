---
title: "CLI Printing Press Phase 4: Catalog, Plugin, and Community"
type: feat
status: active
date: 2026-03-23
origin: docs/plans/2026-03-23-feat-cli-printing-press-plan.md
---

# CLI Printing Press Phase 4: Catalog, Plugin, and Community

## Goal

Turn cli-printing-press from a local tool into a distributable Claude Code plugin with a curated catalog of pre-built CLI definitions. Users install the plugin, browse the catalog, and get working CLIs for popular APIs without generating from scratch.

By the end of Phase 4: `/plugin marketplace add mvanhorn/cli-printing-press` installs the plugin, `/printing-press-catalog install stripe` installs a pre-built Stripe CLI, and `/printing-press submit` opens a PR to contribute back.

## What Changed Since the Master Plan

### 1. The known-specs registry is the catalog seed

Phase 3 built `skills/printing-press/references/known-specs.md` with 12 verified OpenAPI spec URLs. This is the v1 catalog - not YAML definitions, but pointers to official specs that the generator already handles. The master plan imagined hand-written `catalog/stripe.yaml` files in the internal format, but OpenAPI-first is better: less maintenance, always in sync with the real API.

**Master plan said:** 10 Official catalog entries as internal YAML definitions.
**Updated:** Catalog entries are spec URLs (OpenAPI preferred) + metadata. The generator handles the rest.

### 2. Three-tier trust model simplifies to two tiers for v1

The master plan proposed Official / Verified / Unverified tiers. For a v1 catalog with 10-15 entries, the Unverified tier adds complexity without value. Start with Official (maintainer-verified) and Community (PR-reviewed).

### 3. Security validation is partially built

Phase 3 added resource name sanitization and Swagger 2.0 detection. The remaining security work is CI-level: HTTPS enforcement on spec URLs, endpoint domain verification, and generated code scanning.

### 4. The plugin format is well-established

Claude Code plugins follow the `.claude-plugin/plugin.json` + `skills/` pattern. The existing `skills/printing-press/` directory is already the right structure. Just need to add the plugin manifest and a second skill for catalog browsing.

## What Gets Built

### New: `.claude-plugin/plugin.json`

Plugin manifest for Claude Code marketplace. Points to the skills directory.

### New: `.claude-plugin/marketplace.json`

Marketplace metadata (description, tags, screenshots, install instructions).

### New: `catalog/` directory

Catalog entries as YAML metadata files (not full API specs - just pointers):

```yaml
# catalog/stripe.yaml
name: stripe
display_name: Stripe
description: Payment processing API
category: payments
spec_url: https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json
spec_format: json
openapi_version: "3.0"
tier: official
verified_date: 2026-03-23
homepage: https://stripe.com/docs/api
notes: Very large spec (~500 endpoints). Generator truncates to 50 resources / 20 endpoints.
```

### New: `skills/printing-press-catalog/SKILL.md`

A second skill for browsing and installing from the catalog:
- `/printing-press-catalog` - list all available CLIs
- `/printing-press-catalog install <name>` - generate and build a CLI from a catalog entry
- `/printing-press-catalog search <query>` - search catalog by name/category

### Modified: `skills/printing-press/SKILL.md`

Add `/printing-press submit <name>` workflow:
1. User generates a CLI from an API
2. User tests it, confirms it works
3. `/printing-press submit stripe` creates a catalog entry YAML and opens a PR

### New: `.github/workflows/validate-catalog.yml`

CI pipeline that runs on PRs touching `catalog/`:
1. Validate catalog YAML schema (required fields present)
2. Verify spec URL is accessible (HEAD request returns 200)
3. Verify spec URL uses HTTPS
4. Download spec and run `printing-press generate --validate` to confirm it compiles
5. Check that resource names don't contain path traversal patterns

### New: `README.md`

Project README with:
- What it does (one paragraph)
- Quick start (3 commands)
- Plugin installation
- Catalog browsing
- How to contribute catalog entries
- Architecture overview

### New: `CLAUDE.md`

Project-level conventions for contributors.

## Scope Boundaries

- **No Homebrew tap** - that's a future enhancement
- **No GitHub Action for downstream regeneration** - out of scope for v1
- **No signing of catalog entries** - trust comes from PR review, not crypto
- **No web UI** - catalog is browsed via the skill, not a website
- **No automated catalog expansion** - entries are added manually via PRs

## Implementation Steps

### Step 1: Create plugin manifest

Create `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` following the standard Claude Code plugin format.

```json
// .claude-plugin/plugin.json
{
  "name": "cli-printing-press",
  "version": "0.4.0",
  "description": "Generate production Go CLIs from API descriptions or OpenAPI specs",
  "repository": "mvanhorn/cli-printing-press",
  "skills": ["printing-press", "printing-press-catalog"]
}
```

Files: `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`

### Step 2: Build the catalog directory

Create 10 catalog entry YAML files from the known-specs registry. Each entry includes: name, display_name, description, category, spec_url, spec_format, openapi_version, tier, verified_date, homepage, notes.

Start with the 12 APIs already verified in Phase 3: Petstore, Stytch, Discord, Stripe, Twilio, SendGrid, GitHub, DigitalOcean, Asana, Square, HubSpot, Front.

Files: `catalog/*.yaml` (12 files)

### Step 3: Write the catalog schema validator

A small Go program or script that validates catalog YAML files have all required fields. Can be a Go test in `internal/catalog/` or a standalone script.

Files: `internal/catalog/catalog.go`, `internal/catalog/catalog_test.go`

### Step 4: Write the `/printing-press-catalog` skill

The catalog browsing and installation skill:

**`/printing-press-catalog`** (no args) - list all entries grouped by category:
```
Payments: stripe, square
Auth: stytch
Email: sendgrid
Developer Tools: github, gitlab, digitalocean
Project Management: asana
...
```

**`/printing-press-catalog install <name>`** - find the catalog entry, download the spec, generate the CLI:
```bash
# Read catalog/<name>.yaml
# Download spec from spec_url
# Run printing-press generate --spec <downloaded> --output ./<name>-cli --validate
# Present result
```

**`/printing-press-catalog search <query>`** - grep catalog files for matching names/descriptions.

Files: `skills/printing-press-catalog/SKILL.md`

### Step 5: Add `/printing-press submit` workflow

Update the existing SKILL.md to add a submission workflow:

1. Check that the user has a working CLI in the current directory
2. Ask for the API name and metadata
3. Generate a `catalog/<name>.yaml` entry
4. Fork the repo, create branch, push the catalog entry
5. Open a PR with the catalog entry + test evidence

Files: `skills/printing-press/SKILL.md` (update)

### Step 6: Write CI validation pipeline

GitHub Actions workflow that runs on PRs touching `catalog/`:

```yaml
on:
  pull_request:
    paths: ['catalog/**']
```

Steps:
1. Install Go
2. Build printing-press binary
3. For each changed catalog file:
   - Validate YAML schema
   - HEAD request on spec_url (must return 200)
   - Verify HTTPS
   - Download spec
   - Run `printing-press generate --spec <spec> --validate`
4. Report results

Files: `.github/workflows/validate-catalog.yml`

### Step 7: Write README.md and CLAUDE.md

README: what, why, quick start, plugin install, catalog, contributing.
CLAUDE.md: commit conventions, test commands, architecture notes.

Files: `README.md`, `CLAUDE.md`

### Step 8: Test the full plugin flow

1. Install the plugin locally: copy to `~/.claude/plugins/cache/cli-printing-press/`
2. Verify `/printing-press` still works
3. Verify `/printing-press-catalog` lists entries
4. Verify `/printing-press-catalog install petstore` generates a working CLI
5. Test the CI workflow with a test PR

## Acceptance Criteria

- [ ] `/plugin marketplace add mvanhorn/cli-printing-press` installs the plugin (or manual install works)
- [ ] `/printing-press-catalog` lists all 12 catalog entries grouped by category
- [ ] `/printing-press-catalog install stripe` generates a compilable Stripe CLI
- [ ] `/printing-press-catalog install petstore` generates a compilable Petstore CLI
- [ ] `/printing-press-catalog search auth` returns Stytch
- [ ] `/printing-press submit` creates a catalog entry YAML with correct schema
- [ ] CI validates catalog entries on PRs (spec URL accessible, HTTPS, generates compilable CLI)
- [ ] README explains installation, usage, and contribution
- [ ] Existing `/printing-press` skill still works (no regressions)
- [ ] All Go tests pass

## Technical Decisions

### Catalog format: URL pointers, not embedded specs

Catalog entries point to OpenAPI spec URLs rather than containing inline YAML definitions. Advantages:
- Specs stay in sync with the real API (no stale copies)
- Smaller repo (some specs are 1MB+)
- The generator already handles OpenAPI - no need for a second format
- CI can verify URL accessibility on every PR

Disadvantage: if a URL goes stale, the catalog entry breaks. Mitigated by CI checks.

### Plugin structure

```
cli-printing-press/
  .claude-plugin/
    plugin.json
    marketplace.json
  skills/
    printing-press/
      SKILL.md
      references/
        known-specs.md
        spec-format.md
    printing-press-catalog/
      SKILL.md
  catalog/
    stripe.yaml
    stytch.yaml
    ...
  internal/
    ...
```

### Two-tier trust model

| Tier | Who adds it | Review required | CI required |
|------|-------------|-----------------|-------------|
| **Official** | Maintainer (mvanhorn) | No (self-merge) | Yes |
| **Community** | Anyone via PR | Yes (maintainer review) | Yes |

Both tiers pass the same CI validation. The difference is who added it and whether a human reviewed the entry.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Spec URLs go stale | Catalog entries break | CI checks URL accessibility. Stale entries get flagged and updated. |
| Large specs (Stripe, Discord) take long to generate | Poor UX during install | Warn user about large specs. Show progress. |
| Plugin marketplace submission process unclear | Can't distribute | Document manual install as fallback. Submit to marketplace when ready. |
| Catalog grows beyond maintainer review capacity | Review bottleneck | Start small (12 entries). Only expand when demand warrants. |
| PR submission workflow is complex | Users don't contribute | Make it dead simple: one command, auto-generate the YAML, auto-open the PR. |

## Sources

- Master plan Phase 4: `docs/plans/2026-03-23-feat-cli-printing-press-plan.md` (lines 502-513)
- Phase 3 known-specs registry: `skills/printing-press/references/known-specs.md`
- Claude Code plugin format: compound-engineering-plugin as reference (`~/.claude/plugins/`)
- Phase 1-3 implementation: 13 commits on main (454739b..f369ea6)
