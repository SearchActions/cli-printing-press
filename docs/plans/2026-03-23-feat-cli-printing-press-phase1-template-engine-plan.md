---
title: "CLI Printing Press Phase 1: Go Template Engine"
type: feat
status: active
date: 2026-03-23
origin: docs/plans/2026-03-23-feat-cli-printing-press-plan.md
---

# CLI Printing Press Phase 1: Go Template Engine

## Goal

Build the core template engine that generates production Go + Cobra CLI projects from a structured API description. Not from OpenAPI yet - from a simpler internal YAML format that we control. OpenAPI parsing comes in Phase 2.

By the end of Phase 1: `printing-press generate --spec stytch.yaml` produces a complete, compilable, runnable Go CLI tool with Steinberger-quality patterns.

## What Gets Built

A Go CLI tool called `printing-press` that:
1. Reads a YAML API description file
2. Generates a complete Go + Cobra project directory
3. The generated project compiles with `go build`
4. The generated project runs: `<tool> --help`, `<tool> <command> --json`, `<tool> doctor`
5. The generated project follows every Steinberger pattern extracted from wacli/discrawl/sag

## The Internal API Description Format

This is the input format. Simpler than OpenAPI, designed to be easy for agents to generate.

```yaml
# stytch.yaml - API definition for CLI Printing Press
name: stytch
description: "Stytch authentication API CLI"
version: "0.1.0"
base_url: "https://api.stytch.com/v1"

auth:
  type: api_key  # api_key | oauth2 | bearer_token | none
  header: "Authorization"
  format: "Basic {project_id}:{secret}"
  env_vars:
    - STYTCH_PROJECT_ID
    - STYTCH_SECRET

config:
  format: toml  # toml | yaml
  path: "~/.config/stytch-cli/config.toml"

resources:
  users:
    description: "Manage Stytch users"
    endpoints:
      list:
        method: GET
        path: "/users"
        description: "List all users"
        params:
          - name: limit
            type: int
            default: 100
            description: "Max users to return"
          - name: cursor
            type: string
            description: "Pagination cursor"
        response:
          type: array
          item: User
        pagination:
          type: cursor
          cursor_field: "cursor"
          has_more_field: "results.has_more"

      get:
        method: GET
        path: "/users/{user_id}"
        description: "Get a user by ID"
        params:
          - name: user_id
            type: string
            required: true
            positional: true
            description: "User ID"
        response:
          type: object
          item: User

      create:
        method: POST
        path: "/users"
        description: "Create a new user"
        body:
          - name: email
            type: string
            description: "User email"
          - name: phone_number
            type: string
            description: "User phone"
          - name: name
            type: object
            fields:
              - name: first_name
                type: string
              - name: last_name
                type: string
        response:
          type: object
          item: User

      delete:
        method: DELETE
        path: "/users/{user_id}"
        description: "Delete a user"
        params:
          - name: user_id
            type: string
            required: true
            positional: true

  sessions:
    description: "Manage user sessions"
    endpoints:
      list:
        method: GET
        path: "/sessions"
        params:
          - name: user_id
            type: string
            required: true
        response:
          type: array
          item: Session

      revoke:
        method: POST
        path: "/sessions/revoke"
        body:
          - name: session_id
            type: string
            required: true

types:
  User:
    fields:
      - name: user_id
        type: string
      - name: email
        type: string
      - name: phone_number
        type: string
      - name: status
        type: string
      - name: created_at
        type: string
  Session:
    fields:
      - name: session_id
        type: string
      - name: user_id
        type: string
      - name: started_at
        type: string
      - name: expires_at
        type: string
```

## What Gets Generated

From the YAML above, `printing-press generate --spec stytch.yaml` produces:

```
stytch-cli/
  cmd/stytch-cli/main.go              # Thin main (~15 lines, discrawl pattern)
  internal/cli/
    root.go                            # Cobra root, persistent flags (--json, --timeout, --config)
    output.go                          # Tri-mode: human/json/plain via type-switch
    helpers.go                         # parseTime, truncate, csvList, typed exit codes
    version.go                         # var version = "0.1.0"
    doctor.go                          # Health check: config, auth, API connectivity
    users.go                           # users list, users get, users create, users delete
    sessions.go                        # sessions list, sessions revoke
  internal/config/
    config.go                          # TOML config, Load(), Write(), auth resolution
  internal/client/
    client.go                          # HTTP client: base URL, auth, timeout, retry
    pagination.go                      # Cursor-based pagination helper
  internal/types/
    types.go                           # User, Session structs with JSON tags
  go.mod                               # cobra, go-toml, testify
  go.sum
  .goreleaser.yaml                     # 6 targets, ldflags, checksums
  .golangci.yml                        # Standard linter config
  Makefile                             # build, test, lint, install
  README.md                            # Generated from API description
  SPEC.md                              # Agent build contract (discrawl pattern)
```

### Key Generated Patterns (from Steinberger source code analysis)

**main.go** (discrawl pattern):
```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/USER/stytch-cli/internal/cli"
)

func main() {
    if err := cli.Execute(context.Background(), os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(cli.ExitCode(err))
    }
}
```

**root.go** (wacli pattern):
```go
func Execute(ctx context.Context, args []string) error {
    var flags rootFlags
    rootCmd := &cobra.Command{
        Use:           "stytch-cli",
        Short:         "Stytch authentication API CLI",
        SilenceUsage:  true,
        SilenceErrors: true,
        Version:       version,
    }
    rootCmd.PersistentFlags().BoolVar(&flags.asJSON, "json", false, "Output as JSON")
    rootCmd.PersistentFlags().BoolVar(&flags.plain, "plain", false, "Output as plain text")
    rootCmd.PersistentFlags().BoolVar(&flags.quiet, "quiet", false, "Bare output")
    rootCmd.PersistentFlags().StringVar(&flags.configPath, "config", "", "Config file path")
    rootCmd.PersistentFlags().DurationVar(&flags.timeout, "timeout", 30*time.Second, "Request timeout")
    // ... add subcommands ...
    rootCmd.SetArgs(args)
    return rootCmd.Execute()
}
```

**output.go** (discrawl tri-mode pattern):
```go
func (r *rootFlags) print(cmd *cobra.Command, value any) error {
    if r.asJSON {
        enc := json.NewEncoder(cmd.OutOrStdout())
        enc.SetIndent("", "  ")
        return enc.Encode(value)
    }
    if r.plain {
        return printPlain(cmd.OutOrStdout(), value)
    }
    return printHuman(cmd.OutOrStdout(), value)
}
```

**Typed exit codes** (discrawl pattern):
```go
type cliError struct {
    code int
    err  error
}

func usageErr(err error) error  { return &cliError{code: 2, err: err} }
func configErr(err error) error { return &cliError{code: 3, err: err} }
func authErr(err error) error   { return &cliError{code: 4, err: err} }
func apiErr(err error) error    { return &cliError{code: 5, err: err} }
```

**doctor.go** (wacli/discrawl pattern):
```go
func newDoctorCmd(flags *rootFlags) *cobra.Command {
    return &cobra.Command{
        Use:   "doctor",
        Short: "Check CLI health",
        RunE: func(cmd *cobra.Command, args []string) error {
            report := map[string]any{}
            // Check config
            // Check auth credentials
            // Check API connectivity (GET base_url with timeout)
            return flags.print(cmd, report)
        },
    }
}
```

## Implementation Steps

### Step 1: Project scaffold

Create the `cli-printing-press` Go project:
```
cli-printing-press/
  cmd/printing-press/main.go
  internal/
    spec/spec.go              # Parse the YAML spec format
    generator/generator.go    # Orchestrate generation
    templates/                # Go template files
      main.go.tmpl
      root.go.tmpl
      output.go.tmpl
      helpers.go.tmpl
      version.go.tmpl
      doctor.go.tmpl
      command.go.tmpl         # Per-resource command file
      config.go.tmpl
      client.go.tmpl
      pagination.go.tmpl
      types.go.tmpl
      go.mod.tmpl
      goreleaser.yaml.tmpl
      golangci.yml.tmpl
      makefile.tmpl
      readme.md.tmpl
      spec.md.tmpl
  testdata/
    stytch.yaml               # Test spec (from above)
    clerk.yaml                # Second test spec
    loops.yaml                # Third test spec (simplest)
  go.mod
  go.sum
```

### Step 2: Spec parser

Parse the YAML spec into Go structs. Validate: required fields present, types valid, no duplicate resource names, auth type recognized.

### Step 3: Templates

Write Go templates for each generated file. Use `text/template` with custom functions:
- `toLower`, `toTitle`, `toCamel`, `toSnake`, `toKebab`
- `pluralize`, `singularize`
- `cobraFlagType` (maps spec types to Cobra flag types)
- `goType` (maps spec types to Go types)
- `jsonTag` (generates `json:"name"` tags)

### Step 4: Generator

Wire spec parser -> templates -> file writer. For each resource in the spec, generate a command file. Wire all commands into root.go.

### Step 5: Validation

After generation, run the quality gates:
1. `go mod tidy` on generated project
2. `go vet ./...`
3. `go build ./...`
4. Run `<tool> --help` and verify output
5. Run `<tool> --version` and verify
6. Run `<tool> doctor` and verify it reports config/auth status

### Step 6: Test with 3 real APIs

Write spec files for Stytch, Clerk, and Loops. Generate all 3. Verify all compile and run.

## Acceptance Criteria

- [ ] `printing-press generate --spec stytch.yaml` produces a directory
- [ ] Generated project compiles with `go build ./...`
- [ ] Generated project has `--help` on every command
- [ ] Generated project has `--json` on every command
- [ ] Generated project has `--version`
- [ ] Generated project has `doctor` command
- [ ] Generated project has tri-mode output (human/json/plain)
- [ ] Generated project has typed exit codes (0, 2, 3, 4, 5)
- [ ] Generated project has TOML config support
- [ ] Generated project has auth resolution (env var -> config -> flag)
- [ ] Generated project has `.goreleaser.yaml` for 6 targets
- [ ] Generated project has `.golangci.yml`
- [ ] Generated project has `Makefile`
- [ ] Generated project has `README.md` from API description
- [ ] Generated project has `SPEC.md` agent build contract
- [ ] 3 test specs (Stytch, Clerk, Loops) all generate and compile
- [ ] `go vet` passes on all generated projects
- [ ] Quality gate tracker verifies all gates pass

## What Phase 1 Does NOT Include

- OpenAPI spec parsing (Phase 2)
- Natural language API description (Phase 3)
- Claude Code skill / plugin packaging (Phase 4)
- Community catalog / PR submission flow (Phase 4)
- Security validation for catalog entries (Phase 4)
- Homebrew formula (Phase 4)
- OAuth2 auth template (Phase 2 - only api_key and bearer_token in Phase 1)

## Technical Decisions

- **Language**: Go (we're generating Go CLIs, the generator should also be Go)
- **Template engine**: `text/template` (stdlib, no dependencies)
- **Spec format**: YAML via `gopkg.in/yaml.v3`
- **CLI framework for printing-press itself**: Cobra (dogfooding our own patterns)
- **Testing**: testify + golden file tests (compare generated output against expected)

## Learnings to Capture for Phase 2

After Phase 1 ships, document:
1. Which Steinberger patterns were hardest to template?
2. Which parts of the YAML spec format were awkward? What should change?
3. How well do the generated CLIs actually work against real APIs?
4. What's the delta between generated and hand-written for the 3 test APIs?
5. What did the quality gates catch?

These feed directly into the Phase 2 plan (OpenAPI parser).

## Origin

Phase 1 of: docs/plans/2026-03-23-feat-cli-printing-press-plan.md
