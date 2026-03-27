---
title: "Audit: Linear CLI"
type: fix
status: active
date: 2026-03-25
---

# Audit: Linear CLI

## Automated Scorecard Baseline

| Dimension | Score | Gap |
|-----------|-------|-----|
| Output Modes | 8/10 | |
| Auth | 6/10 | No OAuth, no --stdin on login |
| Error Handling | 4/10 | Missing "did you mean?", no help link |
| Terminal UX | 8/10 | |
| README | 0/10 | No README at all |
| Doctor | 8/10 | |
| Agent Native | 8/10 | |
| Local Cache | 3/10 | No file cache |
| Breadth | 4/10 | Missing documents, comments list, webhooks |
| Vision | 0/10 | Scorecard doesn't detect workflow commands |
| Workflows | 0/10 | Scorecard doesn't detect workflow commands |

**Baseline: 49/110 (44%) - Grade D**

## GOAT Improvement Plan

### Priority 1: README (0 -> 8+)
Write full README with install, quickstart, commands, cookbook, agent usage, output formats.

### Priority 2: Breadth (4 -> 7+)
Add documents command group, add comments list, add shell completions docs.

### Priority 3: Error Handling (4 -> 7+)
Add --help link in errors, improve error messages with suggestions.

### Priority 4: Auth (6 -> 8+)
Login --stdin flag already exists. Add env var docs.

### Priority 5: Vision/Workflows (0 -> 6+)
Add export command (JSONL), add vision-type commands the scorecard recognizes.

### Priority 6: Local Cache (3 -> 5+)
Add file cache for GET responses.
