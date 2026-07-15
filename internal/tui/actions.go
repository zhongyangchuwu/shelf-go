package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func (m Model) submitForm() (tea.Model, tea.Cmd) {
	if m.form.editor == nil {
		m.screen = screenError
		return m, nil
	}
	editor := m.form.editor
	if err := editor.validate(); err != nil {
		editor.err = err.Error()
		return m, nil
	}
	switch m.form.kind {
	case formCreate:
		req := editor.createRequest()
		editor.clearSecrets()
		m.screen = screenLoading
		return m, func() tea.Msg {
			_, err := m.service.CreateSecret(m.options.ConfigPath, m.options.VaultPath, req)
			return actionDoneMsg{status: "Created secret item.", refresh: true, err: err}
		}
	case formEdit:
		req, changed := editor.updateRequest(m.detail, m.currentID())
		editor.clearSecrets()
		if !changed {
			m.form = formState{}
			m.screen = screenWorkspace
			m.status = "No changes."
			return m, nil
		}
		m.screen = screenLoading
		return m, func() tea.Msg {
			err := m.service.UpdateSecret(m.options.ConfigPath, m.options.VaultPath, req)
			return actionDoneMsg{status: "Updated secret item.", refresh: true, err: err}
		}
	default:
		m.screen = screenError
		return m, nil
	}
}

func (editor itemEditor) createRequest() app.CreateSecretRequest {
	req := app.CreateSecretRequest{
		Title:       editor.value("title"),
		Description: editor.value("description"),
		Labels:      splitTokens(editor.value("labels")),
		Template:    editor.template,
	}
	if editor.template == "api_credential" && editor.profile != "" {
		req.Profile = editor.profile
	}
	primaryID := "value"
	if editor.template == "api_credential" {
		primaryID = "credential"
	}
	if secret, ok := editor.field(primaryID); ok && secret.input.Value() != "" {
		req.Stdin = true
		req.In = strings.NewReader(secret.input.Value())
	}
	if env, ok := editor.field("env_" + primaryID); ok {
		req.Env = env.input.Value()
	}
	if field, ok := editor.field("provider"); ok {
		req.Provider = field.input.Value()
	}
	if field, ok := editor.field("website"); ok {
		req.Website = field.input.Value()
	}
	if field, ok := editor.field("validity_" + primaryID); ok {
		req.Validity = field.input.Value()
	}
	if field, ok := editor.field("scopes"); ok {
		req.Scopes = splitTokens(field.input.Value())
	}
	if field, ok := editor.field("base_url"); ok {
		req.BaseURL = field.input.Value()
	}
	if field, ok := editor.field("env_base_url"); ok {
		req.BaseURLEnv = field.input.Value()
	}
	if field, ok := editor.field("client_id"); ok {
		req.ClientID = field.input.Value()
	}
	if field, ok := editor.field("env_client_id"); ok {
		req.ClientIDEnv = field.input.Value()
	}
	return req
}

func (editor itemEditor) updateRequest(detail app.SecretInfo, selector string) (app.UpdateSecretRequest, bool) {
	req := app.UpdateSecretRequest{Selector: selector}
	title := editor.value("title")
	if title != detail.Title {
		req.Title = title
		req.TitleSet = true
	}
	description := editor.value("description")
	if description != detail.Description {
		req.Description = description
		req.DescriptionSet = true
	}
	desiredLabels := splitTokens(editor.value("labels"))
	req.AddLabels = additions(detail.Labels, desiredLabels)
	req.RemoveLabels = additions(desiredLabels, detail.Labels)
	for _, original := range detail.Fields {
		field, ok := editor.field(original.ID)
		if !ok {
			continue
		}
		value := field.input.Value()
		if string(original.Sensitivity) == "secret" {
			if value != "" {
				req.FieldUpdates = append(req.FieldUpdates, app.FieldUpdate{Field: original.ID, In: strings.NewReader(value), ReadField: true})
			}
			continue
		}
		initial := editableFieldValue(original)
		if value == initial {
			continue
		}
		if value == "" {
			req.FieldUpdates = append(req.FieldUpdates, app.FieldUpdate{Field: original.ID, ClearField: true})
			continue
		}
		req.FieldUpdates = append(req.FieldUpdates, app.FieldUpdate{
			Field:         original.ID,
			Values:        editorValues(field.fieldType, value),
			FieldValueSet: true,
		})
	}
	changed := req.TitleSet || req.DescriptionSet || len(req.AddLabels) > 0 || len(req.RemoveLabels) > 0 || len(req.FieldUpdates) > 0
	return req, changed
}

func editorValues(fieldType string, value string) []string {
	if fieldType == "string_list" {
		return splitTokens(value)
	}
	return []string{value}
}

func (m Model) copySelectedField() (tea.Model, tea.Cmd) {
	if m.currentID() == "" {
		return m, nil
	}
	field := m.selectedFieldID()
	return m, func() tea.Msg {
		value, err := m.service.GetSecretField(m.options.ConfigPath, m.options.VaultPath, app.GetSecretFieldRequest{Selector: m.currentID(), Field: field})
		if err != nil {
			return actionDoneMsg{err: err}
		}
		if err := m.clipboardWriter.CopyText(value); err != nil {
			return actionDoneMsg{err: err}
		}
		return actionDoneMsg{status: "Copied " + displayField(field) + " to clipboard."}
	}
}

func (m Model) handleDeleteConfirm(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenWorkspace
		return m, nil
	case "enter":
		id := m.currentID()
		m.screen = screenLoading
		return m, func() tea.Msg {
			err := m.service.RemoveSecret(m.options.ConfigPath, m.options.VaultPath, id)
			return actionDoneMsg{status: "Removed secret item.", refresh: true, err: err}
		}
	default:
		return m, nil
	}
}

func (m Model) selectedFieldID() string {
	if m.fieldSelected >= 0 && m.fieldSelected < len(m.detail.Fields) {
		return m.detail.Fields[m.fieldSelected].ID
	}
	return ""
}

func splitTokens(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func additions(current, desired []string) []string {
	have := make(map[string]struct{}, len(current))
	for _, value := range current {
		have[value] = struct{}{}
	}
	out := make([]string, 0, len(desired))
	for _, value := range desired {
		if _, exists := have[value]; !exists {
			out = append(out, value)
		}
	}
	return out
}

func displayField(field string) string {
	if field == "" {
		return "secret"
	}
	return field
}
