package codexconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectsReturnsCodexProjectEntries(t *testing.T) {
	cfg := Config{Document: parseDoc(t, `model = "gpt-5.5"

[projects."/Users/example/repo"]
tags = [
  ["tag"]
]
trust_level = "trusted"

[projects."/Users/example/other"]
trust_level = "untrusted" # comment
`)}

	projects := cfg.Projects()
	if len(projects) != 2 {
		t.Fatalf("expected two projects, got %d", len(projects))
	}
	if projects[0].Path != "/Users/example/repo" || projects[0].TrustLevel != "trusted" {
		t.Fatalf("unexpected first project: %#v", projects[0])
	}
	if projects[1].Path != "/Users/example/other" || projects[1].TrustLevel != "untrusted" {
		t.Fatalf("unexpected second project: %#v", projects[1])
	}
}

func TestLoadReadsConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`[projects."/repo"]
trust_level = "trusted"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	projects := cfg.Projects()
	if len(projects) != 1 || projects[0].Path != "/repo" {
		t.Fatalf("unexpected projects: %#v", projects)
	}
}

func TestLoadDocumentReadsConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`model = "gpt-5.5"

[projects."/repo"]
trust_level = "trusted"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Blocks) != 2 {
		t.Fatalf("unexpected blocks: %#v", doc.Blocks)
	}
}
