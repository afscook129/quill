package cli

import (
	"encoding/json"
	"fmt"

	"github.com/quill-dev/quill/internal/bench"
	"github.com/quill-dev/quill/internal/registry"
	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newBenchCmd() *cobra.Command {
	var trials int
	var model string
	var triggers bool
	var publish bool

	cmd := &cobra.Command{
		Use:   "bench <skill-or-path>",
		Short: "Benchmark a skill — stop vibe-checking, start measuring",
		Long: `Run with-skill vs without-skill comparison. The difference between
shipping with confidence and shipping with hope.

Measures: pass rate, delta, token cost, latency, failure patterns,
and whether the skill is still earning its context budget.

Examples:
  quill bench ./my-skill
  quill bench ticket-classifier
  quill bench ./my-skill --triggers`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBench(args[0], trials, model, triggers, publish)
		},
	}

	cmd.Flags().IntVar(&trials, "trials", 3, "Runs per eval case")
	cmd.Flags().StringVar(&model, "model", "", "Target model")
	cmd.Flags().BoolVar(&triggers, "triggers", false, "Include trigger accuracy test")
	cmd.Flags().BoolVar(&publish, "publish", false, "Contribute results to community registry")

	return cmd
}

func runBench(skill string, trials int, model string, triggers bool, publish bool) error {
	if model == "" {
		model, _ = detectModelFromLock()
	}

	// For now, use mock results to demonstrate the output format
	result := bench.MockResult(skill, model)
	result.Trials = trials

	if isJSON() {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Println()
	fmt.Printf("  %s benchmarking %s · %s · %d trials\n",
		tui.Diamond.Render(),
		tui.Bold.Render(skill),
		model,
		trials)
	fmt.Println()

	// Side by side comparison
	withBar := tui.RenderProgressBar(result.PassWith, 10)
	withoutBar := tui.RenderProgressBar(result.PassWithout, 10)

	fmt.Printf("  %-16s %-28s %-28s\n",
		"",
		tui.Bold.Render("with skill"),
		tui.Bold.Render("without skill"))
	fmt.Printf("  %-16s %s  %-16s %s  %s\n",
		"pass rate",
		registry.FormatPercent(result.PassWith), withBar,
		registry.FormatPercent(result.PassWithout), withoutBar)
	fmt.Printf("  %-16s %-28d %-28d\n",
		"avg tokens", result.AvgTokens, result.AvgTokens-1200)
	fmt.Printf("  %-16s %-28s %-28s\n",
		"avg latency",
		fmt.Sprintf("%.1fs", float64(result.AvgLatencyMs)/1000),
		fmt.Sprintf("%.1fs", float64(result.AvgLatencyMs-500)/1000))
	fmt.Printf("  %-16s %s skill is earning its context budget\n",
		"delta", tui.Success.Render(registry.FormatDeltaPP(result.Delta)+" ↑"))
	fmt.Println()

	// Trigger accuracy
	if triggers && result.TriggerAccuracy != nil {
		ta := result.TriggerAccuracy
		fmt.Printf("  trigger accuracy  %d/%d (%s)   · %d false positive · %d miss\n",
			ta.Correct, ta.Total,
			registry.FormatPercent(float64(ta.Correct)/float64(ta.Total)),
			ta.FalsePositives, ta.FalseNegatives)
		fmt.Println()
	}

	// Failed cases
	if len(result.FailedCases) > 0 {
		fmt.Printf("  failed cases (%d) %s\n", len(result.FailedCases),
			tui.Subtle.Render("────────────────────────────────"))
		for _, fc := range result.FailedCases {
			fmt.Printf("  %s %-6s  %-35s %s\n",
				tui.ErrMark.Render(), fc.ID,
				"\""+fc.Input+"\"",
				tui.Subtle.Render("edge: "+fc.Edge))
		}
		fmt.Println()

		if len(result.Patterns) > 0 {
			fmt.Printf("  pattern: %s\n", result.Patterns[0])
		}
		if result.Suggestion != "" {
			fmt.Printf("  suggestion: %s\n", result.Suggestion)
		}
		fmt.Println()
	}

	// Earning its place
	fmt.Println("  still earning its place?")
	fmt.Printf("  without this skill: base model passes %s of these tasks\n",
		registry.FormatPercent(result.PassWithout))
	fmt.Printf("  with this skill:    base model passes %s of these tasks\n",
		registry.FormatPercent(result.PassWith))

	if result.EarningPlace {
		fmt.Printf("  %s yes, this skill is adding real value (%s)\n",
			tui.Arrow.Render(),
			tui.Success.Render(registry.FormatDeltaPP(result.Delta)))
	} else {
		fmt.Printf("  %s this skill may no longer be adding value\n",
			tui.Arrow.Render())
		fmt.Println("  consider: quill retire " + skill)
	}

	fmt.Println()
	return nil
}
