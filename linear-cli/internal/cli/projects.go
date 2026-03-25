package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newProjectsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects and milestones",
	}

	cmd.AddCommand(newProjectsListCmd(flags))
	cmd.AddCommand(newProjectsGetCmd(flags))
	cmd.AddCommand(newProjectsCreateCmd(flags))
	cmd.AddCommand(newProjectsUpdateCmd(flags))
	return cmd
}

const projectFields = `id name description state startDate targetDate progress lead { name } members { nodes { name } } teams { nodes { name key } }`
const projectDetailFields = `id name description url state startDate targetDate progress lead { name email } members { nodes { name } } teams { nodes { name key } } issues(first: 50) { nodes { identifier title state { name } assignee { name } priority } } documents { nodes { id title } }`

func newProjectsListCmd(flags *rootFlags) *cobra.Command {
	var state string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Example: `  linear-cli projects list
  linear-cli projects list --state planned
  linear-cli projects list --state started --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			filterClause := ""
			if state != "" {
				filterClause = fmt.Sprintf(`, filter: { state: { eq: %q } }`, state)
			}

			query := fmt.Sprintf(`{ projects(first: 50, orderBy: updatedAt%s) { nodes { %s } pageInfo { hasNextPage endCursor } } }`, filterClause, projectFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.projects.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "Filter by project state (planned, started, paused, completed, canceled)")

	return cmd
}

func newProjectsGetCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get project details by ID",
		Example: `  linear-cli projects get abc123-def456
  linear-cli projects get abc123-def456 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`{ project(id: %q) { %s } }`, args[0], projectDetailFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			project, _ := extractData(data, "data.project")
			return printOutput(cmd.OutOrStdout(), project, flags.asJSON)
		},
	}
	return cmd
}

func newProjectsCreateCmd(flags *rootFlags) *cobra.Command {
	var (
		name        string
		team        string
		description string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Example: `  linear-cli projects create --name "Q1 Launch" --team ENG
  linear-cli projects create --name "Redesign" --team DES --description "Full UI overhaul"`,
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
			inputParts = append(inputParts, fmt.Sprintf(`teamIds: [%q]`, teams[0].ID))
			if description != "" {
				inputParts = append(inputParts, fmt.Sprintf(`description: %q`, description))
			}

			query := fmt.Sprintf(`mutation { projectCreate(input: { %s }) { success project { id name url state } } }`, strings.Join(inputParts, ", "))

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			project, _ := extractData(data, "data.projectCreate.project")
			return printOutput(cmd.OutOrStdout(), project, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Project name (required)")
	cmd.Flags().StringVar(&team, "team", "", "Team key (required, e.g. 'ENG')")
	cmd.Flags().StringVar(&description, "description", "", "Project description")

	return cmd
}

func newProjectsUpdateCmd(flags *rootFlags) *cobra.Command {
	var (
		name        string
		state       string
		description string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update project details",
		Example: `  linear-cli projects update abc123 --name "New Name"
  linear-cli projects update abc123 --state completed
  linear-cli projects update abc123 --description "Updated scope"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			var inputParts []string
			if name != "" {
				inputParts = append(inputParts, fmt.Sprintf(`name: %q`, name))
			}
			if state != "" {
				inputParts = append(inputParts, fmt.Sprintf(`state: %q`, state))
			}
			if description != "" {
				inputParts = append(inputParts, fmt.Sprintf(`description: %q`, description))
			}

			if len(inputParts) == 0 {
				return usageErr(fmt.Errorf("no update flags provided"))
			}

			query := fmt.Sprintf(`mutation { projectUpdate(id: %q, input: { %s }) { success project { id name state progress } } }`, args[0], strings.Join(inputParts, ", "))

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			project, _ := extractData(data, "data.projectUpdate.project")
			return printOutput(cmd.OutOrStdout(), project, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New project name")
	cmd.Flags().StringVar(&state, "state", "", "Project state (planned, started, paused, completed, canceled)")
	cmd.Flags().StringVar(&description, "description", "", "New project description")

	return cmd
}
