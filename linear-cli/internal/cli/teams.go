package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newTeamsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teams",
		Short: "Manage teams and their settings",
	}

	cmd.AddCommand(newTeamsListCmd(flags))
	cmd.AddCommand(newTeamsGetCmd(flags))
	return cmd
}

const teamFields = `id name key description members { nodes { name email } } states { nodes { name type position color } } labels { nodes { name color } } activeCycle { number name startsAt endsAt }`
const teamDetailFields = `id name key description timezone members { nodes { id name email displayName active } } states { nodes { id name type position color } } labels { nodes { id name color } } activeCycle { id number name startsAt endsAt progress }`

func newTeamsListCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all teams",
		Example: `  linear-cli teams list
  linear-cli teams list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`{ teams(first: 50) { nodes { %s } } }`, teamFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.teams.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}
	return cmd
}

func newTeamsGetCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get team details by key (e.g. ENG)",
		Example: `  linear-cli teams get ENG
  linear-cli teams get DES --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			key := args[0]

			// Resolve team key to get the team
			query := fmt.Sprintf(`{ teams(filter: { key: { eqIgnoreCase: %q } }) { nodes { %s } } }`, key, teamDetailFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.teams.nodes")
			var teams []struct{ ID string `json:"id"` }
			if err := jsonUnmarshal(nodes, &teams); err != nil || len(teams) == 0 {
				return notFoundErr(fmt.Errorf("team with key %q not found", key))
			}

			// Return first match
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}
	return cmd
}
