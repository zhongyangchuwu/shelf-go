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
)

var vaultTestConfigMu sync.Mutex

func runShelf(t *testing.T, args ...string) (string, error) {
	t.Helper()
	args = withVaultTestConfig(t, args)
	cmd := NewRootCmd()
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
	cmd := NewRootCmd()
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
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--config" {
			configPath = args[i+1]
			break
		}
	}
	vaultPath := ""
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--vault" {
			vaultPath = args[i+1]
			break
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
func withPromptPassword(t *testing.T, password string) {
	t.Helper()
	origIsTerminal := secretAddIsTerminal
	origReadPassword := secretAddReadPassword
	secretAddIsTerminal = func(int) bool { return true }
	secretAddReadPassword = func(int) ([]byte, error) { return []byte(password), nil }
	t.Cleanup(func() {
		secretAddIsTerminal = origIsTerminal
		secretAddReadPassword = origReadPassword
	})
}
func setupProjectTest(t *testing.T) (dir, data string) {
	t.Helper()
	dir = t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data = filepath.Join(dir, "secrets.json")
	// Init the project manifest.
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
