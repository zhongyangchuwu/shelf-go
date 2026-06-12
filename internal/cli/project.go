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
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Project utilities"}
	cmd.AddCommand(newProjectIDCmd())
	cmd.AddCommand(newProjectInitCmd())
	cmd.AddCommand(newProjectExplainCmd())
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
			if failed {
				return fmt.Errorf("project manifest check failed")
			}
			return nil
		},
	}
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
