package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newPrunePreviewCmd(flags *rootFlags) *cobra.Command {
	var guildID string
	var days int
	var roles string

	cmd := &cobra.Command{
		Use:   "prune-preview",
		Short: "Preview how many members would be pruned without executing",
		Long: `Checks how many members would be removed by a prune operation
without actually executing it. Shows the count of members who have not
been active in N days. Always a dry-run - never modifies the guild.

Uses: GET /guilds/{id}/prune?days=N (preview only, never POST).`,
		Example: `  # Preview 7-day prune
  discord-cli prune-preview --guild 123456789012345678 --days 7

  # Preview 30-day prune with role filter
  discord-cli prune-preview --guild 123456789012345678 --days 30 --roles 111222333444555666

  # JSON output
  discord-cli prune-preview --guild 123456789012345678 --days 7 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if guildID == "" {
				return usageErr(fmt.Errorf("--guild is required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Checking prune count for guild %s (inactive %d+ days)...\n", guildID, days)

			params := map[string]string{"days": fmt.Sprintf("%d", days)}
			if roles != "" {
				params["include_roles"] = roles
			}

			data, err := c.Get(fmt.Sprintf("/guilds/%s/prune", guildID), params)
			if err != nil {
				return classifyAPIError(err)
			}

			var result struct {
				Pruned int `json:"pruned"`
			}
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}

			if flags.asJSON {
				output := map[string]any{
					"guild_id":       guildID,
					"days_inactive":  days,
					"would_prune":    result.Pruned,
					"preview_only":   true,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			if result.Pruned == 0 {
				fmt.Printf("No members would be pruned (all active within %d days)\n", days)
			} else {
				fmt.Printf("%s %d members would be pruned (inactive %d+ days)\n",
					yellow("WARNING:"), result.Pruned, days)
				fmt.Println("This is a preview only. No members were removed.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&guildID, "guild", "", "Guild (server) ID (required)")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days of inactivity")
	cmd.Flags().StringVar(&roles, "roles", "", "Comma-separated role IDs to include in prune count")

	return cmd
}
