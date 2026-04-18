package megamcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// DefaultBaseURL is the default GitHub raw content URL for the library repo.
const DefaultBaseURL = "https://raw.githubusercontent.com/mvanhorn/printing-press-library/main"

// FetchRegistry fetches registry.json from baseURL and parses it into a Registry.
// baseURL is injectable for testing (use httptest.NewServer).
// When GITHUB_TOKEN is set, an Authorization header is attached so private-repo
// reads via raw.githubusercontent.com succeed.
func FetchRegistry(baseURL string) (*Registry, error) {
	registryURL := baseURL + "/registry.json"

	req, err := http.NewRequest(http.MethodGet, registryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building registry request: %w", err)
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching registry: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching registry: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 10MB limit
	if err != nil {
		return nil, fmt.Errorf("reading registry body: %w", err)
	}

	var registry Registry
	if err := json.Unmarshal(body, &registry); err != nil {
		return nil, fmt.Errorf("parsing registry JSON: %w", err)
	}

	return &registry, nil
}
