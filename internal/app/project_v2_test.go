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

func TestProjectExportV2WithTempV2Vault(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	_, err := testApp().ProjectExport(configPath, vaultPath, "env")
	if err != nil {
		t.Fatalf("v2 project export: %v", err)
	}
}

func TestProjectExportV2MaterializesBoundFields(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	result, err := testApp().ProjectExport(configPath, vaultPath, "env")
	if err != nil {
		t.Fatalf("v2 project export: %v", err)
	}
	if !strings.Contains(result.Output, "GITHUB_TOKEN=") {
		t.Fatalf("expected GITHUB_TOKEN in export, got:\n%s", result.Output)
	}
	if !strings.Contains(result.Output, "ghp_fake_token") {
		t.Fatalf("expected actual secret value in export, got:\n%s", result.Output)
	}
}

func TestProjectExportV2ShellFormat(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	result, err := testApp().ProjectExport(configPath, vaultPath, "shell")
	if err != nil {
		t.Fatalf("v2 project export shell: %v", err)
	}
	if !strings.Contains(result.Output, "export GITHUB_TOKEN=") {
		t.Fatalf("expected shell export, got:\n%s", result.Output)
	}
}

func TestProjectExportV2JSONFormat(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	result, err := testApp().ProjectExport(configPath, vaultPath, "json")
	if err != nil {
		t.Fatalf("v2 project export json: %v", err)
	}
	if !strings.Contains(result.Output, `"GITHUB_TOKEN"`) {
		t.Fatalf("expected JSON export, got:\n%s", result.Output)
	}
}

func TestProjectExportV2FailsOnMissingItem(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	m := project.NewV2()
	m.AddV2Binding(project.V2Binding{Item: "nonexistent", Field: "credential", Env: "NOPE"})
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	result, err := testApp().ProjectExport(configPath, vaultPath, "env")
	if err == nil {
		t.Fatal("expected error for missing item")
	}
	if !strings.Contains(result.Diagnostics, "missing item") {
		t.Fatalf("expected value-free diagnostic, got: %s", result.Diagnostics)
	}
}

func TestProjectExportV2FailsOnMissingField(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	m := project.NewV2()
	m.AddV2Binding(project.V2Binding{Item: "github", Field: "nonexistent", Env: "NOPE"})
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	result, err := testApp().ProjectExport(configPath, vaultPath, "env")
	if err == nil {
		t.Fatal("expected error for missing field")
	}
	if !strings.Contains(result.Diagnostics, "missing field") {
		t.Fatalf("expected value-free diagnostic, got: %s", result.Diagnostics)
	}
}

func TestProjectExportV2OptionalMissingDoesNotFail(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	optional := false
	m := project.NewV2()
	m.AddV2Binding(project.V2Binding{Item: "nonexistent", Field: "credential", Env: "NOPE", Required: &optional})
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	result, err := testApp().ProjectExport(configPath, vaultPath, "env")
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

func TestProjectExportV2DuplicateEnvIsValueFreeDiagnostic(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	m := project.NewV2()
	m.AddV2Binding(project.V2Binding{Item: "github", Field: "credential", Env: "DUPE"})
	m.AddV2Binding(project.V2Binding{Item: "openai", Field: "credential", Env: "DUPE"})
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	result, err := testApp().ProjectExport(configPath, vaultPath, "env")
	if err == nil {
		t.Fatal("expected error for duplicate env")
	}
	if strings.Contains(result.Diagnostics, "ghp_") || strings.Contains(result.Diagnostics, "sk-") {
		t.Fatalf("diagnostics leaked secret values: %s", result.Diagnostics)
	}
}

func TestProjectStatusV2ReportsDiagnostics(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	out, err := testApp().ProjectStatus(configPath, vaultPath, nil)
	if err != nil {
		t.Fatalf("v2 project status: %v", err)
	}
	if !strings.Contains(out, "github.credential") {
		t.Fatalf("expected binding in status output, got:\n%s", out)
	}
	if !strings.Contains(out, "GITHUB_TOKEN") {
		t.Fatalf("expected env name in status output, got:\n%s", out)
	}
}

func TestProjectStatusV2ValueFree(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

	out, err := testApp().ProjectStatus(configPath, vaultPath, nil)
	if err != nil {
		t.Fatalf("v2 project status: %v", err)
	}
	if strings.Contains(out, "ghp_fake_token") || strings.Contains(out, "sk-") {
		t.Fatalf("status output leaked secret values:\n%s", out)
	}
}

func TestProjectRunV2InjectsSecret(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

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
		t.Fatalf("v2 project run: %v (stderr: %s)", err, stderr.String())
	}
	if stdout.String() != "ghp_fake_token" {
		t.Fatalf("expected 'ghp_fake_token', got %q", stdout.String())
	}
}

func TestProjectRunV2ReturnsChildExitCode(t *testing.T) {
	dir, configPath := setupV2ProjectTest(t)
	vaultPath := filepath.Join(dir, "vault.v2.json")

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

// --- helpers ---

// setupV2ProjectTest creates a temp dir with a v2 .shelf.json manifest
// and a v2 vault JSON. Returns (dir, configPath).
func setupV2ProjectTest(t *testing.T) (dir, configPath string) {
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
	vaultPath := filepath.Join(dir, "vault.v2.json")
	// Write a minimal YAML config pointing at the vault path
	configPath = filepath.Join(dir, "shelf-test-config.yaml")
	configContent := "vault_path: " + vaultPath + "\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Create v2 vault
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

	// Create v2 manifest
	m := project.NewV2()
	if err := m.AddV2Binding(project.V2Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	manifestPath := filepath.Join(dir, ".shelf.json")
	if err := project.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	return dir, configPath
}
