package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

func TestCreateActionPassesMetadataAndDropsSecretDraft(t *testing.T) {
	m, service, _ := readyModel(t)
	secretDraft := fakeSecret
	m.form = formState{kind: formCreate, create: &createDraft{
		kind: "api", title: "New API", id: "item_new_api", env: "NEW_API_KEY",
		labels: "api, work", aliases: "new-api", provider: "Example", scopes: "read, write", secret: secretDraft,
	}}
	updated, cmd := m.submitForm()
	m = updated.(Model)
	if m.form.create.secret != "" {
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
	if request.Template != vaultmodel.TemplateAPICredential || strings.Join(request.Labels, ",") != "api,work" || strings.Join(request.Aliases, ",") != "new-api" {
		t.Fatalf("create request = %#v", request)
	}
	if got := m.render(); strings.Contains(got, fakeSecret) {
		t.Fatalf("create view exposed secret: %q", got)
	}
}

func TestFieldUpdateAndDeleteUseAppWorkflows(t *testing.T) {
	m, service, _ := readyModel(t)
	m.pane = paneDetail
	m.fieldSelected = 1
	m.form = formState{kind: formField, field: &fieldDraft{id: "provider", fieldType: "string", sensitivity: "public", value: "Example"}}
	_, cmd := m.submitForm()
	if result := cmd().(actionDoneMsg); result.err != nil || !result.refresh {
		t.Fatalf("field update result = %#v", result)
	}
	if len(service.updated) != 1 || !service.updated[0].ReadField || service.updated[0].Field != "provider" {
		t.Fatalf("update requests = %#v", service.updated)
	}

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

func TestSecretFieldEditorIsMaskedAndNeverPrefilled(t *testing.T) {
	m, _, _ := readyModel(t)
	secretField := m.detail.Fields[0]
	updated, _ := m.startFieldForm(secretField)
	m = updated.(Model)
	if m.form.field.value != "" {
		t.Fatalf("secret editor value = %q, want empty", m.form.field.value)
	}
	if got := m.renderForm(80); strings.Contains(got, fakeSecret) {
		t.Fatalf("secret editor view exposed secret: %q", got)
	}
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.screen != screenWorkspace || m.form.kind != formNone {
		t.Fatalf("cancelled form state = screen %v form %v", m.screen, m.form.kind)
	}
}
