package manifest

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

const FileName = "quill.manifest.yaml"

type Manifest struct {
	Name         string              `yaml:"name"`
	Version      string              `yaml:"version"`
	Source       string              `yaml:"source,omitempty"`
	Description  string              `yaml:"description,omitempty"`
	Dependencies *Dependencies       `yaml:"dependencies,omitempty"`
	Variants     map[string]*Variant `yaml:"variants,omitempty"`
	Permissions  *Permissions        `yaml:"permissions,omitempty"`
	Context      *ContextBudget      `yaml:"context,omitempty"`
	Interface    *Interface          `yaml:"interface,omitempty"`
	Evals        *EvalConfig         `yaml:"evals,omitempty"`
	Bench        *BenchConfig        `yaml:"bench,omitempty"`
	Conflicts    *Conflicts          `yaml:"conflicts,omitempty"`
	Provenance   *Provenance         `yaml:"provenance,omitempty"`
}

type Dependencies struct {
	Skills []SkillDep `yaml:"skills,omitempty"`
	Models []ModelDep `yaml:"models,omitempty"`
}

type SkillDep struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Required bool   `yaml:"required"`
}

type ModelDep struct {
	Family        string `yaml:"family"`
	MinCapability string `yaml:"min_capability,omitempty"`
}

type Variant struct {
	Prompt       string   `yaml:"prompt"`
	TestedModels []string `yaml:"tested_models,omitempty"`
	PassRate     float64  `yaml:"pass_rate,omitempty"`
	PassRateDlt  float64  `yaml:"pass_rate_delta,omitempty"`
}

type Permissions struct {
	Filesystem   string `yaml:"filesystem,omitempty"`
	Network      string `yaml:"network,omitempty"`
	Shell        string `yaml:"shell,omitempty"`
	ContextScope string `yaml:"context_scope,omitempty"`
}

type ContextBudget struct {
	EstimatedTokens int    `yaml:"estimated_tokens,omitempty"`
	BudgetPriority  string `yaml:"budget_priority,omitempty"`
}

type Interface struct {
	OutputSchema string `yaml:"output_schema,omitempty"`
}

type EvalConfig struct {
	Suite         string  `yaml:"suite,omitempty"`
	PassThreshold float64 `yaml:"pass_threshold,omitempty"`
	SmokeCount    int     `yaml:"smoke_count,omitempty"`
}

type BenchConfig struct {
	WorthKeeping *WorthKeeping `yaml:"worth_keeping,omitempty"`
}

type WorthKeeping struct {
	MinDelta            float64 `yaml:"min_delta,omitempty"`
	ConsecutiveVersions int     `yaml:"consecutive_versions,omitempty"`
}

type Conflicts struct {
	IncompatibleWith []IncompatibleSkill `yaml:"incompatible_with,omitempty"`
}

type IncompatibleSkill struct {
	Skill  string `yaml:"skill"`
	Reason string `yaml:"reason,omitempty"`
}

type Provenance struct {
	Publisher string `yaml:"publisher,omitempty"`
	Signed    bool   `yaml:"signed,omitempty"`
	License   string `yaml:"license,omitempty"`
}

func Parse(r io.Reader) (*Manifest, error) {
	var m Manifest
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}

func Write(w io.Writer, m *Manifest) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	return enc.Close()
}

func Load(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func Save(path string, m *Manifest) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Write(f, m)
}

func Default(name string) *Manifest {
	return &Manifest{
		Name:    name,
		Version: "0.1.0",
		Permissions: &Permissions{
			Filesystem:   "none",
			Network:      "none",
			Shell:        "never",
			ContextScope: "input_only",
		},
		Context: &ContextBudget{
			BudgetPriority: "normal",
		},
		Evals: &EvalConfig{
			PassThreshold: 0.87,
			SmokeCount:    8,
		},
		Bench: &BenchConfig{
			WorthKeeping: &WorthKeeping{
				MinDelta:            0.08,
				ConsecutiveVersions: 2,
			},
		},
	}
}
