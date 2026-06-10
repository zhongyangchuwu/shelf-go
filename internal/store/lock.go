package store

import (
	"os"
	"path/filepath"
	"syscall"
)

type Lock struct {
	file *os.File
}

func LockFile(dataPath string) (*Lock, error) {
	if err := os.MkdirAll(filepath.Dir(dataPath), 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(dataPath+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return nil, err
	}
	return &Lock{file: file}, nil
}

func (l *Lock) Unlock() error {
	if l == nil || l.file == nil {
		return nil
	}
	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	if err != nil {
		return err
	}
	return closeErr
}
