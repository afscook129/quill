package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newAuditCmd() *cobra.Command {
	var strict bool

	cmd := &cobra.Command{
		Use:   "audit [path]",
		Short: "Full security scan — permissions, signatures, integrity",
		Long: `Complete security surface scan across all installed skills.
Works on skills from any source, with or without manifests.
--strict exits 1 on any warning (use as CI gate).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(tui.FormatStep("auditing skill stack..."))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — security audit in development"))
			fmt.Println()
			_ = strict
			return nil
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "Exit 1 on any warning")

	return cmd
}
