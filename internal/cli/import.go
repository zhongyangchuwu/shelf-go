package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func newVaultImportCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "import", Short: "Import external secrets into the local vault"}
	cmd.AddCommand(newGopassImportCmd())
	return cmd
}

func newGopassImportCmd() *cobra.Command {
	var opts app.GopassImportOptions
	cmd := &cobra.Command{
		Use:   "gopass",
		Short: "Import gopass entries into the local vault",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			result, err := app.ImportGopassForRuntime(configPath, vaultPath, opts)
			writeImportResult(cmd, result)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.Prefix, "prefix", "", "Only import gopass entries under prefix")
	cmd.Flags().StringVar(&opts.Command, "command", "", "gopass binary path or name")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite existing local vault secrets")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview import without writing the vault")
	return cmd
}

func writeImportResult(cmd *cobra.Command, result app.ImportResult) {
	verb := "imported"
	if result.DryRun {
		verb = "would import"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", verb, len(result.Imported))
	writeImportPaths(cmd, verb, result.Imported)
	if len(result.SkippedExisting) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "skipped existing %d\n", len(result.SkippedExisting))
		writeImportPaths(cmd, "skip existing", result.SkippedExisting)
	}
	if len(result.SkippedInvalid) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "skipped invalid %d\n", len(result.SkippedInvalid))
		for _, skipped := range result.SkippedInvalid {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s (%s)\n", skipped.Path, skipped.Reason)
		}
	}
}

func writeImportPaths(cmd *cobra.Command, label string, paths []string) {
	label = strings.TrimSpace(label)
	for _, path := range paths {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", label, path)
	}
}
