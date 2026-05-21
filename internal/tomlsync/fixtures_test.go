package tomlsync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSynthesizeFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		preserve []PreserveRule
	}{
		{name: "generic-merge"},
		{name: "codex-preserve", preserve: []PreserveRule{"projects.*"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join("testdata", "synthesize", tc.name)
			source := loadFixtureDocument(t, filepath.Join(dir, "source.toml"))
			target := loadFixtureDocument(t, filepath.Join(dir, "target.toml"))
			want := loadFixtureText(t, filepath.Join(dir, "expect.toml"))

			got := Synthesize(source, target, tc.preserve).String()
			if got != want {
				t.Fatalf("unexpected synthesized document\nwant:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}

func loadFixtureDocument(t *testing.T, path string) Document {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	doc, err := ParseDocument(f)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func loadFixtureText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
