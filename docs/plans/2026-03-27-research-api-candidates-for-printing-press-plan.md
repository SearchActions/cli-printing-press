---
title: "API Candidates for CLI Printing Press: Scored Rankings"
type: research
status: active
date: 2026-03-27
---

# API Candidates for CLI Printing Press

20 APIs researched across 3 categories. Scored on 5 dimensions, ranked by composite score.

## CORRECTION LOG (2026-03-27, second pass)

The first research pass missed existing CLIs because agents searched GitHub but not npm. Second pass verified every top candidate against npm + GitHub. Key corrections:

| API | First Pass Said | Reality | Impact |
|-----|----------------|---------|--------|
| **PostHog** | "ZERO CLI" | `@posthog/cli` exists - login, query, sourcemap, endpoints-as-code | Drops from #1. CLI gap is narrower than claimed. |
| **Linear** | "No CLI at all" | `linearis` (npm, published 2 days ago), `@schpet/linear-cli`, `@anoncam/linear-cli`, Rust CLI | Drops significantly. Multiple community CLIs exist. |
| **Cal.com** | "NO CLI" | `@calcom/cal-mcp` (MCP server, all API endpoints) | Drops. MCP covers the API surface. |
| **Discord** | "No admin CLI" | Confirmed - no admin/API CLI. Only bot frameworks (discord.js). | Holds - still true. |
| **Notion** | "No official CLI" | Community CLIs exist (notion-cli by MrRichRobinson, bash CLI by nitaiaharoni1, @coastal-programs/notion-cli) | Slight drop but still a gap - no dominant CLI. |
| **HubSpot** | "CLI is CMS-only" | Confirmed - `@hubspot/cli` v8.2.0 (4 days ago) is CMS dev tooling. Zero CRM data ops. | Holds. |
| **PagerDuty** | "95 stars community" | Confirmed - `pagerduty-cli` npm, community-maintained, not official. | Holds. |
| **Plaid** | "57 star community" | Confirmed - `plaid-cli` on npm (6 years old), `landakram/plaid-cli` on GitHub. | Holds. |

## Scoring Dimensions

| Dimension | What it measures | 1 (worst) | 10 (best) |
|-----------|-----------------|-----------|-----------|
| **Ease** | How easy to generate (spec quality, REST vs GraphQL, endpoint count) | No spec, GraphQL-only, 1000+ endpoints | Clean official OpenAPI, <200 endpoints |
| **Impact** | How much it matters to the world (breadth of use, pain solved) | Niche tool, mild convenience | Millions of users, solves real daily pain |
| **Popularity** | Developer reach (npm downloads, stars, community size) | <100k npm/wk, <5k stars | >1M npm/wk, >30k stars |
| **CLI Gap** | How underserved the CLI space is (existing tools vs need) | Official CLI covers everything | No CLI exists, massive demand |
| **SQLite Fit** | How much local persistence + offline search adds value | Ephemeral data, no search need | Accumulating data, high search/analytics need |

## The Rankings

| # | API | Ease | Impact | Pop | Gap | SQLite | **Total** | Category | CLI Exists? |
|---|-----|------|--------|-----|-----|--------|-----------|----------|-------------|
| 1 | **HubSpot** | 8 | 8 | 8 | 9 | 10 | **43** | CRM | CMS-only (zero CRM data ops) |
| 2 | **Discord** | 9 | 7 | 7 | 10 | 9 | **42** | Communication | No admin CLI (only bot libs) |
| 3 | **Plaid** | 9 | 9 | 7 | 9 | 10 | **44** | Financial Data | 57-star community, 6yr old |
| 4 | **PagerDuty** | 8 | 8 | 5 | 9 | 9 | **39** | Incident Mgmt | 95-star community, unsupported |
| 5 | **Notion** | 5 | 8 | 8 | 7 | 10 | **38** | Productivity | Fragmented community CLIs |
| 6 | **Stripe** | 10 | 9 | 10 | 6 | 8 | **43** | Payments | Official (event-focused, no data) |
| 7 | **PostHog** | 8 | 9 | 8 | 6 | 9 | **40** | Analytics | Official (query+sourcemap+endpoints) |
| 8 | **Cloudflare** | 7 | 8 | 9 | 7 | 7 | **38** | Infrastructure | Wrangler (Workers only, 20% coverage) |
| 9 | **Slack** | 3 | 7 | 9 | 7 | 9 | **35** | Communication | Official (app-dev only, not data) |
| 10 | **Jira** | 8 | 7 | 6 | 5 | 8 | **34** | Project Mgmt | jira-cli 5.4k stars (no data layer) |
| 11 | **Dub.co** | 7 | 5 | 7 | 10 | 5 | **34** | Link Mgmt | No CLI |
| 12 | **Datadog** | 6 | 7 | 8 | 5 | 7 | **33** | Observability | pup 538 stars (AI-agent focused) |
| 13 | **Linear** | 4 | 8 | 7 | 5 | 10 | **34** | Project Mgmt | Multiple community CLIs (linearis, etc) |
| 14 | **Twilio** | 8 | 6 | 10 | 3 | 5 | **32** | Communications | Official CLI (188 stars, maintained) |
| 15 | **Shopify** | 3 | 7 | 7 | 6 | 8 | **31** | E-commerce | Official (dev-only), GraphQL-only now |
| 16 | **Cal.com** | 7 | 7 | 8 | 5 | 6 | **33** | Scheduling | MCP server covers API surface |
| 17 | **Vercel** | 7 | 5 | 8 | 4 | 5 | **29** | Deployment | Official (15k stars, comprehensive) |
| 18 | **Supabase** | 4 | 6 | 10 | 3 | 4 | **27** | BaaS | Official (1.8k stars, active) |
| 19 | **Neon** | 7 | 4 | 7 | 4 | 4 | **26** | Postgres | Official (105 stars, adequate) |
| 20 | **Turso** | 4 | 4 | 6 | 6 | 4 | **24** | Edge DB | Official (290 stars, 111 issues) |

## Tier 1: Build These First (Score 38+, verified CLI gap)

### 1. HubSpot (43/50) - BIGGEST ENTERPRISE OPPORTUNITY
- **Why #1:** 935k npm/wk, official OpenAPI spec, CLI only covers CMS (zero CRM data ops)
- **Spec:** Official OpenAPI (36 stars), maintained by HubSpot
- **CLI gap:** VERIFIED. `@hubspot/cli` v8.2.0 (published 4 days ago) handles themes, modules, serverless functions. Zero commands for contacts, deals, companies, tickets, or pipeline analytics.
- **Killer commands:** `hs deals --closing-this-month --min=50000`, `hs contacts --no-activity=30d`, `hs pipeline --conversion`
- **SQLite value:** EXTREME. CRM data is the canonical local-sync use case. Contacts, deals, companies, tickets.
- **HN angle:** "I built a HubSpot CLI that lets you SQL query your CRM data offline"
- **Risk:** Enterprise audience may prefer web UI. But 935k npm downloads/wk = massive developer surface.

### 2. Discord (42/50) - BEST SPEC + WIDEST GAP
- **Why #2:** Official OpenAPI 3.1 spec (303 stars), literally no admin/API CLI exists
- **Spec:** Official, auto-generated daily, both stable and preview versions
- **CLI gap:** VERIFIED. Only bot frameworks (discord.js 26.6k stars). Zero CLI for server admin, channel management, message search, audit logs.
- **Killer commands:** `discord search --channel support --contains "crash"`, `discord members --role admin`, `discord audit --last-24h`
- **SQLite value:** Very high. Messages, channels, members, roles, audit logs. Offline message search is Discord's #1 user complaint.
- **HN angle:** "I built a Discord CLI with offline message search (the feature Discord won't build)"
- **Risk:** Bot token required. Rate limits are aggressive.

### 3. Plaid (44/50) - BEST PRODUCT-MARKET FIT
- **Why top tier:** Personal finance + local SQLite = privacy-first architecture that users WANT
- **Spec:** Official OpenAPI (111 stars), all SDKs generated from it
- **CLI gap:** Massive. One 57-star community CLI, one 17-star Go+SQLite project (validates concept)
- **Killer commands:** `plaid sync && plaid spend --vs-last-month`, `plaid balances`, `plaid export --year=2025`
- **SQLite value:** EXTREME. Transactions, balances, investments - all private financial data that should live locally
- **HN angle:** "I built a personal finance CLI that syncs your bank transactions to SQLite"
- **Risk:** Plaid API requires paid access for production. Sandbox testing is free.

### 3. HubSpot (43/50) - BIGGEST ENTERPRISE OPPORTUNITY
- **Why top tier:** 935k npm downloads/wk, official OpenAPI spec, CLI only covers CMS (not CRM)
- **Spec:** Official OpenAPI (36 stars), maintained by HubSpot
- **CLI gap:** Official CLI (183 stars) is CMS-only. Zero CRM data operations from terminal.
- **Killer commands:** `hs deals --closing-this-month --min=50000`, `hs contacts --no-activity=30d`, `hs pipeline --conversion`
- **SQLite value:** EXTREME. Contacts, deals, companies, tickets - CRM data is the canonical local-sync use case
- **HN angle:** "I built a HubSpot CLI that lets you SQL query your CRM data offline"
- **Risk:** Enterprise audience may prefer web UI. But sales ops teams are technical enough.

### 4. Stripe (43/50) - EASIEST TO BUILD
- **Why top tier:** Perfect 10/10 spec quality, 35M npm downloads/month, massive reach
- **Spec:** Official OpenAPI (467 stars), best-maintained spec in the industry
- **CLI gap:** Official CLI (1,947 stars) is webhook/event-focused, not data-querying
- **Killer commands:** `stripe charges --failed --last-7d`, `stripe subscriptions --churned`, `stripe revenue --trend`
- **SQLite value:** High. Charges, subscriptions, invoices, disputes - financial query workload
- **HN angle:** "I built a Stripe CLI that syncs your payment data to SQLite for offline analytics"
- **Risk:** Official CLI is good enough for many users. Need clear differentiation.

### 5. Discord (42/50) - BEST SPEC READINESS
- **Why top tier:** Official OpenAPI 3.1 spec (303 stars), ZERO admin CLI exists
- **Spec:** Official, auto-generated, updated daily
- **CLI gap:** Total for admin operations. discord.js is a library, not a CLI.
- **Killer commands:** `discord search --channel support --contains "crash"`, `discord members --role admin`, `discord audit --last-24h`
- **SQLite value:** Very high. Messages, channels, members, roles, audit logs - search is Discord's #1 complaint
- **HN angle:** "I built a Discord CLI with offline message search (the feature Discord won't build)"
- **Risk:** Discord's API has aggressive rate limits. Bot token required.

### 6. Notion (38/50) - HIGHEST SEARCH VALUE
- **Why top tier:** 5.3M npm/wk, cross-database search is a dream feature
- **Spec:** Partial - REST API but no published OpenAPI file. Would need to write from docs.
- **CLI gap:** CORRECTED. Community CLIs exist (MrRichRobinson/notion-cli, nitaiaharoni1/notion-cli in bash, @coastal-programs/notion-cli). But none are dominant and none have a data layer. The gap is narrower than "zero CLI" but still wide for a SQLite-backed search CLI.
- **Killer commands:** `notion search "roadmap" --all-databases`, `notion sync && notion sql "SELECT * FROM tasks WHERE status='Done'"`, `notion export --format=md`
- **SQLite value:** EXTREME. Pages, databases, rows, comments - offline cross-database search
- **HN angle:** "I built a Notion CLI that lets you search across ALL your databases from the terminal"
- **Risk:** No OpenAPI spec means writing spec from docs. Higher effort. Community CLIs exist but none have the data layer.

### 7. PostHog (40/50) - CORRECTED: CLI EXISTS BUT DATA LAYER GAP
- **Correction:** `@posthog/cli` exists with login, query (SQL), sourcemap upload, and endpoints-as-code management. This is NOT a "zero CLI" situation.
- **Remaining gap:** No local data layer. No offline search. No feature flag toggling. No event tailing. No cohort/experiment management. The existing CLI is query+deploy-focused, not data-layer-focused.
- **Spec:** OpenAPI 3.0, auto-generated
- **CLI gap:** NARROWER than first reported. The existing CLI handles auth, SQL queries, and sourcemaps. A printing-press CLI would need to differentiate on: SQLite sync, feature flag management, event tailing, experiment/cohort CRUD.
- **Killer commands that DON'T exist yet:** `posthog flags toggle signup-v2 --on`, `posthog tail --event $pageview`, `posthog experiments --running`, `posthog cohorts list`
- **Risk:** PostHog is actively developing their CLI. They could add these features.

### 8. PagerDuty (39/50) - BEST NICHE OPPORTUNITY
- **Why top tier:** On-call engineers LIVE in terminals, no official CLI, incident data is perfect for SQLite
- **Spec:** Official OpenAPI (32 stars)
- **CLI gap:** Huge. Best community CLI has 95 stars. No official offering.
- **Killer commands:** `pd oncall --me --next-7d`, `pd incidents --mttr --service=api`, `pd ack --all`
- **SQLite value:** Very high. Incidents, schedules, escalation policies, change events
- **HN angle:** "I built a PagerDuty CLI that tracks your on-call burden and MTTR from SQLite"
- **Risk:** Smaller developer reach (201k npm/wk). But the audience is highly engaged and terminal-native.

### 9. Linear (34/50) - CORRECTED: MULTIPLE CLIs EXIST
- **Correction:** Multiple community CLIs exist: `linearis` (npm, updated 2 days ago - JSON output, smart ID resolution, agent-friendly), `@schpet/linear-cli` (list/start/create PRs), `@anoncam/linear-cli` (comprehensive), plus a Rust CLI.
- **Spec:** GraphQL only. No REST, no OpenAPI.
- **CLI gap:** NARROWER than claimed. linearis in particular is actively maintained and agent-focused. The data layer gap still exists (no SQLite sync, no offline search) but the command gap is partially filled.
- **Remaining opportunity:** SQLite sync + offline velocity tracking + cross-team analytics. But GraphQL-only + existing CLIs = lower priority.
- **Risk:** GraphQL-only means harder to generate. Multiple active competitors.

## Tier 2: Consider Next (Score 31-38)

### Cal.com (33/50) - CORRECTED: MCP SERVER EXISTS
- **Correction:** `@calcom/cal-mcp` (official MCP server) covers all API endpoints via natural language. Not a traditional CLI, but fills the "programmatic access" gap for agent workflows.
- **Remaining opportunity:** A traditional CLI with `cal book` and `cal schedule` commands. The MCP server is agent-focused, not human-terminal-focused.
- **Risk:** The MCP server reduces the urgency. Self-hosters might still want a CLI, but the gap is narrower.

## Tier 2: Consider Next (Score 31-38)

| API | Score | Quick Take |
|-----|-------|-----------|
| Cloudflare (38) | Wrangler covers 20%, full API surface is huge. Good spec. |
| Slack (35) | Archived spec is a problem. Massive demand but hard to generate from. |
| Jira (34) | jira-cli at 5.4k stars covers commands but has no data layer. |
| Dub.co (34) | Small but zero CLI. Clean opportunity for a simple CLI. |
| Datadog (33) | pup at 538 stars covers a lot. Rate limit pain is real though. |
| Twilio (32) | Official CLI exists and is maintained. Hard to differentiate. |
| Shopify (31) | GraphQL-only push kills the OpenAPI generation path. |

## Tier 3: Skip (Score <31)

| API | Score | Why Skip |
|-----|-------|---------|
| Vercel (29) | Official CLI at 15k stars is comprehensive. |
| Supabase (27) | Official CLI at 1.8k stars, actively maintained. |
| Neon (26) | Official CLI works fine, only 5 open issues. |
| Turso (24) | No clean OpenAPI spec, small community. |

## Recommended Build Order (Corrected)

Optimized for: verified CLI gap, spec readiness, and maximum impact per build.

| Order | API | Verified Gap | Spec | Why This Order |
|-------|-----|-------------|------|---------------|
| **1st** | **Discord** | No admin CLI at all | Official OpenAPI 3.1 | Best spec + widest verified gap. Offline message search angle. |
| **2nd** | **HubSpot** | CLI is CMS-only | Official OpenAPI | Biggest enterprise opportunity. CRM data layer = instant value. |
| **3rd** | **Plaid** | Tiny community CLI | Official OpenAPI | Privacy-first personal finance. Best product-market fit. |
| **4th** | **PagerDuty** | Unsupported community CLI | Official OpenAPI | On-call engineers live in terminals. Small API = fast build. |
| **5th** | **Stripe** | CLI is event-focused | Best OpenAPI in industry | Easiest build. Use as the "showcase" for printing press quality. |
| **6th** | **Notion** | Fragmented community | No spec (write from docs) | Highest SQLite search value. More effort (no spec). |
| **7th** | **Cloudflare** | Wrangler covers 20% | Official OpenAPI | Huge API surface beyond Workers. DNS/security/analytics gap. |

**Dropped from build list:**
- PostHog: Official CLI exists with query+sourcemap+endpoints. Gap is narrower than claimed.
- Linear: Multiple active community CLIs (linearis updated 2 days ago). GraphQL-only.
- Cal.com: Official MCP server covers API surface.

## The Showcase Strategy

Each CLI serves a different role:

| Role | CLI | Why |
|------|-----|-----|
| **Flagship** | PostHog | Gets the most stars, proves the concept |
| **Privacy story** | Plaid | "Your financial data, on your machine" |
| **Enterprise proof** | HubSpot | Shows the press works for B2B/CRM |
| **Spec showcase** | Discord or Stripe | Shows what happens with a perfect OpenAPI spec |
| **GraphQL proof** | Linear | Proves the press handles GraphQL, not just REST |
| **Speed run** | PagerDuty | Small API, can build in one evening, proves velocity |

## The Steinberger Portfolio (The Quality Bar)

Peter Steinberger (steipete) is the reference for what a great API CLI looks like. His tools are the printing press's 10/10 benchmark. Understanding his portfolio reveals which APIs he's ALREADY covered and where the gaps remain.

| CLI | Stars | API | Architecture | Commands | What Makes It Special |
|-----|-------|-----|-------------|----------|----------------------|
| **gogcli** | 6,600 | Google Suite (Gmail, Calendar, Drive, Contacts, Tasks, Sheets, Docs, Slides, People, Keep, Admin, Groups) | Stateless, JSON-first, multi-auth | 100+ across 17 services | Least-privilege auth, command allowlist for agents, multi-account |
| **discrawl** | 583 | Discord | SQLite + FTS5, Gateway tail | 11 | Bot-token crawler, full-history backfill, offline member directory, raw SQL access |
| **wacli** | 687 | WhatsApp | SQLite + FTS5, continuous sync | 11 | Best-effort local sync, offline search, send messages, media download |
| **spogo** | 159 | Spotify | Stateless, browser cookies (no API key!) | 15+ | Bypasses official API rate limits via internal web endpoints |
| **sonoscli** | 108 | Sonos | Stateless, mDNS discovery | ~10 | Discover, group, queue, play. Smart home CLI. |
| **ordercli** | 57 | Foodora/Deliveroo | Unknown | ~5 | Food delivery order history in terminal |
| **blucli** | 29 | BluOS | Stateless | ~5 | Audio speaker control |

### Patterns From the Portfolio

1. **Two architectures, not one:** gogcli is stateless (no SQLite), discrawl/wacli use SQLite+FTS5. The right architecture depends on the data profile.
2. **12 commands beats 316:** discrawl has 11 commands and 583 stars. Depth (sync+search+sql+tail) beats breadth (one command per endpoint).
3. **Bot tokens for communication APIs:** discrawl uses a bot token, not user tokens. Legal, TOS-compliant, and scalable.
4. **Agent-native from day one:** gogcli has --json, --select, command allowlists for sandboxed agent runs. Not an afterthought.
5. **Doctor command is universal:** Every Steinberger CLI has a `doctor` that validates auth, connectivity, and config.

### What Steinberger HASN'T Built (Real Gaps)

| Category | APIs Without a Steinberger CLI | Best Candidate |
|----------|-------------------------------|----------------|
| **CRM** | HubSpot, Salesforce | HubSpot (official spec, CMS-only CLI) |
| **Payments** | Stripe, Plaid | Plaid (privacy angle, personal finance) |
| **Project Management** | Jira, Linear, Asana | Jira (largest user base, data layer gap) |
| **Incident Management** | PagerDuty, OpsGenie | PagerDuty (on-call engineers, no CLI) |
| **Product Analytics** | PostHog, Amplitude, Mixpanel | PostHog (open source, partial CLI) |
| **Infra** | Cloudflare (beyond Workers), AWS | Cloudflare (80% of API uncovered) |
| **Email** | SendGrid, Mailgun | Resend already built their own |

### How This Changes Our Rankings

**Discord drops:** discrawl (583 stars) already IS the Discord data-layer CLI. Building another one would compete directly with the quality benchmark. Not smart.

**The real opportunity is the categories Steinberger hasn't touched:** CRM, payments, incident management. These have high-value data, no data-layer CLIs, and large developer audiences.

## CORRECTED Rankings (v3 - After Steinberger Analysis)

| # | API | Score | Why | Steinberger Coverage |
|---|-----|-------|-----|---------------------|
| 1 | **HubSpot** | 43 | CRM data ops gap, official spec, 935k npm/wk | None |
| 2 | **Plaid** | 44 | Privacy-first personal finance, tiny existing CLI | None |
| 3 | **PagerDuty** | 39 | On-call engineers, no official CLI, small API = fast | None |
| 4 | **Stripe** | 43 | Best spec quality, massive reach, event-focused CLI gap | None |
| 5 | **Notion** | 38 | Cross-database offline search, fragmented community CLIs | None |
| 6 | **Cloudflare** | 38 | 80% of API uncovered by Wrangler | None |
| 7 | **Jira** | 34 | Huge user base, jira-cli has no data layer | None |
| ~~8~~ | ~~Discord~~ | ~~42~~ | ~~discrawl (583 stars) already covers this with SQLite+FTS5+sync+search+sql~~ | **COVERED by discrawl** |

**Discord is off the list.** discrawl does exactly what we'd build. Competing with the quality benchmark is the wrong strategy - we should build CLIs for APIs Steinberger hasn't touched.

## GitHub: Why It Belongs On The List

GitHub was our first printing press run. The CLI scored 73/100 but sync was broken. Despite that, the research reveals a genuine 1500-3000 star opportunity:

- gh (43.4k stars): Zero local persistence. The gh team [explicitly rejected offline mode](https://github.com/cli/cli/issues/2967) - open 5 years, no movement.
- gh-dash (11.2k stars): Gorgeous TUI but zero persistence. HN comments wished for local caching.
- github-to-sqlite (462 stars): Proves the model but dead since Dec 2023, Python-only, needs Datasette.
- **Nobody has combined** sync + SQLite + FTS5 + search + sql for GitHub in a Go binary.

**Score: 42/50** (Ease 6, Impact 9, Pop 10, Gap 7, SQLite 10)
**Name:** `ghdb` or `ghx`
**HN angle:** "Show HN: ghdb - sync your GitHub to SQLite, search offline, query with SQL"
**Realistic stars:** 1500-3000

## Easy Wins: Small APIs for Showcase Builds

APIs where the press can generate a polished CLI in 1-2 hours. Small specs, no competition.

| API | Stars | Est. Endpoints | Existing CLI | Why Easy |
|-----|-------|---------------|-------------|---------|
| **Dub.co** | 23k | ~40 | No CLI | Link analytics, clean spec in repo, zero competition |
| **Novu** | 39k | ~50 | No CLI | Notifications, massive community, delivery analytics |
| **OpenStatus** | 8.5k | ~20 | No CLI | Tiny API, monitoring trends perfect for SQLite |
| **Unkey** | 5.2k | ~25 | No CLI | API key mgmt, very small surface |

## Recommended Build Order (v4 - Final)

### Tier A: Flagship Builds (high effort, high reward)

| # | API | Score | HN Angle | Notes |
|---|-----|-------|----------|-------|
| 1 | **HubSpot** | 43 | "SQL query your CRM offline" | Enterprise gap, official spec |
| 2 | **GitHub** | 42 | "ghdb - GitHub in SQLite" | Scope to 40-50 ops. We have the learnings from v1. |
| 3 | **Plaid** | 44 | "Your bank data, on your machine" | Privacy-first, financial data |

### Tier B: Strong Opportunities (medium effort)

| # | API | Score | HN Angle | Notes |
|---|-----|-------|----------|-------|
| 4 | **PagerDuty** | 39 | "Track on-call burden from SQLite" | Small API = fast build |
| 5 | **Stripe** | 43 | "Offline payment analytics" | Best spec quality, showcase build |
| 6 | **Notion** | 38 | "Search ALL databases offline" | No spec = write from docs |

### Tier C: Easy Wins (1-2 hour showcase builds)

| # | API | Stars | Est. Endpoints | Why |
|---|-----|-------|---------------|-----|
| 7 | **Dub.co** | 23k | ~40 | Link analytics, zero CLI |
| 8 | **Novu** | 39k | ~50 | Notification data, huge community |
| 9 | **OpenStatus** | 8.5k | ~20 | Tiny API, monitoring trends |

### Dropped (with reasons)
- **Discord:** discrawl (583 stars) already covers SQLite+FTS5+sync+search+sql
- **PostHog:** Official CLI exists (query+sourcemap+endpoints)
- **Linear:** Multiple community CLIs (linearis updated 2 days ago), GraphQL-only
- **Cal.com:** Official MCP server covers API surface
- **Resend:** Official CLI shipped 6 weeks ago, 53 commands

## Sources

- All npm download numbers from npmjs.com (March 2026)
- GitHub star counts verified via gh CLI and web
- OpenAPI specs verified by checking repos and attempting downloads
- Pain points sourced from GitHub issues, Reddit, HN, and Stack Overflow
- CLI existence verified by searching GitHub for `<api-name> CLI` and checking official repos
