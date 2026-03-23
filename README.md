# CLI Printing Press

Describe your API. Get a production CLI.

CLI Printing Press generates production-ready Go CLI tools from OpenAPI specs or API descriptions. Write a single YAML spec (or point to an existing OpenAPI spec), and get a complete Cobra-based CLI with auth management, output formatting, config loading, and build tooling.

## Quick Start

```bash
# Clone the repo
git clone https://github.com/mvanhorn/cli-printing-press.git
cd cli-printing-press

# Build
go build -o ./printing-press ./cmd/printing-press

# Generate a CLI from an OpenAPI spec
./printing-press generate --spec testdata/openapi/petstore.yaml --output ./petstore-cli

# Try it
cd petstore-cli && go install ./cmd/petstore-cli
petstore-cli --help
```

## Claude Code Plugin

```
/plugin marketplace add mvanhorn/cli-printing-press
```

Then use:
- `/printing-press Stytch auth API` - generate from natural language
- `/printing-press --spec ./openapi.yaml` - generate from spec file
- `/printing-press-catalog` - browse pre-built CLIs
- `/printing-press-catalog install stripe` - install from catalog

## Catalog

12 pre-built CLI definitions for popular APIs:

| API | Category | Spec |
|-----|----------|------|
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

## What Gets Generated

Every generated CLI includes:

- Cobra subcommand hierarchy (`<name>-cli <resource> <action>`)
- `--json`, `--plain`, `--quiet` output modes
- Auth management (API key, bearer token, basic auth)
- Config file at `~/.config/<name>-cli/config.toml`
- `doctor` command for health checks
- `version` command
- Makefile, .goreleaser.yaml, .golangci.yml, README.md
- Compiles and passes 7 quality gates automatically

## Supported Input Formats

- **OpenAPI 3.0+** (YAML or JSON)
- **Swagger 2.0** (YAML or JSON)
- **Internal YAML format** (simpler, hand-written)

Auto-detection: the generator reads the first 500 bytes and routes to the right parser.

## Contributing Catalog Entries

1. Generate a CLI: `./printing-press generate --spec <your-openapi-spec>`
2. Test it works
3. Create `catalog/<api-name>.yaml` with the spec URL and metadata
4. Open a PR - CI validates the entry automatically

Or use the skill: `/printing-press submit <name>`

## Development

```bash
# Run tests
go test ./...

# Build
go build -o ./printing-press ./cmd/printing-press

# Generate from internal spec
./printing-press generate --spec testdata/stytch.yaml

# Generate from OpenAPI
./printing-press generate --spec testdata/openapi/petstore.yaml
```

## License

MIT
