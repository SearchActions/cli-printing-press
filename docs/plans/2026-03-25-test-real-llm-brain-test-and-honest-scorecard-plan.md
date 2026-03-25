---
title: "Turn On the Brain - Real LLM Test and Honest Scorecard"
type: test
status: active
date: 2026-03-25
---

# Turn On the Brain

## The Problem

We built three LLM brains but never turned them on. Every test mocks the LLM or skips it. The Notion CLI has 6 commands because the regex found 6 endpoints - but the LLM brain was supposed to find 30+. We don't know if it works because we've never tried.

The scorecard is also lying. It measures template features, not CLI quality. A 6-command Notion CLI scores the same as a 51-command Plaid CLI. That's not honest scoring.

## What This Plan Does

1. Actually call the LLM (Claude CLI) during generation and measure the difference
2. Fix the scorecard so it penalizes empty CLIs
3. Run the full test with LLM on vs LLM off and compare

## Implementation Units

### Unit 1: Verify Claude CLI Works for LLM Calls

Before anything else, verify the `internal/llm` package can actually call Claude.

```bash
# Check claude is installed
which claude

# Test a simple prompt
echo "Say hello" | claude -p "Say hello in one word"
```

If `claude` CLI doesn't accept `-p` for prompt, fix `internal/llm/llm.go` to use the correct flags. The current code guesses at flags - we need to verify what actually works.

**File:** `internal/llm/llm.go`

Test: `go test ./internal/llm/ -run TestRunWithRealLLM -v` (new test, gated behind `LLM_TEST=1`)

```go
func TestRunWithRealLLM(t *testing.T) {
    if os.Getenv("LLM_TEST") == "" {
        t.Skip("Set LLM_TEST=1 to test with real LLM")
    }
    if !Available() {
        t.Skip("No LLM CLI available")
    }
    response, err := Run("Respond with exactly the word 'hello' and nothing else.")
    require.NoError(t, err)
    assert.Contains(t, strings.ToLower(response), "hello")
}
```

**Verification:** `LLM_TEST=1 go test ./internal/llm/ -run TestRunWithRealLLM -v` prints "hello".

### Unit 2: Test Doc-to-Spec with Real LLM on Notion

The big proof. Does the LLM brain actually find more endpoints than regex?

```bash
# Regex mode (current - finds 6 endpoints)
./printing-press generate --docs "https://developers.notion.com/reference" --name notion --output /tmp/notion-regex --force

# LLM mode (should find 20+)
./printing-press generate --docs "https://developers.notion.com/reference" --name notion --output /tmp/notion-llm --force
# (LLM mode activates automatically when claude CLI is available)
```

**Measure the difference:**
- Count commands in `/tmp/notion-regex/internal/cli/*.go`
- Count commands in `/tmp/notion-llm/internal/cli/*.go`
- Both must pass 7/7 quality gates

**If the LLM doc-to-spec fails:** Fix the prompt in `internal/docspec/docspec.go`. The prompt might need:
- Better format instructions (show the exact YAML format expected)
- Truncation handling (40K char limit might cut off important endpoints)
- Multi-page support (Notion docs span many pages)

**File:** `internal/docspec/docspec.go` (fix prompts if needed)

**Verification:** Notion LLM CLI has 15+ commands (3x what regex finds). Both pass 7/7 gates.

### Unit 3: Test LLM Polish with Real LLM

Test that `--polish` actually improves the help text and examples.

```bash
# Generate without polish
./printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json --output /tmp/petstore-raw --force

# Generate with polish
./printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json --output /tmp/petstore-polished --force --polish
```

**Measure the difference:**
- Compare `Short:` descriptions in `/tmp/petstore-raw/internal/cli/*.go` vs polished
- Compare `Example:` strings
- Compare README.md content
- Polished should have better descriptions and more examples

**If polish fails:** Fix prompts in `internal/llmpolish/prompts.go`. Common issues:
- LLM returns markdown-wrapped JSON instead of raw JSON
- LLM returns too-long descriptions
- LLM hallucinates flag names that don't exist

**Verification:** At least 3 help descriptions are different (improved) between raw and polished.

### Unit 4: Fix Scorecard - Add Breadth Dimension

The scorecard needs to penalize empty CLIs. A 6-command Notion CLI should NOT score the same as 51-command Plaid.

**File:** `internal/pipeline/scorecard.go`

Add a new dimension:

```go
type SteinerScore struct {
    // ... existing 8 dimensions ...
    Breadth int `json:"breadth"` // 0-10: how many commands vs expected
}
```

Scoring:
- Count commands in the generated CLI (files in internal/cli/ minus infrastructure)
- Compare against a baseline:
  - < 5 commands: 0/10 (basically empty)
  - 5-10: 3/10 (minimal)
  - 11-20: 5/10 (decent)
  - 21-40: 7/10 (good)
  - 41-60: 9/10 (comprehensive)
  - 60+: 10/10 (Steinberger-tier)

Update total to be out of 90 (not 80). Update grade thresholds accordingly.

**Also update the comparison table** in `fullrun.go` to show the breadth dimension.

**Verification:** Notion regex (6 commands) scores 0/10 on breadth. Plaid (51 commands) scores 9/10. Grade reflects this.

### Unit 5: Full Run With LLM vs Without

**File:** `internal/pipeline/fullrun_test.go` (extend)

New test: `TestFullRunWithLLM` (gated behind `FULL_RUN_LLM=1`)

Runs the same 3 APIs (petstore, plaid, notion) but with LLM enabled. Compares results against the non-LLM baseline.

```go
func TestFullRunWithLLM(t *testing.T) {
    if os.Getenv("FULL_RUN_LLM") == "" {
        t.Skip("Set FULL_RUN_LLM=1 to run with real LLM calls")
    }
    // ... same as TestFullRun but with LLM available
    // Compare: notion should have 3x more commands with LLM
}
```

Output: side-by-side table showing LLM vs no-LLM for each API.

```
                    | Without LLM        | With LLM
--------------------|--------------------|-----------------
Notion Commands     | 6                  | 25+
Notion Breadth      | 0/10               | 7/10
Notion Grade        | B (inflated)       | A (earned)
Plaid Commands      | 51                 | 51 (same - spec-based)
Petstore Commands   | 8                  | 8 (same - spec-based)
```

The LLM brain should only improve doc-to-spec APIs (Notion). OpenAPI-based APIs (Plaid, Petstore) should be unchanged since the spec already has all endpoints.

## Expected Token Cost

| LLM Call | Tokens | Cost |
|----------|--------|------|
| Notion doc-to-spec | ~30K in, ~5K out | ~$0.30 |
| Petstore polish (help text) | ~10K in, ~2K out | ~$0.10 |
| Petstore polish (examples) | ~10K in, ~3K out | ~$0.15 |
| Petstore polish (README) | ~5K in, ~3K out | ~$0.10 |
| **Total for full run** | | **~$0.65** |

## Acceptance Criteria

- [ ] `internal/llm/llm.go` correctly calls `claude` CLI (verified with real call)
- [ ] Notion CLI with LLM doc-to-spec has 15+ commands (vs 6 with regex)
- [ ] Both Notion CLIs (regex and LLM) pass 7/7 quality gates
- [ ] `--polish` visibly improves at least 3 help descriptions on Petstore
- [ ] Scorecard breadth dimension penalizes 6-command CLIs (0-3/10)
- [ ] Scorecard breadth rewards 51-command CLIs (9-10/10)
- [ ] Full run comparison table shows LLM vs no-LLM side by side
- [ ] Total LLM cost for full run under $2

## Scope Boundaries

- Do NOT fix doc-to-spec to crawl multiple pages (just use what's on the docs URL)
- Do NOT change the template engine (this tests the LLM brains, not the templates)
- Do NOT run LLM tests in regular `go test ./...` (gate behind env vars)
- If Claude CLI flags are wrong, fix them - but don't build a fallback to API calls
