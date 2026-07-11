package clipboard

import (
	"errors"
	"os"
	"testing"
)

func TestCopyTextUsesWindowsBridgeOnWSL(t *testing.T) {
	var windowsCalls, nativeCalls int
	err := copyText("secret", true,
		func(text string) error {
			windowsCalls++
			if text != "secret" {
				t.Fatalf("windows text = %q", text)
			}
			return nil
		},
		func(string) error {
			nativeCalls++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("copy text: %v", err)
	}
	if windowsCalls != 1 || nativeCalls != 0 {
		t.Fatalf("calls: windows=%d native=%d", windowsCalls, nativeCalls)
	}
}

func TestCopyTextUsesNativeClipboardOffWSL(t *testing.T) {
	var windowsCalls, nativeCalls int
	err := copyText("secret", false,
		func(string) error {
			windowsCalls++
			return nil
		},
		func(text string) error {
			nativeCalls++
			if text != "secret" {
				t.Fatalf("native text = %q", text)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("copy text: %v", err)
	}
	if windowsCalls != 0 || nativeCalls != 1 {
		t.Fatalf("calls: windows=%d native=%d", windowsCalls, nativeCalls)
	}
}

func TestCopyTextPropagatesBackendError(t *testing.T) {
	want := errors.New("clipboard unavailable")
	err := copyText("secret", false, func(string) error { return nil }, func(string) error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestWindowsClipboardCommandUsesPATH(t *testing.T) {
	got, err := windowsClipboardCommand(
		func(name string) (string, error) {
			if name != "clip.exe" {
				t.Fatalf("look path name = %q", name)
			}
			return "/custom/clip.exe", nil
		},
		func(string) (os.FileInfo, error) {
			t.Fatal("stat should not run when clip.exe is on PATH")
			return nil, nil
		},
	)
	if err != nil || got != "/custom/clip.exe" {
		t.Fatalf("command = %q, err = %v", got, err)
	}
}

func TestWindowsClipboardCommandFallsBackToSystemPath(t *testing.T) {
	got, err := windowsClipboardCommand(
		func(string) (string, error) { return "", errors.New("not on PATH") },
		func(path string) (os.FileInfo, error) {
			if path != wslClipboardPath {
				t.Fatalf("stat path = %q", path)
			}
			return nil, nil
		},
	)
	if err != nil || got != wslClipboardPath {
		t.Fatalf("command = %q, err = %v", got, err)
	}
}

func TestWindowsClipboardCommandReturnsFallbackError(t *testing.T) {
	want := errors.New("not found")
	_, err := windowsClipboardCommand(
		func(string) (string, error) { return "", errors.New("not on PATH") },
		func(string) (os.FileInfo, error) { return nil, want },
	)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
