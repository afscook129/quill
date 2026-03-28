package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newRetireCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "retire <skill>",
		Short: "Check if a skill is still earning its place — models improve, skills may not need to",
		Long: `Runs the eval suite with and without the skill loaded. Shows the result
in plain English: how much the skill adds, what it costs in context,
whether it's worth keeping.

Models get better. A skill that was essential six months ago may now be
dead weight. This command answers "does this still make my agent better?"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skill := args[0]
			fmt.Println()
			fmt.Printf("  %s checking if %s is still earning its place...\n",
				tui.Diamond.Render(), tui.Bold.Render(skill))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — requires benchmark runner"))
			fmt.Println()
			return nil
		},
	}
}
