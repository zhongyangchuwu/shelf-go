package clipboard

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReadTextSelectsClipboardBackend(t *testing.T) {
	unavailable := errors.New("clipboard unavailable")
	for _, test := range []struct {
		name       string
		wsl        bool
		windowsErr error
		waylandErr error
		nativeErr  error
		wantCalls  string
		wantErr    error
	}{
		{name: "Windows on WSL", wsl: true, wantCalls: "windows"},
		{name: "Wayland fallback", wsl: true, windowsErr: unavailable, wantCalls: "windows,wayland"},
		{name: "both WSL backends fail", wsl: true, windowsErr: unavailable, waylandErr: unavailable, wantCalls: "windows,wayland", wantErr: unavailable},
		{name: "native elsewhere", wantCalls: "native"},
		{name: "native failure", nativeErr: unavailable, wantCalls: "native", wantErr: unavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls []string
			backend := func(name string, err error) func() (string, error) {
				return func() (string, error) {
					calls = append(calls, name)
					return "  密钥 🔑\n\n", err
				}
			}
			text, err := readText(test.wsl, backend("windows", test.windowsErr), backend("wayland", test.waylandErr), backend("native", test.nativeErr))
			if !errors.Is(err, test.wantErr) || strings.Join(calls, ",") != test.wantCalls {
				t.Fatalf("read error = %v, calls = %v", err, calls)
			}
			if test.wantErr != nil {
				if text != "" || strings.Contains(err.Error(), "密钥") {
					t.Fatal("failed read exposed clipboard output")
				}
			} else if text != "  密钥 🔑\n\n" {
				t.Fatalf("clipboard whitespace or Unicode changed: %q", text)
			}
		})
	}
}

func TestReadTextDoesNotFallBackForEmptyWindowsClipboard(t *testing.T) {
	text, err := readText(true,
		func() (string, error) { return "", nil },
		func() (string, error) {
			t.Fatal("empty Windows clipboard must not read another clipboard")
			return "", nil
		},
		func() (string, error) { t.Fatal("native clipboard must not run on WSL"); return "", nil },
	)
	if text != "" || err != nil {
		t.Fatalf("empty clipboard = %q, %v", text, err)
	}
}

func TestWindowsClipboardReaderPreservesOutput(t *testing.T) {
	installClipboardReader(t, "powershell.exe", `
[ "$1" = "-NoLogo" ] && [ "$2" = "-NoProfile" ] && [ "$3" = "-NonInteractive" ] && [ "$4" = "-Command" ] || exit 2
printf '  密钥 🔑\n\n'
`)
	text, err := readWithWindowsClipboard()
	if err != nil || text != "  密钥 🔑\n\n" {
		t.Fatalf("Windows clipboard read = %q, %v", text, err)
	}
}

func TestWaylandClipboardReaderPreservesOutput(t *testing.T) {
	installClipboardReader(t, "wl-paste", `
[ "$#" -eq 1 ] && [ "$1" = "--no-newline" ] || exit 2
printf '  密钥 🔑\n\n'
`)
	text, err := readWithWaylandClipboard()
	if err != nil || text != "  密钥 🔑\n\n" {
		t.Fatalf("Wayland clipboard read = %q, %v", text, err)
	}
}

func TestClipboardReadersDiscardFailedCommandOutput(t *testing.T) {
	for _, test := range []struct {
		name string
		read func() (string, error)
	}{
		{name: "powershell.exe", read: readWithWindowsClipboard},
		{name: "wl-paste", read: readWithWaylandClipboard},
	} {
		t.Run(test.name, func(t *testing.T) {
			installClipboardReader(t, test.name, "printf 'sensitive-output'; printf 'sensitive-error' >&2; exit 1\n")
			text, err := test.read()
			if err == nil || text != "" || strings.Contains(err.Error(), "sensitive") {
				t.Fatal("failed clipboard command must discard stdout and stderr")
			}
		})
	}
}

func installClipboardReader(t *testing.T, name, script string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("command adapter fixture requires a POSIX shell")
	}
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, name), []byte("#!/bin/sh\n"+script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
}
