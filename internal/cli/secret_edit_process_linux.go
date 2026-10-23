package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func secretEditorGroupRunning(group int) (bool, error) {
	if err := unix.Kill(-group, 0); errors.Is(err, unix.ESRCH) {
		return false, nil
	} else if err != nil {
		return true, err
	}
	// A terminated wrapper's children may briefly remain as orphan zombies.
	// kill(-pgid, 0) includes those, although they can no longer use a draft.
	// Inspect only process metadata, never editor contents or command lines.
	proc, err := os.Open("/proc")
	if err != nil {
		return true, err
	}
	defer proc.Close()
	for {
		names, err := proc.Readdirnames(128)
		if err != nil && !errors.Is(err, io.EOF) {
			return true, err
		}
		for _, name := range names {
			pid, parseErr := strconv.Atoi(name)
			if parseErr != nil {
				continue
			}
			pgid, groupErr := unix.Getpgid(pid)
			if errors.Is(groupErr, unix.ESRCH) {
				continue
			}
			if groupErr != nil {
				return true, groupErr
			}
			if pgid != group {
				continue
			}
			stat, readErr := os.ReadFile(filepath.Join("/proc", name, "stat"))
			if errors.Is(readErr, os.ErrNotExist) {
				continue
			}
			if readErr != nil {
				return true, readErr
			}
			end := bytes.LastIndexByte(stat, ')')
			if end < 0 || end+2 >= len(stat) {
				return true, errors.New("invalid process status")
			}
			status := strings.Fields(string(stat[end+2:]))
			if len(status) < 3 {
				return true, errors.New("invalid process status")
			}
			// Recheck the group in the same snapshot as the process state.
			actualGroup, parseErr := strconv.Atoi(status[2])
			if parseErr != nil {
				return true, parseErr
			}
			if actualGroup == group && status[0] != "Z" && status[0] != "X" {
				return true, nil
			}
		}
		if errors.Is(err, io.EOF) {
			return false, nil
		}
	}
}
