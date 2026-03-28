package cli

import (
	"encoding/json"
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

Examples:
  quill search "extract action items from meeting transcripts"
  quill search "classify tickets" --model gpt-5.4
  quill search "summarize docs" --verified`,
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
	client := registry.NewMockClient()

	if model == "" {
		model, _ = detectModelFromLock()
	}

	resp, err := client.Search(query, model)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if isJSON() {
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Println()
	fmt.Printf("  %s results for %s on %s\n",
		tui.Diamond.Render(),
		tui.Bold.Render("\""+query+"\""),
		model)
	fmt.Println()

	// Table header
	fmt.Printf("  %-3s %-30s %-10s %7s %6s %6s  %s\n",
		"",
		tui.Subtle.Render("skill"),
		tui.Subtle.Render("source"),
		tui.Subtle.Render("delta"),
		tui.Subtle.Render("pass"),
		tui.Subtle.Render("tokens"),
		tui.Subtle.Render(""))

	for i, r := range resp.Results {
		verifiedStr := ""
		if r.Verified {
			verifiedStr = tui.Success.Render("✓")
		}

		icon := " "
		if i == 0 {
			icon = tui.Success.Render("★")
		}

		deltaStr := registry.FormatDeltaPP(r.Delta)
		passStr := registry.FormatPercent(r.PassRate)

		fmt.Printf("  %s  %-30s %-10s %7s %6s %5dt  %s\n",
			icon,
			r.Name+"@"+r.Version,
			r.Source,
			deltaStr,
			passStr,
			r.TokenCost,
			verifiedStr,
		)
	}

	if resp.Recommended != nil {
		fmt.Println()
		fmt.Printf("  %s recommended: %s\n",
			tui.Diamond.Render(),
			tui.Bold.Render(resp.Recommended.Name+"@"+resp.Recommended.Version))
		fmt.Printf("    %s\n", resp.Reasoning)
		fmt.Println()
		fmt.Printf("  install: %s\n",
			tui.Bold.Render("quill add "+resp.Recommended.Name))
	}

	fmt.Println()
	return nil
}
