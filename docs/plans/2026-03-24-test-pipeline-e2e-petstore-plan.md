---
title: "E2E Test - Run Pipeline on Petstore"
type: test
status: active
date: 2026-03-24
---

# E2E Test - Run Pipeline on Petstore

## Overview

Run `printing-press print petstore` end-to-end. Verify the pipeline directory is created with plan seeds, state.json tracks all phases, and then manually execute each phase's plan to verify the full pipeline produces a working CLI.

Petstore is the simplest spec (3 resources, 13 endpoints, no OAuth2) - ideal for validating the pipeline machinery without complexity from the API itself.

## Acceptance Criteria

- [ ] `printing-press print petstore` creates `docs/plans/petstore-pipeline/` with 6 plan seeds
- [ ] `state.json` has all 6 phases with status "planned" and correct plan paths
- [ ] Each plan seed has proper frontmatter (title, type, status, date)
- [ ] Each plan seed has acceptance criteria and implementation units
- [ ] Plan seeds reference correct spec URL (petstore3.swagger.io)
- [ ] Preflight plan can be executed: Go check passes, spec downloads, conventions cached
- [ ] Scaffold plan can be executed: CLI generates, 7 quality gates pass
- [ ] Generated petstore-cli compiles and `--help` works
- [ ] Generated petstore-cli `doctor` runs
- [ ] Pipeline cleanup: remove test artifacts after verification

## Implementation Units

### Unit 1: Build and Run Print Command

**Approach:**
```bash
cd ~/cli-printing-press
go build -o ./printing-press ./cmd/printing-press
./printing-press print petstore --output /tmp/petstore-e2e-test
```

**Verify:**
- Directory `docs/plans/petstore-pipeline/` exists
- Contains: `00-preflight-plan.md` through `05-ship-plan.md` (6 files)
- Contains: `state.json`
- No errors in output

### Unit 2: Validate State File

**Approach:**
```bash
cat docs/plans/petstore-pipeline/state.json | python3 -m json.tool
```

**Verify:**
- `api_name` is "petstore"
- `spec_url` is the Petstore swagger URL
- All 6 phases present with status "planned"
- Each phase has a `plan_path` pointing to the correct file

### Unit 3: Validate Plan Seed Quality

**Approach:**
Read each plan seed and verify:

```bash
for f in docs/plans/petstore-pipeline/*.md; do
  echo "=== $(basename $f) ==="
  head -10 "$f"
  grep -c "Acceptance Criteria\|Implementation Units" "$f"
  echo ""
done
```

**Verify:**
- Each plan has YAML frontmatter with title, type: feat, status: active, date
- Each plan has `## Acceptance Criteria` section with checkboxes
- Each plan has `## Implementation Units` section
- Preflight plan references the Petstore spec URL
- Scaffold plan references the output directory

### Unit 4: Execute Preflight Phase Manually

**Approach:**
Execute the preflight plan's acceptance criteria by hand:

```bash
# Check Go
go version

# Check press compiles
go build -o /tmp/pp-check ./cmd/printing-press

# Download and verify spec
curl -sL "https://petstore3.swagger.io/api/v3/openapi.json" -o /tmp/petstore-spec.json
head -5 /tmp/petstore-spec.json  # Should show openapi: 3.x
```

**Verify:**
- Go 1.23+ installed
- Press binary compiles
- Spec downloads and starts with `openapi` or `"openapi"`

### Unit 5: Execute Scaffold Phase Manually

**Approach:**
```bash
./printing-press generate --spec "https://petstore3.swagger.io/api/v3/openapi.json" --output /tmp/petstore-e2e-test
```

**Verify:**
- All 7 quality gates pass (PASS in output)
- CLI generated at /tmp/petstore-e2e-test

### Unit 6: Dogfood the Generated CLI

**Approach:**
```bash
cd /tmp/petstore-e2e-test
go build -o petstore-cli ./cmd/petstore-cli
./petstore-cli --help
./petstore-cli pet --help
./petstore-cli store --help
./petstore-cli user --help
./petstore-cli doctor
./petstore-cli version
```

**Verify:**
- Binary compiles
- `--help` shows pet, store, user resources
- `--no-color` flag present (from GOAT color work)
- `doctor` runs without crash
- `version` prints version

### Unit 7: Cleanup

**Approach:**
```bash
rm -rf docs/plans/petstore-pipeline/
rm -rf /tmp/petstore-e2e-test
rm -f /tmp/petstore-spec.json
```

## Scope Boundaries

- Don't run ce:plan or ce:work on the plan seeds (that tests the full pipeline, not the infrastructure)
- Don't test the nightnight chaining (that's a separate manual test after the skill is updated)
- Don't test with Gmail or other complex specs (Petstore is sufficient for infrastructure validation)
- Don't modify any code - this is pure verification

## Sources

- Pipeline init: `internal/pipeline/pipeline.go:Init()` - creates the pipeline directory
- Plan seeds: `internal/pipeline/seeds.go` - seed templates per phase
- State types: `internal/pipeline/state.go` - PipelineState, PhaseOrder
- Discovery: `internal/pipeline/discover.go` - KnownSpecs registry with Petstore URL
