package codexconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config is a parsed Codex config file.
type Config struct {
	Document Document
}

// Project is a Codex project entry from the config file.
type Project struct {
	Path       string
	TrustLevel string
}

// LoadDocument reads and parses a TOML config document from path.
//
// A leading "~/" is expanded to the current user's home directory.
func LoadDocument(path string) (Document, error) {
	data, err := os.ReadFile(expandHome(path))
	if err != nil {
		return Document{}, fmt.Errorf("read config: %w", err)
	}
	doc, err := ParseDocument(strings.NewReader(string(data)))
	if err != nil {
		return Document{}, err
	}
	return doc, nil
}

// Load reads and parses a Codex config file from path.
func Load(path string) (Config, error) {
	doc, err := LoadDocument(path)
	if err != nil {
		return Config{}, err
	}
	return Config{Document: doc}, nil
}

// Projects returns project trust entries from the config file.
func (c Config) Projects() []Project {
	var projects []Project
	for _, block := range c.Document.Blocks {
		if len(block.Path) != 2 || block.Path[0] != "projects" {
			continue
		}
		projects = append(projects, Project{
			Path:       block.Path[1],
			TrustLevel: tableStringValue(block.Text, "trust_level"),
		})
	}
	return projects
}

func tableStringValue(text, wantKey string) string {
	for raw := range strings.SplitSeq(text, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(raw), "=")
		if !ok || strings.TrimSpace(key) != wantKey {
			continue
		}
		value = strings.TrimSpace(stripInlineComment(value))
		return strings.Trim(value, `"`)
	}
	return ""
}

func stripInlineComment(line string) string {
	inQuote := false
	escaped := false
	for i, r := range line {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && inQuote:
			escaped = true
		case r == '"':
			inQuote = !inQuote
		case r == '#' && !inQuote:
			return strings.TrimSpace(line[:i])
		}
	}
	return line
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if after, ok := strings.CutPrefix(path, "~/"); ok {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, after)
		}
	}
	return path
}
