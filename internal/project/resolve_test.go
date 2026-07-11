package project

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestResolveBindingsResolvesItemField(t *testing.T) {
	v := validVault(t)
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"})

	entries, diagnostics := ResolveBindings(m, v)
	if len(diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(entries))
	}
	if entries[0].Path != "github.credential" {
		t.Fatalf("unexpected path: %s", entries[0].Path)
	}
	if entries[0].EnvName != "GITHUB_TOKEN" {
		t.Fatalf("unexpected env: %s", entries[0].EnvName)
	}
	if entries[0].Value != "ghp_fake_token" {
		t.Fatalf("unexpected value: %s", entries[0].Value)
	}
}

func TestResolveBindingsMultiple(t *testing.T) {
	v := validVault(t)
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"})
	m.AddBinding(Binding{Item: "openai", Field: "credential", Env: "OPENAI_API_KEY"})

	entries, diagnostics := ResolveBindings(m, v)
	if len(diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(entries))
	}
}

func TestResolveBindingsMissingItemIsFail(t *testing.T) {
	v := validVault(t)
	m := New()
	m.AddBinding(Binding{Item: "nonexistent", Field: "credential", Env: "GITHUB_TOKEN"})

	entries, diagnostics := ResolveBindings(m, v)
	if len(entries) != 0 {
		t.Fatalf("expected 0 bindings on missing item, got %d", len(entries))
	}
	if !hasDiagnostic(diagnostics, Diagnostic{Status: "fail", Path: "nonexistent.credential", Message: "missing item required"}) {
		t.Fatalf("missing fail diagnostic: %+v", diagnostics)
	}
}

func TestResolveBindingsMissingItemOptionalIsWarn(t *testing.T) {
	v := validVault(t)
	f := false
	m := New()
	m.AddBinding(Binding{Item: "nonexistent", Field: "credential", Env: "GITHUB_TOKEN", Required: &f})

	entries, diagnostics := ResolveBindings(m, v)
	if len(entries) != 0 {
		t.Fatalf("expected 0 bindings on optional missing item, got %d", len(entries))
	}
	if !hasDiagnostic(diagnostics, Diagnostic{Status: "warn", Path: "nonexistent.credential", Message: "missing item optional"}) {
		t.Fatalf("missing warn diagnostic: %+v", diagnostics)
	}
}

func TestResolveBindingsMissingFieldIsFail(t *testing.T) {
	v := validVault(t)
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "nonexistent", Env: "GITHUB_TOKEN"})

	entries, diagnostics := ResolveBindings(m, v)
	if len(entries) != 0 {
		t.Fatalf("expected 0 bindings on missing field, got %d", len(entries))
	}
	if !hasDiagnostic(diagnostics, Diagnostic{Status: "fail", Path: "github.nonexistent", Message: "missing field required"}) {
		t.Fatalf("missing fail diagnostic: %+v", diagnostics)
	}
}

func TestResolveBindingsMissingFieldOptionalIsWarn(t *testing.T) {
	v := validVault(t)
	f := false
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "nonexistent", Env: "GITHUB_TOKEN", Required: &f})

	entries, diagnostics := ResolveBindings(m, v)
	if len(entries) != 0 {
		t.Fatalf("expected 0 bindings on optional missing field, got %d", len(entries))
	}
	if !hasDiagnostic(diagnostics, Diagnostic{Status: "warn", Path: "github.nonexistent", Message: "missing field optional"}) {
		t.Fatalf("missing warn diagnostic: %+v", diagnostics)
	}
}

func TestResolveBindingsMissingValueIsFail(t *testing.T) {
	v := validVault(t)
	// Empty the credential value on the github item
	for i, item := range v.Items {
		if item.ID == "github" {
			for j, f := range item.Fields {
				if f.ID == "credential" {
					v.Items[i].Fields[j].Value = nil
				}
			}
		}
	}
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN"})

	entries, diagnostics := ResolveBindings(m, v)
	if len(entries) != 0 {
		t.Fatalf("expected 0 bindings on missing value, got %d", len(entries))
	}
	if !hasDiagnostic(diagnostics, Diagnostic{Status: "fail", Path: "github.credential", Message: "missing value required"}) {
		t.Fatalf("missing fail diagnostic: %+v", diagnostics)
	}
}

func TestResolveBindingsMissingValueOptionalIsWarn(t *testing.T) {
	v := validVault(t)
	for i, item := range v.Items {
		if item.ID == "github" {
			for j, f := range item.Fields {
				if f.ID == "credential" {
					v.Items[i].Fields[j].Value = nil
				}
			}
		}
	}
	f := false
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "credential", Env: "GITHUB_TOKEN", Required: &f})

	entries, diagnostics := ResolveBindings(m, v)
	if len(entries) != 0 {
		t.Fatalf("expected 0 bindings on optional missing value, got %d", len(entries))
	}
	if !hasDiagnostic(diagnostics, Diagnostic{Status: "warn", Path: "github.credential", Message: "missing value optional"}) {
		t.Fatalf("missing warn diagnostic: %+v", diagnostics)
	}
}

func TestResolveBindingsDuplicateEnvIsFail(t *testing.T) {
	v := validVault(t)
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "credential", Env: "DUPLICATE_ENV"})
	m.AddBinding(Binding{Item: "openai", Field: "credential", Env: "DUPLICATE_ENV"})

	entries, diagnostics := ResolveBindings(m, v)
	// First binding resolves, second fails with conflict
	if len(entries) != 1 {
		t.Fatalf("expected 1 binding on duplicate env, got %d", len(entries))
	}
	found := false
	for _, d := range diagnostics {
		if d.Status == "fail" && strings.Contains(d.Message, "conflicts") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected conflict diagnostic: %+v", diagnostics)
	}
}

func TestResolveBindingsNonSecretFieldIsFail(t *testing.T) {
	v := validVault(t)
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "provider", Env: "GITHUB_PROVIDER"})

	entries, diagnostics := ResolveBindings(m, v)
	if len(entries) != 0 {
		t.Fatalf("expected 0 bindings for non-secret field, got %d", len(entries))
	}
	if !hasDiagnostic(diagnostics, Diagnostic{Status: "fail", Path: "github.provider", Message: "field is not secret"}) {
		t.Fatalf("expected field-not-secret diagnostic: %+v", diagnostics)
	}
}

func TestResolveBindingsDuplicateEnvDiagnosticsAreValueFree(t *testing.T) {
	v := validVault(t)
	m := New()
	m.AddBinding(Binding{Item: "github", Field: "credential", Env: "DUPE"})
	m.AddBinding(Binding{Item: "openai", Field: "credential", Env: "DUPE"})

	_, diagnostics := ResolveBindings(m, v)
	for _, d := range diagnostics {
		if strings.Contains(d.Message, "ghp_") || strings.Contains(d.Message, "sk-") {
			t.Fatalf("diagnostic leaks secret value: %+v", d)
		}
	}
}

func TestResolveBindingsAllDiagnosticsAreValueFree(t *testing.T) {
	v := validVault(t)
	m := New()
	m.AddBinding(Binding{Item: "nonexistent", Field: "credential", Env: "NOPE"})
	m.AddBinding(Binding{Item: "github", Field: "nonexistent", Env: "NOPE2"})
	m.AddBinding(Binding{Item: "github", Field: "provider", Env: "PUB"})

	_, diagnostics := ResolveBindings(m, v)
	for _, d := range diagnostics {
		if strings.Contains(d.Message, "ghp_") || strings.Contains(d.Message, "sk-") {
			t.Fatalf("diagnostic leaks secret value: %+v", d)
		}
	}
}

// --- helpers ---

func hasDiagnostic(diagnostics []Diagnostic, want Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic == want {
			return true
		}
	}
	return false
}

func validVault(t *testing.T) vaultmodel.Vault {
	t.Helper()
	return vaultmodel.Vault{
		Version:   vaultmodel.Version,
		Templates: vaultmodel.BuiltInTemplates(),
		Profiles:  vaultmodel.BuiltInProfiles(),
		Items: []vaultmodel.Item{
			{
				ID:       "github",
				Title:    "GitHub Token",
				Template: vaultmodel.TemplateAPICredential,
				Fields: []vaultmodel.Field{
					{ID: "credential", Label: "Credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Value: json.RawMessage(`"ghp_fake_token"`)},
					{ID: "credential_env", Label: "Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Value: json.RawMessage(`"GITHUB_TOKEN"`)},
					{ID: "provider", Label: "Provider", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Value: json.RawMessage(`"GitHub"`)},
				},
			},
			{
				ID:       "openai",
				Title:    "OpenAI API Key",
				Template: vaultmodel.TemplateAPICredential,
				Fields: []vaultmodel.Field{
					{ID: "credential", Label: "Credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Value: json.RawMessage(`"sk-openai-test"`)},
					{ID: "credential_env", Label: "Env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Value: json.RawMessage(`"OPENAI_API_KEY"`)},
				},
			},
		},
	}
}
