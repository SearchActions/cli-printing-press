package vision

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteReport(plan *VisionaryPlan, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Visionary Research: %s CLI\n\n", plan.APIName))

	// API Identity
	b.WriteString("## API Identity\n\n")
	b.WriteString(fmt.Sprintf("- **Domain:** %s\n", plan.Identity.DomainCategory))
	b.WriteString(fmt.Sprintf("- **Primary users:** %s\n", strings.Join(plan.Identity.PrimaryUsers, ", ")))
	b.WriteString(fmt.Sprintf("- **Core entities:** %s\n", strings.Join(plan.Identity.CoreEntities, ", ")))
	b.WriteString(fmt.Sprintf("- **Data profile:** %s volume, %s writes, realtime=%v, search need=%s\n\n",
		plan.Identity.DataProfile.Volume,
		plan.Identity.DataProfile.WritePattern,
		plan.Identity.DataProfile.Realtime,
		plan.Identity.DataProfile.SearchNeed))

	// Usage Patterns
	if len(plan.UsagePatterns) > 0 {
		b.WriteString("## Usage Patterns (by Evidence)\n\n")
		for i, p := range plan.UsagePatterns {
			b.WriteString(fmt.Sprintf("### %d. %s (Evidence: %d/10)\n\n", i+1, p.Name, p.EvidenceScore))
			b.WriteString(p.Description + "\n\n")
			if len(p.EvidenceSources) > 0 {
				for _, src := range p.EvidenceSources {
					b.WriteString(fmt.Sprintf("- %s\n", src))
				}
				b.WriteString("\n")
			}
			if len(p.Requirements) > 0 {
				b.WriteString("**Requires:** " + strings.Join(p.Requirements, ", ") + "\n\n")
			}
		}
	}

	// Tool Landscape
	if len(plan.ToolLandscape) > 0 {
		b.WriteString("## Tool Landscape\n\n")
		b.WriteString("| Tool | Stars | Type | Language | Features |\n")
		b.WriteString("|------|-------|------|----------|----------|\n")
		for _, t := range plan.ToolLandscape {
			b.WriteString(fmt.Sprintf("| [%s](%s) | %d | %s | %s | %s |\n",
				t.Name, t.URL, t.Stars, t.ToolType, t.Language,
				strings.Join(t.Features, ", ")))
		}
		b.WriteString("\n")
	}

	// Workflows
	if len(plan.Workflows) > 0 {
		b.WriteString("## Workflows\n\n")
		for i, w := range plan.Workflows {
			b.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, w.Name))
			if len(w.Steps) > 0 {
				b.WriteString("**Steps:** ")
				stepDescs := make([]string, len(w.Steps))
				for j, s := range w.Steps {
					stepDescs[j] = s.Description
				}
				b.WriteString(strings.Join(stepDescs, " -> ") + "\n")
			}
			if w.Frequency != "" {
				b.WriteString(fmt.Sprintf("**Frequency:** %s\n", w.Frequency))
			}
			if w.PainPoint != "" {
				b.WriteString(fmt.Sprintf("**Pain:** %s\n", w.PainPoint))
			}
			if w.ProposedCLIFeature != "" {
				b.WriteString(fmt.Sprintf("**Proposed:** `%s`\n", w.ProposedCLIFeature))
			}
			b.WriteString("\n")
		}
	}

	// Architecture Decisions
	if len(plan.Architecture) > 0 {
		b.WriteString("## Architecture Decisions\n\n")
		b.WriteString("| Area | Need | Decision | Rationale |\n")
		b.WriteString("|------|------|----------|----------|\n")
		for _, a := range plan.Architecture {
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				a.Area, a.NeedLevel, a.Decision, a.Rationale))
		}
		b.WriteString("\n")
	}

	// Top 5 Features
	if len(plan.Features) > 0 {
		b.WriteString("## Top Features for the World\n\n")
		b.WriteString("| # | Feature | Score | Evidence | Impact | Templates |\n")
		b.WriteString("|---|---------|-------|----------|--------|----------|\n")
		for i, f := range plan.Features {
			f.TotalScore = f.ComputeScore()
			b.WriteString(fmt.Sprintf("| %d | %s | %d/16 | %d/3 | %d/3 | %s |\n",
				i+1, f.Name, f.TotalScore,
				f.EvidenceStrength, f.UserImpact,
				strings.Join(f.TemplateNames, ", ")))
		}
		b.WriteString("\n")

		// Detail for each feature
		for i, f := range plan.Features {
			f.TotalScore = f.ComputeScore()
			b.WriteString(fmt.Sprintf("### %d. %s (Score: %d/16)\n\n", i+1, f.Name, f.TotalScore))
			b.WriteString(f.Description + "\n\n")
			if len(f.TemplateNames) > 0 {
				b.WriteString("**Templates:** " + strings.Join(f.TemplateNames, ", ") + "\n\n")
			}
		}
	}

	path := filepath.Join(outputDir, "visionary-research.md")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
