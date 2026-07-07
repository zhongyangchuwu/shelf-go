package vaultmodel_test

import (
	"encoding/json"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

// validGenericSecret returns a valid generic_secret item for use in tests.
func validGenericSecret() vaultmodel.Item {
	return vaultmodel.Item{
		ID:       "item_secret_001",
		Title:    "My Generic Secret",
		Template: vaultmodel.TemplateGenericSecret,
		Fields: []vaultmodel.Field{
			{
				ID:          "value",
				Label:       "Value",
				Type:        vaultmodel.FieldTypeConcealed,
				Sensitivity: vaultmodel.SensitivitySecret,
				Purpose:     vaultmodel.PurposePrimarySecret,
				Value:       json.RawMessage(`"fake-secret-value"`),
				Bindings: []vaultmodel.Binding{
					{Type: vaultmodel.BindingTypeEnv, Name: "MY_SECRET"},
				},
			},
		},
	}
}

// validAPICredential returns a valid api_credential item for use in tests.
func validAPICredential() vaultmodel.Item {
	return vaultmodel.Item{
		ID:       "item_api_001",
		Title:    "OpenAI Personal API Key",
		Template: vaultmodel.TemplateAPICredential,
		Labels:   []string{"ai", "personal"},
		Aliases:  []string{"openai-key"},
		Fields: []vaultmodel.Field{
			{
				ID:          "credential",
				Label:       "Credential",
				Type:        vaultmodel.FieldTypeConcealed,
				Sensitivity: vaultmodel.SensitivitySecret,
				Purpose:     vaultmodel.PurposePrimarySecret,
				Value:       json.RawMessage(`"sk-example"`),
				Bindings: []vaultmodel.Binding{
					{Type: vaultmodel.BindingTypeEnv, Name: "OPENAI_API_KEY"},
				},
			},
			{
				ID:          "provider",
				Label:       "Provider",
				Type:        vaultmodel.FieldTypeString,
				Sensitivity: vaultmodel.SensitivityPublic,
				Purpose:     vaultmodel.PurposeProvider,
				Value:       json.RawMessage(`"openai"`),
			},
			{
				ID:          "website",
				Label:       "Website",
				Type:        vaultmodel.FieldTypeURL,
				Sensitivity: vaultmodel.SensitivityPublic,
				Purpose:     vaultmodel.PurposeWebsite,
				Value:       json.RawMessage(`"https://platform.openai.com/api-keys"`),
			},
			{
				ID:          "expires_at",
				Label:       "Expires",
				Type:        vaultmodel.FieldTypeDate,
				Sensitivity: vaultmodel.SensitivityPublic,
				Purpose:     vaultmodel.PurposeExpiry,
				Value:       json.RawMessage(`"2026-12-31"`),
			},
			{
				ID:          "scopes",
				Label:       "Scopes",
				Type:        vaultmodel.FieldTypeStringList,
				Sensitivity: vaultmodel.SensitivityPublic,
				Purpose:     vaultmodel.PurposeScopes,
				Value:       json.RawMessage(`["responses","embeddings"]`),
			},
		},
	}
}

func TestValidate_ValidGenericSecret(t *testing.T) {
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{validGenericSecret()},
	}
	if err := vaultmodel.Validate(v); err != nil {
		t.Errorf("expected valid generic_secret, got: %v", err)
	}
}

func TestValidate_ValidAPICredential(t *testing.T) {
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{validAPICredential()},
	}
	if err := vaultmodel.Validate(v); err != nil {
		t.Errorf("expected valid api_credential, got: %v", err)
	}
}

func TestValidate_InvalidVersion(t *testing.T) {
	v := vaultmodel.Vault{
		Version: 99,
		Items:   []vaultmodel.Item{validGenericSecret()},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for unsupported version, got nil")
	} else if err.Error() != "unsupported vault version: 99" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_DuplicateItemID(t *testing.T) {
	item := validGenericSecret()
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item, item}, // same ID twice
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for duplicate item ID, got nil")
	} else if err.Error() != "duplicate item id: item_secret_001" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_EmptyItemID(t *testing.T) {
	item := validGenericSecret()
	item.ID = ""
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for empty item ID, got nil")
	}
}

func TestValidate_EmptyTitle(t *testing.T) {
	item := validGenericSecret()
	item.Title = ""
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
}

func TestValidate_UnknownTemplate(t *testing.T) {
	item := validGenericSecret()
	item.Template = "nonexistent_template"
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for unknown template, got nil")
	} else if err.Error() != "items[0].template unknown: nonexistent_template" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_MissingRequiredField(t *testing.T) {
	// generic_secret requires field "value"
	item := validGenericSecret()
	item.Fields = nil // drop all fields
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for missing required field, got nil")
	}
}

func TestValidate_DuplicateFieldID(t *testing.T) {
	item := validGenericSecret()
	// Add a second field with same ID "value"
	item.Fields = append(item.Fields, vaultmodel.Field{
		ID:          "value",
		Label:       "Duplicate",
		Type:        vaultmodel.FieldTypeConcealed,
		Sensitivity: vaultmodel.SensitivitySecret,
	})
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for duplicate field ID, got nil")
	}
}

func TestValidate_FieldTypeMismatch(t *testing.T) {
	// api_credential.field "provider" must be type "string"
	item := validAPICredential()
	for i := range item.Fields {
		if item.Fields[i].ID == "provider" {
			item.Fields[i].Type = vaultmodel.FieldTypeURL
			break
		}
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for field type mismatch, got nil")
	}
}

func TestValidate_FieldSensitivityMismatch(t *testing.T) {
	// api_credential.field "provider" must be sensitivity "public"
	item := validAPICredential()
	for i := range item.Fields {
		if item.Fields[i].ID == "provider" {
			item.Fields[i].Sensitivity = vaultmodel.SensitivitySecret
			break
		}
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for field sensitivity mismatch, got nil")
	}
}

func TestValidate_EnvBindingOnPublicField(t *testing.T) {
	item := validAPICredential()
	for i := range item.Fields {
		if item.Fields[i].ID == "provider" {
			item.Fields[i].Bindings = []vaultmodel.Binding{
				{Type: vaultmodel.BindingTypeEnv, Name: "PROVIDER_ENV"},
			}
			break
		}
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for env binding on public field, got nil")
	}
}

func TestValidate_InvalidEnvBindingName(t *testing.T) {
	item := validGenericSecret()
	item.Fields[0].Bindings = []vaultmodel.Binding{
		{Type: vaultmodel.BindingTypeEnv, Name: "1invalid"},
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for invalid env binding name, got nil")
	}
}

func TestValidate_DuplicateEnvBindingName(t *testing.T) {
	item1 := validGenericSecret()
	item2 := validGenericSecret()
	item2.ID = "item_secret_002"
	item2.Fields[0].Bindings = []vaultmodel.Binding{
		{Type: vaultmodel.BindingTypeEnv, Name: "MY_SECRET"}, // same name as item1
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item1, item2},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for duplicate env binding name, got nil")
	}
}

func TestValidate_InvalidURLValue(t *testing.T) {
	item := validAPICredential()
	for i := range item.Fields {
		if item.Fields[i].ID == "website" {
			item.Fields[i].Value = json.RawMessage(`"not-a-url"`)
			break
		}
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for invalid URL value, got nil")
	}
}

func TestValidate_InvalidDateValue(t *testing.T) {
	item := validAPICredential()
	for i := range item.Fields {
		if item.Fields[i].ID == "expires_at" {
			item.Fields[i].Value = json.RawMessage(`"not-a-date"`)
			break
		}
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for invalid date value, got nil")
	}
}

func TestValidate_InvalidStringListValue(t *testing.T) {
	item := validAPICredential()
	for i := range item.Fields {
		if item.Fields[i].ID == "scopes" {
			item.Fields[i].Value = json.RawMessage(`"not-an-array"`)
			break
		}
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for invalid string_list value, got nil")
	}
}

func TestValidate_StringListElementNotString(t *testing.T) {
	item := validAPICredential()
	for i := range item.Fields {
		if item.Fields[i].ID == "scopes" {
			item.Fields[i].Value = json.RawMessage(`["responses", 42]`)
			break
		}
	}
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item},
	}
	if err := vaultmodel.Validate(v); err == nil {
		t.Fatal("expected error for non-string element in string_list, got nil")
	}
}

// --- Index tests ---

func makeTestVault() vaultmodel.Vault {
	item1 := vaultmodel.Item{
		ID:       "id-alpha",
		Title:    "Alpha Secret",
		Template: vaultmodel.TemplateGenericSecret,
		Aliases:  []string{"alpha", "a-key"},
		Fields: []vaultmodel.Field{
			{
				ID: "value", Label: "Value",
				Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret,
				Purpose: vaultmodel.PurposePrimarySecret,
				Value:   json.RawMessage(`"alpha-value"`),
				Bindings: []vaultmodel.Binding{
					{Type: vaultmodel.BindingTypeEnv, Name: "ALPHA_KEY"},
				},
			},
		},
	}
	item2 := vaultmodel.Item{
		ID:       "id-beta",
		Title:    "Beta Secret",
		Template: vaultmodel.TemplateGenericSecret,
		Fields: []vaultmodel.Field{
			{
				ID: "value", Label: "Value",
				Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret,
				Purpose: vaultmodel.PurposePrimarySecret,
				Value:   json.RawMessage(`"beta-value"`),
				Bindings: []vaultmodel.Binding{
					{Type: vaultmodel.BindingTypeEnv, Name: "BETA_KEY"},
				},
			},
		},
	}
	return vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{item1, item2},
	}
}

func TestIndex_ItemByID(t *testing.T) {
	v := makeTestVault()
	idx, err := vaultmodel.NewIndex(v)
	if err != nil {
		t.Fatal(err)
	}

	item, ok := idx.ItemByID("id-alpha")
	if !ok {
		t.Fatal("expected to find item by ID")
	}
	if item.Title != "Alpha Secret" {
		t.Errorf("expected title 'Alpha Secret', got %q", item.Title)
	}

	if _, ok := idx.ItemByID("nonexistent"); ok {
		t.Error("expected false for nonexistent ID")
	}
}

func TestIndex_UniqueTitle(t *testing.T) {
	v := makeTestVault()
	idx, err := vaultmodel.NewIndex(v)
	if err != nil {
		t.Fatal(err)
	}

	item, err := idx.UniqueTitle("Beta Secret")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "id-beta" {
		t.Errorf("expected ID 'id-beta', got %q", item.ID)
	}

	if _, err := idx.UniqueTitle("Nonexistent"); err == nil {
		t.Error("expected error for nonexistent title")
	}
}

func TestIndex_ItemsByAlias(t *testing.T) {
	v := makeTestVault()
	idx, err := vaultmodel.NewIndex(v)
	if err != nil {
		t.Fatal(err)
	}

	items := idx.ItemsByAlias("alpha")
	if len(items) != 1 {
		t.Fatalf("expected 1 item by alias, got %d", len(items))
	}
	if items[0].ID != "id-alpha" {
		t.Errorf("expected ID 'id-alpha', got %q", items[0].ID)
	}

	if items := idx.ItemsByAlias("nonexistent"); len(items) != 0 {
		t.Error("expected empty for nonexistent alias")
	}
}

func TestIndex_FieldByEnv(t *testing.T) {
	v := makeTestVault()
	idx, err := vaultmodel.NewIndex(v)
	if err != nil {
		t.Fatal(err)
	}

	ref, ok := idx.FieldByEnv("ALPHA_KEY")
	if !ok {
		t.Fatal("expected to find field by env binding")
	}
	if ref.Item.ID != "id-alpha" {
		t.Errorf("expected item ID 'id-alpha', got %q", ref.Item.ID)
	}
	if ref.Field.ID != "value" {
		t.Errorf("expected field ID 'value', got %q", ref.Field.ID)
	}

	if _, ok := idx.FieldByEnv("NONEXISTENT"); ok {
		t.Error("expected false for nonexistent env binding")
	}
}

func TestIndex_ValidateError(t *testing.T) {
	v := vaultmodel.Vault{
		Version: vaultmodel.Version,
		Items:   []vaultmodel.Item{{ID: "", Title: "no template", Template: ""}},
	}
	if _, err := vaultmodel.NewIndex(v); err == nil {
		t.Error("expected error from NewIndex on invalid vault")
	}
}
