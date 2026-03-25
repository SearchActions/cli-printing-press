package generator

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/mvanhorn/cli-printing-press/internal/spec"
)

//go:embed templates
var templateFS embed.FS

type Generator struct {
	Spec      *spec.APISpec
	OutputDir string
	funcs     template.FuncMap
}

func New(s *spec.APISpec, outputDir string) *Generator {
	g := &Generator{Spec: s, OutputDir: outputDir}
	g.funcs = template.FuncMap{
		"title":             strings.Title,
		"lower":             strings.ToLower,
		"upper":             strings.ToUpper,
		"camel":             toCamel,
		"snake":             toSnake,
		"goType":            goType,
		"cobraFlagFunc":     cobraFlagFunc,
		"defaultVal":        defaultVal,
		"zeroVal":           zeroVal,
		"positionalArgs":    positionalArgs,
		"configTag":         configTag,
		"envVarField":       envVarField,
		"envVarPlaceholder": envVarPlaceholder,
		"add":               func(a, b int) int { return a + b },
		"oneline":           oneline,
		"flagName":          flagName,
		"safeTypeName":      safeTypeName,
		"exampleLine":       g.exampleLine,
	}
	return g
}

func (g *Generator) Generate() error {
	dirs := []string{
		filepath.Join("cmd", g.Spec.Name+"-cli"),
		filepath.Join("internal", "cli"),
		filepath.Join("internal", "client"),
		filepath.Join("internal", "config"),
		filepath.Join("internal", "types"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(g.OutputDir, d), 0755); err != nil {
			return fmt.Errorf("creating dir %s: %w", d, err)
		}
	}

	// Generate single files
	singleFiles := map[string]string{
		"main.go.tmpl":         filepath.Join("cmd", g.Spec.Name+"-cli", "main.go"),
		"root.go.tmpl":         filepath.Join("internal", "cli", "root.go"),
		"helpers.go.tmpl":      filepath.Join("internal", "cli", "helpers.go"),
		"doctor.go.tmpl":       filepath.Join("internal", "cli", "doctor.go"),
		"config.go.tmpl":       filepath.Join("internal", "config", "config.go"),
		"client.go.tmpl":       filepath.Join("internal", "client", "client.go"),
		"types.go.tmpl":        filepath.Join("internal", "types", "types.go"),
		"go.mod.tmpl":          "go.mod",
		"goreleaser.yaml.tmpl": ".goreleaser.yaml",
		"golangci.yml.tmpl":    ".golangci.yml",
		"makefile.tmpl":        "Makefile",
		"readme.md.tmpl":       "README.md",
	}

	for tmplName, outPath := range singleFiles {
		if err := g.renderTemplate(tmplName, outPath, g.Spec); err != nil {
			return fmt.Errorf("rendering %s: %w", tmplName, err)
		}
	}

	// Generate per-resource command files
	for name, resource := range g.Spec.Resources {
		data := struct {
			ResourceName string
			FuncPrefix   string
			CommandPath  string
			Resource     spec.Resource
			*spec.APISpec
		}{
			ResourceName: name,
			FuncPrefix:   name,
			CommandPath:  name,
			Resource:     resource,
			APISpec:      g.Spec,
		}
		outPath := filepath.Join("internal", "cli", name+".go")
		if err := g.renderTemplate("command.go.tmpl", outPath, data); err != nil {
			return fmt.Errorf("rendering command %s: %w", name, err)
		}

		// Generate sub-resource command files
		for subName, subResource := range resource.SubResources {
			subData := struct {
				ResourceName string
				FuncPrefix   string
				CommandPath  string
				Resource     spec.Resource
				*spec.APISpec
			}{
				ResourceName: subName,
				FuncPrefix:   name + "-" + subName,
				CommandPath:  name + " " + subName,
				Resource:     subResource,
				APISpec:      g.Spec,
			}
			subOutPath := filepath.Join("internal", "cli", name+"_"+subName+".go")
			if err := g.renderTemplate("command.go.tmpl", subOutPath, subData); err != nil {
				return fmt.Errorf("rendering sub-command %s/%s: %w", name, subName, err)
			}
		}
	}

	// Conditionally render auth command when OAuth2 is detected
	if g.Spec.Auth.AuthorizationURL != "" {
		authPath := filepath.Join("internal", "cli", "auth.go")
		if err := g.renderTemplate("auth.go.tmpl", authPath, g.Spec); err != nil {
			return fmt.Errorf("rendering auth: %w", err)
		}
	}

	return nil
}

func (g *Generator) renderTemplate(tmplName, outPath string, data any) error {
	content, err := templateFS.ReadFile(filepath.Join("templates", tmplName))
	if err != nil {
		return fmt.Errorf("reading template %s: %w", tmplName, err)
	}

	tmpl, err := template.New(tmplName).Funcs(g.funcs).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", tmplName, err)
	}

	fullPath := filepath.Join(g.OutputDir, outPath)
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", fullPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template %s: %w", tmplName, err)
	}

	return nil
}

// Template helper functions

func toCamel(s string) string {
	// Strip characters that are invalid in Go identifiers
	s = strings.TrimLeft(s, "$")
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	result := strings.Join(parts, "")
	// Ensure starts with letter
	if len(result) > 0 && !unicode.IsLetter(rune(result[0])) {
		result = "V" + result
	}
	return result
}

func toSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func goType(t string) string {
	switch t {
	case "string":
		return "string"
	case "int":
		return "int"
	case "bool":
		return "bool"
	case "float":
		return "float64"
	default:
		return "string"
	}
}

func cobraFlagFunc(t string) string {
	switch t {
	case "string":
		return "StringVar"
	case "int":
		return "IntVar"
	case "bool":
		return "BoolVar"
	case "float":
		return "Float64Var"
	default:
		return "StringVar"
	}
}

func defaultVal(p spec.Param) string {
	if p.Default != nil {
		// Coerce the default value to match the declared param type
		switch p.Type {
		case "string":
			return fmt.Sprintf("%q", fmt.Sprintf("%v", p.Default))
		case "bool":
			switch v := p.Default.(type) {
			case bool:
				return fmt.Sprintf("%t", v)
			case string:
				if v == "true" || v == "false" {
					return v
				}
			}
			return "false"
		case "int":
			switch v := p.Default.(type) {
			case float64:
				return fmt.Sprintf("%d", int(v))
			case int:
				return fmt.Sprintf("%d", v)
			}
			return "0"
		case "float":
			switch v := p.Default.(type) {
			case float64:
				return fmt.Sprintf("%f", v)
			case int:
				return fmt.Sprintf("%f", float64(v))
			}
			return "0.0"
		}
	}
	return zeroVal(p.Type)
}

func zeroVal(t string) string {
	switch t {
	case "string":
		return `""`
	case "int":
		return "0"
	case "bool":
		return "false"
	case "float":
		return "0.0"
	default:
		return `""`
	}
}

func positionalArgs(e spec.Endpoint) string {
	var args []string
	for _, p := range e.Params {
		if p.Positional {
			args = append(args, "<"+p.Name+">")
		}
	}
	if len(args) > 0 {
		return " " + strings.Join(args, " ")
	}
	return ""
}

func configTag(format string) string {
	switch format {
	case "toml":
		return "toml"
	case "yaml":
		return "yaml"
	default:
		return "json"
	}
}

func envVarField(envVar string) string {
	// STYTCH_PROJECT_ID -> ProjectID
	parts := strings.Split(strings.ToLower(envVar), "_")
	var result string
	for _, p := range parts {
		if len(p) > 0 {
			result += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return result
}

func oneline(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, `"`, `'`)
	s = strings.ReplaceAll(s, "\\", "")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	s = strings.TrimSpace(s)
	if len(s) > 120 {
		cut := s[:117]
		if idx := strings.LastIndex(cut, " "); idx > 60 {
			s = cut[:idx] + "..."
		} else {
			s = cut + "..."
		}
	}
	return s
}

func (g *Generator) exampleLine(commandPath, endpointName string, endpoint spec.Endpoint) string {
	var parts []string
	parts = append(parts, g.Spec.Name+"-cli")
	parts = append(parts, strings.Fields(commandPath)...)
	parts = append(parts, endpointName)

	// Add positional arg placeholders
	for _, p := range endpoint.Params {
		if p.Positional {
			parts = append(parts, "<"+p.Name+">")
		}
	}

	// Add a sample flag for POST/PUT/PATCH
	switch endpoint.Method {
	case "POST", "PUT", "PATCH":
		for _, p := range endpoint.Body {
			if p.Required && p.Type == "string" {
				parts = append(parts, "--"+strings.ReplaceAll(p.Name, "_", "-"), "value")
				break
			}
		}
	}

	return "  " + strings.Join(parts, " ")
}

func flagName(name string) string {
	name = strings.TrimLeft(name, "$")
	// Replace common separators with hyphens, strip anything not alphanumeric or hyphen
	var b strings.Builder
	lastHyphen := true
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			lastHyphen = false
		} else if !lastHyphen && b.Len() > 0 {
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func safeTypeName(name string) string {
	name = strings.TrimLeft(name, "$")
	name = strings.NewReplacer(".", "_", "/", "_", "\\", "_", "-", "_", " ", "_").Replace(name)
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if len(result) > 0 && !unicode.IsLetter(rune(result[0])) {
		result = "T" + result
	}
	return result
}

func envVarPlaceholder(envVar string) string {
	// STYTCH_PROJECT_ID -> project_id (the placeholder in the format string)
	parts := strings.Split(envVar, "_")
	if len(parts) <= 1 {
		return strings.ToLower(envVar)
	}
	// Skip the first part (tool name prefix) and join the rest
	var lower []string
	for _, p := range parts[1:] {
		lower = append(lower, strings.ToLower(p))
	}
	return strings.Join(lower, "_")
}
