package client

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/USER/linear-cli/internal/config"
)

type Client struct {
	BaseURL    string
	Config     *config.Config
	HTTPClient *http.Client
	DryRun     bool
}

// APIError carries HTTP status information for structured exit codes.
type APIError struct {
	Method     string
	Path       string
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s %s returned HTTP %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

func New(cfg *config.Config, timeout time.Duration) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		Config:     cfg,
		HTTPClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Get(path string, params map[string]string) (json.RawMessage, error) {
	return c.do("GET", path, params, nil)
}

func (c *Client) Post(path string, body any) (json.RawMessage, error) {
	return c.do("POST", path, nil, body)
}

func (c *Client) Delete(path string) (json.RawMessage, error) {
	return c.do("DELETE", path, nil, nil)
}

func (c *Client) Put(path string, body any) (json.RawMessage, error) {
	return c.do("PUT", path, nil, body)
}

func (c *Client) Patch(path string, body any) (json.RawMessage, error) {
	return c.do("PATCH", path, nil, body)
}

func (c *Client) do(method, path string, params map[string]string, body any) (json.RawMessage, error) {
	url := c.BaseURL + path

	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling body: %w", err)
		}
		bodyBytes = b
	}

	// Build the request for dry-run display or actual execution
	if c.DryRun {
		return c.dryRun(method, url, params, bodyBytes)
	}

	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = strings.NewReader(string(bodyBytes))
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		authHeader, err := c.authHeader()
		if err != nil {
			return nil, err
		}
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("User-Agent", "linear-cli/1.0.0")

		if params != nil {
			q := req.URL.Query()
			for k, v := range params {
				if v != "" {
					q.Set(k, v)
				}
			}
			req.URL.RawQuery = q.Encode()
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("%s %s: %w", method, path, err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		// Success
		if resp.StatusCode < 400 {
			return json.RawMessage(respBody), nil
		}

		apiErr := &APIError{
			Method:     method,
			Path:       path,
			StatusCode: resp.StatusCode,
			Body:       truncateBody(respBody),
		}

		// Rate limited - wait and retry
		if resp.StatusCode == 429 && attempt < maxRetries {
			wait := retryAfter(resp)
			fmt.Fprintf(os.Stderr, "rate limited, waiting %s (attempt %d/%d)\n", wait, attempt+1, maxRetries)
			time.Sleep(wait)
			lastErr = apiErr
			continue
		}

		// Server error - retry with backoff
		if resp.StatusCode >= 500 && attempt < maxRetries {
			wait := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			fmt.Fprintf(os.Stderr, "server error %d, retrying in %s (attempt %d/%d)\n", resp.StatusCode, wait, attempt+1, maxRetries)
			time.Sleep(wait)
			lastErr = apiErr
			continue
		}

		// Client error or retries exhausted - return the error
		return nil, apiErr
	}

	return nil, lastErr
}

func (c *Client) dryRun(method, url string, params map[string]string, body []byte) (json.RawMessage, error) {
	fmt.Fprintf(os.Stderr, "%s %s\n", method, url)
	if params != nil {
		for k, v := range params {
			if v != "" {
				fmt.Fprintf(os.Stderr, "  ?%s=%s\n", k, v)
			}
		}
	}
	authHeader, err := c.authHeader()
	if err != nil {
		return nil, err
	}
	if authHeader != "" {
		// Mask token for safety
		auth := authHeader
		if len(auth) > 20 {
			auth = auth[:15] + "..."
		}
		fmt.Fprintf(os.Stderr, "  Authorization: %s\n", auth)
	}
	if body != nil {
		var pretty json.RawMessage
		if json.Unmarshal(body, &pretty) == nil {
			enc := json.NewEncoder(os.Stderr)
			enc.SetIndent("  ", "  ")
			fmt.Fprintf(os.Stderr, "  Body:\n")
			enc.Encode(pretty)
		}
	}
	fmt.Fprintf(os.Stderr, "\n(dry run - no request sent)\n")
	return json.RawMessage(`{"dry_run": true}`), nil
}

func (c *Client) authHeader() (string, error) {
	if c.Config == nil {
		return "", nil
	}
	if c.Config.AccessToken != "" && !c.Config.TokenExpiry.IsZero() && time.Now().After(c.Config.TokenExpiry) && c.Config.RefreshToken != "" {
		if err := c.refreshAccessToken(); err != nil {
			return "", err
		}
	}
	return c.Config.AuthHeader(), nil
}

func (c *Client) refreshAccessToken() error {
	if c.Config == nil {
		return nil
	}
	if c.Config.RefreshToken == "" {
		return nil
	}

	tokenURL := ""
	if tokenURL == "" {
		return nil
	}

	params := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.Config.RefreshToken},
		"client_id":     {c.Config.ClientID},
	}
	if c.Config.ClientSecret != "" {
		params.Set("client_secret", c.Config.ClientSecret)
	}

	resp, err := c.HTTPClient.PostForm(tokenURL, params)
	if err != nil {
		return fmt.Errorf("refreshing access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("refreshing access token: HTTP %d: %s", resp.StatusCode, truncateBody(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("parsing refresh response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("refreshing access token: no access token in response")
	}

	refreshToken := c.Config.RefreshToken
	if tokenResp.RefreshToken != "" {
		refreshToken = tokenResp.RefreshToken
	}

	expiry := time.Time{}
	if tokenResp.ExpiresIn > 0 {
		expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	if err := c.Config.SaveTokens(c.Config.ClientID, c.Config.ClientSecret, tokenResp.AccessToken, refreshToken, expiry); err != nil {
		return fmt.Errorf("saving refreshed token: %w", err)
	}

	return nil
}

func retryAfter(resp *http.Response) time.Duration {
	header := resp.Header.Get("Retry-After")
	if header == "" {
		return 5 * time.Second
	}
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(header); err == nil {
		wait := time.Until(t)
		if wait > 0 {
			return wait
		}
	}
	return 5 * time.Second
}

func truncateBody(b []byte) string {
	s := string(b)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
