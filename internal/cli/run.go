package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/project"
)

type exitCoder interface {
	ExitCode() int
}

type exitCodeError struct {
	code int
}

func (e exitCodeError) Error() string {
	return fmt.Sprintf("child process exited with status %d", e.code)
}

func (e exitCodeError) ExitCode() int {
	return e.code
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr exitCoder
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func newRunCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "run -- command args...",
		Short: "Run a command with project secrets injected",
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := project.Root()
			if err != nil {
				return err
			}
			m, err := project.Load(filepath.Join(root, project.FileName))
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found in %s; run `shelf project init`", project.FileName, root)
			}
			if err != nil {
				return err
			}
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}

			resolvedEntries, diagnostics := project.ResolveEntries(m, st)
			project.RenderDiagnostics(cmd.OutOrStderr(), diagnostics)
			if project.HasFailures(diagnostics) {
				return fmt.Errorf("project run failed")
			}

			if dryRun {
				for _, warning := range project.EnvOverrideWarnings(resolvedEntries, os.Environ()) {
					fmt.Fprintln(cmd.OutOrStderr(), warning)
				}
				for _, entry := range resolvedEntries {
					fmt.Fprintf(cmd.OutOrStdout(), "inject %s\n", entry.EnvName)
				}
				return nil
			}

			child := exec.Command(args[0], args[1:]...)
			child.Env = project.ChildEnv(os.Environ(), resolvedEntries)
			child.Stdin = os.Stdin
			child.Stdout = cmd.OutOrStdout()
			child.Stderr = cmd.OutOrStderr()
			if err := child.Run(); err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					return exitCodeError{code: exitErr.ExitCode()}
				}
				return err
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print injected env names without executing the command")
	return cmd
}
