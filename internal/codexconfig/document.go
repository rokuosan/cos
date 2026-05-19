package codexconfig

import (
	"bufio"
	"fmt"
	"io"
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

// ParseDocument reads TOML text and splits it into blocks without normalizing
// formatting.
//
// This parser intentionally recognizes only the structure needed to inspect a
// Codex config: root key/value entries and table headers. It does not validate
// the full TOML grammar.
func ParseDocument(r io.Reader) (Document, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	var blocks []Block
	var pending []string
	var current *Block

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

		if isTableHeader(trimmed) {
			flushCurrent()
			flushPending()

			path, err := parseHeaderPath(trimmed)
			if err != nil {
				return Document{}, err
			}
			current = &Block{Path: path, Text: line}
			continue
		}

		if key, ok := parseRootKey(trimmed); ok && (current == nil || current.RootKV) {
			flushCurrent()
			current = &Block{
				Key:    key,
				Text:   strings.Join(append(pending, line), ""),
				RootKV: true,
			}
			pending = nil
			continue
		}

		if current != nil {
			current.Text += line
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
		content, end, ok := bracketContent(trimmed, 2, true)
		return content, end, ok
	case strings.HasPrefix(trimmed, "["):
		content, end, ok := bracketContent(trimmed, 1, false)
		return content, end, ok
	default:
		return "", 0, false
	}
}

func bracketContent(input string, start int, arrayTable bool) (string, int, bool) {
	inQuote := false
	escaped := false
	for i := start; i < len(input); i++ {
		switch {
		case escaped:
			escaped = false
		case input[i] == '\\' && inQuote:
			escaped = true
		case input[i] == '"':
			inQuote = !inQuote
		case input[i] == ']' && !inQuote && arrayTable:
			if i+1 < len(input) && input[i+1] == ']' {
				return input[start:i], i + 2, true
			}
		case input[i] == ']' && !inQuote:
			return input[start:i], i + 1, true
		}
	}
	return "", 0, false
}

func splitTOMLPath(path string) ([]string, error) {
	var parts []string
	var b strings.Builder
	inQuote := false
	escaped := false

	for _, r := range path {
		switch {
		case escaped:
			b.WriteRune(r)
			escaped = false
		case r == '\\' && inQuote:
			escaped = true
		case r == '"':
			inQuote = !inQuote
		case r == '.' && !inQuote:
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
	if inQuote {
		return nil, fmt.Errorf("unterminated quote in TOML path: %s", path)
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

	inQuote := false
	escaped := false
	for i, r := range trimmed {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && inQuote:
			escaped = true
		case r == '"':
			inQuote = !inQuote
		case r == '=' && !inQuote:
			key := strings.TrimSpace(trimmed[:i])
			if key == "" {
				return "", false
			}
			return strings.Trim(key, `"`), true
		}
	}
	return "", false
}
