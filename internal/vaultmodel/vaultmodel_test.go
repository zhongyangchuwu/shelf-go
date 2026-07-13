package vaultmodel_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func raw(value string) json.RawMessage { return json.RawMessage(value) }

func mustJSON(value string) json.RawMessage {
	content, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return content
}

func validGenericSecret(id, title, env string) vaultmodel.Item {
	return vaultmodel.Item{
		ID: id, Title: title, Template: vaultmodel.TemplateGenericSecret,
		Fields: []vaultmodel.Field{
			{ID: "value", Value: raw(`"fake-secret-value"`)},
			{ID: vaultmodel.EnvFieldID("value"), Value: mustJSON(env)},
		},
	}
}

func validAPICredential() vaultmodel.Item {
	return vaultmodel.Item{
		ID: "item_api_001", Title: "OpenAI Personal API Key", Template: vaultmodel.TemplateAPICredential,
		Labels: []string{"ai", "personal"}, Aliases: []string{"openai-key"},
		Fields: []vaultmodel.Field{
			{ID: "credential", Value: raw(`"sk-example"`)},
			{ID: vaultmodel.EnvFieldID("credential"), Value: raw(`"OPENAI_API_KEY"`)},
			{ID: vaultmodel.ValidityFieldID("credential"), Value: raw(`"2026-12-31"`)},
			{ID: "base_url", Value: raw(`"https://platform.openai.com"`)},
			{ID: vaultmodel.EnvFieldID("base_url"), Value: raw(`"OPENAI_BASE_URL"`)},
			{ID: "provider", Value: raw(`"OpenAI"`)},
			{ID: "website", Value: raw(`"https://platform.openai.com"`)},
			{ID: "scopes", Value: raw(`["responses","embeddings"]`)},
		},
	}
}

func TestValidateV3BuiltIns(t *testing.T) {
	vault := vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{
		validGenericSecret("item_secret_001", "My Generic Secret", "MY_SECRET"),
		validAPICredential(),
	}}
	if err := vaultmodel.Validate(vault); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateRejectsV2Version(t *testing.T) {
	vault := vaultmodel.Vault{Version: 2, Items: []vaultmodel.Item{validGenericSecret("item_secret_001", "Secret", "MY_SECRET")}}
	if err := vaultmodel.Validate(vault); err == nil || err.Error() != "unsupported vault version: 2" {
		t.Fatalf("Validate error = %v", err)
	}
}

func TestValidateAtomicFieldShape(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*vaultmodel.Item)
		wantErr string
	}{
		{name: "missing required", mutate: func(item *vaultmodel.Item) { item.Fields = nil }, wantErr: "missing required field: value"},
		{name: "duplicate field", mutate: func(item *vaultmodel.Item) { item.Fields = append(item.Fields, item.Fields[0]) }, wantErr: "duplicate field id: value"},
		{name: "unknown field", mutate: func(item *vaultmodel.Item) { item.Fields = append(item.Fields, vaultmodel.Field{ID: "unexpected"}) }, wantErr: "id unknown"},
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

func TestValidateEnvironmentAndValidityConventions(t *testing.T) {
	tests := []struct {
		name     string
		template vaultmodel.Template
		wantErr  string
	}{
		{
			name: "env target can be public URL",
			template: vaultmodel.Template{ID: "endpoint", Name: "Endpoint", PrimarySecret: "credential", Fields: []vaultmodel.FieldTemplate{
				{ID: "credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret},
				{ID: vaultmodel.EnvFieldID("credential"), Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
				{ID: "base_url", Type: vaultmodel.FieldTypeURL, Sensitivity: vaultmodel.SensitivityPublic},
				{ID: vaultmodel.EnvFieldID("base_url"), Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
			}},
		},
		{
			name: "env target is non-projectable",
			template: vaultmodel.Template{ID: "bad_env", Name: "Bad Env", PrimarySecret: "credential", Fields: []vaultmodel.FieldTemplate{
				{ID: "credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret},
				{ID: "scopes", Type: vaultmodel.FieldTypeStringList, Sensitivity: vaultmodel.SensitivityPublic},
				{ID: vaultmodel.EnvFieldID("scopes"), Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
			}},
			wantErr: "not projectable",
		},
		{
			name: "validity target is public",
			template: vaultmodel.Template{ID: "bad_validity", Name: "Bad Validity", PrimarySecret: "credential", Fields: []vaultmodel.FieldTemplate{
				{ID: "credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret},
				{ID: "provider", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
				{ID: vaultmodel.ValidityFieldID("provider"), Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Required: true},
			}},
			wantErr: "validity target must be secret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.Version, Templates: []vaultmodel.Template{tt.template}})
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("Validate error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRejectsDanglingStructuralFieldInstances(t *testing.T) {
	tests := []struct {
		name     string
		template vaultmodel.Template
		field    vaultmodel.Field
		wantErr  string
	}{
		{
			name: "environment mapping", wantErr: "env target missing: credential",
			template: vaultmodel.Template{ID: "optional_credential", Name: "Optional Credential", PrimarySecret: "credential", Fields: []vaultmodel.FieldTemplate{
				{ID: "credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret},
				{ID: vaultmodel.EnvFieldID("credential"), Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
			}},
			field: vaultmodel.Field{ID: vaultmodel.EnvFieldID("credential"), Value: mustJSON("")},
		},
		{
			name: "validity mapping", wantErr: "validity target missing: credential",
			template: vaultmodel.Template{ID: "tracked_credential", Name: "Tracked Credential", PrimarySecret: "credential", Fields: []vaultmodel.FieldTemplate{
				{ID: "credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret},
				{ID: vaultmodel.ValidityFieldID("credential"), Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Required: true},
			}},
			field: vaultmodel.Field{ID: vaultmodel.ValidityFieldID("credential"), Value: mustJSON("unknown")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault := vaultmodel.Vault{Version: vaultmodel.Version, Templates: []vaultmodel.Template{tt.template}, Items: []vaultmodel.Item{{
				ID: "item", Title: "Item", Template: tt.template.ID, Fields: []vaultmodel.Field{tt.field},
			}}}
			if err := vaultmodel.Validate(vault); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateValidityValues(t *testing.T) {
	for _, validity := range []string{"unknown", "permanent", "2027-03-01"} {
		t.Run(validity, func(t *testing.T) {
			item := validAPICredential()
			for i := range item.Fields {
				if item.Fields[i].ID == vaultmodel.ValidityFieldID("credential") {
					item.Fields[i].Value = mustJSON(validity)
				}
			}
			if err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item}}); err != nil {
				t.Fatalf("Validate %s: %v", validity, err)
			}
		})
	}
	item := validAPICredential()
	for i := range item.Fields {
		if item.Fields[i].ID == vaultmodel.ValidityFieldID("credential") {
			item.Fields[i].Value = mustJSON("forever")
		}
	}
	if err := vaultmodel.Validate(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item}}); err == nil || !strings.Contains(err.Error(), "unknown, permanent, or YYYY-MM-DD") {
		t.Fatalf("validity error = %v", err)
	}
}

func TestIndexResolvesPublicAndSecretEnvironmentFields(t *testing.T) {
	item := validAPICredential()
	vault := vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item}}
	idx, err := vaultmodel.NewIndex(vault)
	if err != nil {
		t.Fatal(err)
	}
	for _, env := range []string{"OPENAI_API_KEY", "OPENAI_BASE_URL"} {
		refs := idx.FieldsByEnv(env)
		if len(refs) != 1 {
			t.Fatalf("FieldsByEnv(%s) = %d", env, len(refs))
		}
	}
}

func TestRegistryPrimarySecretLookup(t *testing.T) {
	registry := vaultmodel.NewDefaultRegistry()
	primary, err := registry.PrimarySecretField(vaultmodel.TemplateAPICredential)
	if err != nil || primary.ID != "credential" {
		t.Fatalf("primary = %#v, err = %v", primary, err)
	}
}
