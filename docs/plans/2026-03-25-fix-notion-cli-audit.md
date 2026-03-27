---
title: "Audit: Notion CLI"
type: fix
status: active
date: 2026-03-25
---

# Audit: Notion CLI

## Automated Steinberger Scorecard (Baseline)

| Dimension | Score |
|-----------|-------|
| Output modes | 10/10 |
| Auth | 10/10 |
| Error handling | 10/10 |
| Terminal UX | 10/10 |
| README | 10/10 |
| Doctor | 10/10 |
| Agent Native | 10/10 |
| Local Cache | 10/10 |
| Breadth | 7/10 |

**Baseline Total: 87/90 (Grade A)**

## Fix Plan
1. Breadth: raise from 7/10 to 10/10
2. Command names: delete-a/retrieve-a/create-a -> clean names
3. Root description: developer-friendly
4. Examples: realistic Notion UUIDs
5. Complex body --stdin examples for top 3 endpoints
