package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
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
	const want = "{\n  \"version\": 1,\n  \"secrets\": []\n}\n"
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
func TestProjectStatusReportsResolutionDiagnostics(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
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
	out, err := runShelf(t, "--vault", data, "project", "status")
	if err == nil {
		t.Fatalf("expected project status failure with missing required + conflict")
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
			t.Fatalf("project status output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "sk-openai") || strings.Contains(out, "sk-openrouter") {
		t.Fatalf("project status should not print secret values:\n%s", out)
	}
}
func TestProjectInitFailsOutsideGit(t *testing.T) {
	chdirTest(t, t.TempDir())
	if _, err := runShelf(t, "project", "init"); err == nil {
		t.Fatalf("expected project init to fail outside git")
	} else if !strings.Contains(err.Error(), "not inside a Git worktree") {
		t.Fatalf("unexpected init error: %v", err)
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
func TestProjectStatusPromptsInitWhenManifestMissing(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runShelf(t, "project", "status"); err == nil {
		t.Fatalf("expected project status to fail without manifest")
	} else if !strings.Contains(err.Error(), "run `shelf project init`") {
		t.Fatalf("unexpected status error: %v", err)
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

func TestProjectAddKeepsManifestValueFree(t *testing.T) {
	dir, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-secret-value", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("project add: %v", err)
	}
	manifest, err := os.ReadFile(filepath.Join(dir, ".shelf.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if strings.Contains(string(manifest), "sk-secret-value") {
		t.Fatalf("manifest contains secret value:\n%s", manifest)
	}
	for _, want := range []string{"providers/openai/accounts/personal:api_key", "OPENAI_API_KEY"} {
		if !strings.Contains(string(manifest), want) {
			t.Fatalf("manifest missing non-secret binding %q:\n%s", want, manifest)
		}
	}
}

func TestProjectExportUsesEncryptedVaultWithoutPlaintextSideData(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-encrypted-project", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("project export env: %v\n%s", err, out)
	}
	if !strings.Contains(out, "OPENAI_API_KEY=sk-encrypted-project") {
		t.Fatalf("project export missing encrypted-vault value:\n%s", out)
	}
	content, err := os.ReadFile(data)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if strings.Contains(string(content), "sk-encrypted-project") || strings.Contains(string(content), "providers/openai/accounts/personal:api_key") {
		t.Fatalf("encrypted vault contains plaintext project data")
	}
}
func TestProjectAddPromptsInitWhenNoManifest(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
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
func TestProjectExportDefaultsToShell(t *testing.T) {
	_, data := setupProjectTest(t)
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-test", "--env", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai/accounts/personal:api_key"); err != nil {
		t.Fatalf("add: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export")
	if err != nil {
		t.Fatalf("project export default: %v", err)
	}
	if out != "export OPENAI_API_KEY=sk-test\n" {
		t.Fatalf("unexpected default output:\n%s", out)
	}
}
func TestProjectAddV2BindsItemEnvField(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "vault.v2.json")
	if _, err := runShelfWithInput(t, "fake-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin"); err != nil {
		t.Fatalf("create v2 secret: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(`{"version":2,"bindings":[]}`), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "add", "item_github")
	if err != nil {
		t.Fatalf("project add v2: %v\n%s", err, out)
	}
	if !strings.Contains(out, "added item_github.value") {
		t.Fatalf("unexpected add output: %q", out)
	}
	list, err := runShelf(t, "project", "list")
	if err != nil {
		t.Fatalf("project list v2: %v", err)
	}
	if !strings.Contains(list, "item   item_github.value -> GITHUB_TOKEN") {
		t.Fatalf("unexpected v2 list output: %q", list)
	}
}

func TestProjectExportV2MaterializesBoundField(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "vault.v2.json")
	if _, err := runShelfWithInput(t, "fake-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin"); err != nil {
		t.Fatalf("create v2 secret: %v", err)
	}
	manifest := `{"version":2,"bindings":[{"item":"item_github","field":"value","env":"GITHUB_TOKEN"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("project export v2: %v\n%s", err, out)
	}
	if out != "GITHUB_TOKEN=fake-token\n" {
		t.Fatalf("unexpected export: %q", out)
	}
}

func TestProjectRunV2InjectsBoundField(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "vault.v2.json")
	if _, err := runShelfWithInput(t, "fake-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin"); err != nil {
		t.Fatalf("create v2 secret: %v", err)
	}
	manifest := `{"version":2,"bindings":[{"item":"item_github","field":"value","env":"GITHUB_TOKEN"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "run", "--", "sh", "-c", "printf %s \"$GITHUB_TOKEN\"")
	if err != nil {
		t.Fatalf("project run v2: %v\n%s", err, out)
	}
	if out != "fake-token" {
		t.Fatalf("unexpected child output: %q", out)
	}
}

func TestProjectStatusV2IsValueFree(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "vault.v2.json")
	if _, err := runShelfWithInput(t, "fake-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin"); err != nil {
		t.Fatalf("create v2 secret: %v", err)
	}
	manifest := `{"version":2,"bindings":[{"item":"item_github","field":"value","env":"GITHUB_TOKEN"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "status")
	if err != nil {
		t.Fatalf("project status v2: %v\n%s", err, out)
	}
	if strings.Contains(out, "fake-token") {
		t.Fatalf("status leaked secret value: %q", out)
	}
	if !strings.Contains(out, "ok   item_github.value -> GITHUB_TOKEN") {
		t.Fatalf("missing v2 binding status: %q", out)
	}
}

func TestProjectExportFailsOnRequiredMissing(t *testing.T) {
	dir, data := setupProjectTest(t)
	// Add a path entry but don't create the secret.
	manifest := `{"version":1,"secrets":[{"path":"providers/openai/accounts/personal:api_key","required":true}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err == nil {
		t.Fatalf("expected export to fail with missing required")
	}
}
func TestProjectStatusWarnsAboutParentEnvOverride(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	parentValue := "leak-check-parent-value-123"
	secretValue := "leak-check-secret-value-123"
	t.Setenv("APP_TOKEN", parentValue)
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", secretValue, "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "status")
	if err != nil {
		t.Fatalf("project status: %v\n%s", err, out)
	}
	if !strings.Contains(out, "warn APP_TOKEN overrides existing environment variable") {
		t.Fatalf("missing override warning:\n%s", out)
	}
	for _, leaked := range []string{secretValue, parentValue} {
		if strings.Contains(out, leaked) {
			t.Fatalf("project status leaked env value %q:\n%s", leaked, out)
		}
	}
}

func TestProjectStatusScansEnvFilesWithoutLeakingValues(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret-value-789", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	// Write env files with known keys but keep values out of assertions.
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("DB_HOST=localhost\nDB_PORT=5432\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("DB_HOST=\nDB_PORT=\n"), 0o600); err != nil {
		t.Fatalf("write .env.example: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "status")
	if err != nil {
		t.Fatalf("project status: %v\n%s", err, out)
	}
	if !strings.Contains(out, ".env") {
		t.Fatalf("project status output should mention .env:\n%s", out)
	}
	if !strings.Contains(out, ".env.example") {
		t.Fatalf("project status output should mention .env.example:\n%s", out)
	}
	// Must not print any env file keys or values.
	for _, leaked := range []string{"DB_HOST", "localhost", "DB_PORT", "5432"} {
		if strings.Contains(out, leaked) {
			t.Fatalf("project status leaked env file content %q:\n%s", leaked, out)
		}
	}
	// Must not print vault secret values either.
	if strings.Contains(out, "secret-value-789") {
		t.Fatalf("project status leaked vault secret value:\n%s", out)
	}
}

func TestProjectStatusEnvFileFindingsDoNotFailWhenBindingsHealthy(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "healthy-value", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"path":"app:token"}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	// Env file with empty and unbound entries — findings are informational.
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("EMPTY_VAL=\nUNSET_VAR\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "status"); err != nil {
		t.Fatalf("expected project status to succeed despite env-file findings:\n%v", err)
	}
}

func TestProjectInitOutputIncludesEnvSummary(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("SOME_KEY=some_value\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	out, err := runShelf(t, "project", "init")
	if err != nil {
		t.Fatalf("project init: %v", err)
	}
	if !strings.Contains(out, "manifest:") {
		t.Fatalf("project init missing manifest line:\n%s", out)
	}
	// Init output should include env-file summary section.
	if !strings.Contains(out, ".env") {
		t.Fatalf("project init output should mention .env:\n%s", out)
	}
	// Must not print env file values.
	if strings.Contains(out, "some_value") {
		t.Fatalf("project init leaked env file value:\n%s", out)
	}
}

func TestProjectExplainIsRemoved(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runShelf(t, "project", "init"); err != nil {
		t.Fatalf("project init: %v", err)
	}
	_, err := runShelf(t, "project", "explain")
	if err == nil {
		t.Fatalf("expected project explain to fail; it has been removed")
	}
	if !strings.Contains(err.Error(), "unknown command") &&
		!strings.Contains(err.Error(), "not found") &&
		!strings.Contains(err.Error(), "explain") {
		t.Fatalf("unexpected error for removed command: %v", err)
	}
}

func TestProjectAddCompletionSuggestsVaultSecrets(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret"); err != nil {
		t.Fatalf("set app token: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai:api_key", "sk"); err != nil {
		t.Fatalf("set provider key: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "init"); err != nil {
		t.Fatalf("project init: %v", err)
	}

	out, err := runShelf(t, "--vault", data, "__complete", "project", "add", "")
	if err != nil {
		t.Fatalf("complete project add: %v\n%s", err, out)
	}
	for _, want := range []string{"app:", "providers/openai:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing completion %q in:\n%s", want, out)
		}
	}
}

func TestProjectRmCompletionSuggestsManifestEntries(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret"); err != nil {
		t.Fatalf("set app token: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai:api_key", "sk"); err != nil {
		t.Fatalf("set provider key: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "init"); err != nil {
		t.Fatalf("project init: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "app:token"); err != nil {
		t.Fatalf("project add path: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "providers/openai"); err != nil {
		t.Fatalf("project add prefix: %v", err)
	}

	out, err := runShelf(t, "--vault", data, "__complete", "project", "rm", "")
	if err != nil {
		t.Fatalf("complete project rm: %v\n%s", err, out)
	}
	for _, want := range []string{"app:token", "providers/openai"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing completion %q in:\n%s", want, out)
		}
	}
}

// --- v0.4: project configure ---

func TestProjectConfigureAppearsInHelp(t *testing.T) {
	dir, _ := setupProjectTest(t)

	out, err := runShelfIn(t, dir, "project", "--help")
	if err != nil {
		t.Fatalf("project --help: %v\n%s", err, out)
	}
	if !strings.Contains(out, "configure") {
		t.Fatalf("project --help should list configure:\n%s", out)
	}

	out, err = runShelfIn(t, dir, "project", "configure", "--help")
	if err != nil {
		t.Fatalf("project configure --help: %v\n%s", err, out)
	}
	if !strings.Contains(out, "configure") {
		t.Fatalf("project configure --help should include 'configure':\n%s", out)
	}
}

func TestProjectConfigureModelQueuesEnvBindingAndApplies(t *testing.T) {
	var applied []app.ProjectConfigureAction
	model := newProjectConfigureModel(configureTestState(), func(actions []app.ProjectConfigureAction) (string, error) {
		applied = append([]app.ProjectConfigureAction(nil), actions...)
		return "applied 1 change(s)\n", nil
	})

	model = updateConfigureModel(t, model, "enter")
	model = updateConfigureModel(t, model, "enter")
	model = updateConfigureModel(t, model, "enter")
	if len(model.pending) != 1 {
		t.Fatalf("expected pending bind, got %#v", model.pending)
	}
	if model.pending[0].Kind != app.ProjectConfigureAddPath || model.pending[0].Env != "OPENAI_API_KEY" || model.pending[0].Path != "providers/openai:api_key" {
		t.Fatalf("unexpected pending bind: %#v", model.pending[0])
	}

	model = updateConfigureModel(t, model, "j")
	model = updateConfigureModel(t, model, "j")
	model = updateConfigureModel(t, model, "j")
	model = updateConfigureModel(t, model, "enter")
	model = updateConfigureModel(t, model, "enter")
	if len(applied) != 1 {
		t.Fatalf("expected one applied action, got %#v", applied)
	}
	if !model.done {
		t.Fatalf("expected model to quit after apply")
	}
}

func TestProjectConfigureModelQueuesAddAndRemove(t *testing.T) {
	model := newProjectConfigureModel(configureTestState(), func(actions []app.ProjectConfigureAction) (string, error) {
		return "applied\n", nil
	})

	model = updateConfigureModel(t, model, "j")
	model = updateConfigureModel(t, model, "enter")
	model = updateConfigureModel(t, model, "j")
	model = updateConfigureModel(t, model, "enter")

	model = updateConfigureModel(t, model, "j")
	model = updateConfigureModel(t, model, "j")
	model = updateConfigureModel(t, model, "enter")
	model = updateConfigureModel(t, model, "enter")

	if len(model.pending) != 2 {
		t.Fatalf("expected add and remove pending actions, got %#v", model.pending)
	}
	if model.pending[0].Kind != app.ProjectConfigureAddPath || model.pending[0].Path != "app:token" || model.pending[0].Env != "" {
		t.Fatalf("unexpected add action: %#v", model.pending[0])
	}
	if model.pending[1].Kind != app.ProjectConfigureRemove || model.pending[1].Key != "old:token" {
		t.Fatalf("unexpected remove action: %#v", model.pending[1])
	}
}

func TestProjectConfigureApplyWritesManifestWithoutSecretValues(t *testing.T) {
	dir, data := setupProjectTest(t)
	manifestPath := filepath.Join(dir, ".shelf.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai:api_key", "sk-test-key"); err != nil {
		t.Fatalf("set secret: %v", err)
	}

	appSvc := testApp()
	_, err := appSvc.ApplyProjectConfigureActions(filepath.Join(dir, "shelf-test-config.yaml"), data, []app.ProjectConfigureAction{{Kind: app.ProjectConfigureAddPath, Env: "OPENAI_API_KEY", Path: "providers/openai:api_key"}})
	if err != nil {
		t.Fatalf("apply configure actions: %v", err)
	}
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	for _, want := range []string{"providers/openai:api_key", "OPENAI_API_KEY"} {
		if !strings.Contains(string(manifest), want) {
			t.Fatalf("manifest missing %q:\n%s", want, manifest)
		}
	}
	if strings.Contains(string(manifest), "sk-test-key") {
		t.Fatalf("manifest leaks secret value:\n%s", manifest)
	}
}

func TestProjectConfigureViewDoesNotLeakValues(t *testing.T) {
	model := newProjectConfigureModel(configureTestState(), func(actions []app.ProjectConfigureAction) (string, error) { return "", nil })
	out := model.View()
	for _, leaked := range []string{"sk-test-key", "OPENAI_API_KEY="} {
		if strings.Contains(out, leaked) {
			t.Fatalf("configure view leaked %q:\n%s", leaked, out)
		}
	}
}

func configureTestState() app.ProjectConfigureState {
	return app.ProjectConfigureState{
		EnvFiles:    []app.ProjectConfigureEnvFile{{Name: ".env", Vars: []app.ProjectConfigureEnvVar{{Name: "OPENAI_API_KEY", Empty: true}, {Name: "APP_TOKEN"}}}},
		SecretPaths: []string{"providers/openai:api_key", "app:token"},
		Entries:     []app.ProjectConfigureEntry{{Key: "old:token", Kind: "path", Selector: "old:token", Required: true}},
	}
}

func updateConfigureModel(t *testing.T, model projectConfigureModel, key string) projectConfigureModel {
	t.Helper()
	var msg tea.KeyMsg
	if key == "enter" {
		msg = tea.KeyMsg(tea.Key{Type: tea.KeyEnter})
	} else {
		msg = tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune(key)})
	}
	updated, _ := model.Update(msg)
	got, ok := updated.(projectConfigureModel)
	if !ok {
		t.Fatalf("unexpected model type %T", updated)
	}
	return got
}
