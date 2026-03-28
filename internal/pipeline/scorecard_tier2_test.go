package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreDeadCode(t *testing.T) {
	t.Run("penalizes dead flags and helper functions", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/root.go", `
package cli

var flags struct {
	jsonOutput bool
	csvOutput bool
	stdinInput bool
}

func init() {
	rootCmd.Flags().BoolVar(&flags.jsonOutput, "json", false, "JSON output")
	rootCmd.Flags().BoolVar(&flags.csvOutput, "csv", false, "CSV output")
	rootCmd.Flags().BoolVar(&flags.stdinInput, "stdin", false, "Read stdin")
}
`)
		writeScorecardFixture(t, dir, "internal/cli/messages.go", `
package cli

func runMessages() {
	if flags.jsonOutput {
		println("json")
	}
}
`)
		writeScorecardFixture(t, dir, "internal/cli/helpers.go", `
package cli

func filterFields() {}

func outputCSV() {}
`)

		// 2 dead flags (csvOutput, stdinInput), 2 dead functions (filterFields, outputCSV)
		assert.Equal(t, 1, scoreDeadCode(dir))
	})

	t.Run("returns full score when nothing is dead", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/root.go", `
package cli

var flags struct {
	jsonOutput bool
}

func init() {
	rootCmd.Flags().BoolVar(&flags.jsonOutput, "json", false, "JSON output")
}
`)
		writeScorecardFixture(t, dir, "internal/cli/messages.go", `
package cli

func runMessages() {
	if flags.jsonOutput {
		println("json")
	}
}
`)

		assert.Equal(t, 5, scoreDeadCode(dir))
	})
}

func TestScoreDataPipelineIntegrity(t *testing.T) {
	t.Run("scores generic store methods and tables low", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/sync.go", `
package cli

import "example.com/project/internal/store"

func runSync(db *store.DB) {
	db.Upsert("messages", nil)
}
`)
		writeScorecardFixture(t, dir, "internal/cli/search.go", `
package cli

func runSearch(db interface{ Search(string) error }) {
	_ = db.Search("term")
}
`)
		writeScorecardFixture(t, dir, "internal/store/store.go", `
package store

const schema = "`+`
CREATE TABLE sync_records (
	id TEXT,
	data JSON,
	synced_at TEXT
);
`+`"
`)

		assert.Equal(t, 1, scoreDataPipelineIntegrity(dir))
	})

	t.Run("scores domain specific pipelines high", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/sync.go", `
package cli

func runSync(db interface {
	UpsertMessage(any) error
	UpsertChannel(any) error
}) {
	_ = db.UpsertMessage(nil)
	_ = db.UpsertChannel(nil)
}
`)
		writeScorecardFixture(t, dir, "internal/cli/search.go", `
package cli

func runSearch(db interface{ SearchMessages(string) error }) {
	_ = db.SearchMessages("hello")
}
`)
		writeScorecardFixture(t, dir, "internal/store/store.go", `
package store

const schema = "`+`
CREATE TABLE messages (
	id TEXT,
	channel_id TEXT,
	author_id TEXT,
	content TEXT,
	created_at TEXT,
	updated_at TEXT
);

CREATE TABLE channels (
	id TEXT,
	guild_id TEXT,
	name TEXT,
	type TEXT,
	position INTEGER,
	synced_at TEXT
);
`+`"
`)

		assert.Equal(t, 9, scoreDataPipelineIntegrity(dir))
	})
}

func TestScoreSyncCorrectness(t *testing.T) {
	t.Run("scores empty resource selection and missing state tracking low", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/sync.go", `
package cli

func defaultSyncResources() []string {
	return []string{}
}

func runSync(resource string) string {
	path := "/" + resource
	return path
}
`)

		assert.LessOrEqual(t, scoreSyncCorrectness(dir), 3)
	})

	t.Run("scores resource defaults pagination and sync state high", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/sync.go", `
package cli

func defaultSyncResources() []string {
	return []string{"channels", "messages"}
}

func runSync(store interface {
	GetSyncState(string) string
	SaveSyncState(string, string)
}) {
	path := "/guilds/{guild_id}/messages"
	cursor := store.GetSyncState("messages")
	paginatedGet(path, cursor)
	store.SaveSyncState("messages", "next")
}
`)

		assert.Equal(t, 10, scoreSyncCorrectness(dir))
	})
}

func TestScorePathValidity(t *testing.T) {
	t.Run("matches short variable path declarations used by generated commands", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/links.go", `
package cli

func runLinks() string {
	path := "/links"
	return path
}
`)

		specPath := filepath.Join(dir, "spec.json")
		writeScorecardFixture(t, dir, "spec.json", `{
  "paths": {
    "/links": {}
  },
  "components": {
    "securitySchemes": {}
  }
}`)

		assert.Equal(t, 10, scorePathValidity(dir, specPath))
	})
}

func TestScoreTypeFidelity(t *testing.T) {
	t.Run("scores wrong id flag types and dummy guards low", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/messages.go", `
package cli

import "strings"

var _ = strings.ReplaceAll

func init() {
	cmd := messagesCmd
	cmd.Flags().IntVar(&flagAfterID, "after-id", 0, "After")
}
`)

		assert.Equal(t, 0, scoreTypeFidelity(dir))
	})

	t.Run("scores string id flags required markers and clear descriptions high", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/messages.go", `
package cli

func init() {
	cmd := messagesCmd
	cmd.Flags().StringVar(&flagAfterID, "after-id", "", "Snowflake ID to fetch results after the given message")
	cmd.Flags().StringVar(&flagChannelID, "channel-id", "", "Channel ID containing the messages to fetch for sync")
	cmd.Flags().StringVar(&flagGuildID, "guild-id", "", "Guild ID used to scope channel and message syncing")
	_ = cmd.MarkFlagRequired("after-id")
	_ = cmd.MarkFlagRequired("channel-id")
	_ = cmd.MarkFlagRequired("guild-id")
}
`)

		assert.GreaterOrEqual(t, scoreTypeFidelity(dir), 4)
	})
}

func TestScoreSyncCorrectness_NonSyncFilename(t *testing.T) {
	t.Run("finds sync patterns in non-sync.go files", func(t *testing.T) {
		dir := t.TempDir()

		// Sync logic lives in channel_workflow.go, not sync.go
		writeScorecardFixture(t, dir, "internal/cli/channel_workflow.go", `
package cli

func defaultSyncResources() []string {
	return []string{"bookings", "event_types"}
}

func runChannelSync(store interface {
	GetSyncState(string) string
	SaveSyncState(string, string)
}) {
	path := "/v2/bookings"
	cursor := store.GetSyncState("bookings")
	paginatedGet(path, cursor)
	store.SaveSyncState("bookings", "next")
}
`)

		score := scoreSyncCorrectness(dir)
		assert.GreaterOrEqual(t, score, 7, "sync logic in non-sync.go should score high")
	})
}

func TestScoreDataPipelineIntegrity_NonSyncFilename(t *testing.T) {
	t.Run("finds upsert patterns in non-sync.go files", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/sync_cmd.go", `
package cli

import "example.com/project/internal/store"

func runSync(db *store.DB) {
	_ = db.UpsertBooking(nil)
}
`)
		writeScorecardFixture(t, dir, "internal/cli/search_cmd.go", `
package cli

func runSearch(db interface{ SearchBookings(string) error }) {
	_ = db.SearchBookings("term")
}
`)
		writeScorecardFixture(t, dir, "internal/store/store.go", `
package store

const schema = `+"`"+`
CREATE TABLE bookings (
	id TEXT,
	user_id TEXT,
	event_type_id TEXT,
	title TEXT,
	start_time TEXT,
	end_time TEXT
);
`+"`"+`
`)

		score := scoreDataPipelineIntegrity(dir)
		assert.GreaterOrEqual(t, score, 7, "domain upserts in non-sync.go should score high")
	})
}

func TestScoreDeadCode_FlagsPassedAsArg(t *testing.T) {
	t.Run("flags struct passed to function counts all fields as used", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/root.go", `
package cli

var flags struct {
	asJSON   bool
	csvOutput bool
	verbose  bool
}

func init() {
	rootCmd.Flags().BoolVar(&flags.asJSON, "json", false, "JSON output")
	rootCmd.Flags().BoolVar(&flags.csvOutput, "csv", false, "CSV output")
	rootCmd.Flags().BoolVar(&flags.verbose, "verbose", false, "Verbose")
}
`)
		writeScorecardFixture(t, dir, "internal/cli/messages.go", `
package cli

func runMessages() {
	printOutput(cmd, flags, data, statusCode)
}
`)

		assert.Equal(t, 5, scoreDeadCode(dir))
	})
}

func TestScoreDeadCode_IntraFileHelperCalls(t *testing.T) {
	t.Run("helpers calling other helpers are not dead", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/root.go", `
package cli
`)
		writeScorecardFixture(t, dir, "internal/cli/helpers.go", `
package cli

func formatOutput(data interface{}) string {
	return applyFormat(data)
}

func applyFormat(data interface{}) string {
	return ""
}
`)
		writeScorecardFixture(t, dir, "internal/cli/messages.go", `
package cli

func runMessages() {
	formatOutput(data)
}
`)

		assert.Equal(t, 5, scoreDeadCode(dir))
	})
}

func TestScorecard_VerifyCalibration(t *testing.T) {
	t.Run("verify pass rate sets floor on total score", func(t *testing.T) {
		dir := t.TempDir()

		// Minimal CLI that would score low on static analysis
		writeScorecardFixture(t, dir, "internal/cli/root.go", `
package cli
`)
		writeScorecardFixture(t, dir, "internal/cli/helpers.go", `
package cli
`)
		writeScorecardFixture(t, dir, "README.md", `# Test CLI`)

		pipelineDir := t.TempDir()
		verifyReport := &VerifyReport{
			PassRate:     91.0, // PassRate is 0-100, not 0.0-1.0
			Total:        33,
			Passed:       30,
			DataPipeline: true,
			Verdict:      "PASS",
		}

		sc, err := RunScorecard(dir, pipelineDir, "", verifyReport)
		assert.NoError(t, err)
		// int(91.0) * 80 / 100 = 72 floor
		assert.GreaterOrEqual(t, sc.Steinberger.Total, 72)
		assert.Contains(t, sc.Steinberger.CalibrationNote, "verify pass rate")
	})

	t.Run("verify failure caps data pipeline dimension", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/root.go", `package cli`)
		writeScorecardFixture(t, dir, "internal/cli/sync.go", `
package cli

import "example.com/project/internal/store"

func runSync(db *store.DB) {
	_ = db.UpsertBooking(nil)
}

func defaultSyncResources() []string {
	return []string{"bookings"}
}
`)
		writeScorecardFixture(t, dir, "internal/cli/search.go", `
package cli

func runSearch(db interface{ SearchBookings(string) error }) {
	_ = db.SearchBookings("term")
}
`)
		writeScorecardFixture(t, dir, "internal/store/store.go", `
package store

const schema = `+"`"+`
CREATE TABLE bookings (
	id TEXT,
	user_id TEXT,
	event_type_id TEXT,
	title TEXT,
	start_time TEXT,
	end_time TEXT
);
`+"`"+`
`)

		pipelineDir := t.TempDir()
		verifyReport := &VerifyReport{
			PassRate:     50.0, // PassRate is 0-100
			DataPipeline: false,
			Verdict:      "FAIL",
		}

		sc, err := RunScorecard(dir, pipelineDir, "", verifyReport)
		assert.NoError(t, err)
		assert.LessOrEqual(t, sc.Steinberger.DataPipelineIntegrity, 5)
	})

	t.Run("nil verify report has no effect", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/root.go", `package cli`)
		writeScorecardFixture(t, dir, "README.md", `# Test`)

		pipelineDir := t.TempDir()
		sc, err := RunScorecard(dir, pipelineDir, "", nil)
		assert.NoError(t, err)
		assert.Empty(t, sc.Steinberger.CalibrationNote)
	})
}

func TestScoreWorkflows(t *testing.T) {
	t.Run("counts files matching expanded prefixes", func(t *testing.T) {
		dir := t.TempDir()

		// 3 workflow files by prefix
		writeScorecardFixture(t, dir, "internal/cli/stale_tasks.go", `package cli`)
		writeScorecardFixture(t, dir, "internal/cli/agenda.go", `package cli`)
		writeScorecardFixture(t, dir, "internal/cli/conflicts.go", `package cli`)

		assert.GreaterOrEqual(t, scoreWorkflows(dir), 6) // 3 compound commands → 6
	})

	t.Run("skips infra files", func(t *testing.T) {
		dir := t.TempDir()

		// helpers.go should not count as a workflow
		writeScorecardFixture(t, dir, "internal/cli/helpers.go", `package cli`)
		writeScorecardFixture(t, dir, "internal/cli/root.go", `package cli`)

		assert.Equal(t, 0, scoreWorkflows(dir))
	})

	t.Run("detects store-using commands structurally", func(t *testing.T) {
		dir := t.TempDir()

		// File doesn't match any prefix but imports store
		writeScorecardFixture(t, dir, "internal/cli/bookings_report.go", `
package cli

import "example.com/project/internal/store"

func runReport(db *store.DB) {}
`)
		writeScorecardFixture(t, dir, "internal/cli/availability.go", `
package cli

func runAvailability() {
	db := store.Open()
	_ = db
}
`)

		assert.GreaterOrEqual(t, scoreWorkflows(dir), 4) // 2 compound → 4
	})

	t.Run("counts multi-API-call files", func(t *testing.T) {
		dir := t.TempDir()

		// File makes 2+ different API calls
		writeScorecardFixture(t, dir, "internal/cli/transfer.go", `
package cli

func runTransfer() {
	resp1 := c.Get("/source")
	_ = c.Post("/destination", resp1)
}
`)

		assert.GreaterOrEqual(t, scoreWorkflows(dir), 2) // 1 compound → 2
	})
}

func TestScoreInsight(t *testing.T) {
	t.Run("counts files matching expanded prefixes", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/stats.go", `package cli`)
		writeScorecardFixture(t, dir, "internal/cli/health.go", `package cli`)
		writeScorecardFixture(t, dir, "internal/cli/trends.go", `package cli`)

		assert.GreaterOrEqual(t, scoreInsight(dir), 6) // 3 found → 6
	})

	t.Run("skips infra files", func(t *testing.T) {
		dir := t.TempDir()

		writeScorecardFixture(t, dir, "internal/cli/helpers.go", `package cli`)
		writeScorecardFixture(t, dir, "internal/cli/root.go", `package cli`)
		writeScorecardFixture(t, dir, "internal/cli/doctor.go", `package cli`)

		assert.Equal(t, 0, scoreInsight(dir))
	})

	t.Run("detects store plus aggregation structurally", func(t *testing.T) {
		dir := t.TempDir()

		// File uses store AND aggregation — should count as insight
		writeScorecardFixture(t, dir, "internal/cli/usage_report.go", `
package cli

import "example.com/project/internal/store"

func runUsageReport(db *store.DB) {
	rows := db.Query("SELECT COUNT(*) FROM bookings GROUP BY status")
	_ = rows
}
`)

		assert.GreaterOrEqual(t, scoreInsight(dir), 2) // 1 found → 2
	})

	t.Run("store without aggregation does not count", func(t *testing.T) {
		dir := t.TempDir()

		// File uses store but NO aggregation — should not count
		writeScorecardFixture(t, dir, "internal/cli/lookup.go", `
package cli

import "example.com/project/internal/store"

func runLookup(db *store.DB) {
	row := db.Query("SELECT * FROM bookings WHERE id = ?")
	_ = row
}
`)

		assert.Equal(t, 0, scoreInsight(dir))
	})
}

func writeScorecardFixture(t *testing.T, root, relPath, content string) {
	t.Helper()

	path := filepath.Join(root, relPath)
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("mkdir %s: %v", relPath, err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}
