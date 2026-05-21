package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunRoutesTopLevelSync(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.toml")
	if err := os.WriteFile(source, []byte("model = \"gpt-5.5\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run(
		[]string{"sync", "--source", source, "--target", filepath.Join(dir, "missing.toml")},
		&stdout,
		&stderr,
	); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "model = \"gpt-5.5\"\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run([]string{"unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUsageShowsTopLevelSync(t *testing.T) {
	var stderr bytes.Buffer
	if err := usage(&stderr); err != nil {
		t.Fatal(err)
	}
	if got := stderr.String(); got != "usage: cos <sync|codex-config read> ...\n" {
		t.Fatalf("unexpected usage: %q", got)
	}
}
