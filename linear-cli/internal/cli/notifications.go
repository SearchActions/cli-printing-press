package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newNotificationsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notifications",
		Short: "Manage notifications",
	}

	cmd.AddCommand(newNotificationsListCmd(flags))
	return cmd
}

func newNotificationsListCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent notifications",
		Example: `  linear-cli notifications list
  linear-cli notifications list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := `{ notifications(first: 50, orderBy: createdAt) { nodes { id type readAt createdAt ... on IssueNotification { issue { identifier title state { name } } comment { body } actor { name } } } } }`

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.notifications.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}
	return cmd
}
