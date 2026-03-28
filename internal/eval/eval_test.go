package eval

import (
	"strings"
	"testing"
)

func TestParseSuite(t *testing.T) {
	input := `{
  "version": "1.0",
  "skill": "ticket-classifier",
  "skill_version": "2.1.0",
  "description": "Test suite for ticket classification",
  "cases": [
    {
      "id": "tc-01",
      "description": "Standard billing complaint",
      "input": {
        "prompt": "My account was charged twice for order #4521"
      },
      "expected": {
        "category": "billing",
        "urgency": "high"
      },
      "grading": {
        "method": "llm-judge",
        "rubric": "Must identify as billing category with high urgency."
      },
      "tags": ["positive", "standard"],
      "generated": false
    },
    {
      "id": "tc-20",
      "description": "Should not trigger - off-topic",
      "input": {
        "prompt": "What's the weather like today?"
      },
      "expected": {
        "should_trigger": false
      },
      "grading": {
        "method": "deterministic",
        "assertion": "output_empty_or_declined"
      },
      "tags": ["negative", "trigger-test"],
      "generated": false
    }
  ]
}`

	s, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if s.Skill != "ticket-classifier" {
		t.Errorf("expected skill 'ticket-classifier', got %q", s.Skill)
	}
	if len(s.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(s.Cases))
	}

	c1 := s.Cases[0]
	if c1.ID != "tc-01" {
		t.Errorf("expected id 'tc-01', got %q", c1.ID)
	}
	if c1.Grading.Method != "llm-judge" {
		t.Errorf("expected method 'llm-judge', got %q", c1.Grading.Method)
	}
	if !c1.HasTag("positive") {
		t.Error("expected case to have tag 'positive'")
	}

	c2 := s.Cases[1]
	if c2.Grading.Method != "deterministic" {
		t.Errorf("expected method 'deterministic', got %q", c2.Grading.Method)
	}
	if c2.Grading.Assertion != "output_empty_or_declined" {
		t.Errorf("expected assertion 'output_empty_or_declined', got %q", c2.Grading.Assertion)
	}
}

func TestParseEmpty(t *testing.T) {
	input := `{"version": "1.0", "skill": "test", "cases": []}`
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for empty cases")
	}
}
