---
title: "Power User Workflows: GitHub CLI"
type: feat
status: active
date: 2026-03-27
phase: "0.5"
api: "github"
---

# Power User Workflows: GitHub CLI

## Overview

GitHub is a **Developer Platform** archetype. Power users want compound commands that combine multiple API calls into single operations - not 1,107 individual endpoint wrappers. The official `gh` CLI already covers interactive workflows well. Our opportunity is data-layer-powered analytics and bulk operations that `gh` can't do because it has no local persistence.

All 16 endpoints required by the workflows below have been validated against the OpenAPI spec.

## API Archetype Classification

**Developer Platform** - repos, PRs, CI runs, releases, security alerts.

Key workflow categories:
- PR triage and review management
- CI monitoring and flaky test detection
- Release management and changelog generation
- Repository hygiene (stale issues, orphaned PRs)
- Contributor analytics and team health
- Security alert aggregation

## Workflow Ideas (13 total)

### 1. pr-triage - Cross-Repo PR Triage Report
**Steps:** List org repos -> fetch open PRs per repo -> enrich with review status -> sort by age/size -> group by reviewer
**API calls:** GET /orgs/{org}/repos + GET /repos/{owner}/{repo}/pulls + GET /repos/{owner}/{repo}/pulls/{pull_number}/reviews
**Frequency:** Daily (3) | Pain: High (3) | Feasibility: Easy (3) | Uniqueness: Partial (2) = **11/12**
**Note:** gh-dash does this as TUI but not as JSON-outputting CLI command

### 2. stale - Stale Issue/PR Detection
**Steps:** Fetch issues across repos -> filter by updated_at -> exclude labeled -> report or optionally comment
**API calls:** GET /repos/{owner}/{repo}/issues (with since param) or local DB query
**Frequency:** Weekly (2) | Pain: High (3) | Feasibility: Easy (3) | Uniqueness: No tool (3) = **11/12**

### 3. actions-health - CI/CD Health Report
**Steps:** Fetch workflow runs -> compute success/failure rate per workflow -> identify longest runs -> detect flaky patterns
**API calls:** GET /repos/{owner}/{repo}/actions/runs + GET /repos/{owner}/{repo}/actions/workflows
**Frequency:** Weekly (2) | Pain: High (3) | Feasibility: Easy (3) | Uniqueness: No tool (3) = **11/12**

### 4. changelog - Release Changelog Generator
**Steps:** Get compare between tags -> classify commits by conventional commit prefix -> include PR links -> format
**API calls:** GET /repos/{owner}/{repo}/compare/{basehead} + GET /repos/{owner}/{repo}/releases
**Frequency:** Per release (1) | Pain: High (3) | Feasibility: Easy (3) | Uniqueness: No CLI (3) = **10/12**

### 5. contributors - Contributor Leaderboard
**Steps:** Fetch commits + PRs + reviews per user -> score -> rank -> output table
**API calls:** GET /repos/{owner}/{repo}/contributors + GET /repos/{owner}/{repo}/commits + search/issues
**Frequency:** Monthly (1) | Pain: Medium (2) | Feasibility: Medium (2) | Uniqueness: No tool (3) = **8/12**

### 6. security - Security Alert Aggregation
**Steps:** Fetch code-scanning + dependabot alerts across repos -> group by severity -> prioritize
**API calls:** GET /repos/{owner}/{repo}/code-scanning/alerts + GET /repos/{owner}/{repo}/dependabot/alerts
**Frequency:** Weekly (2) | Pain: High (3) | Feasibility: Easy (3) | Uniqueness: Partial (2) = **10/12**

### 7. activity - User/Org Activity Timeline
**Steps:** Fetch events for user or org -> filter by type -> display as timeline
**API calls:** GET /users/{username}/events or GET /repos/{owner}/{repo}/events
**Frequency:** Daily (3) | Pain: Low (1) | Feasibility: Easy (3) | Uniqueness: No CLI (3) = **10/12**

### 8. repo-health - Repository Health Score
**Steps:** Check: has README? has LICENSE? has CI? open issue ratio? PR merge time? last commit age?
**API calls:** GET /repos/{owner}/{repo} + multiple checks
**Frequency:** Monthly (1) | Pain: Medium (2) | Feasibility: Easy (3) | Uniqueness: No CLI (3) = **9/12**

### 9. review-load - Reviewer Load Balancing
**Steps:** Count pending reviews per team member -> identify overloaded reviewers -> suggest redistribution
**API calls:** Search PRs by reviewer + GET reviews
**Frequency:** Weekly (2) | Pain: Medium (2) | Feasibility: Medium (2) | Uniqueness: No tool (3) = **9/12**

### 10. orphans - Orphaned PR Detection
**Steps:** Find PRs where branch is deleted, author left org, or base branch changed
**API calls:** GET /repos/{owner}/{repo}/pulls + branch checks
**Frequency:** Monthly (1) | Pain: Medium (2) | Feasibility: Medium (2) | Uniqueness: No tool (3) = **8/12**

### 11. label-audit - Label Consistency Check
**Steps:** List labels across org repos -> find inconsistencies (different colors, missing labels)
**API calls:** GET /repos/{owner}/{repo}/labels per repo
**Frequency:** Quarterly (1) | Pain: Low (1) | Feasibility: Easy (3) | Uniqueness: No tool (3) = **8/12**

### 12. burndown - Sprint Burndown (Local DB)
**Steps:** Query local DB for issues in milestone -> track close rate over time -> render chart data
**API calls:** Local DB query (issues synced via sync command)
**Frequency:** Daily during sprint (3) | Pain: High (3) | Feasibility: Hard (1) | Uniqueness: Partial (2) = **9/12**

### 13. dependents - Dependent Repository Discovery
**Steps:** Find repos that depend on a given repo (used_by graph)
**API calls:** GitHub dependency graph API or scrape
**Frequency:** Monthly (1) | Pain: Low (1) | Feasibility: Hard (1) | Uniqueness: Partial (2) = **5/12**

## Validation Results

All required endpoints confirmed present in OpenAPI spec with necessary query parameters:
- `since` param on issues: YES
- `state`, `sort`, `direction` on PRs: YES
- `since`, `until`, `author` on commits: YES
- Search API with `q` param for cross-repo queries: YES
- Compare API for changelog generation: YES
- Code-scanning and dependabot alert endpoints: YES

## Full Scoring Table

| # | Workflow | Freq | Pain | Feas | Uniq | Total | Status |
|---|---------|------|------|------|------|-------|--------|
| 1 | pr-triage | 3 | 3 | 3 | 2 | **11** | BUILD |
| 2 | stale | 2 | 3 | 3 | 3 | **11** | BUILD |
| 3 | actions-health | 2 | 3 | 3 | 3 | **11** | BUILD |
| 4 | changelog | 1 | 3 | 3 | 3 | **10** | BUILD |
| 5 | security | 2 | 3 | 3 | 2 | **10** | BUILD |
| 6 | activity | 3 | 1 | 3 | 3 | **10** | BUILD |
| 7 | contributors | 1 | 2 | 2 | 3 | **8** | BUILD |
| 8 | burndown | 3 | 3 | 1 | 2 | 9 | FUTURE |
| 9 | review-load | 2 | 2 | 2 | 3 | 9 | FUTURE |
| 10 | repo-health | 1 | 2 | 3 | 3 | 9 | FUTURE |
| 11 | orphans | 1 | 2 | 2 | 3 | 8 | FUTURE |
| 12 | label-audit | 1 | 1 | 3 | 3 | 8 | FUTURE |
| 13 | dependents | 1 | 1 | 1 | 2 | 5 | SKIP |

## Top 7 for Implementation (Phase 4 Mandatory)

1. **pr-triage** (11/12) - Cross-repo PR triage with age, size, review status grouping
2. **stale** (11/12) - Find stale issues/PRs across repos with configurable thresholds
3. **actions-health** (11/12) - CI success rates, flaky test detection, build duration trends
4. **changelog** (10/12) - Generate release changelogs from commit comparison
5. **security** (10/12) - Aggregate code-scanning + dependabot alerts across repos by severity
6. **activity** (10/12) - User/org activity timeline from events API
7. **contributors** (8/12) - Contributor leaderboard with commit/PR/review scoring

## Acceptance Criteria
- [x] API archetype classified (Developer Platform)
- [x] 13 workflow ideas generated
- [x] All required endpoints validated against OpenAPI spec
- [x] Each workflow scored on 4 dimensions
- [x] Top 7 selected for Phase 4 implementation

## Sources
- GitHub REST API OpenAPI spec (1,107 operations validated)
- gh-dash (11.2k stars) for PR triage prior art
- github-to-sqlite (462 stars) for data layer prior art
- github-actions-watcher for CI monitoring prior art
