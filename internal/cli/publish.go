package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newPublishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "publish",
		Short: "Publish to Quill registry — Sigstore signed, eval-verified",
		Long: `Publish a skill to the community or private registry.
Requires: manifest, Sigstore signing, passing smoke evals.
Validates output schema semver rules at publish time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(tui.FormatStep("preparing to publish..."))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — registry publishing in development"))
			fmt.Println()
			return nil
		},
	}
}
