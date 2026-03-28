package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Quill version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("quill v%s\n", version)
		},
	}
}
