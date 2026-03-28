package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultModel   string `yaml:"default_model,omitempty"`
	SignalsEnabled bool   `yaml:"signals_enabled"`
	Registry       string `yaml:"registry,omitempty"`
}

func Default() *Config {
	return &Config{
		SignalsEnabled: true,
		Registry:      "https://registry.quill.dev",
	}
}

func Load() (*Config, error) {
	path := filepath.Join(ConfigDir(), "config.yaml")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func Parse(r io.Reader) (*Config, error) {
	var c Config
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &c, nil
}

func Save(c *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "config.yaml")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return err
	}
	return enc.Close()
}

// DetectModel detects the default model from environment variables.
// Priority: QUILL_DEFAULT_MODEL > ANTHROPIC_API_KEY > OPENAI_API_KEY
func DetectModel() (model string, source string) {
	if m := os.Getenv("QUILL_DEFAULT_MODEL"); m != "" {
		return m, "QUILL_DEFAULT_MODEL"
	}
	hasAnthropic := os.Getenv("ANTHROPIC_API_KEY") != ""
	hasOpenAI := os.Getenv("OPENAI_API_KEY") != ""
	if hasAnthropic {
		return "claude-sonnet-4-6", "ANTHROPIC_API_KEY"
	}
	if hasOpenAI {
		return "gpt-5.4", "OPENAI_API_KEY"
	}
	return "", ""
}
