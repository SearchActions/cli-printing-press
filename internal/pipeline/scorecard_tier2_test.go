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
