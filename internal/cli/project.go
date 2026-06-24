package cli

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/manifest"
	"github.com/zhongyangchuwu/shelf-go/internal/render"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

type resolved struct {
	path    string
	envName string
	value   string
}

type projectDiagnostic struct {
	status  string
	path    string
	message string
}

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
			id, err := projectID()
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
			root, err := gitRoot()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, manifest.FileName)
			existed := false
			if _, err := os.Stat(manifestPath); err == nil {
				existed = true
				if !force {
					return fmt.Errorf("%s already exists (use --force to overwrite)", manifest.FileName)
				}
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err := manifest.Save(manifestPath, manifest.New()); err != nil {
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
			root, err := gitRoot()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, manifest.FileName)
			m, err := manifest.Load(manifestPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found in %s; run `shelf project init`", manifest.FileName, root)
			}
			if err != nil {
				return err
			}

			project := projectIDBestEffort(root)
			fmt.Fprintf(cmd.OutOrStdout(), "project: %s\n", project)
			fmt.Fprintf(cmd.OutOrStdout(), "root:    %s\n", root)
			fmt.Fprintf(cmd.OutOrStdout(), "config:  %s\n\n", manifest.FileName)

			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			resolvedEntries, diagnostics := resolveProjectEntries(m, st)
			failed := false
			for _, diagnostic := range diagnostics {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s\n", diagnostic.status, diagnostic.path, diagnostic.message)
				if diagnostic.status == "fail" {
					failed = true
				}
			}
			for _, entry := range resolvedEntries {
				fmt.Fprintf(cmd.OutOrStdout(), "ok   %s -> %s\n", entry.path, entry.envName)
			}
			for _, warning := range envOverrideWarnings(resolvedEntries, os.Environ()) {
				fmt.Fprintln(cmd.OutOrStdout(), warning)
			}
			if failed {
				return fmt.Errorf("project manifest check failed")
			}
			return nil
		},
	}
}

func newProjectAddCmd() *cobra.Command {
	var envName string
	var optional bool
	cmd := &cobra.Command{
		Use:   "add <path-or-prefix>",
		Short: "Add a secret path or prefix to project manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := gitRoot()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, manifest.FileName)
			if _, err := os.Stat(manifestPath); errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found; run `shelf project init` first", manifest.FileName)
			} else if err != nil {
				return err
			}

			input := args[0]
			isPrefix := !strings.Contains(input, ":")
			if isPrefix && envName != "" {
				return fmt.Errorf("--env is only valid for path entries")
			}

			// Validate against the store.
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}
			if isPrefix {
				matches := st.List(input)
				if len(matches) == 0 {
					return fmt.Errorf("no secrets match prefix: %s", input)
				}
			} else {
				if _, ok := st.Get(input); !ok {
					return fmt.Errorf("secret not found: %s", input)
				}
			}

			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}

			entry := manifest.Entry{}
			if isPrefix {
				entry.Prefix = input
			} else {
				entry.Path = input
				if envName != "" {
					entry.Env = envName
				}
			}
			if optional {
				entry.Required = new(bool)
			}

			if err := m.AddEntry(entry); err != nil {
				return err
			}
			if err := manifest.Save(manifestPath, m); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added %s\n", entry.Key())
			return nil
		},
	}
	cmd.Flags().StringVar(&envName, "env", "", "Environment variable name override (path entries only)")
	cmd.Flags().BoolVar(&optional, "optional", false, "Mark entry as optional")
	return cmd
}

func newProjectRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <path-or-prefix>",
		Short: "Remove an entry from project manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := gitRoot()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, manifest.FileName)
			m, err := manifest.Load(manifestPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found", manifest.FileName)
			}
			if err != nil {
				return err
			}
			if !m.RemoveEntry(args[0]) {
				return fmt.Errorf("entry not found: %s", args[0])
			}
			if err := manifest.Save(manifestPath, m); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", args[0])
			return nil
		},
	}
}

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List project manifest entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := gitRoot()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, manifest.FileName)
			m, err := manifest.Load(manifestPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found", manifest.FileName)
			}
			if err != nil {
				return err
			}
			for _, entry := range m.Secrets {
				if entry.IsPrefix() {
					req := "required"
					if !entry.IsRequired() {
						req = "optional"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "prefix %s (%s)\n", entry.Prefix, req)
				} else {
					req := "required"
					if !entry.IsRequired() {
						req = "optional"
					}
					if entry.Env != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "path   %s -> %s (%s)\n", entry.Path, entry.Env, req)
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "path   %s (%s)\n", entry.Path, req)
					}
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
			root, err := gitRoot()
			if err != nil {
				return err
			}
			manifestPath := filepath.Join(root, manifest.FileName)
			m, err := manifest.Load(manifestPath)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found", manifest.FileName)
			}
			if err != nil {
				return err
			}

			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}

			resolvedEntries, diagnostics := resolveProjectEntries(m, st)
			failed := false
			for _, diagnostic := range diagnostics {
				fmt.Fprintf(cmd.OutOrStderr(), "%s %s %s\n", diagnostic.status, diagnostic.path, diagnostic.message)
				if diagnostic.status == "fail" {
					failed = true
				}
			}
			if failed {
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

func resolveProjectEntries(m manifest.Manifest, st *store.Store) ([]resolved, []projectDiagnostic) {
	entries := make([]resolved, 0)
	diagnostics := make([]projectDiagnostic, 0)
	envOwners := map[string]string{}

	for _, entry := range m.Secrets {
		if entry.IsPrefix() {
			matches := st.List(entry.Prefix)
			if len(matches) == 0 {
				path := entry.Prefix + " (prefix)"
				if entry.IsRequired() {
					diagnostics = append(diagnostics, projectDiagnostic{status: "fail", path: path, message: "no matches required"})
				} else {
					diagnostics = append(diagnostics, projectDiagnostic{status: "warn", path: path, message: "no matches optional"})
				}
				continue
			}
			for _, path := range matches {
				secret, ok := st.Get(path)
				if !ok {
					continue
				}
				appendResolvedEntry(&entries, &diagnostics, envOwners, path, secret, "")
			}
			continue
		}

		secret, ok := st.Get(entry.Path)
		if !ok {
			if entry.IsRequired() {
				diagnostics = append(diagnostics, projectDiagnostic{status: "fail", path: entry.Path, message: "missing required"})
			} else {
				diagnostics = append(diagnostics, projectDiagnostic{status: "warn", path: entry.Path, message: "missing optional"})
			}
			continue
		}
		appendResolvedEntry(&entries, &diagnostics, envOwners, entry.Path, secret, entry.Env)
	}

	return entries, diagnostics
}

func appendResolvedEntry(entries *[]resolved, diagnostics *[]projectDiagnostic, envOwners map[string]string, path string, secret store.Secret, envOverride string) {
	envName := envOverride
	if envName == "" {
		var err error
		envName, err = render.EnvName(path, secret)
		if err != nil {
			*diagnostics = append(*diagnostics, projectDiagnostic{status: "fail", path: path, message: err.Error()})
			return
		}
	}
	if owner, exists := envOwners[envName]; exists {
		*diagnostics = append(*diagnostics, projectDiagnostic{status: "fail", path: path, message: fmt.Sprintf("env name %s conflicts with %s", envName, owner)})
		return
	}
	envOwners[envName] = path
	value, err := render.ValueString(secret.Value)
	if err != nil {
		*diagnostics = append(*diagnostics, projectDiagnostic{status: "fail", path: path, message: err.Error()})
		return
	}
	*entries = append(*entries, resolved{path: path, envName: envName, value: value})
}

func renderProjectExportEnv(cmd *cobra.Command, entries []resolved) error {
	bindings := make([]render.Binding, 0, len(entries))
	for _, entry := range entries {
		bindings = append(bindings, render.Binding{EnvName: entry.envName, Value: entry.value})
	}
	out, err := render.EnvBindings(bindings)
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), out)
	return nil
}

func renderProjectExportShell(cmd *cobra.Command, entries []resolved) error {
	bindings := make([]render.Binding, 0, len(entries))
	for _, entry := range entries {
		bindings = append(bindings, render.Binding{EnvName: entry.envName, Value: entry.value})
	}
	out, err := render.ShellBindings(bindings)
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), out)
	return nil
}

func renderProjectExportJSON(cmd *cobra.Command, entries []resolved) error {
	bindings := make([]render.Binding, 0, len(entries))
	for _, entry := range entries {
		bindings = append(bindings, render.Binding{EnvName: entry.envName, Value: entry.value})
	}
	out, err := render.JSONBindings(bindings)
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), out)
	return nil
}

func projectID() (string, error) {
	root, err := gitRoot()
	if err != nil {
		return "", err
	}
	return projectIDFromRoot(root)
}

func projectIDBestEffort(root string) string {
	id, err := projectIDFromRoot(root)
	if err != nil {
		return root
	}
	return id
}

func projectIDFromRoot(root string) (string, error) {
	remoteBytes, err := exec.Command("git", "-C", root, "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("remote origin url not found")
	}
	return normalizeRemote(strings.TrimSpace(string(remoteBytes)))
}

func gitRoot() (string, error) {
	rootBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a Git worktree")
	}
	root := strings.TrimSpace(string(rootBytes))
	if root == "" {
		return "", fmt.Errorf("not inside a Git worktree")
	}
	return root, nil
}

func normalizeRemote(remote string) (string, error) {
	if remote == "" {
		return "", fmt.Errorf("remote url is empty")
	}
	if strings.HasPrefix(remote, "git@") && strings.Contains(remote, ":") {
		rest := strings.TrimPrefix(remote, "git@")
		parts := strings.SplitN(rest, ":", 2)
		return cleanRemotePath(parts[0], parts[1])
	}
	if strings.HasPrefix(remote, "ssh://") || strings.HasPrefix(remote, "https://") || strings.HasPrefix(remote, "http://") {
		u, err := url.Parse(remote)
		if err != nil {
			return "", err
		}
		host := u.Hostname()
		path := strings.TrimPrefix(u.Path, "/")
		return cleanRemotePath(host, path)
	}
	return "", fmt.Errorf("unsupported remote url: %s", remote)
}

func cleanRemotePath(host, path string) (string, error) {
	host = strings.ToLower(strings.TrimSpace(host))
	path = strings.TrimSpace(path)
	path = strings.TrimSuffix(path, ".git")
	path = strings.Trim(path, "/")
	if host == "" || path == "" {
		return "", fmt.Errorf("invalid remote identity")
	}
	return host + "/" + path, nil
}
