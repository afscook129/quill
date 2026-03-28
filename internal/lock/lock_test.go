package lock

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseLock(t *testing.T) {
	input := `
meta:
  quill_version: 1.0.0
  generated_at: "2026-03-23T14:22:00Z"
  model_family: claude-3
  model_version: claude-sonnet-4.6-20260301
resolved:
  - skill: ticket-classifier@2.1.0
    source: skills.sh
    installed_via: quill-add
    signature_verified: true
    variant_selected: claude
    model_tested: claude-sonnet-4.6-20260301
    eval_pass_rate: 0.910
    eval_pass_rate_delta: 0.380
    eval_generated: false
    context_budget_tokens: 1240
system:
  total_context_budget_tokens: 1240
  context_budget_limit: 16000
  signals_enabled: true
  signals_scope: anonymous_aggregate
`
	l, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if l.Meta.QuillVersion != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", l.Meta.QuillVersion)
	}

	if len(l.Resolved) != 1 {
		t.Fatalf("expected 1 resolved skill, got %d", len(l.Resolved))
	}

	s := l.Resolved[0]
	if s.Skill != "ticket-classifier@2.1.0" {
		t.Errorf("expected skill 'ticket-classifier@2.1.0', got %q", s.Skill)
	}
	if s.EvalPassRate == nil || *s.EvalPassRate != 0.910 {
		t.Errorf("expected pass rate 0.910")
	}
	if !s.SignatureVerified {
		t.Error("expected signature verified")
	}
}

func TestRoundtrip(t *testing.T) {
	l := New("1.0.0")
	rate := 0.91
	delta := 0.38
	gen := false
	l.Resolved = append(l.Resolved, ResolvedSkill{
		Skill:             "test@1.0.0",
		Source:            "local",
		InstalledVia:      "quill-add",
		EvalPassRate:      &rate,
		EvalPassRateDelta: &delta,
		EvalGenerated:     &gen,
		ContextBudgetTkns: 500,
	})

	var buf bytes.Buffer
	if err := Write(&buf, l); err != nil {
		t.Fatal(err)
	}

	l2, err := Parse(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if len(l2.Resolved) != 1 {
		t.Fatalf("expected 1 resolved, got %d", len(l2.Resolved))
	}
	if l2.Resolved[0].Skill != "test@1.0.0" {
		t.Errorf("skill mismatch: %q", l2.Resolved[0].Skill)
	}
}

func TestNewLock(t *testing.T) {
	l := New("1.0.0")

	if l.Meta.QuillVersion != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", l.Meta.QuillVersion)
	}
	if l.System.ContextBudgetLimit != 200000 {
		t.Errorf("expected budget 200000, got %d", l.System.ContextBudgetLimit)
	}
	if !l.System.SignalsEnabled {
		t.Error("expected signals enabled by default")
	}
}
