package manifest

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseManifest(t *testing.T) {
	input := `
name: ticket-classifier
version: 2.1.0
source: skills.sh
description: Classifies support tickets
permissions:
  filesystem: none
  network: none
  shell: never
context:
  estimated_tokens: 1240
  budget_priority: normal
evals:
  pass_threshold: 0.87
  smoke_count: 8
`
	m, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if m.Name != "ticket-classifier" {
		t.Errorf("expected name 'ticket-classifier', got %q", m.Name)
	}
	if m.Version != "2.1.0" {
		t.Errorf("expected version '2.1.0', got %q", m.Version)
	}
	if m.Permissions.Filesystem != "none" {
		t.Errorf("expected filesystem 'none', got %q", m.Permissions.Filesystem)
	}
	if m.Context.EstimatedTokens != 1240 {
		t.Errorf("expected 1240 tokens, got %d", m.Context.EstimatedTokens)
	}
	if m.Evals.PassThreshold != 0.87 {
		t.Errorf("expected threshold 0.87, got %f", m.Evals.PassThreshold)
	}
}

func TestRoundtrip(t *testing.T) {
	m := Default("test-skill")
	m.Description = "A test skill for roundtrip"

	var buf bytes.Buffer
	if err := Write(&buf, m); err != nil {
		t.Fatal(err)
	}

	m2, err := Parse(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if m.Name != m2.Name {
		t.Errorf("name mismatch: %q vs %q", m.Name, m2.Name)
	}
	if m.Version != m2.Version {
		t.Errorf("version mismatch: %q vs %q", m.Version, m2.Version)
	}
	if m.Description != m2.Description {
		t.Errorf("description mismatch: %q vs %q", m.Description, m2.Description)
	}
}

func TestDefault(t *testing.T) {
	m := Default("my-project")

	if m.Name != "my-project" {
		t.Errorf("expected name 'my-project', got %q", m.Name)
	}
	if m.Version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", m.Version)
	}
	if m.Permissions == nil {
		t.Fatal("expected default permissions")
	}
	if m.Permissions.Shell != "never" {
		t.Errorf("expected shell 'never', got %q", m.Permissions.Shell)
	}
}

func TestParseVariants(t *testing.T) {
	input := `
name: my-skill
version: 1.0.0
variants:
  claude:
    prompt: ./prompts/claude.md
    tested_models: [claude-sonnet-4.6, haiku-4.5]
    pass_rate: 0.91
    pass_rate_delta: 0.38
  openai:
    prompt: ./prompts/openai.md
    tested_models: [gpt-5.4]
    pass_rate: 0.85
`
	m, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(m.Variants))
	}

	claude := m.Variants["claude"]
	if claude.PassRate != 0.91 {
		t.Errorf("expected pass rate 0.91, got %f", claude.PassRate)
	}
	if claude.PassRateDelta != 0.38 {
		t.Errorf("expected delta 0.38, got %f", claude.PassRateDelta)
	}
}
