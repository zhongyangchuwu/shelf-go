package project

import (
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vault"
)

func TestResolveEntriesReportsStatuses(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"providers/openai/accounts/personal:api_key":     {Value: projectTestValue(t, "sk-openai")},
		"providers/openrouter/accounts/personal:api_key": {Value: projectTestValue(t, "sk-openrouter")},
	})
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{
		{Path: "providers/openai/accounts/personal:api_key", Env: "OPENAI_API_KEY"},
		{Path: "providers/anthropic/accounts/personal:api_key", Required: boolPtr(false)},
		{Path: "providers/deepseek/accounts/personal:api_key"},
		{Path: "providers/openrouter/accounts/personal:api_key", Env: "OPENAI_API_KEY"},
	}}
	bindings, diagnostics := ResolveEntries(m, st)
	if len(bindings) != 1 || bindings[0].Path != "providers/openai/accounts/personal:api_key" || bindings[0].EnvName != "OPENAI_API_KEY" {
		t.Fatalf("unexpected bindings: %+v", bindings)
	}
	for _, want := range []Diagnostic{
		{Status: "warn", Path: "providers/anthropic/accounts/personal:api_key", Message: "missing optional"},
		{Status: "fail", Path: "providers/deepseek/accounts/personal:api_key", Message: "missing required"},
		{Status: "fail", Path: "providers/openrouter/accounts/personal:api_key", Message: "env name OPENAI_API_KEY conflicts with providers/openai/accounts/personal:api_key"},
	} {
		if !hasDiagnostic(diagnostics, want) {
			t.Fatalf("missing diagnostic %+v in %+v", want, diagnostics)
		}
	}
	if !HasFailures(diagnostics) {
		t.Fatalf("expected failures")
	}
}

func TestResolveEntriesExpandsPrefixSorted(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"providers/openai/accounts/work:api_key":     {Value: projectTestValue(t, "sk-work")},
		"providers/openai/accounts/personal:api_key": {Value: projectTestValue(t, "sk-personal")},
	})
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{{Prefix: "providers/openai/accounts"}}}
	bindings, diagnostics := ResolveEntries(m, st)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	got := bindingLines(bindings)
	want := []string{
		"providers/openai/accounts/personal:api_key=PROVIDERS_OPENAI_ACCOUNTS_PERSONAL_API_KEY:sk-personal",
		"providers/openai/accounts/work:api_key=PROVIDERS_OPENAI_ACCOUNTS_WORK_API_KEY:sk-work",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected bindings:\n%v", got)
	}
}

func TestResolveEntriesReportsEmptyPrefixByRequiredState(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"app:token": {Value: projectTestValue(t, "secret"), Env: "APP_TOKEN"},
	})
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{
		{Path: "app:token"},
		{Prefix: "providers/required"},
		{Prefix: "providers/optional", Required: boolPtr(false)},
	}}
	_, diagnostics := ResolveEntries(m, st)
	for _, want := range []Diagnostic{
		{Status: "fail", Path: "providers/required (prefix)", Message: "no matches required"},
		{Status: "warn", Path: "providers/optional (prefix)", Message: "no matches optional"},
	} {
		if !hasDiagnostic(diagnostics, want) {
			t.Fatalf("missing diagnostic %+v in %+v", want, diagnostics)
		}
	}
}

func TestResolveEntriesExpandsTagSelector(t *testing.T) {
	st := projectTestStore(t, map[string]vault.Secret{
		"providers/openai:api_key":     {Value: projectTestValue(t, "sk-openai"), Tags: []string{"ai", "prod"}},
		"providers/anthropic:api_key":  {Value: projectTestValue(t, "sk-anthropic"), Tags: []string{"ai"}},
		"providers/openrouter:api_key": {Value: projectTestValue(t, "sk-openrouter"), Tags: []string{"ai", "prod"}},
	})
	m := Manifest{Version: CurrentVersion, Secrets: []Entry{{Tags: []string{"ai", "prod"}}}}
	bindings, diagnostics := ResolveEntries(m, st)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	got := bindingLines(bindings)
	want := []string{
		"providers/openai:api_key=PROVIDERS_OPENAI_API_KEY:sk-openai",
		"providers/openrouter:api_key=PROVIDERS_OPENROUTER_API_KEY:sk-openrouter",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected bindings:\n%v", got)
	}
}

func hasDiagnostic(diagnostics []Diagnostic, want Diagnostic) bool {
	for _, got := range diagnostics {
		if got == want {
			return true
		}
	}
	return false
}

func bindingLines(bindings []Binding) []string {
	lines := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		lines = append(lines, binding.Path+"="+binding.EnvName+":"+binding.Value)
	}
	return lines
}
