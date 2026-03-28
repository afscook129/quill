package bench

// Result holds the outcome of a benchmark run.
type Result struct {
	SkillName    string       `json:"skill_name"`
	Model        string       `json:"model"`
	Trials       int          `json:"trials"`
	CaseCount    int          `json:"case_count"`
	PassWith     float64      `json:"pass_with"`
	PassWithout  float64      `json:"pass_without"`
	Delta        float64      `json:"delta"`
	AvgTokens    int          `json:"avg_tokens"`
	AvgLatencyMs int          `json:"avg_latency_ms"`
	FailedCases  []FailedCase `json:"failed_cases,omitempty"`
	Patterns     []string     `json:"patterns,omitempty"`
	EarningPlace bool         `json:"earning_place"`
	Suggestion   string       `json:"suggestion,omitempty"`

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

// EstimateCost returns approximate USD cost for a bench run.
// Based on typical token usage and provider pricing.
func EstimateCost(cases int, trials int, model string) float64 {
	// Average ~1,700 tokens per with-skill call, ~700 per without-skill call
	// 2 conditions * trials * cases
	callsPerCase := 2 * trials
	totalCalls := cases * callsPerCase

	// Conservative: ~1,200 tokens average per call
	avgTokensPerCall := 1200
	totalTokens := totalCalls * avgTokensPerCall

	// Approximate pricing (input + output blend)
	pricePerMToken := 9.0 // ~$3 input + $15 output, blended
	if isSmallModel(model) {
		pricePerMToken = 1.5
	}

	return float64(totalTokens) / 1_000_000 * pricePerMToken
}

func isSmallModel(model string) bool {
	for _, s := range []string{"haiku", "mini", "flash"} {
		if len(model) >= len(s) {
			for i := 0; i <= len(model)-len(s); i++ {
				if model[i:i+len(s)] == s {
					return true
				}
			}
		}
	}
	return false
}
