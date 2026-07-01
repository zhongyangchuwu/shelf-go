package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func newProjectCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Project utilities", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() }}
	cmd.AddCommand(newProjectIDCmd())
	cmd.AddCommand(newProjectInitCmd())
	cmd.AddCommand(newProjectStatusCmd(appSvc))
	cmd.AddCommand(newProjectAddCmd(appSvc))
	cmd.AddCommand(newProjectRmCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectExportCmd(appSvc))
	cmd.AddCommand(newRunCmd(appSvc))
	return cmd
}

func newProjectIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "id",
		Short: "Print current Git project identity",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := app.ProjectID()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
}

func newProjectInitCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize project manifest",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := app.ProjectInit(force)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing .shelf.json")
	return cmd
}

func newProjectStatusCmd(appSvc *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show project manifest and env-file status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			vaultPath, _ := cmd.Flags().GetString("vault")
			out, err := appSvc.ProjectStatus(configPath, vaultPath, os.Environ())
			if out != "" {
				fmt.Fprint(cmd.OutOrStdout(), out)
			}
			return err
		},
	}
}

func newProjectAddCmd(appSvc *app.App) *cobra.Command {
	var envName string
	var optional bool
	var tags []string
	cmd := &cobra.Command{
		Use:   "add [path-or-prefix]",
		Short: "Add a secret path, prefix, or tag selector to project manifest",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeProjectAddArg(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			selector := ""
			if len(args) > 0 {
				selector = args[0]
			}

			configPath, _ := cmd.Flags().GetString("config")
			vaultPath, _ := cmd.Flags().GetString("vault")
			out, err := appSvc.ProjectAdd(configPath, vaultPath, app.ProjectAddRequest{Selector: selector, Env: envName, Optional: optional, Tags: tags})
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&envName, "env", "", "Environment variable name override (path entries only)")
	cmd.Flags().BoolVar(&optional, "optional", false, "Mark entry as optional")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag selector for this project; repeat for AND matching")
	_ = cmd.RegisterFlagCompletionFunc("env", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("tag", cobra.NoFileCompletions)
	return cmd
}

func newProjectRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "rm <path-or-prefix>",
		Short:             "Remove an entry from project manifest",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeProjectEntries,
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := app.ProjectRm(args[0])
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	return cmd
}

func completeProjectAddArg(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	configPath, vaultPath := runtimePaths(cmd)
	paths, err := appSvc.AllSecretPaths(configPath, vaultPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretSetPath(paths, toComplete)
}

func completeProjectEntries(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	comps, err := app.ProjectEntryCompletions(toComplete)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	completions := make([]cobra.Completion, 0, len(comps))
	for _, comp := range comps {
		completions = append(completions, cobra.Completion(comp))
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List project manifest entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := app.ProjectList()
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
}

func newProjectExportCmd(appSvc *app.App) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export environment variables from project manifest",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			vaultPath, _ := cmd.Flags().GetString("vault")
			result, err := appSvc.ProjectExport(configPath, vaultPath, format)
			if result.Diagnostics != "" {
				fmt.Fprint(cmd.OutOrStderr(), result.Diagnostics)
			}
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), result.Output)
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "shell", "Output format")
	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{"env", "shell", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}
