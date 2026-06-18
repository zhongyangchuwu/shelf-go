package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-openai"); err != nil {
		t.Fatalf("set openai: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openrouter/accounts/personal:api_key", "sk-openrouter"); err != nil {
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
	out, err := runShelf(t, "--vault", data, "project", "explain")
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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-openai"); err != nil {
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
	out, err := runShelf(t, "--vault", data, "project", "explain")
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
func TestProjectAddPathEntry(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key", "--env", "OPENAI_API_KEY")
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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal", "--optional")
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
func TestProjectAddRejectsEnvForPrefix(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	_, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal", "--env", "OPENAI_API_KEY")
	if err == nil {
		t.Fatalf("expected --env for prefix to fail")
	}
	if !strings.Contains(err.Error(), "--env is only valid for path entries") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestProjectAddRejectsDuplicatePath(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("first add: %v", err)
	}
	_, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key")
	if err == nil {
		t.Fatalf("expected duplicate add to fail")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestProjectAddRejectsMissingSecret(t *testing.T) {
	_, data := setupProjectTest(t)
	_, err := runShelf(t, "--vault", data, "project", "add", "providers/nonexistent/accounts/personal:api_key")
	if err == nil {
		t.Fatalf("expected add of missing secret to fail")
	}
	if !strings.Contains(err.Error(), "secret not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestProjectAddRejectsEmptyPrefix(t *testing.T) {
	_, data := setupProjectTest(t)
	_, err := runShelf(t, "--vault", data, "project", "add", "providers/nonexistent")
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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	_, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key")
	if err == nil {
		t.Fatalf("expected add to fail without manifest")
	}
	if !strings.Contains(err.Error(), "run `shelf project init` first") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestProjectRmRemovesEntry(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("project add: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "rm", "providers/openai/accounts/personal:api_key")
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
	_, err := runShelf(t, "--vault", data, "project", "rm", "providers/nonexistent:api_key")
	if err == nil {
		t.Fatalf("expected rm of missing entry to fail")
	}
	if !strings.Contains(err.Error(), "entry not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestProjectRmRemovesPrefixEntry(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal", "--optional"); err != nil {
		t.Fatalf("project add prefix: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "rm", "providers/openai/accounts/personal")
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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openrouter/accounts/personal:api_key", "sk-router"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("add path: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openrouter/accounts/personal", "--optional"); err != nil {
		t.Fatalf("add prefix: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "list")
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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("project export env: %v", err)
	}
	if !strings.Contains(out, "OPENAI_API_KEY=sk-test") {
		t.Fatalf("unexpected env output:\n%s", out)
	}
}
func TestProjectExportShell(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "shell")
	if err != nil {
		t.Fatalf("project export shell: %v", err)
	}
	if !strings.Contains(out, "export OPENAI_API_KEY=sk-test") {
		t.Fatalf("unexpected shell output:\n%s", out)
	}
}
func TestProjectExportJSON(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "json")
	if err != nil {
		t.Fatalf("project export json: %v", err)
	}
	if !strings.Contains(out, `"OPENAI_API_KEY"`) || !strings.Contains(out, `"sk-test"`) {
		t.Fatalf("unexpected json output:\n%s", out)
	}
}
func TestProjectExportJSONConvertsValuesToStrings(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:number", "123"); err != nil {
		t.Fatalf("set number: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "app:number", "--env", "APP_NUMBER"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "json")
	if err != nil {
		t.Fatalf("project export json: %v", err)
	}
	if !strings.Contains(out, `"APP_NUMBER": "123"`) {
		t.Fatalf("expected string-converted JSON value:\n%s", out)
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
	_, err := runShelf(t, "--vault", filepath.Join(dir, "secrets.json"), "project", "export", "--format", "env")
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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
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
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/work:api_key", "sk-work"); err != nil {
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
	_, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
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
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-personal"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/work:api_key", "sk-work"); err != nil {
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
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
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
func TestProjectExportFailsOnRequiredEmptyPrefix(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"},{"prefix":"providers/missing","required":true}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err == nil {
		t.Fatalf("expected required empty prefix to fail")
	}
	if !strings.Contains(out, "fail providers/missing (prefix) no matches required") {
		t.Fatalf("missing required prefix failure:\n%s", out)
	}
}
func TestProjectExportWarnsOnOptionalEmptyPrefix(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"},{"prefix":"providers/missing","required":false}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("expected optional empty prefix to succeed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "APP_TOKEN=secret") {
		t.Fatalf("missing present secret:\n%s", out)
	}
	if !strings.Contains(out, "warn providers/missing (prefix) no matches optional") {
		t.Fatalf("missing optional prefix warning:\n%s", out)
	}
}
func TestProjectExplainHandlesPrefixEntries(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test"); err != nil {
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
	out, err := runShelf(t, "--vault", data, "project", "explain")
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
func TestProjectExplainWarnsAboutParentEnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	t.Setenv("APP_TOKEN", "parent")
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "explain")
	if err != nil {
		t.Fatalf("project explain: %v\n%s", err, out)
	}
	if !strings.Contains(out, "warn APP_TOKEN overrides existing environment variable") {
		t.Fatalf("missing override warning:\n%s", out)
	}
	if strings.Contains(out, "secret") || strings.Contains(out, "parent") {
		t.Fatalf("explain leaked env value:\n%s", out)
	}
}
