package cli

import (
	"fmt"
	"strings"

	"github.com/quill-dev/quill/internal/registry"
	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var model string
	var minScore float64
	var verified bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Find skills by behavioral description — ranked by how much they actually help",
		Long: `Search across all registries using natural language. Results are ranked by
delta (how much the skill improves outcomes) × pass rate × freshness,
specifically for your model.

Not ranked by stars. Not ranked by downloads. Ranked by measured outcomes.

Requires the Quill registry (coming soon). For now, use quill bench to
measure skills you already have.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			return runSearch(query, model, minScore, verified)
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "Target model")
	cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum combined score")
	cmd.Flags().BoolVar(&verified, "verified", false, "Publisher-signed only")

	return cmd
}

func runSearch(query string, model string, minScore float64, verified bool) error {
	if model == "" {
		model, _ = detectModelFromLock()
	}

	registryURL := registry.BaseURL()

	fmt.Println()
	fmt.Printf("  %s searching for %s on %s\n",
		tui.Diamond.Render(),
		tui.Bold.Render("\""+query+"\""),
		model)
	fmt.Println()
	fmt.Println(tui.FormatWarning("registry not yet connected"))
	fmt.Println()
	fmt.Printf("  Search requires the Quill registry at %s\n", registryURL)
	fmt.Println("  The registry is being built. For now, use quill bench to")
	fmt.Println("  measure skills you already have.")
	fmt.Println()
	fmt.Println(tui.Subtle.Render("  What works today:"))
	fmt.Println(tui.Subtle.Render("    quill bench ./my-skill    benchmark a local skill"))
	fmt.Println(tui.Subtle.Render("    quill status              check installed skills"))
	fmt.Println()
	return nil
}
