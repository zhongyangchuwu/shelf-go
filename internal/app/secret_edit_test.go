// Whole-item editing exercises the real app, canonical model and atomic repository.
package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestSecretEditWholeItemPreservesMetadata(t *testing.T) {
	path, config, before := writeEditScenario(t)
	app := testApp()
	session, err := app.BeginSecretEdit(config, path, SecretEditRequest{Selector: "Service account"})
	if err != nil {
		t.Fatal(err)
	}
	var initial struct {
		ID       string                     `json:"id"`
		Template string                     `json:"template"`
		Fields   map[string]json.RawMessage `json:"fields"`
	}
	if err := json.Unmarshal(session.Document, &initial); err != nil {
		t.Fatal(err)
	}
	if initial.ID != before.Items[0].ID || initial.Template != before.Items[0].Template || string(initial.Fields["token"]) != `"old-primary"` {
		t.Fatal("draft omitted identity, schema context, or stored secret")
	}
	if _, exists := initial.Fields["optional"]; exists {
		t.Fatal("draft invented a value for an unset field")
	}
	if bytes.Contains(session.Document, []byte("created_at")) || bytes.Contains(session.Document, []byte("updated_at")) {
		t.Fatal("draft exposes readonly timestamps as editable keys")
	}
	changed, err := app.SaveSecretEdit(session, []byte(`{
		"id":"stable_service", "template":"service_custom",
		"title":"  Renamed service  ", "description":"  New organization  ", "labels":[" work ","edited","work"],
		"fields":{"token":"new-primary", "backup":"new-secondary", "endpoint":"https://new.example.test", "optional":"new public string", "date":"2026-10-23", "scopes":["read","write"], "notes":{"nested":[true,null,123]}, "env_token":"RENAMED_TOKEN", "validity_token":"permanent"}
	}`))
	if err != nil || !changed {
		t.Fatalf("save = %v, %v", changed, err)
	}
	after := loadEditVault(t, path)
	for id, want := range map[string]string{
		"token": `"new-primary"`, "backup": `"new-secondary"`, "endpoint": `"https://new.example.test"`,
		"optional": `"new public string"`, "date": `"2026-10-23"`, "scopes": `["read","write"]`,
		"notes": `{"nested":[true,null,123]}`, "env_token": `"RENAMED_TOKEN"`, "validity_token": `"permanent"`,
	} {
		assertEditField(t, after.Items[0], id, want)
	}
	// Remove only the deliberately changed values and organization, then compare
	// the entire remaining vault (timestamps, schema, profiles and other items).
	gotItem := after.Items[0]
	if gotItem.Title != "Renamed service" || gotItem.Description != "New organization" || !reflect.DeepEqual(gotItem.Labels, []string{"work", "edited"}) {
		t.Fatal("item organization was not normalized and saved")
	}
	assertEditField(t, gotItem, "env_backup", `"SERVICE_BACKUP"`)
	assertEditField(t, gotItem, "env_endpoint", `"SERVICE_URL"`)
	gotItem.Title, gotItem.Description, gotItem.Labels, gotItem.Fields = before.Items[0].Title, before.Items[0].Description, before.Items[0].Labels, before.Items[0].Fields
	after.Items[0] = gotItem
	if !reflect.DeepEqual(after, before) {
		t.Fatal("whole-item edit changed readonly metadata or unrelated vault data")
	}
}

func TestSecretEditSemanticNoOpDoesNotRewrite(t *testing.T) {
	path, config, vault := writeEditScenario(t)
	if err := setFieldValueIfPresent(&vault.Items[0], "notes", json.RawMessage(`{"integer":9007199254740993,"object":{"a":1,"b":[true,null]}}`)); err != nil {
		t.Fatal(err)
	}
	if err := vaultjson.NewRepository(path).Save(vault); err != nil {
		t.Fatal(err)
	}
	session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
	if err != nil {
		t.Fatal(err)
	}
	for _, document := range [][]byte{
		session.Document,
		[]byte(`{}`),
		[]byte(`{"fields":{"notes":{"object":{"b":[true,null],"a":1.0},"integer":9007199254740993}},"title":" Service account ","labels":["work","team","work"]}`),
	} {
		before, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		changed, err := testApp().SaveSecretEdit(session, document)
		if err != nil || changed {
			t.Fatalf("no-op save = %v, %v", changed, err)
		}
		after, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		afterContent, err := os.ReadFile(path)
		if err != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) || !bytes.Equal(content, afterContent) {
			t.Fatal("semantic no-op rewrote storage")
		}
	}
	changed, err := testApp().SaveSecretEdit(session, []byte(`{"fields":{"notes":{"integer":9007199254740992,"object":{"a":1,"b":[true,null]}}}}`))
	if err != nil || !changed {
		t.Fatalf("distinct large integer save = %v, %v", changed, err)
	}
	assertEditField(t, loadEditVault(t, path).Items[0], "notes", `{"integer":9007199254740992,"object":{"a":1,"b":[true,null]}}`)
}

func TestSecretEditExactNumbersWithHugeExponents(t *testing.T) {
	for _, tc := range []struct {
		name, original, equivalent, different string
	}{
		{"positive exponent", "1e1000002", "10e1000001", "2e1000002"},
		{"negative exponent", "1e-1000002", "10e-1000003", "1e-1000003"},
		{"negative number", "-1e1000002", "-10e1000001", "1e1000002"},
		{"fraction and trailing zeros", "1.2300e1000002", "123e1000000", "1.2301e1000002"},
		{"exponent beyond machine integer", "1e100000000000000000000000000002", "10e100000000000000000000000000001", "1e100000000000000000000000000001"},
		{"negative exponent beyond machine integer", "1e-100000000000000000000000000002", "10e-100000000000000000000000000003", "1e-100000000000000000000000000003"},
		{"signed zero", "-0e100000000000000000000000000002", "0.000e-100000000000000000000000000002", "1e-100000000000000000000000000002"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path, config, vault := writeEditScenario(t)
			if err := setFieldValueIfPresent(&vault.Items[0], "notes", json.RawMessage(tc.original)); err != nil {
				t.Fatal(err)
			}
			repo := vaultjson.NewRepository(path)
			if err := repo.Save(vault); err != nil {
				t.Fatal(err)
			}
			session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
			if err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := testApp().SaveSecretEdit(session, []byte(`{"fields":{"notes":`+tc.equivalent+`}}`))
			if changed || err != nil {
				t.Fatalf("equivalent huge-exponent save = %v, %v", changed, err)
			}
			after, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			afterContent, err := os.ReadFile(path)
			if err != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) || !bytes.Equal(content, afterContent) {
				t.Fatal("equivalent huge-exponent number rewrote storage")
			}
			// An equivalent concurrent representation is not a selected-item
			// change and must not cause a false conflict on a secret-only edit.
			if err := setFieldValueIfPresent(&vault.Items[0], "notes", json.RawMessage(tc.equivalent)); err != nil {
				t.Fatal(err)
			}
			if err := repo.Save(vault); err != nil {
				t.Fatal(err)
			}
			changed, err = testApp().SaveSecretEdit(session, []byte(`{"fields":{"token":"updated-secret"}}`))
			if !changed || err != nil {
				t.Fatalf("equivalent concurrent numeric representation = %v, %v", changed, err)
			}
			assertEditField(t, loadEditVault(t, path).Items[0], "token", `"updated-secret"`)
			session, err = testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
			if err != nil {
				t.Fatal(err)
			}
			changed, err = testApp().SaveSecretEdit(session, []byte(`{"fields":{"notes":`+tc.different+`}}`))
			if !changed || err != nil {
				t.Fatalf("different huge-exponent number save = %v, %v", changed, err)
			}
			assertEditField(t, loadEditVault(t, path).Items[0], "notes", tc.different)
		})
	}
}

func TestSecretEditPreservesUnchangedNonNormalizedOrganization(t *testing.T) {
	path, config, original := writeEditScenario(t)
	original.Items[0].Title = " \tLegacy service \n "
	original.Items[0].Description = " \nLegacy description\t "
	original.Items[0].Labels = []string{" work ", "", "team", " work ", "  "}
	if err := vaultjson.NewRepository(path).Save(original); err != nil {
		t.Fatal(err)
	}
	session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(session.Document, &document); err != nil {
		t.Fatal(err)
	}
	compact, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	for _, draft := range [][]byte{session.Document, compact} {
		before, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		changed, err := testApp().SaveSecretEdit(session, draft)
		if changed || err != nil {
			t.Fatalf("unchanged legacy organization save = %v, %v", changed, err)
		}
		after, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		afterContent, err := os.ReadFile(path)
		if err != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) || !bytes.Equal(content, afterContent) {
			t.Fatal("unchanged or format-only legacy document rewrote storage")
		}
	}
	// Null is not a label string even if decoding it as a string would happen
	// to match a pre-existing empty label.
	if changed, err := testApp().SaveSecretEdit(session, []byte(`{"labels":[" work ",null,"team"," work ","  "]}`)); changed || !errors.Is(err, ErrSecretEditInvalidDocument) {
		t.Fatalf("null legacy label = %v, %v", changed, err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(document["fields"], &fields); err != nil {
		t.Fatal(err)
	}
	fields["token"] = rawString("changed-secret-only")
	document["fields"], err = json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	secretOnly, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := testApp().SaveSecretEdit(session, secretOnly)
	if !changed || err != nil {
		t.Fatalf("legacy organization secret-only save = %v, %v", changed, err)
	}
	if err := setFieldValueIfPresent(&original.Items[0], "token", rawString("changed-secret-only")); err != nil {
		t.Fatal(err)
	}
	if got := loadEditVault(t, path); !reflect.DeepEqual(got, original) {
		t.Fatal("secret-only edit altered original whitespace, empty or duplicate labels, or other metadata")
	}
}

func TestSecretEditOmissionClearingAndLiteralNull(t *testing.T) {
	path, config, before := writeEditScenario(t)
	session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := testApp().SaveSecretEdit(session, []byte(`{"description":"", "labels":[], "fields":{"notes":null}, "clear_fields":["endpoint","optional"]}`))
	if err != nil || !changed {
		t.Fatalf("clear save = %v, %v", changed, err)
	}
	after := loadEditVault(t, path)
	item := after.Items[0]
	if item.Title != before.Items[0].Title || item.Description != "" || len(item.Labels) != 0 {
		t.Fatal("metadata omission or explicit clearing semantics failed")
	}
	assertEditField(t, item, "notes", `null`)
	assertEditField(t, item, "endpoint", "")
	assertEditField(t, item, "token", `"old-primary"`)
	assertEditField(t, item, "backup", `"old-secondary"`)
	assertEditField(t, item, "env_endpoint", `"SERVICE_URL"`)
	// The optional field did not exist before and clearing it must not create it.
	for _, field := range item.Fields {
		if field.ID == "optional" {
			t.Fatal("clearing an absent optional field added a storage slot")
		}
	}
	session, err = testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
	if err != nil {
		t.Fatal(err)
	}
	var draft struct {
		Fields map[string]json.RawMessage `json:"fields"`
	}
	if err := json.Unmarshal(session.Document, &draft); err != nil {
		t.Fatal(err)
	}
	if string(draft.Fields["notes"]) != "null" {
		t.Fatal("stored JSON null was omitted from the next draft")
	}
	if _, exists := draft.Fields["endpoint"]; exists {
		t.Fatal("cleared field was included as a null placeholder")
	}
}

func TestSecretEditInvalidDocumentsAreConfidentialAndRetryable(t *testing.T) {
	const sensitive = "private-edit-canary"
	cases := map[string]string{
		"malformed":               `{"fields":{"token":"private-edit-canary"`,
		"unknown top level":       `{"private-edit-canary":true}`,
		"unknown field":           `{"fields":{"private-edit-canary":"value"}}`,
		"duplicate top level":     `{"title":"private-edit-canary","title":"other"}`,
		"duplicate escaped key":   `{"title":"private-edit-canary","\u0074itle":"other"}`,
		"duplicate field":         `{"fields":{"token":"private-edit-canary","token":"other"}}`,
		"duplicate nested object": `{"fields":{"notes":{"key":"private-edit-canary","key":1}}}`,
		"trailing document":       `{} {"title":"private-edit-canary"}`,
		"trailing garbage":        `{} private-edit-canary`,
		"readonly identity":       `{"id":"private-edit-canary"}`,
		"readonly template":       `{"template":"private-edit-canary"}`,
		"readonly timestamp":      `{"created_at":"private-edit-canary"}`,
		"field schema injection":  `{"fields":{"token":{"type":"private-edit-canary","value":"secret"}}}`,
		"null title":              `{"title":null}`,
		"null description":        `{"description":null}`,
		"null labels":             `{"labels":null}`,
		"null label element":      `{"labels":[null]}`,
		"wrong title type":        `{"title":["private-edit-canary"]}`,
		"empty title":             `{"title":"  "}`,
		"empty label":             `{"labels":["  "]}`,
		"wrong description type":  `{"description":42}`,
		"wrong labels type":       `{"labels":"private-edit-canary"}`,
		"null fields":             `{"fields":null}`,
		"array fields":            `{"fields":[]}`,
		"null clear list":         `{"clear_fields":null}`,
		"wrong clear list":        `{"clear_fields":"private-edit-canary"}`,
		"unknown clear field":     `{"clear_fields":["private-edit-canary"]}`,
		"duplicate clear field":   `{"clear_fields":["endpoint","endpoint"]}`,
		"secret clearing":         `{"clear_fields":["token"]}`,
		"set and clear":           `{"fields":{"endpoint":"https://private-edit-canary.test"},"clear_fields":["endpoint"]}`,
		"invalid url":             `{"fields":{"endpoint":"private-edit-canary"}}`,
		"invalid date":            `{"fields":{"date":"private-edit-canary"}}`,
		"invalid list":            `{"fields":{"scopes":["private-edit-canary",1]}}`,
		"invalid mapping":         `{"fields":{"env_token":"private-edit-canary"}}`,
		"invalid validity":        `{"fields":{"validity_token":"private-edit-canary"}}`,
		"clear required validity": `{"clear_fields":["validity_token"]}`,
		"array document":          `[]`,
		"null document":           `null`,
		"empty document":          ``,
	}
	for name, document := range cases {
		t.Run(name, func(t *testing.T) {
			path, config, _ := writeEditScenario(t)
			session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
			if err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := testApp().SaveSecretEdit(session, []byte(document))
			if changed || !errors.Is(err, ErrSecretEditInvalidDocument) || strings.Contains(err.Error(), sensitive) || strings.Contains(err.Error(), "old-primary") {
				t.Fatalf("invalid document must return a value-free sentinel: %v, %v", changed, err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("invalid document changed storage")
			}
			// Mutating the caller-owned initial draft cannot mutate the baseline.
			for i := range session.Document {
				session.Document[i] = 'x'
			}
			changed, err = testApp().SaveSecretEdit(session, []byte(`{"fields":{"token":"corrected-secret"}}`))
			if err != nil || !changed {
				t.Fatalf("corrected document retry = %v, %v", changed, err)
			}
			assertEditField(t, loadEditVault(t, path).Items[0], "backup", `"old-secondary"`)
		})
	}
}

func TestSecretEditConflictsDoNotOverwrite(t *testing.T) {
	for name, change := range map[string]func(*vaultmodel.Vault){
		"item removal":      func(v *vaultmodel.Vault) { v.Items = nil },
		"item organization": func(v *vaultmodel.Vault) { v.Items[0].Description = "concurrent description" },
		"item timestamps":   func(v *vaultmodel.Vault) { v.Items[0].UpdatedAt = "2026-10-23" },
		"item field":        func(v *vaultmodel.Vault) { v.Items[0].Fields[0].Value = rawString("concurrent-secret") },
		"schema label":      func(v *vaultmodel.Vault) { v.Templates[0].Fields[0].Label = "Concurrent label" },
		"schema sensitivity": func(v *vaultmodel.Vault) {
			v.Templates[0].Fields[4].Sensitivity = vaultmodel.SensitivitySecret
			delete(v.Profiles[0].Defaults, "endpoint")
		},
		"schema optional field": func(v *vaultmodel.Vault) {
			v.Templates[0].Fields = append(v.Templates[0].Fields, vaultmodel.FieldTemplate{ID: "new_field", Label: "New", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic})
		},
	} {
		t.Run(name, func(t *testing.T) {
			path, config, concurrent := writeEditScenario(t)
			session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
			if err != nil {
				t.Fatal(err)
			}
			change(&concurrent)
			if err := vaultjson.NewRepository(path).Save(concurrent); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := testApp().SaveSecretEdit(session, []byte(`{"title":"Edited item", "fields":{"token":"private-edit-canary"}}`))
			if changed || !errors.Is(err, ErrSecretEditConflict) || strings.Contains(err.Error(), "private-edit-canary") || strings.Contains(err.Error(), "concurrent-secret") {
				t.Fatalf("conflict result = %v, %v", changed, err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("conflicting edit overwrote concurrent data")
			}
		})
	}
}

func TestSecretEditPreservesUnrelatedConcurrentChanges(t *testing.T) {
	path, config, concurrent := writeEditScenario(t)
	session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
	if err != nil {
		t.Fatal(err)
	}
	concurrent.Items = append(concurrent.Items, vaultmodel.Item{ID: "other", Title: "Other", Template: vaultmodel.TemplateGenericSecret, Description: "Concurrent item", Fields: []vaultmodel.Field{{ID: "value", Value: rawString("unrelated-secret")}}})
	concurrent.Templates = append(concurrent.Templates, vaultmodel.Template{ID: "unrelated", Name: "Unrelated", PrimarySecret: "secret", Fields: []vaultmodel.FieldTemplate{{ID: "secret", Label: "Secret", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret}}})
	concurrent.Profiles[0].Name = "Concurrent profile"
	if err := vaultjson.NewRepository(path).Save(concurrent); err != nil {
		t.Fatal(err)
	}
	changed, err := testApp().SaveSecretEdit(session, []byte(`{"title":"Edited item"}`))
	if err != nil || !changed {
		t.Fatalf("concurrent unrelated save = %v, %v", changed, err)
	}
	concurrent.Items[0].Title = "Edited item"
	if got := loadEditVault(t, path); !reflect.DeepEqual(got, concurrent) {
		t.Fatal("edit lost unrelated concurrent items, templates, profiles or values")
	}
}

func TestSecretEditSaveFailureCanRetry(t *testing.T) {
	path, config, _ := writeEditScenario(t)
	session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Block the repository backup with a directory, independent of root privileges.
	if err := os.Remove(path + ".bak"); err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if err := os.Mkdir(path+".bak", 0o700); err != nil {
		t.Fatal(err)
	}
	document := []byte(`{"title":"private-edit-canary", "fields":{"token":"new-secret"}}`)
	changed, err := testApp().SaveSecretEdit(session, document)
	if changed || err == nil || errors.Is(err, ErrSecretEditInvalidDocument) || errors.Is(err, ErrSecretEditConflict) || strings.Contains(err.Error(), "private-edit-canary") || strings.Contains(err.Error(), "new-secret") {
		t.Fatalf("storage failure = %v, %v", changed, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed save changed storage")
	}
	if err := os.Remove(path + ".bak"); err != nil {
		t.Fatal(err)
	}
	changed, err = testApp().SaveSecretEdit(session, document)
	if err != nil || !changed {
		t.Fatalf("save retry = %v, %v", changed, err)
	}
	assertEditField(t, loadEditVault(t, path).Items[0], "token", `"new-secret"`)
	assertEditField(t, loadEditVault(t, path).Items[0], "backup", `"old-secondary"`)
}

func TestSecretEditPinsVaultDespiteConfigAndWorkingDirectory(t *testing.T) {
	path, config, _ := writeEditScenario(t)
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Dir(path))
	session, err := testApp().BeginSecretEdit(config, "vault.json", SecretEditRequest{Selector: "stable_service"})
	if err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(t.TempDir(), "other.json")
	writeEmptyVault(t, other)
	if err := os.WriteFile(config, []byte("vault_path: "+other+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(other)
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(originalDir)
	changed, err := testApp().SaveSecretEdit(session, []byte(`{"title":"Pinned item"}`))
	if err != nil || !changed {
		t.Fatalf("pinned save = %v, %v", changed, err)
	}
	if got := loadEditVault(t, path).Items[0].Title; got != "Pinned item" {
		t.Fatal("save did not use the original resolved vault")
	}
	after, err := os.ReadFile(other)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("save was redirected by config or working directory")
	}
}

func TestSecretEditTargetResolution(t *testing.T) {
	path, config, vault := writeEditScenario(t)
	for _, req := range []SecretEditRequest{
		{Selector: "stable_service"}, {Selector: "Service account"}, {Env: "SERVICE_TOKEN"},
		{Env: "SERVICE_BACKUP", Target: "stable_service.backup"}, {Env: "SERVICE_URL", Target: "stable_service.endpoint"},
	} {
		session, err := testApp().BeginSecretEdit(config, path, req)
		if err != nil || session == nil || !bytes.Contains(session.Document, []byte(`"id": "stable_service"`)) {
			t.Fatalf("existing selector/target resolution failed: %v", err)
		}
	}
	for _, req := range []SecretEditRequest{
		{}, {Selector: "missing"}, {Selector: "stable_service", Env: "SERVICE_TOKEN"},
		{Selector: "stable_service", Target: "stable_service.token"}, {Env: "MISSING"}, {Env: "bad-name"},
		{Env: "SERVICE_TOKEN", Target: "Service account.token"}, {Env: "SERVICE_TOKEN", Target: "stable_service.backup"},
		{Env: "SERVICE_TOKEN", Target: "stable_service.notes"}, {Env: "SERVICE_TOKEN", Target: "stable_service.env_token"},
	} {
		if session, err := testApp().BeginSecretEdit(config, path, req); err == nil || session != nil {
			t.Fatal("invalid, missing, or unmapped target was accepted")
		}
	}
	if err := setFieldValueIfPresent(&vault.Items[0], "env_backup", rawString("SERVICE_TOKEN")); err != nil {
		t.Fatal(err)
	}
	if err := vaultjson.NewRepository(path).Save(vault); err != nil {
		t.Fatal(err)
	}
	if _, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Env: "SERVICE_TOKEN"}); err == nil {
		t.Fatal("ambiguous mapping was accepted without a concrete target")
	}
	for _, target := range []string{"stable_service.token", "stable_service.backup"} {
		if _, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Env: "SERVICE_TOKEN", Target: target}); err != nil {
			t.Fatalf("explicit same-item ambiguous target failed: %v", err)
		}
	}
	// A title matching another item's ID must be ambiguous in selector mode,
	// but concrete env targets must continue to identify the exact owning ID.
	vault.Items = append(vault.Items, vaultmodel.Item{ID: "other", Title: "stable_service", Template: vaultmodel.TemplateGenericSecret, Fields: []vaultmodel.Field{{ID: "value", Value: rawString("other-secret")}}})
	if err := vaultjson.NewRepository(path).Save(vault); err != nil {
		t.Fatal(err)
	}
	if _, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"}); err == nil {
		t.Fatal("ambiguous ID/title selector was accepted")
	}
	if _, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Env: "SERVICE_TOKEN", Target: "stable_service.token"}); err != nil {
		t.Fatalf("stable ID target failed amid title collision: %v", err)
	}
}

func TestSecretEditMissingVaultAndTitleCollision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	if session, err := testApp().BeginSecretEdit(writeConfig(t, path), path, SecretEditRequest{Selector: "missing"}); err == nil || session != nil {
		t.Fatal("missing vault produced a session")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("begin initialized a missing vault")
	}
	path, config, vault := writeEditScenario(t)
	session, err := testApp().BeginSecretEdit(config, path, SecretEditRequest{Selector: "stable_service"})
	if err != nil {
		t.Fatal(err)
	}
	vault.Items = append(vault.Items, vaultmodel.Item{ID: "private-edit-canary", Title: "Other", Template: vaultmodel.TemplateGenericSecret, Fields: []vaultmodel.Field{{ID: "value"}}})
	if err := vaultjson.NewRepository(path).Save(vault); err != nil {
		t.Fatal(err)
	}
	changed, err := testApp().SaveSecretEdit(session, []byte(`{"title":"private-edit-canary"}`))
	if changed || !errors.Is(err, ErrSecretEditInvalidDocument) || strings.Contains(err.Error(), "private-edit-canary") {
		t.Fatalf("title cross-collision = %v, %v", changed, err)
	}
	if got := loadEditVault(t, path); !reflect.DeepEqual(got, vault) {
		t.Fatal("title collision mutated storage")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if changed, err := testApp().SaveSecretEdit(session, []byte(`{"title":"Edited"}`)); changed || !errors.Is(err, ErrSecretEditConflict) {
		t.Fatalf("removed vault result = %v, %v", changed, err)
	}
	if changed, err := testApp().SaveSecretEdit(nil, []byte(`{}`)); changed || !errors.Is(err, ErrSecretEditInvalidDocument) {
		t.Fatalf("missing session result = %v, %v", changed, err)
	}
}

func writeEditScenario(t *testing.T) (string, string, vaultmodel.Vault) {
	t.Helper()
	path, config, vault := writeSetScenario(t)
	vault.Templates[0].Fields = append(vault.Templates[0].Fields,
		vaultmodel.FieldTemplate{ID: "optional", Label: "Optional", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
		vaultmodel.FieldTemplate{ID: "date", Label: "Date", Type: vaultmodel.FieldTypeDate, Sensitivity: vaultmodel.SensitivityPublic},
		vaultmodel.FieldTemplate{ID: "scopes", Label: "Scopes", Type: vaultmodel.FieldTypeStringList, Sensitivity: vaultmodel.SensitivityPublic},
		vaultmodel.FieldTemplate{ID: "validity_token", Label: "Validity", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Required: true},
	)
	vault.Items[0].Fields = append(vault.Items[0].Fields,
		vaultmodel.Field{ID: "date", Value: rawString("2025-01-01")},
		vaultmodel.Field{ID: "scopes", Value: json.RawMessage(`["old"]`)},
		vaultmodel.Field{ID: "validity_token", Value: rawString("unknown")},
	)
	if err := vaultjson.NewRepository(path).Save(vault); err != nil {
		t.Fatal(err)
	}
	return path, config, loadEditVault(t, path)
}

func loadEditVault(t *testing.T, path string) vaultmodel.Vault {
	t.Helper()
	vault, err := vaultjson.NewRepository(path).Load()
	if err != nil {
		t.Fatal(err)
	}
	return vault
}

func assertEditField(t *testing.T, item vaultmodel.Item, id, want string) {
	t.Helper()
	for _, field := range item.Fields {
		if field.ID != id {
			continue
		}
		if want == "" {
			if len(field.Value) != 0 {
				t.Fatalf("field %s was not cleared", id)
			}
			return
		}
		var gotValue, wantValue any
		decoder := json.NewDecoder(bytes.NewReader(field.Value))
		decoder.UseNumber()
		if err := decoder.Decode(&gotValue); err != nil {
			t.Fatal(err)
		}
		decoder = json.NewDecoder(strings.NewReader(want))
		decoder.UseNumber()
		if err := decoder.Decode(&wantValue); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotValue, wantValue) {
			t.Fatalf("field %s did not persist its expected JSON value", id)
		}
		return
	}
	t.Fatalf("field %s is missing", id)
}

func TestSecretEditBuiltInSchemaAndReloadFailureRetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	config := writeConfig(t, path)
	app := testApp()
	id, err := app.CreateSecret(config, path, CreateSecretRequest{Title: "Built-in secret", Env: "BUILTIN_TOKEN", Stdin: true, In: strings.NewReader("original-secret")})
	if err != nil {
		t.Fatal(err)
	}
	session, err := app.BeginSecretEdit(config, path, SecretEditRequest{Env: "BUILTIN_TOKEN"})
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"private-edit-canary":`), 0o600); err != nil {
		t.Fatal(err)
	}
	document := []byte(`{"fields":{"value":"replacement-secret"}}`)
	changed, err := app.SaveSecretEdit(session, document)
	if changed || err == nil || strings.Contains(err.Error(), "private-edit-canary") || strings.Contains(err.Error(), "replacement-secret") {
		t.Fatalf("reload failure must be confidential: %v, %v", changed, err)
	}
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err = app.SaveSecretEdit(session, document)
	if err != nil || !changed {
		t.Fatalf("reload failure retry = %v, %v", changed, err)
	}
	vault := loadEditVault(t, path)
	if len(vault.Items) != 1 || vault.Items[0].ID != id || vault.Items[0].Template != vaultmodel.TemplateGenericSecret {
		t.Fatal("built-in edit changed identity or template")
	}
	assertEditField(t, vault.Items[0], "value", `"replacement-secret"`)
}
