package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

const projectFakeSecret = "project-fake-secret"

type fakeProjectService struct {
	workspace  app.ProjectWorkspace
	candidates []app.ProjectBindingCandidate
	err        error

	added       []app.ProjectAddRequest
	removed     []string
	initialized int
}

func (s *fakeProjectService) ProjectWorkspace(_, _ string) (app.ProjectWorkspace, error) {
	return s.workspace, s.err
}

func (s *fakeProjectService) ProjectBindingCandidates(_, _ string) ([]app.ProjectBindingCandidate, error) {
	return s.candidates, s.err
}

func (s *fakeProjectService) ProjectAdd(_, _ string, req app.ProjectAddRequest) (string, error) {
	s.added = append(s.added, req)
	return "added", s.err
}

func (s *fakeProjectService) InitializeProject(bool) (string, error) {
	s.initialized++
	return "initialized", s.err
}

func (s *fakeProjectService) RemoveProjectBinding(key string) (string, error) {
	s.removed = append(s.removed, key)
	return "removed", s.err
}

func readyProjectModel(t *testing.T) (ProjectModel, *fakeProjectService) {
	t.Helper()
	service := &fakeProjectService{
		workspace: app.ProjectWorkspace{
			ProjectID:       "demo-project",
			Root:            "/safe/project",
			ManifestPresent: true,
			Bindings: []app.ProjectBindingStatus{
				{Key: "openai.credential", Item: "openai", Title: "OpenAI personal", Field: "credential", Env: "OPENAI_API_KEY", Required: true, Status: "ready"},
				{Key: "github.credential", Item: "github", Title: "GitHub work", Field: "credential", Env: "GITHUB_TOKEN", Required: false, Status: "warn", Message: "missing value optional"},
			},
		},
		candidates: []app.ProjectBindingCandidate{
			{Item: "github", Title: "GitHub work", Field: "credential", Env: "GITHUB_TOKEN"},
			{Item: "openai", Title: "OpenAI personal", Field: "credential", Env: "OPENAI_API_KEY"},
		},
	}
	m := NewProjectModel(service, Options{})
	m = updateProjectModel(t, m, tea.WindowSizeMsg{Width: 120, Height: 32})
	m = updateProjectModel(t, m, projectWorkspaceLoadedMsg{workspace: service.workspace})
	return m, service
}

func updateProjectModel(t *testing.T, m ProjectModel, msg tea.Msg) ProjectModel {
	t.Helper()
	updated, _ := m.Update(msg)
	result, ok := updated.(ProjectModel)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.ProjectModel", updated)
	}
	return result
}

func TestProjectModelPickerFiltersOnlySafeCandidateMetadata(t *testing.T) {
	m, service := readyProjectModel(t)
	updated, _ := m.Update(tea.KeyPressMsg{Text: "a"})
	m = updated.(ProjectModel)
	if m.screen != projectPicker || !m.loadingCandidates {
		t.Fatalf("picker state = screen %v loading %t", m.screen, m.loadingCandidates)
	}
	m = updateProjectModel(t, m, projectCandidatesLoadedMsg{candidates: service.candidates})
	m.candidateQuery.SetValue("github")
	if got := m.filteredCandidates(); len(got) != 1 || got[0].Item != "github" {
		t.Fatalf("filtered candidates = %#v", got)
	}
	m.candidateQuery.SetValue(projectFakeSecret)
	if got := m.filteredCandidates(); len(got) != 0 {
		t.Fatalf("secret query candidates = %#v", got)
	}
	if view := m.renderProject(); strings.Contains(view, projectFakeSecret) {
		t.Fatalf("picker view leaked secret: %q", view)
	}
}

func TestProjectModelAddAndRemoveUseAppWorkflows(t *testing.T) {
	m, service := readyProjectModel(t)
	candidate := service.candidates[0]
	updated, _ := m.startProjectBindingForm(candidate)
	m = updated.(ProjectModel)
	m.form.draft.env = "OVERRIDE_TOKEN"
	m.form.draft.optional = true
	updated, cmd := m.submitProjectBindingForm()
	m = updated.(ProjectModel)
	result := cmd().(projectActionDoneMsg)
	if result.err != nil || result.status != "Added project binding." || m.screen != projectLoading {
		t.Fatalf("add result = %#v, screen = %v", result, m.screen)
	}
	if got := service.added; len(got) != 1 || got[0].Selector != "github" || got[0].Field != "credential" || got[0].Env != "OVERRIDE_TOKEN" || !got[0].Optional {
		t.Fatalf("add requests = %#v", got)
	}

	m, service = readyProjectModel(t)
	m.screen = projectConfirmRemove
	updated, cmd = m.handleProjectRemoveConfirm("enter")
	m = updated.(ProjectModel)
	result = cmd().(projectActionDoneMsg)
	if result.err != nil || result.status != "Removed project binding." || m.screen != projectLoading {
		t.Fatalf("remove result = %#v, screen = %v", result, m.screen)
	}
	if got := service.removed; len(got) != 1 || got[0] != "openai.credential" {
		t.Fatalf("removed keys = %#v", got)
	}
}

func TestProjectModelInitializesMissingManifestWithConfirmation(t *testing.T) {
	m, service := readyProjectModel(t)
	m.workspace.ManifestPresent = false
	m.workspace.Bindings = nil
	m = updateProjectModel(t, m, tea.KeyPressMsg{Text: "i"})
	if m.screen != projectConfirmInit {
		t.Fatalf("init state = %v, want confirmation", m.screen)
	}
	if view := m.renderProject(); strings.Contains(view, projectFakeSecret) || !strings.Contains(view, "contains no secret values") {
		t.Fatalf("init confirmation = %q", view)
	}
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(ProjectModel)
	result := cmd().(projectActionDoneMsg)
	if result.err != nil || service.initialized != 1 || m.screen != projectLoading {
		t.Fatalf("init result = %#v calls=%d screen=%v", result, service.initialized, m.screen)
	}
}

func TestProjectModelSafeErrorAndResizeViews(t *testing.T) {
	m, service := readyProjectModel(t)
	service.err = fmt.Errorf("failed with %s", projectFakeSecret)
	m = updateProjectModel(t, m, projectActionDoneMsg{err: service.err})
	if view := m.renderProject(); strings.Contains(view, projectFakeSecret) || !strings.Contains(view, "No secret value was displayed.") {
		t.Fatalf("error view = %q", view)
	}
	m = updateProjectModel(t, m, tea.WindowSizeMsg{Width: 80, Height: 20})
	if m.screen != projectResize || !strings.Contains(m.renderProject(), "100 columns") {
		t.Fatalf("resize state = %v view=%q", m.screen, m.renderProject())
	}
}
