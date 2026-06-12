package manifest

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
			in:      Manifest{Version: 2, Secrets: []Entry{}},
			wantErr: "unsupported project manifest version",
		},
		{
			name:    "missing secrets array",
			in:      Manifest{Version: CurrentVersion},
			wantErr: "secrets array is required",
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
			name: "prefix not supported in v0.2",
			in: Manifest{Version: CurrentVersion, Secrets: []Entry{{
				Prefix: "providers/openai/accounts/personal",
			}}},
			wantErr: "prefix is not supported in v0.2",
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
