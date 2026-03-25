package docspec

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleDocs = `<html><body>
<h1>My API</h1>
<p>Base URL: https://api.myservice.com/v1</p>
<p>Authentication: Pass your API key in the X-API-Key header.</p>

<h2>Users</h2>
<pre>GET /v1/users</pre>
<pre>POST /v1/users</pre>
<pre>GET /v1/users/{id}</pre>
<pre>PUT /v1/users/{id}</pre>
<pre>DELETE /v1/users/{id}</pre>

<h2>Projects</h2>
<pre>GET /v1/projects</pre>
<pre>POST /v1/projects</pre>

<h3>Parameters</h3>
<table>
<tr><td>name</td><td>string</td></tr>
<tr><td>email</td><td>string</td></tr>
<tr><td>age</td><td>integer</td></tr>
<tr><td>active</td><td>boolean</td></tr>
</table>

<h3>Example</h3>
<pre>{"title": "My Project", "description": "A project"}</pre>
</body></html>`

func TestGenerateFromDocs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(sampleDocs))
	}))
	defer srv.Close()

	apiSpec, err := GenerateFromDocs(srv.URL, "myapi")
	require.NoError(t, err)

	assert.Equal(t, "myapi", apiSpec.Name)
	assert.Equal(t, "https://api.myservice.com", apiSpec.BaseURL)
	assert.Equal(t, "api_key", apiSpec.Auth.Type)
	assert.Contains(t, apiSpec.Resources, "users")
	assert.Contains(t, apiSpec.Resources, "projects")

	users := apiSpec.Resources["users"]
	assert.GreaterOrEqual(t, len(users.Endpoints), 3)

	projects := apiSpec.Resources["projects"]
	assert.GreaterOrEqual(t, len(projects.Endpoints), 2)
}

func TestExtractEndpoints(t *testing.T) {
	body := `GET /users POST /users GET /users/{id} DELETE /items/{id}`
	eps := extractEndpoints(body)
	assert.Len(t, eps, 4)
	assert.Equal(t, "GET", eps[0].Method)
	assert.Equal(t, "/users", eps[0].Path)
}

func TestDetectAuth(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"bearer", "Use Authorization: Bearer token", "bearer_token"},
		{"api_key", "Pass your API key in the header", "api_key"},
		{"oauth", "We support OAuth 2.0 flows", "oauth2"},
		{"default", "No auth mentioned here", "bearer_token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := detectAuth(tt.body)
			assert.Equal(t, tt.want, auth.Type)
		})
	}
}

func TestDetectBaseURL(t *testing.T) {
	assert.Equal(t, "https://api.stripe.com", detectBaseURL("Base URL: https://api.stripe.com/v1"))
	assert.Equal(t, "https://api.example.com", detectBaseURL("No URL here"))
}

func TestFirstSegment(t *testing.T) {
	assert.Equal(t, "users", firstSegment("/users"))
	assert.Equal(t, "users", firstSegment("/v1/users"))
	assert.Equal(t, "users", firstSegment("/v2/users/{id}"))
	assert.Equal(t, "items", firstSegment("/items/search"))
}

func TestEndpointName(t *testing.T) {
	assert.Equal(t, "get_users", endpointName("GET", "/v1/users"))
	assert.Equal(t, "create_users", endpointName("POST", "/v1/users"))
	assert.Equal(t, "get_users", endpointName("GET", "/v1/users/{id}"))
	assert.Equal(t, "delete_items", endpointName("DELETE", "/items/{id}"))
}

func TestExtractParams(t *testing.T) {
	body := `<table><tr><td>name</td><td>string</td></tr><tr><td>count</td><td>integer</td></tr></table>
<pre>{"title": "hello", "active": true}</pre>`
	params := extractParams(body)
	assert.GreaterOrEqual(t, len(params), 2)

	names := map[string]bool{}
	for _, p := range params {
		names[p.Name] = true
	}
	assert.True(t, names["name"])
	assert.True(t, names["count"])
}

func TestExtractPathParams(t *testing.T) {
	params := extractPathParams("/users/{user_id}/posts/{post_id}")
	assert.Len(t, params, 2)
	assert.Equal(t, "user_id", params[0].Name)
	assert.True(t, params[0].Required)
	assert.True(t, params[0].Positional)
	assert.Equal(t, "post_id", params[1].Name)
}

func TestNoEndpointsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body>Nothing here</body></html>"))
	}))
	defer srv.Close()

	_, err := GenerateFromDocs(srv.URL, "empty")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no endpoints found")
}
