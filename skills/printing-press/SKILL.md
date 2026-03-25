---
name: printing-press
description: Generate production Go CLIs from any API. 5-phase loop - research, generate, audit, fix, score. Each phase is mandatory with artifact gates.
version: 1.0.0
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - WebSearch
  - AskUserQuestion
  - Agent
---

# /printing-press

Generate a production Go CLI from any API. Five mandatory phases. No shortcuts.

```
/printing-press Notion
/printing-press Plaid payments API
/printing-press --spec ./openapi.yaml
/printing-press --spec https://raw.githubusercontent.com/.../openapi.json
```

## Prerequisites

- Go 1.21+ installed
- The printing-press repo at `~/cli-printing-press`
- Build binary if missing: `cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press`

## How This Works

Every run goes through 5 mandatory phases. You cannot skip phases. Each phase produces an artifact that the next phase needs.

```
PHASE 1: RESEARCH  ->  PHASE 2: GENERATE  ->  PHASE 3: AUDIT  ->  PHASE 4: FIX  ->  PHASE 5: SCORE
 (3-5 min)              (1-2 min)               (3-5 min)           (2-3 min)          (1 min)
```

Total expected time: 10-20 minutes. If a run completes in under 5 minutes, something was skipped.

---

## Workflow: `--spec` shortcut

When the user provides `--spec <path-or-url>`, skip Phase 1 (spec is provided). Run Phases 2-5.

## Workflow: Natural Language (Primary)

When the user provides an API name, run ALL five phases.

### Step 0: Parse intent and check known specs

Extract the API name from the user's message. Check `~/cli-printing-press/skills/printing-press/references/known-specs.md` for a known spec URL.

If found in registry: note the URL for Phase 2, but STILL run Phase 1 research (competitors, demand signals).
If not found: Phase 1 will also search for the spec.

---

# PHASE 1: RESEARCH

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Research the API landscape before generating anything. You need to know what exists so you can beat it.

### Step 1.1: Search for the OpenAPI spec

If not found in known-specs registry:

1. **WebSearch**: `"<API name>" openapi spec site:github.com`
2. **WebSearch**: `"<API name>" openapi.yaml OR openapi.json specification`
3. Try common URL patterns:
   - `https://raw.githubusercontent.com/<org>/openapi/main/openapi.yaml`
   - `https://api.<domain>/openapi.json`
4. If a URL is found, **WebFetch** the first 500 bytes to verify it contains `openapi:` or `"openapi"` or `swagger:`

If no spec found after these searches: plan to write one from docs in Phase 2.

### Step 1.2: Search for competing CLIs

**WebSearch**: `"<API name>" CLI tool github`
**WebSearch**: `"<API name>" command line client`

For each competitor found:
- Note the repo URL, star count, language
- **WebFetch** their README to count commands and note features
- Look for: how many resources/commands, what auth methods, output formats, any unique features

### Step 1.3: Check demand signals

**WebSearch**: `"<API name>" "need a CLI" OR "command line" OR "CLI tool" site:reddit.com OR site:news.ycombinator.com`

Note any posts asking for a CLI for this API. This tells us there's demand.

### Step 1.4: Write the research artifact

**Write** to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-research.md`:

```markdown
---
title: "Research: <API> CLI"
type: feat
status: active
date: <today>
---

# Research: <API> CLI

## Spec Discovery
- Official OpenAPI spec: <url or "none found - will write from docs">
- Source: <where found - registry, GitHub search, etc.>
- Format: <OpenAPI 3.x / Swagger 2.0 / internal YAML>

## Competitors
| Name | Stars | Language | Commands | Notable Features |
|------|-------|----------|----------|-----------------|
| <name> | <stars> | <lang> | <count> | <features> |

## Auth Method
- Type: <api_key / oauth2 / bearer_token>
- Header: <name>
- Env var convention: <what competitors use, e.g. NOTION_API_KEY>

## Demand Signals
- <Reddit/HN posts, or "none found">

## Recommendation
- Spec source: <OpenAPI URL / write from docs>
- Target command count: <N - match or beat best competitor>
- Key differentiator: agent-native (--json, --select, --dry-run, --stdin, typed exit codes)
```

### PHASE GATE 1

**STOP.** Before proceeding to Phase 2, verify:
1. The research artifact file exists at `docs/plans/<today>-feat-<api>-cli-research.md`
2. It has a Spec Discovery section with a URL or "write from docs" decision
3. It has at least one competitor listed (or explicitly "no competitors found")

Tell the user: "Phase 1 complete: Found [spec/no spec], [N] competitors. Best competitor: [name] with [N] commands. Proceeding to generation."

---

# PHASE 2: GENERATE

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Generate the CLI using the printing-press binary.

### Step 2.1: Get the spec ready

**If OpenAPI spec was found:**
```bash
curl -sL -o /tmp/printing-press-spec-<api>.json "<spec-url>" && head -c 200 /tmp/printing-press-spec-<api>.json
```

**If no spec (write from docs):**
1. **WebFetch** the API documentation URL
2. **Read** `~/cli-printing-press/skills/printing-press/references/spec-format.md`
3. Write a YAML spec to `/tmp/<api>-spec.yaml` following the format exactly
4. Include ALL endpoints found in the docs - resources, methods, paths, params, body fields, auth

### Step 2.2: Check for existing output directory

```bash
cd ~/cli-printing-press && ls -la <api>-cli 2>/dev/null && echo "EXISTS" || echo "CLEAN"
```

If EXISTS: remove it first (`rm -rf <api>-cli`).

### Step 2.3: Run the generator

```bash
cd ~/cli-printing-press && ./printing-press generate \
  --spec /tmp/printing-press-spec-<api>.json \
  --output ./<api>-cli \
  --force --lenient --validate 2>&1
```

If the binary doesn't exist:
```bash
cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press
```

### Step 2.4: Handle quality gate failures

If gates fail: read the error output carefully. Common fixes:
- Spec description issues: clean the spec, retry
- Module errors: `cd <api>-cli && go mod tidy`
- Template errors: check if the spec has unusual types

Max 3 retries. If still failing, present the error to the user.

### PHASE GATE 2

**STOP.** Before proceeding to Phase 3, verify:
1. The directory `~/cli-printing-press/<api>-cli/` exists
2. `cd ~/cli-printing-press/<api>-cli && go build ./...` succeeds
3. All 7 quality gates passed (or you know which warnings are acceptable)

Tell the user: "Phase 2 complete: Generated <api>-cli with [N] resources, [M] endpoints. All quality gates passed. Proceeding to audit."

---

# PHASE 3: AUDIT

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Code-review the generated CLI against the research findings. This is where you find problems.

### Step 3.1: Read the generated code

You MUST actually read these files (not just check they exist):

```
Read ~/cli-printing-press/<api>-cli/internal/cli/root.go
Read ~/cli-printing-press/<api>-cli/README.md
```

Also read at least 2 resource command files:
```
Read ~/cli-printing-press/<api>-cli/internal/cli/<resource1>.go
Read ~/cli-printing-press/<api>-cli/internal/cli/<resource2>.go
```

### Step 3.2: Count commands and compare

Run:
```bash
cd ~/cli-printing-press/<api>-cli && grep -r "Use:" internal/cli/*.go | grep -v "root.go" | wc -l
```

Compare this count against the best competitor from Phase 1 research.

### Step 3.3: Check help text quality

Read the command files and check:
- Are descriptions developer-friendly or raw OpenAPI spec jargon?
- Do examples use realistic values or placeholder garbage like "string", "0", "example"?
- Does the root command description explain what the API does?

### Step 3.4: Check for missing endpoints

Compare the API's actual endpoints (from spec or docs) against what was generated. Look for:
- Major resources that are missing entirely
- Important CRUD operations that were skipped
- Pagination endpoints that were missed

### Step 3.5: Check agent-native features

Verify these are present in root.go:
- `--json` flag
- `--select` flag
- `--dry-run` flag
- `--stdin` flag
- `--yes` flag
- `--no-cache` flag
- `doctor` subcommand

### Step 3.6: Review README quality

Read the generated README.md. Check:
- Does it explain how to install?
- Does it have realistic examples?
- Does it mention auth setup?
- Does it document output formats?
- Would you actually use this CLI based on the README?

### Step 3.7: Write the audit artifact

**Write** to `~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-audit.md`:

```markdown
---
title: "Audit: <API> CLI"
type: fix
status: active
date: <today>
---

# Audit: <API> CLI

## Command Comparison
- Our CLI: <N> commands across <M> resources
- Best competitor: <name> with <N> commands
- Gap: <what we're missing, or "we match/beat them">

## Help Text Quality
- [ ] Descriptions are developer-friendly: <PASS/FAIL + details>
- [ ] Examples use realistic values: <PASS/FAIL + details>
- [ ] Resource descriptions explain what the resource is: <PASS/FAIL + details>

## Agent-Native Checklist
- [x/] --json flag present
- [x/] --select flag present
- [x/] --dry-run flag present
- [x/] --stdin flag present
- [x/] Typed exit codes
- [x/] doctor command

## README Quality
- <honest assessment>

## Specific Fixes Needed
1. File: `internal/cli/<file>.go` - Change: <what to fix and why>
2. File: `README.md` - Change: <what to fix and why>
...

## Fixes NOT Needed (Regeneration Required)
- <anything that can't be fixed with Edit and requires re-running the generator>
```

### PHASE GATE 3

**STOP.** Before proceeding to Phase 4, verify:
1. The audit artifact exists at `docs/plans/<today>-fix-<api>-cli-audit.md`
2. It has specific fixes listed (file paths + what to change)
3. You actually Read the generated code files (not just ran commands)

Tell the user: "Phase 3 complete: Found [N] issues to fix. [summary of top issues]. Proceeding to fixes."

If the audit found ZERO issues (rare but possible): still write the artifact noting "No fixes needed - generated output meets quality bar." Then proceed to Phase 5 (skip Phase 4).

---

# PHASE 4: FIX

## THIS PHASE IS MANDATORY (unless audit found zero issues).

Execute the specific fixes identified in Phase 3's audit.

### Step 4.1: Fix each issue from the audit

For each fix listed in the audit artifact:

1. **Read** the file that needs fixing
2. **Edit** the specific lines (use the Edit tool, not Write - surgical changes)
3. Verify the edit is correct

Common fixes:
- **Help text jargon**: Edit `Short` and `Long` descriptions in command files to be developer-friendly
- **Bad examples**: Edit `Example` fields with realistic API values
- **Missing resource description**: Edit the resource command's `Short` field
- **README weak**: Edit sections of README.md to be more useful
- **Missing endpoints**: Add new command entries (this may require more extensive edits)

### Step 4.2: Verify compilation after ALL fixes

```bash
cd ~/cli-printing-press/<api>-cli && go build ./... && go vet ./... && echo "FIXES VERIFIED"
```

If compilation fails after edits: read the error, fix it, retry.

### PHASE GATE 4

**STOP.** Before proceeding to Phase 5, verify:
1. All fixes from the audit have been applied
2. `go build ./...` passes in the CLI directory
3. `go vet ./...` passes

Tell the user: "Phase 4 complete: Applied [N] fixes. Compilation verified. Proceeding to scoring."

---

# PHASE 5: SCORE + REPORT

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

### Step 5.1: Quality assessment

Score the CLI against these 8 dimensions (0-10 each):

| Dimension | What to check | Score |
|-----------|---------------|-------|
| Output modes | --json, --select, --plain, --human-friendly all present? | /10 |
| Auth | Correct auth type, doctor validates credentials, env var convention? | /10 |
| Error handling | Typed exit codes (0,2,3,4,5,7,10), helpful error messages? | /10 |
| Terminal UX | Color support, pagination, progress indicators? | /10 |
| README | Install instructions, examples, troubleshooting, sells the tool? | /10 |
| Doctor | Health check validates auth, API reachability? | /10 |
| Agent-native | Non-interactive, pipeable, cacheable, idempotent creates/deletes? | /10 |
| Breadth | Command count vs competitors, missing major endpoints? | /10 |

Total: /80. Grade: 65+ = A, 50-64 = B, <50 = needs work.

Optionally, also try running the Go scorecard:
```bash
cd ~/cli-printing-press && SCORECARD_CLI_DIR=./<api>-cli SCORECARD_PIPELINE_DIR=/tmp/<api>-score \
  go test ./internal/pipeline/ -run TestScorecardOnRealCLI -v -timeout 60s 2>&1 | tail -30
```

### Step 5.2: Present the final report

Show ALL of these sections to the user:

**1. Summary table:**
```
Generated <api>-cli with <N> resources and <M> endpoints.

Resources: <comma-separated list>
```

**2. Quality score:**
```
Steinberger Score: <total>/80 (Grade <A/B/C>)

| Dimension | Score | Notes |
|-----------|-------|-------|
| Output modes | X/10 | ... |
| Auth | X/10 | ... |
...
```

**3. Competitor comparison:**
```
Found <N> competing CLIs.
Best competitor: <name> (<stars> stars, <commands> commands)
We beat them on: <what we do better>
We're missing: <what they have that we don't, or "nothing - we match or beat">
```

**4. Example commands:**
```bash
cd ~/cli-printing-press/<api>-cli
go install ./cmd/<api>-cli

export <AUTH_ENV_VAR>="..."

<api>-cli --help
<api>-cli <resource> list
<api>-cli <resource> get --<id_param> <realistic_value>
<api>-cli doctor
```

**5. Spec source:**
```
Spec: <source description and URL>
```

**6. Limitations (if any):**
- Complex body fields skipped as CLI flags
- Any endpoints that couldn't be represented
- GraphQL APIs wrapped in YAML spec

---

## Writing Specs from Docs

When no OpenAPI spec exists and you need to write one:

1. **WebFetch** the API documentation URL
2. **Read** `~/cli-printing-press/skills/printing-press/references/spec-format.md`
3. Read the docs and identify EVERY endpoint: method, path, description, params, body fields, auth
4. Write YAML spec to `/tmp/<api>-spec.yaml` following the format exactly
5. Generate from it

This is where you ARE the brain. No shelling out to an LLM. Read the docs yourself and write the spec.

## Submit to Catalog

`/printing-press submit <name>`

1. Gather metadata (name, description, category, spec_url, tier)
2. Write `catalog/<name>.yaml`
3. `git checkout -b catalog/<name> && git add catalog/<name>.yaml && git commit && gh pr create`

## Safety Gates

- **Preview before generating**: Show API name, base URL, estimated resource count before running
- **Output directory conflict**: Check before overwriting
- **Untrusted specs**: Note if not from known-specs registry
- **Max 3 retries** on quality gate failure

## Limitations

- Go CLIs only (no Bash, Python, etc.)
- OpenAPI 3.0+ and Swagger 2.0 supported
- 50 resources / 50 endpoints per resource limit
- No native GraphQL support (but can wrap GraphQL in YAML spec)
- Complex body fields (nested objects, arrays) skipped as CLI flags
