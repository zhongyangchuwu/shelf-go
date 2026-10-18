package clipboard

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyTextUsesWindowsBridgeOnWSL(t *testing.T) {
	var windowsCalls, waylandCalls, nativeCalls int
	err := copyText("secret", true,
		func(text string) error {
			windowsCalls++
			if text != "secret" {
				t.Fatalf("windows text = %q", text)
			}
			return nil
		},
		func(string) error {
			waylandCalls++
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
	if windowsCalls != 1 || waylandCalls != 0 || nativeCalls != 0 {
		t.Fatalf("calls: windows=%d wayland=%d native=%d", windowsCalls, waylandCalls, nativeCalls)
	}
}

func TestCopyTextFallsBackToWaylandOnWSL(t *testing.T) {
	windowsErr := errors.New("Windows clipboard unavailable")
	var waylandCalls int
	err := copyText("secret", true,
		func(string) error { return windowsErr },
		func(text string) error {
			waylandCalls++
			if text != "secret" {
				t.Fatalf("wayland text = %q", text)
			}
			return nil
		},
		func(string) error {
			t.Fatal("native backend should not run on WSL")
			return nil
		},
	)
	if err != nil {
		t.Fatalf("copy text: %v", err)
	}
	if waylandCalls != 1 {
		t.Fatalf("wayland calls = %d", waylandCalls)
	}
}

func TestCopyTextUsesNativeClipboardOffWSL(t *testing.T) {
	var windowsCalls, waylandCalls, nativeCalls int
	err := copyText("secret", false,
		func(string) error {
			windowsCalls++
			return nil
		},
		func(string) error {
			waylandCalls++
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
	if windowsCalls != 0 || waylandCalls != 0 || nativeCalls != 1 {
		t.Fatalf("calls: windows=%d wayland=%d native=%d", windowsCalls, waylandCalls, nativeCalls)
	}
}

func TestCopyTextPropagatesBackendError(t *testing.T) {
	want := errors.New("clipboard unavailable")
	err := copyText("secret", false,
		func(string) error { return nil },
		func(string) error { return nil },
		func(string) error { return want },
	)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestCopyTextReportsBothWSLBackendsFailing(t *testing.T) {
	windowsErr := errors.New("Windows clipboard unavailable")
	waylandErr := errors.New("Wayland clipboard unavailable")
	err := copyText("secret", true,
		func(string) error { return windowsErr },
		func(string) error { return waylandErr },
		func(string) error { return nil },
	)
	if !errors.Is(err, windowsErr) || !strings.Contains(err.Error(), "Wayland fallback failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error leaked copied text: %v", err)
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

func TestWindowsClipboardCopyWritesUTF16LEWithoutBOM(t *testing.T) {
	for _, test := range []struct {
		name string
		text string
		want []byte
	}{
		{name: "ASCII", text: "Shelf", want: []byte{'S', 0, 'h', 0, 'e', 0, 'l', 0, 'f', 0}},
		{name: "Chinese and supplementary rune", text: "用户 🔑", want: []byte{0x28, 0x75, 0x37, 0x62, 0x20, 0x00, 0x3d, 0xd8, 0x11, 0xdd}},
		{name: "combining mark and trailing newline", text: "e\u0301\r\n", want: []byte{0x65, 0x00, 0x01, 0x03, 0x0d, 0x00, 0x0a, 0x00}},
		{name: "empty"},
	} {
		t.Run(test.name, func(t *testing.T) {
			capture := filepath.Join(t.TempDir(), "clipboard.bin")
			t.Setenv("SHELF_TEST_CLIPBOARD_CAPTURE", capture)
			installClipboardReader(t, "clip.exe", "/bin/cat > \"$SHELF_TEST_CLIPBOARD_CAPTURE\"\n")
			if err := copyWithWindowsClipboard(test.text); err != nil {
				t.Fatalf("copy to Windows clipboard: %v", err)
			}
			got, err := os.ReadFile(capture)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, test.want) {
				t.Fatalf("Windows clipboard bytes = %x, want %x", got, test.want)
			}
		})
	}
}
