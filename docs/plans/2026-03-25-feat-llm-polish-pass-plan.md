---
title: "LLM Polish Pass - Make the Press Smart, Not Just Fast"
type: feat
status: active
date: 2026-03-25
---

# LLM Polish Pass

## The Insight

The printing press generates CLIs from templates. Templates give you the 80% - correct code structure, auth, error handling, output formats. But the last 20% requires understanding: better help text, smarter examples, missing endpoints, a README that actually sells the tool.

LLMs are good at that 20%. The press should have a two-pass architecture:

```
Pass 1: Template (fast, free, deterministic)
  - Parse spec -> generate Go code from templates
  - Takes 10 seconds, costs $0
  - Produces a CLI that compiles and passes quality gates

Pass 2: LLM Polish (smart, costs tokens, non-deterministic)
  - Review the generated code with an LLM
  - Improve descriptions, examples, README
  - Add missing endpoints from docs
  - Takes 2-5 minutes, costs ~$0.50-2.00
  - Produces a CLI that's actually good to use
```

## Where LLM Polish Helps Most

### 1. Help Text Rewriting

**Before (template - copies spec verbatim):**
```
$ plaid-cli transactions list --help
Retrieve a paginated listing of transactions associated with a given access token
```

**After (LLM polish):**
```
$ plaid-cli transactions list --help
List transactions for a linked bank account. Returns the most recent transactions
first. Use --start-date and --end-date to filter by date range.
```

The LLM reads the spec description and rewrites it for a developer reading --help at 2am.

### 2. Example Generation

**Before (template - generic values):**
```
Examples:
  plaid-cli transactions list --count 25
```

**After (LLM polish):**
```
Examples:
  # Get last 30 days of transactions
  plaid-cli transactions list --start-date 2026-02-25 --end-date 2026-03-25

  # Get transactions for a specific account
  plaid-cli transactions list --account-id acc_abc123 --count 100

  # Export as JSON for processing
  plaid-cli transactions list --start-date 2026-01-01 --json > transactions.json
```

The LLM understands what developers actually DO with the command and writes examples for real workflows.

### 3. Missing Endpoint Discovery

**Before (template - only what's in the spec):**
Notion CLI from doc-to-spec: 6 commands (agents, threads)

**After (LLM polish):**
LLM reads the Notion API docs page, identifies ALL endpoints (databases, pages, blocks, search, users, comments), writes the missing endpoint definitions, regenerates.

This is the doc-to-spec problem solved properly. Instead of regex, the LLM reads the docs and understands them.

### 4. README That Sells

**Before (template):**
```markdown
# plaid-cli

The Plaid REST API. Please see https://plaid.com/docs/api for more details.

## Quick Start
...
```

**After (LLM polish):**
```markdown
# plaid-cli

Connect to bank accounts, pull transactions, and verify identities - all from your terminal.
No more spinning up a React app just to test your Plaid integration.

## Why This Exists

Plaid has no official CLI. The community CLI (plaid-cli, 57 stars) was abandoned in 2021.
This CLI covers 51 endpoints and works with sandbox, development, and production environments.

## Quick Start
...
```

## Implementation Units

### Unit 1: LLM Polish Interface

**File:** `internal/llmpolish/polish.go` (new package)

Define the interface - what goes in, what comes out:

```go
type PolishRequest struct {
    APIName     string
    OutputDir   string     // path to generated CLI
    SpecSource  string     // OpenAPI spec or docs URL
    Research    *pipeline.ResearchResult  // competitor insights
}

type PolishResult struct {
    HelpTextsImproved   int
    ExamplesAdded       int
    EndpointsAdded      int
    READMERewritten     bool
    TokensUsed          int
    Duration            time.Duration
}

func Polish(req PolishRequest) (*PolishResult, error)
```

The `Polish` function does NOT call an LLM API directly. Instead, it writes a structured prompt to a file and shells out to the LLM tool (Claude Code, Codex, or any CLI that accepts a prompt). This keeps the press LLM-agnostic.

**How it shells out:**

```go
prompt := buildPolishPrompt(req)
promptFile := filepath.Join(os.TempDir(), "polish-prompt.md")
os.WriteFile(promptFile, []byte(prompt), 0644)

// Try Claude Code first, fall back to Codex
cmd := exec.Command("claude", "--print", "-p", prompt)
output, err := cmd.CombinedOutput()
if err != nil {
    cmd = exec.Command("codex", "--quiet", prompt)
    output, err = cmd.CombinedOutput()
}
```

### Unit 2: Help Text Polish Prompt

**File:** `internal/llmpolish/prompts.go`

Build a prompt that asks the LLM to improve help descriptions:

```markdown
You are improving the --help text for a CLI tool called {{.APIName}}-cli.

For each command below, rewrite the "Short" description to be:
- Written for a developer reading --help at 2am
- Under 80 characters
- Starts with a verb (List, Get, Create, Update, Delete)
- No jargon from the API spec

Current commands and descriptions:
{{range .Commands}}
- {{.Name}}: "{{.CurrentDescription}}"
{{end}}

Output format: JSON array of {name, description} objects.
```

The LLM returns improved descriptions. The polish function reads the generated Go files, finds the `Short:` strings, and replaces them.

### Unit 3: Example Polish Prompt

Build a prompt that asks the LLM to write real-world examples:

```markdown
You are writing --help examples for {{.APIName}}-cli.

For each command, write 2-3 examples showing real workflows a developer would use.
Each example should have a comment explaining what it does.

Commands:
{{range .Commands}}
- {{.Name}} ({{.Method}} {{.Path}})
  Flags: {{.Flags}}
{{end}}

Output: JSON array of {command, examples: [{comment, invocation}]}
```

### Unit 4: README Polish Prompt

Rewrite the generated README with understanding:

```markdown
Rewrite this CLI's README to sell the tool. You are writing for developers
who will find this on GitHub.

API: {{.APIName}}
Commands: {{.CommandCount}}
Competitors: {{.Competitors}}
Killer workflow: {{.KillerWorkflow}}

Current README:
{{.CurrentREADME}}

Write a README with:
1. A one-line hook that makes developers want to install it
2. Why this exists (what gap it fills)
3. Quick start with a real 3-command workflow
4. The killer workflow (the one command that makes it worth installing)
5. Full command reference
```

### Unit 5: Endpoint Discovery Prompt (for doc-to-spec)

When the press generates from docs and gets few endpoints:

```markdown
Read this API documentation and list ALL endpoints.

URL: {{.DocsURL}}
Content: {{.DocsHTML}}

For each endpoint, output:
- method (GET/POST/PUT/PATCH/DELETE)
- path (e.g. /v1/databases/{id})
- description (one line)
- parameters (name, type, required)

Current spec has {{.CurrentEndpoints}} endpoints. Find more.
```

### Unit 6: Wire Polish into MakeBestCLI

**File:** `internal/pipeline/fullrun.go` (modify)

Add a polish step between Generate and Dogfood:

```go
// Step 2: Generate (template pass)
// ... existing code ...

// Step 2.5: LLM Polish (if available)
if llmAvailable() {
    polishResult, err := llmpolish.Polish(llmpolish.PolishRequest{
        APIName:    apiName,
        OutputDir:  outputDir,
        SpecSource: specURL,
        Research:   result.Research,
    })
    if err != nil {
        result.Errors = append(result.Errors, "LLM polish failed: "+err.Error())
    } else {
        result.PolishResult = polishResult
    }
}

// Step 3: Dogfood
// ... existing code ...
```

Polish is optional. If no LLM CLI is available, the press skips it and generates a template-only CLI (what it does today). If an LLM is available, it polishes the output.

### Unit 7: Add --polish Flag to Generate Command

**File:** `internal/cli/root.go`

```go
var polish bool
cmd.Flags().BoolVar(&polish, "polish", false, "Run LLM polish pass on generated CLI (requires claude or codex CLI)")
```

When `--polish` is set, the generate command runs the polish step after template generation.

### Unit 8: Polish Scorecard Dimension

**File:** `internal/pipeline/scorecard.go`

Add a new dimension to the scorecard that detects LLM polish:

```go
type SteinerScore struct {
    // ... existing 8 dimensions ...
    Polish int `json:"polish"` // 0-10: quality of descriptions, examples, README
}
```

Polish dimension measures:
- Help descriptions: are they under 80 chars and start with a verb? (2pts)
- Examples: does each command have 2+ examples? (3pts)
- README: does it have a "Why This Exists" section? (2pts)
- README: does it have a killer workflow section? (3pts)

This dimension ONLY scores well with LLM polish. Template-only CLIs score 0-2/10 here.

## The Full Architecture After This Plan

```
printing-press generate --spec <url> --polish
  |
  v
  [Template Pass - 10 seconds, $0]
  Parse spec -> Go templates -> compile -> 7 quality gates
  |
  v
  [LLM Polish Pass - 2-5 minutes, ~$1]
  Improve help text -> Add examples -> Rewrite README -> Add missing endpoints
  |
  v
  [Dogfood + Scorecard]
  Run commands -> Score against Steinberger -> Grade
```

## Acceptance Criteria

- [ ] `llmpolish` package with Polish() function that shells out to claude/codex
- [ ] Help text prompt: rewrites spec jargon into developer-friendly descriptions
- [ ] Example prompt: generates 2-3 real-world examples per command
- [ ] README prompt: rewrites README to sell the tool
- [ ] Endpoint discovery prompt: finds missing endpoints from docs
- [ ] --polish flag on generate command
- [ ] MakeBestCLI calls polish when LLM is available
- [ ] New scorecard dimension measures polish quality
- [ ] Petstore with --polish has better help text than without
- [ ] Notion with --polish has more than 6 commands

## Scope Boundaries

- Do NOT build an LLM API client - shell out to existing CLIs (claude, codex)
- Do NOT make polish mandatory - template-only still works
- Do NOT call the LLM during compilation or template rendering
- Polish edits generated FILES, not templates
- Keep total polish cost under $2 per CLI
