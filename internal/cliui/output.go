package cliui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

type SummaryField struct {
	Key   string
	Value string
}

type Table struct {
	Title   string
	Headers []string
	Rows    [][]string
}

var (
	titleStyle = lipgloss.NewStyle()
	keyStyle   = lipgloss.NewStyle()
	cellStyle  = lipgloss.NewStyle().PaddingRight(2)
)

func Title(w io.Writer, title string) {
	fmt.Fprintln(w, titleStyle.Render(title))
}

func Empty(w io.Writer, message string) {
	fmt.Fprintln(w, message)
}

func Summary(w io.Writer, title string, fields []SummaryField) {
	Title(w, title)
	width := 0
	for _, field := range fields {
		if len(field.Key) > width {
			width = len(field.Key)
		}
	}
	for _, field := range fields {
		if field.Value == "" {
			continue
		}
		fmt.Fprintf(w, "  %s: %s\n", keyStyle.Render(padRight(field.Key, width)), field.Value)
	}
}

func RenderTable(w io.Writer, t Table) {
	if t.Title != "" {
		fmt.Fprintf(w, "%s (%d):\n", t.Title, len(t.Rows))
	}
	if len(t.Rows) == 0 {
		Empty(w, "No rows found.")
		return
	}
	tableView := table.New().
		Headers(t.Headers...).
		Rows(t.Rows...).
		Border(lipgloss.HiddenBorder()).
		BorderTop(false).
		BorderBottom(false).
		BorderHeader(false).
		BorderColumn(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			return cellStyle
		})
	writeRendered(w, tableView.String())
}

func writeRendered(w io.Writer, rendered string) {
	for _, line := range strings.Split(rendered, "\n") {
		fmt.Fprintln(w, strings.TrimRight(strings.TrimLeft(line, " "), " "))
	}
}

func padRight(value string, width int) string {
	for len(value) < width {
		value += " "
	}
	return value
}
