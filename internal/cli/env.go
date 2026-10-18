package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func newEnvCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environment projections",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newEnvExportCmd(appSvc))
	return cmd
}

func newEnvExportCmd(appSvc *app.App) *cobra.Command {
	var labels []string
	var format string
	cmd := &cobra.Command{
		Use:   "export --label LABEL",
		Short: "Export mapped environment fields by label",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			out, err := appSvc.ExportEnvironment(configPath, vaultPath, app.ExportEnvironmentRequest{Labels: labels, Format: format})
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Label to export; repeat for AND matching")
	cmd.Flags().StringVar(&format, "format", "shell", "Output format: shell or json")
	_ = cmd.RegisterFlagCompletionFunc("label", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completeSecretLabels(appSvc, cmd, args, toComplete)
	})
	_ = cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions([]string{"shell", "json"}, cobra.ShellCompDirectiveNoFileComp))
	return cmd
}
