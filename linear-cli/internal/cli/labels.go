package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newLabelsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "labels",
		Short: "Manage issue labels",
	}

	cmd.AddCommand(newLabelsListCmd(flags))
	cmd.AddCommand(newLabelsCreateCmd(flags))
	return cmd
}

const labelFields = `id name color description team { name key } parent { name } children { nodes { name } }`

func newLabelsListCmd(flags *rootFlags) *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all labels",
		Example: `  linear-cli labels list
  linear-cli labels list --team ENG
  linear-cli labels list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			filterClause := ""
			if team != "" {
				filterClause = fmt.Sprintf(`, filter: { team: { key: { eqIgnoreCase: %q } } }`, team)
			}

			query := fmt.Sprintf(`{ issueLabels(first: 100%s) { nodes { %s } } }`, filterClause, labelFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.issueLabels.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&team, "team", "", "Filter by team key (e.g. 'ENG')")

	return cmd
}

func newLabelsCreateCmd(flags *rootFlags) *cobra.Command {
	var (
		name  string
		color string
		team  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new label",
		Example: `  linear-cli labels create --name bug --color "#eb5757" --team ENG
  linear-cli labels create --name feature --color "#4ea7fc" --team DES`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return usageErr(fmt.Errorf("--name is required"))
			}
			if team == "" {
				return usageErr(fmt.Errorf("--team is required (team key like 'ENG')"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			// Resolve team key to ID
			teamQuery := fmt.Sprintf(`{ teams(filter: { key: { eqIgnoreCase: %q } }) { nodes { id } } }`, team)
			teamData, err := c.Post("/graphql", gql(teamQuery, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(teamData); err != nil {
				return err
			}

			teamNodes, _ := extractData(teamData, "data.teams.nodes")
			var teams []struct{ ID string `json:"id"` }
			if err := jsonUnmarshal(teamNodes, &teams); err != nil || len(teams) == 0 {
				return usageErr(fmt.Errorf("team %q not found", team))
			}

			var inputParts []string
			inputParts = append(inputParts, fmt.Sprintf(`name: %q`, name))
			inputParts = append(inputParts, fmt.Sprintf(`teamId: %q`, teams[0].ID))
			if color != "" {
				inputParts = append(inputParts, fmt.Sprintf(`color: %q`, color))
			}

			query := fmt.Sprintf(`mutation { issueLabelCreate(input: { %s }) { success issueLabel { id name color } } }`, strings.Join(inputParts, ", "))

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			label, _ := extractData(data, "data.issueLabelCreate.issueLabel")
			return printOutput(cmd.OutOrStdout(), label, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Label name (required)")
	cmd.Flags().StringVar(&color, "color", "", "Label color hex (e.g. '#eb5757')")
	cmd.Flags().StringVar(&team, "team", "", "Team key (required, e.g. 'ENG')")

	return cmd
}
