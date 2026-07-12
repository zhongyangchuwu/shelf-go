package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
)

func (m *Model) configureDetailViewport() {
	leftWidth := max(34, m.width*36/100)
	m.detailViewport.SetWidth(max(10, m.width-leftWidth-3))
	m.detailViewport.SetHeight(max(4, m.height-9))
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
