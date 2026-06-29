package vault

import (
	"os"

	"github.com/zhongyangchuwu/shelf-go/internal/atomicfile"
)

type Options struct {
	FileMode os.FileMode
	DirMode  os.FileMode
	Sync     bool
	Backup   bool
}

func Write(path string, content []byte, opts Options) error {
	return atomicfile.Write(path, content, atomicfile.Options{FileMode: opts.FileMode, DirMode: opts.DirMode, Sync: opts.Sync, Backup: opts.Backup})
}
