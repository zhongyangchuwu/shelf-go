package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBindingIsRequiredDefaultsToTrue(t *testing.T) {
	binding := V2Binding{}
	if !binding.IsRequired() {
		t.Fatalf("expected default required=true")
	}
	falseValue := false
	binding.Required = &falseValue
	if binding.IsRequired() {
		t.Fatalf("expected required=false when explicitly set")
	}
}

func TestBindingKey(t *testing.T) {
	binding := V2Binding{Item: "github", Field: "credential"}
	if binding.Key() != "github.credential" {
		t.Fatalf("unexpected key: %s", binding.Key())
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	required := false
	original := Manifest{
		Version: CurrentVersion,
		Bindings: []V2Binding{{
			Item:     "item_openai",
			Field:    "value",
			Env:      "OPENAI_API_KEY",
			Required: &required,
		}},
	}
	if err := Save(path, original); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if loaded.Version != CurrentVersion {
		t.Fatalf("unexpected version: %d", loaded.Version)
	}
	if len(loaded.Bindings) != 1 {
		t.Fatalf("unexpected bindings length: %d", len(loaded.Bindings))
	}
	binding := loaded.Bindings[0]
	if binding.Item != original.Bindings[0].Item || binding.Field != original.Bindings[0].Field || binding.Env != original.Bindings[0].Env {
		t.Fatalf("unexpected binding: %#v", binding)
	}
	if binding.IsRequired() {
		t.Fatalf("expected required=false to round-trip")
	}
}

func TestValidateRejectsInvalidManifestRules(t *testing.T) {
	tests := []struct {
		name    string
		in      Manifest
		wantErr string
	}{
		{
			name:    "unsupported version",
			in:      Manifest{Version: 2, Bindings: []V2Binding{}},
			wantErr: "unsupported project manifest version",
		},
		{
			name:    "missing bindings array",
			in:      Manifest{Version: CurrentVersion},
			wantErr: "bindings array is required",
		},
		{
			name:    "missing item",
			in:      Manifest{Version: CurrentVersion, Bindings: []V2Binding{{Field: "value", Env: "APP_TOKEN"}}},
			wantErr: "item is required",
		},
		{
			name:    "missing field",
			in:      Manifest{Version: CurrentVersion, Bindings: []V2Binding{{Item: "item_app", Env: "APP_TOKEN"}}},
			wantErr: "field is required",
		},
		{
			name:    "invalid env",
			in:      Manifest{Version: CurrentVersion, Bindings: []V2Binding{{Item: "item_app", Field: "value", Env: "1INVALID"}}},
			wantErr: "invalid bindings[0].env",
		},
		{
			name: "duplicate item field",
			in: Manifest{Version: CurrentVersion, Bindings: []V2Binding{
				{Item: "item_app", Field: "value", Env: "APP_TOKEN"},
				{Item: "item_app", Field: "value", Env: "APP_TOKEN_2"},
			}},
			wantErr: "duplicate binding item/field",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	content := `{"version":1,"bindings":[{"item":"item_app","field":"value","env":"APP_TOKEN","value":"secret"}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected unknown field error")
	}
	if !strings.Contains(err.Error(), "unknown field \"value\"") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCreatesItemManifest(t *testing.T) {
	m := New()
	if !m.IsV2() {
		t.Fatalf("expected IsV2 to be true")
	}
	if m.Version != CurrentVersion {
		t.Fatalf("expected version %d, got %d", CurrentVersion, m.Version)
	}
	if len(m.Bindings) != 0 {
		t.Fatalf("expected empty bindings")
	}
}

func boolPtr(b bool) *bool { return &b }
