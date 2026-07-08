package vaultmigrate_test

import (
	"encoding/json"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/project"
	"github.com/zhongyangchuwu/shelf-go/internal/vault"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmigrate"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

// ---------------------------------------------------------------------------
// ConvertVault
// ---------------------------------------------------------------------------

var valSecret = json.RawMessage(`"s3cret"`)
var valAPIKey = json.RawMessage(`"api-key-123"`)

func TestConvertVault_HappyPath(t *testing.T) {
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"prod/db:password": {
				Value:       valSecret,
				Description: "database password",
				Tags:        []string{"db"},
			},
			"prod/api:key": {
				Value:       valAPIKey,
				Env:         "API_KEY",
				Description: "API key for production",
				Tags:        []string{"api", "prod"},
			},
		},
	}

	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault failed: %v", err)
	}

	if result.Vault.Version != vaultmodel.CurrentVersion {
		t.Fatalf("vault version = %d, want %d", result.Vault.Version, vaultmodel.CurrentVersion)
	}
	if len(result.Vault.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(result.Vault.Items))
	}
	if len(result.Paths) != 2 {
		t.Fatalf("got %d paths, want 2", len(result.Paths))
	}

	// Items are sorted by path: "prod/api:key" < "prod/db:password"
	item0, item1 := result.Vault.Items[0], result.Vault.Items[1]
	if item0.Title != "prod/api:key" || item1.Title != "prod/db:password" {
		t.Fatalf("items not sorted by path: got %q, %q", item0.Title, item1.Title)
	}

	// --- item0: explicit env "API_KEY" ---
	if item0.Template != vaultmodel.TemplateGenericSecret {
		t.Fatalf("item0 template = %q, want %q", item0.Template, vaultmodel.TemplateGenericSecret)
	}
	if len(item0.Fields) != 2 {
		t.Fatalf("item0 has %d fields, want 2 (value + value_env)", len(item0.Fields))
	}
	f0 := item0.Fields[0]
	if f0.ID != "value" {
		t.Fatalf("item0 field id = %q, want %q", f0.ID, "value")
	}
	if f0.Label != "Value" {
		t.Fatalf("item0 field label = %q", f0.Label)
	}
	if f0.Type != vaultmodel.FieldTypeConcealed {
		t.Fatalf("item0 field type = %q, want %q", f0.Type, vaultmodel.FieldTypeConcealed)
	}
	if f0.Sensitivity != vaultmodel.SensitivitySecret {
		t.Fatalf("item0 field sensitivity = %q", f0.Sensitivity)
	}
	if f0.Purpose != vaultmodel.PurposePrimarySecret {
		t.Fatalf("item0 field purpose = %q", f0.Purpose)
	}
	if string(f0.Value) != string(valAPIKey) {
		t.Fatalf("item0 field value = %q, want %q", string(f0.Value), string(valAPIKey))
	}
	// Env name lives in PathMapping, not Field.Bindings

	// Aliases from path
	if len(item0.Aliases) != 1 || item0.Aliases[0] != "prod/api:key" {
		t.Fatalf("item0 aliases = %v, want [prod/api:key]", item0.Aliases)
	}
	// Notes from description
	if item0.Notes != "API key for production" {
		t.Fatalf("item0 notes = %q, want %q", item0.Notes, "API key for production")
	}
	// Labels from tags
	if len(item0.Labels) != 2 || item0.Labels[0] != "api" || item0.Labels[1] != "prod" {
		t.Fatalf("item0 labels = %v, want [api prod]", item0.Labels)
	}

	// --- item1: derived env ---
	f1 := item1.Fields[0]
	// Env name lives in PathMapping, not Field.Bindings
	if string(f1.Value) != string(valSecret) {
		t.Fatalf("item1 field value = %q, want %q", string(f1.Value), string(valSecret))
	}
	if len(item1.Aliases) != 1 || item1.Aliases[0] != "prod/db:password" {
		t.Fatalf("item1 aliases = %v, want [prod/db:password]", item1.Aliases)
	}
	if item1.Notes != "database password" {
		t.Fatalf("item1 notes = %q, want %q", item1.Notes, "database password")
	}
	if len(item1.Labels) != 1 || item1.Labels[0] != "db" {
		t.Fatalf("item1 labels = %v, want [db]", item1.Labels)
	}

	// --- PathMapping entries ---
	pm0, ok0 := result.Paths["prod/api:key"]
	if !ok0 {
		t.Fatal("missing path mapping for prod/api:key")
	}
	if pm0.ItemID == "" {
		t.Fatal("path mapping item id is empty")
	}
	if pm0.FieldID != "value" {
		t.Fatalf("path mapping field id = %q", pm0.FieldID)
	}
	if pm0.EnvName != "API_KEY" {
		t.Fatalf("path mapping env = %q", pm0.EnvName)
	}
	if pm0.Title != "prod/api:key" {
		t.Fatalf("path mapping title = %q", pm0.Title)
	}
	if len(pm0.Labels) != 2 || pm0.Labels[0] != "api" || pm0.Labels[1] != "prod" {
		t.Fatalf("path mapping labels = %v", pm0.Labels)
	}

	pm1, ok1 := result.Paths["prod/db:password"]
	if !ok1 {
		t.Fatal("missing path mapping for prod/db:password")
	}
	if pm1.EnvName != "PROD_DB_PASSWORD" {
		t.Fatalf("path mapping env = %q", pm1.EnvName)
	}
	if pm1.Title != "prod/db:password" {
		t.Fatalf("path mapping title = %q", pm1.Title)
	}
}

func TestConvertVault_DeterministicIDs(t *testing.T) {
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"prod/db:password": {Value: valSecret},
		},
	}

	r1, err1 := vaultmigrate.ConvertVault(data)
	if err1 != nil {
		t.Fatalf("first ConvertVault: %v", err1)
	}
	r2, err2 := vaultmigrate.ConvertVault(data)
	if err2 != nil {
		t.Fatalf("second ConvertVault: %v", err2)
	}

	if r1.Vault.Items[0].ID != r2.Vault.Items[0].ID {
		t.Fatal("item IDs differ between runs (want deterministic)")
	}
	id1 := r1.Vault.Items[0].ID
	if len(id1) < 10 || id1[:5] != "item_" {
		t.Fatalf("item ID format unexpected: %q", id1)
	}
}

func TestConvertVault_ValueIsCopied(t *testing.T) {
	orig := json.RawMessage(`"secret-value"`)
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"app:db_pass": {Value: orig},
		},
	}

	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}

	// Modify the original input
	copy(orig, []byte(`"tampered"`))

	if string(result.Vault.Items[0].Fields[0].Value) != `"secret-value"` {
		t.Fatal("v2 field value changed after modifying source (want copy)")
	}
}

func TestConvertVault_NoAPICredentialInference(t *testing.T) {
	// Even with a value that looks like an API key, the template is always generic_secret
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"prod:stripe_key": {Value: json.RawMessage(`"sk_live_abc123"`)},
		},
	}

	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}

	if result.Vault.Items[0].Template != vaultmodel.TemplateGenericSecret {
		t.Fatalf("template = %q, want %q (no inference)", result.Vault.Items[0].Template, vaultmodel.TemplateGenericSecret)
	}
}

func TestConvertVault_InvalidVersion(t *testing.T) {
	data := vault.Data{
		Version: 0,
		Secrets: map[string]vault.Secret{"a:b": {Value: valSecret}},
	}

	_, err := vaultmigrate.ConvertVault(data)
	if err == nil {
		t.Fatal("expected error for version 0")
	}
}

func TestConvertVault_InvalidPath(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{"no colon", "justkey"},
		{"empty group", ":key"},
		{"empty key", "group:"},
		{"slash in key", "group:key/sub"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := vault.Data{
				Version: vault.CurrentVersion,
				Secrets: map[string]vault.Secret{tc.path: {Value: valSecret}},
			}
			_, err := vaultmigrate.ConvertVault(data)
			if err == nil {
				t.Fatalf("expected error for path %q", tc.path)
			}
		})
	}
}

func TestConvertVault_InvalidSecret(t *testing.T) {
	cases := []struct {
		name   string
		secret vault.Secret
	}{
		{"empty value", vault.Secret{Value: json.RawMessage(nil)}},
		{"non-JSON value", vault.Secret{Value: json.RawMessage(`not json`)}},
		{"bad env", vault.Secret{Value: valSecret, Env: "not var"}},
		{"empty tag", vault.Secret{Value: valSecret, Tags: []string{""}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := vault.Data{
				Version: vault.CurrentVersion,
				Secrets: map[string]vault.Secret{"a:b": tc.secret},
			}
			_, err := vaultmigrate.ConvertVault(data)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

func TestConvertVault_EnvDerivation(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"prod/api:key", "PROD_API_KEY"},
		{"staging:db", "STAGING_DB"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			data := vault.Data{
				Version: vault.CurrentVersion,
				Secrets: map[string]vault.Secret{tc.path: {Value: valSecret}},
			}
			result, err := vaultmigrate.ConvertVault(data)
			if err != nil {
				t.Fatalf("ConvertVault: %v", err)
			}
			got := result.Paths[tc.path].EnvName
			if got != tc.want {
				t.Fatalf("env name = %q, want %q (from PathMapping)", got, tc.want)
			}
		})
	}
}

func TestConvertVault_EnvNameFieldExists(t *testing.T) {
	// After Phase 38, env names live in a value_env public field,
	// not in Field.Bindings (which no longer exists at the type level).
	// Verify the value_env field is created correctly.
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"prod/api:key": {
				Value: valSecret,
				Env:   "MY_SECRET_KEY",
			},
		},
	}

	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault failed: %v", err)
	}

	item := result.Vault.Items[0]
	if len(item.Fields) != 2 {
		t.Fatalf("expected 2 fields (value + value_env), got %d", len(item.Fields))
	}

	// Find the value_env field
	var envField *vaultmodel.Field
	for i := range item.Fields {
		if item.Fields[i].ID == "value_env" {
			envField = &item.Fields[i]
			break
		}
	}
	if envField == nil {
		t.Fatal("value_env field not found on converted item")
	}

	if envField.Type != vaultmodel.FieldTypeString {
		t.Fatalf("value_env type = %q, want %q", envField.Type, vaultmodel.FieldTypeString)
	}
	if envField.Sensitivity != vaultmodel.SensitivityPublic {
		t.Fatalf("value_env sensitivity = %q, want %q", envField.Sensitivity, vaultmodel.SensitivityPublic)
	}
	if envField.Purpose != vaultmodel.PurposeEnvName {
		t.Fatalf("value_env purpose = %q, want %q", envField.Purpose, vaultmodel.PurposeEnvName)
	}
	var envName string
	if err := json.Unmarshal(envField.Value, &envName); err != nil {
		t.Fatalf("value_env value not a valid JSON string: %v", err)
	}
	if envName != "MY_SECRET_KEY" {
		t.Fatalf("value_env = %q, want %q", envName, "MY_SECRET_KEY")
	}
}

func TestConvertVault_ValueIsConcealed(t *testing.T) {
	// The migrated primary secret field must be concealed and secret-sensitivity.
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"app:secret": {Value: json.RawMessage(`"hidden-value"`)},
		},
	}
	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}
	f := result.Vault.Items[0].Fields[0]
	if f.Type != vaultmodel.FieldTypeConcealed {
		t.Fatalf("field type = %q, want %q", f.Type, vaultmodel.FieldTypeConcealed)
	}
	if f.Sensitivity != vaultmodel.SensitivitySecret {
		t.Fatalf("field sensitivity = %q, want %q", f.Sensitivity, vaultmodel.SensitivitySecret)
	}
	if f.Purpose != vaultmodel.PurposePrimarySecret {
		t.Fatalf("field purpose = %q, want %q", f.Purpose, vaultmodel.PurposePrimarySecret)
	}
	// Value must be set (non-empty)
	if len(f.Value) == 0 {
		t.Fatal("primary field value is empty after migration")
	}
}

func TestConvertVault_AliasesPreserveOldPath(t *testing.T) {
	// Old v1 path must appear in the item's Aliases list.
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"prod/staging:db_password": {Value: valSecret},
		},
	}
	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}
	item := result.Vault.Items[0]
	if len(item.Aliases) != 1 || item.Aliases[0] != "prod/staging:db_password" {
		t.Fatalf("aliases = %v, want [prod/staging:db_password]", item.Aliases)
	}
}

func TestConvertVault_DescriptionAndTagsPreserved(t *testing.T) {
	// v1 description → item Notes; v1 Tags → item Labels.
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"app:creds": {
				Value:       valSecret,
				Description: "my app credentials",
				Tags:        []string{"app", "prod", "us-east"},
			},
		},
	}
	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}
	item := result.Vault.Items[0]
	if item.Notes != "my app credentials" {
		t.Fatalf("notes = %q, want %q", item.Notes, "my app credentials")
	}
	if len(item.Labels) != 3 ||
		item.Labels[0] != "app" ||
		item.Labels[1] != "prod" ||
		item.Labels[2] != "us-east" {
		t.Fatalf("labels = %v, want [app prod us-east]", item.Labels)
	}
}

func TestConvertVault_ExplicitEnvNameInPathMapping(t *testing.T) {
	// Explicit Env on a v1 secret must appear in PathMapping.EnvName.
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"prod:key": {
				Value: valSecret,
				Env:   "MY_CUSTOM_VAR",
			},
		},
	}
	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}
	if result.Paths["prod:key"].EnvName != "MY_CUSTOM_VAR" {
		t.Fatalf("env name = %q, want %q", result.Paths["prod:key"].EnvName, "MY_CUSTOM_VAR")
	}
}

func TestConvertVault_DerivedEnvNameInPathMapping(t *testing.T) {
	// Derived env (from path) must appear in PathMapping.EnvName.
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"staging/api:token": {Value: valSecret},
		},
	}
	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}
	// Derived from "staging/api:token" -> "STAGING_API_TOKEN"
	if result.Paths["staging/api:token"].EnvName != "STAGING_API_TOKEN" {
		t.Fatalf("env name = %q, want %q", result.Paths["staging/api:token"].EnvName, "STAGING_API_TOKEN")
	}
}

func TestConvertVault_PathMappingIsValueFree(t *testing.T) {
	// PathMapping must not contain secret values.
	secretVal := json.RawMessage(`"supersecret"`)
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"app:secret": {Value: secretVal},
		},
	}
	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}
	for path, pm := range result.Paths {
		if pm.EnvName == "" {
			t.Fatalf("PathMapping[%q].EnvName is empty, should be set by migration", path)
		}
		if pm.ItemID == "" {
			t.Fatalf("PathMapping[%q].ItemID is empty", path)
		}
		if pm.FieldID != "value" {
			t.Fatalf("PathMapping[%q].FieldID = %q, want %q", path, pm.FieldID, "value")
		}
		if pm.Title != path {
			t.Fatalf("PathMapping[%q].Title = %q, want %q", path, pm.Title, path)
		}
	}
}

func TestConvertVault_SameSecretValuePreserved(t *testing.T) {
	raw := json.RawMessage(`"hello world"`)
	data := vault.Data{
		Version: vault.CurrentVersion,
		Secrets: map[string]vault.Secret{
			"app:greeting": {Value: raw},
		},
	}
	result, err := vaultmigrate.ConvertVault(data)
	if err != nil {
		t.Fatalf("ConvertVault: %v", err)
	}
	if string(result.Vault.Items[0].Fields[0].Value) != string(raw) {
		t.Fatalf("value = %q, want %q", string(result.Vault.Items[0].Fields[0].Value), string(raw))
	}
}

// ---------------------------------------------------------------------------
// ConvertProjectManifest
// ---------------------------------------------------------------------------

func pathMapping(path, env string, labels ...string) vaultmigrate.PathMapping {
	return vaultmigrate.PathMapping{
		OldPath: path,
		ItemID:  "item_" + path,
		FieldID: "value",
		EnvName: env,
		Title:   path,
		Labels:  labels,
	}
}

func TestConvertProjectManifest_PathEntry_ValueFreeBinding(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Path: "prod/db:password"},
		},
	}
	paths := map[string]vaultmigrate.PathMapping{
		"prod/db:password": pathMapping("prod/db:password", "DB_PASS"),
	}

	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}

	bindings := result.Manifest.Bindings
	if len(bindings) != 1 {
		t.Fatalf("got %d bindings, want 1", len(bindings))
	}
	b := bindings[0]
	if b.Item != "item_prod/db:password" {
		t.Fatalf("item = %q", b.Item)
	}
	if b.Field != "value" {
		t.Fatalf("field = %q", b.Field)
	}
	if b.Env != "DB_PASS" {
		t.Fatalf("env = %q, want %q", b.Env, "DB_PASS")
	}
	if !b.Required {
		t.Fatal("binding should be required by default")
	}
}

func TestConvertProjectManifest_PathEntry_ExplicitEnv(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Path: "prod/db:password", Env: "CUSTOM_ENV"},
		},
	}
	paths := map[string]vaultmigrate.PathMapping{
		"prod/db:password": pathMapping("prod/db:password", "DB_PASS"),
	}

	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}

	if result.Manifest.Bindings[0].Env != "CUSTOM_ENV" {
		t.Fatalf("env = %q, want %q", result.Manifest.Bindings[0].Env, "CUSTOM_ENV")
	}
}

func TestConvertProjectManifest_PathEntry_RequiredFalse(t *testing.T) {
	f := false
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Path: "app:key", Required: &f},
		},
	}
	paths := map[string]vaultmigrate.PathMapping{
		"app:key": pathMapping("app:key", "APP_KEY"),
	}

	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}

	if result.Manifest.Bindings[0].Required {
		t.Fatal("expected Required=false for explicitly optional entry")
	}
}

func TestConvertProjectManifest_PrefixExpansion_Sorted(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Prefix: "prod"},
		},
	}
	// Deliberately add in non-alphabetical order
	paths := map[string]vaultmigrate.PathMapping{
		"prod/zzz:key": pathMapping("prod/zzz:key", "ZZZ_KEY"),
		"prod/aaa:db":  pathMapping("prod/aaa:db", "AAA_DB"),
	}
	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}

	bindings := result.Manifest.Bindings
	if len(bindings) != 2 {
		t.Fatalf("got %d bindings, want 2", len(bindings))
	}
	// Sorted by OldPath: "prod/aaa:db" < "prod/zzz:key"
	if bindings[0].Item != "item_prod/aaa:db" || bindings[1].Item != "item_prod/zzz:key" {
		t.Fatalf("bindings not sorted: got items %q, %q", bindings[0].Item, bindings[1].Item)
	}
}

func TestConvertProjectManifest_Prefix_EnvFromMapping(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Prefix: "prod/db"},
		},
	}
	paths := map[string]vaultmigrate.PathMapping{
		"prod/db:password": pathMapping("prod/db:password", "DB_PASS"),
	}
	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}
	if len(result.Manifest.Bindings) != 1 {
		t.Fatalf("got %d bindings, want 1", len(result.Manifest.Bindings))
	}
	// Prefix entries use mapping.EnvName for binding
	if result.Manifest.Bindings[0].Env != "DB_PASS" {
		t.Fatalf("env = %q, want %q", result.Manifest.Bindings[0].Env, "DB_PASS")
	}
}

func TestConvertProjectManifest_TagExpansion_Sorted(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Tags: []string{"prod"}},
		},
	}
	// Non-alphabetical order again
	paths := map[string]vaultmigrate.PathMapping{
		"prod/zzz:key": pathMapping("prod/zzz:key", "ZZZ_KEY", "prod"),
		"prod/aaa:db":  pathMapping("prod/aaa:db", "AAA_DB", "prod"),
	}
	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}

	bindings := result.Manifest.Bindings
	if len(bindings) != 2 {
		t.Fatalf("got %d bindings, want 2", len(bindings))
	}
	if bindings[0].Item != "item_prod/aaa:db" || bindings[1].Item != "item_prod/zzz:key" {
		t.Fatalf("bindings not sorted: got items %q, %q", bindings[0].Item, bindings[1].Item)
	}
}

func TestConvertProjectManifest_Tag_EnvFromMapping(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Tags: []string{"api"}},
		},
	}
	paths := map[string]vaultmigrate.PathMapping{
		"prod:key": pathMapping("prod:key", "PROD_KEY", "api"),
	}
	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}
	if len(result.Manifest.Bindings) != 1 {
		t.Fatalf("got %d bindings, want 1", len(result.Manifest.Bindings))
	}
	if result.Manifest.Bindings[0].Env != "PROD_KEY" {
		t.Fatalf("env = %q, want %q", result.Manifest.Bindings[0].Env, "PROD_KEY")
	}
}

func TestConvertProjectManifest_EmptyPrefixWarning(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Prefix: "nope"},
		},
	}

	result, err := vaultmigrate.ConvertProjectManifest(manifest, map[string]vaultmigrate.PathMapping{})
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("got %d warnings, want 1", len(result.Warnings))
	}
	if result.Warnings[0].Selector != "nope" {
		t.Fatalf("warning selector = %q", result.Warnings[0].Selector)
	}
	if len(result.Manifest.Bindings) != 0 {
		t.Fatalf("expected 0 bindings, got %d", len(result.Manifest.Bindings))
	}
}

func TestConvertProjectManifest_EmptyTagWarning(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Tags: []string{"nonexistent"}},
		},
	}

	result, err := vaultmigrate.ConvertProjectManifest(manifest, map[string]vaultmigrate.PathMapping{})
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("got %d warnings, want 1", len(result.Warnings))
	}
	if result.Warnings[0].Selector != "nonexistent" {
		t.Fatalf("warning selector = %q", result.Warnings[0].Selector)
	}
	if len(result.Manifest.Bindings) != 0 {
		t.Fatalf("expected 0 bindings, got %d", len(result.Manifest.Bindings))
	}
}

func TestConvertProjectManifest_MissingPathMapping(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Path: "prod/db:password"},
		},
	}

	_, err := vaultmigrate.ConvertProjectManifest(manifest, map[string]vaultmigrate.PathMapping{})
	if err == nil {
		t.Fatal("expected error for missing path mapping")
	}
}

func TestConvertProjectManifest_DuplicateBindingDedup(t *testing.T) {
	// Path entry + prefix entry both resolve to the same item:field:env binding
	f := false
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Path: "prod/db:password", Required: &f},
			{Prefix: "prod/db"},
		},
	}
	paths := map[string]vaultmigrate.PathMapping{
		"prod/db:password": pathMapping("prod/db:password", "DB_PASS"),
	}

	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}

	// Two entries, but only one unique binding -> dedup to 1
	if len(result.Manifest.Bindings) != 1 {
		t.Fatalf("got %d bindings (expected dedup to 1)", len(result.Manifest.Bindings))
	}
}

func TestConvertProjectManifest_TitleSnapshot(t *testing.T) {
	manifest := project.Manifest{
		Version: project.CurrentVersion,
		Secrets: []project.Entry{
			{Path: "prod:key"},
		},
	}
	paths := map[string]vaultmigrate.PathMapping{
		"prod:key": {OldPath: "prod:key", ItemID: "item_1", FieldID: "value", EnvName: "PROD_KEY", Title: "prod:key", Labels: []string{"prod"}},
	}

	result, err := vaultmigrate.ConvertProjectManifest(manifest, paths)
	if err != nil {
		t.Fatalf("ConvertProjectManifest: %v", err)
	}

	if result.Manifest.Bindings[0].Title != "prod:key" {
		t.Fatalf("title = %q, want %q", result.Manifest.Bindings[0].Title, "prod:key")
	}
}

func TestConvertProjectManifest_BadProjectManifest(t *testing.T) {
	manifest := project.Manifest{
		Version: 0, // unsupported version
		Secrets: []project.Entry{
			{Path: "a:b"},
		},
	}
	_, err := vaultmigrate.ConvertProjectManifest(manifest, nil)
	if err == nil {
		t.Fatal("expected error for bad project manifest")
	}
}

// ---------------------------------------------------------------------------
// ValidateProjectManifest
// ---------------------------------------------------------------------------

func TestValidateProjectManifest_OK(t *testing.T) {
	m := vaultmigrate.ProjectManifest{
		Version: vaultmigrate.ProjectManifestVersion,
		Bindings: []vaultmigrate.ProjectBinding{
			{Item: "i1", Field: "f1", Env: "MY_ENV"},
		},
	}
	if err := vaultmigrate.ValidateProjectManifest(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateProjectManifest_BadVersion(t *testing.T) {
	m := vaultmigrate.ProjectManifest{Version: 0}
	err := vaultmigrate.ValidateProjectManifest(m)
	if err == nil {
		t.Fatal("expected error for bad version")
	}
}

func TestValidateProjectManifest_EmptyItem(t *testing.T) {
	m := vaultmigrate.ProjectManifest{
		Version: vaultmigrate.ProjectManifestVersion,
		Bindings: []vaultmigrate.ProjectBinding{
			{Field: "f1", Env: "MY_ENV"},
		},
	}
	err := vaultmigrate.ValidateProjectManifest(m)
	if err == nil {
		t.Fatal("expected error for empty item")
	}
}

func TestValidateProjectManifest_EmptyField(t *testing.T) {
	m := vaultmigrate.ProjectManifest{
		Version: vaultmigrate.ProjectManifestVersion,
		Bindings: []vaultmigrate.ProjectBinding{
			{Item: "i1", Env: "MY_ENV"},
		},
	}
	err := vaultmigrate.ValidateProjectManifest(m)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestValidateProjectManifest_EmptyEnv(t *testing.T) {
	m := vaultmigrate.ProjectManifest{
		Version: vaultmigrate.ProjectManifestVersion,
		Bindings: []vaultmigrate.ProjectBinding{
			{Item: "i1", Field: "f1"},
		},
	}
	err := vaultmigrate.ValidateProjectManifest(m)
	if err == nil {
		t.Fatal("expected error for empty env")
	}
}

func TestValidateProjectManifest_InvalidEnvName(t *testing.T) {
	m := vaultmigrate.ProjectManifest{
		Version: vaultmigrate.ProjectManifestVersion,
		Bindings: []vaultmigrate.ProjectBinding{
			{Item: "i1", Field: "f1", Env: "not var name"},
		},
	}
	err := vaultmigrate.ValidateProjectManifest(m)
	if err == nil {
		t.Fatal("expected error for invalid env name")
	}
}

func TestValidateProjectManifest_DuplicateBinding(t *testing.T) {
	m := vaultmigrate.ProjectManifest{
		Version: vaultmigrate.ProjectManifestVersion,
		Bindings: []vaultmigrate.ProjectBinding{
			{Item: "i1", Field: "f1", Env: "MY_ENV"},
			{Item: "i1", Field: "f1", Env: "MY_ENV"},
		},
	}
	err := vaultmigrate.ValidateProjectManifest(m)
	if err == nil {
		t.Fatal("expected error for duplicate binding")
	}
}
