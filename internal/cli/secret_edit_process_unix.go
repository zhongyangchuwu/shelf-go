//go:build unix

package cli

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type secretEditorProcess struct {
	child      *exec.Cmd
	terminal   *os.File
	state      *term.State
	foreground int
}

func prepareSecretEditor(child *exec.Cmd) (*secretEditorProcess, error) {
	process := &secretEditorProcess{child: child}
	child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	input, ok := child.Stdin.(*os.File)
	if !ok || !term.IsTerminal(int(input.Fd())) {
		return process, nil
	}
	state, err := term.GetState(int(input.Fd()))
	if err != nil {
		return nil, errors.New("capture editor terminal state failed")
	}
	process.terminal, process.state = input, state
	foreground, err := unix.IoctlGetInt(int(input.Fd()), unix.TIOCGPGRP)
	if errors.Is(err, unix.ENOTTY) {
		// An explicitly supplied PTY need not be our controlling terminal.
		return process, nil
	}
	if err != nil {
		return nil, errors.New("read editor terminal foreground failed")
	}
	if foreground != unix.Getpgrp() {
		return nil, errors.New("secret edit requires the terminal foreground")
	}
	process.foreground = foreground
	// Go performs the group creation and foreground transfer in the child,
	// before exec and before restoring its signal mask. This avoids both a
	// terminal-read/SIGTTIN race and moving Shelf's own shell into our group.
	child.SysProcAttr.Foreground = true
	child.SysProcAttr.Ctty = int(input.Fd())
	return process, nil
}

func (p *secretEditorProcess) start() error { return p.child.Start() }

func (p *secretEditorProcess) signal(sig os.Signal) error {
	value, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("unsupported editor signal")
	}
	return unix.Kill(-p.child.Process.Pid, value)
}

func (p *secretEditorProcess) running() (bool, error) {
	return secretEditorGroupRunning(p.child.Process.Pid)
}

func (p *secretEditorProcess) kill() error {
	return unix.Kill(-p.child.Process.Pid, unix.SIGKILL)
}

func (p *secretEditorProcess) restore() error {
	if p.terminal == nil {
		return nil
	}
	// Reclaiming the terminal from a background process group requires
	// SIGTTOU to be ignored. Preserve an inherited ignored disposition.
	ignored := signal.Ignored(syscall.SIGTTOU)
	signal.Ignore(syscall.SIGTTOU)
	if !ignored {
		defer signal.Reset(syscall.SIGTTOU)
	}
	var restoreErr error
	if p.foreground != 0 {
		if err := unix.IoctlSetPointerInt(int(p.terminal.Fd()), unix.TIOCSPGRP, p.foreground); err != nil {
			restoreErr = errors.New("restore editor terminal foreground failed")
		}
	}
	if err := term.Restore(int(p.terminal.Fd()), p.state); err != nil {
		restoreErr = errors.New("restore editor terminal state failed")
	}
	return restoreErr
}
