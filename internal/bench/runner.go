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
	skillContent, skillName, err := loadSkill(skillPath)
	if err != nil {
		return nil, fmt.Errorf("loading skill: %w", err)
	}

	suite, err := eval.Load(skillPath)
	if err != nil {
		return nil, fmt.Errorf("loading evals from %s: %w\n\n  Create evals/evals.json in your skill directory.\n  See: quill bench --help", skillPath, err)
	}

	p, err := provider.ForModel(opts.Model)
	if err != nil {
		return nil, err
	}

	if opts.Trials <= 0 {
		opts.Trials = 3
	}

	return runBenchmark(p, opts.Model, skillContent, skillName, suite, opts.Trials)
}

// runBenchmark is the core loop, separated for testability.
func runBenchmark(p provider.Provider, model string, skillContent string, skillName string, suite *eval.Suite, trials int) (*Result, error) {
	var cases []caseResult

	for _, c := range suite.Cases {
		if c.Grading.Method == "human" {
			continue // skip human-graded cases in automated runs
		}

		cr, err := runCase(p, model, skillContent, skillName, c, trials)
		if err != nil {
			return nil, fmt.Errorf("case %s: %w", c.ID, err)
		}
		cases = append(cases, *cr)
	}

	if len(cases) == 0 {
		return nil, fmt.Errorf("no gradeable eval cases (all cases use 'human' grading)")
	}

	return computeResult(skillName, model, trials, cases), nil
}

// computeResult aggregates case results into the final bench result.
// Pass rates are computed per-case then averaged (each case weighted equally).
func computeResult(skillName string, model string, trials int, cases []caseResult) *Result {
	// Per-case pass rates, then average across cases
	sumPassRateWith := 0.0
	sumPassRateWithout := 0.0
	totalTokens := 0
	totalLatency := 0

	for _, cr := range cases {
		sumPassRateWith += float64(cr.passCountWith) / float64(cr.trials)
		sumPassRateWithout += float64(cr.passCountWithout) / float64(cr.trials)
		totalTokens += cr.totalTokensWith
		totalLatency += cr.totalLatencyWith
	}

	n := len(cases)
	passWith := sumPassRateWith / float64(n)
	passWithout := sumPassRateWithout / float64(n)
	delta := passWith - passWithout

	// Total API calls made
	totalCalls := n * trials

	// Average tokens and latency per call (not per case)
	avgTokens := 0
	avgLatency := 0
	if totalCalls > 0 {
		avgTokens = totalTokens / totalCalls
		avgLatency = totalLatency / totalCalls
	}

	// Collect failed cases
	var failed []FailedCase
	for _, cr := range cases {
		if cr.passCountWith < cr.trials {
			failed = append(failed, FailedCase{
				ID:    cr.id,
				Input: truncate(cr.input, 60),
				Edge:  extractEdgeTags(cr.tags),
			})
		}
	}

	patterns := clusterPatterns(failed)

	return &Result{
		SkillName:    skillName,
		Model:        model,
		Trials:       trials,
		CaseCount:    n,
		PassWith:     passWith,
		PassWithout:  passWithout,
		Delta:        delta,
		AvgTokens:    avgTokens,
		AvgLatencyMs: avgLatency,
		FailedCases:  failed,
		Patterns:     patterns,
		EarningPlace: delta >= 0.08,
		Suggestion:   suggestFix(failed, patterns),
	}
}

type caseResult struct {
	id               string
	input            string
	tags             []string
	trials           int
	passCountWith    int
	passCountWithout int
	totalTokensWith  int // raw total, not averaged
	totalTokensWithout int
	totalLatencyWith int // raw total, not averaged
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
			return nil, fmt.Errorf("with-skill trial %d: %w", trial+1, err)
		}

		// WITHOUT skill
		withoutResp, err := p.Call(model, "", []provider.Message{
			{Role: "user", Content: c.Input.Prompt},
		})
		if err != nil {
			return nil, fmt.Errorf("without-skill trial %d: %w", trial+1, err)
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

		cr.totalTokensWith += withResp.InputTokens + withResp.OutputTokens
		cr.totalTokensWithout += withoutResp.InputTokens + withoutResp.OutputTokens
		cr.totalLatencyWith += withResp.LatencyMs
	}

	return cr, nil
}

func grade(c eval.Case, output string, skillName string, p provider.Provider, model string) *eval.GradeResult {
	// Strip common LLM output wrapping before grading
	cleaned := cleanOutput(output)

	switch c.Grading.Method {
	case "deterministic":
		return eval.GradeDeterministic(cleaned, c.Expected, c.Grading.Assertion)

	case "llm-judge":
		prompt := eval.LLMJudgePrompt(skillName, c.Input.Prompt, c.Expected, c.Grading.Rubric, output)
		gradeModel := model
		if c.Grading.Model != "" && c.Grading.Model != "default" {
			gradeModel = c.Grading.Model
		}
		resp, err := p.Call(gradeModel, "You are an eval judge. Respond with exactly:\nVERDICT: PASS or FAIL\nREASONING: <one sentence>", []provider.Message{
			{Role: "user", Content: prompt},
		})
		if err != nil {
			return &eval.GradeResult{Pass: false, Method: "llm-judge", Reasoning: fmt.Sprintf("judge call failed: %s", err)}
		}
		return eval.ParseJudgeVerdict(resp.Content)

	default:
		return &eval.GradeResult{Pass: false, Method: c.Grading.Method, Reasoning: fmt.Sprintf("unknown grading method: %s", c.Grading.Method)}
	}
}

// cleanOutput strips common LLM wrapping (markdown code fences, etc.)
func cleanOutput(s string) string {
	s = strings.TrimSpace(s)
	// Strip ```json ... ``` wrapping
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) >= 3 && strings.HasPrefix(lines[len(lines)-1], "```") {
			s = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	return strings.TrimSpace(s)
}

func loadSkill(skillPath string) (content string, name string, err error) {
	candidates := []string{
		filepath.Join(skillPath, "SKILL.md"),
		skillPath,
	}

	for _, path := range candidates {
		data, readErr := os.ReadFile(path)
		if readErr == nil {
			name = filepath.Base(filepath.Dir(path))
			if name == "." || name == "" {
				name = filepath.Base(skillPath)
			}
			return string(data), name, nil
		}
	}

	return "", "", fmt.Errorf("no SKILL.md found in %s", skillPath)
}

func extractEdgeTags(tags []string) string {
	var edges []string
	for _, tag := range tags {
		switch tag {
		case "positive", "negative", "implicit", "standard":
			continue
		default:
			edges = append(edges, tag)
		}
	}
	return strings.Join(edges, "/")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func clusterPatterns(failed []FailedCase) []string {
	if len(failed) == 0 {
		return nil
	}

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
		var edges []string
		for _, f := range failed {
			for _, e := range strings.Split(f.Edge, "/") {
				if e != "" && !containsStr(edges, e) {
					edges = append(edges, e)
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
