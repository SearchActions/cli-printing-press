---
title: "Research: Notion CLI"
type: feat
status: active
date: 2026-03-25
---

# Research: Notion CLI

## Spec Discovery
- Official OpenAPI spec: https://raw.githubusercontent.com/makenotion/notion-mcp-server/main/scripts/notion-openapi.json
- Source: Notion's official MCP server repo (makenotion/notion-mcp-server)
- Format: OpenAPI 3.1.0
- API Version: 2.0.0 (2025-09-03 "Data Source Edition")
- Endpoint count: 22 operations across 16 paths
- Note: Breaking change - Database endpoints replaced with data source endpoints

## Competitors (Deep Analysis)

### 4ier/notion-cli (87 stars)
- Repo: https://github.com/4ier/notion-cli
- Language: Go (99.9%)
- Commands: 39
- Last commit: 2026-02-24 (v0.3.0)
- Open issues: 1 ("notion block not saved as markdown")
- Maintained: Yes (actively)
- Notable features:
  - Human-readable filtering without JSON complexity
  - Schema-aware property handling (auto-detects data types)
  - Adaptive output formatting (tables interactive, JSON piped)
  - Markdown bidirectional support
  - Recursive block traversal with configurable depth
  - URL and ID flexibility
  - Homebrew, npm, scoop, Docker install
  - 39 commands covering pages, databases, blocks, comments, users, files
  - Raw API access (GET/POST/PATCH/DELETE with direct path)
- Weaknesses:
  - No --dry-run support
  - No --select for field filtering
  - No doctor command for config validation
  - No local caching
  - Uses older database API, not new data_sources API

### lox/notion-cli (15 stars)
- Repo: https://github.com/lox/notion-cli
- Language: Go (100%)
- Commands: ~15
- Last commit: 2026-03-24 (v0.5.0)
- Open issues: 2
- Maintained: Yes (very active, 4 contributors)
- Notable features:
  - MCP-based architecture
  - OAuth browser flow
  - Markdown file sync (bidirectional)
  - Semantic search
  - File attachments
  - Integration with Claude Code and Amp
- Weaknesses:
  - Fewer commands than 4ier
  - MCP dependency adds complexity
  - No --json/--select/--dry-run agent-native features
  - Smaller user base

## User Pain Points
> "the MCP was a little bit short of features for my openclaw usage so I just open sourced a CLI oriented to Agent usage" - lox (HN, 2026-02-03)
> "Have you considered integrating it with tools like fzf for fuzzy searching within Notion workspaces?" - HN user (2026-03)
> "notion block not saved as markdown" - 4ier/notion-cli issue #14

## Auth Method
- Type: Bearer token (Integration token or OAuth)
- Env var convention: NOTION_TOKEN (4ier), NOTION_ACCESS_TOKEN (lox)
- Our convention: NOTION_TOKEN

## Demand Signals
- HN Show HN with 87 stars for 4ier/notion-cli (2026-03)
- HN Show HN for lox/notion-cli agent-focused CLI (2026-02)
- Multiple Go implementations show demand for terminal-native Notion access
- Agent/AI workflow integration is the emerging demand driver

## Strategic Justification
**Why this CLI should exist:** 4ier/notion-cli has 39 commands and 87 stars but lacks agent-native features (--json, --select, --dry-run, --stdin, --yes, --no-cache, doctor). It also uses the older database API rather than the new data_sources API (2025-09-03). Our CLI will be generated from the official Notion MCP server spec (the most up-to-date source), include full agent-native features out of the box, and support the new data_sources endpoints that no competitor has yet adopted. The --stdin support for complex body fields (page creation with nested properties) is a specific gap no CLI addresses well.

## Target
- Command count: 22+ (cover every API operation, match spec breadth)
- Key differentiator: Agent-native features + new data_sources API + --stdin for complex properties
- Quality bar: Steinberger Grade A (72+/90)
