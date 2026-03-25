package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newUsersCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage workspace users",
	}

	cmd.AddCommand(newUsersMeCmd(flags))
	cmd.AddCommand(newUsersListCmd(flags))
	return cmd
}

func newUsersMeCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Get the authenticated user",
		Example: `  linear-cli users me
  linear-cli users me --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := `{ viewer { id name email displayName active admin url organization { id name urlKey } teamMemberships { nodes { team { id name key } } } assignedIssues(first: 10, orderBy: updatedAt, filter: { state: { type: { nin: ["completed", "canceled"] } } }) { nodes { identifier title state { name } priority } } } }`

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			viewer, _ := extractData(data, "data.viewer")
			return printOutput(cmd.OutOrStdout(), viewer, flags.asJSON)
		},
	}
	return cmd
}

func newUsersListCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspace users",
		Example: `  linear-cli users list
  linear-cli users list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := `{ users(first: 100) { nodes { id name email displayName active admin createdAt lastSeen } } }`

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.users.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}
	return cmd
}
