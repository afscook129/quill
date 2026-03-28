package registry

import "strings"

// Client is the interface for registry operations.
type Client interface {
	Search(query string, model string) (*SearchResponse, error)
	Resolve(name string, version string) (*SkillResult, error)
}

// MockClient returns realistic stub data for development.
type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) Search(query string, model string) (*SearchResponse, error) {
	q := strings.ToLower(query)

	results := []SkillResult{
		{
			Name:        "ticket-classifier",
			Version:     "2.1.0",
			Source:      "skills.sh",
			Description: "Classifies support tickets by topic and urgency using two-pass approach",
			Delta:       0.38,
			PassRate:    0.94,
			TokenCost:   1240,
			Verified:    true,
			LastTested:  "2026-03-27",
		},
		{
			Name:        "support-router",
			Version:     "3.0.0",
			Source:      "ClawHub",
			Description: "Routes support requests to appropriate teams with confidence scoring",
			Delta:       0.31,
			PassRate:    0.84,
			TokenCost:   980,
			Verified:    true,
			LastTested:  "2026-03-26",
		},
		{
			Name:        "ticket-sorter",
			Version:     "1.1.0",
			Source:      "npm",
			Description: "Basic ticket sorting by category and priority level",
			Delta:       0.19,
			PassRate:    0.76,
			TokenCost:   620,
			Verified:    false,
			LastTested:  "2026-03-20",
		},
	}

	// Vary results based on query keywords
	if strings.Contains(q, "summariz") || strings.Contains(q, "summary") {
		results = []SkillResult{
			{
				Name:        "summarize-findings",
				Version:     "1.2.1",
				Source:      "skills.sh",
				Description: "Summarizes research findings with key takeaways and confidence levels",
				Delta:       0.34,
				PassRate:    0.90,
				TokenCost:   890,
				Verified:    true,
				LastTested:  "2026-03-28",
			},
			{
				Name:        "doc-summarizer",
				Version:     "2.0.0",
				Source:      "ClawHub",
				Description: "Condensed document summaries with configurable detail level",
				Delta:       0.28,
				PassRate:    0.87,
				TokenCost:   1100,
				Verified:    true,
				LastTested:  "2026-03-25",
			},
			{
				Name:        "quick-recap",
				Version:     "1.0.3",
				Source:      "npm",
				Description: "Fast one-paragraph summaries for meeting notes and documents",
				Delta:       0.15,
				PassRate:    0.79,
				TokenCost:   450,
				Verified:    false,
				LastTested:  "2026-03-18",
			},
		}
	} else if strings.Contains(q, "code") || strings.Contains(q, "format") || strings.Contains(q, "review") {
		results = []SkillResult{
			{
				Name:        "code-reviewer",
				Version:     "3.1.0",
				Source:      "skills.sh",
				Description: "Structured code review with security, performance, and style analysis",
				Delta:       0.41,
				PassRate:    0.92,
				TokenCost:   1560,
				Verified:    true,
				LastTested:  "2026-03-28",
			},
			{
				Name:        "pr-analyzer",
				Version:     "2.0.1",
				Source:      "ClawHub",
				Description: "Pull request analysis with impact assessment and suggested improvements",
				Delta:       0.33,
				PassRate:    0.88,
				TokenCost:   1200,
				Verified:    true,
				LastTested:  "2026-03-27",
			},
			{
				Name:        "lint-helper",
				Version:     "1.4.0",
				Source:      "npm",
				Description: "Code style enforcement and auto-fix suggestions",
				Delta:       0.12,
				PassRate:    0.82,
				TokenCost:   380,
				Verified:    false,
				LastTested:  "2026-03-22",
			},
		}
	}

	if model == "" {
		model = "claude-sonnet-4.6"
	}

	resp := &SearchResponse{
		Results:     results,
		Query:       query,
		Model:       model,
		Recommended: &results[0],
		Reasoning:   formatReasoning(results[0]),
	}

	return resp, nil
}

func (c *MockClient) Resolve(name string, version string) (*SkillResult, error) {
	return &SkillResult{
		Name:       name,
		Version:    version,
		Source:     "skills.sh",
		Delta:      0.35,
		PassRate:   0.89,
		TokenCost:  1000,
		Verified:   true,
		LastTested: "2026-03-28",
	}, nil
}

func formatReasoning(top SkillResult) string {
	return "highest delta (" + FormatDeltaPP(top.Delta) + ") at " +
		itoa(top.TokenCost) + " tokens context cost"
}

func FormatDeltaPP(d float64) string {
	pp := int(d * 100)
	if pp > 0 {
		return "+" + itoa(pp) + "pp"
	}
	return itoa(pp) + "pp"
}

func FormatPercent(f float64) string {
	return itoa(int(f*100)) + "%"
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
