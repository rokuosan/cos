package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunRejectsUnknownCommand(t *testing.T) {
	err := run([]string{"unknown"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunCodexConfigSyncRequiresSource(t *testing.T) {
	var out bytes.Buffer
	err := runCodexConfigSync(nil, &out)
	if err == nil || err.Error() != "--source is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCodexConfigSyncSynthesizesToStdout(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.toml")
	target := filepath.Join(dir, "target.toml")
	if err := os.WriteFile(source, []byte(`model = "gpt-5.5"

[features]
apps = true
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(`[projects."/repo"]
trust_level = "trusted"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := runCodexConfigSync([]string{"--source", source, "--target", target}, &out)
	if err != nil {
		t.Fatal(err)
	}

	want := `model = "gpt-5.5"

[features]
apps = true

[projects."/repo"]
trust_level = "trusted"
`
	if out.String() != want {
		t.Fatalf("unexpected output\nwant:\n%s\ngot:\n%s", want, out.String())
	}
}

func TestRunCodexConfigSyncAllowsMissingTarget(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.toml")
	target := filepath.Join(dir, "missing.toml")
	if err := os.WriteFile(source, []byte(`model = "gpt-5.5"

[features]
apps = true
`), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := runCodexConfigSync([]string{"--source", source, "--target", target}, &out)
	if err != nil {
		t.Fatal(err)
	}

	want := `model = "gpt-5.5"

[features]
apps = true
`
	if out.String() != want {
		t.Fatalf("unexpected output\nwant:\n%s\ngot:\n%s", want, out.String())
	}
}
