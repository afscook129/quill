package discovery

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SkillInfo represents a discovered SKILL.md file with parsed frontmatter.
type SkillInfo struct {
	Path        string
	Name        string
	Description string
	Version     string
	Dir         string
}

// skipDirs are directories that should never be scanned.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".quill":       true,
	"dist":         true,
	"build":        true,
}

// Scan walks the directory tree from root and finds all SKILL.md files,
// parsing their YAML frontmatter for metadata.
func Scan(root string) ([]SkillInfo, error) {
	var skills []SkillInfo

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Name() == "SKILL.md" {
			skill, err := parseSkillMD(path)
			if err != nil {
				return nil // skip unparseable files
			}
			skills = append(skills, *skill)
		}
		return nil
	})

	return skills, err
}

func parseSkillMD(path string) (*SkillInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	skill := &SkillInfo{
		Path: path,
		Dir:  filepath.Dir(path),
		Name: filepath.Base(filepath.Dir(path)),
	}

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if lineNum == 1 && strings.TrimSpace(line) == "---" {
			inFrontmatter = true
			continue
		}

		if inFrontmatter {
			if strings.TrimSpace(line) == "---" {
				break
			}
			parseFrontmatterLine(line, skill)
		}
	}

	if skill.Name == "" {
		skill.Name = filepath.Base(filepath.Dir(path))
	}

	return skill, nil
}

func parseFrontmatterLine(line string, skill *SkillInfo) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Strip quotes
	value = strings.Trim(value, `"'`)

	switch key {
	case "name":
		skill.Name = value
	case "description":
		// Handle multi-line descriptions with > indicator
		if value != "" && value != ">" {
			skill.Description = value
		}
	case "version":
		skill.Version = value
	}
}

// FormatSkillList renders a list of discovered skills for display.
func FormatSkillList(skills []SkillInfo) string {
	if len(skills) == 0 {
		return "  no skills found"
	}

	var sb strings.Builder
	for _, s := range skills {
		name := s.Name
		if s.Version != "" {
			name = fmt.Sprintf("%s@%s", name, s.Version)
		}
		relPath := s.Dir
		sb.WriteString(fmt.Sprintf("    %s  (%s)\n", name, relPath))
	}
	return sb.String()
}
