package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	Header string
	Width  int
}

type Row struct {
	Icon   string
	Values []string
}

func RenderTable(columns []Column, rows []Row) string {
	if len(columns) == 0 || len(rows) == 0 {
		return ""
	}

	var sb strings.Builder

	// Header
	headerParts := make([]string, len(columns))
	for i, col := range columns {
		headerParts[i] = lipgloss.NewStyle().
			Width(col.Width).
			Foreground(ColorMuted).
			Render(col.Header)
	}
	sb.WriteString("  " + strings.Join(headerParts, "  ") + "\n")

	// Rows
	for _, row := range rows {
		parts := make([]string, len(columns))
		for i, col := range columns {
			val := ""
			if i < len(row.Values) {
				val = row.Values[i]
			}
			parts[i] = lipgloss.NewStyle().
				Width(col.Width).
				Render(val)
		}
		icon := row.Icon
		if icon == "" {
			icon = " "
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", icon, strings.Join(parts, "  ")))
	}

	return sb.String()
}

func RenderProgressBar(pct float64, width int) string {
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := lipgloss.NewStyle().Foreground(ColorPrimary).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("░", empty))

	return bar
}
