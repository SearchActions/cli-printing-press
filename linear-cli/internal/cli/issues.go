package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newIssuesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issues",
		Short: "Manage issues - create, update, search, and track work",
	}

	cmd.AddCommand(newIssuesListCmd(flags))
	cmd.AddCommand(newIssuesMineCmd(flags))
	cmd.AddCommand(newIssuesGetCmd(flags))
	cmd.AddCommand(newIssuesCreateCmd(flags))
	cmd.AddCommand(newIssuesUpdateCmd(flags))
	cmd.AddCommand(newIssuesSearchCmd(flags))
	cmd.AddCommand(newIssuesArchiveCmd(flags))
	cmd.AddCommand(newIssuesDeleteCmd(flags))
	return cmd
}

const issueFields = `id identifier title state { name } assignee { name } priority priorityLabel project { name } cycle { number } labels { nodes { name } } estimate dueDate createdAt updatedAt`
const issueDetailFields = `id identifier title description url state { name } assignee { name email } priority priorityLabel project { name } cycle { name number } labels { nodes { name } } estimate dueDate createdAt updatedAt completedAt canceledAt parent { identifier title } children { nodes { identifier title state { name } } } relations { nodes { type relatedIssue { identifier title } } } comments(first: 20) { nodes { body user { name } createdAt } } attachments { nodes { title url } }`

func newIssuesListCmd(flags *rootFlags) *cobra.Command {
	var (
		state    string
		assignee string
		team     string
		project  string
		label    string
		priority string
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues with filters",
		Example: `  linear-cli issues list
  linear-cli issues list --state "In Progress"
  linear-cli issues list --assignee me --team ENG
  linear-cli issues list --priority urgent --limit 10
  linear-cli issues list --project "Q1 Launch" --label bug`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			filter := buildFilter(state, assignee, team, project, label, priority)
			filterClause := ""
			if filter != "" {
				filterClause = ", " + filter
			}

			query := fmt.Sprintf(`{ issues(first: %d, orderBy: updatedAt%s) { nodes { %s } pageInfo { hasNextPage endCursor } } }`, limit, filterClause, issueFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.issues.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "Filter by state name (e.g. 'In Progress', 'Done', 'Todo')")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee name (use 'me' for yourself)")
	cmd.Flags().StringVar(&team, "team", "", "Filter by team key (e.g. 'ENG', 'DES')")
	cmd.Flags().StringVar(&project, "project", "", "Filter by project name")
	cmd.Flags().StringVar(&label, "label", "", "Filter by label name")
	cmd.Flags().StringVar(&priority, "priority", "", "Filter by priority (urgent, high, medium, low, none)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Max results to return")

	return cmd
}

func newIssuesMineCmd(flags *rootFlags) *cobra.Command {
	var (
		state string
		limit int
		all   bool
	)

	cmd := &cobra.Command{
		Use:   "mine",
		Short: "List issues assigned to me",
		Example: `  linear-cli issues mine
  linear-cli issues mine --state "In Progress"
  linear-cli issues mine --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			filterClause := ""
			if !all {
				// By default, exclude completed and canceled
				filterClause = `, filter: { state: { type: { nin: ["completed", "canceled"] } } }`
			}
			if state != "" {
				filterClause = fmt.Sprintf(`, filter: { state: { name: { eqIgnoreCase: %q } } }`, state)
			}

			query := fmt.Sprintf(`{ viewer { assignedIssues(first: %d, orderBy: updatedAt%s) { nodes { %s } } } }`, limit, filterClause, issueFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.viewer.assignedIssues.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "Filter by state name")
	cmd.Flags().IntVar(&limit, "limit", 50, "Max results")
	cmd.Flags().BoolVar(&all, "all", false, "Include completed and canceled issues")

	return cmd
}

func newIssuesGetCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <identifier>",
		Short: "Get a single issue by identifier (e.g. ENG-123)",
		Example: `  linear-cli issues get ENG-123
  linear-cli issues get ENG-123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			identifier := args[0]

			// Linear's issue() query accepts the identifier directly
			query := fmt.Sprintf(`{ issueVcsBranchSearch(branchName: %q) { %s } }`, identifier, issueDetailFields)

			// Try by identifier first using search
			searchQuery := fmt.Sprintf(`{ searchIssues(term: %q, first: 1) { nodes { %s } } }`, identifier, issueDetailFields)

			// If it looks like a UUID, use issue(id:) directly
			if len(identifier) == 36 && strings.Count(identifier, "-") == 4 {
				query = fmt.Sprintf(`{ issue(id: %q) { %s } }`, identifier, issueDetailFields)
				data, err := c.Post("/graphql", gql(query, nil))
				if err != nil {
					return classifyAPIError(err)
				}
				if err := extractErrors(data); err != nil {
					return err
				}
				issue, _ := extractData(data, "data.issue")
				return printOutput(cmd.OutOrStdout(), issue, flags.asJSON)
			}

			// For identifiers like ENG-123, search for it
			data, err := c.Post("/graphql", gql(searchQuery, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			_ = query // suppress unused
			nodes, _ := extractData(data, "data.searchIssues.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}
	return cmd
}

func newIssuesCreateCmd(flags *rootFlags) *cobra.Command {
	var (
		title       string
		team        string
		description string
		assignee    string
		state       string
		priority    string
		project     string
		label       string
		estimate    int
		dueDate     string
		parentID    string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		Example: `  linear-cli issues create --title "Fix login bug" --team ENG
  linear-cli issues create --title "Add dark mode" --team DES --priority high --label feature
  linear-cli issues create --title "Sub-task" --team ENG --parent ENG-123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return usageErr(fmt.Errorf("--title is required"))
			}
			if team == "" {
				return usageErr(fmt.Errorf("--team is required (team key like 'ENG')"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			// First resolve team key to ID
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

			// Build input
			var inputParts []string
			inputParts = append(inputParts, fmt.Sprintf(`title: %q`, title))
			inputParts = append(inputParts, fmt.Sprintf(`teamId: %q`, teams[0].ID))

			if description != "" {
				inputParts = append(inputParts, fmt.Sprintf(`description: %q`, description))
			}
			if priority != "" {
				p := priorityNumber(priority)
				if p >= 0 {
					inputParts = append(inputParts, fmt.Sprintf(`priority: %d`, p))
				}
			}
			if estimate > 0 {
				inputParts = append(inputParts, fmt.Sprintf(`estimate: %d`, estimate))
			}
			if dueDate != "" {
				inputParts = append(inputParts, fmt.Sprintf(`dueDate: %q`, dueDate))
			}
			if parentID != "" {
				inputParts = append(inputParts, fmt.Sprintf(`parentId: %q`, parentID))
			}

			query := fmt.Sprintf(`mutation { issueCreate(input: { %s }) { success issue { id identifier title url state { name } assignee { name } } } }`, strings.Join(inputParts, ", "))

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			issue, _ := extractData(data, "data.issueCreate.issue")
			return printOutput(cmd.OutOrStdout(), issue, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Issue title (required)")
	cmd.Flags().StringVar(&team, "team", "", "Team key (required, e.g. 'ENG')")
	cmd.Flags().StringVar(&description, "description", "", "Issue description (markdown)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee name")
	cmd.Flags().StringVar(&state, "state", "", "Initial state name")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority (urgent, high, medium, low, none)")
	cmd.Flags().StringVar(&project, "project", "", "Project name")
	cmd.Flags().StringVar(&label, "label", "", "Label name")
	cmd.Flags().IntVar(&estimate, "estimate", 0, "Story points estimate")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&parentID, "parent", "", "Parent issue ID for sub-issues")
	_ = assignee
	_ = state
	_ = project
	_ = label

	return cmd
}

func newIssuesUpdateCmd(flags *rootFlags) *cobra.Command {
	var (
		title       string
		state       string
		assignee    string
		priority    string
		description string
		estimate    int
		dueDate     string
	)

	cmd := &cobra.Command{
		Use:   "update <identifier>",
		Short: "Update an existing issue",
		Example: `  linear-cli issues update ENG-123 --state "In Progress"
  linear-cli issues update ENG-123 --assignee "Jane Doe" --priority high
  linear-cli issues update ENG-123 --title "New title" --estimate 3`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			identifier := args[0]
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			// Resolve identifier to ID
			issueID, err := resolveIssueID(c, identifier)
			if err != nil {
				return err
			}

			var inputParts []string
			if title != "" {
				inputParts = append(inputParts, fmt.Sprintf(`title: %q`, title))
			}
			if description != "" {
				inputParts = append(inputParts, fmt.Sprintf(`description: %q`, description))
			}
			if priority != "" {
				p := priorityNumber(priority)
				if p >= 0 {
					inputParts = append(inputParts, fmt.Sprintf(`priority: %d`, p))
				}
			}
			if estimate > 0 {
				inputParts = append(inputParts, fmt.Sprintf(`estimate: %d`, estimate))
			}
			if dueDate != "" {
				inputParts = append(inputParts, fmt.Sprintf(`dueDate: %q`, dueDate))
			}
			if state != "" {
				// Resolve state name to ID - need to find via the issue's team
				stateID, err := resolveStateID(c, issueID, state)
				if err == nil && stateID != "" {
					inputParts = append(inputParts, fmt.Sprintf(`stateId: %q`, stateID))
				}
			}
			if assignee != "" {
				assigneeID, err := resolveAssigneeID(c, assignee)
				if err == nil && assigneeID != "" {
					inputParts = append(inputParts, fmt.Sprintf(`assigneeId: %q`, assigneeID))
				}
			}

			if len(inputParts) == 0 {
				return usageErr(fmt.Errorf("no update flags provided"))
			}

			query := fmt.Sprintf(`mutation { issueUpdate(id: %q, input: { %s }) { success issue { id identifier title state { name } assignee { name } priority priorityLabel } } }`, issueID, strings.Join(inputParts, ", "))

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			issue, _ := extractData(data, "data.issueUpdate.issue")
			return printOutput(cmd.OutOrStdout(), issue, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&state, "state", "", "Move to state (e.g. 'In Progress', 'Done')")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assign to user (name or 'me')")
	cmd.Flags().StringVar(&priority, "priority", "", "Set priority (urgent, high, medium, low, none)")
	cmd.Flags().StringVar(&description, "description", "", "New description (markdown)")
	cmd.Flags().IntVar(&estimate, "estimate", 0, "Story points estimate")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "Due date (YYYY-MM-DD)")

	return cmd
}

func newIssuesSearchCmd(flags *rootFlags) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search <term>",
		Short: "Search issues by text",
		Example: `  linear-cli issues search "login bug"
  linear-cli issues search "performance" --limit 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			term := args[0]
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`{ searchIssues(term: %q, first: %d) { nodes { %s } } }`, term, limit, issueFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.searchIssues.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 25, "Max results")
	return cmd
}

func newIssuesArchiveCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <identifier>",
		Short: "Archive an issue",
		Example: `  linear-cli issues archive ENG-123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			issueID, err := resolveIssueID(c, args[0])
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`mutation { issueArchive(id: %q) { success } }`, issueID)
			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Archived %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func newIssuesDeleteCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <identifier>",
		Short: "Delete an issue permanently",
		Example: `  linear-cli issues delete ENG-123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			issueID, err := resolveIssueID(c, args[0])
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`mutation { issueDelete(id: %q) { success } }`, issueID)
			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s\n", args[0])
			return nil
		},
	}
	return cmd
}

// resolveIssueID takes an identifier like "ENG-123" and returns the UUID.
func resolveIssueID(c interface {
	Post(path string, body any) (jsonRaw, error)
}, identifier string) (string, error) {
	// If already a UUID, return directly
	if len(identifier) == 36 && strings.Count(identifier, "-") == 4 {
		return identifier, nil
	}

	query := fmt.Sprintf(`{ searchIssues(term: %q, first: 1) { nodes { id identifier } } }`, identifier)
	data, err := c.Post("/graphql", gql(query, nil))
	if err != nil {
		return "", classifyAPIError(err)
	}
	if err := extractErrors(data); err != nil {
		return "", err
	}

	nodes, _ := extractData(data, "data.searchIssues.nodes")
	var issues []struct {
		ID         string `json:"id"`
		Identifier string `json:"identifier"`
	}
	if err := jsonUnmarshal(nodes, &issues); err != nil || len(issues) == 0 {
		return "", notFoundErr(fmt.Errorf("issue %q not found", identifier))
	}
	return issues[0].ID, nil
}

func resolveStateID(c interface {
	Post(path string, body any) (jsonRaw, error)
}, issueID, stateName string) (string, error) {
	// Get the issue's team, then find the state by name
	query := fmt.Sprintf(`{ issue(id: %q) { team { states { nodes { id name } } } } }`, issueID)
	data, err := c.Post("/graphql", gql(query, nil))
	if err != nil {
		return "", err
	}
	states, _ := extractData(data, "data.issue.team.states.nodes")
	var stateList []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := jsonUnmarshal(states, &stateList); err != nil {
		return "", err
	}
	for _, s := range stateList {
		if strings.EqualFold(s.Name, stateName) {
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("state %q not found", stateName)
}

func resolveAssigneeID(c interface {
	Post(path string, body any) (jsonRaw, error)
}, name string) (string, error) {
	if name == "me" {
		query := `{ viewer { id } }`
		data, err := c.Post("/graphql", gql(query, nil))
		if err != nil {
			return "", err
		}
		id, _ := extractData(data, "data.viewer.id")
		var s string
		if err := jsonUnmarshal(id, &s); err != nil {
			return "", err
		}
		return s, nil
	}

	query := fmt.Sprintf(`{ users(filter: { name: { containsIgnoreCase: %q } }) { nodes { id name } } }`, name)
	data, err := c.Post("/graphql", gql(query, nil))
	if err != nil {
		return "", err
	}
	nodes, _ := extractData(data, "data.users.nodes")
	var users []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := jsonUnmarshal(nodes, &users); err != nil || len(users) == 0 {
		return "", fmt.Errorf("user %q not found", name)
	}
	return users[0].ID, nil
}
