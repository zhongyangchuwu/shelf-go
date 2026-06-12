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

func TestSecretSetGetListInfoExport(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")

	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openrouter/accounts/personal:api_key", "sk-xxx", "--env", "OPENROUTER_API_KEY", "--tag", "ai"); err != nil {
		t.Fatalf("set secret: %v", err)
	}

	out, err := runShelf(t, "--data", data, "secret", "get", "providers/openrouter/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if out != "sk-xxx\n" {
		t.Fatalf("unexpected get output: %q", out)
	}

	out, err = runShelf(t, "--data", data, "secret", "list", "providers/openrouter")
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	if out != "providers/openrouter/accounts/personal:api_key\n" {
		t.Fatalf("unexpected list output: %q", out)
	}

	out, err = runShelf(t, "--data", data, "secret", "info", "providers/openrouter/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("info secret: %v", err)
	}
	if strings.Contains(out, "sk-xxx") {
		t.Fatalf("info leaked value: %s", out)
	}
	for _, want := range []string{"\"path\"", "\"group_path\": \"providers/openrouter/accounts/personal\"", "\"key\": \"api_key\"", "\"value_set\": true", "\"env\": \"OPENROUTER_API_KEY\"", "\"tags\""} {
		if !strings.Contains(out, want) {
			t.Fatalf("info output missing %q: %s", want, out)
		}
	}

	out, err = runShelf(t, "--data", data, "export", "providers/openrouter/accounts/personal:api_key", "--format", "shell")
	if err != nil {
		t.Fatalf("export shell: %v", err)
	}
	if out != "export OPENROUTER_API_KEY=sk-xxx\n" {
		t.Fatalf("unexpected export output: %q", out)
	}
}

func TestSecretGetAndExportPreserveJSONNumbers(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	want := "12345678901234567890"
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:big_number", want); err != nil {
		t.Fatalf("set big number: %v", err)
	}
	out, err := runShelf(t, "--data", data, "secret", "get", "app:big_number")
	if err != nil {
		t.Fatalf("get big number: %v", err)
	}
	if out != want+"\n" {
		t.Fatalf("secret get changed number: %q", out)
	}
	out, err = runShelf(t, "--data", data, "export", "app:big_number", "--format", "json", "--all")
	if err != nil {
		t.Fatalf("export json: %v", err)
	}
	if !strings.Contains(out, want) {
		t.Fatalf("export json changed number: %s", out)
	}
}

func TestExportExactPathDoesNotIncludeLongerPrefixMatch(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set exact: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token_extra", "two", "--env", "APP_TOKEN_EXTRA"); err != nil {
		t.Fatalf("set prefix: %v", err)
	}
	out, err := runShelf(t, "--data", data, "export", "app:token", "--format", "shell")
	if err != nil {
		t.Fatalf("export exact: %v", err)
	}
	if out != "export APP_TOKEN=one\n" {
		t.Fatalf("exact export included extra paths: %q", out)
	}
}

func TestExportFiltersEnvOnlyByDefaultAndAllFlag(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:with_env", "one", "--env", "WITH_ENV"); err != nil {
		t.Fatalf("set with env: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:no_env", "two"); err != nil {
		t.Fatalf("set no env: %v", err)
	}
	out, err := runShelf(t, "--data", data, "export", "app", "--format", "shell")
	if err != nil {
		t.Fatalf("export without --all: %v", err)
	}
	if strings.Contains(out, "app:no_env") || strings.Contains(out, "NO_ENV") || strings.Contains(out, "=two") {
		t.Fatalf("default export leaked secret without env: %q", out)
	}
	if !strings.Contains(out, "WITH_ENV=one") {
		t.Fatalf("default export missing env secret: %q", out)
	}
	out, err = runShelf(t, "--data", data, "export", "app", "--format", "shell", "--all")
	if err != nil {
		t.Fatalf("export with --all: %v", err)
	}
	if !strings.Contains(out, "WITH_ENV=one") || !strings.Contains(out, "APP_NO_ENV=two") {
		t.Fatalf("--all export missing secret: %q", out)
	}
}

func TestSecretSetRefusesOverwriteWithoutForce(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "two"); err == nil {
		t.Fatalf("expected overwrite without --force to fail")
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "two", "--force"); err != nil {
		t.Fatalf("force set: %v", err)
	}
	out, err := runShelf(t, "--data", data, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get after force: %v", err)
	}
	if out != "two\n" {
		t.Fatalf("unexpected value after force: %q", out)
	}
}

func TestSecretEditUsesEditorAndValidatesJSON(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"app\",\"key\":\"token\",\"value\":\"edited\",\"env\":\"APP_TOKEN\",\"tags\":[\"app\"]}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--data", data, "secret", "edit", "app:token"); err != nil {
		t.Fatalf("edit secret: %v", err)
	}
	out, err := runShelf(t, "--data", data, "export", "app:token", "--format", "shell")
	if err != nil {
		t.Fatalf("export edited: %v", err)
	}
	if out != "export APP_TOKEN=edited\n" {
		t.Fatalf("unexpected edited export: %q", out)
	}
}

func TestSecretEditCanRenamePath(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"services/app\",\"key\":\"api_key\",\"value\":\"one\",\"env\":\"APP_API_KEY\"}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--data", data, "secret", "edit", "app:token"); err != nil {
		t.Fatalf("edit rename: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "get", "app:token"); err == nil {
		t.Fatalf("old path still exists")
	}
	out, err := runShelf(t, "--data", data, "secret", "get", "services/app:api_key")
	if err != nil {
		t.Fatalf("get renamed path: %v", err)
	}
	if out != "one\n" {
		t.Fatalf("unexpected renamed value: %q", out)
	}
}

func TestSecretEditRefusesRenameCollision(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set token: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:other", "two"); err != nil {
		t.Fatalf("set other: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"app\",\"key\":\"other\",\"value\":\"one\"}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--data", data, "secret", "edit", "app:token"); err == nil {
		t.Fatalf("expected edit rename collision to fail")
	}
	out, err := runShelf(t, "--data", data, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("old path missing after failed rename: %v", err)
	}
	if out != "one\n" {
		t.Fatalf("unexpected old value after failed rename: %q", out)
	}
}

func TestConcurrentSecretSetKeepsAllWrites(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	const count = 20
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := runShelf(t, "--data", data, "secret", "set", fmt.Sprintf("app:key_%02d", i), fmt.Sprintf("value-%02d", i))
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent set: %v", err)
		}
	}
	out, err := runShelf(t, "--data", data, "secret", "list", "app")
	if err != nil {
		t.Fatalf("list after concurrent set: %v", err)
	}
	lines := strings.Fields(out)
	if len(lines) != count {
		t.Fatalf("lost writes: got %d paths, want %d\n%s", len(lines), count, out)
	}
}
func TestCompletionCommandGeneratesZsh(t *testing.T) {
	out, err := runShelf(t, "completion", "zsh")
	if err != nil {
		t.Fatalf("completion zsh: %v", err)
	}
	if !strings.Contains(out, "#compdef shelf") {
		t.Fatalf("unexpected zsh completion output")
	}
}

func TestProjectIDNormalizesGitRemote(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
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
	t.Chdir(dir)
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
	const want = "{\n  \"version\": 1,\n  \"secrets\": []\n}\n"
	if string(bytes) != want {
		t.Fatalf("unexpected manifest content:\n%s", string(bytes))
	}
}

func TestProjectInitRequiresForceToOverwrite(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := os.WriteFile(manifestPath, []byte("{\n  \"version\": 1,\n  \"secrets\": []\n}\n"), 0o600); err != nil {
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

func TestProjectExplainReportsStatuses(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-openai"); err != nil {
		t.Fatalf("set openai: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openrouter/accounts/personal:api_key", "sk-openrouter"); err != nil {
		t.Fatalf("set openrouter: %v", err)
	}
	manifest := `{
  "version": 1,
  "secrets": [
    {
      "path": "providers/openai/accounts/personal:api_key",
      "env": "OPENAI_API_KEY",
      "required": true
    },
    {
      "path": "providers/anthropic/accounts/personal:api_key",
      "required": false
    },
    {
      "path": "providers/deepseek/accounts/personal:api_key"
    },
    {
      "path": "providers/openrouter/accounts/personal:api_key",
      "env": "OPENAI_API_KEY",
      "required": true
    }
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "explain")
	if err == nil {
		t.Fatalf("expected project explain failure")
	}
	for _, want := range []string{
		"project:",
		"root:    ",
		"config:  .shelf.json",
		"ok   providers/openai/accounts/personal:api_key -> OPENAI_API_KEY",
		"warn providers/anthropic/accounts/personal:api_key missing optional",
		"fail providers/deepseek/accounts/personal:api_key missing required",
		"fail providers/openrouter/accounts/personal:api_key env name OPENAI_API_KEY conflicts with providers/openai/accounts/personal:api_key",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("project explain output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "sk-openai") || strings.Contains(out, "sk-openrouter") {
		t.Fatalf("project explain should not print secret values:\n%s", out)
	}
}

func TestProjectExplainWarnsOptionalMissingWithoutFail(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-openai"); err != nil {
		t.Fatalf("set openai: %v", err)
	}
	manifest := `{
  "version": 1,
  "secrets": [
    {
      "path": "providers/openai/accounts/personal:api_key",
      "required": true
    },
    {
      "path": "providers/anthropic/accounts/personal:api_key",
      "required": false
    }
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "explain")
	if err != nil {
		t.Fatalf("project explain: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ok   providers/openai/accounts/personal:api_key") {
		t.Fatalf("missing ok line:\n%s", out)
	}
	if !strings.Contains(out, "warn providers/anthropic/accounts/personal:api_key missing optional") {
		t.Fatalf("missing optional warning:\n%s", out)
	}
}

func TestProjectInitAndExplainFailOutsideGit(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runShelf(t, "project", "init"); err == nil {
		t.Fatalf("expected project init to fail outside git")
	} else if !strings.Contains(err.Error(), "not inside a Git worktree") {
		t.Fatalf("unexpected init error: %v", err)
	}
	if _, err := runShelf(t, "project", "explain"); err == nil {
		t.Fatalf("expected project explain to fail outside git")
	} else if !strings.Contains(err.Error(), "not inside a Git worktree") {
		t.Fatalf("unexpected explain error: %v", err)
	}
}

func TestProjectExplainPromptsInitWhenManifestMissing(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runShelf(t, "project", "explain"); err == nil {
		t.Fatalf("expected project explain to fail without manifest")
	} else if !strings.Contains(err.Error(), "run `shelf project init`") {
		t.Fatalf("unexpected explain error: %v", err)
	}
}

// --- v0.3: project add/rm/list/export ---

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

func TestProjectAddPathEntry(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key", "--env", "OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("project add: %v", err)
	}
	if !strings.Contains(out, "added providers/openai/accounts/personal:api_key") {
		t.Fatalf("unexpected add output: %q", out)
	}
	// Verify manifest persisted.
	content, err := os.ReadFile(filepath.Join(dir, ".shelf.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !strings.Contains(string(content), "OPENAI_API_KEY") {
		t.Fatalf("manifest missing env override:\n%s", string(content))
	}
}

func TestProjectAddPrefixEntry(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal", "--optional")
	if err != nil {
		t.Fatalf("project add prefix: %v", err)
	}
	if !strings.Contains(out, "added providers/openai/accounts/personal") {
		t.Fatalf("unexpected add output: %q", out)
	}
	content, err := os.ReadFile(filepath.Join(dir, ".shelf.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !strings.Contains(string(content), `"prefix"`) {
		t.Fatalf("manifest missing prefix entry:\n%s", string(content))
	}
	if !strings.Contains(string(content), `"required": false`) {
		t.Fatalf("manifest missing optional flag:\n%s", string(content))
	}
}

func TestProjectAddRejectsDuplicatePath(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("first add: %v", err)
	}
	_, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key")
	if err == nil {
		t.Fatalf("expected duplicate add to fail")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectAddRejectsMissingSecret(t *testing.T) {
	_, data := setupProjectTest(t)
	_, err := runShelf(t, "--data", data, "project", "add", "providers/nonexistent/accounts/personal:api_key")
	if err == nil {
		t.Fatalf("expected add of missing secret to fail")
	}
	if !strings.Contains(err.Error(), "secret not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectAddRejectsEmptyPrefix(t *testing.T) {
	_, data := setupProjectTest(t)
	_, err := runShelf(t, "--data", data, "project", "add", "providers/nonexistent")
	if err == nil {
		t.Fatalf("expected add of non-matching prefix to fail")
	}
	if !strings.Contains(err.Error(), "no secrets match prefix") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectAddPromptsInitWhenNoManifest(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	_, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key")
	if err == nil {
		t.Fatalf("expected add to fail without manifest")
	}
	if !strings.Contains(err.Error(), "run `shelf project init` first") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectRmRemovesEntry(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("project add: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "rm", "providers/openai/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("project rm: %v", err)
	}
	if !strings.Contains(out, "removed providers/openai/accounts/personal:api_key") {
		t.Fatalf("unexpected rm output: %q", out)
	}
	content, err := os.ReadFile(filepath.Join(dir, ".shelf.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if strings.Contains(string(content), "openai") {
		t.Fatalf("manifest should not contain removed entry:\n%s", string(content))
	}
}

func TestProjectRmFailsOnMissingEntry(t *testing.T) {
	_, data := setupProjectTest(t)
	_, err := runShelf(t, "--data", data, "project", "rm", "providers/nonexistent:api_key")
	if err == nil {
		t.Fatalf("expected rm of missing entry to fail")
	}
	if !strings.Contains(err.Error(), "entry not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectRmRemovesPrefixEntry(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal", "--optional"); err != nil {
		t.Fatalf("project add prefix: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "rm", "providers/openai/accounts/personal")
	if err != nil {
		t.Fatalf("project rm prefix: %v", err)
	}
	if !strings.Contains(out, "removed providers/openai/accounts/personal") {
		t.Fatalf("unexpected rm output: %q", out)
	}
	content, err := os.ReadFile(filepath.Join(dir, ".shelf.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if strings.Contains(string(content), "openai") {
		t.Fatalf("manifest should not contain removed entry:\n%s", string(content))
	}
}

func TestProjectListShowsEntries(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openrouter/accounts/personal:api_key", "sk-router"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("add path: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "project", "add", "providers/openrouter/accounts/personal", "--optional"); err != nil {
		t.Fatalf("add prefix: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "list")
	if err != nil {
		t.Fatalf("project list: %v", err)
	}
	if !strings.Contains(out, "path   providers/openai/accounts/personal:api_key -> OPENAI_API_KEY (required)") {
		t.Fatalf("missing path line:\n%s", out)
	}
	if !strings.Contains(out, "prefix providers/openrouter/accounts/personal (optional)") {
		t.Fatalf("missing prefix line:\n%s", out)
	}
	if strings.Contains(out, "sk-test") || strings.Contains(out, "sk-router") {
		t.Fatalf("project list should not print secret values:\n%s", out)
	}
}

func TestProjectExportEnv(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("project export env: %v", err)
	}
	if !strings.Contains(out, "OPENAI_API_KEY=sk-test") {
		t.Fatalf("unexpected env output:\n%s", out)
	}
}

func TestProjectExportShell(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "export", "--format", "shell")
	if err != nil {
		t.Fatalf("project export shell: %v", err)
	}
	if !strings.Contains(out, "export OPENAI_API_KEY=sk-test") {
		t.Fatalf("unexpected shell output:\n%s", out)
	}
}

func TestProjectExportJSON(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "export", "--format", "json")
	if err != nil {
		t.Fatalf("project export json: %v", err)
	}
	if !strings.Contains(out, `"OPENAI_API_KEY"`) || !strings.Contains(out, `"sk-test"`) {
		t.Fatalf("unexpected json output:\n%s", out)
	}
}

func TestProjectExportFailsOnRequiredMissing(t *testing.T) {
	_, data := setupProjectTest(t)
	// Add a path entry but don't create the secret.
	manifest := `{"version":1,"secrets":[{"path":"providers/openai/accounts/personal:api_key","required":true}]}`
	dir, _ := setupProjectTest(t)
	_ = data // use the dir from setupProjectTest
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, err := runShelf(t, "--data", filepath.Join(dir, "secrets.json"), "project", "export", "--format", "env")
	if err == nil {
		t.Fatalf("expected export to fail with missing required")
	}
}

func TestProjectExportSkipsOptionalMissing(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{
  "version": 1,
  "secrets": [
    {"path": "providers/openai/accounts/personal:api_key"},
    {"path": "providers/anthropic/accounts/personal:api_key", "required": false}
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("expected export to succeed with optional missing: %v\n%s", err, out)
	}
	if !strings.Contains(out, "OPENAI_API_KEY=sk-test") {
		t.Fatalf("missing present secret:\n%s", out)
	}
	if strings.Contains(out, "ANTHROPIC") {
		t.Fatalf("should not export missing optional:\n%s", out)
	}
	if !strings.Contains(out, "warn providers/anthropic/accounts/personal:api_key missing optional") {
		t.Fatalf("should warn about missing optional:\n%s", out)
	}
}

func TestProjectExportFailsOnEnvConflict(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/work:api_key", "sk-work"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	// Both explicitly override to same env name.
	manifest := `{
  "version": 1,
  "secrets": [
    {"path": "providers/openai/accounts/personal:api_key", "env": "OPENAI_API_KEY"},
    {"path": "providers/openai/accounts/work:api_key", "env": "OPENAI_API_KEY"}
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, err := runShelf(t, "--data", data, "project", "export", "--format", "env")
	if err == nil {
		t.Fatalf("expected env name conflict to fail")
	}
}

func TestProjectExportExpandsPrefixSorted(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-personal"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/work:api_key", "sk-work"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{
  "version": 1,
  "secrets": [
    {"prefix": "providers/openai/accounts"}
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("project export: %v\n%s", err, out)
	}
	// Prefix expansion should produce stable sorted output.
	personalIdx := strings.Index(out, "PROVIDERS_OPENAI_ACCOUNTS_PERSONAL_API_KEY=sk-personal")
	workIdx := strings.Index(out, "PROVIDERS_OPENAI_ACCOUNTS_WORK_API_KEY=sk-work")
	if personalIdx == -1 || workIdx == -1 {
		t.Fatalf("missing expanded secrets:\n%s", out)
	}
	if personalIdx > workIdx {
		t.Fatalf("expected personal before work in sorted output:\n%s", out)
	}
}

func TestProjectExplainHandlesPrefixEntries(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{
  "version": 1,
  "secrets": [
    {"prefix": "providers/openai/accounts", "required": true},
    {"prefix": "providers/nonexistent", "required": false}
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--data", data, "project", "explain")
	if err != nil {
		t.Fatalf("project explain: %v\n%s", err, out)
	}
	if !strings.Contains(out, "ok   providers/openai/accounts/personal:api_key ->") {
		t.Fatalf("missing ok line for prefix expansion:\n%s", out)
	}
	if !strings.Contains(out, "warn providers/nonexistent (prefix) no matches optional") {
		t.Fatalf("missing warn for empty optional prefix:\n%s", out)
	}
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

func TestDoctorReportsHealthyStore(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--config", configPath, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	for _, want := range []string{"ok   config resolves", "ok   version", "ok   data file exists", "ok   store loads", "ok   data file mode"} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, out)
		}
	}
}

func TestDoctorFailsInvalidStore(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	if err := os.WriteFile(data, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write invalid data: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err == nil {
		t.Fatalf("expected doctor to fail invalid store")
	}
	if !strings.Contains(out, "fail store loads") {
		t.Fatalf("doctor output missing store failure:\n%s", out)
	}
}

func TestDoctorChecksCompletionFromFpath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	completionDir := filepath.Join(dir, "zfunc")
	if err := os.MkdirAll(completionDir, 0o700); err != nil {
		t.Fatalf("mkdir completion dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(completionDir, "_shelf"), []byte("#compdef shelf\n"), 0o600); err != nil {
		t.Fatalf("write completion: %v", err)
	}
	t.Setenv("FPATH", filepath.Join(dir, "missing")+":"+completionDir)
	if _, err := runShelf(t, "--config", configPath, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	want := "ok   completion installed (" + filepath.Join(completionDir, "_shelf") + ")"
	if !strings.Contains(out, want) {
		t.Fatalf("doctor did not use FPATH completion path %q:\n%s", want, out)
	}
}

func TestSecretRmRemovesPathAndFailsOnMissing(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "rm", "app:token"); err != nil {
		t.Fatalf("rm: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "get", "app:token"); err == nil {
		t.Fatalf("expected get after rm to fail")
	}
	if _, err := runShelf(t, "--data", data, "secret", "rm", "app:token"); err == nil {
		t.Fatalf("expected rm on missing to fail")
	}
}

func TestInitCreatesFilesAndReportsStatus(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "secrets.json")
	cfg := filepath.Join(dir, "shelf.yaml")

	out, err := runShelf(t, "--config", cfg, "--data", data, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	for _, want := range []string{data, "(created)", cfg, "(created)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("init output missing %q: %s", want, out)
		}
	}

	out2, err := runShelf(t, "--config", cfg, "--data", data, "init")
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	for _, want := range []string{data, "(exists)", cfg, "(exists)"} {
		if !strings.Contains(out2, want) {
			t.Fatalf("second init output missing %q: %s", want, out2)
		}
	}

	out3, err := runShelf(t, "--config", cfg, "--data", data, "init", "--minimal")
	if err != nil {
		t.Fatalf("minimal init: %v", err)
	}
	if strings.Contains(out3, "config") {
		t.Fatalf("minimal init mentioned config: %s", out3)
	}
}

func TestInitForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "secrets.json")
	cfg := filepath.Join(dir, "config.yaml")
	if _, err := runShelf(t, "--config", cfg, "--data", data, "init"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runShelf(t, "--config", cfg, "--data", data, "secret", "set", "app:token", "val"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := runShelf(t, "--config", cfg, "--data", data, "init", "--force"); err != nil {
		t.Fatalf("force init: %v", err)
	}
	out, err := runShelf(t, "--config", cfg, "--data", data, "secret", "list")
	if err != nil {
		t.Fatalf("list after force init: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty store after force init, got: %s", out)
	}
}

func TestVersionFlagPrintsVersion(t *testing.T) {
	out, err := runShelf(t, "--version")
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !strings.HasPrefix(out, "shelf") {
		t.Fatalf("unexpected version output: %q", out)
	}
}
