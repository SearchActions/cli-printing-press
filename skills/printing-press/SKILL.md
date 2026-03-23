---
name: printing-press
description: Generate production Go CLIs from API descriptions or OpenAPI specs. Say an API name and get a compiled CLI binary.
version: 0.3.0
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - WebFetch
  - WebSearch
  - AskUserQuestion
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

## Safety Gates

- **Preview before generating**: Show the API name, base URL, and estimated resource count before running the generator
- **Output directory conflict**: If the output directory already exists, ask the user before overwriting
- **Untrusted specs**: If the spec was downloaded from a URL not in the known-specs registry, note this to the user

## Limitations

- Only generates Go CLIs (no Bash, Python, etc.)
- OpenAPI 3.0+ and Swagger 2.0 supported; other formats are not
- Large APIs (500+ endpoints) are automatically truncated to 50 resources / 20 endpoints per resource
- OAuth2 flows are simplified to bearer_token
- No GraphQL support
- Generated CLIs do not include retry/rate-limiting (basic HTTP client only)
