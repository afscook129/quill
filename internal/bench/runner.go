package bench

// Runner defines the interface for running benchmarks.
// The actual implementation will call provider APIs in Go.
type Runner interface {
	Run(skillPath string, model string, trials int) (*Result, error)
}
