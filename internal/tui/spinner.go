package tui

import "fmt"

// FormatStep renders a step in progress output.
// e.g. "◆ searching 4 registries for behavioral match..."
func FormatStep(msg string) string {
	return fmt.Sprintf("%s %s", Diamond.Render(), msg)
}

// FormatSubStep renders an indented finding.
// e.g. "  → ticket-classifier@2.1.0  skills.sh  +38pp delta  94%"
func FormatSubStep(msg string) string {
	return fmt.Sprintf("    %s %s", Arrow.Render(), msg)
}

// FormatResult renders a final result line.
// e.g. "✓ ticket-classifier@2.1.0 installed"
func FormatResult(msg string) string {
	return fmt.Sprintf("%s %s", Check.Render(), msg)
}

// FormatWarning renders a warning line.
func FormatWarning(msg string) string {
	return fmt.Sprintf("%s %s", Warn.Render(), msg)
}

// FormatError renders an error line.
func FormatError(msg string) string {
	return fmt.Sprintf("%s %s", ErrMark.Render(), msg)
}
