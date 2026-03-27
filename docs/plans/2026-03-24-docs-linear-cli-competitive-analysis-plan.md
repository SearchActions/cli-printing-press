---
title: "Linear CLI Competitive Analysis - Our CLI vs Community vs MCP"
type: docs
status: active
date: 2026-03-24
---

# Linear CLI Competitive Analysis

## Overview

We generated a Linear CLI from the printing press (12 resources, 45 commands, Go binary). This document compares it against the three strongest community CLIs and the official Linear MCP server to understand where we win, where we lose, and what to do about it.

## The Contenders

| Tool | Language | Stars | Commands | Install | Primary Audience |
|------|----------|-------|----------|---------|-----------------|
| **printing-press linear-cli** (ours) | Go | - | 45 | `go install` | Developers + agents |
| **schpet/linear-cli** | TypeScript/Deno | 507 | ~20 | Homebrew, npm, binaries | Developers + agents |
| **Finesssee/linear-cli** | Rust | 58 | 16+ subcommands, 38 AI skills | cargo install, binaries | Power users + agents |
| **czottmann/linearis** | TypeScript/Node | 163 | ~15-20 | npm install | AI agents (token-optimized) |
| **Linear MCP** (official) | Remote server | - | 23 tools | `claude mcp add` | AI agents (OAuth) |

## Head-to-Head Comparison

### Command Coverage

| Category | Ours (45) | schpet (~20) | Finesssee (16+) | linearis (~18) | MCP (23) |
|----------|-----------|-------------|-----------------|---------------|----------|
| Issues CRUD | list, mine, get, create, update, search, archive, delete | list, view, create, update, start, comment | Full lifecycle + bulk ops | CRUD + relations | list, get, create, update |
| Comments | list, create, delete | Inline with issues | Inline | CRUD | create only |
| Projects | list, get, create, update | list, view | Full CRUD + labels | CRUD | list, get, create, update |
| Cycles | list, current, get | - | Full + burndown charts | list, active | list |
| Teams | list, get | list | Full + workflows | list | list, get |
| Users | me, list | - | - | list | list, get |
| Labels | list, create | - | Full CRUD | Batch operations | list, create |
| Workflows | list | - | - | - | list (as statuses) |
| Documents | list, get, create | list, view, create | - | CRUD | list, get |
| Notifications | list | - | - | - | - |
| Webhooks | list, create, delete | - | Full + HMAC verification | - | - |
| Organization | get | - | - | - | - |
| **Smart filters** | --state, --assignee me, --team, --priority, --project, --label | Basic | Extensive | Basic | Basic |
| **Git integration** | No | Branch creation, PR gen | Branch + jujutsu | No | No |
| **Watch mode** | No | No | Real-time polling | No | No |

**Verdict: We have the widest command coverage (45 vs next-best ~20).** Finesssee has deeper features per command (burndown charts, watch mode). schpet has the best git workflow integration. We beat all of them on breadth.

### Developer Experience

| Aspect | Ours | schpet | Finesssee | linearis | MCP |
|--------|------|--------|-----------|---------|-----|
| **Startup time** | <50ms (Go binary) | <100ms (Deno binary) | <50ms (Rust binary) | 0.5-1s (Node.js) | N/A (remote) |
| **Install friction** | `go install` (needs Go) | `brew install` (zero friction) | `cargo install` (needs Rust) | `npm install -g` (needs Node) | `claude mcp add` |
| **Zero deps?** | Yes (static binary) | Yes (pre-built binaries) | Yes (pre-built binaries) | No (Node.js runtime) | No (OAuth + SSE) |
| **Auth setup** | `export LINEAR_API_KEY=...` | `linear auth` (interactive) | `linear-cli login` | `export LINEAR_API_KEY=...` | OAuth browser flow |
| **Output format** | Auto-table + --json | Text + --json | Text + --json | JSON-first | JSON (to agent) |
| **Dry-run mode** | Yes (shows GraphQL query) | No | No | No | No |
| **Error messages** | Structured exit codes | Good | Good | Minimal | Opaque |
| **Shell completion** | Yes (Cobra built-in) | Yes | Yes | No | N/A |
| **Rate limit handling** | Auto-retry with backoff | Manual | Manual | Manual | Server-side |

**Verdict: We're competitive on DX.** schpet wins on install friction (Homebrew). Our dry-run mode showing the exact GraphQL query is unique and valuable for debugging. Our auto-retry on rate limits is a differentiator.

### Agent Efficiency (Token Cost)

This is the real battleground - how much does it cost an AI agent to use each tool?

| Metric | CLI (any) | MCP (official) | Ratio |
|--------|-----------|---------------|-------|
| **Initial context load** | ~300 tokens | ~15,540 tokens | 51x cheaper |
| **Per-tool invocation** | ~50-100 tokens | ~200-500 tokens | 3-5x cheaper |
| **50-device compliance check** | ~4,150 tokens | ~145,000 tokens | 35x cheaper |
| **Token Efficiency Score** | 202 | 152 | 33% advantage |
| **Task completion rate** | 28% higher | Baseline | - |

Sources: UBOS benchmarks, Jannik Reinhard Intune study, Fiberplane MCP analysis

**Verdict: Any CLI approach crushes MCP on token efficiency.** The 35-51x difference is structural - MCP loads full tool schemas into context; CLI tools are invoked as-needed.

### Reliability

| Issue | CLI (any) | MCP (official) |
|-------|-----------|---------------|
| **Connection stability** | Stateless HTTP - never disconnects | SSE connections degrade after ~1 hour |
| **Auth token refresh** | N/A (API key) | OAuth tokens expire, auto-refresh fails for SSE |
| **Status visibility** | Exit code 0 or error | Green checkmark masks failures |
| **Error recovery** | Retry and continue | Toggle off/on, restart agent session |
| **Multi-tool conflicts** | None | Linear MCP breaks detection of other MCPs |

**Known MCP issues (with sources):**
- Claude Code #36307: OAuth browser auth flow never triggers for Linear MCP
- Cursor Forum #131713: "Linear MCP constantly going red, eventually fails in agent chat"
- Cursor Forum #148816: "SSE error: undefined" after ~1 hour, requires toggle off/on
- Cursor Forum #152224: Linear MCP interferes with Jira/Atlassian MCP detection
- Gemini CLI #4031: Linear MCP breaks tool-calling

**Verdict: CLI wins decisively on reliability.** MCP's persistent connection model is fundamentally fragile. The "constantly going red" problem is systemic, not a bug that will be fixed.

### What We're Missing (Gaps to Close)

| Gap | Who Has It | Priority | Difficulty |
|-----|-----------|----------|------------|
| **Homebrew tap** | schpet | High | Easy - just create a tap repo |
| **Git branch integration** | schpet, Finesssee | Medium | Medium - detect branch name, resolve to issue |
| **Pre-built binaries** | schpet, Finesssee | High | Easy - GoReleaser already scaffolded |
| **Assignee/label resolution in create** | Finesssee | Medium | Medium - need resolve-then-set pattern |
| **Watch mode** | Finesssee | Low | Medium - polling loop |
| **Burndown charts** | Finesssee | Low | Medium - ASCII chart rendering |
| **Claude Code skill** | schpet | High | Easy - we already have the skill system |
| **Token-optimized output** | linearis | Medium | Easy - --json already works |

## Strategic Position

### Where We Win

1. **Broadest command coverage** (45 commands vs next-best ~20)
2. **Smart GraphQL filter composition** (--state + --team + --assignee compose into one optimized query)
3. **Dry-run mode** (unique - shows exact GraphQL query, invaluable for debugging)
4. **Auto-retry with rate limit backoff** (none of the others do this)
5. **Identifier resolution** (ENG-123 auto-resolves to UUID for mutations)
6. **State/assignee name resolution** (pass human-readable names, not UUIDs)
7. **Go binary** (instant startup, static binary, cross-compile trivially)
8. **Generated from spec** (can regenerate when Linear's API changes, not hand-maintained)

### Where We Lose

1. **Install friction** - `go install` requires Go toolchain; schpet has Homebrew
2. **Git workflow** - schpet auto-creates branches from issues; we don't
3. **Community/stars** - schpet has 507 stars and active community
4. **No pre-built binaries yet** - GoReleaser is scaffolded but not configured
5. **No Claude Code skill integration** - schpet ships as a Claude Code plugin

### Where We Tie

1. **Token efficiency** - All CLIs are roughly equivalent (any CLI >> MCP)
2. **Startup time** - Go and Rust binaries are both <50ms
3. **JSON output** - Everyone has --json
4. **Error handling** - Similar quality across Go/Rust/Deno CLIs

## The MCP Comparison (Why CLI Wins)

The official Linear MCP server has 23 tools and OAuth-based auth. On paper, it's the easiest integration for AI agents. In practice:

1. **It disconnects constantly.** SSE connections degrade after ~1 hour. Users report "constantly going red." The fix is toggling it off and on.

2. **It costs 35-51x more tokens.** The full tool schema loads into context on every session. A CLI invocation is a single bash command.

3. **OAuth auth flow is broken.** Claude Code issue #36307 - the browser auth flow never triggers. Users can't authenticate.

4. **It breaks other MCPs.** Installing Linear MCP can prevent detection of Jira, GitHub, and Atlassian MCPs in the same session.

5. **Errors are invisible.** The status indicator shows green while requests silently fail. There's no structured error reporting.

**Bottom line: The MCP is architecturally unsuitable for reliable agent workflows.** A stateless CLI with API key auth is simpler, cheaper, and more reliable by design.

## Recommended Next Steps

### Phase 1: Ship It (make it installable)

- [ ] Configure GoReleaser for pre-built binaries (darwin/linux/windows, arm64/amd64)
- [ ] Create Homebrew tap (`mvanhorn/tap/linear-cli`)
- [ ] Publish to GitHub Releases
- [ ] Write a README with install instructions and usage examples

### Phase 2: Close the Gaps

- [ ] Git branch integration: detect `ENG-123-feature` branch, auto-set context
- [ ] Assignee/label resolution in `issues create` (resolve names to IDs)
- [ ] Claude Code skill: ship as `/linear` slash command
- [ ] `issues start` command: create branch + move to "In Progress"

### Phase 3: Differentiate Further

- [ ] `issues triage` command: show unassigned + untriaged issues with priority suggestions
- [ ] `cycles burndown` command: ASCII burndown chart in terminal
- [ ] `issues bulk` command: update multiple issues at once
- [ ] Local caching: cache team/user/label IDs to avoid repeated lookups

## Sources

### MCP Issues
- [Claude Code #36307: OAuth never triggers](https://github.com/anthropics/claude-code/issues/36307)
- [Cursor Forum #131713: Constantly going red](https://forum.cursor.com/t/linear-mcp-constantly-going-red-eventually-fails-in-agent-chat/131713)
- [Cursor Forum #148816: SSE error undefined](https://forum.cursor.com/t/linear-mcp-commonly-errors-out-and-requires-turning-off-then-on/148816)
- [Gemini CLI #4031: Breaks tool-calling](https://github.com/google-gemini/gemini-cli/issues/4031)

### Token Benchmarks
- [UBOS: CLI vs MCP Token Cost Savings](https://ubos.tech/news/cli-vs-mcp-token-cost-savings-and-lazy-loading-explained/)
- [Jannik Reinhard: CLI beats MCP 35x](https://jannikreinhard.com/2026/02/22/why-cli-tools-are-beating-mcp-for-ai-agents/)
- [Fiberplane MCP Analysis](https://blog.fiberplane.com/blog/mcp-server-analysis-linear/)

### Community CLIs
- [schpet/linear-cli](https://github.com/schpet/linear-cli) - 507 stars, TypeScript/Deno
- [Finesssee/linear-cli](https://github.com/Finesssee/linear-cli) - 58 stars, Rust
- [czottmann/linearis](https://github.com/czottmann/linearis) - 163 stars, TypeScript/Node
