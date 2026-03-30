---
date: 2026-03-29
topic: sniffed-api-rate-limiting
---

# Adaptive Rate Limiting for Sniffed APIs

## Problem Frame

Sniffed APIs are undocumented public endpoints reverse-engineered from browser traffic. They are designed for one browser making occasional calls, but our CLIs make bulk sequential requests — especially during sync (100-500+ paginated calls) and during the sniff discovery phase itself (10-30 probing calls). Without proactive throttling, these CLIs reliably hit 429 rate limits, wasting time on retries and risking IP bans.

The current generated client handles 429s reactively (detect, wait 5s, retry up to 3x). This is insufficient for sniffed APIs because:
- Each 429 wastes 5+ seconds in retry waits
- Undocumented APIs may not send `Retry-After` headers
- Some APIs escalate to IP bans after repeated 429s
- Without auth, limits are per-IP — shared IPs (corp networks, VPNs) share the budget
- Sync operations make hundreds of calls, so hitting the wall is guaranteed

This was observed concretely during the postman-explore-pp-cli build: 429s hit during endpoint discovery (rapid XHR probing) and during sync pagination across categories.

## Requirements

**Adaptive Rate Limiter Algorithm**

The same algorithm applies in two contexts — the sniff skill (behavioral instructions for Claude) and the generated CLI binary (compiled Go code). The implementation differs but the logic is identical.

- R1. Start at a conservative floor: 1 req/s for sniff phase, 2 req/s for generated CLI
- R2. After N consecutive successful requests (N=10 for CLI, N=5 for sniff), increase the rate by 25%
- R3. On a 429 response, immediately halve the rate and record this as the "discovered ceiling"
- R4. After 429 cooldown, resume at the halved rate. Cap future increases at 90% of the discovered ceiling
- R5. The ceiling is per-session only — not persisted across runs. Each invocation starts fresh at the conservative floor

**Sniff Phase (Skill Instructions)**

- R6. Update the printing-press skill (SKILL.md) to instruct Claude to pace API probing during sniff using the adaptive algorithm (R1-R5 with sniff-phase defaults)
- R7. Between browser-use eval calls that make API requests, apply the current rate delay before proceeding
- R8. On 429 during sniffing, log the event, apply the backoff, and continue — do not abort discovery

**Generated CLI (Client Template)**

- R9. Add a `rate.Limiter` (from `golang.org/x/time/rate`) to the client struct in `client.go.tmpl`
- R10. The limiter is active by default for sniffed APIs (`spec_source: sniffed` in the catalog). For official APIs, no limiter is active by default
- R11. The limiter integrates with the existing retry loop — it runs before each request, not after. A request is only sent after the limiter grants a token
- R12. Add a `--rate-limit` flag: accepts a number (req/s) for manual override, or `0` to disable. Default is `2` for sniffed APIs

**Sync Pacing**

- R13. Sync operations use the same rate limiter as all other requests (no separate mechanism)
- R14. Sync progress output should show the effective rate when human-friendly: `"Syncing collections: 50/600 (category: AI) [1.8 req/s]"`

## Success Criteria

- A `printing-press generate` from a sniffed catalog entry produces a CLI that can `sync` 500+ pages without hitting any 429s at the default rate
- The sniff phase in the skill can probe 30+ endpoints without hitting 429s
- An agent running the CLI in a loop (search, browse, search, browse) does not accumulate 429s
- `--rate-limit 0` disables the limiter for users who know their API can handle high throughput

## Scope Boundaries

- No distributed rate limiting across users or processes (server's job)
- No persistence of discovered ceiling across sessions (undocumented APIs can change limits)
- No rate limit detection via response time analysis (unreliable)
- No changes to cache TTL — keep 5 min for all (defer if needed later)
- No per-endpoint rate limit configuration (too granular, unknown values)

## Key Decisions

- **Same algorithm, two implementations**: The sniff skill and the generated CLI use the same adaptive logic. This is conceptually clean and avoids separate mental models.
- **Conservative floor, not reactive-only**: Proactive throttling prevents 429s rather than recovering from them. The one-time cost of a slower first sync is worth the reliability.
- **Per-session ceiling**: Not persisting the discovered ceiling keeps things simple and safe — undocumented APIs can change limits without notice.
- **spec_source as the signal**: No new catalog fields for rate limiting. The existing `spec_source: sniffed` field (from PR #61) is sufficient to trigger the limiter.

## Dependencies / Assumptions

- `golang.org/x/time/rate` is the Go standard for token-bucket rate limiting — zero external dependencies beyond the Go extended library
- The `spec_source` field is already in the catalog schema (PR #61 merged)
- The generator templates have access to the catalog entry metadata at generation time (need to verify this is plumbed through)

## Outstanding Questions

### Deferred to Planning
- [Affects R9][Technical] How does the generator template access the catalog entry's `spec_source` at generation time? Is it available in the template context, or does it need to be plumbed through?
- [Affects R10][Technical] Should the limiter be a compile-time decision (template conditional) or a runtime decision (config file check)? Compile-time is simpler but means official APIs can't opt in without regeneration.
- [Affects R6][Needs research] What specific sections of the printing-press skill need updating for the sniff pacing instructions? Review the current sniff gate flow in SKILL.md.
- [Affects R12][Technical] Should `--rate-limit` accept `auto` as a value to explicitly enable adaptive mode, or is adaptive always-on the only mode?

## Next Steps

→ `/ce:plan` for structured implementation planning
