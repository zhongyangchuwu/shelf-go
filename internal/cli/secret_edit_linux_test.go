package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func secretEditTerminal(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("PTY unavailable: %v", err)
	}
	t.Cleanup(func() { _ = master.Close() })
	var unlock int32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&unlock))); errno != 0 {
		t.Fatalf("unlock PTY: %v", errno)
	}
	var number uint32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&number))); errno != 0 {
		t.Fatalf("get PTY: %v", errno)
	}
	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", number), os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = slave.Close() })
	return master, slave
}

type secretEditTerminalOutput struct {
	mu      sync.Mutex
	text    bytes.Buffer
	prompts chan struct{}
}

func (w *secretEditTerminalOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.text.Write(p)
	if bytes.Contains(p, []byte("[d]iscard:")) || bytes.Contains(p, []byte("Target number")) {
		select {
		case w.prompts <- struct{}{}:
		default:
		}
	}
	return n, err
}

func runSecretEditTerminal(t *testing.T, path string, args []string, answers []string, beforeAnswer func(int)) (string, error) {
	t.Helper()
	master, slave := secretEditTerminal(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := NewRootCmd(testApp())
	cmd.SetContext(ctx)
	cmd.SetIn(slave)
	out := &secretEditTerminalOutput{prompts: make(chan struct{}, 4)}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(append([]string{"--vault", path, "secret", "edit"}, args...))
	done := make(chan error, 1)
	go func() { done <- cmd.Execute() }()
	index := 0
	for {
		select {
		case <-ctx.Done():
			_ = slave.Close()
			_ = master.Close()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
			}
			t.Fatal("interactive edit did not finish before deadline")
		case <-out.prompts:
			if index >= len(answers) {
				cancel()
				t.Fatal("unexpected retry prompt")
			}
			if beforeAnswer != nil {
				beforeAnswer(index)
			}
			if _, err := master.Write([]byte(answers[index])); err != nil {
				cancel()
				t.Fatal(err)
			}
			index++
		case err := <-done:
			if ctx.Err() != nil {
				t.Fatal("interactive edit did not finish before deadline")
			}
			if index != len(answers) {
				t.Fatalf("answered %d prompts, expected %d", index, len(answers))
			}
			out.mu.Lock()
			text := out.text.String()
			out.mu.Unlock()
			return text, err
		}
	}
}

func TestSecretEditValidationReeditDiscardAndEOF(t *testing.T) {
	for _, action := range []string{"reedit", "discard", "eof"} {
		t.Run(action, func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
			mode, answer, invocations := "invalid", "d\n", 1
			if action == "reedit" {
				mode, answer, invocations = "reedit", "e\n", 2
			} else if action == "eof" {
				answer = "\x04"
			}
			record := configureSecretEditor(t, mode)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			out, err := runSecretEditTerminal(t, path, []string{"alpha"}, []string{answer}, nil)
			assertSecretEditValueFree(t, out, err)
			if action == "eof" {
				if err == nil {
					t.Fatal("terminal EOF accepted an invalid edit")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			records := assertSecretEditorCleaned(t, record, invocations)
			if action == "reedit" {
				if records[0].Path != records[1].Path {
					t.Fatal("re-edit did not retain the edited draft")
				}
				got, err := runShelf(t, "--vault", path, "secret", "get", "item_alpha")
				if err != nil || got != "edited-private-value\n" {
					t.Fatal("corrected document was not saved")
				}
			} else {
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("discard or EOF changed storage")
				}
			}
		})
	}
}

func TestSecretEditSaveFailureChoices(t *testing.T) {
	for _, action := range []string{"retry", "edit", "discard"} {
		t.Run(action, func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
			record := configureSecretEditor(t, "savefail")
			t.Setenv("SHELF_TEST_EDITOR_VAULT", path)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			answer := map[string]string{"retry": "r\n", "edit": "e\n", "discard": "d\n"}[action]
			out, err := runSecretEditTerminal(t, path, []string{"alpha"}, []string{answer}, func(int) {
				if err := os.WriteFile(path, before, 0o600); err != nil {
					t.Fatal(err)
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			assertSecretEditValueFree(t, out, err)
			invocations := 1
			if action == "edit" {
				invocations = 2
			}
			assertSecretEditorCleaned(t, record, invocations)
			got, err := runShelf(t, "--vault", path, "secret", "get", "item_alpha")
			want := "edited-private-value\n"
			if action == "discard" {
				want = "alpha-private-value\n"
			}
			if err != nil || got != want {
				t.Fatal("save failure choice produced the wrong persisted value")
			}
		})
	}
}

func TestSecretEditInteractiveAmbiguityUsesExistingPicker(t *testing.T) {
	path := writeLabeledVault(t, map[string]labeledItem{"alpha": {env: "SHARED_TOKEN", value: "alpha-private-value"}, "beta": {env: "SHARED_TOKEN", value: "beta-private-value"}})
	record := configureSecretEditor(t, "edit")
	out, err := runSecretEditTerminal(t, path, []string{"--env", "SHARED_TOKEN"}, []string{"2\n"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertSecretEditValueFree(t, out, err)
	if !strings.Contains(out, "item_alpha.value") || !strings.Contains(out, "item_beta.value") {
		t.Fatal("ambiguity picker omitted concrete candidates")
	}
	assertSecretEditorCleaned(t, record, 1)
	got, err := runShelf(t, "--vault", path, "secret", "get", "item_beta")
	if err != nil || got != "edited-private-value\n" {
		t.Fatal("ambiguity picker selected the wrong item")
	}
}

func TestSecretEditCLIProcess(t *testing.T) {
	if os.Getenv("SHELF_TEST_CLI_PROCESS") == "" {
		return
	}
	cmd := NewRootCmd(testApp())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd.SetContext(ctx)
	if cancelPath := os.Getenv("SHELF_TEST_CLI_CANCEL"); cancelPath != "" {
		go func() {
			for {
				if _, err := os.Stat(cancelPath); err == nil {
					cancel()
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Millisecond):
				}
			}
		}()
	}
	var terminalState *term.State
	if os.Getenv("SHELF_TEST_EDITOR_TERMINAL_RESTORED") != "" {
		var err error
		terminalState, err = term.GetState(int(os.Stdin.Fd()))
		if err != nil {
			os.Exit(86)
		}
	}
	cmd.SetArgs([]string{"--vault", os.Getenv("SHELF_TEST_EDITOR_VAULT"), "secret", "edit", "alpha"})
	runErr := cmd.Execute()
	if terminalState != nil {
		after, err := term.GetState(int(os.Stdin.Fd()))
		foreground, groupErr := unix.IoctlGetInt(int(os.Stdin.Fd()), unix.TIOCGPGRP)
		if err != nil || groupErr != nil || !reflect.DeepEqual(terminalState, after) || foreground != unix.Getpgrp() {
			os.Exit(87)
		}
		if os.WriteFile(os.Getenv("SHELF_TEST_EDITOR_TERMINAL_RESTORED"), []byte("restored"), 0o600) != nil {
			os.Exit(88)
		}
	}
	if runErr != nil {
		fmt.Fprintln(os.Stderr, runErr)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestSecretEditSignalsKeepActiveDraftAndCleanAfterExit(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP} {
		t.Run(sig.String(), func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
			record := configureSecretEditor(t, "interrupt")
			t.Setenv("SHELF_TEST_CLI_PROCESS", "1")
			t.Setenv("SHELF_TEST_EDITOR_VAULT", path)
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			child := exec.CommandContext(ctx, executable, "-test.run=^TestSecretEditCLIProcess$")
			child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			child.WaitDelay = time.Second
			var out bytes.Buffer
			child.Stdout, child.Stderr = &out, &out
			if err := child.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = syscall.Kill(-child.Process.Pid, syscall.SIGKILL) }()
			for {
				data, err := os.ReadFile(record)
				if err == nil && bytes.Contains(data, []byte{'\n'}) {
					break
				}
				select {
				case <-ctx.Done():
					_ = syscall.Kill(-child.Process.Pid, syscall.SIGKILL)
					_ = child.Wait()
					t.Fatal("external editor did not become ready")
				case <-time.After(10 * time.Millisecond):
				}
			}
			if err := syscall.Kill(-child.Process.Pid, sig); err != nil {
				t.Fatal(err)
			}
			err = child.Wait()
			if ctx.Err() != nil {
				t.Fatal("interrupted editor did not exit before deadline")
			}
			assertSecretEditValueFree(t, out.String(), err)
			assertSecretEditorCleaned(t, record, 1)
			got, getErr := runShelf(t, "--vault", path, "secret", "get", "item_alpha")
			if sig == syscall.SIGINT {
				if err != nil || getErr != nil || got != "edited-private-value\n" {
					t.Fatal("editor could not handle Ctrl+C and finish using its live draft")
				}
			} else if err == nil || getErr != nil || got != "alpha-private-value\n" {
				t.Fatal("termination saved a draft or failed to report interruption")
			}
		})
	}
}

func waitSecretEditFile(t *testing.T, ctx context.Context, path string) {
	t.Helper()
	for {
		if data, err := os.ReadFile(path); err == nil && len(data) != 0 {
			if !strings.HasSuffix(path, ".jsonl") || bytes.HasSuffix(data, []byte{'\n'}) {
				return
			}
		}
		select {
		case <-ctx.Done():
			t.Fatal("editor lifecycle readiness deadline exceeded")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestSecretEditAttachedWrapperLifecycleAndTerminal(t *testing.T) {
	for _, action := range []string{"term", "hup", "cancel", "tty-exit", "tty-control-c", "tty-term", "tty-cancel"} {
		t.Run(action, func(t *testing.T) {
			path := writeLabeledVault(t, map[string]labeledItem{"alpha": {value: "alpha-private-value"}})
			mode := "stubborn"
			if action == "tty-exit" {
				mode = "tty"
			} else if action == "tty-control-c" {
				mode = "tty-control-c"
			} else if strings.HasPrefix(action, "tty") {
				mode = "tty-stubborn"
			}
			record := configureSecretEditor(t, mode)
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			wrapper := filepath.Join(dir, "attached wrapper.sh")
			script := "#!/bin/sh\n'" + strings.ReplaceAll(executable, "'", "'\\''") + "' -test.run=^TestSecretEditEditorProcess$ -- \"$@\"\nstatus=$?\nexit \"$status\"\n"
			if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("VISUAL", "/bin/sh '"+wrapper+"'")
			t.Setenv("SHELF_TEST_CLI_PROCESS", "1")
			t.Setenv("SHELF_TEST_EDITOR_VAULT", path)
			cancelPath := filepath.Join(dir, "cancel")
			t.Setenv("SHELF_TEST_CLI_CANCEL", cancelPath)
			readPath := filepath.Join(dir, "read")
			t.Setenv("SHELF_TEST_EDITOR_READ", readPath)
			earlyCleanup := filepath.Join(dir, "early-cleanup")
			t.Setenv("SHELF_TEST_EDITOR_EARLY_CLEANUP", earlyCleanup)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			child := exec.CommandContext(ctx, executable, "-test.run=^TestSecretEditCLIProcess$")
			child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			var out bytes.Buffer
			child.Stdout, child.Stderr = &out, &out
			var master *os.File
			restored := filepath.Join(dir, "restored")
			if strings.HasPrefix(action, "tty") {
				var slave *os.File
				master, slave = secretEditTerminal(t)
				child.Stdin = slave
				child.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
				t.Setenv("SHELF_TEST_EDITOR_TERMINAL_RESTORED", restored)
				// Command.Env is assigned after the per-case terminal marker.
				child.Env = os.Environ()
			}
			child.WaitDelay = time.Second
			if err := child.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = child.Process.Kill() }()
			waitSecretEditFile(t, ctx, record)
			editor := secretEditorRecords(t, record)[0]
			group, err := unix.Getpgid(editor.PID)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = unix.Kill(-group, unix.SIGKILL) }()
			if master != nil {
				if _, err := master.Write([]byte{'x'}); err != nil {
					t.Fatal(err)
				}
				waitSecretEditFile(t, ctx, readPath)
			}
			switch action {
			case "tty-exit":
			case "tty-control-c":
				if _, err := master.Write([]byte{3}); err != nil {
					t.Fatal(err)
				}
			case "cancel", "tty-cancel":
				if err := os.WriteFile(cancelPath, []byte("cancel"), 0o600); err != nil {
					t.Fatal(err)
				}
			default:
				sig := syscall.SIGTERM
				if action == "hup" {
					sig = syscall.SIGHUP
				}
				if err := child.Process.Signal(sig); err != nil {
					t.Fatal(err)
				}
			}
			err = child.Wait()
			if ctx.Err() != nil {
				t.Fatal("attached invocation did not stop before deadline")
			}
			if action == "tty-exit" || action == "tty-control-c" {
				if err != nil {
					t.Fatal("foreground editor could not finish normally")
				}
			} else if err == nil {
				t.Fatal("terminated attached invocation reported success")
			}
			assertSecretEditValueFree(t, out.String(), err)
			assertSecretEditorCleaned(t, record, 1)
			if running, probeErr := secretEditorGroupRunning(group); probeErr != nil || running {
				t.Fatal("attached editor invocation remains active after cleanup")
			}
			if _, err := os.Stat(earlyCleanup); !os.IsNotExist(err) {
				t.Fatal("draft was removed before its attached editor stopped")
			}
			if master != nil {
				if data, err := os.ReadFile(restored); err != nil || string(data) != "restored" {
					t.Fatal("terminal foreground or attributes were not restored")
				}
			}
		})
	}
}
