package cli

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Project utilities"}
	cmd.AddCommand(&cobra.Command{
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
	})
	return cmd
}

func projectID() (string, error) {
	rootBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a Git worktree")
	}
	_ = strings.TrimSpace(string(rootBytes))
	remoteBytes, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("remote origin url not found")
	}
	return normalizeRemote(strings.TrimSpace(string(remoteBytes)))
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
