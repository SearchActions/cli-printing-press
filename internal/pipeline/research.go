package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mvanhorn/cli-printing-press/internal/catalog"
)

// ResearchResult holds the output of the research phase.
type ResearchResult struct {
	APIName        string        `json:"api_name"`
	NoveltyScore   int           `json:"novelty_score"` // 1-10
	Alternatives   []Alternative `json:"alternatives"`
	Gaps           []string      `json:"gaps"`           // what alternatives miss
	Patterns       []string      `json:"patterns"`       // what alternatives do well
	Recommendation string        `json:"recommendation"` // "proceed", "proceed-with-gaps", "skip"
	ResearchedAt   time.Time     `json:"researched_at"`
}

// Alternative represents a known competing CLI tool.
type Alternative struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	Language      string `json:"language"`
	InstallMethod string `json:"install_method"` // brew, npm, pip, cargo, binary
	Stars         int    `json:"stars"`
	LastUpdated   string `json:"last_updated"`
	CommandCount  int    `json:"command_count"`
	HasJSON       bool   `json:"has_json_output"`
	HasAuth       bool   `json:"has_auth_support"`
}

// RunResearch executes the research phase for an API.
// It checks the catalog for known alternatives, then optionally
// queries the GitHub API for additional CLI tools.
func RunResearch(apiName, catalogDir, pipelineDir string) (*ResearchResult, error) {
	result := &ResearchResult{
		APIName:      apiName,
		ResearchedAt: time.Now(),
	}

	// Step 1: Check catalog for known alternatives
	catalogAlts := loadCatalogAlternatives(apiName, catalogDir)
	for _, alt := range catalogAlts {
		result.Alternatives = append(result.Alternatives, Alternative{
			Name:     alt.Name,
			URL:      alt.URL,
			Language: alt.Language,
		})
	}

	// Step 2: Search GitHub for "<api-name> cli" repos
	ghAlts, err := searchGitHubCLIs(apiName)
	if err != nil {
		// Non-fatal: log and continue with catalog-only results
		fmt.Fprintf(os.Stderr, "warning: GitHub search failed: %v\n", err)
	} else {
		result.Alternatives = append(result.Alternatives, ghAlts...)
	}

	// Step 3: Deduplicate by URL
	result.Alternatives = deduplicateAlts(result.Alternatives)

	// Step 4: Score novelty and produce recommendation
	result.NoveltyScore = scoreNovelty(result.Alternatives)
	result.Recommendation = recommend(result.NoveltyScore)

	// Step 5: Analyze gaps and patterns
	result.Gaps, result.Patterns = analyzeAlternatives(result.Alternatives)

	// Step 6: Write research.json
	if err := writeResearchJSON(result, pipelineDir); err != nil {
		return result, fmt.Errorf("writing research.json: %w", err)
	}

	return result, nil
}

// LoadResearch reads research.json from a pipeline directory.
func LoadResearch(pipelineDir string) (*ResearchResult, error) {
	data, err := os.ReadFile(filepath.Join(pipelineDir, "research.json"))
	if err != nil {
		return nil, err
	}
	var r ResearchResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func loadCatalogAlternatives(apiName, catalogDir string) []catalog.KnownAlt {
	if catalogDir == "" {
		catalogDir = "catalog"
	}
	path := filepath.Join(catalogDir, apiName+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	entry, err := catalog.ParseEntry(data)
	if err != nil {
		return nil
	}
	return entry.KnownAlternatives
}

// ghSearchResponse models the GitHub search API response.
type ghSearchResponse struct {
	Items []ghRepo `json:"items"`
}

type ghRepo struct {
	FullName    string    `json:"full_name"`
	HTMLURL     string    `json:"html_url"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Stars       int       `json:"stargazers_count"`
	PushedAt    time.Time `json:"pushed_at"`
}

func searchGitHubCLIs(apiName string) ([]Alternative, error) {
	query := fmt.Sprintf("%s+cli+language:go+language:python+language:typescript+language:rust", apiName)
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&per_page=5", query)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	// Use GITHUB_TOKEN if available for higher rate limits
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var searchResp ghSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	var alts []Alternative
	for _, repo := range searchResp.Items {
		alt := Alternative{
			Name:        repo.FullName,
			URL:         repo.HTMLURL,
			Language:    strings.ToLower(repo.Language),
			Stars:       repo.Stars,
			LastUpdated: repo.PushedAt.Format("2006-01-02"),
		}
		// Infer install method from language
		switch strings.ToLower(repo.Language) {
		case "go":
			alt.InstallMethod = "binary"
		case "python":
			alt.InstallMethod = "pip"
		case "typescript", "javascript":
			alt.InstallMethod = "npm"
		case "rust":
			alt.InstallMethod = "cargo"
		default:
			alt.InstallMethod = "source"
		}
		alts = append(alts, alt)
	}
	return alts, nil
}

func deduplicateAlts(alts []Alternative) []Alternative {
	seen := make(map[string]bool)
	var unique []Alternative
	for _, a := range alts {
		key := strings.ToLower(a.URL)
		if key == "" {
			key = strings.ToLower(a.Name)
		}
		if !seen[key] {
			seen[key] = true
			unique = append(unique, a)
		}
	}
	return unique
}

func scoreNovelty(alts []Alternative) int {
	if len(alts) == 0 {
		return 10 // No alternatives - maximum novelty
	}

	// Check for high-star official CLIs
	hasOfficialCLI := false
	maxStars := 0
	for _, a := range alts {
		if a.Stars > maxStars {
			maxStars = a.Stars
		}
		if a.Stars > 5000 {
			hasOfficialCLI = true
		}
	}

	if hasOfficialCLI {
		return 2 // Official CLI exists and is popular
	}
	if maxStars > 1000 {
		return 4 // Popular community CLI exists
	}
	if maxStars > 100 {
		return 6 // Some community alternatives but not dominant
	}
	if len(alts) > 0 {
		return 7 // Alternatives exist but are small/stale
	}
	return 10
}

func recommend(noveltyScore int) string {
	switch {
	case noveltyScore <= 3:
		return "skip"
	case noveltyScore <= 6:
		return "proceed-with-gaps"
	default:
		return "proceed"
	}
}

func analyzeAlternatives(alts []Alternative) (gaps, patterns []string) {
	hasGo := false
	hasJSON := false
	hasDryRun := false

	for _, a := range alts {
		if a.Language == "go" {
			hasGo = true
		}
		if a.HasJSON {
			hasJSON = true
		}
	}

	// Our CLI always has these - identify where alternatives fall short
	if !hasGo {
		patterns = append(patterns, "no Go-based alternative exists - our binary has zero runtime deps")
	}
	if !hasJSON {
		gaps = append(gaps, "most alternatives lack --json output mode")
	}
	if !hasDryRun {
		gaps = append(gaps, "no alternative offers --dry-run mode")
	}

	// Check freshness
	staleCount := 0
	for _, a := range alts {
		if a.LastUpdated != "" {
			if t, err := time.Parse("2006-01-02", a.LastUpdated); err == nil {
				if time.Since(t) > 365*24*time.Hour {
					staleCount++
				}
			}
		}
	}
	if staleCount > 0 && len(alts) > 0 {
		gaps = append(gaps, fmt.Sprintf("%d/%d alternatives are stale (>1 year since last update)", staleCount, len(alts)))
	}

	if len(gaps) == 0 {
		gaps = append(gaps, "alternatives cover the basics")
	}
	if len(patterns) == 0 {
		patterns = append(patterns, "standard CLI patterns (list, get, create, update, delete)")
	}

	return gaps, patterns
}

func writeResearchJSON(result *ResearchResult, pipelineDir string) error {
	if err := os.MkdirAll(pipelineDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(pipelineDir, "research.json"), data, 0o644)
}
