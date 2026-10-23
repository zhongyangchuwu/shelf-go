//go:build unix && !linux

package cli

import (
	"errors"

	"golang.org/x/sys/unix"
)

func secretEditorGroupRunning(group int) (bool, error) {
	err := unix.Kill(-group, 0)
	if errors.Is(err, unix.ESRCH) {
		return false, nil
	}
	return true, err
}
