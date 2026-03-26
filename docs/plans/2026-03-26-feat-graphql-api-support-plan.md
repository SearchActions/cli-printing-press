---
title: "GraphQL API Support for Printing Press"
type: feat
status: completed
date: 2026-03-26
---

# GraphQL API Support for Printing Press

## Overview

The printing-press previously refused to run on GraphQL APIs (Linear, GitHub GraphQL, Shopify Admin). It told users to "pick a different REST API." This is wrong - the printing-press's value is the 8-phase research + prediction + dogfood pipeline, not just the REST generator. GraphQL APIs should go through all phases with Phase 2 adapted to produce scaffolding instead of generated commands.

## Changes Made

### Phase 2: GraphQL Mode

When the API is GraphQL-only, Phase 2 now:
- Creates the project structure (go.mod, client, config, helpers, store)
- Generates a GraphQL client wrapper (`Client.Query(query, variables)`)
- Generates root.go with global flags
- Skips per-endpoint command generation (no REST endpoints exist)
- Proceeds to Phase 3 with scaffolding only

### Phase 1: GraphQL Spec Search

Step 1.1 now searches for GraphQL schemas alongside OpenAPI specs:
- `"<API name>" graphql schema site:github.com`
- `"<API name>" graphql introspection schema SDL`
- Fetch schema from API docs or introspection endpoint

### Phase 4: Hand-Written Commands

For GraphQL APIs, Phase 4 writes ALL commands by hand using the GraphQL schema + Phase 0.5 workflows + Phase 0.7 data layer. Workflow commands prioritized over CRUD wrappers.

### Anti-Shortcut Rule Updated

Old: "The API is GraphQL-only but I'll write a REST spec anyway" (STOP)
New: "The API is GraphQL-only so we can't use printing-press" (Wrong. Skip REST generator, hand-write with GraphQL client.)

### Known-Specs Registry

Linear entry updated from "Skipped" to GraphQL SDL schema URL.

## Acceptance Criteria

- [x] Phase 2 handles GraphQL APIs without stopping
- [x] GraphQL client wrapper generated in scaffolding
- [x] Phase 1 searches for GraphQL schemas
- [x] Phase 4 instructions note GraphQL hand-writing
- [x] Anti-shortcut rule updated
- [x] Linear in known-specs registry with GraphQL schema URL
