package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/render"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

type editableSecret struct {
	GroupPath   string          `json:"group_path"`
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Env         string          `json:"env,omitempty"`
	Description string          `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
}

func newEditableSecret(path string, secret store.Secret) (editableSecret, error) {
	id, err := store.ParseSecretID(path)
	if err != nil {
		return editableSecret{}, err
	}
	return editableSecret{GroupPath: id.GroupPath, Key: id.Key, Value: secret.Value, Env: secret.Env, Description: secret.Description, Tags: secret.Tags}, nil
}

func (e editableSecret) secret() (store.SecretID, store.Secret) {
	return store.SecretID{GroupPath: e.GroupPath, Key: e.Key}, store.Secret{Value: e.Value, Env: e.Env, Description: e.Description, Tags: e.Tags}
}

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
		Use:               "set <path> <value>",
		Short:             "Create a secret",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completeSecretSetPathArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, st, unlock, err := loadRuntimeForWrite(cmd)
			if err != nil {
				return err
			}
			defer unlock()
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
			runtime, st, unlock, err := loadRuntimeForWrite(cmd)
			if err != nil {
				return err
			}
			defer unlock()
			secret, ok := st.Get(args[0])
			if !ok {
				return fmt.Errorf("secret not found: %s", args[0])
			}
			editable, err := newEditableSecret(args[0], secret)
			if err != nil {
				return err
			}
			bytes, err := json.MarshalIndent(editable, "", "  ")
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
			var updated editableSecret
			if err := json.Unmarshal(edited, &updated); err != nil {
				return err
			}
			id, secret := updated.secret()
			if err := st.Update(args[0], id, secret); err != nil {
				return err
			}
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

func completeSecretSetPathArg(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	_, st, err := loadRuntime(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSecretSetPathForArgs(st.List(""), args, toComplete)
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

func newSecretRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "rm <path>",
		Short:             "Remove a secret",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeSecretPaths,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, st, unlock, err := loadRuntimeForWrite(cmd)
			if err != nil {
				return err
			}
			defer unlock()
			if !st.Delete(args[0]) {
				return fmt.Errorf("secret not found: %s", args[0])
			}
			return st.Save()
		},
	}
	return cmd
}
