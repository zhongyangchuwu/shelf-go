package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEntryIsRequiredDefaultsToTrue(t *testing.T) {
	entry := Entry{}
	if !entry.IsRequired() {
		t.Fatalf("expected default required=true")
	}
	falseValue := false
	entry.Required = &falseValue
	if entry.IsRequired() {
		t.Fatalf("expected required=false when explicitly set")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	required := false
	original := Manifest{
		Version: CurrentVersion,
		Secrets: []Entry{{
			Path:     "providers/openai/accounts/personal:api_key",
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
	if len(loaded.Secrets) != 1 {
		t.Fatalf("unexpected secrets length: %d", len(loaded.Secrets))
	}
	entry := loaded.Secrets[0]
	if entry.Path != original.Secrets[0].Path {
		t.Fatalf("unexpected path: %s", entry.Path)
	}
	if entry.Env != original.Secrets[0].Env {
		t.Fatalf("unexpected env: %s", entry.Env)
	}
	if entry.IsRequired() {
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
			in:      Manifest{Version: 0, Secrets: []Entry{}},
			wantErr: "unsupported project manifest version",
		},
		{
			name:    "missing secrets array",
			in:      Manifest{Version: CurrentVersion},
			wantErr: "secrets array is required",
		},
		{
			name:    "entry without selector",
			in:      Manifest{Version: CurrentVersion, Secrets: []Entry{{}}},
			wantErr: "path, prefix, or tags is required",
		},
		{
			name: "path and prefix both set",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{{
				Path:   "providers/openai/accounts/personal:api_key",
				Prefix: "providers/openai",
			}}},
			wantErr: "path, prefix, and tags are mutually exclusive",
		},
		{
			name: "invalid path",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{{
				Path: "bad",
			}}},
			wantErr: "invalid secrets[0].path",
		},
		{
			name: "invalid env",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{{
				Path: "providers/openai/accounts/personal:api_key",
				Env:  "1INVALID",
			}}},
			wantErr: "invalid secrets[0].env",
		},
		{
			name: "duplicate paths",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{
				{Path: "providers/openai/accounts/personal:api_key"},
				{Path: "providers/openai/accounts/personal:api_key"},
			}},
			wantErr: "duplicate secrets entry path",
		},
		{
			name: "duplicate prefix",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{
				{Prefix: "providers/openai"},
				{Prefix: "providers/openai"},
			}},
			wantErr: "duplicate secrets entry prefix",
		},
		{
			name: "prefix with colon",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{{
				Prefix: "providers/openai:api_key",
			}}},
			wantErr: "must not contain ':'",
		},
		{
			name: "prefix with empty segment",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{{
				Prefix: "providers//openai",
			}}},
			wantErr: "empty segment",
		},
		{
			name: "prefix with env",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{{
				Prefix: "providers/openai",
				Env:    "OPENAI_KEY",
			}}},
			wantErr: "prefix entries must not carry env",
		},
		{
			name: "prefix with invalid segment",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{{
				Prefix: "providers/open ai",
			}}},
			wantErr: "unsupported characters in segment",
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

func TestValidateAcceptsPrefixEntry(t *testing.T) {
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{{
		Prefix:   "providers/openai/accounts/personal",
		Required: boolPtr(false),
	}}}
	if err := Validate(m); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidateAcceptsMixedPathAndPrefix(t *testing.T) {
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{
		{Path: "providers/openai/accounts/personal:api_key", Env: "OPENAI_API_KEY"},
		{Prefix: "providers/openrouter/accounts/personal", Required: boolPtr(false)},
	}}
	if err := Validate(m); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	content := `{"version":1,"secrets":[{"path":"providers/openai/accounts/personal:api_key","value":"secret"}]}`
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

func boolPtr(b bool) *bool { return &b }

func TestV2BindingIsRequiredDefaultsToTrue(t *testing.T) {
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

func TestV2BindingKey(t *testing.T) {
	binding := V2Binding{Item: "github", Field: "credential"}
	if binding.Key() != "github.credential" {
		t.Fatalf("unexpected key: %s", binding.Key())
	}
}

func TestNewV2CreatesV2Manifest(t *testing.T) {
	m := NewV2()
	if !m.IsV2() {
		t.Fatalf("expected IsV2 to be true")
	}
	if m.Version != V2Version {
		t.Fatalf("expected version %d, got %d", V2Version, m.Version)
	}
	if m.Bindings == nil || len(m.Bindings) != 0 {
		t.Fatalf("expected empty bindings array")
	}
	if m.Secrets != nil {
		t.Fatalf("expected nil secrets")
	}
}

func TestValidateV2AcceptsValidManifest(t *testing.T) {
	m := NewV2()
	m.AddV2Binding(V2Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"})
	if err := Validate(m); err != nil {
		t.Fatalf("expected valid v2 manifest, got: %v", err)
	}
}

func TestValidateV2AcceptsMultipleBindings(t *testing.T) {
	m := NewV2()
	m.AddV2Binding(V2Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN", Required: boolPtr(true)})
	m.AddV2Binding(V2Binding{Item: "openai", Field: "credential", Env: "OPENAI_API_KEY", Title: "OpenAI Key"})
	m.AddV2Binding(V2Binding{Item: "slack", Field: "webhook", Env: "SLACK_WEBHOOK", Required: boolPtr(false)})
	if err := Validate(m); err != nil {
		t.Fatalf("expected valid v2 manifest, got: %v", err)
	}
}

func TestValidateV2RejectsMissingItem(t *testing.T) {
	m := Manifest{Version: V2Version, Bindings: []V2Binding{
		{Field: "credential", Env: "GITHUB_TOKEN"},
	}}
	err := Validate(m)
	if err == nil {
		t.Fatalf("expected error for missing item")
	}
	if !strings.Contains(err.Error(), "item is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateV2RejectsMissingField(t *testing.T) {
	m := Manifest{Version: V2Version, Bindings: []V2Binding{
		{Item: "github", Env: "GITHUB_TOKEN"},
	}}
	err := Validate(m)
	if err == nil {
		t.Fatalf("expected error for missing field")
	}
	if !strings.Contains(err.Error(), "field is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateV2RejectsMissingEnv(t *testing.T) {
	m := Manifest{Version: V2Version, Bindings: []V2Binding{
		{Item: "github", Field: "credential"},
	}}
	err := Validate(m)
	if err == nil {
		t.Fatalf("expected error for missing env")
	}
	if !strings.Contains(err.Error(), "env is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateV2RejectsInvalidEnv(t *testing.T) {
	m := Manifest{Version: V2Version, Bindings: []V2Binding{
		{Item: "github", Field: "credential", Env: "1INVALID"},
	}}
	err := Validate(m)
	if err == nil {
		t.Fatalf("expected error for invalid env")
	}
	if !strings.Contains(err.Error(), "invalid bindings[0].env") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateV2RejectsSecretsWithV2Version(t *testing.T) {
	m := Manifest{Version: V2Version, Secrets: []Entry{{Path: "app:token"}}}
	err := Validate(m)
	if err == nil {
		t.Fatalf("expected error for v2 with secrets")
	}
	if !strings.Contains(err.Error(), "secrets require version 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateV2RejectsBindingsWithV1Version(t *testing.T) {
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{}, Bindings: []V2Binding{{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"}}}
	err := Validate(m)
	if err == nil {
		t.Fatalf("expected error for v1 with bindings")
	}
	if !strings.Contains(err.Error(), "bindings require version 2") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateV2AcceptsOptionalTitle(t *testing.T) {
	m := NewV2()
	m.AddV2Binding(V2Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN", Title: "GitHub Personal Access Token"})
	if err := Validate(m); err != nil {
		t.Fatalf("expected valid v2 manifest with title, got: %v", err)
	}
}

func TestValidateV2AcceptsOptionalMissingBindings(t *testing.T) {
	f := false
	m := NewV2()
	m.AddV2Binding(V2Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN", Required: &f})
	if err := Validate(m); err != nil {
		t.Fatalf("expected valid v2 manifest with optional binding, got: %v", err)
	}
}

func TestSaveAndLoadV2RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	optional := false
	original := Manifest{
		Version: V2Version,
		Bindings: []V2Binding{
			{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"},
			{Item: "openai", Field: "credential", Env: "OPENAI_API_KEY", Required: &optional, Title: "OpenAI Key"},
		},
	}
	if err := Save(path, original); err != nil {
		t.Fatalf("save v2 manifest: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load v2 manifest: %v", err)
	}
	if loaded.Version != V2Version {
		t.Fatalf("unexpected version: %d", loaded.Version)
	}
	if len(loaded.Bindings) != 2 {
		t.Fatalf("unexpected bindings length: %d", len(loaded.Bindings))
	}
	if loaded.Bindings[0].Item != "github" || loaded.Bindings[0].Field != "credential" || loaded.Bindings[0].Env != "GITHUB_TOKEN" {
		t.Fatalf("unexpected binding[0]: %+v", loaded.Bindings[0])
	}
	if loaded.Bindings[1].Item != "openai" || loaded.Bindings[1].Env != "OPENAI_API_KEY" || loaded.Bindings[1].Title != "OpenAI Key" {
		t.Fatalf("unexpected binding[1]: %+v", loaded.Bindings[1])
	}
	if loaded.Bindings[1].IsRequired() {
		t.Fatalf("expected optional=false to round-trip")
	}
}

func TestSaveAndLoadV2RejectsVersion1WithBindings(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{}, Bindings: []V2Binding{{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"}}}
	err := Save(path, m)
	if err == nil {
		t.Fatalf("expected save to reject v1 with bindings")
	}
	if !strings.Contains(err.Error(), "bindings require version 2") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveAndLoadV2RejectsVersion2WithSecrets(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	m := Manifest{Version: V2Version, Secrets: []Entry{{Path: "app:token"}}}
	err := Save(path, m)
	if err == nil {
		t.Fatalf("expected save to reject v2 with secrets")
	}
	if !strings.Contains(err.Error(), "secrets require version 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}
