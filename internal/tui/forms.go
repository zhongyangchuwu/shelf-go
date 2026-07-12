package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

type formKind uint8

const (
	formNone formKind = iota
	formCreate
	formMetadata
	formField
)

type formState struct {
	kind     formKind
	form     *huh.Form
	create   *createDraft
	metadata *metadataDraft
	field    *fieldDraft
}

type createDraft struct {
	kind     string
	title    string
	id       string
	env      string
	labels   string
	aliases  string
	provider string
	website  string
	expires  string
	scopes   string
	secret   string
}

type metadataDraft struct {
	title   string
	labels  string
	aliases string
}

type fieldDraft struct {
	id          string
	fieldType   string
	sensitivity string
	value       string
	clear       bool
}

func (m Model) startCreateForm() (tea.Model, tea.Cmd) {
	draft := &createDraft{kind: "generic"}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Item type").
				Options(
					huh.NewOption("Generic secret", "generic"),
					huh.NewOption("API credential", "api"),
					huh.NewOption("OpenAI profile", "openai"),
				).
				Value(&draft.kind),
			huh.NewInput().Title("Title").Value(&draft.title).Validate(requireValue("title")),
			huh.NewInput().Title("Stable ID (optional)").Value(&draft.id),
			huh.NewInput().Title("Environment variable name (optional)").Value(&draft.env),
			huh.NewInput().Title("Labels (comma separated)").Value(&draft.labels),
			huh.NewInput().Title("Aliases (comma separated)").Value(&draft.aliases),
		),
		huh.NewGroup(
			huh.NewInput().Title("Provider").Value(&draft.provider),
			huh.NewInput().Title("Website").Value(&draft.website),
			huh.NewInput().Title("Expiry date (YYYY-MM-DD)").Value(&draft.expires),
			huh.NewInput().Title("Scopes (comma separated)").Value(&draft.scopes),
		).WithHideFunc(func() bool { return draft.kind == "generic" }),
		huh.NewGroup(
			huh.NewInput().Title("Primary secret value").EchoMode(huh.EchoModePassword).Value(&draft.secret),
		),
		huh.NewGroup(
			huh.NewNote().Title("Review").DescriptionFunc(func() string { return createReview(draft) }, draft).Next(true).NextLabel("Create item"),
		),
	)
	m.form = formState{kind: formCreate, form: form, create: draft}
	m.screen = screenForm
	return m, form.Init()
}

func (m Model) startEditForm() (tea.Model, tea.Cmd) {
	if m.currentID() == "" {
		return m, nil
	}
	if m.pane == paneDetail && m.fieldSelected >= 0 && m.fieldSelected < len(m.detail.Fields) {
		return m.startFieldForm(m.detail.Fields[m.fieldSelected])
	}
	draft := &metadataDraft{
		title:   m.detail.Title,
		labels:  strings.Join(m.detail.Labels, ", "),
		aliases: strings.Join(m.detail.Aliases, ", "),
	}
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Title").Value(&draft.title).Validate(requireValue("title")),
		huh.NewInput().Title("Labels (comma separated)").Value(&draft.labels),
		huh.NewInput().Title("Aliases (comma separated)").Value(&draft.aliases),
	))
	m.form = formState{kind: formMetadata, form: form, metadata: draft}
	m.screen = screenForm
	return m, form.Init()
}

func (m Model) startFieldForm(field app.SecretFieldInfo) (tea.Model, tea.Cmd) {
	draft := &fieldDraft{
		id:          field.ID,
		fieldType:   string(field.Type),
		sensitivity: string(field.Sensitivity),
		value:       editableFieldValue(field),
	}
	input := huh.NewInput().Title("Value for " + field.ID).Value(&draft.value)
	if draft.sensitivity == "secret" {
		input.EchoMode(huh.EchoModePassword).Placeholder("Enter replacement value")
	} else if draft.fieldType == "string_list" {
		input.Description("Comma-separated values")
	}
	fields := []huh.Field{input}
	if draft.sensitivity == "public" {
		fields = append(fields, huh.NewConfirm().Title("Clear this public field").Value(&draft.clear))
	}
	form := huh.NewForm(huh.NewGroup(fields...))
	m.form = formState{kind: formField, form: form, field: draft}
	m.screen = screenForm
	return m, form.Init()
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "esc" {
		m.form = formState{}
		m.screen = screenWorkspace
		return m, nil
	}
	updated, cmd := m.form.form.Update(msg)
	form, ok := updated.(*huh.Form)
	if !ok {
		m.screen = screenError
		return m, nil
	}
	m.form.form = form
	if form.State != huh.StateCompleted {
		return m, cmd
	}
	return m.submitForm()
}

func (m Model) renderForm(width int) string {
	if m.form.form == nil {
		return "Form unavailable."
	}
	return m.form.form.View()
}

func requireValue(name string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
		return nil
	}
}

func editableFieldValue(field app.SecretFieldInfo) string {
	if string(field.Sensitivity) == "secret" || !field.ValueSet {
		return ""
	}
	if string(field.Type) != "string_list" {
		return field.Value
	}
	var values []string
	if err := json.Unmarshal([]byte(field.Value), &values); err != nil {
		return field.Value
	}
	return strings.Join(values, ", ")
}

func createReview(draft *createDraft) string {
	itemType := map[string]string{
		"generic": "Generic secret",
		"api":     "API credential",
		"openai":  "OpenAI profile",
	}[draft.kind]
	lines := []string{
		"Type: " + itemType,
		"Title: " + draft.title,
		"Stable ID: " + optionalValue(draft.id),
		"Environment variable: " + optionalValue(draft.env),
		"Labels: " + optionalValue(draft.labels),
		"Aliases: " + optionalValue(draft.aliases),
		"Primary secret: will be set on save",
	}
	if draft.kind != "generic" {
		lines = append(lines,
			"Provider: "+optionalValue(draft.provider),
			"Website: "+optionalValue(draft.website),
			"Expiry date: "+optionalValue(draft.expires),
			"Scopes: "+optionalValue(draft.scopes),
		)
	}
	return strings.Join(lines, "\n")
}

func optionalValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(not set)"
	}
	return value
}
