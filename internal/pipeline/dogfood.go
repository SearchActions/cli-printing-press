package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type DogfoodReport struct {
	Dir           string          `json:"dir"`
	SpecPath      string          `json:"spec_path,omitempty"`
	Verdict       string          `json:"verdict"`
	PathCheck     PathCheckResult `json:"path_check"`
	AuthCheck     AuthCheckResult `json:"auth_check"`
	DeadFlags     DeadCodeResult  `json:"dead_flags"`
	DeadFuncs     DeadCodeResult  `json:"dead_functions"`
	PipelineCheck PipelineResult  `json:"pipeline_check"`
	Issues        []string        `json:"issues"`
}

type PathCheckResult struct {
	Tested  int      `json:"tested"`
	Valid   int      `json:"valid"`
	Invalid []string `json:"invalid,omitempty"`
	Pct     int      `json:"valid_pct"`
}

type AuthCheckResult struct {
	SpecScheme   string `json:"spec_scheme"`
	GeneratedFmt string `json:"generated_format"`
	Match        bool   `json:"match"`
	Detail       string `json:"detail"`
}

type DeadCodeResult struct {
	Total int      `json:"total"`
	Dead  int      `json:"dead"`
	Items []string `json:"items,omitempty"`
}

type PipelineResult struct {
	SyncCallsDomain   bool   `json:"sync_calls_domain"`
	SearchCallsDomain bool   `json:"search_calls_domain"`
	DomainTables      int    `json:"domain_tables"`
	Detail            string `json:"detail"`
}

type openAPISpec struct {
	Paths      map[string]json.RawMessage `json:"paths"`
	Components struct {
		SecuritySchemes map[string]struct {
			Type   string `json:"type"`
			Scheme string `json:"scheme"`
		} `json:"securitySchemes"`
	} `json:"components"`
}

func RunDogfood(dir, specPath string) (*DogfoodReport, error) {
	report := &DogfoodReport{
		Dir:      dir,
		SpecPath: specPath,
		Verdict:  "PASS",
	}

	var spec *openAPISpec
	if specPath != "" {
		loaded, err := loadDogfoodOpenAPISpec(specPath)
		if err != nil {
			return nil, err
		}
		spec = loaded

		report.PathCheck = checkPaths(dir, spec.Paths)
		report.AuthCheck = checkAuth(dir, spec.Components.SecuritySchemes)
	} else {
		report.AuthCheck = AuthCheckResult{
			Match:  true,
			Detail: "spec not provided; auth protocol check skipped",
		}
	}

	report.DeadFlags = checkDeadFlags(dir)
	report.DeadFuncs = checkDeadFunctions(dir)
	report.PipelineCheck = checkPipelineIntegrity(dir)
	report.Issues = collectDogfoodIssues(report, spec != nil)
	report.Verdict = deriveDogfoodVerdict(report, spec != nil)

	if err := writeDogfoodResults(report, dir); err != nil {
		return nil, err
	}

	return report, nil
}

func LoadDogfoodResults(dir string) (*DogfoodReport, error) {
	data, err := os.ReadFile(filepath.Join(dir, "dogfood-results.json"))
	if err != nil {
		return nil, err
	}

	var report DogfoodReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func writeDogfoodResults(report *DogfoodReport, dir string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "dogfood-results.json"), data, 0o644)
}

func loadDogfoodOpenAPISpec(specPath string) (*openAPISpec, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("reading spec: %w", err)
	}

	var spec openAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing spec JSON: %w", err)
	}
	return &spec, nil
}

func checkPaths(dir string, paths map[string]json.RawMessage) PathCheckResult {
	result := PathCheckResult{}
	if len(paths) == 0 {
		return result
	}

	cliDir := filepath.Join(dir, "internal", "cli")
	files := listGoFiles(cliDir)
	var commandFiles []string
	for _, file := range files {
		base := filepath.Base(file)
		switch base {
		case "root.go", "helpers.go", "doctor.go", "auth.go", "dogfood.go", "scorecard.go", "vision.go":
			continue
		default:
			commandFiles = append(commandFiles, file)
		}
	}
	sort.Strings(commandFiles)
	if len(commandFiles) > 10 {
		commandFiles = commandFiles[:10]
	}

	specPatterns := compileSpecPathPatterns(paths)
	pathAssignmentRe := regexp.MustCompile(`(?m)\bpath\s*(?::=|=)\s*"([^"]+)"`)

	for _, file := range commandFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		matches := pathAssignmentRe.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			path := match[1]
			result.Tested++
			if pathMatchesSpec(path, specPatterns) {
				result.Valid++
				continue
			}
			result.Invalid = append(result.Invalid, path)
		}
	}

	if result.Tested > 0 {
		result.Pct = (result.Valid * 100) / result.Tested
	}
	result.Invalid = uniqueSorted(result.Invalid)
	return result
}

func checkAuth(dir string, schemes map[string]struct {
	Type   string `json:"type"`
	Scheme string `json:"scheme"`
}) AuthCheckResult {
	result := AuthCheckResult{
		Match:  true,
		Detail: "no recognized auth scheme in spec",
	}

	expectedPrefix := ""
	for name, scheme := range schemes {
		if strings.Contains(strings.ToLower(name), "bot") {
			result.SpecScheme = name + ` scheme (expects "Bot " prefix)`
			expectedPrefix = "Bot "
			break
		}
		if strings.EqualFold(scheme.Type, "http") && strings.EqualFold(scheme.Scheme, "bearer") {
			result.SpecScheme = `http bearer scheme (expects "Bearer " prefix)`
			expectedPrefix = "Bearer "
		}
	}

	clientData, err := os.ReadFile(filepath.Join(dir, "internal", "client", "client.go"))
	if err != nil {
		result.Match = false
		result.Detail = fmt.Sprintf("failed to read client.go: %v", err)
		return result
	}

	clientSource := string(clientData)
	switch {
	case strings.Contains(clientSource, `"Bot "`):
		result.GeneratedFmt = "Bot "
	case strings.Contains(clientSource, `"Bearer "`):
		result.GeneratedFmt = "Bearer "
	default:
		result.GeneratedFmt = "unknown"
	}

	if expectedPrefix == "" {
		result.Detail = "spec not provided or no bot/bearer scheme detected"
		return result
	}

	result.Match = result.GeneratedFmt == expectedPrefix
	if result.Match {
		result.Detail = fmt.Sprintf(`spec and generated client both use %q`, strings.TrimSpace(expectedPrefix))
	} else {
		result.Detail = fmt.Sprintf(`spec expects %q but generated client uses %q`, strings.TrimSpace(expectedPrefix), strings.TrimSpace(result.GeneratedFmt))
	}
	return result
}

func checkDeadFlags(dir string) DeadCodeResult {
	rootData, err := os.ReadFile(filepath.Join(dir, "internal", "cli", "root.go"))
	if err != nil {
		return DeadCodeResult{}
	}

	fieldRe := regexp.MustCompile(`&flags\.(\w+)`)
	matches := fieldRe.FindAllStringSubmatch(string(rootData), -1)
	if len(matches) == 0 {
		return DeadCodeResult{}
	}

	fields := make(map[string]struct{})
	for _, match := range matches {
		fields[match[1]] = struct{}{}
	}

	files := listGoFiles(filepath.Join(dir, "internal", "cli"))
	var otherSources []string
	for _, file := range files {
		if filepath.Base(file) == "root.go" {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		otherSources = append(otherSources, string(data))
	}

	result := DeadCodeResult{Total: len(fields)}
	for _, field := range sortedKeys(fields) {
		needle := "flags." + field
		if containsAny(otherSources, needle) {
			continue
		}
		result.Dead++
		result.Items = append(result.Items, field)
	}
	return result
}

func checkDeadFunctions(dir string) DeadCodeResult {
	helpersPath := filepath.Join(dir, "internal", "cli", "helpers.go")
	data, err := os.ReadFile(helpersPath)
	if err != nil {
		return DeadCodeResult{}
	}

	funcRe := regexp.MustCompile(`(?m)^func\s+([A-Za-z_]\w*)\s*\(`)
	matches := funcRe.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		return DeadCodeResult{}
	}

	names := make(map[string]struct{})
	for _, match := range matches {
		names[match[1]] = struct{}{}
	}

	files := listGoFiles(filepath.Join(dir, "internal", "cli"))
	var otherSources []string
	for _, file := range files {
		if filepath.Base(file) == "helpers.go" {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		otherSources = append(otherSources, string(content))
	}

	result := DeadCodeResult{Total: len(names)}
	for _, name := range sortedKeys(names) {
		callRe := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\s*\(`)
		used := false
		for _, source := range otherSources {
			if callRe.MatchString(source) {
				used = true
				break
			}
		}
		if used {
			continue
		}
		result.Dead++
		result.Items = append(result.Items, name)
	}
	return result
}

func checkPipelineIntegrity(dir string) PipelineResult {
	result := PipelineResult{
		Detail: "sync/search/store files not found",
	}

	syncData, _ := os.ReadFile(filepath.Join(dir, "internal", "cli", "sync.go"))
	searchData, _ := os.ReadFile(filepath.Join(dir, "internal", "cli", "search.go"))
	storeData, _ := os.ReadFile(filepath.Join(dir, "internal", "store", "store.go"))

	syncSource := string(syncData)
	searchSource := string(searchData)
	storeSource := string(storeData)

	domainUpsertRe := regexp.MustCompile(`\.Upsert[A-Z]\w*\s*\(`)
	domainSearchRe := regexp.MustCompile(`\.Search[A-Z]\w*\s*\(`)

	result.SyncCallsDomain = domainUpsertRe.MatchString(syncSource)
	result.SearchCallsDomain = domainSearchRe.MatchString(searchSource)
	result.DomainTables = countDomainTables(storeSource)

	var parts []string
	switch {
	case result.SyncCallsDomain:
		parts = append(parts, "sync uses domain-specific Upsert methods")
	case strings.Contains(syncSource, ".Upsert("):
		parts = append(parts, "sync uses generic Upsert only")
	default:
		parts = append(parts, "sync Upsert calls not found")
	}

	switch {
	case result.SearchCallsDomain:
		parts = append(parts, "search uses domain-specific Search methods")
	case strings.Contains(searchSource, ".Search("):
		parts = append(parts, "search uses generic Search only")
	default:
		parts = append(parts, "search methods not found")
	}

	if storeSource != "" {
		parts = append(parts, fmt.Sprintf("%d domain tables found", result.DomainTables))
	}

	result.Detail = strings.Join(parts, "; ")
	return result
}

func deriveDogfoodVerdict(report *DogfoodReport, hasSpec bool) string {
	if hasSpec && report.PathCheck.Tested > 0 && report.PathCheck.Pct < 70 {
		return "FAIL"
	}
	if hasSpec && !report.AuthCheck.Match {
		return "FAIL"
	}
	if report.DeadFlags.Dead >= 3 {
		return "FAIL"
	}
	if report.DeadFlags.Dead >= 1 && report.DeadFlags.Dead <= 2 {
		return "WARN"
	}
	if report.DeadFuncs.Dead >= 1 {
		return "WARN"
	}
	if !report.PipelineCheck.SyncCallsDomain {
		return "WARN"
	}
	return "PASS"
}

func collectDogfoodIssues(report *DogfoodReport, hasSpec bool) []string {
	var issues []string
	if hasSpec && report.PathCheck.Tested > 0 && report.PathCheck.Pct < 70 {
		issues = append(issues, fmt.Sprintf("%d%% path validity against spec", report.PathCheck.Pct))
	}
	if hasSpec && !report.AuthCheck.Match {
		issues = append(issues, "auth protocol mismatch")
	}
	if report.DeadFlags.Dead >= 3 {
		issues = append(issues, fmt.Sprintf("%d dead flags found", report.DeadFlags.Dead))
	} else if report.DeadFlags.Dead > 0 {
		issues = append(issues, fmt.Sprintf("%d dead flags found", report.DeadFlags.Dead))
	}
	if report.DeadFuncs.Dead > 0 {
		issues = append(issues, fmt.Sprintf("%d dead helper functions found", report.DeadFuncs.Dead))
	}
	if !report.PipelineCheck.SyncCallsDomain {
		issues = append(issues, "sync uses generic Upsert only")
	}
	return issues
}

func compileSpecPathPatterns(paths map[string]json.RawMessage) []*regexp.Regexp {
	keys := make([]string, 0, len(paths))
	for path := range paths {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	paramRe := regexp.MustCompile(`\\\{[^/]+\\\}`)
	var patterns []*regexp.Regexp
	for _, path := range keys {
		quoted := regexp.QuoteMeta(path)
		regex := "^" + paramRe.ReplaceAllString(quoted, `[^/]+`) + "$"
		patterns = append(patterns, regexp.MustCompile(regex))
	}
	return patterns
}

func pathMatchesSpec(path string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

func listGoFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	return files
}

func countDomainTables(storeSource string) int {
	if storeSource == "" {
		return 0
	}

	tableRe := regexp.MustCompile(`(?is)CREATE TABLE IF NOT EXISTS\s+\w+\s*\((.*?)\);`)
	matches := tableRe.FindAllStringSubmatch(storeSource, -1)
	count := 0
	for _, match := range matches {
		columns := 0
		for _, line := range strings.Split(match[1], "\n") {
			line = strings.TrimSpace(strings.TrimSuffix(line, ","))
			if line == "" {
				continue
			}
			upper := strings.ToUpper(line)
			if strings.HasPrefix(upper, "PRIMARY KEY") || strings.HasPrefix(upper, "FOREIGN KEY") || strings.HasPrefix(upper, "UNIQUE") || strings.HasPrefix(upper, "CONSTRAINT") || strings.HasPrefix(upper, "CHECK") {
				continue
			}
			columns++
		}
		if columns > 3 {
			count++
		}
	}
	return count
}

func uniqueSorted(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return sortedKeys(set)
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func containsAny(sources []string, needle string) bool {
	for _, source := range sources {
		if strings.Contains(source, needle) {
			return true
		}
	}
	return false
}
