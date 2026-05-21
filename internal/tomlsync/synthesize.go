package tomlsync

import "strings"

// PreserveRule is a dotted TOML path pattern for target-owned config blocks.
//
// A "*" segment matches any single path segment. For example, "projects.*"
// matches tables such as [projects."/path/to/repo"].
type PreserveRule string

// Synthesize builds a config document from source and preserved target blocks.
func Synthesize(source, target Document, preserve []PreserveRule) Document {
	sourceIndex := indexDocument(source)
	targetIndex := indexDocument(target)

	var mergedRoot []Block
	var mergedTables []Block
	seenRoot := map[string]struct{}{}
	consumedTables := map[string]int{}

	for _, block := range sourceIndex.rootOrder {
		key := block.Key
		seenRoot[key] = struct{}{}
		targetBlock, hasTarget := targetIndex.rootByKey[key]

		switch {
		case block.matchesAny(preserve):
			if hasTarget {
				mergedRoot = append(mergedRoot, targetBlock)
			}
		case hasTarget:
			mergedRoot = append(mergedRoot, block)
		default:
			mergedRoot = append(mergedRoot, block)
		}
	}
	for _, block := range targetIndex.rootOrder {
		key := block.Key
		if _, ok := seenRoot[key]; ok {
			continue
		}
		mergedRoot = append(mergedRoot, block)
	}

	for _, block := range sourceIndex.tableOrder {
		pathKey := blockPathKey(block.Path)
		targetBlock, hasTarget := targetIndex.nextTable(pathKey)

		switch {
		case block.matchesAny(preserve):
			if hasTarget {
				mergedTables = append(mergedTables, targetBlock)
				consumedTables[pathKey]++
			}
		case hasTarget:
			mergedTables = append(mergedTables, mergeTableBlock(block, targetBlock))
			consumedTables[pathKey]++
		default:
			mergedTables = append(mergedTables, block)
		}
	}
	appendedTables := map[string]int{}
	for _, block := range targetIndex.tableOrder {
		pathKey := blockPathKey(block.Path)
		if appendedTables[pathKey] < consumedTables[pathKey] {
			appendedTables[pathKey]++
			continue
		}
		mergedTables = append(mergedTables, block)
	}

	blocks := make([]Block, 0, len(mergedRoot)+len(mergedTables))
	blocks = append(blocks, mergedRoot...)
	blocks = append(blocks, mergedTables...)
	return Document{Blocks: blocks}
}

type documentIndex struct {
	rootOrder    []Block
	rootByKey    map[string]Block
	tableOrder   []Block
	tableByPath  map[string][]Block
	tableIndices map[string]int
}

func indexDocument(doc Document) documentIndex {
	out := documentIndex{
		rootByKey:    map[string]Block{},
		tableByPath:  map[string][]Block{},
		tableIndices: map[string]int{},
	}
	for _, block := range doc.Blocks {
		switch {
		case block.RootKV:
			out.rootOrder = append(out.rootOrder, block)
			out.rootByKey[block.Key] = block
		case len(block.Path) > 0:
			out.tableOrder = append(out.tableOrder, block)
			key := blockPathKey(block.Path)
			out.tableByPath[key] = append(out.tableByPath[key], block)
		}
	}
	return out
}

func (i *documentIndex) nextTable(pathKey string) (Block, bool) {
	blocks := i.tableByPath[pathKey]
	index := i.tableIndices[pathKey]
	if index >= len(blocks) {
		return Block{}, false
	}
	i.tableIndices[pathKey] = index + 1
	return blocks[index], true
}

func blockPathKey(path []string) string {
	return strings.Join(path, "\x00")
}

func mergeTableBlock(source, target Block) Block {
	sourceHeader, sourceBody := splitTableBlock(source.Text)
	_, targetBody := splitTableBlock(target.Text)

	sourceEntries, sourceTrailer := parseTableEntries(sourceBody)
	targetEntries, _ := parseTableEntries(targetBody)

	var body strings.Builder
	seen := map[string]struct{}{}
	for _, entry := range sourceEntries {
		seen[entry.Key] = struct{}{}
		appendSection(&body, entry.Text)
	}
	for _, entry := range targetEntries {
		if _, ok := seen[entry.Key]; ok {
			continue
		}
		appendSection(&body, entry.Text)
	}
	body.WriteString(sourceTrailer)

	return Block{
		Path: source.Path,
		Text: sourceHeader + body.String(),
	}
}

type tableEntry struct {
	Key  string
	Text string
}

func splitTableBlock(text string) (string, string) {
	idx := strings.IndexByte(text, '\n')
	if idx == -1 {
		return text, ""
	}
	return text[:idx+1], text[idx+1:]
}

func parseTableEntries(body string) ([]tableEntry, string) {
	lines := strings.SplitAfter(body, "\n")
	var entries []tableEntry
	var pending []string
	var current *tableEntry
	state := tableParseState{}

	flushCurrent := func() {
		if current == nil {
			return
		}
		entries = append(entries, *current)
		current = nil
	}

	for _, line := range lines {
		if line == "" {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if state.canStartEntry() {
			if key, ok := parseRootKey(trimmed); ok {
				flushCurrent()
				current = &tableEntry{
					Key:  key,
					Text: strings.Join(append(pending, line), ""),
				}
				pending = nil
				state.consumeLine(line)
				continue
			}
		}
		if current != nil {
			current.Text += line
			state.consumeLine(line)
			continue
		}
		pending = append(pending, line)
		state.consumeLine(line)
	}
	flushCurrent()
	return entries, strings.Join(pending, "")
}

type tableParseState struct {
	bracketDepth   int
	multilineQuote string
}

func (s tableParseState) canStartEntry() bool {
	return s.bracketDepth == 0 && s.multilineQuote == ""
}

func (s *tableParseState) consumeLine(line string) {
	if line == "" {
		return
	}

	quote := rune(0)
	escaped := false
	for i := 0; i < len(line); i++ {
		if s.multilineQuote != "" {
			if strings.HasPrefix(line[i:], s.multilineQuote) &&
				!isEscapedDelimiter(line, i, s.multilineQuote) {
				i += len(s.multilineQuote) - 1
				s.multilineQuote = ""
			}
			continue
		}

		r := rune(line[i])
		switch {
		case escaped:
			escaped = false
		case quote == '"' && r == '\\':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case strings.HasPrefix(line[i:], `"""`):
			s.multilineQuote = `"""`
			i += 2
		case strings.HasPrefix(line[i:], `'''`):
			s.multilineQuote = `'''`
			i += 2
		case r == '"':
			quote = r
		case r == '\'':
			quote = r
		case r == '#':
			return
		case r == '[':
			s.bracketDepth++
		case r == ']':
			s.bracketDepth--
		}
	}
}

func isEscapedDelimiter(line string, index int, delimiter string) bool {
	if delimiter != `"""` {
		return false
	}
	backslashes := 0
	for i := index - 1; i >= 0 && line[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func appendSection(b *strings.Builder, text string) {
	if text == "" {
		return
	}
	if b.Len() > 0 && !strings.HasPrefix(text, "\n") && !strings.HasSuffix(b.String(), "\n\n") {
		if !strings.HasSuffix(b.String(), "\n") {
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	b.WriteString(text)
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
		return len(rulePath) == 1 && normalizeRootKey(rulePath[0]) == b.Key
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
