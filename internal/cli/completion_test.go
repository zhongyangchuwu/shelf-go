package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestSecretSelectorCompletionsSkipAmbiguousTitles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	repo := vaultjson.NewRepository(path)
	vault := vaultmodel.Vault{
		Version: vaultmodel.CurrentVersion,
		Items: []vaultmodel.Item{
			{
				ID: "item_one", Title: "shared", Template: vaultmodel.TemplateGenericSecret,
				Fields: []vaultmodel.Field{
					{ID: "value", Value: rawJSON("first-secret")},
					{ID: vaultmodel.EnvFieldID("value"), Value: rawJSON("FIRST_TOKEN")},
				},
			},
			{
				ID: "item_two", Title: "shared", Template: vaultmodel.TemplateGenericSecret,
				Fields: []vaultmodel.Field{
					{ID: "value", Value: rawJSON("second-secret")},
					{ID: vaultmodel.EnvFieldID("value"), Value: rawJSON("SECOND_TOKEN")},
				},
			},
		},
	}
	if err := repo.Save(vault); err != nil {
		t.Fatalf("write duplicate-title vault: %v", err)
	}
	got := dynamicCompletions(t, []string{"--vault", path}, "secret", "info", "sha")
	if len(got) != 0 {
		t.Fatalf("duplicate-title selector completions = %v, want ambiguous title omitted", got)
	}
}

func TestSecretSelectorCompletionsDeduplicateIDAndTitle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	repo := vaultjson.NewRepository(path)
	vault := vaultmodel.Vault{
		Version: vaultmodel.CurrentVersion,
		Items: []vaultmodel.Item{{
			ID:       "item_shared",
			Title:    "item_shared",
			Template: vaultmodel.TemplateGenericSecret,
			Fields: []vaultmodel.Field{
				{ID: "value", Value: rawJSON("selector-secret")},
				{ID: vaultmodel.EnvFieldID("value"), Value: rawJSON("SHARED_TOKEN")},
			},
		}},
	}
	if err := repo.Save(vault); err != nil {
		t.Fatalf("write selector vault: %v", err)
	}
	got := dynamicCompletions(t, []string{"--vault", path}, "secret", "info", "item_shared")
	if len(got) != 1 || got[0] != "item_shared" {
		t.Fatalf("id/title selector completions = %v, want one item_shared", got)
	}
}

func TestSecretCompletionsUseValueFreeVaultMetadata(t *testing.T) {
	secretValue := "never-suggest-this-secret-value"
	apiSecretValue := "never-suggest-this-api-secret"
	vault := writeLabeledVault(t, map[string]labeledItem{
		"alpha":        {env: "ALPHA_TOKEN", value: secretValue, labels: []string{"development", "shared"}},
		"alpha-backup": {env: "ALPHA_BACKUP_TOKEN", value: "another-secret-value", labels: []string{"development"}},
	})
	addAPICompletionItem(t, vault, apiSecretValue)
	flags := []string{"--vault", vault}
	var outputs []string

	for _, command := range [][]string{
		{"secret", "get", "item_alpha"},
		{"secret", "copy", "item_alpha"},
		{"secret", "info", "item_alpha"},
		{"secret", "update", "item_alpha"},
		{"secret", "rm", "item_alpha"},
	} {
		got := dynamicCompletions(t, flags, command...)
		if !hasCompletion(got, "item_alpha") {
			t.Errorf("%v completions = %v, want item_alpha", command, got)
		}
		outputs = append(outputs, got...)
	}

	selectors := dynamicCompletions(t, flags, "secret", "get", "item_alpha")
	if len(selectors) != 2 || !hasCompletion(selectors, "item_alpha") || !hasCompletion(selectors, "item_alpha-backup") {
		t.Errorf("prefix-filtered selectors = %v, want exactly the two matching IDs", selectors)
	}
	outputs = append(outputs, selectors...)
	if got := dynamicCompletions(t, flags, "secret", "get", "alpha", "item"); len(got) != 0 {
		t.Errorf("extra positional selector completions = %v, want none", got)
	}
	if got := dynamicCompletions(t, flags, "secret", "list", "alpha", "item"); len(got) != 0 {
		t.Errorf("extra list query completions = %v, want none", got)
	}

	for _, args := range [][]string{
		{"secret", "get", "alpha", "--field", "v"},
		{"secret", "update", "alpha", "--field", "v"},
		{"project", "add", "alpha.value", "--field", "v"},
	} {
		got := dynamicCompletions(t, flags, args...)
		if !hasCompletion(got, "value") {
			t.Errorf("%v field completions = %v, want value", args, got)
		}
		outputs = append(outputs, got...)
	}

	for _, args := range [][]string{
		{"secret", "get", "--env", "AL"},
		{"secret", "copy", "--env", "AL"},
		{"secret", "get", "--env", "ALPHA_TOKEN", "--field", "v"},
		{"secret", "copy", "--env", "ALPHA_TOKEN", "--field", "v"},
	} {
		got := dynamicCompletions(t, flags, args...)
		want := "ALPHA_TOKEN"
		if strings.Contains(strings.Join(args, " "), "--field") {
			want = "value"
		}
		if !hasCompletion(got, want) {
			t.Errorf("%v completions = %v, want %q", args, got, want)
		}
		outputs = append(outputs, got...)
	}
	got := dynamicCompletions(t, flags, "secret", "get", "--env", "API_B")
	if !hasCompletion(got, "API_BASE_URL") {
		t.Errorf("secondary env completions = %v, want API_BASE_URL", got)
	}
	outputs = append(outputs, got...)
	got = dynamicCompletions(t, flags, "secret", "get", "--env", "API_BASE_URL", "--field", "b")
	if !hasCompletion(got, "base_url") {
		t.Errorf("field completions for secondary env = %v, want base_url", got)
	}
	outputs = append(outputs, got...)

	for _, args := range [][]string{
		{"secret", "list", "alpha"},
		{"env", "export", "--label", "dev"},
		{"secret", "update", "alpha", "--add-label", "dev"},
		{"secret", "update", "alpha", "--remove-label", "dev"},
	} {
		got := dynamicCompletions(t, flags, args...)
		want := "alpha"
		if strings.Contains(strings.Join(args, " "), "--label") || strings.Contains(strings.Join(args, " "), "--add-label") || strings.Contains(strings.Join(args, " "), "--remove-label") {
			want = "development"
			if len(got) != 1 {
				t.Errorf("%v label completions = %v, want one deduplicated label", args, got)
			}
		} else if len(got) != 2 {
			t.Errorf("list query completions = %v, want the two prefix-matching titles", got)
		}
		if !hasCompletion(got, want) {
			t.Errorf("%v completions = %v, want %q", args, got, want)
		}
		outputs = append(outputs, got...)
	}
	got = dynamicCompletions(t, flags, "secret", "list", "test")
	if !hasCompletion(got, "test provider details") {
		t.Errorf("description query completions = %v, want matching public description", got)
	}
	outputs = append(outputs, got...)

	if strings.Contains(strings.Join(outputs, "\n"), secretValue) || strings.Contains(strings.Join(outputs, "\n"), "another-secret-value") || strings.Contains(strings.Join(outputs, "\n"), apiSecretValue) {
		t.Fatal("dynamic completion suggested a stored secret value")
	}
}

func addAPICompletionItem(t *testing.T, path, secretValue string) {
	t.Helper()
	repo := vaultjson.NewRepository(path)
	vault, err := repo.Load()
	if err != nil {
		t.Fatalf("load completion vault: %v", err)
	}
	vault.Items = append(vault.Items, vaultmodel.Item{
		ID:    "item_api",
		Title: "api", Description: "test provider details",
		Template: vaultmodel.TemplateAPICredential,
		Fields: []vaultmodel.Field{
			{ID: "credential", Value: rawJSON(secretValue)},
			{ID: vaultmodel.EnvFieldID("credential"), Value: rawJSON("API_TOKEN")},
			{ID: vaultmodel.ValidityFieldID("credential"), Value: rawJSON("unknown")},
			{ID: "base_url", Value: rawJSON("https://api.example.test")},
			{ID: vaultmodel.EnvFieldID("base_url"), Value: rawJSON("API_BASE_URL")},
		},
	})
	if err := repo.Save(vault); err != nil {
		t.Fatalf("write completion vault: %v", err)
	}
}

func TestSecretCreateCompletionsUseCompatibleOptions(t *testing.T) {
	got := dynamicCompletions(t, nil, "secret", "create", "alpha", "--template", "api_")
	if len(got) != 1 || got[0] != "api_credential" {
		t.Fatalf("template completions = %v, want api_credential", got)
	}

	got = dynamicCompletions(t, nil, "secret", "create", "alpha", "--template", "api_credential", "--profile", "op")
	if len(got) != 1 || got[0] != "openai" {
		t.Fatalf("profile completions = %v, want openai", got)
	}

	got = dynamicCompletions(t, nil, "secret", "create", "alpha", "--template", "generic_secret", "--profile", "o")
	if len(got) != 0 {
		t.Fatalf("incompatible profile completions = %v, want none", got)
	}

	got = dynamicCompletions(t, nil, "secret", "create", "alpha", "--template", "api", "--profile", "op")
	if len(got) != 1 || got[0] != "openai" {
		t.Fatalf("profile completions for template alias = %v, want openai", got)
	}

}

func TestCompletionWithMissingConfigAndVaultIsReadOnly(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "missing-config.yaml")
	vault := filepath.Join(dir, "missing-vault.json")
	got := dynamicCompletions(t, []string{"--config", config, "--vault", vault}, "secret", "get", "item")
	if len(got) != 0 {
		t.Fatalf("completion with missing files = %v, want none", got)
	}
	for _, path := range []string{config, vault} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("completion created or changed %s: stat error = %v", path, err)
		}
	}
}

func TestInvalidSelectorFieldCompletionIsEmpty(t *testing.T) {
	vault := writeLabeledVault(t, map[string]labeledItem{
		"alpha": {env: "ALPHA_TOKEN", value: "never-suggest-this-secret-value"},
	})
	got := dynamicCompletions(t, []string{"--vault", vault}, "secret", "get", "missing", "--field", "v")
	if len(got) != 0 {
		t.Fatalf("invalid-selector field completions = %v, want none", got)
	}
}

func dynamicCompletions(t *testing.T, flags []string, args ...string) []string {
	t.Helper()
	command := append(append([]string(nil), flags...), "__complete")
	command = append(command, args...)
	cmd := NewRootCmd(testApp())
	var output, diagnostics bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&diagnostics)
	cmd.SetArgs(command)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("completion %v: %v\n%s", args, err, diagnostics.String())
	}
	out := output.String()
	var values []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if value, _, found := strings.Cut(line, "\t"); found {
			line = value
		}
		values = append(values, line)
	}
	return values
}

func hasCompletion(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
