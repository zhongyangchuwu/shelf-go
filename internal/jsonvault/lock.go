package jsonvault

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

type Lock struct {
	file *flock.Flock
}

func LockFile(dataPath string) (*Lock, error) {
	if err := os.MkdirAll(filepath.Dir(dataPath), 0o700); err != nil {
		return nil, err
	}
	file := flock.New(dataPath+".lock", flock.SetPermissions(0o600))
	if err := file.Lock(); err != nil {
		return nil, fmt.Errorf("lock %s: %w", dataPath, err)
	}
	return &Lock{file: file}, nil
}

func (l *Lock) Unlock() error {
	if l == nil || l.file == nil {
		return nil
	}
	if err := l.file.Unlock(); err != nil {
		return err
	}
	l.file = nil
	return nil
}
