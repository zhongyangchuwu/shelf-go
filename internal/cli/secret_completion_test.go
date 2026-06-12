package cli

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

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
