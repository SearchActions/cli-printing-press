---
title: "Dogfood the Press Until It Prints Steinberger-Quality CLIs"
type: fix
status: active
date: 2026-03-23
---

# Dogfood the Press Until It Prints Steinberger-Quality CLIs

## The Problem

We've written 7 plans today and shipped 3 phases of features. The press compiles CLIs from Petstore, Stytch, and Discord. But when you actually USE the generated CLIs, they're not good enough. We keep finding new bugs by dogfooding, writing a plan, fixing some, dogfooding again, finding more. The loop is right but the overhead of planning each round is wrong.

This is the last plan. After this: just fix and dogfood until it's done.

## What "Done" Looks Like

A generated CLI that Steinberger would put his name on. Concretely:

1. `stytch-cli users create --email test@example.com` actually calls the Stytch API and returns JSON
2. `stytch-cli doctor` shows all green (config OK, auth configured, API reachable)
3. `stytch-cli users --help` shows clean names like `create`, `delete`, `get` with readable descriptions
4. `petstore-cli pet add --name Fido --status available` actually creates a pet on the Petstore API
5. `discord-cli --help` shows organized command groups that make sense
6. No flag is rendered that a human can't actually use (no `--trusted_metadata string` for JSON objects)
7. No description is cut mid-word, CamelCase-garbled, or empty

## Known Remaining Issues (from last dogfood round)

1. **Petstore base URL is relative** (`/api/v3`) - doctor shows "unsupported protocol scheme". Need to either error clearly or prepend the spec's host.
2. **Stytch "Connectedapps" description** - humanize function doesn't split all-lowercase concatenated words. Needs the description from the spec, not a mangled operationId.
3. **Some Stytch resources have empty descriptions** - tag-to-resource mapping still incomplete.
4. **No usage examples in --help** - every Steinberger CLI shows examples.

## The Execution Loop

This plan has no numbered steps. It has one loop:

```
while (not steinberger quality):
  1. Build the press: go build -o ./printing-press ./cmd/printing-press
  2. Delete all output: rm -rf *-cli/
  3. Generate: ./printing-press generate --spec <spec> --output ./<name>-cli
  4. Build the CLI: cd <name>-cli && go build -o ./<name> ./cmd/<name>-cli
  5. USE IT:
     - Run --help on every command group
     - Run doctor
     - Try to actually call the API
     - Read every flag description
  6. Write down what's wrong
  7. Fix the PRESS (parser.go, templates, generator.go) - never the output
  8. Run go test ./... to catch regressions
  9. Go to 1
```

**Test specs in order:**
- Petstore first (smallest, canonical)
- Stytch second (real API, medium complexity)
- Discord last (hardest, most edge cases)

Don't move to the next spec until the previous one is genuinely good.

## Acceptance Criteria

All measured by regenerating from scratch:

- [ ] `petstore-cli doctor` shows API reachable (not "unsupported protocol scheme")
- [ ] `petstore-cli pet add --name Fido --status available` returns valid JSON
- [ ] `petstore-cli pet --help` has readable descriptions on every command
- [ ] `stytch-cli users create --help` shows only flags a human can actually use
- [ ] `stytch-cli users --help` has no empty or garbled descriptions
- [ ] `stytch-cli doctor` shows base_url=https://api.stytch.com and API reachable
- [ ] `discord-cli --help` compiles and shows organized command groups
- [ ] All 3 specs pass 7 quality gates
- [ ] No generated *-cli/ directories committed to repo

## Scope

Fix the press. Nothing else. No new features, no catalog work, no skill updates, no README changes. Just make the printer print good CLIs.

## How to Run This

Start a fresh Claude Code session (this one is huge). Run:

```
cd ~/cli-printing-press
/ce-work-beta codex docs/plans/2026-03-23-fix-press-dogfood-until-steinberger-quality-plan.md
```

Or just say: "dogfood the press against Petstore until it's good, then Stytch, then Discord."
