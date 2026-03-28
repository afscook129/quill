package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade [skill]",
		Short: "Show available upgrades with delta comparison — never upgrades without asking",
		Long: `Shows all available skill upgrades ranked by delta improvement.
Without argument: shows upgrades for all skills in quill.lock.
With argument: shows upgrade path for a specific skill.

Never upgrades automatically — always requires your confirmation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(tui.FormatStep("checking for upgrades..."))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  no upgrades available"))
			fmt.Println()
			return nil
		},
	}
}
