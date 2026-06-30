package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"golang.org/x/term"
)

var secretAddIsTerminal = term.IsTerminal
var secretAddReadPassword = term.ReadPassword

func newSecretCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{Use: "secret", Short: "Manage secrets"}
	cmd.AddCommand(newSecretAddCmd(appSvc))
	cmd.AddCommand(newSecretSetCmd(appSvc))
	cmd.AddCommand(newSecretGetCmd(appSvc))
	cmd.AddCommand(newSecretListCmd(appSvc))
	cmd.AddCommand(newSecretInfoCmd(appSvc))
	cmd.AddCommand(newSecretEditCmd(appSvc))
	cmd.AddCommand(newExportCmd(appSvc))
	cmd.AddCommand(newSecretRmCmd(appSvc))
	return cmd
}

func newSecretAddCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [path-or-group]",
		Short: "Interactively create a secret",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretSetPathArg(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !secretAddIsTerminal(int(os.Stdin.Fd())) {
				return fmt.Errorf("secret add requires a terminal; use `shelf secret set` for scripts")
			}
			configPath, vaultPath := runtimePaths(cmd)
			path, err := appSvc.AddSecret(configPath, vaultPath, app.AddSecretRequest{
				Args: args,
				In:   cmd.InOrStdin(),
				Out:  cmd.OutOrStdout(),
				ReadPassword: func() ([]byte, error) {
					return secretAddReadPassword(int(os.Stdin.Fd()))
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added %s\n", path)
			return nil
		},
	}
	return cmd
}

func newSecretSetCmd(appSvc *app.App) *cobra.Command {
	var envName, description string
	var tags []string
	var force bool
	cmd := &cobra.Command{
		Use:   "set <path> <value>",
		Short: "Create a secret",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretSetPathArg(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			return appSvc.SetSecret(configPath, vaultPath, app.SetSecretRequest{Path: args[0], Value: args[1], Env: envName, Description: description, Tags: tags, Force: force})
		},
	}
	cmd.Flags().StringVar(&envName, "env", "", "Environment variable name")
	cmd.Flags().StringVar(&description, "description", "", "Human-readable description")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag for this secret")
	cmd.Flags().BoolVar(&force, "force", false, "Replace existing secret")
	_ = cmd.RegisterFlagCompletionFunc("env", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("description", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("tag", cobra.NoFileCompletions)
	return cmd
}

func newSecretGetCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <path>",
		Short: "Print a secret value",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretPaths(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			value, err := appSvc.GetSecretValue(configPath, vaultPath, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}
	return cmd
}

func newSecretListCmd(appSvc *app.App) *cobra.Command {
	var tags []string
	cmd := &cobra.Command{
		Use:   "list [prefix]",
		Short: "List secret paths",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretPathPrefixes(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
			}
			configPath, vaultPath := runtimePaths(cmd)
			paths, err := appSvc.ListSecretPaths(configPath, vaultPath, app.ListSecretsRequest{Prefix: prefix, Tags: tags})
			if err != nil {
				return err
			}
			for _, path := range paths {
				fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Select secrets with tag; repeat for AND matching")
	_ = cmd.RegisterFlagCompletionFunc("tag", cobra.NoFileCompletions)
	return cmd
}

func newSecretInfoCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <path>",
		Short: "Print secret metadata as JSON",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretPaths(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			info, err := appSvc.SecretInfoJSON(configPath, vaultPath, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), info)
			return nil
		},
	}
	return cmd
}

func newSecretEditCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <path>",
		Short: "Edit a secret object as JSON",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretPaths(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			return appSvc.EditSecret(configPath, vaultPath, app.EditSecretRequest{Path: args[0], Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
		},
	}
	return cmd
}

func completeSecretPaths(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	configPath, vaultPath := runtimePaths(cmd)
	paths, err := appSvc.SecretPaths(configPath, vaultPath, toComplete)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	comps := make([]cobra.Completion, 0, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(path, toComplete) {
			comps = append(comps, cobra.Completion(path))
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func completeSecretPathPrefixes(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
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

func completeSecretSetPathArg(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	configPath, vaultPath := runtimePaths(cmd)
	paths, err := appSvc.AllSecretPaths(configPath, vaultPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretSetPathForArgs(paths, args, toComplete)
}

func completeSecretSetPathForArgs(paths []string, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeSecretSetPath(paths, toComplete)
	}
	if len(args) == 1 && !strings.Contains(args[0], ":") {
		return completeSecretSetKeyForGroup(paths, args[0], toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeSecretSetKeyForGroup(paths []string, groupPath, keyPrefix string) ([]cobra.Completion, cobra.ShellCompDirective) {
	prefix := groupPath + ":" + keyPrefix
	comps := make([]cobra.Completion, 0, len(paths))
	for _, path := range paths {
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		_, key, ok := strings.Cut(path, ":")
		if !ok {
			continue
		}
		comps = append(comps, cobra.Completion(key))
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func completeSecretSetPath(paths []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if strings.Contains(toComplete, ":") {
		comps := make([]cobra.Completion, 0, len(paths))
		for _, path := range paths {
			if strings.HasPrefix(path, toComplete) {
				comps = append(comps, cobra.Completion(path))
			}
		}
		return comps, cobra.ShellCompDirectiveNoFileComp
	}
	seen := make(map[string]struct{}, len(paths))
	groups := make([]string, 0, len(paths))
	for _, path := range paths {
		groupPath, _, ok := strings.Cut(path, ":")
		if !ok || groupPath == "" {
			continue
		}
		if !strings.HasPrefix(groupPath, toComplete) {
			continue
		}
		completion := groupPath + ":"
		if _, ok := seen[completion]; ok {
			continue
		}
		seen[completion] = struct{}{}
		groups = append(groups, completion)
	}
	sort.Strings(groups)
	comps := make([]cobra.Completion, 0, len(groups))
	for _, group := range groups {
		comps = append(comps, cobra.Completion(group))
	}
	directive := cobra.ShellCompDirectiveNoFileComp
	if len(comps) > 0 {
		directive |= cobra.ShellCompDirectiveNoSpace
	}
	return comps, directive
}

func newSecretRmCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <path>",
		Short: "Remove a secret",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretPaths(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			return appSvc.RemoveSecret(configPath, vaultPath, args[0])
		},
	}
	return cmd
}
