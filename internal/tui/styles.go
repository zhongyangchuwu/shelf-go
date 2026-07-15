package tui

import "charm.land/lipgloss/v2"

const (
	mochaBase     = "#1e1e2e"
	mochaMantle   = "#181825"
	mochaSurface1 = "#45475a"
	mochaOverlay0 = "#6c7086"
	mochaText     = "#cdd6f4"
	mochaSubtext1 = "#bac2de"
	mochaBlue     = "#89b4fa"
	mochaMauve    = "#cba6f7"
	mochaGreen    = "#a6e3a1"
	mochaYellow   = "#f9e2af"
	mochaRed      = "#f38ba8"
)

type workspaceStyles struct {
	header           lipgloss.Style
	footer           lipgloss.Style
	status           lipgloss.Style
	panel            lipgloss.Style
	activePanel      lipgloss.Style
	panelTitle       lipgloss.Style
	activePanelTitle lipgloss.Style
}

func newWorkspaceStyles() workspaceStyles {
	return workspaceStyles{
		header:           lipgloss.NewStyle().Foreground(lipgloss.Color(mochaText)).Background(lipgloss.Color(mochaMantle)).Bold(true),
		footer:           lipgloss.NewStyle().Foreground(lipgloss.Color(mochaSubtext1)).Background(lipgloss.Color(mochaMantle)),
		status:           lipgloss.NewStyle().Foreground(lipgloss.Color(mochaGreen)).Background(lipgloss.Color(mochaMantle)),
		panel:            lipgloss.NewStyle().Foreground(lipgloss.Color(mochaText)).Background(lipgloss.Color(mochaBase)).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(mochaSurface1)).Padding(0, 1),
		activePanel:      lipgloss.NewStyle().Foreground(lipgloss.Color(mochaText)).Background(lipgloss.Color(mochaBase)).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(mochaBlue)).Padding(0, 1),
		panelTitle:       lipgloss.NewStyle().Foreground(lipgloss.Color(mochaMauve)).Bold(true),
		activePanelTitle: lipgloss.NewStyle().Foreground(lipgloss.Color(mochaBlue)).Bold(true),
	}
}

func (s workspaceStyles) renderHeader(content string, width int) string {
	return s.header.Width(max(1, width)).Render(content)
}

func (s workspaceStyles) renderFooter(content string, width int) string {
	return s.footer.Width(max(1, width)).Render(content)
}

func (s workspaceStyles) renderStatus(content string, width int) string {
	return s.status.Width(max(1, width)).Render(content)
}

func (s workspaceStyles) renderPanel(title, content string, active bool, outerWidth, outerHeight int) string {
	frame := s.panel
	titleStyle := s.panelTitle
	prefix := "  "
	if active {
		frame = s.activePanel
		titleStyle = s.activePanelTitle
		prefix = "› "
	}
	body := titleStyle.Render(prefix + title)
	if content != "" {
		body += "\n\n" + content
	}
	return frame.Width(max(1, outerWidth)).Height(max(1, outerHeight)).Render(body)
}

func workspacePanelWidths(width int) (left, right int) {
	left = max(34, width*36/100)
	right = max(50, width-left-1)
	return left, right
}

func workspacePanelHeight(height int, hasStatus bool) int {
	footerLines := 1
	if hasStatus {
		footerLines++
	}
	return max(8, height-footerLines-3)
}

func panelContentWidth(outerWidth int) int {
	return max(1, outerWidth-4)
}

func (s workspaceStyles) renderCatalogStatus(status string) string {
	symbol := "●"
	color := mochaGreen
	switch status {
	case "missing":
		symbol, color = "×", mochaRed
	case "attention":
		symbol, color = "!", mochaYellow
	case "unavailable":
		symbol, color = "·", mochaOverlay0
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(symbol)

}
