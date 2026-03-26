---
title: "docs: Add agent-native thesis to README with research-backed arguments"
type: docs
status: completed
date: 2026-03-26
---

# docs: Add agent-native thesis to README with research-backed arguments

## Overview

The README explains WHAT the printing press does but not WHO it's for or WHY CLIs matter. The thesis is: the printing press is an **agent infrastructure factory** - it prints the CLIs that power users' agents need. Every API that gets a CLI becomes instantly accessible to every agent framework. The human is in the loop as the architect; the agent is the operator; the CLI is the interface between them.

## Research Findings

### The Token Economics Argument

From [MCP vs CLI for AI Agents](https://manveerc.substack.com/p/mcp-vs-cli-ai-agents) (2026):
- A single GitHub MCP server exposes **93 tools** costing **~55,000 tokens** just to load tool definitions
- At scale (10,000 sessions/day): **$1,600/day** spent on tool definitions alone
- CLI alternative: `gh issue create --help` costs **~200 tokens**. Full execution cycle: **<500 tokens**
- Concrete example: summing totals across 150 order IDs - MCP approach: ~50,000 tokens; CLI approach: ~500 tokens (**1% of MCP cost**)
- One practitioner cut token count to **60%** by reformatting JSON responses as plain text

### The Training Data Argument

From the same source and [Firecrawl's analysis](https://www.firecrawl.dev/blog/why-clis-are-better-for-agents):
- LLMs trained on **enormous volumes of shell interactions** - Unix pipe chains are deeply embedded in model weights
- MCP composition patterns have **zero training data** and **zero production hardening**
- Unix tool output formats described as **"information-theoretically optimal for LLM reasoning"**
- IDE agents send entire conversation history + all open files each interaction - **scales poorly**
- CLI agents practice **"progressive disclosure"** - loading only necessary context

### The Delegation Model Argument

From [Why CLIs Are Better for AI Agents](https://www.firecrawl.dev/blog/why-clis-are-better-for-agents):
- "IDE agents are designed for **suggestion**. CLI agents are designed for **delegation**."
- Terminal agents "run for hours without supervision, coordinate changes across dozens of files, execute shell commands to verify their work"
- CLI operations provide **deterministic feedback** - exit code 1 means failure, agent can self-heal
- "You can invoke terminal AI agents from scripts but you cannot do the same with a VS Code sidebar"

### The Composability Argument

From [CLI-Anything](https://github.com/HKUDS/CLI-Anything) (HKU, 1,839+ passing tests):
- CLI is "the universal interface for both humans and AI agents"
- Text commands match LLM output format and enable workflow chaining
- JSON output eliminates parsing complexity
- "UI automation breaks constantly" vs CLI deterministic reliability
- Claude Code executes "thousands of real workflows through CLI daily"

### The Agentic Engineering Era

From [Andrej Karpathy](https://www.nextbigfuture.com/2026/03/andrej-karpathy-on-code-agents-autoresearch-and-the-self-improvement-loopy-era-of-ai.html) and [The New Stack](https://thenewstack.io/ai-coding-tools-in-2025-welcome-to-the-agentic-cli-era/):
- 2026 is the era of **agentic engineering** - humans don't write most code, they direct agents
- Karpathy: LLMs are a new "Operating System" - we're in the "1960s of OS design"
- CLI coding agents are "no longer experimental products but standard productivity tools"

## Proposed Changes

### 1. Rewrite the opening section

The current opening ("Give it an API name...") is good but needs the agent thesis woven in. The new opening should answer: "Why does every API need a CLI?"

### 2. Replace "Why CLIs Matter Now" section

The current section (lines 13-19 of the old README, removed in last update) had the right idea but was too brief and had no evidence. Bring it back stronger with research-backed arguments:

**Structure:**
1. The token economics (numbers from MCP vs CLI research)
2. The training data advantage (models know Unix, don't know MCP)
3. The delegation model (agents run CLIs autonomously, can't run GUIs)
4. The composability (pipes, jq, scripting)
5. The conclusion: every API that gets a CLI becomes agent-accessible

### 3. Add "The Human + Agent Model"

New section explaining the power user -> agent -> CLI relationship:
- Power user is the architect (decides what to build, sets constraints)
- Agent is the operator (executes tasks, chains commands, self-heals)
- CLI is the interface (structured, deterministic, composable, token-efficient)
- The printing press manufactures the interface layer

### 4. Update GitHub About description

Current: "Give it an API name. Get back a CLI that sees what the API's own creators missed."
New: Include the agent angle.

## Acceptance Criteria

- [ ] Opening section establishes the agent-native thesis
- [ ] "Why CLIs Matter" section includes 3+ specific statistics with sources
- [ ] Token economics comparison (MCP ~55k tokens vs CLI ~200 tokens) cited
- [ ] Training data argument included (models know Unix pipes)
- [ ] Human + Agent + CLI relationship explained
- [ ] No made-up statistics - everything sourced from research
- [ ] Existing NOI and Creativity Ladder sections preserved

## Files to Modify

- `README.md` - Add/rewrite sections

## Sources

- [MCP vs CLI for AI Agents (Manveer, 2026)](https://manveerc.substack.com/p/mcp-vs-cli-ai-agents) - Token economics, 55k vs 200 token comparison
- [Why CLIs Are Better for AI Agents (Firecrawl, 2026)](https://www.firecrawl.dev/blog/why-clis-are-better-for-agents) - Delegation model, progressive disclosure
- [CLI-Anything (HKU, 2026)](https://github.com/HKUDS/CLI-Anything) - Universal interface thesis, 1839+ tests
- [Agentic CLI Era (The New Stack, 2025)](https://thenewstack.io/ai-coding-tools-in-2025-welcome-to-the-agentic-cli-era/) - Standard productivity tools
- [Karpathy on Code Agents (2026)](https://www.nextbigfuture.com/2026/03/andrej-karpathy-on-code-agents-autoresearch-and-the-self-improvement-loopy-era-of-ai.html) - OS analogy
- [CLI-Based Agents vs MCP (2026)](https://lalatenduswain.medium.com/cli-based-agents-vs-mcp-the-2026-showdown-that-every-ai-engineer-needs-to-understand-7dfbc9e3e1f9) - Showdown analysis
