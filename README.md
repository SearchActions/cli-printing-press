# CLI Printing Press

Just making a CLI is not hard. Making a CLI that understands the power user is extremely hard.

```bash
/printing-press Discord
/printing-press Stripe
/printing-press Linear
```

One command. 8 phases. ~1 hour. Produces a production-ready Go CLI + 7 analysis documents. REST or GraphQL.

## 323 Commands vs. 12 Commands

Discord's API has 300+ endpoints. The printing press can generate a CLI that wraps all 323 of them. But [discrawl](https://github.com/steipete/discrawl) - Peter Steinberger's Discord tool - has **12 commands** and **551 stars**.

Why does the 12-command tool win?

Because discrawl saw something in Discord's data that Discord itself didn't design for: **conversations are institutional knowledge**. Every message thread is a document that should be archived, indexed, and searched. Discrawl mirrors Discord into local SQLite with FTS5 full-text search, and those 12 commands - `sync`, `search`, `messages`, `mentions`, `members`, `sql`, `tail` - are worth more than 323 endpoint wrappers because they embody a deep understanding of what power users actually need.

That understanding is the **Non-Obvious Insight**. And until now, discovering it required being Peter Steinberger.

## The Non-Obvious Insight

Every API has a secret identity. The data it exposes is useful for something its creators never designed for. The printing press finds that secret and builds a CLI around it.

The **Non-Obvious Insight (NOI)** is a one-sentence reframe:

```
"[API] isn't just [obvious thing]. It's [non-obvious thing].
 Every [data point] is a signal about [hidden truth]."
```

| API | What they think it is | What it actually is |
|-----|----------------------|---------------------|
| Discord | A chat app | A **searchable knowledge base**. Every message thread is institutional memory. |
| Linear | An issue tracker | A **team behavior observatory**. Every state change is a signal about how your team actually works vs. how they think they work. |
| Stripe | A payment processor | A **business health monitor**. Every failed charge and churn event is a signal about product-market fit. |
| GitHub | A code host | An **engineering culture fingerprint**. Every review turnaround and merge pattern is a signal about how your team ships. |
| Notion | A doc editor | A **knowledge decay detector**. Every stale page and orphaned database is a signal about what your team has forgotten. |
| Slack | Messaging | An **organizational nervous system**. Every response time and channel silence is a signal about team health. |

The NOI is the creative DNA of every CLI the press generates. Phase 0 cannot complete without one. If the LLM can't write an NOI, the research wasn't deep enough.

The printing press automates what Steinberger does intuitively: look at an API, see what power users actually do with it, and build the 12 commands that matter instead of the 323 that don't.

## How I Knew This Was Real

I was deciding which Google Workspace CLI to use. Peter Steinberger's [gogcli](https://github.com/steipete/gogcli) (6.5K stars, Go) or Google's official [Workspace CLI](https://github.com/googleworkspace/cli) (10K+ stars in a week, Rust, dynamically generated from Google's Discovery Service).

I ran [/last30days](https://github.com/mvanhorn/last30days-skill) - my recency research skill that searches Reddit, X, YouTube, and the web for what people actually say about tools. It searched 34 X posts (1,437 likes), 5 YouTube videos (57K views), and 10 web sources.

The verdict surprised me: **use gogcli**. The newer, official tool with 10x the API coverage lost to the older third-party one. As [@7dyhn4542y put it on X](https://x.com): "my preference is 100% gogcli since I have my agent working a lot with Google Docs and sheets, and gogcli just makes him able to do what he needs to do."

Google's CLI wraps every endpoint but doesn't understand the user. Steinberger's CLI understands what people actually do with Gmail, Calendar, and Sheets - and builds human-friendly commands around those workflows. Setup is `brew install gogcli` vs. a multi-step Google Cloud Console OAuth dance.

That's the NOI again. Breadth doesn't beat depth. Understanding the user beats understanding the API. And /last30days saw it in the community data before I could see it myself.

-- [Matt Van Horn](https://github.com/mvanhorn)

## The Creativity Ladder

Most API CLIs stop at Rung 1. The printing press climbs to Rung 5.

| Rung | What It Is | Auto-Generated? | Example |
|------|-----------|-----------------|---------|
| 1 | API wrapper commands | Yes (from spec) | `issue create --title "..."` |
| 2 | Output formatting | Yes (always) | `--json`, `--select`, `--csv`, `--dry-run` |
| 3 | Local persistence | Yes (conditional) | `sync`, `search`, `export`, `tail` |
| 4 | **Domain analytics** | **Yes (from archetype)** | `stale --days 30`, `orphans`, `load` |
| 5 | **Behavioral insights** | **Yes (from archetype)** | `health` (composite score), `similar` (duplicate detection) |

Rung 3 is table stakes. Rung 4 is where discrawl lives. Rung 5 is where nobody else is yet.

The press generates the 323-command wrapper in Phase 2. Then it generates the discrawl-style commands automatically from domain archetype templates. That's the difference between a spec compiler and an intelligence engine.

## Why CLIs (Not APIs, Not MCP)

The NOI is the creative intelligence. CLIs are the delivery mechanism. Here's why they win for agents:

**100x fewer tokens.** An MCP server loads [~55,000 tokens](https://manveerc.substack.com/p/mcp-vs-cli-ai-agents) of tool definitions per session. A CLI `--help` costs ~200 tokens. At 10K sessions/day, that's $1,600/day saved.

**Training data advantage.** LLMs were trained on millions of shell interactions. When an agent sees `mycli list --json | jq '.[] | select(.status == "active")'`, it already knows. [MCP composition patterns have zero training data](https://manveerc.substack.com/p/mcp-vs-cli-ai-agents).

**Self-healing delegation.** [CLI agents are designed for delegation, not suggestion.](https://www.firecrawl.dev/blog/why-clis-are-better-for-agents) Exit code 0 = done. Exit code 1 = try again. No screenshots, no clicking, no UI fragility.

```
Power User (architect)  -->  Agent (operator)  -->  CLI (interface)  -->  API
  "Find stale issues"      runs the command       linear-cli stale      GraphQL
  "Who's overloaded?"      parses JSON output     linear-cli load       queries
  "Fix the auth bug"       chains 5 commands      linear-cli issue...   mutations
```

Every API that gets a CLI becomes instantly accessible to Claude Code, Codex, Gemini CLI, and every open source agent. The printing press is the factory.

## Domain Archetypes

The profiler classifies every API into a domain archetype and auto-generates the right workflow + insight commands:

| Archetype | Detected By | Auto-Generated Commands |
|-----------|------------|------------------------|
| **Project Management** | issue/task/ticket resources, assignee fields, priority levels | `stale`, `orphans`, `load`, `health`, `similar` |
| **Communication** | message/channel/thread resources, threading fields | `channel-health`, `message-stats`, `health`, `similar` |
| **Payments** | charge/payment/invoice resources, amount/currency fields | `reconcile`, `revenue`, `health`, `similar` |
| **Infrastructure** | server/deploy/instance resources | `health`, `similar` |
| **Content** | document/page/block resources | `health`, `similar` |

The archetype is detected automatically from the spec. The entity mapper figures out which resource is the "primary entity" (issues for PM, messages for comms, charges for payments) and wires the templates accordingly.

## How It Works

8 mandatory phases. Each phase writes a plan document. The artifacts are the product.

```
Phase 0    Visionary Research       (3-5 min)    NOI + domain identity + usage patterns
Phase 0.5  Power User Workflows     (2-3 min)    Compound commands power users want
Phase 0.7  Prediction Engine        (15-25 min)  SQLite schema + FTS5 + sync strategy
Phase 1    Deep Research            (5-8 min)    Competitors, strategic justification
Phase 2    Generate                 (1-2 min)    Go CLI from spec + archetype templates
Phase 3    Steinberger Audit        (5-8 min)    Two-tier quality scoring (100 points)
Phase 4    GOAT Build               (5-10 min)   Domain tables, workflow commands, insights
Phase 4.5  Dogfood Emulation        (10-20 min)  Test every command against spec mocks
Phase 4.6  Hallucination Audit      (5 min)      Dead flags, dead functions, ghost tables
Phase 5    Final Steinberger        (2-3 min)    Before/after delta + report
```

## What Gets Generated

**Agent-first flags** (every command): `--json`, `--select`, `--dry-run`, `--stdin`, `--csv`, `--compact`, `--quiet`, `--yes`, `--no-input`, `--no-cache`, `--no-color`. Auto-JSON when piped (no `--json` needed). Typed exit codes (`0`=success, `2`=usage, `3`=not found, `4`=auth, `5`=API, `7`=rate limited).

**Actionable errors**: errors include the specific flag/arg that's wrong, the correct usage pattern, and the command path. Agents self-correct in one retry.

**Bounded output**: list commands show "Showing N results. To narrow: add --limit, --json --select, or filter flags." Token-conscious `--compact` mode returns only high-gravity fields (id, name, status, timestamps) - 60-80% fewer tokens.

**Data layer** (high-gravity entities): domain-specific SQLite tables with proper columns (not JSON blobs), FTS5 full-text search, incremental sync with cursor tracking, `sql` command for raw queries, domain-specific `UpsertX()` and `SearchX()` methods.

**Workflow commands** (from archetype): `stale`, `orphans`, `load`, `channel-health`, `reconcile`, etc.

**Insight commands** (Rung 5): `health` (composite score), `similar` (duplicate detection), `trends`, `bottleneck`, `forecast`, `patterns`.

**REST + GraphQL**: OpenAPI specs generate full CLIs. GraphQL SDL files are parsed with Relay pagination detection and produce the same domain-specific output.

## The Steinberger Bar (v2 - Honest Scoring)

Named after Peter Steinberger. Two tiers, 100 points max, weighted 50/50. Grade A = 85+.

**Tier 1: Infrastructure** (50 points) - does the skeleton have the right patterns?

| Dimension | What It Checks |
|-----------|---------------|
| Output Modes | --json, --csv, --select, --quiet, --compact, auto-JSON when piped |
| Auth | OAuth flow, format-aware headers (Bot/Bearer/Basic from spec) |
| Error Handling | Typed exits, retry with backoff, actionable error messages |
| Agent-Native | --json, --select, --dry-run, --stdin, --no-input, --compact, --yes |
| + 5 more | Terminal UX, README, Doctor, Local Cache, Breadth |

**Tier 2: Domain Correctness** (50 points) - does the code actually work?

| Dimension | What It Checks |
|-----------|---------------|
| Path Validity | Generated paths exist in the OpenAPI spec |
| Auth Protocol | Auth format matches spec's securitySchemes |
| Data Pipeline | Sync calls domain-specific UpsertX(), not generic Upsert() |
| Sync Correctness | Real resources, nested paths, pagination, incremental cursors |
| Type Fidelity | String IDs (not int), required params marked, quality descriptions |
| Dead Code | No unwired flags, no uncalled functions, no ghost tables |

**Why two tiers?** The original scorecard tested syntax (does this string exist in the file?) not semantics (does this code actually work?). Generated CLIs scored Grade A and failed on the first real API call. The v2 scorecard catches that.

```bash
# Run the honest scorecard
printing-press scorecard --dir ./discord-cli --spec /tmp/discord-spec.json

# Run the mechanical dogfood validator
printing-press dogfood --dir ./discord-cli --spec /tmp/discord-spec.json
```

## Quick Start

```bash
git clone https://github.com/mvanhorn/cli-printing-press.git
cd cli-printing-press
go build -o ./printing-press ./cmd/printing-press
```

Then in Claude Code:

```bash
/printing-press Discord
/printing-press Stripe
/printing-press --spec ./openapi.yaml
```

## Verification Tools

Two Go commands for mechanical validation - no vibes, no self-assessment.

```bash
# Scorecard: two-tier scoring (infrastructure + domain correctness)
printing-press scorecard --dir ./my-cli --spec ./openapi.json

# Dogfood: catches dead flags, dead functions, auth mismatches, invalid paths
printing-press dogfood --dir ./my-cli --spec ./openapi.json
```

The dogfood command exists because the v1 scorecard tested syntax, not semantics. Generated CLIs scored Grade A and failed on the first real API call. The v2 scorecard + dogfood command make that impossible.

## Credits

- **Peter Steinberger** ([@steipete](https://github.com/steipete)) - [discrawl](https://github.com/steipete/discrawl) and [gogcli](https://github.com/steipete/gogcli) set the bar. The Steinberger quality scoring system is named after him.
- **Trevin Chow** ([@trevin](https://x.com/trevin)) - [7 Principles for Agent-Friendly CLIs](https://x.com/trevin) shaped the agent-first template design.
- **Ramp** ([@tryramp](https://github.com/ramp-public/ramp-cli)) - Their agent-first CLI inspired auto-JSON piping, --no-input, and --compact output.
- **Matt Van Horn** ([@mvanhorn](https://github.com/mvanhorn)) - Author of the printing press and [/last30days](https://github.com/mvanhorn/last30days-skill) recency research skill.

## License

MIT
