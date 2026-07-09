package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func newSecretCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage v2 secret items",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown secret command: %s", args[0])
			}
			return cmd.Help()
		},
	}
	cmd.AddCommand(newSecretCreateCmd(appSvc))
	cmd.AddCommand(newSecretSetCmd(appSvc))
	cmd.AddCommand(newSecretGetCmd(appSvc))
	cmd.AddCommand(newSecretListCmd(appSvc))
	cmd.AddCommand(newSecretInfoCmd(appSvc))
	cmd.AddCommand(newSecretExportCmd(appSvc))
	cmd.AddCommand(newSecretRmCmd(appSvc))
	return cmd
}

func newSecretCreateCmd(appSvc *app.App) *cobra.Command {
	var req app.CreateSecretRequest
	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a v2 secret item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req.Title = args[0]
			req.In = cmd.InOrStdin()
			configPath, vaultPath := runtimePaths(cmd)
			id, err := appSvc.CreateSecretV2(configPath, vaultPath, req)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Template, "template", "", "Template id or alias (generic, api)")
	cmd.Flags().StringVar(&req.Profile, "profile", "", "Profile id or alias")
	cmd.Flags().StringVar(&req.ID, "id", "", "Stable item id")
	cmd.Flags().StringVar(&req.Env, "env", "", "Environment variable name for the primary secret field")
	cmd.Flags().BoolVar(&req.Stdin, "stdin", false, "Read primary secret field value from stdin")
	cmd.Flags().StringVar(&req.Provider, "provider", "", "Provider/source public field")
	cmd.Flags().StringVar(&req.Website, "website", "", "Website public field")
	cmd.Flags().StringVar(&req.Expires, "expires", "", "Expiration date public field (YYYY-MM-DD)")
	cmd.Flags().StringArrayVar(&req.Scopes, "scope", nil, "Scope public field; repeat for multiple scopes")
	_ = cmd.RegisterFlagCompletionFunc("template", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("profile", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("id", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("env", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("provider", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("website", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("expires", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("scope", cobra.NoFileCompletions)
	return cmd
}

func newSecretSetCmd(appSvc *app.App) *cobra.Command {
	var field string
	var stdin bool
	cmd := &cobra.Command{
		Use:   "set <selector>",
		Short: "Set a v2 secret item field from stdin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			return appSvc.SetSecretFieldV2(configPath, vaultPath, app.SetSecretFieldRequest{Selector: args[0], Field: field, Stdin: stdin, In: cmd.InOrStdin()})
		},
	}
	cmd.Flags().StringVar(&field, "field", "", "Field id; defaults to the template primary secret field")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "Read field value from stdin")
	_ = cmd.RegisterFlagCompletionFunc("field", cobra.NoFileCompletions)
	return cmd
}

func newSecretGetCmd(appSvc *app.App) *cobra.Command {
	var field, envName string
	cmd := &cobra.Command{
		Use:   "get [selector]",
		Short: "Print a v2 secret item field value",
		Args: func(cmd *cobra.Command, args []string) error {
			if envName != "" {
				return cobra.MaximumNArgs(1)(cmd, args)
			}
			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			selector := ""
			if len(args) > 0 {
				selector = args[0]
			}
			configPath, vaultPath := runtimePaths(cmd)
			value, err := appSvc.GetSecretFieldV2(configPath, vaultPath, app.GetSecretFieldRequest{Selector: selector, Env: envName, Field: field})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}
	cmd.Flags().StringVar(&field, "field", "", "Field id; defaults to the template primary secret field")
	cmd.Flags().StringVar(&envName, "env", "", "Find the secret field by env-name field value")
	_ = cmd.RegisterFlagCompletionFunc("field", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("env", cobra.NoFileCompletions)
	return cmd
}

func newSecretExportCmd(appSvc *app.App) *cobra.Command {
	var labels []string
	var format string
	cmd := &cobra.Command{
		Use:   "export --label LABEL",
		Short: "Export v2 secret items by label",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			out, err := appSvc.ExportSecretItemsV2(configPath, vaultPath, app.ExportSecretItemsRequest{Labels: labels, Format: format})
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Label to export; repeat for AND matching")
	cmd.Flags().StringVar(&format, "format", "shell", "Output format: shell, env, json")
	_ = cmd.RegisterFlagCompletionFunc("label", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions([]string{"shell", "env", "json"}, cobra.ShellCompDirectiveNoFileComp))
	return cmd
}

func newSecretListCmd(appSvc *app.App) *cobra.Command {
	var long bool
	cmd := &cobra.Command{
		Use:   "list [query]",
		Short: "List v2 secret items without secret values",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			configPath, vaultPath := runtimePaths(cmd)
			items, err := appSvc.ListSecretItemsV2(configPath, vaultPath, app.ListSecretItemsRequest{Query: query, Long: long})
			if err != nil {
				return err
			}
			for _, item := range items {
				if long {
					state := "unset"
					if item.ValueSet {
						state = "set"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\t%s\n", item.ID, item.Title, item.Template, item.Env, state)
					continue
				}
				fmt.Fprintln(cmd.OutOrStdout(), item.Title)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&long, "long", false, "Show id, template, primary env name, and value-set state")
	return cmd
}

func newSecretInfoCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <selector>",
		Short: "Print v2 secret item metadata without secret values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			info, err := appSvc.SecretInfoV2(configPath, vaultPath, args[0])
			if err != nil {
				return err
			}
			content, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(content))
			return nil
		},
	}
	return cmd
}

func newSecretRmCmd(appSvc *app.App) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "rm <selector>",
		Short: "Remove a v2 secret item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("secret rm requires --yes")
			}
			configPath, vaultPath := runtimePaths(cmd)
			return appSvc.RemoveSecretV2(configPath, vaultPath, args[0])
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm removal")
	return cmd
}

func completeSecretSetPathForArgs(paths []string, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretSetPath(paths, toComplete)
}

func completeSecretSetKeyForGroup(paths []string, groupPath, keyPrefix string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeSecretPaths(appSvc *app.App, cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeSecretSetPath(paths []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	comps := make([]cobra.Completion, 0, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(path, toComplete) {
			comps = append(comps, cobra.Completion(path))
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}
