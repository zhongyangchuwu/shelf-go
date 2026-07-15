package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
)

type editorEditMode uint8

const (
	editNormal editorEditMode = iota
	editInsert
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
	editMode editorEditMode
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
		if m.form.editor != nil && m.form.editor.editMode == editInsert {
			cmd, _ := m.form.editor.update(msg)
			return m, cmd
		}
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
		editor.fields = append(editor.fields, newItemEditorFields(field.ID, editorFieldLabel(field.ID), value, placeholder, string(field.Type), string(field.Sensitivity))...)
	}
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

func newItemEditorFields(fieldID, label, value, placeholder, fieldType, sensitivity string) []editorField {
	if _, validity := strings.CutPrefix(fieldID, "validity_"); !validity {
		field := newEditorInput("field:"+fieldID, label, value, placeholder, fieldType, sensitivity)
		field.fieldID = fieldID
		return []editorField{field}
	}
	choice := 0
	date := ""
	switch value {
	case "permanent":
		choice = 1
	case "", "unknown":
		choice = 0
	default:
		choice, date = 2, value
	}
	dateInput := newEditorInput("validity-date:"+fieldID, "Expiry date", date, "YYYY-MM-DD", "string", sensitivity)
	return []editorField{
		{id: "validity:" + fieldID, label: label, kind: editorChoice, fieldID: fieldID, fieldType: "string", sensitivity: sensitivity, optionValues: []string{"unknown", "permanent", "date"}, optionLabels: []string{"Unknown", "Permanent", "Expires on date"}, option: choice},
		dateInput,
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
		editor.fields = append(editor.fields, newItemEditorFields(field.ID, field.Label, value, placeholder, field.Type, field.Sensitivity)...)
	}
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

func (editor *itemEditor) clearPresetDefaults(profile string) {
	if profile == "" {
		return
	}
	option, ok := editor.createOption()
	if !ok {
		return
	}
	for _, preset := range option.Presets {
		if preset.ID != profile {
			continue
		}
		for fieldID, defaultValue := range preset.Defaults {
			for index := range editor.fields {
				field := &editor.fields[index]
				if field.id == "field:"+fieldID && field.kind == editorInput && field.input.Value() == defaultValue {
					field.input.SetValue("")
				}
			}
		}
		return
	}
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
		editor.fields[i].input.Blur()
	}
	if editor.editMode == editInsert && editor.fields[editor.selected].kind == editorInput {
		editor.fields[editor.selected].input.Focus()
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
	if editor.editMode == editInsert {
		if key.String() == "esc" {
			editor.editMode = editNormal
			for i := range editor.fields {
				editor.fields[i].input.Blur()
			}
			return nil, false
		}
		current := &editor.fields[editor.selected]
		if current.kind != editorInput {
			return nil, false
		}
		updated, cmd := current.input.Update(msg)
		current.input = updated
		return cmd, false
	}
	current := &editor.fields[editor.selected]
	switch key.String() {
	case "j", "down":
		editor.focus(editor.nextVisible(1))
	case "k", "up":
		editor.focus(editor.nextVisible(-1))
	}
	if current.kind == editorChoice {
		switch key.String() {
		case "h", "left":
			editor.cycleChoice(-1)
		case "l", "right":
			editor.cycleChoice(1)
		}
		return nil, false
	}
	if current.kind == editorInput && (key.String() == "i" || key.String() == "a") {
		editor.editMode = editInsert
		if key.String() == "a" {
			current.input.CursorEnd()
		} else {
			current.input.CursorStart()
		}
		editor.focus(editor.selected)
		return nil, false
	}
	return nil, false
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
		previous := editor.profile
		editor.clearPresetDefaults(previous)
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

func (editor itemEditor) fieldValue(id string) (string, bool) {
	for _, field := range editor.fields {
		if field.fieldID != id {
			continue
		}
		if field.kind == editorInput {
			return field.input.Value(), true
		}
		if field.kind == editorChoice && strings.HasPrefix(field.id, "validity:") {
			value := field.optionValues[field.option]
			if value == "date" {
				return editor.value("validity-date:" + id), true
			}
			return value, true
		}
	}
	return "", false
}

func (editor itemEditor) validityMode(fieldID string) string {
	for _, field := range editor.fields {
		if field.id == "validity:"+fieldID && field.kind == editorChoice {
			return field.optionValues[field.option]
		}
	}
	return ""
}

func (editor itemEditor) isVisible(field editorField) bool {
	if field.kind != editorInput {
		return true
	}
	target, dateField := strings.CutPrefix(field.id, "validity-date:")
	return !dateField || editor.validityMode(target) == "date"
}

func (editor itemEditor) nextVisible(delta int) int {
	for steps, index := 0, editor.selected; steps < len(editor.fields); steps++ {
		index = (index + delta + len(editor.fields)) % len(editor.fields)
		if editor.isVisible(editor.fields[index]) {
			return index
		}
	}
	return editor.selected
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
	for _, field := range editor.fields {
		if !strings.HasPrefix(field.id, "validity:") || field.kind != editorChoice || field.optionValues[field.option] != "date" {
			continue
		}
		date := strings.TrimSpace(editor.value("validity-date:" + field.fieldID))
		if date == "" {
			return fmt.Errorf("expiry date is required")
		}
		if _, err := time.Parse(time.DateOnly, date); err != nil {
			return fmt.Errorf("expiry date must use YYYY-MM-DD")
		}
	}
	return nil
}

func (editor itemEditor) view(height int) string {
	if len(editor.fields) == 0 {
		return "Editor unavailable."
	}
	indices := make([]int, 0, len(editor.fields))
	selected := 0
	for index, field := range editor.fields {
		if !editor.isVisible(field) {
			continue
		}
		if index == editor.selected {
			selected = len(indices)
		}
		indices = append(indices, index)
	}
	visible := max(4, height-5)
	start := editor.scroll
	if selected < start {
		start = selected
	}
	if selected >= start+visible {
		start = selected - visible + 1
	}
	start = min(max(0, start), max(0, len(indices)-visible))
	end := min(len(indices), start+visible)
	var b strings.Builder
	if start > 0 {
		b.WriteString("↑ more fields\n")
	}
	for _, index := range indices[start:end] {
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
		}
	}
	if end < len(indices) {
		b.WriteString("↓ more fields\n")
	}
	if editor.err != "" {
		b.WriteString("\n" + editor.err + "\n")
	}
	if editor.editMode == editInsert {
		b.WriteString("\n-- INSERT --  esc to finish")
	} else {
		b.WriteString("\nj/k move  h/l choose  i/a edit  ctrl+s save  esc cancel")
	}
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
