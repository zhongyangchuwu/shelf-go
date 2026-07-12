package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/vaultmodel"
)

const fakeSecret = "fake-secret-value"

type fakeService struct {
	catalog []app.SecretCatalogItem
	info    map[string]app.SecretInfo

	created []app.CreateSecretRequest
	updated []app.UpdateSecretRequest
	removed []string
	value   string
	err     error
}

func (s *fakeService) SecretCatalog(_, _ string) ([]app.SecretCatalogItem, error) {
	return s.catalog, s.err
}

func (s *fakeService) SecretInfo(_, _, selector string) (app.SecretInfo, error) {
	if s.err != nil {
		return app.SecretInfo{}, s.err
	}
	return s.info[selector], nil
}

func (s *fakeService) CreateSecret(_, _ string, req app.CreateSecretRequest) (string, error) {
	s.created = append(s.created, req)
	if s.err != nil {
		return "", s.err
	}
	return "item_created", nil
}

func (s *fakeService) UpdateSecret(_, _ string, req app.UpdateSecretRequest) error {
	s.updated = append(s.updated, req)
	return s.err
}

func (s *fakeService) GetSecretField(_, _ string, _ app.GetSecretFieldRequest) (string, error) {
	return s.value, s.err
}

func (s *fakeService) RemoveSecret(_, _, selector string) error {
	s.removed = append(s.removed, selector)
	return s.err
}

type fakeClipboard struct {
	copied string
	err    error
}

func (c *fakeClipboard) CopyText(value string) error {
	c.copied = value
	return c.err
}

func readyModel(t *testing.T) (Model, *fakeService, *fakeClipboard) {
	t.Helper()
	service := &fakeService{
		catalog: []app.SecretCatalogItem{
			{ID: "item_openai", Title: "OpenAI personal", Template: vaultmodel.TemplateAPICredential, Labels: []string{"ai", "personal"}, Aliases: []string{"openai"}, ValueSet: true},
			{ID: "item_github", Title: "GitHub work", Template: vaultmodel.TemplateGenericSecret, Labels: []string{"work"}, Aliases: []string{"github"}, ValueSet: true},
		},
		info: map[string]app.SecretInfo{
			"item_openai": {
				ID: "item_openai", Title: "OpenAI personal", Template: vaultmodel.TemplateAPICredential,
				Labels: []string{"ai", "personal"}, Aliases: []string{"openai"},
				Fields: []app.SecretFieldInfo{
					{ID: "credential", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, ValueSet: true},
					{ID: "provider", Type: vaultmodel.FieldTypeString, Sensitivity: vaultmodel.SensitivityPublic, Value: "OpenAI", ValueSet: true},
				},
			},
			"item_github": {
				ID: "item_github", Title: "GitHub work", Template: vaultmodel.TemplateGenericSecret,
				Fields: []app.SecretFieldInfo{{ID: "value", Type: vaultmodel.FieldTypeConcealed, Sensitivity: vaultmodel.SensitivitySecret, ValueSet: true}},
			},
		},
		value: fakeSecret,
	}
	clipboard := &fakeClipboard{}
	m := NewModel(service, clipboard, Options{})
	m = updateModel(t, m, tea.WindowSizeMsg{Width: 120, Height: 32})
	m = updateModel(t, m, catalogLoadedMsg{items: service.catalog})
	m = updateModel(t, m, detailLoadedMsg{id: "item_openai", info: service.info["item_openai"]})
	return m, service, clipboard
}

func updateModel(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	updated, _ := m.Update(msg)
	result, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.Model", updated)
	}
	return result
}

func TestModelFiltersSearchAndNeverSearchesFieldValues(t *testing.T) {
	m, _, _ := readyModel(t)
	m.query.SetValue("openai")
	m.applyFilters()
	if got := len(m.filtered); got != 1 {
		t.Fatalf("search result count = %d, want 1", got)
	}
	m.query.SetValue("OpenAI")
	m.applyFilters()
	if got := len(m.filtered); got != 1 {
		t.Fatalf("case-insensitive search result count = %d, want 1", got)
	}
	m.query.SetValue(fakeSecret)
	m.applyFilters()
	if got := len(m.filtered); got != 0 {
		t.Fatalf("secret-value search result count = %d, want 0", got)
	}
}

func TestModelLabelFilterUsesANDSemantics(t *testing.T) {
	m, _, _ := readyModel(t)
	m.labels = map[string]bool{"ai": true, "personal": true}
	m.applyFilters()
	if got := len(m.filtered); got != 1 || m.filtered[0].ID != "item_openai" {
		t.Fatalf("AND label filter = %#v, want only item_openai", m.filtered)
	}
	m.labels["work"] = true
	m.applyFilters()
	if got := len(m.filtered); got != 0 {
		t.Fatalf("three-label AND filter = %d results, want 0", got)
	}
}

func TestModelResizeAndPaneSelection(t *testing.T) {
	m, _, _ := readyModel(t)
	m = updateModel(t, m, tea.WindowSizeMsg{Width: 80, Height: 23})
	if m.screen != screenResize || !strings.Contains(m.render(), "100 columns") {
		t.Fatalf("small terminal state = %v, view = %q", m.screen, m.render())
	}
	m = updateModel(t, m, tea.WindowSizeMsg{Width: 120, Height: 32})
	if m.screen != screenWorkspace {
		t.Fatalf("resized state = %v, want workspace", m.screen)
	}
	m = updateModel(t, m, tea.KeyPressMsg{Text: "l"})
	m = updateModel(t, m, tea.KeyPressMsg{Text: "j"})
	if m.pane != paneDetail || m.fieldSelected != 0 {
		t.Fatalf("detail selection = pane %v field %d, want detail field 0", m.pane, m.fieldSelected)
	}
}

func TestModelCancelsUnappliedFilters(t *testing.T) {
	m, _, _ := readyModel(t)
	m.template = vaultmodel.TemplateAPICredential
	m.labels = map[string]bool{"ai": true}
	m.applyFilters()
	m = updateModel(t, m, tea.KeyPressMsg{Text: "f"})
	m = updateModel(t, m, tea.KeyPressMsg{Text: "c"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.template != vaultmodel.TemplateAPICredential || !m.labels["ai"] {
		t.Fatalf("cancelled filters = template %q labels %#v", m.template, m.labels)
	}
}
