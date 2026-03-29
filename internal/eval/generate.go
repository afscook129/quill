package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/quill-dev/quill/internal/provider"
)

const generatePrompt = `You are generating eval test cases for an AI agent skill.

SKILL CONTENT:
%s

Generate exactly 5 eval cases in JSON format for testing this skill.
Include:
- 3 positive cases (skill should trigger and produce correct output)
- 1 implicit case (skill should trigger without explicit request)
- 1 negative case (skill should NOT trigger — off-topic input)

For positive and implicit cases, use "llm-judge" grading with a clear rubric.
For the negative case, use "deterministic" grading with assertion "output_empty_or_declined".

Respond with ONLY valid JSON matching this exact schema:
{
  "version": "1.0",
  "skill": "<infer skill name from content>",
  "skill_version": "0.1.0",
  "description": "<one sentence describing what these evals test>",
  "cases": [
    {
      "id": "gen-01",
      "description": "<what this case tests>",
      "input": {"prompt": "<the user message>"},
      "expected": {},
      "grading": {"method": "llm-judge", "rubric": "<clear pass/fail criteria>"},
      "tags": ["positive", "standard"],
      "generated": true
    }
  ]
}

Use realistic, diverse inputs. Make rubrics specific and measurable.
IDs must be gen-01 through gen-05. All cases must have "generated": true.`

// GenerateStarter creates 5 starter eval cases from a SKILL.md using an LLM.
func GenerateStarter(skillContent string, skillDir string, model string) (*Suite, error) {
	p, err := provider.ForModel(model)
	if err != nil {
		return nil, fmt.Errorf("generating evals: %w", err)
	}

	prompt := fmt.Sprintf(generatePrompt, skillContent)

	resp, err := p.Call(model, "", []provider.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("calling LLM for eval generation: %w", err)
	}

	// Clean output (strip markdown code fences)
	content := cleanJSON(resp.Content)

	var suite Suite
	if err := json.Unmarshal([]byte(content), &suite); err != nil {
		return nil, fmt.Errorf("parsing generated evals (model returned invalid JSON): %w\n\nRaw output:\n%s", err, resp.Content[:min(len(resp.Content), 500)])
	}

	// Ensure all cases are marked as generated
	for i := range suite.Cases {
		suite.Cases[i].Generated = true
	}

	if len(suite.Cases) == 0 {
		return nil, fmt.Errorf("model generated 0 eval cases — try again or write evals manually")
	}

	// Write to disk
	evalsDir := filepath.Join(skillDir, "evals")
	if err := os.MkdirAll(evalsDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating evals directory: %w", err)
	}

	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling evals: %w", err)
	}

	path := filepath.Join(evalsDir, FileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", path, err)
	}

	return &suite, nil
}

func cleanJSON(s string) string {
	// Strip markdown code fences
	if len(s) > 6 && s[:3] == "```" {
		lines := splitLines(s)
		if len(lines) >= 3 {
			// Find closing fence
			end := len(lines) - 1
			for end > 0 && len(lines[end]) < 3 {
				end--
			}
			if end > 0 && len(lines[end]) >= 3 && lines[end][:3] == "```" {
				s = joinLines(lines[1:end])
			}
		}
	}
	return s
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
