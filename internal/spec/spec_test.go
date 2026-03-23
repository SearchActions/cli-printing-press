package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStytch(t *testing.T) {
	s, err := Parse("../../testdata/stytch.yaml")
	require.NoError(t, err)

	assert.Equal(t, "stytch", s.Name)
	assert.Equal(t, "Stytch authentication API CLI", s.Description)
	assert.Equal(t, "0.1.0", s.Version)
	assert.Equal(t, "https://api.stytch.com/v1", s.BaseURL)

	// Auth
	assert.Equal(t, "api_key", s.Auth.Type)
	assert.Equal(t, "Authorization", s.Auth.Header)
	assert.Len(t, s.Auth.EnvVars, 2)

	// Resources
	assert.Len(t, s.Resources, 2)
	users := s.Resources["users"]
	assert.Equal(t, "Manage Stytch users", users.Description)
	assert.Len(t, users.Endpoints, 4) // list, get, create, delete

	// Users list endpoint
	list := users.Endpoints["list"]
	assert.Equal(t, "GET", list.Method)
	assert.Equal(t, "/users", list.Path)
	assert.NotNil(t, list.Pagination)
	assert.Equal(t, "cursor", list.Pagination.Type)

	// Users create endpoint
	create := users.Endpoints["create"]
	assert.Equal(t, "POST", create.Method)
	assert.Len(t, create.Body, 2)

	// Sessions
	sessions := s.Resources["sessions"]
	assert.Len(t, sessions.Endpoints, 2)

	// Types
	assert.Len(t, s.Types, 2)
	assert.Len(t, s.Types["User"].Fields, 5)
	assert.Len(t, s.Types["Session"].Fields, 4)
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		spec    APISpec
		wantErr string
	}{
		{
			name:    "empty name",
			spec:    APISpec{BaseURL: "http://x", Resources: map[string]Resource{"a": {Endpoints: map[string]Endpoint{"b": {Method: "GET", Path: "/"}}}}},
			wantErr: "name is required",
		},
		{
			name:    "empty base_url",
			spec:    APISpec{Name: "x", Resources: map[string]Resource{"a": {Endpoints: map[string]Endpoint{"b": {Method: "GET", Path: "/"}}}}},
			wantErr: "base_url is required",
		},
		{
			name:    "no resources",
			spec:    APISpec{Name: "x", BaseURL: "http://x"},
			wantErr: "at least one resource is required",
		},
		{
			name:    "endpoint missing method",
			spec:    APISpec{Name: "x", BaseURL: "http://x", Resources: map[string]Resource{"a": {Endpoints: map[string]Endpoint{"b": {Path: "/"}}}}},
			wantErr: "method is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNewFields(t *testing.T) {
	s := APISpec{
		Name:    "x",
		BaseURL: "http://x",
		Auth: AuthConfig{
			Type:    "api_key",
			Scheme:  "ApiKeyAuth",
			In:      "header",
			EnvVars: []string{"API_KEY"},
		},
		Resources: map[string]Resource{
			"users": {
				Endpoints: map[string]Endpoint{
					"list": {
						Method:       "GET",
						Path:         "/users",
						ResponsePath: "results.items",
						Params: []Param{
							{
								Name:   "status",
								Type:   "string",
								Enum:   []string{"active", "inactive"},
								Format: "email",
							},
						},
					},
				},
			},
		},
	}

	require.NoError(t, s.Validate())

	endpoint := s.Resources["users"].Endpoints["list"]
	assert.Equal(t, "results.items", endpoint.ResponsePath)
	assert.Equal(t, []string{"active", "inactive"}, endpoint.Params[0].Enum)
	assert.Equal(t, "email", endpoint.Params[0].Format)
	assert.Equal(t, "ApiKeyAuth", s.Auth.Scheme)
	assert.Equal(t, "header", s.Auth.In)
}
