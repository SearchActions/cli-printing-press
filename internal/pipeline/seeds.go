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
status: seed
pipeline_phase: preflight
pipeline_api: {{.APIName}}
date: {{now}}
---

# Phase Goal

Verify the local environment and source inputs needed to run the {{.APIName}} CLI pipeline.

## Context

- Pipeline directory: {{.PipelineDir}}
- Output directory: {{.OutputDir}}
- Spec URL: {{.SpecURL}}
- Spec source: {{.SpecSource}}

## What This Phase Must Produce

- Verified Go environment for the pipeline run
- Verified printing-press binary for local generation work
- Downloaded and validated OpenAPI spec for {{.APIName}}
- conventions.json in {{.PipelineDir}}

## Prior Phase Outputs

None.

## Codebase Pointers

- Build entrypoint: go build ./cmd/printing-press
- OpenAPI parsing: internal/openapi/parser.go
- Pipeline discovery flow: internal/pipeline/discover.go
`,
	PhaseScaffold: `---
title: "{{.APIName}} CLI Pipeline - Phase 1: Scaffold"
type: feat
status: seed
pipeline_phase: scaffold
pipeline_api: {{.APIName}}
date: {{now}}
---

# Phase Goal

Generate the first working {{.APIName}} CLI from the validated OpenAPI spec.

## Context

- Pipeline directory: {{.PipelineDir}}
- Output directory: {{.OutputDir}}
- Spec URL: {{.SpecURL}}
- Spec source: {{.SpecSource}}

## What This Phase Must Produce

- Generated CLI source tree in {{.OutputDir}}
- All seven generator quality gates passing
- Working CLI binary for {{.APIName}}

## Prior Phase Outputs

- conventions.json from preflight in {{.PipelineDir}}
- Validated spec URL and downloaded spec source for generation

## Codebase Pointers

- Generator entrypoint: printing-press generate --spec <url> --output <dir>
- Generator implementation: internal/generator/
- Quality gate logic in the generator flow under internal/generator/
`,
	PhaseEnrich: `---
title: "{{.APIName}} CLI Pipeline - Phase 2: Enrich"
type: feat
status: seed
pipeline_phase: enrich
pipeline_api: {{.APIName}}
date: {{now}}
---

# Phase Goal

Produce a focused overlay that captures useful spec enrichments missing from the original generation pass.

## Context

- Pipeline directory: {{.PipelineDir}}
- Output directory: {{.OutputDir}}
- Spec URL: {{.SpecURL}}
- Spec source: {{.SpecSource}}

## What This Phase Must Produce

- overlay.yaml in {{.PipelineDir}}
- At least one verified enrichment for the source spec
- Overlay content that is valid for downstream merge and regeneration

## Prior Phase Outputs

- conventions.json from preflight in {{.PipelineDir}}
- Scaffold-generated CLI in {{.OutputDir}}

## Codebase Pointers

- Overlay model and helpers: internal/pipeline/overlay.go
- Overlay merge preparation: internal/pipeline/merge.go
- Source spec artifact downloaded during preflight
`,
	PhaseRegenerate: `---
title: "{{.APIName}} CLI Pipeline - Phase 3: Regenerate"
type: feat
status: seed
pipeline_phase: regenerate
pipeline_api: {{.APIName}}
date: {{now}}
---

# Phase Goal

Merge the enrichments into the source spec and regenerate the CLI without losing quality.

## Context

- Pipeline directory: {{.PipelineDir}}
- Output directory: {{.OutputDir}}
- Spec URL: {{.SpecURL}}
- Spec source: {{.SpecSource}}

## What This Phase Must Produce

- Re-generated CLI in {{.OutputDir}} using the merged overlay
- Merged spec artifact suitable for regeneration
- All seven quality gates still passing after regeneration

## Prior Phase Outputs

- overlay.yaml from enrich in {{.PipelineDir}}
- Original scaffolded CLI in {{.OutputDir}}

## Codebase Pointers

- Overlay merge implementation: internal/pipeline/merge.go
- MergeOverlay function in internal/pipeline/merge.go
- Generator entrypoint: printing-press generate
`,
	PhaseReview: `---
title: "{{.APIName}} CLI Pipeline - Phase 4: Review"
type: feat
status: seed
pipeline_phase: review
pipeline_api: {{.APIName}}
date: {{now}}
---

# Phase Goal

Evaluate the generated CLI with static scoring and dogfooding evidence that determines ship readiness.

## Context

- Pipeline directory: {{.PipelineDir}}
- Output directory: {{.OutputDir}}
- Spec URL: {{.SpecURL}}
- Spec source: {{.SpecSource}}
- Sandbox note: petstore is sandbox-safe for Tier 3 dogfooding

## What This Phase Must Produce

- Static quality score from 0 to 50
- Dogfood score from 0 to 50
- review.md in {{.PipelineDir}}
- dogfood-results.json in {{.PipelineDir}}

## Prior Phase Outputs

- Working CLI binary from regenerate, or scaffold if regenerate was skipped

## Codebase Pointers

- Review scoring rules defined by the pipeline review plan for this phase
- Dogfood model uses three tiers: Tier 1 no credentials, Tier 2 read-only, Tier 3 sandbox write
- Generated CLI binary and help surfaces in {{.OutputDir}}
`,
	PhaseShip: `---
title: "{{.APIName}} CLI Pipeline - Phase 5: Ship"
type: feat
status: seed
pipeline_phase: ship
pipeline_api: {{.APIName}}
date: {{now}}
---

# Phase Goal

Package the generated CLI output and produce the final handoff report for humans.

## Context

- Pipeline directory: {{.PipelineDir}}
- Output directory: {{.OutputDir}}
- Spec URL: {{.SpecURL}}
- Spec source: {{.SpecSource}}

## What This Phase Must Produce

- Git repository initialized in {{.OutputDir}}
- Morning report written in {{.PipelineDir}}

## Prior Phase Outputs

- Review score and review.md from the review phase
- Working CLI binary ready for packaging and handoff

## Codebase Pointers

- Output CLI tree in {{.OutputDir}}
- Review artifacts in {{.PipelineDir}}
- Morning report format from SKILL.md Workflow 4 Step 6
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
