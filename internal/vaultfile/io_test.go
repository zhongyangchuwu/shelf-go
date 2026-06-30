package vaultfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRejectsUnknownFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "top level",
			content: `{"version":1,"secrets":{},"extra":true}`,
		},
		{
			name:    "secret field",
			content: `{"version":1,"secrets":{"app:token":{"value":"secret","source":"manual"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "secrets.json")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("write store: %v", err)
			}
			_, err := Load(path)
			if err == nil {
				t.Fatalf("expected unknown field to fail")
			}
			if !strings.Contains(err.Error(), "unknown field") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadAcceptsValidStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"secrets":{"app:token":{"value":"secret"}}}`), 0o600); err != nil {
		t.Fatalf("write store: %v", err)
	}
	st, err := Load(path)
	if err != nil {
		t.Fatalf("load valid store: %v", err)
	}
	if _, ok := st.Get("app:token"); !ok {
		t.Fatalf("missing loaded secret")
	}
}
