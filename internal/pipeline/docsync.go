package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type novelFeatureDocGroup struct {
	Name     string
	Features []NovelFeature
}

// SyncCLITranscendenceDocs rewrites generated README/SKILL transcendence
// blocks from dogfood-verified features. Empty verified sets remove the blocks.
func SyncCLITranscendenceDocs(dir string, features []NovelFeature) error {
	if err := syncMarkdownFeatureSection(
		filepath.Join(dir, "README.md"),
		"## Unique Features",
		renderNovelFeatureDocSection("## Unique Features", features),
		[]string{"## Usage"},
	); err != nil {
		return err
	}

	return syncMarkdownFeatureSection(
		filepath.Join(dir, "SKILL.md"),
		"## Unique Capabilities",
		renderNovelFeatureDocSection("## Unique Capabilities", features),
		[]string{"## HTTP Transport", "## Discovery Signals", "## Command Reference", "## Auth Setup"},
	)
}

func syncMarkdownFeatureSection(path, heading, replacement string, insertBefore []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading %s: %w", path, err)
	}

	updated := replaceMarkdownSection(string(data), heading, replacement, insertBefore)
	if updated == string(data) {
		return nil
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func renderNovelFeatureDocSection(heading string, features []NovelFeature) string {
	if len(features) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(heading)
	b.WriteString("\n\nThese capabilities aren't available in any other tool for this API.\n")

	if groups := groupNovelFeaturesForDocs(features); len(groups) > 0 {
		for _, group := range groups {
			b.WriteString("\n### ")
			b.WriteString(group.Name)
			b.WriteString("\n")
			for _, feature := range group.Features {
				writeNovelFeatureDocLine(&b, feature)
			}
		}
	} else {
		for _, feature := range features {
			writeNovelFeatureDocLine(&b, feature)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func writeNovelFeatureDocLine(b *strings.Builder, feature NovelFeature) {
	b.WriteString("- **`")
	b.WriteString(feature.Command)
	b.WriteString("`** \u2014 ")
	b.WriteString(feature.Description)
	b.WriteString("\n")
	if feature.WhyItMatters != "" {
		b.WriteString("\n  _")
		b.WriteString(feature.WhyItMatters)
		b.WriteString("_\n")
	}
	if feature.Example != "" {
		b.WriteString("\n  ```bash\n  ")
		b.WriteString(feature.Example)
		b.WriteString("\n  ```\n")
	}
}

func groupNovelFeaturesForDocs(features []NovelFeature) []novelFeatureDocGroup {
	canonGroup := func(s string) string {
		return strings.Join(strings.Fields(strings.ToLower(s)), " ")
	}

	anyGrouped := false
	for _, feature := range features {
		if canonGroup(feature.Group) != "" {
			anyGrouped = true
			break
		}
	}
	if !anyGrouped {
		return nil
	}

	order := []string{}
	displayName := map[string]string{}
	byGroup := map[string][]NovelFeature{}
	for _, feature := range features {
		display := feature.Group
		key := canonGroup(display)
		if key == "" {
			key = "more"
			display = "More"
		}
		if _, seen := byGroup[key]; !seen {
			order = append(order, key)
			displayName[key] = display
		}
		byGroup[key] = append(byGroup[key], feature)
	}

	out := make([]novelFeatureDocGroup, 0, len(order))
	for _, key := range order {
		out = append(out, novelFeatureDocGroup{Name: displayName[key], Features: byGroup[key]})
	}
	return out
}

func replaceMarkdownSection(content, heading, replacement string, insertBefore []string) string {
	start := findMarkdownHeading(content, heading)
	if start >= 0 {
		end := findNextLevelTwoHeading(content, start+len(heading))
		return joinMarkdownParts(content[:start], replacement, content[end:])
	}

	if strings.TrimSpace(replacement) == "" {
		return content
	}

	insertAt := -1
	for _, anchor := range insertBefore {
		if idx := findMarkdownHeading(content, anchor); idx >= 0 && (insertAt == -1 || idx < insertAt) {
			insertAt = idx
		}
	}
	if insertAt == -1 {
		return joinMarkdownParts(content, replacement, "")
	}
	return joinMarkdownParts(content[:insertAt], replacement, content[insertAt:])
}

func joinMarkdownParts(prefix, middle, suffix string) string {
	prefix = strings.TrimRight(prefix, "\n")
	middle = strings.Trim(middle, "\n")
	suffix = strings.TrimLeft(suffix, "\n")

	switch {
	case middle == "":
		if prefix == "" {
			if suffix == "" {
				return ""
			}
			return suffix
		}
		if suffix == "" {
			return prefix + "\n"
		}
		return prefix + "\n\n" + suffix
	case prefix == "" && suffix == "":
		return middle + "\n"
	case prefix == "":
		return middle + "\n\n" + suffix
	case suffix == "":
		return prefix + "\n\n" + middle + "\n"
	default:
		return prefix + "\n\n" + middle + "\n\n" + suffix
	}
}

func findMarkdownHeading(content, heading string) int {
	offset := 0
	for {
		idx := strings.Index(content[offset:], heading)
		if idx == -1 {
			return -1
		}
		idx += offset
		beforeOK := idx == 0 || content[idx-1] == '\n'
		after := idx + len(heading)
		afterOK := after == len(content) || content[after] == '\n' || content[after] == '\r'
		if beforeOK && afterOK {
			return idx
		}
		offset = idx + len(heading)
	}
}

func findNextLevelTwoHeading(content string, after int) int {
	if after >= len(content) {
		return len(content)
	}
	if idx := strings.Index(content[after:], "\n## "); idx >= 0 {
		return after + idx + 1
	}
	return len(content)
}
