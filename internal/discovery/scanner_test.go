package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsSkills(t *testing.T) {
	dir := t.TempDir()

	// Create a skill
	skillDir := filepath.Join(dir, "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := `---
name: my-skill
description: A test skill
version: "1.0.0"
---
# My Skill
Does stuff.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	s := skills[0]
	if s.Name != "my-skill" {
		t.Errorf("expected name 'my-skill', got %q", s.Name)
	}
	if s.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", s.Version)
	}
	if s.Description != "A test skill" {
		t.Errorf("expected description 'A test skill', got %q", s.Description)
	}
}

func TestScanSkipsNodeModules(t *testing.T) {
	dir := t.TempDir()

	// Create a skill in node_modules (should be skipped)
	nmDir := filepath.Join(dir, "node_modules", "some-skill")
	if err := os.MkdirAll(nmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nmDir, "SKILL.md"), []byte("---\nname: hidden\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(skills) != 0 {
		t.Fatalf("expected 0 skills (node_modules should be skipped), got %d", len(skills))
	}
}

func TestScanNoFrontmatter(t *testing.T) {
	dir := t.TempDir()

	skillDir := filepath.Join(dir, "plain-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// SKILL.md without frontmatter
	content := "# Plain Skill\nJust a markdown file.\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	// Name should be inferred from directory
	if skills[0].Name != "plain-skill" {
		t.Errorf("expected name 'plain-skill', got %q", skills[0].Name)
	}
}

func TestScanEmpty(t *testing.T) {
	dir := t.TempDir()

	skills, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(skills) != 0 {
		t.Fatalf("expected 0 skills, got %d", len(skills))
	}
}
