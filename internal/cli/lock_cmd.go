package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newLockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Generate quill.lock from all installed skills — the behavioral contract",
		Long: `Generates quill.lock covering skills from every source: npx skills,
skillpm, bun, manual install. Non-destructive. Safe to run after
any external skill installation.

The lock file is the behavioral contract for your system.
Commit it to your repo.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(tui.FormatStep("generating quill.lock..."))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — lock generation in development"))
			fmt.Println()
			return nil
		},
	}
}
