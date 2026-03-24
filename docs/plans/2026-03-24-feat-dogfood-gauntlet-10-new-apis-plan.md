---
title: "Dogfood Gauntlet - Generate and Test 10 New API CLIs"
type: feat
status: completed
date: 2026-03-24
---

# Dogfood Gauntlet - Generate and Test 10 New API CLIs

## Overview

The printing press has been tested against 4 APIs (Petstore, Stytch, Discord, Gmail). That's not enough. We need to throw diverse, real-world specs at it and see what breaks. Not Steinberger's APIs (Google suite, Discord, WhatsApp) - those are covered. Not the ones already in our registry. Fresh specs that stress different parts of the generator.

The goal: run `printing-press generate` against 10 new APIs, capture what passes and what breaks, fix the generator (not the output), and add the working ones to the catalog.

## The 10 APIs

Selected for diversity of auth, spec size, resource nesting, and pagination patterns:

| # | API | Spec URL | Size | Auth | Free? | Why It's Interesting |
|---|-----|----------|------|------|-------|---------------------|
| 1 | Cloudflare | `https://raw.githubusercontent.com/cloudflare/api-schemas/main/openapi.json` | 1716 paths | API key, Bearer | Yes | Largest spec - stress test truncation, zone-scoped resources |
| 2 | Jira | `https://api.apis.guru/v2/specs/atlassian.com/jira/1001.0.0-SNAPSHOT/openapi.json` | 317 paths | OAuth2, Basic | Yes | Deep nesting (project/board/sprint/issue/comment), JQL |
| 3 | Spotify | `https://api.apis.guru/v2/specs/spotify.com/sonallux/2023.2.27/openapi.json` | 68 paths | OAuth2 | Yes | Cursor pagination, nested resources, search with filters |
| 4 | Telegram Bot | `https://api.apis.guru/v2/specs/telegram.org/5.0.0/openapi.json` | 74 paths | Token-in-URL | Yes | File uploads, inline keyboards, webhook setup |
| 5 | Trello | `https://api.apis.guru/v2/specs/trello.com/1.0/openapi.json` | 264 paths | API key+token | Yes | Board/list/card/checklist nesting, batch endpoint |
| 6 | Vercel | `https://api.apis.guru/v2/specs/vercel.com/0.0.1/openapi.json` | 85 paths | Bearer, OAuth2 | Yes | Deploy lifecycle, env vars, domain mgmt |
| 7 | Sentry | `https://raw.githubusercontent.com/getsentry/sentry-api-schema/main/openapi-derefed.json` | 126 paths | Bearer | Yes | Org/project/issue hierarchy, cursor pagination |
| 8 | Fly.io Machines | `https://docs.machines.dev/spec/openapi3.json` | 51 paths | Bearer | Yes | Machine lifecycle, volumes, leases, long-poll waits |
| 9 | LaunchDarkly | `https://app.launchdarkly.com/api/v2/openapi.json` | 221 paths | API key | Yes (trial) | Feature flags, semantic patching, experiments |
| 10 | Supabase | `https://api.supabase.com/api/v1-json` | 105 paths | Bearer | Yes | Project lifecycle, database mgmt, edge functions |

## Acceptance Criteria

- [ ] All 10 specs downloaded and parsed without crash
- [ ] At least 7/10 pass all 7 quality gates (compile, vet, build, binary, help, version, doctor)
- [ ] Bugs found in the generator are documented with spec name and reproduction steps
- [ ] Each passing CLI has been dogfooded: `--help`, `doctor`, at least 1 resource `--help`
- [ ] Working CLIs added to catalog/ as community tier entries
- [ ] Known specs registry updated with all 10 URLs
- [ ] Bugs-found list written to docs/plans/dogfood-gauntlet-findings.md

## Implementation Units

### Unit 1: Add All 10 to Known Specs Registry

**Files:** `internal/pipeline/discover.go`

Add the 10 new APIs to `KnownSpecs` map. None are sandbox-safe (no Tier 3 dogfooding).

### Unit 2: Run the Gauntlet

**Approach:** For each API, in order from smallest to largest:

```bash
# Build the press
go build -o ./printing-press ./cmd/printing-press

# For each API:
./printing-press generate \
  --spec <spec-url> \
  --output /tmp/dogfood-<name>-cli \
  2>&1 | tee /tmp/dogfood-<name>-output.txt

# If 7 gates pass:
cd /tmp/dogfood-<name>-cli
go build -o <name>-cli ./cmd/<name>-cli
./<name>-cli --help
./<name>-cli doctor
./<name>-cli <first-resource> --help

# Record: pass/fail, gate output, help output, any crashes
```

Run order (smallest first, fail fast):
1. Fly.io (51 paths)
2. Spotify (68 paths)
3. Telegram (74 paths)
4. Vercel (85 paths)
5. Supabase (105 paths)
6. Sentry (126 paths)
7. LaunchDarkly (221 paths)
8. Trello (264 paths)
9. Jira (317 paths)
10. Cloudflare (1716 paths)

### Unit 3: Fix Generator Bugs Found

**Files:** `internal/openapi/parser.go`, `internal/generator/templates/*.tmpl`

For each API that fails:
1. Read the error output
2. Identify root cause in the parser or templates (NOT in the generated output)
3. Fix the generator
4. Run `go test ./...` (don't break existing specs)
5. Re-run the failing API
6. Repeat until it passes or document as known limitation

Common expected failures:
- Huge specs hitting resource/endpoint caps (Cloudflare at 1716 paths)
- Complex auth schemes not mapping cleanly (Telegram token-in-URL)
- Deep nesting exceeding sub-resource detection (Jira 5-level nesting)
- Unusual parameter types or content types

### Unit 4: Add Passing CLIs to Catalog

**Files:** `catalog/<name>.yaml` for each passing API

Create catalog entry with:
- name, display_name, description, category, spec_url, spec_format
- tier: community
- verified_date: 2026-03-24

### Unit 5: Write Findings Report

**Files:** `docs/plans/dogfood-gauntlet-findings.md`

Document:
- Pass/fail per API with gate output
- Bugs found and fixed (with commit SHAs)
- Bugs found but NOT fixed (known limitations)
- Generator improvements needed
- Quality score per CLI (if dogfood phase was run)
- Recommendations for the Steinberger Parity plan

## Scope Boundaries

- Fix the generator, never the generated output
- Don't call any real APIs (no auth needed, just generate and compile)
- Don't run the full pipeline (`printing-press print`) - just `generate`
- Don't test all endpoints of each CLI - just help, doctor, and one resource
- If a spec can't be downloaded (404, rate limit), skip it and note why
- Max 2 retries per spec after generator fixes

## Success Metrics

- **7+ of 10 pass all gates** = the press handles diverse specs well
- **5-6 pass** = generator needs targeted fixes for common patterns
- **< 5 pass** = fundamental parser/template gaps to address

## Sources

- Known specs registry: `internal/pipeline/discover.go`
- Catalog entries: `catalog/*.yaml`
- Generator: `internal/generator/generator.go`, `internal/generator/templates/`
- Parser: `internal/openapi/parser.go`
- Quality gates: `internal/generator/validate.go`
- apis.guru directory: https://apis.guru/
