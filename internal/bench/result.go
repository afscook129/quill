package bench

type Result struct {
	SkillName     string       `json:"skill_name"`
	Model         string       `json:"model"`
	Trials        int          `json:"trials"`
	PassWith      float64      `json:"pass_with"`
	PassWithout   float64      `json:"pass_without"`
	Delta         float64      `json:"delta"`
	AvgTokens     int          `json:"avg_tokens"`
	AvgLatencyMs  int          `json:"avg_latency_ms"`
	FailedCases   []FailedCase `json:"failed_cases,omitempty"`
	Patterns      []string     `json:"patterns,omitempty"`
	EarningPlace  bool         `json:"earning_place"`
	Suggestion    string       `json:"suggestion,omitempty"`
	TriggerAccuracy *TriggerResult `json:"trigger_accuracy,omitempty"`
}

type FailedCase struct {
	ID    string `json:"id"`
	Input string `json:"input"`
	Edge  string `json:"edge"`
}

type TriggerResult struct {
	Total          int `json:"total"`
	Correct        int `json:"correct"`
	FalsePositives int `json:"false_positives"`
	FalseNegatives int `json:"false_negatives"`
}

// MockResult returns a realistic benchmark result for demonstration.
func MockResult(skillName string, model string) *Result {
	return &Result{
		SkillName:   skillName,
		Model:       model,
		Trials:      3,
		PassWith:    0.81,
		PassWithout: 0.42,
		Delta:       0.39,
		AvgTokens:   1847,
		AvgLatencyMs: 2300,
		FailedCases: []FailedCase{
			{ID: "tc-07", Input: "plz help cant pay", Edge: "emoji/short"},
			{ID: "tc-12", Input: "aide moi avec mon compte", Edge: "non-english"},
			{ID: "tc-15", Input: ".", Edge: "minimal input"},
			{ID: "tc-19", Input: "ambiguous billing vs technical", Edge: "category boundary"},
		},
		Patterns:     []string{"non-standard input (3 of 4 failures)"},
		EarningPlace: true,
		Suggestion:   "add explicit handling for emoji, short, non-english",
		TriggerAccuracy: &TriggerResult{
			Total:          20,
			Correct:        18,
			FalsePositives: 1,
			FalseNegatives: 1,
		},
	}
}
