package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newTeamCmd() *cobra.Command {
	team := &cobra.Command{
		Use:   "team",
		Short: "Team coordination — git-native, no telemetry",
		Long:  `Team commands for coordinating skill stacks across developers. Purely git-native.`,
	}

	team.AddCommand(newTeamStatusCmd())

	return team
}

func newTeamStatusCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Branch divergence — compares quill.lock on main vs every open branch",
		Long: `Compares the committed quill.lock on main against quill.lock on every
open branch and PR. Purely git-native. No telemetry. No developer tracking.
Requires GITHUB_TOKEN or GITLAB_TOKEN env var.

"Divergence" means a branch has a different lock file than main.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			fmt.Println(tui.FormatStep("comparing quill.lock across branches..."))
			fmt.Println()
			fmt.Println(tui.Subtle.Render("  coming soon — requires GITHUB_TOKEN or GITLAB_TOKEN"))
			fmt.Println()
			_ = repo
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Specify repo if not auto-detected")

	return cmd
}
