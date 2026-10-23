package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"golang.org/x/term"
)

var errSecretEditorActive = errors.New("external editor did not stop; private draft retained")

func newSecretEditCmd(appSvc *app.App) *cobra.Command {
	var env, target string
	cmd := &cobra.Command{
		Use:   "edit [selector]",
		Short: "Edit a whole secret item in your external editor",
		Long:  "Edit an existing item as plaintext JSON using VISUAL, EDITOR, or vi. Editor commands support quoted arguments, without shell evaluation. GUI editors must be configured to wait. Private drafts are normally removed after the editor stops; uncertain termination retains them for recovery. Editor backups and histories outside the draft directory are not managed by Shelf.",
		Args: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("env") {
				if strings.TrimSpace(env) == "" || len(args) != 0 {
					return errors.New("use either a selector or --env ENV")
				}
				return nil
			}
			if cmd.Flags().Changed("target") {
				return errors.New("--target requires --env ENV")
			}
			return cobra.ExactArgs(1)(cmd, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return completeSecretSelectors(appSvc, cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, vaultPath := runtimePaths(cmd)
			req := app.SecretEditRequest{Env: env, Target: target}
			if len(args) != 0 {
				req.Selector = args[0]
			}
			if env != "" && !cmd.Flags().Changed("target") {
				targets, err := appSvc.SecretSetTargets(configPath, vaultPath, env)
				if err != nil {
					return err
				}
				if len(targets) > 1 {
					req.Target, err = pickSecretSetTarget(cmd, env, targets)
					if err != nil {
						return err
					}
				}
			}
			session, err := appSvc.BeginSecretEdit(configPath, vaultPath, req)
			if err != nil {
				return err
			}
			return editSecretItem(cmd, appSvc, session)
		},
	}
	cmd.Flags().StringVar(&env, "env", "", "Edit the item owning an existing environment mapping")
	cmd.Flags().StringVar(&target, "target", "", "Concrete item-ID.field mapped to ENV (required for noninteractive ambiguity)")
	_ = cmd.RegisterFlagCompletionFunc("env", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretSetEnvNames})
	})
	_ = cmd.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if env == "" || len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completeSecretMetadata(appSvc, cmd, toComplete, app.SecretCompletionRequest{Target: app.CompleteSecretSetTargets, Env: env})
	})
	return cmd
}

func editSecretItem(cmd *cobra.Command, appSvc *app.App, session *app.SecretEditSession) (result error) {
	command, err := secretEditorCommand()
	if err != nil {
		return err
	}
	interrupts := make(chan os.Signal, 4)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupts)
	dir, err := os.MkdirTemp("", "shelf-edit-")
	if err != nil {
		return errors.New("create private secret edit directory failed")
	}
	defer func() {
		if errors.Is(result, errSecretEditorActive) {
			result = fmt.Errorf("%w: %s", result, dir)
			return
		}
		if err := os.RemoveAll(dir); err != nil {
			if result != nil {
				result = fmt.Errorf("%w; secret edit cleanup failed: %s", result, dir)
			} else {
				result = fmt.Errorf("secret edit cleanup failed: %s", dir)
			}
		}
	}()
	if err := os.Chmod(dir, 0o700); err != nil {
		return errors.New("protect secret edit directory failed")
	}
	path := filepath.Join(dir, "item.json")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return errors.New("create secret edit draft failed")
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return errors.New("protect secret edit draft failed")
	}
	_, writeErr := file.Write(session.Document)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return errors.New("write secret edit draft failed")
	}
	edit := true
	for {
		if err := secretEditInterrupted(cmd.Context(), interrupts, false); err != nil {
			return err
		}
		if edit {
			if err := runSecretEditor(cmd, command, path, interrupts); err != nil {
				return err
			}
		}
		if err := secretEditInterrupted(cmd.Context(), interrupts, true); err != nil {
			return err
		}
		document, err := os.ReadFile(path)
		if err != nil {
			return errors.New("read secret edit draft failed")
		}
		changed, err := appSvc.SaveSecretEdit(session, document)
		if err == nil {
			if changed {
				fmt.Fprintln(cmd.OutOrStdout(), "Updated secret item.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "No changes to secret item.")
			}
			return nil
		}
		if errors.Is(err, app.ErrSecretEditConflict) {
			return errors.New("secret edit conflict: item or schema changed; no edit was saved")
		}
		invalid := errors.Is(err, app.ErrSecretEditInvalidDocument)
		choice, err := secretEditChoice(cmd, invalid, interrupts)
		if err != nil {
			return err
		}
		if choice == "discard" {
			fmt.Fprintln(cmd.OutOrStdout(), "Discarded secret edit.")
			return nil
		}
		edit = choice == "edit"
	}
}

func secretEditorCommand() ([]string, error) {
	value := os.Getenv("VISUAL")
	if strings.TrimSpace(value) == "" {
		value = os.Getenv("EDITOR")
	}
	if strings.TrimSpace(value) == "" {
		value = "vi"
	}
	// Split words only. Shell operators, expansions and substitutions remain
	// literal argument text; no shell is ever invoked by Shelf.
	var words []string
	var word strings.Builder
	var quote rune
	escaped, started := false, false
	for _, char := range value {
		if escaped {
			if quote == '"' && char != '$' && char != '`' && char != '"' && char != '\\' && char != '\n' {
				word.WriteRune('\\')
			}
			if char != '\n' {
				word.WriteRune(char)
				started = true
			}
			escaped = false
			continue
		}
		if char == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			} else {
				word.WriteRune(char)
			}
			continue
		}
		switch {
		case char == '\'' || char == '"':
			quote, started = char, true
		case char == ' ' || char == '\t' || char == '\n':
			if started {
				words = append(words, word.String())
				word.Reset()
				started = false
			}
		default:
			word.WriteRune(char)
			started = true
		}
	}
	if quote != 0 || escaped {
		return nil, errors.New("invalid editor command: unmatched quote or trailing escape")
	}
	if started {
		words = append(words, word.String())
	}
	if len(words) == 0 || words[0] == "" {
		return nil, errors.New("invalid editor command: missing executable")
	}
	return words, nil
}

func runSecretEditor(cmd *cobra.Command, command []string, path string, interrupts <-chan os.Signal) (result error) {
	args := make([]string, len(command))
	copy(args, command[1:])
	args[len(args)-1] = path
	child := exec.Command(command[0], args...)
	child.Stdin, child.Stdout, child.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
	child.WaitDelay = 2 * time.Second
	process, err := prepareSecretEditor(child)
	if err != nil {
		return err
	}
	defer func() {
		if err := process.restore(); err != nil {
			if result == nil {
				result = err
			} else {
				result = fmt.Errorf("%w; restore editor terminal failed", result)
			}
		}
	}()
	if err := process.start(); err != nil {
		if errors.Is(err, errSecretEditorActive) {
			return err
		}
		return errors.New("start external editor failed")
	}
	done := make(chan error, 1)
	go func() { done <- child.Wait() }()
	var waiting <-chan error = done
	var ticker *time.Ticker
	var groupTick <-chan time.Time
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()
	for {
		if waiting == nil {
			running, err := process.running()
			if err != nil {
				if stopErr := stopSecretEditor(process, waiting, syscall.SIGTERM); stopErr != nil {
					return stopErr
				}
				return errors.New("inspect external editor invocation failed")
			}
			if !running {
				return nil
			}
			if ticker == nil {
				ticker = time.NewTicker(25 * time.Millisecond)
				groupTick = ticker.C
			}
		}
		select {
		case err := <-waiting:
			waiting = nil
			if err != nil {
				if err := stopSecretEditor(process, waiting, syscall.SIGTERM); err != nil {
					return err
				}
				return errors.New("external editor failed; no edit was saved")
			}
		case sig := <-interrupts:
			if sig == os.Interrupt {
				// Unix terminals deliver Ctrl+C directly to the editor's
				// foreground group. Forward parent-only interrupts too, but
				// never remove its draft until the entire invocation exits.
				_ = process.signal(sig)
				continue
			}
			if err := stopSecretEditor(process, waiting, sig); err != nil {
				return err
			}
			return errors.New("secret edit interrupted; no edit was saved")
		case <-cmd.Context().Done():
			if err := stopSecretEditor(process, waiting, syscall.SIGTERM); err != nil {
				return err
			}
			return errors.New("secret edit interrupted; no edit was saved")
		case <-groupTick:
		}
	}
}

func stopSecretEditor(process *secretEditorProcess, done <-chan error, sig os.Signal) error {
	_ = process.signal(sig)
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	killed := false
	for {
		running, err := process.running()
		if err == nil && !running && done == nil {
			return nil
		}
		select {
		case <-done:
			done = nil
		case <-ticker.C:
		case <-timer.C:
			if killed {
				return errSecretEditorActive
			}
			_ = process.kill()
			killed = true
			timer.Reset(2 * time.Second)
		}
	}
}

func secretEditInterrupted(ctx context.Context, interrupts <-chan os.Signal, editorFinished bool) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("secret edit interrupted; no edit was saved")
		case sig := <-interrupts:
			// Delivery to the parent can lag the editor's successful exit.
			// A Ctrl+C handled by that editor must not cancel its result.
			if editorFinished && sig == os.Interrupt {
				continue
			}
			return errors.New("secret edit interrupted; no edit was saved")
		default:
			return nil
		}
	}
}

func secretEditChoice(cmd *cobra.Command, invalid bool, interrupts <-chan os.Signal) (string, error) {
	failure := "save secret edit failed"
	prompt := "Could not save secret edit. [r]etry save, [e]dit, or [d]iscard: "
	if invalid {
		failure = "secret edit document is invalid"
		prompt = "Secret edit document is invalid. [e]dit or [d]iscard: "
	}
	input, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(input.Fd())) {
		return "", errors.New(failure + "; no edit was saved (use an interactive terminal to retry)")
	}
	for {
		fmt.Fprint(cmd.ErrOrStderr(), prompt)
		// Do not prefetch input intended for the next editor invocation.
		line := make(chan string, 1)
		go func() {
			var text strings.Builder
			var one [1]byte
			for {
				if _, err := io.ReadFull(input, one[:]); err != nil {
					line <- ""
					return
				}
				if one[0] == '\n' {
					line <- strings.TrimSpace(text.String())
					return
				}
				text.WriteByte(one[0])
			}
		}()
		select {
		case <-cmd.Context().Done():
			return "", errors.New("secret edit interrupted; no edit was saved")
		case <-interrupts:
			return "", errors.New("secret edit interrupted; no edit was saved")
		case choice := <-line:
			switch strings.ToLower(choice) {
			case "e", "edit":
				return "edit", nil
			case "d", "discard":
				return "discard", nil
			case "r", "retry":
				if !invalid {
					return "retry", nil
				}
			case "":
				return "", errors.New(failure + "; secret edit cancelled without saving")
			}
		}
	}
}
