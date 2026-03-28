package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compare <skill-a> <skill-b>",
		Short: "Head-to-head behavioral comparison — which is stronger and why",
		Long: `Side-by-side comparison of two skills: which is stronger on what
dimensions, whether they're composable, and which to choose for
your use case and model.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Printf("  %s comparing %s vs %s\n",
				tui.Diamond.Render(),
				tui.Bold.Render(args[0]),
				tui.Bold.Render(args[1]))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — requires registry connection"))
			fmt.Println()
			return nil
		},
	}
}
