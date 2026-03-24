# CLI Printing Press

One command. Any API. A production CLI walks out the other side.

Peter Steinberger hand-crafts [gogcli](https://github.com/steipete/gogcli), [discrawl](https://github.com/steipete/discrawl), [wacli](https://github.com/steipete/wacli) - beautiful Go CLIs that take weeks each. The printing press does it in 60 seconds. Point it at any OpenAPI spec in the world and get a finished CLI with auth, pagination, tables, retries, color, doctor, completions. Not a scaffold. Not a starting point. A finished product.

```bash
./printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json
```

```
PASS go mod tidy
PASS go vet ./...
PASS go build ./...
PASS build runnable binary
PASS petstore-cli --help
PASS petstore-cli version
PASS petstore-cli doctor
Generated petstore-cli at ./petstore-cli
```

Seven quality gates. All green. Ship it.

## The 60-Second Demo

```bash
# Build the press
git clone https://github.com/mvanhorn/cli-printing-press.git
cd cli-printing-press
go build -o ./printing-press ./cmd/printing-press

# Print a CLI from any OpenAPI spec
./printing-press generate \
  --spec https://petstore3.swagger.io/api/v3/openapi.json \
  --output ./petstore-cli

# Use it
cd petstore-cli && go build -o petstore-cli ./cmd/petstore-cli
./petstore-cli --help
```

```
Usage:
  petstore-cli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  doctor      Check CLI health
  help        Help about any command
  pet         Everything about your Pets
  store       Access to Petstore orders
  user        Operations about user
  version     Print version

Flags:
      --config string      Config file path
      --dry-run            Show request without sending
  -h, --help               help for petstore-cli
      --json               Output as JSON
      --no-color           Disable colored output
      --plain              Output as plain tab-separated text
      --quiet              Bare output, one value per line
      --timeout duration   Request timeout (default 30s)
  -v, --version            version for petstore-cli
```

```bash
./petstore-cli doctor
```

```
  OK Config: ok
  FAIL Auth: not configured
  OK API: reachable (HTTP 404)
  config_path: ~/.config/petstore-cli/config.toml
  base_url: https://petstore3.swagger.io/api/v3
  version: 1.0.27
```

Every resource, every endpoint, every flag - auto-generated from the spec.

```bash
./petstore-cli pet --help
```

```
Available Commands:
  add              Add a new pet to the store.
  delete           Deletes a pet.
  find-by-status   Finds Pets by status.
  find-by-tags     Finds Pets by tags.
  get-by-id        Find pet by ID.
  update           Update an existing pet.
  update-with-form Updates a pet in the store with form data.
  upload-image     Manage upload image
```

## What Gets Generated

Every printed CLI ships with everything a Steinberger-quality CLI needs:

### Commands

- **Cobra subcommand tree** - `<name>-cli <resource> <action>` hierarchy, auto-generated from spec paths
- **Sub-resource grouping** - nested API paths become `resource sub-resource action`
- **doctor** - formatted health dashboard with status checks, config paths, and version
- **version** - semantic version from config
- **Shell completions** - bash, zsh, fish, powershell via Cobra

### Output

- **Table output** - list responses auto-detected from array schemas, rendered as formatted tables
- **`--json`** - raw JSON for piping and scripting
- **`--plain`** - tab-separated values for `awk`/`cut`
- **`--quiet`** - bare values, one per line
- **Color and TTY detection** - ANSI color when interactive, plain when piped. Respects `NO_COLOR`, `TERM=dumb`, and `--no-color`

### Auth

- **API key** - header or query parameter
- **Bearer token** - standard Authorization header
- **Basic auth** - username:password
- **OAuth2** - full browser-based authorization flow with token refresh, authorization URL, and token URL support

### Reliability

- **Retry with exponential backoff** - automatic retry on 5xx responses
- **Rate limit detection** - 429 status + Retry-After header handling
- **Structured exit codes** - 0 (ok), 1 (general error), 3 (not found), 4 (auth error), 7 (rate limited), 130 (interrupted)
- **Request timeout** - configurable via `--timeout` flag (default 30s)
- **`--dry-run`** - preview mutation requests without sending them

### Pagination

- **Auto-detected from spec** - `--limit` and `--all` flags generated when pagination is detected
- **Cursor-based and offset-based** - handles both pagination styles

### Build Tooling

- **go.mod** - clean module with pinned dependencies
- **Makefile** - build, test, lint, install targets
- **.goreleaser.yaml** - cross-platform release automation
- **.golangci.yml** - linter configuration
- **README.md** - auto-generated documentation for the CLI itself
- **Config** - TOML config at `~/.config/<name>-cli/config.toml`

## Usage

### From OpenAPI URL

```bash
./printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json
```

Specs are cached locally for 24 hours at `~/.cache/printing-press/specs/`.

### From Local File

```bash
./printing-press generate --spec ./my-api-spec.yaml --output ./my-cli
```

### From Internal YAML Format

For APIs without OpenAPI specs, write a simpler YAML format:

```bash
./printing-press generate --spec ./my-api.yaml
```

See `skills/printing-press/references/spec-format.md` for the YAML format reference.

### Multi-Spec Composition

Merge multiple APIs into a single CLI:

```bash
./printing-press generate \
  --spec https://api1.example.com/openapi.json \
  --spec https://api2.example.com/openapi.json \
  --name combined-cli
```

Auth configuration is taken from the first spec that defines OAuth2.

### Supported Formats

- **OpenAPI 3.0+** (YAML or JSON)
- **Swagger 2.0** (YAML or JSON)
- **Internal YAML** (hand-written, simpler schema)

Auto-detection reads the first 500 bytes and routes to the correct parser.

## Autonomous Pipeline

For complex APIs, the press doesn't just generate - it thinks.

```bash
./printing-press print gmail
```

```
Pipeline created for gmail
  Spec: https://gmail.googleapis.com/$discovery/rest?version=v1
  Output: ./gmail-cli
  Plans:
    0. docs/plans/gmail-pipeline/00-preflight-plan.md
    1. docs/plans/gmail-pipeline/01-scaffold-plan.md
    2. docs/plans/gmail-pipeline/02-enrich-plan.md
    3. docs/plans/gmail-pipeline/03-regenerate-plan.md
    4. docs/plans/gmail-pipeline/04-review-plan.md
    5. docs/plans/gmail-pipeline/05-ship-plan.md
```

This kicks off a 6-phase autonomous pipeline. Each phase runs in a fresh AI session, chains automatically, and tracks state across sessions:

| Phase | What Happens |
|-------|-------------|
| Preflight | Verify Go installed, download spec, cache project conventions |
| Scaffold | Generate CLI, pass all 7 quality gates |
| Enrich | Research API docs, discover missing endpoints, auth flows, pagination patterns |
| Regenerate | Merge enrichments back into the spec, regenerate with full knowledge |
| Review | Lint, test, benchmark, fix issues found |
| Ship | Build, tag, generate release notes |

The pipeline uses [Compound Engineering](https://github.com/EveryInc/compound-engineering-plugin) for planning and execution. Each phase produces a plan that gets expanded with parallel research agents, then executed with full test coverage.

Budget gate stops at 3 hours. Heartbeat safety net resumes if a session dies. Morning report tells you what happened.

### Known APIs

16 APIs are pre-registered and discoverable by name:

| API | Spec Format |
|-----|-------------|
| Petstore | OpenAPI 3.0 |
| Stripe | OpenAPI 3.0 |
| Square | OpenAPI 3.0 |
| Stytch | OpenAPI 3.0 |
| Discord | OpenAPI 3.1 |
| Twilio | OpenAPI 3.0 |
| SendGrid | OpenAPI 3.0 |
| GitHub | OpenAPI 3.0 |
| DigitalOcean | OpenAPI 3.0 |
| Asana | OpenAPI 3.0 |
| HubSpot | OpenAPI 3.0 |
| Front | OpenAPI 3.0 |

Unknown APIs fall back to the [apis-guru](https://apis.guru/) directory.

## Claude Code Plugin

```
/plugin marketplace add mvanhorn/cli-printing-press
```

Then just say what you want:

```
/printing-press Stripe payments API
/printing-press Discord bot management
/printing-press --spec ./openapi.yaml
/printing-press print gmail
```

The skill finds the spec (or writes one from API docs), generates the CLI, and presents the result.

### Catalog Skill

Browse and install pre-built CLIs:

```
/printing-press-catalog
/printing-press-catalog search payments
/printing-press-catalog install stripe
```

## Catalog

12 pre-built CLI definitions for popular APIs:

| API | Category | Spec Format |
|-----|----------|-------------|
| Stripe | Payments | OpenAPI 3.0 |
| Square | Payments | OpenAPI 3.0 |
| Stytch | Auth | OpenAPI 3.0 |
| Discord | Communication | OpenAPI 3.1 |
| Twilio | Communication | OpenAPI 3.0 |
| SendGrid | Email | OpenAPI 3.0 |
| GitHub | Developer Tools | OpenAPI 3.0 |
| DigitalOcean | Developer Tools | OpenAPI 3.0 |
| Asana | Project Management | OpenAPI 3.0 |
| HubSpot | CRM | OpenAPI 3.0 |
| Front | Communication | OpenAPI 3.0 |
| Petstore | Example | OpenAPI 3.0 |

## Quality Gates

Every generated CLI must pass 7 gates before the press considers it done:

| Gate | What It Checks |
|------|---------------|
| `go mod tidy` | Clean, minimal dependencies |
| `go vet ./...` | Static analysis passes |
| `go build ./...` | All packages compile |
| Binary build | Produces a runnable binary |
| `--help` | Usage renders without crash |
| `version` | Prints version string |
| `doctor` | Health check runs |

If any gate fails, the press retries with fixes. If it still fails, it tells you exactly what broke and why.

## Contributing Catalog Entries

1. Generate a CLI from an API spec:
   ```bash
   ./printing-press generate --spec <your-openapi-spec>
   ```
2. Verify it compiles and passes all 7 gates
3. Create `catalog/<api-name>.yaml`:
   ```yaml
   name: my-api
   display_name: My API
   description: One-line description
   category: developer-tools
   spec_url: https://example.com/openapi.json
   spec_format: json
   openapi_version: "3.0"
   tier: community
   verified_date: "2026-03-24"
   homepage: https://example.com
   ```
4. Open a PR - CI validates the entry automatically (schema + URL accessibility + full generate + compile)

Or use the skill: `/printing-press submit <name>`

## Development

```bash
# Build
go build -o ./printing-press ./cmd/printing-press

# Run tests
go test ./...

# Generate from internal YAML spec
./printing-press generate --spec testdata/stytch.yaml

# Generate from OpenAPI spec
./printing-press generate --spec testdata/openapi/petstore.yaml

# Generate from URL
./printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json

# Initialize a pipeline
./printing-press print petstore
```

## License

MIT
