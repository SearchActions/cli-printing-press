package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEntry(t *testing.T) {
	data := []byte(`
name: test-api
display_name: Test API
description: Test API for catalog parser validation
category: developer-tools
spec_url: https://example.com/openapi.yaml
spec_format: yaml
openapi_version: "3.0"
tier: community
verified_date: "2026-03-23"
homepage: https://example.com
notes: Example fixture.
`)

	entry, err := ParseEntry(data)
	require.NoError(t, err)

	assert.Equal(t, "test-api", entry.Name)
	assert.Equal(t, "Test API", entry.DisplayName)
	assert.Equal(t, "Test API for catalog parser validation", entry.Description)
	assert.Equal(t, "developer-tools", entry.Category)
	assert.Equal(t, "https://example.com/openapi.yaml", entry.SpecURL)
	assert.Equal(t, "yaml", entry.SpecFormat)
	assert.Equal(t, "3.0", entry.OpenAPIVersion)
	assert.Equal(t, "community", entry.Tier)
	assert.Equal(t, "2026-03-23", entry.VerifiedDate)
	assert.Equal(t, "https://example.com", entry.Homepage)
	assert.Equal(t, "Example fixture.", entry.Notes)
}

func TestValidateEntry(t *testing.T) {
	base := Entry{
		Name:        "test-api",
		DisplayName: "Test API",
		Description: "A valid catalog entry",
		Category:    "developer-tools",
		SpecURL:     "https://example.com/openapi.yaml",
		SpecFormat:  "yaml",
		Tier:        "official",
	}

	tests := []struct {
		name    string
		mutate  func(*Entry)
		wantErr string
	}{
		{
			name: "empty name",
			mutate: func(e *Entry) {
				e.Name = ""
			},
			wantErr: "name is required",
		},
		{
			name: "invalid name format",
			mutate: func(e *Entry) {
				e.Name = "Not_Kebab"
			},
			wantErr: "name must be lowercase kebab-case",
		},
		{
			name: "invalid category",
			mutate: func(e *Entry) {
				e.Category = "finance"
			},
			wantErr: "category must be one of",
		},
		{
			name: "non https spec url",
			mutate: func(e *Entry) {
				e.SpecURL = "http://example.com/openapi.yaml"
			},
			wantErr: `spec_url must start with "https://"`,
		},
		{
			name: "invalid spec format",
			mutate: func(e *Entry) {
				e.SpecFormat = "xml"
			},
			wantErr: "spec_format must be one of",
		},
		{
			name: "invalid tier",
			mutate: func(e *Entry) {
				e.Tier = "partner"
			},
			wantErr: "tier must be one of",
		},
		{
			name: "missing display_name",
			mutate: func(e *Entry) {
				e.DisplayName = ""
			},
			wantErr: "display_name is required",
		},
		{
			name: "missing description",
			mutate: func(e *Entry) {
				e.Description = ""
			},
			wantErr: "description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := base
			tt.mutate(&entry)

			err := entry.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestParseDir(t *testing.T) {
	entries, err := ParseDir("../../testdata/catalog")
	require.NoError(t, err)
	require.Len(t, entries, 2)

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name)
	}

	assert.Contains(t, names, "test-api")
	assert.Contains(t, names, "petstore")
}
