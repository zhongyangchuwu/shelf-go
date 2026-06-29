package shelfvault

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLockFileBlocksConcurrentLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.age")
	lock, err := LockFile(path)
	if err != nil {
		t.Fatalf("lock: %v", err)
	}
	defer lock.Unlock()

	locked := make(chan struct{})
	errc := make(chan error, 1)
	go func() {
		second, err := LockFile(path)
		if err != nil {
			errc <- err
			return
		}
		defer second.Unlock()
		close(locked)
	}()

	select {
	case <-locked:
		t.Fatalf("second lock acquired before first unlock")
	case err := <-errc:
		t.Fatalf("second lock failed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	if err := lock.Unlock(); err != nil {
		t.Fatalf("unlock: %v", err)
	}

	select {
	case <-locked:
	case err := <-errc:
		t.Fatalf("second lock failed after unlock: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatalf("second lock did not acquire after unlock")
	}
}
