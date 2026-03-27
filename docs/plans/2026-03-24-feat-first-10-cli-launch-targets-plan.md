---
title: "First 10 CLI Launch Targets - Filling the World's Biggest API Gaps"
type: feat
status: active
date: 2026-03-24
deepened: 2026-03-24
---

# First 10 CLI Launch Targets - Filling the World's Biggest API Gaps

## Enhancement Summary

**Deepened on:** 2026-03-24
**Research agents used:** 5 (spec verification, killer workflows, Steinberger gap analysis, brand strategy, architecture review)

### Key Improvements from Deepening
1. **All 7 spec URLs verified** - every original URL was wrong. Corrected with fetchable raw GitHub URLs.
2. **Reordered targets** - Plaid moved to #1 launch (HN virality) instead of Intercom. PayPal demoted (13 separate specs = complex).
3. **Honest quality gap scored** - our generated CLIs are 4.9/10 vs Steinberger's 10/10. Three template fixes identified to reach 6.9.
4. **Killer workflow per target** - each CLI now has a specific "install it for this" command.
5. **Architecture decisions locked** - separate repos, no shared library, no meta-CLI, `USER` placeholder must be fixed first.

### Blockers Found
- **`USER` placeholder in 6+ templates** must become `{{.Owner}}` before generating any public CLI
- **PayPal has 13 separate specs**, not one - needs multi-spec merge support
- **ClickUp has no official spec** - best is a community repo on a `develop` branch

---

## The Steinberger Quality Bar (10/10)

Based on deep analysis of Peter Steinberger's 21+ CLI tools (gogcli at 6.5K stars, wacli at 681, imsg at 923):

1. **Fill a vacuum** - target services with NO existing CLI
2. **Three output modes** - human (colored), `--json` (machine), `--plain` (pipeable)
3. **Auth that works everywhere** - env var + config file + keyring + browser OAuth
4. **One binary, zero deps** - Go, static, `brew install` or download tarball
5. **README as product page** - 150+ real examples, quickstart, troubleshooting
6. **Agent-native** - `--json` + `--select` + `--dry-run` + non-interactive
7. **doctor command** - validates auth, connectivity, config
8. **Typed errors with hints** - 401 says "run auth login", 404 suggests valid IDs
9. **Conventional commits + AGENTS.md** - ready for AI-assisted maintenance
10. **Relentless scope expansion** - ship v1, keep adding commands

### Honest Gap: Our Generated CLIs Score 4.9/10

| Dimension | Steinberger | Our CLIs | Score | Gap |
|-----------|-------------|----------|-------|-----|
| Output modes | human + --json + --plain + --select | --json + --plain + --quiet | 6/10 | No --select for field filtering |
| Auth | env var + config + keyring + OAuth | env var + config + OAuth | 7/10 | No keyring (nice-to-have) |
| Error handling | typed + hints + suggestions | typed exit codes, no hints | 5/10 | 401 doesn't say "run auth login" |
| Terminal UX | color + NO_COLOR + spinner | color + NO_COLOR, no spinner | 6/10 | No spinner for long ops |
| README | 150+ real examples, product page | skeleton, ~30 lines | **2/10** | Biggest gap |
| Doctor | auth + connectivity + config | auth + connectivity + config | 7/10 | Decent |
| Agent-native | --json + --select + --dry-run | --json + --dry-run | 5/10 | No --select |
| Local caching | SQLite FTS5, offline search | None | **1/10** | Feature, not template |

### 3 Template Fixes Before Launch (4.9 -> 6.9)

1. **`--select` flag** (~30 lines in helpers.go.tmpl) - field filtering for --json output. Closes agent-native + output mode gaps.
2. **Error hints map** (~20 lines in helpers.go.tmpl) - 401 -> "run auth login", 404 -> "run list to see valid IDs"
3. **README rewrite** (readme.md.tmpl) - quickstart workflow, per-command examples, troubleshooting section

---

## The 10 Targets (Reordered by Launch Impact)

### Launch CLI #1: Plaid - Gap Score 8/10

**The first CLI we ship publicly. Maximum HN/X virality.**

| Dimension | Detail |
|-----------|--------|
| **What it does** | Banking - account linking, transactions, identity, income |
| **Developer population** | ~100K developers, 12K+ fintech companies (Venmo, Robinhood, etc.) |
| **Official CLI** | None |
| **Best community CLI** | landakram/plaid-cli (57 stars, Go, abandoned 2021) - proves demand |
| **OpenAPI spec** | `https://raw.githubusercontent.com/plaid/plaid-openapi/master/2020-09-14.yml` (VERIFIED) |
| **Auth** | clientId + secret + plaidVersion (all per-request) |

**Killer workflow:** Create a sandbox bank connection and fetch transactions without spinning up a frontend.
```
plaid-cli sandbox create-item --institution ins_109508 --products transactions,auth | plaid-cli transactions sync --days 30
```
**Pain without CLI:** Every Plaid integration cycle requires a React frontend with Link UI just to create a test Item. Developers write disposable 6-step curl scripts every time.

**Why ship first:** Fintech devs are HN's core audience. The 10-second demo of replacing a 15-minute browser workflow is visceral. Zero CLI competition. The abandoned 57-star Go CLI proves demand.

**Persona:** Fintech developer building banking integrations (uses terminal daily, 10-20 times during active Plaid work)

### Launch CLI #2: PagerDuty - Gap Score 7/10

**The "3am incident ack" CLI. Emotionally powerful story.**

| Dimension | Detail |
|-----------|--------|
| **What it does** | Incident management - on-call, alerts, escalations |
| **Developer population** | ~200K developers |
| **Official CLI** | None (explicitly "not endorsed or supported") |
| **Best community CLI** | martindstone/pagerduty-cli (95 stars, Node.js, abandoned Oct 2024) |
| **OpenAPI spec** | `https://raw.githubusercontent.com/PagerDuty/api-schema/main/reference/REST/openapiv3.json` (VERIFIED) |
| **Auth** | API Key token (`Token token=<token>` or `Bearer <token>`) |

**Killer workflow:** Acknowledge all your incidents and check who's on-call before opening your laptop.
```
pagerduty-cli incident ack -m && pagerduty-cli oncall list --escalation-policy "Backend"
```
**Pain without CLI:** At 3am, fumbling through browser login, finding the incident, clicking ack - 60+ seconds when the escalation window is ticking.

**Why ship second:** Every SRE would share this. Differentiator vs dead Node.js CLI: Go binary starts in 50ms vs 2 seconds. Zero deps at 3am matters.

### Launch CLI #3: Intercom - Gap Score 9/10

| Dimension | Detail |
|-----------|--------|
| **What it does** | Customer messaging - conversations, contacts, articles, help center |
| **Developer population** | ~100K developers, 25K+ businesses |
| **Official CLI** | None |
| **Best community CLI** | kouk/intercom-cli (1 star, CoffeeScript, 2014). Dead. |
| **OpenAPI spec** | `https://raw.githubusercontent.com/intercom/Intercom-OpenAPI/main/descriptions/2.11/api.intercom.io.yaml` (VERIFIED) |
| **Auth** | Bearer token |

**Killer workflow:** Bulk-close stale conversations that have been idle 7+ days.
```
intercom-cli conversations list --status open --tag "billing" --idle-days 7 | intercom-cli conversations close --note "Auto-closed: no reply in 7 days"
```
**Persona:** Support Operations Lead at a SaaS company (weekly inbox hygiene)

### Launch CLI #4: Pipedrive - Gap Score 9/10

| Dimension | Detail |
|-----------|--------|
| **What it does** | CRM - deals, contacts, pipelines, activities |
| **Developer population** | ~50K developers, 100K+ companies |
| **Official CLI** | None |
| **Best community CLI** | None found |
| **OpenAPI spec** | `https://developers.pipedrive.com/docs/api/v1/openapi.yaml` (VERIFIED) |
| **Auth** | API Key + OAuth 2.0 |

**Killer workflow:** Log a sales call and advance the deal stage without opening the browser.
```
pipedrive-cli deals update "Acme Corp" --stage "Proposal Sent" --note "Demo went well, sending SOW by Friday" --activity call
```
**Persona:** Account Executive who delays CRM updates because of browser friction (5-10x daily)

### Launch CLI #5: Square - Gap Score 8/10

| Dimension | Detail |
|-----------|--------|
| **What it does** | Payments, inventory, orders, customers, catalog |
| **Developer population** | ~200K developers, millions of merchants |
| **Official CLI** | None |
| **Best community CLI** | nickrobinson/square-cli (3 stars, Go, minimal) |
| **OpenAPI spec** | `https://raw.githubusercontent.com/square/connect-api-specification/master/api.json` (VERIFIED) |
| **Auth** | OAuth 2.0 with 50+ permission scopes |

**Killer workflow:** Reconcile today's payments across all locations as CSV.
```
square-cli payments list --location all --begin-time today --format csv > daily-reconciliation.csv
```
**Persona:** Multi-location merchant bookkeeper (daily end-of-day reconciliation)

### Launch CLI #6: ClickUp - Gap Score 8/10

| Dimension | Detail |
|-----------|--------|
| **What it does** | Project management - tasks, docs, goals, time tracking |
| **Developer population** | ~100K developers, 800K+ teams |
| **Official CLI** | None |
| **Best community CLI** | None with traction |
| **OpenAPI spec** | `https://raw.githubusercontent.com/rksilvergreen/clickup_openapi_spec_v2/develop/clickup-api-v2-reference.yaml` (VERIFIED - community, ~80 endpoints) |
| **Auth** | Personal API Token or OAuth 2.0 |

**Killer workflow:** Bulk-create a sprint's worth of tasks from a YAML file.
```
clickup-cli tasks import --list "Sprint 47" --file sprint-tasks.yaml --space "Engineering"
```
**Risk:** No official spec. Community spec on `develop` branch could go stale.

### Launch CLI #7: PayPal - Gap Score 8/10 (COMPLEX)

| Dimension | Detail |
|-----------|--------|
| **What it does** | Payments, invoicing, payouts, subscriptions, disputes |
| **Developer population** | ~1M+ developers, 35M merchant accounts |
| **Official CLI** | None |
| **Best community CLI** | jeffharrell/paypal-cli (3 stars, 2016). Dead. |
| **OpenAPI spec** | 13 SEPARATE specs at `https://raw.githubusercontent.com/paypal/paypal-rest-api-specifications/main/openapi/*.json` (VERIFIED) |
| **Auth** | OAuth 2.0 client credentials |

**Killer workflow:** Look up a failed payment and replay its webhook.
```
paypal-cli payments get PAY-1AB23456CD --webhooks | paypal-cli webhooks replay --event-id WH-7890
```
**Complexity warning:** PayPal publishes 13 separate spec files (orders, invoicing, payouts, subscriptions, disputes, transactions, etc.), not one monolithic spec. Needs multi-spec merge via `printing-press generate --spec orders.json --spec invoicing.json --spec payouts.json --name paypal`.

### Launch CLI #8: Airtable - Gap Score 9/10 (CUSTOM SPEC)

| Dimension | Detail |
|-----------|--------|
| **What it does** | Spreadsheet-database hybrid for ops/project management |
| **Developer population** | ~500K developers, 300K+ businesses |
| **Official CLI** | Only @airtable/blocks-cli (extensions, not core API) |
| **Best community CLI** | None with traction |
| **OpenAPI spec** | Per-base schema available. Needs custom YAML spec. |
| **Auth** | Personal Access Token (Bearer) |

**Killer workflow:** Bulk upsert 1,000 records from CSV with automatic batching and rate-limit backoff.
```
airtable-cli records upsert --base appXYZ123 --table "Leads" --file import.csv --merge-on "Email"
```
**Persona:** Operations lead who outgrew the web UI (weekly data syncs)

### Launch CLI #9: Notion - Gap Score 9/10 (CUSTOM SPEC, HIGHEST CEILING)

| Dimension | Detail |
|-----------|--------|
| **What it does** | All-in-one workspace - databases, pages, wikis |
| **Developer population** | ~4M developers, 100M+ users |
| **Official CLI** | None |
| **Best community CLI** | 4ier/notion-cli (87 stars, Go, active), nitaiaharoni1/notion-cli (bash, 2025 API) |
| **OpenAPI spec** | NOT published. Must write YAML spec from docs. |
| **Auth** | Internal integration token (Bearer) |

**Killer workflow:** Search across all pages and databases, output as structured JSON - "grep for Notion."
```
notion-cli search "Q2 OKR" --type page --format json | jq '.[].title'
```
**Risk:** Block-based content model is deeply nested. Multiple active community CLIs exist (87 stars). Highest ceiling (100M users) but hardest to build.

### Launch CLI #10: Mixpanel - Gap Score 9/10 (CUSTOM SPEC)

| Dimension | Detail |
|-----------|--------|
| **What it does** | Product analytics - events, funnels, retention, user profiles |
| **Developer population** | ~200K developers, 30K+ companies |
| **Official CLI** | None |
| **Best community CLI** | mixpanel-cli (14 stars, JS, abandoned 2020) |
| **OpenAPI spec** | No formal spec. Must write YAML from docs. |
| **Auth** | Service Account credentials |

**Killer workflow:** Query your key product metric without loading the dashboard.
```
mixpanel-cli query events --event "signup_completed" --from 7d --group-by "utm_source" --format table
```
**Persona:** Product Manager / Growth Engineer checking morning metrics (daily)

---

## Architecture Decisions (Locked)

| Decision | Choice | Reason |
|----------|--------|--------|
| Repo structure | **Separate repos** per CLI | Discoverability ("intercom-cli" on GitHub), independent stars, independent releases. Steinberger does this. |
| Shared library | **No** (standalone CLIs) | At 10 CLIs, regeneration is cheap. Extract at 50+. |
| Homebrew tap | **One formula per CLI** in `mvanhorn/homebrew-tap` | Matches steipete/homebrew-tap pattern |
| Meta-CLI installer | **No** | `brew install mvanhorn/tap/plaid-cli` is already one command |
| Versioning | **Independent semver** per CLI | GoReleaser injects via ldflags per repo |
| Config sharing | **No** | Each CLI gets `~/.config/<name>-cli/config.toml` independently |
| Generation metadata | Add `// Code generated by CLI Printing Press vX.Y.Z` to every file | Traceability for bug reports |

---

## Pre-Launch Blockers

Before generating any public CLI:

- [ ] **Fix `USER` placeholder** in 6+ templates - replace with `{{.Owner}}` variable
- [ ] **Add `--select` flag** to helpers.go.tmpl (~30 lines)
- [ ] **Add error hints map** to helpers.go.tmpl (~20 lines)
- [ ] **Rewrite readme.md.tmpl** with quickstart, real examples, troubleshooting
- [ ] **Add `// Code generated by CLI Printing Press` comment** to all templates
- [ ] **Create `mvanhorn/homebrew-tap`** GitHub repo
- [ ] **Create `regenerate-all.sh`** script in the printing-press repo

---

## Implementation Plan

### Phase 0: Template Quality Push (get from 4.9 to 6.9)

Fix the 3 template gaps + blockers above. No CLI generation until this is done.

### Phase 1: First 3 CLIs (Maximum Launch Impact)

1. **Plaid** - generate from verified OpenAPI spec, ship with Show HN
2. **PagerDuty** - generate from verified OpenAPI spec, ship on-call story on X
3. **Intercom** - generate from verified OpenAPI spec

### Phase 2: Next 4 CLIs (OpenAPI auto-generation)

4. **Pipedrive** - generate from verified spec
5. **Square** - regenerate from verified spec (already in catalog)
6. **ClickUp** - generate from community spec (risk: may go stale)
7. **PayPal** - merge 13 specs then generate (needs multi-spec support)

### Phase 3: Final 3 CLIs (Custom YAML specs)

8. **Airtable** - write YAML spec from docs (~40 commands)
9. **Notion** - write YAML spec from docs (block model complexity)
10. **Mixpanel** - write YAML spec for export/query/profile endpoints

### Phase 4: Distribution Blitz

- All 10 published to `mvanhorn/homebrew-tap`
- GoReleaser configured for all 10
- README for each with real dogfood output
- Landing page linking all 10 CLIs

---

## Verified OpenAPI Spec URLs

| # | API | Verified Fetchable URL | Format | Endpoints |
|---|-----|----------------------|--------|-----------|
| 1 | Plaid | `https://raw.githubusercontent.com/plaid/plaid-openapi/master/2020-09-14.yml` | YAML | 15+ (large, truncated) |
| 2 | PagerDuty | `https://raw.githubusercontent.com/PagerDuty/api-schema/main/reference/REST/openapiv3.json` | JSON | Many (incidents, users, schedules, etc.) |
| 3 | Intercom | `https://raw.githubusercontent.com/intercom/Intercom-OpenAPI/main/descriptions/2.11/api.intercom.io.yaml` | YAML | ~40+ paths |
| 4 | Pipedrive | `https://developers.pipedrive.com/docs/api/v1/openapi.yaml` | YAML | 20+ visible, 40+ resource groups |
| 5 | Square | `https://raw.githubusercontent.com/square/connect-api-specification/master/api.json` | JSON | Many (very large spec) |
| 6 | ClickUp | `https://raw.githubusercontent.com/rksilvergreen/clickup_openapi_spec_v2/develop/clickup-api-v2-reference.yaml` | YAML | ~80 (community, not official) |
| 7 | PayPal | 13 files at `https://raw.githubusercontent.com/paypal/paypal-rest-api-specifications/main/openapi/` | JSON | ~9 per file, ~100+ total |
| 8 | Airtable | Needs custom YAML spec | N/A | ~40 (estimated) |
| 9 | Notion | Needs custom YAML spec | N/A | ~30 (estimated) |
| 10 | Mixpanel | Needs custom YAML spec | N/A | ~20 (estimated) |

---

## Acceptance Criteria

- [ ] 3 template improvements shipped (--select, error hints, README rewrite)
- [ ] `USER` placeholder replaced with `{{.Owner}}` in all templates
- [ ] 10 CLIs generated, all passing 7 quality gates
- [ ] Each CLI has a killer workflow documented in its README
- [ ] Each CLI scores >= 75/100 on comparative analysis
- [ ] All 10 published to Homebrew tap
- [ ] Plaid CLI shipped with Show HN post
- [ ] At least 7 auto-generated from OpenAPI specs
- [ ] At least 3 hand-written YAML specs (Linear, Airtable, Notion or Mixpanel)

## Success Metrics

- **Show HN for Plaid CLI**: 50+ points
- **GitHub stars**: 100+ combined within first month
- **Homebrew installs**: 50+ within first month
- **"First CLI for X"**: 9/10 targets have no prior CLI with >10 stars

## Sources

### Research Agents (2026-03-24)
- Spec verification agent: verified all 7 URLs, found correct raw GitHub URLs
- Killer workflow agent: identified #1 use case per API with persona and frequency
- Steinberger gap agent: scored templates 4.9/10, identified 3 highest-ROI fixes
- Brand strategy agent: recommended Plaid as #1 launch for HN virality
- Architecture agent: locked separate repos, no shared lib, fix USER placeholder

### Steinberger Reference
- steipete/gogcli (6.5K stars) - 150+ commands, 16 Google services, 6 auth paths
- steipete/wacli (681 stars) - SQLite FTS5 offline search, QR auth
- steipete/imsg (923 stars) - only iMessage CLI in existence
- steipete/homebrew-tap - 21 formulas, separate repo per CLI

### API Documentation
- Plaid: github.com/plaid/plaid-openapi
- PagerDuty: github.com/PagerDuty/api-schema
- Intercom: github.com/intercom/Intercom-OpenAPI
- Pipedrive: developers.pipedrive.com
- Square: github.com/square/connect-api-specification
- ClickUp: github.com/rksilvergreen/clickup_openapi_spec_v2 (community)
- PayPal: github.com/paypal/paypal-rest-api-specifications (13 specs)
