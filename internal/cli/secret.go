package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/render"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
	"golang.org/x/term"
)

var secretAddIsTerminal = term.IsTerminal
var secretAddReadPassword = term.ReadPassword

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
	cmd.AddCommand(newSecretAddCmd())
	cmd.AddCommand(newSecretSetCmd())
	cmd.AddCommand(newSecretGetCmd())
	cmd.AddCommand(newSecretListCmd())
	cmd.AddCommand(newSecretInfoCmd())
	cmd.AddCommand(newSecretEditCmd())
	cmd.AddCommand(newExportCmd())
	cmd.AddCommand(newSecretRmCmd())
	return cmd
}

func newSecretAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "add [path-or-group]",
		Short:             "Interactively create a secret",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeSecretSetPathArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !secretAddIsTerminal(int(os.Stdin.Fd())) {
				return fmt.Errorf("secret add requires a terminal; use `shelf secret set` for scripts")
			}
			return updateVault(cmd, func(st *store.Store) error {
				prompt := newSecretAddPrompt(cmd.InOrStdin(), cmd.OutOrStdout(), st)
				path, secret, force, err := prompt.collect(args)
				if err != nil {
					return err
				}
				if err := st.Set(path, secret, force); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "added %s\n", path)
				return nil
			})
		},
	}
	return cmd
}

type secretAddPrompt struct {
	in  *bufio.Reader
	out io.Writer
	st  *store.Store
}

func newSecretAddPrompt(in io.Reader, out io.Writer, st *store.Store) secretAddPrompt {
	return secretAddPrompt{in: bufio.NewReader(in), out: out, st: st}
}

func (p secretAddPrompt) collect(args []string) (string, store.Secret, bool, error) {
	p.printGroupHints()
	path, err := p.collectPath(args)
	if err != nil {
		return "", store.Secret{}, false, err
	}
	force := false
	if _, exists := p.st.Get(path); exists {
		overwrite, err := p.confirm("secret exists; overwrite? [y/N]: ")
		if err != nil {
			return "", store.Secret{}, false, err
		}
		if !overwrite {
			return "", store.Secret{}, false, fmt.Errorf("secret already exists: %s", path)
		}
		force = true
	}
	value, err := p.password("value: ")
	if err != nil {
		return "", store.Secret{}, false, err
	}
	if value == "" {
		return "", store.Secret{}, false, fmt.Errorf("secret value is required")
	}
	envName, err := p.line("env (optional): ")
	if err != nil {
		return "", store.Secret{}, false, err
	}
	description, err := p.line("description (optional): ")
	if err != nil {
		return "", store.Secret{}, false, err
	}
	tagText, err := p.line("tags comma-separated (optional): ")
	if err != nil {
		return "", store.Secret{}, false, err
	}
	raw, err := store.ParseValue(value)
	if err != nil {
		return "", store.Secret{}, false, err
	}
	secret := store.Secret{Value: raw, Env: strings.TrimSpace(envName), Description: strings.TrimSpace(description), Tags: parsePromptTags(tagText)}
	return path, secret, force, nil
}

func (p secretAddPrompt) collectPath(args []string) (string, error) {
	if len(args) == 0 {
		path, err := p.line("path (group/key as group:path): ")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(path), nil
	}
	input := strings.TrimSpace(args[0])
	if strings.Contains(input, ":") {
		return input, nil
	}
	key, err := p.line("key: ")
	if err != nil {
		return "", err
	}
	return input + ":" + strings.TrimSpace(key), nil
}

func (p secretAddPrompt) printGroupHints() {
	groups := existingGroups(p.st.List(""))
	if len(groups) == 0 {
		return
	}
	fmt.Fprintln(p.out, "existing groups:")
	limit := len(groups)
	if limit > 8 {
		limit = 8
	}
	for _, group := range groups[:limit] {
		fmt.Fprintf(p.out, "  %s\n", group)
	}
	if len(groups) > limit {
		fmt.Fprintf(p.out, "  ... %d more\n", len(groups)-limit)
	}
}

func (p secretAddPrompt) line(label string) (string, error) {
	fmt.Fprint(p.out, label)
	text, err := p.in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(text, "\r\n"), nil
}

func (p secretAddPrompt) password(label string) (string, error) {
	fmt.Fprint(p.out, label)
	bytes, err := secretAddReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(p.out)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (p secretAddPrompt) confirm(label string) (bool, error) {
	answer, err := p.line(label)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func parsePromptTags(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func existingGroups(paths []string) []string {
	seen := map[string]struct{}{}
	groups := make([]string, 0, len(paths))
	for _, path := range paths {
		group, _, ok := strings.Cut(path, ":")
		if !ok || group == "" {
			continue
		}
		if _, exists := seen[group]; exists {
			continue
		}
		seen[group] = struct{}{}
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return groups
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
			return updateVault(cmd, func(st *store.Store) error {
				value, err := store.ParseValue(args[1])
				if err != nil {
					return err
				}
				secret := store.Secret{Value: value, Env: envName, Description: description, Tags: tags}
				return st.Set(args[0], secret, force)
			})
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
			runtime, vault, err := loadVault(cmd)
			if err != nil {
				return err
			}
			return vault.Update(func(st *store.Store) error {
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
				if err := tmp.Chmod(0o600); err != nil {
					tmp.Close()
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
				return st.Update(args[0], id, secret)
			})
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
			return updateVault(cmd, func(st *store.Store) error {
				if !st.Delete(args[0]) {
					return fmt.Errorf("secret not found: %s", args[0])
				}
				return nil
			})
		},
	}
	return cmd
}
