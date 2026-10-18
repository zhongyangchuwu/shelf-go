package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/clipboard"
	"github.com/zhongyangchuwu/shelf-go/internal/cliui"
	"golang.org/x/term"
)

func newSecretCmd(appSvc *app.App, clipboardWriter clipboard.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage secret items",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown secret command: %s", args[0])
			}
			return cmd.Help()
		},
	}
	cmd.AddCommand(newSecretCreateCmd(appSvc))
	cmd.AddCommand(newSecretSetCmd(appSvc))
	cmd.AddCommand(newSecretUpdateCmd(appSvc))
	cmd.AddCommand(newSecretGetCmd(appSvc))
	cmd.AddCommand(newSecretCopyCmd(appSvc, clipboardWriter))
	cmd.AddCommand(newSecretListCmd(appSvc))
	cmd.AddCommand(newSecretInfoCmd(appSvc))
	cmd.AddCommand(newSecretRmCmd(appSvc))
	cmd.AddCommand(newSecretAuditCmd(appSvc))
	return cmd
}

func newSecretCreateCmd(appSvc *app.App) *cobra.Command {
	var req app.CreateSecretRequest
	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a secret item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req.Title = args[0]
			req.In = cmd.InOrStdin()
			configPath, vaultPath := runtimePaths(cmd)
			id, err := appSvc.CreateSecret(configPath, vaultPath, req)
			if err != nil {
				return err
			}
			info, err := appSvc.SecretInfo(configPath, vaultPath, id)
			if err != nil {
				return err
			}
			renderSecretCreateResult(cmd.OutOrStdout(), info)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Template, "template", "", "Template id or alias (generic, api)")
	cmd.Flags().StringVar(&req.Profile, "profile", "", "Profile id or alias")
	cmd.Flags().StringVar(&req.ID, "id", "", "Stable item id")
	cmd.Flags().StringVar(&req.Env, "env", "", "Environment variable name for the primary secret field")
	cmd.Flags().BoolVar(&req.Stdin, "stdin", false, "Read primary secret field value from stdin")
	cmd.Flags().StringVar(&req.Description, "description", "", "Human-readable item description")
	cmd.Flags().StringVar(&req.Provider, "provider", "", "Provider/source public field")
	cmd.Flags().StringVar(&req.Website, "website", "", "Website public field")
	cmd.Flags().StringVar(&req.Validity, "validity", "", "Credential validity: unknown, permanent, or YYYY-MM-DD")
	cmd.Flags().StringArrayVar(&req.Scopes, "scope", nil, "Scope public field; repeat for multiple scopes")
	_ = cmd.RegisterFlagCompletionFunc("template", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeSecretTemplates(appSvc, cmd, args, toComplete)
	})
	_ = cmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeSecretProfiles(appSvc, cmd, args, toComplete)
	})
	_ = cmd.RegisterFlagCompletionFunc("id", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("env", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("description", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("provider", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("website", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("validity", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("scope", cobra.NoFileCompletions)
	return cmd
}

func newSecretUpdateCmd(appSvc *app.App) *cobra.Command {
	var (
		req        app.UpdateSecretRequest
		field      string
		values     []string
		clearField bool
	)
	cmd := &cobra.Command{
		Use:   "update <selector>",
		Short: "Update a secret item",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretSelectors(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			req.Selector = args[0]
			req.TitleSet = cmd.Flags().Changed("title")
			req.DescriptionSet = cmd.Flags().Changed("description")
			fieldValueSet := cmd.Flags().Changed("value")
			metadataChanged := req.TitleSet || req.DescriptionSet || cmd.Flags().Changed("add-label") || cmd.Flags().Changed("remove-label")
			readField := !fieldValueSet && !clearField && (cmd.Flags().Changed("field") || !metadataChanged)
			if readField {
				input, err := updateFieldInput(cmd, field)
				if err != nil {
					return err
				}
				req.FieldUpdates = []app.FieldUpdate{{Field: field, In: input, ReadField: true}}
			} else if fieldValueSet || clearField {
				req.FieldUpdates = []app.FieldUpdate{{Field: field, Values: values, FieldValueSet: fieldValueSet, ClearField: clearField}}
			}
			configPath, vaultPath := runtimePaths(cmd)
			if err := appSvc.UpdateSecret(configPath, vaultPath, req); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Updated secret item.")
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Title, "title", "", "New item title")
	cmd.Flags().StringVar(&req.Description, "description", "", "Human-readable item description")
	cmd.Flags().StringArrayVar(&req.AddLabels, "add-label", nil, "Label to add; repeat for multiple labels")
	cmd.Flags().StringArrayVar(&req.RemoveLabels, "remove-label", nil, "Label to remove; repeat for multiple labels")
	cmd.Flags().StringVar(&field, "field", "", "Field id; defaults to the primary secret field when reading stdin")
	cmd.Flags().StringArrayVar(&values, "value", nil, "Public field value; repeat for a string-list field")
	cmd.Flags().BoolVar(&clearField, "clear-field", false, "Clear the selected public field")
	_ = cmd.RegisterFlagCompletionFunc("description", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("add-label", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeSecretLabels(appSvc, cmd, args, toComplete)
	})
	_ = cmd.RegisterFlagCompletionFunc("remove-label", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretLabels, Selector: args[0]})
	})
	_ = cmd.RegisterFlagCompletionFunc("field", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeSecretFields(appSvc, cmd, args, toComplete, secretFieldCompletionOptions{})
	})
	_ = cmd.RegisterFlagCompletionFunc("value", cobra.NoFileCompletions)
	return cmd
}

func updateFieldInput(cmd *cobra.Command, field string) (io.Reader, error) {
	input := cmd.InOrStdin()
	file, ok := input.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return input, nil
	}
	prompt := field
	if prompt == "" {
		prompt = "the primary secret field"
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Enter value for %s: ", prompt)
	value, err := term.ReadPassword(int(file.Fd()))
	fmt.Fprintln(cmd.ErrOrStderr())
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", prompt, err)
	}
	return bytes.NewReader(value), nil
}

func newSecretGetCmd(appSvc *app.App) *cobra.Command {
	var field, envName string
	cmd := &cobra.Command{
		Use:   "get [selector]",
		Short: "Print a secret item field value",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretSelectors(appSvc, cmd, args, toComplete)
		},
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
			value, err := appSvc.GetSecretField(configPath, vaultPath, app.GetSecretFieldRequest{Selector: selector, Env: envName, Field: field})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}
	cmd.Flags().StringVar(&field, "field", "", "Field id; defaults to the template primary secret field")
	cmd.Flags().StringVar(&envName, "env", "", "Find the secret field by env-name field value")
	_ = cmd.RegisterFlagCompletionFunc("field", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeSecretFields(appSvc, cmd, args, toComplete, secretFieldCompletionOptions{useEnvSelector: true})
	})
	_ = cmd.RegisterFlagCompletionFunc("env", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeSecretEnvNames(appSvc, cmd, args, toComplete)
	})
	return cmd
}

func newSecretCopyCmd(appSvc *app.App, clipboardWriter clipboard.Writer) *cobra.Command {
	var field, envName string
	cmd := &cobra.Command{
		Use:   "copy [selector]",
		Short: "Copy a secret item field to the clipboard",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretSelectors(appSvc, cmd, args, toComplete)
		},
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
			value, err := appSvc.GetSecretField(configPath, vaultPath, app.GetSecretFieldRequest{Selector: selector, Env: envName, Field: field})
			if err != nil {
				return err
			}
			if err := clipboardWriter.CopyText(value); err != nil {
				return fmt.Errorf("copy secret to clipboard: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Copied secret to clipboard.")
			return nil
		},
	}
	cmd.Flags().StringVar(&field, "field", "", "Field id; defaults to the template primary secret field")
	cmd.Flags().StringVar(&envName, "env", "", "Find the secret field by env-name field value")
	_ = cmd.RegisterFlagCompletionFunc("field", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeSecretFields(appSvc, cmd, args, toComplete, secretFieldCompletionOptions{useEnvSelector: true})
	})
	_ = cmd.RegisterFlagCompletionFunc("env", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return completeSecretEnvNames(appSvc, cmd, args, toComplete)
	})
	return cmd
}

func newSecretListCmd(appSvc *app.App) *cobra.Command {
	var long bool
	cmd := &cobra.Command{
		Use:   "list [query]",
		Short: "List secret items without secret values",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretListQueries(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			configPath, vaultPath := runtimePaths(cmd)
			items, err := appSvc.ListSecretItems(configPath, vaultPath, app.ListSecretItemsRequest{Query: query, Long: long})
			if err != nil {
				return err
			}
			renderSecretList(cmd.OutOrStdout(), items, long)
			return nil
		},
	}
	cmd.Flags().BoolVar(&long, "long", false, "Show id, template, primary env name, and value-set state")
	return cmd
}

func newSecretInfoCmd(appSvc *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <selector>",
		Short: "Print secret item metadata without secret values",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretSelectors(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			info, err := appSvc.SecretInfo(configPath, vaultPath, args[0])
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
		Short: "Remove a secret item",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretSelectors(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("secret rm requires --yes")
			}
			configPath, vaultPath := runtimePaths(cmd)
			return appSvc.RemoveSecret(configPath, vaultPath, args[0])
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm removal")
	return cmd
}

func renderSecretCreateResult(w io.Writer, info app.SecretInfo) {
	cliui.Summary(w, "created secret", []cliui.SummaryField{
		{Key: "title", Value: info.Title},
		{Key: "id", Value: info.ID},
		{Key: "template", Value: info.Template},
		{Key: "env", Value: info.PrimaryEnv},
	})
}

func renderSecretList(w io.Writer, items []app.SecretItemSummary, long bool) {
	if len(items) == 0 {
		cliui.Empty(w, "No secrets found.")
		return
	}
	if !long {
		cliui.Title(w, fmt.Sprintf("Secrets (%d):", len(items)))
		for _, item := range items {
			fmt.Fprintf(w, "  %s\n", item.Title)
		}
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		state := "unset"
		if item.ValueSet {
			state = "set"
		}
		rows = append(rows, []string{item.ID, item.Title, item.Template, item.Env, state})
	}
	cliui.RenderTable(w, cliui.Table{Title: "Secrets", Headers: []string{"ID", "TITLE", "TEMPLATE", "ENV", "VALUE"}, Rows: rows})
}
