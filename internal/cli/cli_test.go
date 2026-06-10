package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func runShelf(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestSecretSetGetListInfoExport(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")

	if _, err := runShelf(t, "--data", data, "secret", "set", "providers/openrouter/accounts/personal:api_key", "sk-xxx", "--env", "OPENROUTER_API_KEY", "--tag", "ai"); err != nil {
		t.Fatalf("set secret: %v", err)
	}

	out, err := runShelf(t, "--data", data, "secret", "get", "providers/openrouter/accounts/personal:api_key")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if out != "sk-xxx\n" {
		t.Fatalf("unexpected get output: %q", out)
	}

	out, err = runShelf(t, "--data", data, "secret", "list", "providers/openrouter")
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	if out != "providers/openrouter/accounts/personal:api_key\n" {
		t.Fatalf("unexpected list output: %q", out)
	}

	out, err = runShelf(t, "--data", data, "secret", "info", "providers/openrouter/accounts/personal:api_key")
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

	out, err = runShelf(t, "--data", data, "export", "providers/openrouter/accounts/personal:api_key", "--format", "shell")
	if err != nil {
		t.Fatalf("export shell: %v", err)
	}
	if out != "export OPENROUTER_API_KEY=sk-xxx\n" {
		t.Fatalf("unexpected export output: %q", out)
	}
}

func TestSecretGetAndExportPreserveJSONNumbers(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	want := "12345678901234567890"
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:big_number", want); err != nil {
		t.Fatalf("set big number: %v", err)
	}
	out, err := runShelf(t, "--data", data, "secret", "get", "app:big_number")
	if err != nil {
		t.Fatalf("get big number: %v", err)
	}
	if out != want+"\n" {
		t.Fatalf("secret get changed number: %q", out)
	}
	out, err = runShelf(t, "--data", data, "export", "app:big_number", "--format", "json", "--all")
	if err != nil {
		t.Fatalf("export json: %v", err)
	}
	if !strings.Contains(out, want) {
		t.Fatalf("export json changed number: %s", out)
	}
}

func TestExportExactPathDoesNotIncludeLongerPrefixMatch(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one", "--env", "APP_TOKEN"); err != nil {
		t.Fatalf("set exact: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token_extra", "two", "--env", "APP_TOKEN_EXTRA"); err != nil {
		t.Fatalf("set prefix: %v", err)
	}
	out, err := runShelf(t, "--data", data, "export", "app:token", "--format", "shell")
	if err != nil {
		t.Fatalf("export exact: %v", err)
	}
	if out != "export APP_TOKEN=one\n" {
		t.Fatalf("exact export included extra paths: %q", out)
	}
}

func TestExportFiltersEnvOnlyByDefaultAndAllFlag(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:with_env", "one", "--env", "WITH_ENV"); err != nil {
		t.Fatalf("set with env: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:no_env", "two"); err != nil {
		t.Fatalf("set no env: %v", err)
	}
	out, err := runShelf(t, "--data", data, "export", "app", "--format", "shell")
	if err != nil {
		t.Fatalf("export without --all: %v", err)
	}
	if strings.Contains(out, "app:no_env") || strings.Contains(out, "NO_ENV") || strings.Contains(out, "=two") {
		t.Fatalf("default export leaked secret without env: %q", out)
	}
	if !strings.Contains(out, "WITH_ENV=one") {
		t.Fatalf("default export missing env secret: %q", out)
	}
	out, err = runShelf(t, "--data", data, "export", "app", "--format", "shell", "--all")
	if err != nil {
		t.Fatalf("export with --all: %v", err)
	}
	if !strings.Contains(out, "WITH_ENV=one") || !strings.Contains(out, "APP_NO_ENV=two") {
		t.Fatalf("--all export missing secret: %q", out)
	}
}

func TestSecretSetRefusesOverwriteWithoutForce(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "two"); err == nil {
		t.Fatalf("expected overwrite without --force to fail")
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "two", "--force"); err != nil {
		t.Fatalf("force set: %v", err)
	}
	out, err := runShelf(t, "--data", data, "secret", "get", "app:token")
	if err != nil {
		t.Fatalf("get after force: %v", err)
	}
	if out != "two\n" {
		t.Fatalf("unexpected value after force: %q", out)
	}
}

func TestSecretEditUsesEditorAndValidatesJSON(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"app\",\"key\":\"token\",\"value\":\"edited\",\"env\":\"APP_TOKEN\",\"tags\":[\"app\"]}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--data", data, "secret", "edit", "app:token"); err != nil {
		t.Fatalf("edit secret: %v", err)
	}
	out, err := runShelf(t, "--data", data, "export", "app:token", "--format", "shell")
	if err != nil {
		t.Fatalf("export edited: %v", err)
	}
	if out != "export APP_TOKEN=edited\n" {
		t.Fatalf("unexpected edited export: %q", out)
	}
}

func TestSecretEditCanRenamePath(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("initial set: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"services/app\",\"key\":\"api_key\",\"value\":\"one\",\"env\":\"APP_API_KEY\"}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--data", data, "secret", "edit", "app:token"); err != nil {
		t.Fatalf("edit rename: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "get", "app:token"); err == nil {
		t.Fatalf("old path still exists")
	}
	out, err := runShelf(t, "--data", data, "secret", "get", "services/app:api_key")
	if err != nil {
		t.Fatalf("get renamed path: %v", err)
	}
	if out != "one\n" {
		t.Fatalf("unexpected renamed value: %q", out)
	}
}

func TestSecretEditRefusesRenameCollision(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set token: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:other", "two"); err != nil {
		t.Fatalf("set other: %v", err)
	}
	editor := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(editor, []byte("#!/bin/sh\ncat > \"$1\" <<'JSON'\n{\"group_path\":\"app\",\"key\":\"other\",\"value\":\"one\"}\nJSON\n"), 0o700); err != nil {
		t.Fatalf("write editor: %v", err)
	}
	t.Setenv("EDITOR", editor)
	if _, err := runShelf(t, "--data", data, "secret", "edit", "app:token"); err == nil {
		t.Fatalf("expected edit rename collision to fail")
	}
	out, err := runShelf(t, "--data", data, "secret", "get", "app:token")
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
			_, err := runShelf(t, "--data", data, "secret", "set", fmt.Sprintf("app:key_%02d", i), fmt.Sprintf("value-%02d", i))
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
	out, err := runShelf(t, "--data", data, "secret", "list", "app")
	if err != nil {
		t.Fatalf("list after concurrent set: %v", err)
	}
	lines := strings.Fields(out)
	if len(lines) != count {
		t.Fatalf("lost writes: got %d paths, want %d\n%s", len(lines), count, out)
	}
}
func TestCompletionCommandGeneratesZsh(t *testing.T) {
	out, err := runShelf(t, "completion", "zsh")
	if err != nil {
		t.Fatalf("completion zsh: %v", err)
	}
	if !strings.Contains(out, "#compdef shelf") {
		t.Fatalf("unexpected zsh completion output")
	}
}

func TestProjectIDNormalizesGitRemote(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runGit(t, "remote", "add", "origin", "git@github.com:Owner/Repo.git"); err != nil {
		t.Fatalf("git remote add: %v", err)
	}
	out, err := runShelf(t, "project", "id")
	if err != nil {
		t.Fatalf("project id: %v", err)
	}
	if out != "github.com/Owner/Repo\n" {
		t.Fatalf("unexpected project id: %q", out)
	}
}

func runGit(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func TestDoctorReportsHealthyStore(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--config", configPath, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	for _, want := range []string{"ok   config resolves", "ok   version", "ok   data file exists", "ok   store loads", "ok   data file mode"} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, out)
		}
	}
}

func TestDoctorFailsInvalidStore(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	if err := os.WriteFile(data, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write invalid data: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err == nil {
		t.Fatalf("expected doctor to fail invalid store")
	}
	if !strings.Contains(out, "fail store loads") {
		t.Fatalf("doctor output missing store failure:\n%s", out)
	}
}

func TestDoctorChecksCompletionFromFpath(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := filepath.Join(dir, "secrets.json")
	completionDir := filepath.Join(dir, "zfunc")
	if err := os.MkdirAll(completionDir, 0o700); err != nil {
		t.Fatalf("mkdir completion dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(completionDir, "_shelf"), []byte("#compdef shelf\n"), 0o600); err != nil {
		t.Fatalf("write completion: %v", err)
	}
	t.Setenv("FPATH", filepath.Join(dir, "missing")+":"+completionDir)
	if _, err := runShelf(t, "--config", configPath, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	out, err := runShelf(t, "--config", configPath, "--data", data, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, out)
	}
	want := "ok   completion installed (" + filepath.Join(completionDir, "_shelf") + ")"
	if !strings.Contains(out, want) {
		t.Fatalf("doctor did not use FPATH completion path %q:\n%s", want, out)
	}
}

func TestSecretRmRemovesPathAndFailsOnMissing(t *testing.T) {
	data := filepath.Join(t.TempDir(), "secrets.json")
	if _, err := runShelf(t, "--data", data, "secret", "set", "app:token", "one"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "rm", "app:token"); err != nil {
		t.Fatalf("rm: %v", err)
	}
	if _, err := runShelf(t, "--data", data, "secret", "get", "app:token"); err == nil {
		t.Fatalf("expected get after rm to fail")
	}
	if _, err := runShelf(t, "--data", data, "secret", "rm", "app:token"); err == nil {
		t.Fatalf("expected rm on missing to fail")
	}
}

func TestInitCreatesFilesAndReportsStatus(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "secrets.json")
	cfg := filepath.Join(dir, "shelf.yaml")

	out, err := runShelf(t, "--config", cfg, "--data", data, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	for _, want := range []string{data, "(created)", cfg, "(created)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("init output missing %q: %s", want, out)
		}
	}

	out2, err := runShelf(t, "--config", cfg, "--data", data, "init")
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	for _, want := range []string{data, "(exists)", cfg, "(exists)"} {
		if !strings.Contains(out2, want) {
			t.Fatalf("second init output missing %q: %s", want, out2)
		}
	}

	out3, err := runShelf(t, "--config", cfg, "--data", data, "init", "--minimal")
	if err != nil {
		t.Fatalf("minimal init: %v", err)
	}
	if strings.Contains(out3, "config") {
		t.Fatalf("minimal init mentioned config: %s", out3)
	}
}

func TestInitForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "secrets.json")
	cfg := filepath.Join(dir, "config.yaml")
	if _, err := runShelf(t, "--config", cfg, "--data", data, "init"); err != nil {
		t.Fatalf("first init: %v", err)
	}
	if _, err := runShelf(t, "--config", cfg, "--data", data, "secret", "set", "app:token", "val"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := runShelf(t, "--config", cfg, "--data", data, "init", "--force"); err != nil {
		t.Fatalf("force init: %v", err)
	}
	out, err := runShelf(t, "--config", cfg, "--data", data, "secret", "list")
	if err != nil {
		t.Fatalf("list after force init: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty store after force init, got: %s", out)
	}
}

func TestVersionFlagPrintsVersion(t *testing.T) {
	out, err := runShelf(t, "--version")
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !strings.HasPrefix(out, "shelf") {
		t.Fatalf("unexpected version output: %q", out)
	}
}
