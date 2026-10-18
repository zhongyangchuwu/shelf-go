package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestSetSecretCreatesGenericMappedItem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	config := writeConfig(t, path)
	svc := testApp()
	targets, err := svc.SecretSetTargets(config, path, "NEW_TOKEN")
	if err != nil || len(targets) != 0 {
		t.Fatalf("initial targets = %v, %v", targets, err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("target lookup initialized storage: %v", err)
	}
	id, err := svc.SetSecret(config, path, SetSecretRequest{Env: "NEW_TOKEN", In: strings.NewReader("new-secret\r\n")})
	if err != nil {
		t.Fatal(err)
	}
	vault, err := vaultjson.NewRepository(path).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(vault.Items) != 1 || vault.Items[0].ID != id || vault.Items[0].Title != "NEW_TOKEN" || vault.Items[0].Template != vaultmodel.TemplateGenericSecret {
		t.Fatalf("created item identity = %#v", vault.Items)
	}
	value, err := svc.GetSecretField(config, path, GetSecretFieldRequest{Env: "NEW_TOKEN"})
	if err != nil || value != "new-secret" {
		t.Fatalf("mapped value = %q, %v", value, err)
	}
}

func TestSetSecretPreservesVaultAndCustomSchema(t *testing.T) {
	path, config, before := writeSetScenario(t)
	for _, update := range []struct{ env, field, value string }{
		{"SERVICE_TOKEN", "token", "new-primary"},
		{"SERVICE_BACKUP", "backup", "new-secondary"},
		{"SERVICE_URL", "endpoint", "https://new.example.test"},
	} {
		targets, err := testApp().SecretSetTargets(config, path, update.env)
		if err != nil || !reflect.DeepEqual(targets, []string{"stable_service." + update.field}) {
			t.Fatalf("targets for %s = %v, %v", update.env, targets, err)
		}
		if _, err := testApp().SetSecret(config, path, SetSecretRequest{Env: update.env, Target: targets[0], In: strings.NewReader(update.value)}); err != nil {
			t.Fatal(err)
		}
		if err := setFieldValueIfPresent(&before.Items[0], update.field, rawString(update.value)); err != nil {
			t.Fatal(err)
		}
		after, err := vaultjson.NewRepository(path).Load()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(after, before) {
			t.Fatalf("setting %s changed unrelated metadata, fields, organization, or definitions", update.env)
		}
	}
	catalog, err := testApp().SecretCatalog(config, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 1 || catalog[0].Env != "SERVICE_TOKEN" || !reflect.DeepEqual(catalog[0].EnvNames, []string{"SERVICE_BACKUP", "SERVICE_TOKEN", "SERVICE_URL"}) || catalog[0].EnvMappings != 3 {
		t.Fatalf("catalog env metadata = %#v", catalog)
	}
	names, err := testApp().SecretCompletions(config, path, SecretCompletionRequest{Target: CompleteSecretSetEnvNames})
	if err != nil || len(names) != 3 || !hasString(names, "SERVICE_TOKEN") || !hasString(names, "SERVICE_BACKUP") || !hasString(names, "SERVICE_URL") {
		t.Fatalf("primary and secondary env completions = %v, %v", names, err)
	}
	targets, err := testApp().SecretCompletions(config, path, SecretCompletionRequest{Target: CompleteSecretSetTargets, Env: "SERVICE_URL"})
	if err != nil || !reflect.DeepEqual(targets, []string{"stable_service.endpoint"}) {
		t.Fatalf("secondary target completions = %v, %v", targets, err)
	}
	encoded, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"new-primary", "new-secondary", "https://new.example.test"} {
		if bytes.Contains(encoded, []byte(value)) {
			t.Fatalf("catalog leaked a mapped value")
		}
	}
}

func TestSetSecretRejectsInvalidRequestsWithoutMutation(t *testing.T) {
	for _, tc := range []struct {
		name, env, target string
		input             io.Reader
	}{
		{name: "invalid env", env: "1INVALID", input: strings.NewReader("sensitive-input")},
		{name: "empty env", input: strings.NewReader("sensitive-input")},
		{name: "title not ID", env: "SERVICE_TOKEN", target: "Service account.token", input: strings.NewReader("sensitive-input")},
		{name: "unrelated target", env: "SERVICE_TOKEN", target: "stable_service.backup", input: strings.NewReader("sensitive-input")},
		{name: "structural target", env: "SERVICE_TOKEN", target: "stable_service.env_token", input: strings.NewReader("sensitive-input")},
		{name: "missing target", env: "SERVICE_TOKEN", target: "missing.token", input: strings.NewReader("sensitive-input")},
		{name: "invalid public URL", env: "SERVICE_URL", target: "stable_service.endpoint", input: strings.NewReader("sensitive-input")},
		{name: "failed input", env: "SERVICE_TOKEN", target: "stable_service.token", input: failingSetInput{}},
		{name: "nil input", env: "SERVICE_TOKEN", target: "stable_service.token"},
		{name: "ineligible JSON field", env: "SERVICE_TOKEN", target: "stable_service.notes", input: strings.NewReader("sensitive-input")},
		{name: "stale create", env: "SERVICE_TOKEN", input: strings.NewReader("sensitive-input")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path, config, _ := writeSetScenario(t)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = testApp().SetSecret(config, path, SetSecretRequest{Env: tc.env, Target: tc.target, In: tc.input})
			if err == nil || strings.Contains(err.Error(), "sensitive-input") {
				t.Fatalf("expected value-free error, got %v", err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("failed request mutated storage: %v", err)
			}
		})
	}
}

func TestSetSecretRejectsUnrelatedCreateCollisions(t *testing.T) {
	for _, identity := range []struct{ id, title string }{
		{"item_new_token", "Unrelated"}, {"other", "NEW_TOKEN"},
		{"NEW_TOKEN", "Unrelated"}, {"other", "item_new_token"},
	} {
		t.Run(identity.id+"-"+identity.title, func(t *testing.T) {
			path, config, vault := writeSetScenario(t)
			vault.Items[0].ID, vault.Items[0].Title = identity.id, identity.title
			repo := vaultjson.NewRepository(path)
			if err := repo.Save(vault); err != nil {
				t.Fatal(err)
			}
			_, err := testApp().SetSecret(config, path, SetSecretRequest{Env: "NEW_TOKEN", In: strings.NewReader("new-value")})
			if err == nil || !strings.Contains(err.Error(), "distinct title and --id") {
				t.Fatalf("collision error = %v", err)
			}
			after, err := repo.Load()
			if err != nil || !reflect.DeepEqual(after, vault) {
				t.Fatalf("collision mutated unrelated item: %v", err)
			}
		})
	}
}

func TestSetSecretRechecksMappingAfterInput(t *testing.T) {
	path, config, vault := writeSetScenario(t)
	reader := &changingSetInput{change: func() {
		if err := setFieldValueIfPresent(&vault.Items[0], "env_token", rawString("RENAMED_TOKEN")); err != nil {
			t.Fatal(err)
		}
		if err := vaultjson.NewRepository(path).Save(vault); err != nil {
			t.Fatal(err)
		}
	}}
	_, err := testApp().SetSecret(config, path, SetSecretRequest{Env: "SERVICE_TOKEN", Target: "stable_service.token", In: reader})
	if err == nil {
		t.Fatal("stale mapping was accepted")
	}
	after, err := vaultjson.NewRepository(path).Load()
	if err != nil || !reflect.DeepEqual(after, vault) {
		t.Fatalf("stale selection overwrote a concurrent change: %v", err)
	}
}

func writeSetScenario(t *testing.T) (string, string, vaultmodel.Vault) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.json")
	vault := vaultmodel.Vault{
		Version: vaultmodel.CurrentVersion,
		Templates: []vaultmodel.Template{{ID: "service_custom", Name: "Service", PrimarySecret: "token", Fields: []vaultmodel.FieldTemplate{
			{ID: "token", Label: "Token", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, Required: true},
			{ID: "env_token", Label: "Token env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
			{ID: "backup", Label: "Backup", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret},
			{ID: "env_backup", Label: "Backup env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
			{ID: "endpoint", Label: "Endpoint", Type: vaultmodel.FieldTypeURL, Sensitivity: vaultmodel.SensitivityPublic},
			{ID: "env_endpoint", Label: "Endpoint env", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
			{ID: "notes", Label: "Notes", Type: vaultmodel.FieldTypeJSON, Sensitivity: vaultmodel.SensitivityPublic},
		}}},
		Profiles: []vaultmodel.Profile{{ID: "service_profile", Name: "Service profile", Template: "service_custom", Defaults: map[string]json.RawMessage{"endpoint": rawString("https://default.example.test")}}},
		Items: []vaultmodel.Item{{ID: "stable_service", Title: "Service account", Template: "service_custom", Labels: []string{"work", "team"}, Description: "Keep this organization", CreatedAt: "2025-01-01", UpdatedAt: "2025-02-01", Fields: []vaultmodel.Field{
			{ID: "token", Value: rawString("old-primary")}, {ID: "env_token", Value: rawString("SERVICE_TOKEN")},
			{ID: "backup", Value: rawString("old-secondary")}, {ID: "env_backup", Value: rawString("SERVICE_BACKUP")},
			{ID: "endpoint", Value: rawString("https://old.example.test")}, {ID: "env_endpoint", Value: rawString("SERVICE_URL")},
			{ID: "notes", Value: json.RawMessage(`{"keep":true}`)},
		}}},
	}
	if err := vaultjson.NewRepository(path).Save(vault); err != nil {
		t.Fatal(err)
	}
	// Repository serialization may normalize JSON formatting.
	loaded, err := vaultjson.NewRepository(path).Load()
	if err != nil {
		t.Fatal(err)
	}
	return path, writeConfig(t, path), loaded
}

type failingSetInput struct{}

func (failingSetInput) Read([]byte) (int, error) {
	return 0, errors.New("sensitive-input")
}

type changingSetInput struct {
	change func()
	done   bool
}

func (r *changingSetInput) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	r.done = true
	r.change()
	return copy(p, "new-secret"), io.EOF
}

func TestSetSecretSaveFailureDoesNotMutateOrLeak(t *testing.T) {
	path, config, _ := writeSetScenario(t)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// A directory at the backup path makes the real repository save fail,
	// without relying on permissions (which behave differently as root).
	if err := os.Mkdir(path+".bak", 0o700); err != nil {
		t.Fatal(err)
	}
	_, err = testApp().SetSecret(config, path, SetSecretRequest{Env: "SERVICE_TOKEN", Target: "stable_service.token", In: strings.NewReader("sensitive-save-value")})
	if err == nil || strings.Contains(err.Error(), "sensitive-save-value") {
		t.Fatalf("expected value-free save failure, got %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("failed save mutated vault: %v", err)
	}
}

func TestSetSecretRetainsMetadataChangedDuringInput(t *testing.T) {
	path, config, vault := writeSetScenario(t)
	reader := &changingSetInput{change: func() {
		vault.Items[0].Labels = []string{"concurrently-organized"}
		vault.Items[0].Description = "New description"
		if err := vaultjson.NewRepository(path).Save(vault); err != nil {
			t.Fatal(err)
		}
	}}
	if _, err := testApp().SetSecret(config, path, SetSecretRequest{Env: "SERVICE_TOKEN", Target: "stable_service.token", In: reader}); err != nil {
		t.Fatal(err)
	}
	if err := setFieldValueIfPresent(&vault.Items[0], "token", rawString("new-secret")); err != nil {
		t.Fatal(err)
	}
	after, err := vaultjson.NewRepository(path).Load()
	if err != nil || !reflect.DeepEqual(after, vault) {
		t.Fatalf("set lost a concurrent metadata change: %v", err)
	}
}
