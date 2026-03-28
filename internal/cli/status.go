package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/quill-dev/quill/internal/config"
	"github.com/quill-dev/quill/internal/discovery"
	"github.com/quill-dev/quill/internal/lock"
	"github.com/quill-dev/quill/internal/manifest"
	"github.com/quill-dev/quill/internal/registry"
	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

type statusOutput struct {
	Model          string        `json:"model"`
	Skills         []skillStatus `json:"skills"`
	TotalContext    int           `json:"total_context_tokens"`
	ContextLimit   int           `json:"context_limit"`
	DriftCount     int           `json:"drift_count"`
	UnbenchedCount int           `json:"unbenchmarked_count"`
}

type skillStatus struct {
	Name       string   `json:"name"`
	Source     string   `json:"source"`
	Delta      string   `json:"delta,omitempty"`
	PassRate   string   `json:"pass_rate,omitempty"`
	Status     string   `json:"status"`
	Note       string   `json:"note,omitempty"`
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Health summary — what's working, what's drifted, what needs attention",
		Long: `Shows the health of your agent's skill stack. For each skill: pass rate,
delta (how much it helps), drift warnings, and whether it's still earning
its context budget.

This is the command that tells you if your agent is credible.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}

	return cmd
}

func runStatus() error {
	// Check for manifest
	if _, err := os.Stat(manifest.FileName); os.IsNotExist(err) {
		fmt.Println(tui.FormatWarning("no quill.manifest.yaml found"))
		fmt.Println()
		fmt.Println("  Run " + tui.Bold.Render("quill init") + " to set up this project.")
		fmt.Println("  30 seconds from zero to configured.")
		fmt.Println()
		return nil
	}

	// Detect model
	model, _ := config.DetectModel()
	if model == "" {
		model = "unknown"
	}

	// Load lock file
	var lf *lock.Lock
	if _, err := os.Stat(lock.FileName); err == nil {
		lf, _ = lock.Load(lock.FileName)
	}

	// Scan for skills on disk
	skills, _ := discovery.Scan(".")

	output := statusOutput{
		Model:        model,
		ContextLimit: 16000,
	}

	if isJSON() {
		output.Skills = buildSkillStatuses(lf, skills)
		if lf != nil {
			for _, s := range lf.Resolved {
				output.TotalContext += s.ContextBudgetTkns
				if s.EvalPassRate == nil {
					output.UnbenchedCount++
				}
			}
			output.DriftCount = len(lf.System.DriftWarnings)
			output.ContextLimit = lf.System.ContextBudgetLimit
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// TUI output
	fmt.Println()
	fmt.Printf("  %s agent system · %s\n", tui.Diamond.Render(), tui.Bold.Render(model))
	fmt.Println()

	if lf != nil && len(lf.Resolved) > 0 {
		totalTokens := 0
		driftCount := 0
		unbenchedCount := 0

		for _, s := range lf.Resolved {
			icon := tui.Check.Render()
			name := s.Skill
			deltaStr := ""
			passStr := ""
			source := s.Source
			note := ""

			if s.EvalPassRateDelta != nil && *s.EvalPassRateDelta > 0 {
				deltaStr = registry.FormatDeltaPP(*s.EvalPassRateDelta)
			}
			if s.EvalPassRate != nil {
				passStr = registry.FormatPercent(*s.EvalPassRate)
			}

			if s.EvalPassRate == nil {
				icon = tui.Pending.Render()
				note = "no bench"
				unbenchedCount++
			} else if s.EvalPassRateDelta != nil && *s.EvalPassRateDelta < 0.08 {
				icon = tui.Pending.Render()
				note = "may no longer be adding value"
			}

			totalTokens += s.ContextBudgetTkns

			age := formatAge(s.InstalledAt)

			fmt.Printf("  %s  %-35s %6s  %4s  %-10s %s  %s\n",
				icon, name, deltaStr, passStr, source, tui.Subtle.Render(age), tui.Subtle.Render(note))

			if note == "may no longer be adding value" {
				fmt.Printf("     %s quill retire %s\n",
					tui.Arrow.Render(), s.Skill)
			}
		}

		// Check for drift in System warnings
		if lf.System.DriftWarnings != nil {
			driftCount = len(lf.System.DriftWarnings)
		}

		fmt.Println()
		fmt.Printf("  context   %d / %d tokens (%d%%)\n",
			totalTokens, lf.System.ContextBudgetLimit,
			totalTokens*100/max(lf.System.ContextBudgetLimit, 1))

		if driftCount > 0 || unbenchedCount > 0 {
			fmt.Println()
			parts := []string{}
			if driftCount > 0 {
				parts = append(parts, fmt.Sprintf("%d drift", driftCount))
			}
			if unbenchedCount > 0 {
				parts = append(parts, fmt.Sprintf("%d unbenchmarked", unbenchedCount))
			}
			summary := ""
			for i, p := range parts {
				if i > 0 {
					summary += " · "
				}
				summary += p
			}
			fmt.Printf("  %s · run %s\n", summary, tui.Bold.Render("quill fix"))
		}
	} else if len(skills) > 0 {
		// Skills found on disk but not in lock file
		for _, s := range skills {
			name := s.Name
			if s.Version != "" {
				name += "@" + s.Version
			}
			fmt.Printf("  %s  %-35s  %s  %s\n",
				tui.Pending.Render(), name, s.Dir, tui.Subtle.Render("not tracked"))
		}
		fmt.Println()
		fmt.Println("  " + tui.Subtle.Render("run quill add <skill> to track and benchmark"))
	} else {
		fmt.Printf("  %s no skills installed yet\n", tui.Pending.Render())
		fmt.Println()
		fmt.Println("  " + tui.Subtle.Render("try: quill add \"describe what you want to build\""))
	}

	fmt.Println()
	return nil
}

func buildSkillStatuses(lf *lock.Lock, skills []discovery.SkillInfo) []skillStatus {
	var out []skillStatus

	if lf != nil {
		for _, s := range lf.Resolved {
			ss := skillStatus{
				Name:   s.Skill,
				Source: s.Source,
				Status: "ok",
			}
			if s.EvalPassRate != nil {
				ss.PassRate = registry.FormatPercent(*s.EvalPassRate)
			}
			if s.EvalPassRateDelta != nil {
				ss.Delta = registry.FormatDeltaPP(*s.EvalPassRateDelta)
			}
			if s.EvalPassRate == nil {
				ss.Status = "unbenchmarked"
			}
			out = append(out, ss)
		}
	}

	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatAge(installedAt string) string {
	if installedAt == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, installedAt)
	if err != nil {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return "just now"
	case d < 24*time.Hour:
		return "today"
	case d < 48*time.Hour:
		return "1d ago"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/24/7))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/24/30))
	}
}
