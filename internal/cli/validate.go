package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var full bool
	var publish bool
	var strict bool

	cmd := &cobra.Command{
		Use:   "validate <skill>",
		Short: "Full eval suite — for CI on merge",
		Long: `Runs the complete eval suite against a skill. Use in CI pipelines:
quill validate --full --strict --publish

--strict exits 1 on any warning (use as CI gate).
--publish contributes results to the community registry.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Printf("  %s validating %s\n", tui.Diamond.Render(), tui.Bold.Render(args[0]))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — eval runner in development"))
			fmt.Println()
			_ = full
			_ = publish
			_ = strict
			return nil
		},
	}

	cmd.Flags().BoolVar(&full, "full", false, "Complete suite, not just smoke tests")
	cmd.Flags().BoolVar(&publish, "publish", false, "Contribute results to community registry")
	cmd.Flags().BoolVar(&strict, "strict", false, "Exit 1 on any warning")

	return cmd
}
