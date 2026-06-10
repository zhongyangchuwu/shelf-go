package cli

import (
	"fmt"

	"github.com/han/shelf-go/internal/render"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var format string
	var all bool
	cmd := &cobra.Command{
		Use:   "export <path-or-prefix>",
		Short: "Export secret values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			var paths []string
			if _, ok := st.Get(args[0]); ok {
				paths = []string{args[0]}
			} else {
				paths = st.List(args[0])
			}
			if !all {
				filtered := make([]string, 0, len(paths))
				for _, p := range paths {
					if s, ok := st.Data.Secrets[p]; ok && s.Env != "" {
						filtered = append(filtered, p)
					}
				}
				paths = filtered
			}
			if len(paths) == 0 {
				return fmt.Errorf("no secrets matched: %s", args[0])
			}
			var out string
			switch format {
			case "json":
				out, err = render.JSON(paths, st.Data.Secrets)
			case "env":
				out, err = render.Env(paths, st.Data.Secrets)
			case "shell":
				out, err = render.Shell(paths, st.Data.Secrets)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "shell", "Output format")
	cmd.Flags().BoolVar(&all, "all", false, "Export all secrets, including those without env")
	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{"shell", "env", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.ValidArgsFunction = completeSecretPaths
	return cmd
}
