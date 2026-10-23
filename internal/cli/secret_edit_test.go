package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/zhongyangchuwu/shelf-go/internal/vaultjson"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
	"golang.org/x/term"
)

type secretEditorRecord struct {
	Path          string
	Args          []string
	DirectoryMode os.FileMode
	FileMode      os.FileMode
	PID           int
}

// The test binary is a real external executable, configured through its
// environment rather than placing edited plaintext on the command line.
func TestSecretEditEditorProcess(t *testing.T) {
	mode := os.Getenv("SHELF_TEST_EDITOR_MODE")
	if mode == "" {
		return
	}
	var interrupts chan os.Signal
	if mode == "interrupt" {
		interrupts = make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
	}
	if mode == "stubborn" || mode == "tty-stubborn" {
		signal.Ignore(syscall.SIGTERM, syscall.SIGHUP)
	}
	if strings.HasPrefix(mode, "tty") {
		if _, err := term.MakeRaw(int(os.Stdin.Fd())); err != nil {
			os.Exit(82)
		}
	}
	path := os.Args[len(os.Args)-1]
	info, err := os.Stat(path)
	if err != nil {
		os.Exit(71)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		os.Exit(72)
	}
	record := secretEditorRecord{Path: path, Args: os.Args[1:], DirectoryMode: dirInfo.Mode().Perm(), FileMode: info.Mode().Perm(), PID: os.Getpid()}
	recordFile, err := os.OpenFile(os.Getenv("SHELF_TEST_EDITOR_RECORD"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		os.Exit(73)
	}
	if json.NewEncoder(recordFile).Encode(record) != nil || recordFile.Close() != nil {
		os.Exit(74)
	}
	if strings.HasPrefix(mode, "tty") {
		var one [1]byte
		if _, err := io.ReadFull(os.Stdin, one[:]); err != nil || one[0] != 'x' {
			os.Exit(83)
		}
		if os.WriteFile(os.Getenv("SHELF_TEST_EDITOR_READ"), []byte("read"), 0o600) != nil {
			os.Exit(84)
		}
		if mode == "tty-control-c" {
			if _, err := io.ReadFull(os.Stdin, one[:]); err != nil || one[0] != 3 {
				os.Exit(89)
			}
		}
	}
	if mode == "stubborn" || mode == "tty-stubborn" {
		for {
			if _, err := os.Stat(path); err != nil {
				_ = os.WriteFile(os.Getenv("SHELF_TEST_EDITOR_EARLY_CLEANUP"), []byte("draft removed while editor active"), 0o600)
				os.Exit(85)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	if mode == "interrupt" {
		<-interrupts
		if _, err := os.Stat(path); err != nil {
			os.Exit(81)
		}
	}
	if mode == "fail" {
		os.Exit(19)
	}
	document, err := os.ReadFile(path)
	if err != nil {
		os.Exit(75)
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(document, &object) != nil {
		// A re-edit may deliberately receive an invalid document.
		object = make(map[string]json.RawMessage)
	}
	if mode == "conflict" || mode == "unrelated" {
		repo := vaultjson.NewRepository(os.Getenv("SHELF_TEST_EDITOR_VAULT"))
		vault, err := repo.Load()
		if err != nil {
			os.Exit(76)
		}
		for index := range vault.Items {
			if (mode == "conflict" && vault.Items[index].ID == "item_alpha") || (mode == "unrelated" && vault.Items[index].ID == "item_beta") {
				vault.Items[index].Description = "concurrent change"
			}
		}
		if repo.Save(vault) != nil {
			os.Exit(77)
		}
	}
	if mode == "invalid" || mode == "reedit" {
		count, _ := os.ReadFile(os.Getenv("SHELF_TEST_EDITOR_RECORD"))
		if mode == "invalid" || bytes.Count(count, []byte{'\n'}) == 1 {
			document = []byte(`{"title":"edited-private-value","fields":{"unknown":"edited-private-value"}}`)
		} else {
			document = []byte(os.Getenv("SHELF_TEST_EDITOR_DOCUMENT"))
		}
	} else if mode != "noop" {
		document = []byte(os.Getenv("SHELF_TEST_EDITOR_DOCUMENT"))
	} else {
		document, err = json.Marshal(object)
		if err != nil {
			os.Exit(78)
		}
	}
	if os.WriteFile(path, document, 0o600) != nil {
		os.Exit(79)
	}
	if mode == "savefail" {
		count, _ := os.ReadFile(os.Getenv("SHELF_TEST_EDITOR_RECORD"))
		if bytes.Count(count, []byte{'\n'}) == 1 {
			if os.WriteFile(os.Getenv("SHELF_TEST_EDITOR_VAULT"), []byte("{"), 0o600) != nil {
				os.Exit(80)
			}
		}
	}
	os.Exit(0)
}

func TestSecretEditWrapperProcess(t *testing.T) {
	if os.Getenv("SHELF_TEST_EDITOR_MODE") == "" {
		return
	}
	child := exec.Command(os.Args[0], "-test.run=^TestSecretEditEditorProcess$", "--", os.Args[len(os.Args)-1])
	child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := child.Run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func configureSecretEditor(t *testing.T, mode string) string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(t.TempDir(), "editor-record.jsonl")
	t.Setenv("VISUAL", fmt.Sprintf("%q -test.run=^TestSecretEditEditorProcess$ --", executable))
	t.Setenv("EDITOR", "must-not-be-used")
	t.Setenv("SHELF_TEST_EDITOR_MODE", mode)
	t.Setenv("SHELF_TEST_EDITOR_RECORD", record)
	t.Setenv("SHELF_TEST_EDITOR_DOCUMENT", `{"title":"Edited title","description":"Edited description","labels":["edited"],"fields":{"value":"edited-private-value"}}`)
	return record
}

func secretEditorRecords(t *testing.T, path string) []secretEditorRecord {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var records []secretEditorRecord
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte{'\n'}) {
		var record secretEditorRecord
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	return records
}

func assertSecretEditorCleaned(t *testing.T, recordPath string, count int) []secretEditorRecord {
	t.Helper()
	records := secretEditorRecords(t, recordPath)
	if len(records) != count {
		t.Fatalf("editor invocations = %d, want %d", len(records), count)
	}
	for _, record := range records {
		if filepath.Base(record.Path) != "item.json" || record.DirectoryMode != 0o700 || record.FileMode != 0o600 {
			t.Fatalf("draft privacy = %#v", record)
		}
		if _, err := os.Stat(filepath.Dir(record.Path)); !os.IsNotExist(err) {
			t.Fatalf("draft directory was not removed: %v", err)
		}
		for _, arg := range record.Args {
			if strings.Contains(arg, "private-value") {
				t.Fatal("editor argv leaked plaintext")
			}
		}
	}
	return records
}

func assertSecretEditValueFree(t *testing.T, out string, err error) {
	t.Helper()
	if err != nil {
		out += err.Error()
	}
	for _, value := range []string{"alpha-private-value", "beta-private-value", "edited-private-value"} {
		if strings.Contains(out, value) {
			t.Fatal("Shelf output or diagnostic leaked a secret value")
		}
	}
}

func TestSecretEditWholeItemAndNoop(t *testing.T) {
	for _, mode := range []string{"edit", "noop"} {
		t.Run(mode, func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {env: "ALPHA_TOKEN", value: "alpha-private-value"}})
			record := configureSecretEditor(t, mode)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			beforeInfo, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "alpha")
			assertSecretEditValueFree(t, out, err)
			if err != nil {
				t.Fatal(err)
			}
			assertSecretEditorCleaned(t, record, 1)
			vault, err := vaultjson.NewRepository(path).Load()
			if err != nil {
				t.Fatal(err)
			}
			if mode == "edit" {
				item := vault.Items[0]
				if item.Title != "Edited title" || item.Description != "Edited description" || !reflect.DeepEqual(item.Labels, []string{"edited"}) {
					t.Fatalf("whole-item metadata was not saved: %#v", item.ID)
				}
				got, err := runShelf(t, "--vault", path, "secret", "get", "--env", "ALPHA_TOKEN")
				if err != nil || got != "edited-private-value\n" {
					t.Fatal("edited value or omitted environment mapping was not retained")
				}
				if !strings.Contains(out, "Updated secret item.") {
					t.Fatalf("save result = %q", out)
				}
			} else {
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("formatting-only edit changed storage")
				}
				afterInfo, err := os.Stat(path)
				if err != nil || !os.SameFile(beforeInfo, afterInfo) || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
					t.Fatal("no-op edit rewrote storage")
				}
				if !strings.Contains(out, "No changes") {
					t.Fatalf("no-op result = %q", out)
				}
			}
		})
	}
}

func TestSecretEditQuotedCommandAndVisualPrecedence(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
	record := configureSecretEditor(t, "noop")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	withSpaces := filepath.Join(t.TempDir(), "editor with spaces")
	if err := os.Symlink(executable, withSpaces); err != nil {
		t.Skipf("executable symlink unavailable: %v", err)
	}
	marker := filepath.Join(t.TempDir(), "must-not-exist")
	t.Setenv("VISUAL", fmt.Sprintf("'%s' -test.run=^TestSecretEditEditorProcess$ -- 'two words' \"double words\" escaped\\ space '' '$HOME' '$(touch %s)' ';' '>'", withSpaces, marker))
	out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "alpha")
	if err != nil {
		t.Fatal(err)
	}
	assertSecretEditValueFree(t, out, err)
	records := assertSecretEditorCleaned(t, record, 1)
	args := records[0].Args
	want := []string{"-test.run=^TestSecretEditEditorProcess$", "--", "two words", "double words", "escaped space", "", "$HOME", "$(touch " + marker + ")", ";", ">"}
	if !reflect.DeepEqual(args[:len(args)-1], want) {
		t.Fatalf("quoted editor argv = %q", args)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("editor command was evaluated by a shell")
	}
}

func TestSecretEditEditorFallbacks(t *testing.T) {
	for _, fallback := range []string{"EDITOR", "vi"} {
		t.Run(fallback, func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
			record := configureSecretEditor(t, "noop")
			visual := os.Getenv("VISUAL")
			t.Setenv("VISUAL", " \t ")
			if fallback == "EDITOR" {
				t.Setenv("EDITOR", visual)
			} else {
				// A tiny executable wrapper gives vi its usual basename while
				// forwarding to the controlled editor process.
				executable, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				dir := t.TempDir()
				wrapper := "#!/bin/sh\nexec '" + strings.ReplaceAll(executable, "'", "'\\''") + "' -test.run=^TestSecretEditEditorProcess$ -- \"$@\"\n"
				if err := os.WriteFile(filepath.Join(dir, "vi"), []byte(wrapper), 0o700); err != nil {
					t.Fatal(err)
				}
				t.Setenv("PATH", dir)
				t.Setenv("EDITOR", "")
			}
			out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "alpha")
			if err != nil {
				t.Fatal(err)
			}
			assertSecretEditValueFree(t, out, err)
			assertSecretEditorCleaned(t, record, 1)
		})
	}
}

func TestSecretEditFailuresDoNotSaveOrLeak(t *testing.T) {
	for _, failure := range []string{"invalid", "fail", "launch", "quote"} {
		t.Run(failure, func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
			record := configureSecretEditor(t, failure)
			temp := t.TempDir()
			t.Setenv("TMPDIR", temp)
			if failure == "launch" {
				t.Setenv("VISUAL", filepath.Join(t.TempDir(), "missing-editor"))
			}
			if failure == "quote" {
				t.Setenv("VISUAL", "editor 'edited-private-value")
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "alpha")
			if err == nil {
				t.Fatal("editor failure succeeded")
			}
			assertSecretEditValueFree(t, out, err)
			if strings.Contains(out, "[e]dit") {
				t.Fatal("noninteractive failure prompted")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("failed edit changed storage")
			}
			entries, err := os.ReadDir(temp)
			if err != nil || len(entries) != 0 {
				t.Fatal("failed edit left a draft directory")
			}
			if failure == "invalid" || failure == "fail" {
				assertSecretEditorCleaned(t, record, 1)
			}
		})
	}
}

func TestSecretEditConflictAndUnrelatedChanges(t *testing.T) {
	for _, mode := range []string{"conflict", "unrelated"} {
		t.Run(mode, func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}, "beta": {value: "beta-private-value"}})
			record := configureSecretEditor(t, mode)
			t.Setenv("SHELF_TEST_EDITOR_VAULT", path)
			out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "alpha")
			assertSecretEditValueFree(t, out, err)
			if mode == "conflict" && (err == nil || !strings.Contains(err.Error(), "conflict")) {
				t.Fatalf("conflict result = %v", err)
			}
			if mode == "unrelated" && err != nil {
				t.Fatal(err)
			}
			assertSecretEditorCleaned(t, record, 1)
			vault, err := vaultjson.NewRepository(path).Load()
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range vault.Items {
				if mode == "conflict" && item.ID == "item_alpha" && (item.Title != "alpha" || item.Description != "concurrent change") {
					t.Fatal("conflict overwrote the concurrent item")
				}
				if mode == "unrelated" && item.ID == "item_beta" && item.Description != "concurrent change" {
					t.Fatal("save lost an unrelated concurrent change")
				}
			}
		})
	}
}

func TestSecretEditMappingResolutionAndCompletions(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{"alpha": {env: "SHARED_TOKEN", value: "alpha-private-value"}, "beta": {env: "SHARED_TOKEN", value: "beta-private-value"}})
	record := configureSecretEditor(t, "edit")
	for _, args := range [][]string{
		{"--env", "SHARED_TOKEN"},
		{"--env", "MISSING_TOKEN"},
		{"--env", "SHARED_TOKEN", "--target", "alpha.value"},
		{"--env", "SHARED_TOKEN", "--target", "item_alpha.env_value"},
		{"alpha", "--env", "SHARED_TOKEN"},
		{"alpha", "--target", "item_alpha.value"},
		{"--env", ""},
		{"missing"},
	} {
		out, err := runShelfWithInput(t, "", append([]string{"--vault", path, "secret", "edit"}, args...)...)
		if err == nil {
			t.Fatalf("invalid selection accepted: %q", args)
		}
		assertSecretEditValueFree(t, out, err)
	}
	if _, err := os.Stat(record); !os.IsNotExist(err) {
		t.Fatal("invalid or ambiguous selection opened an editor")
	}
	flags := []string{"--vault", path}
	if got := dynamicCompletions(t, flags, "secret", "edit", "--env", "SHARED"); !reflect.DeepEqual(got, []string{"SHARED_TOKEN"}) {
		t.Fatalf("edit env completions = %v", got)
	}
	targets := dynamicCompletions(t, flags, "secret", "edit", "--env", "SHARED_TOKEN", "--target", "")
	if len(targets) != 2 || !hasCompletion(targets, "item_alpha.value") || !hasCompletion(targets, "item_beta.value") {
		t.Fatalf("edit target completions = %v", targets)
	}
	if got := dynamicCompletions(t, flags, "secret", "edit", "alpha"); !hasCompletion(got, "alpha") {
		t.Fatalf("edit selector completions = %v", got)
	}
	out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "--env", "SHARED_TOKEN", "--target", "item_beta.value")
	if err != nil {
		t.Fatal(err)
	}
	assertSecretEditValueFree(t, out, err)
	assertSecretEditorCleaned(t, record, 1)
	got, err := runShelf(t, "--vault", path, "secret", "get", "item_beta")
	if err != nil || got != "edited-private-value\n" {
		t.Fatal("explicit environment target edited the wrong item")
	}
}

func TestSecretEditUsesEligibleMappingCandidates(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{"alpha": {env: "VALID_TOKEN", value: "alpha-private-value"}})
	repo := vaultjson.NewRepository(path)
	vault, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	// A stored, projectable field slot can be mapped before it has a value.
	template, ok := vaultmodel.NewDefaultRegistry().Template(vaultmodel.TemplateGenericSecret)
	if !ok {
		t.Fatal("generic template is unavailable")
	}
	template.ID = "editor_unset"
	template.Name = "Editor unset fields"
	template.Aliases = []string{"editor_unset_alias"}
	template.Fields = append(template.Fields,
		vaultmodel.FieldTemplate{ID: "optional", Label: "Optional", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
		vaultmodel.FieldTemplate{ID: "env_optional", Label: "Optional mapping", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic})
	vault.Templates = append(vault.Templates, template)
	vault.Items[0].Template = template.ID
	vault.Items[0].Fields = append(vault.Items[0].Fields, vaultmodel.Field{ID: "optional"}, vaultmodel.Field{ID: "env_optional", Value: rawJSON("UNSET_TOKEN")})
	if err := repo.Save(vault); err != nil {
		t.Fatal(err)
	}
	record := configureSecretEditor(t, "noop")
	flags := []string{"--vault", path}
	if got := dynamicCompletions(t, flags, "secret", "edit", "--env", "UNSET"); !reflect.DeepEqual(got, []string{"UNSET_TOKEN"}) {
		t.Fatalf("unset mapped slot completions = %v", got)
	}
	out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "--env", "UNSET_TOKEN")
	if err != nil {
		t.Fatal(err)
	}
	assertSecretEditValueFree(t, out, err)
	assertSecretEditorCleaned(t, record, 1)
}

func TestSecretEditTypedFieldsAndExplicitClearing(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{"alpha": {env: "ALPHA_TOKEN", value: "alpha-private-value", labels: []string{"old"}}})
	repo := vaultjson.NewRepository(path)
	vault, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	template, ok := vaultmodel.NewDefaultRegistry().Template(vaultmodel.TemplateGenericSecret)
	if !ok {
		t.Fatal("generic template is unavailable")
	}
	template.ID = "editor_typed"
	template.Name = "Editor typed fields"
	template.Aliases = []string{"editor_typed_alias"}
	template.Fields = append(template.Fields,
		vaultmodel.FieldTemplate{ID: "edit_public", Label: "Public", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic},
		vaultmodel.FieldTemplate{ID: "edit_list", Label: "List", Type: vaultmodel.FieldTypeStringList, Sensitivity: vaultmodel.SensitivityPublic},
		vaultmodel.FieldTemplate{ID: "edit_json", Label: "JSON", Type: vaultmodel.FieldTypeJSON, Sensitivity: vaultmodel.SensitivityPublic},
		vaultmodel.FieldTemplate{ID: "edit_clear", Label: "Clear", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic})
	vault.Templates = append(vault.Templates, template)
	vault.Items[0].Template = template.ID
	vault.Items[0].Description = "old description"
	vault.Items[0].Fields = append(vault.Items[0].Fields,
		vaultmodel.Field{ID: "edit_public", Value: rawJSON("old public")},
		vaultmodel.Field{ID: "edit_clear", Value: rawJSON("clear me")})
	if err := repo.Save(vault); err != nil {
		t.Fatal(err)
	}
	record := configureSecretEditor(t, "edit")
	t.Setenv("SHELF_TEST_EDITOR_DOCUMENT", `{"description":"","labels":[],"fields":{"value":"edited-private-value","edit_public":"new public","edit_list":["one","two"],"edit_json":{"nested":[1,true,null]}},"clear_fields":["edit_clear"]}`)
	out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "alpha")
	if err != nil {
		t.Fatal(err)
	}
	assertSecretEditValueFree(t, out, err)
	assertSecretEditorCleaned(t, record, 1)
	after, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	item := after.Items[0]
	if item.Title != "alpha" || item.Description != "" || len(item.Labels) != 0 {
		t.Fatal("metadata omission or explicit clearing was not respected")
	}
	want := map[string]string{
		"value":       `"edited-private-value"`,
		"env_value":   `"ALPHA_TOKEN"`,
		"edit_public": `"new public"`,
		"edit_list":   `["one","two"]`,
		"edit_json":   `{"nested":[1,true,null]}`,
	}
	for _, field := range item.Fields {
		if field.ID == "edit_clear" && len(field.Value) != 0 {
			t.Fatal("explicit public-field clearing retained its value")
		}
		if raw, exists := want[field.ID]; exists {
			var actual, expected any
			if json.Unmarshal(field.Value, &actual) != nil || json.Unmarshal([]byte(raw), &expected) != nil || !reflect.DeepEqual(actual, expected) {
				t.Fatalf("typed field %s was not saved", field.ID)
			}
			delete(want, field.ID)
		}
	}
	if len(want) != 0 {
		t.Fatal("edited fields or omitted mapping were lost")
	}
}

func TestSecretEditInvalidDocumentsStayValueFree(t *testing.T) {
	for name, document := range map[string]string{
		"unknown top-level":   `{"edited-private-value":"edited-private-value"}`,
		"unknown field":       `{"fields":{"edited-private-value":"edited-private-value"}}`,
		"duplicate top-level": `{"title":"edited-private-value","title":"second"}`,
		"duplicate field":     `{"fields":{"value":"edited-private-value","value":"second"}}`,
		"trailing JSON":       `{} {"title":"edited-private-value"}`,
		"readonly identity":   `{"id":"edited-private-value"}`,
		"readonly schema":     `{"template":"edited-private-value"}`,
		"null metadata":       `{"description":null}`,
		"wrong field type":    `{"fields":{"value":["edited-private-value"]}}`,
		"secret clearing":     `{"clear_fields":["value"]}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
			record := configureSecretEditor(t, "edit")
			t.Setenv("SHELF_TEST_EDITOR_DOCUMENT", document)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			out, err := runShelfWithInput(t, "", "--vault", path, "secret", "edit", "alpha")
			if err == nil {
				t.Fatal("invalid document was accepted")
			}
			assertSecretEditValueFree(t, out, err)
			assertSecretEditorCleaned(t, record, 1)
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("invalid document changed storage")
			}
		})
	}
}
