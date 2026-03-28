package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	var from string
	var to string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Analyze skill stack for a model switch — see what breaks before you break it",
		Long: `Shows expected pass rate and delta changes when switching models.
Suggests variant swaps where available. Identifies skills with no
eval coverage for the target model.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Printf("  %s analyzing migration from %s to %s\n",
				tui.Diamond.Render(), tui.Bold.Render(from), tui.Bold.Render(to))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — migration analysis in development"))
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Source model")
	cmd.Flags().StringVar(&to, "to", "", "Target model")

	return cmd
}
