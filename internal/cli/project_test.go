package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectIDNormalizesGitRemote(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runGit(t, "remote", "add", "origin", "git@github.com:Owner/Repo.git"); err != nil {
		t.Fatalf("git remote add: %v", err)
	}
	out, err := runShelf(t, "project", "id")
	if err != nil {
		t.Fatalf("project id: %v", err)
	}
	if out != "github.com/Owner/Repo\n" {
		t.Fatalf("unexpected project id: %q", out)
	}
}

func TestProjectInitCreatesManifestAtGitRoot(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	out, err := runShelf(t, "project", "init")
	if err != nil {
		t.Fatalf("project init: %v", err)
	}
	manifestPath := filepath.Join(dir, ".shelf.json")
	if !strings.Contains(out, "manifest: "+manifestPath+" (created)") {
		t.Fatalf("unexpected init output: %q", out)
	}
	bytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	const want = "{\n  \"version\": 1,\n  \"bindings\": []\n}\n"
	if string(bytes) != want {
		t.Fatalf("unexpected manifest content:\n%s", string(bytes))
	}
}

func TestProjectInitRequiresForceToOverwrite(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := os.WriteFile(manifestPath, []byte("{\n  \"version\": 1,\n  \"bindings\": []\n}\n"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if _, err := runShelf(t, "project", "init"); err == nil {
		t.Fatalf("expected project init to refuse overwrite without --force")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := runShelf(t, "project", "init", "--force")
	if err != nil {
		t.Fatalf("project init --force: %v", err)
	}
	if !strings.Contains(out, "manifest: "+manifestPath+" (overwritten)") {
		t.Fatalf("unexpected init --force output: %q", out)
	}
}

func TestProjectAddKeepsManifestValueFree(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelfWithInput(t, "sk-secret-value\n", "--vault", data, "secret", "create", "openai", "--id", "item_openai", "--env", "OPENAI_API_KEY", "--stdin"); err != nil {
		t.Fatalf("create secret: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "add", "item_openai")
	if err != nil {
		t.Fatalf("project add: %v", err)
	}
	if !strings.Contains(out, "added item_openai.value") {
		t.Fatalf("unexpected add output: %q", out)
	}
	manifest, err := os.ReadFile(filepath.Join(dir, ".shelf.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if strings.Contains(string(manifest), "sk-secret-value") {
		t.Fatalf("manifest contains secret value:\n%s", manifest)
	}
	for _, want := range []string{"item_openai", "OPENAI_API_KEY"} {
		if !strings.Contains(string(manifest), want) {
			t.Fatalf("manifest missing non-secret binding %q:\n%s", want, manifest)
		}
	}
}

func TestProjectExportUsesItemVault(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelfWithInput(t, "sk-project\n", "--vault", data, "secret", "create", "openai", "--id", "item_openai", "--env", "OPENAI_API_KEY", "--stdin"); err != nil {
		t.Fatalf("create secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "item_openai"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "shell")
	if err != nil {
		t.Fatalf("project export shell: %v\n%s", err, out)
	}
	if !strings.Contains(out, "export OPENAI_API_KEY=sk-project") {
		t.Fatalf("project export missing value:\n%s", out)
	}
}

func TestProjectStatusFailsOutsideGit(t *testing.T) {
	chdirTest(t, t.TempDir())
	if _, err := runShelf(t, "project", "status"); err == nil {
		t.Fatalf("expected project status to fail outside git")
	} else if !strings.Contains(err.Error(), "not inside a Git worktree") {
		t.Fatalf("unexpected status error: %v", err)
	}
}
