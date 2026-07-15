package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

type formKind uint8

const (
	formNone formKind = iota
	formCreate
	formEdit
)

type formState struct {
	kind   formKind
	editor *itemEditor
}

type editorFieldKind uint8

const (
	editorChoice editorFieldKind = iota
	editorInput
	editorSave
)

type editorField struct {
	id           string
	label        string
	kind         editorFieldKind
	input        textinput.Model
	initial      string
	fieldID      string
	fieldType    string
	sensitivity  string
	optionValues []string
	optionLabels []string
	option       int
}

type itemEditor struct {
	mode     formKind
	template string
	profile  string
	options  []app.SecretCreateOption
	fields   []editorField
	selected int
	scroll   int
	err      string
}

func (m Model) startCreateForm() (tea.Model, tea.Cmd) {
	m.form = formState{kind: formCreate, editor: newCreateEditor(m.service.SecretCreateOptions())}
	m.screen = screenForm
	return m, nil
}

func (m Model) startEditForm() (tea.Model, tea.Cmd) {
	if m.currentID() == "" {
		return m, nil
	}
	m.form = formState{kind: formEdit, editor: newEditEditor(m.detail)}
	m.screen = screenForm
	return m, nil
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "esc" {
		m.form = formState{}
		m.screen = screenWorkspace
		return m, nil
	}
	if m.form.editor == nil {
		m.screen = screenError
		return m, nil
	}
	cmd, submit := m.form.editor.update(msg)
	if submit {
		return m.submitForm()
	}
	return m, cmd
}

func (m Model) renderForm(_ int) string {
	if m.form.editor == nil {
		return "Editor unavailable."
	}
	height := workspacePanelHeight(m.height, m.status != "") - 4
	return m.form.editor.view(height)
}

func requireValue(name string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
		return nil
	}
}

func newCreateEditor(options []app.SecretCreateOption) *itemEditor {
	editor := &itemEditor{mode: formCreate, options: options}
	if len(options) == 0 {
		editor.err = "No create options are available."
		return editor
	}
	editor.template = options[0].ID
	editor.rebuildCreateFields("")
	return editor
}

func newEditEditor(info app.SecretInfo) *itemEditor {
	editor := &itemEditor{mode: formEdit, template: info.Template}
	editor.fields = append(editor.fields,
		newEditorInput("title", "Title", info.Title, "", "string", "public"),
		newEditorInput("description", "Description", info.Description, "Optional context", "string", "public"),
		newEditorInput("labels", "Labels", strings.Join(info.Labels, ", "), "Comma-separated", "string_list", "public"),
	)
	for _, field := range info.Fields {
		value := editableFieldValue(field)
		placeholder := "Optional"
		if string(field.Sensitivity) == "secret" {
			value = ""
			placeholder = "Leave blank to keep the current secret"
		}
		editor.fields = append(editor.fields, newEditorInput("field:"+field.ID, editorFieldLabel(field.ID), value, placeholder, string(field.Type), string(field.Sensitivity)))
		editor.fields[len(editor.fields)-1].fieldID = field.ID
	}
	editor.fields = append(editor.fields, editorField{id: "save", label: "Save changes", kind: editorSave})
	editor.focus(0)
	return editor
}

func newEditorInput(id, label, value, placeholder, fieldType, sensitivity string) editorField {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = placeholder
	input.CharLimit = 4096
	input.SetWidth(40)
	input.SetValue(value)
	if sensitivity == "secret" {
		input.EchoMode = textinput.EchoPassword
		input.EchoCharacter = '•'
	}
	return editorField{
		id:          id,
		label:       label,
		kind:        editorInput,
		input:       input,
		initial:     value,
		fieldType:   fieldType,
		sensitivity: sensitivity,
	}
}

func (editor itemEditor) createOption() (app.SecretCreateOption, bool) {
	for _, option := range editor.options {
		if option.ID == editor.template {
			return option, true
		}
	}
	return app.SecretCreateOption{}, false
}

func (editor *itemEditor) rebuildCreateFields(selectedID string) {
	values := editor.valuesByID()
	if selectedID == "" && editor.selected >= 0 && editor.selected < len(editor.fields) {
		selectedID = editor.fields[editor.selected].id
	}
	option, ok := editor.createOption()
	if !ok {
		editor.err = "Selected create option is unavailable."
		return
	}
	kindValues := make([]string, 0, len(editor.options))
	kindLabels := make([]string, 0, len(editor.options))
	kindIndex := 0
	for index, candidate := range editor.options {
		kindValues = append(kindValues, candidate.ID)
		kindLabels = append(kindLabels, candidate.Label)
		if candidate.ID == editor.template {
			kindIndex = index
		}
	}
	editor.fields = []editorField{{id: "kind", label: "Kind", kind: editorChoice, optionValues: kindValues, optionLabels: kindLabels, option: kindIndex}}
	if len(option.Presets) > 0 {
		presetValues := []string{""}
		presetLabels := []string{"No preset"}
		presetIndex := 0
		for index, preset := range option.Presets {
			presetValues = append(presetValues, preset.ID)
			presetLabels = append(presetLabels, preset.Label)
			if preset.ID == editor.profile {
				presetIndex = index + 1
			}
		}
		editor.fields = append(editor.fields, editorField{id: "preset", label: "Preset", kind: editorChoice, optionValues: presetValues, optionLabels: presetLabels, option: presetIndex})
	}
	editor.fields = append(editor.fields,
		newEditorInput("title", "Title", values["title"], "Required", "string", "public"),
		newEditorInput("description", "Description", values["description"], "Optional context", "string", "public"),
		newEditorInput("labels", "Labels", values["labels"], "Comma-separated", "string_list", "public"),
	)
	defaults := createDefaults(option, editor.profile)
	for _, field := range option.Fields {
		value, exists := values["field:"+field.ID]
		if !exists || value == "" {
			value = defaults[field.ID]
		}
		placeholder := "Optional"
		if field.Required {
			placeholder = "Required"
		}
		if field.Sensitivity == "secret" {
			placeholder = "Enter secret"
		}
		editor.fields = append(editor.fields, newEditorInput("field:"+field.ID, editorFieldLabel(field.ID), value, placeholder, field.Type, field.Sensitivity))
		editor.fields[len(editor.fields)-1].fieldID = field.ID
	}
	editor.fields = append(editor.fields, editorField{id: "save", label: "Create item", kind: editorSave})
	editor.focus(editor.indexOf(selectedID))
}

func createDefaults(option app.SecretCreateOption, profile string) map[string]string {
	defaults := make(map[string]string, len(option.Defaults))
	for id, value := range option.Defaults {
		defaults[id] = value
	}
	for _, preset := range option.Presets {
		if preset.ID != profile {
			continue
		}
		for id, value := range preset.Defaults {
			defaults[id] = value
		}
		break
	}
	return defaults
}

func editorFieldLabel(id string) string {
	if target, ok := strings.CutPrefix(id, "env_"); ok && target != "" {
		return strings.ReplaceAll(target, "_", " ") + " environment variable"
	}
	if target, ok := strings.CutPrefix(id, "validity_"); ok && target != "" {
		return strings.ReplaceAll(target, "_", " ") + " validity"
	}
	return strings.ReplaceAll(id, "_", " ")
}

func (editor *itemEditor) valuesByID() map[string]string {
	values := make(map[string]string, len(editor.fields))
	for _, field := range editor.fields {
		if field.kind == editorInput {
			values[field.id] = field.input.Value()
		}
	}
	return values
}

func (editor *itemEditor) indexOf(id string) int {
	for index, field := range editor.fields {
		if field.id == id {
			return index
		}
	}
	return 0
}

func (editor *itemEditor) focus(index int) {
	if len(editor.fields) == 0 {
		editor.selected = 0
		return
	}
	editor.selected = min(max(0, index), len(editor.fields)-1)
	for i := range editor.fields {
		if editor.fields[i].kind != editorInput {
			continue
		}
		if i == editor.selected {
			editor.fields[i].input.Focus()
		} else {
			editor.fields[i].input.Blur()
		}
	}
}

func (editor *itemEditor) update(msg tea.Msg) (tea.Cmd, bool) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, false
	}
	if key.String() == "ctrl+s" {
		return nil, true
	}
	if len(editor.fields) == 0 {
		return nil, false
	}
	current := &editor.fields[editor.selected]
	switch key.String() {
	case "tab", "down", "enter":
		if current.kind == editorSave {
			return nil, true
		}
		editor.focus((editor.selected + 1) % len(editor.fields))
		return nil, false
	case "shift+tab", "up":
		editor.focus((editor.selected - 1 + len(editor.fields)) % len(editor.fields))
		return nil, false
	}
	if current.kind == editorChoice {
		switch key.String() {
		case "left":
			editor.cycleChoice(-1)
		case "right", "space":
			editor.cycleChoice(1)
		}
		return nil, false
	}
	if current.kind != editorInput {
		return nil, false
	}
	updated, cmd := current.input.Update(msg)
	current.input = updated
	return cmd, false
}

func (editor *itemEditor) cycleChoice(delta int) {
	field := &editor.fields[editor.selected]
	if len(field.optionValues) == 0 {
		return
	}
	field.option = (field.option + delta + len(field.optionValues)) % len(field.optionValues)
	switch field.id {
	case "kind":
		editor.template = field.optionValues[field.option]
		editor.profile = ""
		editor.rebuildCreateFields(field.id)
	case "preset":
		editor.profile = field.optionValues[field.option]
		editor.rebuildCreateFields(field.id)
	}
}

func (editor itemEditor) value(id string) string {
	for _, field := range editor.fields {
		if field.id == id && field.kind == editorInput {
			return field.input.Value()
		}
	}
	return ""
}

func (editor itemEditor) field(id string) (editorField, bool) {
	for _, field := range editor.fields {
		if field.fieldID == id {
			return field, true
		}
	}
	return editorField{}, false
}

func (editor *itemEditor) clearSecrets() {
	for i := range editor.fields {
		if editor.fields[i].sensitivity == "secret" {
			editor.fields[i].input.SetValue("")
		}
	}
}

func (editor *itemEditor) validate() error {
	if strings.TrimSpace(editor.value("title")) == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

func (editor itemEditor) view(height int) string {
	if len(editor.fields) == 0 {
		return "Editor unavailable."
	}
	visible := max(4, height-5)
	start := editor.scroll
	if editor.selected < start {
		start = editor.selected
	}
	if editor.selected >= start+visible {
		start = editor.selected - visible + 1
	}
	start = min(max(0, start), max(0, len(editor.fields)-visible))
	end := min(len(editor.fields), start+visible)
	var b strings.Builder
	if start > 0 {
		b.WriteString("↑ more fields\n")
	}
	for index := start; index < end; index++ {
		field := editor.fields[index]
		cursor := " "
		if index == editor.selected {
			cursor = "›"
		}
		switch field.kind {
		case editorChoice:
			fmt.Fprintf(&b, "%s %s: %s\n", cursor, field.label, field.optionLabels[field.option])
		case editorInput:
			fmt.Fprintf(&b, "%s %s: %s\n", cursor, field.label, field.input.View())
		case editorSave:
			fmt.Fprintf(&b, "%s %s\n", cursor, field.label)
		}
	}
	if end < len(editor.fields) {
		b.WriteString("↓ more fields\n")
	}
	if editor.err != "" {
		b.WriteString("\n" + editor.err + "\n")
	}
	b.WriteString("\nTab next  ↑/↓ move  ←/→ choose  ctrl+s save  esc cancel")
	return b.String()
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
