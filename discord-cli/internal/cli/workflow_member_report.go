package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

func newMemberReportCmd(flags *rootFlags) *cobra.Command {
	var guildID string
	var maxMembers int

	cmd := &cobra.Command{
		Use:   "member-report",
		Short: "Analyze member roles, join dates, and activity across a guild",
		Long: `Fetches guild members and produces a report showing role distribution,
join date timeline, and member counts per role. Useful for understanding
your community composition and identifying role bloat.

Combines: GET /guilds/{id}/members + GET /guilds/{id}/roles.`,
		Example: `  # Member report for a guild
  discord-cli member-report --guild 123456789012345678

  # Fetch more members (default: 100)
  discord-cli member-report --guild 123456789012345678 --max 500

  # JSON output
  discord-cli member-report --guild 123456789012345678 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if guildID == "" {
				return usageErr(fmt.Errorf("--guild is required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			c.NoCache = true

			// Fetch roles
			fmt.Fprintf(os.Stderr, "Fetching roles...\n")
			rolesData, err := c.Get(fmt.Sprintf("/guilds/%s/roles", guildID), nil)
			if err != nil {
				return classifyAPIError(err)
			}
			var roles []map[string]any
			json.Unmarshal(rolesData, &roles)

			roleNames := map[string]string{}
			for _, r := range roles {
				id := fmt.Sprintf("%v", r["id"])
				name, _ := r["name"].(string)
				roleNames[id] = name
			}

			// Fetch members
			fmt.Fprintf(os.Stderr, "Fetching members (max %d)...\n", maxMembers)
			params := map[string]string{"limit": fmt.Sprintf("%d", maxMembers)}
			membersData, err := c.Get(fmt.Sprintf("/guilds/%s/members", guildID), params)
			if err != nil {
				return classifyAPIError(err)
			}

			var members []map[string]any
			if err := json.Unmarshal(membersData, &members); err != nil {
				return fmt.Errorf("parsing members: %w", err)
			}

			// Count members per role
			roleCounts := map[string]int{}
			botCount := 0
			for _, m := range members {
				user, _ := m["user"].(map[string]any)
				if user != nil {
					if bot, ok := user["bot"].(bool); ok && bot {
						botCount++
					}
				}
				if roleIDs, ok := m["roles"].([]any); ok {
					for _, rid := range roleIDs {
						roleID := fmt.Sprintf("%v", rid)
						name := roleNames[roleID]
						if name == "" {
							name = roleID
						}
						roleCounts[name]++
					}
				}
			}

			if flags.asJSON {
				result := map[string]any{
					"guild_id":      guildID,
					"total_members": len(members),
					"bot_count":     botCount,
					"human_count":   len(members) - botCount,
					"total_roles":   len(roles),
					"role_counts":   roleCounts,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("\nMember Report (guild: %s)\n\n", guildID)
			fmt.Printf("Total Members: %d (%d humans, %d bots)\n", len(members), len(members)-botCount, botCount)
			fmt.Printf("Total Roles: %d\n\n", len(roles))

			fmt.Println("Members per Role:")
			type kv struct {
				Role  string
				Count int
			}
			var sorted []kv
			for k, v := range roleCounts {
				sorted = append(sorted, kv{k, v})
			}
			sort.Slice(sorted, func(i, j int) bool { return sorted[i].Count > sorted[j].Count })
			for _, item := range sorted {
				fmt.Printf("  %-30s %d members\n", item.Role, item.Count)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&guildID, "guild", "", "Guild (server) ID (required)")
	cmd.Flags().IntVar(&maxMembers, "max", 100, "Maximum members to fetch")

	return cmd
}
