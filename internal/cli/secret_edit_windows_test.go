package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestSecretEditWindowsJobStopsAttachedWrapperChild(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
	recordPath := configureSecretEditor(t, "stubborn")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("VISUAL", fmt.Sprintf("%q -test.run=^TestSecretEditWrapperProcess$ --", executable))
	earlyCleanup := filepath.Join(t.TempDir(), "early-cleanup")
	t.Setenv("SHELF_TEST_EDITOR_EARLY_CLEANUP", earlyCleanup)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := NewRootCmd(testApp())
	cmd.SetContext(ctx)
	cmd.SetIn(bytes.NewReader(nil))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--vault", path, "secret", "edit", "alpha"})
	done := make(chan error, 1)
	go func() { done <- cmd.Execute() }()
	for {
		data, err := os.ReadFile(recordPath)
		if err == nil && bytes.HasSuffix(data, []byte{'\n'}) {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("attached editor did not become ready")
		case <-time.After(10 * time.Millisecond):
		}
	}
	record := secretEditorRecords(t, recordPath)[0]
	process, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_TERMINATE, false, uint32(record.PID))
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(process)
	defer windows.TerminateProcess(process, 1)
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled editor invocation reported success")
		}
		assertSecretEditValueFree(t, out.String(), err)
	case <-time.After(5 * time.Second):
		t.Fatal("attached editor job did not terminate before deadline")
	}
	state, err := windows.WaitForSingleObject(process, 0)
	if err != nil || state != windows.WAIT_OBJECT_0 {
		t.Fatal("attached editor remains active after cancellation")
	}
	if _, err := os.Stat(filepath.Dir(record.Path)); !os.IsNotExist(err) {
		t.Fatal("terminated editor job left its private draft directory")
	}
	if _, err := os.Stat(earlyCleanup); !os.IsNotExist(err) {
		t.Fatal("draft was removed before its attached editor stopped")
	}
	got, err := runShelf(t, "--vault", path, "secret", "get", "item_alpha")
	if err != nil || got != "alpha-private-value\n" {
		t.Fatal("cancelled editor job changed the vault")
	}
}
