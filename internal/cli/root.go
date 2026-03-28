package cli

import (
	"fmt"
	"os"

	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

var (
	formatFlag  string
	quietFlag   bool
	verboseFlag bool
)

func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "quill",
		Short: "Outcome intelligence for AI agent skills",
		Long:  "Quill — the credibility layer for AI agent skills.\nMakes agent behavior measurable, improvable, and defensible.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(tui.RenderHelp(version))
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&formatFlag, "format", "", "Output format: text, json (default: auto-detect)")
	root.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "Minimal output")
	root.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Detailed output")

	root.AddCommand(
		newInitCmd(),
		newStatusCmd(),
		newAddCmd(),
		newSearchCmd(),
		newBenchCmd(),
		newFixCmd(),
		newUpgradeCmd(),
		newExplainCmd(),
		newCompareCmd(),
		newValidateCmd(),
		newLockCmd(),
		newRetireCmd(),
		newPublishCmd(),
		newAuditCmd(),
		newMigrateCmd(),
		newSBOMCmd(),
		newVersionCmd(version),
	)

	return root
}

func isJSON() bool {
	return formatFlag == "json"
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
