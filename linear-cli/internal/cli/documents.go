package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newDocumentsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "documents",
		Short: "Manage project documents",
	}

	cmd.AddCommand(newDocumentsListCmd(flags))
	cmd.AddCommand(newDocumentsGetCmd(flags))
	cmd.AddCommand(newDocumentsCreateCmd(flags))
	return cmd
}

const documentFields = `id title creator { name } project { name } updatedAt createdAt`
const documentDetailFields = `id title content url creator { name } project { name } updatedAt createdAt`

func newDocumentsListCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		Example: `  linear-cli documents list
  linear-cli documents list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`{ documents(first: 50, orderBy: updatedAt) { nodes { %s } } }`, documentFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			nodes, _ := extractData(data, "data.documents.nodes")
			return printOutput(cmd.OutOrStdout(), nodes, flags.asJSON)
		},
	}
	return cmd
}

func newDocumentsGetCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get document content by ID",
		Example: `  linear-cli documents get abc123-def456
  linear-cli documents get abc123-def456 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := fmt.Sprintf(`{ document(id: %q) { %s } }`, args[0], documentDetailFields)

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			doc, _ := extractData(data, "data.document")
			return printOutput(cmd.OutOrStdout(), doc, flags.asJSON)
		},
	}
	return cmd
}

func newDocumentsCreateCmd(flags *rootFlags) *cobra.Command {
	var (
		title   string
		content string
		project string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a document",
		Example: `  linear-cli documents create --title "Design Spec" --content "## Overview" --project abc123
  linear-cli documents create --title "Meeting Notes" --content "Discussed roadmap"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return usageErr(fmt.Errorf("--title is required"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			var inputParts []string
			inputParts = append(inputParts, fmt.Sprintf(`title: %q`, title))
			if content != "" {
				inputParts = append(inputParts, fmt.Sprintf(`content: %q`, content))
			}
			if project != "" {
				inputParts = append(inputParts, fmt.Sprintf(`projectId: %q`, project))
			}

			query := fmt.Sprintf(`mutation { documentCreate(input: { %s }) { success document { id title url } } }`, strings.Join(inputParts, ", "))

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			doc, _ := extractData(data, "data.documentCreate.document")
			return printOutput(cmd.OutOrStdout(), doc, flags.asJSON)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Document title (required)")
	cmd.Flags().StringVar(&content, "content", "", "Document content (markdown)")
	cmd.Flags().StringVar(&project, "project", "", "Project ID to attach document to")

	return cmd
}
