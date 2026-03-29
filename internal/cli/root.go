package cli

import (
	"fmt"
	"os"

	"github.com/quill-dev/quill/internal/config"
	"github.com/quill-dev/quill/internal/lock"
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

	SetMCPVersion(version)

	root.AddCommand(
		newInitCmd(),
		newStatusCmd(),
		newAddCmd(),
		newSearchCmd(),
		newBenchCmd(),
		newBenchHistoryCmd(),
		newBenchCompareCmd(),
		newMCPCmd(),
		newFixCmd(),
		newUpgradeCmd(),
		newExplainCmd(),
		newCompareCmd(),
		newValidateCmd(),
		newLockCmd(),
		newRetireCmd(),
		newTeamCmd(),
		newPublishCmd(),
		newRegistryCmd(),
		newAuditCmd(),
		newMigrateCmd(),
		newSBOMCmd(),
		newVersionCmd(version),
	)

	return root
}

func isJSON() bool {
	if formatFlag == "json" {
		return true
	}
	if formatFlag == "text" {
		return false
	}
	// Auto-detect: non-TTY defaults to JSON for piping
	if formatFlag == "" && !isTTY() {
		return true
	}
	return false
}

func forceText() {
	formatFlag = "text"
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func detectModelFromLock() (string, string) {
	if _, err := os.Stat(lock.FileName); err == nil {
		lf, err := lock.Load(lock.FileName)
		if err == nil && lf.Meta.ModelVersion != "" {
			return lf.Meta.ModelVersion, "quill.lock"
		}
	}
	model, source := config.DetectModel()
	if model != "" {
		return model, source
	}
	return "claude-sonnet-4-6", "default"
}
