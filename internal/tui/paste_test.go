package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

type fakeClipboardReader struct {
	text  string
	err   error
	reads int
}

func (r *fakeClipboardReader) ReadText() (string, error) {
	r.reads++
	return r.text, r.err
}

func editingSecretModel(t *testing.T) (Model, *fakeService, *fakeClipboard) {
	t.Helper()
	m, service, clipboard := readyModel(t)
	m = updateModel(t, m, tea.KeyPressMsg{Text: "e"})
	m.form.editor.focus(m.form.editor.indexOf("field:credential"))
	m = updateModel(t, m, tea.KeyPressMsg{Text: "a"})
	return m, service, clipboard
}

func TestWorkspaceInputsAcceptTerminalAndAsyncClipboardPaste(t *testing.T) {
	for _, search := range []bool{false, true} {
		for _, clipboardPaste := range []bool{false, true} {
			t.Run(fmt.Sprintf("search=%t/clipboard=%t", search, clipboardPaste), func(t *testing.T) {
				m, service, clipboard := editingSecretModel(t)
				if search {
					m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
					m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
					m = updateModel(t, m, tea.KeyPressMsg{Text: "/"})
					m.query.SetValue("beforeafter")
					m.query.SetCursor(6)
				} else {
					m.activeEditorInput().SetValue("beforeafter")
					m.activeEditorInput().SetCursor(6)
				}
				payload := "qjcdr/?用户\n\x13tail"
				reader := &fakeClipboardReader{text: payload}
				m.clipboardReader = reader
				if clipboardPaste {
					updated, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
					m = updated.(Model)
					if cmd == nil || reader.reads != 0 {
						t.Fatal("clipboard read must be an asynchronous command")
					}
					m = updateModel(t, m, cmd())
					if reader.reads != 1 {
						t.Fatalf("clipboard reads = %d", reader.reads)
					}
				} else {
					m = updateModel(t, m, tea.PasteStartMsg{})
					m = updateModel(t, m, tea.PasteMsg{Content: payload})
					m = updateModel(t, m, tea.PasteEndMsg{})
					if reader.reads != 0 {
						t.Fatal("terminal paste read the system clipboard")
					}
				}
				want := "beforeqjcdr/?用户 tailafter"
				if search {
					if m.screen != screenSearch || m.query.Value() != want {
						t.Fatalf("search paste = screen %v, value %q", m.screen, m.query.Value())
					}
				} else {
					if m.screen != screenForm || m.form.editor.value("field:credential") != want || m.form.editor.editMode != editInsert {
						t.Fatal("form paste did not stay in the focused insert-mode field")
					}
					if strings.Contains(m.render(), "qjcdr/?用户") {
						t.Fatal("secret paste was displayed in the form")
					}
				}
				if len(service.created)+len(service.updated)+len(service.removed)+len(service.retrieved) != 0 || clipboard.copied != "" {
					t.Fatal("pasted characters executed a workspace shortcut")
				}
			})
		}
	}
}

func TestWorkspaceQIsTextWhileEditingOrSearching(t *testing.T) {
	m, _, _ := editingSecretModel(t)
	m = updateModel(t, m, tea.KeyPressMsg{Text: "q"})
	if m.screen != screenForm || m.form.editor.value("field:credential") != "q" {
		t.Fatal("q did not enter the form as text")
	}
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	m = updateModel(t, m, tea.KeyPressMsg{Text: "y"})
	m = updateModel(t, m, tea.KeyPressMsg{Text: "/"})
	m = updateModel(t, m, tea.KeyPressMsg{Text: "q"})
	if m.screen != screenSearch || m.query.Value() != "q" {
		t.Fatal("q did not enter search as text")
	}
}

func TestWorkspacePasteFailurePreservesDraftAndSearch(t *testing.T) {
	for _, search := range []bool{false, true} {
		t.Run(fmt.Sprintf("search=%t", search), func(t *testing.T) {
			m, _, _ := editingSecretModel(t)
			m.activeEditorInput().SetValue(fakeSecret)
			if search {
				m.screen = screenSearch
				m.query.Focus()
				m.query.SetValue("unsaved query")
			}
			m.clipboardReader = &fakeClipboardReader{text: fakeSecret, err: fmt.Errorf("clipboard failed: %s", fakeSecret)}
			updated, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
			m = updateModel(t, updated.(Model), cmd())
			if m.form.editor.value("field:credential") != fakeSecret || (search && (m.screen != screenSearch || m.query.Value() != "unsaved query")) || (!search && m.screen != screenForm) {
				t.Fatal("clipboard failure discarded an unsaved input")
			}
			if m.status == "" || strings.Contains(m.render(), fakeSecret) {
				t.Fatal("clipboard failure did not report a value-free status")
			}
		})
	}
}

func TestWorkspaceAsyncPasteDoesNotFollowFocusToPublicInput(t *testing.T) {
	for _, search := range []bool{false, true} {
		t.Run(fmt.Sprintf("search=%t", search), func(t *testing.T) {
			m, _, _ := editingSecretModel(t)
			m.clipboardReader = &fakeClipboardReader{text: fakeSecret}
			updated, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
			m = updated.(Model)
			m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
			if search {
				m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
				m = updateModel(t, m, tea.KeyPressMsg{Text: "/"})
			} else {
				m.form.editor.focus(m.form.editor.indexOf("title"))
				m = updateModel(t, m, tea.KeyPressMsg{Text: "a"})
			}
			m = updateModel(t, m, cmd())
			if strings.Contains(m.render(), fakeSecret) || strings.Contains(m.query.Value(), fakeSecret) || (!search && strings.Contains(m.form.editor.value("title"), fakeSecret)) {
				t.Fatal("late secret paste followed focus to a public input")
			}
		})
	}
}

func TestWorkspaceCopyKeysUseSelectedStoredField(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{{Text: "c"}, {Code: 'c', Mod: tea.ModCtrl}} {
		t.Run(key.String(), func(t *testing.T) {
			m, service, clipboard := readyModel(t)
			m.fieldSelected = 1
			updated, cmd := m.Update(key)
			m = updated.(Model)
			if cmd == nil {
				t.Fatal("copy key did not return a command")
			}
			m = updateModel(t, m, cmd())
			if len(service.retrieved) != 1 || service.retrieved[0].Selector != "item_openai" || service.retrieved[0].Field != "provider" || clipboard.copied != fakeSecret {
				t.Fatal("copy did not send the selected stored field to the clipboard")
			}
			if m.screen != screenWorkspace || strings.Contains(m.render(), fakeSecret) {
				t.Fatal("copy quit or displayed the stored value")
			}
		})
	}
}

func TestWorkspaceCopyResultPreservesNewerDraftAndSearch(t *testing.T) {
	for _, search := range []bool{false, true} {
		for _, failure := range []string{"none", "read", "write"} {
			t.Run(fmt.Sprintf("search=%t/failure=%s", search, failure), func(t *testing.T) {
				m, service, clipboard := readyModel(t)
				updated, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
				m = updated.(Model)
				if failure == "read" {
					service.err = fmt.Errorf("read failed: %s", fakeSecret)
				} else if failure == "write" {
					clipboard.err = fmt.Errorf("copy failed: %s", fakeSecret)
				}
				if search {
					m = updateModel(t, m, tea.KeyPressMsg{Text: "/"})
					m.query.SetValue("unsaved query")
				} else {
					m = updateModel(t, m, tea.KeyPressMsg{Text: "e"})
					setEditorValue(t, m.form.editor, "field:credential", fakeSecret)
				}
				m = updateModel(t, m, cmd())
				if search {
					if m.screen != screenSearch || m.query.Value() != "unsaved query" {
						t.Fatal("copy result discarded the active search")
					}
				} else if m.screen != screenForm || m.form.editor == nil || m.form.editor.value("field:credential") != fakeSecret {
					t.Fatal("copy result discarded the active form")
				}
				if m.status == "" || strings.Contains(m.render(), fakeSecret) {
					t.Fatal("copy result did not report a value-free status")
				}
			})
		}
	}
}

func TestWorkspaceCtrlCQuitsOutsideWorkspace(t *testing.T) {
	for _, state := range []screen{screenSearch, screenForm, screenFilters, screenConfirmDelete, screenHelp, screenError, screenLoading} {
		m, service, _ := readyModel(t)
		m.screen = state
		_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		if cmd == nil {
			t.Fatalf("ctrl+c on screen %v did not quit", state)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok || len(service.retrieved) != 0 {
			t.Fatalf("ctrl+c on screen %v copied instead of quitting", state)
		}
	}
}

func TestWorkspaceTerminalPasteCannotExecuteWorkspaceShortcuts(t *testing.T) {
	m, service, clipboard := readyModel(t)
	m = updateModel(t, m, tea.PasteMsg{Content: "qcad/\x03\x13\n"})
	if m.screen != screenWorkspace || len(service.created)+len(service.updated)+len(service.removed)+len(service.retrieved) != 0 || clipboard.copied != "" {
		t.Fatal("terminal paste executed a workspace shortcut")
	}
}

func TestWorkspaceClipboardPasteRespectsInputLimitAndEmptyClipboard(t *testing.T) {
	for _, text := range []string{"", "用户123"} {
		t.Run(fmt.Sprintf("empty=%t", text == ""), func(t *testing.T) {
			m, _, _ := editingSecretModel(t)
			input := m.activeEditorInput()
			input.CharLimit = 4
			input.SetValue("ab")
			input.SetCursor(1)
			m.clipboardReader = &fakeClipboardReader{text: text}
			updated, cmd := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
			m = updateModel(t, updated.(Model), cmd())
			want := "a用户b"
			if text == "" {
				want = "ab"
			}
			if m.form.editor.value("field:credential") != want {
				t.Fatal("clipboard paste changed the input's character-limit or empty-paste behavior")
			}
		})
	}
}

func TestWorkspacePendingPasteSurvivesTypingInSameInput(t *testing.T) {
	for _, search := range []bool{false, true} {
		t.Run(fmt.Sprintf("search=%t", search), func(t *testing.T) {
			m, _, _ := editingSecretModel(t)
			if search {
				m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
				m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
				m = updateModel(t, m, tea.KeyPressMsg{Text: "/"})
			}
			m.clipboardReader = &fakeClipboardReader{text: "clipboard fragment"}
			updated, paste := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
			m = updateModel(t, updated.(Model), tea.KeyPressMsg{Text: "x"})
			m = updateModel(t, m, paste())
			value := m.query.Value()
			if !search {
				value = m.form.editor.value("field:credential")
				if strings.Contains(m.render(), "clipboard fragment") {
					t.Fatal("secret input paste was revealed")
				}
			}
			if value != "xclipboard fragment" {
				t.Fatal("typing in the same active input cancelled a pending clipboard paste")
			}
		})
	}
}
