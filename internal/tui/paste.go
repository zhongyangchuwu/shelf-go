package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Clipboard results belong to the input that requested them, not whichever
// screen or field happens to be focused when the system clipboard responds.
type workspacePasteMsg struct {
	content string
	err     error
	screen  screen
	input   *textinput.Model
}

type projectPasteMsg struct {
	content string
	err     error
	screen  projectScreen
	form    *huh.Form
	input   *huh.Input
}

func (m Model) activeEditorInput() *textinput.Model {
	if m.screen != screenForm || m.form.editor == nil {
		return nil
	}
	editor := m.form.editor
	if editor.editMode != editInsert || len(editor.fields) == 0 {
		return nil
	}
	field := &editor.fields[editor.selected]
	if field.kind != editorInput {
		return nil
	}
	return &field.input
}

func (m Model) readClipboard() tea.Cmd {
	reader := m.clipboardReader
	result := workspacePasteMsg{screen: m.screen, input: m.activeEditorInput()}
	return func() tea.Msg {
		result.content, result.err = reader.ReadText()
		return result
	}
}

func (m Model) finishClipboardPaste(msg workspacePasteMsg) (tea.Model, tea.Cmd) {
	if m.screen != msg.screen || (m.screen != screenSearch && (msg.input == nil || m.activeEditorInput() != msg.input)) {
		return m, nil
	}
	if msg.err != nil {
		m.status = "Could not read system clipboard."
		return m, nil
	}
	m.status = ""
	return m.Update(tea.PasteMsg{Content: msg.content})
}

func (m ProjectModel) projectTextInput() *huh.Input {
	if m.screen != projectForm || m.form.form == nil {
		return nil
	}
	input, _ := m.form.form.GetFocusedField().(*huh.Input)
	return input
}

func (m ProjectModel) readProjectClipboard() tea.Cmd {
	reader := m.clipboardReader
	result := projectPasteMsg{screen: m.screen, form: m.form.form, input: m.projectTextInput()}
	return func() tea.Msg {
		result.content, result.err = reader.ReadText()
		return result
	}
}

func (m ProjectModel) finishProjectClipboardPaste(msg projectPasteMsg) (tea.Model, tea.Cmd) {
	if m.screen != msg.screen || (m.screen != projectPickerSearch && (msg.form != m.form.form || msg.input == nil || msg.input != m.projectTextInput())) {
		return m, nil
	}
	if msg.err != nil {
		m.status = "Could not read system clipboard."
		return m, nil
	}
	m.status = ""
	return m.Update(tea.PasteMsg{Content: msg.content})
}
