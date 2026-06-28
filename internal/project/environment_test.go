package project

import (
	"strings"
	"testing"
)

func TestChildEnvReplacesParentAndAppendsMissing(t *testing.T) {
	entries := []Binding{{EnvName: "APP_TOKEN", Value: "secret"}, {EnvName: "NEW_TOKEN", Value: "new-secret"}}
	env := ChildEnv([]string{"APP_TOKEN=parent", "OTHER=value"}, entries)
	want := []string{"APP_TOKEN=secret", "OTHER=value", "NEW_TOKEN=new-secret"}
	if strings.Join(env, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected env:\n%q", env)
	}
}

func TestChildEnvDropsMalformedParentEntryWhenShelfOverridesIt(t *testing.T) {
	entries := []Binding{{EnvName: "APP_TOKEN", Value: "secret"}}
	env := ChildEnv([]string{"APP_TOKEN", "OTHER=value"}, entries)
	want := []string{"OTHER=value", "APP_TOKEN=secret"}
	if strings.Join(env, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected env:\n%q", env)
	}
}

func TestEnvOverrideWarningsReportsParentConflicts(t *testing.T) {
	entries := []Binding{{EnvName: "APP_TOKEN", Value: "secret"}, {EnvName: "NEW_TOKEN", Value: "new-secret"}}
	warnings := EnvOverrideWarnings(entries, []string{"APP_TOKEN=parent", "OTHER=value"})
	if len(warnings) != 1 || warnings[0] != "warn APP_TOKEN overrides existing environment variable" {
		t.Fatalf("unexpected warnings: %#v", warnings)
	}
}
