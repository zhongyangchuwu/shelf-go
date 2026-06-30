package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zhongyangchuwu/shelf-go/internal/project"
)

type ProjectRunRequest struct {
	ConfigPath string
	VaultPath  string
	Command    []string
	DryRun     bool
	ParentEnv  []string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
}

type ChildExitError struct {
	Code int
}

func (e ChildExitError) Error() string {
	return fmt.Sprintf("child process exited with status %d", e.Code)
}

func (a *App) ProjectRun(req ProjectRunRequest) error {
	root, err := project.Root()
	if err != nil {
		return err
	}
	manifest, err := project.Load(filepath.Join(root, project.FileName))
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s not found in %s; run `shelf project init`", project.FileName, root)
	}
	if err != nil {
		return err
	}
	_, st, err := a.LoadRuntime(req.ConfigPath, req.VaultPath)
	if err != nil {
		return err
	}

	resolvedEntries, diagnostics := project.ResolveEntries(manifest, st)
	project.RenderDiagnostics(req.Stderr, diagnostics)
	if project.HasFailures(diagnostics) {
		return fmt.Errorf("project run failed")
	}

	if req.DryRun {
		for _, warning := range project.EnvOverrideWarnings(resolvedEntries, req.ParentEnv) {
			fmt.Fprintln(req.Stderr, warning)
		}
		for _, entry := range resolvedEntries {
			fmt.Fprintf(req.Stdout, "inject %s\n", entry.EnvName)
		}
		return nil
	}

	if len(req.Command) == 0 {
		return fmt.Errorf("command is required")
	}
	child := exec.Command(req.Command[0], req.Command[1:]...)
	child.Env = project.ChildEnv(req.ParentEnv, resolvedEntries)
	child.Stdin = req.Stdin
	child.Stdout = req.Stdout
	child.Stderr = req.Stderr
	if err := child.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ChildExitError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}
