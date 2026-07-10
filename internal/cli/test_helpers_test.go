package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func runShelf(t *testing.T, args ...string) (string, error) {
	t.Helper()
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

func runShelfWithInput(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd(testApp())
	var out bytes.Buffer
	cmd.SetIn(bytes.NewBufferString(input))
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
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

func setupProjectTest(t *testing.T) (dir, data string) {
	t.Helper()
	dir = t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data = filepath.Join(dir, "vault.json")
	if _, err := runShelf(t, "--vault", data, "project", "init"); err != nil {
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

func testApp() *app.App { return app.NewDefault() }
