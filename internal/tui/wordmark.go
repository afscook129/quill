package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderWordmark(version string) string {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3)

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("◆ quill"),
		lipgloss.NewStyle().Foreground(ColorSecondary).Render("outcome intelligence for AI agent skills"),
		"",
		Subtle.Render(fmt.Sprintf("v%s · registry.quill.dev · MIT", version)),
	}

	return border.Render(strings.Join(lines, "\n"))
}

func RenderHelp(version string) string {
	wordmark := RenderWordmark(version)

	sections := []string{
		wordmark,
		"",
		Title.Render("Everyday"),
		fmt.Sprintf("  %-28s %s", "init", "Set up a project"),
		fmt.Sprintf("  %-28s %s", "add <skill-or-description>", "Install a skill"),
		fmt.Sprintf("  %-28s %s", "status", "Health summary"),
		fmt.Sprintf("  %-28s %s", "search <query>", "Find skills by description"),
		fmt.Sprintf("  %-28s %s", "fix", "Resolve issues from status"),
		fmt.Sprintf("  %-28s %s", "upgrade [skill]", "Show available upgrades"),
		"",
		Title.Render("Benchmarking"),
		fmt.Sprintf("  %-28s %s", "bench <skill-or-path>", "Benchmark a skill"),
		fmt.Sprintf("  %-28s %s", "bench-history <skill>", "Results over time"),
		fmt.Sprintf("  %-28s %s", "bench-compare <v1> <v2>", "A/B between versions"),
		fmt.Sprintf("  %-28s %s", "retire <skill>", "Check if still earning its place"),
		"",
		Title.Render("Understanding"),
		fmt.Sprintf("  %-28s %s", "explain <skill>", "What a skill does and where it fails"),
		fmt.Sprintf("  %-28s %s", "compare <a> <b>", "Head-to-head comparison"),
		"",
		Title.Render("Team & CI"),
		fmt.Sprintf("  %-28s %s", "validate <skill>", "Full eval suite for CI"),
		fmt.Sprintf("  %-28s %s", "lock", "Generate quill.lock"),
		fmt.Sprintf("  %-28s %s", "team status", "Branch divergence (git-native)"),
		fmt.Sprintf("  %-28s %s", "audit [path]", "Full security scan"),
		fmt.Sprintf("  %-28s %s", "migrate --from <m> --to <m>", "Analyze model switch"),
		"",
		Title.Render("Publishing"),
		fmt.Sprintf("  %-28s %s", "publish", "Publish to registry"),
		fmt.Sprintf("  %-28s %s", "registry init", "Scaffold private registry"),
		fmt.Sprintf("  %-28s %s", "sbom", "Security surface inventory"),
		"",
		Subtle.Render("  Use quill <command> --help for details on any command."),
		"",
	}

	return strings.Join(sections, "\n")
}
