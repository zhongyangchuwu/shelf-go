package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
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

func newRunCmd(appSvc *app.App) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "run -- command args...",
		Short: "Run a command with project secrets injected",
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			err := appSvc.ProjectRun(app.ProjectRunRequest{
				ConfigPath: configPath,
				VaultPath:  vaultPath,
				Command:    args,
				DryRun:     dryRun,
				ParentEnv:  os.Environ(),
				Stdin:      os.Stdin,
				Stdout:     cmd.OutOrStdout(),
				Stderr:     cmd.OutOrStderr(),
			})
			var childExit app.ChildExitError
			if errors.As(err, &childExit) {
				return exitCodeError{code: childExit.Code}
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print injected env names without executing the command")
	return cmd
}
