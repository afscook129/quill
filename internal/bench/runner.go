package bench

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quill-dev/quill/internal/eval"
	"github.com/quill-dev/quill/internal/provider"
)

// Options configures a bench run.
type Options struct {
	Model   string
	Trials  int
	Publish bool
}

// Run executes a full benchmark: with-skill vs without-skill comparison.
func Run(skillPath string, opts Options) (*Result, error) {
	// Load SKILL.md content
	skillContent, skillName, err := loadSkill(skillPath)
	if err != nil {
		return nil, fmt.Errorf("loading skill: %w", err)
	}

	// Load eval suite
	suite, err := eval.Load(skillPath)
	if err != nil {
		return nil, fmt.Errorf("loading evals: %w", err)
	}

	// Get provider
	p, err := provider.ForModel(opts.Model)
	if err != nil {
		return nil, err
	}

	if opts.Trials <= 0 {
		opts.Trials = 3
	}

	// Run each case
	var caseResults []caseResult
	totalTokensWith := 0
	totalTokensWithout := 0
	totalLatencyWith := 0

	for _, c := range suite.Cases {
		cr, err := runCase(p, opts.Model, skillContent, skillName, c, opts.Trials)
		if err != nil {
			return nil, fmt.Errorf("running case %s: %w", c.ID, err)
		}
		caseResults = append(caseResults, *cr)
		totalTokensWith += cr.avgTokensWith
		totalTokensWithout += cr.avgTokensWithout
		totalLatencyWith += cr.avgLatencyWith
	}

	// Compute pass rates
	totalWith := 0
	totalWithout := 0
	totalTrials := 0
	for _, cr := range caseResults {
		totalWith += cr.passCountWith
		totalWithout += cr.passCountWithout
		totalTrials += cr.trials
	}

	passWith := float64(totalWith) / float64(totalTrials)
	passWithout := float64(totalWithout) / float64(totalTrials)
	delta := passWith - passWithout

	// Collect failed cases
	var failed []FailedCase
	for _, cr := range caseResults {
		if cr.passCountWith < cr.trials {
			edge := ""
			for _, tag := range cr.tags {
				if tag != "positive" && tag != "negative" && tag != "implicit" && tag != "standard" {
					if edge != "" {
						edge += "/"
					}
					edge += tag
				}
			}
			failed = append(failed, FailedCase{
				ID:    cr.id,
				Input: cr.input,
				Edge:  edge,
			})
		}
	}

	// Cluster failure patterns
	patterns := clusterPatterns(failed)

	avgTokens := 0
	avgLatency := 0
	n := len(suite.Cases)
	if n > 0 {
		avgTokens = totalTokensWith / n
		avgLatency = totalLatencyWith / n
	}

	return &Result{
		SkillName:    skillName,
		Model:        opts.Model,
		Trials:       opts.Trials,
		PassWith:     passWith,
		PassWithout:  passWithout,
		Delta:        delta,
		AvgTokens:    avgTokens,
		AvgLatencyMs: avgLatency,
		FailedCases:  failed,
		Patterns:     patterns,
		EarningPlace: delta >= 0.08,
		Suggestion:   suggestFix(failed, patterns),
	}, nil
}

type caseResult struct {
	id               string
	input            string
	tags             []string
	trials           int
	passCountWith    int
	passCountWithout int
	avgTokensWith    int
	avgTokensWithout int
	avgLatencyWith   int
}

func runCase(p provider.Provider, model string, skillContent string, skillName string, c eval.Case, trials int) (*caseResult, error) {
	cr := &caseResult{
		id:     c.ID,
		input:  c.Input.Prompt,
		tags:   c.Tags,
		trials: trials,
	}

	for trial := 0; trial < trials; trial++ {
		// WITH skill
		withResp, err := p.Call(model, skillContent, []provider.Message{
			{Role: "user", Content: c.Input.Prompt},
		})
		if err != nil {
			return nil, fmt.Errorf("with-skill call (trial %d): %w", trial+1, err)
		}

		// WITHOUT skill
		withoutResp, err := p.Call(model, "", []provider.Message{
			{Role: "user", Content: c.Input.Prompt},
		})
		if err != nil {
			return nil, fmt.Errorf("without-skill call (trial %d): %w", trial+1, err)
		}

		// Grade both
		withGrade := grade(c, withResp.Content, skillName, p, model)
		withoutGrade := grade(c, withoutResp.Content, skillName, p, model)

		if withGrade.Pass {
			cr.passCountWith++
		}
		if withoutGrade.Pass {
			cr.passCountWithout++
		}

		cr.avgTokensWith += withResp.InputTokens + withResp.OutputTokens
		cr.avgTokensWithout += withoutResp.InputTokens + withoutResp.OutputTokens
		cr.avgLatencyWith += withResp.LatencyMs
	}

	if trials > 0 {
		cr.avgTokensWith /= trials
		cr.avgTokensWithout /= trials
		cr.avgLatencyWith /= trials
	}

	return cr, nil
}

func grade(c eval.Case, output string, skillName string, p provider.Provider, model string) *eval.GradeResult {
	switch c.Grading.Method {
	case "deterministic":
		return eval.GradeDeterministic(output, c.Expected, c.Grading.Assertion)

	case "llm-judge":
		prompt := eval.LLMJudgePrompt(skillName, c.Input.Prompt, c.Expected, c.Grading.Rubric, output)
		gradeModel := model
		if c.Grading.Model != "" && c.Grading.Model != "default" {
			gradeModel = c.Grading.Model
		}
		resp, err := p.Call(gradeModel, "", []provider.Message{
			{Role: "user", Content: prompt},
		})
		if err != nil {
			return &eval.GradeResult{Pass: false, Method: "llm-judge", Reasoning: fmt.Sprintf("judge error: %s", err)}
		}
		return eval.ParseJudgeVerdict(resp.Content)

	case "human":
		return &eval.GradeResult{Pass: false, Method: "human", Reasoning: "pending human review"}

	default:
		return &eval.GradeResult{Pass: false, Method: c.Grading.Method, Reasoning: "unknown grading method"}
	}
}

func loadSkill(skillPath string) (content string, name string, err error) {
	// Try SKILL.md in the path
	candidates := []string{
		filepath.Join(skillPath, "SKILL.md"),
		skillPath, // might be a direct path to SKILL.md
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			name = filepath.Base(filepath.Dir(path))
			if name == "." || name == "" {
				name = filepath.Base(skillPath)
			}
			return string(data), name, nil
		}
	}

	return "", "", fmt.Errorf("no SKILL.md found in %s", skillPath)
}

func clusterPatterns(failed []FailedCase) []string {
	if len(failed) == 0 {
		return nil
	}

	// Count edge type occurrences
	edgeCounts := map[string]int{}
	for _, f := range failed {
		if f.Edge != "" {
			for _, e := range strings.Split(f.Edge, "/") {
				edgeCounts[e]++
			}
		}
	}

	var patterns []string
	for edge, count := range edgeCounts {
		if count >= 2 || float64(count)/float64(len(failed)) > 0.5 {
			patterns = append(patterns, fmt.Sprintf("%s (%d of %d failures)", edge, count, len(failed)))
		}
	}

	return patterns
}

func suggestFix(failed []FailedCase, patterns []string) string {
	if len(failed) == 0 {
		return ""
	}
	if len(patterns) > 0 {
		edges := []string{}
		for _, f := range failed {
			if f.Edge != "" {
				for _, e := range strings.Split(f.Edge, "/") {
					if !containsStr(edges, e) {
						edges = append(edges, e)
					}
				}
			}
		}
		if len(edges) > 0 {
			return "add explicit handling for " + strings.Join(edges, ", ")
		}
	}
	return fmt.Sprintf("%d cases failed — review failure patterns", len(failed))
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
