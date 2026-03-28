package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/quill-dev/quill/internal/config"
	"github.com/quill-dev/quill/internal/discovery"
	"github.com/quill-dev/quill/internal/lock"
	"github.com/quill-dev/quill/internal/manifest"
	"github.com/quill-dev/quill/internal/tui"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force bool
	var harness string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up a project — 30 seconds from zero to configured",
		Long: `Initialize Quill for this project. Auto-detects installed harnesses,
generates configs, creates quill.manifest.yaml, and scans for existing skills.

One command. Everything configured.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(force, harness)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing configuration")
	cmd.Flags().StringVar(&harness, "harness", "", "Target specific harness (or 'all')")

	return cmd
}

func runInit(force bool, harness string) error {
	// Check for existing manifest
	if _, err := os.Stat(manifest.FileName); err == nil && !force {
		fmt.Println(tui.FormatWarning("quill.manifest.yaml already exists"))
		fmt.Println("  use --force to overwrite")
		return nil
	}

	projectName := filepath.Base(mustCwd())

	fmt.Println(tui.FormatStep("initializing quill..."))
	fmt.Println()

	// Detect harnesses
	fmt.Println(tui.FormatStep("detecting harnesses..."))
	harnesses := detectHarnesses()
	for _, h := range harnesses {
		fmt.Println(tui.FormatSubStep(h.Name + " " + tui.Success.Render("✓")))
	}
	if len(harnesses) == 0 {
		fmt.Println(tui.FormatSubStep("none detected (you can configure manually)"))
	}
	fmt.Println()

	// Generate harness configs
	for _, h := range harnesses {
		configPath := h.ConfigPath
		fmt.Printf("  %s generating %s... ", tui.Diamond.Render(), configPath)
		if err := writeHarnessConfig(h); err != nil {
			fmt.Println(tui.Error.Render("failed"))
			fmt.Printf("    %s\n", err)
		} else {
			fmt.Println(tui.Success.Render("done"))
		}
	}

	// Generate manifest
	fmt.Printf("  %s generating quill.manifest.yaml... ", tui.Diamond.Render())
	m := manifest.Default(projectName)
	if err := manifest.Save(manifest.FileName, m); err != nil {
		fmt.Println(tui.Error.Render("failed"))
		return fmt.Errorf("creating manifest: %w", err)
	}
	fmt.Println(tui.Success.Render("done"))

	// Create .quill directory
	if err := os.MkdirAll(config.ProjectDir(), 0o755); err != nil {
		return fmt.Errorf("creating .quill directory: %w", err)
	}

	// Create lock file
	l := lock.New("1.0.0")
	model, source := config.DetectModel()
	if model != "" {
		l.Meta.ModelVersion = model
		fmt.Printf("  %s detected model: %s (from %s)\n", tui.Diamond.Render(), model, source)
	}
	if err := lock.Save(lock.FileName, l); err != nil {
		return fmt.Errorf("creating lock file: %w", err)
	}

	// Scan for existing skills
	fmt.Println()
	fmt.Println(tui.FormatStep("scanning for existing skills..."))
	skills, _ := discovery.Scan(".")
	if len(skills) > 0 {
		fmt.Printf("  found %d skill(s):\n", len(skills))
		fmt.Print(discovery.FormatSkillList(skills))
	} else {
		fmt.Println(tui.FormatSubStep("no skills found yet"))
	}

	// Signal contribution prompt
	fmt.Println()
	fmt.Println("  Help improve skill recommendations?")
	fmt.Println("  Anonymous signals (pass/fail, latency — never content)")
	fmt.Println("  shared with registry. Default: yes")
	fmt.Println(tui.Subtle.Render("  Change anytime in ~/.quill/config.yaml"))

	// Run status as final step (always TUI, even in non-TTY)
	forceText()
	return runStatus()
}

type harnessInfo struct {
	Name       string
	ConfigPath string
	ConfigType string
}

func detectHarnesses() []harnessInfo {
	var found []harnessInfo

	// Claude Code
	if _, err := os.Stat(".claude"); err == nil {
		found = append(found, harnessInfo{
			Name:       "Claude Code",
			ConfigPath: ".claude/settings.json",
			ConfigType: "claude",
		})
	} else if _, err := os.Stat(".mcp.json"); err == nil {
		found = append(found, harnessInfo{
			Name:       "Claude Code",
			ConfigPath: ".mcp.json",
			ConfigType: "claude-mcp",
		})
	}

	// Cursor
	if _, err := os.Stat(".cursor"); err == nil {
		found = append(found, harnessInfo{
			Name:       "Cursor",
			ConfigPath: ".cursor/mcp.json",
			ConfigType: "cursor",
		})
	}

	// VS Code
	if _, err := os.Stat(".vscode"); err == nil {
		found = append(found, harnessInfo{
			Name:       "VS Code",
			ConfigPath: ".vscode/mcp.json",
			ConfigType: "vscode",
		})
	}

	// Windsurf
	if _, err := os.Stat(".windsurf"); err == nil {
		found = append(found, harnessInfo{
			Name:       "Windsurf",
			ConfigPath: ".windsurf/mcp.json",
			ConfigType: "windsurf",
		})
	}

	// Zed
	home, _ := os.UserHomeDir()
	if home != "" {
		if _, err := os.Stat(filepath.Join(home, ".config", "zed")); err == nil {
			found = append(found, harnessInfo{
				Name:       "Zed",
				ConfigPath: filepath.Join(home, ".config", "zed", "settings.json"),
				ConfigType: "zed",
			})
		}
	}

	// Codex CLI
	if _, err := os.Stat(".codex"); err == nil {
		found = append(found, harnessInfo{
			Name:       "Codex CLI",
			ConfigPath: ".codex/config.yaml",
			ConfigType: "codex",
		})
	}

	// Gemini CLI
	if _, err := os.Stat(".gemini"); err == nil {
		found = append(found, harnessInfo{
			Name:       "Gemini CLI",
			ConfigPath: ".gemini/settings.json",
			ConfigType: "gemini",
		})
	}

	return found
}

func writeHarnessConfig(h harnessInfo) error {
	dir := filepath.Dir(h.ConfigPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var cfg interface{}

	switch h.ConfigType {
	case "claude":
		cfg = map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"quill": map[string]interface{}{
					"command": "quill",
					"args":    []string{"mcp", "--serve"},
					"env": map[string]string{
						"QUILL_REGISTRY": "https://registry.quill.dev",
					},
				},
			},
			"hooks": map[string]interface{}{
				"SessionStart": []map[string]interface{}{
					{
						"hooks": []map[string]interface{}{
							{
								"type":    "command",
								"command": "quill hook session-start --quiet",
								"timeout": 3,
							},
						},
					},
				},
				"PreToolUse": []map[string]interface{}{
					{
						"matcher": ".*",
						"hooks": []map[string]interface{}{
							{
								"type":    "command",
								"command": "quill hook pre-tool --event \"$CLAUDE_HOOK_INPUT\"",
								"timeout": 1,
							},
						},
					},
				},
				"PostToolUse": []map[string]interface{}{
					{
						"matcher": ".*",
						"hooks": []map[string]interface{}{
							{
								"type":    "command",
								"command": "quill hook post-tool --event \"$CLAUDE_HOOK_INPUT\" --async",
								"timeout": 0,
							},
						},
					},
				},
				"Stop": []map[string]interface{}{
					{
						"hooks": []map[string]interface{}{
							{
								"type":    "command",
								"command": "quill hook session-stop --async",
								"timeout": 0,
							},
						},
					},
				},
			},
		}
	case "zed":
		cfg = map[string]interface{}{
			"context_servers": map[string]interface{}{
				"quill": map[string]interface{}{
					"command": map[string]interface{}{
						"path": "quill",
						"args": []string{"mcp", "--serve"},
					},
				},
			},
		}
	case "codex":
		// Codex uses YAML
		yamlContent := "mcpServers:\n  quill:\n    command: quill\n    args: [mcp, --serve]\n"
		return os.WriteFile(h.ConfigPath, []byte(yamlContent), 0o644)
	default:
		// MCP-only config for other harnesses
		cfg = map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"quill": map[string]interface{}{
					"command": "quill",
					"args":    []string{"mcp", "--serve"},
				},
			},
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(h.ConfigPath, append(data, '\n'), 0o644)
}

func mustCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
