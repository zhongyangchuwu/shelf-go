package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/exportfmt"
	"github.com/zhongyangchuwu/shelf-go/internal/project"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Project utilities"}
	cmd.AddCommand(newProjectIDCmd())
	cmd.AddCommand(newProjectInitCmd())
	cmd.AddCommand(newProjectExplainCmd())
	cmd.AddCommand(newProjectAddCmd())
	cmd.AddCommand(newProjectRmCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectExportCmd())
	cmd.AddCommand(newRunCmd())
	return cmd
}

func newProjectIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "id",
		Short: "Print current Git project identity",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := project.ID()
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
			root, err := project.Root()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, project.FileName)
			existed := false
			if _, err := os.Stat(manifestPath); err == nil {
				existed = true
				if !force {
					return fmt.Errorf("%s already exists (use --force to overwrite)", project.FileName)
				}
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err := project.Save(manifestPath, project.New()); err != nil {
				return err
			}
			label := map[bool]string{true: "overwritten", false: "created"}
			fmt.Fprintf(cmd.OutOrStdout(), "manifest: %s (%s)\n", manifestPath, label[existed])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing .shelf.json")
	return cmd
}

func newProjectExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain",
		Short: "Explain project manifest resolution",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := project.Root()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, project.FileName)
			m, err := project.Load(manifestPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found in %s; run `shelf project init`", project.FileName, root)
			}
			if err != nil {
				return err
			}

			projectID := project.IDBestEffort(root)
			fmt.Fprintf(cmd.OutOrStdout(), "project: %s\n", projectID)
			fmt.Fprintf(cmd.OutOrStdout(), "root:    %s\n", root)
			fmt.Fprintf(cmd.OutOrStdout(), "config:  %s\n\n", project.FileName)

			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			resolvedEntries, diagnostics := project.ResolveEntries(m, st)
			project.RenderDiagnostics(cmd.OutOrStdout(), diagnostics)
			for _, entry := range resolvedEntries {
				fmt.Fprintf(cmd.OutOrStdout(), "ok   %s -> %s\n", entry.Path, entry.EnvName)
			}
			for _, warning := range project.EnvOverrideWarnings(resolvedEntries, os.Environ()) {
				fmt.Fprintln(cmd.OutOrStdout(), warning)
			}
			if project.HasFailures(diagnostics) {
				return fmt.Errorf("project manifest check failed")
			}
			return nil
		},
	}
}

func newProjectAddCmd() *cobra.Command {
	var envName string
	var optional bool
	var tags []string
	cmd := &cobra.Command{
		Use:               "add [path-or-prefix]",
		Short:             "Add a secret path, prefix, or tag selector to project manifest",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeProjectAddArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := project.Root()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, project.FileName)
			if _, err := os.Stat(manifestPath); errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found; run `shelf project init` first", project.FileName)
			} else if err != nil {
				return err
			}

			selector := ""
			if len(args) > 0 {
				selector = args[0]
			}

			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}

			m, err := project.Load(manifestPath)
			if err != nil {
				return err
			}
			m, entry, err := project.AddEntry(m, st, project.AddEntryRequest{Selector: selector, Env: envName, Optional: optional, Tags: tags})
			if err != nil {
				return err
			}

			if err := project.Save(manifestPath, m); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added %s\n", entry.Key())
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
			root, err := project.Root()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, project.FileName)
			m, err := project.Load(manifestPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found", project.FileName)
			}
			if err != nil {
				return err
			}
			if !m.RemoveEntry(args[0]) {
				return fmt.Errorf("entry not found: %s", args[0])
			}
			if err := project.Save(manifestPath, m); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func completeProjectAddArg(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	_, st, err := loadRuntime(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretSetPath(st.List(""), toComplete)
}

func completeProjectEntries(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	root, err := project.Root()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	m, err := project.Load(filepath.Join(root, project.FileName))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	comps := make([]cobra.Completion, 0, len(m.Secrets))
	for _, entry := range m.Secrets {
		key := entry.Key()
		if strings.HasPrefix(key, toComplete) {
			comps = append(comps, cobra.Completion(key))
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List project manifest entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := project.Root()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, project.FileName)
			m, err := project.Load(manifestPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found", project.FileName)
			}
			if err != nil {
				return err
			}
			for _, entry := range m.Secrets {
				req := "required"
				if !entry.IsRequired() {
					req = "optional"
				}
				if entry.IsPrefix() {
					fmt.Fprintf(cmd.OutOrStdout(), "prefix %s (%s)\n", entry.Prefix, req)
				} else if entry.IsTag() {
					fmt.Fprintf(cmd.OutOrStdout(), "tag    %s (%s)\n", entry.Key(), req)
				} else if entry.Env != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "path   %s -> %s (%s)\n", entry.Path, entry.Env, req)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "path   %s (%s)\n", entry.Path, req)
				}
			}
			return nil
		},
	}
}

func newProjectExportCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export environment variables from project manifest",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := project.Root()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, project.FileName)
			m, err := project.Load(manifestPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found", project.FileName)
			}
			if err != nil {
				return err
			}

			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}

			resolvedEntries, diagnostics := project.ResolveEntries(m, st)
			project.RenderDiagnostics(cmd.OutOrStderr(), diagnostics)
			if project.HasFailures(diagnostics) {
				return fmt.Errorf("project export failed")
			}
			if len(resolvedEntries) == 0 {
				return fmt.Errorf("no secrets to export")
			}

			switch format {
			case "env":
				return renderProjectExportEnv(cmd, resolvedEntries)
			case "shell":
				return renderProjectExportShell(cmd, resolvedEntries)
			case "json":
				return renderProjectExportJSON(cmd, resolvedEntries)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "shell", "Output format")
	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{"env", "shell", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}

func renderProjectExportEnv(cmd *cobra.Command, entries []project.Binding) error {
	out, err := exportfmt.EnvBindings(project.BindingsForRender(entries))
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), out)
	return nil
}

func renderProjectExportShell(cmd *cobra.Command, entries []project.Binding) error {
	out, err := exportfmt.ShellBindings(project.BindingsForRender(entries))
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), out)
	return nil
}

func renderProjectExportJSON(cmd *cobra.Command, entries []project.Binding) error {
	out, err := exportfmt.JSONBindings(project.BindingsForRender(entries))
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), out)
	return nil
}
