package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretCreateSetGetListInfoRmV2(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
	out, err := runShelfWithInput(t, "fake-token", "--vault", data, "secret", "create", "github", "--env", "GITHUB_TOKEN", "--stdin")
	if err != nil {
		t.Fatalf("create: %v\n%s", err, out)
	}
	if !strings.Contains(out, "created item_github") {
		t.Fatalf("create output = %q", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "get", "github")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if out != "fake-token\n" {
		t.Fatalf("get = %q", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "list", "--long")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "item_github\tgithub\tgeneric_secret\tGITHUB_TOKEN\tset") {
		t.Fatalf("list output = %q", out)
	}
	if strings.Contains(out, "fake-token") {
		t.Fatalf("list leaked value: %q", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "info", "github")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	for _, want := range []string{"\"id\": \"item_github\"", "\"id\": \"value_env\"", "GITHUB_TOKEN", "\"value_set\": true"} {
		if !strings.Contains(out, want) {
			t.Fatalf("info missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "fake-token") {
		t.Fatalf("info leaked value: %s", out)
	}

	if _, err := runShelf(t, "--vault", data, "secret", "rm", "github"); err == nil {
		t.Fatalf("rm without --yes should fail")
	}
	if _, err := runShelf(t, "--vault", data, "secret", "rm", "github", "--yes"); err != nil {
		t.Fatalf("rm --yes: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "get", "github"); err == nil {
		t.Fatalf("get deleted item should fail")
	}
}

func TestSecretCreateProfileOpenAIV2(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
	out, err := runShelfWithInput(t, "fake-openai-key", "--vault", data, "secret", "create", "openai", "--profile", "openai", "--stdin")
	if err != nil {
		t.Fatalf("create profile: %v\n%s", err, out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "info", "openai")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	for _, want := range []string{"api_credential", "credential_env", "OPENAI_API_KEY", "provider", "OpenAI", "website", "https://platform.openai.com"} {
		if !strings.Contains(out, want) {
			t.Fatalf("info missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "fake-openai-key") {
		t.Fatalf("info leaked value: %s", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "get", "--env", "OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("get --env: %v", err)
	}
	if out != "fake-openai-key\n" {
		t.Fatalf("get --env = %q", out)
	}
}

func TestSecretSetRequiresStdinAndCanSetPublicField(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
	if _, err := runShelf(t, "--vault", data, "secret", "create", "app", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app"); err == nil {
		t.Fatalf("set without --stdin should fail")
	}
	if _, err := runShelfWithInput(t, "updated-token", "--vault", data, "secret", "set", "app", "--stdin"); err != nil {
		t.Fatalf("set primary: %v", err)
	}
	if _, err := runShelfWithInput(t, "APP_TOKEN_RENAMED", "--vault", data, "secret", "set", "app", "--field", "value_env", "--stdin"); err != nil {
		t.Fatalf("set env field: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "--env", "APP_TOKEN_RENAMED")
	if err != nil {
		t.Fatalf("get env renamed: %v", err)
	}
	if out != "updated-token\n" {
		t.Fatalf("get = %q", out)
	}
}

func TestSecretEnvSelectorAmbiguityIsValueFree(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
	if _, err := runShelfWithInput(t, "first-secret", "--vault", data, "secret", "create", "one", "--env", "DUPLICATE_ENV", "--stdin"); err != nil {
		t.Fatalf("create one: %v", err)
	}
	if _, err := runShelfWithInput(t, "second-secret", "--vault", data, "secret", "create", "two", "--env", "DUPLICATE_ENV", "--stdin"); err != nil {
		t.Fatalf("create two: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "--env", "DUPLICATE_ENV")
	if err == nil {
		t.Fatalf("expected ambiguous env error")
	}
	if !strings.Contains(err.Error(), "env selector matched multiple fields") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out+err.Error(), "first-secret") || strings.Contains(out+err.Error(), "second-secret") {
		t.Fatalf("ambiguity leaked secret values: out=%q err=%v", out, err)
	}
}

func TestSecretLegacyCommandsRemovedV2(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.v2.json")
	for _, subcommand := range []string{"add", "edit", "export"} {
		_, err := runShelf(t, "--vault", data, "secret", subcommand)
		if err == nil {
			t.Fatalf("secret %s should be removed in v2", subcommand)
		}
	}
}
