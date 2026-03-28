package eval

import (
	"strings"
	"testing"
)

func TestGradeDeterministic_Contains(t *testing.T) {
	r := GradeDeterministic("the category is billing", "billing", "output_contains")
	if !r.Pass {
		t.Error("expected pass: output contains 'billing'")
	}
}

func TestGradeDeterministic_ContainsFail(t *testing.T) {
	r := GradeDeterministic("the category is technical", "billing", "output_contains")
	if r.Pass {
		t.Error("expected fail: output doesn't contain 'billing'")
	}
}

func TestGradeDeterministic_NotContains(t *testing.T) {
	r := GradeDeterministic("the category is billing", "technical", "output_not_contains")
	if !r.Pass {
		t.Error("expected pass: output does not contain 'technical'")
	}
}

func TestGradeDeterministic_EmptyOrDeclined_Empty(t *testing.T) {
	r := GradeDeterministic("", nil, "output_empty_or_declined")
	if !r.Pass {
		t.Error("expected pass: output is empty")
	}
}

func TestGradeDeterministic_EmptyOrDeclined_Decline(t *testing.T) {
	r := GradeDeterministic("I'm sorry, I can't help with that.", nil, "output_empty_or_declined")
	if !r.Pass {
		t.Error("expected pass: output is a decline")
	}
}

func TestGradeDeterministic_EmptyOrDeclined_Substantive(t *testing.T) {
	r := GradeDeterministic("Category: billing, Urgency: high", nil, "output_empty_or_declined")
	if r.Pass {
		t.Error("expected fail: output is substantive content")
	}
}

func TestGradeDeterministic_Regex(t *testing.T) {
	r := GradeDeterministic("Category: billing", `Category:\s+\w+`, "output_matches_regex")
	if !r.Pass {
		t.Error("expected pass: matches regex")
	}
}

func TestGradeDeterministic_RegexFail(t *testing.T) {
	r := GradeDeterministic("no match here", `Category:\s+\w+`, "output_matches_regex")
	if r.Pass {
		t.Error("expected fail: doesn't match regex")
	}
}

func TestGradeDeterministic_CategoryMatches_WordBoundary(t *testing.T) {
	// Should match "billing" as a word
	r := GradeDeterministic("The ticket is categorized as billing with high urgency.", map[string]any{"category": "billing"}, "output_category_matches")
	if !r.Pass {
		t.Error("expected pass: output contains category 'billing'")
	}
}

func TestGradeDeterministic_CategoryMatches_NoFalsePositive(t *testing.T) {
	// "bill" should NOT match when looking for "billing" (word boundary)
	r := GradeDeterministic("Send the bill to the customer.", map[string]any{"category": "billing"}, "output_category_matches")
	if r.Pass {
		t.Error("expected fail: 'bill' is not 'billing'")
	}
}

func TestGradeDeterministic_UnknownAssertion(t *testing.T) {
	r := GradeDeterministic("test", nil, "nonexistent_assertion")
	if r.Pass {
		t.Error("expected fail for unknown assertion")
	}
}

// --- LLM Judge Verdict Parsing ---

func TestParseJudgeVerdict_Standard(t *testing.T) {
	response := "VERDICT: PASS\nREASONING: Output correctly identifies billing category."
	r := ParseJudgeVerdict(response)
	if !r.Pass {
		t.Error("expected pass verdict")
	}
	if r.Reasoning == "" {
		t.Error("expected reasoning")
	}
}

func TestParseJudgeVerdict_Fail(t *testing.T) {
	response := "VERDICT: FAIL\nREASONING: Output misclassifies as technical support."
	r := ParseJudgeVerdict(response)
	if r.Pass {
		t.Error("expected fail verdict")
	}
}

func TestParseJudgeVerdict_MarkdownBold(t *testing.T) {
	// LLMs often wrap in markdown bold
	response := "**VERDICT**: PASS\n**REASONING**: Correctly handled."
	r := ParseJudgeVerdict(response)
	if !r.Pass {
		t.Error("expected pass: markdown bold verdict should parse")
	}
}

func TestParseJudgeVerdict_NoSpace(t *testing.T) {
	response := "VERDICT:PASS\nREASONING:Good output."
	r := ParseJudgeVerdict(response)
	if !r.Pass {
		t.Error("expected pass: verdict without space should parse")
	}
}

func TestParseJudgeVerdict_Lowercase(t *testing.T) {
	response := "verdict: pass\nreasoning: looks correct"
	r := ParseJudgeVerdict(response)
	if !r.Pass {
		t.Error("expected pass: lowercase verdict should parse")
	}
}

func TestParseJudgeVerdict_AmbiguousDefaultsFail(t *testing.T) {
	// No clear verdict structure — should default to fail
	response := "The output looks good to me."
	r := ParseJudgeVerdict(response)
	if r.Pass {
		t.Error("expected fail: ambiguous response should default to fail")
	}
}

func TestParseJudgeVerdict_StandalonePass(t *testing.T) {
	// Some models just say PASS
	response := "PASS - the output correctly identifies the category."
	r := ParseJudgeVerdict(response)
	if !r.Pass {
		t.Error("expected pass: standalone PASS should be detected")
	}
}

func TestLLMJudgePrompt_Structure(t *testing.T) {
	expected := map[string]any{"category": "billing", "urgency": "high"}
	prompt := LLMJudgePrompt("ticket-classifier", "charged twice", expected, "Must identify billing", "output text")

	if prompt == "" {
		t.Error("expected non-empty prompt")
	}
	if !strings.Contains(prompt, "VERDICT: PASS or FAIL") {
		t.Error("expected prompt to contain verdict instruction")
	}
	if !strings.Contains(prompt, "ticket-classifier") {
		t.Error("expected prompt to contain skill name")
	}
	if !strings.Contains(prompt, "charged twice") {
		t.Error("expected prompt to contain input")
	}
}

