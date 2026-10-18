package cli

import (
	"path/filepath"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestCompletionUsesCustomTemplateAndProfileAliases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	vault := vaultmodel.Vault{
		Version: vaultmodel.CurrentVersion,
		Templates: []vaultmodel.Template{{
			ID: "service_token", Name: "Service Token", Aliases: []string{"service"}, PrimarySecret: "token",
			Fields: []vaultmodel.FieldTemplate{{ID: "token", Label: "Token", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Required: true}},
		}},
		Profiles: []vaultmodel.Profile{{ID: "service_personal", Name: "Personal", Aliases: []string{"personal"}, Template: "service_token"}},
	}
	if err := vaultjson.NewRepository(path).Save(vault); err != nil {
		t.Fatal(err)
	}
	flags := []string{"--vault", path}
	profiles := dynamicCompletions(t, flags, "secret", "create", "example", "--template", "service", "--profile", "per")
	if len(profiles) != 1 || profiles[0] != "personal" {
		t.Fatalf("custom profile alias completions = %v", profiles)
	}
	templates := dynamicCompletions(t, flags, "secret", "create", "example", "--profile", "personal", "--template", "service")
	if len(templates) != 2 || !hasCompletion(templates, "service_token") || !hasCompletion(templates, "service") {
		t.Fatalf("custom compatible template completions = %v", templates)
	}
	if got := dynamicCompletions(t, flags, "secret", "create", "example", "--template", "generic", "--profile", "per"); len(got) != 0 {
		t.Fatalf("incompatible custom profile completions = %v", got)
	}
}

func TestCompletionExcludesAmbiguousEnvsAndRestrictsProjectFields(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{
		"alpha": {env: "SHARED_TOKEN", value: "alpha-secret", labels: []string{"personal"}},
		"beta":  {env: "SHARED_TOKEN", value: "beta-secret", labels: []string{"work"}},
	})
	flags := []string{"--vault", path}
	if got := dynamicCompletions(t, flags, "secret", "get", "--env", "SHARED"); len(got) != 0 {
		t.Fatalf("ambiguous env completions = %v", got)
	}
	if got := dynamicCompletions(t, flags, "secret", "get", "--env", "SHARED_TOKEN", "--field", "v"); len(got) != 0 {
		t.Fatalf("ambiguous env field completions = %v", got)
	}
	fields := dynamicCompletions(t, flags, "secret", "update", "alpha", "--field", "env_")
	if len(fields) != 1 || fields[0] != "env_value" {
		t.Fatalf("editable mapping field completions = %v", fields)
	}
	if got := dynamicCompletions(t, flags, "project", "add", "alpha", "--field", "env_"); len(got) != 0 {
		t.Fatalf("non-projectable mapping field completions = %v", got)
	}
	labels := dynamicCompletions(t, flags, "secret", "update", "alpha", "--remove-label", "")
	if len(labels) != 1 || labels[0] != "personal" {
		t.Fatalf("removable labels for selected item = %v", labels)
	}
}
