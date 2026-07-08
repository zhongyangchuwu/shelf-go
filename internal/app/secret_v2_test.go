package app

import (
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

func TestSecretV2UsesDistinctDefaultPath(t *testing.T) {
	legacyVaultPath := filepath.Join(t.TempDir(), "vault.age")
	configPath := writeV2Config(t, legacyVaultPath)

	if _, err := testApp().CreateSecretV2(configPath, "", CreateSecretRequest{Title: "default-path", Stdin: true, In: strings.NewReader("fake-default-secret")}); err != nil {
		t.Fatalf("CreateSecretV2: %v", err)
	}
	if _, err := os.Stat(legacyVaultPath); !os.IsNotExist(err) {
		t.Fatalf("legacy vault path was touched: %v", err)
	}
	if _, err := os.Stat(legacyVaultPath + ".v2.json"); err != nil {
		t.Fatalf("v2 vault path missing: %v", err)
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
