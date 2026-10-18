package tui

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func TestSecretDraftSaveFailureAndRetry(t *testing.T) {
	for _, create := range []bool{false, true} {
		t.Run(fmt.Sprintf("create=%t", create), func(t *testing.T) {
			m, service, _ := readyModel(t)
			key, secretID, publicID := "e", "field:credential", "field:provider"
			if create {
				key, secretID, publicID = "a", "field:value", "field:env_value"
			}
			m = updateModel(t, m, tea.KeyPressMsg{Text: key})
			editor := m.form.editor
			setEditorValue(t, editor, "title", "  Retry title  ")
			setEditorValue(t, editor, "description", "Exact public description")
			setEditorValue(t, editor, "labels", "work, retry")
			setEditorValue(t, editor, publicID, "EXACT_PUBLIC_INPUT")
			secret := "  " + fakeSecret + " 用户  "
			setEditorValue(t, editor, secretID, secret)
			input := &editor.fields[editor.indexOf(secretID)].input
			service.err = fmt.Errorf("write failed: %s", secret)

			updated, cmd := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = updated.(Model)
			if cmd == nil || !m.form.saving || m.screen != screenForm || input.Value() != secret {
				t.Fatal("save did not retain the pending form")
			}
			failed := cmd().(formSavedMsg)
			m = updateModel(t, m, failed)
			if m.screen != screenForm || m.form.saving || m.form.editor != editor || editor.err == "" {
				t.Fatal("failed save did not return to the same retryable editor")
			}
			if input.Value() != secret || input.EchoMode != textinput.EchoPassword || editor.value("title") != "  Retry title  " || editor.value("description") != "Exact public description" || editor.value("labels") != "work, retry" || editor.value(publicID) != "EXACT_PUBLIC_INPUT" {
				t.Fatal("failed save changed a secret or public draft input")
			}
			if strings.Contains(m.render(), fakeSecret) || strings.Contains(editor.err, secret) {
				t.Fatal("failed save displayed a secret")
			}

			// Consume the failed attempt's input, as the real app does. Retry must
			// build a fresh reader rather than reusing an exhausted request.
			var first io.Reader
			if create {
				first = service.created[0].In
			} else {
				for _, field := range service.updated[0].FieldUpdates {
					if field.ReadField {
						first = field.In
					}
				}
			}
			if first == nil {
				t.Fatal("secret was not supplied through a reader")
			}
			if got, err := io.ReadAll(first); err != nil || string(got) != secret {
				t.Fatal("failed attempt did not retain the exact submitted secret")
			}

			service.err = nil
			updated, cmd = m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = updated.(Model)
			if cmd == nil || !m.form.saving || input.Value() != secret {
				t.Fatal("retry required re-entering the draft")
			}
			// A duplicate result from the failed attempt must not complete retry.
			failed.err = nil
			m = updateModel(t, m, failed)
			if !m.form.saving || m.form.editor != editor || input.Value() != secret {
				t.Fatal("stale save result completed a newer attempt")
			}
			m = updateModel(t, m, cmd())
			if m.screen != screenLoading || m.form.editor != nil || len(editor.fields) != 0 || len(editor.baseline) != 0 || input.Value() != "" {
				t.Fatal("successful retry did not clear editor-held inputs before reload")
			}
			var second io.Reader
			if create {
				if len(service.created) != 2 {
					t.Fatal("create retry did not make exactly two attempts")
				}
				second = service.created[1].In
			} else {
				if len(service.updated) != 2 {
					t.Fatal("edit retry did not make exactly two attempts")
				}
				for _, field := range service.updated[1].FieldUpdates {
					if field.ReadField {
						second = field.In
					}
				}
			}
			if second == nil {
				t.Fatal("retry did not provide a new secret reader")
			}
			if got, err := io.ReadAll(second); err != nil || string(got) != secret {
				t.Fatal("retry did not submit the exact original secret")
			}
		})
	}
}

func TestSecretDraftValidationRetainsInputs(t *testing.T) {
	for _, invalidDate := range []bool{false, true} {
		t.Run(fmt.Sprintf("invalidDate=%t", invalidDate), func(t *testing.T) {
			m, service, _ := readyModel(t)
			m = updateModel(t, m, tea.KeyPressMsg{Text: "a"})
			editor := m.form.editor
			secretField := "field:value"
			if invalidDate {
				editor.focus(editor.indexOf("kind"))
				editor.cycleChoice(1)
				secretField = "field:credential"
				setEditorValue(t, editor, "title", "Valid title")
				editor.focus(editor.indexOf("validity:validity_credential"))
				editor.cycleChoice(1)
				editor.cycleChoice(1)
				setEditorValue(t, editor, "validity-date:validity_credential", "not-a-date")
			}
			setEditorValue(t, editor, secretField, fakeSecret)
			updated, cmd := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = updated.(Model)
			if cmd != nil || m.form.saving || m.screen != screenForm || editor.err == "" || editor.value(secretField) != fakeSecret || len(service.created) != 0 {
				t.Fatal("validation failure lost the draft or started a write")
			}
			if strings.Contains(m.render(), fakeSecret) {
				t.Fatal("validation failure displayed a secret")
			}
			if invalidDate {
				setEditorValue(t, editor, "validity-date:validity_credential", "2027-12-31")
			} else {
				setEditorValue(t, editor, "title", "Corrected title")
			}
			updated, cmd = m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
			m = updated.(Model)
			if cmd == nil || !m.form.saving || editor.err != "" {
				t.Fatal("corrected validation could not retry the retained draft")
			}
			m = updateModel(t, m, cmd())
			if m.form.editor != nil || len(editor.fields) != 0 {
				t.Fatal("validated save did not clear its editor")
			}
		})
	}
}

func TestSecretDraftCancelAndQuit(t *testing.T) {
	for _, create := range []bool{false, true} {
		for _, state := range []string{"clean", "dirty", "reverted"} {
			for _, exit := range []tea.KeyPressMsg{{Code: tea.KeyEscape}, {Text: "q"}, {Code: 'c', Mod: tea.ModCtrl}} {
				t.Run(fmt.Sprintf("create=%t/%s/%s", create, state, exit.String()), func(t *testing.T) {
					m, service, _ := readyModel(t)
					key, secretID := "e", "field:credential"
					if create {
						key, secretID = "a", "field:value"
					}
					m = updateModel(t, m, tea.KeyPressMsg{Text: key})
					editor := m.form.editor
					input := &editor.fields[editor.indexOf(secretID)].input
					original := editor.value("title")
					if state != "clean" {
						setEditorValue(t, editor, "title", "Changed")
						setEditorValue(t, editor, secretID, fakeSecret)
					}
					if state == "reverted" {
						setEditorValue(t, editor, "title", original)
						setEditorValue(t, editor, secretID, "")
					}
					quit := exit.String() != "esc"
					updated, cmd := m.Update(exit)
					m = updated.(Model)
					if state == "dirty" {
						if cmd != nil || m.screen != screenConfirmDiscard || m.form.discardQuit != quit || input.Value() != fakeSecret {
							t.Fatal("dirty form exited without preserving the requested discard confirmation")
						}
						for _, keep := range []tea.KeyPressMsg{{Code: tea.KeyEnter}, {Text: "n"}, {Code: tea.KeyEscape}} {
							m = updateModel(t, m, keep)
							if keep.Code == tea.KeyEnter {
								if m.screen != screenConfirmDiscard {
									t.Fatal("Enter accidentally discarded or accepted the confirmation")
								}
							} else {
								if m.screen != screenForm || m.form.editor != editor || input.Value() != fakeSecret {
									t.Fatal("declining discard did not retain the editor")
								}
								m = updateModel(t, m, exit)
							}
						}
						updated, cmd = m.Update(tea.KeyPressMsg{Text: "y"})
						m = updated.(Model)
					}
					if m.form.editor != nil || len(editor.fields) != 0 || len(editor.baseline) != 0 || input.Value() != "" || m.screen != screenWorkspace {
						t.Fatal("discard or clean exit retained editor inputs")
					}
					if quit {
						if cmd == nil {
							t.Fatal("confirmed or clean quit did not quit")
						}
						if _, ok := cmd().(tea.QuitMsg); !ok {
							t.Fatal("quit returned a non-quit command")
						}
					} else if cmd != nil {
						t.Fatal("cancel unexpectedly quit or started work")
					}
					if len(service.created)+len(service.updated) != 0 {
						t.Fatal("cancel wrote a draft")
					}
				})
			}
		}
	}
}

func TestSecretDraftInsertEscapeAndQ(t *testing.T) {
	m, _, _ := editingSecretModel(t)
	m = updateModel(t, m, tea.KeyPressMsg{Text: "q"})
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.screen != screenForm || m.form.editor.editMode != editNormal || m.form.editor.value("field:credential") != "q" {
		t.Fatal("insert-mode q or Escape discarded the draft")
	}
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.screen != screenConfirmDiscard || m.form.discardQuit {
		t.Fatal("normal-mode Escape did not confirm returning to the workspace")
	}
}

func TestSecretDraftUnchangedEditDoesNotWrite(t *testing.T) {
	m, service, _ := readyModel(t)
	m = updateModel(t, m, tea.KeyPressMsg{Text: "e"})
	editor := m.form.editor
	original := editor.value("description")
	setEditorValue(t, editor, "description", "Temporary")
	setEditorValue(t, editor, "description", original)
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = updated.(Model)
	if cmd != nil || m.screen != screenWorkspace || m.form.editor != nil || len(editor.fields) != 0 || len(service.updated) != 0 {
		t.Fatal("unchanged edit wrote or retained the editor")
	}
}

func TestSecretDraftRevertedChoicesAreClean(t *testing.T) {
	m, _, _ := readyModel(t)
	m = updateModel(t, m, tea.KeyPressMsg{Text: "a"})
	editor := m.form.editor
	editor.focus(editor.indexOf("kind"))
	editor.cycleChoice(1)
	editor.focus(editor.indexOf("validity:validity_credential"))
	editor.cycleChoice(1)
	editor.cycleChoice(1)
	setEditorValue(t, editor, "validity-date:validity_credential", "2027-12-31")
	editor.cycleChoice(1)
	editor.focus(editor.indexOf("kind"))
	editor.cycleChoice(-1)
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.screen != screenWorkspace || m.form.editor != nil {
		t.Fatal("reverted kind/validity choices or an inactive date made an unchanged form dirty")
	}
}

func TestSecretDraftSaveIsSingleFlightAndRejectsLateMessages(t *testing.T) {
	m, service, _ := editingSecretModel(t)
	editor := m.form.editor
	setEditorValue(t, editor, "field:credential", fakeSecret)
	m.clipboardReader = &fakeClipboardReader{text: "late secret"}
	updated, paste := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	m = updated.(Model)
	updated, save := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = updated.(Model)
	if save == nil || !m.form.saving {
		t.Fatal("save was not started")
	}
	for _, key := range []tea.KeyPressMsg{{Text: "q"}, {Code: 'c', Mod: tea.ModCtrl}, {Code: tea.KeyEscape}, {Code: 's', Mod: tea.ModCtrl}, {Code: 'v', Mod: tea.ModCtrl}, {Text: "x"}, {Text: "r"}} {
		updated, cmd := m.Update(key)
		m = updated.(Model)
		if cmd != nil || !m.form.saving || m.screen != screenForm || editor.value("field:credential") != fakeSecret {
			t.Fatal("pending save accepted a key, discarded input, or started duplicate work")
		}
	}
	for _, msg := range []tea.Msg{
		tea.PasteMsg{Content: "terminal paste"}, paste(),
		catalogLoadedMsg{err: fmt.Errorf("late catalog failure")},
		detailLoadedMsg{id: m.currentID(), err: fmt.Errorf("late detail failure")},
		actionDoneMsg{refresh: true},
		formSavedMsg{editor: editor, saveID: m.form.saveID + 1},
	} {
		m = updateModel(t, m, msg)
		if !m.form.saving || m.screen != screenForm || m.form.editor != editor || editor.value("field:credential") != fakeSecret {
			t.Fatal("late message changed a pending draft")
		}
	}
	result := save().(formSavedMsg)
	m = updateModel(t, m, result)
	if len(service.updated) != 1 || m.form.editor != nil {
		t.Fatal("single-flight save did not complete exactly one write")
	}
	m = updateModel(t, m, catalogLoadedMsg{items: service.catalog})
	m = updateModel(t, m, detailLoadedMsg{id: "item_openai", info: service.info["item_openai"]})
	m = updateModel(t, m, tea.KeyPressMsg{Text: "e"})
	newEditor := m.form.editor
	setEditorValue(t, newEditor, "field:credential", "new secret draft")
	m = updateModel(t, m, result)
	if m.screen != screenForm || m.form.editor != newEditor || newEditor.value("field:credential") != "new secret draft" {
		t.Fatal("old save result cleared a newer form")
	}
}

func TestSecretDraftResizePreservesLifecycle(t *testing.T) {
	for _, state := range []string{"editing", "confirm", "saving", "failed"} {
		t.Run(state, func(t *testing.T) {
			m, service, _ := readyModel(t)
			m = updateModel(t, m, tea.KeyPressMsg{Text: "e"})
			editor := m.form.editor
			setEditorValue(t, editor, "field:credential", fakeSecret)
			var save tea.Cmd
			if state == "confirm" {
				m = updateModel(t, m, tea.KeyPressMsg{Text: "q"})
			} else if state == "saving" || state == "failed" {
				updated, cmd := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
				m, save = updated.(Model), cmd
			}
			expectedScreen := m.screen
			m = updateModel(t, m, tea.WindowSizeMsg{Width: 60, Height: 15})
			m = updateModel(t, m, tea.PasteMsg{Content: "must not change input"})
			if state == "failed" {
				service.err = fmt.Errorf("permission denied: %s", fakeSecret)
				m = updateModel(t, m, save())
			}
			if m.screen != expectedScreen || m.form.editor != editor || editor.value("field:credential") != fakeSecret || strings.Contains(m.render(), fakeSecret) {
				t.Fatal("small terminal lost the draft or lifecycle state")
			}
			m = updateModel(t, m, tea.WindowSizeMsg{Width: 120, Height: 32})
			if m.screen != expectedScreen || m.form.editor != editor || editor.value("field:credential") != fakeSecret || m.form.discardQuit != (state == "confirm") || m.form.saving != (state == "saving") {
				t.Fatal("resize did not restore the same form/confirmation/save")
			}
			if state == "saving" {
				m = updateModel(t, m, save())
				if m.form.editor != nil || len(editor.fields) != 0 {
					t.Fatal("save could not finish after resize")
				}
			}
		})
	}
}

func TestSecretDraftSaveSuccessWhileSmallDoesNotReopenForm(t *testing.T) {
	m, service, _ := readyModel(t)
	m = updateModel(t, m, tea.KeyPressMsg{Text: "e"})
	editor := m.form.editor
	setEditorValue(t, editor, "field:credential", fakeSecret)
	updated, save := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	m = updateModel(t, updated.(Model), tea.WindowSizeMsg{Width: 60, Height: 15})
	m = updateModel(t, m, save())
	m = updateModel(t, m, catalogLoadedMsg{items: service.catalog})
	m = updateModel(t, m, tea.WindowSizeMsg{Width: 120, Height: 32})
	if m.screen != screenWorkspace || m.form.editor != nil || len(editor.fields) != 0 {
		t.Fatal("resize reopened an already saved draft")
	}
}

func TestWorkspaceLatePasteCannotReenterInactiveInput(t *testing.T) {
	for _, search := range []bool{false, true} {
		t.Run(fmt.Sprintf("search=%t", search), func(t *testing.T) {
			m, _, _ := editingSecretModel(t)
			if search {
				m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
				m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
				m = updateModel(t, m, tea.KeyPressMsg{Text: "/"})
			}
			m.clipboardReader = &fakeClipboardReader{text: fakeSecret}
			updated, paste := m.Update(tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
			m = updated.(Model)
			m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
			if search {
				m = updateModel(t, m, tea.KeyPressMsg{Text: "/"})
			} else {
				m = updateModel(t, m, tea.KeyPressMsg{Text: "a"})
			}
			m = updateModel(t, m, paste())
			if (search && m.query.Value() != "") || (!search && m.form.editor.value("field:credential") != "") {
				t.Fatal("late paste was accepted after its input became inactive")
			}
		})
	}
}

func TestCatalogSearchIncludesAllEnvNamesAndKeepsFilters(t *testing.T) {
	m, _, _ := readyModel(t)
	m.items[0].EnvNames = []string{"PRIMARY_API_KEY", "SECONDARY_BASE_URL", "SECONDARY_CLIENT_ID"}
	m.items[1].EnvNames = []string{"OTHER_TOKEN"}
	for _, query := range []string{"primary_api_key", "SeCoNdArY_bAsE_uRl", "secondary_client_id"} {
		m.query.SetValue(query)
		m.applyFilters()
		if len(m.filtered) != 1 || m.filtered[0].ID != "item_openai" {
			t.Fatal("catalog search missed a primary or secondary env mapping")
		}
	}
	m.labels = map[string]bool{"work": true}
	m.applyFilters()
	if len(m.filtered) != 0 {
		t.Fatal("env-name search bypassed label filters")
	}
	m.labels = map[string]bool{}
	m.template = m.items[1].Template
	m.applyFilters()
	if len(m.filtered) != 0 {
		t.Fatal("env-name search bypassed template filters")
	}
	m.template = ""
	for _, query := range []string{"personal account", "item_openai", "personal", fakeSecret} {
		m.query.SetValue(query)
		m.applyFilters()
		want := 1
		if query == fakeSecret {
			want = 0
		}
		if len(m.filtered) != want {
			t.Fatal("env-name search changed existing metadata or value-exclusion behavior")
		}
	}
}
