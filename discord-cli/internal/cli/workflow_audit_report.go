package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

func newAuditReportCmd(flags *rootFlags) *cobra.Command {
	var guildID string
	var actionType string
	var userID string
	var limit int

	cmd := &cobra.Command{
		Use:   "audit-report",
		Short: "Analyze audit log entries grouped by action type, user, and date",
		Long: `Fetches the audit log for a guild and produces a summary report.
Groups entries by action type and user, showing who did what and when.
Useful for security reviews, moderation tracking, and compliance.

Combines: GET /guilds/{id}/audit-logs with analysis and grouping.`,
		Example: `  # Full audit report for a guild
  discord-cli audit-report --guild 123456789012345678

  # Filter by action type (e.g., MEMBER_BAN_ADD=22)
  discord-cli audit-report --guild 123456789012345678 --action 22

  # Filter by user who performed the action
  discord-cli audit-report --guild 123456789012345678 --user 987654321098765432

  # JSON output
  discord-cli audit-report --guild 123456789012345678 --json --limit 100`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if guildID == "" {
				return usageErr(fmt.Errorf("--guild is required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			c.NoCache = true

			fmt.Fprintf(os.Stderr, "Fetching audit log for guild %s...\n", guildID)
			params := map[string]string{"limit": fmt.Sprintf("%d", limit)}
			if actionType != "" {
				params["action_type"] = actionType
			}
			if userID != "" {
				params["user_id"] = userID
			}

			data, err := c.Get(fmt.Sprintf("/guilds/%s/audit-logs", guildID), params)
			if err != nil {
				return classifyAPIError(err)
			}

			var auditLog struct {
				Entries []map[string]any `json:"audit_log_entries"`
				Users   []map[string]any `json:"users"`
			}
			if err := json.Unmarshal(data, &auditLog); err != nil {
				return fmt.Errorf("parsing audit log: %w", err)
			}

			// Build user lookup
			userNames := map[string]string{}
			for _, u := range auditLog.Users {
				id := fmt.Sprintf("%v", u["id"])
				username, _ := u["username"].(string)
				userNames[id] = username
			}

			// Group by action type
			actionCounts := map[string]int{}
			userActionCounts := map[string]int{}
			for _, entry := range auditLog.Entries {
				action := fmt.Sprintf("%v", entry["action_type"])
				actionCounts[action]++

				uid := fmt.Sprintf("%v", entry["user_id"])
				name := userNames[uid]
				if name == "" {
					name = uid
				}
				userActionCounts[name]++
			}

			if flags.asJSON {
				result := map[string]any{
					"guild_id":      guildID,
					"total_entries": len(auditLog.Entries),
					"by_action":     actionCounts,
					"by_user":       userActionCounts,
					"entries":       auditLog.Entries,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("\nAudit Report (guild: %s, %d entries)\n\n", guildID, len(auditLog.Entries))

			fmt.Println("By Action Type:")
			printSortedCounts(actionCounts)

			fmt.Println("\nBy User:")
			printSortedCounts(userActionCounts)

			return nil
		},
	}

	cmd.Flags().StringVar(&guildID, "guild", "", "Guild (server) ID (required)")
	cmd.Flags().StringVar(&actionType, "action", "", "Filter by action type number")
	cmd.Flags().StringVar(&userID, "user", "", "Filter by user ID who performed the action")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum audit log entries to fetch")

	return cmd
}

func printSortedCounts(counts map[string]int) {
	type kv struct {
		Key   string
		Count int
	}
	var sorted []kv
	for k, v := range counts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Count > sorted[j].Count })
	for _, item := range sorted {
		fmt.Printf("  %-40s %d\n", item.Key, item.Count)
	}
}
