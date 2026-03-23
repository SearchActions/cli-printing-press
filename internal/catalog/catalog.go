package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var namePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var validCategories = map[string]struct{}{
	"auth":               {},
	"payments":           {},
	"email":              {},
	"developer-tools":    {},
	"project-management": {},
	"communication":      {},
	"crm":                {},
	"example":            {},
}

var validSpecFormats = map[string]struct{}{
	"yaml": {},
	"json": {},
}

var validTiers = map[string]struct{}{
	"official":  {},
	"community": {},
}

type Entry struct {
	Name           string `yaml:"name"`
	DisplayName    string `yaml:"display_name"`
	Description    string `yaml:"description"`
	Category       string `yaml:"category"`
	SpecURL        string `yaml:"spec_url"`
	SpecFormat     string `yaml:"spec_format"`
	OpenAPIVersion string `yaml:"openapi_version"`
	Tier           string `yaml:"tier"`
	VerifiedDate   string `yaml:"verified_date"`
	Homepage       string `yaml:"homepage"`
	Notes          string `yaml:"notes"`
}

func ParseEntry(data []byte) (*Entry, error) {
	var e Entry
	if err := yaml.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}
	if err := e.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	return &e, nil
}

func ParseDir(dir string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	entries := make([]Entry, 0, len(dirEntries))
	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		if filepath.Ext(de.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(dir, de.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", de.Name(), err)
		}

		entry, err := ParseEntry(data)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", de.Name(), err)
		}
		entries = append(entries, *entry)
	}

	return entries, nil
}

func (e *Entry) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !namePattern.MatchString(e.Name) {
		return fmt.Errorf("name must be lowercase kebab-case (letters, digits, hyphens only)")
	}
	if e.DisplayName == "" {
		return fmt.Errorf("display_name is required")
	}
	if e.Description == "" {
		return fmt.Errorf("description is required")
	}
	if e.Category == "" {
		return fmt.Errorf("category is required")
	}
	if _, ok := validCategories[e.Category]; !ok {
		return fmt.Errorf("category must be one of: auth, payments, email, developer-tools, project-management, communication, crm, example")
	}
	if e.SpecURL == "" {
		return fmt.Errorf("spec_url is required")
	}
	if !strings.HasPrefix(e.SpecURL, "https://") {
		return fmt.Errorf(`spec_url must start with "https://"`)
	}
	if e.SpecFormat == "" {
		return fmt.Errorf("spec_format is required")
	}
	if _, ok := validSpecFormats[e.SpecFormat]; !ok {
		return fmt.Errorf("spec_format must be one of: yaml, json")
	}
	if e.Tier == "" {
		return fmt.Errorf("tier is required")
	}
	if _, ok := validTiers[e.Tier]; !ok {
		return fmt.Errorf("tier must be one of: official, community")
	}

	return nil
}
