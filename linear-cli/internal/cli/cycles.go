package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newCyclesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cycles",
		Short: "Manage sprint cycles",
	}

	cmd.AddCommand(newCyclesListCmd(flags))
	cmd.AddCommand(newCyclesCurrentCmd(flags))
	cmd.AddCommand(newCyclesGetCmd(flags))
	return cmd
}

const cycleFields = `id number name startsAt endsAt completedAt progress team { name key }`
const cycleDetailFields = `id number name startsAt endsAt completedAt progress completedScopeHistory scopeHistory team { name key } issues(first: 100) { nodes { identifier title state { name } assignee { name } priority estimate } }`

func newCyclesListCmd(flags *rootFlags) *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cycles",
		Example: `  linear-cli cycles list
  linear-cli cycles list --team ENG
  linear-cli cycles list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			filterClause := ""
			if team != "" {
				filterClause = fmt.Sprintf(`, filter: { team: { key: { eqIgnoreCase: %q } } }`, team)
			}

			query := fmt.Sprintf(`{ cycles(first: 20, orderBy: createdAt%s) { nodes { %s } } }`, filterClause, cycleFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.cycles.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&team, "team", "", "Filter by team key (e.g. 'ENG')")

	return cmd
}

func newCyclesCurrentCmd(flags *rootFlags) *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:   "current",
		Short: "Get the active cycle",
		Example: `  linear-cli cycles current
  linear-cli cycles current --team ENG
  linear-cli cycles current --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			filterParts := []string{`isActive: { eq: true }`}
			if team != "" {
				filterParts = append(filterParts, fmt.Sprintf(`team: { key: { eqIgnoreCase: %q } }`, team))
			}

			query := fmt.Sprintf(`{ cycles(first: 1, filter: { %s }) { nodes { %s } } }`, strings.Join(filterParts, ", "), cycleDetailFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.cycles.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&team, "team", "", "Filter by team key (e.g. 'ENG')")

	return cmd
}

func newCyclesGetCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get cycle details with issues",
		Example: `  linear-cli cycles get abc123-def456
  linear-cli cycles get abc123-def456 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`{ cycle(id: %q) { %s } }`, args[0], cycleDetailFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			cycle, _ := extractData(data, "data.cycle")
			return printOutput(cmd.OutOrStdout(), cycle, flags.asJSON)
		},
	}
	return cmd
}
