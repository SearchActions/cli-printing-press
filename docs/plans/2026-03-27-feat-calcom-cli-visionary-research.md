---
title: "Visionary Research: Cal.com CLI"
type: feat
status: active
date: 2026-03-27
phase: "0"
api: "cal.com"
---

# Visionary Research: Cal.com CLI

## Overview

Cal.com is open-source scheduling infrastructure ("the open-source Calendly alternative") with 34k+ GitHub stars. It serves businesses, freelancers, and developers who need scheduling automation. The API v2 has 181 paths / 285 operations covering bookings, event types, schedules, calendars, teams, organizations, webhooks, conferencing, slots, and more. The official @calcom/cli exists in the companion repo (8 stars, TypeScript, 46 command categories) but has near-zero adoption — it's an auto-generated API wrapper with no workflow intelligence. API v1 is deprecated and shuts down April 8, 2026 — making v2-native tooling timely.

The strategic opportunity: build a CLI that makes scheduling management instant from the terminal — with a local data layer that enables offline booking search, availability analysis, schedule health monitoring, and team scheduling insights that the web UI makes tedious. gcalcli (3k+ stars for Google Calendar) proves developer demand for calendar CLIs exists.

## API Identity

- **Domain:** Scheduling / Calendar Management
- **Primary users:** (1) Business owners managing their booking flow, (2) Engineering managers coordinating team availability, (3) Developers building scheduling integrations, (4) Ops/support staff handling booking changes
- **Core entities:** Bookings, Event Types, Schedules, Calendars, Teams, Organizations, Webhooks, Slots, Attendees, Conferencing
- **Base URL:** `https://api.cal.com`
- **Auth:** Bearer token (API key prefixed `cal_live_`), OAuth 2.0 with PKCE
- **Rate limit:** 120 req/min (API key), upgradeable to 200-800
- **Versioning:** `cal-api-version` header (date-based, latest: `2026-02-25`)
- **Pagination:** `take` + `skip` with envelope: `{ status, data, pagination: { totalItems, hasNextPage, ... } }`
- **v1 deprecated:** Shutting down April 8, 2026
- **Spec:** OpenAPI 3.0.0 at `docs/api-reference/v2/openapi.json` in repo

### Data Profile

| Dimension | Assessment |
|-----------|-----------|
| Write pattern | Mutable (bookings transition: upcoming -> past/cancelled/rescheduled; event types are editable) |
| Volume | Medium (hundreds to low thousands of bookings per account, tens of event types) |
| Real-time | Webhooks (outbound only — booking.created, booking.cancelled, etc.), NO WebSocket/SSE for clients |
| Search need | HIGH — users need to find bookings by attendee, date range, status; find available slots |

## Usage Patterns (Top 5 by Evidence)

| # | Pattern | Evidence Score | What It Needs |
|---|---------|---------------|---------------|
| 1 | **Booking management** (create, cancel, reschedule, list upcoming) | 8/10 | CRUD + status filtering + date ranges |
| 2 | **Calendar sync troubleshooting** (Google/Outlook integration issues) | 7/10 | Doctor command checking calendar connections + busy-time inspection |
| 3 | **Availability/slot checking** (when am I free? when can someone book me?) | 7/10 | Slot queries with time range + timezone + event type |
| 4 | **Schedule/availability configuration** (set working hours, block time, OOO) | 6/10 | Schedule CRUD + OOO management |
| 5 | **Team scheduling management** (who's available, team event types) | 6/10 | Team queries, member schedules, team event types |

**Evidence sources:**
- Cal.com GitHub issues (#14706: missing endpoints, #25560: workflow API, #3546/#8376: sync issues)
- n8n/Zapier Cal.com integrations (cross-platform evidence)
- Official companion CLI structure (46 command categories)
- Trustpilot reviews citing booking management and sync pain points
- Cal.com docs featuring availability as core product value prop

## Tool Landscape

### Forge/Platform CLIs
| Tool | Stars | Language | Type | Commands | Maintained |
|------|-------|----------|------|----------|------------|
| @calcom/cli (official, in calcom/companion) | 8 | TypeScript | API Wrapper | 46 categories | Yes (2026) |

### Workflow Overlays
None developer-native. n8n, Zapier, Make.com provide no-code automation but not CLI workflows.

### Alternative UX (Calendar Domain)
| Tool | Stars | Language | Relevance |
|------|-------|----------|-----------|
| gcalcli | 3k+ | Python | Google Calendar CLI — proves demand exists |
| calcure | 500+ | Python | TUI calendar/task manager — generic, not API-connected |
| calcurse | 1.1k+ | C | TUI calendar — local only |

### Data Tools
None. Zero tools sync Cal.com data locally for offline search/analysis.

### Integration Tools
- n8n Cal.com trigger (workflow automation platform)
- Zapier Cal.com integration (no-code)
- Make.com Cal.com app (visual workflow)
- cal-pr-agent (5 stars, PR automation for Cal.com dev)

**Key finding:** The Cal.com ecosystem has ZERO non-wrapper tools. No discrawl-equivalent. No offline search, no analytics dashboards, no backup tools. The entire data-tool and workflow-tool lanes are empty. gcalcli's 3k stars for Google Calendar proves developer appetite for scheduling CLIs.

## Workflows

### 1. Daily Agenda
- **Steps:** GET /v2/bookings?status=upcoming&afterStart=today&beforeEnd=tomorrow → group by time → show attendees
- **Frequency:** Daily
- **Pain point:** Must open web UI every morning to see today's meetings
- **Proposed:** `calcom-pp-cli agenda` — today's bookings with attendee info, times, event types

### 2. Availability Finder
- **Steps:** GET /v2/slots?username=X&eventSlug=Y&startTime=A&endTime=B → parse windows → format as time blocks
- **Frequency:** Multiple times daily when coordinating
- **Pain point:** Need to open Cal.com or share link just to see own availability
- **Proposed:** `calcom-pp-cli free --days 7` — show free windows for next N days

### 3. Schedule Health Check
- **Steps:** GET /v2/schedules → GET /v2/bookings?status=upcoming → cross-reference → check for conflicts
- **Frequency:** Weekly
- **Pain point:** No way to audit schedule health; sync issues are #1 complaint
- **Proposed:** `calcom-pp-cli health` — check calendar connections, find conflicts, report utilization

### 4. Booking Analytics
- **Steps:** Sync all bookings → aggregate by event type → compute show/cancel rates → identify trends
- **Frequency:** Weekly/Monthly
- **Pain point:** No built-in analytics; callytics web dashboard is the only option (0 stars)
- **Proposed:** `calcom-pp-cli stats --days 30` — show rate, cancellation rate, busiest days

### 5. Stale Event Type Cleanup
- **Steps:** GET /v2/event-types → GET /v2/bookings?eventTypeId=X for each → flag unused
- **Frequency:** Monthly
- **Pain point:** Event types accumulate, no easy way to identify unused ones
- **Proposed:** `calcom-pp-cli stale --days 90` — find event types with zero bookings in N days

## Architecture Decisions

| Decision Area | Need Level | Choice | Rationale |
|---------------|-----------|--------|-----------|
| **Persistence** | HIGH | SQLite with domain tables (bookings, event_types, schedules, attendees) | Bookings accumulate; local DB enables offline search, trend analysis, and analytics without API calls |
| **Real-time** | LOW | REST polling with `afterUpdatedAt` cursor | No WebSocket/SSE for clients; API supports date-range filtering for incremental sync |
| **Search** | HIGH | FTS5 on booking titles/descriptions, attendee names/emails, event type names | Users need to find bookings by attendee, keyword, date range |
| **Bulk** | MEDIUM | Paginate with `take=100` + `skip`, incremental sync | 120 req/min rate limit; sync fetches all pages incrementally with cursor |
| **Cache** | MEDIUM | SQLite IS the cache; read commands query local DB; `--sync` triggers refresh | Avoids rate limits; enables offline analytics |

## Top 5 Features for the World

| # | Feature | Evidence | Impact | Feasibility | Uniqueness | Composability | Data Fit | Maintain | Moat | **Total** |
|---|---------|----------|--------|-------------|------------|---------------|----------|----------|------|-----------|
| 1 | **Booking search** (FTS5 across all bookings) | 3 | 3 | 2 | 2 | 2 | 2 | 1 | 1 | **16/16** |
| 2 | **Agenda** (today's bookings, formatted) | 2 | 3 | 2 | 2 | 2 | 2 | 1 | 0 | **14/16** |
| 3 | **Schedule health** (gaps, conflicts, utilization) | 2 | 2 | 2 | 2 | 2 | 2 | 1 | 1 | **14/16** |
| 4 | **Free windows** (show availability) | 2 | 3 | 1 | 2 | 1 | 1 | 1 | 1 | **12/16** |
| 5 | **Stale finder** (unused event types) | 1 | 2 | 2 | 2 | 2 | 2 | 1 | 0 | **12/16** |

All 5 features score >= 12 — all are must-haves.

## Sources

- [Cal.com GitHub](https://github.com/calcom/cal.com) — 34k+ stars
- [Cal.com API v2 Docs](https://cal.com/docs/api-reference/v2/introduction) — 181 paths, 285 operations
- [OpenAPI Spec](https://github.com/calcom/cal.com/blob/main/docs/api-reference/v2/openapi.json) — 1.1MB JSON
- [Cal.com Companion/CLI](https://github.com/calcom/companion) — 8 stars, 46 command categories
- [API v1→v2 Migration](https://cal.com/docs/api-reference/v2/v1-v2-differences) — v1 shutdown April 8, 2026
- [GitHub Issue #14706](https://github.com/calcom/cal.com/issues/14706) — Missing API endpoints
- [GitHub Issue #25560](https://github.com/calcom/cal.com/issues/25560) — Workflow API request
- [GitHub Issues #3546, #8376](https://github.com/calcom/cal.com/issues/3546) — Calendar sync problems
- [n8n Cal.com Integration](https://n8n.io/integrations/cal-trigger/) — Workflow automation
- [gcalcli](https://github.com/insanum/gcalcli) — 3k+ stars, proves calendar CLI demand
- [calcure TUI](https://github.com/anufrievroman/calcure) — Terminal calendar alternative
