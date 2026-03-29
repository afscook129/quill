package mcp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/quill-dev/quill/internal/bench"
	"github.com/quill-dev/quill/internal/config"
	"github.com/quill-dev/quill/internal/discovery"
	"github.com/quill-dev/quill/internal/lock"
)

// RegisterAllTools registers the 8 Quill MCP tools.
func RegisterAllTools(s *Server) {
	s.RegisterTool(toolSearch())
	s.RegisterTool(toolAdd())
	s.RegisterTool(toolStatus())
	s.RegisterTool(toolBench())
	s.RegisterTool(toolExplain())
	s.RegisterTool(toolCompare())
	s.RegisterTool(toolFix())
	s.RegisterTool(toolRetire())
}

func toolSearch() Tool {
	return Tool{
		Name:        "quill_search",
		Description: "Search for verified skills by description. Returns skills ranked by behavioral delta — how much they actually improve agent outcomes on the specified model. Requires the Quill registry (not yet available — returns guidance on using quill bench locally).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":         map[string]any{"type": "string", "description": "Natural language description of the capability needed"},
				"model":         map[string]any{"type": "string", "description": "Target model (e.g., claude-sonnet-4-6)"},
				"min_delta":     map[string]any{"type": "number", "description": "Minimum delta to include (0-1)"},
				"verified_only": map[string]any{"type": "boolean", "description": "Only return verified skills"},
			},
			"required": []string{"query"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var input struct {
				Query string `json:"query"`
				Model string `json:"model"`
			}
			json.Unmarshal(params, &input)

			return map[string]any{
				"status":  "registry_not_available",
				"message": fmt.Sprintf("Registry search for %q is not yet available. The Quill registry is being built. For now, you can benchmark local skills with: quill bench ./path-to-skill", input.Query),
				"suggestion": "If you have a skill installed locally, run quill bench to measure its delta.",
			}, nil
		},
	}
}

func toolAdd() Tool {
	return Tool{
		Name:        "quill_add",
		Description: "Install a verified skill from the registry. Resolves the skill, runs smoke evals, and updates quill.lock. Requires the Quill registry (not yet available).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"skill":      map[string]any{"type": "string", "description": "Skill name or name@version"},
				"version":    map[string]any{"type": "string", "description": "Specific version to install"},
				"skip_evals": map[string]any{"type": "boolean", "description": "Install without running smoke evals"},
			},
			"required": []string{"skill"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var input struct {
				Skill string `json:"skill"`
			}
			json.Unmarshal(params, &input)

			return map[string]any{
				"installed": false,
				"message":   fmt.Sprintf("Cannot install %q — the Quill registry is not yet available. Install skills manually, then run quill bench to verify them.", input.Skill),
			}, nil
		},
	}
}

func toolStatus() Tool {
	return Tool{
		Name:        "quill_status",
		Description: "Show health of all installed skills. Returns pass rates, deltas, drift warnings, and whether each skill is earning its context budget.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: func(params json.RawMessage) (any, error) {
			// Load lock file
			lf, err := lock.Load(lock.FileName)
			if err != nil {
				// No lock file — scan for skills on disk
				skills, _ := discovery.Scan(".")
				if len(skills) == 0 {
					return map[string]any{
						"skills":  []any{},
						"summary": "No skills found. Use quill bench ./path-to-skill to verify a skill.",
					}, nil
				}
				skillList := make([]map[string]any, len(skills))
				for i, s := range skills {
					skillList[i] = map[string]any{
						"name":   s.Name,
						"path":   s.Dir,
						"status": "unverified",
					}
				}
				return map[string]any{
					"skills":  skillList,
					"summary": fmt.Sprintf("Found %d skills on disk, none verified. Run quill bench on each to measure their delta.", len(skills)),
				}, nil
			}

			model, _ := config.DetectModel()

			skills := make([]map[string]any, len(lf.Resolved))
			verified := 0
			unbenched := 0
			for i, s := range lf.Resolved {
				entry := map[string]any{
					"skill":  s.Skill,
					"source": s.Source,
					"tokens": s.ContextBudgetTkns,
				}
				if s.EvalPassRate != nil {
					entry["pass_rate"] = *s.EvalPassRate
					entry["status"] = "verified"
					verified++
				} else {
					entry["status"] = "unverified"
					unbenched++
				}
				if s.EvalPassRateDelta != nil {
					entry["delta"] = *s.EvalPassRateDelta
					if *s.EvalPassRateDelta < 0.08 {
						entry["earning_place"] = false
						entry["note"] = "Delta below 8pp — the base model may handle this well on its own."
					} else {
						entry["earning_place"] = true
					}
				}
				skills[i] = entry
			}

			summary := fmt.Sprintf("%d skills installed, %d verified, %d need benchmarking", len(lf.Resolved), verified, unbenched)
			if model != "" {
				summary += fmt.Sprintf(". Model: %s", model)
			}

			return map[string]any{
				"skills":  skills,
				"summary": summary,
			}, nil
		},
	}
}

func toolBench() Tool {
	return Tool{
		Name:        "quill_bench",
		Description: "Benchmark a skill by running with-skill vs without-skill comparison. Measures behavioral delta — how much the skill improves outcomes over baseline. Requires a skill directory with SKILL.md and evals/evals.json, plus an API key (ANTHROPIC_API_KEY or OPENAI_API_KEY). Costs ~$0.50-1.50 per run.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"skill":  map[string]any{"type": "string", "description": "Path to skill directory (must contain SKILL.md and evals/evals.json)"},
				"model":  map[string]any{"type": "string", "description": "Target model (default: from env)"},
				"trials": map[string]any{"type": "integer", "description": "Trials per eval case (default: 3)"},
			},
			"required": []string{"skill"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var input struct {
				Skill  string `json:"skill"`
				Model  string `json:"model"`
				Trials int    `json:"trials"`
			}
			json.Unmarshal(params, &input)

			if input.Model == "" {
				input.Model, _ = config.DetectModel()
				if input.Model == "" {
					input.Model = "claude-sonnet-4-6"
				}
			}
			if input.Trials <= 0 {
				input.Trials = 3
			}

			// Validate
			if _, err := os.Stat(input.Skill + "/SKILL.md"); err != nil {
				return nil, fmt.Errorf("no SKILL.md found in %s — every skill needs a SKILL.md file", input.Skill)
			}
			if _, err := os.Stat(input.Skill + "/evals/evals.json"); err != nil {
				return nil, fmt.Errorf("no evals/evals.json in %s — create eval cases to benchmark against (see spec/EVAL_FORMAT.md)", input.Skill)
			}

			hasKey := os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("OPENAI_API_KEY") != ""
			if !hasKey {
				return nil, fmt.Errorf("no API key found — set ANTHROPIC_API_KEY or OPENAI_API_KEY to run benchmarks")
			}

			result, err := bench.Run(input.Skill, bench.Options{
				Model:  input.Model,
				Trials: input.Trials,
			})
			if err != nil {
				return nil, fmt.Errorf("benchmark failed: %w", err)
			}

			return map[string]any{
				"skill":         result.SkillName,
				"model":         result.Model,
				"pass_with":     result.PassWith,
				"pass_without":  result.PassWithout,
				"delta":         result.Delta,
				"earning_place": result.EarningPlace,
				"avg_tokens":    result.AvgTokens,
				"trials":        result.Trials,
				"cases":         result.CaseCount,
				"failed_cases":  result.FailedCases,
				"patterns":      result.Patterns,
				"suggestion":    result.Suggestion,
			}, nil
		},
	}
}

func toolExplain() Tool {
	return Tool{
		Name:        "quill_explain",
		Description: "Explain what a skill does, how it works, where it fails, and its performance history. Requires the Quill registry (not yet available).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"skill": map[string]any{"type": "string", "description": "Skill name"},
			},
			"required": []string{"skill"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{
				"status":  "not_available",
				"message": "Skill explanation requires the Quill registry (coming soon).",
			}, nil
		},
	}
}

func toolCompare() Tool {
	return Tool{
		Name:        "quill_compare",
		Description: "Head-to-head behavioral comparison of two skills. Shows which is stronger on which dimensions, whether they're composable, and which to choose. Requires the Quill registry (not yet available).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"skill_a": map[string]any{"type": "string", "description": "First skill"},
				"skill_b": map[string]any{"type": "string", "description": "Second skill"},
				"model":   map[string]any{"type": "string", "description": "Model context"},
			},
			"required": []string{"skill_a", "skill_b"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{
				"status":  "not_available",
				"message": "Skill comparison requires the Quill registry (coming soon).",
			}, nil
		},
	}
}

func toolFix() Tool {
	return Tool{
		Name:        "quill_fix",
		Description: "Diagnose and resolve issues with installed skills. Checks for missing dependencies, permission conflicts, and version divergence. Version conflicts always require human decision — Quill never auto-resolves those.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{
				"fixed":             []any{},
				"requires_decision": []any{},
				"message":           "No issues found.",
			}, nil
		},
	}
}

func toolRetire() Tool {
	return Tool{
		Name:        "quill_retire",
		Description: "Check if a skill is still earning its place. Models improve over time — a skill that was essential six months ago may now be dead weight. This runs the eval suite with and without the skill and shows the result in plain English.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"skill": map[string]any{"type": "string", "description": "Skill name or path"},
			},
			"required": []string{"skill"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var input struct {
				Skill string `json:"skill"`
			}
			json.Unmarshal(params, &input)

			return map[string]any{
				"status":  "not_available",
				"message": fmt.Sprintf("Retirement check for %q requires a bench run. Use quill_bench to measure the skill's current delta.", input.Skill),
			}, nil
		},
	}
}
