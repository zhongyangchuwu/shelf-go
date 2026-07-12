package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

type projectFormState struct {
	form      *huh.Form
	candidate app.ProjectBindingCandidate
	draft     *projectBindingDraft
}

type projectBindingDraft struct {
	env      string
	optional bool
}

func (m ProjectModel) startProjectBindingForm(candidate app.ProjectBindingCandidate) (tea.Model, tea.Cmd) {
	draft := &projectBindingDraft{env: candidate.Env}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Environment variable name").Value(&draft.env).Validate(requireValue("environment variable name")),
			huh.NewConfirm().Title("Mark this binding optional").Value(&draft.optional),
		),
		huh.NewGroup(
			huh.NewNote().Title("Review").DescriptionFunc(func() string { return projectBindingReview(candidate, draft) }, draft).Next(true).NextLabel("Add binding"),
		),
	)
	m.form = projectFormState{form: form, candidate: candidate, draft: draft}
	m.screen = projectForm
	return m, form.Init()
}

func (m ProjectModel) updateProjectForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "esc" {
		m.form = projectFormState{}
		m.screen = projectPicker
		return m, nil
	}
	updated, cmd := m.form.form.Update(msg)
	form, ok := updated.(*huh.Form)
	if !ok {
		m.screen = projectError
		return m, nil
	}
	m.form.form = form
	if form.State != huh.StateCompleted {
		return m, cmd
	}
	return m.submitProjectBindingForm()
}

func (m ProjectModel) submitProjectBindingForm() (tea.Model, tea.Cmd) {
	form := m.form
	draft := *form.draft
	m.screen = projectLoading
	return m, func() tea.Msg {
		_, err := m.service.ProjectAdd(m.options.ConfigPath, m.options.VaultPath, app.ProjectAddRequest{
			Selector: form.candidate.Item,
			Field:    form.candidate.Field,
			Env:      draft.env,
			Optional: draft.optional,
		})
		return projectActionDoneMsg{status: "Added project binding.", err: err}
	}
}

func (m ProjectModel) handleProjectRemoveConfirm(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = projectWorkspace
		return m, nil
	case "enter":
		binding, ok := m.currentProjectBinding()
		if !ok {
			m.screen = projectWorkspace
			return m, nil
		}
		m.screen = projectLoading
		return m, func() tea.Msg {
			_, err := m.service.RemoveProjectBinding(binding.Key)
			return projectActionDoneMsg{status: "Removed project binding.", err: err}
		}
	default:
		return m, nil
	}
}

func (m ProjectModel) renderProjectForm(width int) string {
	if m.form.form == nil {
		return "Form unavailable."
	}
	return m.form.form.View()
}

func projectBindingReview(candidate app.ProjectBindingCandidate, draft *projectBindingDraft) string {
	required := "yes"
	if draft.optional {
		required = "no (optional)"
	}
	title := candidate.Title
	if title == "" {
		title = candidate.Item
	}
	return strings.Join([]string{
		"Item: " + title,
		"Item ID: " + candidate.Item,
		"Secret field: " + candidate.Field,
		"Environment variable: " + draft.env,
		"Required: " + required,
	}, "\n")
}
