package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

type projectScreen uint8

const (
	projectLoading projectScreen = iota
	projectWorkspace
	projectPicker
	projectPickerSearch
	projectForm
	projectConfirmInit
	projectConfirmRemove
	projectError
	projectResize
	projectHelp
)

type projectWorkspaceLoadedMsg struct {
	workspace app.ProjectWorkspace
	err       error
}

type projectCandidatesLoadedMsg struct {
	candidates []app.ProjectBindingCandidate
	err        error
}

type projectActionDoneMsg struct {
	status string
	err    error
}

type ProjectModel struct {
	service ProjectService
	options Options

	screen projectScreen
	width  int
	height int
	styles workspaceStyles

	workspace app.ProjectWorkspace
	selected  int

	candidates        []app.ProjectBindingCandidate
	candidateSelected int
	candidateQuery    textinput.Model
	loadingCandidates bool

	status string
	form   projectFormState
}

func NewProjectModel(service ProjectService, options Options) ProjectModel {
	query := textinput.New()
	query.Prompt = "/ "
	query.Placeholder = "Filter item, field, or env name"
	return ProjectModel{
		service:        service,
		options:        options,
		screen:         projectLoading,
		styles:         newWorkspaceStyles(),
		candidateQuery: query,
	}
}

func (m ProjectModel) Init() tea.Cmd {
	return m.loadWorkspace()
}

func (m ProjectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.projectTooSmall() {
			m.screen = projectResize
		} else if m.screen == projectResize {
			m.screen = projectWorkspace
		}
		return m, nil
	case projectWorkspaceLoadedMsg:
		if msg.err != nil {
			m.screen = projectError
			return m, nil
		}
		m.workspace = msg.workspace
		if m.selected >= len(m.workspace.Bindings) {
			m.selected = max(0, len(m.workspace.Bindings)-1)
		}
		m.screen = m.projectScreenAfterLoad()
		return m, nil
	case projectCandidatesLoadedMsg:
		if msg.err != nil {
			m.screen = projectError
			return m, nil
		}
		m.candidates = msg.candidates
		m.candidateSelected = 0
		m.loadingCandidates = false
		m.screen = projectPicker
		return m, nil
	case projectActionDoneMsg:
		if msg.err != nil {
			m.screen = projectError
			return m, nil
		}
		m.status = msg.status
		m.form = projectFormState{}
		m.screen = projectLoading
		return m, m.loadWorkspace()
	case tea.KeyPressMsg:
		return m.handleProjectKey(msg)
	}

	if m.screen == projectPickerSearch {
		var cmd tea.Cmd
		m.candidateQuery, cmd = m.candidateQuery.Update(msg)
		m.clampCandidateSelection()
		return m, cmd
	}
	if m.screen == projectForm {
		return m.updateProjectForm(msg)
	}
	return m, nil
}

func (m ProjectModel) handleProjectKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" || key == "q" {
		return m, tea.Quit
	}
	if m.screen == projectResize {
		if key == "r" {
			return m, m.loadWorkspace()
		}
		return m, nil
	}
	if m.screen == projectError {
		if key == "r" {
			m.screen = projectLoading
			return m, m.loadWorkspace()
		}
		return m, nil
	}
	if m.screen == projectHelp {
		if key == "?" || key == "esc" {
			m.screen = projectWorkspace
		}
		return m, nil
	}
	if m.screen == projectForm {
		return m.updateProjectForm(msg)
	}
	if m.screen == projectConfirmInit {
		if key == "esc" {
			m.screen = projectWorkspace
			return m, nil
		}
		if key == "enter" {
			m.screen = projectLoading
			return m, func() tea.Msg {
				_, err := m.service.InitializeProject(false)
				return projectActionDoneMsg{status: "Initialized project manifest.", err: err}
			}
		}
		return m, nil
	}
	if m.screen == projectConfirmRemove {
		return m.handleProjectRemoveConfirm(key)
	}
	if m.screen == projectPickerSearch {
		if key == "esc" {
			m.candidateQuery.Blur()
			m.screen = projectPicker
			return m, nil
		}
		if key == "enter" {
			m.candidateQuery.Blur()
			m.screen = projectPicker
			return m, nil
		}
		var cmd tea.Cmd
		m.candidateQuery, cmd = m.candidateQuery.Update(msg)
		m.clampCandidateSelection()
		return m, cmd
	}
	if m.screen == projectPicker {
		if m.loadingCandidates {
			return m, nil
		}
		switch key {
		case "esc":
			m.candidateQuery.SetValue("")
			m.candidateQuery.Blur()
			m.screen = projectWorkspace
			return m, nil
		case "/":
			m.candidateQuery.Focus()
			m.screen = projectPickerSearch
			return m, nil
		case "up", "k":
			m.moveCandidate(-1)
			return m, nil
		case "down", "j":
			m.moveCandidate(1)
			return m, nil
		case "enter":
			candidate, ok := m.selectedCandidate()
			if !ok {
				return m, nil
			}
			return m.startProjectBindingForm(candidate)
		}
		return m, nil
	}

	switch key {
	case "?":
		m.screen = projectHelp
		return m, nil
	case "r":
		m.screen = projectLoading
		return m, m.loadWorkspace()
	case "up", "k":
		m.moveProjectBinding(-1)
		return m, nil
	case "down", "j":
		m.moveProjectBinding(1)
		return m, nil
	case "i":
		if !m.workspace.ManifestPresent {
			m.screen = projectConfirmInit
		}
		return m, nil
	case "a":
		if !m.workspace.ManifestPresent {
			return m, nil
		}
		m.loadingCandidates = true
		m.screen = projectPicker
		return m, m.loadCandidates()
	case "d":
		if _, ok := m.currentProjectBinding(); ok {
			m.screen = projectConfirmRemove
		}
		return m, nil
	}
	return m, nil
}

func (m ProjectModel) View() tea.View {
	view := tea.NewView(m.renderProject())
	view.AltScreen = true
	return view
}

func (m ProjectModel) projectTooSmall() bool {
	return m.width > 0 && m.height > 0 && (m.width < minWidth || m.height < minHeight)
}

func (m ProjectModel) projectScreenAfterLoad() projectScreen {
	if m.projectTooSmall() {
		return projectResize
	}
	return projectWorkspace
}

func (m ProjectModel) loadWorkspace() tea.Cmd {
	return func() tea.Msg {
		workspace, err := m.service.ProjectWorkspace(m.options.ConfigPath, m.options.VaultPath)
		return projectWorkspaceLoadedMsg{workspace: workspace, err: err}
	}
}

func (m ProjectModel) loadCandidates() tea.Cmd {
	return func() tea.Msg {
		candidates, err := m.service.ProjectBindingCandidates(m.options.ConfigPath, m.options.VaultPath)
		return projectCandidatesLoadedMsg{candidates: candidates, err: err}
	}
}

func (m *ProjectModel) moveProjectBinding(delta int) {
	if len(m.workspace.Bindings) == 0 {
		return
	}
	m.selected = min(max(0, m.selected+delta), len(m.workspace.Bindings)-1)
}

func (m *ProjectModel) moveCandidate(delta int) {
	candidates := m.filteredCandidates()
	if len(candidates) == 0 {
		return
	}
	m.candidateSelected = min(max(0, m.candidateSelected+delta), len(candidates)-1)
}

func (m *ProjectModel) clampCandidateSelection() {
	candidates := m.filteredCandidates()
	m.candidateSelected = min(max(0, m.candidateSelected), max(0, len(candidates)-1))
}

func (m ProjectModel) currentProjectBinding() (app.ProjectBindingStatus, bool) {
	if m.selected < 0 || m.selected >= len(m.workspace.Bindings) {
		return app.ProjectBindingStatus{}, false
	}
	return m.workspace.Bindings[m.selected], true
}

func (m ProjectModel) filteredCandidates() []app.ProjectBindingCandidate {
	query := strings.ToLower(strings.TrimSpace(m.candidateQuery.Value()))
	if query == "" {
		return m.candidates
	}
	out := make([]app.ProjectBindingCandidate, 0, len(m.candidates))
	for _, candidate := range m.candidates {
		if strings.Contains(strings.ToLower(candidate.Item), query) ||
			strings.Contains(strings.ToLower(candidate.Title), query) ||
			strings.Contains(strings.ToLower(candidate.Field), query) ||
			strings.Contains(strings.ToLower(candidate.Env), query) {
			out = append(out, candidate)
		}
	}
	return out
}

func (m ProjectModel) selectedCandidate() (app.ProjectBindingCandidate, bool) {
	candidates := m.filteredCandidates()
	if m.candidateSelected < 0 || m.candidateSelected >= len(candidates) {
		return app.ProjectBindingCandidate{}, false
	}
	return candidates[m.candidateSelected], true
}
