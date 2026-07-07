package cli

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
)

type projectConfigureApplyFunc func([]app.ProjectConfigureAction) (string, error)

type projectConfigureScreen int

const (
	projectConfigureMain projectConfigureScreen = iota
	projectConfigureEnvVar
	projectConfigureSecret
	projectConfigureAddSecret
	projectConfigureRemove
	projectConfigureReview
)

type projectConfigureModel struct {
	state   app.ProjectConfigureState
	apply   projectConfigureApplyFunc
	screen  projectConfigureScreen
	cursor  int
	message string
	err     error
	done    bool

	selectedEnv app.ProjectConfigureEnvVar
	pending     []app.ProjectConfigureAction
}

func newProjectConfigureCmd(appSvc *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Interactively configure project secret bindings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			state, err := appSvc.ProjectConfigureState(configPath, vaultPath)
			if err != nil {
				return err
			}
			apply := func(actions []app.ProjectConfigureAction) (string, error) {
				return appSvc.ApplyProjectConfigureActions(configPath, vaultPath, actions)
			}
			model, err := runProjectConfigureTUI(cmd.InOrStdin(), cmd.OutOrStdout(), state, apply)
			if err != nil {
				return err
			}
			if model.err != nil {
				return model.err
			}
			return nil
		},
	}
}

func runProjectConfigureTUI(in io.Reader, out io.Writer, state app.ProjectConfigureState, apply projectConfigureApplyFunc) (projectConfigureModel, error) {
	initial := newProjectConfigureModel(state, apply)
	program := tea.NewProgram(initial, tea.WithInput(in), tea.WithOutput(out))
	finalModel, err := program.Run()
	if err != nil {
		return projectConfigureModel{}, err
	}
	model, ok := finalModel.(projectConfigureModel)
	if !ok {
		return projectConfigureModel{}, fmt.Errorf("unexpected project configure model: %T", finalModel)
	}
	return model, nil
}

func newProjectConfigureModel(state app.ProjectConfigureState, apply projectConfigureApplyFunc) projectConfigureModel {
	return projectConfigureModel{state: state, apply: apply, screen: projectConfigureMain}
}

func (m projectConfigureModel) Init() tea.Cmd { return nil }

func (m projectConfigureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit
		case "esc", "backspace":
			m.back()
			return m, nil
		case "up", "k":
			m.move(-1)
			return m, nil
		case "down", "j":
			m.move(1)
			return m, nil
		case "enter":
			return m.selectItem()
		case "a":
			if m.screen == projectConfigureReview {
				return m.applyPending()
			}
		}
	}
	return m, nil
}

func (m projectConfigureModel) View() string {
	if m.done {
		if m.message != "" {
			return m.message + "\n"
		}
		return ""
	}
	var out strings.Builder
	fmt.Fprintln(&out, "Shelf Project Configure")
	fmt.Fprintln(&out)
	if m.message != "" {
		fmt.Fprintf(&out, "%s\n\n", m.message)
	}
	if m.err != nil {
		fmt.Fprintf(&out, "error: %s\n\n", m.err)
	}
	switch m.screen {
	case projectConfigureMain:
		renderMenu(&out, "Actions", m.cursor, m.mainItems())
	case projectConfigureEnvVar:
		renderMenu(&out, "Resolve env-file variable", m.cursor, m.envVarItems())
	case projectConfigureSecret:
		renderMenu(&out, fmt.Sprintf("Bind %s to secret", m.selectedEnv.Name), m.cursor, m.secretItems())
	case projectConfigureAddSecret:
		renderMenu(&out, "Add existing secret binding", m.cursor, m.secretItems())
	case projectConfigureRemove:
		renderMenu(&out, "Remove project binding", m.cursor, m.removeItems())
	case projectConfigureReview:
		renderMenu(&out, "Pending changes", m.cursor, m.reviewItems())
	}
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "↑/k ↓/j move · enter select · esc back · q quit")
	if m.screen == projectConfigureReview && len(m.pending) > 0 {
		fmt.Fprintln(&out, "a apply pending changes")
	}
	return out.String()
}

func (m *projectConfigureModel) back() {
	m.message = ""
	m.err = nil
	if m.screen == projectConfigureMain {
		return
	}
	m.screen = projectConfigureMain
	m.cursor = 0
}

func (m *projectConfigureModel) move(delta int) {
	items := m.currentItems()
	if len(items) == 0 {
		m.cursor = 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = len(items) - 1
	}
	if m.cursor >= len(items) {
		m.cursor = 0
	}
}

func (m projectConfigureModel) selectItem() (tea.Model, tea.Cmd) {
	m.message = ""
	m.err = nil
	switch m.screen {
	case projectConfigureMain:
		return m.selectMain()
	case projectConfigureEnvVar:
		vars := m.envVars()
		if len(vars) == 0 {
			m.message = "no unbound env-file variables"
			return m, nil
		}
		m.selectedEnv = vars[m.cursor]
		m.screen = projectConfigureSecret
		m.cursor = 0
	case projectConfigureSecret:
		secrets := m.state.SecretPaths
		if len(secrets) == 0 {
			m.message = "no vault secrets found"
			return m, nil
		}
		m.pending = append(m.pending, app.ProjectConfigureAction{Kind: app.ProjectConfigureAddPath, Env: m.selectedEnv.Name, Path: secrets[m.cursor]})
		m.message = fmt.Sprintf("queued bind %s -> %s", m.selectedEnv.Name, secrets[m.cursor])
		m.screen = projectConfigureMain
		m.cursor = 0
	case projectConfigureAddSecret:
		secrets := m.state.SecretPaths
		if len(secrets) == 0 {
			m.message = "no vault secrets found"
			return m, nil
		}
		m.pending = append(m.pending, app.ProjectConfigureAction{Kind: app.ProjectConfigureAddPath, Path: secrets[m.cursor]})
		m.message = fmt.Sprintf("queued add %s", secrets[m.cursor])
		m.screen = projectConfigureMain
		m.cursor = 0
	case projectConfigureRemove:
		entries := m.state.Entries
		if len(entries) == 0 {
			m.message = "no project bindings to remove"
			return m, nil
		}
		m.pending = append(m.pending, app.ProjectConfigureAction{Kind: app.ProjectConfigureRemove, Key: entries[m.cursor].Key})
		m.message = fmt.Sprintf("queued remove %s", entries[m.cursor].Key)
		m.screen = projectConfigureMain
		m.cursor = 0
	case projectConfigureReview:
		return m.applyPending()
	}
	return m, nil
}

func (m projectConfigureModel) selectMain() (tea.Model, tea.Cmd) {
	items := m.mainItems()
	if len(items) == 0 {
		return m, nil
	}
	switch items[m.cursor] {
	case "Resolve env-file variable":
		m.screen = projectConfigureEnvVar
	case "Add existing secret binding":
		m.screen = projectConfigureAddSecret
	case "Remove project binding":
		m.screen = projectConfigureRemove
	case "Review pending changes":
		m.screen = projectConfigureReview
	case "Apply pending changes":
		return m.applyPending()
	case "Quit":
		m.done = true
		return m, tea.Quit
	}
	m.cursor = 0
	return m, nil
}

func (m projectConfigureModel) applyPending() (tea.Model, tea.Cmd) {
	if len(m.pending) == 0 {
		m.message = "no pending changes"
		return m, nil
	}
	result, err := m.apply(m.pending)
	if err != nil {
		m.err = err
		return m, nil
	}
	m.message = strings.TrimSpace(result)
	m.pending = nil
	m.done = true
	return m, tea.Quit
}

func (m projectConfigureModel) mainItems() []string {
	return []string{"Resolve env-file variable", "Add existing secret binding", "Remove project binding", "Review pending changes", "Apply pending changes", "Quit"}
}

func (m projectConfigureModel) envVars() []app.ProjectConfigureEnvVar {
	vars := make([]app.ProjectConfigureEnvVar, 0)
	for _, file := range m.state.EnvFiles {
		if file.Error != "" {
			continue
		}
		for _, variable := range file.Vars {
			if !variable.Bound {
				vars = append(vars, variable)
			}
		}
	}
	return vars
}

func (m projectConfigureModel) envVarItems() []string {
	vars := m.envVars()
	items := make([]string, 0, len(vars))
	for _, variable := range vars {
		state := "set"
		if variable.Empty {
			state = "empty"
		}
		items = append(items, fmt.Sprintf("%s (%s)", variable.Name, state))
	}
	return items
}

func (m projectConfigureModel) secretItems() []string {
	return append([]string(nil), m.state.SecretPaths...)
}

func (m projectConfigureModel) removeItems() []string {
	items := make([]string, 0, len(m.state.Entries))
	for _, entry := range m.state.Entries {
		required := "optional"
		if entry.Required {
			required = "required"
		}
		if entry.Env != "" {
			items = append(items, fmt.Sprintf("%s %s -> %s (%s)", entry.Kind, entry.Selector, entry.Env, required))
		} else {
			items = append(items, fmt.Sprintf("%s %s (%s)", entry.Kind, entry.Selector, required))
		}
	}
	return items
}

func (m projectConfigureModel) reviewItems() []string {
	if len(m.pending) == 0 {
		return []string{"No pending changes"}
	}
	items := make([]string, 0, len(m.pending))
	for _, action := range m.pending {
		switch action.Kind {
		case app.ProjectConfigureAddPath:
			if action.Env != "" {
				items = append(items, fmt.Sprintf("add %s -> %s", action.Path, action.Env))
			} else {
				items = append(items, fmt.Sprintf("add %s", action.Path))
			}
		case app.ProjectConfigureRemove:
			items = append(items, fmt.Sprintf("remove %s", action.Key))
		}
	}
	return items
}

func (m projectConfigureModel) currentItems() []string {
	switch m.screen {
	case projectConfigureMain:
		return m.mainItems()
	case projectConfigureEnvVar:
		return m.envVarItems()
	case projectConfigureSecret, projectConfigureAddSecret:
		return m.secretItems()
	case projectConfigureRemove:
		return m.removeItems()
	case projectConfigureReview:
		return m.reviewItems()
	default:
		return nil
	}
}

func renderMenu(out *strings.Builder, title string, cursor int, items []string) {
	fmt.Fprintln(out, title)
	if len(items) == 0 {
		fmt.Fprintln(out, "  none")
		return
	}
	for i, item := range items {
		marker := "  "
		if i == cursor {
			marker = "❯ "
		}
		fmt.Fprintf(out, "%s%s\n", marker, item)
	}
}
