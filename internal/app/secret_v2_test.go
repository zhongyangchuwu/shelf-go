package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

// writeV2Config writes a minimal YAML config pointing vault_path at jsonPath
// and returns the config file path.
func writeV2Config(t *testing.T, jsonPath string) string {
	t.Helper()
	dir := filepath.Dir(jsonPath)
	configPath := filepath.Join(dir, "shelf-test-config.yaml")
	content := fmt.Sprintf("vault_path: %s\n", jsonPath)
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write v2 config: %v", err)
	}
	return configPath
}

// writeV2EmptyVault writes a minimal valid v2 vault JSON file.
func writeV2EmptyVault(t *testing.T, jsonPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
		t.Fatalf("mkdir vault dir: %v", err)
	}
	repo := vaultjson.NewRepository(jsonPath)
	if err := repo.Save(vaultmodel.Vault{Version: vaultmodel.CurrentVersion}); err != nil {
		t.Fatalf("write empty vault: %v", err)
	}
}

func TestSecretUsesConfiguredVaultPath(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeV2Config(t, vaultPath)

	if _, err := testApp().CreateSecretV2(configPath, "", CreateSecretRequest{Title: "default-path", Stdin: true, In: strings.NewReader("fake-default-secret")}); err != nil {
		t.Fatalf("CreateSecretV2: %v", err)
	}
	if _, err := os.Stat(vaultPath); err != nil {
		t.Fatalf("vault path missing: %v", err)
	}
	value, err := testApp().GetSecretFieldV2(configPath, "", GetSecretFieldRequest{Selector: "default-path"})
	if err != nil {
		t.Fatalf("GetSecretFieldV2: %v", err)
	}
	if value != "fake-default-secret" {
		t.Fatalf("value = %q, want fake-default-secret", value)
	}
}

func TestCreateSecretV2(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	// Create a basic secret with stdin value
	req := CreateSecretRequest{
		Title: "my-token",
		Env:   "MY_TOKEN",
		Stdin: true,
		In:    strings.NewReader("sk-fake-value"),
	}
	id, err := testApp().CreateSecretV2(configPath, vaultPath, req)
	if err != nil {
		t.Fatalf("CreateSecretV2: %v", err)
	}
	if id != "item_my_token" {
		t.Fatalf("id = %q, want item_my_token", id)
	}
	// Duplicate title should fail (ID uniqueness is checked first since same ID is derived)
	_, err = testApp().CreateSecretV2(configPath, vaultPath, req)
	if err == nil {
		t.Fatal("expected error on duplicate, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already-exists error, got: %v", err)
	}
	_, err = testApp().CreateSecretV2(configPath, vaultPath, CreateSecretRequest{
		Title: "  ",
		Stdin: true,
		In:    strings.NewReader("x"),
	})
	if err == nil || !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("expected title required error, got: %v", err)
	}
}

func TestCreateSecretV2ProfileOpenAI(t *testing.T) {
	// Equivalent to `--profile openai --stdin` in the CLI
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	req := CreateSecretRequest{
		Title:   "openai-key",
		Profile: "openai",
		Stdin:   true,
		In:      strings.NewReader("sk-fake-openai-key"),
	}
	id, err := testApp().CreateSecretV2(configPath, vaultPath, req)
	if err != nil {
		t.Fatalf("CreateSecretV2 with openai profile: %v", err)
	}
	if id != "item_openai_key" {
		t.Fatalf("id = %q, want item_openai_key", id)
	}

	// Verify the item was created with openai template defaults
	info, err := testApp().SecretInfoV2(configPath, vaultPath, "openai-key")
	if err != nil {
		t.Fatalf("SecretInfoV2: %v", err)
	}
	if info.Template != "api_credential" {
		t.Fatalf("template = %q, want api_credential", info.Template)
	}
	// provider field should be publicly visible
	foundProvider := false
	for _, f := range info.Fields {
		if f.ID == "provider" && f.Value == "OpenAI" {
			foundProvider = true
		}
	}
	if !foundProvider {
		t.Fatalf("expected provider field with value OpenAI, got fields: %#v", info.Fields)
	}
	// credential value must NOT be visible in info
	for _, f := range info.Fields {
		if f.ID == "credential" && f.Value != "" {
			t.Fatalf("info leaked credential value: %s", f.Value)
		}
	}
}

func TestCreateSecretV2RejectsUnsupportedMetadataFlags(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	_, err := testApp().CreateSecretV2(configPath, vaultPath, CreateSecretRequest{
		Title:    "generic-with-provider",
		Provider: "Example",
		Stdin:    true,
		In:       strings.NewReader("fake-secret"),
	})
	if err == nil || !strings.Contains(err.Error(), "--provider is not supported by template generic_secret") {
		t.Fatalf("expected unsupported provider error, got: %v", err)
	}
}

func TestCreateSecretV2AcceptsAPIMetadataFlags(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	_, err := testApp().CreateSecretV2(configPath, vaultPath, CreateSecretRequest{
		Title:    "api-token",
		Template: vaultmodel.TemplateAPICredential,
		Provider: "Example",
		Website:  "https://example.test",
		Expires:  "2026-12-31",
		Scopes:   []string{"read", "write"},
		Stdin:    true,
		In:       strings.NewReader("fake-secret"),
	})
	if err != nil {
		t.Fatalf("CreateSecretV2 with api metadata: %v", err)
	}
	info, err := testApp().SecretInfoV2(configPath, vaultPath, "api-token")
	if err != nil {
		t.Fatalf("SecretInfoV2: %v", err)
	}
	want := map[string]string{"provider": "Example", "website": "https://example.test", "expires_at": "2026-12-31"}
	for _, field := range info.Fields {
		if want[field.ID] != "" && field.Value != want[field.ID] {
			t.Fatalf("%s = %q, want %q", field.ID, field.Value, want[field.ID])
		}
	}
}

func TestSetSecretFieldV2(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	// Create a secret first
	_, err := testApp().CreateSecretV2(configPath, vaultPath, CreateSecretRequest{
		Title: "test-set",
		Stdin: true,
		In:    strings.NewReader("initial-value"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Set a new value via stdin
	err = testApp().SetSecretFieldV2(configPath, vaultPath, SetSecretFieldRequest{
		Selector: "test-set",
		Stdin:    true,
		In:       strings.NewReader("updated-value"),
	})
	if err != nil {
		t.Fatalf("SetSecretFieldV2: %v", err)
	}

	// Verify the updated value is returned
	val, err := testApp().GetSecretFieldV2(configPath, vaultPath, GetSecretFieldRequest{
		Selector: "test-set",
	})
	if err != nil {
		t.Fatalf("GetSecretFieldV2 after set: %v", err)
	}
	if val != "updated-value" {
		t.Fatalf("value = %q, want updated-value", val)
	}

	// Set without Stdin should fail
	err = testApp().SetSecretFieldV2(configPath, vaultPath, SetSecretFieldRequest{
		Selector: "test-set",
		Stdin:    false,
	})
	if err == nil || !strings.Contains(err.Error(), "--stdin") {
		t.Fatalf("expected --stdin error, got: %v", err)
	}
}

func TestGetSecretFieldV2(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	// Create with a value
	_, err := testApp().CreateSecretV2(configPath, vaultPath, CreateSecretRequest{
		Title: "get-test",
		Env:   "GET_TEST_TOKEN",
		Stdin: true,
		In:    strings.NewReader("secret-value-42"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Get by selector should return only the raw value, no metadata framing
	val, err := testApp().GetSecretFieldV2(configPath, vaultPath, GetSecretFieldRequest{
		Selector: "get-test",
	})
	if err != nil {
		t.Fatalf("GetSecretFieldV2: %v", err)
	}
	if val != "secret-value-42" {
		t.Fatalf("value = %q, want secret-value-42", val)
	}

	// Get by env name should also work and return only the value
	val, err = testApp().GetSecretFieldV2(configPath, vaultPath, GetSecretFieldRequest{
		Env: "GET_TEST_TOKEN",
	})
	if err != nil {
		t.Fatalf("GetSecretFieldV2 by env: %v", err)
	}
	if val != "secret-value-42" {
		t.Fatalf("env get value = %q, want secret-value-42", val)
	}

	// Non-existent selector
	_, err = testApp().GetSecretFieldV2(configPath, vaultPath, GetSecretFieldRequest{
		Selector: "nonexistent",
	})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got: %v", err)
	}
}

func TestListSecretItemsV2_ValueFree(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	// Create two secrets
	for _, tc := range []struct {
		title string
		env   string
		value string
	}{
		{"alpha", "ALPHA_VAR", "secret-alpha"},
		{"beta", "BETA_VAR", "secret-beta"},
	} {
		_, err := testApp().CreateSecretV2(configPath, vaultPath, CreateSecretRequest{
			Title: tc.title,
			Env:   tc.env,
			Stdin: true,
			In:    strings.NewReader(tc.value),
		})
		if err != nil {
			t.Fatalf("create %s: %v", tc.title, err)
		}
	}

	// List all items
	items, err := testApp().ListSecretItemsV2(configPath, vaultPath, ListSecretItemsRequest{})
	if err != nil {
		t.Fatalf("ListSecretItemsV2: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}

	// List output has ID, Title, Template, Env, ValueSet — not the secret values themselves
	for _, item := range items {
		if item.Env == "ALPHA_VAR" && item.ID != "item_alpha" {
			t.Fatalf("alpha: id=%q title=%q", item.ID, item.Title)
		}
		if !item.ValueSet {
			t.Errorf("item %s should have value_set=true", item.ID)
		}
	}

	// Query filtering
	items, err = testApp().ListSecretItemsV2(configPath, vaultPath, ListSecretItemsRequest{
		Query: "alpha",
	})
	if err != nil {
		t.Fatalf("ListSecretItemsV2 with query: %v", err)
	}
	if len(items) != 1 || items[0].ID != "item_alpha" {
		t.Fatalf("query gave %#v, want [item_alpha]", items)
	}
}

func TestSecretInfoV2_ValueFree(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	_, err := testApp().CreateSecretV2(configPath, vaultPath, CreateSecretRequest{
		Title: "info-test",
		Env:   "INFO_TOKEN",
		Stdin: true,
		In:    strings.NewReader("secret-info-value"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	info, err := testApp().SecretInfoV2(configPath, vaultPath, "info-test")
	if err != nil {
		t.Fatalf("SecretInfoV2: %v", err)
	}
	if info.ID != "item_info_test" {
		t.Fatalf("id = %q, want item_info_test", info.ID)
	}
	if info.Title != "info-test" {
		t.Fatalf("title = %q, want info-test", info.Title)
	}
	if info.Template != "generic_secret" {
		t.Fatalf("template = %q, want generic_secret", info.Template)
	}

	// The primary secret field value must NOT appear in info output (it's concealed/secret)
	for _, f := range info.Fields {
		if f.ID == "value" && f.Value != "" {
			t.Fatalf("info leaked secret field value: %s", f.Value)
		}
	}
	// value_env is public and should appear when set
	foundEnv := false
	for _, f := range info.Fields {
		if f.ID == "value_env" {
			foundEnv = true
			if f.Value != "INFO_TOKEN" {
				t.Fatalf("value_env = %q, want INFO_TOKEN", f.Value)
			}
		}
	}
	if !foundEnv {
		t.Fatal("missing value_env field in info")
	}
}

func TestV2GetSecretEnvAmbiguityIsValueFree(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)

	// Create two items with the same env name
	for _, tc := range []struct {
		title string
		value string
	}{
		{"first", "first-secret-val"},
		{"second", "second-secret-val"},
	} {
		_, err := testApp().CreateSecretV2(configPath, vaultPath, CreateSecretRequest{
			Title: tc.title,
			Env:   "DUPLICATE_ENV",
			Stdin: true,
			In:    strings.NewReader(tc.value),
		})
		if err != nil {
			t.Fatalf("create %s: %v", tc.title, err)
		}
	}

	// Get by env name should fail with ambiguity
	val, err := testApp().GetSecretFieldV2(configPath, vaultPath, GetSecretFieldRequest{
		Env: "DUPLICATE_ENV",
	})
	if err == nil {
		t.Fatalf("expected ambiguous env error, got value: %s", val)
	}
	// The error message must NOT contain secret values
	errMsg := err.Error()
	if strings.Contains(errMsg, "first-secret-val") || strings.Contains(errMsg, "second-secret-val") {
		t.Fatalf("ambiguity error leaked secret values: %s", errMsg)
	}
	if !strings.Contains(errMsg, "env selector matched multiple fields") {
		t.Fatalf("unexpected error message: %s", errMsg)
	}
	// Item ID/title in the error are safe (non-secret)
	if !strings.Contains(errMsg, "item_first") || !strings.Contains(errMsg, "first") {
		t.Fatalf("error should contain item identities: %s", errMsg)
	}
}

func TestV2SecretInfoAndListNonExistent(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.v2.json")
	configPath := writeV2Config(t, vaultPath)
	writeV2EmptyVault(t, vaultPath)

	// Info on non-existent item
	_, err := testApp().SecretInfoV2(configPath, vaultPath, "nothing")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found, got: %v", err)
	}

	// List on empty vault
	items, err := testApp().ListSecretItemsV2(configPath, vaultPath, ListSecretItemsRequest{})
	if err != nil {
		t.Fatalf("ListSecretItemsV2 on empty vault: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

// --- v2 label-scoped secret export tests ---

func TestExportSecretItemsV2_NoLabels(t *testing.T) {
	_, err := testApp().ExportSecretItemsV2("", "", ExportSecretItemsRequest{Labels: nil, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "requires --label") {
		t.Fatalf("expected --label required error, got: %v", err)
	}
	_, err = testApp().ExportSecretItemsV2("", "", ExportSecretItemsRequest{Labels: []string{}, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "requires --label") {
		t.Fatalf("expected --label required error on empty, got: %v", err)
	}
}

func TestExportSecretItemsV2_BySingleLabel(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_db", "DB Credentials", "DATABASE_URL", "postgres://localhost", []string{"db", "prod"}),
		genericSecretItem("item_cache", "Cache Key", "REDIS_URL", "redis://localhost", []string{"cache", "prod"}),
		genericSecretItem("item_local", "Local Config", "LOCAL_VAR", "local-value", []string{"local"}),
	})
	out, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"prod"}, Format: "shell"})
	if err != nil {
		t.Fatalf("export by label: %v", err)
	}
	if !strings.Contains(out, "DATABASE_URL=postgres://localhost") {
		t.Fatalf("missing DATABASE_URL in:\n%s", out)
	}
	if !strings.Contains(out, "REDIS_URL=redis://localhost") {
		t.Fatalf("missing REDIS_URL in:\n%s", out)
	}
	if strings.Contains(out, "LOCAL_VAR") {
		t.Fatalf("local item should not match label prod:\n%s", out)
	}
}

func TestExportSecretItemsV2_RepeatedLabelsUseAND(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "value-a", []string{"db", "prod"}),
		genericSecretItem("item_b", "Item B", "B_VAR", "value-b", []string{"db", "staging"}),
		genericSecretItem("item_c", "Item C", "C_VAR", "value-c", []string{"prod"}),
	})
	out, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"db", "prod"}, Format: "shell"})
	if err != nil {
		t.Fatalf("export by AND labels: %v", err)
	}
	if !strings.Contains(out, "A_VAR=value-a") {
		t.Fatalf("missing A_VAR in:\n%s", out)
	}
	if strings.Contains(out, "B_VAR") {
		t.Fatalf("B_VAR should not match (no prod label):\n%s", out)
	}
	if strings.Contains(out, "C_VAR") {
		t.Fatalf("C_VAR should not match (no db label):\n%s", out)
	}
}

func TestExportSecretItemsV2_NoMatchingLabels(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "value-a", []string{"prod"}),
	})
	_, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"nonexistent"}, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "no items matched") {
		t.Fatalf("expected no-match error, got: %v", err)
	}
}

func TestExportSecretItemsV2_SkipsMissingEnvName(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "", "value-a", []string{"prod"}),
		genericSecretItem("item_b", "Item B", "B_VAR", "value-b", []string{"prod"}),
	})
	out, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"prod"}, Format: "shell"})
	if err != nil {
		t.Fatalf("export with missing env item: %v", err)
	}
	if strings.Contains(out, "value-a") || strings.Contains(out, "Item A") {
		t.Fatalf("item without env should be skipped:\n%s", out)
	}
	if !strings.Contains(out, "B_VAR=value-b") {
		t.Fatalf("item with env missing from output:\n%s", out)
	}
}

func TestExportSecretItemsV2_MissingValue(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		emptyValueSecretItem("item_a", "Item A", "A_VAR", []string{"prod"}),
	})
	_, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"prod"}, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "missing value") {
		t.Fatalf("expected missing value error, got: %v", err)
	}
}

func TestExportSecretItemsV2_DuplicateEnvName(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "DUPE_VAR", "value-a", []string{"prod"}),
		genericSecretItem("item_b", "Item B", "DUPE_VAR", "value-b", []string{"prod"}),
	})
	_, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"prod"}, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "duplicate env name") {
		t.Fatalf("expected duplicate env error, got: %v", err)
	}
	if strings.Contains(err.Error(), "value-a") || strings.Contains(err.Error(), "value-b") {
		t.Fatalf("duplicate error leaked secret values: %v", err)
	}
}

func TestExportSecretItemsV2_SortsByEnvName(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_z", "Z Item", "ZZZ_VAR", "zzz-value", []string{"all"}),
		genericSecretItem("item_a", "A Item", "AAA_VAR", "aaa-value", []string{"all"}),
		genericSecretItem("item_m", "M Item", "MMM_VAR", "mmm-value", []string{"all"}),
	})
	out, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"all"}, Format: "shell"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], "export AAA_VAR=") {
		t.Fatalf("first line should be AAA_VAR, got: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "export MMM_VAR=") {
		t.Fatalf("second line should be MMM_VAR, got: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "export ZZZ_VAR=") {
		t.Fatalf("third line should be ZZZ_VAR, got: %s", lines[2])
	}
}

func TestExportSecretItemsV2_ShellFormat(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "simple", []string{"all"}),
		genericSecretItem("item_b", "Item B", "B_VAR", "has spaces and $pecial", []string{"all"}),
	})
	out, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{
		Labels: []string{"all"},
		Format: "shell",
	})
	if err != nil {
		t.Fatalf("export shell: %v", err)
	}
	if !strings.Contains(out, "export A_VAR=simple") {
		t.Fatalf("missing export A_VAR in:\n%s", out)
	}
	if !strings.Contains(out, "export B_VAR='has spaces and $pecial'") {
		t.Fatalf("missing quoted export B_VAR in:\n%s", out)
	}
}

func TestExportSecretItemsV2_JSONFormat(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "alpha", []string{"all"}),
		genericSecretItem("item_b", "Item B", "B_VAR", "beta", []string{"all"}),
	})
	out, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{
		Labels: []string{"all"},
		Format: "json",
	})
	if err != nil {
		t.Fatalf("export json: %v", err)
	}
	if !strings.Contains(out, `"A_VAR": "alpha"`) {
		t.Fatalf("missing A_VAR in JSON:\n%s", out)
	}
	if !strings.Contains(out, `"B_VAR": "beta"`) {
		t.Fatalf("missing B_VAR in JSON:\n%s", out)
	}
}

func TestExportSecretItemsV2_UnsupportedFormat(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "value-a", []string{"all"}),
	})
	_, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{
		Labels: []string{"all"},
		Format: "toml",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported export format") {
		t.Fatalf("expected unsupported format error, got: %v", err)
	}
}

func TestExportSecretItemsV2_ValueFreeDiagnostics(t *testing.T) {
	vaultPath, configPath := writeV2VaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "actual-secret-value", []string{"prod"}),
	})
	_, err := testApp().ExportSecretItemsV2(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"nonexistent"}, Format: "shell"})
	if err != nil && (strings.Contains(err.Error(), "actual-secret-value") || strings.Contains(err.Error(), "sk-")) {
		t.Fatalf("error leaked secret value: %v", err)
	}
}

// --- helpers ---

func genericSecretItem(id, title, envName, value string, labels []string) vaultmodel.Item {
	return vaultmodel.Item{
		ID:       id,
		Title:    title,
		Template: vaultmodel.TemplateGenericSecret,
		Labels:   labels,
		Fields: []vaultmodel.Field{
			{ID: "value", Label: "Value", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Value: rawJSON(value)},
			{ID: "value_env", Label: "Value Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeEnvName, Value: rawJSON(envName)},
		},
	}
}

func emptyValueSecretItem(id, title, envName string, labels []string) vaultmodel.Item {
	return vaultmodel.Item{
		ID:       id,
		Title:    title,
		Template: vaultmodel.TemplateGenericSecret,
		Labels:   labels,
		Fields: []vaultmodel.Field{
			{ID: "value", Label: "Value", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret},
			{ID: "value_env", Label: "Value Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeEnvName, Value: rawJSON(envName)},
		},
	}
}

func writeV2VaultWithItems(t *testing.T, items []vaultmodel.Item) (vaultPath, configPath string) {
	t.Helper()
	vaultPath = filepath.Join(t.TempDir(), "vault.v2.json")
	configPath = writeV2Config(t, vaultPath)
	repo := vaultjson.NewRepository(vaultPath)
	if err := repo.Save(vaultmodel.Vault{Version: vaultmodel.CurrentVersion, Items: items}); err != nil {
		t.Fatalf("write vault: %v", err)
	}
	return vaultPath, configPath
}

func rawJSON(value string) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
