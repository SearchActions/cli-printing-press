---
title: "Top 50 API Candidates for CLI Printing Press"
type: research
status: active
date: 2026-03-27
---

# Top 50 API Candidates for CLI Printing Press

Which APIs should we build CLIs for? Ranked by composite score across 5 dimensions.

## Scoring Dimensions (each 1-10)

| Dimension | What it measures | 1 (worst) | 10 (best) |
|-----------|-----------------|-----------|-----------|
| **Ease** | Spec quality, REST vs GraphQL, endpoint count | No spec, GraphQL-only, 1000+ endpoints | Clean official OpenAPI, <200 endpoints |
| **Impact** | How much it matters (breadth of use, pain solved) | Niche tool, mild convenience | Millions of users, solves real daily pain |
| **Popularity** | Developer reach (npm downloads, stars, community) | <100k npm/wk, <5k stars | >1M npm/wk, >30k stars |
| **CLI Gap** | How underserved the CLI space is | Official CLI covers everything | No CLI exists, massive demand |
| **SQLite Fit** | How much local persistence + offline search adds value | Ephemeral data, no search need | Accumulating data, high search/analytics need |

---

## The Rankings

### Tier A: Flagship Builds (Score 38+)

High effort, high reward. These are the CLIs worth spending a full printing press run on.

| # | API | Category | Ease | Impact | Pop | Gap | SQLite | **Total** | Existing CLI | HN Angle |
|---|-----|----------|------|--------|-----|-----|--------|-----------|-------------|----------|
| 1 | **Plaid** | Finance | 9 | 9 | 7 | 9 | 10 | **44** | 57-star community, 6yr old | "Your bank data, on your machine" |
| 2 | **HubSpot** | CRM | 8 | 8 | 8 | 9 | 10 | **43** | CMS-only (zero CRM data) | "SQL query your CRM offline" |
| 3 | **Stripe** | Payments | 10 | 9 | 10 | 6 | 8 | **43** | Official (event-focused) | "Offline payment analytics" |
| 4 | **GitHub** | Developer | 6 | 9 | 10 | 7 | 10 | **42** | gh (workflow, no data layer) | "ghdb - GitHub in SQLite" |
| 5 | **PostHog** | Analytics | 8 | 9 | 8 | 6 | 9 | **40** | Official (query+sourcemap) | "Toggle feature flags from terminal" |
| 6 | **PagerDuty** | Incidents | 8 | 8 | 5 | 9 | 9 | **39** | 95-star community | "Track on-call burden from SQLite" |
| 7 | **Novu** | Notifications | 7 | 7 | 9 | 9 | 7 | **39** | No CLI (39k stars!) | "Notification delivery analytics" |
| 8 | **Notion** | Productivity | 5 | 8 | 8 | 7 | 10 | **38** | Fragmented community | "Search ALL databases offline" |
| 9 | **Cloudflare** | Infrastructure | 7 | 8 | 9 | 7 | 7 | **38** | Wrangler (Workers only, 20%) | "The other 80% of Cloudflare" |

### Tier B: Strong Opportunities (Score 33-37)

Medium effort, clear value. Good for building the printing press portfolio.

| # | API | Category | Ease | Impact | Pop | Gap | SQLite | **Total** | Existing CLI | Notes |
|---|-----|----------|------|--------|-----|-----|--------|-----------|-------------|-------|
| 10 | **Slack** | Communication | 3 | 7 | 9 | 7 | 9 | **35** | Official (app-dev only) | Archived spec is a problem |
| 11 | **Jira** | Project Mgmt | 8 | 7 | 6 | 5 | 8 | **34** | jira-cli 5.4k stars (no data layer) | Huge user base |
| 12 | **Linear** | Project Mgmt | 4 | 8 | 7 | 5 | 10 | **34** | Multiple community CLIs | GraphQL-only |
| 13 | **Dub.co** | Link Mgmt | 7 | 5 | 7 | 10 | 5 | **34** | No CLI (23k stars) | Easy win - small spec |
| 14 | **Datadog** | Observability | 6 | 7 | 8 | 5 | 7 | **33** | pup 538 stars | Rate limit pain |
| 15 | **Cal.com** | Scheduling | 7 | 7 | 8 | 5 | 6 | **33** | MCP server exists | 41k stars, self-hosters |
| 16 | **Twilio** | Communication | 8 | 6 | 10 | 3 | 5 | **32** | Official (188 stars) | Hard to differentiate |
| 17 | **Shopify** | E-commerce | 3 | 7 | 7 | 6 | 8 | **31** | Official (dev-only) | GraphQL-only now |

### Tier C: Easy Wins (Score 28-32)

Small APIs, 1-2 hour builds. Good for showcasing what the press can do.

| # | API | Category | Ease | Impact | Pop | Gap | SQLite | **Total** | Existing CLI | Notes |
|---|-----|----------|------|--------|-----|-----|--------|-----------|-------------|-------|
| 18 | **OpenStatus** | Monitoring | 9 | 4 | 5 | 9 | 5 | **32** | No CLI (8.5k stars) | Tiny API, monitoring trends |
| 19 | **Unkey** | API Keys | 8 | 4 | 4 | 9 | 4 | **29** | No CLI (5.2k stars) | Very small surface |
| 20 | **Vercel** | Deployment | 7 | 5 | 8 | 4 | 5 | **29** | Official (15k stars) | CLI is comprehensive |

### Tier D: Skip (Score <28)

Official CLIs already cover these well, or the opportunity is too small.

| # | API | Category | Ease | Impact | Pop | Gap | SQLite | **Total** | Why Skip |
|---|-----|----------|------|--------|-----|-----|--------|-----------|---------|
| -- | Supabase | BaaS | 4 | 6 | 10 | 3 | 4 | **27** | Official CLI (1.8k stars), vendor-owned |
| -- | Neon | Postgres | 7 | 4 | 7 | 4 | 4 | **26** | Official CLI works fine |
| -- | Turso | Edge DB | 4 | 4 | 6 | 6 | 4 | **24** | No clean spec, small community |
| -- | Discord | Communication | 9 | 7 | 7 | 10 | 9 | **42** | discrawl (583 stars) already covers this |
| -- | Resend | Email | 8 | 5 | 6 | 2 | 3 | **24** | Official CLI shipped 6 weeks ago, 53 commands |

### APIs 21-50 (From Extended Research)

| # | API | Category | Ease | Impact | Pop | Gap | SQLite | **Total** | Existing CLI | Notes |
|---|-----|----------|------|--------|-----|-----|--------|-----------|-------------|-------|
| 21 | **Airtable** | SaaS/Database | 6 | 7 | 7 | 9 | 8 | **37** | None (dead npm pkg) | 315k npm/wk, bases+records+automations |
| 22 | **SendGrid** | Email | 7 | 7 | 9 | 9 | 5 | **37** | None | 3.3M npm/wk, email templates+stats |
| 23 | **Intercom** | Support | 6 | 7 | 7 | 9 | 8 | **37** | None | Conversations, contacts, articles |
| 24 | **OpenAI** | AI/ML | 8 | 9 | 10 | 7 | 5 | **39** | Weak (90 dl/wk) | 16M npm/wk, spec has 2.3k stars |
| 25 | **Directus** | Headless CMS | 7 | 6 | 8 | 9 | 7 | **37** | None (admin UI only) | 34.6k stars, auto-generated spec |
| 26 | **Algolia** | Search | 6 | 6 | 9 | 6 | 5 | **32** | Weak (106 stars) | 5.4M npm/wk, index+rules mgmt |
| 27 | **Grafana** | Monitoring | 6 | 8 | 9 | 8 | 7 | **38** | None for API | 72.8k stars, dashboards+alerts |
| 28 | **Sentry** | Error Tracking | 5 | 8 | 9 | 6 | 7 | **35** | sentry-cli (sourcemaps only) | 43.5k stars, issue querying gap |
| 29 | **Square** | Payments | 7 | 7 | 5 | 9 | 8 | **36** | None | Payments, orders, inventory |
| 30 | **Anthropic** | AI/ML | 6 | 8 | 9 | 9 | 3 | **35** | None | 9.8M npm/wk, ~20 endpoints |
| 31 | **Appwrite** | BaaS | 5 | 6 | 9 | 5 | 5 | **30** | Weak (SDK wrapper) | 55.3k stars |
| 32 | **Kong** | API Gateway | 6 | 6 | 9 | 7 | 5 | **33** | None for Admin API | 43.1k stars, services+routes+plugins |
| 33 | **Livekit** | Video/Audio | 5 | 6 | 7 | 7 | 4 | **29** | Weak (334 stars, rooms only) | 17.8k stars |
| 34 | **Trigger.dev** | Background Jobs | 5 | 5 | 6 | 7 | 5 | **28** | None (SDK only) | 14.3k stars, jobs+runs mgmt |
| 35 | **Logto** | Auth | 7 | 5 | 5 | 9 | 5 | **31** | None | 11.8k stars, users+roles+SSO |
| 36 | **LaunchDarkly** | Feature Flags | 7 | 6 | 7 | 7 | 5 | **32** | Weak (22 stars) | 645k npm/wk, flags+segments |
| 37 | **Contentful** | Headless CMS | 6 | 6 | 8 | 5 | 6 | **31** | 352 stars (space mgmt only) | 1.1M npm/wk, entries+assets gap |
| 38 | **Highlight** | Monitoring | 5 | 5 | 5 | 9 | 6 | **30** | None | 9.2k stars, sessions+errors+logs |
| 39 | **Svix** | Webhooks | 8 | 4 | 3 | 9 | 4 | **28** | None | 3.1k stars, clean OpenAPI |
| 40 | **Mistral** | AI/ML | 6 | 7 | 7 | 9 | 3 | **32** | None | 2.1M npm/wk |
| 41 | **Replicate** | AI/ML | 5 | 6 | 6 | 8 | 4 | **29** | Weak (23 dl/wk) | 386k npm/wk, run models |
| 42 | **Mux** | Video | 7 | 5 | 6 | 9 | 5 | **32** | None | 214k npm/wk, assets+streams |
| 43 | **Cohere** | AI/ML | 5 | 6 | 6 | 9 | 3 | **29** | None | 395k npm/wk |
| 44 | **Mapbox** | Maps | 6 | 5 | 7 | 7 | 4 | **29** | Weak (172 stars, Python) | 317k npm/wk, tilesets+styles |
| 45 | **Sanity** | Headless CMS | 3 | 5 | 7 | 5 | 5 | **25** | Good (@sanity/cli) | GROQ not REST |
| 46 | **Upstash** | Serverless DB | 5 | 4 | 7 | 7 | 3 | **26** | Weak (24 stars) | 1.9M npm/wk |
| 47 | **Hasura** | GraphQL Engine | 3 | 5 | 8 | 4 | 4 | **24** | Good (hasura-cli, 49k dl/wk) | Migration-focused |
| 48 | **Inngest** | Background Jobs | 5 | 4 | 4 | 5 | 4 | **22** | Good (202k dl/wk) | Dev server focused |
| 49 | **ClickHouse** | Database | 4 | 5 | 9 | 3 | 3 | **24** | Built-in client | SQL-focused already |
| 50 | **Resend** | Email | 8 | 5 | 6 | 2 | 3 | **24** | Official (53 cmds, 6 weeks old) | Too late - they shipped |

**Notable additions to Tier A/B from the extended list:**

- **OpenAI** (39/50) jumps into Tier A - 16M npm/wk, official OpenAPI spec (2.3k stars), weak CLI. Manage models, fine-tunes, files, batches, assistants.
- **Grafana** (38/50) enters Tier A - 72.8k stars, no API CLI (separate from Grafana server). Dashboard+alert+datasource management.
- **Airtable** (37/50), **SendGrid** (37/50), **Intercom** (37/50), **Directus** (37/50) all strong Tier B with zero CLIs.

---

## Build Order

### Phase 1: Prove the Press (Easy wins, 1-2 hours each)

| # | API | Stars | Endpoints | Why | Effort |
|---|-----|-------|-----------|-----|--------|
| 1 | **Dub.co** | 23k | ~30 | Zero CLI, clean spec, link analytics | 1-2 hours |
| 2 | **OpenStatus** | 8.5k | ~20 | Tiny API, monitoring trends | 1 hour |
| 3 | **Unkey** | 5.2k | ~25 | API key mgmt, very small | 1 hour |

### Phase 2: Flagships (Full press runs, half day to full day each)

| # | API | Score | Why This Order | Effort |
|---|-----|-------|---------------|--------|
| 4 | **HubSpot** | 43 | Biggest CRM gap, official spec, 935k npm/wk | Full run |
| 5 | **OpenAI** | 39 | 16M npm/wk, official spec (2.3k stars), weak CLI | Full run |
| 6 | **GitHub** | 42 | We have v1 learnings. Scope to 40 ops. `ghdb`. | Full run (v2) |
| 7 | **Plaid** | 44 | Privacy-first personal finance | Full run |
| 8 | **Grafana** | 38 | 72.8k stars, no API CLI, dashboards+alerts | Full run |
| 9 | **Stripe** | 43 | Best spec in industry, showcase quality | Full run |

### Phase 3: Fill the Portfolio (Medium effort)

| # | API | Score | Why | Effort |
|---|-----|-------|-----|--------|
| 10 | **PagerDuty** | 39 | Small API, passionate on-call niche | Half day |
| 11 | **Novu** | 39 | 39k stars, zero CLI, notification analytics | Half day |
| 12 | **Airtable** | 37 | 315k npm/wk, zero CLI, base/record mgmt | Half day |
| 13 | **SendGrid** | 37 | 3.3M npm/wk, zero CLI, email templates+stats | Half day |
| 14 | **Intercom** | 37 | 298k npm/wk, zero CLI, conversations+contacts | Half day |
| 15 | **Notion** | 38 | Highest search value, no spec = write from docs | Full run |

**Strategy:** Start with 3 easy wins to prove the press + verify loop works. Then alternate flagships with portfolio builders. Each shipped CLI validates the printing press and feeds back learnings.

---

## The Steinberger Portfolio (Quality Bar)

Peter Steinberger's CLIs are the 10/10 benchmark:

| CLI | Stars | API | Architecture | What Makes It Special |
|-----|-------|-----|-------------|----------------------|
| **gogcli** | 6,600 | Google Suite | Stateless, JSON-first | 17 services, multi-auth, agent allowlists |
| **wacli** | 687 | WhatsApp | SQLite + FTS5 | Offline search, continuous sync, send messages |
| **discrawl** | 583 | Discord | SQLite + FTS5 | Bot-token crawler, full-history backfill, raw SQL |
| **spogo** | 159 | Spotify | Stateless | Bypasses API via browser cookies |
| **sonoscli** | 108 | Sonos | Stateless | mDNS discovery, speaker control |
| **ordercli** | 57 | Foodora/Deliveroo | Unknown | Food delivery history |
| **blucli** | 29 | BluOS | Stateless | Audio speaker control |

**What Steinberger hasn't built:** CRM, payments, incident management, project management, product analytics, infrastructure. These are our opportunity.

**Key patterns:** Two architectures (stateless for small APIs, SQLite+FTS5 for data-heavy ones). Doctor command in every CLI. 12 commands beats 316. Agent-native from day one.

---

## GitHub: Deep Analysis

The data-layer gap between gh (43.4k stars, no persistence) and github-to-sqlite (462 stars, dead since Dec 2023) is wide open.

- gh team [explicitly rejected offline mode](https://github.com/cli/cli/issues/2967) - open 5 years
- gh-dash (11.2k stars) proves demand for dashboards but has zero persistence
- Nobody has sync + SQLite + FTS5 + search + sql for GitHub in a Go binary
- Realistic star potential: 1500-3000
- Name: `ghdb` or `ghx`
- Scope to 40-50 operations (Tier 1: issues, PRs, commits, repos, notifications)
- We have v1 learnings: test with `printing-press verify`, product thesis before code

---

## Research Methodology

### What we checked for each API
- GitHub repo stars (via `gh api`)
- npm package downloads (via npmjs.com)
- OpenAPI spec existence (searched GitHub + official docs)
- Existing CLI tools (searched npm + GitHub + web)
- Developer pain points (GitHub issues, Reddit, HN, Stack Overflow)
- SQLite sync value (data profile analysis)

### Correction log
- PostHog: First pass said "ZERO CLI" - actually has `@posthog/cli` (query+sourcemap+endpoints)
- Linear: First pass said "no CLI at all" - actually has linearis, @schpet/linear-cli, etc.
- Cal.com: First pass said "NO CLI" - actually has `@calcom/cal-mcp` MCP server
- Discord: Confirmed no admin CLI, BUT discrawl (583 stars) already covers SQLite+FTS5+sync

### Sources
- npm download numbers from npmjs.com (March 2026)
- Star counts verified via gh CLI
- OpenAPI specs verified by checking repos and attempting downloads
- Steinberger portfolio from github.com/steipete
- GitHub gap analysis from cli/cli issues, gh-dash HN thread, github-to-sqlite repo
