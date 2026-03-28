package bench

import (
	"math"
	"testing"
)

func TestComputeResult_BasicDelta(t *testing.T) {
	// 2 cases, 3 trials each
	// Case 1: with passes 3/3, without passes 1/3
	// Case 2: with passes 2/3, without passes 0/3
	cases := []caseResult{
		{id: "c1", trials: 3, passCountWith: 3, passCountWithout: 1, totalTokensWith: 3600, totalLatencyWith: 6000, tags: []string{"positive"}},
		{id: "c2", trials: 3, passCountWith: 2, passCountWithout: 0, totalTokensWith: 3600, totalLatencyWith: 6000, tags: []string{"positive", "edge-case"}},
	}

	r := computeResult("test-skill", "claude-sonnet-4.6", 3, cases)

	// Case 1 pass rate: with=1.0, without=0.333
	// Case 2 pass rate: with=0.667, without=0.0
	// Average with: (1.0 + 0.667) / 2 = 0.833
	// Average without: (0.333 + 0.0) / 2 = 0.167
	// Delta: 0.833 - 0.167 = 0.667
	assertClose(t, r.PassWith, 0.833, 0.01, "PassWith")
	assertClose(t, r.PassWithout, 0.167, 0.01, "PassWithout")
	assertClose(t, r.Delta, 0.667, 0.01, "Delta")

	if !r.EarningPlace {
		t.Error("expected EarningPlace=true for delta 0.667")
	}
}

func TestComputeResult_NegativeDelta(t *testing.T) {
	// Skill makes things worse
	cases := []caseResult{
		{id: "c1", trials: 3, passCountWith: 1, passCountWithout: 3, totalTokensWith: 3600, totalLatencyWith: 3000},
	}

	r := computeResult("bad-skill", "gpt-5.4", 3, cases)

	assertClose(t, r.PassWith, 0.333, 0.01, "PassWith")
	assertClose(t, r.PassWithout, 1.0, 0.01, "PassWithout")

	if r.Delta >= 0 {
		t.Errorf("expected negative delta, got %f", r.Delta)
	}
	if r.EarningPlace {
		t.Error("expected EarningPlace=false for negative delta")
	}
}

func TestComputeResult_ZeroDelta(t *testing.T) {
	cases := []caseResult{
		{id: "c1", trials: 3, passCountWith: 2, passCountWithout: 2, totalTokensWith: 3600, totalLatencyWith: 3000},
	}

	r := computeResult("no-effect", "claude-sonnet-4.6", 3, cases)

	assertClose(t, r.Delta, 0.0, 0.01, "Delta")
	if r.EarningPlace {
		t.Error("expected EarningPlace=false for zero delta")
	}
}

func TestComputeResult_TokenAveraging(t *testing.T) {
	// 2 cases, 2 trials each = 4 total calls
	// Total tokens: 1000 + 2000 = 3000
	cases := []caseResult{
		{id: "c1", trials: 2, passCountWith: 2, passCountWithout: 0, totalTokensWith: 1000, totalLatencyWith: 2000},
		{id: "c2", trials: 2, passCountWith: 2, passCountWithout: 0, totalTokensWith: 2000, totalLatencyWith: 4000},
	}

	r := computeResult("test", "claude-sonnet-4.6", 2, cases)

	// Total tokens: 3000, total calls: 4
	// Average per call: 750
	if r.AvgTokens != 750 {
		t.Errorf("expected avg tokens 750, got %d", r.AvgTokens)
	}
	if r.AvgLatencyMs != 1500 {
		t.Errorf("expected avg latency 1500, got %d", r.AvgLatencyMs)
	}
}

func TestComputeResult_FailedCases(t *testing.T) {
	cases := []caseResult{
		{id: "c1", trials: 3, passCountWith: 3, passCountWithout: 0, tags: []string{"positive", "standard"}},
		{id: "c2", trials: 3, passCountWith: 1, passCountWithout: 0, tags: []string{"positive", "edge-case", "emoji"}},
		{id: "c3", trials: 3, passCountWith: 0, passCountWithout: 0, tags: []string{"positive", "edge-case", "short"}},
	}

	r := computeResult("test", "claude-sonnet-4.6", 3, cases)

	if len(r.FailedCases) != 2 {
		t.Fatalf("expected 2 failed cases, got %d", len(r.FailedCases))
	}

	if r.FailedCases[0].ID != "c2" {
		t.Errorf("expected first failure c2, got %s", r.FailedCases[0].ID)
	}
	if r.FailedCases[0].Edge != "edge-case/emoji" {
		t.Errorf("expected edge 'edge-case/emoji', got %q", r.FailedCases[0].Edge)
	}
}

func TestComputeResult_PatternClustering(t *testing.T) {
	cases := []caseResult{
		{id: "c1", trials: 3, passCountWith: 1, tags: []string{"positive", "edge-case"}},
		{id: "c2", trials: 3, passCountWith: 0, tags: []string{"positive", "edge-case"}},
		{id: "c3", trials: 3, passCountWith: 0, tags: []string{"positive", "short"}},
	}

	r := computeResult("test", "claude-sonnet-4.6", 3, cases)

	if len(r.Patterns) == 0 {
		t.Fatal("expected at least one pattern")
	}
	// edge-case appears in 2/3 failures, should be clustered
	found := false
	for _, p := range r.Patterns {
		if len(p) > 0 && p[0:4] == "edge" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'edge-case' pattern, got %v", r.Patterns)
	}
}

func TestEstimateCost(t *testing.T) {
	cost := EstimateCost(8, 3, "claude-sonnet-4.6")
	// 8 cases * 2 conditions * 3 trials = 48 calls
	// 48 * 1200 tokens = 57,600 tokens
	// 57,600 / 1M * $9 = ~$0.52
	if cost < 0.3 || cost > 1.0 {
		t.Errorf("expected cost $0.30-1.00, got $%.2f", cost)
	}

	// Small model should be cheaper
	costSmall := EstimateCost(8, 3, "haiku-4.5")
	if costSmall >= cost {
		t.Errorf("expected small model cheaper: $%.2f >= $%.2f", costSmall, cost)
	}
}

func assertClose(t *testing.T, got, want, tolerance float64, name string) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("%s: got %.3f, want %.3f (tolerance %.3f)", name, got, want, tolerance)
	}
}
