package secret

import (
	"os"
	"path/filepath"
)

func createPrivateTempFile(pattern string) (*os.File, error) {
	dir, err := os.MkdirTemp("", "shelf-secret-*")
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, pattern)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}
	return file, nil
}

func removePrivateTempFile(path string) {
	_ = os.RemoveAll(filepath.Dir(path))
}
