package project

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

func ID() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return IDFromRoot(root)
}

func IDBestEffort(root string) string {
	id, err := IDFromRoot(root)
	if err != nil {
		return root
	}
	return id
}

func IDFromRoot(root string) (string, error) {
	remoteBytes, err := exec.Command("git", "-C", root, "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("remote origin url not found")
	}
	return NormalizeRemote(strings.TrimSpace(string(remoteBytes)))
}

func Root() (string, error) {
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

func NormalizeRemote(remote string) (string, error) {
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
