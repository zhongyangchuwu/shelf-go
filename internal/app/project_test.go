package app

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/project"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestProjectExportWithTempVault(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	_, err := testApp().ProjectExport(configPath, vaultPath, "shell")
	if err != nil {
		t.Fatalf("item project export: %v", err)
	}
}

func TestProjectExportMaterializesBoundFields(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	result, err := testApp().ProjectExport(configPath, vaultPath, "shell")
	if err != nil {
		t.Fatalf("item project export: %v", err)
	}
	if !strings.Contains(result.Output, "GITHUB_TOKEN=") {
		t.Fatalf("expected GITHUB_TOKEN in export, got:\n%s", result.Output)
	}
	if !strings.Contains(result.Output, "ghp_fake_token") {
		t.Fatalf("expected actual secret value in export, got:\n%s", result.Output)
	}
}

func TestProjectExportShellFormat(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	result, err := testApp().ProjectExport(configPath, vaultPath, "shell")
	if err != nil {
		t.Fatalf("item project export shell: %v", err)
	}
	if !strings.Contains(result.Output, "export GITHUB_TOKEN=") {
		t.Fatalf("expected shell export, got:\n%s", result.Output)
	}
}

func TestProjectExportRejectsRemovedEnvFormat(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	_, err := testApp().ProjectExport(configPath, vaultPath, "env")
	if err == nil || !strings.Contains(err.Error(), "unsupported format: env") {
		t.Fatalf("expected removed env format error, got: %v", err)
	}
}

func TestProjectExportJSONFormat(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	result, err := testApp().ProjectExport(configPath, vaultPath, "json")
	if err != nil {
		t.Fatalf("item project export json: %v", err)
	}
	if !strings.Contains(result.Output, `"GITHUB_TOKEN"`) {
		t.Fatalf("expected JSON export, got:\n%s", result.Output)
	}
}

func TestProjectExportFailsOnMissingItem(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	m := project.New()
	m.AddBinding(project.Binding{Item: "nonexistent", Field: "credential", Env: "NOPE"})
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	result, err := testApp().ProjectExport(configPath, vaultPath, "shell")
	if err == nil {
		t.Fatal("expected error for missing item")
	}
	if !strings.Contains(result.Diagnostics, "missing item") {
		t.Fatalf("expected value-free diagnostic, got: %s", result.Diagnostics)
	}
}

func TestProjectExportFailsOnMissingField(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	m := project.New()
	m.AddBinding(project.Binding{Item: "github", Field: "nonexistent", Env: "NOPE"})
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	result, err := testApp().ProjectExport(configPath, vaultPath, "shell")
	if err == nil {
		t.Fatal("expected error for missing field")
	}
	if !strings.Contains(result.Diagnostics, "missing field") {
		t.Fatalf("expected value-free diagnostic, got: %s", result.Diagnostics)
	}
}

func TestProjectExportOptionalMissingDoesNotFail(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	optional := false
	m := project.New()
	m.AddBinding(project.Binding{Item: "nonexistent", Field: "credential", Env: "NOPE", Required: &optional})
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	result, err := testApp().ProjectExport(configPath, vaultPath, "shell")
	if err == nil {
		t.Fatal("expected error for optional missing with no resolved entries")
	}
	if !strings.Contains(err.Error(), "no secrets to export") {
		t.Fatalf("expected no-secrets error, got: %v (diag: %s)", err, result.Diagnostics)
	}
	if !strings.Contains(result.Diagnostics, "warn") {
		t.Fatalf("expected warn diagnostic for optional missing, got: %s", result.Diagnostics)
	}
}

func TestProjectExportDuplicateEnvIsValueFreeDiagnostic(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	m := project.New()
	m.AddBinding(project.Binding{Item: "github", Field: "credential", Env: "DUPE"})
	m.AddBinding(project.Binding{Item: "openai", Field: "credential", Env: "DUPE"})
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	result, err := testApp().ProjectExport(configPath, vaultPath, "shell")
	if err == nil {
		t.Fatal("expected error for duplicate env")
	}
	if strings.Contains(result.Diagnostics, "ghp_") || strings.Contains(result.Diagnostics, "sk-") {
		t.Fatalf("diagnostics leaked secret values: %s", result.Diagnostics)
	}
}

func TestProjectStatusReportsDiagnostics(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	out, err := testApp().ProjectStatus(configPath, vaultPath, nil)
	if err != nil {
		t.Fatalf("item project status: %v", err)
	}
	if !strings.Contains(out, "github.credential") {
		t.Fatalf("expected binding in status output, got:\n%s", out)
	}
	if !strings.Contains(out, "GITHUB_TOKEN") {
		t.Fatalf("expected env name in status output, got:\n%s", out)
	}
}

func TestProjectStatusValueFree(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	out, err := testApp().ProjectStatus(configPath, vaultPath, nil)
	if err != nil {
		t.Fatalf("item project status: %v", err)
	}
	if strings.Contains(out, "ghp_fake_token") || strings.Contains(out, "sk-") {
		t.Fatalf("status output leaked secret values:\n%s", out)
	}
}

func TestProjectWorkspaceIsValueFreeAndProvidesCandidates(t *testing.T) {
	_, configPath := setupProjectTest(t)
	appSvc := testApp()

	workspace, err := appSvc.ProjectWorkspace(configPath, "")
	if err != nil {
		t.Fatalf("ProjectWorkspace: %v", err)
	}
	if !workspace.ManifestPresent || len(workspace.Bindings) != 1 {
		t.Fatalf("workspace = %#v", workspace)
	}
	binding := workspace.Bindings[0]
	if binding.Key != "github.credential" || binding.Env != "GITHUB_TOKEN" || binding.Status != "ready" || !binding.Required {
		t.Fatalf("binding = %#v", binding)
	}
	candidates, err := appSvc.ProjectBindingCandidates(configPath, "")
	if err != nil {
		t.Fatalf("ProjectBindingCandidates: %v", err)
	}
	if len(candidates) != 2 || candidates[0].Field != "credential" || candidates[0].Env == "" {
		t.Fatalf("candidates = %#v", candidates)
	}
	encoded, err := json.Marshal(struct {
		Workspace  ProjectWorkspace
		Candidates []ProjectBindingCandidate
	}{workspace, candidates})
	if err != nil {
		t.Fatalf("marshal workspace: %v", err)
	}
	if strings.Contains(string(encoded), "ghp_fake_token") || strings.Contains(string(encoded), "sk-openai-test") {
		t.Fatalf("workspace DTO leaked secret values: %s", encoded)
	}
}

func TestProjectWorkspaceReportsDuplicateEnvWithoutValues(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	manifest := project.New()
	if err := manifest.AddBinding(project.Binding{Item: "github", Field: "credential", Env: "DUPLICATE"}); err != nil {
		t.Fatalf("add github binding: %v", err)
	}
	if err := manifest.AddBinding(project.Binding{Item: "openai", Field: "credential", Env: "DUPLICATE", Required: new(false)}); err != nil {
		t.Fatalf("add openai binding: %v", err)
	}
	if err := project.Save(filepath.Join(dir, project.FileName), manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	workspace, err := testApp().ProjectWorkspace(configPath, "")
	if err != nil {
		t.Fatalf("ProjectWorkspace: %v", err)
	}
	if got := workspace.Bindings[1]; got.Status != "fail" || !strings.Contains(got.Message, "conflicts") || strings.Contains(got.Message, "sk-openai-test") {
		t.Fatalf("duplicate status = %#v", got)
	}
}

func TestProjectWorkspaceHandlesMissingManifest(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	if err := os.Remove(filepath.Join(dir, project.FileName)); err != nil {
		t.Fatalf("remove manifest: %v", err)
	}
	workspace, err := testApp().ProjectWorkspace(configPath, "")
	if err != nil {
		t.Fatalf("ProjectWorkspace: %v", err)
	}
	if workspace.ManifestPresent || len(workspace.Bindings) != 0 {
		t.Fatalf("missing-manifest workspace = %#v", workspace)
	}
}

func TestProjectRunInjectsSecret(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	var stdout strings.Builder
	var stderr strings.Builder
	err := testApp().ProjectRun(ProjectRunRequest{
		ConfigPath: configPath,
		VaultPath:  vaultPath,
		Command:    []string{"sh", "-c", `printf '%s' "$GITHUB_TOKEN"`},
		Stdin:      nil,
		Stdout:     &stdout,
		Stderr:     &stderr,
	})
	if err != nil {
		t.Fatalf("item project run: %v (stderr: %s)", err, stderr.String())
	}
	if stdout.String() != "ghp_fake_token" {
		t.Fatalf("expected 'ghp_fake_token', got %q", stdout.String())
	}
}

func TestProjectRunReturnsChildExitCode(t *testing.T) {
	dir, configPath := setupProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.json")

	var stdout, stderr strings.Builder
	err := testApp().ProjectRun(ProjectRunRequest{
		ConfigPath: configPath,
		VaultPath:  vaultPath,
		Command:    []string{"sh", "-c", "exit 7"},
		Stdin:      nil,
		Stdout:     &stdout,
		Stderr:     &stderr,
	})
	if err == nil {
		t.Fatal("expected child exit error")
	}
	if !strings.Contains(err.Error(), "status 7") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectAddAndSecretGetAcceptTheSameTitleSelector(t *testing.T) {
	_, configPath := setupProjectTest(t)

	value, err := testApp().GetSecretField(configPath, "", GetSecretFieldRequest{Selector: "OpenAI API Key"})
	if err != nil {
		t.Fatalf("get by title: %v", err)
	}
	if value == "" {
		t.Fatal("get by title returned an empty value")
	}

	out, err := testApp().ProjectAdd(configPath, "", ProjectAddRequest{Selector: "OpenAI API Key"})
	if err != nil {
		t.Fatalf("project add by title: %v", err)
	}
	if out != "added openai.credential\n" {
		t.Fatalf("project add output = %q", out)
	}
}

// --- helpers ---

// setupProjectTest creates a temp dir with a item .shelf.json manifest
// and a item vault JSON. Returns (dir, configPath).
func setupProjectTest(t *testing.T) (dir, configPath string) {
	t.Helper()
	dir = t.TempDir()

	// Chdir to the temp dir so project.Root() resolves correctly
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Git init so loadProjectManifest (via project.Root()) works
	if _, err := exec.LookPath("git"); err == nil {
		if out, err := exec.Command("git", "init").CombinedOutput(); err != nil {
			t.Fatalf("git init: %v\n%s", err, out)
		}
	}
	vaultPath := filepath.Join(dir, "vault.json")
	// Write a minimal YAML config pointing at the vault path
	configPath = filepath.Join(dir, "shelf-test-config.yaml")
	configContent := "vault_path: " + vaultPath + "\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Create item vault
	vault := vaultmodel.Vault{
		Version:   vaultmodel.Version,
		Templates: vaultmodel.BuiltInTemplates(),
		Profiles:  vaultmodel.BuiltInProfiles(),
		Items: []vaultmodel.Item{
			{
				ID:       "github",
				Title:    "GitHub Token",
				Template: vaultmodel.TemplateAPICredential,
				Fields: []vaultmodel.Field{
					{ID: "credential", Label: "Credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Value: json.RawMessage(`"ghp_fake_token"`)},
					{ID: "credential_env", Label: "Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeEnvName, Value: json.RawMessage(`"GITHUB_TOKEN"`)},
				},
			},
			{
				ID:       "openai",
				Title:    "OpenAI API Key",
				Template: vaultmodel.TemplateAPICredential,
				Fields: []vaultmodel.Field{
					{ID: "credential", Label: "Credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Value: json.RawMessage(`"sk-openai-test"`)},
					{ID: "credential_env", Label: "Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeEnvName, Value: json.RawMessage(`"OPENAI_API_KEY"`)},
				},
			},
		},
	}
	repo := vaultjson.NewRepository(vaultPath)
	if err := repo.Save(vault); err != nil {
		t.Fatalf("save vault: %v", err)
	}

	// Create item manifest
	m := project.New()
	if err := m.AddBinding(project.Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	return dir, configPath
}
