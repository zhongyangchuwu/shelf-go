package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectAddListAndRmTagEntry(t *testing.T) {
	dir, data := setupProjectTest(t)
	setProjectTagSecrets(t, data)

	out, err := runShelf(t, "--vault", data, "project", "add", "--tag", "ai", "--tag", "prod", "--optional")
	if err != nil {
		t.Fatalf("project add tag: %v\n%s", err, out)
	}
	if !strings.Contains(out, "added ai,prod") {
		t.Fatalf("unexpected add output: %q", out)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(dir, ".shelf.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	manifestText := string(manifestBytes)
	for _, want := range []string{`"tags"`, `"ai"`, `"prod"`, `"required": false`} {
		if !strings.Contains(manifestText, want) {
			t.Fatalf("manifest missing %q:\n%s", want, manifestText)
		}
	}
	for _, forbidden := range []string{"secret-token", "deploy-secret", "OPENAI_API_KEY"} {
		if strings.Contains(manifestText, forbidden) {
			t.Fatalf("manifest leaked value/env %q:\n%s", forbidden, manifestText)
		}
	}

	out, err = runShelf(t, "--vault", data, "project", "list")
	if err != nil {
		t.Fatalf("project list: %v\n%s", err, out)
	}
	if !strings.Contains(out, "tag    ai,prod (optional)") {
		t.Fatalf("missing tag list line:\n%s", out)
	}

	out, err = runShelf(t, "--vault", data, "project", "rm", "ai,prod")
	if err != nil {
		t.Fatalf("project rm tag: %v\n%s", err, out)
	}
	if !strings.Contains(out, "removed ai,prod") {
		t.Fatalf("unexpected rm output: %q", out)
	}
}

func TestProjectExportExpandsTagEntryWithAndSemantics(t *testing.T) {
	_, data := setupProjectTest(t)
	setProjectTagSecrets(t, data)
	if _, err := runShelf(t, "--vault", data, "project", "add", "--tag", "ai", "--tag", "prod"); err != nil {
		t.Fatalf("project add tag: %v", err)
	}

	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("project export tag: %v\n%s", err, out)
	}
	for _, want := range []string{"OPENAI_API_KEY=secret-token", "DEPLOY_KEY=deploy-secret"} {
		if !strings.Contains(out, want) {
			t.Fatalf("tag export missing %q:\n%s", want, out)
		}
	}
	for _, forbidden := range []string{"APP_URL=https://example.test", "OPS_PASSWORD=ops-secret"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("tag export included non-AND match %q:\n%s", forbidden, out)
		}
	}
}

func TestProjectStatusShowsTagExpansion(t *testing.T) {
	_, data := setupProjectTest(t)
	setProjectTagSecrets(t, data)
	if _, err := runShelf(t, "--vault", data, "project", "add", "--tag", "ai", "--tag", "prod"); err != nil {
		t.Fatalf("project add tag: %v", err)
	}

	out, err := runShelf(t, "--vault", data, "project", "status")
	if err != nil {
		t.Fatalf("project status tag: %v\n%s", err, out)
	}
	for _, want := range []string{"ok   app:token -> OPENAI_API_KEY", "ok   ops:deploy -> DEPLOY_KEY"} {
		if !strings.Contains(out, want) {
			t.Fatalf("project status missing %q:\n%s", want, out)
		}
	}
}

func TestProjectTagEntryReportsEmptyRequiredAndOptional(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:token", "secret", "--env", "APP_TOKEN", "--tag", "ai"); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	requiredManifest := `{"version":1,"secrets":[{"tags":["missing"],"required":true}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(requiredManifest), 0o600); err != nil {
		t.Fatalf("write required manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err == nil {
		t.Fatalf("expected required empty tag to fail")
	}
	if !strings.Contains(out, "fail missing (tags) no matches required") {
		t.Fatalf("missing required tag failure:\n%s", out)
	}

	optionalManifest := `{"version":1,"secrets":[{"path":"app:token"},{"tags":["missing"],"required":false}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(optionalManifest), 0o600); err != nil {
		t.Fatalf("write optional manifest: %v", err)
	}
	out, err = runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err != nil {
		t.Fatalf("expected optional empty tag to pass: %v\n%s", err, out)
	}
	if !strings.Contains(out, "warn missing (tags) no matches optional") || !strings.Contains(out, "APP_TOKEN=secret") {
		t.Fatalf("missing optional tag warning or present export:\n%s", out)
	}
}

func TestProjectTagEntryFailsOnEnvConflict(t *testing.T) {
	dir := t.TempDir()
	chdirTest(t, dir)
	if _, err := runGit(t, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	data := filepath.Join(dir, "secrets.json")
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:one", "one", "--env", "APP_TOKEN", "--tag", "ai"); err != nil {
		t.Fatalf("set one: %v", err)
	}
	if _, err := runShelf(t, "--vault", data, "secret", "set", "app:two", "two", "--env", "APP_TOKEN", "--tag", "ai"); err != nil {
		t.Fatalf("set two: %v", err)
	}
	manifest := `{"version":1,"secrets":[{"tags":["ai"]}]}`
	if err := os.WriteFile(filepath.Join(dir, ".shelf.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	out, err := runShelf(t, "--vault", data, "project", "export", "--format", "env")
	if err == nil {
		t.Fatalf("expected tag env conflict to fail")
	}
	if !strings.Contains(out, "env name APP_TOKEN conflicts") {
		t.Fatalf("missing conflict diagnostic:\n%s", out)
	}
}

func TestProjectAddRejectsInvalidTagCombinations(t *testing.T) {
	_, data := setupProjectTest(t)
	setProjectTagSecrets(t, data)
	if _, err := runShelf(t, "--vault", data, "project", "add", "app", "--tag", "ai"); err == nil {
		t.Fatalf("expected path plus --tag to fail")
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "--tag", "ai", "--env", "APP_TOKEN"); err == nil {
		t.Fatalf("expected --env plus --tag to fail")
	}
	if _, err := runShelf(t, "--vault", data, "project", "add", "--tag", "missing"); err == nil {
		t.Fatalf("expected missing tag selector to fail")
	}
}

func setProjectTagSecrets(t *testing.T, data string) {
	t.Helper()
	commands := [][]string{
		{"secret", "set", "app:token", "secret-token", "--env", "OPENAI_API_KEY", "--tag", "ai", "--tag", "prod"},
		{"secret", "set", "app:url", "https://example.test", "--env", "APP_URL", "--tag", "ai"},
		{"secret", "set", "ops:deploy", "deploy-secret", "--env", "DEPLOY_KEY", "--tag", "ai", "--tag", "prod"},
		{"secret", "set", "ops:password", "ops-secret", "--env", "OPS_PASSWORD", "--tag", "prod"},
	}
	for _, args := range commands {
		fullArgs := append([]string{"--vault", data}, args...)
		if out, err := runShelf(t, fullArgs...); err != nil {
			t.Fatalf("set tagged secret %v: %v\n%s", args, err, out)
		}
	}
}
