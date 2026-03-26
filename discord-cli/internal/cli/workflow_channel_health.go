package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

type channelHealthReport struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RecentCount int    `json:"recent_messages"`
	LastMessage string `json:"last_message_at"`
	Stale       bool   `json:"stale"`
	DaysSilent  int    `json:"days_silent"`
}

func newChannelHealthCmd(flags *rootFlags) *cobra.Command {
	var guildID string
	var days int

	cmd := &cobra.Command{
		Use:   "channel-health",
		Short: "Detect stale channels and show activity trends across a guild",
		Long: `Fetches all text channels in a guild, samples recent messages from each,
and reports activity levels. Identifies stale channels (no messages in N days)
and ranks channels by message volume. Useful for server cleanup and understanding
where your community is most active.

Combines: GET /guilds/{id}/channels + GET /channels/{id}/messages for each channel.`,
		Example: `  # Check channel health for a guild (default: 30 days)
  discord-cli channel-health --guild 123456789012345678

  # Check with custom staleness threshold
  discord-cli channel-health --guild 123456789012345678 --days 14

  # JSON output for agent consumption
  discord-cli channel-health --guild 123456789012345678 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if guildID == "" {
				return usageErr(fmt.Errorf("--guild is required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			c.NoCache = true

			fmt.Fprintf(os.Stderr, "Fetching channels for guild %s...\n", guildID)
			channelsData, err := c.Get(fmt.Sprintf("/guilds/%s/channels", guildID), nil)
			if err != nil {
				return classifyAPIError(err)
			}

			var channels []map[string]any
			if err := json.Unmarshal(channelsData, &channels); err != nil {
				return fmt.Errorf("parsing channels: %w", err)
			}

			var textChannels []map[string]any
			for _, ch := range channels {
				chType, _ := ch["type"].(float64)
				if chType == 0 || chType == 5 {
					textChannels = append(textChannels, ch)
				}
			}

			fmt.Fprintf(os.Stderr, "Found %d text channels. Sampling messages...\n", len(textChannels))
			cutoff := time.Now().AddDate(0, 0, -days)

			var reports []channelHealthReport

			for _, ch := range textChannels {
				chID := fmt.Sprintf("%v", ch["id"])
				chName, _ := ch["name"].(string)

				params := map[string]string{"limit": "5"}
				msgData, err := c.Get(fmt.Sprintf("/channels/%s/messages", chID), params)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  #%s: skipped (access denied or error)\n", chName)
					continue
				}

				var messages []map[string]any
				if err := json.Unmarshal(msgData, &messages); err != nil {
					continue
				}

				report := channelHealthReport{ID: chID, Name: chName}

				if len(messages) == 0 {
					report.Stale = true
					report.DaysSilent = days
				} else {
					if ts, ok := messages[0]["timestamp"].(string); ok {
						if t, err := time.Parse(time.RFC3339, ts); err == nil {
							report.LastMessage = ts
							report.DaysSilent = int(time.Since(t).Hours() / 24)
							report.Stale = t.Before(cutoff)
						}
					}
					for _, msg := range messages {
						if ts, ok := msg["timestamp"].(string); ok {
							if t, err := time.Parse(time.RFC3339, ts); err == nil && t.After(cutoff) {
								report.RecentCount++
							}
						}
					}
				}
				reports = append(reports, report)
			}

			sort.Slice(reports, func(i, j int) bool {
				if reports[i].Stale != reports[j].Stale {
					return !reports[i].Stale
				}
				return reports[i].RecentCount > reports[j].RecentCount
			})

			staleCount := 0
			for _, r := range reports {
				if r.Stale {
					staleCount++
				}
			}

			if flags.asJSON {
				result := map[string]any{
					"guild_id":        guildID,
					"threshold_days":  days,
					"total_channels":  len(reports),
					"stale_channels":  staleCount,
					"active_channels": len(reports) - staleCount,
					"channels":        reports,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("\nChannel Health Report (guild: %s, threshold: %d days)\n", guildID, days)
			fmt.Printf("%-30s  %8s  %12s  %s\n", "CHANNEL", "RECENT", "DAYS SILENT", "STATUS")
			fmt.Printf("%-30s  %8s  %12s  %s\n", "-------", "------", "-----------", "------")
			for _, r := range reports {
				status := green("active")
				if r.Stale {
					status = red("STALE")
				}
				fmt.Printf("%-30s  %8d  %12d  %s\n", "#"+r.Name, r.RecentCount, r.DaysSilent, status)
			}
			fmt.Printf("\nTotal: %d channels (%d active, %d stale)\n",
				len(reports), len(reports)-staleCount, staleCount)

			return nil
		},
	}

	cmd.Flags().StringVar(&guildID, "guild", "", "Guild (server) ID (required)")
	cmd.Flags().IntVar(&days, "days", 30, "Number of days to consider a channel stale")

	return cmd
}
