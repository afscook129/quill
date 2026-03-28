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
		reasoning = fmt.Sprintf("output %s substring %q", boolVerb(pass, "contains", "missing"), substr)

	case "output_not_contains":
		substr := expectedString(expected)
		pass = !strings.Contains(output, substr)
		reasoning = fmt.Sprintf("output %s %q", boolVerb(pass, "excludes", "contains"), substr)

	case "output_empty_or_declined":
		trimmed := strings.TrimSpace(output)
		pass = trimmed == "" || isDecline(trimmed)
		if pass {
			reasoning = "output is empty or a polite decline"
		} else {
			reasoning = "output is substantive (expected empty or decline)"
		}

	case "output_matches_regex":
		pattern := expectedString(expected)
		re, err := regexp.Compile(pattern)
		if err != nil {
			return &GradeResult{Pass: false, Method: "deterministic", Reasoning: fmt.Sprintf("invalid regex %q: %s", pattern, err)}
		}
		pass = re.MatchString(output)
		reasoning = fmt.Sprintf("output %s regex %q", boolVerb(pass, "matches", "does not match"), pattern)

	case "output_equals":
		exp := expectedString(expected)
		pass = strings.TrimSpace(output) == strings.TrimSpace(exp)
		reasoning = fmt.Sprintf("output %s expected value", boolVerb(pass, "equals", "differs from"))

	case "output_matches_json_schema":
		pass = json.Valid([]byte(output))
		reasoning = fmt.Sprintf("output is %s", boolVerb(pass, "valid JSON", "not valid JSON"))

	case "output_category_matches":
		cat := expectedString(expected)
		// Look for the category as a word boundary, not just substring
		pass = containsWordCI(output, cat)
		reasoning = fmt.Sprintf("output %s category %q", boolVerb(pass, "contains", "missing"), cat)

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
// Handles various formatting: "VERDICT: PASS", "verdict:pass", "**VERDICT**: PASS", etc.
func ParseJudgeVerdict(response string) *GradeResult {
	// Normalize: strip markdown bold, collapse whitespace
	cleaned := strings.ReplaceAll(response, "**", "")
	cleaned = strings.ReplaceAll(cleaned, "__", "")
	upper := strings.ToUpper(cleaned)

	pass := false

	// Try structured patterns first
	verdictPatterns := []string{
		"VERDICT: PASS", "VERDICT:PASS", "VERDICT : PASS",
		"RESULT: PASS", "RESULT:PASS",
	}
	for _, p := range verdictPatterns {
		if strings.Contains(upper, p) {
			pass = true
			break
		}
	}

	// If no PASS found, check if it explicitly says FAIL
	// (absence of both PASS and FAIL means we can't determine — default to fail)
	if !pass {
		hasFail := false
		failPatterns := []string{"VERDICT: FAIL", "VERDICT:FAIL", "VERDICT : FAIL", "RESULT: FAIL"}
		for _, p := range failPatterns {
			if strings.Contains(upper, p) {
				hasFail = true
				break
			}
		}
		if !hasFail {
			// Last resort: look for standalone PASS/FAIL
			if strings.Contains(upper, "PASS") && !strings.Contains(upper, "FAIL") {
				pass = true
			}
		}
	}

	// Extract reasoning
	reasoning := extractReasoning(response)

	return &GradeResult{Pass: pass, Method: "llm-judge", Reasoning: reasoning}
}

func extractReasoning(response string) string {
	// Try "REASONING:" prefix (case-insensitive)
	lower := strings.ToLower(response)
	for _, prefix := range []string{"reasoning:", "reason:", "explanation:"} {
		if idx := strings.Index(lower, prefix); idx >= 0 {
			rest := strings.TrimSpace(response[idx+len(prefix):])
			if nl := strings.Index(rest, "\n"); nl >= 0 {
				rest = rest[:nl]
			}
			return strings.TrimSpace(rest)
		}
	}
	// Fall back to last line if multi-line
	lines := strings.Split(strings.TrimSpace(response), "\n")
	if len(lines) > 1 {
		last := strings.TrimSpace(lines[len(lines)-1])
		if len(last) > 10 && !strings.HasPrefix(strings.ToUpper(last), "VERDICT") {
			return last
		}
	}
	return ""
}

func expectedString(expected any) string {
	switch v := expected.(type) {
	case string:
		return v
	case map[string]any:
		// Try common field names
		for _, key := range []string{"value", "substring", "pattern", "category"} {
			if s, ok := v[key].(string); ok {
				return s
			}
		}
		b, _ := json.Marshal(v)
		return string(b)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func boolVerb(b bool, trueV, falseV string) string {
	if b {
		return trueV
	}
	return falseV
}

// containsWordCI checks if output contains the word (case-insensitive)
// with word-boundary-like matching (not just substring).
func containsWordCI(output, word string) bool {
	lo := strings.ToLower(output)
	lw := strings.ToLower(word)

	idx := 0
	for {
		pos := strings.Index(lo[idx:], lw)
		if pos < 0 {
			return false
		}
		pos += idx

		// Check word boundaries
		before := pos == 0 || !isAlpha(lo[pos-1])
		after := pos+len(lw) >= len(lo) || !isAlpha(lo[pos+len(lw)])

		if before && after {
			return true
		}
		idx = pos + 1
	}
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
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
