package eval

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// GradeResult holds the outcome of grading an eval case.
type GradeResult struct {
	Pass      bool   `json:"pass"`
	Method    string `json:"method"`
	Reasoning string `json:"reasoning,omitempty"`
}

// GradeDeterministic evaluates output using built-in assertion functions.
func GradeDeterministic(output string, expected any, assertion string) *GradeResult {
	pass := false
	reasoning := ""

	switch assertion {
	case "output_contains":
		substr := expectedString(expected)
		pass = strings.Contains(output, substr)
		reasoning = fmt.Sprintf("output %s substring %q", passVerb(pass), substr)

	case "output_not_contains":
		substr := expectedString(expected)
		pass = !strings.Contains(output, substr)
		reasoning = fmt.Sprintf("output %s not contain %q", passVerb(pass), substr)

	case "output_empty_or_declined":
		trimmed := strings.TrimSpace(output)
		pass = trimmed == "" || isDecline(trimmed)
		reasoning = fmt.Sprintf("output is %s", describeEmpty(trimmed))

	case "output_matches_regex":
		pattern := expectedString(expected)
		re, err := regexp.Compile(pattern)
		if err != nil {
			return &GradeResult{Pass: false, Method: "deterministic", Reasoning: fmt.Sprintf("invalid regex: %s", err)}
		}
		pass = re.MatchString(output)
		reasoning = fmt.Sprintf("output %s regex %q", passVerb(pass), pattern)

	case "output_equals":
		expected := expectedString(expected)
		pass = strings.TrimSpace(output) == strings.TrimSpace(expected)
		reasoning = fmt.Sprintf("output %s expected", passVerb(pass))

	case "output_matches_json_schema":
		pass = json.Valid([]byte(output))
		reasoning = fmt.Sprintf("output is %s JSON", passVerb(pass))

	case "output_category_matches":
		cat := expectedString(expected)
		pass = strings.Contains(strings.ToLower(output), strings.ToLower(cat))
		reasoning = fmt.Sprintf("output %s category %q", passVerb(pass), cat)

	default:
		return &GradeResult{Pass: false, Method: "deterministic", Reasoning: fmt.Sprintf("unknown assertion: %s", assertion)}
	}

	return &GradeResult{Pass: pass, Method: "deterministic", Reasoning: reasoning}
}

// LLMJudgePrompt builds the prompt for LLM-judge grading.
func LLMJudgePrompt(skillName string, input string, expected any, rubric string, output string) string {
	expectedJSON, _ := json.Marshal(expected)

	return fmt.Sprintf(`You are evaluating an AI agent's output for a specific task.

SKILL BEING TESTED: %s
TASK INPUT: %s
EXPECTED BEHAVIOR: %s
RUBRIC: %s

ACTUAL OUTPUT:
%s

Based on the rubric, does this output PASS or FAIL?

Respond with exactly:
VERDICT: PASS or FAIL
REASONING: <one sentence explaining why>`, skillName, input, string(expectedJSON), rubric, output)
}

// ParseJudgeVerdict extracts PASS/FAIL from an LLM judge response.
func ParseJudgeVerdict(response string) *GradeResult {
	upper := strings.ToUpper(response)

	pass := false
	if strings.Contains(upper, "VERDICT: PASS") || strings.Contains(upper, "VERDICT:PASS") {
		pass = true
	}

	reasoning := ""
	if idx := strings.Index(upper, "REASONING:"); idx >= 0 {
		reasoning = strings.TrimSpace(response[idx+len("REASONING:"):])
		// Take first line only
		if nl := strings.Index(reasoning, "\n"); nl >= 0 {
			reasoning = reasoning[:nl]
		}
	}

	return &GradeResult{Pass: pass, Method: "llm-judge", Reasoning: reasoning}
}

func expectedString(expected any) string {
	switch v := expected.(type) {
	case string:
		return v
	case map[string]any:
		if s, ok := v["value"].(string); ok {
			return s
		}
		if s, ok := v["substring"].(string); ok {
			return s
		}
		if s, ok := v["pattern"].(string); ok {
			return s
		}
		if s, ok := v["category"].(string); ok {
			return s
		}
		b, _ := json.Marshal(v)
		return string(b)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func passVerb(pass bool) string {
	if pass {
		return "does"
	}
	return "does not"
}

func isDecline(s string) bool {
	lower := strings.ToLower(s)
	declines := []string{
		"i can't", "i cannot", "i'm not able", "i am not able",
		"i don't", "i do not", "not within my", "outside my",
		"i'm unable", "sorry", "apolog",
	}
	for _, d := range declines {
		if strings.Contains(lower, d) {
			return true
		}
	}
	return false
}

func describeEmpty(s string) string {
	if s == "" {
		return "empty"
	}
	if isDecline(s) {
		return "a polite decline"
	}
	return "non-empty and not a decline"
}
