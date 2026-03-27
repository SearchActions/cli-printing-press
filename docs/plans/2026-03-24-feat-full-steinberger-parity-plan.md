---
title: "Full Steinberger Parity - Make the Press Print Finished CLIs, Not Scaffolds"
type: feat
status: active
date: 2026-03-24
---

# Full Steinberger Parity - Make the Press Print Finished CLIs, Not Scaffolds

## The Vision

Anyone downloads the printing press, points it at any API documentation, and gets a **finished CLI** - not a scaffold, not a starting point. A CLI that Peter Steinberger would put his name on. Ship it. Brew install it. Done.

The previous report drew the wrong line. It called sync/search/members "domain logic" and said the press shouldn't generate them. That's wrong. The press should generate everything a finished CLI needs. If discrawl has table output, pagination, retry logic, and color - then the press should print CLIs with table output, pagination, retry logic, and color.

## What's Missing (Ranked by Impact)

### Tier 1: Makes It Feel Like a Real CLI (Not a JSON Dumper)

**1. Table output for list responses**
- Currently: `fmt.Fprintln(cmd.OutOrStdout(), string(data))` - dumps raw JSON
- Steinberger: Formatted tables with column headers, alignment, truncation
- The `printTable()` helper already exists in `root.go.tmpl` but is never called
- Fix: Detect array responses, extract field names from first item, render as table
- Files: `command.go.tmpl`, `helpers.go.tmpl`

**2. Pagination**
- Currently: Single request, no page handling. Spec already has `Pagination` struct (cursor_field, has_more_field) but templates ignore it
- Steinberger: `--limit`, `--cursor`/`--after`, `--all` flags auto-generated for paginated endpoints
- Fix: When endpoint has pagination config, add flags and loop logic
- Files: `command.go.tmpl`, `client.go.tmpl`

**3. TTY detection + color**
- Currently: No color, no terminal detection
- Steinberger: ANSI color for tables, doctor, errors. Respects `NO_COLOR` env and `TERM=dumb`
- Fix: Add `isTerminal()` helper, use it for color/table decisions
- Files: `helpers.go.tmpl`, `root.go.tmpl`, `go.mod.tmpl` (add `mattn/go-isatty`)

**4. Formatted doctor output**
- Currently: `key=value` plain text or JSON blob
- Steinberger: Health dashboard with checkmarks, colors, alignment
- Fix: Rewrite doctor.go.tmpl to use formatted output with symbols
- Files: `doctor.go.tmpl`

### Tier 2: Makes It Production-Grade (Not a Toy)

**5. Retry logic with backoff**
- Currently: Single HTTP request, fail or succeed
- Steinberger: Exponential backoff on 5xx, rate limit detection (429 + Retry-After header)
- Fix: Add retry loop in client.go.tmpl `do()` method
- Files: `client.go.tmpl`

**6. Proper exit codes**
- Currently: 2=usage, 3=config, 4=auth, 5=api (generic)
- Steinberger: 3=not-found, 7=rate-limited, 8=retryable, 130=cancelled
- Fix: Parse HTTP status codes in client, map to exit codes. Handle SIGINT for 130
- Files: `helpers.go.tmpl`, `client.go.tmpl`

**7. Shell completions**
- Currently: Cobra's `completion` shows up but no custom completions
- Steinberger: bash/zsh/fish completions for resource names and flags
- Fix: Already works via Cobra's built-in - just ensure it's in the help
- Files: Already done (Cobra adds it automatically)

**8. Dry-run for mutations**
- Currently: POST/PUT/PATCH/DELETE always execute
- Steinberger: `--dry-run` shows what would be sent without sending
- Fix: Add `--dry-run` persistent flag, check in command before making request
- Files: `root.go.tmpl`, `command.go.tmpl`

### Tier 3: Makes It Delightful (Steinberger Polish)

**9. Smart defaults and config validation**
- Currently: Load TOML, no validation
- Fix: Validate base_url is HTTPS, warn on missing auth, suggest config on first run
- Files: `config.go.tmpl`

**10. "Next command" suggestions**
- Currently: Command finishes silently
- Steinberger: "Created user. Next: try `stytch-cli users get <id>`"
- Fix: Add contextual hints after mutations
- Files: `command.go.tmpl`

### Tier 4: Makes It Universal (Not Just OpenAPI)

**11. Accept any API documentation**
- Currently: OpenAPI 3.x and Swagger 2.0 only
- Vision: Markdown docs, curl examples, Postman collections, HTML pages
- Fix: Add parsers for additional formats. Use the internal spec as the universal intermediate representation
- Files: New parsers in `internal/`

**12. Natural language mode**
- Currently: `--spec file.yaml` only
- Vision: `printing-press "Stripe API"` - fetches docs, generates spec, builds CLI
- Fix: Agent integration that fetches and parses API docs
- Files: New command in `cmd/`

## Implementation Order

```
Phase 1: Table output + pagination  (makes list commands usable)
Phase 2: Color + TTY + doctor       (makes it feel polished)
Phase 3: Retries + exit codes       (makes it production-grade)
Phase 4: Dry-run + suggestions      (makes it delightful)
Phase 5: Universal input formats    (makes it for everyone)
```

## Acceptance Criteria (Full Parity)

### Table Output
- [ ] `discord-cli users list-my-guilds` shows a formatted table by default (not raw JSON)
- [ ] `discord-cli users list-my-guilds --json` still shows JSON
- [ ] Table auto-detects columns from response fields
- [ ] Long values are truncated with `...`

### Pagination
- [ ] Paginated endpoints have `--limit` and `--after`/`--cursor` flags
- [ ] `--all` fetches all pages and concatenates results
- [ ] Progress shown on stderr during `--all` fetch

### Color + TTY
- [ ] Table headers are bold when in a terminal
- [ ] Doctor checks show green checkmarks / red X marks
- [ ] `NO_COLOR=1` disables all color
- [ ] Piping to `jq` disables color automatically

### Retries
- [ ] 429 responses wait and retry (using Retry-After header)
- [ ] 5xx responses retry 3 times with exponential backoff
- [ ] Retry count shown on stderr

### Exit Codes
- [ ] 401/403 -> exit 4 (auth)
- [ ] 404 -> exit 3 (not found)
- [ ] 429 after retries exhausted -> exit 7 (rate limited)
- [ ] Ctrl-C -> exit 130 (cancelled)

### Dry-Run
- [ ] `discord-cli channels messages create 123 --content "test" --dry-run` shows the request without sending
- [ ] Shows method, URL, headers, body

### Universal Input (Future)
- [ ] `printing-press generate --spec petstore.yaml` (OpenAPI - current)
- [ ] `printing-press generate --spec api-docs.md` (Markdown)
- [ ] `printing-press generate --spec collection.json` (Postman)
- [ ] `printing-press "Stripe API"` (natural language - fetches docs)

## Scope

Only modify files in `internal/` and `internal/generator/templates/`. The press prints the CLI. The CLI is the product. The press is the machine.

## How to Run

Same dogfood loop as always:
```
while (not steinberger):
  1. Fix the press (internal/ files only)
  2. go test ./...
  3. rm -rf discord-cli/
  4. Generate + build + dogfood against Discord API
  5. Find what's still not good enough
  6. Go to 1
```
