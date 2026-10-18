package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m ProjectModel) renderProject() string {
	if m.width > 0 && m.height > 0 && m.projectTooSmall() {
		return renderResize(m.width, m.height)
	}
	switch m.screen {
	case projectLoading:
		return "Shelf  ·  Project bindings\n\nLoading current project…\n"
	case projectError:
		return "Shelf  ·  Project bindings\n\nThe project workspace action could not be completed.\nNo secret value was displayed.\n\nr retry   q quit\n"
	case projectHelp:
		return m.renderProjectHelp()
	}
	return m.renderProjectWorkspace()
}

func (m ProjectModel) renderProjectWorkspace() string {
	projectID := m.workspace.ProjectID
	if projectID == "" {
		projectID = "current worktree"
	}
	header := m.styles.renderHeader(fmt.Sprintf("Shelf  ·  Project bindings    %s  ·  %d bindings", projectID, len(m.workspace.Bindings)), m.width)
	leftWidth, rightWidth := workspacePanelWidths(m.width)
	panelHeight := workspacePanelHeight(m.height, m.status != "")
	left := m.styles.renderPanel("Bindings", m.renderProjectBindings(panelContentWidth(leftWidth)), m.projectLeftPanelActive(), leftWidth, panelHeight)
	right := m.styles.renderPanel(m.projectRightPanelTitle(), m.renderProjectRight(panelContentWidth(rightWidth)), m.projectRightPanelActive(), rightWidth, panelHeight)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	footer := m.styles.renderFooter("a add  d remove  i initialize  r reload  ? help  q quit", m.width)
	if m.status != "" {
		footer = m.styles.renderStatus(m.status, m.width) + "\n" + footer
	}
	return header + "\n" + body + "\n" + footer + "\n"
}

func (m ProjectModel) renderProjectBindings(width int) string {
	var b strings.Builder
	if !m.workspace.ManifestPresent {
		b.WriteString("No .shelf.json in this worktree.\n\nPress i to initialize it.")
		return b.String()
	}
	if len(m.workspace.Bindings) == 0 {
		b.WriteString("No bindings.\n\nPress a to add a secret field.")
		return b.String()
	}
	for index, binding := range m.workspace.Bindings {
		cursor := " "
		if index == m.selected {
			cursor = "›"
		}
		required := "required"
		if !binding.Required {
			required = "optional"
		}
		health := binding.Status
		if binding.Health != "" {
			health += " · " + healthLabel(binding.Health)
		}
		row := fmt.Sprintf("%s %-18s %-16s %s %s", cursor, truncate(binding.Key, 18), truncate(binding.Env, 16), required, health)
		b.WriteString(truncate(row, width) + "\n")
	}
	return b.String()
}

func (m ProjectModel) renderProjectRight(width int) string {
	switch m.screen {
	case projectPicker, projectPickerSearch:
		return m.renderProjectPicker(width)
	case projectForm:
		return m.renderProjectForm(width)
	case projectConfirmInit:
		return "Initialize .shelf.json for this worktree?\n\nThe new manifest contains no secret values.\n\nenter confirm   esc cancel"
	case projectConfirmRemove:
		return m.renderProjectRemoveConfirm(width)
	}
	return m.renderProjectDetail(width)
}

func (m ProjectModel) projectLeftPanelActive() bool {
	return !m.projectRightPanelActive()
}

func (m ProjectModel) projectRightPanelActive() bool {
	switch m.screen {
	case projectPicker, projectPickerSearch, projectForm, projectConfirmInit, projectConfirmRemove:
		return true
	default:
		return false
	}
}

func (m ProjectModel) projectRightPanelTitle() string {
	switch m.screen {
	case projectPicker, projectPickerSearch, projectForm:
		return "Add binding"
	case projectConfirmInit:
		return "Initialize manifest"
	case projectConfirmRemove:
		return "Remove binding"
	default:
		return "Detail"
	}
}

func (m ProjectModel) renderProjectDetail(_ int) string {
	if !m.workspace.ManifestPresent {
		return "Project manifest\n\nNo .shelf.json exists for this Git worktree.\n\nPress i to create an empty, value-free manifest."
	}
	binding, ok := m.currentProjectBinding()
	if !ok {
		return "Project manifest\n\nNo bindings yet.\n\nPress a to choose a secret field."
	}
	title := binding.Title
	if title == "" {
		title = binding.Item
	}
	required := "yes"
	if !binding.Required {
		required = "no (optional)"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", title)
	fmt.Fprintf(&b, "item:     %s\n", binding.Item)
	fmt.Fprintf(&b, "field:    %s\n", binding.Field)
	fmt.Fprintf(&b, "env:      %s\n", binding.Env)
	fmt.Fprintf(&b, "required: %s\n", required)
	fmt.Fprintf(&b, "status:   %s\n", binding.Status)
	if binding.Health != "" {
		fmt.Fprintf(&b, "health:   %s", healthLabel(binding.Health))
		if binding.Expires != "" {
			fmt.Fprintf(&b, " · expires %s", binding.Expires)
		}
		b.WriteByte('\n')
	}
	if binding.Message != "" {
		fmt.Fprintf(&b, "\n%s\n", binding.Message)
	}
	return b.String()
}

func (m ProjectModel) renderProjectPicker(width int) string {
	var b strings.Builder
	if m.loadingCandidates {
		b.WriteString("Loading secret fields…")
		return b.String()
	}
	if m.screen == projectPickerSearch {
		b.WriteString(m.candidateQuery.View() + "\n\n")
	}
	candidates := m.filteredCandidates()
	if len(candidates) == 0 {
		b.WriteString("No secret fields match.\n\n/ filter   esc cancel")
		return b.String()
	}
	for index, candidate := range candidates {
		cursor := " "
		if index == m.candidateSelected {
			cursor = "›"
		}
		env := candidate.Env
		if env == "" {
			env = "env required"
		}
		row := fmt.Sprintf("%s %-16s %-14s %s", cursor, truncate(candidate.Title, 16), truncate(candidate.Field, 14), env)
		b.WriteString(truncate(row, width) + "\n")
	}
	b.WriteString("\n/ filter   enter choose   esc cancel")
	return b.String()
}

func (m ProjectModel) renderProjectRemoveConfirm(_ int) string {
	binding, ok := m.currentProjectBinding()
	if !ok {
		return "No binding selected."
	}
	required := "required"
	if !binding.Required {
		required = "optional"
	}
	return fmt.Sprintf("Remove %s?\n\nenv:      %s\nrequired: %s\n\nThis changes only the value-free project manifest.\n\nenter confirm   esc cancel", binding.Key, binding.Env, required)
}

func (m ProjectModel) renderProjectHelp() string {
	return "Shelf project TUI help\n\n" +
		"j / k          move binding or candidate selection\n" +
		"a              choose a secret field and add a binding\n" +
		"d              remove selected binding after confirmation\n" +
		"i              initialize an empty manifest when missing\n" +
		"/              filter the candidate picker\n" +
		"ctrl+v         paste into the active text input\n" +
		"r              reload\n" +
		"esc            cancel\n" +
		"q / ctrl+c     quit (q is text while editing/searching)\n\n" +
		"? / esc return\n"
}
