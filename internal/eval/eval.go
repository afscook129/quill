package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const FileName = "evals.json"

type Suite struct {
	Version      string `json:"version"`
	Skill        string `json:"skill"`
	SkillVersion string `json:"skill_version"`
	Description  string `json:"description"`
	Cases        []Case `json:"cases"`
}

type Case struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Input       Input   `json:"input"`
	Expected    any     `json:"expected"`
	Grading     Grading `json:"grading"`
	Tags        []string `json:"tags"`
	Generated   bool    `json:"generated"`
}

type Input struct {
	Prompt  string         `json:"prompt"`
	Context map[string]any `json:"context,omitempty"`
}

type Grading struct {
	Method    string `json:"method"`    // deterministic, llm-judge, human
	Rubric    string `json:"rubric,omitempty"`
	Assertion string `json:"assertion,omitempty"`
	Model     string `json:"model,omitempty"`
}

func Parse(r io.Reader) (*Suite, error) {
	var s Suite
	dec := json.NewDecoder(r)
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("parsing eval suite: %w", err)
	}
	if len(s.Cases) == 0 {
		return nil, fmt.Errorf("eval suite has no cases")
	}
	return &s, nil
}

func Load(skillDir string) (*Suite, error) {
	path := filepath.Join(skillDir, "evals", FileName)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// HasTag returns true if the case has the given tag.
func (c *Case) HasTag(tag string) bool {
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
