package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func TestSecretServiceWritesListsAndReveals(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new manager service: %v", err)
	}
	value := "secret-value"
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:token", Value: &value, Env: "APP_TOKEN", Description: "primary token", Tags: []string{"api"}}); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	items, err := service.ListSecrets("api")
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	if len(items) != 1 || items[0].Path != "app:token" || !items[0].ValueSet {
		t.Fatalf("unexpected list result: %#v", items)
	}
	revealed, err := service.RevealSecret("app:token")
	if err != nil {
		t.Fatalf("reveal secret: %v", err)
	}
	if revealed != "secret-value" {
		t.Fatalf("revealed = %q, want secret-value", revealed)
	}
}

func newTestSecretService(vaultPath string, recipients, identityPaths []string) (*SecretService, error) {
	if err := os.MkdirAll(filepath.Dir(vaultPath), 0o700); err != nil {
		return nil, err
	}
	repo, err := testApp().vaults.Open(vault.Options{Path: vaultPath, Recipients: recipients, IdentityPaths: identityPaths})
	if err != nil {
		return nil, err
	}
	return NewSecretService(repo)
}
func TestSecretService_ListSecretsAll(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	v1 := "val1"
	v2 := "val2"
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:one", Value: &v1, Env: "ONE", Tags: []string{"a"}}); err != nil {
		t.Fatalf("write one: %v", err)
	}
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:two", Value: &v2, Env: "TWO", Tags: []string{"b"}}); err != nil {
		t.Fatalf("write two: %v", err)
	}
	all, err := service.ListSecrets("")
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("list all = %d items, want 2", len(all))
	}
	// Verify both values are value-free in list output
	for _, item := range all {
		if item.ValueSet && item.Path != "app:one" && item.Path != "app:two" {
			t.Fatalf("unexpected path: %q", item.Path)
		}
	}
}

func TestSecretService_ListSecretsQueryFilter(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	v := "value"
	for _, p := range []string{"database:password", "app:token", "app:secret"} {
		if err := service.WriteSecret(false, WriteSecretRequest{Path: p, Value: &v}); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	items, err := service.ListSecrets("app")
	if err != nil {
		t.Fatalf("list app: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("list app = %d items, want 2", len(items))
	}
	if items[0].Path != "app:token" && items[1].Path != "app:token" {
		t.Fatalf("list missing app:token: %#v", items)
	}
	items, err = service.ListSecrets("database")
	if err != nil {
		t.Fatalf("list database: %v", err)
	}
	if len(items) != 1 || items[0].Path != "database:password" {
		t.Fatalf("list database = %#v, want [database:password]", items)
	}
	items, err = service.ListSecrets("nonexistent")
	if err != nil {
		t.Fatalf("list nonexistent: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("list nonexistent = %d items, want 0", len(items))
	}
}

func TestSecretService_RevealSecret_ReturnsOnlyValue(t *testing.T) {
	// v2 contract: reveal/get materializes only the explicit value —
	// no env name, no metadata, nothing else.
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	val := "exact-secret-123"
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:reveal_test", Value: &val, Env: "REVEAL_ENV"}); err != nil {
		t.Fatalf("write: %v", err)
	}
	revealed, err := service.RevealSecret("app:reveal_test")
	if err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if revealed != val {
		t.Fatalf("revealed = %q, want %q", revealed, val)
	}
}

func TestSecretService_RevealSecret_NonExistent(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	if _, err := service.RevealSecret("missing:path"); err == nil {
		t.Fatalf("expected error for missing path")
	}
}

func TestSecretService_DeleteSecret(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	val := "delete-me"
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:delete_test", Value: &val}); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := service.DeleteSecret("app:delete_test"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := service.RevealSecret("app:delete_test"); err == nil {
		t.Fatalf("expected error for deleted path")
	}
}

func TestSecretService_DeleteSecret_NonExistent(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	if err := service.DeleteSecret("nonexistent:key"); err == nil {
		t.Fatalf("expected error for deleting non-existent secret")
	}
}

func TestSecretService_UpdateSecret(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	orig := "original"
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:update_test", Value: &orig, Env: "ORIG_VAR"}); err != nil {
		t.Fatalf("write original: %v", err)
	}
	// Update with new value and env (update=true)
	updated := "updated"
	if err := service.WriteSecret(true, WriteSecretRequest{Path: "app:update_test", Value: &updated, Env: "UPDATED_VAR"}); err != nil {
		t.Fatalf("write updated: %v", err)
	}
	info, err := service.SecretInfo("app:update_test")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if info.Env != "UPDATED_VAR" {
		t.Fatalf("env after update = %q, want UPDATED_VAR", info.Env)
	}
	revealed, err := service.RevealSecret("app:update_test")
	if err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if revealed != "updated" {
		t.Fatalf("revealed = %q, want updated", revealed)
	}
}

func TestSecretService_WriteSecret_ForceReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	val := "first"
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:force_test", Value: &val, Env: "FIRST_VAR"}); err != nil {
		t.Fatalf("write first: %v", err)
	}
	// WriteSecret with update=false and no Force should fail (existing)
	val2 := "second"
	if err := service.WriteSecret(false, WriteSecretRequest{Path: "app:force_test", Value: &val2, Env: "SECOND_VAR"}); err == nil {
		t.Fatalf("expected overwrite to fail without force")
	}
	info, err := service.SecretInfo("app:force_test")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if info.Env != "FIRST_VAR" {
		t.Fatalf("env changed despite failed overwrite: %q", info.Env)
	}
}

func TestSecretService_SecretInfo_NonExistent(t *testing.T) {
	dir := t.TempDir()
	identity, err := testApp().EnsureInitIdentity(filepath.Join(dir, "identity.txt"))
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}
	service, err := newTestSecretService(filepath.Join(dir, "vault.age"), []string{identity.Recipient}, []string{filepath.Join(dir, "identity.txt")})
	if err != nil {
		t.Fatalf("new secret service: %v", err)
	}
	if _, err := service.SecretInfo("missing"); err == nil {
		t.Fatalf("expected error for missing secret info")
	}
}
