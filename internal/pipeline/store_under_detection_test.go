package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	apispec "github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpecHasCollectionShapedResource(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  bool
	}{
		{"list+detail pair", []string{"/items", "/items/{id}"}, true},
		{"detail only", []string{"/items/{id}"}, false},
		{"list only", []string{"/items"}, false},
		{"nested collection pair", []string{"/orgs/{org}/repos", "/orgs/{org}/repos/{id}"}, true},
		{"RPC actions", []string{"/v1/load", "/v1/sql", "/v1/meta", "/v1/running-query/{requestId}"}, false},
		{"nil paths", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, specHasCollectionShapedResource(tc.paths))
		})
	}
}

func TestStoreUnderDetected(t *testing.T) {
	collectionPaths := []string{"/items", "/items/{id}"}

	newDir := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		writeClientPkgGo(t, dir)
		return dir
	}

	t.Run("collection spec with no store is under-detection", func(t *testing.T) {
		assert.True(t, storeUnderDetected(newDir(t), collectionPaths, apispec.KindREST, false))
	})

	t.Run("store present silences the guard", func(t *testing.T) {
		dir := newDir(t)
		writeStubFile(t, filepath.Join(dir, "internal", "store", "store.go"), "package store\n")
		assert.False(t, storeUnderDetected(dir, collectionPaths, apispec.KindREST, false))
	})

	t.Run("synthetic kind silences the guard", func(t *testing.T) {
		assert.False(t, storeUnderDetected(newDir(t), collectionPaths, apispec.KindSynthetic, false))
	})

	t.Run("GraphQL spec flag silences the guard", func(t *testing.T) {
		assert.False(t, storeUnderDetected(newDir(t), collectionPaths, apispec.KindREST, true))
	})

	t.Run("GraphQL CLI dir silences the guard", func(t *testing.T) {
		dir := newDir(t)
		writeStubFile(t, filepath.Join(dir, "internal", "client", "graphql.go"), "package client\n")
		assert.False(t, storeUnderDetected(dir, collectionPaths, apispec.KindREST, false))
	})

	t.Run("HTML page-mode sync stub silences the guard", func(t *testing.T) {
		dir := newDir(t)
		writeStubFile(t, filepath.Join(dir, "internal", "cli", "sync.go"), "package cli\n\n// sync is not implemented for this CLI; write internal/cli/sync_pages.go because the generic spec-driven sync template does not fit predominantly HTML page-mode endpoints\n")
		assert.False(t, storeUnderDetected(dir, collectionPaths, apispec.KindREST, false))
	})

	t.Run("device CLI silences the guard", func(t *testing.T) {
		dir := t.TempDir()
		writeDeviceSpecGo(t, dir)
		assert.False(t, storeUnderDetected(dir, collectionPaths, apispec.KindREST, false))
	})

	t.Run("local-datastore manifest silences the guard", func(t *testing.T) {
		dir := newDir(t)
		writeStubFile(t, filepath.Join(dir, CLIManifestFilename), `{"spec_format": "sqlite"}`)
		assert.False(t, storeUnderDetected(dir, collectionPaths, apispec.KindREST, false))
	})

	t.Run("no GET paths silences the guard", func(t *testing.T) {
		assert.False(t, storeUnderDetected(newDir(t), nil, apispec.KindREST, false))
	})

	t.Run("no collection pair silences the guard", func(t *testing.T) {
		assert.False(t, storeUnderDetected(newDir(t), []string{"/v1/load", "/v1/sql"}, apispec.KindREST, false))
	})
}

// TestStoreUnderDetectedCrossLoader pins that the scorecard loader (raw paths
// map) and the verify loader (lenient parsed resources) agree on the guard's
// answer for the same spec bytes. The two loaders build path lists from
// different sources, so only the final answer is asserted, never the path
// sets themselves.
func TestStoreUnderDetectedCrossLoader(t *testing.T) {
	writeSpec := func(t *testing.T, body string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "spec.json")
		require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
		return path
	}

	t.Run("readable collection pair is detected by both loaders", func(t *testing.T) {
		specPath := writeSpec(t, `{
  "openapi": "3.0.3",
  "info": {"title": "Items", "version": "1.0.0"},
  "paths": {
    "/items": {
      "get": {
        "operationId": "listItems",
        "responses": {"200": {"description": "ok"}}
      }
    },
    "/items/{id}": {
      "get": {
        "operationId": "getItem",
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
        "responses": {"200": {"description": "ok"}}
      }
    }
  }
}`)
		dir := t.TempDir()
		writeClientPkgGo(t, dir)

		scorecardSpec, err := loadOpenAPISpec(specPath)
		require.NoError(t, err)
		require.NotNil(t, scorecardSpec)
		verifySpec, err := loadDogfoodOpenAPISpec(specPath, "")
		require.NoError(t, err)
		require.NotNil(t, verifySpec)

		scorecardAnswer := storeUnderDetected(dir, scorecardSpec.GETPaths, scorecardSpec.Kind, scorecardSpec.IsGraphQL)
		verifyAnswer := storeUnderDetected(dir, verifySpec.GETPaths, verifySpec.Kind, false)
		assert.True(t, scorecardAnswer, "scorecard loader must see the GET collection pair")
		assert.Equal(t, scorecardAnswer, verifyAnswer, "both loaders must agree on the guard's answer")
	})

	t.Run("write-only collection is not under-detection via either loader", func(t *testing.T) {
		specPath := writeSpec(t, `{
  "openapi": "3.0.3",
  "info": {"title": "Orders", "version": "1.0.0"},
  "paths": {
    "/orders": {
      "post": {
        "operationId": "createOrder",
        "responses": {"201": {"description": "created"}}
      }
    },
    "/orders/{id}": {
      "delete": {
        "operationId": "deleteOrder",
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
        "responses": {"204": {"description": "deleted"}}
      }
    }
  }
}`)
		dir := t.TempDir()
		writeClientPkgGo(t, dir)

		scorecardSpec, err := loadOpenAPISpec(specPath)
		require.NoError(t, err)
		require.NotNil(t, scorecardSpec)
		assert.False(t, storeUnderDetected(dir, scorecardSpec.GETPaths, scorecardSpec.Kind, scorecardSpec.IsGraphQL),
			"a write-only P + P/{param} pair must not read as a collection")

		verifySpec, err := loadDogfoodOpenAPISpec(specPath, "")
		require.NoError(t, err)
		require.NotNil(t, verifySpec)
		assert.False(t, storeUnderDetected(dir, verifySpec.GETPaths, verifySpec.Kind, false))
	})
}
