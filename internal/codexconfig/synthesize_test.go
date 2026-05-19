package codexconfig

import (
	"strings"
	"testing"
)

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
