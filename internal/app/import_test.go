package app

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/jsonvault"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

type fakeGopassImportClient struct {
	paths     []string
	passwords map[string]string
	errs      map[string]error
	shown     []string
}

func (c *fakeGopassImportClient) ListFlat(prefix string) ([]string, error) {
	return append([]string(nil), c.paths...), nil
}

func (c *fakeGopassImportClient) ShowPassword(path string) (string, error) {
	c.shown = append(c.shown, path)
	if err := c.errs[path]; err != nil {
		return "", err
	}
	return c.passwords[path], nil
}

func TestMapGopassPathUsesLastSlash(t *testing.T) {
	got, ok, reason := MapGopassPath("providers/openai/api_key")
	if !ok {
		t.Fatalf("map failed: %s", reason)
	}
	if got != "providers/openai:api_key" {
		t.Fatalf("path = %s, want providers/openai:api_key", got)
	}
	if _, ok, _ := MapGopassPath("token"); ok {
		t.Fatalf("single segment path should be skipped")
	}
}

func TestImportGopassToVaultImportsStringsAndSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	v := newTestVaultFile(t, dir)
	st := &vault.Store{Data: vault.NewData()}
	if err := st.Set("app:existing", vault.Secret{Value: []byte(`"old"`)}, false); err != nil {
		t.Fatalf("set existing: %v", err)
	}
	if err := v.Save(st); err != nil {
		t.Fatalf("save seed: %v", err)
	}
	client := &fakeGopassImportClient{
		paths: []string{"app/token", "app/existing", "invalid"},
		passwords: map[string]string{
			"app/token":    `{"looks":"json"}`,
			"app/existing": "new",
		},
	}
	result, err := ImportGopassToVault(v, client, GopassImportOptions{})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if !reflect.DeepEqual(result.Imported, []string{"app:token"}) {
		t.Fatalf("imported = %#v", result.Imported)
	}
	if !reflect.DeepEqual(result.SkippedExisting, []string{"app:existing"}) {
		t.Fatalf("skipped existing = %#v", result.SkippedExisting)
	}
	if len(result.SkippedInvalid) != 1 || result.SkippedInvalid[0].Path != "invalid" {
		t.Fatalf("skipped invalid = %#v", result.SkippedInvalid)
	}
	loaded, err := v.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	secret, ok := loaded.Get("app:token")
	if !ok {
		t.Fatalf("missing imported secret")
	}
	if string(secret.Value) != `"{\"looks\":\"json\"}"` {
		t.Fatalf("value = %s, want JSON string", secret.Value)
	}
	existing, _ := loaded.Get("app:existing")
	if string(existing.Value) != `"old"` {
		t.Fatalf("existing overwritten without force: %s", existing.Value)
	}
}

func TestImportGopassToVaultAbortsBeforeWriteOnReadFailure(t *testing.T) {
	dir := t.TempDir()
	v := newTestVaultFile(t, dir)
	if err := v.Save(&vault.Store{Data: vault.NewData()}); err != nil {
		t.Fatalf("save seed: %v", err)
	}
	client := &fakeGopassImportClient{
		paths:     []string{"app/token", "app/broken"},
		passwords: map[string]string{"app/token": "secret"},
		errs:      map[string]error{"app/broken": errors.New("boom")},
	}
	if _, err := ImportGopassToVault(v, client, GopassImportOptions{}); err == nil {
		t.Fatalf("expected read failure")
	}
	loaded, err := v.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Data.Secrets) != 0 {
		t.Fatalf("vault changed after failed import: %#v", loaded.Data.Secrets)
	}
}

func newTestVaultFile(t *testing.T, dir string) *jsonvault.Vault {
	t.Helper()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("identity: %v", err)
	}
	v, err := jsonvault.NewVault(filepath.Join(dir, "vault.age"), vault.Options{Path: filepath.Join(dir, "vault.age"), Recipients: []string{identity.Recipient}, IdentityPaths: []string{filepath.Join(dir, "identity.txt")}})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	return v
}
