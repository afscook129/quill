package registry

type SkillResult struct {
	Name        string  `json:"name"`
	Version     string  `json:"version"`
	Source      string  `json:"source"`
	Description string  `json:"description"`
	Delta       float64 `json:"delta"`
	PassRate    float64 `json:"pass_rate"`
	TokenCost   int     `json:"token_cost"`
	Verified    bool    `json:"verified"`
	LastTested  string  `json:"last_tested"`
}

type SearchResponse struct {
	Results     []SkillResult `json:"results"`
	Recommended *SkillResult  `json:"recommended,omitempty"`
	Reasoning   string        `json:"reasoning"`
	Query       string        `json:"query"`
	Model       string        `json:"model"`
}
