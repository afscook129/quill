package lock

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const FileName = "quill.lock"

type Lock struct {
	Meta     Meta            `yaml:"meta"`
	Resolved []ResolvedSkill `yaml:"resolved"`
	System   System          `yaml:"system"`
}

type Meta struct {
	QuillVersion string `yaml:"quill_version"`
	GeneratedAt  string `yaml:"generated_at"`
	GeneratedBy  string `yaml:"generated_by,omitempty"`
	ModelFamily  string `yaml:"model_family,omitempty"`
	ModelVersion string `yaml:"model_version,omitempty"`
}

type ResolvedSkill struct {
	Skill              string       `yaml:"skill"`
	Source             string       `yaml:"source"`
	InstalledVia       string       `yaml:"installed_via"`
	ContentHash        string       `yaml:"content_hash,omitempty"`
	ManifestHash       string       `yaml:"manifest_hash,omitempty"`
	SignatureVerified   bool         `yaml:"signature_verified"`
	VariantSelected    string       `yaml:"variant_selected,omitempty"`
	ModelTested        string       `yaml:"model_tested,omitempty"`
	EvalPassRate       *float64     `yaml:"eval_pass_rate"`
	EvalPassRateDelta  *float64     `yaml:"eval_pass_rate_delta"`
	EvalSource         string       `yaml:"eval_source,omitempty"`
	EvalGenerated      *bool        `yaml:"eval_generated"`
	ContextBudgetTkns  int          `yaml:"context_budget_tokens,omitempty"`
	PermissionsGranted *Permissions `yaml:"permissions_granted,omitempty"`
	InstalledAt        string       `yaml:"installed_at,omitempty"`
	InstalledBy        string       `yaml:"installed_by,omitempty"`
}

type Permissions struct {
	Filesystem string `yaml:"filesystem,omitempty"`
	Network    string `yaml:"network,omitempty"`
	Shell      string `yaml:"shell,omitempty"`
}

type System struct {
	TotalContextTokens int      `yaml:"total_context_budget_tokens"`
	ContextBudgetLimit int      `yaml:"context_budget_limit"`
	SkillsWithoutBench []string `yaml:"skills_without_bench,omitempty"`
	UnsignedSkills     []string `yaml:"unsigned_skills,omitempty"`
	DriftWarnings      []string `yaml:"drift_warnings,omitempty"`
	UpgradeAvailable   []string `yaml:"upgrade_available,omitempty"`
	SignalsEnabled     bool     `yaml:"signals_enabled"`
	SignalsScope       string   `yaml:"signals_scope,omitempty"`
}

func Parse(r io.Reader) (*Lock, error) {
	var l Lock
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&l); err != nil {
		return nil, fmt.Errorf("parsing lock file: %w", err)
	}
	return &l, nil
}

func Write(w io.Writer, l *Lock) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(l); err != nil {
		return fmt.Errorf("writing lock file: %w", err)
	}
	return enc.Close()
}

func Load(path string) (*Lock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func Save(path string, l *Lock) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Write(f, l)
}

func New(version string) *Lock {
	return &Lock{
		Meta: Meta{
			QuillVersion: version,
			GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		},
		Resolved: []ResolvedSkill{},
		System: System{
			ContextBudgetLimit: 200000, // conservative default for modern models
			SignalsEnabled:     true,
			SignalsScope:       "anonymous_aggregate",
		},
	}
}

// ContextLimitForModel returns the context window size for known models.
func ContextLimitForModel(model string) int {
	switch {
	case contains(model, "claude"):
		return 200000
	case contains(model, "haiku"):
		return 200000
	case contains(model, "gpt-5"), contains(model, "gpt-4o"):
		return 128000
	case contains(model, "gpt-4"):
		return 128000
	case contains(model, "gemini"):
		return 1000000
	default:
		return 200000
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
