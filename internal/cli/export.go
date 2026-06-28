package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func newExportCmd() *cobra.Command {
	var format string
	var all bool
	var tags []string
	cmd := &cobra.Command{
		Use:   "export [path-or-prefix]",
		Short: "Export secret values",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			selector := ""
			if len(args) > 0 {
				selector = args[0]
			}
			configPath, vaultPath := runtimePaths(cmd)
			out, err := app.ExportSecretsForRuntime(configPath, vaultPath, app.ExportRequest{Selector: selector, Tags: tags, All: all, Format: format})
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "shell", "Output format")
	cmd.Flags().BoolVar(&all, "all", false, "Export all secrets, including those without env")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Select secrets with tag; repeat for AND matching")
	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{"shell", "env", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.ValidArgsFunction = completeSecretPaths
	return cmd
}
