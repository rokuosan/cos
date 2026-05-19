package codexconfig

import (
	"strings"
	"testing"
)

func TestParseDocumentReadsRootKeysAndTables(t *testing.T) {
	doc := parseDoc(t, `model = "gpt-5.5"

[plugins."github@openai-curated"]
enabled = true
`)
	if len(doc.Blocks) != 2 {
		t.Fatalf("expected two blocks, got %d", len(doc.Blocks))
	}
	if !doc.Blocks[0].RootKV || doc.Blocks[0].Key != "model" {
		t.Fatalf("unexpected root block: %#v", doc.Blocks[0])
	}
	got := strings.Join(doc.Blocks[1].Path, "|")
	want := "plugins|github@openai-curated"
	if got != want {
		t.Fatalf("unexpected path: want %q got %q", want, got)
	}
}

func TestRootMultilineValueStaysInOneBlock(t *testing.T) {
	doc := parseDoc(t, `items = [
  "a",
]
model = "new"
`)
	if len(doc.Blocks) != 2 {
		t.Fatalf("expected two blocks, got %d", len(doc.Blocks))
	}
	if doc.Blocks[0].Key != "items" || !strings.Contains(doc.Blocks[0].Text, `"a"`) {
		t.Fatalf("unexpected multiline root block: %#v", doc.Blocks[0])
	}
}

func parseDoc(t *testing.T, input string) Document {
	t.Helper()
	doc, err := ParseDocument(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}
