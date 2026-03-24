---
name: printing-press
description: Generate production Go CLIs from API descriptions or OpenAPI specs. Say an API name and get a compiled CLI binary. Supports autonomous multi-phase pipeline mode.
version: 0.4.0
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - WebFetch
  - WebSearch
  - AskUserQuestion
  - Skill
  - Agent
  - Edit
  - CronCreate
  - CronList
  - CronDelete
---

# /printing-press

Generate a production Go CLI from an API description or spec file.

## Quick Start

```
/printing-press Stytch authentication API
/printing-press --spec ./openapi.yaml
/printing-press --spec https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json
```

## Prerequisites

- Go 1.21+ installed (`go version`)
- The printing-press repo at `~/cli-printing-press`

## Workflows

### Workflow 0: Natural Language (Primary)

When the user provides an API name or description (no --spec flag):

**Step 1: Parse intent**
Extract the API name from the user's message. Examples:
- "Stytch authentication API" -> API name: "Stytch"
- "Stripe payments" -> API name: "Stripe"
- "the Loops email API" -> API name: "Loops"

**Step 2: Check known-specs registry**
Read `~/cli-printing-press/skills/printing-press/references/known-specs.md` and search for the API name.

If found: use the spec URL from the registry. Go to Step 4.

**Step 3: Search for OpenAPI spec online**
If not in registry, search in this order:

1. WebSearch: `"<api-name>" openapi spec site:github.com`
2. WebSearch: `"<api-name>" openapi.yaml OR openapi.json`
3. Try common URL patterns:
   - `https://raw.githubusercontent.com/<org>/openapi/main/openapi.yaml`
   - `https://api.<domain>/openapi.json`

If a URL is found, verify it's accessible with a brief WebFetch check (first 200 bytes should contain `openapi:` or `"openapi"` or `swagger:`).

If no spec found: go to Step 6 (generate from docs).

**Step 4: Download and generate from OpenAPI spec**

```bash
# Download the spec
curl -sL -o /tmp/printing-press-spec-$$.yaml "<spec-url>"

# Generate the CLI
cd ~/cli-printing-press && ./printing-press generate \
  --spec /tmp/printing-press-spec-$$.yaml \
  --output ./<name>-cli \
  --validate
```

If the binary doesn't exist, build it first:
```bash
cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press
```

If all 7 quality gates pass: go to Step 7 (present result).
If gates fail: go to Step 5 (retry).

**Step 5: Retry on quality gate failure**

Read the error output carefully. Common fixes:
- "newline in string" -> descriptions have unescaped newlines (should not happen with current templates, but if it does, the spec description needs cleaning)
- "undefined" -> a template function is missing a type mapping
- Module errors -> run `go mod tidy` in the output directory

If the error is in the spec (not the templates):
1. Read the spec file
2. Fix the issue (e.g., remove problematic descriptions, fix types)
3. Delete the output directory
4. Re-run `printing-press generate`

Max 2 retries. If still failing after retries, present the error to the user and suggest they inspect the generated code.

**Step 6: Generate internal YAML spec from documentation**

When no OpenAPI spec exists:

1. Read the API documentation using WebFetch
2. Read `~/cli-printing-press/skills/printing-press/references/spec-format.md` for the YAML format
3. Generate a YAML spec following the format exactly. Include:
   - name (kebab-case)
   - base_url (from API docs)
   - auth config (type, header, env_vars)
   - resources (group endpoints logically)
   - endpoints (method, path, params, body)
   - types (for response objects)
4. Save to `./<name>-spec.yaml`
5. Run `printing-press generate --spec ./<name>-spec.yaml`
6. On failure: read error, fix YAML, retry (max 2 attempts)
7. Present result with note: "Generated spec saved to ./<name>-spec.yaml - you can edit and regenerate."

**Step 7: Present result**

Show the user:
1. What was generated (directory name, number of resources/endpoints)
2. Example commands they can try
3. How to install: `cd <name>-cli && go install ./cmd/<name>-cli`
4. If resource/endpoint limits were hit, note what was truncated

Example output:
```
Generated stytch-cli with 8 resources and 42 endpoints.

Try it:
  cd stytch-cli
  go install ./cmd/stytch-cli
  stytch-cli --help
  stytch-cli users list --limit 10
  stytch-cli doctor

All 7 quality gates passed.
```

### Workflow 1: From Spec File

When the user provides `--spec <local-path>`:

1. Verify the file exists
2. Run `cd ~/cli-printing-press && ./printing-press generate --spec <path> [--output <dir>]`
3. Present result (Step 7 above)

### Workflow 2: From URL

When the user provides `--spec <url>`:

1. Download with `curl -sL -o /tmp/printing-press-spec-$$.yaml "<url>"`
2. Run `cd ~/cli-printing-press && ./printing-press generate --spec /tmp/printing-press-spec-$$.yaml [--output <dir>]`
3. Present result (Step 7 above)

### Workflow 3: Submit to Catalog

When the user invokes `/printing-press submit <name>`:

**Step 1: Gather metadata**
Ask the user for:
- API display name (e.g., "Stripe")
- One-line description
- Category (payments, auth, email, developer-tools, project-management, communication, crm, example)
- Homepage URL
- OpenAPI spec URL (if they used one)

**Step 2: Create catalog entry**
Write a YAML file to `~/cli-printing-press/catalog/<name>.yaml`:

```yaml
name: <name>
display_name: <display_name>
description: <description>
category: <category>
spec_url: <spec_url>
spec_format: <yaml|json>
openapi_version: "3.0"
tier: community
verified_date: "<today's date>"
homepage: <homepage>
notes: <any notes>
```

**Step 3: Open a PR**
```bash
cd ~/cli-printing-press
git checkout -b catalog/<name>
git add catalog/<name>.yaml
git commit -m "feat(catalog): add <display_name> catalog entry"
git push -u origin catalog/<name>
gh pr create --title "feat(catalog): add <display_name>" --body "Adds catalog entry for <display_name> API.

Spec URL: <spec_url>
Category: <category>
Tier: community

Tested locally - generated CLI compiles and passes all quality gates."
```

**Step 4: Present the PR URL**
Show the user the PR link and note that CI will validate the entry.

### Workflow 4: Autonomous Pipeline

| Phase | Purpose |
|-------|---------|
| 0. Preflight | Validate environment and discover the OpenAPI spec |
| 1. Scaffold | Generate the initial CLI and verify quality gates |
| 2. Enrich | Deep-read the spec and research missing hints |
| 3. Regenerate | Merge enrichments and re-generate the CLI |
| 4. Review | Static quality checks + autonomous dogfooding against real API |
| 5. Ship | Initialize git repo, commit, write report |

The Review phase dogfoods the generated CLI in three tiers:
- **Tier 1** (always): version, doctor, help, dry-run, output modes - no credentials needed
- **Tier 2** (if credentials available): list, get, auth error handling - read-only API calls
- **Tier 3** (sandbox APIs only): create/delete roundtrip with cleanup - write operations on safe test servers

Results feed a combined quality score (static 0-50 + dogfood 0-50 = total 0-100).

When the user says "print <api-name>":

**Step 1: Initialize**
- Build the press binary: `go build -o ./printing-press ./cmd/printing-press`
- Run: `./printing-press print <api-name> [--output <dir>] [--force]`
- This creates `docs/plans/<api-name>-pipeline/` with 6 plan seeds + `state.json`
- Show the user: API name, spec URL, output directory, and phase list

**Step 2: Heartbeat safety net**
Schedule a heartbeat to resume if the session dies unexpectedly:
```bash
RESUME_MIN=$(date -v+45M '+%M')
RESUME_HOUR=$(date -v+45M '+%H')
RESUME_DAY=$(date -v+45M '+%d')
RESUME_MONTH=$(date -v+45M '+%m')
```
Then CronCreate with cron expression `$RESUME_MIN $RESUME_HOUR $RESUME_DAY $RESUME_MONTH *` and prompt: `/printing-press resume <api-name>`

**Step 3: Phase execution loop**
For each phase in state.json where `plan_status` is not "completed":

a. Read the plan file at the phase's `plan_path`
b. Check the `status` field in the plan's YAML frontmatter:
   - If `status: seed` - the plan is a thin prompt that needs expansion:
     1. Run `Skill("compound-engineering:ce:plan", plan_path)` to expand the seed into a full plan with research
     2. ce:plan overwrites the file with the expanded plan (frontmatter becomes `status: active`)
     3. Update state.json: set `plan_status` to "expanded":
     ```bash
     python3 -c "
     import json
     s = json.load(open('docs/plans/<api-name>-pipeline/state.json'))
     s['phases']['<phase>']['plan_status'] = 'expanded'
     json.dump(s, open('docs/plans/<api-name>-pipeline/state.json', 'w'), indent=2)
     "
     ```
   - If `status: active` - the plan is already expanded, ready for execution:
     1. Run `Skill("compound-engineering:ce:work", plan_path)` to implement, test, and check off criteria
     2. Update state.json: mark phase completed:
     ```bash
     python3 -c "
     import json
     s = json.load(open('docs/plans/<api-name>-pipeline/state.json'))
     s['phases']['<phase>']['status'] = 'completed'
     s['phases']['<phase>']['plan_status'] = 'completed'
     json.dump(s, open('docs/plans/<api-name>-pipeline/state.json', 'w'), indent=2)
     "
     ```
c. Run budget gate (Step 4)
d. If more phases remain and budget gate says CONTINUE:
   - CronCreate 30 seconds from now: `/printing-press resume <api-name>`
   - Print: `"[phase] complete. Chaining to [next_phase] in 30s with fresh context..."`
   - END SESSION (the cron fires a new session with fresh context)

**Note:** Each phase may take TWO sessions - one for ce:plan expansion, one for ce:work execution. This is by design: ce:plan gets a fresh context window for research, ce:work gets a fresh context window for implementation.

**Step 4: Budget gate (between every phase)**
```bash
python3 -c "
import json, datetime
s = json.load(open('docs/plans/<api-name>-pipeline/state.json'))
started = datetime.datetime.fromisoformat(s['started_at'].replace('Z', '+00:00'))
elapsed = (datetime.datetime.now(datetime.timezone.utc) - started).total_seconds() / 3600
print('STOP' if elapsed > 3 else 'CONTINUE')
print(f'Elapsed: {elapsed:.1f}h / 3h budget')
"
```
If STOP: go to Step 6 (morning report), do NOT chain.

**Step 5: Error handling**
- If ce:plan fails on a phase: log error to state.json errors array, retry once. If still fails, mark phase "failed", skip to next phase.
- If ce:work fails: retry once with same plan. If still fails, mark "failed", skip to next.
- If 2+ consecutive phases fail: write morning report and STOP. Do not chain.
- NEVER die silently. Always update state.json before ending a session.
- If CronCreate fails: print manual resume command as fallback: `claude "/printing-press resume <api-name>"`

**Step 6: Morning report**
Write `docs/plans/<api-name>-pipeline/report.md`:
```markdown
# Pipeline Report: <api-name>

- **API:** <api-name>
- **Spec URL:** <spec_url from state.json>
- **Output:** <output_dir from state.json>
- **Phases completed:** N/M
- **Total elapsed:** Xh Ym
- **Status:** completed | budget-exhausted | failed

## Phase Results
| Phase | Status | Notes |
|-------|--------|-------|
| preflight | completed | ... |
| scaffold | completed | ... |
| ... | ... | ... |

## Next Steps
- [ ] Review generated CLI at <output_dir>
- [ ] Run `<cli-name> doctor` to verify
- [ ] Submit to catalog: `/printing-press submit <api-name>`
```

### Workflow 5: Resume Pipeline

When the user says "resume <api-name>" or the `--resume` flag is detected:

1. Load `docs/plans/<api-name>-pipeline/state.json`
2. **MANDATORY:** Run budget gate FIRST (Workflow 4 Step 4) before any work
3. If STOP: write morning report if not already written, print summary, exit
4. If CONTINUE:
   - Show status table: which phases are done, which is next
   - Delete stale heartbeat crons: CronList, then CronDelete any matching `/printing-press resume`
   - Schedule new heartbeat (45 min from now, same as Workflow 4 Step 2)
   - Go to Workflow 4 Step 3 (phase execution loop) starting from the next incomplete phase

## Safety Gates

- **Preview before generating**: Show the API name, base URL, and estimated resource count before running the generator
- **Output directory conflict**: If the output directory already exists, ask the user before overwriting
- **Untrusted specs**: If the spec was downloaded from a URL not in the known-specs registry, note this to the user

## Limitations

- Only generates Go CLIs (no Bash, Python, etc.)
- OpenAPI 3.0+ and Swagger 2.0 supported; other formats are not
- Large APIs (500+ endpoints) are automatically truncated to 50 resources / 50 endpoints per resource
- No GraphQL support
- Pipeline mode requires Compound Engineering plugin (`compound-engineering:ce:plan` and `compound-engineering:ce:work`)
- Pipeline budget gate is 3 hours max
