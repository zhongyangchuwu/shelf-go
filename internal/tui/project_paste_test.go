package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func editingProjectBindingModel(t *testing.T) (ProjectModel, *fakeProjectService) {
	t.Helper()
	m, service := readyProjectModel(t)
	updated, _ := m.startProjectBindingForm(service.candidates[0])
	m = updated.(ProjectModel)
	// Focus directly instead of executing the form's terminal-size and blink commands.
	m.form.form.GetFocusedField().Focus()
	return m, service
}

func searchingProjectPickerModel(t *testing.T) (ProjectModel, *fakeProjectService) {
	t.Helper()
	m, service := readyProjectModel(t)
	m = updateProjectModel(t, m, tea.KeyPressMsg{Text: "a"})
	m = updateProjectModel(t, m, projectCandidatesLoadedMsg{candidates: service.candidates})
	m = updateProjectModel(t, m, tea.KeyPressMsg{Text: "/"})
	return m, service
}

func TestProjectInputsAcceptTerminalAndAsyncClipboardPaste(t *testing.T) {
	for _, search := range []bool{false, true} {
		for _, clipboardPaste := range []bool{false, true} {
			t.Run(fmt.Sprintf("search=%t/clipboard=%t", search, clipboardPaste), func(t *testing.T) {
				m, service := editingProjectBindingModel(t)
				initial := service.candidates[0].Env
				if search {
					m, service = searchingProjectPickerModel(t)
					initial = ""
				} else {
					m = updateProjectModel(t, m, tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
				}
				reader := &fakeClipboardReader{text: "qjcdr/?用户\n\x03tail"}
				m.clipboardReader = reader
				if clipboardPaste {
					updated, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
					m = updated.(ProjectModel)
					if cmd == nil || reader.reads != 0 {
						t.Fatal("project clipboard read must be asynchronous")
					}
					m = updateProjectModel(t, m, cmd())
					if reader.reads != 1 {
						t.Fatalf("clipboard reads = %d", reader.reads)
					}
				} else {
					m = updateProjectModel(t, m, tea.PasteStartMsg{})
					m = updateProjectModel(t, m, tea.PasteMsg{Content: reader.text})
					m = updateProjectModel(t, m, tea.PasteEndMsg{})
					if reader.reads != 0 {
						t.Fatal("project terminal paste read the system clipboard")
					}
				}
				want := "qjcdr/?用户 tail" + initial
				if search {
					if m.screen != projectPickerSearch || m.candidateQuery.Value() != want {
						t.Fatalf("picker search paste = screen %v, value %q", m.screen, m.candidateQuery.Value())
					}
				} else if m.screen != projectForm || m.form.draft.env != want {
					t.Fatal("project form did not receive pasted text in the active input")
				}
				if len(service.added)+len(service.removed)+service.initialized != 0 {
					t.Fatal("project paste executed a shortcut or submitted a form")
				}
			})
		}
	}
}

func TestProjectQIsTextWhileEditingOrSearching(t *testing.T) {
	m, service := editingProjectBindingModel(t)
	m = updateProjectModel(t, m, tea.KeyPressMsg{Text: "q"})
	if m.screen != projectForm || !strings.Contains(m.form.draft.env, "q") || len(service.added) != 0 {
		t.Fatal("q was not text in the project form")
	}
	m, _ = searchingProjectPickerModel(t)
	m = updateProjectModel(t, m, tea.KeyPressMsg{Text: "q"})
	if m.screen != projectPickerSearch || m.candidateQuery.Value() != "q" {
		t.Fatal("q was not text in project picker search")
	}
}

func TestProjectPasteFailurePreservesDraftAndSearch(t *testing.T) {
	for _, search := range []bool{false, true} {
		t.Run(fmt.Sprintf("search=%t", search), func(t *testing.T) {
			m, _ := editingProjectBindingModel(t)
			initial := m.form.draft.env
			if search {
				m, _ = searchingProjectPickerModel(t)
				m.candidateQuery.SetValue("unsaved query")
			}
			m.clipboardReader = &fakeClipboardReader{text: projectFakeSecret, err: fmt.Errorf("clipboard failed: %s", projectFakeSecret)}
			updated, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
			m = updateProjectModel(t, updated.(ProjectModel), cmd())
			if search {
				if m.screen != projectPickerSearch || m.candidateQuery.Value() != "unsaved query" {
					t.Fatal("clipboard failure discarded project search")
				}
			} else if m.screen != projectForm || m.form.draft.env != initial {
				t.Fatal("clipboard failure discarded the project form")
			}
			if m.status == "" || strings.Contains(m.renderProject(), projectFakeSecret) {
				t.Fatal("project clipboard failure did not report a value-free status")
			}
		})
	}
}

func TestProjectAsyncPasteDoesNotFollowCancelledForm(t *testing.T) {
	m, service := editingProjectBindingModel(t)
	m.clipboardReader = &fakeClipboardReader{text: projectFakeSecret}
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	m = updated.(ProjectModel)
	m = updateProjectModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	updated, _ = m.startProjectBindingForm(service.candidates[1])
	m = updated.(ProjectModel)
	m.form.form.GetFocusedField().Focus()
	m = updateProjectModel(t, m, cmd())
	if m.form.draft.env != service.candidates[1].Env || strings.Contains(m.renderProject(), projectFakeSecret) {
		t.Fatal("late project paste followed focus into a replacement form")
	}
}

func TestProjectCtrlCQuitsInsteadOfEditing(t *testing.T) {
	m, service := editingProjectBindingModel(t)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("ctrl+c did not return a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok || len(service.added) != 0 {
		t.Fatal("ctrl+c in a project form did not quit")
	}
}

func TestProjectTerminalPasteCannotExecuteWorkspaceShortcuts(t *testing.T) {
	m, service := readyProjectModel(t)
	m = updateProjectModel(t, m, tea.PasteMsg{Content: "qiad\x03\n"})
	if m.screen != projectWorkspace || len(service.added)+len(service.removed)+service.initialized != 0 {
		t.Fatal("terminal paste executed a project workspace shortcut")
	}
}
