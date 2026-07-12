package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

type recordingClipboard struct {
	text string
	err  error
}

func (c *recordingClipboard) CopyText(text string) error {
	c.text = text
	return c.err
}

func runShelfWithClipboard(t *testing.T, copier *recordingClipboard, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCmd(testApp(), copier)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestRenderSecretCreateResultUsesPrimaryEnv(t *testing.T) {
	var out bytes.Buffer
	renderSecretCreateResult(&out, app.SecretInfo{ID: "item_service", Title: "service", Template: "service_token", PrimaryEnv: "SERVICE_TOKEN"})
	if !strings.Contains(out.String(), "env     : SERVICE_TOKEN") {
		t.Fatalf("create summary = %q, want role-derived env", out.String())
	}
}

func TestSecretCreateSetGetListInfoRm(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.json")
	out, err := runShelfWithInput(t, "fake-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin")
	if err != nil {
		t.Fatalf("create: %v\n%s", err, out)
	}
	for _, want := range []string{"created secret", "title   : github", "id      : item_github", "template: generic_secret", "env     : GITHUB_TOKEN"} {
		if !strings.Contains(out, want) {
			t.Fatalf("create output missing %q:\n%s", want, out)
		}
	}

	out, err = runShelf(t, "--vault", data, "secret", "get", "github")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if out != "fake-token\n" {
		t.Fatalf("get = %q", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "list", "--long")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, want := range []string{"Secrets (1):", "ID", "TITLE", "TEMPLATE", "ENV", "VALUE", "item_github", "github", "generic_secret", "GITHUB_TOKEN", "set"} {
		if !strings.Contains(out, want) {
			t.Fatalf("list output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "fake-token") {
		t.Fatalf("list leaked value: %q", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "info", "github")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	for _, want := range []string{"\"id\": \"item_github\"", "\"id\": \"value_env\"", "GITHUB_TOKEN", "\"value_set\": true"} {
		if !strings.Contains(out, want) {
			t.Fatalf("info missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "fake-token") {
		t.Fatalf("info leaked value: %s", out)
	}

	if _, err := runShelf(t, "--vault", data, "secret", "rm", "github"); err == nil {
		t.Fatalf("rm without --yes should fail")
	}
	if _, err := runShelf(t, "--vault", data, "secret", "rm", "github", "--yes"); err != nil {
		t.Fatalf("rm --yes: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "get", "github"); err == nil {
		t.Fatalf("get deleted item should fail")
	}
}

func TestSecretCopyCopiesWithoutPrintingValue(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.json")
	if _, err := runShelfWithInput(t, "fake-copy-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin"); err != nil {
		t.Fatalf("create: %v", err)
	}
	copier := &recordingClipboard{}
	out, err := runShelfWithClipboard(t, copier, "--vault", data, "secret", "copy", "--env", "GITHUB_TOKEN")
	if err != nil {
		t.Fatalf("copy: %v\n%s", err, out)
	}
	if out != "Copied secret to clipboard.\n" {
		t.Fatalf("copy output = %q", out)
	}
	if copier.text != "fake-copy-token" {
		t.Fatalf("clipboard text = %q", copier.text)
	}
}

func TestSecretCopyClipboardFailureDoesNotPrintValue(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.json")
	if _, err := runShelfWithInput(t, "fake-copy-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin"); err != nil {
		t.Fatalf("create: %v", err)
	}
	copier := &recordingClipboard{err: errors.New("clipboard unavailable")}
	out, err := runShelfWithClipboard(t, copier, "--vault", data, "secret", "copy", "github")
	if err == nil || !strings.Contains(err.Error(), "copy secret to clipboard") {
		t.Fatalf("expected copy error, got: %v", err)
	}
	if strings.Contains(out, "fake-copy-token") || strings.Contains(err.Error(), "fake-copy-token") {
		t.Fatalf("copy failure leaked secret: out=%q err=%v", out, err)
	}
}

func TestSecretCreateProfileOpenAI(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.json")
	out, err := runShelfWithInput(t, "fake-openai-key", "--vault", data, "secret", "create", "openai", "--profile", "openai", "--stdin")
	if err != nil {
		t.Fatalf("create profile: %v\n%s", err, out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "info", "openai")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	for _, want := range []string{"api_credential", "credential_env", "OPENAI_API_KEY", "provider", "OpenAI", "website", "https://platform.openai.com"} {
		if !strings.Contains(out, want) {
			t.Fatalf("info missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "fake-openai-key") {
		t.Fatalf("info leaked value: %s", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "get", "--env", "OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("get --env: %v", err)
	}
	if out != "fake-openai-key\n" {
		t.Fatalf("get --env = %q", out)
	}
}

func TestSecretUpdateReadsPipedInput(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.json")
	if _, err := runShelf(t, "--vault", data, "secret", "create", "app", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if out, err := runShelfWithInput(t, "updated-token", "--vault", data, "secret", "update", "app"); err != nil || out != "Updated secret item.\n" {
		t.Fatalf("update primary: out=%q err=%v", out, err)
	}
	if _, err := runShelfWithInput(t, "APP_TOKEN_RENAMED", "--vault", data, "secret", "update", "app", "--field", "value_env"); err != nil {
		t.Fatalf("update public field: %v", err)
	}
	if out, err := runShelf(t, "--vault", data, "secret", "update", "app", "--add-label", "local"); err != nil || out != "Updated secret item.\n" {
		t.Fatalf("metadata-only update: out=%q err=%v", out, err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "--env", "APP_TOKEN_RENAMED")
	if err != nil {
		t.Fatalf("get env renamed: %v", err)
	}
	if out != "updated-token\n" {
		t.Fatalf("get = %q", out)
	}
}

func TestSecretUpdateMaintainsValueFreeOutput(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.json")
	if _, err := runShelfWithInput(t, "fake-openai-token", "--vault", data, "secret", "create", "openai", "--profile", "openai", "--stdin"); err != nil {
		t.Fatalf("create: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "update", "openai",
		"--title", "OpenAI Personal",
		"--add-label", "ai",
		"--add-label", "personal",
		"--add-alias", "openai-personal",
		"--field", "provider",
		"--value", "OpenAI personal account",
	)
	if err != nil {
		t.Fatalf("update: %v\n%s", err, out)
	}
	if out != "Updated secret item.\n" {
		t.Fatalf("update output = %q", out)
	}
	if strings.Contains(out, "fake-openai-token") {
		t.Fatalf("update leaked secret value: %q", out)
	}
	out, err = runShelf(t, "--vault", data, "secret", "info", "openai-personal")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	for _, want := range []string{"OpenAI Personal", "ai", "personal", "OpenAI personal account"} {
		if !strings.Contains(out, want) {
			t.Fatalf("info missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "fake-openai-token") {
		t.Fatalf("info leaked updated secret value: %q", out)
	}
	_, err = runShelf(t, "--vault", data, "secret", "update", "openai-personal", "--value", "attempted-secret")
	if err == nil || !strings.Contains(err.Error(), "requires --field") {
		t.Fatalf("missing field error = %v", err)
	}
	if strings.Contains(err.Error(), "attempted-secret") {
		t.Fatalf("update error leaked value: %v", err)
	}
}

func TestSecretEnvSelectorAmbiguityIsValueFree(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.json")
	if _, err := runShelfWithInput(t, "first-secret", "--vault", data, "secret", "create", "one", "--env", "DUPLICATE_ENV", "--stdin"); err != nil {
		t.Fatalf("create one: %v", err)
	}
	if _, err := runShelfWithInput(t, "second-secret", "--vault", data, "secret", "create", "two", "--env", "DUPLICATE_ENV", "--stdin"); err != nil {
		t.Fatalf("create two: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "--env", "DUPLICATE_ENV")
	if err == nil {
		t.Fatalf("expected ambiguous env error")
	}
	if !strings.Contains(err.Error(), "env selector matched multiple fields") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out+err.Error(), "first-secret") || strings.Contains(out+err.Error(), "second-secret") {
		t.Fatalf("ambiguity leaked secret values: out=%q err=%v", out, err)
	}
}

func TestSecretLegacyMutatingCommandsRemoved(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.json")
	for _, subcommand := range []string{"add", "edit", "set"} {
		_, err := runShelf(t, "--vault", data, "secret", subcommand)
		if err == nil {
			t.Fatalf("secret %s should be removed", subcommand)
		}
	}
}

func TestSecretExport_ByLabelShell(t *testing.T) {
	data := writeLabeledVault(t, map[string]labeledItem{
		"db":    {env: "DATABASE_URL", value: "postgres://db", labels: []string{"db", "prod"}},
		"local": {env: "LOCAL_VAR", value: "local-val", labels: []string{"local"}},
	})
	out, err := runShelf(t, "--vault", data, "secret", "export", "--label", "prod", "--format", "shell")
	if err != nil {
		t.Fatalf("export: %v\n%s", err, out)
	}
	if !strings.Contains(out, "export DATABASE_URL=postgres://db") {
		t.Fatalf("missing export DATABASE_URL in:\n%s", out)
	}
	if strings.Contains(out, "LOCAL_VAR") {
		t.Fatalf("local item leaked into prod export:\n%s", out)
	}
}

func TestSecretExportRejectsRemovedEnvFormat(t *testing.T) {
	data := writeLabeledVault(t, map[string]labeledItem{
		"api": {env: "API_KEY", value: "sk-abc123", labels: []string{"env", "test"}},
	})
	out, err := runShelf(t, "--vault", data, "secret", "export", "--label", "env", "--format", "env")
	if err == nil || !strings.Contains(err.Error(), "unsupported export format: env") {
		t.Fatalf("expected removed env format error, got: %v\n%s", err, out)
	}
	if strings.Contains(out, "sk-abc123") || strings.Contains(err.Error(), "sk-abc123") {
		t.Fatalf("env format error leaked secret: out=%q err=%v", out, err)
	}
}

func TestSecretExport_ByLabelJSON(t *testing.T) {
	data := writeLabeledVault(t, map[string]labeledItem{
		"a": {env: "ALPHA", value: "one", labels: []string{"all"}},
		"b": {env: "BETA", value: "two", labels: []string{"all"}},
	})
	out, err := runShelf(t, "--vault", data, "secret", "export", "--label", "all", "--format", "json")
	if err != nil {
		t.Fatalf("export json: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"ALPHA": "one"`) {
		t.Fatalf("missing ALPHA in json:\n%s", out)
	}
	if !strings.Contains(out, `"BETA": "two"`) {
		t.Fatalf("missing BETA in json:\n%s", out)
	}
}

func TestSecretExport_MultipleLabelsAND(t *testing.T) {
	data := writeLabeledVault(t, map[string]labeledItem{
		"shared":     {env: "SHARED_VAR", value: "shared-val", labels: []string{"prod", "us-east"}},
		"onlyprod":   {env: "ONLY_PROD", value: "prod-val", labels: []string{"prod"}},
		"onlyregion": {env: "ONLY_REGION", value: "region-val", labels: []string{"us-east"}},
	})
	out, err := runShelf(t, "--vault", data, "secret", "export", "--label", "prod", "--label", "us-east", "--format", "shell")
	if err != nil {
		t.Fatalf("export AND: %v\n%s", err, out)
	}
	if !strings.Contains(out, "SHARED_VAR=shared-val") {
		t.Fatalf("missing SHARED_VAR:\n%s", out)
	}
	if strings.Contains(out, "ONLY_PROD") {
		t.Fatalf("ONLY_PROD should not match AND query:\n%s", out)
	}
	if strings.Contains(out, "ONLY_REGION") {
		t.Fatalf("ONLY_REGION should not match AND query:\n%s", out)
	}
}

func TestSecretExport_NoMatchingLabels(t *testing.T) {
	data := writeLabeledVault(t, map[string]labeledItem{
		"x": {env: "X_VAR", value: "x-val", labels: []string{"alpha"}},
	})
	_, err := runShelf(t, "--vault", data, "secret", "export", "--label", "nonexistent")
	if err == nil {
		t.Fatal("expected error for no matching labels, got nil")
	}
	if !strings.Contains(err.Error(), "no items matched labels") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSecretExport_NoLabelsFlag(t *testing.T) {
	data := writeLabeledVault(t, nil)
	_, err := runShelf(t, "--vault", data, "secret", "export")
	if err == nil {
		t.Fatal("expected error for missing --label, got nil")
	}
}

func TestSecretExport_DuplicateEnvNameIsValueFree(t *testing.T) {
	data := writeLabeledVault(t, map[string]labeledItem{
		"first":  {env: "DUPE", value: "secret-one", labels: []string{"group"}},
		"second": {env: "DUPE", value: "secret-two", labels: []string{"group"}},
	})
	_, err := runShelf(t, "--vault", data, "secret", "export", "--label", "group")
	if err == nil {
		t.Fatal("expected duplicate env error, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate env name") {
		t.Fatalf("expected duplicate env error, got: %v", err)
	}
	if strings.Contains(err.Error(), "secret-one") || strings.Contains(err.Error(), "secret-two") {
		t.Fatalf("error leaked secret values: %v", err)
	}
}

// labeledItem describes a secret item for test vault creation.
type labeledItem struct {
	env    string
	value  string
	labels []string
}

// writeLabeledVault creates a item vault JSON file with labeled generic_secret items.
func writeLabeledVault(t *testing.T, items map[string]labeledItem) string {
	t.Helper()
	data := filepath.Join(t.TempDir(), "vault.json")
	repo := vaultjson.NewRepository(data)
	vault := vaultmodel.Vault{Version: vaultmodel.CurrentVersion}
	// Sort titles for deterministic output in vault.
	titles := make([]string, 0, len(items))
	for title := range items {
		titles = append(titles, title)
	}
	sort.Strings(titles)
	for _, title := range titles {
		it := items[title]
		fields := []vaultmodel.Field{
			{ID: "value", Label: "Value", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Value: rawJSON(it.value)},
		}
		if it.env != "" {
			fields = append(fields, vaultmodel.Field{ID: "value_env", Label: "Value Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeEnvName, Value: rawJSON(it.env)})
		}
		vault.Items = append(vault.Items, vaultmodel.Item{
			ID:       "item_" + title,
			Title:    title,
			Template: vaultmodel.TemplateGenericSecret,
			Labels:   it.labels,
			Fields:   fields,
		})
	}
	if err := repo.Save(vault); err != nil {
		t.Fatalf("write vault: %v", err)
	}
	return data
}

func rawJSON(value string) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}

func TestSecretAuditReportsValueFreeHealthAndExitCodes(t *testing.T) {
	expiredVault := filepath.Join(t.TempDir(), "expired-vault.json")
	if _, err := runShelfWithInput(t, "fake-expired-token", "--vault", expiredVault, "secret", "create", "expired", "--template", "api", "--expires", "2000-01-01", "--stdin"); err != nil {
		t.Fatalf("create expired credential: %v", err)
	}
	out, err := runShelf(t, "--vault", expiredVault, "secret", "audit", "--format", "json")
	if err == nil || ExitCode(err) != 3 {
		t.Fatalf("expired audit error = %v, exit code = %d", err, ExitCode(err))
	}
	if !strings.Contains(out, `"status":"expired"`) || strings.Contains(out, "fake-expired-token") {
		t.Fatalf("expired audit output = %q", out)
	}

	missingVault := filepath.Join(t.TempDir(), "missing-vault.json")
	if _, err := runShelfWithInput(t, "fake-missing-token", "--vault", missingVault, "secret", "create", "missing", "--template", "api", "--stdin"); err != nil {
		t.Fatalf("create missing credential: %v", err)
	}
	out, err = runShelf(t, "--vault", missingVault, "secret", "audit")
	if err == nil || ExitCode(err) != 2 {
		t.Fatalf("missing audit error = %v, exit code = %d", err, ExitCode(err))
	}
	if !strings.Contains(out, "missing") || strings.Contains(out, "fake-missing-token") {
		t.Fatalf("missing audit output = %q", out)
	}

	healthyVault := filepath.Join(t.TempDir(), "healthy-vault.json")
	if _, err := runShelfWithInput(t, "fake-healthy-token", "--vault", healthyVault, "secret", "create", "healthy", "--template", "api", "--expires", "2099-01-01", "--stdin"); err != nil {
		t.Fatalf("create healthy credential: %v", err)
	}
	out, err = runShelf(t, "--vault", healthyVault, "secret", "audit")
	if err != nil {
		t.Fatalf("healthy audit error = %v", err)
	}
	if !strings.Contains(out, "healthy") || strings.Contains(out, "fake-healthy-token") {
		t.Fatalf("healthy audit output = %q", out)
	}
}

func TestSecretAuditRejectsNegativeHorizon(t *testing.T) {
	_, err := runShelf(t, "--vault", filepath.Join(t.TempDir(), "vault.json"), "secret", "audit", "--within", "-1")
	if err == nil || !strings.Contains(err.Error(), "horizon") {
		t.Fatalf("negative audit horizon error = %v", err)
	}
}
