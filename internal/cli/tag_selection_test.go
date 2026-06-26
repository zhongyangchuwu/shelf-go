package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretListFiltersByTags(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.age")
	setTaggedSecrets(t, data)

	out, err := runShelf(t, "--vault", data, "secret", "list", "--tag", "ai")
	if err != nil {
		t.Fatalf("list by tag: %v\n%s", err, out)
	}
	want := "app:hidden\napp:token\napp:url\nops:deploy\n"
	if out != want {
		t.Fatalf("tag list output:\ngot  %q\nwant %q", out, want)
	}
	if strings.Contains(out, "secret-token") || strings.Contains(out, "deploy-secret") {
		t.Fatalf("tag list leaked secret value: %q", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "list", "app", "--tag", "ai", "--tag", "prod")
	if err != nil {
		t.Fatalf("list by two tags: %v\n%s", err, out)
	}
	if out != "app:token\n" {
		t.Fatalf("AND tag list output = %q, want app:token", out)
	}
}

func TestSecretExportFiltersByTag(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.age")
	setTaggedSecrets(t, data)

	out, err := runShelf(t, "--vault", data, "secret", "export", "--tag", "ai", "--format", "env")
	if err != nil {
		t.Fatalf("export by tag: %v\n%s", err, out)
	}
	for _, want := range []string{"APP_TOKEN=secret-token", "APP_URL=https://example.test", "DEPLOY_KEY=deploy-secret"} {
		if !strings.Contains(out, want) {
			t.Fatalf("tag export missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "hidden-value") || strings.Contains(out, "APP_HIDDEN") {
		t.Fatalf("tag export without --all included secret without env:\n%s", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "export", "--tag", "ai", "--format", "env", "--all")
	if err != nil {
		t.Fatalf("export by tag all: %v\n%s", err, out)
	}
	if !strings.Contains(out, "APP_HIDDEN=hidden-value") {
		t.Fatalf("tag export --all missing derived env binding:\n%s", out)
	}
}

func TestSecretExportCombinesPrefixAndTagsWithAndSemantics(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.age")
	setTaggedSecrets(t, data)

	out, err := runShelf(t, "--vault", data, "secret", "export", "app", "--tag", "ai", "--tag", "prod", "--format", "json")
	if err != nil {
		t.Fatalf("export prefix by tags: %v\n%s", err, out)
	}
	if !strings.Contains(out, `"APP_TOKEN": "secret-token"`) {
		t.Fatalf("AND tag export missing token:\n%s", out)
	}
	for _, forbidden := range []string{"APP_URL", "DEPLOY_KEY", "OPS_PASSWORD", "hidden-value"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("AND tag export included %q:\n%s", forbidden, out)
		}
	}
}

func TestSecretExportRequiresPathPrefixOrTag(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.age")
	out, err := runShelf(t, "--vault", data, "secret", "export")
	if err == nil {
		t.Fatalf("expected export without selector to fail: %q", out)
	}
	if !strings.Contains(err.Error(), "path, prefix, or --tag is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func setTaggedSecrets(t *testing.T, data string) {
	t.Helper()
	commands := [][]string{
		{"secret", "set", "app:token", "secret-token", "--env", "APP_TOKEN", "--tag", "ai", "--tag", "prod"},
		{"secret", "set", "app:url", "https://example.test", "--env", "APP_URL", "--tag", "ai"},
		{"secret", "set", "app:password", "password-secret", "--env", "APP_PASSWORD", "--tag", "prod"},
		{"secret", "set", "app:hidden", "hidden-value", "--tag", "ai"},
		{"secret", "set", "ops:deploy", "deploy-secret", "--env", "DEPLOY_KEY", "--tag", "ai", "--tag", "prod"},
	}
	for _, args := range commands {
		fullArgs := append([]string{"--vault", data}, args...)
		if out, err := runShelf(t, fullArgs...); err != nil {
			t.Fatalf("set tagged secret %v: %v\n%s", args, err, out)
		}
	}
}
