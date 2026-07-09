package cli

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestSecretCreateSetGetListInfoRmV2(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
	out, err := runShelfWithInput(t, "fake-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin")
	if err != nil {
		t.Fatalf("create: %v\n%s", err, out)
	}
	if !strings.Contains(out, "created item_github") {
		t.Fatalf("create output = %q", out)
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
	if !strings.Contains(out, "item_github\tgithub\tgeneric_secret\tGITHUB_TOKEN\tset") {
		t.Fatalf("list output = %q", out)
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

func TestSecretCreateProfileOpenAIV2(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
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

func TestSecretSetRequiresStdinAndCanSetPublicField(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
	if _, err := runShelf(t, "--vault", data, "secret", "create", "app", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app"); err == nil {
		t.Fatalf("set without --stdin should fail")
	}
	if _, err := runShelfWithInput(t, "updated-token", "--vault", data, "secret", "set", "app", "--stdin"); err != nil {
		t.Fatalf("set primary: %v", err)
	}
	if _, err := runShelfWithInput(t, "APP_TOKEN_RENAMED", "--vault", data, "secret", "set", "app", "--field", "value_env", "--stdin"); err != nil {
		t.Fatalf("set env field: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "--env", "APP_TOKEN_RENAMED")
	if err != nil {
		t.Fatalf("get env renamed: %v", err)
	}
	if out != "updated-token\n" {
		t.Fatalf("get = %q", out)
	}
}

func TestSecretEnvSelectorAmbiguityIsValueFree(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
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

func TestSecretLegacyMutatingCommandsRemovedV2(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
	for _, subcommand := range []string{"add", "edit"} {
		_, err := runShelf(t, "--vault", data, "secret", subcommand)
		if err == nil {
			t.Fatalf("secret %s should be removed in v2", subcommand)
		}
	}
}

func TestSecretExportV2_ByLabelShell(t *testing.T) {
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

func TestSecretExportV2_ByLabelEnv(t *testing.T) {
	data := writeLabeledVault(t, map[string]labeledItem{
		"api": {env: "API_KEY", value: "sk-abc123", labels: []string{"env", "test"}},
	})
	out, err := runShelf(t, "--vault", data, "secret", "export", "--label", "env", "--format", "env")
	if err != nil {
		t.Fatalf("export env: %v\n%s", err, out)
	}
	if !strings.Contains(out, "API_KEY=sk-abc123") {
		t.Fatalf("missing API_KEY in env output:\n%s", out)
	}
}

func TestSecretExportV2_ByLabelJSON(t *testing.T) {
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

func TestSecretExportV2_MultipleLabelsAND(t *testing.T) {
	data := writeLabeledVault(t, map[string]labeledItem{
		"shared":     {env: "SHARED_VAR", value: "shared-val", labels: []string{"prod", "us-east"}},
		"onlyprod":   {env: "ONLY_PROD", value: "prod-val", labels: []string{"prod"}},
		"onlyregion": {env: "ONLY_REGION", value: "region-val", labels: []string{"us-east"}},
	})
	out, err := runShelf(t, "--vault", data, "secret", "export", "--label", "prod", "--label", "us-east", "--format", "env")
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

func TestSecretExportV2_NoMatchingLabels(t *testing.T) {
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

func TestSecretExportV2_NoLabelsFlag(t *testing.T) {
	data := writeLabeledVault(t, nil)
	_, err := runShelf(t, "--vault", data, "secret", "export")
	if err == nil {
		t.Fatal("expected error for missing --label, got nil")
	}
}

func TestSecretExportV2_DuplicateEnvNameIsValueFree(t *testing.T) {
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

// labeledItem describes a v2 secret item for test vault creation.
type labeledItem struct {
	env    string
	value  string
	labels []string
}

// writeLabeledVault creates a v2 vault JSON file with labeled generic_secret items.
func writeLabeledVault(t *testing.T, items map[string]labeledItem) string {
	t.Helper()
	data := filepath.Join(t.TempDir(), "vault.v2.json")
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
