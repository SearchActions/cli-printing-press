package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newWorkflowsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflows",
		Short: "Manage workflow states",
	}

	cmd.AddCommand(newWorkflowsListCmd(flags))
	return cmd
}

const workflowFields = `id name type position color team { name key }`

func newWorkflowsListCmd(flags *rootFlags) *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflow states across teams",
		Example: `  linear-cli workflows list
  linear-cli workflows list --team ENG
  linear-cli workflows list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			filterClause := ""
			if team != "" {
				filterClause = fmt.Sprintf(`, filter: { team: { key: { eqIgnoreCase: %q } } }`, team)
			}

			query := fmt.Sprintf(`{ workflowStates(first: 100%s) { nodes { %s } } }`, filterClause, workflowFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.workflowStates.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&team, "team", "", "Filter by team key (e.g. 'ENG')")

	return cmd
}
