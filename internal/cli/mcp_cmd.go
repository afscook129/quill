package cli

import (
	"fmt"

	"github.com/quill-dev/quill/internal/mcp"
	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

var mcpVersion = "dev"

func SetMCPVersion(v string) {
	mcpVersion = v
}

func newMCPCmd() *cobra.Command {
	var serve bool

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server — the agent-facing interface to Quill",
		Long: `Start the Quill MCP server. Agents in Claude Code, Cursor, VS Code,
and any MCP-compatible harness can call Quill tools as structured tool calls.

Tools available:
  quill_search   Search for verified skills (needs registry)
  quill_add      Install a skill (needs registry)
  quill_status   Show installed skill health
  quill_bench    Benchmark a skill (with/without delta)
  quill_explain  Explain a skill's behavior (needs registry)
  quill_compare  Compare two skills (needs registry)
  quill_fix      Diagnose and fix issues
  quill_retire   Check if a skill is still earning its place

Usage:
  quill mcp --serve              Start stdio MCP server

Configure in .mcp.json:
  {"mcpServers": {"quill": {"command": "quill", "args": ["mcp", "--serve"]}}}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !serve {
				fmt.Println()
				fmt.Println(tui.FormatStep("Quill MCP server"))
				fmt.Println()
				fmt.Println("  Start with: quill mcp --serve")
				fmt.Println()
				fmt.Println("  Or add to your harness MCP config:")
				fmt.Println(`  {"mcpServers": {"quill": {"command": "quill", "args": ["mcp", "--serve"]}}}`)
				fmt.Println()
				return nil
			}

			s := mcp.New(mcpVersion)
			mcp.RegisterAllTools(s)
			return s.Serve()
		},
	}

	cmd.Flags().BoolVar(&serve, "serve", false, "Start the MCP server (stdio transport)")

	return cmd
}
