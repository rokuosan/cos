package tomlsync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDocumentCreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	doc := parseDoc(t, `model = "gpt-5.5"
`)

	if err := WriteDocument(path, doc, false); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != doc.String() {
		t.Fatalf("unexpected file contents\nwant:\n%s\ngot:\n%s", doc.String(), string(got))
	}
}

func TestWriteDocumentCreatesDirectoryWithPrivatePermissions(t *testing.T) {
	dir := t.TempDir()
	targetDir := filepath.Join(dir, "nested")
	path := filepath.Join(targetDir, "config.toml")
	doc := parseDoc(t, `model = "gpt-5.5"
`)

	if err := WriteDocument(path, doc, false); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(targetDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("unexpected directory mode: got %o", info.Mode().Perm())
	}
}

func TestWriteDocumentPreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("old\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	doc := parseDoc(t, `model = "gpt-5.5"
`)

	if err := WriteDocument(path, doc, false); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("unexpected mode: got %o", info.Mode().Perm())
	}
}

func TestWriteDocumentRejectsSymlinkByDefault(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.toml")
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(realPath, []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, path); err != nil {
		t.Fatal(err)
	}
	doc := parseDoc(t, `model = "gpt-5.5"
`)

	err := WriteDocument(path, doc, false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteDocumentReplacesSymlinkWhenAllowed(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.toml")
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(realPath, []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, path); err != nil {
		t.Fatal(err)
	}
	doc := parseDoc(t, `model = "gpt-5.5"
`)

	if err := WriteDocument(path, doc, true); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("target is still a symlink")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != doc.String() {
		t.Fatalf("unexpected file contents\nwant:\n%s\ngot:\n%s", doc.String(), string(got))
	}

	realGot, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(realGot) != "old\n" {
		t.Fatalf("unexpected real target contents: %q", string(realGot))
	}
}

func TestWriteDocumentKeepsSymlinkOnReplacementFailure(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "real.toml")
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(realPath, []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, path); err != nil {
		t.Fatal(err)
	}
	doc := parseDoc(t, `model = "gpt-5.5"
`)

	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(dir, 0o700)
	}()

	err := WriteDocument(path, doc, true)
	if err == nil {
		t.Fatal("expected error")
	}

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink was unexpectedly replaced on failure")
	}

	realGot, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(realGot) != "old\n" {
		t.Fatalf("unexpected real target contents: %q", string(realGot))
	}
}
