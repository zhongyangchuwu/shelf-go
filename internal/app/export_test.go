package app

import (
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func TestExportSecretsSelectsPrefixAndFiltersEnvByDefault(t *testing.T) {
	st := appTestStore(t, map[string]vault.Secret{
		"app/api:token":    {Value: appTestValue(t, "token-secret"), Env: "APP_TOKEN"},
		"app/api:url":      {Value: appTestValue(t, "https://example.test"), Env: "APP_URL"},
		"app/api:internal": {Value: appTestValue(t, "hidden")},
	})
	out, err := ExportSecrets(st, ExportRequest{Selector: "app/api", Format: "env"})
	if err != nil {
		t.Fatalf("export secrets: %v", err)
	}
	for _, want := range []string{"APP_TOKEN=token-secret", "APP_URL=https://example.test"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "hidden") {
		t.Fatalf("export included value without env:\n%s", out)
	}
}

func TestExportSecretsAllDerivesEnvForSecretsWithoutEnv(t *testing.T) {
	st := appTestStore(t, map[string]vault.Secret{
		"app/api:internal": {Value: appTestValue(t, "hidden")},
	})
	out, err := ExportSecrets(st, ExportRequest{Selector: "app/api", All: true, Format: "env"})
	if err != nil {
		t.Fatalf("export secrets: %v", err)
	}
	if !strings.Contains(out, "APP_API_INTERNAL=hidden") {
		t.Fatalf("missing derived env binding:\n%s", out)
	}
}

func TestExportSecretsSelectsTags(t *testing.T) {
	st := appTestStore(t, map[string]vault.Secret{
		"providers/openai:api_key":    {Value: appTestValue(t, "sk-openai"), Env: "OPENAI_API_KEY", Tags: []string{"ai", "prod"}},
		"providers/anthropic:api_key": {Value: appTestValue(t, "sk-anthropic"), Env: "ANTHROPIC_API_KEY", Tags: []string{"ai"}},
	})
	out, err := ExportSecrets(st, ExportRequest{Tags: []string{"ai", "prod"}, Format: "env"})
	if err != nil {
		t.Fatalf("export secrets: %v", err)
	}
	if !strings.Contains(out, "OPENAI_API_KEY=sk-openai") || strings.Contains(out, "ANTHROPIC") {
		t.Fatalf("unexpected tag export:\n%s", out)
	}
}

func TestExportSecretsRejectsMissingSelectorAndBadFormat(t *testing.T) {
	st := appTestStore(t, map[string]vault.Secret{})
	if _, err := ExportSecrets(st, ExportRequest{Format: "env"}); err == nil || !strings.Contains(err.Error(), "path, prefix, or --tag is required") {
		t.Fatalf("expected selector error, got %v", err)
	}
	if _, err := ExportSecrets(st, ExportRequest{Selector: "app", All: true, Format: "bad"}); err == nil || !strings.Contains(err.Error(), "no secrets matched: app") {
		t.Fatalf("expected no match before format error, got %v", err)
	}
	st = appTestStore(t, map[string]vault.Secret{"app:token": {Value: appTestValue(t, "secret"), Env: "APP_TOKEN"}})
	if _, err := ExportSecrets(st, ExportRequest{Selector: "app:token", Format: "bad"}); err == nil || !strings.Contains(err.Error(), "unsupported format: bad") {
		t.Fatalf("expected bad format error, got %v", err)
	}
}

func appTestStore(t *testing.T, secrets map[string]vault.Secret) *vault.Store {
	t.Helper()
	st := &vault.Store{Data: vault.NewData()}
	for path, secret := range secrets {
		if err := st.Set(path, secret, false); err != nil {
			t.Fatalf("set %s: %v", path, err)
		}
	}
	return st
}

func appTestValue(t *testing.T, value string) []byte {
	t.Helper()
	raw, err := vault.ParseValue(value)
	if err != nil {
		t.Fatalf("parse value: %v", err)
	}
	return raw
}
