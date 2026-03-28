package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newBenchCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bench-compare <skill-v1> <skill-v2>",
		Short: "A/B between two versions of the same skill — blind grading",
		Long: `Run blind A/B comparison between two versions of the same skill.
Neither version is labeled during eval. Use when iterating on a
skill to confirm your changes actually improved outcomes.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Printf("  %s comparing %s vs %s\n",
				tui.Diamond.Render(),
				tui.Bold.Render(args[0]),
				tui.Bold.Render(args[1]))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — requires benchmark runner"))
			fmt.Println()
			return nil
		},
	}
}
