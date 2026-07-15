package vaultjson

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestJSONCodecRoundTripsVaultOrderAndMetadata(t *testing.T) {
	codec := JSONCodec{}
	vault := validVault()
	content, err := codec.Encode(vault)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := codec.Decode(content)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(decoded.Items) != 2 {
		t.Fatalf("items length = %d", len(decoded.Items))
	}
	if decoded.Items[0].ID != "item_app" || decoded.Items[1].ID != "item_openai" {
		t.Fatalf("item order changed: %+v", decoded.Items)
	}
	fields := decoded.Items[1].Fields
	if len(fields) != 6 {
		t.Fatalf("fields length = %d", len(fields))
	}
	if fields[0].ID != "credential" || fields[5].ID != "scopes" {
		t.Fatalf("field order changed: %+v", fields)
	}
	// env_credential stores the credential's projected environment name.
	var envName string
	if err := json.Unmarshal(fields[1].Value, &envName); err != nil {
		t.Fatalf("unmarshal env_credential value: %v", err)
	}
	if envName != "OPENAI_API_KEY" {
		t.Fatalf("env_credential value = %q, want OPENAI_API_KEY", envName)
	}
	if decoded.Items[0].Description != "CLI deployment token" {
		t.Fatalf("items[0].description = %q", decoded.Items[0].Description)
	}

	// Metadata on second item
	item1 := decoded.Items[1]
	if len(item1.Labels) != 2 || item1.Labels[0] != "ai" || item1.Labels[1] != "personal" {
		t.Fatalf("items[1].labels = %v, want [ai personal]", item1.Labels)
	}
	if item1.Description != "rotated quarterly" {
		t.Fatalf("items[1].description = %q, want rotated quarterly", item1.Description)
	}
	if item1.CreatedAt != "2026-01-15T00:00:00Z" {
		t.Fatalf("items[1].created_at = %q, want 2026-01-15T00:00:00Z", item1.CreatedAt)
	}
	if item1.UpdatedAt != "2026-06-20T00:00:00Z" {
		t.Fatalf("items[1].updated_at = %q, want 2026-06-20T00:00:00Z", item1.UpdatedAt)
	}
}

func TestJSONCodecRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{name: "unknown top level field", content: `{"version":4,"items":[],"extra":true}`, wantErr: "unknown field"},
		{name: "unknown nested item field", content: `{"version":4,"items":[{"id":"item","title":"Title","template":"generic_secret","fields":[{"id":"value","value":"fake"}],"extra":true}]}`, wantErr: "unknown field"},
		{name: "trailing json", content: `{"version":4,"items":[]} {}`, wantErr: "trailing"},
		{name: "prior version", content: `{"version":3,"items":[]}`, wantErr: "unsupported vault version"},
		{name: "invalid model data", content: `{"version":4,"items":[{"id":"","title":"Title","template":"generic_secret","fields":[]}]}`, wantErr: "id is empty"},
		{name: "unsupported version zero", content: `{"version":0,"items":[]}`, wantErr: "unsupported vault version"},
		{name: "unsupported version negative", content: `{"version":-1,"items":[]}`, wantErr: "unsupported vault version"},
		{name: "unsupported version future", content: `{"version":42,"items":[]}`, wantErr: "unsupported vault version"},
		{name: "invalid model empty title", content: `{"version":4,"items":[{"id":"x","title":"","template":"generic_secret","fields":[{"id":"value","value":"s"}]}]}`, wantErr: "title is empty"},
		{name: "invalid model unknown template", content: `{"version":4,"items":[{"id":"x","title":"X","template":"bogus_template","fields":[]}]}`, wantErr: "template unknown"},
		{name: "invalid model missing required field", content: `{"version":4,"items":[{"id":"x","title":"X","template":"generic_secret","fields":[]}]}`, wantErr: "missing required field"},
		{name: "stale V3 field metadata", content: `{"version":4,"items":[{"id":"x","title":"X","template":"generic_secret","fields":[{"id":"value","label":"Value","value":"s"}]}]}`, wantErr: "unknown field"},
		{name: "removed item aliases", content: `{"version":4,"items":[{"id":"x","title":"X","template":"generic_secret","aliases":["legacy"],"fields":[{"id":"value","value":"s"}]}]}`, wantErr: "unknown field"},
		{name: "removed item notes", content: `{"version":4,"items":[{"id":"x","title":"X","template":"generic_secret","notes":"legacy","fields":[{"id":"value","value":"s"}]}]}`, wantErr: "unknown field"},
		{name: "invalid model bad field value type", content: `{"version":4,"items":[{"id":"x","title":"X","template":"api_credential","fields":[{"id":"credential","value":"x"},{"id":"validity_credential","value":"unknown"},{"id":"scopes","value":"not_an_array"}]}]}`, wantErr: "value must be string array"},
	}
	codec := JSONCodec{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := codec.Decode([]byte(tt.content))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Decode error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRepositoryLoadMissingFileReturnsNotExist(t *testing.T) {
	repo := NewRepository(filepath.Join(t.TempDir(), "missing.json"))
	_, err := repo.Load()
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load error = %v, want os.ErrNotExist", err)
	}
}
func TestRepositorySaveThenLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	repo := NewRepository(path)
	vault := validVault()

	if err := repo.Save(vault); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := repo.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Version != vaultmodel.CurrentVersion {
		t.Errorf("version = %d, want %d", loaded.Version, vaultmodel.CurrentVersion)
	}
	if len(loaded.Items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(loaded.Items))
	}
	if loaded.Items[0].ID != "item_app" || loaded.Items[1].ID != "item_openai" {
		t.Fatalf("item IDs = [%q %q], want [item_app item_openai]", loaded.Items[0].ID, loaded.Items[1].ID)
	}
	if loaded.Items[1].Labels[0] != "ai" {
		t.Fatalf("labels[0] = %q, want ai", loaded.Items[1].Labels[0])
	}
}

func TestRepositorySaveLoadAndBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	repo := NewRepository(path)
	if err := repo.Save(validVault()); err != nil {
		t.Fatalf("first save: %v", err)
	}
	firstContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read first save: %v", err)
	}
	vault := validVault()
	vault.Items[0].Title = "Updated token"
	if err := repo.Save(vault); err != nil {
		t.Fatalf("second save: %v", err)
	}
	backup, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backup) != string(firstContent) {
		t.Fatalf("backup did not preserve first save")
	}
	loaded, err := repo.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Items[0].Title != "Updated token" {
		t.Fatalf("loaded title = %q", loaded.Items[0].Title)
	}
}

func TestRepositorySaveRejectsInvalidVaultWithoutReplacingTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	repo := NewRepository(path)
	if err := repo.Save(validVault()); err != nil {
		t.Fatalf("first save: %v", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}
	invalid := validVault()
	invalid.Items[0].Fields = nil
	if err := repo.Save(invalid); err == nil {
		t.Fatalf("expected invalid save to fail")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("invalid save replaced target")
	}
}

func validVault() vaultmodel.Vault {
	return vaultmodel.Vault{Version: vaultmodel.CurrentVersion, Items: []vaultmodel.Item{
		{
			ID: "item_app", Title: "Local app token", Template: vaultmodel.TemplateGenericSecret, Description: "CLI deployment token",
			Fields: []vaultmodel.Field{
				{ID: "value", Value: []byte(`"fake-app-token"`)},
				{ID: vaultmodel.EnvFieldID("value"), Value: []byte(`"APP_TOKEN"`)},
			},
		},
		{
			ID: "item_openai", Title: "OpenAI personal API key", Template: vaultmodel.TemplateAPICredential,
			Labels: []string{"ai", "personal"}, Description: "rotated quarterly", CreatedAt: "2026-01-15T00:00:00Z", UpdatedAt: "2026-06-20T00:00:00Z",
			Fields: []vaultmodel.Field{
				{ID: "credential", Value: []byte(`"fake-openai-key"`)},
				{ID: vaultmodel.EnvFieldID("credential"), Value: []byte(`"OPENAI_API_KEY"`)},
				{ID: vaultmodel.ValidityFieldID("credential"), Value: []byte(`"2026-12-31"`)},
				{ID: "provider", Value: []byte(`"openai"`)},
				{ID: "website", Value: []byte(`"https://platform.openai.com/api-keys"`)},
				{ID: "scopes", Value: []byte(`["responses","embeddings"]`)},
			},
		},
	}}
}

func TestJSONCodecRoundTripsMinimalVault(t *testing.T) {
	codec := JSONCodec{}
	vault := vaultmodel.Vault{Version: vaultmodel.CurrentVersion}
	content, err := codec.Encode(vault)
	if err != nil {
		t.Fatalf("encode empty vault: %v", err)
	}
	decoded, err := codec.Decode(content)
	if err != nil {
		t.Fatalf("decode empty vault: %v", err)
	}
	if decoded.Version != vaultmodel.CurrentVersion {
		t.Errorf("version = %d, want %d", decoded.Version, vaultmodel.CurrentVersion)
	}
	if len(decoded.Items) != 0 {
		t.Errorf("items = %d, want 0", len(decoded.Items))
	}
}

func TestJSONCodecRejectsUnknownFieldOnFieldObject(t *testing.T) {
	codec := JSONCodec{}
	content := `{"version":3,"items":[{"id":"x","title":"X","template":"generic_secret","fields":[{"id":"value","value":"s","unknown_field":true}]}]}`
	if _, err := codec.Decode([]byte(content)); err == nil {
		t.Fatal("expected error for unknown field on field, got nil")
	}
}

func TestJSONCodecRoundTripsEnvNameValuesPreserved(t *testing.T) {
	codec := JSONCodec{}
	vault := validVault()
	content, err := codec.Encode(vault)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := codec.Decode(content)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// item 0: generic_secret — env_value holds the environment name.
	item0 := decoded.Items[0]
	if len(item0.Fields) != 2 {
		t.Fatalf("items[0] fields = %d, want 2", len(item0.Fields))
	}
	var envName string
	if err := json.Unmarshal(item0.Fields[1].Value, &envName); err != nil {
		t.Fatalf("unmarshal env_value: %v", err)
	}
	if envName != "APP_TOKEN" {
		t.Errorf("env_value = %q, want APP_TOKEN", envName)
	}

	// item 1: api_credential — env_credential holds the environment name.
	item1 := decoded.Items[1]
	if len(item1.Fields) != 6 {
		t.Fatalf("items[1] fields = %d, want 6", len(item1.Fields))
	}
	if err := json.Unmarshal(item1.Fields[1].Value, &envName); err != nil {
		t.Fatalf("unmarshal env_credential: %v", err)
	}
	if envName != "OPENAI_API_KEY" {
		t.Errorf("env_credential = %q, want OPENAI_API_KEY", envName)
	}
}

func TestJSONCodecRoundTripsFieldValuesPreserved(t *testing.T) {
	codec := JSONCodec{}
	vault := validVault()
	content, err := codec.Encode(vault)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := codec.Decode(content)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	item1 := decoded.Items[1]
	var provider string
	if err := json.Unmarshal(item1.Fields[3].Value, &provider); err != nil {
		t.Fatalf("unmarshal provider value: %v", err)
	}
	if provider != "openai" {
		t.Errorf("provider = %q, want openai", provider)
	}
	var website string
	if err := json.Unmarshal(item1.Fields[4].Value, &website); err != nil {
		t.Fatalf("unmarshal website value: %v", err)
	}
	if website != "https://platform.openai.com/api-keys" {
		t.Errorf("website = %q", website)
	}
}

func TestJSONCodecEncodeRejectsInvalidVault(t *testing.T) {
	codec := JSONCodec{}
	invalid := vaultmodel.Vault{Version: vaultmodel.CurrentVersion, Items: []vaultmodel.Item{
		{ID: "x", Title: "X", Template: "nonexistent"},
	}}
	if _, err := codec.Encode(invalid); err == nil {
		t.Fatal("expected error encoding vault with unknown template, got nil")
	}
}

func TestJSONCodecDecodeRejectsTrailingJSON(t *testing.T) {
	codec := JSONCodec{}
	content := `{"version":3,"items":[]} trailing`
	if _, err := codec.Decode([]byte(content)); err == nil {
		t.Fatal("expected error for trailing content, got nil")
	}
}

func TestJSONCodecDecodeRejectsMultipleValues(t *testing.T) {
	codec := JSONCodec{}
	content := `{"version":3,"items":[]}{"version":3,"items":[]}`
	if _, err := codec.Decode([]byte(content)); err == nil {
		t.Fatal("expected error for multiple JSON values, got nil")
	}
}

func TestJSONCodecRoundTripsTemplateFieldOrderStrict(t *testing.T) {
	codec := JSONCodec{}
	vault := validVault()
	content, err := codec.Encode(vault)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := codec.Decode(content)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// api_credential field order must persist as written.
	item1 := decoded.Items[1]
	expected := []string{"credential", "env_credential", "validity_credential", "provider", "website", "scopes"}
	for i, f := range item1.Fields {
		if f.ID != expected[i] {
			t.Errorf("field[%d].id = %q, want %q", i, f.ID, expected[i])
		}
	}
}
