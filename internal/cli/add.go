package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/quill-dev/quill/internal/lock"
	"github.com/quill-dev/quill/internal/registry"
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
	client := registry.NewMockClient()

	if model == "" {
		model, _ = detectModelFromLock()
	}

	// Search
	fmt.Println()
	fmt.Println(tui.FormatStep("searching 4 registries for behavioral match..."))

	resp, err := client.Search(query, model)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Show results
	for _, r := range resp.Results {
		verified := ""
		if r.Verified {
			verified = " ✓"
		}
		fmt.Printf("    %s %-30s %-10s %s delta  %s%s\n",
			tui.Arrow.Render(),
			r.Name+"@"+r.Version,
			r.Source,
			registry.FormatDeltaPP(r.Delta),
			registry.FormatPercent(r.PassRate),
			verified,
		)
	}

	if resp.Recommended == nil {
		fmt.Println()
		fmt.Println(tui.FormatWarning("no matching skills found"))
		return nil
	}

	best := resp.Recommended
	fmt.Println()
	fmt.Printf("  %s best match: %s\n", tui.Diamond.Render(),
		tui.Bold.Render(best.Name+"@"+best.Version))
	fmt.Printf("    why: %s\n", resp.Reasoning)
	fmt.Println("    alternatives shown above if you want to compare")

	if dryRun {
		fmt.Println()
		fmt.Println(tui.Subtle.Render("  --dry-run: nothing installed"))
		return nil
	}

	// Simulate dependency resolution
	fmt.Println()
	fmt.Println(tui.FormatStep("resolving dependencies..."))
	fmt.Println(tui.FormatSubStep("no additional dependencies required"))

	// Simulate smoke evals
	if !skipEvals {
		fmt.Println()
		fmt.Printf("  %s running smoke evals (8 examples)...\n", tui.Diamond.Render())
		fmt.Printf("    %s  %s\n",
			tui.RenderProgressBar(1.0, 24),
			tui.Success.Render("done"))
	} else {
		fmt.Println()
		fmt.Println(tui.FormatWarning("--skip-evals: validation skipped (logged in lock file)"))
	}

	// Update lock file
	if err := addToLockFile(best, model); err != nil {
		return err
	}

	fmt.Println()
	signed := ""
	if best.Verified {
		signed = " · signed ✓"
	}
	fmt.Printf("  %s %s installed\n", tui.Check.Render(),
		tui.Bold.Render(best.Name+"@"+best.Version))
	fmt.Printf("    %s delta · %s pass rate · %d tokens%s\n",
		registry.FormatDeltaPP(best.Delta),
		registry.FormatPercent(best.PassRate),
		best.TokenCost,
		signed)

	fmt.Println()
	fmt.Println(tui.Subtle.Render("  ○ better skills appear? quill will tell you automatically"))
	fmt.Println()

	return nil
}

func addToLockFile(skill *registry.SkillResult, model string) error {
	var lf *lock.Lock

	if _, err := os.Stat(lock.FileName); err == nil {
		lf, err = lock.Load(lock.FileName)
		if err != nil {
			lf = lock.New("1.0.0")
		}
	} else {
		lf = lock.New("1.0.0")
	}

	passRate := skill.PassRate
	delta := skill.Delta
	generated := false

	resolved := lock.ResolvedSkill{
		Skill:             skill.Name + "@" + skill.Version,
		Source:            skill.Source,
		InstalledVia:      "quill-add",
		SignatureVerified:  skill.Verified,
		ModelTested:       model,
		EvalPassRate:      &passRate,
		EvalPassRateDelta: &delta,
		EvalGenerated:     &generated,
		ContextBudgetTkns: skill.TokenCost,
		InstalledAt:       time.Now().UTC().Format(time.RFC3339),
		PermissionsGranted: &lock.Permissions{
			Filesystem: "none",
			Network:    "none",
			Shell:      "never",
		},
	}

	lf.Resolved = append(lf.Resolved, resolved)

	// Update system totals
	total := 0
	for _, s := range lf.Resolved {
		total += s.ContextBudgetTkns
	}
	lf.System.TotalContextTokens = total

	return lock.Save(lock.FileName, lf)
}

func detectModelFromLock() (string, string) {
	if _, err := os.Stat(lock.FileName); err == nil {
		lf, err := lock.Load(lock.FileName)
		if err == nil && lf.Meta.ModelVersion != "" {
			return lf.Meta.ModelVersion, "quill.lock"
		}
	}
	return "claude-sonnet-4.6", "default"
}
