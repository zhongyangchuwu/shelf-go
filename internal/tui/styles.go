package tui

import "charm.land/lipgloss/v2"

type workspaceStyles struct {
	header lipgloss.Style
	pane   lipgloss.Style
	focus  lipgloss.Style
}

func newWorkspaceStyles(dark bool) workspaceStyles {
	accent := lipgloss.Color("62")
	muted := lipgloss.Color("245")
	if !dark {
		accent = lipgloss.Color("27")
		muted = lipgloss.Color("241")
	}
	return workspaceStyles{
		header: lipgloss.NewStyle().Foreground(accent).Bold(true),
		pane:   lipgloss.NewStyle().Foreground(muted).Bold(true),
		focus:  lipgloss.NewStyle().Foreground(accent).Bold(true),
	}
}
