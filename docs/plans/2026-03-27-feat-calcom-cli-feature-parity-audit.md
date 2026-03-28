---
title: "Feature Parity Audit: Cal.com CLI"
type: feat
status: active
date: 2026-03-27
phase: "0.6"
api: "cal.com"
---

# Feature Parity Audit: Cal.com CLI

## Competitors Analyzed

1. **@calcom/cli** (8 stars, TypeScript) — Official Cal.com CLI in companion repo. 46 command categories. Pure API wrapper.
2. **gcalcli** (3k+ stars, Python) — Google Calendar CLI. Domain reference for calendar CLIs. 11 commands.

## Feature Matrix

### @calcom/cli (Official) — 46 Command Categories

| Feature | @calcom/cli | gcalcli | Ours | Classification |
|---------|-------------|---------|------|----------------|
| **Booking CRUD** (list, get, create, cancel, reschedule) | YES | N/A (different API) | YES | TABLE STAKES |
| **Event Type CRUD** (list, get, create, update, delete) | YES | N/A | YES | TABLE STAKES |
| **Schedule CRUD** (list, get, create, update, delete) | YES | N/A | YES | TABLE STAKES |
| **Calendar connections** (list, check, connect, disconnect) | YES | list calendars | YES | TABLE STAKES |
| **Webhooks CRUD** | YES | N/A | YES | TABLE STAKES |
| **Teams CRUD** | YES | N/A | YES | TABLE STAKES |
| **Team event types** | YES | N/A | YES | NICE-TO-HAVE |
| **Team memberships** | YES | N/A | YES | NICE-TO-HAVE |
| **Conferencing apps** (list, connect, default) | YES | N/A | YES | TABLE STAKES |
| **OAuth clients** | YES | N/A | NO | ANTI-SCOPE (platform feature, not user workflow) |
| **Managed users** | YES | N/A | NO | ANTI-SCOPE (platform feature) |
| **Organization-level operations** (131 endpoints) | YES | N/A | NO | ANTI-SCOPE (enterprise, org plan required) |
| **Routing forms** | YES | N/A | NO | NICE-TO-HAVE |
| **Verified resources** (email/phone verification) | YES | N/A | NO | ANTI-SCOPE (platform setup, not daily workflow) |
| **Stripe integration** | YES | N/A | NO | NICE-TO-HAVE |
| **Delegation credentials** | YES | N/A | NO | ANTI-SCOPE (enterprise only) |
| **Login/auth** (OAuth PKCE flow) | YES | OAuth setup | YES | TABLE STAKES |
| **Slots/availability query** | YES | N/A | YES | TABLE STAKES |
| **Destination calendars** | YES | N/A | YES | NICE-TO-HAVE |
| **Selected calendars** | YES | N/A | YES | NICE-TO-HAVE |
| **Booking attendees** | YES | N/A | YES | TABLE STAKES |
| **Booking confirm/decline** | YES | N/A | YES | TABLE STAKES |
| **Private links** | YES | N/A | NO | NICE-TO-HAVE |
| **Out-of-office** | YES (org only) | N/A | YES | NICE-TO-HAVE |
| **Me/profile** | YES | N/A | YES | TABLE STAKES |

### gcalcli (Domain Reference) — What Calendar CLI Users Expect

| Feature | gcalcli | @calcom/cli | Ours | Classification |
|---------|---------|-------------|------|----------------|
| **Agenda view** (today's/week's events) | YES (core feature) | NO (just list) | YES | TABLE STAKES |
| **Week calendar view** (calw) | YES | NO | NO | NICE-TO-HAVE |
| **Month calendar view** (calm) | YES | NO | NO | NICE-TO-HAVE |
| **Quick-add event** (natural language) | YES | NO | NO | NICE-TO-HAVE |
| **Event reminders** | YES | NO | NO | NICE-TO-HAVE |
| **ICS import** | YES | NO | NO | NICE-TO-HAVE |
| **Updates since datetime** | YES | NO | YES (sync) | TABLE STAKES |
| **Search events** | YES (text search) | NO | YES (FTS5) | TABLE STAKES |
| **Colored output** | YES | NO | YES | TABLE STAKES |
| **Configurable output format** | YES (--tsv, etc.) | JSON only | YES (--json, table) | TABLE STAKES |

## Table Stakes Summary

Features that MUST be in our CLI (from both competitors + domain expectations):

| # | Feature | Source | Priority |
|---|---------|--------|----------|
| 1 | Booking CRUD (list, get, create, cancel, reschedule) | @calcom/cli | P1 |
| 2 | Event Type CRUD (list, get, create, update, delete) | @calcom/cli | P1 |
| 3 | Schedule CRUD (list, get, create, update, delete) | @calcom/cli | P1 |
| 4 | Calendar connections (list, check) | @calcom/cli | P1 |
| 5 | Agenda view (today's bookings, formatted) | gcalcli | P1 |
| 6 | Search (full-text across bookings) | gcalcli | P1 |
| 7 | Slot/availability query | @calcom/cli | P1 |
| 8 | Conferencing apps (list, default) | @calcom/cli | P1 |
| 9 | Webhooks CRUD | @calcom/cli | P1 |
| 10 | Teams CRUD | @calcom/cli | P1 |
| 11 | Me/profile | @calcom/cli | P1 |
| 12 | Booking confirm/decline | @calcom/cli | P1 |
| 13 | Colored/formatted output | gcalcli | P1 |
| 14 | --json output mode | Standard | P1 |
| 15 | Attendee management | @calcom/cli | P1 |

## Anti-Scope Items (with cost analysis)

| Feature | Why Anti-Scope | % Users Need | Competitor has it? |
|---------|---------------|-------------|-------------------|
| OAuth client management | Platform developer feature, not scheduling workflow | <5% | @calcom/cli (platform tier) |
| Managed users | Platform feature for multi-tenant apps | <5% | @calcom/cli (platform tier) |
| Organization-level ops (131 endpoints) | Requires org plan, enterprise-only | <10% | @calcom/cli |
| Verified resources | One-time setup, not daily workflow | <5% | @calcom/cli |
| Delegation credentials | Enterprise SSO feature | <2% | @calcom/cli |
| Routing forms | Niche feature, web UI is better for form building | <10% | @calcom/cli |

None of these anti-scope items have demand from users with >100 stars. The @calcom/cli has 8 stars total, so even its full feature set has no proven user demand. Our anti-scope decisions are safe.

## Key Insight

@calcom/cli has 46 command categories but is a pure API wrapper — no agenda, no search, no analytics, no health checks. gcalcli proves that calendar CLI users expect **agenda views and search**, not just CRUD. Our differentiator is combining @calcom/cli's API coverage with gcalcli's workflow intelligence, plus a local data layer that neither has.
