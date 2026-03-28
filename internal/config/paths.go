package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the Quill config directory (~/.quill).
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "quill")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".quill"
	}
	return filepath.Join(home, ".quill")
}

// DataDir returns the Quill data directory.
func DataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "quill")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".quill/data"
	}
	return filepath.Join(home, ".local", "share", "quill")
}

// ProjectDir returns the .quill directory in the current project.
func ProjectDir() string {
	return ".quill"
}
