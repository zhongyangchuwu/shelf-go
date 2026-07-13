package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

func (m Model) submitForm() (tea.Model, tea.Cmd) {
	form := m.form
	m.screen = screenLoading
	switch form.kind {
	case formCreate:
		draft := *form.create
		form.create.secret = ""
		return m, func() tea.Msg {
			req := app.CreateSecretRequest{
				Title:    draft.title,
				ID:       draft.id,
				Env:      draft.env,
				Labels:   splitTokens(draft.labels),
				Aliases:  splitTokens(draft.aliases),
				Provider: draft.provider,
				Website:  draft.website,
				Validity: draft.validity,
				Scopes:   splitTokens(draft.scopes),
				Stdin:    true,
				In:       strings.NewReader(draft.secret),
			}
			switch draft.kind {
			case "api":
				req.Template = "api_credential"
			case "openai":
				req.Profile = "openai"
			default:
				req.Template = "generic_secret"
			}
			_, err := m.service.CreateSecret(m.options.ConfigPath, m.options.VaultPath, req)
			return actionDoneMsg{status: "Created secret item.", refresh: true, err: err}
		}
	case formMetadata:
		draft := *form.metadata
		return m, func() tea.Msg {
			addLabels := additions(m.detail.Labels, splitTokens(draft.labels))
			removeLabels := additions(splitTokens(draft.labels), m.detail.Labels)
			addAliases := additions(m.detail.Aliases, splitTokens(draft.aliases))
			removeAliases := additions(splitTokens(draft.aliases), m.detail.Aliases)
			titleSet := draft.title != m.detail.Title
			if !titleSet && len(addLabels) == 0 && len(removeLabels) == 0 && len(addAliases) == 0 && len(removeAliases) == 0 {
				return actionDoneMsg{status: "No changes."}
			}
			err := m.service.UpdateSecret(m.options.ConfigPath, m.options.VaultPath, app.UpdateSecretRequest{
				Selector:      m.currentID(),
				Title:         draft.title,
				TitleSet:      titleSet,
				AddLabels:     addLabels,
				RemoveLabels:  removeLabels,
				AddAliases:    addAliases,
				RemoveAliases: removeAliases,
			})
			return actionDoneMsg{status: "Updated secret item.", refresh: true, err: err}
		}
	case formField:
		draft := *form.field
		form.field.value = ""
		return m, func() tea.Msg {
			req := app.UpdateSecretRequest{Selector: m.currentID(), Field: draft.id}
			if draft.clear {
				req.ClearField = true
			} else {
				req.ReadField = true
				req.In = strings.NewReader(fieldInput(draft.fieldType, draft.value))
			}
			err := m.service.UpdateSecret(m.options.ConfigPath, m.options.VaultPath, req)
			return actionDoneMsg{status: "Updated secret item.", refresh: true, err: err}
		}
	default:
		m.screen = screenError
		return m, nil
	}
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

func fieldInput(fieldType string, value string) string {
	if fieldType != "string_list" {
		return value
	}
	return strings.Join(splitTokens(value), "\n")
}

func displayField(field string) string {
	if field == "" {
		return "secret"
	}
	return field
}
