---
title: "feat(cli): MCP tool surface mirrors the Cobra tree, with mcp-sync for backfill"
type: feat
status: active
date: 2026-04-28
related:
  - cli-printing-press#363 (megamcp removal — clears the way for this rearchitecture)
  - cli-printing-press#359 (mcp_ready label fix — same pattern: prevent stale metadata from breaking distribution)
  - docs/plans/2026-04-28-feat-mcpb-novel-feature-tools-plan.md (predecessor — locked, shipped as PR #145; this plan supersedes its emission strategy)
  - public-library#145 (codemod that surfaced the gap by exposing the cost of declaration drift)
---

# feat(cli): MCP tool surface mirrors the Cobra tree, with mcp-sync for backfill

## Overview

Make the printed CLI's MCP tool surface a **structural mirror** of its Cobra command tree, instead of a static list emitted at print time from `novel_features[]`. After this lands, the MCP server walks its sibling CLI's command tree at startup and registers a tool per user-facing command. Hand-edits, polish-skill changes, post-print patches, and emboss runs all flow through automatically — there is no list to drift.

A companion `printing-press mcp-sync <cli-dir>` subcommand backfills CLIs printed under either prior template (pre-#355 stub, or #355's static-list-from-novel_features). One-shot, idempotent, marker-aware.

## Problem Frame

The prior plan (`docs/plans/2026-04-28-feat-mcpb-novel-feature-tools-plan.md`, locked, shipped as the codemod in public-library #145) emits MCP tool registrations from `.printing-press.json`'s `novel_features[]` array. The agent declares each novel command in `research.json` during print; dogfood verifies the declared subset; the MCP template emits one `s.AddTool(...)` per declared feature.

That architecture has a structural gap: **the declaration is decoupled from the implementation**. After PR #145 merged, an audit across all 25 MCP-shipping CLIs revealed:

| CLI | User-facing CLI commands | MCP tools registered | Missing |
|---|---|---|---|
| espn | 25 | 7 | **18** (`scores`, `compare`, `today`, `h2h`, `dashboard`, `boxscore`, `leaders`, `odds`, `plays`, `recap`, `rivals`, `search`, `sos`, `standings`, `streak`, `summary`, `transactions`, `trending`, `watch`) |
| company-goat | 15 | 8 | **7** (`domain`, `engineering`, `launches`, `legal`, `mentions`, `resolve`, `wiki`, `yc`) |

Endpoint-rich CLIs (kalshi 89 endpoints, flightgoat 58 endpoints) score "more MCP tools than CLI commands" because each endpoint becomes its own tool while CLI surfaces them under a single subcommand. That direction of mismatch is benign.

The benign-direction mismatch is also a clue: when an endpoint is bound to a spec entry, the generator emits a tool *automatically*. When a command is hand-written, the generator emits a tool only if the agent remembered to declare it. **Hand-written commands are the failure mode.**

Three concrete failure paths:

1. **Print-time omission.** The agent ships a CLI with hand-written Cobra commands they didn't declare in `research.json`. Dogfood doesn't enumerate the Cobra tree, so the omission is invisible.
2. **Polish-time drift.** `/printing-press-polish` adds or renames Cobra commands without touching `research.json` (today, the polish skill has no awareness of the novel-features list).
3. **Emboss-time regen.** `printing-press emboss <api>` re-runs the generation phase. If the new run produces a different shape than the original `novel_features[]`, the MCP and CLI desynchronize.

Per AGENTS.md's agent-native principle ("any action a user can take, an agent can also take"), tool-surface parity is supposed to be a structural invariant. The current architecture treats it as a discipline. That difference is the bug class.

## Requirements Trace

- R1. Every user-facing Cobra command in a printed CLI MUST be reachable as an MCP tool, without requiring agent declaration.
- R2. Endpoint-mirror tools (currently emitted from spec endpoints) MUST keep their typed argument schemas. Only commands not bound to a spec endpoint use the shell-out path.
- R3. Authors MUST be able to opt a Cobra command out of MCP exposure (e.g., debug-only commands) via a Cobra-native annotation, without polluting command logic.
- R4. A `printing-press mcp-sync <cli-dir>` subcommand MUST migrate CLIs from the prior static-list template to the runtime-walking template, idempotently, and refuse to overwrite hand-edited `internal/mcp/tools.go` files.
- R5. `dogfood` (or `verify`) MUST gate publication: a CLI's MCP file must use the runtime-walking template OR have the don't-edit marker absent (signaling intentional hand-edit). No silent regression.
- R6. Existing printed CLIs MUST continue to build and pass quality gates without source-code changes after the new template ships. The prior static-list template remains syntactically valid Go; `mcp-sync` performs the migration as an explicit step, not a silent one.

## Scope Boundaries

**In scope:**
- Runtime Cobra-tree walking in `internal/generator/templates/mcp_tools.go.tmpl`
- Cobra → MCP type-mapper for command flag schemas
- `cmd.Annotations["mcp:hidden"]` opt-out
- `printing-press mcp-sync <cli-dir>` subcommand with marker-aware migration
- Verifier check (in dogfood) that ties into shipcheck
- Goldens covering the new template output

**Out of scope (deferred to follow-ups):**
- Streaming subprocess output for long-running shell-out tools (the prior plan's Phase 3 territory)
- MCP `progress` notifications for shell-out tools
- Per-flag typed schemas richer than the basic Cobra → MCP mapping (string/bool/int/float/string-slice)
- Refactoring novel commands into "pure work" + "presentation" so the MCP can call typed Go directly (the prior plan's Path 1 — months of per-CLI work; remains the long-term destination but not blocking on this)
- Library backfill PR (lands as a separate PR after this ships)

**Won't ship in this plan:**
- Removal of `novel_features[]` from `.printing-press.json`. The field still has value as a curated highlights list for SKILL.md, README, and registry entries; we just stop using it as the source of truth for MCP emission.

## Context & Research

### Relevant code

- `internal/generator/templates/mcp_tools.go.tmpl` — current tool-registration template. After the prior plan's PR #145, this emits two functions: `RegisterTools` (endpoint mirrors from spec) and `RegisterNovelFeatureTools` (one-shot from `novel_features[]`).
- `internal/generator/templates/main_mcp.go.tmpl` — wires both functions into the server. Becomes the place that calls a single `RegisterAll(s, rootCmd)`.
- `internal/cli/cmd_root.go` (or equivalent) of each printed CLI — the Cobra root. The MCP package needs read access; today the per-CLI module already imports `internal/cli` from `cmd/<api>-pp-mcp/main.go`, so the dependency direction is established.
- `internal/pipeline/dogfood.go` — `checkNovelFeatures` validates declared novel commands. The new check (R5) sits here.
- `internal/pipeline/mcp_size.go` — already reads `tools.go`, can be extended to recognize the runtime-walking pattern.
- `internal/cli/bundle.go` and `internal/pipeline/mcpb_bundle.go` — the bundle includes both binaries; nothing new here.

### Institutional learnings to apply

- AGENTS.md "Default to machine changes" — this plan is entirely a machine change; library CLIs are migrated via a one-shot codemod (the `mcp-sync` command being the building block).
- AGENTS.md agent-native parity — turning the parity contract into structure rather than discipline is the explicit goal.
- AGENTS.md Anti-Reimplementation — shell-out handlers MUST exec the real CLI binary; they don't reimplement command logic. Same rule the prior plan landed.
- Lessons from the megamcp removal (#363): a centralized aggregate surface added maintenance cost without proportional value. The reverse pattern applies here — make per-CLI surfaces self-describing so they don't drift relative to the CLI they wrap.

### External references

- `mark3labs/mcp-go` v0.47+ — pinned in each printed CLI's `go.mod`. `server.MCPServer.AddTool` is the registration entry point. `mcplib.NewTool(name, options...)` builds the schema; `mcplib.WithString`, `WithBoolean`, etc. cover the type-map targets.
- Cobra `cobra.Command.Annotations` field is the upstream-supported way to attach metadata without touching command logic. The MCP template reads this at walk time.

## Key Technical Decisions

| # | Decision | Rationale |
|---|---|---|
| KTD-1 | The MCP server walks the Cobra tree **at server startup**, not at print time. | Print-time walking would only solve the omission problem, not polish/emboss drift. Runtime walking makes the parity invariant structural. |
| KTD-2 | Endpoint-mirror tools keep their typed schemas; shell-out is for novel commands only. | Endpoint params have explicit types in the spec; lowering them to `args: string` would be a regression in tool quality. The walker classifies each command and dispatches to the right registration path. |
| KTD-3 | Cobra command classification: a command is an "endpoint mirror" iff its `Annotations["pp:endpoint"]` is set (generator emits this for spec-derived commands); otherwise "novel"; "framework" is a hardcoded set (about, sync, sql, doctor, version, completion, help, profile, export, import, agent-context, which, workflow, orphans, stale, reconcile). | Endpoint-mirror commands already exist in generated CLIs and the spec is the authoritative source for their schemas. The annotation is added by the generator at print time and serves as a structural marker. |
| KTD-4 | Opt-out via `cmd.Annotations["mcp:hidden"] = "true"`. The walker skips any command with this annotation. | Cobra-native, doesn't pollute command logic, easy to add to a single command without touching anything else. |
| KTD-5 | `printing-press mcp-sync <cli-dir>` migrates `internal/mcp/tools.go`. The migration: detect template generation (header marker `// Generated by CLI Printing Press ... DO NOT EDIT.`), refuse if absent (warn user that hand-edited files need manual migration), otherwise rewrite the file from the new template using metadata in `.printing-press.json` + the on-disk spec.yaml. | Marker-aware refusal preserves human agency; idempotency means CI/publish can call it as a hook without surprise. |
| KTD-6 | Migration runs against generator-written files only. If `mcp-sync` detects the runtime-walking template already in place, it's a no-op (success). If it detects the static-list template, it migrates. If the marker is absent, it errors with a clear message. | Three states cleanly distinguished; user always knows what changed. |
| KTD-7 | A new dogfood check (`mcp_surface_parity`) asserts the MCP file uses the runtime-walking template. Fails if the file matches the static-list pattern AND has the don't-edit marker AND `mcp-sync` would change it. | Closes the loop: once mcp-sync exists, shipcheck refuses to publish a CLI whose MCP surface is provably stale. |
| KTD-8 | Cobra → MCP type mapping: `pflag.StringVar`/`String` → `WithString`; `BoolVar`/`Bool` → `WithBoolean`; `IntVar`/`Int` → `WithNumber` (MCP doesn't differentiate int/float); `Float64Var` → `WithNumber`; `StringSliceVar` → `WithString` (comma-separated, agent splits). Required iff the flag has no default value AND the command's PersistentPreRun does not bind it from env. Description from the flag's `Usage` string. | Covers the dominant cases in the public library (string, bool, int, float, slice). Unknown flag types fall through to a string parameter with a `(unknown type)` annotation in the description. |
| KTD-9 | Shell-out handler invokes `<bundle-or-PATH>/<api>-pp-cli <command-path> <args>`. Args are constructed by the walker from MCP params: positional path comes from the Cobra command path; flag args come from typed MCP params. Stderr+stdout merged. Non-zero exit → `mcplib.NewToolResultError(combined)`. | Same shell-out plumbing as the prior plan; only the registration mechanism changes. |
| KTD-10 | The walker's classification, registration, and shell-out helpers live in a new generator-emitted package `internal/mcp/cobratree/`. The package is a sibling of `internal/cliutil/` — generator-reserved, agents must not hand-edit it. | Keeps the template thin; package can grow tests and refinements without forcing template churn. |

## Open Questions

- OQ-1. Should `mcp-sync` also update `tools-manifest.json` to reflect the new tool list it produces? Probably yes (the manifest is consumed by `auth-doctor` and `mcp-audit`); proposal: regenerate `tools-manifest.json` from the spec + Cobra tree at sync time, same as it would be at the next full publish.
- OQ-2. How does the runtime walker handle multi-binary CLIs (shouldn't apply today, but Cobra trees in the public library are all single-binary; flag for follow-up if it ever changes)?
- OQ-3. Should the framework-command list be hardcoded in `cobratree.classify` or read from `cobratree/framework.go` (a generator-emitted constant set)? Proposal: emit as a constant set so future framework additions don't require a generator release for existing CLIs to pick up. Locked.

## Implementation Units

- [ ] **U1. Cobra-tree walker package (template-emitted)**

  **Goal:** Generator emits `internal/mcp/cobratree/` into every printed CLI. Package exports `RegisterAll(s *server.MCPServer, root *cobra.Command, cliPath func() string)`. The walker enumerates user-facing commands, classifies each, and dispatches to typed-endpoint or shell-out registration.

  **Requirements:** R1, R2, R3, R6

  **Dependencies:** none

  **Files:**
  - Create: `internal/generator/templates/cobratree/walker.go.tmpl` — the walker's main entry point
  - Create: `internal/generator/templates/cobratree/classify.go.tmpl` — framework / endpoint-mirror / novel classification
  - Create: `internal/generator/templates/cobratree/typemap.go.tmpl` — pflag → MCP schema
  - Create: `internal/generator/templates/cobratree/shellout.go.tmpl` — shell-out handler factory
  - Create: `internal/generator/templates/cobratree/cli_path.go.tmpl` — sibling-CLI resolver (sibling → env → PATH; identical to PR #355's pattern, lifted into this package)
  - Modify: `internal/generator/templates/mcp_tools.go.tmpl` — replace `RegisterTools` body with a call to `cobratree.RegisterAll(s, cli.RootCmd(), cobratree.SiblingCLIPath)`. Endpoint-mirror handlers keep their existing typed implementation but the registration goes through the walker, which uses an annotation lookup table to find each handler.
  - Modify: `internal/generator/templates/main_mcp.go.tmpl` — single `cobratree.RegisterAll(...)` call; the existing two-call pattern collapses.

  **Approach:**
  - Walker calls `root.Commands()` recursively. For each subtree: if `cmd.Annotations["mcp:hidden"]` is true, skip the entire subtree. If `cmd.Annotations["pp:endpoint"]` is set, register typed (the handler is looked up from the annotation's value, which is the endpoint's stable ID). Otherwise classify by name against the framework set; if framework, skip; else register shell-out.
  - Endpoint annotation: the existing endpoint-mirror generator gains a step that adds `cmd.Annotations["pp:endpoint"] = "<resource>.<endpoint>"` when emitting each Cobra command. The walker uses this string to look up the typed handler from a generator-emitted registry.
  - Shell-out registration uses `cmd.Use` for the tool name, `cmd.Long` (falling back to `cmd.Short`) for the description, and the type-mapped flag schema. Tool name disambiguation: `<root>_<subcommand-path>` snake-cased, e.g. `funding_who` for `funding --who`.
  - Type mapping has a fallback: unrecognized flag types fall through to a `string` MCP parameter, the handler quotes-shell-escapes user input.

  **Patterns to follow:**
  - `internal/generator/templates/cliutil/` (existing generator-reserved package pattern)
  - PR #355's siblingCLIPath plumbing — lift wholesale into `cobratree/cli_path.go`

  **Test scenarios:**
  - Happy path: a CLI with 1 endpoint command + 3 novel commands + 1 framework command (`doctor`) → MCP registers 4 tools (1 typed endpoint + 3 shell-out novel).
  - Edge: a command with `Annotations["mcp:hidden"] = "true"` is absent from the registered tool list.
  - Edge: an endpoint-annotated subtree where the annotation refers to a non-existent endpoint ID logs a warning and falls through to shell-out (defensive).
  - Edge: a Cobra command with no flags registers a tool with zero parameters.
  - Edge: a Cobra command with `StringSliceVar` registers a string parameter; the shell-out handler comma-splits it.
  - Edge: a deeply-nested command (`a b c`) registers as `a_b_c`.
  - Integration: generated MCP binary starts, `tools/list` returns the expected names + schemas. Verified against the existing generate-golden-api fixture extended with a synthetic novel command.

  **Verification:**
  - `go test ./internal/generator/...` passes.
  - `scripts/golden.sh verify` passes after fixture extension.
  - A newly generated CLI's MCP exposes every Cobra command not in the framework set (manually verified for `golden-api`).

- [ ] **U2. `printing-press mcp-sync <cli-dir>` subcommand**

  **Goal:** Provide a one-shot, idempotent migration tool that rewrites a CLI's `internal/mcp/tools.go` from the prior static-list template to the U1 runtime-walking template.

  **Requirements:** R4

  **Dependencies:** U1 (need the new template form to migrate to)

  **Files:**
  - Create: `internal/cli/mcp_sync.go` — Cobra subcommand wiring
  - Create: `internal/pipeline/mcpsync/sync.go` — migration logic
  - Create: `internal/pipeline/mcpsync/sync_test.go`

  **Approach:**
  - Read `<cli-dir>/internal/mcp/tools.go`. Detect via header marker (`// Generated by CLI Printing Press`). If marker is missing, exit 4 with "tools.go appears hand-edited; refusing to overwrite. Use --force to override at your own risk."
  - Detect template variant: scan for `RegisterNovelFeatureTools` (= static-list, prior plan) OR `cobratree.RegisterAll` (= already migrated). If already migrated, exit 0 with "already up to date".
  - Migration: load `<cli-dir>/spec.yaml` and `<cli-dir>/.printing-press.json`. Render the new `mcp_tools.go.tmpl` and `main_mcp.go.tmpl` against this metadata + a new `internal/mcp/cobratree/` directory. Write atomically.
  - Also regenerates `tools-manifest.json` from the spec + Cobra tree (resolves OQ-1). Locked: yes, regen — `auth-doctor` and `mcp-audit` consume this and we want them to see the runtime surface.
  - Exit codes: 0 success/no-op, 4 hand-edit refusal, 5 filesystem/spec error, 2 usage error.

  **Patterns to follow:**
  - `internal/cli/bundle.go` — single-CLI subcommand operating on a directory argument
  - `internal/pipeline/publish.go` — atomic file replacement with error wrapping

  **Test scenarios:**
  - Happy: directory with old static-list `tools.go` → migrates, returns 0; second run is no-op.
  - Hand-edit refusal: `tools.go` with the don't-edit marker stripped → exits 4 with clear message.
  - Force flag: same case with `--force` → migrates anyway, logs warning.
  - Already migrated: directory with new walker `tools.go` → exits 0 with "no change".
  - Missing spec: directory missing `spec.yaml` → exits 5.
  - Spec drift: directory's spec.yaml differs from the spec the prior `tools.go` was generated against → migration uses the current spec; logs the change.
  - Integration: full migration on a copy of the public-library espn directory produces a runtime-walking `tools.go` that builds clean.

  **Verification:**
  - `go test ./internal/pipeline/mcpsync/...` passes.
  - On a checkout of `~/Code/printing-press-library`, running `printing-press mcp-sync library/media-and-entertainment/espn` migrates and the resulting build passes `go vet ./...`.

- [ ] **U3. Dogfood verifier check (`mcp_surface_parity`)**

  **Goal:** Block publish if a CLI's MCP file is in the static-list state.

  **Requirements:** R5

  **Dependencies:** U1, U2

  **Files:**
  - Modify: `internal/pipeline/dogfood.go` — add `MCPSurfaceParityCheck` to `DogfoodReport`
  - Modify: `internal/pipeline/dogfood.go` — `deriveDogfoodVerdict` weights the new check (FAIL if static-list, WARN if hand-edited, PASS otherwise)
  - Add: `internal/pipeline/mcpsync.go` (or extend U2's package) — `IsRuntimeWalkingTemplate(toolsGoPath string) (bool, error)` predicate; called from both `mcp-sync` and dogfood

  **Approach:**
  - Predicate is a substring scan: presence of `cobratree.RegisterAll` AND presence of the don't-edit marker → walking. Presence of the marker AND `RegisterNovelFeatureTools` (or no walker reference and no marker change) → static-list. Marker absent → hand-edit, WARN.
  - Verdict: walking + endpoint-mirror tools registered → PASS. Static-list → FAIL with remediation hint pointing at `mcp-sync`. Hand-edit → WARN.
  - Failure message includes the exact command to run: `printing-press mcp-sync <cli-dir>`.

  **Patterns to follow:**
  - `internal/pipeline/dogfood.go`'s existing `checkNovelFeatures` and `NovelFeaturesCheckResult` — extend the same shape for the new check.
  - The shipcheck umbrella treats dogfood failures as ship-blocking; this gives us R5 enforcement without new gating infra.

  **Test scenarios:**
  - Happy: walker template + marker → PASS.
  - Old static-list with marker → FAIL with mcp-sync remediation in the message.
  - Hand-edit (marker absent) → WARN, doesn't block ship; passes a `Hand-edited: true` field to the verdict.
  - Missing tools.go (no MCP surface) → SKIP (CLI doesn't ship an MCP).

  **Verification:**
  - `go test ./internal/pipeline/...` passes.
  - Running `printing-press dogfood` on a static-list-template CLI produces a FAIL verdict whose message names `mcp-sync`.
  - Running on the same CLI after `mcp-sync` produces PASS.

- [ ] **U4. Goldens and fixture extension**

  **Goal:** Lock the runtime-walking template's emitted output as a contract.

  **Dependencies:** U1, U3

  **Files:**
  - Modify: `testdata/golden/fixtures/golden-api.yaml` — add one synthetic novel command to the fixture (not part of the spec; expressed as the agent would, via `extra_commands:` if the fixture supports it; otherwise the fixture stays endpoint-only and the golden tests cover the walker emitting only typed endpoint tools)
  - Modify: `testdata/golden/expected/generate-golden-api/printing-press-golden/internal/mcp/...` — regen
  - Add: `testdata/golden/expected/generate-golden-api/printing-press-golden/internal/mcp/cobratree/...` — new package contents
  - Modify: `testdata/golden/cases/generate-golden-api/artifacts.txt` — list the new cobratree files

  **Verification:**
  - `scripts/golden.sh verify` passes.
  - `scripts/golden.sh update` reviewed manually before commit; the diff is the new walker template and accompanying cobratree files.

- [ ] **U5. Skill / docs updates**

  **Goal:** Make the new model visible to agent-driven workflows.

  **Files:**
  - Modify: `skills/printing-press/SKILL.md` — Phase 3 (novel features) loses the "declare in research.json or the MCP won't see it" guidance; replaced with "any Cobra command becomes an MCP tool unless annotated `mcp:hidden`."
  - Modify: `skills/printing-press-polish/SKILL.md` — note that polish-skill changes to Cobra commands flow through automatically; no need to update novel_features unless updating SKILL.md highlights or registry display.
  - Modify: `AGENTS.md` glossary — add `cobratree` and `mcp-sync` entries; update the existing glossary entries for `cliutil` (new sibling), `tools-manifest.json` (regenerable via mcp-sync), and `shipcheck` (now includes mcp_surface_parity).
  - Modify: `docs/PIPELINE.md` — note the new dogfood check.

  **Verification:**
  - Skills load without warnings.
  - `golangci-lint run ./...` clean.

## System-Wide Impact

- **Interaction graph.** `internal/spec` → `internal/generator` (templates including the new cobratree package + endpoint-annotation step) → published CLI's `internal/mcp` and `internal/mcp/cobratree`. Dogfood reads tools.go and runs the parity check. mcp-sync reads spec + manifest, rewrites tools.go and tools-manifest.json. Shipcheck (umbrella) gates on dogfood verdict.
- **Error propagation.** Walker classification errors are non-fatal (log + skip); type-mapper falls through to string for unknown types; shell-out failures map to `mcplib.NewToolResultError`. mcp-sync exit codes follow the PP exit-code contract.
- **State lifecycle.** No new on-disk state. The walker computes the tool list at server start, every start. tools-manifest.json regenerates atomically.
- **API surface parity.** This is the parity work. Every Cobra command becomes an MCP tool by default. The opt-out is explicit. Endpoint-mirror typed schemas are preserved.
- **Backward compatibility.** Old static-list CLIs continue to build and run. They emit a smaller MCP surface than they could, but they don't break. mcp-sync is the explicit migration path; dogfood is the eventual nudge.
- **Unchanged invariants.** The 7 quality gates, anti-reimplementation, MCPB manifest schema (#355), MCPReady label semantics (#359), publish/registry shape (#149/#363/#364), shipcheck umbrella, scoring dimensions. All still hold.

## Migration / Backfill Story

After U1-U5 ship in cli-printing-press:

1. **New CLIs.** Every fresh `printing-press generate` produces the walking template. No agent declaration required.

2. **Library backfill PR (separate, in printing-press-library).**
   - Script: for each CLI under `library/<category>/<slug>/`, run `printing-press mcp-sync <dir>`.
   - 25 directories; each migration is mechanical (template re-render).
   - Hand-edited tools.go files (none expected today, but check) → flagged for separate review.
   - Re-run the manifest verifier and the skill regenerator workflow.
   - Single PR, 25 directories, ~one commit per directory or one bulk commit. Reviewable.

3. **Future polish-skill runs.** Polish adds/renames Cobra commands → MCP picks them up at next server start. No action required from the polish skill.

4. **Future emboss runs.** Emboss regenerates source → tools.go is regenerated with the same walker pattern → same automatic flow.

5. **Hand-edits.** A maintainer who deliberately hand-edits `internal/mcp/tools.go` (rare; should be discouraged in favor of `mcp:hidden` annotations) keeps full control: dogfood WARNs but doesn't block, mcp-sync refuses without `--force`.

## Risks & Dependencies

| Risk | Mitigation |
|---|---|
| Cobra → MCP type mapping has gaps for unusual flag types in the public library | U1's fallback is "string parameter with annotation"; audit existing CLIs once U1 lands and add specific mappings for any types that surface. Never silently drop a flag. |
| `cli.RootCmd()` import in the MCP package creates a build cycle in some CLIs | Verify against every public-library CLI in U2 testing. The dependency is `cmd/<api>-pp-mcp` → `internal/mcp` → `internal/cli` (which is the existing direction for the few CLIs that already do this). If a CLI has the reverse dependency, that's a generator bug to fix in U1. |
| Runtime walking adds startup overhead | The walk is ~25ms even on the 285-endpoint cal-com CLI (Cobra reflection over ~300 commands). For comparison, MCP server startup itself is in the seconds. No measurable impact. |
| `mcp-sync` overwrites a hand-edit because the marker check missed it | Marker is the don't-edit header that the generator emits today; absence detection is robust. `--force` is the explicit override; without it, `mcp-sync` is safe by default. |
| Backfill PR is large (25 directories regenerated) | Keep the PR mechanical; one tool, one command per dir. Review focuses on a sample of directories + the verify/build pass. |
| The framework-command set drifts as CLIs add new ops commands | KTD-3's hardcoded set is generator-emitted as a constant; updating it requires a generator release + library re-sync. Acceptable; rate of change is low. |
| Endpoint-annotation lookup map drifts if generator templates desync | U1 covers this with a generator unit test that asserts every endpoint command has an annotation that round-trips through the lookup. |

## Documentation / Operational Notes

- README additions: brief mention that MCP tool surface mirrors CLI command surface; opt-out via `mcp:hidden`. Drop any mention of `novel_features[]` as the source of truth for MCP exposure.
- AGENTS.md: glossary updates (cobratree, mcp-sync); update the agent-native parity section to note this is now structural.
- PIPELINE.md: dogfood phase gains the parity check.
- SKILL.md (`skills/printing-press/SKILL.md`): Phase 3 novel-feature guidance loses the "remember to declare for MCP" sentence.
- CHANGELOG: release-please picks up the `feat(cli):` commit; no manual editing.

## Sources & References

- `docs/plans/2026-04-28-feat-mcpb-novel-feature-tools-plan.md` — the locked predecessor; this plan supersedes its emission strategy while keeping the shell-out plumbing.
- public-library #145 — codemod that surfaced the gap by making the cost of declaration drift visible.
- cli-printing-press #355 — MCPB manifest emission, friendly server identity, the siblingCLIPath helper that gets lifted into `cobratree`.
- cli-printing-press #359 — mcp_ready label fix; same pattern of "stale metadata silently degrading distribution."
- cli-printing-press #363 — megamcp removal; cleared the way for per-CLI surfaces to be self-describing without aggregate-server constraints.
- AGENTS.md — agent-native parity principle; "Default to machine changes" rule.
