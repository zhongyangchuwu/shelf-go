package project

import (
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func TestBuildEntryCreatesPathEntry(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"providers/openai/accounts/personal:api_key": {Value: projectTestValue(t, "sk-test")},
	})
	entry, err := BuildEntry(st, AddEntryRequest{Selector: "providers/openai/accounts/personal:api_key", Env: "OPENAI_API_KEY"})
	if err != nil {
		t.Fatalf("build entry: %v", err)
	}
	if entry.Path != "providers/openai/accounts/personal:api_key" || entry.Env != "OPENAI_API_KEY" || entry.IsPrefix() || entry.IsTag() {
		t.Fatalf("unexpected path entry: %+v", entry)
	}
	if !entry.IsRequired() {
		t.Fatalf("path entries should default to required")
	}
}

func TestBuildEntryCreatesOptionalPrefixEntry(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"providers/openai/accounts/personal:api_key": {Value: projectTestValue(t, "sk-test")},
	})
	entry, err := BuildEntry(st, AddEntryRequest{Selector: "providers/openai/accounts/personal", Optional: true})
	if err != nil {
		t.Fatalf("build entry: %v", err)
	}
	if entry.Prefix != "providers/openai/accounts/personal" || !entry.IsPrefix() || entry.IsRequired() {
		t.Fatalf("unexpected prefix entry: %+v", entry)
	}
}

func TestBuildEntryCreatesTagEntry(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"providers/openai/accounts/personal:api_key": {Value: projectTestValue(t, "sk-test"), Tags: []string{"ai", "prod"}},
	})
	entry, err := BuildEntry(st, AddEntryRequest{Tags: []string{"ai", "prod"}})
	if err != nil {
		t.Fatalf("build entry: %v", err)
	}
	if !entry.IsTag() || entry.Key() != "ai,prod" || !entry.IsRequired() {
		t.Fatalf("unexpected tag entry: %+v", entry)
	}
}

func TestBuildEntryRejectsInvalidRequests(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"providers/openai/accounts/personal:api_key": {Value: projectTestValue(t, "sk-test"), Tags: []string{"ai"}},
	})
	tests := []struct {
		name    string
		req     AddEntryRequest
		wantErr string
	}{
		{name: "path with tags", req: AddEntryRequest{Selector: "providers/openai/accounts/personal:api_key", Tags: []string{"ai"}}, wantErr: "path-or-prefix must not be set with --tag"},
		{name: "missing selector", req: AddEntryRequest{}, wantErr: "path-or-prefix or --tag is required"},
		{name: "env with tags", req: AddEntryRequest{Env: "OPENAI_API_KEY", Tags: []string{"ai"}}, wantErr: "--env is only valid for path entries"},
		{name: "env with prefix", req: AddEntryRequest{Selector: "providers/openai/accounts/personal", Env: "OPENAI_API_KEY"}, wantErr: "--env is only valid for path entries"},
		{name: "missing path", req: AddEntryRequest{Selector: "providers/missing/accounts/personal:api_key"}, wantErr: "secret not found"},
		{name: "empty prefix", req: AddEntryRequest{Selector: "providers/missing"}, wantErr: "no secrets match prefix"},
		{name: "empty tags", req: AddEntryRequest{Tags: []string{"missing"}}, wantErr: "no secrets match tags"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildEntry(st, tt.req)
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddEntryRejectsDuplicate(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"app:token": {Value: projectTestValue(t, "secret")},
	})
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{{Path: "app:token"}}}
	_, _, err := AddEntry(m, st, AddEntryRequest{Selector: "app:token"})
	if err == nil {
		t.Fatalf("expected duplicate add to fail")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func projectTestStore(t *testing.T, secrets map[string]vault.Secret) *vault.Store {
	t.Helper()
	st := &vault.Store{Data: vault.NewData()}
	for path, secret := range secrets {
		if err := st.Set(path, secret, false); err != nil {
			t.Fatalf("set %s: %v", path, err)
		}
	}
	return st
}

func projectTestValue(t *testing.T, value string) []byte {
	t.Helper()
	raw, err := vault.ParseValue(value)
	if err != nil {
		t.Fatalf("parse value: %v", err)
	}
	return raw
}
