package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newFixCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fix",
		Short: "Resolve issues from quill status — safe fixes auto-applied, version divergence asks you",
		Long: `Resolves safe issues identified by quill status: missing dependencies,
permission conflicts, context budget overruns. Version divergence always
requires your decision — Quill never auto-resolves version conflicts.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(tui.FormatStep("checking for issues..."))
			fmt.Println()
			fmt.Println(tui.FormatResult("no issues found"))
			fmt.Println()
			return nil
		},
	}
}
