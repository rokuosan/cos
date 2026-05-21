package tomlsync

import (
	"strings"
	"testing"
)

func TestSynthesizeUsesSourceForOverlappingRootKeysAndPreservesTargetOnlyKeys(t *testing.T) {
	source := parseDoc(t, `a = "A"
b = "B"
c = "C"
`)
	target := parseDoc(t, `a = "AA"
b = "BBB"
c = "C"
d = "D"
`)

	got := Synthesize(source, target, nil).String()
	want := `a = "A"

b = "B"

c = "C"

d = "D"
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizeMergesTableEntriesRecursively(t *testing.T) {
	source := parseDoc(t, `[foo]
a = "A"

[foo.bar]
c = "C"
`)
	target := parseDoc(t, `[foo]
a = "AA"
b = "B"

[foo.bar]
c = "old"
d = "D"

[foo.baz]
e = "E"
`)

	got := Synthesize(source, target, nil).String()
	want := `[foo]
a = "A"

b = "B"

[foo.bar]
c = "C"

d = "D"

[foo.baz]
e = "E"
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizePreservesCommentsOnTargetOnlyTableEntries(t *testing.T) {
	source := parseDoc(t, `[foo]
a = "A"
`)
	target := parseDoc(t, `[foo]
# keep me
b = [
  "B",
]
`)

	got := Synthesize(source, target, nil).String()
	want := `[foo]
a = "A"

# keep me
b = [
  "B",
]
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizeMergesArrayTableInstancesByOccurrence(t *testing.T) {
	source := parseDoc(t, `[[plugins.instances]]
name = "github"
enabled = true
`)
	target := parseDoc(t, `[[plugins.instances]]
name = "github"
timeout = 5

[[plugins.instances]]
name = "slack"
enabled = false
`)

	got := Synthesize(source, target, nil).String()
	want := `[[plugins.instances]]
name = "github"

enabled = true

timeout = 5

[[plugins.instances]]
name = "slack"
enabled = false
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizeDoesNotSplitTripleQuotedTableValue(t *testing.T) {
	source := parseDoc(t, `[foo]
message = """
left = right
"""
enabled = true
`)
	target := parseDoc(t, `[foo]
message = """
old = value
"""
extra = "keep"
`)

	got := Synthesize(source, target, nil).String()
	want := `[foo]
message = """
left = right
"""

enabled = true

extra = "keep"
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizeDoesNotEndTripleQuotedBasicStringOnEscapedDelimiter(t *testing.T) {
	source := parseDoc(t, `[foo]
message = """
keep \""" inside
still here
"""
enabled = true
`)
	target := parseDoc(t, `[foo]
message = """
old
"""
extra = "keep"
`)

	got := Synthesize(source, target, nil).String()
	want := `[foo]
message = """
keep \""" inside
still here
"""

enabled = true

extra = "keep"
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizePreservesProjectsFromTarget(t *testing.T) {
	source := parseDoc(t, `model = "gpt-5.5"

[projects."/source-only"]
trust_level = "trusted"

[features]
apps = true
`)
	target := parseDoc(t, `model = "old"

[projects."/target"]
trust_level = "trusted"
`)

	got := Synthesize(source, target, []PreserveRule{"projects.*"}).String()
	want := `model = "gpt-5.5"

[features]
apps = true

[projects."/target"]
trust_level = "trusted"
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizePreservesRootKeyFromTarget(t *testing.T) {
	source := parseDoc(t, `model = "gpt-5.5"
approval_policy = "untrusted"
`)
	target := parseDoc(t, `model = "old"
approval_policy = "never"
`)

	got := Synthesize(source, target, []PreserveRule{"approval_policy"}).String()
	want := `model = "gpt-5.5"

approval_policy = "never"
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizePreservesQuotedRootKeyAcrossQuotingStyles(t *testing.T) {
	source := parseDoc(t, `"approval policy" = "untrusted"
`)
	target := parseDoc(t, `approval policy = "never"
`)

	got := Synthesize(source, target, []PreserveRule{`"approval policy"`}).String()
	want := `approval policy = "never"
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizeKeepsPreservedRootKeyBeforeTables(t *testing.T) {
	source := parseDoc(t, `model = "gpt-5.5"
approval_policy = "untrusted"

[features]
apps = true
`)
	target := parseDoc(t, `approval_policy = "never"
`)

	got := Synthesize(source, target, []PreserveRule{"approval_policy"}).String()
	want := `model = "gpt-5.5"

approval_policy = "never"

[features]
apps = true
`
	if got != want {
		t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSynthesizePreservesParentTableFromTarget(t *testing.T) {
	source := parseDoc(t, `[tui]
status_line = ["model"]

[tui.model_availability_nux]
"gpt-5.5" = 4
`)
	target := parseDoc(t, `[tui]
status_line = ["current-dir"]

[tui.model_availability_nux]
"gpt-5.5" = 10
`)

	got := Synthesize(source, target, []PreserveRule{"tui"}).String()
	if strings.Contains(got, `status_line = ["model"]`) {
		t.Fatalf("source tui block was not removed:\n%s", got)
	}
	if !strings.Contains(got, `status_line = ["current-dir"]`) {
		t.Fatalf("target tui block was not preserved:\n%s", got)
	}
	if !strings.Contains(got, `"gpt-5.5" = 10`) {
		t.Fatalf("target child table was not preserved:\n%s", got)
	}
}
