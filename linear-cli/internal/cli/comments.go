package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import

func newCommentsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comments",
		Short: "Manage issue comments",
	}
	cmd.AddCommand(newCommentsListCmd(flags))
	cmd.AddCommand(newCommentsCreateCmd(flags))
	cmd.AddCommand(newCommentsDeleteCmd(flags))
	return cmd
}

func newCommentsListCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <issue-identifier>",
		Short:   "List comments on an issue",
		Example: "  linear-cli comments list ENG-123",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			issueID, err := resolveIssueID(c, args[0])
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`{ issue(id: %q) { comments(first: 50) { nodes { id body user { name } createdAt updatedAt } } } }`, issueID)
			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}
			nodes, _ := extractData(data, "data.issue.comments.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}
	return cmd
}

func newCommentsCreateCmd(flags *rootFlags) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:     "create <issue-identifier>",
		Short:   "Add a comment to an issue",
		Example: `  linear-cli comments create ENG-123 --body "Looks good, shipping it"`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return usageErr(fmt.Errorf("--body is required"))
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			issueID, err := resolveIssueID(c, args[0])
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`mutation { commentCreate(input: { issueId: %q, body: %q }) { success comment { id body user { name } createdAt } } }`, issueID, body)
			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}
			comment, _ := extractData(data, "data.commentCreate.comment")
			return printOutput(cmd.OutOrStdout(), comment, flags.asJSON)
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "Comment body (markdown)")
	return cmd
}

func newCommentsDeleteCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <comment-id>",
		Short: "Delete a comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			query := fmt.Sprintf(`mutation { commentDelete(id: %q) { success } }`, args[0])
			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}
			_ = data
			fmt.Fprintln(cmd.OutOrStdout(), "Comment deleted")
			return nil
		},
	}
	return cmd
}
