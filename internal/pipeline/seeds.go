package pipeline

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

// SeedData holds the context for rendering plan seeds.
type SeedData struct {
	APIName     string
	OutputDir   string
	SpecURL     string
	SpecSource  string
	PipelineDir string
}

var seedTemplates = map[string]string{
	PhasePreflight: `---
title: "{{.APIName}} CLI Pipeline - Phase 0: Preflight"
type: feat
status: active
date: {{now}}
---

# Preflight: {{.APIName}} CLI

## Goal
Validate the environment and discover the OpenAPI spec for {{.APIName}}.

## Acceptance Criteria
- [ ] Go is installed and working (go version succeeds)
- [ ] printing-press binary compiles (go build ./cmd/printing-press)
- [ ] OpenAPI spec downloaded and validated
- [ ] Spec has 3+ endpoints and a base URL
- [ ] Conventions cache written with auth type, resource count, pagination patterns

## Implementation Units

### Unit 1: Environment Check
- Run ` + "`go version`" + ` - verify Go 1.23+
- Run ` + "`go build -o /tmp/pp-check ./cmd/printing-press`" + ` - verify press compiles
- Check output dir {{.OutputDir}} doesn't exist (or --force was used)

### Unit 2: Spec Discovery
- Spec URL: {{.SpecURL}} (source: {{.SpecSource}})
- Download with: ` + "`printing-press generate --spec {{.SpecURL}} --output /dev/null 2>&1 | head -5`" + ` to test parsing
- If parse fails: try alternative sources

### Unit 3: Write Conventions Cache
- After successful parse, write conventions.json to {{.PipelineDir}}/
- Include: auth type, endpoint count, resource names, pagination patterns detected, global params found

## Context
Pipeline directory: {{.PipelineDir}}
`,
	PhaseScaffold: `---
title: "{{.APIName}} CLI Pipeline - Phase 1: Scaffold"
type: feat
status: active
date: {{now}}
---

# Scaffold: {{.APIName}} CLI

## Goal
Generate the initial CLI from the discovered OpenAPI spec.

## Acceptance Criteria
- [ ] CLI generated at {{.OutputDir}}
- [ ] All 7 quality gates pass (go mod tidy, go vet, go build, binary, --help, version, doctor)
- [ ] CLI compiles to a working binary

## Implementation Units

### Unit 1: Generate CLI
Run: ` + "`printing-press generate --spec <spec-path-from-preflight> --output {{.OutputDir}}`" + `

### Unit 2: Validate
- Build the binary: ` + "`cd {{.OutputDir}} && go build -o ./{{.APIName}}-cli ./cmd/{{.APIName}}-cli`" + `
- Run: ` + "`./{{.APIName}}-cli --help`" + `
- Run: ` + "`./{{.APIName}}-cli doctor`" + `
- Count resources and endpoints

### Unit 3: Document
- List all top-level resources
- List total endpoint count
- Note any warnings from generation
`,
	PhaseEnrich: `---
title: "{{.APIName}} CLI Pipeline - Phase 2: Enrich"
type: feat
status: active
date: {{now}}
---

# Enrich: {{.APIName}} CLI

## Goal
Deep-read the original spec for hints the parser missed. Research API docs. Produce a spec overlay.

## Acceptance Criteria
- [ ] overlay.yaml written to {{.PipelineDir}}/
- [ ] At least 1 enrichment found
- [ ] Overlay is valid YAML

## Implementation Units

### Unit 1: Description Hints
- Read every parameter description in the spec
- Find default value hints (e.g., "The special value 'me' can be used")
- Extract as ParamPatch with Default field

### Unit 2: Upload Detection
- Scan for mediaUpload fields, x-google extensions, multipart content types
- Flag endpoints that support file upload (note in overlay comments)

### Unit 3: Sync Token Patterns
- Scan response schemas for historyId, syncToken, nextSyncToken fields
- Note resources that support incremental sync

### Unit 4: Better Descriptions
- Find endpoints with empty or generic descriptions
- WebSearch for "{{.APIName}} API" to find better descriptions

### Unit 5: Write Overlay
- Compile all enrichments into overlay.yaml at {{.PipelineDir}}/overlay.yaml
`,
	PhaseRegenerate: `---
title: "{{.APIName}} CLI Pipeline - Phase 3: Regenerate"
type: feat
status: active
date: {{now}}
---

# Regenerate: {{.APIName}} CLI with Enrichments

## Goal
Merge the spec overlay with the original spec and re-generate the CLI.

## Acceptance Criteria
- [ ] Overlay merged with original spec
- [ ] CLI re-generated at {{.OutputDir}}
- [ ] All 7 quality gates still pass
- [ ] Enrichments visible in CLI help output

## Implementation Units

### Unit 1: Load and Merge
- Load original spec
- Load overlay.yaml from {{.PipelineDir}}/
- Apply overlay using MergeOverlay function
- Write merged spec to {{.PipelineDir}}/merged-spec.yaml

### Unit 2: Regenerate
- Run ` + "`printing-press generate --spec {{.PipelineDir}}/merged-spec.yaml --output {{.OutputDir}}`" + `
- If quality gates fail, fall back to original spec

### Unit 3: Verify Enrichments
- Check CLI help for enriched defaults (e.g., userId now shows default)
- Compare before/after help output
`,
	PhaseReview: `---
title: "{{.APIName}} CLI Pipeline - Phase 4: Review"
type: feat
status: active
date: {{now}}
---

# Review: {{.APIName}} CLI Quality

## Goal
Static quality analysis and autonomous dogfooding of the generated CLI.

## Acceptance Criteria
- [ ] Static quality score calculated (0-50)
- [ ] Dogfood score calculated (0-50)
- [ ] Combined score written to review.md (0-100)
- [ ] dogfood-results.json written with per-test results
- [ ] All critical static checks pass

## Implementation Units

### Unit 1: Help Completeness
- Run ` + "`{{.APIName}}-cli --help`" + ` and check exit code
- Run ` + "`{{.APIName}}-cli <resource> --help`" + ` for every top-level resource
- Verify non-empty output for each

### Unit 2: Name Quality
- No command name > 40 characters
- No raw operationId passthrough (no dots or underscores in command names)
- No duplicate command names

### Unit 3: Description Quality
- No empty descriptions on top-level resources
- No descriptions that just repeat the command name

### Unit 4: Static Scoring
- +10 compiles cleanly
- +10 all help commands work
- +10 no name quality issues
- +10 no empty descriptions
- +5 doctor works
- +5 binary < 50MB

### Unit 5: Dogfood Tier 1 - No Credentials Required
Build the binary and run zero-config tests:

a. Build: ` + "`cd {{.OutputDir}} && go build -o {{.APIName}}-cli ./cmd/{{.APIName}}-cli`" + `
b. Version: ` + "`./{{.APIName}}-cli version`" + ` - expect exit code 0
c. Doctor: ` + "`./{{.APIName}}-cli doctor`" + ` - expect runs without crash
d. Dry-run: pick first POST/PUT endpoint, run with --dry-run
e. Output modes: run any list command with --json, --plain, --quiet

Record all results to dogfood-results.json.

### Unit 6: Dogfood Tier 2 - Read-Only API Calls
Check for API credentials in env vars or config file.

If no credentials: print "Tier 2 skipped: no credentials" and score as N/A.

If credentials found:
a. List: ` + "`./{{.APIName}}-cli <resource> list --limit 5`" + `
b. Get: ` + "`./{{.APIName}}-cli <resource> get <id-from-list>`" + `
c. Auth error: run with invalid credentials, expect exit code 4
d. Record response shape, latency, exit codes

### Unit 7: Dogfood Tier 3 - Write Operations (Sandbox APIs Only)
Only run on sandbox-safe APIs (petstore, stripe test mode, stytch test).
If not sandbox-safe: skip entirely.

If sandbox-safe:
a. Create test resource
b. Read it back
c. Delete it (cleanup)
d. Verify deletion returns exit code 3

### Unit 8: Combined Scoring
Write dogfood-results.json with per-test results.
Write review.md with combined score:

Static (0-50): from Unit 4
Dogfood (0-50):
  +10 version prints correctly
  +10 doctor reports API status
  +10 list/get returns data (Tier 2, 0 if skipped)
  +10 create/delete roundtrip (Tier 3, 0 if skipped)
  +5 dry-run works
  +5 auth error handling correct

Grade: A (90+), B (75+), C (60+), D (40+), F (<40)
`,
	PhaseShip: `---
title: "{{.APIName}} CLI Pipeline - Phase 5: Ship"
type: feat
status: active
date: {{now}}
---

# Ship: {{.APIName}} CLI

## Goal
Finalize the CLI for human use.

## Acceptance Criteria
- [ ] Git repo initialized in {{.OutputDir}}
- [ ] Initial commit created
- [ ] Morning report written to {{.PipelineDir}}/report.md

## Implementation Units

### Unit 1: Git Init
- Run ` + "`cd {{.OutputDir}} && git init && git add -A && git commit -m 'Initial CLI generated by printing-press'`" + `

### Unit 2: Morning Report
Write {{.PipelineDir}}/report.md with:
- API name and spec source
- Resource and endpoint count
- Quality score from review phase
- Enrichments applied
- Time per phase
- Next steps: configure auth, test against real API, publish

### Unit 3: Summary
Print to stderr:
- "{{.APIName}}-cli ready at {{.OutputDir}}/"
- Resource list
- Auth type
- Total pipeline time
`,
}

// RenderSeed renders a plan seed template for the given phase.
func RenderSeed(phase string, data SeedData) (string, error) {
	tmplStr, ok := seedTemplates[phase]
	if !ok {
		return "", fmt.Errorf("no seed template for phase %q", phase)
	}

	tmpl, err := template.New(phase).Funcs(template.FuncMap{
		"now": func() string {
			return time.Now().Format("2006-01-02")
		},
	}).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parsing seed template for %s: %w", phase, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering seed template for %s: %w", phase, err)
	}

	return buf.String(), nil
}
