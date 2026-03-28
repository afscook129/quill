package eval

import "testing"

func TestGradeDeterministic_Contains(t *testing.T) {
	r := GradeDeterministic("the category is billing", "billing", "output_contains")
	if !r.Pass {
		t.Error("expected pass: output contains 'billing'")
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

func TestGradeDeterministic_EmptyOrDeclined_NotEmpty(t *testing.T) {
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

func TestGradeDeterministic_CategoryMatches(t *testing.T) {
	expected := map[string]any{"category": "billing"}
	r := GradeDeterministic("The ticket is categorized as Billing with high urgency.", expected, "output_category_matches")
	if !r.Pass {
		t.Error("expected pass: output contains category 'billing'")
	}
}

func TestParseJudgeVerdict_Pass(t *testing.T) {
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

func TestLLMJudgePrompt(t *testing.T) {
	expected := map[string]any{"category": "billing", "urgency": "high"}
	prompt := LLMJudgePrompt("ticket-classifier", "charged twice", expected, "Must identify billing", "output text")
	if prompt == "" {
		t.Error("expected non-empty prompt")
	}
	if !contains(prompt, "VERDICT: PASS or FAIL") {
		t.Error("expected prompt to contain verdict instruction")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr)
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
