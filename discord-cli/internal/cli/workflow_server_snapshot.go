package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func newServerSnapshotCmd(flags *rootFlags) *cobra.Command {
	var guildID string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "server-snapshot",
		Short: "Backup guild configuration (roles, channels, emojis) to JSON",
		Long: `Creates a complete snapshot of a guild's configuration including roles,
channels, emojis, and guild settings. Exports to a JSON file that can be
used for backup, migration planning, or configuration diffing.

Combines: GET /guilds/{id} + GET /guilds/{id}/channels + GET /guilds/{id}/roles + GET /guilds/{id}/emojis.`,
		Example: `  # Snapshot to file
  discord-cli server-snapshot --guild 123456789012345678 -o backup.json

  # Snapshot to stdout as JSON
  discord-cli server-snapshot --guild 123456789012345678 --json

  # Snapshot with dry-run to see what would be fetched
  discord-cli server-snapshot --guild 123456789012345678 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if guildID == "" {
				return usageErr(fmt.Errorf("--guild is required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			c.NoCache = true

			snapshot := map[string]any{
				"snapshot_at": time.Now().UTC().Format(time.RFC3339),
				"guild_id":    guildID,
			}

			// Fetch guild info
			fmt.Fprintf(os.Stderr, "Fetching guild info...\n")
			guildData, err := c.Get(fmt.Sprintf("/guilds/%s", guildID), map[string]string{"with_counts": "true"})
			if err != nil {
				return classifyAPIError(err)
			}
			var guild map[string]any
			json.Unmarshal(guildData, &guild)
			snapshot["guild"] = guild

			// Fetch channels
			fmt.Fprintf(os.Stderr, "Fetching channels...\n")
			channelsData, err := c.Get(fmt.Sprintf("/guilds/%s/channels", guildID), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  warning: channels: %v\n", err)
			} else {
				var channels []any
				json.Unmarshal(channelsData, &channels)
				snapshot["channels"] = channels
			}

			// Fetch roles
			fmt.Fprintf(os.Stderr, "Fetching roles...\n")
			rolesData, err := c.Get(fmt.Sprintf("/guilds/%s/roles", guildID), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  warning: roles: %v\n", err)
			} else {
				var roles []any
				json.Unmarshal(rolesData, &roles)
				snapshot["roles"] = roles
			}

			// Fetch emojis
			fmt.Fprintf(os.Stderr, "Fetching emojis...\n")
			emojisData, err := c.Get(fmt.Sprintf("/guilds/%s/emojis", guildID), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  warning: emojis: %v\n", err)
			} else {
				var emojis []any
				json.Unmarshal(emojisData, &emojis)
				snapshot["emojis"] = emojis
			}

			result, _ := json.MarshalIndent(snapshot, "", "  ")

			if outputFile != "" {
				if err := os.WriteFile(outputFile, result, 0o644); err != nil {
					return fmt.Errorf("writing file: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Snapshot saved to %s\n", outputFile)
				return nil
			}

			fmt.Fprintln(os.Stdout, string(result))
			return nil
		},
	}

	cmd.Flags().StringVar(&guildID, "guild", "", "Guild (server) ID (required)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")

	return cmd
}
