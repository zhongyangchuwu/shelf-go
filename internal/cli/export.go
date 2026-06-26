package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/render"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
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
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			var paths []string
			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
				if len(tags) == 0 {
					if _, ok := st.Get(prefix); ok {
						paths = []string{prefix}
					} else {
						paths = st.List(prefix)
					}
				} else if secret, ok := st.Get(prefix); ok {
					if store.HasTags(secret, tags) {
						paths = []string{prefix}
					}
				} else {
					paths = st.ListByTags(prefix, tags)
				}
			} else if len(tags) > 0 {
				paths = st.ListByTags("", tags)
			} else {
				return fmt.Errorf("path, prefix, or --tag is required")
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
				return fmt.Errorf("no secrets matched: %s", exportSelector(prefix, tags))
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
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Select secrets with tag; repeat for AND matching")
	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{"shell", "env", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.ValidArgsFunction = completeSecretPaths
	return cmd
}

func exportSelector(prefix string, tags []string) string {
	if prefix == "" {
		return "tag " + strings.Join(tags, ",")
	}
	if len(tags) == 0 {
		return prefix
	}
	return prefix + " with tag " + strings.Join(tags, ",")
}
