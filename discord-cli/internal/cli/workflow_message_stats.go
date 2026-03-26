package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func newMessageStatsCmd(flags *rootFlags) *cobra.Command {
	var guildID string
	var channelID string
	var limit int

	cmd := &cobra.Command{
		Use:   "message-stats",
		Short: "Show message volume and top contributors for a channel or guild",
		Long: `Fetches recent messages and calculates statistics including message
count per author, average message length, and activity by hour of day.
Can analyze a single channel or sample across all text channels in a guild.

Combines: GET /guilds/{id}/channels + GET /channels/{id}/messages with analysis.`,
		Example: `  # Stats for a specific channel
  discord-cli message-stats --channel 123456789012345678 --limit 100

  # Stats across all channels in a guild (samples 20 messages per channel)
  discord-cli message-stats --guild 123456789012345678

  # JSON output
  discord-cli message-stats --channel 123456789012345678 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if guildID == "" && channelID == "" {
				return usageErr(fmt.Errorf("--guild or --channel is required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			c.NoCache = true

			var allMessages []map[string]any

			if channelID != "" {
				// Single channel mode
				fmt.Fprintf(os.Stderr, "Fetching messages from channel %s...\n", channelID)
				params := map[string]string{"limit": fmt.Sprintf("%d", limit)}
				data, err := c.Get(fmt.Sprintf("/channels/%s/messages", channelID), params)
				if err != nil {
					return classifyAPIError(err)
				}
				json.Unmarshal(data, &allMessages)
			} else {
				// Guild mode - sample all channels
				fmt.Fprintf(os.Stderr, "Fetching channels for guild %s...\n", guildID)
				chData, err := c.Get(fmt.Sprintf("/guilds/%s/channels", guildID), nil)
				if err != nil {
					return classifyAPIError(err)
				}
				var channels []map[string]any
				json.Unmarshal(chData, &channels)

				for _, ch := range channels {
					chType, _ := ch["type"].(float64)
					if chType != 0 && chType != 5 {
						continue
					}
					chID := fmt.Sprintf("%v", ch["id"])
					params := map[string]string{"limit": "20"}
					data, err := c.Get(fmt.Sprintf("/channels/%s/messages", chID), params)
					if err != nil {
						continue
					}
					var msgs []map[string]any
					json.Unmarshal(data, &msgs)
					allMessages = append(allMessages, msgs...)
				}
			}

			// Analyze
			authorCounts := map[string]int{}
			hourCounts := map[int]int{}
			totalLen := 0

			for _, msg := range allMessages {
				if author, ok := msg["author"].(map[string]any); ok {
					name, _ := author["username"].(string)
					if name == "" {
						name = fmt.Sprintf("%v", author["id"])
					}
					authorCounts[name]++
				}
				if content, ok := msg["content"].(string); ok {
					totalLen += len(content)
				}
				if ts, ok := msg["timestamp"].(string); ok {
					if t, err := time.Parse(time.RFC3339, ts); err == nil {
						hourCounts[t.Hour()]++
					}
				}
			}

			avgLen := 0
			if len(allMessages) > 0 {
				avgLen = totalLen / len(allMessages)
			}

			if flags.asJSON {
				result := map[string]any{
					"total_messages": len(allMessages),
					"avg_length":     avgLen,
					"by_author":      authorCounts,
					"by_hour":        hourCounts,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Printf("\nMessage Stats (%d messages)\n\n", len(allMessages))
			fmt.Printf("Average message length: %d chars\n\n", avgLen)

			fmt.Println("Top Contributors:")
			type kv struct {
				Name  string
				Count int
			}
			var sorted []kv
			for k, v := range authorCounts {
				sorted = append(sorted, kv{k, v})
			}
			sort.Slice(sorted, func(i, j int) bool { return sorted[i].Count > sorted[j].Count })
			topN := 15
			if len(sorted) < topN {
				topN = len(sorted)
			}
			for _, item := range sorted[:topN] {
				fmt.Printf("  %-25s %d messages\n", item.Name, item.Count)
			}

			fmt.Println("\nActivity by Hour (UTC):")
			for h := 0; h < 24; h++ {
				bar := ""
				count := hourCounts[h]
				maxCount := 1
				for _, v := range hourCounts {
					if v > maxCount {
						maxCount = v
					}
				}
				barLen := 0
				if maxCount > 0 {
					barLen = count * 30 / maxCount
				}
				for i := 0; i < barLen; i++ {
					bar += "#"
				}
				fmt.Printf("  %02d:00  %3d  %s\n", h, count, bar)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&guildID, "guild", "", "Guild (server) ID")
	cmd.Flags().StringVar(&channelID, "channel", "", "Channel ID (for single-channel mode)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum messages to fetch per channel")

	return cmd
}
