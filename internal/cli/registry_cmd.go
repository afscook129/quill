package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newRegistryCmd() *cobra.Command {
	reg := &cobra.Command{
		Use:   "registry",
		Short: "Private registry management",
		Long:  `Commands for managing private Quill registries.`,
	}

	reg.AddCommand(newRegistryInitCmd())

	return reg
}

func newRegistryInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Scaffold a private Cloudflare Workers registry — live in under 10 minutes",
		Long: `Scaffolds the full stack for a private Quill registry:
Cloudflare Workers + D1 + R2 + Vectorize. Including semantic search.
Fork, deploy, done. Full setup in under 10 minutes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(tui.FormatStep("scaffolding private registry..."))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — registry template in development"))
			fmt.Println()
			return nil
		},
	}
}
