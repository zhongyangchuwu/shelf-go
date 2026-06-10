package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/han/shelf-go/internal/render"
	"github.com/han/shelf-go/internal/store"
	"github.com/spf13/cobra"
)

func newSecretCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "secret", Short: "Manage secrets"}
	cmd.AddCommand(newSecretSetCmd())
	cmd.AddCommand(newSecretGetCmd())
	cmd.AddCommand(newSecretListCmd())
	cmd.AddCommand(newSecretInfoCmd())
	cmd.AddCommand(newSecretEditCmd())
	cmd.AddCommand(newSecretRmCmd())
	return cmd
}

func newSecretSetCmd() *cobra.Command {
	var envName, description string
	var tags []string
	var force bool
	cmd := &cobra.Command{
		Use:   "set <path> <value>",
		Short: "Create a secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			value, err := store.ParseValue(args[1])
			if err != nil {
				return err
			}
			secret := store.Secret{Value: value, Env: envName, Description: description, Tags: tags}
			if err := st.Set(args[0], secret, force); err != nil {
				return err
			}
			return st.Save()
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

func newSecretGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get <path>",
		Short:             "Print a secret value",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeSecretPaths,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			secret, ok := st.Get(args[0])
			if !ok {
				return fmt.Errorf("secret not found: %s", args[0])
			}
			value, err := render.ValueString(secret.Value)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}
	return cmd
}

func newSecretListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [prefix]",
		Short: "List secret paths",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
			}
			for _, path := range st.List(prefix) {
				fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}
	return cmd
}

func newSecretInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "info <path>",
		Short:             "Print secret metadata as JSON",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeSecretPaths,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			info, ok := st.Info(args[0])
			if !ok {
				return fmt.Errorf("secret not found: %s", args[0])
			}
			bytes, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(bytes))
			return nil
		},
	}
	return cmd
}

func newSecretEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "edit <path>",
		Short:             "Edit a secret object as JSON",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeSecretPaths,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			secret, ok := st.Get(args[0])
			if !ok {
				return fmt.Errorf("secret not found: %s", args[0])
			}
			bytes, err := json.MarshalIndent(secret, "", "  ")
			if err != nil {
				return err
			}
			tmp, err := os.CreateTemp("", "shelf-secret-*.json")
			if err != nil {
				return err
			}
			tmpName := tmp.Name()
			defer os.Remove(tmpName)
			if _, err := tmp.Write(append(bytes, '\n')); err != nil {
				tmp.Close()
				return err
			}
			if err := tmp.Close(); err != nil {
				return err
			}
			editor := runtime.Editor
			editorCmd := exec.Command("sh", "-c", "$SHELF_EDITOR \"$SHELF_EDIT_FILE\"")
			editorCmd.Env = append(os.Environ(), "SHELF_EDITOR="+editor, "SHELF_EDIT_FILE="+tmpName)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr
			if err := editorCmd.Run(); err != nil {
				return err
			}
			edited, err := os.ReadFile(tmpName)
			if err != nil {
				return err
			}
			var updated store.Secret
			if err := json.Unmarshal(edited, &updated); err != nil {
				return err
			}
			if err := store.ValidateSecret(updated); err != nil {
				return err
			}
			st.Data.Secrets[args[0]] = updated
			return st.Save()
		},
	}
	return cmd
}

func completeSecretPaths(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	_, st, err := loadRuntime(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	paths := st.List(toComplete)
	comps := make([]cobra.Completion, 0, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(path, toComplete) {
			comps = append(comps, cobra.Completion(path))
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func newSecretRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "rm <path>",
		Short:             "Remove a secret",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeSecretPaths,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			if !st.Delete(args[0]) {
				return fmt.Errorf("secret not found: %s", args[0])
			}
			return st.Save()
		},
	}
	return cmd
}
