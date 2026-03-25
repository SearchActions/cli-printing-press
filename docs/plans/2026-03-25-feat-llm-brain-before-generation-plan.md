---
title: "LLM Brain Before Generation - The Press Understands Before It Builds"
type: feat
status: active
date: 2026-03-25
---

# LLM Brain Before Generation

## The Problem

The press has a brain AFTER generation (LLM polish pass) but is dumb BEFORE. It uses regex to read API docs, GitHub API to find competitors, and templates to generate code. The result: Notion gets 6 commands because regex found 6 endpoints. An LLM reading the same docs page would find 30+.

## The Architecture

```
BEFORE (LLM understands)         TEMPLATE (generates)         AFTER (LLM polishes)
  LLM reads API docs         ->   Go templates render    ->   LLM improves help text
  LLM reads competitor READMEs     deterministic code          LLM adds examples
  LLM writes the YAML spec                                    LLM rewrites README
  LLM plans what to build
```

We built the AFTER brain. This plan builds the BEFORE brain.

## What Changes

| Step | Current (dumb) | After (smart) |
|------|---------------|---------------|
| Read API docs | Regex for `GET /path` patterns | LLM reads docs, understands every endpoint |
| Read competitor READMEs | Regex for command patterns | LLM reads README, extracts all commands |
| Write YAML spec | Regex output -> YAML | LLM writes complete spec from understanding |
| Plan what to build | Nothing - generates whatever spec has | LLM decides: "this API needs these 30 commands" |

## Implementation Units

### Unit 1: LLM-Powered Doc-to-Spec (Replace Regex)

**File:** `internal/docspec/docspec.go` (rewrite core function)

**Current:** `GenerateFromDocs` fetches HTML, runs regex, produces a spec with whatever regex finds.

**New:** `GenerateFromDocsLLM` fetches the docs page(s), sends the content to an LLM with a structured prompt, and gets back a complete YAML spec.

```go
func GenerateFromDocsLLM(docsURL, apiName string) (*spec.APISpec, error) {
    // 1. Fetch the docs page
    html, err := fetchDocs(docsURL)

    // 2. Build prompt for the LLM
    prompt := buildDocSpecPrompt(apiName, html)

    // 3. Send to LLM (claude or codex CLI)
    response, err := runLLM(prompt)

    // 4. Parse the YAML spec from the LLM response
    yamlSpec := extractYAMLFromResponse(response)

    // 5. Parse through standard spec parser
    return spec.ParseBytes([]byte(yamlSpec))
}
```

**The prompt:**

```markdown
You are generating a CLI spec for the {{apiName}} API.

Read this API documentation and produce a YAML spec in this exact format:

name: {{apiName}}
description: "one line description"
base_url: "https://api.example.com"
auth:
  type: bearer_token
  header: "Authorization"
  format: "Bearer {token}"
  env_vars:
    - APINAME_API_KEY
resources:
  resource_name:
    description: "what this resource is"
    endpoints:
      list:
        method: GET
        path: /v1/resources
        description: "List all resources"
        params:
          - name: limit
            type: integer
            description: "Max results"

API Documentation:
{{docs content - first 50K chars}}

Rules:
- Find EVERY endpoint in the docs
- Group endpoints by resource (users, projects, tasks, etc.)
- Include all query parameters and body fields
- Get the auth method right (Bearer, API key, OAuth)
- Find the correct base URL
- Output ONLY valid YAML, no markdown fences
```

**Key difference from regex:** The LLM understands context. When Notion docs say "Query a database" with a POST to `/v1/databases/{id}/query`, the LLM knows this is an endpoint with a database_id path param and a filter body param. Regex just sees `POST /v1/databases`.

**Fallback:** If no LLM is available, fall back to the existing regex-based `GenerateFromDocs`. Same behavior as today.

**Multi-page crawling:** The prompt can ask the LLM to identify links to additional API reference pages. Or we fetch the docs sitemap first and concatenate all API reference pages before sending to the LLM.

### Unit 2: LLM-Powered Competitor Analysis

**File:** `internal/pipeline/research.go` (enhance `analyzeCompetitorRepo`)

**Current:** Fetches README via GitHub API, runs regex to find command patterns, counts them.

**New:** Send the competitor's README to an LLM and ask it to extract:

```markdown
Read this CLI tool's README and tell me:

1. What commands does it have? (list each command with its description)
2. What auth methods does it support?
3. What output formats? (JSON, table, etc.)
4. What's the install method? (brew, npm, pip, etc.)
5. What features are mentioned but not implemented? (TODOs, roadmap items)
6. What do users complain about in the issues? (you can see issue titles below)

README:
{{readme content}}

Open Issues:
{{issue titles}}

Output as JSON.
```

This replaces the regex-based `parseCommandsFromReadme` with actual understanding. Instead of finding "lines that look like commands," the LLM reads the README like a developer would and extracts the real command list.

### Unit 3: LLM-Powered Spec Planning

**File:** `internal/pipeline/planner.go` (enhance `generateScaffoldPlan`)

**Current:** The dynamic planner writes a scaffold plan using research data, but the plan is a static markdown template filled with numbers.

**New:** The planner sends all research data to an LLM and asks it to write the actual scaffold plan:

```markdown
You are planning a CLI for the {{apiName}} API.

Research findings:
- Novelty score: {{score}}/10
- Competitors: {{competitors}}
- Their commands: {{competitor commands}}
- Their missing features: {{unmet features}}
- Community demand: {{demand signals}}

OpenAPI spec summary: {{endpoint count}} endpoints across {{resource count}} resources

Write a scaffold plan that:
1. Lists every resource the CLI should have
2. Lists every command per resource (CRUD + any special operations)
3. Notes which commands competitors have that we should match
4. Notes which commands no competitor has (our differentiator)
5. Sets a command count target

Output the plan as markdown.
```

This means the scaffold plan isn't just "run `printing-press generate`" - it's "generate these specific resources with these specific commands because competitors X and Y don't have Z."

### Unit 4: LLM Router (shared infrastructure)

**File:** `internal/llmpolish/llm.go` (rename from polish-specific to shared)

Actually, better to create a shared package: `internal/llm/llm.go`

```go
package llm

// Run sends a prompt to the best available LLM CLI and returns the response.
func Run(prompt string) (string, error) {
    // Try claude first (Claude Code CLI)
    if path, err := exec.LookPath("claude"); err == nil {
        return runClaude(path, prompt)
    }
    // Try codex
    if path, err := exec.LookPath("codex"); err == nil {
        return runCodex(path, prompt)
    }
    return "", fmt.Errorf("no LLM CLI found (install claude or codex)")
}

// Available returns true if any LLM CLI is installed.
func Available() bool {
    _, err1 := exec.LookPath("claude")
    _, err2 := exec.LookPath("codex")
    return err1 == nil || err2 == nil
}
```

Both the BEFORE brain (doc-to-spec, competitor analysis, planning) and the AFTER brain (polish) use this shared runner. The `llmpolish` package would import `llm` instead of having its own LLM discovery logic.

### Unit 5: Wire Into --docs and --spec Paths

**File:** `internal/cli/root.go`

When `--docs` is set:
```go
if llm.Available() {
    // Smart path: LLM reads docs and writes spec
    docSpec, err = docspec.GenerateFromDocsLLM(docsURL, apiName)
} else {
    // Dumb path: regex extracts endpoints
    docSpec, err = docspec.GenerateFromDocs(docsURL, apiName)
}
```

When `--spec` is set and research found competitors:
```go
// After generation, if LLM available, run competitor-informed enrichment
if llm.Available() && research != nil && research.CompetitorInsights != nil {
    // LLM suggests additional endpoints based on competitor analysis
    enrichments := llm.Run(buildEnrichmentPrompt(research, generatedSpec))
    // Apply enrichments to spec, regenerate
}
```

### Unit 6: Update MakeBestCLI to Use LLM Brain

**File:** `internal/pipeline/fullrun.go`

The full run sequence becomes:

```go
// Step 1: Research (GitHub API + LLM competitor analysis)
// Step 2: Generate
//   2a: If --docs, use LLM doc-to-spec (not regex)
//   2b: If --spec, use template generation
//   2c: If LLM available, run spec planning to check for missing endpoints
// Step 2.5: LLM Polish (help text, examples, README)
// Step 3: Dogfood
// Step 4: Scorecard
```

This is the FULL brain:
- LLM understands the API docs (before)
- Template generates correct code (middle)
- LLM polishes the output (after)

## Expected Impact

| API | Current (regex) | After (LLM brain) |
|-----|----------------|-------------------|
| Notion (docs) | 6 commands, 2 resources | 30+ commands, 8+ resources |
| Plaid (OpenAPI) | 51 commands (spec-limited) | 51 + enrichments from competitor analysis |
| Petstore (OpenAPI) | 8 commands | 8 (small API, nothing to add) |

The hard test (Notion) is where the LLM brain matters most. Regex found 6 endpoints. An LLM reading the same page would find every endpoint listed in the documentation.

## Token Cost Estimate

| Step | Tokens (input) | Tokens (output) | Cost (~) |
|------|---------------|-----------------|----------|
| Doc-to-spec LLM | ~30K (docs page) | ~5K (YAML spec) | $0.30 |
| Competitor README analysis | ~5K per competitor | ~1K | $0.10 |
| Spec planning | ~10K (research + spec) | ~3K (plan) | $0.15 |
| Polish (already built) | ~20K (generated code) | ~5K | $0.25 |
| **Total per CLI** | | | **~$0.80** |

Under $1 per CLI. Cheaper than a cup of coffee.

## Acceptance Criteria

- [ ] `GenerateFromDocsLLM` produces a Notion spec with 20+ endpoints (vs 6 with regex)
- [ ] Competitor analysis via LLM extracts actual command lists from READMEs
- [ ] LLM-powered scaffold plan references competitor commands and sets targets
- [ ] Shared `internal/llm` package used by both before-brain and after-brain
- [ ] Falls back to dumb mode (regex/template-only) when no LLM CLI is available
- [ ] Full run with LLM: Notion goes from 6 commands to 20+ commands
- [ ] Total cost per CLI under $2
- [ ] `go test ./...` passes

## Scope Boundaries

- Do NOT call LLM APIs directly (use claude/codex CLI)
- Do NOT make LLM required - dumb mode must still work
- Do NOT re-generate the entire CLI during polish (polish edits files, doesn't regenerate)
- Keep each LLM call under 50K input tokens (truncate docs if needed)
- Do NOT implement multi-page crawling in this plan (send what we have, let LLM work with it)
