# CLI Printing Press - Development Conventions

## Build & Test

```bash
go build -o ./printing-press ./cmd/printing-press
go test ./...
```

## Project Structure

- `cmd/printing-press/` - CLI entry point
- `internal/spec/` - Internal YAML spec parser
- `internal/openapi/` - OpenAPI 3.0+ parser
- `internal/generator/` - Template engine + quality gates
- `internal/catalog/` - Catalog schema validator
- `catalog/` - API catalog entries (YAML)
- `skills/` - Claude Code skill definitions
- `testdata/` - Test fixtures (internal + OpenAPI specs)

## Commit Style

Conventional commits: `feat(scope):`, `fix(scope):`, `docs:`, `refactor:`

## Adding Catalog Entries

Catalog entries in `catalog/` must pass `internal/catalog` validation:
- Required fields: name, display_name, description, category, spec_url, spec_format, tier
- spec_url must use HTTPS
- category must be: auth, payments, email, developer-tools, project-management, communication, crm, example
- tier must be: official or community

## Quality Gates

Generated CLIs must pass 7 gates: go mod tidy, go vet, go build, binary build, --help, version, doctor.
