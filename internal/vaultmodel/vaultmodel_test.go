package vaultmodel_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func raw(value string) json.RawMessage { return json.RawMessage(value) }

func validGenericSecret(id, title, env string) vaultmodel.Item {
	fields := []vaultmodel.Field{
		{
			ID:          "value",
			Label:       "Value",
			Type:        vaultmodel.FieldTypeConcealed,
			Sensitivity: vaultmodel.SensitivitySecret,
			Purpose:     vaultmodel.PurposePrimarySecret,
			Value:       raw(`"fake-secret-value"`),
		},
	}
	if env != "" {
		fields = append(fields, vaultmodel.Field{
			ID:          "value_env",
			Label:       "Value Env",
			Type:        vaultmodel.FieldTypeString,
			Sensitivity: vaultmodel.SensitivityPublic,
			Purpose:     vaultmodel.PurposeEnvName,
			Value:       mustJSON(env),
		})
	}
	return vaultmodel.Item{ID: id, Title: title, Template: vaultmodel.TemplateGenericSecret, Fields: fields}
}

func validAPICredential() vaultmodel.Item {
	return vaultmodel.Item{
		ID:       "item_api_001",
		Title:    "OpenAI Personal API Key",
		Template: vaultmodel.TemplateAPICredential,
		Labels:   []string{"ai", "personal"},
		Aliases:  []string{"openai-key"},
		Fields: []vaultmodel.Field{
			{ID: "credential", Label: "Credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Purpose: vaultmodel.PurposePrimarySecret, Value: raw(`"sk-example"`)},
			{ID: "credential_env", Label: "Credential Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeEnvName, Value: raw(`"OPENAI_API_KEY"`)},
			{ID: "provider", Label: "Provider", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeProvider, Value: raw(`"OpenAI"`)},
			{ID: "website", Label: "Website", Type: vaultmodel.FieldTypeURL, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeWebsite, Value: raw(`"https://platform.openai.com"`)},
			{ID: "expires_at", Label: "Expires", Type: vaultmodel.FieldTypeDate, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeExpiry, Value: raw(`"2026-12-31"`)},
			{ID: "scopes", Label: "Scopes", Type: vaultmodel.FieldTypeStringList, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeScopes, Value: raw(`["responses","embeddings"]`)},
		},
	}
}

func mustJSON(value string) json.RawMessage {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return b
}

func TestValidate_ValidBuiltIns(t *testing.T) {
	v := vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{
		validGenericSecret("item_secret_001", "My Generic Secret", "MY_SECRET"),
		validAPICredential(),
	}}
	if err := vaultmodel.Validate(v); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidate_InvalidVersion(t *testing.T) {
	v := vaultmodel.Vault{Version: 99, Items: []vaultmodel.Item{validGenericSecret("item_secret_001", "Secret", "MY_SECRET")}}
	err := vaultmodel.Validate(v)
	if err == nil || err.Error() != "unsupported vault version: 99" {
		t.Fatalf("Validate error = %v", err)
	}
}

func TestValidate_DuplicateItemID(t *testing.T) {
	item := validGenericSecret("item_secret_001", "Secret", "MY_SECRET")
	v := vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item, item}}
	err := vaultmodel.Validate(v)
	if err == nil || err.Error() != "duplicate item id: item_secret_001" {
		t.Fatalf("Validate error = %v", err)
	}
}

func TestValidate_RequiredFieldsAndFieldIdentity(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*vaultmodel.Item)
		wantErr string
	}{
		{name: "missing required", mutate: func(item *vaultmodel.Item) { item.Fields = nil }, wantErr: "missing required field: value"},
		{name: "duplicate field", mutate: func(item *vaultmodel.Item) { item.Fields = append(item.Fields, item.Fields[0]) }, wantErr: "duplicate field id: value"},
		{name: "type mismatch", mutate: func(item *vaultmodel.Item) { item.Fields[0].Type = vaultmodel.FieldTypeURL }, wantErr: "type for value"},
		{name: "sensitivity mismatch", mutate: func(item *vaultmodel.Item) { item.Fields[0].Sensitivity = vaultmodel.SensitivityPublic }, wantErr: "sensitivity for value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := validGenericSecret("item_secret_001", "Secret", "MY_SECRET")
			tt.mutate(&item)
			err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item}})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_FieldValues(t *testing.T) {
	tests := []struct {
		name    string
		fieldID string
		value   json.RawMessage
		wantErr string
	}{
		{name: "invalid url", fieldID: "website", value: raw(`"not-a-url"`), wantErr: "invalid url"},
		{name: "invalid date", fieldID: "expires_at", value: raw(`"not-a-date"`), wantErr: "invalid date"},
		{name: "invalid string list", fieldID: "scopes", value: raw(`"not-an-array"`), wantErr: "value must be string array"},
		{name: "non string element", fieldID: "scopes", value: raw(`["responses",42]`), wantErr: "value must be string array"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := validAPICredential()
			for i := range item.Fields {
				if item.Fields[i].ID == tt.fieldID {
					item.Fields[i].Value = tt.value
				}
			}
			err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item}})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_EnvNameField(t *testing.T) {
	item := validGenericSecret("item_secret_001", "Secret", "1INVALID")
	err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item}})
	if err == nil || !strings.Contains(err.Error(), "env name invalid") {
		t.Fatalf("Validate error = %v, want invalid env name", err)
	}
}

func TestValidate_DuplicateEnvNamesAllowed(t *testing.T) {
	v := vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{
		validGenericSecret("item_one", "One", "DUPLICATE_ENV"),
		validGenericSecret("item_two", "Two", "DUPLICATE_ENV"),
	}}
	if err := vaultmodel.Validate(v); err != nil {
		t.Fatalf("duplicate env-name field values should be valid vault data: %v", err)
	}
}

func TestValidate_CustomEnvTemplate(t *testing.T) {
	bad := vaultmodel.Template{ID: "bad", Name: "Bad", Fields: []vaultmodel.FieldTemplate{
		{ID: "name", Label: "Name", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
		{ID: "name_env", Label: "Name Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Purpose: vaultmodel.PurposeEnvName, EnvFor: "name"},
	}}
	err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.Version, Templates: []vaultmodel.Template{bad}})
	if err == nil || !strings.Contains(err.Error(), "env_for target must be secret") {
		t.Fatalf("Validate error = %v, want public env_for target rejection", err)
	}
}

func TestRegistry_ProfileLookup(t *testing.T) {
	registry := vaultmodel.NewDefaultRegistry()
	profile, ok := registry.Profile(vaultmodel.ProfileOpenAI)
	if !ok {
		t.Fatal("missing openai profile")
	}
	if profile.Template != vaultmodel.TemplateAPICredential {
		t.Fatalf("profile template = %q", profile.Template)
	}
	var env string
	if err := json.Unmarshal(profile.Defaults["credential_env"], &env); err != nil {
		t.Fatalf("credential_env default: %v", err)
	}
	if env != "OPENAI_API_KEY" {
		t.Fatalf("credential_env = %q", env)
	}
	if _, ok := registry.Profile("openai"); !ok {
		t.Fatal("openai alias did not resolve")
	}
}

func TestValidate_ProfileDefaults(t *testing.T) {
	tests := []struct {
		name    string
		profile vaultmodel.Profile
		wantErr string
	}{
		{name: "unknown field", profile: vaultmodel.Profile{ID: "bad", Name: "Bad", Template: vaultmodel.TemplateAPICredential, Defaults: map[string]json.RawMessage{"missing": raw(`"x"`)}}, wantErr: "field unknown"},
		{name: "secret default", profile: vaultmodel.Profile{ID: "bad", Name: "Bad", Template: vaultmodel.TemplateAPICredential, Defaults: map[string]json.RawMessage{"credential": raw(`"secret"`)}}, wantErr: "must not set secret field defaults"},
		{name: "invalid env", profile: vaultmodel.Profile{ID: "bad", Name: "Bad", Template: vaultmodel.TemplateAPICredential, Defaults: map[string]json.RawMessage{"credential_env": raw(`"1INVALID"`)}}, wantErr: "env name invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.Version, Profiles: []vaultmodel.Profile{tt.profile}})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	v := vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{
		func() vaultmodel.Item {
			item := validGenericSecret("id-alpha", "Alpha Secret", "SHARED_ENV")
			item.Aliases = []string{"alpha", "a-key"}
			return item
		}(),
		validGenericSecret("id-beta", "Beta Secret", "SHARED_ENV"),
	}}
	idx, err := vaultmodel.NewIndex(v)
	if err != nil {
		t.Fatal(err)
	}
	if item, ok := idx.ItemByID("id-alpha"); !ok || item.Title != "Alpha Secret" {
		t.Fatalf("ItemByID = %+v, %v", item, ok)
	}
	if item, err := idx.UniqueTitle("Beta Secret"); err != nil || item.ID != "id-beta" {
		t.Fatalf("UniqueTitle = %+v, %v", item, err)
	}
	if items := idx.ItemsByAlias("alpha"); len(items) != 1 || items[0].ID != "id-alpha" {
		t.Fatalf("ItemsByAlias = %+v", items)
	}
	refs := idx.FieldsByEnv("SHARED_ENV")
	if len(refs) != 2 {
		t.Fatalf("FieldsByEnv len = %d, want 2", len(refs))
	}
	if refs[0].Field.ID != "value" || refs[0].EnvField.ID != "value_env" {
		t.Fatalf("FieldsByEnv[0] = %+v", refs[0])
	}
}

func TestIndex_ValidateError(t *testing.T) {
	_, err := vaultmodel.NewIndex(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{{ID: "", Title: "no template", Template: ""}}})
	if err == nil {
		t.Fatal("expected error from NewIndex on invalid vault")
	}
}
