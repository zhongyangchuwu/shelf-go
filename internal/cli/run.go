package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/manifest"
	projectsvc "github.com/zhongyangchuwu/shelf-go/internal/project"
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
			root, err := projectsvc.Root()
			if err != nil {
				return err
			}
			m, err := manifest.Load(filepath.Join(root, manifest.FileName))
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found in %s; run `shelf project init`", manifest.FileName, root)
			}
			if err != nil {
				return err
			}
			_, st, err := loadRuntime(cmd)
			if err != nil {
				return err
			}

			resolvedEntries, diagnostics := projectsvc.ResolveEntries(m, st)
			projectsvc.RenderDiagnostics(cmd.OutOrStderr(), diagnostics)
			if projectsvc.HasFailures(diagnostics) {
				return fmt.Errorf("project run failed")
			}

			if dryRun {
				for _, warning := range envOverrideWarnings(resolvedEntries, os.Environ()) {
					fmt.Fprintln(cmd.OutOrStderr(), warning)
				}
				for _, entry := range resolvedEntries {
					fmt.Fprintf(cmd.OutOrStdout(), "inject %s\n", entry.EnvName)
				}
				return nil
			}

			child := exec.Command(args[0], args[1:]...)
			child.Env = childEnv(os.Environ(), resolvedEntries)
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

func childEnv(parent []string, entries []projectsvc.Binding) []string {
	values := make(map[string]string, len(entries))
	for _, entry := range entries {
		values[entry.EnvName] = entry.Value
	}

	out := make([]string, 0, len(parent)+len(entries))
	seen := make(map[string]struct{}, len(parent)+len(entries))
	for _, item := range parent {
		key, _, ok := splitEnv(item)
		if !ok {
			if _, exists := values[item]; exists {
				continue
			}
			out = append(out, item)
			continue
		}
		if value, exists := values[key]; exists {
			out = append(out, key+"="+value)
			seen[key] = struct{}{}
			continue
		}
		out = append(out, item)
		seen[key] = struct{}{}
	}
	for _, entry := range entries {
		if _, exists := seen[entry.EnvName]; exists {
			continue
		}
		out = append(out, entry.EnvName+"="+entry.Value)
	}
	return out
}

func envOverrideWarnings(entries []projectsvc.Binding, parent []string) []string {
	parentNames := make(map[string]struct{}, len(parent))
	for _, item := range parent {
		key, _, ok := splitEnv(item)
		if ok {
			parentNames[key] = struct{}{}
		}
	}
	warnings := make([]string, 0)
	for _, entry := range entries {
		if _, exists := parentNames[entry.EnvName]; exists {
			warnings = append(warnings, fmt.Sprintf("warn %s overrides existing environment variable", entry.EnvName))
		}
	}
	return warnings
}

func splitEnv(item string) (string, string, bool) {
	for i := 0; i < len(item); i++ {
		if item[i] == '=' {
			return item[:i], item[i+1:], true
		}
	}
	return "", "", false
}
