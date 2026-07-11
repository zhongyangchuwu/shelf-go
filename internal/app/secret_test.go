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

// writeConfig writes a minimal YAML config pointing vault_path at jsonPath
// and returns the config file path.
func writeConfig(t *testing.T, jsonPath string) string {
	t.Helper()
	dir := filepath.Dir(jsonPath)
	configPath := filepath.Join(dir, "shelf-test-config.yaml")
	content := fmt.Sprintf("vault_path: %s\n", jsonPath)
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write item config: %v", err)
	}
	return configPath
}

// writeEmptyVault writes a minimal valid item vault JSON file.
func writeEmptyVault(t *testing.T, jsonPath string) {
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
	configPath := writeConfig(t, vaultPath)

	if _, err := testApp().CreateSecret(configPath, "", CreateSecretRequest{Title: "default-path", Stdin: true, In: strings.NewReader("fake-default-secret")}); err != nil {
		t.Fatalf("CreateSecret: %v", err)
	}
	if _, err := os.Stat(vaultPath); err != nil {
		t.Fatalf("vault path missing: %v", err)
	}
	value, err := testApp().GetSecretField(configPath, "", GetSecretFieldRequest{Selector: "default-path"})
	if err != nil {
		t.Fatalf("GetSecretField: %v", err)
	}
	if value != "fake-default-secret" {
		t.Fatalf("value = %q, want fake-default-secret", value)
	}
}

func TestCreateSecret(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)

	// Create a basic secret with stdin value
	req := CreateSecretRequest{
		Title: "my-token",
		Env:   "MY_TOKEN",
		Stdin: true,
		In:    strings.NewReader("sk-fake-value"),
	}
	id, err := testApp().CreateSecret(configPath, vaultPath, req)
	if err != nil {
		t.Fatalf("CreateSecret: %v", err)
	}
	if id != "item_my_token" {
		t.Fatalf("id = %q, want item_my_token", id)
	}
	// Duplicate title should fail (ID uniqueness is checked first since same ID is derived)
	_, err = testApp().CreateSecret(configPath, vaultPath, req)
	if err == nil {
		t.Fatal("expected error on duplicate, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already-exists error, got: %v", err)
	}
	_, err = testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
		Title: "  ",
		Stdin: true,
		In:    strings.NewReader("x"),
	})
	if err == nil || !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("expected title required error, got: %v", err)
	}
}

func TestCreateSecretCustomTemplate(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)
	custom := vaultmodel.Template{
		ID:   "service_token",
		Name: "Service Token",
		Fields: []vaultmodel.FieldTemplate{
			{ID: "token", Label: "Token", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Required: true},
			{ID: "token_env", Label: "Token Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeEnvName, EnvFor: "token"},
		},
	}
	repo := vaultjson.NewRepository(vaultPath)
	if err := repo.Save(vaultmodel.Vault{Version: vaultmodel.CurrentVersion, Templates: []vaultmodel.Template{custom}}); err != nil {
		t.Fatalf("write custom template vault: %v", err)
	}
	if _, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
		Title:    "service credential",
		Template: custom.ID,
		Env:      "SERVICE_TOKEN",
		Stdin:    true,
		In:       strings.NewReader("fake-custom-token"),
	}); err != nil {
		t.Fatalf("create custom-template secret: %v", err)
	}
	info, err := testApp().SecretInfo(configPath, vaultPath, "service credential")
	if err != nil {
		t.Fatalf("custom secret info: %v", err)
	}
	if info.Template != custom.ID {
		t.Fatalf("template = %q, want %q", info.Template, custom.ID)
	}
	for _, field := range info.Fields {
		if field.ID == "token_env" && field.Value == "SERVICE_TOKEN" {
			return
		}
	}
	t.Fatalf("custom template env field missing from info: %#v", info.Fields)
}

func TestCreateSecretProfileOpenAI(t *testing.T) {
	// Equivalent to `--profile openai --stdin` in the CLI
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)

	req := CreateSecretRequest{
		Title:   "openai-key",
		Profile: "openai",
		Stdin:   true,
		In:      strings.NewReader("sk-fake-openai-key"),
	}
	id, err := testApp().CreateSecret(configPath, vaultPath, req)
	if err != nil {
		t.Fatalf("CreateSecret with openai profile: %v", err)
	}
	if id != "item_openai_key" {
		t.Fatalf("id = %q, want item_openai_key", id)
	}

	// Verify the item was created with openai template defaults
	info, err := testApp().SecretInfo(configPath, vaultPath, "openai-key")
	if err != nil {
		t.Fatalf("SecretInfo: %v", err)
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

func TestCreateSecretRejectsUnsupportedMetadataFlags(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)

	_, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
		Title:    "generic-with-provider",
		Provider: "Example",
		Stdin:    true,
		In:       strings.NewReader("fake-secret"),
	})
	if err == nil || !strings.Contains(err.Error(), "--provider is not supported by template generic_secret") {
		t.Fatalf("expected unsupported provider error, got: %v", err)
	}
}

func TestCreateSecretAcceptsAPIMetadataFlags(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)

	_, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
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
		t.Fatalf("CreateSecret with api metadata: %v", err)
	}
	info, err := testApp().SecretInfo(configPath, vaultPath, "api-token")
	if err != nil {
		t.Fatalf("SecretInfo: %v", err)
	}
	want := map[string]string{"provider": "Example", "website": "https://example.test", "expires_at": "2026-12-31"}
	for _, field := range info.Fields {
		if want[field.ID] != "" && field.Value != want[field.ID] {
			t.Fatalf("%s = %q, want %q", field.ID, field.Value, want[field.ID])
		}
	}
}

func TestUpdateSecretReadsFieldInput(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)
	if _, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
		Title: "test-update",
		Env:   "APP_TOKEN",
		Stdin: true,
		In:    strings.NewReader("initial-value"),
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{
		Selector:  "test-update",
		ReadField: true,
		In:        strings.NewReader("updated-value"),
	}); err != nil {
		t.Fatalf("update primary from stdin: %v", err)
	}
	if err := testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{
		Selector:  "test-update",
		Field:     "value_env",
		ReadField: true,
		In:        strings.NewReader("APP_TOKEN_RENAMED"),
	}); err != nil {
		t.Fatalf("update public field from stdin: %v", err)
	}
	val, err := testApp().GetSecretField(configPath, vaultPath, GetSecretFieldRequest{Selector: "test-update"})
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if val != "updated-value" {
		t.Fatalf("value = %q, want updated-value", val)
	}
	if _, err := testApp().GetSecretField(configPath, vaultPath, GetSecretFieldRequest{Env: "APP_TOKEN_RENAMED"}); err != nil {
		t.Fatalf("updated public env field unavailable: %v", err)
	}
	if err := testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{Selector: "test-update"}); err == nil || !strings.Contains(err.Error(), "at least one change") {
		t.Fatalf("empty update error = %v", err)
	}
}

func TestUpdateSecretMaintainsItemAndPublicFields(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)
	if _, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
		Title:    "openai-personal",
		Template: vaultmodel.TemplateAPICredential,
		Env:      "OPENAI_API_KEY",
		Provider: "OpenAI",
		Stdin:    true,
		In:       strings.NewReader("fake-openai-token"),
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	title := "OpenAI Personal"
	if err := testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{
		Selector:      "openai-personal",
		Title:         title,
		TitleSet:      true,
		AddLabels:     []string{"ai", "personal", "ai"},
		AddAliases:    []string{"openai", "my-openai"},
		Field:         "provider",
		Values:        []string{"OpenAI personal account"},
		FieldValueSet: true,
	}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{
		Selector:      title,
		RemoveLabels:  []string{"personal"},
		RemoveAliases: []string{"my-openai"},
		Field:         "scopes",
		Values:        []string{"responses", "embeddings"},
		FieldValueSet: true,
	}); err != nil {
		t.Fatalf("update scopes: %v", err)
	}
	info, err := testApp().SecretInfo(configPath, vaultPath, "openai")
	if err != nil {
		t.Fatalf("info by alias: %v", err)
	}
	if info.Title != title || !hasString(info.Labels, "ai") || hasString(info.Labels, "personal") || hasString(info.Aliases, "my-openai") {
		t.Fatalf("updated item metadata = %#v", info)
	}
	publicValues := map[string]string{}
	for _, field := range info.Fields {
		publicValues[field.ID] = field.Value
	}
	if publicValues["provider"] != "OpenAI personal account" {
		t.Fatalf("provider = %q", publicValues["provider"])
	}
	var scopes []string
	if err := json.Unmarshal([]byte(publicValues["scopes"]), &scopes); err != nil {
		t.Fatalf("decode scopes: %v", err)
	}
	if strings.Join(scopes, ",") != "responses,embeddings" {
		t.Fatalf("scopes = %#v", scopes)
	}
	value, err := testApp().GetSecretField(configPath, vaultPath, GetSecretFieldRequest{Selector: "openai"})
	if err != nil {
		t.Fatalf("secret value unavailable after update: %v", err)
	}
	if value != "fake-openai-token" {
		t.Fatalf("secret value changed during metadata update: %q", value)
	}
}

func TestUpdateSecretClearAndFailureAreAtomicAndValueFree(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)
	if _, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
		Title:    "api-token",
		Template: vaultmodel.TemplateAPICredential,
		Website:  "https://example.test",
		Stdin:    true,
		In:       strings.NewReader("actual-secret-value"),
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{Selector: "api-token", Field: "website", ClearField: true}); err != nil {
		t.Fatalf("clear website: %v", err)
	}
	info, err := testApp().SecretInfo(configPath, vaultPath, "api-token")
	if err != nil {
		t.Fatalf("info after clear: %v", err)
	}
	for _, field := range info.Fields {
		if field.ID == "website" && field.ValueSet {
			t.Fatalf("website remains set after clear: %#v", field)
		}
	}
	before, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("read vault before invalid update: %v", err)
	}
	err = testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{Selector: "api-token", Field: "expires_at", Values: []string{"not-a-date"}, FieldValueSet: true})
	if err == nil || !strings.Contains(err.Error(), "invalid date") {
		t.Fatalf("invalid public update error = %v", err)
	}
	after, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("read vault after invalid update: %v", err)
	}
	if string(after) != string(before) {
		t.Fatal("invalid update changed vault")
	}
	err = testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{Selector: "api-token", Field: "credential", Values: []string{"attempted-secret"}, FieldValueSet: true})
	if err == nil || !strings.Contains(err.Error(), "stdin input") {
		t.Fatalf("secret field update error = %v", err)
	}
	if strings.Contains(err.Error(), "actual-secret-value") || strings.Contains(err.Error(), "attempted-secret") {
		t.Fatalf("secret field error leaked a value: %v", err)
	}
}

func TestUpdateSecretRejectsSelectorConflicts(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)
	for _, title := range []string{"first", "second"} {
		if _, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{Title: title}); err != nil {
			t.Fatalf("create %s: %v", title, err)
		}
	}
	err := testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{Selector: "first", AddAliases: []string{"second"}})
	if err == nil || !strings.Contains(err.Error(), "selector already exists") {
		t.Fatalf("alias conflict error = %v", err)
	}
	title := "second"
	err = testApp().UpdateSecret(configPath, vaultPath, UpdateSecretRequest{Selector: "first", Title: title, TitleSet: true})
	if err == nil || !strings.Contains(err.Error(), "selector already exists") {
		t.Fatalf("title conflict error = %v", err)
	}
}

func TestGetSecretField(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)

	// Create with a value
	_, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
		Title: "get-test",
		Env:   "GET_TEST_TOKEN",
		Stdin: true,
		In:    strings.NewReader("secret-value-42"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Get by selector should return only the raw value, no metadata framing
	val, err := testApp().GetSecretField(configPath, vaultPath, GetSecretFieldRequest{
		Selector: "get-test",
	})
	if err != nil {
		t.Fatalf("GetSecretField: %v", err)
	}
	if val != "secret-value-42" {
		t.Fatalf("value = %q, want secret-value-42", val)
	}

	// Get by env name should also work and return only the value
	val, err = testApp().GetSecretField(configPath, vaultPath, GetSecretFieldRequest{
		Env: "GET_TEST_TOKEN",
	})
	if err != nil {
		t.Fatalf("GetSecretField by env: %v", err)
	}
	if val != "secret-value-42" {
		t.Fatalf("env get value = %q, want secret-value-42", val)
	}

	// Non-existent selector
	_, err = testApp().GetSecretField(configPath, vaultPath, GetSecretFieldRequest{
		Selector: "nonexistent",
	})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got: %v", err)
	}
}

func TestListSecretItems_ValueFree(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)

	// Create two secrets
	for _, tc := range []struct {
		title string
		env   string
		value string
	}{
		{"alpha", "ALPHA_VAR", "secret-alpha"},
		{"beta", "BETA_VAR", "secret-beta"},
	} {
		_, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
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
	items, err := testApp().ListSecretItems(configPath, vaultPath, ListSecretItemsRequest{})
	if err != nil {
		t.Fatalf("ListSecretItems: %v", err)
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
	items, err = testApp().ListSecretItems(configPath, vaultPath, ListSecretItemsRequest{
		Query: "alpha",
	})
	if err != nil {
		t.Fatalf("ListSecretItems with query: %v", err)
	}
	if len(items) != 1 || items[0].ID != "item_alpha" {
		t.Fatalf("query gave %#v, want [item_alpha]", items)
	}
}

func TestSecretInfo_ValueFree(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)

	_, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
		Title: "info-test",
		Env:   "INFO_TOKEN",
		Stdin: true,
		In:    strings.NewReader("secret-info-value"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	info, err := testApp().SecretInfo(configPath, vaultPath, "info-test")
	if err != nil {
		t.Fatalf("SecretInfo: %v", err)
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

func TestGetSecretEnvAmbiguityIsValueFree(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)

	// Create two items with the same env name
	for _, tc := range []struct {
		title string
		value string
	}{
		{"first", "first-secret-val"},
		{"second", "second-secret-val"},
	} {
		_, err := testApp().CreateSecret(configPath, vaultPath, CreateSecretRequest{
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
	val, err := testApp().GetSecretField(configPath, vaultPath, GetSecretFieldRequest{
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

func TestSecretInfoAndListNonExistent(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.json")
	configPath := writeConfig(t, vaultPath)
	writeEmptyVault(t, vaultPath)

	// Info on non-existent item
	_, err := testApp().SecretInfo(configPath, vaultPath, "nothing")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found, got: %v", err)
	}

	// List on empty vault
	items, err := testApp().ListSecretItems(configPath, vaultPath, ListSecretItemsRequest{})
	if err != nil {
		t.Fatalf("ListSecretItems on empty vault: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

// --- item label-scoped secret export tests ---

func TestExportSecretItems_NoLabels(t *testing.T) {
	_, err := testApp().ExportSecretItems("", "", ExportSecretItemsRequest{Labels: nil, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "requires --label") {
		t.Fatalf("expected --label required error, got: %v", err)
	}
	_, err = testApp().ExportSecretItems("", "", ExportSecretItemsRequest{Labels: []string{}, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "requires --label") {
		t.Fatalf("expected --label required error on empty, got: %v", err)
	}
}

func TestExportSecretItems_BySingleLabel(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_db", "DB Credentials", "DATABASE_URL", "postgres://localhost", []string{"db", "prod"}),
		genericSecretItem("item_cache", "Cache Key", "REDIS_URL", "redis://localhost", []string{"cache", "prod"}),
		genericSecretItem("item_local", "Local Config", "LOCAL_VAR", "local-value", []string{"local"}),
	})
	out, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"prod"}, Format: "shell"})
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

func TestExportSecretItems_RepeatedLabelsUseAND(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "value-a", []string{"db", "prod"}),
		genericSecretItem("item_b", "Item B", "B_VAR", "value-b", []string{"db", "staging"}),
		genericSecretItem("item_c", "Item C", "C_VAR", "value-c", []string{"prod"}),
	})
	out, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"db", "prod"}, Format: "shell"})
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

func TestExportSecretItems_NoMatchingLabels(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "value-a", []string{"prod"}),
	})
	_, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"nonexistent"}, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "no items matched") {
		t.Fatalf("expected no-match error, got: %v", err)
	}
}

func TestExportSecretItems_SkipsMissingEnvName(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "", "value-a", []string{"prod"}),
		genericSecretItem("item_b", "Item B", "B_VAR", "value-b", []string{"prod"}),
	})
	out, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"prod"}, Format: "shell"})
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

func TestExportSecretItems_MissingValue(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		emptyValueSecretItem("item_a", "Item A", "A_VAR", []string{"prod"}),
	})
	_, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"prod"}, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "missing value") {
		t.Fatalf("expected missing value error, got: %v", err)
	}
}

func TestExportSecretItems_DuplicateEnvName(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "DUPE_VAR", "value-a", []string{"prod"}),
		genericSecretItem("item_b", "Item B", "DUPE_VAR", "value-b", []string{"prod"}),
	})
	_, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"prod"}, Format: "shell"})
	if err == nil || !strings.Contains(err.Error(), "duplicate env name") {
		t.Fatalf("expected duplicate env error, got: %v", err)
	}
	if strings.Contains(err.Error(), "value-a") || strings.Contains(err.Error(), "value-b") {
		t.Fatalf("duplicate error leaked secret values: %v", err)
	}
}

func TestExportSecretItems_SortsByEnvName(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_z", "Z Item", "ZZZ_VAR", "zzz-value", []string{"all"}),
		genericSecretItem("item_a", "A Item", "AAA_VAR", "aaa-value", []string{"all"}),
		genericSecretItem("item_m", "M Item", "MMM_VAR", "mmm-value", []string{"all"}),
	})
	out, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"all"}, Format: "shell"})
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

func TestExportSecretItems_ShellFormat(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "simple", []string{"all"}),
		genericSecretItem("item_b", "Item B", "B_VAR", "has spaces and $pecial", []string{"all"}),
	})
	out, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{
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

func TestExportSecretItems_JSONFormat(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "alpha", []string{"all"}),
		genericSecretItem("item_b", "Item B", "B_VAR", "beta", []string{"all"}),
	})
	out, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{
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

func TestExportSecretItems_UnsupportedFormat(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "value-a", []string{"all"}),
	})
	_, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{
		Labels: []string{"all"},
		Format: "toml",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported export format") {
		t.Fatalf("expected unsupported format error, got: %v", err)
	}
}

func TestExportSecretItems_ValueFreeDiagnostics(t *testing.T) {
	vaultPath, configPath := writeVaultWithItems(t, []vaultmodel.Item{
		genericSecretItem("item_a", "Item A", "A_VAR", "actual-secret-value", []string{"prod"}),
	})
	_, err := testApp().ExportSecretItems(configPath, vaultPath, ExportSecretItemsRequest{Labels: []string{"nonexistent"}, Format: "shell"})
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

func writeVaultWithItems(t *testing.T, items []vaultmodel.Item) (vaultPath, configPath string) {
	t.Helper()
	vaultPath = filepath.Join(t.TempDir(), "vault.json")
	configPath = writeConfig(t, vaultPath)
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
