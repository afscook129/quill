package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newBenchHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bench-history <skill>",
		Short: "Benchmark results over time — pass rate and delta across model versions",
		Long: `Shows how a skill's benchmark results have changed over time.
Pass rate and delta changes across model versions. Whether the skill
is still earning its place over time. Use to track improvement
from your own skill iterations.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Printf("  %s benchmark history for %s\n",
				tui.Diamond.Render(), tui.Bold.Render(args[0]))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — requires registry connection"))
			fmt.Println()
			return nil
		},
	}
}
