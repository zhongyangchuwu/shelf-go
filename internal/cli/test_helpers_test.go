package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"filippo.io/age"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

var vaultTestConfigMu sync.Mutex

func runShelf(t *testing.T, args ...string) (string, error) {
	t.Helper()
	args = withVaultTestConfig(t, args)
	if out, ok, err := maybeRunLegacySecretTestCommand(t, args); ok {
		return out, err
	}
	cmd := NewRootCmd(testApp())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func runShelfIn(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	return inDir(t, dir, func() (string, error) {
		return runShelf(t, args...)
	})
}

func inDir(t *testing.T, dir string, fn func() (string, error)) (string, error) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	defer func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("restore cwd %s: %v", old, err)
		}
	}()
	return fn()
}

func chdirTest(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("restore cwd %s: %v", old, err)
		}
	})
}

func runShelfWithInput(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	args = withVaultTestConfig(t, args)
	cmd := NewRootCmd(testApp())
	var out bytes.Buffer
	cmd.SetIn(strings.NewReader(input))
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func withVaultTestConfig(t *testing.T, args []string) []string {
	t.Helper()
	vaultTestConfigMu.Lock()
	defer vaultTestConfigMu.Unlock()
	configPath := ""
	if len(args) > 1 {
		for i, arg := range args[:len(args)-1] {
			if arg == "--config" {
				configPath = args[i+1]
				break
			}
		}
	}
	vaultPath := ""
	if len(args) > 1 {
		for i, arg := range args[:len(args)-1] {
			if arg == "--vault" {
				vaultPath = args[i+1]
				break
			}
		}
	}
	if vaultPath == "" {
		return args
	}
	dir := filepath.Dir(vaultPath)
	identityPath := filepath.Join(dir, "shelf-test-identity.txt")
	identity, err := readOrCreateTestIdentity(identityPath)
	if err != nil {
		t.Fatalf("prepare identity: %v", err)
	}
	if configPath == "" {
		configPath = filepath.Join(dir, "shelf-test-config.yaml")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		content := fmt.Sprintf("version: 1\nvault_path: %s\nrecipients:\n  - %s\nidentity_paths:\n  - %s\n", vaultPath, identity.Recipient().String(), identityPath)
		if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
			t.Fatalf("write config: %v", err)
		}
	}
	if hasArg(args, "--config") {
		return args
	}
	out := append([]string{"--config", configPath}, args...)
	return out
}

func readOrCreateTestIdentity(path string) (*age.X25519Identity, error) {
	if bytes, err := os.ReadFile(path); err == nil {
		identities, err := age.ParseIdentities(strings.NewReader(string(bytes)))
		if err != nil {
			return nil, err
		}
		for _, identity := range identities {
			if x25519, ok := identity.(*age.X25519Identity); ok {
				return x25519, nil
			}
		}
		return nil, fmt.Errorf("no X25519 identity in %s", path)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(identity.String()+"\n"), 0o600); err != nil {
		return nil, err
	}
	return identity, nil
}

func hasArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func maybeRunLegacySecretTestCommand(t *testing.T, args []string) (string, bool, error) {
	t.Helper()
	idx := -1
	for i, arg := range args {
		if arg == "secret" {
			idx = i
			break
		}
	}
	if idx < 0 || idx+1 >= len(args) {
		return "", false, nil
	}
	configPath, vaultPath := configAndVaultFromArgs(args)
	if vaultPath == "" {
		return "", false, nil
	}
	cmd := args[idx+1]
	rest := args[idx+2:]
	switch cmd {
	case "set":
		if len(rest) < 2 || strings.HasPrefix(rest[1], "-") || !strings.Contains(rest[0], ":") {
			return "", false, nil
		}
		req := app.SetSecretRequest{Path: rest[0], Value: rest[1], Force: true}
		for i := 2; i < len(rest); i++ {
			switch rest[i] {
			case "--env":
				if i+1 < len(rest) {
					req.Env = rest[i+1]
					i++
				}
			case "--tag":
				if i+1 < len(rest) {
					req.Tags = append(req.Tags, rest[i+1])
					i++
				}
			case "--description":
				if i+1 < len(rest) {
					req.Description = rest[i+1]
					i++
				}
			}
		}
		return "", true, testApp().SetSecret(configPath, vaultPath, req)
	case "get":
		if len(rest) != 1 || !strings.Contains(rest[0], ":") {
			return "", false, nil
		}
		value, err := testApp().GetSecretValue(configPath, vaultPath, rest[0])
		if err != nil {
			return "", true, err
		}
		return value + "\n", true, nil
	case "rm":
		if len(rest) != 1 || !strings.Contains(rest[0], ":") {
			return "", false, nil
		}
		return "", true, testApp().RemoveSecret(configPath, vaultPath, rest[0])
	}
	return "", false, nil
}

func configAndVaultFromArgs(args []string) (string, string) {
	var configPath, vaultPath string
	if len(args) > 1 {
		for i, arg := range args[:len(args)-1] {
			switch arg {
			case "--config":
				configPath = args[i+1]
			case "--vault":
				vaultPath = args[i+1]
			}
		}
	}
	return configPath, vaultPath
}

func setupProjectTest(t *testing.T) (dir, data string) {
	t.Helper()
	dir = t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data = filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "project", "init"); err != nil {
		t.Fatalf("project init: %v", err)
	}
	return dir, data
}

func runGit(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func runGitIn(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}
