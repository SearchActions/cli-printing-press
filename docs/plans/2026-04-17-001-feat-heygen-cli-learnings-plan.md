---
title: "feat(cli): Apply HeyGen CLI learnings to the printing press"
type: feat
status: active
date: 2026-04-17
---

# feat(cli): Apply HeyGen CLI learnings to the printing press

## Overview

Joshua Xu's "HeyGen CLI: The Visualization Layer for AI Agents" (2026-04-13) describes an agent-first CLI built around four distinctive ideas: a single command that chases a long-running async job to completion, a persistent agent identity reused across invocations, built-in delivery after the job finishes, and an in-band channel for agents themselves to send feedback back to the tool.

Mapping those ideas onto the printing press (PP) after scanning current `main`:

| HeyGen learning | PP status |
|---|---|
| Agent-first CLI posture (--json, structured exit codes, doctor, agent-context) | ALREADY DONE (#186 SKILL.md, Cloudflare-plan agent-context subcommand, prior agent-first PostHog work) |
| Single-command pipeline that drives an async job to completion (`script → avatar → video → delivery`) | NOT YET - `tail` polls for live streams, but no `--wait` on the submitting command and no `jobs` surface. Closed by Unit 1. |
| Persistent agent identity reused across runs (HeyGen's "Beacon" - same avatar, same voice, different context each day) | NOT YET - no named-profile / named-context system in printed CLIs. Closed by Unit 2. |
| First-class delivery so the agent's output lands where humans will see it (`fewer steps between agent output and a finished video`) | PARTIAL - `export` writes files, `sync` populates SQLite; no `--deliver` routing. Closed by Unit 3. |
| In-band agent feedback channel (`when your agent has feedback, send it our way`) | NOT YET - no `feedback` subcommand, no local capture, no opt-in upstream. Closed by Unit 4. |
| Visualization layer abstraction (text → talking head, motion graphic, live stream) | NOT APPLICABLE - that is HeyGen's domain thesis, not a generator pattern. PP already lets any API produce structured output; the visualization is the consumer's concern. |

The four gaps are additive, independently shippable, and each compounds across every future printed CLI.

## Problem Frame

Printed CLIs are agent-native at the invocation level: a single command returns structured output. They are not agent-native at the *workflow* level:

1. When a printed CLI wraps an API endpoint that returns `202 + job_id`, the current generator emits a command that returns the job ID and stops. An agent that wants the finished artifact must write its own polling loop, which it usually does badly (fixed sleeps, no jitter, no timeout discipline). This is exactly the shape HeyGen's CLI eliminated with one command.
2. Every invocation starts from zero. An agent that briefs its user every morning with Linear + Notion data configures the same flags, output paths, and filters on every run. HeyGen captured this as Beacon: a named, persistent context. Printed CLIs have no equivalent.
3. Output lands on stdout. Agents scheduled in cron, CI, or background loops need it to land somewhere a human will actually see (Slack, a webhook, a file drop). Today the generator leaves this entirely to the agent.
4. Agents notice when a generated CLI is wrong or confusing - wrong flag name, missing endpoint, unhelpful error - but the friction of turning that observation into a GitHub issue means the signal is lost. HeyGen's framing ("send it our way - from you or your agent") makes feedback a programmable action.

## Requirements Trace

- R1. Generated CLIs detect async-job endpoints from the spec and expose a single-command `--wait` path plus a generic `jobs` surface for listing, inspecting, and cancelling in-flight work.
- R2. Generated CLIs support named profiles: save a set of flags and env under a label, reuse them with `--profile <name>`, list and describe them.
- R3. Generated CLIs accept `--deliver <target>` that routes structured output to one of a fixed set of sinks (webhook, file, stdout) without the agent having to pipe.
- R4. Generated CLIs ship a `feedback` subcommand that captures agent-authored feedback locally and, when the user has opted in, POSTs it to a configured endpoint.
- R5. Scorecard recognises the four new capabilities so a correctly-built CLI is not penalised for having them and a CLI that skips them is not rewarded.
- R6. None of the above break existing commands, existing gates, or already-published CLIs in the public library.

## Scope Boundaries

- Does NOT change the `tail` command (live polling of an API) - `tail` and the new `--wait` / `jobs` surface are different shapes and coexist.
- Does NOT add a telemetry backend. Upstream feedback POSTs are opt-in and point at a user-configured URL; no PP-hosted service is introduced.
- Does NOT add Slack, email, or SaaS delivery targets. `--deliver` starts with `webhook:`, `file:`, and `stdout:` only. Later retros may add more.
- Does NOT touch already-published CLIs in `mvanhorn/printing-press-library`. New generations pick these up; old CLIs inherit them on next regeneration.
- Does NOT change the MCP surface in this plan. If these features need MCP analogues, a follow-up plan covers that.
- Does NOT invent a visual / video / avatar pipeline inside PP. That is HeyGen's domain, not a generator pattern.

## Context & Research

### Relevant Code and Patterns

- `internal/generator/templates/tail.go.tmpl` - polling shape (timer, signal handling, NDJSON) to mirror in the new async-job polling helper. Tail targets resource-listing endpoints; the async-job helper targets single-job status endpoints - the templates are cousins, not duplicates.
- `internal/generator/templates/sync.go.tmpl` - existing long-running client with retry, timeout, and store integration - the natural spot for the generic `jobs` local store.
- `internal/generator/templates/channel_workflow.go.tmpl` - current `workflow` parent command (`workflow archive`, `workflow status`). The new `jobs` and `profile` commands follow the same top-level-parent registration pattern.
- `internal/generator/templates/command_endpoint.go.tmpl`, `command_promoted.go.tmpl` - where per-endpoint commands are emitted. The `--wait` flag lives here, gated on detection done at generation time.
- `internal/generator/templates/config.go.tmpl` - config file loader. Profiles reuse the config mechanism rather than inventing a parallel store.
- `internal/generator/templates/export.go.tmpl` - existing output-to-file primitive; `--deliver file:<path>` reuses this machinery.
- `internal/generator/templates/doctor.go.tmpl`, `internal/generator/templates/agent_context.go.tmpl` - canonical shape for built-in utility subcommands (`doctor`, `agent-context`). `feedback` mirrors this shape.
- `internal/generator/generator.go` - central generator dispatch; new detection pass for async-job endpoints attaches here.
- `internal/pipeline/scorecard.go` - dimension scoring with `UnscoredDimensions` for "not applicable" cases. Extended in Unit 5 for the new capabilities.
- `internal/pipeline/dogfood.go` - structural checks; new `checkAsyncJobCoverage` check verifies every detected async endpoint has `--wait` wired up.

### Institutional Learnings

- `docs/solutions/best-practices/steinberger-scorecard-scoring-architecture-2026-03-27.md` - banding pattern to reuse for the new scorecard dimension in Unit 5.
- `docs/plans/2026-04-10-001-feat-agent-first-posthog-learnings-plan.md` (completed) - front-loading context, one-liner opinions in tool descriptions, keep surfaces additive. The `feedback` subcommand description should carry a PostHog-style opinion ("write what surprised you, not a bug report").
- `docs/plans/2026-04-13-002-feat-cloudflare-cli-learnings-plan.md` (active) - the agent-context subcommand lands a self-describing runtime surface. Profiles (Unit 2) should appear in that JSON payload so introspecting agents see available personas.
- AGENTS.md Machine vs Printed CLI rule - all four changes are machine-layer; every future CLI benefits.
- AGENTS.md glossary - profiles are a new term; add an entry so agents using the skill know what `--profile` means.

### External References

- Joshua Xu, "HeyGen CLI: The Visualization Layer for AI Agents" (2026-04-13) - origin of these learnings.
- HeyGen developer docs at developers.heygen.com - canonical example of a 202-then-poll API shape that a PP-generated CLI for HeyGen would consume once these features exist.
- Karpathy thread (2026-04-10) quoted in the article - "tractable brain upload" framing for identity persistence; useful context for Unit 2's product framing but no implementation content.

## Key Technical Decisions

- **Async-job detection is spec-driven, not runtime probed.** At generation time the generator inspects each operation for one of three signals: (a) response type listed as `202`, (b) response schema with a field whose name matches `job_id|task_id|operation_id|request_id`, (c) sibling `GET /<resource>/{id}` or `GET /<resource>_status/{id}` operation. When at least two signals match, the operation is marked async and gets `--wait`. Runtime detection is fragile and would slow every invocation.
- **`jobs` is a generic parent command backed by the existing local store.** Launching an async-job command stamps a row into the `jobs` table in the CLI's SQLite store. `jobs list/get/cancel/prune` reads that table. No new data layer; reuses `internal/store` from `sync.go.tmpl`.
- **`--wait` uses exponential backoff with jitter and a configurable ceiling.** Fixed-interval polling is what agents do when they write it themselves and it is what has to be replaced. Starting at 2 seconds, capping at 30 seconds, 10 minute default timeout. `--wait-timeout` and `--wait-interval` override.
- **Profiles live in the same config file as existing config, under a `profiles:` key.** Reuses the config loader. No new file format, no new search path. `--profile <name>` overlays that profile's values on top of defaults; explicit flags still win.
- **`--deliver` starts with three sinks: `webhook:<url>`, `file:<path>`, `stdout` (default).** The value is a single string that parses to `sink:target`. Adding Slack or email later is a matter of registering another sink, not reworking the interface.
- **`feedback` writes locally first, upstream second.** `feedback <text>` always appends a JSON line to `~/.<cli>/feedback.jsonl`. If the user has set `feedback_endpoint` in config and passed `--send` (or set `feedback_auto_send: true`), the CLI POSTs the entry. Defaulting to local-only keeps the feature safe by default.
- **Every change is additive.** No existing behavior changes, no existing tests break, no existing CLI shape is renamed.

## Open Questions

### Resolved During Planning

- **Does `--wait` replace `tail`?** No. `tail` is for streaming resource changes (log-tail style). `--wait` is for a single known job ID. They coexist.
- **Should the async detector be permissive or strict?** Strict. False positives add `--wait` to endpoints that do not have a job status and fail verify. Two-of-three signal threshold errs on the side of skipping rather than inventing.
- **Should profiles be env-based or file-based?** File-based, in the existing config file, with env override. Env-only would lose discoverability; a separate file would duplicate the config loader.
- **Should feedback POST by default when an endpoint is configured?** No. Local-only by default; upstream POST requires `--send` or explicit opt-in via config. Silent POSTs are exactly the kind of surprise AGENTS.md warns against.
- **Do we need Slack/email `--deliver` targets day one?** No. webhook + file covers the agent use cases in the article. Slack/email can land in a later retro without changing the interface.
- **Does `jobs` need to survive across machines?** No. Jobs are local to the machine that started them. A follow-up plan can add a shared remote store if agents ask for it.

### Deferred to Implementation

- **Exact regex list for async field names.** Start with `job_id|task_id|operation_id|request_id|async_id`; grow via retro findings.
- **Default `--wait-timeout` per API family.** 10 minutes is the cross-API default. Per-spec override may be useful once real APIs prove it out, but only after seeing usage.
- **`jobs prune` retention window.** Likely 7 days but tuned by measuring typical agent usage.
- **Schema for the upstream feedback POST body.** Draft in Unit 4; finalize with one real receiver (e.g. a Formbricks or Webhook.site URL) before shipping.

## High-Level Technical Design

> *This illustrates the intended approach and is directional guidance for review, not implementation specification. The implementing agent should treat it as context, not code to reproduce.*

End-to-end shape for the async + profile + deliver + feedback story, from an agent's point of view:

```
# First run: create a named profile that captures the daily context
mycli profile save briefing \
  --topic "AI infra funding"            \
  --format json                          \
  --deliver webhook:https://hooks.slack.com/xxx

# Every morning: one command, blocks until job is done, routes output
mycli generate-report --profile briefing --wait

# Internally this is:
#   1. profile 'briefing' loads topic/format/deliver
#   2. POST /reports -> 202 { job_id: "r_123" }
#   3. exponential-backoff poll GET /reports/r_123 until status == done
#   4. fetch result artifact
#   5. route artifact to webhook:<url>
#   6. record the job in local SQLite (mycli jobs list shows it)
#
# Inspect in-flight or past runs:
mycli jobs list
mycli jobs get r_123

# When the agent notices the CLI is wrong about something:
mycli feedback "the --since flag is inclusive but docs say exclusive"
```

Detection flow at generation time:

```
spec operation ──► async-detector ──► 2+ signals? ──► mark async
    │                                     │
    │                                     └── schema has job_id-shaped field
    │                                     └── 202 response listed
    │                                     └── sibling /<res>_status/{id} exists
    │
    ├─► yes ──► emit command with --wait, --wait-timeout, --wait-interval
    │           register in jobs table on launch
    │           poll sibling status endpoint
    │
    └─► no  ──► emit command unchanged
```

## Implementation Units

- [ ] **Unit 1: Async-job detection, `--wait`, and `jobs` parent command**

**Goal:** Generated CLIs treat async endpoints as first-class. Detected async commands accept `--wait` and drive the job to completion (or timeout); a new `jobs` parent command lists, inspects, cancels, and prunes tracked jobs using the existing local store. This is the core HeyGen pattern - one command from submit to finished artifact.

**Requirements:** R1, R6

**Dependencies:** None

**Files:**
- Modify: `internal/generator/generator.go` (add async-detection pass over parsed operations; annotate spec nodes with `IsAsyncJob`, `StatusOperationRef`, `JobIDField`)
- Create: `internal/generator/async_detect.go` (pure detection logic with unit tests; two-of-three signal rule)
- Create: `internal/generator/async_detect_test.go`
- Modify: `internal/generator/templates/command_endpoint.go.tmpl` (conditional `--wait`, `--wait-timeout`, `--wait-interval` flags when endpoint is marked async)
- Create: `internal/generator/templates/jobs.go.tmpl` (new `jobs` parent command: `list`, `get`, `cancel`, `prune`)
- Create: `internal/generator/templates/jobs_store.go.tmpl` (SQLite schema for `jobs` table; reuses `store.Open` from sync)
- Modify: `internal/generator/templates/root.go.tmpl` (register `jobs` parent command)
- Modify: `internal/pipeline/dogfood.go` (new `checkAsyncJobCoverage` that verifies every detected async endpoint has `--wait` wired up)
- Modify: `internal/generator/templates/skill.md.tmpl` (document the `--wait` and `jobs` surface for agents consuming the skill)
- Test: `internal/generator/async_detect_test.go`
- Test: `internal/generator/jobs_template_test.go`
- Test: `internal/pipeline/dogfood_test.go` (add `checkAsyncJobCoverage` cases)

**Approach:**
- Detection is purely spec-driven and runs once per generation, mutating the in-memory operation graph before template emission.
- `--wait` loops against the detected sibling status operation using exponential backoff with jitter; not a new HTTP client.
- Job rows live in the CLI's existing SQLite store, keyed by `(cli_name, job_id, submitted_at)`. `jobs` does not introduce a new data layer.
- `jobs cancel` only attempts a cancel HTTP call when the spec exposes a `DELETE /<resource>/{id}` or `POST /<resource>/{id}/cancel`. Otherwise it marks the local row as orphaned and exits non-zero with a structured message.

**Technical design:** *(directional, not implementation specification)*

```
// Pseudocode for the two-of-three async signal rule
signals := 0
if op.HasResponseStatus(202)                    { signals++ }
if op.ResponseSchemaHasFieldMatching(JobIDRegex) { signals++ }
if spec.HasSiblingStatusOperation(op)            { signals++ }
op.IsAsyncJob = signals >= 2
```

**Patterns to follow:**
- `internal/generator/templates/tail.go.tmpl` for polling loop + signal handling
- `internal/generator/templates/sync.go.tmpl` for SQLite store integration
- `internal/pipeline/dogfood.go` existing `checkDeadFunctions` (post-#183) for structural check pattern with fixed-point analysis

**Test scenarios:**
- Happy path: spec with `POST /videos -> 202 {video_id}` and `GET /videos/{id}` generates a `videos create` command with `--wait` that polls to completion and returns the final resource
- Happy path: `jobs list` shows submitted jobs with submit time, status, and age; `jobs get <id>` returns the latest status row as JSON
- Happy path: `jobs prune --older-than 7d` removes rows older than the threshold and reports a count
- Edge case: operation with only a `202` response but no sibling status endpoint is NOT marked async (one signal only)
- Edge case: operation with a `job_id`-shaped response field AND a 202 but no sibling status still triggers detection (two of three); dogfood surfaces a warning that the status operation is missing
- Edge case: `--wait-timeout 0` means no timeout; negative values are rejected with a structured error
- Error path: status endpoint returns 500 during polling - client retries with backoff up to 3 failures, then emits a structured timeout error and leaves the job row in `errored` state
- Error path: user presses Ctrl-C during `--wait` - loop exits cleanly, job row remains `in_flight`, exit code non-zero
- Error path: `jobs cancel` on a spec without cancel operation exits non-zero with a structured "cancel not supported by this API" message
- Integration: a printed CLI built against a recorded spec passes dogfood's new `checkAsyncJobCoverage` when every detected async endpoint has `--wait`, fails it when the generator regresses

**Verification:**
- A generated CLI with at least one async endpoint exposes `--wait`, `jobs list`, `jobs get`, `jobs cancel`, `jobs prune` and dogfood reports the async coverage check as PASS
- `go test ./internal/generator/...` and `go test ./internal/pipeline/...` both pass on main
- An end-to-end run against a recorded async spec (HeyGen or similar) completes one full submit-poll-download cycle without the agent writing polling code

---

- [ ] **Unit 2: Named profiles system**

**Goal:** Every printed CLI supports named profiles - a block of flags and env values saved under a label and reused with `--profile <name>`. Implements HeyGen's "Beacon" pattern: same configuration, invoked repeatedly, different input each time. Compounds with Unit 1 because long-running async jobs are exactly what benefits from a saved context.

**Requirements:** R2, R6

**Dependencies:** None; composes with Unit 1

**Files:**
- Modify: `internal/generator/templates/config.go.tmpl` (load and merge a `profiles:` section from the existing config file; profile values overlay defaults, explicit flags win)
- Create: `internal/generator/templates/profile.go.tmpl` (new `profile` parent command: `save`, `use`, `list`, `show`, `delete`)
- Modify: `internal/generator/templates/root.go.tmpl` (register `profile` parent; wire `--profile` persistent flag)
- Modify: `internal/generator/templates/agent_context.go.tmpl` (include `available_profiles` in the agent-context JSON so runtime introspecting agents discover profiles)
- Modify: `internal/generator/templates/skill.md.tmpl` (short agent-facing explanation of profiles)
- Modify: `AGENTS.md` glossary (new `profile` entry - canonical term in the generated-CLI vocabulary)
- Test: `internal/generator/profile_template_test.go`
- Test: `internal/generator/templates/config_test.go` (new cases for profile merge precedence)

**Approach:**
- Profiles reuse the existing config loader; no new file, no new search path.
- `--profile <name>` is a persistent root flag implemented in `root.go.tmpl`.
- Precedence: explicit CLI flag > env var > profile > config default.
- `profile save <name>` captures the current invocation's non-default flags; `profile show <name>` emits JSON; `profile delete <name>` requires `--yes` to skip confirmation in agent use.
- The `agent-context` JSON grows an `available_profiles` array listing `{name, description}` so introspecting agents learn which personas exist.

**Patterns to follow:**
- `internal/generator/templates/config.go.tmpl` for YAML/JSON config merging
- `internal/generator/templates/doctor.go.tmpl` for a clean utility subcommand shape

**Test scenarios:**
- Happy path: `profile save briefing --topic foo --format json` persists a profile; later `--profile briefing` applies those values; `profile show briefing` emits them as JSON
- Happy path: `profile list` prints `{name, description, field_count}` as NDJSON when `--json` is set
- Edge case: profile does not exist - `--profile missing` exits non-zero with a structured error naming the missing profile and listing available ones
- Edge case: explicit flag overrides profile value ("--topic bar" beats "--profile briefing" which set topic to foo)
- Edge case: env var overrides profile but is overridden by explicit flag
- Error path: `profile save` called with no non-default flags - exits non-zero with a clear message so agents do not create empty profiles
- Integration: after `profile save briefing`, agent-context JSON contains `available_profiles: [{name: "briefing", ...}]`
- Integration: with Unit 1 live, `--profile briefing --wait` correctly composes profile-supplied flags with async behavior

**Verification:**
- A freshly generated CLI supports the five profile subcommands
- Config precedence holds in all four combinations (flag > env > profile > default)
- `agent-context` JSON surfaces `available_profiles`

---

- [ ] **Unit 3: `--deliver` output routing**

**Goal:** Every command that emits structured output (so: every command) accepts `--deliver <sink>` which routes the output without the agent piping manually. Start with three sinks: `stdout` (default, unchanged), `file:<path>`, `webhook:<url>`. Closes the "fewer steps between agent output and delivered artifact" gap from the article.

**Requirements:** R3, R6

**Dependencies:** None; composes with Unit 1 (async `--wait` output is what most wants delivery) and Unit 2 (profiles persist a preferred sink)

**Files:**
- Create: `internal/generator/templates/deliver.go.tmpl` (sink registry, parse `sink:target` strings, dispatch)
- Modify: `internal/generator/templates/root.go.tmpl` (persistent `--deliver` flag; bind into output path)
- Modify: `internal/generator/templates/command_endpoint.go.tmpl` (route command output through deliver layer instead of directly to stdout)
- Modify: `internal/generator/templates/config.go.tmpl` (allow `deliver:` default in config and per-profile)
- Modify: `internal/generator/templates/skill.md.tmpl` (document the three sinks for agents)
- Test: `internal/generator/deliver_template_test.go`

**Approach:**
- Parse `sink:target` once in root; pass an `io.Writer`-equivalent into command handlers.
- `stdout` = current behavior; `file:<path>` writes atomically via tmp+rename; `webhook:<url>` POSTs NDJSON or JSON depending on `--json`/`--compact` state.
- Webhook failures are surfaced as structured errors, not silenced. `--deliver-on-error <fallback>` defers to a secondary sink (usually `stdout`) so an agent loop still sees output even if the webhook is down.
- Sinks are a registry keyed by scheme. Adding `slack:` or `s3:` later is a matter of registering another handler.

**Patterns to follow:**
- `internal/generator/templates/export.go.tmpl` for file-write atomicity
- `internal/generator/templates/client.go.tmpl` for HTTP POST shape and retry budget

**Test scenarios:**
- Happy path: `--deliver stdout` produces identical output to no flag (backward compat)
- Happy path: `--deliver file:/tmp/out.json` writes atomically; process crash mid-write leaves no partial file
- Happy path: `--deliver webhook:<url>` POSTs with correct Content-Type (`application/x-ndjson` when `--compact`, otherwise `application/json`)
- Edge case: webhook returns 500 - command exits non-zero with structured error naming the URL and status code; with `--deliver-on-error stdout` the output still prints to stdout
- Edge case: unknown scheme `slack:foo` exits non-zero with a structured "unknown deliver sink: slack" listing supported schemes
- Edge case: `--deliver file:<path>` with unwritable path exits non-zero with a structured error
- Integration: with Unit 1, `... --wait --deliver webhook:<url>` POSTs the completed artifact once polling succeeds, not the initial 202 stub
- Integration: with Unit 2, `profile save briefing --deliver webhook:<url>` persists the sink; `--profile briefing` reuses it

**Verification:**
- Generated CLIs accept `--deliver` on every endpoint command and utility command (`doctor`, `agent-context`)
- All three sinks behave per the test scenarios
- Unit 1's `--wait` output routes through the deliver layer

---

- [ ] **Unit 4: `feedback` subcommand**

**Goal:** Every printed CLI ships `feedback <text>` - an in-band channel for agents (and humans) to record friction. Writes locally by default; posts upstream when opted in. Closes the "send it our way - from you or your agent" loop.

**Requirements:** R4, R6

**Dependencies:** None

**Files:**
- Create: `internal/generator/templates/feedback.go.tmpl` (new `feedback` command: accepts positional text or `--stdin`; emits confirmation JSON)
- Modify: `internal/generator/templates/root.go.tmpl` (register `feedback` subcommand)
- Modify: `internal/generator/templates/config.go.tmpl` (new `feedback_endpoint`, `feedback_auto_send` config keys)
- Modify: `internal/generator/templates/skill.md.tmpl` (agent-facing description - PostHog-style opinion: "write what surprised you, not a bug report")
- Modify: `internal/generator/templates/agent_context.go.tmpl` (include `feedback_endpoint_configured: bool` in agent-context so agents know whether feedback will ship upstream)
- Test: `internal/generator/feedback_template_test.go`

**Approach:**
- `feedback <text>` or `feedback --stdin` captures text, metadata (cli name, cli version, timestamp, agent-id env var if set), and appends a JSON line to `~/.<cli-slug>/feedback.jsonl`.
- If `feedback_endpoint` is set and either `--send` is passed or `feedback_auto_send: true` is in config, the entry is POSTed as JSON. Local write always happens first so upstream failure does not lose the entry.
- Never exit non-zero on upstream failure by default - the agent's primary goal was recording the feedback, and local write succeeded.
- `feedback list --json` returns the last N entries so an agent can audit what it has said.

**Patterns to follow:**
- `internal/generator/templates/doctor.go.tmpl` for utility subcommand shape
- `internal/generator/templates/export.go.tmpl` for append-only local file writing

**Test scenarios:**
- Happy path: `feedback "the --since flag is inclusive"` appends one JSON line to `~/.<cli-slug>/feedback.jsonl` and prints a confirmation to stdout
- Happy path: `feedback --stdin` reads from piped input, handles multi-line text, truncates at a reasonable max length with a warning
- Happy path: `feedback list --json --limit 5` returns the last five entries as NDJSON
- Edge case: empty feedback text exits non-zero with a structured error; prevents agents from submitting whitespace-only entries
- Edge case: `feedback_endpoint` unset and `--send` passed - exits zero, logs a warning that upstream delivery was skipped because no endpoint is configured
- Error path: upstream POST returns 500 with `--send` - local write still succeeded, command exits zero but returns a structured warning entry the agent can surface
- Error path: local write fails (disk full, permission denied) - exits non-zero with a structured error
- Integration: after `feedback` with `--send` and configured endpoint, `agent-context` JSON reports `feedback_endpoint_configured: true`
- Integration: `feedback list` entries survive across invocations and across different working directories

**Verification:**
- A generated CLI's `feedback` subcommand works offline by default
- Upstream POST is off by default and requires explicit opt-in
- Local file is created with 0600 permissions (agent feedback may contain sensitive text)

---

- [ ] **Unit 5: Scorecard dimension for agent-workflow readiness**

**Goal:** The scorecard recognises the four new capabilities so a correctly-built CLI is not penalised for having them and a CLI that skips them is not rewarded. Same mechanic Cloudflare learning plan used for MCP token efficiency: a new dimension with banded scoring and graceful "not applicable" handling.

**Requirements:** R5, R6

**Dependencies:** Units 1 through 4 (the features being scored must exist first)

**Files:**
- Modify: `internal/pipeline/scorecard.go` (add `agent_workflow_readiness` dimension)
- Modify: `internal/pipeline/scorecard_test.go`
- Create: `docs/solutions/best-practices/scorecard-agent-workflow-dimension-2026-04-17.md` (capture the banding rationale so future retros can tune it)
- Modify: `README.md`, `CHANGELOG.md` (document the new dimension once it ships)

**Approach:**
- One dimension, four sub-signals, banded scoring similar to the existing Tier-2 pattern: full credit when all four are present and wired, partial credit when some are present, zero when none and the CLI would benefit.
- "Not applicable" path uses the existing `UnscoredDimensions` mechanic for CLIs where the API has no async endpoints and the spec is too small to warrant profiles. The dimension reports "N/A - small synchronous spec" rather than penalising.

**Patterns to follow:**
- Cloudflare plan's `mcp_token_efficiency` dimension addition pattern
- `docs/solutions/best-practices/steinberger-scorecard-scoring-architecture-2026-03-27.md` banding

**Test scenarios:**
- Happy path: CLI with all four capabilities (async + profiles + deliver + feedback) scores full on the dimension
- Happy path: CLI where the spec has no async endpoints scores as N/A for the async sub-signal; still earns full credit on the other three
- Edge case: CLI generated before this plan landed re-scores under the new dimension - missing capabilities show up as gaps, not as hard failures that block the test matrix
- Error path: scorecard struct mismatch between old and new catalog entries - deserialiser handles absent field cleanly
- Integration: a regenerated CLI from the catalog (e.g. Linear) lands on a predicted score; drift triggers a retro

**Verification:**
- Running scorecard against a synthetic CLI fixture produces the expected banded score for each combination
- Existing CLIs in the local library do not regress on their overall grade when the dimension is added (adjust base weights if needed, documented in the solution note)

## System-Wide Impact

- **Interaction graph:** Units 1 through 4 touch template emission and root command registration; Unit 5 touches scoring. No change to the spec parser, catalog schema, or the research/sniff/emboss pipelines. MCP surface is intentionally untouched.
- **Error propagation:** New commands return the standard structured error shape (code, message, machine-readable `error.kind`). No new exit codes beyond the existing palette. Async polling failures and webhook delivery failures share the same error shape as existing HTTP failures.
- **State lifecycle risks:** The `jobs` table and `feedback.jsonl` live in the CLI's existing local-state directory. Both are append-first and safe to lose; neither is authoritative. Profiles live in the config file; corruption is recoverable via `profile delete` or manual edit.
- **API surface parity:** Units 1 through 4 add CLI surface. MCP parity is deferred to a follow-up plan. The agent-context JSON (Cloudflare plan) is extended so MCP-connected introspection still discovers the new surface.
- **Integration coverage:** Cross-layer scenarios worth verifying end to end: `--profile X --wait --deliver webhook:Y` (all four features except feedback composed in one invocation); `feedback` + `agent-context` reporting the feedback endpoint state.
- **Unchanged invariants:** Existing endpoint commands, flags, and output formats do not change. `tail`, `sync`, `workflow archive`, `doctor`, `agent-context`, `sql`, `search` all behave identically. No published CLI in the library changes until regenerated.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Async detection false-positives add `--wait` to endpoints that do not support polling, breaking `verify` | Two-of-three signal rule; dogfood's `checkAsyncJobCoverage` flags detected-but-unwaitable combinations before verify runs. |
| Async detection false-negatives silently drop `--wait` on legitimate job APIs | Capture miss cases in retros; per-spec override in the internal YAML schema (`async_hint: true`) planned as a follow-up if retros find repeated misses. |
| Profiles accumulate stale configs agents never prune | `profile list` includes `last_used`; retro may later add automatic pruning. Non-load-bearing for day one. |
| Webhook `--deliver` leaks sensitive output to the wrong endpoint | Webhook URL is never logged in plaintext; sink parsing validates scheme against the fixed registry; `doctor` surfaces configured sinks. |
| Feedback subcommand becomes a spam channel | Default is local-only; upstream POST requires explicit opt-in and a configured endpoint; no PP-hosted default endpoint is shipped. |
| Scorecard weights shift enough to regress published-CLI grades | The solution note in Unit 5 captures the calibration; grades are re-measured on the full catalog before merging Unit 5. |

## Documentation / Operational Notes

- Update `AGENTS.md` glossary with `profile`, `--wait`, `jobs`, `feedback`, `--deliver` as canonical terms (Unit 2 adds `profile`; later units add the rest as they land).
- Update `skills/printing-press/SKILL.md` to describe the four new generated-CLI capabilities so the printing-press skill itself teaches them to agents running a fresh generation.
- The solution note in Unit 5 (`docs/solutions/best-practices/scorecard-agent-workflow-dimension-2026-04-17.md`) is the durable record of dimension rationale and banding for future retros.
- No migration required: every change is additive, picked up by CLIs on regeneration.

## Sources & References

- **Origin article:** Joshua Xu, "HeyGen CLI: The Visualization Layer for AI Agents" (2026-04-13)
- Related completed plan: `docs/plans/2026-04-10-001-feat-agent-first-posthog-learnings-plan.md`
- Related active plan: `docs/plans/2026-04-13-002-feat-cloudflare-cli-learnings-plan.md` (agent-context subcommand this plan extends)
- Related solution note: `docs/solutions/best-practices/steinberger-scorecard-scoring-architecture-2026-03-27.md`
- Karpathy 2026-04-10 post quoted in the article (framing only; no implementation content)
