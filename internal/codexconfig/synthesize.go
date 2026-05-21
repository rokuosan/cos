package codexconfig

import "strings"

// PreserveRule is a dotted TOML path pattern for target-owned config blocks.
//
// A "*" segment matches any single path segment. For example, "projects.*"
// matches tables such as [projects."/path/to/repo"].
type PreserveRule string

// Synthesize builds a config document from source and preserved target blocks.
func Synthesize(source, target Document, preserve []PreserveRule) Document {
	var sourceBlocks []Block
	var preservedRoot []Block
	var preservedTables []Block

	for _, block := range source.Blocks {
		if block.matchesAny(preserve) {
			continue
		}
		sourceBlocks = append(sourceBlocks, block)
	}
	for _, block := range target.Blocks {
		if !block.matchesAny(preserve) {
			continue
		}
		if block.RootKV {
			preservedRoot = append(preservedRoot, block)
			continue
		}
		preservedTables = append(preservedTables, block)
	}

	insertAt := len(sourceBlocks)
	for i, block := range sourceBlocks {
		if len(block.Path) > 0 {
			insertAt = i
			break
		}
	}

	blocks := make([]Block, 0, len(sourceBlocks)+len(preservedRoot)+len(preservedTables))
	blocks = append(blocks, sourceBlocks[:insertAt]...)
	blocks = append(blocks, preservedRoot...)
	blocks = append(blocks, sourceBlocks[insertAt:]...)
	blocks = append(blocks, preservedTables...)
	return Document{Blocks: blocks}
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

func (b Block) matchesAny(rules []PreserveRule) bool {
	for _, rule := range rules {
		if b.matches(rule) {
			return true
		}
	}
	return false
}

func (b Block) matches(rule PreserveRule) bool {
	rulePath := splitPreserveRule(rule)
	if len(rulePath) == 0 {
		return false
	}
	if b.RootKV {
		return len(rulePath) == 1 && rulePath[0] == b.Key
	}
	if len(b.Path) == 0 {
		return false
	}
	return pathMatches(rulePath, b.Path)
}

func pathMatches(rule, path []string) bool {
	if len(rule) > len(path) {
		return false
	}
	for i := range rule {
		if rule[i] == "*" {
			continue
		}
		if rule[i] != path[i] {
			return false
		}
	}
	return true
}

func splitPreserveRule(rule PreserveRule) []string {
	parts := strings.Split(strings.TrimSpace(string(rule)), ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
