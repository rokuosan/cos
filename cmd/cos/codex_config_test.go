package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultCodexConfigPathUsesCODEXHOME(t *testing.T) {
	t.Setenv("CODEX_HOME", "/tmp/codex-home")
	got := defaultCodexConfigPath()
	want := filepath.Join("/tmp/codex-home", "config.toml")
	if got != want {
		t.Fatalf("unexpected path: want %q got %q", want, got)
	}
}

func TestDefaultCodexConfigPathFallsBackToHomeConfig(t *testing.T) {
	t.Setenv("CODEX_HOME", "")
	got := defaultCodexConfigPath()
	if got != "~/.codex/config.toml" {
		t.Fatalf("unexpected path: %q", got)
	}
}

func TestRunSyncRequiresSource(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSync(nil, &stdout, &stderr)
	if err == nil || err.Error() != "--source is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSyncRequiresTargetWithoutCodexMode(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.toml")
	if err := os.WriteFile(source, []byte(`model = "gpt-5.5"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSync([]string{"--source", source}, &stdout, &stderr)
	if err == nil || err.Error() != "--target is required unless --codex is set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSyncSynthesizesToStdout(t *testing.T) {
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

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSync([]string{"--source", source, "--target", target}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	want := `model = "gpt-5.5"

[features]
apps = true

[projects."/repo"]
trust_level = "trusted"
`
	if stdout.String() != want {
		t.Fatalf("unexpected output\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestRunSyncAllowsMissingTarget(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.toml")
	target := filepath.Join(dir, "missing.toml")
	if err := os.WriteFile(source, []byte(`model = "gpt-5.5"

[features]
apps = true
`), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSync([]string{"--source", source, "--target", target}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	want := `model = "gpt-5.5"

[features]
apps = true
`
	if stdout.String() != want {
		t.Fatalf("unexpected output\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestRunSyncWriteUpdatesTarget(t *testing.T) {
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
`), 0o640); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSync(
		[]string{"--source", source, "--target", target, "--write"},
		&stdout,
		&stderr,
	)
	if err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	want := `model = "gpt-5.5"

[features]
apps = true

[projects."/repo"]
trust_level = "trusted"
`
	if string(got) != want {
		t.Fatalf("unexpected target contents\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestRunCodexConfigReadUsesCODEXHOMEByDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CODEX_HOME", dir)
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`[projects."/repo"]
trust_level = "trusted"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runCodexConfigRead(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "/repo\ttrusted\n" {
		t.Fatalf("unexpected output: %q", stdout.String())
	}
}

func TestRunSyncUsesCODEXHOMEInCodexMode(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CODEX_HOME", dir)
	source := filepath.Join(dir, "source.toml")
	target := filepath.Join(dir, "config.toml")
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

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSync([]string{"--source", source, "--codex"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	want := `model = "gpt-5.5"

[features]
apps = true

[projects."/repo"]
trust_level = "trusted"
`
	if stdout.String() != want {
		t.Fatalf("unexpected output\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestRunSyncCodexModeUsesProjectsPreserveByDefault(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.toml")
	target := filepath.Join(dir, "target.toml")
	if err := os.WriteFile(source, []byte(`model = "gpt-5.5"

[projects."/source-only"]
trust_level = "untrusted"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(`[projects."/repo"]
trust_level = "trusted"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSync([]string{"--source", source, "--target", target, "--codex"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}

	want := `model = "gpt-5.5"

[projects."/repo"]
trust_level = "trusted"
`
	if stdout.String() != want {
		t.Fatalf("unexpected output\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}
