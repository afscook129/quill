package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newSBOMCmd() *cobra.Command {
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "sbom",
		Short: "Security surface inventory — SPDX 2.3 format, one command",
		Long: `Complete security surface inventory covering all installed skills
from all sources. SPDX 2.3 format. One command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(tui.FormatStep("generating SBOM..."))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — SBOM generation in development"))
			fmt.Println()
			_ = format
			_ = output
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "spdx-2.3", "Output format")
	cmd.Flags().StringVar(&output, "output", "", "Output file path")

	return cmd
}
