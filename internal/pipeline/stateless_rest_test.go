package pipeline

import (
	"path/filepath"
	"testing"
)

// TestSpecHasCollectionShapedResource pins the positive spec signal that gates
// the stateless exemption: a list+detail pair (P and P/{id}) is a collection;
// RPC-style action endpoints are not.
func TestSpecHasCollectionShapedResource(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  bool
	}{
		{"RPC actions (Cube shape)", []string{"/v1/load", "/v1/sql", "/v1/meta", "/v1/running-query/{requestId}"}, false},
		{"list+detail pair", []string{"/items", "/items/{id}"}, true},
		{"detail without bare parent", []string{"/items/{id}"}, false},
		{"nested collection pair", []string{"/orgs/{org}/repos", "/orgs/{org}/repos/{id}"}, true},
		{"empty", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := specHasCollectionShapedResource(tc.paths); got != tc.want {
				t.Errorf("specHasCollectionShapedResource(%v) = %v, want %v", tc.paths, got, tc.want)
			}
		})
	}
}

// TestIsStatelessHTTPCLIDir pins the detector that distinguishes a stateless
// REST mirror from a stateful CLI whose sync may merely be broken, and from a
// CLI whose spec has a collection the profiler under-detected.
func TestIsStatelessHTTPCLIDir(t *testing.T) {
	rpcPaths := []string{"/v1/load", "/v1/meta", "/v1/sql"}

	t.Run("HTTP client, no store, no collections -> stateless", func(t *testing.T) {
		dir := t.TempDir()
		writeClientPkgGo(t, dir)
		if !isStatelessHTTPCLIDir(dir, rpcPaths) {
			t.Error("expected an HTTP CLI with no store and RPC-only paths to be stateless")
		}
	})
	t.Run("HTTP client + store/store.go -> not stateless (broken-sync stays tested)", func(t *testing.T) {
		dir := t.TempDir()
		writeClientPkgGo(t, dir)
		writeStubFile(t, filepath.Join(dir, "internal", "store", "store.go"), "package store\n")
		if isStatelessHTTPCLIDir(dir, rpcPaths) {
			t.Error("a CLI with a store layer must stay subject to pipeline checks, not be marked stateless")
		}
	})
	t.Run("spec has collection but no store -> not stateless (profiler under-detection surfaces)", func(t *testing.T) {
		dir := t.TempDir()
		writeClientPkgGo(t, dir)
		if isStatelessHTTPCLIDir(dir, []string{"/items", "/items/{id}"}) {
			t.Error("a spec with a collection resource and no store must NOT be exempted; the missing store should surface")
		}
	})
	t.Run("device CLI -> not stateless", func(t *testing.T) {
		dir := t.TempDir()
		writeDeviceSpecGo(t, dir)
		if isStatelessHTTPCLIDir(dir, rpcPaths) {
			t.Error("a device CLI is not a stateless HTTP CLI")
		}
	})
	t.Run("no HTTP client -> not stateless", func(t *testing.T) {
		if isStatelessHTTPCLIDir(t.TempDir(), rpcPaths) {
			t.Error("a dir with no HTTP client must not be detected as a stateless HTTP CLI")
		}
	})
}

// TestDataPipelineSkipReason covers the verify-side gate: device and stateless
// REST CLIs SKIP with the right detail; a stateful CLI runs the real test.
func TestDataPipelineSkipReason(t *testing.T) {
	t.Run("stateless REST -> SKIP", func(t *testing.T) {
		dir := t.TempDir()
		writeClientPkgGo(t, dir)
		detail, skip := dataPipelineSkipReason(dir, []string{"/v1/load"})
		if !skip || detail != "SKIP (stateless REST CLI: no sync data pipeline)" {
			t.Errorf("stateless gate = (%q, %v), want SKIP stateless detail", detail, skip)
		}
	})
	t.Run("device -> SKIP", func(t *testing.T) {
		dir := t.TempDir()
		writeDeviceSpecGo(t, dir)
		detail, skip := dataPipelineSkipReason(dir, nil)
		if !skip || detail != "SKIP (device CLI: no sync data pipeline)" {
			t.Errorf("device gate = (%q, %v), want SKIP device detail", detail, skip)
		}
	})
	t.Run("stateful (has store) -> run real test", func(t *testing.T) {
		dir := t.TempDir()
		writeClientPkgGo(t, dir)
		writeStubFile(t, filepath.Join(dir, "internal", "store", "store.go"), "package store\n")
		if _, skip := dataPipelineSkipReason(dir, nil); skip {
			t.Error("a stateful CLI must run the real data-pipeline test, not be skipped")
		}
	})
}

// TestScorecardStatelessMarksPipelineDimsUnscored proves a stateless REST mirror
// marks only the store-gated DataPipelineIntegrity and SyncCorrectness N/A
// (unscored), while Workflows stays SCORED. Workflows is not store-gated (it
// credits a "load"-prefixed compound command), so N/A-ing it would drop it from
// the denominator while scoreInfrastructureDimensions still counted it in the
// numerator — inflating the score. The fixture includes a workflow-shaped
// command file to make that leak visible if Workflows were wrongly N/A'd.
func TestScorecardStatelessMarksPipelineDimsUnscored(t *testing.T) {
	dir := t.TempDir()
	writeClientPkgGo(t, dir)
	writeStubFile(t, filepath.Join(dir, "internal", "cli", "root.go"), "package cli\n")
	writeStubFile(t, filepath.Join(dir, "internal", "cli", "load.go"), "package cli\n// newLoadCmd builds the \"load\" command.\n")

	sc := &Scorecard{}
	spec := &openAPISpecInfo{Paths: []string{"/v1/load", "/v1/meta"}}
	scoreDomainDimensions(sc, dir, spec, nil, false)

	for _, dim := range []string{DimDataPipelineIntegrity, DimSyncCorrectness} {
		if !sc.IsDimensionUnscored(dim) {
			t.Errorf("stateless REST CLI should mark store-gated %q N/A (unscored), got scored", dim)
		}
	}
	if sc.IsDimensionUnscored(DimWorkflows) {
		t.Error("Workflows is not store-gated and must stay scored for a stateless mirror, not N/A (N/A would inflate the score)")
	}
}

// TestScorecardStatefulStillScoresPipeline guards the gate: a CLI carrying a
// store layer must NOT be treated as stateless — its pipeline/sync dimensions
// stay scored so a missing or broken sync is still caught.
func TestScorecardStatefulStillScoresPipeline(t *testing.T) {
	dir := t.TempDir()
	writeClientPkgGo(t, dir)
	writeStubFile(t, filepath.Join(dir, "internal", "cli", "root.go"), "package cli\n")
	writeStubFile(t, filepath.Join(dir, "internal", "store", "store.go"), "package store\n")

	sc := &Scorecard{}
	scoreDomainDimensions(sc, dir, &openAPISpecInfo{}, nil, false)

	if sc.IsDimensionUnscored(DimDataPipelineIntegrity) {
		t.Error("a CLI with a store layer must keep DataPipelineIntegrity scored, not N/A")
	}
	if sc.IsDimensionUnscored(DimWorkflows) {
		t.Error("a stateful CLI must keep Workflows scored, not N/A")
	}
}

// TestScorecardCollectionSpecStillScoresPipeline guards Problem 1: a spec with a
// real collection resource that emitted no store (profiler under-detection) must
// NOT be exempted — the missing pipeline must surface, not pass silently.
func TestScorecardCollectionSpecStillScoresPipeline(t *testing.T) {
	dir := t.TempDir()
	writeClientPkgGo(t, dir)
	writeStubFile(t, filepath.Join(dir, "internal", "cli", "root.go"), "package cli\n")

	sc := &Scorecard{}
	spec := &openAPISpecInfo{Paths: []string{"/items", "/items/{id}"}}
	scoreDomainDimensions(sc, dir, spec, nil, false)

	if sc.IsDimensionUnscored(DimDataPipelineIntegrity) {
		t.Error("a spec with a collection resource and no store must keep DataPipelineIntegrity scored, surfacing the gap")
	}
}
