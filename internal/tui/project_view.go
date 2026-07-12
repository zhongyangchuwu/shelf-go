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
	header := m.styles.header.Render(fmt.Sprintf("Shelf  ·  Project bindings    %s  ·  %d bindings", projectID, len(m.workspace.Bindings)))
	leftWidth := max(34, m.width*36/100)
	rightWidth := max(50, m.width-leftWidth-3)
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.renderProjectBindings(leftWidth), "   ", m.renderProjectRight(rightWidth))
	footer := "a add  d remove  i initialize  r reload  ? help  q quit"
	if m.status != "" {
		footer = m.status + "\n" + footer
	}
	return header + "\n" + strings.Repeat("─", max(1, m.width)) + "\n" + body + "\n" + strings.Repeat("─", max(1, m.width)) + "\n" + footer + "\n"
}

func (m ProjectModel) renderProjectBindings(width int) string {
	var b strings.Builder
	b.WriteString(m.styles.focus.Render("> Bindings") + "\n\n")
	if !m.workspace.ManifestPresent {
		b.WriteString("No .shelf.json in this worktree.\n\nPress i to initialize it.")
		return lipgloss.NewStyle().Width(width).Render(b.String())
	}
	if len(m.workspace.Bindings) == 0 {
		b.WriteString("No bindings.\n\nPress a to add a secret field.")
		return lipgloss.NewStyle().Width(width).Render(b.String())
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
		row := fmt.Sprintf("%s %-18s %-16s %s %s", cursor, truncate(binding.Key, 18), truncate(binding.Env, 16), required, binding.Status)
		b.WriteString(truncate(row, width) + "\n")
	}
	return lipgloss.NewStyle().Width(width).Render(b.String())
}

func (m ProjectModel) renderProjectRight(width int) string {
	switch m.screen {
	case projectPicker, projectPickerSearch:
		return m.renderProjectPicker(width)
	case projectForm:
		return m.renderProjectForm(width)
	case projectConfirmInit:
		return lipgloss.NewStyle().Width(width).Render("Initialize .shelf.json for this worktree?\n\nThe new manifest contains no secret values.\n\nenter confirm   esc cancel")
	case projectConfirmRemove:
		return m.renderProjectRemoveConfirm(width)
	}
	return m.renderProjectDetail(width)
}

func (m ProjectModel) renderProjectDetail(width int) string {
	if !m.workspace.ManifestPresent {
		return lipgloss.NewStyle().Width(width).Render("Project manifest\n\nNo .shelf.json exists for this Git worktree.\n\nPress i to create an empty, value-free manifest.")
	}
	binding, ok := m.currentProjectBinding()
	if !ok {
		return lipgloss.NewStyle().Width(width).Render("Project manifest\n\nNo bindings yet.\n\nPress a to choose a secret field.")
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
	b.WriteString(m.styles.pane.Render("Detail") + "\n\n")
	fmt.Fprintf(&b, "%s\n", title)
	fmt.Fprintf(&b, "item:     %s\n", binding.Item)
	fmt.Fprintf(&b, "field:    %s\n", binding.Field)
	fmt.Fprintf(&b, "env:      %s\n", binding.Env)
	fmt.Fprintf(&b, "required: %s\n", required)
	fmt.Fprintf(&b, "status:   %s\n", binding.Status)
	if binding.Message != "" {
		fmt.Fprintf(&b, "\n%s\n", binding.Message)
	}
	return lipgloss.NewStyle().Width(width).Render(b.String())
}

func (m ProjectModel) renderProjectPicker(width int) string {
	var b strings.Builder
	b.WriteString(m.styles.focus.Render("> Add binding") + "\n\n")
	if m.loadingCandidates {
		b.WriteString("Loading secret fields…")
		return lipgloss.NewStyle().Width(width).Render(b.String())
	}
	if m.screen == projectPickerSearch {
		b.WriteString(m.candidateQuery.View() + "\n\n")
	}
	candidates := m.filteredCandidates()
	if len(candidates) == 0 {
		b.WriteString("No secret fields match.\n\n/ filter   esc cancel")
		return lipgloss.NewStyle().Width(width).Render(b.String())
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
	return lipgloss.NewStyle().Width(width).Render(b.String())
}

func (m ProjectModel) renderProjectRemoveConfirm(width int) string {
	binding, ok := m.currentProjectBinding()
	if !ok {
		return "No binding selected."
	}
	required := "required"
	if !binding.Required {
		required = "optional"
	}
	content := fmt.Sprintf("Remove %s?\n\nenv:      %s\nrequired: %s\n\nThis changes only the value-free project manifest.\n\nenter confirm   esc cancel", binding.Key, binding.Env, required)
	return lipgloss.NewStyle().Width(width).Render(content)
}

func (m ProjectModel) renderProjectHelp() string {
	return "Shelf project TUI help\n\n" +
		"j / k          move binding or candidate selection\n" +
		"a              choose a secret field and add a binding\n" +
		"d              remove selected binding after confirmation\n" +
		"i              initialize an empty manifest when missing\n" +
		"/              filter the candidate picker\n" +
		"r              reload\n" +
		"esc            cancel\n" +
		"q / ctrl+c     quit\n\n" +
		"? / esc return\n"
}
