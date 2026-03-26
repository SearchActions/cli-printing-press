package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newWebhookTestCmd(flags *rootFlags) *cobra.Command {
	var webhookID string
	var webhookToken string
	var message string

	cmd := &cobra.Command{
		Use:   "webhook-test",
		Short: "Send a test payload to a Discord webhook",
		Long: `Sends a test message to a webhook URL to verify it's working.
Supports plain text and rich embeds. Useful for validating webhook
integrations before deploying to production.

Uses: POST /webhooks/{id}/{token}.`,
		Example: `  # Send a simple test message
  discord-cli webhook-test --id 123456789012345678 --token abcdef123456 --message "Hello from CLI!"

  # Send a rich embed via stdin
  echo '{"content":"Deploy alert","embeds":[{"title":"Build #42","description":"All tests passed","color":3066993}]}' | discord-cli webhook-test --id 123456789012345678 --token abcdef123456 --stdin

  # Dry run to preview the request
  discord-cli webhook-test --id 123456789012345678 --token abcdef123456 --message "test" --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if webhookID == "" || webhookToken == "" {
				return usageErr(fmt.Errorf("--id and --token are required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			body := map[string]any{"content": message}
			if message == "" {
				body["content"] = fmt.Sprintf("Test message from discord-cli (webhook %s)", webhookID)
			}

			fmt.Fprintf(os.Stderr, "Sending test payload to webhook %s...\n", webhookID)

			path := fmt.Sprintf("/webhooks/%s/%s", webhookID, webhookToken)
			data, err := c.Post(path, body)
			if err != nil {
				return classifyAPIError(err)
			}

			if flags.asJSON {
				return printOutput(os.Stdout, data, true)
			}

			fmt.Fprintln(os.Stdout, green("Webhook test successful!"))
			return nil
		},
	}

	cmd.Flags().StringVar(&webhookID, "id", "", "Webhook ID (required)")
	cmd.Flags().StringVar(&webhookToken, "token", "", "Webhook token (required)")
	cmd.Flags().StringVar(&message, "message", "", "Test message content")

	return cmd
}
