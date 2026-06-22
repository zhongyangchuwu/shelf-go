package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportFormatsReadEncryptedVault(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.age")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret-token", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set token: %v", err)
	}

	cases := []struct {
		name   string
		format string
		want   string
	}{
		{name: "shell", format: "shell", want: "export APP_TOKEN=secret-token\n"},
		{name: "env", format: "env", want: "APP_TOKEN=secret-token\n"},
		{name: "json", format: "json", want: "{\n  \"APP_TOKEN\": \"secret-token\"\n}\n"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runShelf(t, "--vault", data, "secret", "export", "app:token", "--format", tt.format)
			if err != nil {
				t.Fatalf("export %s: %v\n%s", tt.format, err, out)
			}
			if out != tt.want {
				t.Fatalf("unexpected %s output:\ngot  %q\nwant %q", tt.format, out, tt.want)
			}
		})
	}

	content, err := os.ReadFile(data)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if strings.Contains(string(content), "secret-token") || strings.Contains(string(content), "app:token") || strings.Contains(string(content), "APP_TOKEN") {
		t.Fatalf("encrypted vault contains exported plaintext data")
	}
}

func TestExportPrefixReadsEncryptedVault(t *testing.T) {
	data := filepath.Join(t.TempDir(), "vault.age")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app/api:token", "token-secret", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set token: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app/api:url", "https://example.test", "--env", "APP_URL"); err != nil {
		t.Fatalf("set url: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app/api:internal", "hidden"); err != nil {
		t.Fatalf("set internal: %v", err)
	}

	out, err := runShelf(t, "--vault", data, "secret", "export", "app/api", "--format", "env")
	if err != nil {
		t.Fatalf("export prefix: %v\n%s", err, out)
	}
	for _, want := range []string{"APP_TOKEN=token-secret", "APP_URL=https://example.test"} {
		if !strings.Contains(out, want) {
			t.Fatalf("prefix export missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "hidden") || strings.Contains(out, "APP_INTERNAL") {
		t.Fatalf("prefix export included secret without env by default:\n%s", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "export", "app/api", "--format", "env", "--all")
	if err != nil {
		t.Fatalf("export prefix all: %v\n%s", err, out)
	}
	if !strings.Contains(out, "APP_API_INTERNAL=hidden") {
		t.Fatalf("--all prefix export missing derived env binding:\n%s", out)
	}
}
