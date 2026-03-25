# CLI Printing Press

Give it an API. It researches the competition, understands the docs, generates a Go CLI, polishes it with an LLM, scores it against Steinberger, and tells you if it's good enough to ship. One command.

```bash
printing-press generate --spec https://raw.githubusercontent.com/plaid/plaid-openapi/master/2020-09-14.yml --polish
```

```
Using LLM to understand API docs...
PASS go mod tidy
PASS go vet ./...
PASS go build ./...
PASS build runnable binary
PASS plaid-cli --help
PASS plaid-cli version
PASS plaid-cli doctor
Polish: 12 help texts improved, 8 examples added, README rewritten
Generated plaid-cli at ./plaid-cli
```

No official Plaid CLI exists. Now one does. 51 resources. Agent-native. Grade A on the Steinberger bar.

## How It Works

The press has three brains:

```
LLM BEFORE ($0.30)           TEMPLATE ($0)              LLM AFTER ($0.25)
  reads API docs          ->   Go templates render   ->   improves help text
  understands competitors       deterministic code          adds real examples
  writes complete spec          error handling              rewrites README
  plans what to build           agent-native flags          sells the tool
```

**Pass 1 - LLM understands the API.** Reads the docs or OpenAPI spec. Analyzes competing CLIs on GitHub (issues, README, PRs). Decides what commands to generate.

**Pass 2 - Templates generate correct code.** 14 Go templates produce a complete CLI project: auth, pagination, retries, color, doctor, --json, --select, --dry-run, --stdin, caching.

**Pass 3 - LLM polishes the output.** Rewrites help descriptions from spec jargon to developer-friendly. Generates realistic examples. Writes a README that sells the tool.

Without an LLM, the press still works - it falls back to regex-based doc parsing and template-only generation. The LLM makes it smart, not functional.

## Quick Start

```bash
git clone https://github.com/mvanhorn/cli-printing-press.git
cd cli-printing-press
go build -o ./printing-press ./cmd/printing-press

# From an OpenAPI spec (fast, free)
./printing-press generate --spec https://petstore3.swagger.io/api/v3/openapi.json

# From API docs when no spec exists (uses LLM)
./printing-press generate --docs "https://developers.notion.com/reference" --name notion

# With LLM polish pass (better help text, examples, README)
./printing-press generate --spec <url> --polish

# With lenient parsing for messy specs (PagerDuty, Intercom)
./printing-press generate --spec <url> --lenient
```

## What Gets Generated

Every CLI ships with:

| Feature | Flag | What It Does |
|---------|------|-------------|
| JSON output | `--json` | Machine-readable, pipeable to jq |
| Field filtering | `--select id,name` | Only the fields you want |
| Dry run | `--dry-run` | Shows the exact API request without sending |
| Stdin input | `--stdin` | Pipe JSON body: `echo '{}' \| mycli create --stdin` |
| Response cache | `--no-cache` | GET responses cached 5 min, bypass with flag |
| Skip confirmation | `--yes` | For agents and scripts on destructive actions |
| Plain output | `--plain` | Tab-separated for awk/cut |
| Color control | `--no-color` | Respects NO_COLOR, TERM=dumb |

Plus: doctor health check, shell completions, TOML config, OAuth2 browser flow, retry with backoff, rate limit detection, typed exit codes with actionable hints, idempotent creates/deletes.

## Agent-Native by Default

Every generated CLI is designed for AI agents:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - stdout is data, stderr is errors
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Retryable** - creates return "already exists" on retry, deletes return "already deleted"
- **Piped input** - `echo '{"key":"value"}' | mycli create --stdin`
- **Cacheable** - GET responses cached, bypass with `--no-cache`

Exit codes: `0` success, `2` usage, `3` not found, `4` auth, `5` API error, `7` rate limited, `10` config error.

## Scorecard

The press scores every generated CLI against the Steinberger bar (Peter Steinberger's gogcli at 6.5K stars is the 10/10 benchmark):

```bash
FULL_RUN=1 go test ./internal/pipeline/ -run TestFullRun -v -timeout 10m
```

```
Metric                   | petstore (EASY)   | plaid (MEDIUM)    | notion (HARD)
-------------------------|-------------------|-------------------|------------------
Quality Gates            | 7/7 PASS          | 7/7 PASS          | 7/7 PASS
Commands                 | 8                 | 51                | 6
Steinberger Total        | 67/80 (83%)       | 69/80 (86%)       | 67/80 (83%)
Grade                    | A                 | A                 | A
```

8 dimensions scored: output modes, auth, error handling, terminal UX, README, doctor, agent-native, local cache. The scorecard auto-generates fix plans for any dimension below 5/10.

## Intelligence Engine

The press doesn't just generate - it thinks:

- **Research phase** - discovers competing CLIs, analyzes their GitHub repos (issues, README, PRs), identifies what they're missing
- **Competitor intelligence** - reads competitor READMEs to extract command lists, feature requests, pain points
- **Doc-to-spec** - when no OpenAPI spec exists, reads API docs and generates a YAML spec (regex fallback if no LLM)
- **Dynamic planning** - each pipeline phase writes the next phase's plan using what it learned
- **Learning system** - remembers issues from past runs, auto-suggests fixes for future runs
- **Self-improvement** - scorecard gaps auto-generate fix plans for the press templates

## Catalog

18 APIs pre-registered with verified spec URLs:

| API | Category | Known Competitors |
|-----|----------|------------------|
| Plaid | Payments | plaid-cli (57 stars, abandoned) |
| Pipedrive | CRM | None |
| Stripe | Payments | stripe-cli (official) |
| GitHub | Developer Tools | gh (official) |
| Telegram | Communication | telegram-bot-api |
| Square | Payments | square-cli (3 stars) |
| Sentry | Developer Tools | sentry-cli (official) |
| LaunchDarkly | Developer Tools | ldcli (official) |
| + 10 more | Various | See catalog/ |

## Project Structure

```
cmd/printing-press/     CLI entry point
internal/
  catalog/              Catalog schema validator
  cli/                  CLI commands (generate, print, version)
  docspec/              Doc-to-spec generator (regex + LLM)
  generator/            Template engine + quality gates (14 templates)
  llm/                  Shared LLM runner (claude/codex CLI)
  llmpolish/            LLM polish pass (help, examples, README)
  openapi/              OpenAPI 3.0 parser (strict + lenient modes)
  pipeline/             Intelligence engine (research, scorecard, dogfood,
                        comparative, learnings, planner, self-improve)
  spec/                 Internal YAML spec parser
catalog/                18 API catalog entries
skills/                 Claude Code skill definitions
```

## Development

```bash
go build -o ./printing-press ./cmd/printing-press
go test ./...                    # 107 tests across 9 packages
FULL_RUN=1 go test ./internal/pipeline/ -run TestFullRun -v -timeout 10m  # full scorecard
```

## License

MIT
