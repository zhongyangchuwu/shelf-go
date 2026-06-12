package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/manifest"
	"github.com/zhongyangchuwu/shelf-go/internal/render"
	"github.com/zhongyangchuwu/shelf-go/internal/store"
)

type resolved struct {
	path    string
	envName string
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
			envOwners := map[string]string{}
			failed := false
		for _, entry := range m.Secrets {
			if entry.IsPrefix() {
				matches := st.List(entry.Prefix)
				if len(matches) == 0 {
					if entry.IsRequired() {
						fmt.Fprintf(cmd.OutOrStdout(), "fail %s (prefix) no matches required\n", entry.Prefix)
						failed = true
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "warn %s (prefix) no matches optional\n", entry.Prefix)
					}
					continue
				}
				for _, p := range matches {
					secret, ok := st.Get(p)
					if !ok {
						continue
					}
					envName := render.EnvName(p, secret)
					if ownerPath, exists := envOwners[envName]; exists {
						fmt.Fprintf(cmd.OutOrStdout(), "fail %s env name %s conflicts with %s\n", p, envName, ownerPath)
						failed = true
						continue
					}
					envOwners[envName] = p
					fmt.Fprintf(cmd.OutOrStdout(), "ok   %s -> %s\n", p, envName)
				}
			} else {
				secret, ok := st.Get(entry.Path)
				if !ok {
					if entry.IsRequired() {
						fmt.Fprintf(cmd.OutOrStdout(), "fail %s missing required\n", entry.Path)
						failed = true
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "warn %s missing optional\n", entry.Path)
					}
					continue
				}
				envName := entry.Env
				if envName == "" {
					envName = render.EnvName(entry.Path, secret)
				}
				if ownerPath, exists := envOwners[envName]; exists {
					fmt.Fprintf(cmd.OutOrStdout(), "fail %s env name %s conflicts with %s\n", entry.Path, envName, ownerPath)
					failed = true
					continue
				}
				envOwners[envName] = entry.Path
				fmt.Fprintf(cmd.OutOrStdout(), "ok   %s -> %s\n", entry.Path, envName)
			}
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

			// Resolve all entries into ordered path list with env names.
			var paths []string
			var resolvedEntries []resolved
			envOwners := map[string]string{}
			failed := false

			for _, entry := range m.Secrets {
				if entry.IsPrefix() {
					matches := st.List(entry.Prefix)
					sort.Strings(matches)
					for _, p := range matches {
						secret, ok := st.Get(p)
						if !ok {
							continue
						}
						envName := render.EnvName(p, secret)
						if owner, exists := envOwners[envName]; exists {
							fmt.Fprintf(cmd.OutOrStderr(), "fail %s env name %s conflicts with %s\n", p, envName, owner)
							failed = true
							continue
						}
						envOwners[envName] = p
						paths = append(paths, p)
						resolvedEntries = append(resolvedEntries, resolved{path: p, envName: envName})
					}
				} else {
					secret, ok := st.Get(entry.Path)
					if !ok {
						if entry.IsRequired() {
							fmt.Fprintf(cmd.OutOrStderr(), "fail %s missing required\n", entry.Path)
							failed = true
						} else {
							fmt.Fprintf(cmd.OutOrStderr(), "warn %s missing optional\n", entry.Path)
						}
						continue
					}
					envName := entry.Env
					if envName == "" {
						envName = render.EnvName(entry.Path, secret)
					}
					if owner, exists := envOwners[envName]; exists {
						fmt.Fprintf(cmd.OutOrStderr(), "fail %s env name %s conflicts with %s\n", entry.Path, envName, owner)
						failed = true
						continue
					}
					envOwners[envName] = entry.Path
					paths = append(paths, entry.Path)
					resolvedEntries = append(resolvedEntries, resolved{path: entry.Path, envName: envName})
				}
			}

			if failed {
				return fmt.Errorf("project export failed")
			}
			if len(paths) == 0 {
				return fmt.Errorf("no secrets to export")
			}

			switch format {
			case "env":
				return renderProjectExportEnv(cmd, st, resolvedEntries)
			case "shell":
				return renderProjectExportShell(cmd, st, resolvedEntries)
			case "json":
				return renderProjectExportJSON(cmd, st, resolvedEntries)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "env", "Output format")
	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{"env", "shell", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}

func renderProjectExportEnv(cmd *cobra.Command, st *store.Store, entries []resolved) error {
	var b strings.Builder
	for _, e := range entries {
		secret, _ := st.Get(e.path)
		value, err := render.ValueString(secret.Value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s=%s\n", e.envName, value)
	}
	fmt.Fprint(cmd.OutOrStdout(), b.String())
	return nil
}

func renderProjectExportShell(cmd *cobra.Command, st *store.Store, entries []resolved) error {
	var b strings.Builder
	for _, e := range entries {
		secret, _ := st.Get(e.path)
		value, err := render.ValueString(secret.Value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "export %s=%s\n", e.envName, render.ShellQuote(value))
	}
	fmt.Fprint(cmd.OutOrStdout(), b.String())
	return nil
}

func renderProjectExportJSON(cmd *cobra.Command, st *store.Store, entries []resolved) error {
	payload := map[string]json.RawMessage{}
	for _, e := range entries {
		secret, _ := st.Get(e.path)
		payload[e.envName] = secret.Value
	}
	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, key := range keys {
		value, err := json.Marshal(payload[key])
		if err != nil {
			return err
		}
		fmt.Fprintf(&buf, "  %q: %s", key, string(value))
		if i < len(keys)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("}\n")
	fmt.Fprint(cmd.OutOrStdout(), buf.String())
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
