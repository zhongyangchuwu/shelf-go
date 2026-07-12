package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
)

func (m *Model) configureDetailViewport() {
	_, rightWidth := workspacePanelWidths(m.width)
	panelHeight := workspacePanelHeight(m.height, m.status != "")
	m.detailViewport.SetWidth(panelContentWidth(rightWidth))
	m.detailViewport.SetHeight(max(4, panelHeight-4))
	m.detailViewport.SoftWrap = true
}

func (m *Model) syncDetailViewport() {
	m.detailViewport.SetContent(m.detailContent())
}

func (m Model) detailContent() string {
	if m.currentID() == "" {
		return "Select an item from the catalog."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", m.detail.Title)
	fmt.Fprintf(&b, "id:       %s\n", m.detail.ID)
	fmt.Fprintf(&b, "template: %s\n", m.detail.Template)
	if len(m.detail.Labels) > 0 {
		fmt.Fprintf(&b, "labels:   %s\n", strings.Join(m.detail.Labels, ", "))
	}
	if len(m.detail.Aliases) > 0 {
		fmt.Fprintf(&b, "aliases:  %s\n", strings.Join(m.detail.Aliases, ", "))
	}
	if health, ok := m.health[m.detail.ID]; ok {
		fmt.Fprintf(&b, "health:   %s", healthLabel(health.Status))
		if health.Expires != "" {
			fmt.Fprintf(&b, " · expires %s", health.Expires)
		}
		b.WriteByte('\n')
	} else if m.healthUnavailable {
		b.WriteString("health:   unavailable\n")
	}
	b.WriteString("\nfields\n")
	for i, field := range m.detail.Fields {
		cursor := " "
		if m.pane == paneDetail && i == m.fieldSelected {
			cursor = "›"
		}
		if field.Sensitivity == "secret" {
			state := "unset"
			if field.ValueSet {
				state = "set"
			}
			fmt.Fprintf(&b, "%s %-18s secret · %s\n", cursor, field.ID, state)
			continue
		}
		value := "unset"
		if field.ValueSet {
			value = field.Value
		}
		fmt.Fprintf(&b, "%s %-18s %s\n", cursor, field.ID, value)
	}
	return b.String()
}

func newDetailViewport() viewport.Model {
	viewport := viewport.New()
	viewport.SoftWrap = true
	return viewport
}
