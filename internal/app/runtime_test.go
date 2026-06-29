package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/adapters/gopass"
)

func TestLoadSecretReaderSelectsGopass(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `version: 1
source:
  type: gopass
  gopass_command: gopass-test
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	reader, err := LoadSecretReader(configPath, "")
	if err != nil {
		t.Fatalf("load reader: %v", err)
	}
	gp, ok := reader.(gopass.Reader)
	if !ok {
		t.Fatalf("reader type = %T, want gopass.Reader", reader)
	}
	if gp.Binary != "gopass-test" {
		t.Fatalf("binary = %s, want gopass-test", gp.Binary)
	}
}

func TestLoadSecretReaderRejectsUnknownSource(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `version: 1
source:
  type: missing
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := LoadSecretReader(configPath, ""); err == nil {
		t.Fatalf("expected unknown source error")
	}
}
