package codexconfig

import (
	"io"
	"strings"

	"github.com/rokuosan/cos/internal/tomlsync"
)

// Config is a parsed Codex config file.
type Config struct {
	Document tomlsync.Document
}

// Project is a Codex project entry from the config file.
type Project struct {
	Path       string
	TrustLevel string
}

// ParseConfig reads and parses a Codex config file from r.
func ParseConfig(r io.Reader) (Config, error) {
	doc, err := tomlsync.ParseDocument(r)
	if err != nil {
		return Config{}, err
	}
	return Config{Document: doc}, nil
}

// Load reads and parses a Codex config file from path.
func Load(path string) (Config, error) {
	doc, err := tomlsync.LoadDocument(path)
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
