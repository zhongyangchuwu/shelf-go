package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestCreateActionBuildsAPIRequestAndDropsSecretDraft(t *testing.T) {
	m, service, _ := readyModel(t)
	editor := newCreateEditor(app.DefaultSecretCreateOptions())
	editor.template = vaultmodel.TemplateAPICredential
	editor.profile = vaultmodel.ProfileOpenAI
	editor.rebuildCreateFields("")
	setEditorValue(t, editor, "title", "New API")
	setEditorValue(t, editor, "description", "Development credential")
	setEditorValue(t, editor, "labels", "api, work")
	setEditorValue(t, editor, "field:credential", fakeSecret)
	setEditorValue(t, editor, "field:env_credential", "NEW_API_KEY")
	setEditorValue(t, editor, "field:base_url", "https://api.example.test")
	m.form = formState{kind: formCreate, editor: editor}

	updated, cmd := m.submitForm()
	m = updated.(Model)
	if editor.value("field:credential") != "" {
		t.Fatal("create form retained the secret draft after submission")
	}
	message := cmd()
	result, ok := message.(actionDoneMsg)
	if !ok || result.err != nil || !result.refresh {
		t.Fatalf("create result = %#v", message)
	}
	if len(service.created) != 1 {
		t.Fatalf("create calls = %d, want 1", len(service.created))
	}
	request := service.created[0]
	if request.Template != vaultmodel.TemplateAPICredential || request.Profile != vaultmodel.ProfileOpenAI || request.Description != "Development credential" || strings.Join(request.Labels, ",") != "api,work" || request.Env != "NEW_API_KEY" || request.BaseURL != "https://api.example.test" {
		t.Fatalf("create request = %#v", request)
	}
	if got := m.render(); strings.Contains(got, fakeSecret) {
		t.Fatalf("create view exposed secret: %q", got)
	}
}

func TestUnifiedEditActionBuildsAtomicFieldUpdates(t *testing.T) {
	m, service, _ := readyModel(t)
	updated, _ := m.startEditForm()
	m = updated.(Model)
	setEditorValue(t, m.form.editor, "title", "OpenAI work")
	setEditorValue(t, m.form.editor, "description", "Work account")
	setEditorValue(t, m.form.editor, "labels", "ai, work")
	setEditorValue(t, m.form.editor, "field:credential", fakeSecret)
	setEditorValue(t, m.form.editor, "field:provider", "OpenAI work")

	updated, cmd := m.submitForm()
	m = updated.(Model)
	if m.form.editor.value("field:credential") != "" {
		t.Fatal("edit form retained the secret draft after submission")
	}
	result := cmd().(actionDoneMsg)
	if result.err != nil || !result.refresh {
		t.Fatalf("update result = %#v", result)
	}
	if len(service.updated) != 1 {
		t.Fatalf("update calls = %#v", service.updated)
	}
	request := service.updated[0]
	if !request.TitleSet || request.Title != "OpenAI work" || !request.DescriptionSet || request.Description != "Work account" || strings.Join(request.AddLabels, ",") != "work" || len(request.FieldUpdates) != 2 {
		t.Fatalf("update request = %#v", request)
	}
	updates := map[string]bool{}
	for _, update := range request.FieldUpdates {
		updates[update.Field] = update.ReadField || update.FieldValueSet
	}
	if !updates["credential"] || !updates["provider"] {
		t.Fatalf("field updates = %#v", request.FieldUpdates)
	}
}

func TestCopyAndErrorViewsNeverRenderSecretValues(t *testing.T) {
	m, service, clipboard := readyModel(t)
	updated, cmd := m.copySelectedField()
	m = updated.(Model)
	result := cmd().(actionDoneMsg)
	if result.err != nil || result.status != "Copied secret to clipboard." || result.refresh {
		t.Fatalf("copy result = %#v", result)
	}
	if clipboard.copied != fakeSecret {
		t.Fatalf("clipboard value = %q, want fake secret", clipboard.copied)
	}
	m = updateModel(t, m, result)
	if got := m.render(); strings.Contains(got, fakeSecret) || !strings.Contains(got, "Copied secret to clipboard.") {
		t.Fatalf("copy view = %q", got)
	}
	m.screen = screenConfirmDelete
	if got := m.render(); strings.Contains(got, fakeSecret) || !strings.Contains(got, "item_openai") || !strings.Contains(got, vaultmodel.TemplateAPICredential) {
		t.Fatalf("delete confirmation view = %q", got)
	}
	m.screen = screenWorkspace

	service.err = fmt.Errorf("operation failed for %s", fakeSecret)
	m = updateModel(t, m, actionDoneMsg{err: service.err})
	if got := m.render(); strings.Contains(got, fakeSecret) || !strings.Contains(got, "No secret value was displayed.") {
		t.Fatalf("error view = %q", got)
	}
}

func TestUnifiedEditFormMasksSecretsAndCancels(t *testing.T) {
	m, _, _ := readyModel(t)
	updated, _ := m.startEditForm()
	m = updated.(Model)
	if m.form.editor.value("field:credential") != "" {
		t.Fatalf("secret editor value = %q, want empty", m.form.editor.value("field:credential"))
	}
	if got := m.renderForm(80); strings.Contains(got, fakeSecret) {
		t.Fatalf("secret editor view exposed secret: %q", got)
	}
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.screen != screenWorkspace || m.form.kind != formNone {
		t.Fatalf("cancelled form state = screen %v form %v", m.screen, m.form.kind)
	}
}

func TestDeleteUsesAppWorkflow(t *testing.T) {
	m, service, _ := readyModel(t)
	m.screen = screenConfirmDelete
	updated, cmd := m.handleDeleteConfirm("enter")
	m = updated.(Model)
	if m.screen != screenLoading {
		t.Fatalf("delete screen = %v, want loading", m.screen)
	}
	if result := cmd().(actionDoneMsg); result.err != nil || !result.refresh {
		t.Fatalf("delete result = %#v", result)
	}
	if got, want := service.removed, []string{"item_openai"}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("remove selectors = %#v, want %#v", got, want)
	}
}

func TestCreateEditorSeparatesKindFromPreset(t *testing.T) {
	editor := newCreateEditor(app.DefaultSecretCreateOptions())
	if editor.template != vaultmodel.TemplateGenericSecret || editor.indexOf("preset") != 0 {
		t.Fatalf("generic editor = template %q fields %#v", editor.template, editor.fields)
	}
	if _, submit := editor.update(tea.KeyPressMsg{Code: tea.KeyRight}); submit {
		t.Fatal("kind selection submitted the editor")
	}
	if editor.template != vaultmodel.TemplateAPICredential || editor.indexOf("preset") == 0 {
		t.Fatalf("API kind did not reveal preset: template %q fields %#v", editor.template, editor.fields)
	}
	editor.focus(editor.indexOf("preset"))
	editor.update(tea.KeyPressMsg{Code: tea.KeyRight})
	if editor.profile != vaultmodel.ProfileOpenAI {
		t.Fatalf("preset = %q, want %q", editor.profile, vaultmodel.ProfileOpenAI)
	}
	if got := editor.value("field:env_credential"); got != "OPENAI_API_KEY" {
		t.Fatalf("OpenAI preset env = %q", got)
	}
	editor.update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if editor.profile != "" {
		t.Fatalf("preset after deselection = %q", editor.profile)
	}
	for _, fieldID := range []string{"env_credential", "provider", "website"} {
		if got := editor.value("field:" + fieldID); got != "" {
			t.Fatalf("%s retained an OpenAI default after preset deselection: %q", fieldID, got)
		}
	}
}

func TestCreateEditorValidityUsesChoiceAndDatePrompt(t *testing.T) {
	editor := newCreateEditor(app.DefaultSecretCreateOptions())
	editor.update(tea.KeyPressMsg{Code: tea.KeyRight})
	validityID := "validity:validity_credential"
	if editor.indexOf(validityID) == 0 {
		t.Fatalf("validity choice %q is missing: %#v", validityID, editor.fields)
	}
	editor.focus(editor.indexOf(validityID))
	editor.update(tea.KeyPressMsg{Code: tea.KeyRight})
	if choice := editor.fields[editor.selected].optionValues[editor.fields[editor.selected].option]; choice != "permanent" {
		t.Fatalf("validity choice = %q, want permanent", choice)
	}
	editor.update(tea.KeyPressMsg{Code: tea.KeyRight})
	if got := editor.value("validity-date:validity_credential"); got != "" {
		t.Fatalf("expiry date input = %q, want empty prompt", got)
	}
	dateIndex := editor.indexOf("validity-date:validity_credential")
	if dateIndex == 0 || !editor.isVisible(editor.fields[dateIndex]) || editor.fields[dateIndex].input.Placeholder != "YYYY-MM-DD" {
		t.Fatalf("expiry date prompt = %#v", editor.fields)
	}
	setEditorValue(t, editor, "validity-date:validity_credential", "2027-12-31")
	if got := editor.createRequest().Validity; got != "2027-12-31" {
		t.Fatalf("create validity = %q", got)
	}
}

func TestCreateEditorModalEditFlow(t *testing.T) {
	editor := newCreateEditor(app.DefaultSecretCreateOptions())
	if editor.editMode != editNormal {
		t.Fatalf("initial editMode = %v, want normal", editor.editMode)
	}
	titleIdx := editor.indexOf("title")
	editor.focus(titleIdx)
	// i enters insert mode
	editor.update(tea.KeyPressMsg{Text: "i"})
	if editor.editMode != editInsert {
		t.Fatalf("editMode after i = %v, want insert", editor.editMode)
	}
	// esc returns to normal mode
	editor.update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if editor.editMode != editNormal {
		t.Fatalf("editMode after esc = %v, want normal", editor.editMode)
	}
	// a enters insert mode
	editor.update(tea.KeyPressMsg{Text: "a"})
	if editor.editMode != editInsert {
		t.Fatalf("editMode after a = %v, want insert", editor.editMode)
	}
	// j/k don't navigate in insert mode
	editor.update(tea.KeyPressMsg{Text: "j"})
	if editor.selected != titleIdx {
		t.Fatalf("j navigated in insert mode: selected=%d", editor.selected)
	}
	// h/l don't cycle choices in insert mode
	editor.update(tea.KeyPressMsg{Text: "h"})
	// Return to normal mode and verify j/k navigate
	editor.update(tea.KeyPressMsg{Code: tea.KeyEscape})
	editor.update(tea.KeyPressMsg{Text: "j"})
	if editor.selected == titleIdx {
		t.Fatalf("j did not navigate in normal mode")
	}
	editor.update(tea.KeyPressMsg{Text: "k"})
	if editor.selected != titleIdx {
		t.Fatalf("k did not navigate back")
	}
}

func setEditorValue(t *testing.T, editor *itemEditor, id, value string) {
	t.Helper()
	for index := range editor.fields {
		if editor.fields[index].id == id {
			editor.fields[index].input.SetValue(value)
			return
		}
	}
	t.Fatalf("editor field %q not found", id)
}
