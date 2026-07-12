package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/clipboard"
)

const (
	minWidth  = 100
	minHeight = 24
)

type screen uint8

const (
	screenLoading screen = iota
	screenWorkspace
	screenSearch
	screenFilters
	screenForm
	screenConfirmDelete
	screenError
	screenResize
	screenHelp
)

type pane uint8

const (
	paneCatalog pane = iota
	paneDetail
)

type catalogLoadedMsg struct {
	items []app.SecretCatalogItem
	err   error
}

type detailLoadedMsg struct {
	id   string
	info app.SecretInfo
	err  error
}

type healthLoadedMsg struct {
	report app.CredentialAuditReport
	err    error
}

type actionDoneMsg struct {
	status  string
	refresh bool
	err     error
}

type Model struct {
	service         Service
	clipboardWriter clipboard.Writer
	options         Options

	screen screen
	pane   pane
	width  int
	height int

	items             []app.SecretCatalogItem
	filtered          []app.SecretCatalogItem
	selected          int
	detail            app.SecretInfo
	detailViewport    viewport.Model
	health            map[string]app.CredentialHealthItem
	healthUnavailable bool

	fieldSelected  int
	query          textinput.Model
	template       string
	labels         map[string]bool
	filterTemplate string
	filterLabels   map[string]bool
	filterCursor   int
	filterChoices  []filterChoice

	status string
	form   formState
	styles workspaceStyles
}

type filterChoice struct {
	kind  string
	value string
}

func NewModel(service Service, clipboardWriter clipboard.Writer, options Options) Model {
	query := textinput.New()
	query.Prompt = "/ "
	query.Placeholder = "Search title, id, template, label, or alias"
	return Model{
		service:         service,
		clipboardWriter: clipboardWriter,
		options:         options,
		screen:          screenLoading,
		pane:            paneCatalog,
		fieldSelected:   -1,
		query:           query,
		labels:          make(map[string]bool),
		styles:          newWorkspaceStyles(),
		detailViewport:  newDetailViewport(),
		health:          make(map[string]app.CredentialHealthItem),
	}
}

func (m Model) Init() tea.Cmd {
	return m.reloadCatalog()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.tooSmall() {
			m.screen = screenResize
		} else if m.screen == screenResize {
			m.screen = screenWorkspace
		}
		m.configureDetailViewport()
		m.syncDetailViewport()
		return m, nil
	case healthLoadedMsg:
		if msg.err != nil {
			m.health = nil
			m.healthUnavailable = true
			m.syncDetailViewport()
			return m, nil
		}
		m.health = credentialHealthItemsByID(msg.report)
		m.healthUnavailable = false
		m.syncDetailViewport()
		return m, nil
	case catalogLoadedMsg:
		if msg.err != nil {
			m.screen = screenError
			return m, nil
		}
		m.items = msg.items
		m.applyFilters()
		m.screen = m.screenAfterLoad()
		if len(m.filtered) == 0 {
			m.detail = app.SecretInfo{}
			return m, nil
		}
		return m, m.loadDetail(m.filtered[m.selected].ID)
	case detailLoadedMsg:
		if msg.err != nil {
			m.screen = screenError
			return m, nil
		}
		if m.currentID() == msg.id {
			m.detail = msg.info
			m.fieldSelected = -1
			m.syncDetailViewport()
		}
		return m, nil
	case actionDoneMsg:
		if msg.err != nil {
			m.screen = screenError
			return m, nil
		}
		m.status = msg.status
		m.form = formState{}
		if !msg.refresh {
			m.screen = screenWorkspace
			return m, nil
		}
		m.screen = screenLoading
		return m, m.reloadCatalog()
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	if m.screen == screenSearch {
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		m.applyFilters()
		return m, cmd
	}
	if m.screen == screenForm {
		return m.updateForm(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" || key == "q" {
		return m, tea.Quit
	}
	if m.screen == screenResize {
		if key == "r" {
			return m, m.reloadCatalog()
		}
		return m, nil
	}
	if m.screen == screenError {
		if key == "r" {
			m.screen = screenLoading
			return m, m.reloadCatalog()
		}
		return m, nil
	}
	if m.screen == screenHelp {
		if key == "?" || key == "esc" {
			m.screen = screenWorkspace
		}
		return m, nil
	}
	if m.screen == screenForm {
		return m.updateForm(msg)
	}
	if m.screen == screenConfirmDelete {
		return m.handleDeleteConfirm(key)
	}
	if m.screen == screenFilters {
		return m.handleFilterKey(key)
	}
	if m.screen == screenSearch {
		if key == "esc" {
			m.query.Blur()
			m.screen = screenWorkspace
			return m, nil
		}
		if key == "enter" {
			m.query.Blur()
			m.screen = screenWorkspace
			return m, m.loadCurrentDetail()
		}
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		m.applyFilters()
		return m, cmd
	}

	switch key {
	case "?":
		m.screen = screenHelp
		return m, nil
	case "/":
		m.query.Focus()
		m.screen = screenSearch
		return m, nil
	case "f":
		m.filterTemplate = m.template
		m.filterLabels = cloneLabels(m.labels)
		m.buildFilterChoices()
		m.screen = screenFilters
		return m, nil
	case "r":
		m.screen = screenLoading
		return m, m.reloadCatalog()
	case "tab", "h", "l":
		if m.pane == paneCatalog {
			m.pane = paneDetail
		} else {
			m.pane = paneCatalog
		}
		return m, nil
	case "up", "k":
		return m.moveSelection(-1)
	case "pgup", "pgdown", "ctrl+u", "ctrl+d":
		if m.pane != paneDetail {
			return m, nil
		}
		viewport, cmd := m.detailViewport.Update(msg)
		m.detailViewport = viewport
		return m, cmd
	case "down", "j":
		return m.moveSelection(1)
	case "a":
		return m.startCreateForm()
	case "e":
		return m.startEditForm()
	case "c":
		return m.copySelectedField()
	case "d":
		if m.currentID() != "" {
			m.screen = screenConfirmDelete
		}
		return m, nil
	}
	return m, nil
}

func (m Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	return view
}

func (m Model) render() string {
	if m.width > 0 && m.height > 0 && m.tooSmall() {
		return renderResize(m.width, m.height)
	}
	switch m.screen {
	case screenLoading:
		return "Shelf  ·  Catalog\n\nLoading local vault…\n"
	case screenError:
		return "Shelf  ·  Catalog\n\nThe catalog or requested action could not be completed.\nNo secret value was displayed.\n\nr retry   q quit\n"
	case screenHelp:
		return m.renderHelp()
	}
	return m.renderWorkspace()
}

func (m Model) renderWorkspace() string {
	header := m.styles.renderHeader(fmt.Sprintf("Shelf  ·  Catalog    %d items  ·  template: %s  ·  labels: %s  ·  health: %s", len(m.filtered), displayTemplate(m.template), displayLabels(m.labels), m.healthSummary()), m.width)
	leftWidth, rightWidth := workspacePanelWidths(m.width)
	panelHeight := workspacePanelHeight(m.height, m.status != "")
	left := m.styles.renderPanel("Catalog", m.renderCatalog(panelContentWidth(leftWidth)), m.catalogPanelActive(), leftWidth, panelHeight)
	right := m.styles.renderPanel(m.rightPanelTitle(), m.renderRight(panelContentWidth(rightWidth)), m.rightPanelActive(), rightWidth, panelHeight)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	footer := m.styles.renderFooter("Tab pane  a add  e edit  c copy  d delete  / search  f filters  r reload  ? help  q quit", m.width)
	if m.status != "" {
		footer = m.styles.renderStatus(m.status, m.width) + "\n" + footer
	}
	return header + "\n" + body + "\n" + footer + "\n"
}

func (m Model) renderCatalog(width int) string {
	var b strings.Builder
	if m.screen == screenSearch {
		b.WriteString(m.query.View() + "\n\n")
	}
	if len(m.filtered) == 0 {
		b.WriteString("No secrets found.\n\nPress a to create an item.")
		return b.String()
	}
	for i, item := range m.filtered {
		cursor := " "
		if i == m.selected {
			cursor = "›"
		}
		state := "unset"
		if item.ValueSet {
			state = "set"
		}
		if health, ok := m.health[item.ID]; ok {
			state += " · " + healthLabel(health.Status)
		}
		row := fmt.Sprintf("%s %-20s %-16s %s", cursor, truncate(item.Title, 20), truncate(item.Template, 16), state)
		b.WriteString(truncate(row, width) + "\n")
	}
	return b.String()
}

func (m Model) renderRight(width int) string {
	switch m.screen {
	case screenFilters:
		return m.renderFilters(width)
	case screenForm:
		return m.renderForm(width)
	case screenConfirmDelete:
		return m.renderDeleteConfirm(width)
	}
	return m.renderDetail(width)
}

func (m Model) catalogPanelActive() bool {
	return m.screen == screenSearch || (m.screen == screenWorkspace && m.pane == paneCatalog)
}

func (m Model) rightPanelActive() bool {
	switch m.screen {
	case screenFilters, screenForm, screenConfirmDelete:
		return true
	default:
		return m.pane == paneDetail
	}
}

func (m Model) rightPanelTitle() string {
	switch m.screen {
	case screenFilters:
		return "Filters"
	case screenConfirmDelete:
		return "Remove item"
	case screenForm:
		switch m.form.kind {
		case formCreate:
			return "Create item"
		case formMetadata:
			return "Edit item"
		case formField:
			return "Edit field"
		}
	}
	return "Detail"
}

func (m Model) renderDetail(_ int) string {
	return m.detailViewport.View()
}

func (m Model) renderHelp() string {
	return "Shelf TUI help\n\n" +
		"Tab / h / l  switch catalog and detail\n" +
		"j / k          move selection\n" +
		"/              search safe catalog metadata\n" +
		"f              filter template and labels\n" +
		"a              create item\n" +
		"e              edit item or selected field\n" +
		"c              copy selected/default field\n" +
		"d              delete selected item after confirmation\n" +
		"r              reload\n" +
		"esc            cancel\n" +
		"q / ctrl+c     quit\n\n" +
		"? / esc return\n"
}

func (m Model) renderFilters(_ int) string {
	var b strings.Builder
	for i, choice := range m.filterChoices {
		cursor := " "
		if i == m.filterCursor {
			cursor = "›"
		}
		selected := " "
		if choice.kind == "template" && ((choice.value == "All" && m.filterTemplate == "") || choice.value == m.filterTemplate) {
			selected = "x"
		}
		if choice.kind == "label" && m.filterLabels[choice.value] {
			selected = "x"
		}
		fmt.Fprintf(&b, "%s [%s] %s: %s\n", cursor, selected, choice.kind, choice.value)
	}
	b.WriteString("\nspace toggle   enter apply   c clear   esc cancel")
	return b.String()
}

func (m Model) renderDeleteConfirm(_ int) string {
	return fmt.Sprintf("Remove %q?\n\nid:       %s\ntemplate: %s\n\nThis removes the local item and cannot be undone by Shelf.\n\nenter confirm   esc cancel", m.detail.Title, m.detail.ID, m.detail.Template)
}

func (m Model) tooSmall() bool {
	return m.width > 0 && m.height > 0 && (m.width < minWidth || m.height < minHeight)
}

func (m Model) screenAfterLoad() screen {
	if m.tooSmall() {
		return screenResize
	}
	return screenWorkspace
}

func (m *Model) applyFilters() {
	query := strings.ToLower(strings.TrimSpace(m.query.Value()))
	m.filtered = m.filtered[:0]
	for _, item := range m.items {
		if m.template != "" && item.Template != m.template {
			continue
		}
		if !hasAllLabels(item.Labels, m.labels) || !matchesCatalogQuery(item, query) {
			continue
		}
		m.filtered = append(m.filtered, item)
	}
	if m.selected >= len(m.filtered) {
		m.selected = max(0, len(m.filtered)-1)
	}
}

func (m *Model) buildFilterChoices() {
	templates := map[string]struct{}{}
	labels := map[string]struct{}{}
	for _, item := range m.items {
		templates[item.Template] = struct{}{}
		for _, label := range item.Labels {
			labels[label] = struct{}{}
		}
	}
	m.filterChoices = []filterChoice{{kind: "template", value: "All"}}
	for _, value := range sortedKeys(templates) {
		m.filterChoices = append(m.filterChoices, filterChoice{kind: "template", value: value})
	}
	for _, value := range sortedKeys(labels) {
		m.filterChoices = append(m.filterChoices, filterChoice{kind: "label", value: value})
	}
	m.filterCursor = 0
}

func (m Model) handleFilterKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenWorkspace
	case "enter":
		m.template = m.filterTemplate
		m.labels = m.filterLabels
		m.applyFilters()
		m.screen = screenWorkspace
		return m, m.loadCurrentDetail()
	case "c":
		m.filterTemplate = ""
		m.filterLabels = make(map[string]bool)
		m.buildFilterChoices()
	case "up", "k":
		if m.filterCursor > 0 {
			m.filterCursor--
		}
	case "down", "j":
		if m.filterCursor < len(m.filterChoices)-1 {
			m.filterCursor++
		}
	case "space":
		if len(m.filterChoices) == 0 {
			return m, nil
		}
		choice := m.filterChoices[m.filterCursor]
		if choice.kind == "template" {
			if choice.value == "All" {
				m.filterTemplate = ""
			} else {
				m.filterTemplate = choice.value
			}
		} else if m.filterLabels[choice.value] {
			delete(m.filterLabels, choice.value)
		} else {
			m.filterLabels[choice.value] = true
		}
	}
	return m, nil
}

func (m Model) moveSelection(delta int) (tea.Model, tea.Cmd) {
	if m.pane == paneDetail {
		if len(m.detail.Fields) == 0 {
			return m, nil
		}
		m.fieldSelected = min(max(0, m.fieldSelected+delta), len(m.detail.Fields)-1)
		m.syncDetailViewport()
		return m, nil
	}
	if len(m.filtered) == 0 {
		return m, nil
	}
	m.selected = min(max(0, m.selected+delta), len(m.filtered)-1)
	return m, m.loadCurrentDetail()
}

func (m Model) loadCatalog() tea.Cmd {
	return func() tea.Msg {
		items, err := m.service.SecretCatalog(m.options.ConfigPath, m.options.VaultPath)
		return catalogLoadedMsg{items: items, err: err}
	}
}

func (m Model) loadHealth() tea.Cmd {
	return func() tea.Msg {
		report, err := m.service.CredentialAudit(m.options.ConfigPath, m.options.VaultPath, app.CredentialAuditRequest{WithinDays: app.DefaultCredentialAuditWithinDays})
		return healthLoadedMsg{report: report, err: err}
	}
}

func (m Model) reloadCatalog() tea.Cmd {
	return tea.Batch(m.loadCatalog(), m.loadHealth())
}

func credentialHealthItemsByID(report app.CredentialAuditReport) map[string]app.CredentialHealthItem {
	health := make(map[string]app.CredentialHealthItem, len(report.Items))
	for _, candidate := range report.Items {
		current, exists := health[candidate.ID]
		if !exists || healthStatusRank(candidate.Status) < healthStatusRank(current.Status) {
			health[candidate.ID] = candidate
		}
	}
	return health
}

func healthStatusRank(status app.CredentialHealthStatus) int {
	switch status {
	case app.CredentialHealthExpired:
		return 0
	case app.CredentialHealthExpiring:
		return 1
	case app.CredentialHealthMissing:
		return 2
	case app.CredentialHealthMalformed:
		return 3
	default:
		return 4
	}
}

func (m Model) loadDetail(id string) tea.Cmd {
	if id == "" {
		return nil
	}
	return func() tea.Msg {
		info, err := m.service.SecretInfo(m.options.ConfigPath, m.options.VaultPath, id)
		return detailLoadedMsg{id: id, info: info, err: err}
	}
}

func (m Model) loadCurrentDetail() tea.Cmd {
	return m.loadDetail(m.currentID())
}

func (m Model) currentID() string {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return ""
	}
	return m.filtered[m.selected].ID
}

func displayTemplate(template string) string {
	if template == "" {
		return "All"
	}
	return template
}

func displayLabels(labels map[string]bool) string {
	values := make([]string, 0, len(labels))
	for label := range labels {
		values = append(values, label)
	}
	if len(values) == 0 {
		return "All"
	}
	sort.Strings(values)
	return strings.Join(values, " ∩ ")
}

func matchesCatalogQuery(item app.SecretCatalogItem, query string) bool {
	if query == "" {
		return true
	}
	for _, value := range append([]string{item.ID, item.Title, item.Template}, append(item.Labels, item.Aliases...)...) {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}

func hasAllLabels(labels []string, required map[string]bool) bool {
	for label := range required {
		found := false
		for _, candidate := range labels {
			if label == candidate {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func sortedKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func truncate(value string, width int) string {
	if width <= 0 || len(value) <= width {
		return value
	}
	if width == 1 {
		return "…"
	}
	return value[:width-1] + "…"
}

func renderResize(width, height int) string {
	return fmt.Sprintf("Shelf TUI needs a terminal at least %d columns × %d rows.\nCurrent size: %d × %d\n\nResize the terminal, then press r.  q quit\n", minWidth, minHeight, width, height)
}

func cloneLabels(labels map[string]bool) map[string]bool {
	clone := make(map[string]bool, len(labels))
	for label, selected := range labels {
		clone[label] = selected
	}
	return clone
}

func (m Model) healthSummary() string {
	if m.healthUnavailable {
		return "unavailable"
	}
	counts := map[app.CredentialHealthStatus]int{}
	for _, health := range m.health {
		counts[health.Status]++
	}
	parts := make([]string, 0, 4)
	for _, status := range []app.CredentialHealthStatus{app.CredentialHealthExpired, app.CredentialHealthExpiring, app.CredentialHealthMissing, app.CredentialHealthMalformed} {
		if count := counts[status]; count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", count, healthLabel(status)))
		}
	}
	if len(parts) == 0 {
		return "clear"
	}
	return strings.Join(parts, ", ")
}

func healthLabel(status app.CredentialHealthStatus) string {
	switch status {
	case app.CredentialHealthExpired:
		return "expired"
	case app.CredentialHealthExpiring:
		return "due"
	case app.CredentialHealthMissing:
		return "missing"
	case app.CredentialHealthMalformed:
		return "invalid"
	default:
		return "healthy"
	}
}
