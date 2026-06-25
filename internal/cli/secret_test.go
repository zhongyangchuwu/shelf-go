package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"filippo.io/age"
	"github.com/spf13/cobra"
)

func TestSecretSetGetListInfoExport(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")

	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openrouter/accounts/personal:api_key", "sk-xxx", "--env", "OPENROUTER_API_KEY", "--tag", "ai"); err != nil {
		t.Fatalf("set secret: %v", err)
	}

	out, err := runShelf(t, "--vault", data, "secret", "get", "providers/openrouter/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if out != "sk-xxx\n" {
		t.Fatalf("unexpected get output: %q", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "list", "providers/openrouter")
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	if out != "providers/openrouter/accounts/personal:api_key\n" {
		t.Fatalf("unexpected list output: %q", out)
	}

	out, err = runShelf(t, "--vault", data, "secret", "info", "providers/openrouter/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("info secret: %v", err)
	}
	if strings.Contains(out, "sk-xxx") {
		t.Fatalf("info leaked value: %s", out)
	}
	for _, want := range []string{"\"path\"", "\"group_path\": \"providers/openrouter/accounts/personal\"", "\"key\": \"api_key\"", "\"value_set\": true", "\"env\": \"OPENROUTER_API_KEY\"", "\"tags\""} {
		if !strings.Contains(out, want) {
			t.Fatalf("info output missing %q: %s", want, out)
		}
	}

	out, err = runShelf(t, "--vault", data, "secret", "export", "providers/openrouter/accounts/personal:api_key", "--format", "shell")
	if err != nil {
		t.Fatalf("export shell: %v", err)
	}
	if out != "export OPENROUTER_API_KEY=sk-xxx\n" {
		t.Fatalf("unexpected export output: %q", out)
	}
}
func TestSecretCommandsUseEncryptedVaultConfig(t *testing.T) {
	dir := t.TempDir()
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate identity: %v", err)
	}
	identityPath := filepath.Join(dir, "identity.txt")
	if err := os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0o600); err != nil {
		t.Fatalf("write identity: %v", err)
	}
	vaultPath := filepath.Join(dir, "secrets.vault")
	configPath := filepath.Join(dir, "config.yaml")
	config := fmt.Sprintf("version: 1\nvault_path: %s\nrecipients:\n  - %s\nidentity_paths:\n  - %s\n", vaultPath, identity.Recipient().String(), identityPath)
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := runShelf(t, "--config", configPath, "secret", "set", "app:token", "vault-secret", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set vault secret: %v", err)
	}
	content, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if strings.Contains(string(content), "vault-secret") || strings.Contains(string(content), "app:token") {
		t.Fatalf("vault leaked plaintext: %s", content)
	}
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{name: "get", args: []string{"secret", "get", "app:token"}, want: "vault-secret\n"},
		{name: "list", args: []string{"secret", "list"}, want: "app:token\n"},
		{name: "info", args: []string{"secret", "info", "app:token"}, want: "\"env\": \"APP_TOKEN\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"--config", configPath}, tc.args...)
			out, err := runShelf(t, args...)
			if err != nil {
				t.Fatalf("run %s: %v", tc.name, err)
			}
			if !strings.Contains(out, tc.want) {
				t.Fatalf("output missing %q: %s", tc.want, out)
			}
		})
	}

	if _, err := runShelf(t, "--config", configPath, "secret", "set", "app:token", "updated-secret", "--force"); err != nil {
		t.Fatalf("update vault secret: %v", err)
	}
	if _, err := runShelf(t, "--config", configPath, "secret", "rm", "app:token"); err != nil {
		t.Fatalf("remove vault secret: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "secret", "list")
	if err != nil {
		t.Fatalf("list after remove: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty list after remove, got %q", out)
	}
}

func TestSecretAddInteractiveFullPath(t *testing.T) {
	withPromptPassword(t, "sk-test")
	data := filepath.Join(t.TempDir(), "secrets.json")
	out, err := runShelfWithInput(t, "OPENAI_API_KEY\nOpenAI key\nai,personal\n", "--vault", data, "secret", "add", "providers/openai/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("secret add: %v\n%s", err, out)
	}
	if strings.Contains(out, "sk-test") {
		t.Fatalf("secret add leaked value:\n%s", out)
	}
	got, err := runShelf(t, "--vault", data, "secret", "get", "providers/openai/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if got != "sk-test\n" {
		t.Fatalf("unexpected secret value: %q", got)
	}
	info, err := runShelf(t, "--vault", data, "secret", "info", "providers/openai/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("info secret: %v", err)
	}
	for _, want := range []string{`"env": "OPENAI_API_KEY"`, `"description": "OpenAI key"`, `"ai"`, `"personal"`} {
		if !strings.Contains(info, want) {
			t.Fatalf("info missing %q:\n%s", want, info)
		}
	}
}
func TestSecretAddInteractiveGroupPathShowsExistingGroups(t *testing.T) {
	withPromptPassword(t, "sk-work")
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai/accounts/personal:api_key", "sk-personal"); err != nil {
		t.Fatalf("set seed: %v", err)
	}
	out, err := runShelfWithInput(t, "api_key\n\n\n\n", "--vault", data, "secret", "add", "providers/openai/accounts/work")
	if err != nil {
		t.Fatalf("secret add: %v\n%s", err, out)
	}
	if !strings.Contains(out, "existing groups:") || !strings.Contains(out, "providers/openai/accounts/personal") {
		t.Fatalf("missing existing group hint:\n%s", out)
	}
	got, err := runShelf(t, "--vault", data, "secret", "get", "providers/openai/accounts/work:api_key")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if got != "sk-work\n" {
		t.Fatalf("unexpected secret value: %q", got)
	}
}
func TestSecretAddRefusesOverwriteByDefault(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "old"); err != nil {
		t.Fatalf("set seed: %v", err)
	}
	origIsTerminal := secretAddIsTerminal
	origReadPassword := secretAddReadPassword
	secretAddIsTerminal = func(int) bool { return true }
	secretAddReadPassword = func(int) ([]byte, error) { return nil, fmt.Errorf("password prompt should not run") }
	t.Cleanup(func() {
		secretAddIsTerminal = origIsTerminal
		secretAddReadPassword = origReadPassword
	})
	_, err := runShelfWithInput(t, "n\n", "--vault", data, "secret", "add", "app:token")
	if err == nil {
		t.Fatalf("expected overwrite refusal")
	}
	got, err := runShelf(t, "--vault", data, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if got != "old\n" {
		t.Fatalf("secret was overwritten: %q", got)
	}
}
func TestSecretAddOverwritesWhenConfirmed(t *testing.T) {
	withPromptPassword(t, "new")
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "old"); err != nil {
		t.Fatalf("set seed: %v", err)
	}
	out, err := runShelfWithInput(t, "y\n\n\n\n", "--vault", data, "secret", "add", "app:token")
	if err != nil {
		t.Fatalf("secret add overwrite: %v\n%s", err, out)
	}
	got, err := runShelf(t, "--vault", data, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if got != "new\n" {
		t.Fatalf("secret was not overwritten: %q", got)
	}
}
func TestSecretAddFailsWithoutTerminal(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	origIsTerminal := secretAddIsTerminal
	secretAddIsTerminal = func(int) bool { return false }
	t.Cleanup(func() { secretAddIsTerminal = origIsTerminal })
	_, err := runShelfWithInput(t, "", "--vault", data, "secret", "add", "app:token")
	if err == nil {
		t.Fatalf("expected non-terminal add to fail")
	}
	if !strings.Contains(err.Error(), "requires a terminal") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestSecretGetAndExportPreserveJSONNumbers(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	want := "12345678901234567890"
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:big_number", want); err != nil {
		t.Fatalf("set big number: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "app:big_number")
	if err != nil {
		t.Fatalf("get big number: %v", err)
	}
	if out != want+"\n" {
		t.Fatalf("secret get changed number: %q", out)
	}
	out, err = runShelf(t, "--vault", data, "secret", "export", "app:big_number", "--format", "json", "--all")
	if err != nil {
		t.Fatalf("export json: %v", err)
	}
	if !strings.Contains(out, want) {
		t.Fatalf("export json changed number: %s", out)
	}
}
func TestExportRejectsInvalidDerivedEnvName(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "123:token", "secret"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	_, err := runShelf(t, "--vault", data, "secret", "export", "123:token", "--format", "shell", "--all")
	if err == nil {
		t.Fatalf("expected invalid derived env name to fail")
	}
	if !strings.Contains(err.Error(), "derived env name") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestExportExactPathDoesNotIncludeLongerPrefixMatch(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "one", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set exact: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token_extra", "two", "--env", "APP_TOKEN_EXTRA"); err != nil {
		t.Fatalf("set prefix: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "export", "app:token", "--format", "shell")
	if err != nil {
		t.Fatalf("export exact: %v", err)
	}
	if out != "export APP_TOKEN=one\n" {
		t.Fatalf("exact export included extra paths: %q", out)
	}
}
func TestExportFiltersEnvOnlyByDefaultAndAllFlag(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:with_env", "one", "--env", "WITH_ENV"); err != nil {
		t.Fatalf("set with env: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:no_env", "two"); err != nil {
		t.Fatalf("set no env: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "export", "app", "--format", "shell")
	if err != nil {
		t.Fatalf("export without --all: %v", err)
	}
	if strings.Contains(out, "app:no_env") || strings.Contains(out, "NO_ENV") || strings.Contains(out, "=two") {
		t.Fatalf("default export leaked secret without env: %q", out)
	}
	if !strings.Contains(out, "WITH_ENV=one") {
		t.Fatalf("default export missing env secret: %q", out)
	}
	out, err = runShelf(t, "--vault", data, "secret", "export", "app", "--format", "shell", "--all")
	if err != nil {
		t.Fatalf("export with --all: %v", err)
	}
	if !strings.Contains(out, "WITH_ENV=one") || !strings.Contains(out, "APP_NO_ENV=two") {
		t.Fatalf("--all export missing secret: %q", out)
	}
}
func TestSecretSetRefusesOverwriteWithoutForce(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "two"); err == nil {
		t.Fatalf("expected overwrite without --force to fail")
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "two", "--force"); err != nil {
		t.Fatalf("force set: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get after force: %v", err)
	}
	if out != "two\n" {
		t.Fatalf("unexpected value after force: %q", out)
	}
}
func TestSecretEditUsesEditorAndValidatesJSON(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"app\",\"key\":\"token\",\"value\":\"edited\",\"env\":\"APP_TOKEN\",\"tags\":[\"app\"]}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--vault", data, "secret", "edit", "app:token"); err != nil {
		t.Fatalf("edit secret: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "secret", "export", "app:token", "--format", "shell")
	if err != nil {
		t.Fatalf("export edited: %v", err)
	}
	if out != "export APP_TOKEN=edited\n" {
		t.Fatalf("unexpected edited export: %q", out)
	}
}

func TestSecretEditTempFileIsRestrictedAndCleanedOnEditorError(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	dir := t.TempDir()
	editor := filepath.Join(dir, "editor.sh")
	pathFile := filepath.Join(dir, "edit-path")
	modeFile := filepath.Join(dir, "edit-mode")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s' \"$1\" > %q\nstat -c '%%a' \"$1\" > %q\nexit 42\n", pathFile, modeFile)
	if err := os.WriteFile(editor, []byte(script), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--vault", data, "secret", "edit", "app:token"); err == nil {
		t.Fatalf("expected editor failure")
	}
	mode, err := os.ReadFile(modeFile)
	if err != nil {
		t.Fatalf("read mode: %v", err)
	}
	if strings.TrimSpace(string(mode)) != "600" {
		t.Fatalf("temp mode = %q, want 600", strings.TrimSpace(string(mode)))
	}
	tmpPath, err := os.ReadFile(pathFile)
	if err != nil {
		t.Fatalf("read temp path: %v", err)
	}
	if _, err := os.Stat(string(tmpPath)); !os.IsNotExist(err) {
		t.Fatalf("temp file was not cleaned up: %v", err)
	}
}

func TestSecretEditTempFileCleanedAfterInvalidJSON(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	dir := t.TempDir()
	editor := filepath.Join(dir, "editor.sh")
	pathFile := filepath.Join(dir, "edit-path")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s' \"$1\" > %q\nprintf '{' > \"$1\"\n", pathFile)
	if err := os.WriteFile(editor, []byte(script), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--vault", data, "secret", "edit", "app:token"); err == nil {
		t.Fatalf("expected invalid JSON failure")
	}
	tmpPath, err := os.ReadFile(pathFile)
	if err != nil {
		t.Fatalf("read temp path: %v", err)
	}
	if _, err := os.Stat(string(tmpPath)); !os.IsNotExist(err) {
		t.Fatalf("temp file was not cleaned up: %v", err)
	}
}
func TestSecretEditCanRenamePath(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"services/app\",\"key\":\"api_key\",\"value\":\"one\",\"env\":\"APP_API_KEY\"}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--vault", data, "secret", "edit", "app:token"); err != nil {
		t.Fatalf("edit rename: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "get", "app:token"); err == nil {
		t.Fatalf("old path still exists")
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "services/app:api_key")
	if err != nil {
		t.Fatalf("get renamed path: %v", err)
	}
	if out != "one\n" {
		t.Fatalf("unexpected renamed value: %q", out)
	}
}
func TestSecretEditRefusesRenameCollision(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set token: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:other", "two"); err != nil {
		t.Fatalf("set other: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"app\",\"key\":\"other\",\"value\":\"one\"}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--vault", data, "secret", "edit", "app:token"); err == nil {
		t.Fatalf("expected edit rename collision to fail")
	}
	out, err := runShelf(t, "--vault", data, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("old path missing after failed rename: %v", err)
	}
	if out != "one\n" {
		t.Fatalf("unexpected old value after failed rename: %q", out)
	}
}
func TestConcurrentSecretSetKeepsAllWrites(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	const count = 20
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := runShelf(t, "--vault", data, "secret", "set", fmt.Sprintf("app:key_%02d", i), fmt.Sprintf("value-%02d", i))
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent set: %v", err)
		}
	}
	out, err := runShelf(t, "--vault", data, "secret", "list", "app")
	if err != nil {
		t.Fatalf("list after concurrent set: %v", err)
	}
	lines := strings.Fields(out)
	if len(lines) != count {
		t.Fatalf("lost writes: got %d paths, want %d\n%s", len(lines), count, out)
	}
}
func TestSecretRmRemovesPathAndFailsOnMissing(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "rm", "app:token"); err != nil {
		t.Fatalf("rm: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "get", "app:token"); err == nil {
		t.Fatalf("expected get after rm to fail")
	}
	if _, err := runShelf(t, "--vault", data, "secret", "rm", "app:token"); err == nil {
		t.Fatalf("expected rm on missing to fail")
	}
}

func TestCompleteSecretSetPathSuggestsUniqueGroups(t *testing.T) {
	paths := []string{
		"providers/openai/accounts/personal:api_key",
		"providers/openai/accounts/personal:token",
		"providers/openrouter/accounts/work:api_key",
		"github/accounts/personal:token",
	}
	got, directive := completeSecretSetPath(paths, "providers/open")
	want := []cobra.Completion{
		"providers/openai/accounts/personal:",
		"providers/openrouter/accounts/work:",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected completions: got=%v want=%v", got, want)
	}
	if directive&cobra.ShellCompDirectiveNoFileComp == 0 {
		t.Fatalf("expected no file completion directive")
	}
	if directive&cobra.ShellCompDirectiveNoSpace == 0 {
		t.Fatalf("expected no space directive for group completion")
	}
}

func TestCompleteSecretSetPathCompletesExistingPathAfterColon(t *testing.T) {
	paths := []string{
		"providers/openai/accounts/personal:api_key",
		"providers/openai/accounts/personal:token",
		"providers/openrouter/accounts/work:api_key",
	}
	got, directive := completeSecretSetPath(paths, "providers/openai/accounts/personal:a")
	want := []cobra.Completion{"providers/openai/accounts/personal:api_key"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected completions: got=%v want=%v", got, want)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive: %v", directive)
	}
}

func TestCompleteSecretSetPathNoMatchHasNoNoSpaceDirective(t *testing.T) {
	paths := []string{"providers/openai/accounts/personal:api_key"}
	got, directive := completeSecretSetPath(paths, "missing")
	if len(got) != 0 {
		t.Fatalf("expected no completions, got %v", got)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive: %v", directive)
	}
}

func TestCompleteSecretSetPathForArgsHandlesColonWordBreakMode(t *testing.T) {
	paths := []string{
		"providers/openai/accounts/personal:api_key",
		"providers/openai/accounts/personal:token",
		"providers/openrouter/accounts/work:api_key",
	}
	got, directive := completeSecretSetPathForArgs(paths, []string{"providers/openai/accounts/personal"}, "a")
	want := []cobra.Completion{"api_key"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected completions: got=%v want=%v", got, want)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive: %v", directive)
	}
}

func TestCompleteSecretSetPathForArgsSecondArgNoCompletion(t *testing.T) {
	paths := []string{"providers/openai/accounts/personal:api_key"}
	got, directive := completeSecretSetPathForArgs(paths, []string{"providers/openai/accounts/personal:api_key"}, "v")
	if len(got) != 0 {
		t.Fatalf("expected no completions, got %v", got)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive: %v", directive)
	}
}

func TestCompleteSecretPathPrefixesSuggestsGroups(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "vault.age")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret"); err != nil {
		t.Fatalf("set app token: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "providers/openai:api_key", "sk"); err != nil {
		t.Fatalf("set provider key: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "__complete", "secret", "list", "")
	if err != nil {
		t.Fatalf("complete secret list: %v\n%s", err, out)
	}
	for _, want := range []string{"app:", "providers/openai:", ":6"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing completion %q in:\n%s", want, out)
		}
	}
}
