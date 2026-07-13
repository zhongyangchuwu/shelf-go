package vaultmodel_test

import (
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestResolverResolvesItemFieldAndEnvName(t *testing.T) {
	item := validAPICredential()
	resolver, err := vaultmodel.NewResolver(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item}})
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	_, resolved, err := resolver.ResolveItem("openai-key")
	if err != nil {
		t.Fatalf("resolve alias: %v", err)
	}
	if resolved.ID != item.ID {
		t.Fatalf("resolved id = %q, want %q", resolved.ID, item.ID)
	}

	_, field, err := resolver.DefaultField(resolved, "")
	if err != nil {
		t.Fatalf("default field: %v", err)
	}
	if field.ID != "credential" {
		t.Fatalf("default field = %q, want credential", field.ID)
	}

	envName, ok, err := resolver.EnvName(resolved, field.ID)
	if err != nil {
		t.Fatalf("env name: %v", err)
	}
	if !ok || envName != "OPENAI_API_KEY" {
		t.Fatalf("env name = %q, found = %t", envName, ok)
	}

	ref, err := resolver.ResolveEnv("OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("resolve env: %v", err)
	}
	if ref.Item.ID != item.ID || ref.Field.ID != "credential" {
		t.Fatalf("env reference = %#v", ref)
	}
}

func TestResolverTreatsMissingOptionalEnvMappingAsUnmapped(t *testing.T) {
	item := validAPICredential()
	fields := make([]vaultmodel.Field, 0, len(item.Fields)-1)
	for _, field := range item.Fields {
		if field.ID != vaultmodel.EnvFieldID("base_url") {
			fields = append(fields, field)
		}
	}
	item.Fields = fields
	resolver, err := vaultmodel.NewResolver(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{item}})
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	envName, mapped, err := resolver.EnvName(item, "base_url")
	if err != nil || mapped || envName != "" {
		t.Fatalf("missing optional mapping = name %q mapped %t err %v", envName, mapped, err)
	}
}

func TestResolverAmbiguityErrorsAreValueFree(t *testing.T) {
	first := validGenericSecret("first", "First", "DUPLICATE_ENV")
	first.Aliases = []string{"shared"}
	second := validGenericSecret("second", "Second", "DUPLICATE_ENV")
	second.Aliases = []string{"shared"}
	resolver, err := vaultmodel.NewResolver(vaultmodel.Vault{Version: vaultmodel.Version, Items: []vaultmodel.Item{first, second}})
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	_, _, err = resolver.ResolveItem("shared")
	if err == nil || !strings.Contains(err.Error(), "selector matched multiple items: shared") {
		t.Fatalf("expected item ambiguity, got: %v", err)
	}
	if strings.Contains(err.Error(), "fake-secret-value") {
		t.Fatalf("item ambiguity leaked a value: %v", err)
	}

	_, err = resolver.ResolveEnv("DUPLICATE_ENV")
	if err == nil || !strings.Contains(err.Error(), "env selector matched multiple fields: DUPLICATE_ENV") {
		t.Fatalf("expected env ambiguity, got: %v", err)
	}
	if strings.Contains(err.Error(), "fake-secret-value") {
		t.Fatalf("env ambiguity leaked a value: %v", err)
	}
}
