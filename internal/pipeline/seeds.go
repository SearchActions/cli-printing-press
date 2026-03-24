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
Static quality analysis of the generated CLI. No API calls.

## Acceptance Criteria
- [ ] Quality score calculated (0-100)
- [ ] All critical checks pass
- [ ] Issue list written to review.md

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

### Unit 4: Scoring
- +20 points for compiles cleanly
- +20 points for all help commands work
- +20 points for no name quality issues
- +20 points for no empty descriptions
- +10 points for doctor works
- +10 points for binary < 50MB
- Write score and issues to {{.PipelineDir}}/review.md
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
