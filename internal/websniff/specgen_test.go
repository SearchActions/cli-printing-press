package websniff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mvanhorn/cli-printing-press/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	t.Parallel()

	apiSpec, err := Analyze(filepath.Join("..", "..", "testdata", "sniff", "sample-enriched.json"))
	require.NoError(t, err)
	require.NotNil(t, apiSpec)

	assert.Equal(t, "hn-algolia", apiSpec.Name)
	require.NotEmpty(t, apiSpec.Resources)

	foundEndpointWithParams := false
	for _, resource := range apiSpec.Resources {
		require.NotEmpty(t, resource.Endpoints)
		for _, endpoint := range resource.Endpoints {
			if len(endpoint.Params) > 0 {
				foundEndpointWithParams = true
			}
		}
	}

	assert.True(t, foundEndpointWithParams)
	assert.NoError(t, apiSpec.Validate())
}

func TestWriteSpec(t *testing.T) {
	t.Parallel()

	apiSpec := &spec.APISpec{
		Name:        "example",
		Description: "Example API",
		Version:     "0.1.0",
		BaseURL:     "https://api.example.com",
		Auth:        spec.AuthConfig{Type: "none"},
		Config: spec.ConfigSpec{
			Format: "toml",
			Path:   "~/.config/example-pp-cli/config.toml",
		},
		Resources: map[string]spec.Resource{
			"widgets": {
				Description: "Operations on widgets",
				Endpoints: map[string]spec.Endpoint{
					"list_widgets": {
						Method: "GET",
						Path:   "/widgets",
						Response: spec.ResponseDef{
							Type: "object",
							Item: "widgets",
						},
					},
				},
			},
		},
		Types: map[string]spec.TypeDef{},
	}

	outputPath := filepath.Join(t.TempDir(), "nested", "spec.yaml")
	err := WriteSpec(apiSpec, outputPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	parsed, err := spec.ParseBytes(data)
	require.NoError(t, err)
	assert.Equal(t, apiSpec.Name, parsed.Name)
	assert.Equal(t, apiSpec.BaseURL, parsed.BaseURL)
}

func TestDeriveNameFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "strips www and tld",
			raw:  "https://www.youtube.com",
			want: "youtube",
		},
		{
			name: "keeps meaningful subdomain",
			raw:  "https://hn.algolia.com",
			want: "hn-algolia",
		},
		{
			name: "drops generic api prefix",
			raw:  "https://api.example.com",
			want: "example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, deriveNameFromURL(tt.raw))
		})
	}
}
