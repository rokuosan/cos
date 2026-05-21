package codexconfig

import "testing"

func TestParseDocumentAllowsTrailingCommentOnTableHeader(t *testing.T) {
	doc := parseDoc(t, `[projects."/repo"] # trusted repo
trust_level = "trusted"
`)
	if len(doc.Blocks) != 1 {
		t.Fatalf("expected one block, got %d", len(doc.Blocks))
	}
	if len(doc.Blocks[0].Path) != 2 ||
		doc.Blocks[0].Path[0] != "projects" ||
		doc.Blocks[0].Path[1] != "/repo" {
		t.Fatalf("unexpected table path: %#v", doc.Blocks[0].Path)
	}
}

func TestParseDocumentReadsArrayTableHeader(t *testing.T) {
	doc := parseDoc(t, `[[plugins.instances]]
name = "github"
`)
	if len(doc.Blocks) != 1 {
		t.Fatalf("expected one block, got %d", len(doc.Blocks))
	}
	if len(doc.Blocks[0].Path) != 2 ||
		doc.Blocks[0].Path[0] != "plugins" ||
		doc.Blocks[0].Path[1] != "instances" {
		t.Fatalf("unexpected array table path: %#v", doc.Blocks[0].Path)
	}
}

func TestParseDocumentKeepsHashInsideRootStringValue(t *testing.T) {
	doc := parseDoc(t, `model = "gpt-5.5#fast"
approval_policy = "never"
`)
	if len(doc.Blocks) != 2 {
		t.Fatalf("expected two blocks, got %d", len(doc.Blocks))
	}
	if got := doc.Blocks[0].Text; got != "model = \"gpt-5.5#fast\"\n" {
		t.Fatalf("unexpected first block text: %q", got)
	}
}

func TestDocumentStringRoundTripsSupportedShape(t *testing.T) {
	input := `model = "gpt-5.5"

[projects."/repo"] # comment
tags = [
  ["tag"]
]
trust_level = "trusted"
`
	doc := parseDoc(t, input)
	if got := doc.String(); got != input {
		t.Fatalf("unexpected round trip\nwant:\n%s\ngot:\n%s", input, got)
	}
}
