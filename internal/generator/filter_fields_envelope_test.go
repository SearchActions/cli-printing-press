package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFilterFieldsEnvelopeDescent_EmittedHelper guards the runtime behavior
// of filterFields against list-envelope responses inside the cli-printing-press
// repo's own test suite. The function is emitted into every printed CLI's
// internal/cli/helpers.go from helpers.go.tmpl. Without this gate, regressions
// in the envelope-descent fallback only surface when a user runs `go test ./...`
// inside a generated CLI, which slows the feedback loop and risks shipping a
// broken --select to every CLI built from a future bad commit.
//
// The test follows the TestRootFlagsPrintJSONHonorsOutputFlags pattern: it
// generates a CLI to a temp dir, writes a fixture _test.go alongside the
// emitted helpers, then runs `go test` on the generated module.
func TestFilterFieldsEnvelopeDescent_EmittedHelper(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("envelope-descent")
	outputDir := filepath.Join(t.TempDir(), "envelope-descent-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	testPath := filepath.Join(outputDir, "internal", "cli", "filter_fields_envelope_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(`package cli

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// TestFilterFieldsEnvelopeDescent covers the four shapes printed CLIs see in
// practice. The envelope cases pin the regression where wrapper-key + array
// responses returned `+"`{}`"+` because the selector heads matched the inner
// record fields, not the wrapper key.
func TestFilterFieldsEnvelopeDescent(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		fields string
		want   string
	}{
		{
			"bare array element-wise",
			`+"`"+`[{"id":"a","name":"x","other":"y"}]`+"`"+`,
			"id,name",
			`+"`"+`[{"id":"a","name":"x"}]`+"`"+`,
		},
		{
			"envelope single array sibling",
			`+"`"+`{"projects":[{"id":"a","name":"x","other":"y"}]}`+"`"+`,
			"id,name",
			`+"`"+`{"projects":[{"id":"a","name":"x"}]}`+"`"+`,
		},
		{
			"envelope with metadata sibling preserves count",
			`+"`"+`{"total_count":2,"items":[{"id":"a","other":"y"}]}`+"`"+`,
			"id",
			`+"`"+`{"items":[{"id":"a"}],"total_count":2}`+"`"+`,
		},
		{
			"envelope preserves null pagination cursor verbatim",
			`+"`"+`{"items":[{"id":"a"}],"next_cursor":null}`+"`"+`,
			"id",
			`+"`"+`{"items":[{"id":"a"}],"next_cursor":null}`+"`"+`,
		},
		{
			"flat object no match preserves input",
			`+"`"+`{"a":1,"b":2}`+"`"+`,
			"c",
			`+"`"+`{"a":1,"b":2}`+"`"+`,
		},
		{
			"selector matches envelope key suppresses descent",
			`+"`"+`{"projects":[{"id":"a","other":"y"}]}`+"`"+`,
			"projects",
			`+"`"+`{"projects":[{"id":"a","other":"y"}]}`+"`"+`,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := filterFields(json.RawMessage(tc.input), tc.fields)
			var gotV, wantV interface{}
			if err := json.Unmarshal(got, &gotV); err != nil {
				t.Fatalf("invalid json output: %v (raw=%s)", err, string(got))
			}
			if err := json.Unmarshal([]byte(tc.want), &wantV); err != nil {
				t.Fatalf("invalid want json: %v (raw=%s)", err, tc.want)
			}
			gotBytes, _ := json.Marshal(gotV)
			wantBytes, _ := json.Marshal(wantV)
			if string(gotBytes) != string(wantBytes) {
				t.Errorf("filterFields(%q, %q) = %s, want %s",
					tc.input, tc.fields, string(gotBytes), string(wantBytes))
			}
		})
	}
}

func TestFilterFieldsEnvelopeDescent_UnknownSelector(t *testing.T) {
	input := "{\"items\":[{\"id\":\"a\",\"name\":\"Alpha\"},{\"id\":\"b\",\"name\":\"Beta\"}]}"
	got, warning := filterFieldsWithWarning(t, input, "missing")

	var gotV, wantV interface{}
	if err := json.Unmarshal(got, &gotV); err != nil {
		t.Fatalf("invalid json output: %v (raw=%s)", err, string(got))
	}
	if err := json.Unmarshal([]byte(input), &wantV); err != nil {
		t.Fatalf("invalid input json: %v", err)
	}
	gotBytes, _ := json.Marshal(gotV)
	wantBytes, _ := json.Marshal(wantV)
	if string(gotBytes) != string(wantBytes) {
		t.Fatalf("unknown selector changed the payload: got %s, want %s", gotBytes, wantBytes)
	}
	if !strings.Contains(string(warning), "--select \"missing\" matched no fields") {
		t.Fatalf("warning = %q, want unknown-selector warning", warning)
	}
	if !strings.Contains(string(warning), "valid fields: items") {
		t.Fatalf("warning = %q, want valid top-level fields", warning)
	}
}

func TestFilterFieldsEnvelopeDescent_EmptyCollectionsDoNotWarn(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		fields string
		want   string
	}{
		{"top-level array", `+"`"+`[]`+"`"+`, "id", `+"`"+`[]`+"`"+`},
		{"list envelope", `+"`"+`{"items":[]}`+"`"+`, "id", `+"`"+`{"items":[]}`+"`"+`},
		{"known dotted head", `+"`"+`{"events":[],"other":1}`+"`"+`, "events.name", `+"`"+`{"events":[]}`+"`"+`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, warning := filterFieldsWithWarning(t, tc.input, tc.fields)
			if string(warning) != "" {
				t.Fatalf("warning = %q, want no warning for an empty collection", warning)
			}
			assertJSONEqual(t, got, tc.want)
		})
	}
}

func TestFilterFieldsEnvelopeDescent_PartiallyInvalidSelectorWarns(t *testing.T) {
	input := `+"`"+`[{"id":"a","name":"Alpha"}]`+"`"+`
	got, warning := filterFieldsWithWarning(t, input, "id,naem")

	assertJSONEqual(t, got, `+"`"+`[{"id":"a"}]`+"`"+`)
	if !strings.Contains(string(warning), "--select \"naem\" matched no fields") {
		t.Fatalf("warning = %q, want warning naming the unmatched selector", warning)
	}
	if strings.Contains(string(warning), "--select \"id\" matched no fields") {
		t.Fatalf("warning = %q, valid selector id must not be reported", warning)
	}
}

func TestFilterFieldsEnvelopeDescent_EmptyEnvelopeSelectorWarnings(t *testing.T) {
	input := `+"`"+`{"items":[]}`+"`"+`
	cases := []struct {
		name           string
		fields         string
		wantWarnings   []string
		forbidWarnings []string
	}{
		{
			name:           "known prefix and unrelated typo",
			fields:         "items.id,naem",
			wantWarnings:   []string{"naem"},
			forbidWarnings: []string{"items.id"},
		},
		{
			name:         "multiple unrelated selectors",
			fields:       "naem,missing",
			wantWarnings: []string{"naem", "missing"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, warning := filterFieldsWithWarning(t, input, tc.fields)
			assertJSONEqual(t, got, input)
			for _, field := range tc.wantWarnings {
				if !strings.Contains(string(warning), "--select \""+field+"\" matched no fields") {
					t.Fatalf("warning = %q, want warning naming %q", warning, field)
				}
			}
			for _, field := range tc.forbidWarnings {
				if strings.Contains(string(warning), "--select \""+field+"\" matched no fields") {
					t.Fatalf("warning = %q, indeterminate selector %q must not be reported", warning, field)
				}
			}
		})
	}
}

// TestFilterFieldsEnvelopeDescent_MixedSelectors pins the composition rule:
// every --select token that resolves alone also resolves in combination.
// The guarded regression: one sibling token matching a top-level key used
// to suppress list-envelope descent for every other token, silently
// dropping fields while the per-token warning pass believed they matched.
func TestFilterFieldsEnvelopeDescent_MixedSelectors(t *testing.T) {
	cases := []struct {
		name         string
		fields       string
		input        string
		want         string
		wantWarnings []string
	}{
		{"sibling top-level match does not suppress envelope descent", "domain,mailboxes",
			`+"`"+`{"fleet":[{"domain":"a.io","status":"ok"}],"mailboxes":0}`+"`"+`,
			`+"`"+`{"fleet":[{"domain":"a.io"}],"mailboxes":0}`+"`"+`, nil},
		{"token order does not change the result", "mailboxes,domain",
			`+"`"+`{"fleet":[{"domain":"a.io","status":"ok"}],"mailboxes":0}`+"`"+`,
			`+"`"+`{"fleet":[{"domain":"a.io"}],"mailboxes":0}`+"`"+`, nil},
		{"dotted path and sibling whole-value", "fleet.domain,mailboxes",
			`+"`"+`{"fleet":[{"domain":"a.io","status":"ok"}],"mailboxes":0}`+"`"+`,
			`+"`"+`{"fleet":[{"domain":"a.io"}],"mailboxes":0}`+"`"+`, nil},
		{"nested object sibling is not an envelope", "domain,mailboxes",
			`+"`"+`{"fleet":{"domain":"a.io"},"mailboxes":0}`+"`"+`,
			`+"`"+`{"mailboxes":0}`+"`"+`, []string{"domain"}},
		{"collision keeps subpath and fallback token in one array", "items.name,id",
			`+"`"+`{"items":[{"id":1,"name":"x","z":0}]}`+"`"+`,
			`+"`"+`{"items":[{"id":1,"name":"x"}]}`+"`"+`, nil},
		{"whole-value collision keeps the original array", "items,id",
			`+"`"+`{"items":[{"id":1,"name":"x","z":0}]}`+"`"+`,
			`+"`"+`{"items":[{"id":1,"name":"x","z":0}]}`+"`"+`, nil},
		{"unmatched path warns beside a valid whole-value", "total,naem",
			`+"`"+`{"items":[{"id":1}],"total":1}`+"`"+`,
			`+"`"+`{"total":1}`+"`"+`, []string{"naem"}},
		{"kebab token counts as matched head", "order-date,domain",
			`+"`"+`{"orderDate":"x","fleet":[{"domain":"a.io"}]}`+"`"+`,
			`+"`"+`{"orderDate":"x","fleet":[{"domain":"a.io"}]}`+"`"+`, nil},
		{"empty fallback array does not count as a contribution", "domain,mailboxes",
			`+"`"+`{"fleet":[],"mailboxes":0}`+"`"+`,
			`+"`"+`{"mailboxes":0}`+"`"+`, []string{"domain"}},
		{"empty fallback array beside whole-value warns on typo", "total,naem",
			`+"`"+`{"items":[],"total":0}`+"`"+`,
			`+"`"+`{"total":0}`+"`"+`, []string{"naem"}},
		{"record-level narrowing inside a bare array", "number,name",
			`+"`"+`[{"number":1,"title":"t","body":"b","labels":[{"name":"bug","color":"red"}]}]`+"`"+`,
			`+"`"+`[{"number":1,"labels":[{"name":"bug"}]}]`+"`"+`, nil},
		{"embedded collision merges subpath and fallback token", "_embedded.count,id",
			`+"`"+`{"_embedded":{"count":2,"items":[{"id":1,"z":0}]}}`+"`"+`,
			`+"`"+`{"_embedded":{"count":2,"items":[{"id":1}]}}`+"`"+`, nil},
		{"anchored empty collision with unmatched typo", "items.id,naem",
			`+"`"+`{"items":[]}`+"`"+`,
			`+"`"+`{"items":[]}`+"`"+`, []string{"naem"}},
		{"embedded narrowing drops non-qualifying sub-keys", "id,title",
			`+"`"+`{"id":1,"_embedded":{"author":{"n":"a"},"comments":[{"title":"t","z":0}]}}`+"`"+`,
			`+"`"+`{"id":1,"_embedded":{"comments":[{"title":"t"}]}}`+"`"+`, nil},
		{"duplicate spellings union their sub-paths", "orderdate.a,order-date.b",
			`+"`"+`{"orderDate":{"a":1,"b":2,"c":3}}`+"`"+`,
			`+"`"+`{"orderDate":{"a":1,"b":2}}`+"`"+`, nil},
		{"duplicate spelling whole-value wins", "orderdate.a,order-date",
			`+"`"+`{"orderDate":{"a":1,"b":2}}`+"`"+`,
			`+"`"+`{"orderDate":{"a":1,"b":2}}`+"`"+`, nil},
		{"duplicate spelling whole-value suppresses descent", "orderdate,order-date",
			`+"`"+`{"orderDate":"x","fleet":[{"orderDate":"y","z":0}]}`+"`"+`,
			`+"`"+`{"orderDate":"x"}`+"`"+`, nil},
		{"heterogeneous records pick up envelope descent per record", "id,name",
			`+"`"+`[{"id":1,"name":"a","tags":[{"name":"t"}]},{"id":2,"tags":[{"name":"u"}]}]`+"`"+`,
			`+"`"+`[{"id":1,"name":"a"},{"id":2,"tags":[{"name":"u"}]}]`+"`"+`, nil},
		{"second array sibling that matches nothing is dropped", "domain,mailboxes",
			`+"`"+`{"fleet":[{"domain":"a"}],"pending":[{"x":1}],"mailboxes":0}`+"`"+`,
			`+"`"+`{"fleet":[{"domain":"a"}],"mailboxes":0}`+"`"+`, nil},
		{"metadata array is never a descent target", "domain,mailboxes",
			`+"`"+`{"errors":[{"domain":"e"}],"mailboxes":0}`+"`"+`,
			`+"`"+`{"mailboxes":0}`+"`"+`, []string{"domain"}},
		{"null array sibling is not an envelope", "domain,mailboxes",
			`+"`"+`{"fleet":null,"mailboxes":0}`+"`"+`,
			`+"`"+`{"mailboxes":0}`+"`"+`, []string{"domain"}},
		{"partly matching elements keep empty objects", "domain,mailboxes",
			`+"`"+`{"fleet":[{"domain":"a"},{"status":"x"}],"mailboxes":0}`+"`"+`,
			`+"`"+`{"fleet":[{"domain":"a"},{}],"mailboxes":0}`+"`"+`, nil},
		{"scalar array sibling does not resolve the token", "domain,mailboxes",
			`+"`"+`{"tags":["a"],"mailboxes":0}`+"`"+`,
			`+"`"+`{"mailboxes":0}`+"`"+`, []string{"domain"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, warning := filterFieldsWithWarning(t, tc.input, tc.fields)
			assertJSONEqual(t, got, tc.want)
			for _, field := range strings.Split(tc.fields, ",") {
				field = strings.TrimSpace(field)
				msg := "--select \"" + field + "\" matched no fields"
				wantWarned := false
				for _, w := range tc.wantWarnings {
					if w == field {
						wantWarned = true
						break
					}
				}
				if warned := strings.Contains(string(warning), msg); warned != wantWarned {
					t.Fatalf("warning = %q, %q warned=%v, want %v", warning, field, warned, wantWarned)
				}
			}
			if n := strings.Count(string(warning), "warning: "); n != len(tc.wantWarnings) {
				t.Fatalf("warning = %q, want exactly %d warning line(s), got %d", warning, len(tc.wantWarnings), n)
			}
		})
	}
}

func filterFieldsWithWarning(t *testing.T, input, fields string) (json.RawMessage, []byte) {
	t.Helper()
	oldStderr := os.Stderr
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stderr = write
	got := filterFields(json.RawMessage(input), fields)
	_ = write.Close()
	os.Stderr = oldStderr
	warning, _ := io.ReadAll(read)
	_ = read.Close()
	return got, warning
}

func assertJSONEqual(t *testing.T, got json.RawMessage, want string) {
	t.Helper()
	var gotV, wantV interface{}
	if err := json.Unmarshal(got, &gotV); err != nil {
		t.Fatalf("invalid json output: %v (raw=%s)", err, string(got))
	}
	if err := json.Unmarshal([]byte(want), &wantV); err != nil {
		t.Fatalf("invalid want json: %v (raw=%s)", err, want)
	}
	gotBytes, _ := json.Marshal(gotV)
	wantBytes, _ := json.Marshal(wantV)
	if string(gotBytes) != string(wantBytes) {
		t.Fatalf("filterFields output = %s, want %s", gotBytes, wantBytes)
	}
}
`), 0o644))

	runGoCommand(t, outputDir, "test", "./internal/cli", "-run", "^(TestFilterFieldsEnvelopeDescent|TestFilterFieldsEnvelopeDescent_UnknownSelector|TestFilterFieldsEnvelopeDescent_EmptyCollectionsDoNotWarn|TestFilterFieldsEnvelopeDescent_PartiallyInvalidSelectorWarns|TestFilterFieldsEnvelopeDescent_EmptyEnvelopeSelectorWarnings|TestFilterFieldsEnvelopeDescent_MixedSelectors)$", "-count=1")
}
