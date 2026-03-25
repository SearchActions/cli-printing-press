package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import

func newOrganizationCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organization",
		Short: "View workspace/organization info",
	}

	cmd.AddCommand(newOrganizationGetCmd(flags))
	return cmd
}

func newOrganizationGetCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get organization details",
		Example: `  linear-cli organization get
  linear-cli organization get --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			query := `{ organization { id name urlKey logoUrl createdAt subscription { type seats } teams { nodes { name key } } users { nodes { name email active } } } }`

			data, err := c.Post("/graphql", gql(query, nil))
			if err != nil {
				return classifyAPIError(err)
			}
			if err := extractErrors(data); err != nil {
				return err
			}

			org, _ := extractData(data, "data.organization")
			return printOutput(cmd.OutOrStdout(), org, flags.asJSON)
		},
	}
	return cmd
}
