package cli

import (
	"fmt"
	"strings"

	"github.com/mvanhorn/cli-printing-press/internal/websniff"
	"github.com/spf13/cobra"
)

func newSniffCmd() *cobra.Command {
	var harPath string
	var outputPath string
	var name string
	var blocklist string

	cmd := &cobra.Command{
		Use:   "sniff",
		Short: "Analyze captured web traffic to discover API endpoints and generate a spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			websniff.SetAdditionalBlocklist(splitCSV(blocklist))

			apiSpec, err := websniff.Analyze(harPath)
			if err != nil {
				return fmt.Errorf("analyzing capture: %w", err)
			}

			if name != "" {
				apiSpec.Name = name
				apiSpec.Config.Path = fmt.Sprintf("~/.config/%s-pp-cli/config.toml", name)
			}

			if outputPath == "" {
				outputPath = websniff.DefaultCachePath(apiSpec.Name)
			}

			if err := websniff.WriteSpec(apiSpec, outputPath); err != nil {
				return fmt.Errorf("writing spec: %w", err)
			}

			endpoints := 0
			for _, resource := range apiSpec.Resources {
				endpoints += len(resource.Endpoints)
			}

			fmt.Printf("Spec written to %s (%d endpoints across %d resources)\n", outputPath, endpoints, len(apiSpec.Resources))
			fmt.Printf("Run 'printing-press generate --spec %s' to build the CLI\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&harPath, "har", "", "Path to HAR or enriched capture file")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output path for generated spec YAML")
	cmd.Flags().StringVar(&name, "name", "", "Override the auto-detected API name")
	cmd.Flags().StringVar(&blocklist, "blocklist", "", "Comma-separated additional domains to filter")
	_ = cmd.MarkFlagRequired("har")

	return cmd
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}

	return out
}
