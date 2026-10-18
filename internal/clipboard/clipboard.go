package clipboard

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"unicode/utf16"

	nativeclipboard "github.com/atotto/clipboard"
)

type Writer interface {
	CopyText(string) error
}

type Reader interface {
	ReadText() (string, error)
}

type systemWriter struct{}

func New() Writer {
	return systemWriter{}
}

func NewReader() Reader {
	return systemWriter{}
}

func (systemWriter) ReadText() (string, error) {
	return readText(isWSL(), readWithWindowsClipboard, readWithWaylandClipboard, readWithNativeClipboard)
}

func readText(wsl bool, windowsRead, waylandRead, nativeRead func() (string, error)) (string, error) {
	if !wsl {
		text, err := nativeRead()
		if err != nil {
			return "", err
		}
		return text, nil
	}
	text, err := windowsRead()
	if err == nil {
		return text, nil
	}
	text, fallbackErr := waylandRead()
	if fallbackErr == nil {
		return text, nil
	}
	return "", fmt.Errorf("read WSL clipboard: Windows bridge failed: %w; Wayland fallback failed: %v", err, fallbackErr)
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
	// clip.exe expects UTF-16LE on stdin; a BOM would become clipboard text.
	encoded := make([]byte, 0, len(text)*2)
	for _, r := range text {
		if r <= 0xffff {
			encoded = binary.LittleEndian.AppendUint16(encoded, uint16(r))
		} else {
			high, low := utf16.EncodeRune(r)
			encoded = binary.LittleEndian.AppendUint16(encoded, uint16(high))
			encoded = binary.LittleEndian.AppendUint16(encoded, uint16(low))
		}
	}
	cmd.Stdin = bytes.NewReader(encoded)
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

const wslClipboardReadPath = "/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe"

func readWithWindowsClipboard() (string, error) {
	command, err := exec.LookPath("powershell.exe")
	if err != nil {
		if _, err := os.Stat(wslClipboardReadPath); err != nil {
			return "", fmt.Errorf("locate Windows clipboard reader: %w", err)
		}
		command = wslClipboardReadPath
	}
	cmd := exec.Command(command, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command",
		"[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false); [Console]::Out.Write((Get-Clipboard -Raw))")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("read Windows clipboard: %w", err)
	}
	return string(output), nil
}

func readWithWaylandClipboard() (string, error) {
	output, err := exec.Command("wl-paste", "--no-newline").Output()
	if err != nil {
		return "", fmt.Errorf("read Wayland clipboard: %w", err)
	}
	return string(output), nil
}

func readWithNativeClipboard() (string, error) {
	if nativeclipboard.Unsupported {
		return "", fmt.Errorf("system clipboard is unsupported")
	}
	text, err := nativeclipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("read from system clipboard: %w", err)
	}
	return text, nil
}
