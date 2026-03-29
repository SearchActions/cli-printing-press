package websniff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func ParseHAR(path string) (*HAR, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var har HAR
	if err := json.Unmarshal(data, &har); err != nil {
		return nil, fmt.Errorf("parsing har json: %w", err)
	}

	return &har, nil
}

func ParseEnriched(path string) (*EnrichedCapture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var capture EnrichedCapture
	if err := json.Unmarshal(data, &capture); err != nil {
		return nil, fmt.Errorf("parsing enriched json: %w", err)
	}

	return &capture, nil
}

func ParseCapture(path string) ([]EnrichedEntry, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("reading file: %w", err)
	}

	if bytes.Contains(data, []byte(`"log"`)) {
		var har HAR
		if err := json.Unmarshal(data, &har); err != nil {
			return nil, "", fmt.Errorf("parsing har json: %w", err)
		}

		entries := make([]EnrichedEntry, 0, len(har.Log.Entries))
		for _, entry := range har.Log.Entries {
			entries = append(entries, convertHAREntry(entry))
		}

		targetURL := ""
		if len(har.Log.Entries) > 0 {
			targetURL = har.Log.Entries[0].Request.URL
		}

		return entries, targetURL, nil
	}

	if bytes.Contains(data, []byte(`"target_url"`)) {
		var capture EnrichedCapture
		if err := json.Unmarshal(data, &capture); err != nil {
			return nil, "", fmt.Errorf("parsing enriched json: %w", err)
		}

		return capture.Entries, capture.TargetURL, nil
	}

	return nil, "", fmt.Errorf("unknown capture format")
}

func convertHAREntry(entry HAREntry) EnrichedEntry {
	headers := make(map[string]string, len(entry.Request.Headers))
	for _, header := range entry.Request.Headers {
		headers[header.Name] = header.Value
	}

	requestBody := ""
	if entry.Request.PostData != nil {
		requestBody = entry.Request.PostData.Text
	}

	return EnrichedEntry{
		Method:              entry.Request.Method,
		URL:                 entry.Request.URL,
		RequestBody:         requestBody,
		ResponseBody:        entry.Response.Content.Text,
		ResponseStatus:      entry.Response.Status,
		ResponseContentType: entry.Response.Content.MimeType,
		RequestHeaders:      headers,
	}
}
