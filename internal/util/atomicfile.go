package util

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type AtomicWriteOptions struct {
	FileMode os.FileMode
	DirMode  os.FileMode
	Sync     bool
	Backup   bool
}

func AtomicWrite(path string, content []byte, opts AtomicWriteOptions) error {
	fileMode := opts.FileMode
	if fileMode == 0 {
		fileMode = 0o600
	}
	dirMode := opts.DirMode
	if dirMode == 0 {
		dirMode = 0o700
	}
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return err
	}
	if opts.Backup {
		if err := backup(path, fileMode); err != nil {
			return err
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(fileMode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if opts.Sync {
		if err := tmp.Sync(); err != nil {
			tmp.Close()
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func backup(path string, mode os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return copyFile(path, path+".bak", mode)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
