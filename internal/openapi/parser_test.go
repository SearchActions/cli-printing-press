package openapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePetstore(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "openapi", "petstore.yaml"))
	require.NoError(t, err)

	parsed, err := Parse(data)
	require.NoError(t, err)

	assert.Equal(t, "swagger-petstore-openapi-30", parsed.Name)
	assert.Equal(t, "/api/v3", parsed.BaseURL)
	assert.NotEmpty(t, parsed.Resources)

	hasEndpoint := false
	for _, resource := range parsed.Resources {
		if len(resource.Endpoints) > 0 {
			hasEndpoint = true
			break
		}
	}
	assert.True(t, hasEndpoint)

	assert.NotEmpty(t, parsed.Types)
	assert.Contains(t, parsed.Types, "Pet")
}

func TestIsOpenAPI(t *testing.T) {
	t.Parallel()

	openAPIYAML := []byte(`
openapi: 3.0.0
info:
  title: Demo
  version: 1.0.0
paths: {}
`)
	openAPIJSON := []byte(`{"openapi":"3.0.1","info":{"title":"Demo","version":"1.0.0"},"paths":{}}`)
	internalYAML := []byte(`
name: demo
base_url: https://api.example.com
resources:
  users:
    endpoints:
      list:
        method: GET
        path: /users
`)

	assert.True(t, IsOpenAPI(openAPIYAML))
	assert.True(t, IsOpenAPI(openAPIJSON))
	assert.False(t, IsOpenAPI(internalYAML))
}
