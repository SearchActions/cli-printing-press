---
title: "Research: Cal.com CLI"
type: feat
status: active
date: 2026-03-27
phase: "1"
api: "cal.com"
---

# Research: Cal.com CLI

## Spec Discovery
- **Official OpenAPI spec:** https://github.com/calcom/cal.com/blob/main/docs/api-reference/v2/openapi.json
- **Source:** GitHub repo, confirmed via discussion #17565
- **Format:** OpenAPI 3.0.0 JSON (1.1MB)
- **Total paths:** 181 (103 user-level, 78 org-level)
- **Total operations:** 285 (154 user-level)
- **Resource categories:** bookings, event-types, schedules, calendars, conferencing, webhooks, teams, slots, me, oauth-clients, verified-resources, routing-forms, stripe, selected-calendars, destination-calendars, api-keys
- **API version header:** `cal-api-version: 2026-02-25` (date-based, required)

## Competitors (Deep Analysis)

### @calcom/cli (8 stars)
- **Repo:** https://github.com/calcom/companion (packages/cli/)
- **Language:** TypeScript (98.6%)
- **Commands:** 46 command categories (auto-generated from OpenAPI)
- **Last commit:** February 2026
- **Maintained:** Yes (active development)
- **Notable features:** OAuth PKCE login, full API v2 coverage, openapi-ts code generation
- **Weaknesses:** Pure API wrapper. No agenda view, no search, no analytics, no health checks. 8 stars despite Cal.com's 34k. Requires Node runtime.

### gcalcli (3k+ stars) — Domain Reference
- **Repo:** https://github.com/insanum/gcalcli
- **Language:** Python
- **Commands:** 11 (agenda, calw, calm, list, edit, add, quick, import, remind, updates, init)
- **Maintained:** Yes
- **Notable features:** Agenda view, week/month calendars, colored output, reminders, search
- **Relevance:** Proves demand for calendar CLIs

## User Pain Points
> "Some API endpoints don't work properly... resorting to tRPC calls" — Issue #14706
> "Workflows can only be created/edited through the UI" — Issue #25560
> Calendar sync issues (Google #8376, Outlook #3546) are the #1 complaint

## Auth Method
- **Type:** Bearer token (API key prefix: `cal_live_`) or OAuth 2.0 PKCE
- **Env var:** `CAL_COM_API_KEY`
- **Rate limit:** 120 req/min

## Strategic Justification
@calcom/cli has 8 stars vs Cal.com's 34k — that's a vacuum, not competition. It's a pure API wrapper with no workflow intelligence. gcalcli (3k stars) proves calendar CLI demand. Our differentiators: SQLite data layer with FTS5, booking analytics, agenda view, Go single binary, agent-native flags.

## Target
- **Commands:** 40+ (15 CRUD + 7 workflows + 7 data layer + 11 supporting)
- **Differentiator:** Local SQLite + FTS5 search + booking analytics + agenda
- **Quality bar:** Grade A (80+/100)
