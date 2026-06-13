package cli

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runShelf(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}
func runShelfWithInput(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetIn(strings.NewReader(input))
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
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
	t.Chdir(dir)
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
