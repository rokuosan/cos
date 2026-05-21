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

func TestRunCodexConfigSyncWriteUpdatesTarget(t *testing.T) {
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

	var out bytes.Buffer
	err := runCodexConfigSync([]string{"--source", source, "--target", target, "--write"}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("unexpected stdout: %q", out.String())
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

	var out bytes.Buffer
	if err := runCodexConfigRead(nil, &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "/repo\ttrusted\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunCodexConfigSyncUsesCODEXHOMEByDefault(t *testing.T) {
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

	var out bytes.Buffer
	err := runCodexConfigSync([]string{"--source", source}, &out)
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
