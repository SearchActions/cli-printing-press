package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newWebhooksCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhooks",
		Short: "Manage webhooks",
	}

	cmd.AddCommand(newWebhooksListCmd(flags))
	cmd.AddCommand(newWebhooksCreateCmd(flags))
	cmd.AddCommand(newWebhooksDeleteCmd(flags))
	return cmd
}

const webhookFields = `id url label enabled allPublicTeams team { name } resourceTypes createdAt`

func newWebhooksListCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		Example: `  linear-cli webhooks list
  linear-cli webhooks list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`{ webhooks(first: 50) { nodes { %s } } }`, webhookFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.webhooks.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}
	return cmd
}

func newWebhooksCreateCmd(flags *rootFlags) *cobra.Command {
	var (
		url       string
		label     string
		resources []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		Example: `  linear-cli webhooks create --url https://example.com/hook --label "My Hook"
  linear-cli webhooks create --url https://example.com/hook --resources Issue --resources Comment`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if url == "" {
				return usageErr(fmt.Errorf("--url is required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			var inputParts []string
			inputParts = append(inputParts, fmt.Sprintf(`url: %q`, url))
			if label != "" {
				inputParts = append(inputParts, fmt.Sprintf(`label: %q`, label))
			}
			if len(resources) > 0 {
				quoted := make([]string, len(resources))
				for i, r := range resources {
					quoted[i] = fmt.Sprintf("%q", r)
				}
				inputParts = append(inputParts, fmt.Sprintf(`resourceTypes: [%s]`, strings.Join(quoted, ", ")))
			}

			query := fmt.Sprintf(`mutation { webhookCreate(input: { %s }) { success webhook { id url label enabled } } }`, strings.Join(inputParts, ", "))

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			webhook, _ := extractData(data, "data.webhookCreate.webhook")
			return printOutput(cmd.OutOrStdout(), webhook, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "Webhook URL (required)")
	cmd.Flags().StringVar(&label, "label", "", "Webhook label")
	cmd.Flags().StringSliceVar(&resources, "resources", nil, "Resource types to subscribe to (e.g. Issue, Comment, Project)")

	return cmd
}

func newWebhooksDeleteCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a webhook",
		Example: `  linear-cli webhooks delete abc123-def456`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`mutation { webhookDelete(id: %q) { success } }`, args[0])

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}
			_ = data

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted webhook %s\n", args[0])
			return nil
		},
	}
	return cmd
}
