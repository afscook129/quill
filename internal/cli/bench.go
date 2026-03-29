package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/quill-dev/quill/internal/bench"
	"github.com/quill-dev/quill/internal/eval"
	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newBenchCmd() *cobra.Command {
	var trials int
	var model string
	var triggers bool
	var publish bool

	cmd := &cobra.Command{
		Use:   "bench <skill-path>",
		Short: "Benchmark a skill — stop vibe-checking, start measuring",
		Long: `Run with-skill vs without-skill comparison. The core of Quill.

For each eval case, the bench runner makes two API calls:
  WITH skill:    your SKILL.md as system prompt
  WITHOUT skill: no system prompt (baseline)

Both outputs are graded. The delta = how much your skill helps.

Requires:
  1. A skill directory with SKILL.md
  2. Eval cases in evals/evals.json (see spec/EVAL_FORMAT.md)
  3. An API key: ANTHROPIC_API_KEY or OPENAI_API_KEY

Cost: ~$0.50 for smoke (8 cases), ~$1.20 for full (20 cases).

Examples:
  quill bench ./skills/my-skill
  quill bench ./skills/my-skill --trials 5
  quill bench ./skills/my-skill --model gpt-5.4`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBench(args[0], trials, model, triggers, publish)
		},
	}

	cmd.Flags().IntVar(&trials, "trials", 3, "Runs per eval case (default: 3)")
	cmd.Flags().StringVar(&model, "model", "", "Target model (default: from env)")
	cmd.Flags().BoolVar(&triggers, "triggers", false, "Include trigger accuracy test")
	cmd.Flags().BoolVar(&publish, "publish", false, "Contribute results to community registry")

	return cmd
}

func runBench(skill string, trials int, model string, triggers bool, publish bool) error {
	if model == "" {
		model, _ = detectModelFromLock()
	}

	// Validate skill path exists
	if _, err := os.Stat(skill); os.IsNotExist(err) {
		fmt.Println()
		fmt.Println(tui.FormatError(fmt.Sprintf("skill path not found: %s", skill)))
		fmt.Println()
		fmt.Println("  quill bench expects a directory containing SKILL.md and evals/evals.json")
		fmt.Println()
		fmt.Println("  Example directory structure:")
		fmt.Println("    my-skill/")
		fmt.Println("      SKILL.md")
		fmt.Println("      evals/")
		fmt.Println("        evals.json")
		fmt.Println()
		return fmt.Errorf("skill path not found: %s", skill)
	}

	// Check for SKILL.md
	skillMdPath := filepath.Join(skill, "SKILL.md")
	if _, err := os.Stat(skillMdPath); os.IsNotExist(err) {
		fmt.Println()
		fmt.Println(tui.FormatError(fmt.Sprintf("no SKILL.md found in %s", skill)))
		fmt.Println()
		fmt.Println("  Every skill needs a SKILL.md file that defines what it does.")
		fmt.Println("  This file becomes the system prompt for with-skill benchmark runs.")
		fmt.Println()
		return fmt.Errorf("no SKILL.md in %s", skill)
	}

	// Check for API key (needed for evals and bench)
	hasAPIKey := os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("OPENAI_API_KEY") != ""

	// Check for evals — offer to generate if missing
	evalsPath := filepath.Join(skill, "evals", "evals.json")
	if _, err := os.Stat(evalsPath); os.IsNotExist(err) {
		if !hasAPIKey {
			fmt.Println()
			fmt.Println(tui.FormatError(fmt.Sprintf("no evals/evals.json found in %s", skill)))
			fmt.Println()
			fmt.Println("  Set ANTHROPIC_API_KEY or OPENAI_API_KEY to generate starter evals,")
			fmt.Println("  or create evals/evals.json manually (see spec/EVAL_FORMAT.md).")
			fmt.Println()
			return fmt.Errorf("no evals and no API key")
		}

		// Generate starter evals from SKILL.md
		fmt.Println()
		fmt.Println(tui.FormatStep("no evals found — generating 5 starter cases from SKILL.md..."))

		skillContent, err := os.ReadFile(skillMdPath)
		if err != nil {
			return fmt.Errorf("reading SKILL.md: %w", err)
		}

		suite, err := eval.GenerateStarter(string(skillContent), skill, model)
		if err != nil {
			fmt.Println()
			fmt.Println(tui.FormatError("failed to generate starter evals"))
			fmt.Printf("  %s\n", err)
			fmt.Println()
			fmt.Println("  Create evals/evals.json manually. See spec/EVAL_FORMAT.md")
			fmt.Println()
			return err
		}

		fmt.Printf("    generated %d cases [generated] — saved to %s/evals/evals.json\n", len(suite.Cases), skill)
		fmt.Println()
		fmt.Println(tui.FormatWarning("generated evals are flagged [generated] in all output"))
		fmt.Println("  Review and edit them for better coverage. Human-written evals catch more real failures.")
		fmt.Println()
	}
	if !hasAPIKey {
		fmt.Println()
		fmt.Println(tui.FormatError("no API key found"))
		fmt.Println()
		fmt.Println("  Benchmarking calls the model API to compare with-skill vs without-skill.")
		fmt.Println("  Set one of:")
		fmt.Println("    export ANTHROPIC_API_KEY=sk-...")
		fmt.Println("    export OPENAI_API_KEY=sk-...")
		fmt.Println()
		return fmt.Errorf("set ANTHROPIC_API_KEY or OPENAI_API_KEY to run benchmarks")
	}

	// Show cost estimate
	suite, err := eval.Load(skill)
	if err != nil {
		return fmt.Errorf("loading evals: %w", err)
	}
	caseCount := len(suite.Cases)
	estimatedCost := bench.EstimateCost(caseCount, trials, model)

	fmt.Println()
	fmt.Printf("  %s benchmarking %s · %s · %d trials · %d cases\n",
		tui.Diamond.Render(),
		tui.Bold.Render(skill),
		model,
		trials,
		caseCount)
	fmt.Printf("    estimated cost: ~$%.2f (%d API calls)\n",
		estimatedCost, caseCount*trials*2)
	fmt.Println()

	// Run real benchmark
	result, err := bench.Run(skill, bench.Options{
		Model:   model,
		Trials:  trials,
		Publish: publish,
	})
	if err != nil {
		return fmt.Errorf("bench failed: %w", err)
	}

	if isJSON() {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	renderBenchResult(result, triggers, skill)
	return nil
}

func renderBenchResult(result *bench.Result, triggers bool, skill string) {
	// Side by side comparison
	withBar := tui.RenderProgressBar(result.PassWith, 10)
	withoutBar := tui.RenderProgressBar(result.PassWithout, 10)

	fmt.Printf("  %-16s %-28s %-28s\n",
		"",
		tui.Bold.Render("with skill"),
		tui.Bold.Render("without skill"))
	fmt.Printf("  %-16s %s  %-16s %s  %s\n",
		"pass rate",
		tui.FormatPercent(result.PassWith), withBar,
		tui.FormatPercent(result.PassWithout), withoutBar)
	fmt.Printf("  %-16s %-28d\n",
		"avg tokens/call", result.AvgTokens)
	fmt.Printf("  %-16s %-28s\n",
		"avg latency",
		fmt.Sprintf("%.1fs", float64(result.AvgLatencyMs)/1000))

	deltaStr := tui.FormatDeltaPP(result.Delta)
	if result.Delta > 0 {
		fmt.Printf("  %-16s %s skill is earning its context budget\n",
			"delta", tui.Success.Render(deltaStr+" ↑"))
	} else if result.Delta < -0.02 {
		fmt.Printf("  %-16s %s skill is hurting performance\n",
			"delta", tui.Error.Render(deltaStr+" ↓"))
	} else {
		fmt.Printf("  %-16s %s skill has no measurable effect\n",
			"delta", tui.Warning.Render(deltaStr))
	}
	fmt.Println()

	// Trigger accuracy
	if triggers && result.TriggerAccuracy != nil {
		ta := result.TriggerAccuracy
		fmt.Printf("  trigger accuracy  %d/%d (%s)   · %d false positive · %d miss\n",
			ta.Correct, ta.Total,
			tui.FormatPercent(float64(ta.Correct)/float64(ta.Total)),
			ta.FalsePositives, ta.FalseNegatives)
		fmt.Println()
	}

	// Failed cases
	if len(result.FailedCases) > 0 {
		fmt.Printf("  failed cases (%d) %s\n", len(result.FailedCases),
			tui.Subtle.Render("────────────────────────────────"))
		for _, fc := range result.FailedCases {
			fmt.Printf("  %s %-6s  %-40s %s\n",
				tui.ErrMark.Render(), fc.ID,
				"\""+fc.Input+"\"",
				tui.Subtle.Render("edge: "+fc.Edge))
		}
		fmt.Println()

		for _, p := range result.Patterns {
			fmt.Printf("  pattern: %s\n", p)
		}
		if result.Suggestion != "" {
			fmt.Printf("  suggestion: %s\n", result.Suggestion)
		}
		fmt.Println()
	}

	// Earning its place
	fmt.Println("  still earning its place?")
	fmt.Printf("  without this skill: base model passes %s of these tasks\n",
		tui.FormatPercent(result.PassWithout))
	fmt.Printf("  with this skill:    base model passes %s of these tasks\n",
		tui.FormatPercent(result.PassWith))

	if result.EarningPlace {
		fmt.Printf("  %s yes, this skill is adding real value (%s)\n",
			tui.Arrow.Render(),
			tui.Success.Render(tui.FormatDeltaPP(result.Delta)))
	} else if result.Delta < -0.02 {
		fmt.Printf("  %s this skill is making outcomes worse — consider removing it\n",
			tui.Arrow.Render())
	} else {
		fmt.Printf("  %s this skill may no longer be adding value\n",
			tui.Arrow.Render())
		fmt.Println("  consider: quill retire " + skill)
	}

	fmt.Println()
}
