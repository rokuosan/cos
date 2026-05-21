package tomlsync

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Block is a contiguous part of a TOML document.
//
// A block is either a table block, a root key/value block, or unstructured
// whitespace/comment text that should be carried with nearby content.
type Block struct {
	Path   []string
	Key    string
	Text   string
	RootKV bool
}

// Document is a TOML document split into mergeable text blocks.
type Document struct {
	Blocks []Block
}

// LoadDocument reads and parses a TOML config document from path.
//
// A leading "~/" is expanded to the current user's home directory.
func LoadDocument(path string) (Document, error) {
	data, err := os.ReadFile(expandHome(path))
	if err != nil {
		return Document{}, fmt.Errorf("read config: %w", err)
	}
	return ParseDocument(strings.NewReader(string(data)))
}

// ParseDocument reads TOML text and splits it into blocks without normalizing
// formatting.
func ParseDocument(r io.Reader) (Document, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	var blocks []Block
	var pending []string
	var current *Block
	multilineDepth := 0

	flushCurrent := func() {
		if current == nil {
			return
		}
		blocks = append(blocks, *current)
		current = nil
	}
	flushPending := func() {
		if len(pending) == 0 {
			return
		}
		blocks = append(blocks, Block{Text: strings.Join(pending, "")})
		pending = nil
	}

	for scanner.Scan() {
		line := scanner.Text() + "\n"
		trimmed := strings.TrimSpace(line)

		if multilineDepth == 0 && isTableHeader(trimmed) {
			flushCurrent()
			flushPending()

			path, err := parseHeaderPath(trimmed)
			if err != nil {
				return Document{}, err
			}
			current = &Block{Path: path, Text: line}
			continue
		}

		if multilineDepth == 0 {
			if key, ok := parseRootKey(trimmed); ok && (current == nil || current.RootKV) {
				flushCurrent()
				current = &Block{
					Key:    key,
					Text:   strings.Join(append(pending, line), ""),
					RootKV: true,
				}
				pending = nil
				multilineDepth += bracketDelta(trimmed)
				continue
			}
		}

		if current != nil {
			current.Text += line
			multilineDepth += bracketDelta(trimmed)
			continue
		}

		pending = append(pending, line)
	}
	if err := scanner.Err(); err != nil {
		return Document{}, err
	}

	flushCurrent()
	flushPending()
	return Document{Blocks: blocks}, nil
}

// String renders the document by concatenating its original text blocks.
func (d Document) String() string {
	var b strings.Builder
	for i, block := range d.Blocks {
		if i > 0 && block.Text != "" && b.Len() > 0 && !strings.HasSuffix(b.String(), "\n\n") {
			if !strings.HasSuffix(b.String(), "\n") {
				b.WriteByte('\n')
			}
			b.WriteByte('\n')
		}
		b.WriteString(block.Text)
	}
	return b.String()
}

func isTableHeader(trimmed string) bool {
	_, end, ok := tableHeaderContent(trimmed)
	if !ok {
		return false
	}
	rest := strings.TrimSpace(trimmed[end:])
	return rest == "" || strings.HasPrefix(rest, "#")
}

func parseHeaderPath(trimmed string) ([]string, error) {
	content, _, ok := tableHeaderContent(trimmed)
	if !ok {
		return nil, fmt.Errorf("invalid TOML table header: %s", trimmed)
	}
	return splitTOMLPath(content)
}

func tableHeaderContent(trimmed string) (string, int, bool) {
	switch {
	case strings.HasPrefix(trimmed, "[["):
		return bracketContent(trimmed, 2, true)
	case strings.HasPrefix(trimmed, "["):
		return bracketContent(trimmed, 1, false)
	default:
		return "", 0, false
	}
}

func bracketContent(input string, start int, arrayTable bool) (string, int, bool) {
	quote := rune(0)
	escaped := false
	for i, r := range input {
		if i < start {
			continue
		}
		switch {
		case escaped:
			escaped = false
		case r == '\\' && quote == '"':
			escaped = true
		case r == '"' && quote == 0:
			quote = r
		case r == '"' && quote == r:
			quote = 0
		case r == '\'' && quote == 0:
			quote = r
		case r == '\'' && quote == r:
			quote = 0
		case r == ']' && quote == 0 && arrayTable:
			if i+1 < len(input) && input[i+1] == ']' {
				return input[start:i], i + 2, true
			}
		case r == ']' && quote == 0:
			return input[start:i], i + 1, true
		}
	}
	return "", 0, false
}

func bracketDelta(line string) int {
	line = stripComment(line)
	quote := rune(0)
	escaped := false
	depth := 0
	for _, r := range line {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && quote == '"':
			escaped = true
		case r == '"' && quote == 0:
			quote = r
		case r == '"' && quote == r:
			quote = 0
		case r == '\'' && quote == 0:
			quote = r
		case r == '\'' && quote == r:
			quote = 0
		case r == '[' && quote == 0:
			depth++
		case r == ']' && quote == 0:
			depth--
		}
	}
	return depth
}

func stripComment(line string) string {
	quote := rune(0)
	escaped := false
	for i, r := range line {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && quote == '"':
			escaped = true
		case r == '"' && quote == 0:
			quote = r
		case r == '"' && quote == r:
			quote = 0
		case r == '\'' && quote == 0:
			quote = r
		case r == '\'' && quote == r:
			quote = 0
		case r == '#' && quote == 0:
			return line[:i]
		}
	}
	return line
}

func splitTOMLPath(path string) ([]string, error) {
	var parts []string
	var b strings.Builder
	quote := rune(0)
	escaped := false

	for _, r := range path {
		switch {
		case escaped:
			b.WriteRune(r)
			escaped = false
		case r == '\\' && quote == '"':
			escaped = true
		case r == '"' && quote == 0:
			quote = r
		case r == '"' && quote == r:
			quote = 0
		case r == '\'' && quote == 0:
			quote = r
		case r == '\'' && quote == r:
			quote = 0
		case r == '.' && quote == 0:
			part := strings.TrimSpace(b.String())
			if part == "" {
				return nil, fmt.Errorf("invalid TOML path: %s", path)
			}
			parts = append(parts, part)
			b.Reset()
		default:
			b.WriteRune(r)
		}
	}

	part := strings.TrimSpace(b.String())
	if part == "" {
		return nil, fmt.Errorf("invalid TOML path: %s", path)
	}
	parts = append(parts, part)
	return parts, nil
}

func parseRootKey(trimmed string) (string, bool) {
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "[") {
		return "", false
	}
	key, _, ok := strings.Cut(trimmed, "=")
	if !ok {
		return "", false
	}
	key = strings.TrimSpace(key)
	if key == "" || strings.ContainsAny(key, " \t") {
		return "", false
	}
	return key, true
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
