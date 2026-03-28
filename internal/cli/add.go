package cli

import (
	"fmt"
	"strings"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var model string
	var skipEvals bool
	var dryRun bool
	var source string

	cmd := &cobra.Command{
		Use:   "add <skill-or-description>",
		Short: "Install a skill — by name, version, or natural language description",
		Long: `Install by name, version, or describe what you need in plain English.
Quill searches across all registries, ranks by delta (how much the skill
actually improves your agent), and installs the best match.

Requires the Quill registry (coming soon). For now, install skills manually
and use quill bench to measure them.

Examples:
  quill add "classify support tickets by urgency"
  quill add ticket-classifier
  quill add ticket-classifier@2.1.0`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			return runAdd(query, model, skipEvals, dryRun, source)
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "Override detected model")
	cmd.Flags().BoolVar(&skipEvals, "skip-evals", false, "Install without validation (logged)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without installing")
	cmd.Flags().StringVar(&source, "source", "", "Limit to specific registry")

	return cmd
}

func runAdd(query string, model string, skipEvals bool, dryRun bool, source string) error {
	fmt.Println()
	fmt.Printf("  %s add %s\n", tui.Diamond.Render(), tui.Bold.Render(query))
	fmt.Println()
	fmt.Println(tui.FormatWarning("registry not yet connected"))
	fmt.Println()
	fmt.Println("  quill add requires the Quill registry to search and install skills.")
	fmt.Println("  The registry is being built.")
	fmt.Println()
	fmt.Println("  To measure skills you already have installed:")
	fmt.Println()
	fmt.Printf("    1. Place your SKILL.md in a directory (e.g., ./skills/my-skill/)\n")
	fmt.Printf("    2. Create evals: ./skills/my-skill/evals/evals.json\n")
	fmt.Printf("    3. Run: %s\n", tui.Bold.Render("quill bench ./skills/my-skill"))
	fmt.Println()
	fmt.Println(tui.Subtle.Render("  See spec/EVAL_FORMAT.md for the eval case format."))
	fmt.Println()
	return nil
}
