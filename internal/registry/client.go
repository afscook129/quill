package registry

import "os"

// Client is the interface for registry operations.
// The registry backend runs separately (Convex in quill-cloud repo).
// This interface defines what the CLI needs from it.
type Client interface {
	Search(query string, model string) (*SearchResponse, error)
	Resolve(name string, version string) (*SkillResult, error)
}

// BaseURL returns the registry base URL from env or default.
func BaseURL() string {
	if url := os.Getenv("QUILL_REGISTRY"); url != "" {
		return url
	}
	return "https://registry.quill.dev/v1/"
}
