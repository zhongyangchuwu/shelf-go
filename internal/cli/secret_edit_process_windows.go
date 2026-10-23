package cli

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

type secretEditorProcess struct {
	child    *exec.Cmd
	job      windows.Handle
	terminal *os.File
	state    *term.State
}

func prepareSecretEditor(child *exec.Cmd) (*secretEditorProcess, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, errors.New("create external editor job failed")
	}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		_ = windows.CloseHandle(job)
		return nil, errors.New("protect external editor job failed")
	}
	process := &secretEditorProcess{child: child, job: job}
	if input, ok := child.Stdin.(*os.File); ok && term.IsTerminal(int(input.Fd())) {
		state, err := term.GetState(int(input.Fd()))
		if err != nil {
			_ = windows.CloseHandle(job)
			return nil, errors.New("capture editor terminal state failed")
		}
		process.terminal, process.state = input, state
	}
	// Do not let a launcher execute even one instruction before assignment:
	// otherwise it could spawn an untracked editor before joining the job.
	child.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED}
	return process, nil
}

func (p *secretEditorProcess) start() (result error) {
	if err := p.child.Start(); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			_ = windows.TerminateJobObject(p.job, 1)
			_ = p.child.Process.Kill()
			done := make(chan struct{})
			go func() {
				_ = p.child.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				result = errSecretEditorActive
			}
		}
	}()
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(p.child.Process.Pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(process)
	if err := windows.AssignProcessToJobObject(p.job, process); err != nil {
		return err
	}
	// os/exec closes CreateProcess's primary-thread handle. A suspended new
	// process has exactly that thread, which Toolhelp lets us reopen safely.
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ThreadEntry32{Size: uint32(unsafe.Sizeof(windows.ThreadEntry32{}))}
	for err := windows.Thread32First(snapshot, &entry); err == nil; err = windows.Thread32Next(snapshot, &entry) {
		if entry.OwnerProcessID != uint32(p.child.Process.Pid) {
			continue
		}
		thread, err := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
		if err != nil {
			return err
		}
		previous, resumeErr := windows.ResumeThread(thread)
		_ = windows.CloseHandle(thread)
		if resumeErr != nil {
			return resumeErr
		}
		if previous != 1 {
			return errors.New("unexpected external editor suspension state")
		}
		return nil
	}
	return errors.New("external editor primary thread unavailable")
}

func (p *secretEditorProcess) signal(sig os.Signal) error {
	if sig == os.Interrupt {
		// Console Ctrl+C is already broadcast to attached processes. Windows
		// does not implement os.Process.Signal(os.Interrupt); do not turn an
		// editor-handled console event into forced termination.
		return nil
	}
	// Windows has no POSIX graceful termination signal. Cancellation uses
	// the job's atomic whole-invocation termination, never direct-child Kill.
	return windows.TerminateJobObject(p.job, 1)
}

func (p *secretEditorProcess) kill() error { return windows.TerminateJobObject(p.job, 1) }

func (p *secretEditorProcess) running() (bool, error) {
	// JOBOBJECT_BASIC_ACCOUNTING_INFORMATION is not exposed by x/sys.
	var accounting struct {
		TotalUserTime             int64
		TotalKernelTime           int64
		ThisPeriodTotalUserTime   int64
		ThisPeriodTotalKernelTime int64
		TotalPageFaultCount       uint32
		TotalProcesses            uint32
		ActiveProcesses           uint32
		TotalTerminatedProcesses  uint32
	}
	if err := windows.QueryInformationJobObject(p.job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err != nil {
		return true, err
	}
	return accounting.ActiveProcesses != 0, nil
}

func (p *secretEditorProcess) restore() error {
	closeErr := windows.CloseHandle(p.job)
	if p.terminal != nil {
		if err := term.Restore(int(p.terminal.Fd()), p.state); err != nil {
			return errors.New("restore editor terminal state failed")
		}
	}
	if closeErr != nil {
		return errors.New("close external editor job failed")
	}
	return nil
}
