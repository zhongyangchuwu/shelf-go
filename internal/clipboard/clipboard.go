package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	nativeclipboard "github.com/atotto/clipboard"
)

type Writer interface {
	CopyText(string) error
}

type systemWriter struct{}

func New() Writer {
	return systemWriter{}
}

func (systemWriter) CopyText(text string) error {
	return copyText(text, isWSL(), copyWithWindowsClipboard, copyWithWaylandClipboard, copyWithNativeClipboard)
}

func copyText(text string, wsl bool, windowsCopy, waylandCopy, nativeCopy func(string) error) error {
	if !wsl {
		return nativeCopy(text)
	}
	if err := windowsCopy(text); err == nil {
		return nil
	} else if fallbackErr := waylandCopy(text); fallbackErr == nil {
		return nil
	} else {
		return fmt.Errorf("copy to WSL clipboard: Windows bridge failed: %w; Wayland fallback failed: %v", err, fallbackErr)
	}
}

func isWSL() bool {
	return os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != ""
}

const wslClipboardPath = "/mnt/c/Windows/System32/clip.exe"

func copyWithWindowsClipboard(text string) error {
	command, err := windowsClipboardCommand(exec.LookPath, os.Stat)
	if err != nil {
		return fmt.Errorf("locate Windows clipboard bridge: %w", err)
	}
	cmd := exec.Command(command)
	cmd.Stdin = bytes.NewBufferString(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("copy to Windows clipboard: %w", err)
	}
	return nil
}

func windowsClipboardCommand(lookPath func(string) (string, error), stat func(string) (os.FileInfo, error)) (string, error) {
	if command, err := lookPath("clip.exe"); err == nil {
		return command, nil
	}
	if _, err := stat(wslClipboardPath); err != nil {
		return "", err
	}
	return wslClipboardPath, nil
}

func copyWithWaylandClipboard(text string) error {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = bytes.NewBufferString(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("copy to Wayland clipboard: %w", err)
	}
	return nil
}

func copyWithNativeClipboard(text string) error {
	if nativeclipboard.Unsupported {
		return fmt.Errorf("system clipboard is unsupported")
	}
	if err := nativeclipboard.WriteAll(text); err != nil {
		return fmt.Errorf("write to system clipboard: %w", err)
	}
	return nil
}
