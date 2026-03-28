---
title: "Power User Workflows: Cal.com CLI"
type: feat
status: active
date: 2026-03-27
phase: "0.5"
api: "cal.com"
---

# Power User Workflows: Cal.com CLI

## API Archetype: Scheduling Platform (Hybrid)

Combines elements of:
- **Project Management** — Event types = task types, bookings = tasks with states, teams with memberships
- **CRM** — Attendees are contacts, booking pipeline tracks conversion from slot → booking → completion
- **Infrastructure** — Calendar connections are integrations that break, conferencing apps need health checks

## All 12 Workflow Ideas

### 1. `agenda` — Today's Bookings Briefing
- **What:** Show today's/this week's bookings with attendee info, times, event types, conferencing links
- **API calls:** GET /v2/bookings?status=upcoming&afterStart=today&beforeEnd=tomorrow&sortStart=asc
- **Frequency:** Daily (3) | **Pain:** High (3) — must open web UI | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 12/12**

### 2. `search` — Full-Text Booking Search
- **What:** FTS5 search across all synced bookings by attendee name/email, title, description
- **API calls:** Sync bookings to SQLite → FTS5 MATCH query (no live API needed)
- **Frequency:** Daily (3) | **Pain:** High (3) — no offline search exists | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 12/12**

### 3. `stats` — Booking Analytics
- **What:** Aggregate bookings: total, cancellation rate, show rate, most popular event types, busiest days/hours
- **API calls:** Local DB query on synced bookings → aggregate by status/event type
- **Frequency:** Weekly (2) | **Pain:** High (3) — no built-in analytics | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 11/12**

### 4. `doctor` — Calendar + API Health Check
- **What:** Check all calendar connections, conferencing apps, API auth, schedule validity
- **API calls:** GET /v2/me → GET /v2/calendars/connections → GET /v2/calendars/{cal}/check → GET /v2/conferencing → GET /v2/schedules
- **Frequency:** On-demand (2) | **Pain:** High (3) — sync issues are #1 complaint | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 11/12**

### 5. `free` — Available Time Slots
- **What:** Show when you're available for a specific event type over next N days
- **API calls:** GET /v2/slots?username=X&eventTypeSlug=Y&start=A&end=B&timeZone=Z
- **Frequency:** Daily (3) | **Pain:** High (3) — need to open app | **Feasibility:** Medium (2) — needs slot parsing | **Uniqueness:** 3
- **Total: 11/12**

### 6. `conflicts` — Double-Booking Detection
- **What:** Find overlapping bookings across calendars, detect scheduling conflicts
- **API calls:** GET /v2/calendars/busy-times + sync bookings → time range overlap detection in SQLite
- **Frequency:** Weekly (2) | **Pain:** High (3) — silent double-books | **Feasibility:** Medium (2) | **Uniqueness:** 3
- **Total: 10/12**

### 7. `stale` — Unused Event Type Finder
- **What:** Find event types with zero bookings in last N days, candidates for cleanup
- **API calls:** Event types + bookings in local DB → cross-reference on eventTypeId
- **Frequency:** Monthly (1) | **Pain:** Medium (2) | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 9/12**

### 8. `unconfirmed` — Pending Booking Triage
- **What:** List unconfirmed bookings, allow bulk confirm/decline
- **API calls:** GET /v2/bookings?status=unconfirmed → POST confirm/decline
- **Frequency:** Daily (3) | **Pain:** Medium (2) | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 11/12**

### 9. `clone` — Event Type Cloning
- **What:** Clone event type to another team or as personal copy
- **API calls:** GET /v2/event-types/{id} → POST /v2/event-types or /v2/teams/{id}/event-types
- **Frequency:** Monthly (1) | **Pain:** High (3) | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 10/12**

### 10. `noshow` — No-Show Detection
- **What:** Find past bookings not marked completed/absent, track no-show rate
- **API calls:** Local DB → filter past unmarked → POST mark-absent
- **Frequency:** Weekly (2) | **Pain:** Medium (2) | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 10/12**

### 11. `export` — Booking Data Export
- **What:** Export bookings to CSV/JSON for external analytics
- **API calls:** Local DB → SELECT → format
- **Frequency:** Monthly (1) | **Pain:** Medium (2) | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 9/12**

### 12. `reassign` — Bulk Booking Reassignment
- **What:** Move upcoming bookings from one host to another
- **API calls:** GET bookings → filter by host → POST /v2/bookings/{uid}/reassign/{userId}
- **Frequency:** Monthly (1) | **Pain:** High (3) | **Feasibility:** Easy (3) | **Uniqueness:** 3
- **Total: 10/12**

## Validation Against API Capabilities

All 12 workflows validated against OpenAPI spec. Every required endpoint exists with the needed query parameters. Key validations:
- `afterStart`, `beforeEnd`, `status`, `attendeeEmail`, `eventTypeId` filters on GET /v2/bookings ✓
- GET /v2/slots supports username, eventTypeSlug, start, end, timeZone ✓
- GET /v2/calendars/{cal}/check endpoint exists ✓
- POST /v2/bookings/{uid}/confirm, /decline, /mark-absent, /reassign/{userId} all exist ✓
- `afterUpdatedAt` supports incremental sync ✓

## Top 7 for Implementation (Phase 4 Priority 2)

| Priority | Name | Score | What It Does |
|----------|------|-------|-------------|
| 1 | **agenda** | 12 | Today's bookings with attendee info and times |
| 2 | **search** | 12 | FTS5 full-text search across all bookings |
| 3 | **stats** | 11 | Booking analytics: show rate, cancellation rate, busiest times |
| 4 | **doctor** | 11 | Calendar connections, conferencing, API health check |
| 5 | **free** | 11 | Available time slots for next N days |
| 6 | **unconfirmed** | 11 | Pending bookings triage with bulk confirm/decline |
| 7 | **conflicts** | 10 | Detect double-bookings and overlapping meetings |

## Naming Pass

| Command | Completes "I need to check ___" | Max 9 chars | No hyphens | No collision |
|---------|--------------------------------|-------------|------------|-------------|
| agenda | "my agenda" | 6 ✓ | ✓ | ✓ |
| search | "for a meeting" | 6 ✓ | ✓ | ✓ |
| stats | "my booking stats" | 5 ✓ | ✓ | ✓ |
| doctor | "my calendar health" | 6 ✓ | ✓ | Standard |
| free | "when I'm free" | 4 ✓ | ✓ | ✓ |
| unconfirmed | "unconfirmed bookings" | 11 — over limit but clear | ✓ | ✓ |
| conflicts | "for conflicts" | 9 ✓ | ✓ | ✓ |
