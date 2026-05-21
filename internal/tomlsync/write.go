package tomlsync

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteDocument writes doc to path using an atomic rename.
//
// Existing file permissions are preserved where possible. If path is a
// symlink, writing fails unless replaceSymlink is true, in which case the
// symlink itself is replaced by a regular file.
func WriteDocument(path string, doc Document, replaceSymlink bool) error {
	path = expandHome(path)

	info, err := os.Lstat(path)
	mode := os.FileMode(0o600)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			if !replaceSymlink {
				return fmt.Errorf("refusing to write through symlink target %q", path)
			}
		} else {
			mode = info.Mode().Perm()
		}
	case os.IsNotExist(err):
	default:
		return fmt.Errorf("stat target: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".cos-config-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if _, err := tmp.WriteString(doc.String()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace target: %w", err)
	}
	return nil
}
