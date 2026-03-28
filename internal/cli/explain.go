package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain <skill[@version]>",
		Short: "What this skill does, how it works, where it fails — in plain English",
		Long: `Deep dive into a skill: technique, strengths, weaknesses, performance
history across model versions, what it's commonly used alongside,
and whether it's still earning its context budget.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Printf("  %s explain for %s\n", tui.Diamond.Render(), tui.Bold.Render(args[0]))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — requires registry connection"))
			fmt.Println()
			return nil
		},
	}
}
