package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // violet
	ColorSecondary = lipgloss.Color("#A78BFA") // light violet
	ColorSuccess   = lipgloss.Color("#10B981") // green
	ColorWarning   = lipgloss.Color("#F59E0B") // amber
	ColorError     = lipgloss.Color("#EF4444") // red
	ColorMuted     = lipgloss.Color("#6B7280") // gray
	ColorText      = lipgloss.Color("#E5E7EB") // light gray

	// Symbols
	SymbolDiamond = "◆"
	SymbolCheck   = "✓"
	SymbolWarn    = "⚠"
	SymbolError   = "✗"
	SymbolPending = "○"
	SymbolArrow   = "→"

	// Styles
	Bold = lipgloss.NewStyle().Bold(true)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)

	Subtle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	Success = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	Warning = lipgloss.NewStyle().
		Foreground(ColorWarning)

	Error = lipgloss.NewStyle().
		Foreground(ColorError)

	Diamond = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		SetString(SymbolDiamond)

	Check = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		SetString(SymbolCheck)

	Warn = lipgloss.NewStyle().
		Foreground(ColorWarning).
		SetString(SymbolWarn)

	ErrMark = lipgloss.NewStyle().
		Foreground(ColorError).
		SetString(SymbolError)

	Pending = lipgloss.NewStyle().
		Foreground(ColorMuted).
		SetString(SymbolPending)

	Arrow = lipgloss.NewStyle().
		Foreground(ColorMuted).
		SetString(SymbolArrow)
)
