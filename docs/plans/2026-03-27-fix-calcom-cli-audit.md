---
title: "Non-Obvious Insight Review: Cal.com CLI"
type: fix
status: active
date: 2026-03-27
phase: "3"
api: "cal.com"
---

# Non-Obvious Insight Review: Cal.com CLI

## Automated Scorecard Baseline: 55/100 (Grade C)

| Dimension | Score | Gap | Fix |
|-----------|-------|-----|-----|
| Output Modes | 10/10 | — | Good |
| Auth | 8/10 | Missing cal-api-version header | Add header to client.go |
| Error Handling | 10/10 | — | Good |
| Terminal UX | 9/10 | No color | Add color later |
| README | 5/10 | No cookbook, no comparison | Add in Phase 4 P7 |
| Doctor | 10/10 | — | Good |
| Agent Native | 8/10 | — | Good |
| Local Cache | 10/10 | — | Good |
| Breadth | 8/10 | — | Good (222 commands) |
| Vision | 8/10 | Generic store | Replace with domain tables |
| Workflows | 2/10 | No domain workflows | Add 7 workflow commands |
| Insight | 0/10 | No insight commands | Add stats, stale, conflicts |
| Sync Correctness | 0/10 | Generic sync | Rewrite with afterUpdatedAt |
| Dead Code | 0/5 | Dead code present | Audit in Phase 4.6 |

## Critical Issues
1. Missing `cal-api-version` header — API calls will 400
2. Generic store — JSON blob, not domain columns
3. No workflow commands — the product is missing
4. Wrong binary name — cal-com-user-cli, not calcom-pp-cli
5. Complex body fields skipped — booking create needs --stdin

## GOAT Plan (Phase 4 priorities)
- P0: Domain data layer (bookings, attendees, event_types tables + FTS5 + sync)
- P1: Table stakes (cal-api-version header, CRUD verification)
- P2: 7 workflow commands (agenda, search, stats, free, unconfirmed, conflicts, stale)
- P3: Name normalization (calcom-pp-cli) + command name cleanup
- P4: Scorecard fixes + dead code audit
- P5: Tests
- P6-7: Distribution + README polish
