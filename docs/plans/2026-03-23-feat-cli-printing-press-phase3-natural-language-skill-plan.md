---
title: "CLI Printing Press Phase 3: Natural Language Mode + Claude Code Skill"
type: feat
status: completed
date: 2026-03-23
origin: docs/plans/2026-03-23-feat-cli-printing-press-plan.md
---

# CLI Printing Press Phase 3: Natural Language Mode

## Goal

Ship a Claude Code skill (`/printing-press`) that turns a single sentence into a working CLI:

```
/printing-press Stytch authentication API
```

The agent finds an OpenAPI spec (or reads API docs and writes one), runs `printing-press generate`, and hands the user a compiled CLI binary. No spec file required.

By the end of Phase 3: `/printing-press Loops email API` produces a working `loops-cli` binary from just that sentence.

## What Changed Since the Master Plan

Phase 1 and 2 revealed things the master plan didn't anticipate:

### 1. OpenAPI specs are the fast path, not internal YAML

Generating internal YAML from docs is fragile - the agent must infer types, auth schemes, pagination patterns, and response shapes from prose. OpenAPI specs encode all of this structurally. The skill should exhaust all spec-finding strategies before falling back to YAML generation.

**Master plan said:** Agent reads API docs and generates internal spec.
**Updated:** Agent searches for OpenAPI spec first (APIs.guru, GitHub, common URL patterns). Only generates internal YAML as fallback.

### 2. Real specs are messy but the parser handles it

Stytch's OpenAPI spec (47K lines) generated a CLI that compiles on the first try. The resource/endpoint limits (50/20) work correctly - they cap output size and log warnings. The `oneline` template fix handles multiline descriptions. This means: for APIs with published OpenAPI specs, the pipeline is already end-to-end.

**Master plan said:** 5 real-world OpenAPI specs tested.
**Updated:** 3 tested (Petstore, Stytch, Discord all compile). Clerk has no public spec. The skill should track which APIs have known-good specs.

### 3. The binary needs to be pre-built

The skill invokes `printing-press generate` as a shell command. The binary must exist on PATH or at a known location. This is a setup step the master plan didn't address.

### 4. Quality gate failures need a retry loop

When the agent generates YAML from docs, the first attempt often produces specs that fail `go vet` or `go build`. The skill needs a structured retry: read error, fix spec, regenerate. Max 2 retries.

## What Gets Built

### New: `~/.claude/skills/printing-press/SKILL.md`

The Claude Code skill with three workflows:
1. **Natural language** (primary): `/printing-press Stytch auth API`
2. **From spec file**: `/printing-press --spec ./openapi.yaml`
3. **From URL**: `/printing-press --spec https://api.example.com/openapi.yaml`

### New: `~/.claude/skills/printing-press/references/spec-format.md`

Reference doc for the agent describing the internal YAML format with a complete annotated example. Loaded on-demand when the agent needs to generate YAML from docs.

### New: `~/.claude/skills/printing-press/references/known-specs.md`

Registry of APIs with known-good OpenAPI spec URLs. The agent checks this first before searching. Starts with 10-15 entries.

### Modified: `internal/openapi/detect.go`

Add Swagger 2.0 detection (`"swagger:"` keyword) so the parser doesn't misroute Swagger specs.

### Modified: `internal/openapi/parser.go`

Add filename sanitization on resource names to prevent path traversal from untrusted specs.

### New: `internal/cli/discover.go` (optional - stretch goal)

A `printing-press discover <api-name>` subcommand that searches for OpenAPI specs online and prints the URL. Useful standalone, and the skill can invoke it.

## Scope Boundaries (What Phase 3 Does NOT Include)

- **Catalog/community submissions** - that's Phase 4
- **Plugin packaging** (`.claude-plugin/plugin.json`) - Phase 4
- **Multi-language output** (Bash, Python) - the CLI only generates Go
- **OAuth2 flow generation** - simplified to bearer_token
- **GraphQL API support** - OpenAPI/REST only
- **Spec caching between invocations** - each run starts fresh
- **Auto-install of generated binary** - print the command, don't execute it

## Implementation Steps

### Step 1: Fix detect.go for Swagger 2.0

Add `"swagger:"` detection to `IsOpenAPI()`. Trivial change, prevents misclassification of Swagger 2.0 specs.

Files: `internal/openapi/detect.go`, `internal/openapi/parser_test.go`
Execution note: Test-first. Add a Swagger 2.0 test case to TestIsOpenAPI, verify it fails, then fix.

### Step 2: Add resource name sanitization

Validate that `resourceNameFromPath` cannot produce path traversal sequences. Add sanitization: strip `..`, `/`, `\`, and any non-alphanumeric-underscore characters from resource names before using them as filenames.

Files: `internal/openapi/parser.go`, `internal/openapi/parser_test.go`
Execution note: Test-first. Write tests with adversarial paths like `../../../etc/passwd`, `foo/bar`, `foo\bar`.

### Step 3: Build the printing-press binary

Ensure the binary is built and accessible. Add an install target to the Makefile if one doesn't exist. The skill will reference `~/cli-printing-press/printing-press` or `$GOPATH/bin/printing-press`.

Files: Check existing Makefile, update if needed.
Execution note: `go build -o ./printing-press ./cmd/printing-press`

### Step 4: Create the known-specs registry

Write `references/known-specs.md` with 15 APIs that have publicly available OpenAPI specs. For each entry: API name, spec URL, format (YAML/JSON), verified date, notes.

Sources to check:
- APIs.guru directory (https://apis.guru)
- GitHub repos for API providers (search for `openapi.yaml` or `openapi.json`)
- Official API documentation pages

Include at minimum: Petstore, Stytch, Discord, Stripe, Twilio, SendGrid, GitHub, GitLab, DigitalOcean, Asana, Square, Notion, Linear, HubSpot, Front.

Files: `~/.claude/skills/printing-press/references/known-specs.md`

### Step 5: Create the spec-format reference

Write `references/spec-format.md` with:
- The complete internal YAML schema (all fields, types, required/optional)
- An annotated example based on `testdata/stytch.yaml`
- Common mistakes and how to avoid them
- Validation rules from `spec.Validate()`

This is what the agent reads when it needs to generate YAML from API docs.

Files: `~/.claude/skills/printing-press/references/spec-format.md`

### Step 6: Write the SKILL.md

The core deliverable. Three workflows, progressive disclosure, safety gates.

**Workflow 0: Natural Language (Primary)**
```
User: /printing-press <api description>

Agent:
1. Parse intent - extract API name, optional resource filter
2. Check known-specs.md for a cached URL
3. If not found: search online (WebSearch for "<api-name> OpenAPI spec github")
4. If spec URL found:
   a. Download with WebFetch
   b. Save to temp file
   c. Run: printing-press generate --spec <file> --output ./<name>-cli
   d. If quality gates pass: present result
   e. If gates fail: read error, attempt spec fix, retry once
5. If no spec found:
   a. Read API documentation with WebFetch
   b. Read references/spec-format.md
   c. Generate internal YAML spec
   d. Save spec to ./<name>-spec.yaml
   e. Run: printing-press generate --spec ./<name>-spec.yaml
   f. If gates fail: read error, fix YAML, retry (max 2 attempts)
   g. Present result with spec file for user to review/modify
6. Show: binary location, command examples, how to customize
```

**Workflow 1: From Spec File**
```
User: /printing-press --spec ./path/to/spec.yaml

Agent:
1. Verify file exists
2. Run: printing-press generate --spec <file>
3. Present result
```

**Workflow 2: From URL**
```
User: /printing-press --spec https://example.com/openapi.yaml

Agent:
1. Download with WebFetch
2. Save to temp file
3. Run: printing-press generate --spec <file>
4. Present result
```

**Safety gates in SKILL.md:**
- Preview: show the API name, base URL, and resource count before generating
- Confirm: if output directory exists, ask before overwriting
- Warn: if quality gates fail after retries, present the error and the generated project for manual inspection

**Frontmatter:**
```yaml
name: printing-press
description: Generate production Go CLIs from API descriptions or OpenAPI specs
version: 0.3.0
allowed-tools: [Bash, Read, Write, Glob, Grep, WebFetch, WebSearch, AskUserQuestion]
```

Files: `~/.claude/skills/printing-press/SKILL.md`

### Step 7: Test the skill end-to-end

Manual testing in a fresh Claude Code session:
1. `/printing-press Petstore API` - should find the OpenAPI spec and generate
2. `/printing-press Loops email API` - may need to generate YAML from docs
3. `/printing-press --spec testdata/openapi/stytch.yaml` - direct spec path
4. `/printing-press nonexistent-api-xyz` - should fail gracefully

Fix issues discovered during testing.

### Step 8: Sync skill to repo

Copy the skill files into the cli-printing-press repo under a `skills/` directory so they're versioned alongside the generator.

```
cli-printing-press/
  skills/
    printing-press/
      SKILL.md
      references/
        known-specs.md
        spec-format.md
```

Files: `skills/printing-press/SKILL.md`, etc.

## Acceptance Criteria

- [ ] `/printing-press Petstore API` generates a compilable CLI from the canonical OpenAPI spec
- [ ] `/printing-press Stytch auth API` finds the real Stytch OpenAPI spec and generates
- [ ] `/printing-press --spec testdata/stytch.yaml` still works (internal format)
- [ ] `/printing-press --spec testdata/openapi/petstore.yaml` still works (OpenAPI)
- [ ] Swagger 2.0 specs are correctly detected by `IsOpenAPI()`
- [ ] Resource names from untrusted specs cannot cause path traversal
- [ ] Quality gate failures trigger a retry with error-informed spec fixes
- [ ] Graceful failure when no spec or docs are found
- [ ] The agent informs the user about truncated resources (50/20 limits)
- [ ] Generated spec files are saved for user review when YAML is generated from docs
- [ ] The skill works in a fresh Claude Code session with no prior context

## Technical Decisions

### Spec discovery strategy (priority order)

1. **Known-specs registry** - instant, verified, no network latency
2. **GitHub search** - `gh search repos "<api-name> openapi" --json` or WebSearch for `"<api-name>" openapi.yaml site:github.com`
3. **Common URL patterns** - try `https://api.<domain>/openapi.json`, `https://developer.<domain>/openapi.yaml`
4. **APIs.guru** - check `https://api.apis.guru/v2/list.json` (2400+ specs)
5. **Web search fallback** - broad search for API documentation

### Internal YAML generation strategy

When no OpenAPI spec exists, the agent:
1. Fetches the API's documentation page(s) with WebFetch
2. Reads `references/spec-format.md` for the schema
3. Generates YAML matching the internal format
4. Runs `spec.Validate()` mentally (checks name, base_url, resources, method, path)
5. Saves to `<name>-spec.yaml` alongside the generated project
6. On failure: reads the error message, fixes the YAML, retries (max 2)

### Binary location

The skill runs `printing-press` from the repo directory:
```bash
cd ~/cli-printing-press && go run ./cmd/printing-press generate --spec <file> --output <dir>
```

Using `go run` avoids the build/install step entirely. Slightly slower (adds ~1s) but always uses the latest code. No PATH management needed.

### Output directory

Default: `./<name>-cli` in the current working directory (matching the CLI's default behavior). The user can override with `--output`.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Agent generates invalid YAML from docs | CLI won't compile | Structured retry loop with error parsing. Save spec for manual editing. |
| OpenAPI spec URLs go stale | Known-specs registry has broken links | Include verified date. The skill falls back to search if known URL fails. |
| WebFetch blocked by API docs sites | Can't read documentation | Fall back to asking user to paste relevant docs or provide a spec file. |
| Large specs (500+ endpoints) overwhelmed | Resource/endpoint limits truncate silently | Agent checks parsed resource count and warns user about truncation. |
| Adversarial OpenAPI spec from untrusted source | Path traversal in generated files | Sanitize resource names. The generator writes to a user-specified output dir. |
| Agent hallucinates API endpoints when generating YAML | Generated CLI has nonexistent commands | Save the spec for user review. Quality gates catch compilation issues but not semantic correctness. |

## What Phase 3 Enables for Phase 4

Once the skill works, Phase 4 (Catalog + Community) builds on it:
- The known-specs registry becomes the seed for the Official catalog tier
- The spec-format reference becomes the contributor guide
- The submit workflow becomes: generate with skill, test, PR the spec to the catalog
- Plugin packaging wraps the skill for `/plugin marketplace add`

## Sources

- Phase 1 plan: `docs/plans/2026-03-23-feat-cli-printing-press-phase1-template-engine-plan.md`
- Phase 2 plan: `docs/plans/2026-03-23-feat-cli-printing-press-phase2-openapi-parser-plan.md`
- Master plan: `docs/plans/2026-03-23-feat-cli-printing-press-plan.md`
- APIs.guru: https://apis.guru (2400+ OpenAPI spec directory)
- Existing skill patterns: `~/.claude/skills/last30days/SKILL.md`, `~/.claude/skills/osc-work/SKILL.md`
- SpecFlow analysis identified 20 gaps and 14 critical questions (all addressed above)
