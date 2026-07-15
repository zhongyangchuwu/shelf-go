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
	if m.detail.Description != "" {
		fmt.Fprintf(&b, "description: %s\n", m.detail.Description)
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
	structuralByTarget := make(map[string][]int)
	for i, field := range m.detail.Fields {
		if _, target, structural := structuralDetailField(field.ID); structural {
			structuralByTarget[target] = append(structuralByTarget[target], i)
		}
	}
	for i, field := range m.detail.Fields {
		if _, _, structural := structuralDetailField(field.ID); structural {
			continue
		}
		cursor := m.detailCursor(i)
		if field.Sensitivity == "secret" {
			state := "unset"
			if field.ValueSet {
				state = "set"
			}
			fmt.Fprintf(&b, "%s %-18s secret · %s\n", cursor, field.ID, state)
		} else {
			value := "unset"
			if field.ValueSet {
				value = field.Value
			}
			fmt.Fprintf(&b, "%s %-18s %s\n", cursor, field.ID, value)
		}
		for _, structuralIndex := range structuralByTarget[field.ID] {
			structural := m.detail.Fields[structuralIndex]
			kind, _, _ := structuralDetailField(structural.ID)
			value := "unset"
			if structural.ValueSet {
				value = structural.Value
			}
			fmt.Fprintf(&b, "%s   %-15s %s\n", m.detailCursor(structuralIndex), kind+":", value)
		}
	}
	return b.String()
}

func (m Model) detailCursor(index int) string {
	if m.pane == paneDetail && index == m.fieldSelected {
		return "›"
	}
	return " "
}

func structuralDetailField(fieldID string) (kind, target string, structural bool) {
	if target, structural = strings.CutPrefix(fieldID, "env_"); structural && target != "" {
		return "env", target, true
	}
	if target, structural = strings.CutPrefix(fieldID, "validity_"); structural && target != "" {
		return "validity", target, true
	}
	return "", "", false
}

func newDetailViewport() viewport.Model {
	viewport := viewport.New()
	viewport.SoftWrap = true
	return viewport
}
